# Nova Bounty Sweep

You are a bounty hunter. Find real, paid GitHub bounties and draft a review deck.
**READ-ONLY**: never post, comment, or submit anything.

## Step 1: Search GitHub for open bounties

Use the `github` tool to run these search queries. Each returns issues across ALL public repos:

### Query 1: Algora-verified bounties (real money, paid on merge)
```
label:"💎 Bounty" state:open type:issue sort:created
```
Take ALL results — these are the highest quality.

### Query 2: Bounties with dollar amounts
```
label:bounty "$" state:open type:issue sort:updated
```

### Query 3: General bounty label
```
label:bounty state:open type:issue sort:created
```

### Query 4: Reward bounties
```
label:reward state:open type:issue
```

For each query, get the first 30 results. Extract:
- repo (owner/name)
- issue number
- title
- dollar amount (from labels or title/body)
- URL
- created date
- number of existing comments (competition indicator)

## Step 2: Filter

EXCLUDE:
- Issues by known spammers (Nexussyn, tdpeta754, clanker-journalist)
- "AI Growth Engine", "AiMPN", "ClankerNation" promotional posts
- Issues with no dollar amount AND no bounty label
- Questions ("is there a bounty program?")

KEEP only issues that:
- Have a bounty/reward/💎 label, OR
- Contain explicit dollar amounts ($XX), OR
- Are from verified platforms (Algora, Opire)

## Step 3: Score and rank

For each qualifying bounty, estimate:
1. **Amount**: Parse dollar value (default $50 if unknown)
2. **Competition**: Fewer comments = easier to win
3. **Freshness**: Newer = better
4. **Difficulty**: TypeScript/Python/JS = easy, C++/Rust = hard

Keep the **top 10** by total score.

## Step 4: Write output

Write a JSON file to `output/bounty-report-YYYY-MM-DD.json` (use today's date):
```json
[
  {
    "repo": "owner/repo",
    "issueNumber": 123,
    "title": "Issue title",
    "url": "https://github.com/owner/repo/issues/123",
    "amount": "$500",
    "approach": "2-3 sentence solution approach",
    "draftComment": "Professional first comment (3-5 sentences, shows understanding, proposes approach)",
    "score": 8.5
  }
]
```

Also write `output/bounty-sweep-report.md` with a summary of findings.

## Rules
- Never comment, post, or submit. Drafting only.
- If nothing found, write empty JSON array `[]` and explain why.
