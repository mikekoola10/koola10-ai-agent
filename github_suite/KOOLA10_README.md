# 🌀 Koola10 – Autonomous AI Business Engine

[![Go](https://img.shields.io/github/go-mod/go-version/mikekoola10/koola10-ai-agent?style=flat-square)](https://go.dev)
[![Build](https://img.shields.io/github/actions/workflow/status/mikekoola10/koola10-ai-agent/build.yml?style=flat-square&label=Build)](https://github.com/mikekoola10/koola10-ai-agent/actions)
[![Coverage](https://img.shields.io/codecov/c/github/mikekoola10/koola10-ai-agent?style=flat-square)](https://codecov.io/gh/mikekoola10/koola10-ai-agent)
[![License](https://img.shields.io/github/license/mikekoola10/koola10-ai-agent?style=flat-square)](LICENSE)
[![Fly.io](https://img.shields.io/badge/deployed%20on-Fly.io-8A2BE2?style=flat-square&logo=fly.io)](https://fly.io)
[![DeepSeek](https://img.shields.io/badge/powered%20by-DeepSeek-5A29E4?style=flat-square)](https://deepseek.com)

---

**Koola10** is a self‑healing, revenue‑generating AI orchestrator. It manages swarms, a dual‑ledger economic system, full‑duplex email communication, and autonomous infrastructure recovery – all governed by a codified "Secret Agent Manifest".

### 🚀 Key Features

- **🧠 Intelligent Swarms**: Sterling (Finance), Nova (Leads), Forge (Dev), Solara (Content), Sage (Compliance), and a self‑evolving Meta‑Swarm.
- **💰 Economic Ledger**: 30% operations / 70% spendable split with automated payout pipelines (Cash App, Circle USDC).
- **📡 Full‑Duplex AgentMail**: Send and receive emails – control the system via natural language.
- **⚙️ Self‑Healing Infrastructure**: E2E watchdog, external Sentry, automatic rollback to `DEPLOYMENT_LOCK`.
- **🎯 Revenue Swarms**: Affiliate marketing, bug bounty hunting, and BPA API (paid endpoints).
- **🔄 MCP Integration**: Model Context Protocol support – extendable with any MCP server.

---

### 📐 Architecture

```mermaid
graph TD
    A[User / Email / API] --> B[Koola10 Orchestrator]
    B --> C[Swarm Manager]
    B --> D[Economic Ledger]
    B --> E[AgentMail]
    B --> F[Watchdog / Sentry]
    C --> G[Sterling (Finance)]
    C --> H[Nova (Leads)]
    C --> I[Forge (Dev)]
    C --> J[Solara (Content)]
    C --> K[Sage (Compliance)]
    C --> L[Meta‑Swarm]
    D --> M[30% Ops Fund]
    D --> N[70% Spendable]
    E --> O[Incoming / Outgoing Email]
    F --> P[Auto‑Recovery / Rollback]
    G --> Q[AgentCard / Circle USDC / Cash App]
```

---

🛠️ Quick Start

1. Clone the repo:
   ```bash
   git clone https://github.com/mikekoola10/koola10-ai-agent.git
   cd koola10-ai-agent
   ```
2. Set environment variables (Fly.io):
   ```bash
   fly secrets set DEEPSEEK_API_KEY="..."
   fly secrets set AGENTCARD_JWT="..."
   fly secrets set AGENTMAIL_API_KEY="..."
   ```
3. Deploy:
   ```bash
   fly deploy -a koola10
   ```
4. Monitor:
   · Dashboard: https://koola10.fly.dev/vault
   · Monitor: https://koola10.fly.dev/monitor

---

📊 Live Metrics (auto‑fetched)

| Metric | Value |
|--------|-------|
| Total Revenue | ![Total Revenue](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fkoola10.fly.dev%2Fmonitor&query=%24.revenue.total&label=&color=brightgreen) |
| Ops Fund | ![Ops Fund](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fkoola10.fly.dev%2Fmonitor&query=%24.revenue.operations&label=&color=blue) |
| Spendable | ![Spendable](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fkoola10.fly.dev%2Fmonitor&query=%24.revenue.spendable&label=&color=gold) |

---

🧩 Extensibility

· Add a new swarm: Implement the Agent interface in agents/.
· Add a tool: Register in tools/registry.go.
· Connect an MCP server: Use the built‑in tools/mcp_client.go.

---

📜 License

MIT – see LICENSE.

---

"Koola10 doesn't just run – it evolves. Every day it gets smarter, faster, and more profitable."
