import { createCipheriv, createDecipheriv, createHash, randomBytes } from "node:crypto";
import { chmodSync, existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import type { NovaConfig } from "../config.js";

/**
 * Nova credential vault — encrypted storage for API keys and webhook URLs.
 *
 * - Encrypted at rest with AES-256-GCM.
 * - Master key: NOVA_VAULT_KEY env var (any length; hashed to 32 bytes) when
 *   set, otherwise an auto-generated random key at <vault>/.master.key
 *   (chmod 600). Auto key works out of the box; set NOVA_VAULT_KEY if the
 *   vault must survive moving to another machine.
 * - Storage: <NOVA_VAULT_DIR|./vault>/keys.json (chmod 600), gitignored.
 * - The agent sees one `vault` tool (list/set/get/delete). Values are never
 *   exposed to the browser through the UI API.
 */

const ALGO = "aes-256-gcm";
const NAME_RE = /^[A-Za-z0-9_]{1,64}$/;

type VaultEntry = { v: string; created: number; updated: number };
type VaultData = { entries: Record<string, VaultEntry> };

function vaultDir(): string {
  return process.env.NOVA_VAULT_DIR ? process.env.NOVA_VAULT_DIR : join(process.cwd(), "vault");
}

function masterKey(): Buffer {
  const fromEnv = process.env.NOVA_VAULT_KEY;
  if (fromEnv) return createHash("sha256").update(fromEnv).digest();
  const dir = vaultDir();
  mkdirSync(dir, { recursive: true });
  const keyPath = join(dir, ".master.key");
  if (existsSync(keyPath)) return readFileSync(keyPath);
  const key = randomBytes(32);
  writeFileSync(keyPath, key, { mode: 0o600 });
  try {
    chmodSync(keyPath, 0o600);
  } catch {
    /* best effort */
  }
  return key;
}

function storePath(): string {
  return join(vaultDir(), "keys.json");
}

function load(): VaultData {
  try {
    const path = storePath();
    if (!existsSync(path)) return { entries: {} };
    return JSON.parse(readFileSync(path, "utf8")) as VaultData;
  } catch {
    return { entries: {} };
  }
}

function save(data: VaultData): void {
  const dir = vaultDir();
  mkdirSync(dir, { recursive: true });
  writeFileSync(storePath(), JSON.stringify(data, null, 2), { mode: 0o600 });
}

function encrypt(value: string): string {
  const key = masterKey();
  const iv = randomBytes(12);
  const cipher = createCipheriv(ALGO, key, iv);
  const enc = Buffer.concat([cipher.update(value, "utf8"), cipher.final()]);
  const tag = cipher.getAuthTag();
  return `v1:${iv.toString("base64")}:${tag.toString("base64")}:${enc.toString("base64")}`;
}

function decrypt(payload: string): string {
  const parts = payload.split(":");
  if (parts.length !== 4 || parts[0] !== "v1") throw new Error("unsupported vault entry format");
  const key = masterKey();
  const decipher = createDecipheriv(ALGO, key, Buffer.from(parts[1], "base64"));
  decipher.setAuthTag(Buffer.from(parts[2], "base64"));
  return Buffer.concat([decipher.update(Buffer.from(parts[3], "base64")), decipher.final()]).toString("utf8");
}

/* ---------------- Public API ---------------- */

export function vaultInfo(): { dir: string; count: number; usingEnvKey: boolean } {
  return { dir: vaultDir(), count: vaultList().length, usingEnvKey: Boolean(process.env.NOVA_VAULT_KEY) };
}

export function vaultList(): string[] {
  return Object.keys(load().entries).sort();
}

export function vaultSet(name: string, value: string): { ok: boolean; error?: string } {
  if (!NAME_RE.test(name)) return { ok: false, error: `vault name must match ${NAME_RE.source}` };
  if (!value) return { ok: false, error: "value is required" };
  const data = load();
  const now = Date.now();
  const existing = data.entries[name];
  data.entries[name] = { v: encrypt(value), created: existing?.created ?? now, updated: now };
  save(data);
  return { ok: true };
}

export function vaultGet(name: string): string | null {
  const data = load();
  const entry = data.entries[name];
  if (!entry) return null;
  try {
    return decrypt(entry.v);
  } catch {
    return null;
  }
}

export function vaultDelete(name: string): boolean {
  const data = load();
  if (!data.entries[name]) return false;
  delete data.entries[name];
  save(data);
  return true;
}

/* ---------------- Agent tool ---------------- */

export function vaultTool(action: string, args: Record<string, unknown> = {}): string {
  const str = (v: unknown) => (typeof v === "string" ? v : "");
  switch (action) {
    case "list": {
      const names = vaultList();
      return names.length
        ? `Vault entries (${names.length}):\n${names.map((n) => `  • ${n}`).join("\n")}`
        : "Vault is empty. Use action=set with a name and value to store a credential.";
    }
    case "set": {
      const name = str(args.name);
      const value = str(args.value);
      if (!name || !value) return "ERROR: vault set requires name and value.";
      const res = vaultSet(name, value);
      return res.ok ? `Stored ${name} in the vault (encrypted at rest).` : `ERROR: ${res.error}`;
    }
    case "get": {
      const name = str(args.name);
      if (!name) return "ERROR: vault get requires name.";
      const value = vaultGet(name);
      return value === null ? `No vault entry named ${name}.` : value;
    }
    case "delete": {
      const name = str(args.name);
      if (!name) return "ERROR: vault delete requires name.";
      return vaultDelete(name) ? `Deleted ${name} from the vault.` : `No vault entry named ${name}.`;
    }
    default:
      return "ERROR: vault supports actions: list, set, get, delete.";
  }
}

/* ---------------- Connector config overrides ---------------- */
/* Connector values stored in the vault under their env-var name are used as a
 * fallback when the corresponding env var is not set, so Nova can "connect
 * things" from the vault alone (e.g. store ZAPIER_WEBHOOK_URL once). */

const CONNECTOR_FIELDS: Array<[string, string]> = [
  ["GITHUB_TOKEN", "githubToken"],
  ["STRIPE_SECRET_KEY", "stripeKey"],
  ["COMPOSIO_API_KEY", "composioApiKey"],
  ["ZAPIER_WEBHOOK_URL", "zapierWebhookUrl"],
  ["N8N_WEBHOOK_URL", "n8nWebhookUrl"],
  ["N8N_API_KEY", "n8nApiKey"],
  ["MAKE_WEBHOOK_URL", "makeWebhookUrl"],
  ["HUGGINGFACE_TOKEN", "huggingfaceApiKey"],
];

export function applyVaultOverrides(config: NovaConfig): NovaConfig {
  const out: NovaConfig = { ...config };
  const record = out as unknown as Record<string, string | boolean>;
  for (const [envName, field] of CONNECTOR_FIELDS) {
    if (!record[field]) {
      try {
        const value = vaultGet(envName);
        if (value !== null) record[field] = value;
      } catch {
        /* vault unavailable — keep env value */
      }
    }
  }
  return out;
}
