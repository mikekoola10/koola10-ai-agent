# ⚡ Nova — Manus-style Autonomous AI Agent

Give Nova an open-ended task. It plans, uses tools (shell, files, web),
iterates on the results, and delivers a finished artifact — no hand-holding.

Built for the koola10 team as a self-hosted, low-cost alternative to cloud
agents like Manus. Zero runtime dependencies, powered by DeepSeek (OpenAI-compatible).

```
┌─────────────────────────────────────────────────────────────┐
│  you:  nova "audit this repo and write a summary report"    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                ┌──────────────────┐
                │   Agent loop     │
                │  plan → act →    │
                │  observe → repeat│
                └──────────────────┘
                  │  tool calls    │  final report
        ┌─────────┼─────────┬──────┴─────────┐
        ▼         ▼         ▼                ▼
   run_command  read_file  web_search   output/last-report.md
   (bash)      write_file  fetch_url
               list_dir
```

## Quick start

```bash
# 1. Install (Node 20+)
npm install

# 2. Add your key
cp env.example .env        # then edit .env → DEEPSEEK_API_KEY=sk-...
# (or: export DEEPSEEK_API_KEY=sk-...)

# 3. Build & run
npm run build
nova "List the files in this repo and summarize what the project does"

# run without building:
npm run dev -- "your task here"
```

## Try it instantly (no API key)

```bash
nova --mock "explore this directory and write a summary"
```

Mock mode swaps the real brain for a scripted one so you can exercise the full
loop — tool dispatch, artifact writing, final report — for free.

## Nova UI — Manus-style web interface

Nova ships with a web UI that mirrors the Manus desktop layout — dark sidebar
(New task, Agent, Search, Plugins, Scheduled, Library, Projects, Tasks), a
"What can Nova do for you?" home, a task input bar with a provider picker,
quick-action pills, and a live task view that streams steps and the final
report.

```bash
npm run build
npm run ui        # or: nova-ui  (node dist/server.js)
# → http://0.0.0.0:3000   (set PORT to change)
```

No API key set? The UI auto-switches to **mock mode** so you can demo the whole
loop — the top bar shows a 🎭 badge.

The UI is a tiny zero-dependency HTTP server (`src/server.ts`) with a simple
REST API you can script against:

| Endpoint | Description |
|---|---|
| `GET /api/health` | Provider, model, mock flag, enabled connectors, tool count |
| `GET /api/tasks` | Task history (id, status, step/tool counts), newest first |
| `POST /api/tasks` | Start a task — `{ "task": "...", "provider": "deepseek" }` → `{ "id" }` |
| `GET /api/tasks/:id` | Full task: live steps, tool previews, final report |

Tasks live in memory (reset when the server restarts) and run concurrently.
The provider picker really switches brains — the server re-resolves the
matching API key for `anthropic` / `openai`.

Freebuff preview config for this repo:

```
install: npm install
build:   npm run build
preview: npm run ui     (PORT is injected by the platform)
```

## Usage

```
nova "<task description>" [options]

  --mock              scripted brain (no API key; for testing)
  --model <name>      override model (default deepseek-chat)
  --max-steps <n>     override max loop iterations (default 25)
  --verbose, -v       print full tool outputs
  --no-report         don't save output/last-report.md
  --list-tools        list the tools Nova can use
  -h, --help          help
  --version           version
```

### Example tasks

```bash
nova "Audit this repo: find TODO/FIXME markers, list the main entry points, and write a summary to output/audit.md"
nova "Research the top 5 open-source AI agent frameworks, compare them in a table, and save the report"
nova "Check if the dev server on port 3000 is healthy and report the response time"
nova --model deepseek-chat --max-steps 40 "Plan and implement the missing feature described in TODO.md"
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `NOVA_PROVIDER` | `deepseek` | Model provider: `deepseek` \| `anthropic` (Claude) \| `openai` |
| `DEEPSEEK_API_KEY` | — | Required for the `deepseek` provider (platform.deepseek.com → API Keys) |
| `ANTHROPIC_API_KEY` | — | Required for the `anthropic` provider (console.anthropic.com) |
| `OPENAI_API_KEY` | — | Required for the `openai` provider |
| `NOVA_MODEL` | per-provider | `deepseek-chat` \| `claude-sonnet-4-5` \| `gpt-4o-mini` |
| `NOVA_API_BASE` | per-provider | Override the API base URL |
| `NOVA_MAX_STEPS` | `25` | Max agent-loop iterations (bounds cost on long tasks) |
| `NOVA_MAX_TOOL_OUTPUT` | `8000` | Max chars per tool result fed back to the model |
| `NOVA_TOOL_TIMEOUT_MS` | `60000` | Timeout for shell commands |
| `NOVA_SAVE_REPORT` | `1` | Save the final report to `output/last-report.md` |

Env vars are read from `.env` / `.env.local` in the working directory or from
the shell environment.

## Connectors (GitHub · Stripe · Clawdbot)

Nova ships three connector tools, each enabled simply by setting its key:

| Connector | Env var | What Nova can do |
|---|---|---|
| **GitHub** | `GITHUB_TOKEN` | Read repos/issues/PRs, open issues, create PRs, inspect any REST path (`github` tool) |
| **Stripe** | `STRIPE_SECRET_KEY` | Check balance, list customers/subscriptions/charges, read payment intents (`stripe` tool) |
| **Clawdbot** | `CLAWDBOT_CLI` (`openclaw`) | Ask the local OpenClaw agent (formerly Clawdbot) to send chat messages (WhatsApp/Telegram/Discord/Slack) or dispatch tasks to its own agent loop (`clawdbot` tool) |

```bash
# Enable them in .env:
GITHUB_TOKEN=ghp_...          # https://github.com/settings/tokens (fine-grained works best)
STRIPE_SECRET_KEY=sk_test_... # test keys first — never sk_live_ unless you want real charges
npm install -g openclaw       # then CLAWDBOT_CLI=openclaw
```

Example tasks:

```bash
nova "Check my Stripe balance and list the last 5 customers, then summarize"
nova "Open an issue in mikekoola10/koola10-nova-agent titled 'v0.2: add browser tool' with a description"
nova "Send a WhatsApp message via Clawdbot to +15551234567: 'Deploy is green ✅'"
```

> ⚠️ The `stripe` tool can also create charges/refunds. Prefer read-only
> endpoints unless the task explicitly requires a payment action.

## Tools

| Tool | What it does |
|---|---|
| `run_command` | Runs a bash command (build, test, git, install, inspect) with timeout |
| `list_directory` | Lists one directory level with sizes |
| `read_file` | Reads a UTF-8 text file (truncated safely) |
| `write_file` | Writes a file, creating parent directories — for deliverables |
| `web_search` | Keyless DuckDuckGo web search (ranked results) |
| `fetch_url` | Fetches a page and returns readable text |
| `github` | GitHub REST API via `GITHUB_TOKEN` |
| `stripe` | Stripe REST API via `STRIPE_SECRET_KEY` |
| `clawdbot` | OpenClaw (Clawdbot) bridge — send messages / dispatch agent runs |
| `browser` | Browser use — real headless Chromium via Playwright (open/click/type/screenshot) |
| `computer` | Computer use — desktop control via xdotool (type/key/mouse/screenshot) |

Run `nova --list-tools` to see the live schemas.

## Browser & computer use

Two Clawdbot-style powers, both opt-in:

### Browser use (`browser` tool)

Nova drives a real Chromium browser — navigate, click, fill forms, extract page
text, take screenshots — and keeps one tab open across tool calls in a run.

```bash
# One-time setup (~1 min):
npm install playwright && npx playwright install chromium
```

```bash
nova "Open https://news.ycombinator.com, summarize the top 5 stories"
nova "Search for 'nova agent' on Google, open the first result, and save a screenshot to output/"
```

Set `NOVA_BROWSER_HEADLESS=0` to run with a visible window (needs a display).

### Computer use (`computer` tool)

Nova can type, press keys, move/click the mouse, and capture the screen on the
machine it runs on — like a remote-hands operator.

```bash
# Requires (Linux/X11):
sudo apt-get install -y xdotool
# and an X display:  export DISPLAY=:0   (headless servers: run under Xvfb)
```

```bash
nova "Open a terminal, run 'htop', and screenshot the screen to output/"
```

> Both tools degrade gracefully: if Playwright or xdotool is missing, Nova
> returns exact install instructions instead of crashing.

## Development

```bash
npm run typecheck   # strict tsc
npm run smoke       # mock-mode end-to-end run
npm run build       # emit dist/
```

Layout:

```
src/
├── index.ts        # CLI entry (flags, banner, progress, report save)
├── server.ts       # Nova UI — zero-dep HTTP server, REST API, serves web/
├── agent.ts        # the agent loop + system prompt + context trimming
├── llm.ts          # DeepSeek/OpenAI-compatible client (+ mock brain)
├── config.ts       # env loading & configuration
├── util.ts         # output helpers
└── tools/
    ├── index.ts    # tool registry + dispatch
    ├── shell.ts    # run_command
    ├── files.ts    # list_directory / read_file / write_file
    └── web.ts      # web_search / fetch_url
```

## Roadmap

- **Computer-use vision loop** — feed screenshots back to a vision-capable
  model so Nova can *see* the screen and reason about what to click
- **Clawdbot gateway pairing** — talk to a running OpenClaw gateway over its
  WebSocket API (`ws://127.0.0.1:18789`) instead of shelling out to the CLI
- **Persistent sessions** — keep a browser profile + memory across runs
- **Nova UI next steps** — desktop wrapper (Electron/Tauri) around the web
  UI, streaming events (SSE) instead of polling, task persistence across
  restarts, and wiring the Koola10 CEO dashboard chat into the same task API
- **Memory & sessions** — persist conversation state and resume long tasks
- **Multi-step planning** — explicit plan → execute → verify phases with checklists
- **Cost guardrails** — token budgets, per-tool rate limits, step budgets

## License

MIT
