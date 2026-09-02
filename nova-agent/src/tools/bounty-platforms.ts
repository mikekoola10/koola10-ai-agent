/**
 * Multi-platform bounty scanner.
 *
 * Scans multiple bounty platforms for open issues with payouts:
 *   - Algora (algora.io) — $50-$2,500, Stripe payouts
 *   - Boss.dev — GitHub-native bounties, auto-payout on PR merge
 *   - Opire — GitHub-comment-driven bounties, 4% fee
 *   - Immunefi — Security bounties, crypto payouts
 *   - GitHub Issues (existing) — direct repo scanning
 *
 * Each platform returns a normalized BountyItem[] so the sweep prompt
 * can rank and present them uniformly.
 */

export interface BountyItem {
  platform: string;
  repo: string;
  issueNumber: number;
  title: string;
  url: string;
  amount: string;
  amountRaw: number;
  currency: string;
  approach: string;
  draftComment: string;
  score: number;
  labels: string[];
  language: string;
  postedAt: string;
  platformMeta: Record<string, unknown>;
}

const FETCH_TIMEOUT = 12_000;

async function safeFetch(url: string, headers?: Record<string, string>): Promise<Response | null> {
  try {
    const res = await fetch(url, {
      headers: { "User-Agent": "koola10-nova-agent", ...headers },
      signal: AbortSignal.timeout(FETCH_TIMEOUT),
    });
    return res.ok ? res : null;
  } catch {
    return null;
  }
}

/* ── Algora ──────────────────────────────────────────────────────── */

/**
 * Fetch open bounties from Algora.
 * Public API: GET https://algora.io/api/orgs/{org}/bounties
 * Also supports: GET https://algora.io/bounties (web scrape fallback)
 */
export async function scanAlgora(): Promise<BountyItem[]> {
  const bounties: BountyItem[] = [];

  // Known Algora orgs with active bounties
  const orgs = [
    "twentyhq", "coolify", "formbricks", "hoppscotch", "infisical",
    "medusajs", "documenso", "calcom", "highlight", "unkeyed",
  ];

  for (const org of orgs) {
    const res = await safeFetch(`https://algora.io/api/orgs/${org}/bounties`);
    if (!res) continue;
    try {
      const data = await res.json() as Array<Record<string, unknown>>;
      for (const b of data) {
        if (b.status !== "open") continue;
        const amount = Number(b.reward_amount ?? b.amount ?? 0);
        if (amount <= 0) continue;
        const repo = String(b.repo ?? b.repository ?? org);
        const issueNum = Number(b.issue_number ?? b.number ?? 0);
        if (!issueNum) continue;
        bounties.push({
          platform: "algora",
          repo,
          issueNumber: issueNum,
          title: String(b.title ?? b.issue_title ?? ""),
          url: String(b.url ?? b.html_url ?? `https://github.com/${repo}/issues/${issueNum}`),
          amount: `$${amount}`,
          amountRaw: amount,
          currency: "USD",
          approach: String(b.description ?? b.body ?? "").slice(0, 500),
          draftComment: `Hi! I'd like to work on this $${amount} bounty. I've read the acceptance criteria and can deliver a quality solution. Let me know if you'd like me to proceed.`,
          score: 0, // will be scored by caller
          labels: (b.labels as string[]) ?? [],
          language: "",
          postedAt: String(b.created_at ?? ""),
          platformMeta: { org, algoraId: b.id },
        });
      }
    } catch { /* parse error, skip org */ }
  }

  return bounties;
}

/* ── Boss.dev ────────────────────────────────────────────────────── */

/**
 * Scan for Boss.dev bounties.
 * Boss.dev attaches bounty amounts to GitHub issue comments.
 * We scan known repos that use the boss-bounty GitHub app.
 */
export async function scanBossDev(githubToken?: string): Promise<BountyItem[]> {
  const bounties: BountyItem[] = [];

  // Search GitHub for issues with boss bounty comments
  // Boss comments contain "$ bounty" or "💰 bounty" patterns
  const queries = [
    '"boss bounty" is:issue is:open',
    '"💰" "bounty" is:issue is:open label:bounty',
  ];

  if (!githubToken) return bounties;

  for (const q of queries) {
    const res = await safeFetch(
      `https://api.github.com/search/issues?q=${encodeURIComponent(q)}&sort=created&order=desc&per_page=20`,
      { Authorization: `Bearer ${githubToken}`, Accept: "application/vnd.github+json" }
    );
    if (!res) continue;
    try {
      const data = await res.json() as { items?: Array<Record<string, unknown>> };
      for (const item of data.items ?? []) {
        const repoUrl = String(item.repository_url ?? "");
        const repoMatch = repoUrl.match(/repos\/(.+)/);
        if (!repoMatch) continue;
        const repo = repoMatch[1];
        const issueNum = Number(item.number ?? 0);
        const title = String(item.title ?? "");
        const body = String(item.body ?? "").slice(0, 500);

        // Extract bounty amount from title or body
        const combined = `${title} ${body}`;
        const amtMatch = combined.match(/\$(\d[\d,]*)/);
        const amount = amtMatch ? parseInt(amtMatch[1].replace(/,/g, ""), 10) : 0;
        if (amount <= 0) continue;

        bounties.push({
          platform: "boss.dev",
          repo,
          issueNumber: issueNum,
          title,
          url: String(item.html_url ?? ""),
          amount: `$${amount}`,
          amountRaw: amount,
          currency: "USD",
          approach: body,
          draftComment: `Hi! I'd like to work on this $${amount} bounty. I've read the issue and can implement a fix. Let me know if you'd like me to proceed.`,
          score: 0,
          labels: (item.labels as Array<{ name: string }>)?.map((l) => l.name) ?? [],
          language: "",
          postedAt: String(item.created_at ?? ""),
          platformMeta: { source: "github-search" },
        });
      }
    } catch { /* parse error, skip query */ }
  }

  return bounties;
}

/* ── Opire ───────────────────────────────────────────────────────── */

/**
 * Scan for Opire bounties.
 * Opire attaches bounty amounts via GitHub comments with "/bounty" commands.
 */
export async function scanOpire(githubToken?: string): Promise<BountyItem[]> {
  const bounties: BountyItem[] = [];

  if (!githubToken) return bounties;

  // Search for issues with opire bounty comments
  const queries = [
    '"/bounty" is:issue is:open',
    '"opire bounty" is:issue is:open',
  ];

  for (const q of queries) {
    const res = await safeFetch(
      `https://api.github.com/search/issues?q=${encodeURIComponent(q)}&sort=created&order=desc&per_page=20`,
      { Authorization: `Bearer ${githubToken}`, Accept: "application/vnd.github+json" }
    );
    if (!res) continue;
    try {
      const data = await res.json() as { items?: Array<Record<string, unknown>> };
      for (const item of data.items ?? []) {
        const repoUrl = String(item.repository_url ?? "");
        const repoMatch = repoUrl.match(/repos\/(.+)/);
        if (!repoMatch) continue;
        const repo = repoMatch[1];
        const issueNum = Number(item.number ?? 0);
        const title = String(item.title ?? "");
        const body = String(item.body ?? "").slice(0, 500);
        const combined = `${title} ${body}`;
        const amtMatch = combined.match(/\$(\d[\d,]*)/);
        const amount = amtMatch ? parseInt(amtMatch[1].replace(/,/g, ""), 10) : 0;
        if (amount <= 0) continue;

        bounties.push({
          platform: "opire",
          repo,
          issueNumber: issueNum,
          title,
          url: String(item.html_url ?? ""),
          amount: `$${amount}`,
          amountRaw: amount,
          currency: "USD",
          approach: body,
          draftComment: `Hi! I'd like to work on this bounty. I've read the issue and can implement a solution. Let me know if you'd like me to proceed.`,
          score: 0,
          labels: (item.labels as Array<{ name: string }>)?.map((l) => l.name) ?? [],
          language: "",
          postedAt: String(item.created_at ?? ""),
          platformMeta: { source: "opire-search" },
        });
      }
    } catch { /* parse error, skip query */ }
  }

  return bounties;
}

/* ── Immunefi (Security Bounties) ────────────────────────────────── */

/**
 * Scan Immunefi for security bounties.
 * Public API: GET https://immunefi.com/public-api/bounties.json
 * These are higher-value ($100-$100,000+) but require security expertise.
 */
export async function scanImmunefi(): Promise<BountyItem[]> {
  const bounties: BountyItem[] = [];

  const res = await safeFetch("https://immunefi.com/public-api/bounties.json");
  if (!res) return bounties;

  try {
    const data = await res.json() as Array<Record<string, unknown>>;
    for (const b of data) {
      if (b.status !== "active") continue;
      const maxPayout = Number(b.max_payout ?? b.reward ?? 0);
      if (maxPayout <= 0) continue;
      const title = String(b.title ?? b.name ?? "");
      const url = String(b.url ?? b.link ?? "");
      if (!url) continue;

      bounties.push({
        platform: "immunefi",
        repo: String(b.project ?? b.protocol ?? ""),
        issueNumber: 0, // security bounties don't have issue numbers
        title,
        url,
        amount: `$${maxPayout}+`,
        amountRaw: maxPayout,
        currency: "USD",
        approach: String(b.description ?? b.scope ?? "").slice(0, 500),
        draftComment: `Hi! I'd like to investigate this security bounty. I can perform a thorough audit and report any vulnerabilities found.`,
        score: 0,
        labels: (b.tags as string[]) ?? [],
        language: String(b.type ?? "smart-contract"),
        postedAt: String(b.created_at ?? ""),
        platformMeta: { immunefiId: b.id, payoutType: b.payout_type },
      });
    }
  } catch { /* parse error, skip */ }

  return bounties;
}

/* ── Aggregate Scanner ───────────────────────────────────────────── */

/**
 * Scan ALL platforms and return a unified, deduplicated bounty list.
 * Scores each bounty using the smart scoring function.
 */
export async function scanAllPlatforms(githubToken?: string): Promise<BountyItem[]> {
  const [algora, bossDev, opire, immunefi] = await Promise.all([
    scanAlgora().catch(() => [] as BountyItem[]),
    scanBossDev(githubToken).catch(() => [] as BountyItem[]),
    scanOpire(githubToken).catch(() => [] as BountyItem[]),
    scanImmunefi().catch(() => [] as BountyItem[]),
  ]);

  const all = [...algora, ...bossDev, ...opire, ...immunefi];

  // Deduplicate by repo+issue (if issue numbers match)
  const seen = new Set<string>();
  const deduped: BountyItem[] = [];
  for (const b of all) {
    const key = b.issueNumber > 0 ? `${b.repo}#${b.issueNumber}` : `${b.platform}:${b.url}`;
    if (seen.has(key)) continue;
    seen.add(key);
    deduped.push(b);
  }

  return deduped;
}
