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
