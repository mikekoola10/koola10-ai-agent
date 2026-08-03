#!/usr/bin/env node
/**
 * Nova UI server — the Manus-style web interface for the Nova agent.
 *
 * Zero runtime dependencies: built on node:http. Serves the static UI from
 * web/ and exposes a small REST API that drives the real agent loop.
 *
 *   POST /api/tasks          { task, provider? } -> { id, status }
 *   GET  /api/tasks          -> summaries, newest first
 *   GET  /api/tasks/:id      -> full task (steps, report, error, timings)
 *   GET  /api/health         -> runtime info (provider, model, mock, connectors)
 *   GET  /                   -> the web UI
 */

import { createServer, type IncomingMessage, type Server, type ServerResponse } from "node:http";
import { readFileSync } from "node:fs";
import { extname, join, normalize } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { runAgent } from "./agent.js";
import { keyEnvFor, loadConfig, loadDotEnv, type NovaConfig, type Provider } from "./config.js";
import { buildToolDefinitions } from "./tools/index.js";
import { firstLine } from "./util.js";

export const VERSION = "0.4.0";

const WEB_DIR = fileURLToPath(new URL("../web/", import.meta.url));

const MIME: Record<string, string> = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".mjs": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".ico": "image/x-icon",
  ".txt": "text/plain; charset=utf-8",
  ".md": "text/markdown; charset=utf-8",
};

interface ToolLog {
  name: string;
  preview: string;
  elapsedMs: number;
}

interface StepLog {
  step: number;
  toolNames: string[];
  elapsedMs: number;
  tools: ToolLog[];
}

export interface TaskRecord {
  id: string;
  task: string;
  status: "queued" | "running" | "done" | "error";
  provider: Provider;
  mock: boolean;
  createdAt: number;
  updatedAt: number;
  steps: StepLog[];
  report?: string;
  error?: string;
  durationMs?: number;
  toolCalls: number;
}

const MAX_STEPS_KEPT = 300;
const MAX_TASK_CHARS = 5000;
const MAX_BODY_BYTES = 1_048_576;

/** Starts the Nova UI server on 0.0.0.0. Returns the http.Server (already listening). */
export function startServer(config: NovaConfig, port = 0): Server {
  const tasks = new Map<string, TaskRecord>();
  // No API key anywhere? Run in mock mode so the UI is demoable immediately.
  const autoMock = config.mock || config.apiKey === "";

  const serveStatic = (pathname: string, res: ServerResponse): void => {
    const rel = pathname === "/" || pathname === "" ? "index.html" : pathname.replace(/^\/+/, "");
    const file = normalize(join(WEB_DIR, rel));
    if (!file.startsWith(WEB_DIR)) {
      res.writeHead(403, { "content-type": "application/json" });
      res.end(JSON.stringify({ error: "forbidden" }));
      return;
    }
    try {
      const body = readFileSync(file);
      res.writeHead(200, { "content-type": MIME[extname(file)] ?? "application/octet-stream" });
      res.end(body);
    } catch {
      res.writeHead(404, { "content-type": "application/json" });
      res.end(JSON.stringify({ error: "not found" }));
    }
  };

  const sendJson = (res: ServerResponse, code: number, data: unknown): void => {
    res.writeHead(code, { "content-type": "application/json; charset=utf-8" });
    res.end(JSON.stringify(data));
  };

  const readBody = (req: IncomingMessage): Promise<Record<string, unknown>> =>
    new Promise((resolve, reject) => {
      const chunks: Buffer[] = [];
      let size = 0;
      req.on("data", (c) => {
        const chunk = Buffer.isBuffer(c) ? c : Buffer.from(String(c));
        size += chunk.length;
        if (size > MAX_BODY_BYTES) {
          reject(new Error("payload too large"));
          req.destroy();
          return;
        }
        chunks.push(chunk);
      });
      req.on("end", () => {
        const text = Buffer.concat(chunks).toString("utf8").trim();
        if (!text) {
          resolve({});
          return;
        }
        try {
          resolve(JSON.parse(text) as Record<string, unknown>);
        } catch {
          reject(new Error("invalid JSON body"));
        }
      });
      req.on("error", reject);
    });

  const summary = (t: TaskRecord) => ({
    id: t.id,
    task: t.task,
    status: t.status,
    provider: t.provider,
    mock: t.mock,
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
    steps: t.steps.length,
    toolCalls: t.toolCalls,
  });

  const launchTask = (task: string, provider: Provider): TaskRecord => {
    // Re-resolve config for the requested provider so the UI's provider picker
    // really switches brains (reads the right key/model from env).
    const runConfig: NovaConfig =
      provider === config.provider ? { ...config } : loadConfig({ cwd: config.cwd, provider });
    if (!runConfig.apiKey && !runConfig.mock) {
      runConfig.mock = autoMock; // fresh install demo: no key anywhere -> scripted brain
    }
    const rec: TaskRecord = {
      id: `t-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`,
      task,
      status: "queued",
      provider: runConfig.provider,
      mock: runConfig.mock,
      createdAt: Date.now(),
      updatedAt: Date.now(),
      steps: [],
      toolCalls: 0,
    };
    tasks.set(rec.id, rec);
    void (async () => {
      rec.status = "running";
      rec.updatedAt = Date.now();
      let failed = false;
      const result = await runAgent(task, runConfig, {
        onStep: ({ step, toolNames, elapsedMs }) => {
          if (rec.steps.length >= MAX_STEPS_KEPT) rec.steps.shift();
          rec.steps.push({ step, toolNames, elapsedMs, tools: [] });
          rec.updatedAt = Date.now();
        },
        onTool: ({ name, output, elapsedMs }) => {
          const last = rec.steps[rec.steps.length - 1];
          if (last) last.tools.push({ name, preview: firstLine(output, 200), elapsedMs });
          rec.updatedAt = Date.now();
        },
        onError: (err) => {
          failed = true;
          rec.status = "error";
          rec.error = err.message;
          rec.updatedAt = Date.now();
        },
      });
      rec.report = result.report;
      rec.durationMs = result.durationMs;
      rec.toolCalls = result.toolCalls;
      if (!failed) rec.status = "done";
      rec.updatedAt = Date.now();
    })();
    return rec;
  };

  const server = createServer((req, res) => {
    void (async () => {
      const url = new URL(req.url ?? "/", "http://localhost");
      const pathname = url.pathname;

      if (req.method === "GET" && pathname === "/api/health") {
        const tools = buildToolDefinitions();
        sendJson(res, 200, {
          ok: true,
          version: VERSION,
          provider: config.provider,
          model: config.model,
          mock: autoMock,
          tools: tools.length,
          maxSteps: config.maxSteps,
          cwd: config.cwd,
          connectors: {
            github: config.githubToken !== "",
            stripe: config.stripeKey !== "",
            clawdbot: config.clawdbotCli !== "",
            browser: true, // playwright is lazy-loaded; enabled by default
            computer: true,
          },
        });
        return;
      }

      if (req.method === "GET" && pathname === "/api/tasks") {
        const list = [...tasks.values()].sort((a, b) => b.createdAt - a.createdAt).map(summary);
        sendJson(res, 200, list);
        return;
      }

      const m = pathname.match(/^\/api\/tasks\/([^/]+)$/);
      if (m && req.method === "GET") {
        const t = tasks.get(m[1]!);
        if (!t) {
          sendJson(res, 404, { error: "task not found" });
          return;
        }
        sendJson(res, 200, t);
        return;
      }

      if (req.method === "POST" && pathname === "/api/tasks") {
        let body: Record<string, unknown>;
        try {
          body = await readBody(req);
        } catch (err) {
          sendJson(res, 400, { error: (err as Error).message });
          return;
        }
        const task = typeof body.task === "string" ? body.task.trim() : "";
        if (!task) {
          sendJson(res, 400, { error: "task is required" });
          return;
        }
        if (task.length > MAX_TASK_CHARS) {
          sendJson(res, 400, { error: `task too long (max ${MAX_TASK_CHARS} chars)` });
          return;
        }
        const provider: Provider =
          body.provider === "anthropic" || body.provider === "openai" || body.provider === "deepseek"
            ? body.provider
            : config.provider;
        const probe = provider === config.provider ? config : loadConfig({ cwd: config.cwd, provider });
        if (!probe.apiKey && !autoMock && !probe.mock) {
          sendJson(res, 400, {
            error: `${keyEnvFor(provider)} is not set for provider "${provider}". Add it to a .env file or choose another provider.`,
          });
          return;
        }
        const rec = launchTask(task, provider);
        sendJson(res, 201, { id: rec.id, status: rec.status });
        return;
      }

      if (req.method === "GET") {
        serveStatic(pathname, res);
        return;
      }
      sendJson(res, 405, { error: "method not allowed" });
    })().catch((err) => {
      try {
        sendJson(res, 500, { error: (err as Error).message });
      } catch {
        // connection already closed
      }
    });
  });

  server.listen(port, "0.0.0.0");
  return server;
}

function main(): void {
  loadDotEnv(process.cwd());
  const config = loadConfig({ cwd: process.cwd() });
  const autoMock = config.apiKey === "";
  const cfg: NovaConfig = autoMock ? { ...config, mock: true } : config;
  const rawPort = Number.parseInt(process.env.PORT ?? "", 10);
  const port = Number.isFinite(rawPort) && rawPort > 0 ? rawPort : 3000;
  const server = startServer(cfg, port);
  const addr = server.address();
  const shown = typeof addr === "object" && addr !== null ? addr.port : port;
  console.log(`⚡ Nova UI v${VERSION}`);
  console.log(`   brain:    ${cfg.provider}/${cfg.model}${autoMock ? "  (no API key — mock mode)" : ""}`);
  console.log(`   url:      http://0.0.0.0:${shown}`);
  console.log(`   api:      GET /api/health · GET /api/tasks · POST /api/tasks`);
  console.log(`   ctrl-c to stop`);
  server.on("error", (err) => {
    console.error(`nova-ui: ${err.message}`);
    process.exit(1);
  });
  for (const sig of ["SIGINT", "SIGTERM"] as const) {
    process.on(sig, () => {
      server.close(() => process.exit(0));
      setTimeout(() => process.exit(0), 800).unref();
    });
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main();
}
