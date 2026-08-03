import { execFile } from "node:child_process";

export interface ShellOptions {
  cwd?: string;
  timeoutMs?: number;
  maxChars?: number;
}

interface ExecOutcome {
  exitCode: number;
  stdout: string;
  stderr: string;
  elapsedSec: string;
  note?: string;
}

function truncate(s: string, max: number): string {
  if (s.length <= max) return s;
  return `${s.slice(0, max)}\n… [truncated: ${s.length} chars total]`;
}

function format(out: ExecOutcome): string {
  const lines: string[] = [];
  lines.push(`exit code: ${out.exitCode}${out.note ? ` (${out.note})` : ""} · ${out.elapsedSec}s`);
  if (out.stdout) lines.push(`[stdout]\n${truncate(out.stdout, 8000)}`);
  if (out.stderr) lines.push(`[stderr]\n${truncate(out.stderr, 4000)}`);
  if (!out.stdout && !out.stderr) lines.push("(no output)");
  return lines.join("\n");
}

/** Runs `command` via `bash -lc` and returns a summarized result string. */
export async function runCommand(command: string, opts: ShellOptions = {}): Promise<string> {
  const { cwd, timeoutMs = 60_000, maxChars = 8000 } = opts;
  const start = Date.now();
  const elapsed = (): string => ((Date.now() - start) / 1000).toFixed(1);

  try {
    const { stdout, stderr } = await new Promise<{ stdout: string; stderr: string }>(
      (resolve, reject) => {
        execFile(
          "bash",
          ["-lc", command],
          { cwd, timeout: timeoutMs, maxBuffer: 4 * 1024 * 1024 },
          (error, stdout, stderr) => {
            if (error) {
              const err = error as NodeJS.ErrnoException & { killed?: boolean; stdout?: string; stderr?: string };
              reject(
                Object.assign(new Error(err.message), {
                  code: err.code,
                  killed: err.killed,
                  stdout: err.stdout ?? stdout,
                  stderr: err.stderr ?? stderr,
                }),
              );
            } else {
              resolve({ stdout, stderr });
            }
          },
        );
      },
    );
    return truncate(format({ exitCode: 0, stdout, stderr, elapsedSec: elapsed() }), maxChars);
  } catch (err) {
    const e = err as NodeJS.ErrnoException & { killed?: boolean; stdout?: string; stderr?: string };
    const exitCode =
      typeof e.code === "number" ? e.code : e.killed ? 124 : 1;
    const note = e.killed ? "timed out / killed" : undefined;
    return truncate(
      format({
        exitCode,
        stdout: e.stdout ?? "",
        stderr: e.stderr ?? String(e.message),
        elapsedSec: elapsed(),
        note,
      }),
      maxChars,
    );
  }
}
