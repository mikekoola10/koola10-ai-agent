const noColor =
  process.env.NO_COLOR !== undefined ||
  process.env.TERM === "dumb" ||
  !process.stdout.isTTY;

function wrap(code: string) {
  return (s: string): string => (noColor ? s : `\u001b[${code}m${s}\u001b[0m`);
}

export const color = {
  dim: wrap("2"),
  cyan: wrap("36"),
  green: wrap("32"),
  yellow: wrap("33"),
  red: wrap("31"),
  magenta: wrap("35"),
  bold: wrap("1"),
};

export function firstLine(s: string, max = 140): string {
  const line = s.split(/\r?\n/)[0] ?? "";
  return line.length > max ? `${line.slice(0, max)}…` : line;
}

/** Mask secret-looking values so they never land in task previews/logs. */
export function redactSecrets(s: string): string {
  return s
    .replace(/\bgithub_pat_[A-Za-z0-9_]{20,}\b/g, "***REDACTED***")
    .replace(/\bgh[pousr]_[A-Za-z0-9]{20,}\b/g, "***REDACTED***")
    .replace(/\bsk-[A-Za-z0-9]{16,}\b/g, "sk-***REDACTED***")
    .replace(/rediss?:\/\/[^@\s/]+@/g, (m) => m.replace(/:\/\/[^@]+@/, "://***REDACTED***@"))
    .replace(/\bBearer\s+[A-Za-z0-9._-]{12,}\b/gi, "Bearer ***REDACTED***");
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  const s = ms / 1000;
  if (s < 60) return `${s.toFixed(1)}s`;
  const m = Math.floor(s / 60);
  const rest = Math.round(s % 60);
  return `${m}m ${rest}s`;
}
