#!/usr/bin/env node
/**
 * Nova static build — emits the self-contained Nova UI into dist/.
 *
 * Freebuff production hosting runs the build command in a clean checkout with
 * Node available, then serves the static output from dist/. Nova's web UI is a
 * single dependency-free HTML file (nova-agent/web/index.html), so this script
 * just copies it (plus optional crawler files) into dist/.
 *
 *   npm run build   →   node scripts/build-nova-static.mjs
 */
import { copyFileSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const out = join(root, "dist");

// Fresh dist every build.
rmSync(out, { recursive: true, force: true });
mkdirSync(out, { recursive: true });

// Nova UI — the whole app in one HTML file.
copyFileSync(join(root, "nova-agent", "web", "index.html"), join(out, "index.html"));

// Optional crawler files from public/ (robots.txt, sitemap.xml) if present.
for (const f of ["robots.txt", "sitemap.xml"]) {
  const src = join(root, "public", f);
  if (existsSync(src)) copyFileSync(src, join(out, f));
}

console.log("✓ Nova static UI → dist/index.html");
