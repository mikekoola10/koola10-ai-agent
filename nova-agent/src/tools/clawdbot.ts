import { execFile } from "node:child_process";
import type { NovaConfig } from "../config.js";

export type ClawdbotAction = "send" | "agent";

/**
 * Clawdbot (OpenClaw) connector.
 *
 * OpenClaw is the open-source personal AI agent formerly known as Clawdbot
 * (https://openclaw.ai). It runs a local gateway that bridges messaging apps
 * (WhatsApp, Telegram, Discord, Slack, Signal, iMessage, Teams) to an AI brain.
 *
 * - action "send":  `openclaw message send --to <contact> --message <text>`
 *   — Nova asks Clawdbot to deliver a message to a chat contact.
 * - action "agent": `openclaw agent --message <text>`
 *   — Nova dispatches a task to Clawdbot's own agent loop (agent-to-agent).
 */
export async function clawdbot(
  config: NovaConfig,
  action: ClawdbotAction,
  contact: string,
  message: string,
): Promise<string> {
  const cli = config.clawdbotCli;
  const args =
    action === "send"
      ? ["message", "send", "--to", contact, "--message", message]
      : ["agent", "--message", message];

  return new Promise<string>((resolve) => {
    execFile(cli, args, { timeout: 90_000, maxBuffer: 4 * 1024 * 1024 }, (error, stdout, stderr) => {
      if (error) {
        const e = error as NodeJS.ErrnoException;
        if (e.code === "ENOENT") {
          resolve(
            `ERROR: the "${cli}" CLI was not found on PATH. OpenClaw (formerly Clawdbot) powers this tool.\n` +
              `Install it with:  npm install -g openclaw\n` +
              `Then run onboarding once:  openclaw onboard\n` +
              `See https://openclaw.ai for setup, or set CLAWDBOT_CLI to your binary path.`,
          );
          return;
        }
        resolve(
          `ERROR: openclaw ${action} failed (${e.message}).\n` +
            `stdout: ${stdout || "(empty)"}\nstderr: ${stderr || "(empty)"}`,
        );
        return;
      }
      resolve(stdout.trim() || "(no output)");
    });
  });
}
