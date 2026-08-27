# Nova Bounty Sweep — standing order

You are executing the standing bounty sweep for Koola10. **This task is READ-ONLY
against GitHub**: you may scan, read, analyze, and draft — you must NEVER post
comments, open issues/PRs, or take any visible action. Everything you produce is
for a human to review and approve first.

## Mission
Find real, open, winnable GitHub bounties, rank them by (value × likelihood of
winning), and deliver a human-review deck.

## 1. Scope — repositories
Prefer the `github` tool when `GITHUB_TOKEN` is set (5,000 req/hr). Otherwise use
`curl -s` against `https://api.github.com` with a 1–2s pause between requests
(anonymous limit ~10 req/min); if you hit rate limits, note it and continue.

**Dedicated bounty programs (highest priority):**
- TheSCInitiative/bounties
- projectdiscovery/oss-bounty-program
- zama-ai/bounty-program
- stacksgov/critical-bounties
- Concordium/Concordium-Free-Open-Grants-Program

**AI / open-source repos:**
open-webui/open-webui, langchain-ai/langchain, run-llama/llama_index,
microsoft/autogen, crewAIInc/crewAI, Significant-Gravitas/AutoGPT,
Pythagora-io/gpt-pilot, lobehub/lobe-chat, danny-avila/LibreChat,
huggingface/transformers, vllm-project/vllm, ollama/ollama,
ggerganov/llama.cpp, letta-ai/letta, mem0ai/mem0, BerriAI/litellm,
Aider-AI/aider, All-Hands-AI/OpenHands, camel-ai/camel,
ScrapeGraphAI/Scrapegraph-ai, ComposioHQ/composio, phidatahq/phidata,
jina-ai/reader

If `GITHUB_TOKEN` is NOT set, scan only the 5 dedicated bounty programs plus the
first 5 AI repos (10 repos total) and add a note that the full 28-repo sweep
requires `GITHUB_TOKEN`.

## 2. Search strategy (per repo)
1. Open issues labeled `bounty`
2. `"bounty" in:title` open issues; add `"bounty" in:body` if results are thin
3. On dedicated programs, list ALL open issues (they are all bounty material)
4. Verify by listing the repo's labels — if a `bounty` label exists, search by it

Extract dollar amounts from title/body. **Exclude** promotional/spam issues
(third-party "collaboration" posts, random USDC offers from strangers,
auto-generated PRs). Note security-only programs locked behind private
disclosure separately.

## 3. Ranking
Score each candidate by: amount × (confidence you could solve it) × (open and
unassigned) × (recent activity). Keep the **top 5**.

## 4. Deliverable — human-review deck (never auto-submit)
Write `web/artifacts/bounties/bounty-report-YYYY-MM-DD.md` (create the directory
if needed). For each top candidate include:
- issue title, URL, repo
- bounty amount and any deadline
- why it's winnable and what it actually requires
- **Draft solution approach** — a concrete plan, not a code dump
- **Draft first comment** — a professional message (ask a clarifying question or
  signal intent) ready for a human to edit and approve. It must NOT be posted by
  you.
- Footer: "Requires human approval before ANY post or submission."

ALSO write a JSON sidecar file at `web/artifacts/bounties/bounty-report-YYYY-MM-DD.json`
with the same date. This file is an array of objects, one per top bounty:
```json
[
  {
    "repo": "owner/repo",
    "issueNumber": 123,
    "title": "Issue title",
    "url": "https://github.com/owner/repo/issues/123",
    "amount": "$500",
    "approach": "Brief solution approach",
    "draftComment": "The full draft first comment ready to post"
  }
]
```
The JSON is what the Nova UI uses to show Approve buttons. Make sure issueNumber
is the integer issue number extracted from the URL.

Finish with a 5-line summary in your final report, including the deck path.

## Hard rules
- Never comment, post, PR, or submit anything. Drafting only.
- Never claim you can do something you can't.
- If the scan finds nothing, say so plainly — that IS a valid result.
