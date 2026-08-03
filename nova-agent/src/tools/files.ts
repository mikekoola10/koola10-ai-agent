import { promises as fs } from "node:fs";
import { dirname, resolve } from "node:path";

export interface ListOptions {
  maxEntries?: number;
}

/** Lists one directory level with [dir]/[file] markers and sizes. */
export async function listDirectory(path: string, opts: ListOptions = {}): Promise<string> {
  const maxEntries = opts.maxEntries ?? 200;
  try {
    const abs = resolve(path);
    const entries = await fs.readdir(abs, { withFileTypes: true });
    entries.sort((a, b) => {
      if (a.isDirectory() !== b.isDirectory()) return a.isDirectory() ? -1 : 1;
      return a.name.localeCompare(b.name);
    });

    const lines: string[] = [];
    for (const en of entries.slice(0, maxEntries)) {
      if (en.isDirectory()) {
        lines.push(`[dir]  ${en.name}/`);
      } else {
        let size = "";
        try {
          const st = await fs.stat(resolve(abs, en.name));
          size = ` (${st.size.toLocaleString()} B)`;
        } catch {
          /* ignore stat failures */
        }
        lines.push(`[file] ${en.name}${size}`);
      }
    }

    const hidden = entries.length - lines.length;
    if (hidden > 0) lines.push(`… and ${hidden} more entries (showing ${maxEntries})`);
    if (lines.length === 0) lines.push("(empty directory)");
    return lines.join("\n");
  } catch (err) {
    return `ERROR: ${(err as Error).message}`;
  }
}

/** Reads a file as UTF-8, truncated to maxChars. */
export async function readFile(path: string, maxChars = 8000): Promise<string> {
  try {
    const abs = resolve(path);
    const data = await fs.readFile(abs, "utf8");
    if (data.length <= maxChars) return data;
    return `${data.slice(0, maxChars)}\n… [truncated: ${data.length} chars total, showing first ${maxChars}]`;
  } catch (err) {
    return `ERROR: ${(err as Error).message}`;
  }
}

/** Writes a file (creating parent directories). */
export async function writeFile(path: string, content: string): Promise<string> {
  try {
    const abs = resolve(path);
    await fs.mkdir(dirname(abs), { recursive: true });
    await fs.writeFile(abs, content, "utf8");
    return `Wrote ${content.length} chars to ${path}`;
  } catch (err) {
    return `ERROR: ${(err as Error).message}`;
  }
}
