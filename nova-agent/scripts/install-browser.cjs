#!/usr/bin/env node
/**
 * Installs Playwright Chromium on hosted build environments (Render, CI) so
 * Nova's browser tool works in production. Local dev skips the ~160MB download —
 * run `npm run tools:browser` if you want it locally.
 *
 * Safe-by-default: never fails the install step (worst case, the browser tool
 * reports unavailable for that deploy).
 */
const { spawnSync } = require("node:child_process");

const hosted = !!(process.env.RENDER || process.env.RENDER_SERVICE_ID || process.env.CI);
if (!hosted) {
  console.log("[nova] not a hosted build — skipping Chromium download (local dev). Use: npm run tools:browser");
  process.exit(0);
}

console.log("[nova] hosted build — installing Playwright Chromium…");
const attempts = [
  ["playwright", "install", "--with-deps", "chromium"],
  ["playwright", "install", "chromium"],
];
for (const args of attempts) {
  const r = spawnSync("npx", args, { stdio: "inherit", shell: false, timeout: 600_000 });
  if (r.status === 0) {
    console.log("[nova] Chromium ready.");
    process.exit(0);
  }
  console.warn(`[nova] 'npx ${args.join(" ")}' failed (rc=${r.status}); trying next approach…`);
}
console.warn("[nova] Chromium install did not fully succeed — the browser tool may be unavailable this deploy.");
process.exit(0);
