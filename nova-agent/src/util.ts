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

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  const s = ms / 1000;
  if (s < 60) return `${s.toFixed(1)}s`;
  const m = Math.floor(s / 60);
  const rest = Math.round(s % 60);
  return `${m}m ${rest}s`;
}
