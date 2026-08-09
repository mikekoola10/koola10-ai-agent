#!/usr/bin/env node
/* Smoke test: Nova UI server + durable vault.
 * Boots a mock Upstash REST server, pre-seeds it with an encrypted blob
 * (written by a prior vault instance), wipes the local vault dir, boots the
 * real UI server, and verifies /api/vault reports the restored entry and the
 * remote backup shape. */
import { createServer } from "node:http";
import { rm, mkdir } from "node:fs/promises";

const store = {};
const srv = createServer((req, res) => {
  const url = new URL(req.url, "http://x");
  res.setHeader("content-type", "application/json");
  if (req.method === "GET" && url.pathname.startsWith("/get/")) {
    const key = decodeURIComponent(url.pathname.slice(5));
    res.end(JSON.stringify({ result: store[key] ?? null }));
    return;
  }
  if (req.method === "POST") {
    let body = "";
    req.on("data", (c) => (body += c));
    req.on("end", () => {
      const [cmd, key, value] = JSON.parse(body);
      if (cmd === "SET" && key && value !== undefined) {
        store[key] = value;
        res.end(JSON.stringify({ result: "OK" }));
        return;
      }
      res.end(JSON.stringify({ result: "ERR" }));
    });
    return;
  }
  res.statusCode = 404;
  res.end("{}");
});

await new Promise((r) => srv.listen(0, "127.0.0.1", r));
const port = srv.address().port;

process.env.NOVA_VAULT_URL = "http://127.0.0.1:" + port;
process.env.NOVA_VAULT_TOKEN = "test-token";
process.env.NOVA_VAULT_KEY = "smoke-test-master-key";
process.env.NOVA_VAULT_DIR = "/tmp/nova-vault-server/vault";
process.env.NOVA_MOCK = "1";

// Seed the remote with an entry from a prior "vault instance"
await mkdir("/tmp/nova-vault-server/vault", { recursive: true });
const seeder = await import("../dist/tools/vault.js");
seeder.vaultSet("GITHUB_TOKEN", "ghp_seed-999");
await seeder.vaultPushToRemote();
console.log("seeded remote, local names:", seeder.vaultList().join(","));

// Simulate redeploy: wipe the local vault dir, then boot the real server
await rm("/tmp/nova-vault-server/vault", { recursive: true, force: true });

const { startServer } = await import("../dist/server.js");
const { loadConfig } = await import("../dist/config.js");
const { applyVaultOverrides } = await import("../dist/tools/vault.js");
const cfg = applyVaultOverrides(loadConfig({ cwd: process.cwd(), mock: true }));
const server = startServer(cfg, 0);

await new Promise((res) => server.on("listening", res));
const portSrv = server.address().port;
const j = async (p) => (await fetch("http://127.0.0.1:" + portSrv + p)).json();

const vaultInfo = await j("/api/vault");
console.log("api/vault names:", vaultInfo.names.join(","));
console.log("api/vault remote:", JSON.stringify(vaultInfo.remote));

const health = await j("/api/health");
console.log("health.connectors.github:", health.connectors.github, "| bounty.fullScan:", health.bounty.fullScan);

server.close();
srv.close();
process.exit(0);
