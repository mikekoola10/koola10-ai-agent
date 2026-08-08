# Nova Landing Page

A simple, self-contained one-page landing site for the **Nova** autonomous AI
agent. Built with plain HTML, CSS, and vanilla JavaScript — no frameworks, no
build step, no dependencies.

## Contents

```
output/site/
├── index.html   # Single-page landing site
├── styles.css   # All styling (dark, responsive, no framework)
├── script.js    # Small progressive enhancements (year, smooth scroll, reveal)
└── README.md    # This file
```

## Quick start

Because it's plain static files, just open `index.html` in a browser:

```bash
open output/site/index.html      # macOS
xdg-open output/site/index.html  # Linux
start output/site/index.html     # Windows
```

Or serve it locally with any static server:

```bash
cd output/site
python3 -m http.server 8080
# then visit http://localhost:8080
```

## What's on the page

- **Hero** — headline, call to action, and a mock terminal showing a Nova run.
- **Features** — a grid of six capabilities (plan, act, observe, report, private, connectors).
- **How it works** — the four-step plan → act → observe → report loop.
- **Connectors** — GitHub, Stripe, Slack, Gmail, Zapier, n8n, Make, Hugging Face, and more.
- **FAQ** — collapsible Q&A (native `<details>`/`<summary>`).
- **CTA + footer** — install commands and dynamic year.

## Design notes

- **No framework, no build step** — everything is hand-written and dependency-free.
- **Responsive** — grid collapses from 3 → 2 → 1 column on smaller screens.
- **Accessible** — skip link, semantic landmarks, ARIA labels, keyboard-friendly.
- **Progressive enhancement** — the scroll-reveal and sticky-nav effects are
  added by JS only; the page is fully usable without it.
- **System + Google fonts** — `Inter` for text, `JetBrains Mono` for code/terminal.
  The Google Fonts `<link>` is optional; the page falls back to system fonts if offline.

## Customizing

- To change colors, edit the CSS custom properties at the top of `styles.css`
  (e.g. `--primary`, `--accent`, `--bg`).
- To edit copy, sections, or the connector list, edit `index.html`.
- To point the CTA buttons somewhere real, update the `href` attributes in the
  `#start` section of `index.html`.

## License

Part of the Nova project for the koola10 team.
