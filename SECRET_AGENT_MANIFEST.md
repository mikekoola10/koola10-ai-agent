🕵️ Jules Secret Agent Manifest (Koola‑10 / Spiral)

Mandatory Objectives (ranked by priority)

1. Keep both Koola‑10 and Spiral deployed and reachable – the orchestrator (koola10) and the agent (spiral-ai-agent) must respond to /health with HTTP 200 at all times. If either fails, you must initiate recovery within 5 minutes.
2. Ensure revenue‑critical services stay up – the following components must be running 24/7: Economic Ledger (Fly.io), AgentCard API, Browser Agent (for subscriptions), and AgentMail (for alerts). If any of these is down for >10 minutes, escalate to the human with a clear error report and a suggested fix.
3. Run the E2E Watchdog and external Sentry – every 10 minutes, trigger a functional test of the system (e.g., a dummy card creation call, a test transaction). If it fails, attempt the recovery actions defined in the Recovery Map and, if needed, rollback to the DEPLOYMENT_LOCK hash.
4. Send a daily revenue summary – every morning at 8 AM, use AgentMail to send an email to mikekoola10@agentmail.to with the following:
   · Total revenue (from /vault/summary)
   · Operations fund balance (30% of revenue)
   · Spendable balance (70%)
   · List of upcoming bills and overdue payments (if any)
   · Any failed auto‑subscribe attempts from the last 24 hours
5. Proactively patch vulnerabilities – scan the codebase weekly for outdated dependencies (e.g., go mod outdated, npm outdated) and open a PR to update them. Include a test suite run to verify no regression.

---

Operational Schedules (UTC)

Task Frequency Execution Method
Health check & recovery Every 5 minutes Call https://koola10.fly.dev/health; if failure, trigger /system/webhook/recovery with {"action":"rollback_deploy"} (using external Sentry if main app down)
Revenue summary email Daily at 08:00 Use AgentMail SDK to compose and send (via agentmail tool)
Dependency scan Weekly on Sunday 00:00 Run go mod outdated, npm outdated; create PR if updates found
Fire drill (simulated crash) Monthly on 1st 03:00 Run fire_drill.py to test recovery loop and email alert

---

Jules‑specific Execution Rules

· Use your existing capabilities – you already have access to the codebase, Fly.io CLI (via flyctl), and Render API (via render CLI). Use them directly; do not ask for permission for routine actions.
· When you make a change (e.g., an auto‑fix), always create a PR with a clear title like [Autofix] ... and include a test proof (e.g., go test ./... output).
· If you cannot resolve a failure within 30 minutes, send an alert email and tag the issue with #severe – the human will take over.
· All actions must be logged – append a JSON line to /data/agent_audit.log for every action taken (check, fix, email, deploy). This log must be accessible via the Portal's telemetry.

---

Escalation Path (if all else fails)

If the entire Koola‑10 app is unreachable and the external Sentry cannot trigger recovery, you (Jules) should:

1. Use flyctl to redeploy the last known good image (stored in DEPLOYMENT_LOCK).
2. If that fails, create a new Fly.io app (koola10-fallback) using the same Dockerfile and switch the DNS A record to it.
3. Notify the human via both email and a new GitHub issue titled EMERGENCY - Fallback activated.
