import { promises as fs } from "node:fs";
import { dirname, resolve } from "node:path";
import type { Browser, Page } from "playwright";

export type BrowserAction = "open" | "click" | "type" | "extract" | "screenshot" | "back" | "close";

export interface BrowserOptions {
  headless: boolean;
  maxChars: number;
}

let browserPromise: Promise<Browser> | null = null;
let page: Page | null = null;

function noPlaywrightError(): Error {
  return new Error(
    "the playwright module is not installed. Enable the browser tool with:\n" +
      "  npm install playwright && npx playwright install chromium\n" +
      "(one-time setup; then restart nova).",
  );
}

async function getPage(opts: BrowserOptions): Promise<Page> {
  let pw: typeof import("playwright");
  try {
    pw = await import("playwright");
  } catch {
    throw noPlaywrightError();
  }
  if (page && !page.isClosed()) return page;
  if (!browserPromise) {
    browserPromise = pw.chromium.launch({ headless: opts.headless }).catch((e: unknown) => {
      browserPromise = null;
      const msg = (e as Error).message ?? "";
      if (/executable doesn't exist|playwright install|browser_type/i.test(msg)) {
        throw new Error(
          "Chromium is not downloaded. Run once:  npx playwright install chromium",
        );
      }
      throw e;
    });
  }
  const browser = await browserPromise;
  page = await browser.newPage();
  return page;
}

async function pageText(page: Page): Promise<string> {
  return page.evaluate(() => {
    const el = (globalThis as { document?: { body?: { innerText?: string } } }).document?.body;
    return el?.innerText ?? "";
  });
}

/** Browser use tool — drives a real Chromium via Playwright. */
export async function browserUse(
  action: BrowserAction,
  args: Record<string, unknown>,
  opts: BrowserOptions,
): Promise<string> {
  const str = (v: unknown, fb = ""): string => (typeof v === "string" ? v : fb);
  try {
    const p = await getPage(opts);

    switch (action) {
      case "open": {
        const url = str(args.url);
        if (!url) return "ERROR: browser open requires a url.";
        await p.goto(url, { waitUntil: "domcontentloaded", timeout: 30_000 });
        const title = await p.title();
        const text = (await pageText(p)).trim();
        const shown =
          text.length > opts.maxChars
            ? `${text.slice(0, opts.maxChars)}\n… [truncated: ${text.length} chars total]`
            : text;
        return `opened ${p.url()}\ntitle: ${title}\n\n${shown}`;
      }
      case "click": {
        const sel = str(args.selector);
        if (!sel) return "ERROR: browser click requires a selector.";
        await p.click(sel, { timeout: 10_000 });
        return `clicked "${sel}" → now at ${p.url()}`;
      }
      case "type": {
        const sel = str(args.selector);
        const text = str(args.text);
        if (!sel) return "ERROR: browser type requires a selector.";
        await p.click(sel, { timeout: 10_000 });
        await p.fill(sel, text);
        return `typed ${text.length} chars into "${sel}"`;
      }
      case "extract": {
        const text = (await pageText(p)).trim();
        if (!text) return `(page has no readable text) — current URL: ${p.url()}`;
        return text.length > opts.maxChars
          ? `${text.slice(0, opts.maxChars)}\n… [truncated: ${text.length} chars total]`
          : text;
      }
      case "screenshot": {
        const path = str(args.path, "output/browser-screenshot.png");
        await fs.mkdir(dirname(resolve(path)), { recursive: true });
        await p.screenshot({ path, fullPage: false });
        return `screenshot saved to ${path}`;
      }
      case "back": {
        await p.goBack({ timeout: 15_000 }).catch(() => null);
        return `went back → ${p.url()}`;
      }
      case "close": {
        await p.close().catch(() => undefined);
        if (browserPromise) await (await browserPromise).close().catch(() => undefined);
        page = null;
        browserPromise = null;
        return "browser closed";
      }
      default:
        return `ERROR: unknown browser action "${action}".`;
    }
  } catch (err) {
    return `ERROR: ${(err as Error).message}`;
  }
}
