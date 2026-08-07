// Diagnostic probe + Nova server bootstrap.
// Prints presence of connector env vars (never values), then starts the
// Nova UI server so the preview stays functional. Output lands in preview logs.
import { spawn } from "node:child_process";

const keys = [
  "HUGGINGFACE_TOKEN",
  "COMPOSIO_API_KEY",
  "ZAPIER_WEBHOOK_URL",
  "N8N_WEBHOOK_URL",
  "MAKE_WEBHOOK_URL",
  "DEEPSEEK_API_KEY",
  "GITHUB_TOKEN",
  "STRIPE_SECRET_KEY",
  "HF_TOKEN",
];
console.log("=== ENV PROBE (presence only) ===");
for (const k of keys) {
  console.log(`${k} => ${process.env[k] ? "SET" : "missing"}`);
}
console.log("=== starting Nova server ===");
// dist/server.js only runs main() when it is the direct entry point, so spawn
// it as a child process. It reads $PORT (injected by Freebuff) and binds 0.0.0.0.
const child = spawn(process.execPath, ["dist/server.js"], {
  cwd: process.cwd(),
  env: process.env,
  stdio: "inherit",
});
child.on("exit", (code) => process.exit(code ?? 0));
