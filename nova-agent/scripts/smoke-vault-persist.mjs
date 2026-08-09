#!/usr/bin/env node
/* Smoke test: vault durable remote backup round-trip.
 * Boots a mock Upstash-compatible REST server, writes a key, pushes it,
 * wipes the local vault (simulating a Render redeploy), then restores in a
 * brand-new Node process to prove the fresh-boot path really works. */
import { createServer } from "node:http";
import { rm } from "node:fs/promises";
import { spawn } from "node:child_process";

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
console.log("mock upstash on port", port);

process.env.NOVA_VAULT_URL = "http://127.0.0.1:" + port;
process.env.NOVA_VAULT_TOKEN = "test-token";
process.env.NOVA_VAULT_KEY = "smoke-test-master-key";
process.env.NOVA_VAULT_DIR = "/tmp/nova-vault-persist/vault";

const vault = await import("../dist/tools/vault.js");

// 1) Write a key, flush to remote
vault.vaultSet("GITHUB_TOKEN", "ghp_dummy-123");
const push = await vault.vaultPushToRemote();
console.log("push.ok =", push.ok);
console.log("local names after set:", vault.vaultList().join(","));
console.log("vaultInfo.remote:", JSON.stringify(vault.vaultInfo().remote));

// 2) Simulate a redeploy: wipe the local vault dir entirely
await rm("/tmp/nova-vault-persist/vault", { recursive: true, force: true });  // 3) Restore in a brand-new process (no in-memory cache) from the remote.
  //    Async spawn so the mock server keeps answering the child's requests.
  const child = spawn(
    process.execPath,
    [
      "--input-type=module",
      "-e",
      `import { vaultSyncFromRemote, vaultList, vaultGet } from "./dist/tools/vault.js";
     const r = await vaultSyncFromRemote();
     console.log("fresh restore.pulled =", r.pulled);
     console.log("fresh names:", vaultList().join(","));
     console.log("fresh value intact:", vaultGet("GITHUB_TOKEN"));`,
    ],
    { cwd: process.cwd(), env: process.env },
  );
  let childOut = "",
    childErr = "";
  child.stdout.on("data", (c) => (childOut += c));
  child.stderr.on("data", (c) => (childErr += c));
  await new Promise((res, rej) => {
    child.on("close", (code) => (code === 0 ? res() : rej(new Error("child exited " + code))));
    child.on("error", rej);
  });
  console.log(childOut.trim());
  if (childErr.trim()) console.error("child stderr:", childErr.trim());

// 4) No remote configured -> local-only path still works
process.env.NOVA_VAULT_URL = "";
process.env.NOVA_VAULT_TOKEN = "";
const local = await import("../dist/tools/vault.js");
console.log("remote disabled, info:", JSON.stringify(local.vaultInfo().remote));

srv.close();
process.exit(0);
