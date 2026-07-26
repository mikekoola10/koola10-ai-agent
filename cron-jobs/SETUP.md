# Cron-job.org Autotraffic Setup

This folder contains the recurring-job manifest that drives automated traffic, lead capture, and revenue pings to **Koola10 AI Agent**.

## Stack
| Piece | Where |
|---|---|
| Cron scheduler | [cron-job.org](https://cron-job.org) (free, browser-based, supports auth headers, retries, alerting) |
| Backend API | `https://koola10-ai-agent.onrender.com` |
| Frontend | `https://koola10aiagent.freebuff.app` |
| Auth header | `Authorization: Bearer <ADMIN_KEY>` (your Render env var `ADMIN_API_KEY`) |

## Files
- `cron-jobs.json` — full machine-readable manifest, 9 staggered jobs
- `setup.py` — bulk-import script that calls cron-job.org's REST API
- `SETUP.md` — this file

## Schedule Philosophy
1. **Stagger** — none of the jobs fire at `:00` (CPU spikes on Render's shared tier)
2. **Interleave** warm-up pings every ~10 min so Render never idle-spins
3. **Throttle** Stripe/DeepSeek endpoints to ≤ 2×/day each (cost-control)
4. **Cluster** heavy work at distinct off-peak hours

## Auth Header
Replace `${ADMIN_KEY}` with the value of your **Render** env var `ADMIN_API_KEY`
(matches what `src/api.js` reads). Format:
```
Authorization: Bearer <ADMIN_KEY>
Content-Type: application/json
```

## Job Summary

| # | Type | Schedule (cron) | Endpoint | Notes |
|---|---|---|---|---|
| 1 | warm-up | `2,22,42 * * * *` | GET /health | Render keep-alive A |
| 2 | warm-up | `12,32,52 * * * *` | GET /health | Render keep-alive B |
| 3 | traffic | `15,45 * * * *` | POST /admin/analytics/track | Dashboard padding |
| 4 | lead-gen | `18 9,21 * * *` | POST /admin/email/signup | Passive list growth |
| 5 | traffic | `25 */4 * * *` | POST /admin/content/syndicate | Cross-platform push |
| 6 | maintenance | `35 */6 * * *` | POST /blog/generate | DeepSeek post |
| 7 | revenue | `45 14,23 * * *` | POST /admin/trigger_affiliate | Stripe vertical |
| 8 | revenue | `5 9,18 * * *` | POST /admin/trigger_bounty | Grant / bounty |
| 9 | revenue | `11 3,15 * * *` | POST /admin/run-scheduled-sprint | Full sprint |

## Setup — Manual (UI)

1. Sign in at https://cron-job.org (free tier allows 100 jobs)
2. Click **Create new cronjob** for each row above
3. **Title** — use the **Name** column above
4. **URL** — paste the full URL (scheme + host + path)
5. **Schedule** — switch to **Cron expression** mode and paste the cron expression
6. **Request method** — choose GET or POST per row
7. **Request headers** — add `Authorization: Bearer <your ADMIN_KEY>` and (for POSTs) `Content-Type: application/json`
8. **Request body** — for POSTs, paste the JSON body from the manifest
9. **Enabled** — yes
10. Save & repeat x9

## Setup — Automated (API)

`setup.py` reads `cron-jobs.json`, replaces `${ADMIN_KEY}`, and POSTs each job to
`https://api.cron-job.org/jobs`. You'll need a cron-job.org **API key** (Account → API).

```bash
export CRON_JOB_ORG_API_KEY='your-key-here'
export ADMIN_KEY='your-render-admin-key'
python3 cron-jobs/setup.py
```

## Verifying It Worked

After saving, cron-job.org logs each run. To see the live effect on your Render backend:

```bash
curl -s 'https://koola10-ai-agent.onrender.com/admin/cron/jobs' \
  -H "Authorization: Bearer $ADMIN_KEY"
```
You should see the analytics history grow and the email signup tally tick up.

## Pitfalls

- **`/admin/cron/jobs` returned 404 on first test** — your Render build may predate
  the new endpoints. Push the latest `main.go`/Vercel frontend and re-deploy.
- **Don't mount heavier endpoints with sub-hour frequencies** — Stripe + DeepSeek
  will rack up costs within hours.
- **Set timezone to `America/New_York`** in cron-job.org (matches your `CronManager.jsx` preset).
- **Email signups from cron use dummy addresses** — they pad the count but aren't
  real leads. Wire your waitlist form to call the same endpoint with real emails.

## Next Steps

- Add a UptimeRobot or cron-job.org monitoring alert to the frontend URL
- Add an `email.send_drip` cron job to your mailer (Resend SDK) once you wire it
- Add a `business_metrics.daily_summary` cron once you wire analytics storage
