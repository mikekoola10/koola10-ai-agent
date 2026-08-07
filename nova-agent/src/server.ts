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
 *   GET  /api/connectors/verify   -> live ping of every configured connector
 *   GET  /api/connectors/status   -> last scheduled connector self-check
 *   GET  /api/scheduled/status    -> daily report + self-check schedule status
 *   POST /api/scheduled/report    -> send the Nova daily report email now
 *   GET  /api/vault               -> vault entry names (never values)
 *   POST /api/vault               -> { name, value } store an encrypted entry
 *   DELETE /api/vault/:name       -> remove an entry
 *   GET  /                   -> the web UI
 */
import { createServer, type Server } from "node:http";
import { readFileSync } from "node:fs";
import { extname, join, normalize } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { runAgent } from "./agent.js";
import { keyEnvFor, loadConfig, loadDotEnv, type NovaConfig } from "./config.js";
import { buildToolDefinitions, verifyConnectors } from "./tools/index.js";
import { automationTool, buildDailyReport, reportDeliveryProvider } from "./tools/automations.js";
import { applyVaultOverrides, vaultDelete, vaultInfo, vaultList, vaultSet } from "./tools/vault.js";
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

const MAX_STEPS_KEPT = 300;
const MAX_TASK_CHARS = 5000;
const MAX_BODY_BYTES = 1_048_576;

/**
 * Scheduled connector self-check. Pings configured connectors (Composio,
 * Zapier, n8n, Make, Hugging Face) on a fixed cadence and caches the last
 * result so the dashboard and /api/health can surface it. Runs an initial
 * check shortly after startup. Timers are unref'd: they never keep the
 * process alive on their own. Set NOVA_CONNECTOR_CHECK_HOURS to 0 to disable
 * the schedule (manual verification via /api/connectors/verify still works).
 */
function startConnectorMonitor(config: NovaConfig) {
  let lastCheckedAt = 0;
  let ok: boolean | null = null;
  let checks: Awaited<ReturnType<typeof verifyConnectors>> | null = null;
  let error: string | null = null;

  const rawHours = Number.parseInt(process.env.NOVA_CONNECTOR_CHECK_HOURS ?? "", 10);
  const intervalHours = Number.isFinite(rawHours) && rawHours >= 0 ? rawHours : 24;

  const run = async (): Promise<void> => {
    try {
      const result = await verifyConnectors(config);
      checks = result;
      ok = result.filter((c) => c.configured).every((c) => c.ok);
      lastCheckedAt = Date.now();
      error = null;
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    }
  };

  const timers: Array<ReturnType<typeof setTimeout>> = [];
  if (intervalHours > 0) {
    const first = setTimeout(run, 2500);
    const timer = setInterval(run, intervalHours * 60 * 60 * 1000);
    timers.push(first, timer);
    if (typeof first.unref === "function") first.unref();
    if (typeof timer.unref === "function") timer.unref();
  }

  return {
    intervalHours,
    status: () => ({ lastCheckedAt, ok, checks, error, intervalHours }),
    stop: () => timers.forEach((t) => clearTimeout(t)),
  };
}

function summarizeTasks(tasks: Map<string, TaskRecord>) {
  const cutoff = Date.now() - 24 * 60 * 60 * 1000;
  let last24h = 0,
    done = 0,
    failed = 0,
    running = 0;
  for (const t of tasks.values()) {
    if (t.createdAt >= cutoff) {
      last24h += 1;
      if (t.status === "done") done += 1;
      else if (t.status === "error") failed += 1;
      else if (t.status === "running" || t.status === "queued") running += 1;
    }
  }
  return { last24h, done, failed, running, total: tasks.size };
}

/**
 * Daily Nova report job. Builds a text summary (connector health + recent
 * task stats) and POSTs it to the delivery webhook, which emails it (Zapier
 * Catch Hook by default). Fires at NOVA_DAILY_REPORT_TIME (default 09:00);
 * NOVA_DAILY_REPORT=0 disables. Also exposes runNow() for the UI/API test.
 */
function startDailyReporter(config: NovaConfig, tasks: Map<string, TaskRecord>) {
  let lastRunAt = 0;
  let lastOk: boolean | null = null;
  let lastDetail: string | null = null;

  const timeRaw = (process.env.NOVA_DAILY_REPORT_TIME || "09:00").trim();
  const parts = timeRaw.split(":").map((n) => Number.parseInt(n, 10));
  const hour = Number.isFinite(parts[0]) ? parts[0]! : 9;
  const minute = Number.isFinite(parts[1]) ? parts[1]! : 0;
  const enabled = process.env.NOVA_DAILY_REPORT !== "0";
  const provider = reportDeliveryProvider(config);
  const active = enabled && provider !== null;

  const msUntilNext = (): number => {
    const now = new Date();
    const next = new Date(now);
    next.setHours(hour, minute, 0, 0);
    if (next.getTime() <= now.getTime()) next.setDate(next.getDate() + 1);
    return next.getTime() - now.getTime();
  };

  const run = async (): Promise<void> => {
    lastRunAt = Date.now();
    try {
      const checks = await verifyConnectors(config);
      const report = buildDailyReport({ checks, version: VERSION, taskStats: summarizeTasks(tasks) });
      const out = await automationTool(config, provider!, "trigger", { payload: report.payload });
      lastOk = !String(out).startsWith("ERROR");
      lastDetail = String(out).slice(0, 200);
    } catch (err) {
      lastOk = false;
      lastDetail = err instanceof Error ? err.message : String(err);
    }
  };

  const timers: Array<ReturnType<typeof setTimeout>> = [];
  if (active) {
    const first = setTimeout(run, msUntilNext());
    const timer = setInterval(run, 24 * 60 * 60 * 1000);
    timers.push(first, timer);
    if (typeof first.unref === "function") first.unref();
    if (typeof timer.unref === "function") timer.unref();
  }

  const status = () => {
    const nextRunAt = active ? (lastRunAt > 0 ? lastRunAt + 24 * 60 * 60 * 1000 : Date.now() + msUntilNext()) : 0;
    return { active, enabled, provider, time: timeRaw, lastRunAt, lastOk, lastDetail, nextRunAt };
  };

  const runNow = async () => {
    if (!active) {
      return { ok: false, active: false, error: "no report delivery target configured (set NOVA_REPORT_PROVIDER or a webhook URL)" };
    }
    await run();
    return { ok: lastOk === true, active: true, provider, at: lastRunAt, error: lastOk === false ? lastDetail : null };
  };

  return { active, provider, time: timeRaw, enabled, status, runNow, stop: () => timers.forEach((t) => clearTimeout(t)) };
}

interface TaskStep {
  step: number;
  toolNames: string[];
  elapsedMs: number;
  tools: Array<{ name: string; preview: string; elapsedMs: number }>;
}

interface TaskRecord {
  id: string;
  task: string;
  status: string;
  provider: string;
  mock: boolean;
  createdAt: number;
  updatedAt: number;
  steps: TaskStep[];
  toolCalls: number;
  error?: string;
  report?: string;
  durationMs?: number;
}

interface NovaServer extends Server {
  connectorMonitor?: ReturnType<typeof startConnectorMonitor>;
  dailyReporter?: ReturnType<typeof startDailyReporter>;
}

/** Starts the Nova UI server on 0.0.0.0. Returns the http.Server (already listening). */
export function startServer(config: NovaConfig, port = 0): NovaServer {
  const tasks = new Map<string, TaskRecord>();

  // No API key anywhere? Run in mock mode so the UI is demoable immediately.
  const autoMock = config.mock || config.apiKey === "";

  const serveStatic = (pathname: string, res: { writeHead(code: number, headers: Record<string, string>): void; end(body: string | Buffer): void }): void => {
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

  const sendJson = (res: { writeHead(code: number, headers: Record<string, string>): void; end(body: string): void }, code: number, data: unknown): void => {
    res.writeHead(code, { "content-type": "application/json; charset=utf-8" });
    res.end(JSON.stringify(data));
  };

  const readBody = (req: { on(event: "data", cb: (chunk: unknown) => void): unknown; on(event: "end", cb: () => void): unknown; on(event: "error", cb: (err: Error) => void): unknown; destroy(): void }): Promise<Record<string, unknown>> =>
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

  const launchTask = (task: string, provider: string): TaskRecord => {
    // Re-resolve config for the requested provider so the UI's provider picker
    // really switches brains (reads the right key/model from env).
    const runConfig = provider === config.provider ? { ...config } : loadConfig({ cwd: config.cwd, provider: provider as NovaConfig["provider"] });
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

  const server: NovaServer = createServer((req, res) => {
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
            composio: config.composioApiKey !== "",
            zapier: config.zapierWebhookUrl !== "",
            n8n: config.n8nWebhookUrl !== "",
            make: config.makeWebhookUrl !== "",
            huggingface: config.huggingfaceApiKey !== "",
            browser: true, // playwright is lazy-loaded; enabled by default
            computer: true,
            lastCheckedAt: monitor.status().lastCheckedAt,
            lastOk: monitor.status().ok,
            checkIntervalHours: monitor.status().intervalHours,
          },
          dailyReport: reporter.status(),
        });
        return;
      }

      if (req.method === "GET" && pathname === "/api/connectors/verify") {
        sendJson(res, 200, { ok: true, checks: await verifyConnectors(config) });
        return;
      }

      if (req.method === "GET" && pathname === "/api/connectors/status") {
        sendJson(res, 200, monitor.status());
        return;
      }

      if (req.method === "GET" && pathname === "/api/scheduled/status") {
        sendJson(res, 200, { dailyReport: reporter.status(), connectorCheck: monitor.status() });
        return;
      }

      if (req.method === "POST" && pathname === "/api/scheduled/report") {
        try {
          await readBody(req); // drain any body; not used
        } catch {
          /* ignore malformed body */
        }
        sendJson(res, 200, await reporter.runNow());
        return;
      }

      if (req.method === "GET" && pathname === "/api/vault") {
        sendJson(res, 200, { names: vaultList(), count: vaultInfo().count, dir: vaultInfo().dir });
        return;
      }

      if (req.method === "POST" && pathname === "/api/vault") {
        let body: Record<string, unknown>;
        try {
          body = await readBody(req);
        } catch (err) {
          sendJson(res, 400, { error: (err as Error).message });
          return;
        }
        const name = typeof body.name === "string" ? body.name.trim() : "";
        const value = typeof body.value === "string" ? body.value : "";
        if (!name || !value) {
          sendJson(res, 400, { error: "name and value are required" });
          return;
        }
        const result = vaultSet(name, value);
        if (!result.ok) {
          sendJson(res, 400, { error: result.error });
          return;
        }
        sendJson(res, 200, { ok: true, name });
        return;
      }

      const vm = pathname.match(/^\/api\/vault\/([^/]+)$/);
      if (vm && req.method === "DELETE") {
        sendJson(res, 200, { ok: vaultDelete(decodeURIComponent(vm[1]!)) });
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
        const provider =
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
    })().catch((err: unknown) => {
      try {
        sendJson(res, 500, { error: (err as Error).message });
      } catch {
        // connection already closed
      }
    });
  });

  server.listen(port, "0.0.0.0");

  const monitor = startConnectorMonitor(config);
  const reporter = startDailyReporter(config, tasks);
  server.connectorMonitor = monitor;
  server.dailyReporter = reporter;
  return server;
}

function main(): void {
  loadDotEnv(process.cwd());
  const config = applyVaultOverrides(loadConfig({ cwd: process.cwd() }));

  const autoMock = config.apiKey === "";
  const cfg = autoMock ? { ...config, mock: true } : config;

  const rawPort = Number.parseInt(process.env.PORT ?? "", 10);
  const port = Number.isFinite(rawPort) && rawPort > 0 ? rawPort : 3000;

  const server = startServer(cfg, port);
  const addr = server.address();
  const shown = typeof addr === "object" && addr !== null ? addr.port : port;

  console.log(`⚡ Nova UI v${VERSION}`);
  console.log(`   brain:    ${cfg.provider}/${cfg.model}${autoMock ? "  (no API key — mock mode)" : ""}`);
  console.log(`   url:      http://0.0.0.0:${shown}`);
  console.log(`   api:      GET /api/health · GET /api/tasks · POST /api/tasks · GET /api/connectors/verify · GET /api/scheduled/status`);

  const monitorStatus = server.connectorMonitor ? server.connectorMonitor.status() : null;
  console.log(`   monitor:  connector self-check every ${monitorStatus && monitorStatus.intervalHours > 0 ? `${monitorStatus.intervalHours}h` : "disabled"} (NOVA_CONNECTOR_CHECK_HOURS)`);
  const reportStatus = server.dailyReporter ? server.dailyReporter.status() : null;
  console.log(`   report:   daily Nova report ${reportStatus && reportStatus.active ? `→ ${reportStatus.provider} at ${reportStatus.time}` : "disabled (no webhook target)"} (NOVA_DAILY_REPORT / NOVA_DAILY_REPORT_TIME)`);
  const vaultStatus = vaultInfo();
  console.log(`   vault:    ${vaultStatus.count} encrypted entr${vaultStatus.count === 1 ? "y" : "ies"} (${vaultStatus.dir}${vaultStatus.usingEnvKey ? " · NOVA_VAULT_KEY" : ""})`);
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
