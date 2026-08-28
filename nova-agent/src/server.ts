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
 *   GET  /api/health         -> runtime info (provider, model, mock, connectors, bounty sweep)
 *   GET  /api/connectors/verify   -> live ping of every configured connector
 *   GET  /api/connectors/status   -> last scheduled connector self-check
 *   GET  /api/scheduled/status    -> daily report + self-check schedule status
 *   POST /api/scheduled/report    -> send the Nova daily report email now
 *   POST /api/sweep          -> launch the standing bounty sweep now (draft-only)
 *   GET  /api/sweep/status   -> sweep schedule, scope (full vs limited) + last run
 *   GET  /api/vault               -> vault entry names (never values)
 *   POST /api/vault               -> { name, value } store an encrypted entry
 *   DELETE /api/vault/:name       -> remove an entry
 *   GET  /                   -> the web UI
 */
import { createServer, type Server } from "node:http";
import { readFileSync, readdirSync, statSync, existsSync } from "node:fs";
import { extname, join, normalize } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { runAgent } from "./agent.js";
import { keyEnvFor, loadConfig, loadDotEnv, type NovaConfig } from "./config.js";
import { buildToolDefinitions, verifyConnectors } from "./tools/index.js";
import { automationTool, buildDailyReport, reportDeliveryProvider } from "./tools/automations.js";
import { applyVaultOverrides, vaultDelete, vaultInfo, vaultList, vaultPushToRemote, vaultSet, vaultSyncFromRemote } from "./tools/vault.js";
import { firstLine, redactSecrets } from "./util.js";

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

/* ------------------------------------------------------------------ */
/* Bounty sweep — standing order + scheduler                           */
/* ------------------------------------------------------------------ */

/** The canonical bounty sweep prompt, kept next to the server in prompts/. */
function readSweepPrompt(): string {
  try {
    return readFileSync(new URL("../prompts/bounty-sweep.md", import.meta.url), "utf8");
  } catch {
    return "";
  }
}

/** Task text used for both the scheduled and manual sweep triggers. */
export const SWEEP_TASK = (): string =>
  "Read prompts/bounty-sweep.md (relative to the nova-agent project root) and execute the bounty sweep exactly as instructed. The prompt has 6 phases — follow ALL of them. CRITICAL OUTPUT: Write your JSON deck to output/bounty-report-YYYY-MM-DD.json and your markdown report to output/bounty-sweep-report.md (use today's date). The Nova UI searches output/ for these files. Reply with the path of the written review deck and a summary of findings.";

/** Current HH:MM in a fixed timezone (e.g. America/New_York). */
function tzTime(tz: string, d: Date): { hh: string; mm: string } {
  const parts = new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: tz,
  }).formatToParts(d);
  const get = (type: string): string => parts.find((p) => p.type === type)?.value ?? "";
  let hh = get("hour");
  if (hh === "24") hh = "00";
  return { hh, mm: get("minute") };
}

interface SweepStatus {
  enabled: boolean;
  times: string[];
  tz: string;
  promptAvailable: boolean;
  fullScan: boolean;
  repos: number;
  lastRunAt: number;
  lastOk: boolean | null;
  lastDetail: string | null;
  nextRunAt: number;
}

/**
 * Scheduled bounty sweep. Fires the standing sweep at the configured local
 * times (America/New_York by default) so the two-a-day automated run works
 * without cron-job.org. Nova stays READ-ONLY against GitHub: it scans, ranks
 * and drafts a human-review deck — it never posts or submits (see the prompt).
 * Configure with NOVA_SWEEP_TIMES ("HH:MM,HH:MM", default "07:40,19:40"; set
 * to "0" or empty to disable) and NOVA_SWEEP_TZ (default America/New_York).
 */
function startSweepScheduler(
  config: NovaConfig,
  launch: (task: string, provider: string, automated?: boolean) => TaskRecord,
): { status(): SweepStatus; stop(): void } {
  const raw = process.env.NOVA_SWEEP_TIMES;
  const times = (raw === undefined || raw.trim() === "" || raw.trim() === "0"
    ? "07:40,19:40"
    : raw
  )
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
  const enabled = times.length > 0;
  const tz = process.env.NOVA_SWEEP_TZ || "America/New_York";
  const prompt = readSweepPrompt();
  const fullScan = config.githubToken !== "";
  const repos = fullScan ? 28 : 10;

  let lastRunAt = 0;
  let lastOk: boolean | null = null;
  let lastDetail: string | null = null;
  let lastFiredKey = "";
  let nextRunAt = 0;

  const nextFire = (): number => {
    const now = Date.now();
    for (let i = 0; i < 60 * 26; i++) {
      const { hh, mm } = tzTime(tz, new Date(now + i * 60_000));
      if (times.includes(`${hh}:${mm}`)) return now + i * 60_000;
    }
    return 0;
  };

  const fire = (): void => {
    if (!prompt) {
      lastOk = false;
      lastDetail = "prompts/bounty-sweep.md not found on this server";
      return;
    }
    const { hh, mm } = tzTime(tz, new Date());
    lastFiredKey = `${hh}:${mm}`;
    lastRunAt = Date.now();
    const rec = launch(SWEEP_TASK(), config.provider, true);
    lastOk = true;
    lastDetail = `task ${rec.id} launched (${repos}-repo ${fullScan ? "full" : "limited"} scan)`;
  };

  const tick = (): void => {
    if (!enabled || !prompt) return;
    const { hh, mm } = tzTime(tz, new Date());
    const key = `${hh}:${mm}`;
    if (times.includes(key) && key !== lastFiredKey) fire();
    nextRunAt = nextFire();
  };

  const timers: Array<ReturnType<typeof setInterval>> = [];
  if (enabled) {
    nextRunAt = nextFire();
    const timer = setInterval(tick, 60_000);
    timers.push(timer);
    if (typeof timer.unref === "function") timer.unref();
  }

  return {
    status: (): SweepStatus => ({
      enabled,
      times,
      tz,
      promptAvailable: prompt !== "",
      fullScan,
      repos,
      lastRunAt,
      lastOk,
      lastDetail,
      nextRunAt,
    }),
    stop: () => timers.forEach((t) => clearInterval(t)),
  };
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
  /** True when the task was launched by an automated trigger (bounty sweep scheduler / cron). */
  automated?: boolean;
}

interface NovaServer extends Server {
  connectorMonitor?: ReturnType<typeof startConnectorMonitor>;
  dailyReporter?: ReturnType<typeof startDailyReporter>;
  sweepScheduler?: ReturnType<typeof startSweepScheduler>;
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
    automated: t.automated === true,
  });

  const launchTask = (task: string, provider: string, automated = false): TaskRecord => {
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
      automated,
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
          if (last) last.tools.push({ name, preview: redactSecrets(firstLine(output, 200)), elapsedMs });
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
        const sweepStatus = sweep.status();
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
          bounty: {
            fullScan: config.githubToken !== "",
            repos: config.githubToken !== "" ? 28 : 10,
            sweep: sweepStatus,
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
        sendJson(res, 200, { dailyReport: reporter.status(), connectorCheck: monitor.status(), sweep: sweep.status() });
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

      if (req.method === "GET" && pathname === "/api/sweep/status") {
        sendJson(res, 200, sweep.status());
        return;
      }
      if (req.method === "GET" && pathname === "/api/bounties") {
        // Search multiple directories for bounty markdown files
        const searchDirs = [
          join(WEB_DIR, "artifacts", "bounties"),
          join(process.cwd(), "output"),
        ];
        let files: Array<{ name: string; modifiedAt: number; size: number; dir: string }> = [];
        for (const dir of searchDirs) {
          try {
            const found = readdirSync(dir, { withFileTypes: true })
              .filter((d) => d.isFile() && d.name.endsWith(".md") && (d.name.toLowerCase().includes("bounty") || d.name.toLowerCase().includes("sweep")))
              .map((d) => {
                const st = statSync(join(dir, d.name));
                return { name: d.name, modifiedAt: st.mtimeMs, size: st.size, dir };
              });
            files.push(...found);
          } catch { /* dir doesn't exist, skip */ }
        }
        files.sort((a, b) => b.modifiedAt - a.modifiedAt);
        const latest = files[0] ?? null;
        let content = "";
        if (latest) {
          try {
            content = readFileSync(join(latest.dir, latest.name), "utf8");
          } catch {
            content = "";
          }
        }
        sendJson(res, 200, { files, latest: latest ? { ...latest, content } : null });
        return;
      }

      if (req.method === "GET" && pathname === "/api/bounties/deck") {
        let bounties: unknown[] = [];
        // Search ALL directories for bounty JSON or markdown
        const searchDirs = [
          join(WEB_DIR, "artifacts", "bounties"),
          join(process.cwd(), "output"),
          join(process.cwd(), "artifacts", "bounties"),
        ];
        for (const dir of searchDirs) {
          if (bounties.length) break;
          try {
            const files = readdirSync(dir);
            // Try JSON first
            const jsonFile = files.find((f) => f.endsWith(".json") && f.includes("bounty"));
            if (jsonFile) {
              try {
                bounties = JSON.parse(readFileSync(join(dir, jsonFile), "utf8"));
                break;
              } catch { /* parse error, skip */ }
            }
            // Try markdown fallback
            const mdFile = files.find((f) => f.endsWith(".md") && f.includes("bounty"));
            if (mdFile) {
              const md = readFileSync(join(dir, mdFile), "utf8");
              const urlRegex = /https:\/\/github\.com\/([\w.-]+)\/([\w.-]+)\/issues\/(\d+)/g;
              let match;
              while ((match = urlRegex.exec(md)) !== null) {
                const [_, owner, repo, num] = match;
                if (!(bounties as any[]).find((b) => b.url === match![0])) {
                  // Try to extract title from nearby lines
                  const beforeMd = md.slice(Math.max(0, match.index - 200), match.index);
                  const titleMatch = beforeMd.match(/#{1,3}\s+(.+)$/m) || beforeMd.match(/\*\*(.+?)\*\*/);
                  const title = titleMatch ? titleMatch[1].trim() : "";
                  // Try to extract amount
                  const afterMd = md.slice(match.index, match.index + 500);
                  const amtMatch = afterMd.match(/\$[\d,]+/);
                  const amount = amtMatch ? amtMatch[0] : "";
                  bounties.push({ repo: `${owner}/${repo}`, issueNumber: Number(num), url: match[0], title, amount, approach: "", draftComment: "" });
                }
              }
            }
          } catch { /* dir doesn't exist, skip */ }
        }
        sendJson(res, 200, { bounties });
        return;
      }

      if (req.method === "GET" && pathname === "/api/bounties/debug") {
        const debug: Record<string, unknown> = { cwd: process.cwd(), webDir: WEB_DIR };
        // Check artifact dir
        const artifactDir = join(WEB_DIR, "artifacts", "bounties");
        try {
          debug.artifactFiles = readdirSync(artifactDir);
        } catch { debug.artifactFiles = "dir not found"; }
        // Check output dir
        const outputDir = join(process.cwd(), "output");
        try {
          const files = readdirSync(outputDir);
          debug.outputFiles = files;
          // Try to read bounty files
          for (const f of files) {
            if (f.includes("bounty") && f.endsWith(".md")) {
              const content = readFileSync(join(outputDir, f), "utf8");
              debug[`output_${f}_preview`] = content.slice(0, 1500);
            }
          }
        } catch { debug.outputFiles = "dir not found"; }
        sendJson(res, 200, debug);
        return;
      }

      if (req.method === "POST" && pathname === "/api/bounties/approve") {
        const body = await readBody(req);
        const repo = String(body.repo || "");
        const issueNumber = Number(body.issueNumber);
        const comment = String(body.comment || "");
        if (!repo || !issueNumber || !comment) {
          sendJson(res, 400, { error: "Missing repo, issueNumber, or comment" });
          return;
        }
        if (!config.githubToken) {
          sendJson(res, 500, { error: "GITHUB_TOKEN not configured" });
          return;
        }
        try {
          const url = `https://api.github.com/repos/${repo}/issues/${issueNumber}/comments`;
          const ghRes = await fetch(url, {
            method: "POST",
            headers: {
              Authorization: `Bearer ${config.githubToken}`,
              Accept: "application/vnd.github+json",
              "X-GitHub-Api-Version": "2022-11-28",
              "Content-Type": "application/json",
              "User-Agent": "koola10-nova-agent",
            },
            body: JSON.stringify({ body: comment }),
            signal: AbortSignal.timeout(15_000),
          });
          const text = await ghRes.text();
          if (!ghRes.ok) {
            sendJson(res, ghRes.status, { error: `GitHub ${ghRes.status}: ${text.slice(0, 300)}` });
            return;
          }
          let ghData: unknown;
          try { ghData = JSON.parse(text); } catch { ghData = { html_url: "" }; }
          const htmlUrl = (ghData as Record<string, unknown>).html_url ?? `https://github.com/${repo}/issues/${issueNumber}#issuecomment-new`;
          sendJson(res, 200, { ok: true, url: String(htmlUrl) });
        } catch (err) {
          sendJson(res, 500, { error: `Post failed: ${(err as Error).message}` });
        }
        return;
      }

      if (req.method === "POST" && pathname === "/api/sweep") {
        try {
          await readBody(req); // drain any body; not used
        } catch {
          /* ignore malformed body */
        }
        const prompt = readSweepPrompt();
        if (!prompt) {
          sendJson(res, 500, { error: "prompts/bounty-sweep.md not found on this server" });
          return;
        }
        const rec = launchTask(SWEEP_TASK(), config.provider, true);
        const fullScan = config.githubToken !== "";
        sendJson(res, 201, {
          id: rec.id,
          status: rec.status,
          fullScan,
          repos: fullScan ? 28 : 10,
          note: "Nova drafts only — a human reviews the deck before anything is posted.",
        });
        return;
      }

      if (req.method === "GET" && pathname === "/api/vault") {
        const info = vaultInfo();
        sendJson(res, 200, {
          names: vaultList(),
          count: info.count,
          dir: info.dir,
          usingEnvKey: info.usingEnvKey,
          remote: info.remote,
        });
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
        const sync = await vaultPushToRemote();
        sendJson(res, 200, { ok: true, name, remoteSynced: sync.ok, remoteError: sync.error ?? null });
        return;
      }

      const vm = pathname.match(/^\/api\/vault\/([^/]+)$/);
      if (vm && req.method === "DELETE") {
        const ok = vaultDelete(decodeURIComponent(vm[1]!));
        const sync = ok ? await vaultPushToRemote() : { ok: true };
        sendJson(res, 200, { ok, remoteSynced: sync.ok, remoteError: sync.error ?? null });
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
  const sweep = startSweepScheduler(config, (task, provider, automated) => launchTask(task, provider, automated));
  server.connectorMonitor = monitor;
  server.dailyReporter = reporter;
  server.sweepScheduler = sweep;
  return server;
}

async function main(): Promise<void> {
  loadDotEnv(process.cwd());
  // Restore the vault from the durable remote store, but never let a slow or
  // unreachable remote backend delay (or block) the web server from starting.
  // (node-redis retries forever by default, so this race guarantees the server
  // comes up within a few seconds even if REDIS_URL is misconfigured.)
  type SyncResult = Awaited<ReturnType<typeof vaultSyncFromRemote>>;
  const restoreTimeout = new Promise<SyncResult>((resolve) =>
    setTimeout(() => resolve({ pulled: false, error: "remote restore timed out" }), 5000),
  );
  const restored = await Promise.race([vaultSyncFromRemote(), restoreTimeout]);
  if (restored.error) console.log(`   vault:    remote restore skipped — ${restored.error}`);
  const config = applyVaultOverrides(loadConfig({ cwd: process.cwd() }));

  // Deferred restore retries: a slow/cold Redis at boot can make the first
  // restore time out (or silently no-op), leaving the vault empty on a fresh
  // container — exactly what wipes UI-added keys on redeploy. Re-attempt in
  // the background a few times; the server is already listening by then, and
  // each retry builds a fresh client. Skips automatically once the vault is
  // populated, so it never clobbers user edits.
  const retryRestore = (delayMs: number): void => {
    setTimeout(() => {
      void (async () => {
        if (vaultList().length > 0) return;
        const res = await vaultSyncFromRemote();
        if (res.pulled) {
          console.log("   vault:    restored from remote store (background retry)");
          // Re-apply overrides onto the shared config so connector status /
          // health reflect the restored token without needing another restart.
          Object.assign(config, applyVaultOverrides(config));
        } else if (res.error) {
          console.log(`   vault:    background restore retry failed — ${res.error}`);
        }
      })();
    }, delayMs);
  };
  retryRestore(15_000);
  retryRestore(60_000);

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
  console.log(`   api:      GET /api/health · GET /api/tasks · POST /api/tasks · POST /api/sweep · GET /api/sweep/status · GET /api/connectors/verify · GET /api/scheduled/status`);

  const monitorStatus = server.connectorMonitor ? server.connectorMonitor.status() : null;
  console.log(`   monitor:  connector self-check every ${monitorStatus && monitorStatus.intervalHours > 0 ? `${monitorStatus.intervalHours}h` : "disabled"} (NOVA_CONNECTOR_CHECK_HOURS)`);
  const reportStatus = server.dailyReporter ? server.dailyReporter.status() : null;
  console.log(`   report:   daily Nova report ${reportStatus && reportStatus.active ? `→ ${reportStatus.provider} at ${reportStatus.time}` : "disabled (no webhook target)"} (NOVA_DAILY_REPORT / NOVA_DAILY_REPORT_TIME)`);
  const sweepStatus = server.sweepScheduler ? server.sweepScheduler.status() : null;
  console.log(`   sweep:    bounty sweep ${sweepStatus && sweepStatus.enabled ? `scheduled ${sweepStatus.times.join(", ")} ${sweepStatus.tz} · ${sweepStatus.repos}-repo ${sweepStatus.fullScan ? "FULL" : "limited"} scan` : "disabled"} (NOVA_SWEEP_TIMES / GITHUB_TOKEN)`);
  const vaultStatus = vaultInfo();
  const remoteText = vaultStatus.remote.enabled
    ? ` · remote backup ${vaultStatus.remote.host}${vaultStatus.remote.needsMasterKey ? " ⚠ set NOVA_VAULT_KEY" : ""}`
    : "";
  console.log(`   vault:    ${vaultStatus.count} encrypted entr${vaultStatus.count === 1 ? "y" : "ies"} (${vaultStatus.dir}${vaultStatus.usingEnvKey ? " · NOVA_VAULT_KEY" : ""}${remoteText})`);
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
  main().catch((err) => {
    console.error(`nova-ui: ${err instanceof Error ? err.message : String(err)}`);
    process.exit(1);
  });
}
