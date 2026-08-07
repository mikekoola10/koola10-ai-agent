import type { NovaConfig } from "../config.js";
import type { ToolCall, ToolDefinition } from "../types.js";
import { runCommand } from "./shell.js";
import { listDirectory, readFile, writeFile } from "./files.js";
import { fetchUrl, webSearch } from "./web.js";
import { githubApi } from "./github.js";
import { stripeApi } from "./stripe.js";
import { clawdbot, type ClawdbotAction } from "./clawdbot.js";
import { browserUse, type BrowserAction } from "./browser.js";
import { computerUse, type ComputerAction } from "./computer.js";
import { composioTool, checkComposio } from "./composio.js";
import { automationTool, checkAutomation } from "./automations.js";
import { vaultTool } from "./vault.js";
import { huggingfaceTool, checkHuggingface } from "./huggingface.js";

type Props = Record<string, unknown>;

const fn = (name: string, description: string, properties: Props, required: string[] = []): ToolDefinition => ({
  type: "function",
  function: {
    name,
    description,
    parameters: {
      type: "object",
      properties,
      ...(required.length ? { required } : {}),
    },
  },
});

const automationProperties: Props = {
  action: { type: "string", enum: ["trigger", "ping"] },
  payload: { type: "object", description: "JSON payload sent to the provider webhook" },
};

/** Tool schemas sent to the model. */
export function buildToolDefinitions(): ToolDefinition[] {
  return [
    fn("run_command", "Run a bash command in the working directory.", { command: { type: "string" } }, ["command"]),
    fn("list_directory", "List one directory level.", { path: { type: "string" } }),
    fn("read_file", "Read a UTF-8 text file.", { path: { type: "string" } }, ["path"]),
    fn("write_file", "Write a text file, creating parent directories.", { path: { type: "string" }, content: { type: "string" } }, ["path", "content"]),
    fn("web_search", "Search the web for current information.", { query: { type: "string" } }, ["query"]),
    fn("fetch_url", "Fetch a web page and return readable text.", { url: { type: "string" } }, ["url"]),
    fn(
      "github",
      "Call the GitHub REST API using GITHUB_TOKEN. Prefer read-only requests unless a write is explicitly requested.",
      { path: { type: "string" }, method: { type: "string" }, body: { type: "object" } },
      ["path"],
    ),
    fn(
      "stripe",
      "Call the Stripe REST API using STRIPE_SECRET_KEY. Prefer read-only endpoints; never create live charges without explicit confirmation.",
      { path: { type: "string" }, method: { type: "string" }, body: { type: "object" } },
      ["path"],
    ),
    fn(
      "clawdbot",
      "Send a message or dispatch a task through the local OpenClaw/Clawdbot CLI.",
      { action: { type: "string" }, contact: { type: "string" }, message: { type: "string" } },
      ["action", "message"],
    ),
    fn(
      "browser",
      "Use a real Chromium browser through Playwright: open, click, type, extract, screenshot, back, or close.",
      { action: { type: "string" }, url: { type: "string" }, selector: { type: "string" }, text: { type: "string" }, path: { type: "string" } },
      ["action"],
    ),
    fn(
      "computer",
      "Control the local X11 desktop through xdotool.",
      { action: { type: "string" }, text: { type: "string" }, key: { type: "string" }, x: { type: "integer" }, y: { type: "integer" }, path: { type: "string" } },
      ["action"],
    ),
    fn(
      "composio",
      "Use Composio's connected app tools across GitHub, Stripe, Slack, Gmail, and other apps. Prefer read-only tools unless the user explicitly requests a side effect. 'ping' tests the connection.",
      { action: { type: "string", enum: ["list", "execute", "ping"] }, toolkit: { type: "string" }, tool: { type: "string" }, arguments: { type: "object" }, userId: { type: "string" } },
      ["action"],
    ),
    fn(
      "zapier",
      "Trigger the configured Zapier Catch Hook with a JSON payload, or 'ping' to test the connection. Use only when a Zapier webhook URL is configured and the user has authorized the workflow.",
      automationProperties,
      ["action"],
    ),
    fn(
      "n8n",
      "Trigger the configured n8n Cloud production webhook with a JSON payload, or 'ping' to test the connection. Use only when the workflow URL is configured and the user has authorized the workflow.",
      automationProperties,
      ["action"],
    ),
    fn(
      "make",
      "Trigger the configured Make custom webhook with a JSON payload, or 'ping' to test the connection. Use only when the scenario webhook URL is configured and the user has authorized the scenario.",
      automationProperties,
      ["action"],
    ),
    fn(
      "vault",
      "Store, list, read, or delete credentials in Nova's encrypted vault. Use to save API keys and webhook URLs during a job. 'list' returns names only; 'get' returns the secret value (only use when the task needs it); 'set' stores a new value; 'delete' removes one.",
      { action: { type: "string", enum: ["list", "set", "get", "delete"] }, name: { type: "string" }, value: { type: "string" } },
      ["action"],
    ),
    fn(
      "huggingface",
      "Use the Hugging Face API. 'ping' verifies the HUGGINGFACE_TOKEN; 'models' searches the public model Hub (no key needed); 'infer' runs a chat completion through the Hugging Face Inference API (OpenAI-compatible) using the token. Use 'infer' for free/cheap open-model completions and embeddings-style tasks; pick a model with the 'model' argument (e.g. HuggingFaceTB/SmolLM2-1.7B-Instruct, mistralai/Mistral-7B-Instruct-v0.3).",
      {
        action: { type: "string", enum: ["ping", "models", "infer"] },
        query: { type: "string", description: "For models: Hub search query, e.g. 'instruction tuned' or 'embedding'" },
        prompt: { type: "string", description: "For infer: the user prompt to complete" },
        model: { type: "string", description: "For infer: model id (defaults to HuggingFaceTB/SmolLM2-1.7B-Instruct)" },
        maxTokens: { type: "integer", description: "For infer: max output tokens (default 512, max 2048)" },
        limit: { type: "integer", description: "For models: max results (default 5, max 10)" },
      },
      ["action"],
    ),
  ];
}

export type ConnectorCheck = {
  provider: string;
  configured: boolean;
  ok: boolean;
  status?: number;
  detail: string;
};

/**
 * Live end-to-end verification of every configured Nova connector.
 * Pings the automation webhooks, Composio, and Hugging Face; reports key
 * presence for the native connectors. Used by `nova --verify-connectors`
 * and the UI endpoint.
 */
export async function verifyConnectors(config: NovaConfig): Promise<ConnectorCheck[]> {
  const present = (value: string): boolean => Boolean(value);
  const checks: ConnectorCheck[] = [
    {
      provider: "github",
      configured: present(config.githubToken),
      ok: present(config.githubToken),
      detail: present(config.githubToken) ? "key present (no live call made)" : "GITHUB_TOKEN not set",
    },
    {
      provider: "stripe",
      configured: present(config.stripeKey),
      ok: present(config.stripeKey),
      detail: present(config.stripeKey) ? "key present (no live call made)" : "STRIPE_SECRET_KEY not set",
    },
    {
      provider: "clawdbot",
      configured: present(config.clawdbotCli),
      ok: present(config.clawdbotCli),
      detail: present(config.clawdbotCli) ? "CLI configured (not invoked)" : "CLAWDBOT_CLI not set",
    },
  ];
  checks.push(await checkComposio(config));
  checks.push(await checkAutomation(config, "zapier"));
  checks.push(await checkAutomation(config, "n8n"));
  checks.push(await checkAutomation(config, "make"));
  checks.push(await checkHuggingface(config));
  return checks;
}

/** Executes a tool call and returns its output string (never throws). */
export async function dispatchTool(call: ToolCall, config: NovaConfig): Promise<string> {
  let args: Record<string, unknown>;
  try {
    args = JSON.parse(call.arguments || "{}") as Record<string, unknown>;
  } catch {
    return `ERROR: invalid JSON arguments for tool "${call.name}".`;
  }

  const str = (v: unknown, fallback = ""): string =>
    typeof v === "string" ? v : v == null ? fallback : JSON.stringify(v);
  const method = str(args.method, "GET").toUpperCase();
  const body = args.body && typeof args.body === "object" ? (args.body as Record<string, unknown>) : {};

  switch (call.name) {
    case "run_command":
      return runCommand(str(args.command), {
        cwd: config.cwd,
        timeoutMs: config.toolTimeoutMs,
        maxChars: config.maxToolOutputChars,
      });
    case "list_directory":
      return listDirectory(str(args.path, "."));
    case "read_file":
      return readFile(str(args.path), config.maxToolOutputChars);
    case "write_file":
      return writeFile(str(args.path), str(args.content));
    case "web_search":
      return webSearch(str(args.query));
    case "fetch_url":
      return fetchUrl(str(args.url), config.maxToolOutputChars);
    case "github":
      return githubApi(config, str(args.path), method, body);
    case "stripe":
      return stripeApi(config, str(args.path), method, body);
    case "clawdbot":
      return clawdbot(config, str(args.action) as ClawdbotAction, str(args.contact), str(args.message));
    case "browser":
      return browserUse(str(args.action) as BrowserAction, args, {
        headless: config.browserHeadless,
        maxChars: config.maxToolOutputChars,
      });
    case "computer":
      return computerUse(str(args.action) as ComputerAction, args);
    case "composio":
      return composioTool(config, str(args.action), args);
    case "zapier":
      return automationTool(config, "zapier", str(args.action), args);
    case "n8n":
      return automationTool(config, "n8n", str(args.action), args);
    case "make":
      return automationTool(config, "make", str(args.action), args);
    case "vault":
      return vaultTool(str(args.action), args);
    case "huggingface":
      return huggingfaceTool(config, str(args.action), args);
    default:
      return `ERROR: unknown tool "${call.name}".`;
  }
}
