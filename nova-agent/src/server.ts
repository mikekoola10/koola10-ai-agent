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
 *   GET  /api/memory              -> memory summary
 *   GET  /api/memory/daily        -> today's daily note
 *   GET  /api/memory/patterns     -> bounty patterns
 *   GET  /api/memory/repos        -> repo knowledge index
 *   POST /api/memory/patterns     -> add a pattern
 *   POST /api/memory/repos        -> add repo knowledge
 *   POST /api/memory/daily        -> update daily note
 *   POST /api/memory/skills       -> add a skill
 *   GET  /api/memory/profile      -> operator profile
 *   POST /api/memory/profile      -> update operator profile
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
import { applyVaultOverrides, vaultDelete, vaultGet, vaultInfo, vaultList, vaultPushToRemote, vaultSet, vaultSyncFromRemote } from "./tools/vault.js";
import { firstLine, redactSecrets } from "./util.js";
import { getDailyNote, updateDailyNote, writeDailyNoteMarkdown, getPatterns, addPattern, getRepoKnowledge, saveRepoKnowledge, listRepoKnowledge, updateRepoIndex, getProfile, updateProfile, getSkills, addSkill, memorySummary, getRevenueSummary, getEarnings, addEarning, updateEarning, scoreBounty, rescoreBounties } from "./memory.js";

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

/**
 * PR Merge Watcher — periodically checks tracked bounties for PR status changes.
 * When a PR is merged, updates the bounty lifecycle to "paid" status.
 * Checks every 30 minutes to stay within API rate limits.
 */
function startPrWatcher(config: NovaConfig): { status(): PrWatcherStatus; stop(): void } {
  const CHECK_INTERVAL_MS = 30 * 60_000; // 30 minutes
  let lastCheckAt = 0;
  let lastOk: boolean | null = null;
  let lastDetail: string | null = null;
  let mergedCount = 0;
  let totalChecked = 0;

  const checkPrStatuses = async (): Promise<void> => {
    if (!config.githubToken) return;
    lastCheckAt = Date.now();
    try {
      const raw = vaultGet("BOUNTY_DECK");
      if (!raw) { lastDetail = "no deck in vault"; return; }
      const deck = JSON.parse(raw);
      const bounties = (deck.bounties || []) as Array<Record<string, unknown>>;
      const solving = bounties.filter((b) => b.status === "solving" && b.solveTaskId);
      if (solving.length === 0) { lastDetail = `no active solves (${bounties.length} total bounties)`; lastOk = true; return; }
      for (const bounty of solving) {
        totalChecked++;
        const repo = bounty.repo as string;
        const issueNumber = bounty.issueNumber as number;
        const prUrl = bounty.prUrl as string | undefined;
        // Check if we have a PR URL from the solve task report
        if (!prUrl) {
          // Check the solve task status
          const taskId = bounty.solveTaskId as string;
          // Tasks are in-memory, so check by looking at the bounty's solve status
          continue;
        }
        // Check PR status via GitHub API
        try {
          const prMatch = prUrl.match(/github\.com\/(.+?)\/pull\/(\d+)/);
          if (!prMatch) continue;
          const [, prRepo, prNum] = prMatch;
          const ghRes = await fetch(`https://api.github.com/repos/${prRepo}/pulls/${prNum}`, {
            headers: {
              Authorization: `Bearer ${config.githubToken}`,
              Accept: "application/vnd.github+json",
              "X-GitHub-Api-Version": "2022-11-28",
              "User-Agent": "koola10-nova-agent",
            },
            signal: AbortSignal.timeout(10_000),
          });
          if (ghRes.ok) {
            const prData = await ghRes.json() as { state: string; merged: boolean; merged_at: string | null };
            if (prData.merged) {
              bounty.status = "merged";
              bounty.mergedAt = prData.merged_at || new Date().toISOString();
              bounty.lifecycleStep = 5;
              mergedCount++;
              // Track earnings when a bounty PR is merged
              try {
                const amtStr = String(bounty.amount || "");
                const amtMatch = amtStr.match(/\$([\d,]+)/);
                const amt = amtMatch ? parseInt(amtMatch[1].replace(/,/g, ""), 10) : 0;
                if (amt > 0) {
                  addEarning({
                    date: new Date().toISOString().slice(0, 10),
                    repo: bounty.repo as string,
                    issue: Number(bounty.issueNumber),
                    amount: amt,
                    currency: "USD",
                    status: "claimed",
                    prUrl: bounty.prUrl as string,
                    mergedAt: prData.merged_at || new Date().toISOString(),
                  });
                }
              } catch { /* earnings tracking is best effort */ }
            } else if (prData.state === "closed") {
              bounty.status = "pr-closed";
              bounty.lifecycleStep = 4;
            }
          }
        } catch { /* individual PR check failed, continue */ }
      }
      vaultSet("BOUNTY_DECK", JSON.stringify({ bounties, savedAt: Date.now() }));
      lastOk = true;
      lastDetail = `checked ${solving.length} solves, ${mergedCount} merged, ${totalChecked} total checks`;
    } catch (err) {
      lastOk = false;
      lastDetail = `error: ${(err as Error).message}`;
    }
  };

  // Start the periodic check
  const timers: Array<ReturnType<typeof setInterval>> = [];
  if (config.prWatcher && config.githubToken) {
    const timer = setInterval(() => { void checkPrStatuses(); }, CHECK_INTERVAL_MS);
    timers.push(timer);
    if (typeof timer.unref === "function") timer.unref();
  }

  return {
    status: () => ({
      enabled: config.prWatcher && !!config.githubToken,
      lastCheckAt,
      lastOk,
      lastDetail,
      mergedCount,
      totalChecked,
      intervalMinutes: CHECK_INTERVAL_MS / 60_000,
    }),
    stop: () => timers.forEach((t) => clearInterval(t)),
  };
}

interface PrWatcherStatus {
  enabled: boolean;
  lastCheckAt: number;
  lastOk: boolean | null;
  lastDetail: string | null;
  mergedCount: number;
  totalChecked: number;
  intervalMinutes: number;
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
  prWatcher?: ReturnType<typeof startPrWatcher>;
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
      // Auto-save bounty deck to vault if this was a sweep task
      if (rec.automated && rec.status === "done" && /bounty/i.test(rec.task || "")) {
        try {
          const outputDir = join(process.cwd(), "output");
          const files = readdirSync(outputDir);
          // Try JSON first
          const jsonFile = files.find((f) => f.endsWith(".json") && f.includes("bounty"));
          if (jsonFile) {
            const raw = readFileSync(join(outputDir, jsonFile), "utf8");
            const parsed = JSON.parse(raw);
            if (Array.isArray(parsed) && parsed.length > 0) {
              vaultSet("BOUNTY_DECK", JSON.stringify({ bounties: parsed, savedAt: Date.now() }));
              vaultPushToRemote();
            }
          }
          // Also save markdown reports for persistence across deploys
          const mdFiles = files.filter((f) => f.endsWith(".md") && (f.includes("bounty") || f.includes("sweep")));
          for (const mdFile of mdFiles) {
            const md = readFileSync(join(outputDir, mdFile), "utf8");
            vaultSet(`BOUNTY_REPORT_${mdFile.replace(/\.md$/, "")}`, JSON.stringify({ content: md, fileName: mdFile, savedAt: Date.now() }));
            // Also save as the main report for the /api/bounties endpoint
            vaultSet('BOUNTY_REPORT', JSON.stringify({ content: md, filename: mdFile, savedAt: Date.now() }));
          }
          if (mdFiles.length > 0) vaultPushToRemote();
        } catch { /* best effort — vault save is non-critical */ }
      }
      // Auto-write daily note after any completed task
      if (rec.status === "done") {
        try {
          const isSweep = /sweep/i.test(rec.task || "");
          const isSolve = /solve|bounty/i.test(rec.task || "");
          updateDailyNote({
            sweeps: isSweep ? 1 : 0,
            solves: isSolve ? 1 : 0,
            bountiesFound: isSweep ? Math.min(rec.toolCalls, 10) : 0,
            highlights: [`[${isSweep ? 'sweep' : isSolve ? 'solve' : 'task'}] Completed in ${(result.durationMs / 1000).toFixed(0)}s (${result.steps} steps, ${result.toolCalls} tool calls)`],
          });
        } catch { /* best effort */ }
      }
      // AUTO-RETRY: If a solve task failed or ran out of steps, retry once with a simpler approach
      if (rec.status === "error" && /solve|bounty/i.test(rec.task || "") && rec.automated) {
        try {
          // Extract repo/issue from the failed task
          const repoMatch = rec.task.match(/REPO:\s*(.+)/m);
          const issueMatch = rec.task.match(/ISSUE:\s*(\d+)/m);
          const titleMatch = rec.task.match(/TITLE:\s*(.+)/m);
          if (repoMatch && issueMatch) {
            const repo = repoMatch[1].trim();
            const issue = issueMatch[1].trim();
            const title = titleMatch?.[1].trim() || "";
            // Simplified retry prompt — fewer steps, focused on just the fix
            const retryPrompt = `Solve this GitHub bounty with minimal steps. Clone the repo, read the issue, implement the smallest possible fix, and create a PR.\n\nREPO: ${repo}\nISSUE: ${issue}\nTITLE: ${title}\n\nFocus on the MINIMAL change. Don't read the entire codebase — just find the file that needs changing, make the fix, and create a PR.`;
            launchTask(retryPrompt, config.provider, true);
          }
        } catch { /* retry is best effort */ }
      }
      // AUTO-SOLVE: After a sweep completes, auto-approve and solve top bounties
      if (rec.automated && rec.status === "done" && /sweep/i.test(rec.task || "") && config.autoSolve && config.githubToken) {
        try {
          // Re-score all bounties with smarter scoring before picking
          rescoreBounties();
          const raw = vaultGet("BOUNTY_DECK");
          if (raw) {
            const deck = JSON.parse(raw);
            const bounties = (deck.bounties || []) as Array<Record<string, any>>;
            // Filter: has issue number, not already claimed/solved, score >= threshold
            const eligible = bounties
              .filter((b) => Number(b.issueNumber) > 0 && !b.status && Number(b.score) >= config.autoSolveMinScore)
              .sort((a, b) => (Number(b.score) || 0) - (Number(a.score) || 0))
              .slice(0, config.autoSolveMaxPerSweep);
            for (const bounty of eligible) {
              // Mark as claimed to avoid re-solving
              bounty.status = "auto-claimed";
              bounty.claimedAt = Date.now();
              // Post the comment on the issue
              try {
                const commentUrl = `https://api.github.com/repos/${bounty.repo}/issues/${bounty.issueNumber}/comments`;
                const ghRes = await fetch(commentUrl, {
                  method: "POST",
                  headers: {
                    Authorization: `Bearer ${config.githubToken}`,
                    Accept: "application/vnd.github+json",
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Content-Type": "application/json",
                    "User-Agent": "koola10-nova-agent",
                  },
                  body: JSON.stringify({ body: bounty.draftComment || `Hi! I'd like to work on this bounty. I've read the acceptance criteria and can deliver a quality solution.` }),
                  signal: AbortSignal.timeout(15_000),
                });
                if (ghRes.ok) {
                  bounty.commentPostedAt = Date.now();
                  bounty.status = "comment-posted";
                }
              } catch { /* comment failed, skip this bounty */ }
              // Launch a solve task
              if (bounty.status === "comment-posted") {
                const solvePrompt = `Read prompts/bounty-solve.md (relative to the nova-agent project root) and execute the bounty solver exactly as instructed.\n\nREPO: ${bounty.repo}\nISSUE: ${bounty.issueNumber}\nTITLE: ${bounty.title}\nAMOUNT: ${bounty.amount}\nAPPROACH: ${bounty.approach}\nCOMMENT: ${bounty.draftComment}\n\nClone the repo, implement the fix, run tests, and create a PR. Write results to output/bounty-solve-${String(bounty.repo).replace(/\//g, "-")}-${bounty.issueNumber}.md`;
                const solveRec = launchTask(solvePrompt, config.provider, true);
                bounty.solveTaskId = solveRec.id;
                bounty.status = "solving";
                bounty.solveStartedAt = Date.now();
                // Track as pending earnings
                try {
                  const amtStr = String(bounty.amount || "");
                  const amtMatch = amtStr.match(/\$([\d,]+)/);
                  const amt = amtMatch ? parseInt(amtMatch[1].replace(/,/g, ""), 10) : 0;
                  if (amt > 0) {
                    addEarning({
                      date: new Date().toISOString().slice(0, 10),
                      repo: String(bounty.repo),
                      issue: Number(bounty.issueNumber),
                      amount: amt,
                      currency: "USD",
                      status: "pending",
                    });
                  }
                } catch { /* earnings tracking is best effort */ }
              }
            }
            // Save updated deck back to vault
            vaultSet("BOUNTY_DECK", JSON.stringify({ bounties, savedAt: Date.now() }));
            vaultPushToRemote();
          }
        } catch { /* auto-solve is best effort */ }
      }
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
            autoSolve: {
              enabled: config.autoSolve,
              minScore: config.autoSolveMinScore,
              maxPerSweep: config.autoSolveMaxPerSweep,
            },
          },
          prWatcher: prWatcher.status(),
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
        // Fallback: if no files found, try vault for saved report
        if (!content) {
          try {
            const raw = vaultGet('BOUNTY_REPORT');
            if (raw) {
              const report = JSON.parse(raw);
              content = report.content || '';
              files = [{ name: report.filename || 'bounty-report.md', modifiedAt: report.savedAt || Date.now(), size: content.length, dir: 'vault' }];
            }
          } catch { /* vault fallback failed */ }
        }
        sendJson(res, 200, { files, latest: latest ? { ...latest, content } : { name: 'vault-report', modifiedAt: Date.now(), content } });
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
            // Try markdown fallback — extract issue URLs AND repo names
            const mdFile = files.find((f) => f.endsWith(".md") && f.includes("bounty"));
            if (mdFile) {
              const md = readFileSync(join(dir, mdFile), "utf8");
              // Extract GitHub issue URLs
              const urlRegex = /https:\/\/github\.com\/([\w.-]+)\/([\w.-]+)\/issues\/(\d+)/g;
              let match;
              while ((match = urlRegex.exec(md)) !== null) {
                const [_, owner, repo, num] = match;
                if (!(bounties as any[]).find((b) => b.url === match![0])) {
                  const beforeMd = md.slice(Math.max(0, match.index - 200), match.index);
                  const titleMatch = beforeMd.match(/#{1,3}\s+(.+)$/m) || beforeMd.match(/\*\*(.+?)\*\*/);
                  const title = titleMatch ? titleMatch[1].trim() : "";
                  const afterMd = md.slice(match.index, match.index + 500);
                  const amtMatch = afterMd.match(/\$[\d,]+/);
                  const amount = amtMatch ? amtMatch[0] : "";
                  bounties.push({ repo: `${owner}/${repo}`, issueNumber: Number(num), url: match[0], title, amount, approach: "", draftComment: "" });
                }
              }
              // Also extract from Bounty Investigation headings (# Bounty Investigation: owner/repo — "title")
              if (bounties.length === 0) {
                const seen = new Set<string>();
                const headingRegex = /^#\s+Bounty\s+Investigation:\s+([\w.-]+\/([\w.-]+))\s+[—–-]\s+["\u201c](.+?)["\u201d]/gm;
                let hm: RegExpExecArray | null;
                while ((hm = headingRegex.exec(md)) !== null) {
                  const repoKey = hm[1];
                  if (seen.has(repoKey)) continue;
                  seen.add(repoKey);
                  const title = hm[3] || repoKey;
                  // Extract amount from body
                  const amtMatch = md.match(/\$[\d,]+/);
                  const amount = amtMatch ? amtMatch[0] : "";
                  // Extract recommendation/conclusion from later sections
                  const recIdx = md.indexOf('Recommendation');
                  const conclusion = recIdx > -1 ? md.slice(recIdx, recIdx + 400).replace(/#+\s*Recommendation\s*/i, '').trim().split('\n')[0] : '';
                  // Try to find the issue URL in the body
                  const issueUrlMatch = md.match(new RegExp(`https://github\.com/${repoKey.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/issues/(\d+)`));
                  const issueNum = issueUrlMatch ? Number(issueUrlMatch[1]) : 0;
                  const url = issueUrlMatch ? issueUrlMatch[0] : `https://github.com/${repoKey}`;
                  bounties.push({ repo: repoKey, issueNumber: issueNum, url, title, amount, approach: conclusion || "", draftComment: "" });
                }
              }
              // Fallback: extract backtick-wrapped owner/repo patterns
              if (bounties.length === 0) {
                const repoRegex = /`([\w.-]+)\/([\w.-]+)`/g;
                const seen2 = new Set<string>();
                // Exclude invalid repo-like patterns and non-repo patterns
                const invalidRepos = /^(search|issues|pull|labels|wiki|api|graphql|raw|blob|tree|releases|actions|projects|security|insights|settings|packages|orgs|users|notifications|contributors)$/i;
                while ((match = repoRegex.exec(md)) !== null) {
                  const repoKey = `${match[1]}/${match[2]}`;
                  if (seen2.has(repoKey)) continue;
                  if (invalidRepos.test(match[1]) || invalidRepos.test(match[2])) continue;
                  // Validate: owner must not be empty, repo must not have file extensions
                  if (!match[1] || !match[2]) continue;
                  if (/\.json$|\.md$|\.js$|\.ts$|\.py$|\.yaml$|\.yml$/i.test(match[2])) continue;
                  // Validate: owner/repo must look like a real GitHub repo (no repeated names, no table headers)
                  if (match[1] === match[2]) continue;
                  if (/^[|\s-]+$/.test(match[1]) || /^[|\s-]+$/.test(match[2])) continue;
                  // Validate: both parts must be at least 2 chars and contain letters
                  if (match[1].length < 2 || match[2].length < 2) continue;
                  if (!/[a-zA-Z]/.test(match[1]) || !/[a-zA-Z]/.test(match[2])) continue;
                  seen2.add(repoKey);
                  const ctx = md.slice(Math.max(0, match.index - 100), match.index + 300);
                  const amtMatch = ctx.match(/\$[\d,]+/);
                  const amount = amtMatch ? amtMatch[0] : "";
                  // Extract description from nearby context
                  const lines = ctx.split('\n').filter((l: string) => l.trim());
                  const descLine = lines.find((l: string) => l.includes('bounty') || l.includes('Bounty') || l.includes('reward') || l.includes('Reward')) || '';
                  const title = descLine.replace(/^[\s-*#`]+/, '').trim().slice(0, 100) || repoKey;
                  // Generate a draft comment
                  const draftComment = `Hi! I'd like to work on this bounty. I have experience with the relevant technologies and can deliver a quality solution. Let me know if you'd like me to proceed.`;
                  bounties.push({ repo: repoKey, issueNumber: 0, url: `https://github.com/${repoKey}`, title, amount, approach: '', draftComment });
                }
              }
              // Post-process: generate draft comments for bounties with empty drafts
              for (const b of bounties as any[]) {
                if (!b.draftComment) {
                  const amt = b.amount ? ` (${b.amount})` : '';
                  b.draftComment = `Hi! I'd like to work on this${amt} bounty. I can deliver a quality solution. Let me know if you'd like me to proceed.`;
                }
              }
            }
          } catch { /* dir doesn't exist, skip */ }
        }
        // Fallback: read from vault (persists across deploys)
        if (bounties.length === 0) {
          try {
            const raw = vaultGet("BOUNTY_DECK");
            if (raw) {
              const deck = JSON.parse(raw);
              bounties = deck.bounties || [];
            }
          } catch { /* vault parse error, skip */ }
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

      if (req.method === "GET" && pathname === "/api/bounties/track") {
        // Show lifecycle status of all tracked bounties
        try {
          const raw = vaultGet("BOUNTY_DECK");
          if (!raw) { sendJson(res, 200, { bounties: [], summary: {} }); return; }
          const deck = JSON.parse(raw);
          const bounties = (deck.bounties || []) as Array<Record<string, unknown>>;
          const summary = {
            total: bounties.length,
            autoClaimed: bounties.filter((b) => b.status === "auto-claimed").length,
            commentPosted: bounties.filter((b) => b.status === "comment-posted").length,
            solving: bounties.filter((b) => b.status === "solving").length,
            solved: bounties.filter((b) => b.status === "solved").length,
            merged: bounties.filter((b) => b.status === "merged").length,
            prClosed: bounties.filter((b) => b.status === "pr-closed").length,
            unclaimed: bounties.filter((b) => !b.status).length,
          };
          sendJson(res, 200, { bounties, summary });
        } catch { sendJson(res, 500, { error: "Failed to read bounty deck" }); }
        return;
      }

      if (req.method === "POST" && pathname === "/api/bounties/save") {
        const body = await readBody(req);
        const bounties = body.bounties;
        if (!Array.isArray(bounties)) {
          sendJson(res, 400, { error: "bounties array required" });
          return;
        }
        const result = vaultSet("BOUNTY_DECK", JSON.stringify({ bounties, savedAt: Date.now() }));
        if (!result.ok) {
          sendJson(res, 500, { error: result.error });
          return;
        }
        await vaultPushToRemote();
        sendJson(res, 200, { ok: true, count: bounties.length });
        return;
      }

      if (req.method === "POST" && pathname === "/api/bounties/approve") {
        const body = await readBody(req);
        const repo = String(body.repo || "");
        let issueNumber = Number(body.issueNumber);
        const comment = String(body.comment || "");
        const title = String(body.title || "");
        const amount = String(body.amount || "");
        const approach = String(body.approach || "");
        const solve = body.solve === true; // launch full solve task?
        if (!repo || !comment) {
          sendJson(res, 400, { error: "Missing repo or comment" });
          return;
        }
        if (!config.githubToken) {
          sendJson(res, 500, { error: "GITHUB_TOKEN not configured" });
          return;
        }
        // If no issue number, look up the repo's open issues to find a match
        if (!issueNumber && title) {
          try {
            const searchUrl = `https://api.github.com/search/issues?q=repo:${repo}+type:issue+state:open+${encodeURIComponent(title.slice(0, 60))}`;
            const searchRes = await fetch(searchUrl, {
              headers: { Authorization: `Bearer ${config.githubToken}`, Accept: 'application/vnd.github+json', 'User-Agent': 'koola10-nova-agent' },
              signal: AbortSignal.timeout(10_000),
            });
            if (searchRes.ok) {
              const searchData = await searchRes.json() as any;
              const items = searchData.items || [];
              // Find best match by title similarity
              const titleLower = title.toLowerCase();
              const best = items.find((it: any) => it.title && titleLower.includes(it.title.toLowerCase().slice(0, 20))) || items[0];
              if (best) issueNumber = best.number;
            }
          } catch { /* search failed, continue without issue number */ }
        }
        // Step 1: Post the comment on the issue
        let commentUrl = issueNumber ? `https://github.com/${repo}/issues/${issueNumber}` : `https://github.com/${repo}`;
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
          commentUrl = String((ghData as Record<string, unknown>).html_url ?? commentUrl);
        } catch (err) {
          sendJson(res, 500, { error: `Post failed: ${(err as Error).message}` });
          return;
        }
        // Step 2: If solve=true, launch a full bounty-solving task
        if (solve) {
          const solvePrompt = `Read prompts/bounty-solve.md (relative to the nova-agent project root) and execute the bounty solver exactly as instructed.\n\nREPO: ${repo}\nISSUE: ${issueNumber}\nTITLE: ${title}\nAMOUNT: ${amount}\nAPPROACH: ${approach}\nCOMMENT: ${comment}\n\nClone the repo, implement the fix, run tests, and create a PR. Write results to output/bounty-solve-${repo.replace(/\//g, "-")}-${issueNumber}.md`;
          const rec = launchTask(solvePrompt, config.provider, true);
          sendJson(res, 200, { ok: true, url: commentUrl, solveTaskId: rec.id, solveStatus: "queued" });
        } else {
          sendJson(res, 200, { ok: true, url: commentUrl });
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

      // ── Revenue endpoints ────────────────────────────────────────

      if (req.method === "GET" && pathname === "/api/revenue") {
        sendJson(res, 200, getRevenueSummary());
        return;
      }

      if (req.method === "POST" && pathname === "/api/revenue") {
        const body = await readBody(req);
        addEarning({
          date: new Date().toISOString().slice(0, 10),
          repo: String(body.repo || ""),
          issue: Number(body.issue || 0),
          amount: Number(body.amount || 0),
          currency: String(body.currency || "USD"),
          status: "pending",
          prUrl: body.prUrl ? String(body.prUrl) : undefined,
        });
        sendJson(res, 200, { ok: true });
        return;
      }

      if (req.method === "PUT" && pathname === "/api/revenue") {
        const body = await readBody(req);
        updateEarning(String(body.repo || ""), Number(body.issue || 0), body);
        sendJson(res, 200, { ok: true });
        return;
      }

      if (req.method === "POST" && pathname === "/api/bounties/rescore") {
        const updated = rescoreBounties();
        sendJson(res, 200, { ok: true, updated });
        return;
      }

      // ── Memory Vault endpoints ──────────────────────────────────

      if (req.method === "GET" && pathname === "/api/memory") {
        sendJson(res, 200, {
          summary: memorySummary(),
          daily: getDailyNote(),
          patterns: getPatterns(),
          repos: listRepoKnowledge(),
          skills: getSkills(),
          profile: getProfile(),
        });
        return;
      }

      if (req.method === "GET" && pathname === "/api/memory/daily") {
        const date = url.searchParams.get("date") ?? undefined;
        const note = getDailyNote(date);
        sendJson(res, 200, { note, markdown: writeDailyNoteMarkdown(note) });
        return;
      }

      if (req.method === "POST" && pathname === "/api/memory/daily") {
        const body = await readBody(req);
        const updated = updateDailyNote(body as Record<string, unknown>);
        await vaultPushToRemote();
        sendJson(res, 200, { ok: true, note: updated });
        return;
      }

      if (req.method === "GET" && pathname === "/api/memory/patterns") {
        sendJson(res, 200, { patterns: getPatterns() });
        return;
      }

      if (req.method === "POST" && pathname === "/api/memory/patterns") {
        const body = await readBody(req);
        if (!body.pattern) {
          sendJson(res, 400, { error: "pattern is required" });
          return;
        }
        addPattern({
          pattern: String(body.pattern),
          confidence: Number(body.confidence) || 0.5,
          examples: Array.isArray(body.examples) ? body.examples : [],
          lastUpdated: Date.now(),
        });
        await vaultPushToRemote();
        sendJson(res, 200, { ok: true });
        return;
      }

      if (req.method === "GET" && pathname === "/api/memory/repos") {
        sendJson(res, 200, { repos: listRepoKnowledge() });
        return;
      }

      if (req.method === "POST" && pathname === "/api/memory/repos") {
        const body = await readBody(req);
        if (!body.repo) {
          sendJson(res, 400, { error: "repo is required" });
          return;
        }
        const knowledge = {
          repo: String(body.repo),
          verdict: (body.verdict as "pursue" | "avoid" | "monitor") || "monitor",
          reason: String(body.reason || ""),
          bountyHistory: Array.isArray(body.bountyHistory) ? body.bountyHistory : [],
          lastScanned: Date.now(),
        };
        saveRepoKnowledge(knowledge);
        updateRepoIndex(knowledge.repo);
        sendJson(res, 200, { ok: true });
        return;
      }

      if (req.method === "GET" && pathname === "/api/memory/profile") {
        sendJson(res, 200, { profile: getProfile() });
        return;
      }

      if (req.method === "POST" && pathname === "/api/memory/profile") {
        const body = await readBody(req);
        const updated = updateProfile(body as Record<string, unknown>);
        await vaultPushToRemote();
        sendJson(res, 200, { ok: true, profile: updated });
        return;
      }

      if (req.method === "POST" && pathname === "/api/memory/skills") {
        const body = await readBody(req);
        if (!body.name) {
          sendJson(res, 400, { error: "name is required" });
          return;
        }
        addSkill({
          name: String(body.name),
          description: String(body.description || ""),
          learnedFrom: String(body.learnedFrom || ""),
          lastUsed: Date.now(),
          successRate: Number(body.successRate) || 1.0,
        });
        await vaultPushToRemote();
        sendJson(res, 200, { ok: true });
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
  const prWatcher = startPrWatcher(config);
  server.connectorMonitor = monitor;
  server.dailyReporter = reporter;
  server.sweepScheduler = sweep;
  server.prWatcher = prWatcher;
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
  console.log(`   auto:     ${config.autoSolve ? `auto-solve ON (min score ${config.autoSolveMinScore}, max ${config.autoSolveMaxPerSweep}/sweep)` : "auto-solve OFF"} · ${config.prWatcher ? "PR watcher ON (30m interval)" : "PR watcher OFF"}`);
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
