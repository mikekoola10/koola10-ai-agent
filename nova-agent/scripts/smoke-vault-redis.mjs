#!/usr/bin/env node
/* Smoke test: vault durable remote backup over a Redis connection string
 * (REDIS_URL / NOVA_REDIS_URL). Boots a minimal RESP3 Redis-protocol mock
 * (node-redis negotiates HELLO/RESP3 on connect), writes a key, pushes it,
 * wipes the local vault (simulating a Render redeploy), then restores in a
 * brand-new Node process to prove the fresh-boot path really works. */
import { createServer } from "node:net";
import { rm } from "node:fs/promises";
import { spawn } from "node:child_process";

const store = new Map();
const sockets = new Set();

/* --- Minimal RESP3 server (just enough for node-redis handshake + GET/SET) */
function encodeRESP(value) {
  if (value === null) return "_\r\n"; // RESP3 null
  if (typeof value === "string") {
    const b = Buffer.from(value, "utf8");
    return `$${b.length}\r\n${value}\r\n`;
  }
  if (typeof value === "number") return `:${value}\r\n`;
  return `+OK\r\n`;
}

function reply(sock, payload) {
  sock.write(payload);
}

const HELLO_MAP =
  "%7\r\n" +
  "+server\r\n+redis\r\n" +
  "+version\r\n+7.4.0\r\n" +
  "+proto\r\n:3\r\n" +
  "+id\r\n:1\r\n" +
  "+mode\r\n+standalone\r\n" +
  "+role\r\n+master\r\n" +
  "+modules\r\n*0\r\n";

/* Parses a RESP array of bulk strings from a buffer, returning [args, rest] */
function parseArray(buf) {
  if (buf[0] !== 42) return null; // '*'
  let idx = 1;
  let nl = buf.indexOf("\r\n", idx);
  if (nl === -1) return null;
  const count = parseInt(buf.subarray(idx, nl).toString(), 10);
  idx = nl + 2;
  const args = [];
  for (let i = 0; i < count; i++) {
    if (buf[idx] !== 36) return null; // '$'
    nl = buf.indexOf("\r\n", idx);
    if (nl === -1) return null;
    const len = parseInt(buf.subarray(idx + 1, nl).toString(), 10);
    idx = nl + 2;
    if (buf.length < idx + len + 2) return null;
    args.push(buf.subarray(idx, idx + len).toString("utf8"));
    idx += len + 2;
  }
  return [args, buf.subarray(idx)];
}

const srv = createServer((sock) => {
  sockets.add(sock);
  let buf = Buffer.alloc(0);
  sock.on("data", (chunk) => {
    buf = Buffer.concat([buf, chunk]);
    let parsed;
    while ((parsed = parseArray(buf))) {
      const [args, rest] = parsed;
      buf = rest;
      const cmd = (args[0] || "").toUpperCase();
      if (cmd === "HELLO") {
        reply(sock, HELLO_MAP);
      } else if (cmd === "PING") {
        reply(sock, "+PONG\r\n");
      } else if (cmd === "CLIENT") {
        reply(sock, "+OK\r\n");
      } else if (cmd === "SET") {
        store.set(args[1], args[2] ?? "");
        reply(sock, "+OK\r\n");
      } else if (cmd === "GET") {
        const v = store.has(args[1]) ? store.get(args[1]) : null;
        reply(sock, encodeRESP(v));
      } else if (cmd === "QUIT") {
        reply(sock, "+OK\r\n");
        sock.end();
      } else {
        reply(sock, "-ERR unsupported mock command: " + cmd + "\r\n");
      }
    }
  });
  sock.on("error", () => {});
  sock.on("close", () => sockets.delete(sock));
});

await new Promise((r) => srv.listen(0, "127.0.0.1", r));
const port = srv.address().port;
console.log("mock redis on 127.0.0.1:" + port);

process.env.NOVA_REDIS_URL = `redis://127.0.0.1:${port}`;
process.env.NOVA_VAULT_KEY = "smoke-test-master-key";
process.env.NOVA_VAULT_DIR = "/tmp/nova-vault-redis/vault";

const vault = await import("../dist/tools/vault.js");

// 1) Write a key, flush to remote via the Redis backend
console.log("remote.backend =", vault.vaultInfo().remote.backend);
console.log("remote.enabled =", vault.vaultInfo().remote.enabled);
console.log("remote.host =", vault.vaultInfo().remote.host);
vault.vaultSet("GITHUB_TOKEN", "ghp_dummy-redis-1");
const push = await vault.vaultPushToRemote();
console.log("push.ok =", push.ok, push.error ?? "");
console.log("local names after set:", vault.vaultList().join(","));

// 2) Simulate a redeploy: wipe the local vault dir entirely
await rm("/tmp/nova-vault-redis/vault", { recursive: true, force: true });

// 3) Restore in a brand-new process (no in-memory cache) from the Redis mock.
const child = spawn(
  process.execPath,
  [
    "--input-type=module",
    "-e",
    `import { vaultSyncFromRemote, vaultList, vaultGet, vaultInfo } from "./dist/tools/vault.js";
     const r = await vaultSyncFromRemote();
     console.log("fresh backend:", vaultInfo().remote.backend);
     console.log("fresh restore.pulled =", r.pulled);
     console.log("fresh names:", vaultList().join(","));
     console.log("fresh value intact:", vaultGet("GITHUB_TOKEN"));
     process.exit(0);`,
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
process.env.NOVA_REDIS_URL = "";
const local = await import("../dist/tools/vault.js");
console.log("remote disabled, info:", JSON.stringify(local.vaultInfo().remote));

for (const s of sockets) s.destroy();
srv.close();
process.exit(0);
