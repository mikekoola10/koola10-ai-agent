# 🚀 Koola-10 AI Agent Ecosystem: System Catch-Up Brief

Copy and paste this into a new conversation to restore the full context of the project.

---

## 1. Executive Summary
Koola-10 is an autonomous agent ecosystem designed for end-to-end business operations. It coordinates multiple specialized AI swarms to manage money, write software, find leads, and handle compliance autonomously.

## 2. Technical Architecture
- **Orchestrator (Go 1.22)**: The "brain" (`main.go`) using `chi` router. Manages state, the economic ledger, and swarm coordination.
- **Browser Agent (Python/FastAPI)**: UI automation (`browser-agent/`) using Playwright/browser-use.
- **Semantic Agent (Python/FastAPI)**: Memory specialist (`semantic-agent/`) using `sentence-transformers` for vector embeddings.
- **Infrastructure**: Deployed on Fly.io with a persistent `/data` volume and Redis Pub/Sub for inter-agent communication.

## 3. Core Modules & Swarms
- **Economic Ledger**: Implements a **30/70 revenue split** (30% Ops / 70% General Ledger). Automates Fly.io invoice payments.
- **Swarm Manager**: Coordinates 7+ specialized swarms including:
    - **Sterling (Finance)**: Reporting & P&L.
    - **Nova (Leads/Grants)**: Lead gen & Grant writing.
    - **Forge (Dev)**: "Night Shift" autonomous coding/deployment.
    - **Echo (API)**, **Solara (Content)**, **Sage (Compliance)**, **Vale (Research)**.
- **Memory Systems**:
    - **Graph Memory**: Tracks relationships between meetings, entities, and decisions.
    - **Semantic Index**: Vector search across all generated documents and narratives.

## 4. Operational Tools
- **Stripe**: Integrated for checkout sessions and subscription management.
- **Crypto**: Paper trading tools for strategy execution and price monitoring.
- **HuggingFace**: Direct integration for model search and inference.
- **Studio Module**: Lore management and style-governed AI video generation jobs.

## 5. Current State & Known Discrepancies
*Important: Some features mentioned in long-term memory may be inactive or in separate branches:*
- **Active**: Go Orchestrator, Swarm Manager, Economic Ledger, Grant Search/Apply, Studio Lore.
- **Missing/In-Progress**: Privacy.com card generation, automated CashApp payouts, and the 8 AM "Morning Brief" scheduler are not currently active in the `main` code.

## 6. Critical Credentials Needed
- `DEEPSEEK_API_KEY`, `STRIPE_API_KEY`, `FLY_API_TOKEN`, `HUGGINGFACE_API_TOKEN`, `REDIS_URL`.
