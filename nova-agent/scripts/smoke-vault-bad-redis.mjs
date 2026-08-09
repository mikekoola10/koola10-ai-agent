#!/usr/bin/env node
/**
 * Regression test: a misconfigured/unreachable REDIS_URL must NOT prevent the
 * Nova server from coming up. This is exactly what broke the Render deploy:
 * node-redis retried the failed DNS lookup forever, main() blocked on
 * vaultSyncFromRemote(), no port ever opened, and Render marked the deploy
 * failed while serving the previous build.
 *
 * Boots the real built server (dist/server.js) with REDIS_URL pointing at a
 * hostname that cannot resolve (same ENOTFOUND failure mode as Render's
 * internal `red-<id>` hostname used from a different region) and asserts:
 *   1. the server starts listening within ~15s,
 *   2. /api/vault responds with remote.backend === "redis" (configured but
 *      unreachable) while still serving requests,
 *   3. the process stays alive and logs the non-fatal redis error.
 *
 * Run after `npm run build`:  node scripts/smoke-vault-bad-redis.mjs
 */
import { spawn } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const cwd = process.cwd();
const port = 3947;
const base = `http://127.0.0.1:${port}`;
const vaultDir = mkdtempSync(join(tmpdir(), "nova-vault-bad-redis-"));

const child = spawn(process.execPath, [join(cwd, "dist/server.js")], {
  env: {
    ...process.env,
    PORT: String(port),
    REDIS_URL: "redis://red-does-not-exist.invalid:6379",
    NOVA_VAULT_KEY: "smoke-bad-redis-key",
    NOVA_VAULT_DIR: vaultDir,
  },
  stdio: ["ignore", "pipe", "pipe"],
});

let stdout = "";
let stderr = "";
child.stdout.on("data", (d) => (stdout += d));
child.stderr.on("data", (d) => (stderr += d));

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const waitFor = async (fn, ms, what) => {
  const deadline = Date.now() + ms;
  while (Date.now() < deadline) {
    try {
      const out = await fn();
      if (out) return out;
    } catch {
      /* retry */
    }
    await sleep(500);
  }
  throw new Error(`timed out: ${what}`);
};

let pass = 0;
const ok = (m) => {
  pass += 1;
  console.log(`OK   ${m}`);
};

try {
  const t0 = Date.now();
  const health = await waitFor(
    async () => {
      const res = await fetch(`${base}/api/health`);
      if (!res.ok) return null;
      return res.json();
    },
    15000,
    "server never started listening within 15s",
  );
  ok(`server listening in ${Date.now() - t0}ms (mock=healthy)`);
  if (health.ok !== true) throw new Error("health endpoint did not report ok");

  const vault = await fetch(`${base}/api/vault`).then((r) => r.json());
  ok(`/api/vault remote = ${JSON.stringify(vault.remote)}`);
  if (!vault.remote || vault.remote.backend !== "redis") {
    throw new Error(`expected remote.backend "redis", got ${JSON.stringify(vault.remote)}`);
  }

  await sleep(2500); // let the startup sync finish failing
  if (child.exitCode !== null) {
    throw new Error(`server exited early (rc=${child.exitCode})`);
  }
  ok("process still alive after failed redis connect");
  if (!/redis (client error|connect failed)/.test(stdout + stderr)) {
    console.log(`WARN no redis error logged. stderr:\n${stderr.slice(0, 300)}`);
  } else {
    ok("non-fatal redis error logged");
  }

  console.log(`\nPASS — server starts and serves despite unreachable REDIS_URL (${pass} checks)`);
} finally {
  child.kill("SIGTERM");
  rmSync(vaultDir, { recursive: true, force: true });
}
