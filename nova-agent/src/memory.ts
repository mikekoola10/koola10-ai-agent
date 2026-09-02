/**
 * Nova Memory Vault — structured persistent memory system.
 *
 * Inspired by jaredrhod/ai-memory-vault. Stores structured data in the
 * existing vault (encrypted key-value store backed by Redis) using prefixed
 * keys so everything survives redeploys.
 *
 * Vault key layout:
 *   MEM:DAILY:YYYY-MM-DD   — daily activity note (markdown)
 *   MEM:PATTERNS            — bounty patterns / learned intelligence
 *   MEM:REPOS:owner/repo    — per-repo knowledge (bounty history, verdicts)
 *   MEM:PROFILE             — operator profile (preferences, strategies)
 *   MEM:SKILLS              — learned skills / methods
 *   MEM:CONTEXT:taskType    — priming context for a task type
 */
import { vaultGet, vaultSet, vaultPushToRemote } from "./tools/vault.js";

/* ------------------------------------------------------------------ */
/* Types                                                               */
/* ------------------------------------------------------------------ */

export interface DailyNote {
  date: string; // YYYY-MM-DD
  sweeps: number;
  solves: number;
  bountiesFound: number;
  bountiesClaimed: number;
  earnings: number;
  highlights: string[];
  lessons: string[];
}

export interface BountyPattern {
  pattern: string; // e.g. "Algora-verified bounties pay best"
  confidence: number; // 0-1
  examples: string[];
  lastUpdated: number;
}

export interface RepoKnowledge {
  repo: string;
  verdict: "pursue" | "avoid" | "monitor";
  reason: string;
  bountyHistory: Array<{ issue: number; title: string; amount: string; outcome: string }>;
  lastScanned: number;
}

export interface OperatorProfile {
  name: string;
  preferences: Record<string, string>;
  strategies: string[];
  riskTolerance: "conservative" | "moderate" | "aggressive";
  preferredLanguages: string[];
  maxBountyAge: number; // days
}

export interface Skill {
  name: string;
  description: string;
  learnedFrom: string;
  lastUsed: number;
  successRate: number;
}

/* ------------------------------------------------------------------ */
/* Vault helpers                                                       */
/* ------------------------------------------------------------------ */

function vaultGetJSON<T>(key: string): T | null {
  try {
    const raw = vaultGet(key);
    if (!raw) return null;
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

function vaultSetJSON(key: string, data: unknown): boolean {
  try {
    const result = vaultSet(key, JSON.stringify(data));
    return result.ok;
  } catch {
    return false;
  }
}

/* ------------------------------------------------------------------ */
/* Daily notes                                                         */
/* ------------------------------------------------------------------ */

function todayKey(): string {
  return new Date().toISOString().slice(0, 10);
}

export function getDailyNote(date?: string): DailyNote {
  const key = `MEM:DAILY:${date ?? todayKey()}`;
  return vaultGetJSON<DailyNote>(key) ?? {
    date: date ?? todayKey(),
    sweeps: 0,
    solves: 0,
    bountiesFound: 0,
    bountiesClaimed: 0,
    earnings: 0,
    highlights: [],
    lessons: [],
  };
}

export function updateDailyNote(update: Partial<DailyNote>): DailyNote {
  const current = getDailyNote();
  const merged: DailyNote = {
    ...current,
    ...update,
    sweeps: (current.sweeps ?? 0) + (update.sweeps ?? 0),
    solves: (current.solves ?? 0) + (update.solves ?? 0),
    bountiesFound: (current.bountiesFound ?? 0) + (update.bountiesFound ?? 0),
    bountiesClaimed: (current.bountiesClaimed ?? 0) + (update.bountiesClaimed ?? 0),
    earnings: (current.earnings ?? 0) + (update.earnings ?? 0),
    highlights: [...(current.highlights ?? []), ...(update.highlights ?? [])],
    lessons: [...(current.lessons ?? []), ...(update.lessons ?? [])],
  };
  vaultSetJSON(`MEM:DAILY:${merged.date}`, merged);
  return merged;
}

export function writeDailyNoteMarkdown(note: DailyNote): string {
  const lines = [
    `# Daily Note — ${note.date}`,
    "",
    `## Activity`,
    `- Sweeps: ${note.sweeps}`,
    `- Solves: ${note.solves}`,
    `- Bounties found: ${note.bountiesFound}`,
    `- Bounties claimed: ${note.bountiesClaimed}`,
    `- Earnings: $${note.earnings}`,
    "",
  ];
  if (note.highlights.length) {
    lines.push("## Highlights", ...note.highlights.map((h) => `- ${h}`), "");
  }
  if (note.lessons.length) {
    lines.push("## Lessons Learned", ...note.lessons.map((l) => `- ${l}`), "");
  }
  return lines.join("\n");
}

/* ------------------------------------------------------------------ */
/* Bounty patterns                                                     */
/* ------------------------------------------------------------------ */

export function getPatterns(): BountyPattern[] {
  return vaultGetJSON<BountyPattern[]>("MEM:PATTERNS") ?? [];
}

export function addPattern(pattern: BountyPattern): void {
  const existing = getPatterns();
  // Dedupe by pattern text
  const filtered = existing.filter((p) => p.pattern !== pattern.pattern);
  filtered.push({ ...pattern, lastUpdated: Date.now() });
  vaultSetJSON("MEM:PATTERNS", filtered);
}

/* ------------------------------------------------------------------ */
/* Repo knowledge                                                      */
/* ------------------------------------------------------------------ */

export function getRepoKnowledge(repo: string): RepoKnowledge | null {
  return vaultGetJSON<RepoKnowledge>(`MEM:REPOS:${repo}`);
}

export function saveRepoKnowledge(knowledge: RepoKnowledge): void {
  vaultSetJSON(`MEM:REPOS:${knowledge.repo}`, { ...knowledge, lastScanned: Date.now() });
  vaultPushToRemote().catch(() => {});
}

export function listRepoKnowledge(): RepoKnowledge[] {
  // We can't list vault keys by prefix easily, so we store an index
  return vaultGetJSON<RepoKnowledge[]>("MEM:REPOS:INDEX") ?? [];
}

export function updateRepoIndex(repo: string): void {
  const index = vaultGetJSON<string[]>("MEM:REPOS:INDEX_LIST") ?? [];
  if (!index.includes(repo)) {
    index.push(repo);
    vaultSetJSON("MEM:REPOS:INDEX_LIST", index);
  }
}

/* ------------------------------------------------------------------ */
/* Operator profile                                                    */
/* ------------------------------------------------------------------ */

export function getProfile(): OperatorProfile {
  return vaultGetJSON<OperatorProfile>("MEM:PROFILE") ?? {
    name: "Operator",
    preferences: {},
    strategies: [],
    riskTolerance: "moderate",
    preferredLanguages: ["TypeScript", "Python", "JavaScript", "Go", "Rust"],
    maxBountyAge: 30,
  };
}

export function updateProfile(update: Partial<OperatorProfile>): OperatorProfile {
  const current = getProfile();
  const merged = { ...current, ...update };
  vaultSetJSON("MEM:PROFILE", merged);
  return merged;
}

/* ------------------------------------------------------------------ */
/* Skills                                                              */
/* ------------------------------------------------------------------ */

export function getSkills(): Skill[] {
  return vaultGetJSON<Skill[]>("MEM:SKILLS") ?? [];
}

export function addSkill(skill: Skill): void {
  const existing = getSkills();
  const filtered = existing.filter((s) => s.name !== skill.name);
  filtered.push({ ...skill, lastUsed: Date.now() });
  vaultSetJSON("MEM:SKILLS", filtered);
}

/* ------------------------------------------------------------------ */
/* AI Priming — load context before a task                             */
/* ------------------------------------------------------------------ */

/**
 * Generates a context string to prime the agent before a task.
 * This is the core of the ai-memory-vault concept: load relevant notes
 * BEFORE the agent starts working so it has full context.
 */
export function primeContext(taskType: string, extra?: Record<string, string>): string {
  const parts: string[] = [];

  // 1. Operator profile
  const profile = getProfile();
  parts.push(`## Operator Profile\n- Name: ${profile.name}\n- Risk tolerance: ${profile.riskTolerance}\n- Preferred languages: ${profile.preferredLanguages.join(", ")}\n- Max bounty age: ${profile.maxBountyAge} days`);

  // 2. Recent daily notes (last 7 days)
  const recentNotes: string[] = [];
  for (let i = 0; i < 7; i++) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    const note = getDailyNote(d.toISOString().slice(0, 10));
    if (note.sweeps > 0 || note.solves > 0) {
      recentNotes.push(`${note.date}: ${note.sweeps} sweeps, ${note.solves} solves, ${note.bountiesFound} found, $${note.earnings} earned`);
    }
  }
  if (recentNotes.length) {
    parts.push(`## Recent Activity\n${recentNotes.join("\n")}`);
  }

  // 3. Bounty patterns
  const patterns = getPatterns();
  if (patterns.length) {
    parts.push(`## Learned Patterns\n${patterns.sort((a, b) => b.confidence - a.confidence).slice(0, 10).map((p) => `- ${p.pattern} (confidence: ${(p.confidence * 100).toFixed(0)}%)`).join("\n")}`);
  }

  // 4. Repo knowledge (for sweep/solve tasks)
  if (taskType === "sweep" || taskType === "solve") {
    const repos = listRepoKnowledge();
    if (repos.length) {
      const avoid = repos.filter((r) => r.verdict === "avoid");
      const pursue = repos.filter((r) => r.verdict === "pursue");
      if (avoid.length) {
        parts.push(`## Repos to AVOID\n${avoid.map((r) => `- ${r.repo}: ${r.reason}`).join("\n")}`);
      }
      if (pursue.length) {
        parts.push(`## Repos to PRIORITIZE\n${pursue.map((r) => `- ${r.repo}: ${r.reason}`).join("\n")}`);
      }
    }
  }

  // 5. Extra context
  if (extra) {
    const lines = Object.entries(extra).map(([k, v]) => `- ${k}: ${v}`);
    parts.push(`## Additional Context\n${lines.join("\n")}`);
  }

  return parts.join("\n\n");
}

/* ------------------------------------------------------------------ */
/* Revenue Tracking                                                    */
/* ------------------------------------------------------------------ */

export interface EarningsRecord {
  date: string;
  repo: string;
  issue: number;
  amount: number;
  currency: string;
  status: "pending" | "claimed" | "paid";
  /** Which platform the bounty came from */
  platform: string;
  prUrl?: string;
  mergedAt?: string;
  paidAt?: string;
}

export interface RevenueSummary {
  totalEarned: number;
  totalPending: number;
  totalPaid: number;
  bountyCount: number;
  paidCount: number;
  avgBounty: number;
  bestBounty: { repo: string; issue: number; amount: number } | null;
  recentEarnings: EarningsRecord[];
  monthlyEarnings: Array<{ month: string; total: number }>;
  byPlatform: Record<string, { count: number; total: number }>;
}

export function getEarnings(): EarningsRecord[] {
  return vaultGetJSON<EarningsRecord[]>("MEM:EARNINGS") ?? [];
}

export function addEarning(record: EarningsRecord): void {
  const earnings = getEarnings();
  // Deduplicate by repo+issue
  const filtered = earnings.filter((e) => !(e.repo === record.repo && e.issue === record.issue));
  filtered.push(record);
  vaultSetJSON("MEM:EARNINGS", filtered);
}

export function updateEarning(repo: string, issue: number, update: Partial<EarningsRecord>): void {
  const earnings = getEarnings();
  const idx = earnings.findIndex((e) => e.repo === repo && e.issue === issue);
  if (idx >= 0) {
    earnings[idx] = { ...earnings[idx], ...update };
    vaultSetJSON("MEM:EARNINGS", earnings);
  }
}

export function getRevenueSummary(): RevenueSummary {
  const earnings = getEarnings();
  const totalPaid = earnings.filter((e) => e.status === "paid").reduce((s, e) => s + e.amount, 0);
  const totalPending = earnings.filter((e) => e.status !== "paid").reduce((s, e) => s + e.amount, 0);
  const paidCount = earnings.filter((e) => e.status === "paid").length;
  const bestBounty = earnings.length > 0
    ? earnings.reduce((best, e) => e.amount > (best?.amount ?? 0) ? e : best, earnings[0])
    : null;
  // Monthly breakdown
  const monthly: Record<string, number> = {};
  for (const e of earnings.filter((e) => e.status === "paid")) {
    const month = e.paidAt?.slice(0, 7) ?? e.date.slice(0, 7);
    monthly[month] = (monthly[month] ?? 0) + e.amount;
  }
  // Platform breakdown
  const byPlatform: Record<string, { count: number; total: number }> = {};
  for (const e of earnings) {
    const p = e.platform || "github";
    if (!byPlatform[p]) byPlatform[p] = { count: 0, total: 0 };
    byPlatform[p].count++;
    byPlatform[p].total += e.amount;
  }
  return {
    totalEarned: totalPaid,
    totalPending,
    totalPaid,
    bountyCount: earnings.length,
    paidCount,
    avgBounty: paidCount > 0 ? totalPaid / paidCount : 0,
    bestBounty: bestBounty ? { repo: bestBounty.repo, issue: bestBounty.issue, amount: bestBounty.amount } : null,
    recentEarnings: earnings.slice(-10).reverse(),
    monthlyEarnings: Object.entries(monthly).sort(([a], [b]) => b.localeCompare(a)).map(([month, total]) => ({ month, total })),
    byPlatform,
  };
}

/* ------------------------------------------------------------------ */
/* Smarter Bounty Scoring                                              */
/* ------------------------------------------------------------------ */

/** Language difficulty weights — lower = easier for Nova to solve */
const LANGUAGE_WEIGHTS: Record<string, number> = {
  python: 1.0,
  javascript: 1.0,
  typescript: 1.1,
  go: 1.3,
  rust: 1.6,
  java: 1.4,
  cpp: 1.8,
  c: 1.8,
  haskell: 2.0,
  unknown: 1.2,
};

/** Extract likely language from repo or title hints */
function detectLanguage(repo: string, title: string, approach: string): string {
  const combined = `${repo} ${title} ${approach}`.toLowerCase();
  if (combined.includes("python") || combined.includes("pip") || combined.includes(".py")) return "python";
  if (combined.includes("javascript") || combined.includes("npm") || combined.includes("node")) return "javascript";
  if (combined.includes("typescript") || combined.includes("tsc")) return "typescript";
  if (combined.includes("golang") || combined.includes("go ") || combined.includes("go-")) return "go";
  if (combined.includes("rust") || combined.includes("cargo")) return "rust";
  if (combined.includes("java ") || combined.includes("maven")) return "java";
  return "unknown";
}

/** Parse bounty amount to a number */
function parseAmount(amount: string): number {
  const match = amount.match(/\$([\d,]+)/);
  if (!match) return 0;
  return parseInt(match[1].replace(/,/g, ""), 10) || 0;
}

/**
 * Score a bounty for auto-solve priority.
 * Higher score = better candidate for Nova.
 * 
 * Factors:
 * - Amount (higher = better)
 * - Language difficulty (easier = better)
 * - Has issue number (required)
 * - Title clarity (more specific = better)
 * - Has approach description (better defined = better)
 */
export function scoreBounty(bounty: {
  repo: string; issueNumber: number; title: string;
  amount: string; approach: string; url?: string;
}): number {
  let score = 0;

  // 1. Amount score (0-4 points)
  const amt = parseAmount(bounty.amount);
  if (amt >= 1000) score += 4;
  else if (amt >= 500) score += 3.5;
  else if (amt >= 100) score += 3;
  else if (amt >= 50) score += 2.5;
  else if (amt >= 20) score += 2;
  else if (amt > 0) score += 1;
  else score += 0.5; // unknown amount, still worth trying

  // 2. Language difficulty (0-2 points, easier = higher)
  const lang = detectLanguage(bounty.repo, bounty.title, bounty.approach);
  const weight = LANGUAGE_WEIGHTS[lang] ?? 1.2;
  score += Math.max(0, 2 - (weight - 1) * 2);

  // 3. Title clarity (0-1.5 points)
  const titleLen = bounty.title.length;
  if (titleLen > 30 && titleLen < 120) score += 1.5;
  else if (titleLen > 15) score += 1;
  else score += 0.5;

  // 4. Has approach (0-1 point)
  if (bounty.approach && bounty.approach.length > 20) score += 1;
  else if (bounty.approach) score += 0.5;

  // 5. Has issue number (required for auto-solve, -2 if missing)
  if (!bounty.issueNumber) score -= 2;

  return Math.round(score * 100) / 100;
}

/** Re-score all bounties in the deck and update vault */
export function rescoreBounties(): number {
  const raw = vaultGet("BOUNTY_DECK");
  if (!raw) return 0;
  const deck = JSON.parse(raw);
  const bounties = deck.bounties || [];
  let updated = 0;
  for (const b of bounties) {
    const newScore = scoreBounty(b);
    if (b.score !== newScore) {
      b.score = newScore;
      updated++;
    }
  }
  if (updated > 0) {
    vaultSet("BOUNTY_DECK", JSON.stringify({ bounties, savedAt: Date.now() }));
  }
  return updated;
}

/* ------------------------------------------------------------------ */
/* Memory summary for the agent system prompt                          */
/* ------------------------------------------------------------------ */

export function memorySummary(): string {
  const profile = getProfile();
  const patterns = getPatterns();
  const skills = getSkills();
  const note = getDailyNote();

  return [
    `# Nova Memory`,
    `Operator: ${profile.name}`,
    `Today: ${note.sweeps} sweeps, ${note.solves} solves, $${note.earnings} earned`,
    `Patterns learned: ${patterns.length}`,
    `Skills: ${skills.length}`,
    `Risk tolerance: ${profile.riskTolerance}`,
  ].join("\n");
}
