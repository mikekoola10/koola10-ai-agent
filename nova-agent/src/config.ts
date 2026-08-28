import { existsSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";

/**
 * Loads .env / .env.local from `cwd` and each ancestor up to the repo root
 * without clobbering existing env vars. Nearer files win over farther ones;
 * already-set process.env values (e.g. Freebuff Keys injected into the
 * preview/deploy runtime) always win over files.
 */
export function loadDotEnv(cwd: string): void {
  const dirs: string[] = [];
  let dir = cwd;
  for (let i = 0; i < 8 && dir; i++) {
    dirs.push(dir);
    const parent = dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  // Walk from farthest to nearest so closer .env files take precedence.
  for (const d of dirs.reverse()) {
    for (const file of [".env.local", ".env"]) {
      const p = join(d, file);
      if (!existsSync(p)) continue;
      const text = readFileSync(p, "utf8");
      for (const line of text.split(/\r?\n/)) {
        const trimmed = line.trim();
        if (!trimmed || trimmed.startsWith("#")) continue;
        const eq = trimmed.indexOf("=");
        if (eq <= 0) continue;
        const key = trimmed.slice(0, eq).trim();
        let value = trimmed.slice(eq + 1).trim();
        if (
          (value.startsWith('"') && value.endsWith('"')) ||
          (value.startsWith("'") && value.endsWith("'"))
        ) {
          value = value.slice(1, -1);
        }
        if (!(key in process.env)) process.env[key] = value;
      }
    }
  }
}

export type Provider = "deepseek" | "anthropic" | "openai";

export interface NovaConfig {
  provider: Provider;
  /** API key for the active model provider. */
  apiKey: string;
  apiBase: string;
  model: string;
  maxSteps: number;
  maxToolOutputChars: number;
  toolTimeoutMs: number;
  mock: boolean;
  verbose: boolean;
  saveReport: boolean;
  cwd: string;
  /** Connector secrets (optional). */
  githubToken: string;
  stripeKey: string;
  clawdbotCli: string;
  /** Browser tool settings. */
  browserHeadless: boolean;
  /** Shared Composio connector. */
  composioApiKey: string;
  composioUserId: string;
  /** Automation connectors. */
  zapierWebhookUrl: string;
  n8nWebhookUrl: string;
  n8nApiKey: string;
  makeWebhookUrl: string;
  /** Hugging Face access token. */
  huggingfaceApiKey: string;
}

function env(name: string): string | undefined {
  const v = process.env[name];
  return v && v.trim() !== "" ? v : undefined;
}

function intEnv(name: string, fallback: number): number {
  const v = env(name);
  if (!v) return fallback;
  const n = Number.parseInt(v, 10);
  return Number.isFinite(n) && n > 0 ? n : fallback;
}

function boolEnv(name: string, fallback = false): boolean {
  const v = env(name);
  if (!v) return fallback;
  return !["0", "false", "no", "off"].includes(v.toLowerCase());
}

function normalizeProvider(v: string | undefined): Provider {
  if (v === "anthropic" || v === "openai" || v === "deepseek") return v;
  return "deepseek";
}

export interface CliFlags {
  mock?: boolean;
  verbose?: boolean;
  model?: string;
  provider?: string;
  maxSteps?: string;
  saveReport?: boolean;
  cwd?: string;
}

const DEFAULT_MODEL: Record<Provider, string> = {
  deepseek: "deepseek-chat",
  anthropic: "claude-sonnet-4-5",
  openai: "gpt-4o-mini",
};

const DEFAULT_BASE: Record<Provider, string> = {
  deepseek: "https://api.deepseek.com",
  anthropic: "https://api.anthropic.com",
  openai: "https://api.openai.com/v1",
};

const KEY_ENV: Record<Provider, string> = {
  deepseek: "DEEPSEEK_API_KEY",
  anthropic: "ANTHROPIC_API_KEY",
  openai: "OPENAI_API_KEY",
};

export function keyEnvFor(provider: Provider): string {
  return KEY_ENV[provider];
}

export function loadConfig(flags: CliFlags = {}): NovaConfig {
  const cwd = flags.cwd ?? process.cwd();
  const provider = normalizeProvider(flags.provider ?? env("NOVA_PROVIDER"));
  const maxToolOutputChars = intEnv("NOVA_MAX_TOOL_OUTPUT", 8000);
  return {
    provider,
    apiKey: env(KEY_ENV[provider]) ?? "",
    apiBase: (env("NOVA_API_BASE") ?? DEFAULT_BASE[provider]).replace(/\/+$/, ""),
    model: flags.model ?? env("NOVA_MODEL") ?? DEFAULT_MODEL[provider],
    maxSteps: flags.maxSteps ? Number.parseInt(flags.maxSteps, 10) : intEnv("NOVA_MAX_STEPS", 50),
    maxToolOutputChars,
    toolTimeoutMs: intEnv("NOVA_TOOL_TIMEOUT_MS", 60_000),
    mock: flags.mock ?? boolEnv("NOVA_MOCK"),
    verbose: flags.verbose ?? false,
    saveReport: flags.saveReport ?? boolEnv("NOVA_SAVE_REPORT", true),
    cwd,
    githubToken: env("GITHUB_TOKEN") ?? "",
    stripeKey: env("STRIPE_SECRET_KEY") ?? "",
    clawdbotCli: env("CLAWDBOT_CLI") ?? "openclaw",
    browserHeadless: boolEnv("NOVA_BROWSER_HEADLESS", true),
    composioApiKey: env("COMPOSIO_API_KEY") ?? "",
    composioUserId: env("COMPOSIO_USER_ID") ?? "nova-user",
    zapierWebhookUrl: env("ZAPIER_WEBHOOK_URL") ?? "",
    n8nWebhookUrl: env("N8N_WEBHOOK_URL") ?? "",
    n8nApiKey: env("N8N_API_KEY") ?? "",
    makeWebhookUrl: env("MAKE_WEBHOOK_URL") ?? "",
    huggingfaceApiKey: env("HUGGINGFACE_TOKEN") ?? "",
  };
}
