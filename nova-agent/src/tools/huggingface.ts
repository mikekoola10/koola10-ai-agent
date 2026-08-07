import type { NovaConfig } from "../config.js";

/**
 * Hugging Face connector for Nova.
 *
 * - `ping`   verifies the access token via GET /api/whoami-v2 (Bearer auth).
 * - `models` searches the public Hub API (no auth required).
 * - `infer`  runs a chat completion through the unified Inference API
 *            (POST https://router.huggingface.co/v1/chat/completions,
 *            OpenAI-compatible, Bearer auth).
 *
 * Uses node's native fetch — no SDK. The token is server-side only and never
 * exposed to the browser.
 */

type HfConfig = NovaConfig & { huggingfaceApiKey: string };

export type HfCheck = {
  provider: "huggingface";
  configured: boolean;
  ok: boolean;
  status?: number;
  detail: string;
};

const INFER_URL = "https://router.huggingface.co/v1/chat/completions";
const WHOAMI_URL = "https://huggingface.co/api/whoami-v2";
const HUB_API = "https://huggingface.co/api";
const DEFAULT_MODEL = "HuggingFaceTB/SmolLM2-1.7B-Instruct";

function formatResult(value: unknown, maxChars: number): string {
  const text = typeof value === "string" ? value : JSON.stringify(value, null, 2);
  if (!text) return "(empty response)";
  return text.length <= maxChars ? text : `${text.slice(0, maxChars)}\n… [truncated]`;
}

function errMsg(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

/** Live token verification (powers `--verify-connectors` and the `ping` action). */
export async function checkHuggingface(config: HfConfig): Promise<HfCheck> {
  const key = config.huggingfaceApiKey;
  if (!key) return { provider: "huggingface", configured: false, ok: false, detail: "HUGGINGFACE_TOKEN is not set" };
  try {
    const res = await fetch(WHOAMI_URL, {
      headers: { Authorization: `Bearer ${key}` },
      signal: AbortSignal.timeout(config.toolTimeoutMs),
    });
    if (res.status === 200) {
      let name = "token";
      try {
        const data = (await res.json()) as { name?: string; orgs?: unknown };
        if (data.name) name = data.name;
      } catch {
        /* non-JSON body */
      }
      return { provider: "huggingface", configured: true, ok: true, status: 200, detail: `authenticated as ${name}` };
    }
    if (res.status === 401 || res.status === 403) {
      return { provider: "huggingface", configured: true, ok: false, status: res.status, detail: `authentication failed (${res.status}) — check HUGGINGFACE_TOKEN` };
    }
    return { provider: "huggingface", configured: true, ok: res.ok, status: res.status, detail: `HTTP ${res.status}` };
  } catch (error) {
    return { provider: "huggingface", configured: true, ok: false, detail: `request failed — ${errMsg(error)}` };
  }
}

export async function huggingfaceTool(config: HfConfig, action: string, args: Record<string, unknown> = {}): Promise<string> {
  const key = config.huggingfaceApiKey;
  const str = (v: unknown, fallback = "") => (typeof v === "string" ? v : v == null ? fallback : String(v));
  const num = (v: unknown, fallback: number, max: number) => {
    const n = typeof v === "number" ? v : Number.parseInt(String(v ?? ""), 10);
    return Number.isFinite(n) && n > 0 ? Math.min(n, max) : fallback;
  };

  switch (action) {
    case "ping": {
      if (!key) return "ERROR: HUGGINGFACE_TOKEN is not set. Add it to the server environment or the vault.";
      return formatResult(await checkHuggingface(config), config.maxToolOutputChars);
    }
    case "models": {
      const query = str(args.query, "").trim();
      const limit = num(args.limit, 5, 10);
      const url = `${HUB_API}/models?search=${encodeURIComponent(query)}&limit=${limit}&sort=downloads&direction=-1`;
      try {
        const res = await fetch(url, { signal: AbortSignal.timeout(config.toolTimeoutMs) });
        const text = await res.text();
        let data: Array<{ id?: string; downloads?: number; likes?: number }> = [];
        try {
          data = JSON.parse(text);
        } catch {
          return `ERROR: Hugging Face Hub search failed — ${text.slice(0, 160)}`;
        }
        if (!Array.isArray(data) || data.length === 0) return `No models found for "${query}".`;
        const lines = data.map((m, i) => `${i + 1}. ${m.id ?? "?"} — ${m.downloads ?? 0} downloads · ${m.likes ?? 0} likes`);
        return `Top models for "${query}":\n${lines.join("\n")}`;
      } catch (error) {
        return `ERROR: Hub search failed — ${errMsg(error)}`;
      }
    }
    case "infer": {
      if (!key) return "ERROR: HUGGINGFACE_TOKEN is not set. Add it to the server environment or the vault.";
      const prompt = str(args.prompt, "").trim();
      if (!prompt) return "ERROR: huggingface infer requires a prompt.";
      const model = str(args.model, DEFAULT_MODEL).trim() || DEFAULT_MODEL;
      const maxTokens = num(args.maxTokens, 512, 2048);
      try {
        const res = await fetch(INFER_URL, {
          method: "POST",
          headers: { Authorization: `Bearer ${key}`, "content-type": "application/json" },
          body: JSON.stringify({ model, messages: [{ role: "user", content: prompt }], max_tokens: maxTokens }),
          signal: AbortSignal.timeout(config.toolTimeoutMs),
        });
        const text = await res.text();
        let data: { choices?: Array<{ message?: { content?: string } }>; error?: unknown } = {};
        try {
          data = JSON.parse(text);
        } catch {
          data = { error: text.slice(0, 200) };
        }
        if (!res.ok) {
          const detail = typeof data.error === "string" ? data.error : errMsg(data.error ?? `HTTP ${res.status}`);
          return `ERROR: inference failed (${res.status}) — ${detail}. The model may be gated (accept its license on the Hub) or the token may lack inference access.`;
        }
        const content = data.choices?.[0]?.message?.content ?? "(no content)";
        return formatResult({ model, status: res.status, output: content }, config.maxToolOutputChars);
      } catch (error) {
        return `ERROR: inference failed — ${errMsg(error)}`;
      }
    }
    default:
      return "ERROR: huggingface supports actions: ping, models, infer.";
  }
}
