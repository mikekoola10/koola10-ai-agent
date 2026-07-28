# Koola10 Automation Toolkit

Proactive revenue-generation automation tools that work 24/7 on your behalf.

## 🚀 Quick Start

```bash
# Run all scanners
python3 cron-jobs/automation/orchestrator.py

# Run specific scanner
python3 cron-jobs/automation/orchestrator.py --scanners bounty

# Generate full report
python3 cron-jobs/automation/orchestrator.py --report report.md
```

## 📦 Available Scanners

### 1. Bounty Scanner (`bounty_scanner.py`)
Scans GitHub for bounty opportunities across repositories.

```bash
# Scan all targets
python3 cron-jobs/automation/bounty_scanner.py

# Only $500+ bounties
python3 cron-jobs/automation/bounty_scanner.py --min-bounty 500

# Output as JSON
python3 cron-jobs/automation/bounty_scanner.py --json
```

**Features:**
- Searches 30+ high-value repositories
- Extracts bounty amounts from issue text
- Prioritizes high-value opportunities ($500+)
- Respects GitHub API rate limits

### 2. Content Auto-Poster (`content_poster.py`)
Posts articles to Medium, Dev.to, and LinkedIn.

```bash
# Post to all platforms
python3 cron-jobs/automation/content_poster.py \
  --title "My Article" \
  --body-file article.md \
  --tags ai revenue automation

# Dry run (preview only)
python3 cron-jobs/automation/content_poster.py \
  --title "My Article" \
  --body "Content here" \
  --dry-run
```

**Required Environment Variables:**
```bash
export MEDIUM_TOKEN="your-medium-integration-token"
export DEVTO_API_KEY="your-devto-api-key"
export LINKEDIN_TOKEN="your-linkedin-oauth-token"
export LINKEDIN_PERSON_ID="urn:li:person:xxxxx"
```

### 3. Competitor Monitor (`competitor_monitor.py`)
Tracks pricing and feature changes on competitor websites.

```bash
# Monitor all targets
python3 cron-jobs/automation/competitor_monitor.py

# Monitor custom URLs
python3 cron-jobs/automation/competitor_monitor.py \
  --urls https://example.com/pricing https://rival.com/pricing

# Output as JSON
python3 cron-jobs/automation/competitor_monitor.py --json
```

**Monitors:**
- Vercel, Netlify, Railway, Fly.io, Render
- Supabase, PlanetScale, Firebase
- Custom URLs (add your own)

### 4. Grant Opportunity Finder (`grant_finder.py`)
Scans for AI/SaaS business grants and funding.

```bash
# Scan all sources
python3 cron-jobs/automation/grant_finder.py

# Only $10k+ grants
python3 cron-jobs/automation/grant_finder.py --min-amount 10000

# Output as JSON
python3 cron-jobs/automation/grant_finder.py --json
```

**Scans:**
- GitHub Sponsors
- Y Combinator Grants
- SBIR/STTR Programs
- NSF Innovation Corps
- Open Source Collective

## 🔧 Setting Up API Tokens

### GitHub Token (Optional but Recommended)
1. Go to https://github.com/settings/tokens
2. Create a new token with `public_repo` scope
3. Set: `export GITHUB_TOKEN="ghp_xxxxx"`

### Medium Token
1. Go to https://medium.com/me/settings → Integration tokens
2. Generate a new token
3. Set: `export MEDIUM_TOKEN="xxxxx"`

### Dev.to API Key
1. Go to https://dev.to/settings/extensions
2. Generate a new API key
3. Set: `export DEVTO_API_KEY="xxxxx"`

### LinkedIn Token
1. Create app at https://www.linkedin.com/developers/
2. Enable "Share on LinkedIn" product
3. Generate OAuth2 access token
4. Set: `export LINKEDIN_TOKEN="xxxxx"`
5. Set: `export LINKEDIN_PERSON_ID="urn:li:person:xxxxx"`

## 📅 Automated Scheduling

Add to your cron-job.org schedule:

```
# Bounty scan every 6 hours
0 */6 * * * cd /app && python3 cron-jobs/automation/orchestrator.py --scanners bounty --report /tmp/bounty-report.md

# Competitor check daily
0 9 * * * cd /app && python3 cron-jobs/automation/orchestrator.py --scanners competitor

# Grant scan weekly
0 10 * * 1 cd /app && python3 cron-jobs/automation/orchestrator.py --scanners grant

# Full report daily
0 8 * * * cd /app && python3 cron-jobs/automation/orchestrator.py --report /tmp/daily-report.md
```

## 💰 Revenue Impact

| Scanner | Potential Revenue | Frequency |
|---------|------------------|-----------|
| Bounty Scanner | $500-$5000+/bounty | 6x daily |
| Content Poster | Traffic → conversions | Daily |
| Competitor Monitor | Pricing intelligence | Daily |
| Grant Finder | $10k-$100k+/grant | Weekly |

## 🛠️ Adding Custom Scanners

1. Create a new file: `cron-jobs/automation/my_scanner.py`
2. Implement a class with a `find()` or `scan()` method
3. Add to `SCANNERS` dict in `orchestrator.py`
4. Run: `python3 orchestrator.py --scanners my_scanner`

## 📊 Sample Output

```
🤖 KOOLA10 AUTOMATION ORCHESTRATOR v2.0
============================================================
Scanners: bounty, competitor, grant
Time: 2026-07-27 12:00 UTC
============================================================

🔍 Koola10 Bounty Scanner v2.0
   Scanning 30 targets...

  📋 Searching: label:bounty state:open
     → Found 12 bounties
  📋 Searching: "$500" in:body state:open
     → Found 8 bounties

  ✅ Total bounties found: 47

============================================================
📊 ORCHESTRATION COMPLETE
============================================================
  ✅ Bounty: 47 results
  ✅ Competitor: 8 results
  ✅ Grant: 15 results
============================================================
```
