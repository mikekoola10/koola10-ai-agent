# 🧠 Spiral – Jarvis‑like Personal AI Agent

[![Node.js](https://img.shields.io/badge/Node.js-18-339933?style=flat-square&logo=node.js)](https://nodejs.org)
[![Render](https://img.shields.io/badge/deployed%20on-Render-46E3B7?style=flat-square&logo=render)](https://render.com)
[![Python](https://img.shields.io/badge/Python-3.11-3776AB?style=flat-square&logo=python)](https://python.org)
[![Chat](https://img.shields.io/badge/chat%20endpoint-live-00ff88?style=flat-square)](https://spiral-ai-agent.onrender.com/chat)
[![Helix UI](https://img.shields.io/badge/Helix%20UI-available-ffaa00?style=flat-square)](https://spiral-ai-agent.onrender.com)

---

**Spiral** is your personal Jarvis‑like AI assistant – a full‑featured agent that handles chat, health management, schedule planning, video ingestion, and autonomous purchasing.

### ✨ Key Features

- **💬 Natural Language Chat** – Powered by DeepSeek, with a clean Helix UI.
- **🏥 Health Swarm** – Tracks inventory, schedules, supplements, and extracts ingredients from videos (YouTube, Facebook).
- **🛒 Autonomous Purchasing** – Uses AgentCard virtual cards to order supplies when inventory runs low.
- **📅 Smart Scheduling** – Creates daily routines and sends reminders via AgentMail.
- **🎥 Video Ingestion** – Send a link, and Spiral extracts ingredients, notes, and adds to your watch‑later list.
- **📧 Email Commands** – Manage everything by email – `summary`, `schedule`, `purchase`, `video add`.

---

### 🖥️ Helix UI – Your Command Center

![Helix UI](https://via.placeholder.com/800x400?text=Helix+UI+Screenshot+–+Dark+Terminal+Style)

- **Chat with Spiral** – ask anything.
- **Create virtual cards** – for subscriptions or one‑time purchases.
- **View health inventory** – and receive low‑stock alerts.

🔗 **Try it live**: [spiral-ai-agent.onrender.com](https://spiral-ai-agent.onrender.com)

---

### 🛠️ Quick Start

1. **Clone the repo**:
   ```bash
   git clone https://github.com/mikekoola10/spiral-ai-agent.git
   cd spiral-ai-agent
```

2. Set environment variables (Render or local):
   · DEEPSEEK_API_KEY
   · AGENTCARD_JWT
   · AGENTMAIL_API_KEY
3. Deploy to Render (auto‑deploy from GitHub) or run locally:
   ```bash
   cd backend/gateway
   npm install
   node server.js
   ```
4. Open your browser at http://localhost:3000 – Helix UI is served.

---

🧩 Extensibility

· Add new health items: Update /data/health_inventory.json.
· Add new commands: Extend email_parser.go (in Koola10) or the chat handler.
· Integrate with Koola10: Spiral can call Koola10's API for ledger operations.

---

📜 License

MIT – see LICENSE.

---

"Spiral isn't just an assistant – it's your digital twin, always looking ahead."
