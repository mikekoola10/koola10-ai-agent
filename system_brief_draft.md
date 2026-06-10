# Koola-10 AI Agent Ecosystem: System Brief (Draft)

## 1. Technical Architecture
- **Go Orchestrator (v1.22)**: The central brain (`main.go`) using `chi` router. Manages state, economic ledger, and swarm coordination.
- **Browser Agent (Python/FastAPI)**: UI automation specialist (`browser-agent/`) using Playwright and `browser-use`.
- **Semantic Agent (Python/FastAPI)**: Memory specialist (`semantic-agent/`) using `sentence-transformers` for vector embeddings.
- **Data & Persistence**: Persistent `/data` volume on Fly.io. Redis used for Pub/Sub and swarm heartbeats.

## 2. Core Modules
- **Economic Ledger & Fund Manager**:
    - 30/70 revenue split (30% Ops / 70% General Ledger).
    - Automated Fly.io invoice payments.
    - Surplus reinvestment logic.
- **Swarm Network**:
    - **Sterling**: Financial Reporting.
    - **Nova**: Grant Writing & Lead Generation.
    - **Forge**: Developer & Deployment (Night Shift).
    - **Echo**: API Services.
    - **Solara**: Content & Engagement.
    - **Sage**: Compliance & Audit.
    - **Vale**: Research.
- **Memory Systems**:
    - **Memory Graph**: Entity relationships and decision tracking.
    - **Semantic Index**: Vector search for generated narratives and documents.
- **Compliance & Audit**:
    - Cryptographically linked SHA-256 audit logs.
    - Manual approval gates for sensitive actions.
- **Studio Module**: Cinematic universe management (Lorekeeper, Style rules, Video jobs).

## 3. Toolset
- **Stripe**: Checkout automation and subscription management.
- **Crypto**: Paper trading and strategy execution.
- **HuggingFace**: Integrated model search and inference.

## 4. Current Discrepancies (Memory vs. Code)
*Note: These features are documented in system memory but appear missing or moved in the current branch:*
- Privacy.com virtual card integration.
- Morning Brief (8 AM report).
- Portfolio Manager (crypto sweep logic).
- CashApp/PlayStation automated checkout.

## 5. Critical Environment Variables
- `DEEPSEEK_API_KEY` (AI capabilities)
- `STRIPE_API_KEY` (Payments)
- `FLY_API_TOKEN` (Infrastructure payments)
- `HUGGINGFACE_API_TOKEN` (AI Models)
- `REDIS_URL` (Swarm coordination)
