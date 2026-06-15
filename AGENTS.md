# Koola10 Autonomous Architecture & Guidelines

Welcome to the Koola10 ecosystem. This file defines the 24/7 autonomous architecture and providing instructions for agents (like Jules) and developers.

## Core Philosophy: The Non-Stop Workflow

Koola10 is designed for **continuous, multi-device autonomy**. It leverages three core pillars:
1.  **Go Orchestrator:** The brain. Manages swarms, financial ledgers, and global state.
2.  **Google Jules:** The Senior Architect. Operates asynchronously on the codebase to implement features and fix bugs.
3.  **DeepSeek + Browser/Device Agents:** The hands. Executes real-time tasks on the web, desktop, and mobile devices.

## 24/7 Operational Rules

- **Self-Healing:** Every task MUST have a supervisor check. If a task fails, the agent must trigger a `/diagnose` endpoint to find the root cause and retry (limit: 3 attempts).
- **Infrastructure Monitoring:** The `DependencyWatchdog` pings critical services every 5 minutes. If a dependency is down, it reports to the `Supervisor` and logs a `CRITICAL_INFRASTRUCTURE_FAILURE` in the audit chain.
- **Financial Autonomy:** All actions costing > $0.05 must be cleared by the `EconomicLedger` via `EvaluateAction`. Revenue must be gross-recorded to the Global Ledger.
- **Cross-Device Coordination:**
    - **Desktop Swarm:** Controls the host OS for local environment setup and long-running scripts.
    - **Mobile Swarm (Droid Run):** Handles mobile-specific automation (SMS OTPs, app-based payouts).
- **Persistence:** All swarm state must be persisted to `/data` (e.g., `bills.json`, `leads/`, `audit_chain.jsonl`).

## Agent Instructions (Jules)

- **Asynchronous Tasking:** When you receive a complex task, focus on creating a robust plan and implementing the core logic.
- **Tool Creation:** You are encouraged to create Python-based tools in `/home/jules/self_created_tools` to aid your workflow.
- **Context Awareness:** Always read the `EconomicLedger` state before proposing financial operations.
- **Reliability Check:** Before starting high-stakes or long-running tasks, agents should verify system health via `GET /system/health`.
- **Verification:** Every code change must be verified with `go build` or `go test`. Frontend changes require Playwright screenshots.

## System Components

- `/agents`: Autonomous swarm implementations.
- `/browser-agent`: Playwright-based web automation.
- `/device-agent`: (Planned) Desktop and mobile control service.
- `/financial`: Ledger and fund management.
- `/tools`: Reusable utility tools (Crypto, Stripe, etc.).
