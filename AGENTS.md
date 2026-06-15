# Koola10 Autonomous Architecture & Guidelines

Welcome to the Koola10 ecosystem. This file defines the 24/7 autonomous architecture and providing instructions for agents (like Jules) and developers.

## Core Philosophy: The Non-Stop Workflow

Koola10 is designed for **continuous, multi-device autonomy**. It leverages three core pillars:
1.  **Go Orchestrator:** The brain. Manages swarms, financial ledgers, and global state.
2.  **Google Jules:** The Senior Architect. Operates asynchronously on the codebase to implement features and fix bugs.
3.  **DeepSeek + Browser/Device Agents:** The hands. Executes real-time tasks on the web, desktop, and mobile devices.

## 24/7 Operational Rules

- **Self-Healing:** Every task MUST have a supervisor check. If a task fails, the agent must trigger a `/diagnose` endpoint to find the root cause and retry (limit: 3 attempts).
- **E2E Oversight:** The system operates with an "End-to-End" mindset. The watchdog performs functional verification (e.g., tool execution tests) every 5 minutes to ensure the system works "1st time around".
- **Infrastructure Monitoring:** The `DependencyWatchdog` pings critical services. If a dependency is down, it reports to the `Supervisor` and logs a `CRITICAL_INFRASTRUCTURE_FAILURE` in the audit chain.
- **Self-Evolution:** The `MetaSwarm` is responsible for scouting GitHub for future upgrade ideas and analyzing system performance to optimize Koola10's roadmap and code.
- **Financial Autonomy:** All actions costing > $0.05 must be cleared by the `EconomicLedger` via `EvaluateAction`. Revenue must be gross-recorded to the Global Ledger.
- **The Wizard's Code (UI/UX):** All frontend development must adhere to the "High-Fidelity Cyberpunk" aesthetic established in the Master Command Portal. Use neon accents, glitch effects, and real-time telemetry.
- **Cross-Device Coordination:**
    - **Desktop Swarm:** Controls the host OS for local environment setup and long-running scripts.
    - **Mobile Swarm (Droid Run):** Handles mobile-specific automation (SMS OTPs, app-based payouts).
- **Persistence:** All swarm state must be persisted to `/data` (e.g., `bills.json`, `leads/`, `audit_chain.jsonl`).

## Agent Instructions (Jules)

- **Asynchronous Tasking:** When you receive a complex task, focus on creating a robust plan and implementing the core logic.
- **Tool Creation:** You are encouraged to create Python-based tools in `/home/jules/self_created_tools` to aid your workflow.
- **Google Jules Magic:** Use the "Idea -> Prototype -> Validate -> Merge" workflow. Meta-agents (Idea Hunter) propose features, Jules prototypes them, the E2E Watchdog validates, and the Senior Architect (Jules) merges.
- **Wizard's Shield (Safe Autonomy):** Any PR or event that modifies financial logic (`financial/*.go`) or triggers a payout is automatically held for manual approval in the Master Command Portal.
- **Staging Verification:** Fixes are tested in a staging environment (defined by `DEVICE_AGENT_ENV`) before being proposed for production.
- **Autonomous Recovery:** The system uses an AI-readable Recovery Map (`data/recovery_map.json`). Meta-agents should update this map when new failure patterns or recovery actions are discovered.
- **Circuit Breaker:** If a task fails > 5 times, the Engine enters SAFE MODE and sends an alert to `mikekoola10@agentmail.to`. Manual intervention via the Portal is required to reset.
- **Context Awareness:** Always read the `EconomicLedger` state before proposing financial operations.
- **Reliability Check:** Before starting high-stakes or long-running tasks, agents should verify system health via `GET /system/health`.
- **AgentMail Notifications:** The `agentmail` tool is available for sending reports and alerts. The system uses it for escalation when autonomous recovery fails.
- **Two-Way Email Protocol:** Agents can now receive and act on emails. Inbound emails trigger the `E2E Watchdog` via `/webhook/agentmail`. Sensitive outgoing emails are held for manual approval.
- **Global Communication Grid:** Agents are connected via Email, Slack, and the A2A Bridge. Critical alerts are automatically cross-posted. Use the `messaging` tool for Slack/SMS.
- **A2A Interoperability:** Use `/a2a/delegate` to hand off tasks to external peer agents (e.g., Spiral). Ensure the payload follows the `A2AMessage` schema.
- **Infrastructure Mastery:**
    - **Render:** Deploy using the root `render.yaml`. Use the Master Command Portal to monitor cross-service telemetry.
    - **Fly.io:** Managed autonomously by the `Engine` using `flyctl`. Ensure `metaclaw_data` volume is attached.
- **Deployment Reliability (Wizard's Sentry):**
    - **External Liveness:** Use external monitoring (UptimeRobot, etc.) to hit `/system/health`.
    - **Recovery Webhook:** External failures should trigger `POST /system/webhook/recovery` with the `RECOVERY_WEBHOOK_SECRET`.
    - **Smoke Tests:** Hourly internal verification runs via `smoke_test.sh`. Failure triggers automated rollback.
    - **Deployment Lock:** The `data/DEPLOYMENT_LOCK` file stores the last known-good commit hash for `flyctl` restoration.
- **Verification:** Every code change must be verified with `go build` or `go test`. Frontend changes require Playwright screenshots.

## System Components

- `/agents`: Autonomous swarm implementations.
- `/browser-agent`: Playwright-based web automation.
- `/device-agent`: (Planned) Desktop and mobile control service.
- `/financial`: Ledger and fund management.
- `/tools`: Reusable utility tools (Crypto, Stripe, etc.).
