import { execFile } from "node:child_process";
import { promises as fs } from "node:fs";
import { dirname, resolve } from "node:path";

export type ComputerAction = "type" | "key" | "click" | "move" | "screenshot";

function run(cmd: string, args: string[], timeout = 15_000): Promise<string> {
  return new Promise((resolveP, rejectP) => {
    execFile(cmd, args, { timeout, maxBuffer: 4 * 1024 * 1024 }, (err, stdout, stderr) => {
      if (err) rejectP(new Error((stderr || stdout || err.message).trim() || err.message));
      else resolveP(stdout);
    });
  });
}

async function detectXdotool(): Promise<string | null> {
  try {
    const out = await run("bash", ["-lc", "command -v xdotool"]);
    return out.trim() || null;
  } catch {
    return null;
  }
}

async function screenshotAny(path: string): Promise<string> {
  const attempts: Array<[string, string[]]> = [
    ["scrot", [path]],
    ["import", ["-window", "root", path]],
    ["gnome-screenshot", ["-f", path]],
  ];
  for (const [cmd, args] of attempts) {
    try {
      await run(cmd, args, 20_000);
      return path;
    } catch {
      /* try next */
    }
  }
  throw new Error(
    "no screenshot tool found. Install one:  sudo apt-get install -y scrot  (or imagemagick, or gnome-screenshot)",
  );
}

/**
 * Computer use tool — controls the local desktop (X11) via xdotool:
 * type text, press keys, move/click the mouse, capture screenshots.
 */
export async function computerUse(
  action: ComputerAction,
  args: Record<string, unknown>,
): Promise<string> {
  const str = (v: unknown, fb = ""): string => (typeof v === "string" ? v : fb);
  const num = (v: unknown): number => {
    const n = Number(v);
    return Number.isFinite(n) ? Math.round(n) : 0;
  };

  const xdotool = await detectXdotool();
  if (!xdotool) {
    return (
      "ERROR: xdotool is not installed. Computer use needs it:\n" +
      "  sudo apt-get install -y xdotool\n" +
      "It also requires an X display — export DISPLAY=:0 on a desktop session, " +
      "or run under a virtual display: Xvfb :99 & export DISPLAY=:99"
    );
  }
  if (!process.env.DISPLAY) {
    return "ERROR: no DISPLAY is set. Computer use needs an X server (export DISPLAY=:0, or start Xvfb on headless machines).";
  }

  try {
    switch (action) {
      case "type": {
        const text = str(args.text);
        await run(xdotool, ["type", "--delay", "12", text]);
        return `typed ${text.length} chars`;
      }
      case "key": {
        const key = str(args.key);
        await run(xdotool, ["key", key]);
        return `pressed ${key}`;
      }
      case "click": {
        const x = num(args.x);
        const y = num(args.y);
        await run(xdotool, ["mousemove", String(x), String(y), "click", "1"]);
        return `clicked at (${x}, ${y})`;
      }
      case "move": {
        const x = num(args.x);
        const y = num(args.y);
        await run(xdotool, ["mousemove", String(x), String(y)]);
        return `moved mouse to (${x}, ${y})`;
      }
      case "screenshot": {
        const path = str(args.path, "output/computer-screenshot.png");
        await fs.mkdir(dirname(resolve(path)), { recursive: true });
        const saved = await screenshotAny(path);
        return `screenshot saved to ${saved}`;
      }
      default:
        return `ERROR: unknown computer action "${action}".`;
    }
  } catch (err) {
    return `ERROR: ${(err as Error).message}`;
  }
}
