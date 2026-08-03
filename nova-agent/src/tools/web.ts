const UA =
  "Mozilla/5.0 (compatible; NovaAgent/0.1; +https://github.com/mikekoola10/koola10-nova-agent)";

function decodeEntities(s: string): string {
  return s
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&#x27;/g, "'")
    .replace(/&#x2F;/g, "/")
    .replace(/&amp;/g, "&")
    .replace(/&nbsp;/g, " ");
}

function stripTags(s: string): string {
  return s.replace(/<[^>]+>/g, "").trim();
}

function htmlToText(html: string): string {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, " ")
    .replace(/<style[\s\S]*?<\/style>/gi, " ")
    .replace(/<noscript[\s\S]*?<\/noscript>/gi, " ")
    .replace(/<svg[\s\S]*?<\/svg>/gi, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

interface SearchResult {
  title: string;
  url: string;
  snippet: string;
}

/** DuckDuckGo HTML search (keyless). Falls back to the lite endpoint. */
export async function webSearch(query: string, maxResults = 6): Promise<string> {
  const enc = encodeURIComponent(query);

  const parseHtml = (html: string): SearchResult[] => {
    const out: SearchResult[] = [];
    const blocks = html.split(/<div class="result[^"]*"[^>]*>/g).slice(1);
    for (const block of blocks) {
      const titleM = block.match(/class="result__a"[^>]*>(.*?)<\/a>/s);
      const urlM = block.match(/class="result__a"[^>]*href="([^"]+)"/);
      const snipM = block.match(/class="result__snippet"[^>]*>(.*?)<\/a>/s);
      if (!titleM || !urlM) continue;
      const href = decodeEntities(urlM[1]!);
      out.push({
        title: decodeEntities(stripTags(titleM[1]!)),
        url: href.startsWith("//") ? `https:${href}` : href,
        snippet: snipM ? decodeEntities(stripTags(snipM[1]!)) : "",
      });
      if (out.length >= maxResults) break;
    }
    return out;
  };

  const parseLite = (html: string): SearchResult[] => {
    const out: SearchResult[] = [];
    const blocks = html.split(/<table class="result">/g).slice(1);
    for (const block of blocks) {
      const linkM = block.match(/<a[^>]*href="([^"]+)"[^>]*>(.*?)<\/a>/s);
      const snipM = block.match(/class="result-snippet">(.*?)<\/td>/s);
      if (!linkM) continue;
      out.push({
        title: decodeEntities(stripTags(linkM[2]!)),
        url: decodeEntities(linkM[1]!),
        snippet: snipM ? decodeEntities(stripTags(snipM[1]!)) : "",
      });
      if (out.length >= maxResults) break;
    }
    return out;
  };

  const fetchHtml = async (url: string): Promise<string> => {
    const res = await fetch(url, {
      headers: { "User-Agent": UA },
      redirect: "follow",
      signal: AbortSignal.timeout(15_000),
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.text();
  };

  try {
    const html = await fetchHtml(`https://html.duckduckgo.com/html/?q=${enc}`);
    let results = parseHtml(html);
    if (results.length === 0) {
      // Fall back to the lite endpoint.
      const lite = await fetchHtml(`https://lite.duckduckgo.com/lite/?q=${enc}`);
      results = parseLite(lite);
    }
    if (results.length === 0) return "No results found.";
    return results
      .map((r, i) => `${i + 1}. ${r.title}\n   ${r.url}\n   ${r.snippet}`)
      .join("\n");
  } catch (err) {
    return `ERROR: web search failed — ${(err as Error).message}`;
  }
}

/** Fetches a URL and returns readable text, truncated. */
export async function fetchUrl(url: string, maxChars = 8000): Promise<string> {
  try {
    const res = await fetch(url, {
      headers: { "User-Agent": UA, Accept: "text/html,application/xhtml+xml" },
      redirect: "follow",
      signal: AbortSignal.timeout(20_000),
    });
    if (!res.ok) return `ERROR: HTTP ${res.status} for ${url}`;
    const text = htmlToText(await res.text());
    if (!text) return "Page returned no readable text.";
    if (text.length <= maxChars) return text;
    return `${text.slice(0, maxChars)}\n… [truncated: ${text.length} chars total]`;
  } catch (err) {
    return `ERROR: ${(err as Error).message}`;
  }
}
