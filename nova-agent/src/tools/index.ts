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

/** Tool schemas sent to the model. */
export function buildToolDefinitions(): ToolDefinition[] {
  return [
    {
      type: "function",
      function: {
        name: "run_command",
        description:
          "Run a bash shell command in the working directory. Use for anything you would do in a terminal: build, test, install, git, inspect processes. Returns stdout/stderr and the exit code.",
        parameters: {
          type: "object",
          properties: {
            command: { type: "string", description: "The bash command to run." },
          },
          required: ["command"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "list_directory",
        description: "List the entries of a directory (one level) with [dir]/[file] markers.",
        parameters: {
          type: "object",
          properties: {
            path: { type: "string", description: "Directory path. Defaults to the working directory." },
          },
        },
      },
    },
    {
      type: "function",
      function: {
        name: "read_file",
        description: "Read a text file as UTF-8 (truncated to a safe size).",
        parameters: {
          type: "object",
          properties: {
            path: { type: "string", description: "Path of the file to read." },
          },
          required: ["path"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "write_file",
        description:
          "Write content to a file, creating parent directories as needed. Use for deliverables, reports, and code you produce.",
        parameters: {
          type: "object",
          properties: {
            path: { type: "string", description: "Destination file path." },
            content: { type: "string", description: "Full file content to write." },
          },
          required: ["path", "content"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "web_search",
        description:
          "Search the web for current information (keyless DuckDuckGo). Returns ranked results with titles, URLs, and snippets.",
        parameters: {
          type: "object",
          properties: {
            query: { type: "string", description: "The search query." },
          },
          required: ["query"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "fetch_url",
        description:
          "Fetch a web page and return its readable text content (HTML stripped). Use after a web search to read a full page.",
        parameters: {
          type: "object",
          properties: {
            url: { type: "string", description: "The http(s) URL to fetch." },
          },
          required: ["url"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "github",
        description:
          "GitHub connector — call the GitHub REST API using GITHUB_TOKEN. Use for repo/issues/PRs: e.g. GET /repos/mikekoola10/koola10-nova-agent, GET /repos/{owner}/{repo}/issues, POST /repos/{owner}/{repo}/issues, GET /user. Method defaults to GET; include a JSON body for POST/PATCH.",
        parameters: {
          type: "object",
          properties: {
            path: { type: "string", description: "API path, e.g. /repos/mikekoola10/koola10-nova-agent/issues" },
            method: { type: "string", enum: ["GET", "POST", "PATCH", "PUT", "DELETE"], description: "HTTP method (default GET)" },
            body: { type: "object", description: "JSON body for POST/PATCH/PUT" },
          },
          required: ["path"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "stripe",
        description:
          "Stripe connector — call the Stripe REST API using STRIPE_SECRET_KEY. Paths like /balance, /customers, /subscriptions, /charges, /payment_intents (the /v1 prefix is added automatically). POST bodies are form-encoded. WARNING: this tool can also create charges/refunds — prefer read-only endpoints unless the task explicitly requires a payment action.",
        parameters: {
          type: "object",
          properties: {
            path: { type: "string", description: "API path, e.g. /balance or /subscriptions" },
            method: { type: "string", enum: ["GET", "POST", "PATCH", "DELETE"], description: "HTTP method (default GET)" },
            body: { type: "object", description: "Form fields for POST/PATCH" },
          },
          required: ["path"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "clawdbot",
        description:
          "Clawdbot (OpenClaw) connector — ask the local OpenClaw agent (formerly Clawdbot, https://openclaw.ai) to deliver a message to a chat contact (WhatsApp/Telegram/Discord/Slack), or dispatch a task to Clawdbot's own agent loop. Uses the `openclaw` CLI (override with CLAWDBOT_CLI).",
        parameters: {
          type: "object",
          properties: {
            action: { type: "string", enum: ["send", "agent"], description: "send = deliver a message to a contact; agent = dispatch a task to Clawdbot's agent" },
            contact: { type: "string", description: "Recipient for send, e.g. +15551234567 or a chat handle" },
            message: { type: "string", description: "The message text or task description" },
          },
          required: ["action", "message"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "browser",
        description:
          "Browser use — drive a real headless Chromium via Playwright. Use for anything that requires a real web UI: navigate (open), click elements, type into inputs, extract page text, take screenshots. Persists one tab across calls in the same run. Actions: open (url), click (selector), type (selector, text), extract, screenshot (path), back, close. Requires one-time setup: npm install playwright && npx playwright install chromium",
        parameters: {
          type: "object",
          properties: {
            action: { type: "string", enum: ["open", "click", "type", "extract", "screenshot", "back", "close"], description: "What to do in the browser" },
            url: { type: "string", description: "For open: the URL to navigate to" },
            selector: { type: "string", description: "For click/type: CSS selector, e.g. #email or button[type=submit]" },
            text: { type: "string", description: "For type: the text to enter" },
            path: { type: "string", description: "For screenshot: where to save (default output/browser-screenshot.png)" },
          },
          required: ["action"],
        },
      },
    },
    {
      type: "function",
      function: {
        name: "computer",
        description:
          "Computer use — control the local desktop (X11) via xdotool: type text, press key combos (e.g. ctrl+s, Return), move/click the mouse at (x, y) coordinates, capture the screen. Use when a task needs the desktop rather than a browser. Requires xdotool and an X display. Actions: type (text), key (key), click (x, y), move (x, y), screenshot (path).",
        parameters: {
          type: "object",
          properties: {
            action: { type: "string", enum: ["type", "key", "click", "move", "screenshot"], description: "What to do on the computer" },
            text: { type: "string", description: "For type: the text to type" },
            key: { type: "string", description: "For key: a key or combo, e.g. Return, ctrl+s, alt+Tab" },
            x: { type: "integer", description: "For click/move: x coordinate" },
            y: { type: "integer", description: "For click/move: y coordinate" },
            path: { type: "string", description: "For screenshot: where to save (default output/computer-screenshot.png)" },
          },
          required: ["action"],
        },
      },
    },
  ];
}

/** Executes a tool call and returns its output string (never throws). */
export async function dispatchTool(call: ToolCall, config: NovaConfig): Promise<string> {
  let args: Record<string, unknown>;
  try {
    args = JSON.parse(call.arguments || "{}") as Record<string, unknown>;
  } catch {
    return `ERROR: invalid JSON arguments for tool "${call.name}": ${call.arguments}`;
  }

  const str = (v: unknown, fallback = ""): string =>
    typeof v === "string" ? v : v == null ? fallback : JSON.stringify(v);
  const method = str(args.method, "GET").toUpperCase();
  const body = (args.body ?? {}) as Record<string, unknown>;
  const maxChars = config.maxToolOutputChars;

  switch (call.name) {
    case "run_command":
      return runCommand(str(args.command), {
        cwd: config.cwd,
        timeoutMs: config.toolTimeoutMs,
        maxChars,
      });
    case "list_directory":
      return listDirectory(str(args.path, "."));
    case "read_file":
      return readFile(str(args.path), maxChars);
    case "write_file":
      return writeFile(str(args.path), str(args.content));
    case "web_search":
      return webSearch(str(args.query));
    case "fetch_url":
      return fetchUrl(str(args.url), maxChars);
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
    default:
      return `ERROR: unknown tool "${call.name}".`;
  }
}
