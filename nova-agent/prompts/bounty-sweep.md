# Nova Bounty Sweep — standing order (v2.0)

You are executing the standing bounty sweep for Koola10. **This task is READ-ONLY
against GitHub**: you may scan, read, analyze, and draft — you must NEVER post
comments, open issues/PRs, or take any visible action. Everything you produce is
for a human to review and approve first.

## Mission
Find real, open, winnable GitHub bounties worth **$50+**, rank them by
(value × likelihood of winning), and deliver a human-review deck with
ready-to-post first comments.

---

## Phase 1: Global bounty search (HIGHEST PRIORITY)

Use the `github` tool (or `curl` against `https://api.github.com/search/issues`)
to search ALL of GitHub for open bounties. Run these searches IN ORDER:

### Search 1: Algora-verified bounties (real money, paid on merge)
```
label:"💎 Bounty" state:open type:issue sort:created
```
These are the highest-quality bounties — verified by Algora.io, paid via Stripe
on merge. Take ALL results from this search.

### Search 2: Dollar-amount bounties
```
label:bounty "$" state:open type:issue sort:updated
```
Find bounties with explicit dollar amounts. Prioritize $100+.

### Search 3: General bounty label
```
label:bounty state:open type:issue sort:created
```
Cast a wider net for repos that use the standard `bounty` label.

### Search 4: Bounty platforms
```
"algora.io" OR "opire.dev" OR "gitcoin" state:open type:issue sort:updated
```
Catch bounties posted on bounty platforms.

### Search 5: Dollar amounts in body
```
"$" "bounty" state:open type:issue sort:updated
```
Find bounties that mention dollar amounts in the issue body.

**Rate limiting:** If you hit 403/429, pause 30 seconds and continue. You have
`GITHUB_TOKEN` with 5,000 req/hr — use it for all API calls.

---

## Phase 2: Known bounty repositories (supplement global search)

These repos are PROVEN to have active bounty programs. Scan each one for
open issues labeled `bounty`, `reward`, `💰`, or `💎 Bounty`:

### Tier 1 — Active bounty programs (scan ALL open issues)
- `calcom/cal.com` — TypeScript/React, $50-$500 bounties
- `coollabsio/coolify` — PHP/Laravel, $7-$100 bounties
- `activepieces/activepieces` — TypeScript, $50-$100 bounties
- `tenstorrent/tt-metal` — C++/Metal, $3,500-$10,000 bounties
- `zio/zio` — Scala, $150-$1,000 bounties
- `rohitdash08/FinMind` — TypeScript, $200-$1,000 bounties
- `archestra-ai/archestra` — TypeScript, $100-$500 bounties
- `FreezingMoon/AncientBeast` — JS/CSS, 8-10 XTR bounties
- `hashgraph/guardian` — $100K annual bounty program, $3K-$5K per bounty
- `ClickHouse/ClickHouse` — Bug bounty program
- `systeminit/si` — Bug bounty program

### Tier 2 — Bounty-friendly repos (check for bounty labels)
- `vercel/next.js` — occasional bounties
- `supabase/supabase` — occasional bounties
- `oven-sh/bun` — occasional bounties
- `microsoft/vscode` — bug bounties
- `huggingface/transformers` — community bounties
- `langchain-ai/langchain` — community bounties
- `run-llama/llama_index` — community bounties
- `microsoft/autogen` — community bounties
- `crewAIInc/crewAI` — community bounties
- `Significant-Gravitas/AutoGPT` — community bounties
- `lobehub/lobe-chat` — IssueHunt bounties
- `danny-avila/LibreChat` — community bounties
- `BerriAI/litellm` — community bounties
- `Aider-AI/aider` — community bounties
- `All-Hands-AI/OpenHands` — community bounties
- `ComposioHQ/composio` — community bounties
- `phidatahq/phidata` — community bounties

### Tier 3 — Crypto/Web3 bounty programs
- `Scottcjn/rustchain-bounties` — active bounty board
- `TheSCInitiative/bounties` — bounty board
- `projectdiscovery/oss-bounty-program` — OSS bounties
- `zama-ai/bounty-program` — FHE bounties

---

## Phase 3: Filtering (EXCLUDE aggressively)

**EXCLUDE these patterns** (they are spam, not real bounties):
- Issues by `Nexussyn` or `tdpeta754` (serial spammers)
- "AI Growth Engine" or "AiMPN" promotional posts
- Third-party "collaboration" offers from strangers
- Auto-generated dependency update PRs
- Issues asking "is there a bounty program?" (questions, not bounties)
- Issues with no dollar amount AND no bounty label
- Security-only programs that require private disclosure (HackerOne, etc.)

**INCLUDE only issues that:**
- Have a `bounty`, `reward`, `💰`, or `💎 Bounty` label, OR
- Contain an explicit dollar amount ($XX) in title or body, OR
- Are from a known bounty platform (Algora, Opire, Gitcoin)

---

## Phase 4: Scoring and ranking

For each qualifying bounty, score on 4 dimensions (each 1-10):

1. **Value** (amount): $1000+ = 10, $500+ = 8, $100+ = 6, $50+ = 4, <$50 = 2
2. **Winnability**: Is the problem well-defined? Are there few competing PRs?
   Is it in a language we know? Low competition + clear spec = high score.
3. **Freshness**: Created in last 7 days = 10, 30 days = 7, 90 days = 4, older = 2
4. **Confidence**: How confident are we we can actually solve it?
   TypeScript/Python/JS = 9, Go/Rust = 7, C++ = 5, Scala/Haskell = 3

**Final score** = (Value × 0.3) + (Winnability × 0.3) + (Freshness × 0.2) + (Confidence × 0.2)

Keep the **top 10** bounties. If fewer than 10 qualify, keep all of them.

---

## Phase 5: Draft first comments

For each top bounty, draft a **professional first comment** that:
1. Shows you understand the problem (paraphrase the issue)
2. Proposes a concrete approach (2-3 sentences)
3. Asks ONE clarifying question if needed
4. Signals intent to work on it
5. Is concise (3-5 sentences max — no essays)

**Example draft:**
> I'd like to tackle this. My approach: [1-2 sentence plan]. One question before I start — [specific question]. I can have a PR ready within [timeframe].

**Never** include:
- "As an AI..." or "I am an AI assistant..."
- Generic pleasantries or fluff
- Promises you can't keep
- Technical jargon that doesn't add value

---

## Phase 6: Deliverable — human-review deck

**OUTPUT PATHS (MUST follow exactly):**
Write to `output/bounty-report-YYYY-MM-DD.json` (the deck endpoint searches here).

The JSON file is an array of objects:
```json
[
  {
    "repo": "owner/repo",
    "issueNumber": 123,
    "title": "Issue title",
    "url": "https://github.com/owner/repo/issues/123",
    "amount": "$500",
    "approach": "Brief solution approach (2-3 sentences)",
    "draftComment": "The full draft first comment ready to post",
    "score": 8.5,
    "freshness": "2 days ago",
    "competition": "1 existing PR"
  }
]
```

Also write a markdown report to `output/bounty-sweep-report.md` with:
- Executive summary (total found, top 3 highlights)
- Full ranked list with details
- Stats: repos scanned, issues checked, bounties found, spam excluded
- Footer: "Requires human approval before ANY post or submission."

---

## Hard rules
- Never comment, post, PR, or submit anything. Drafting only.
- Never claim you can do something you can't.
- Always use `GITHUB_TOKEN` for API calls (5,000 req/hr vs 10 req/min anonymous).
- If the scan finds nothing, say so plainly — that IS a valid result.
- Focus on QUALITY over quantity — 5 real bounties beat 50 spam hits.
