import type { NovaConfig } from "../config.js";

/**
 * GitHub connector. Calls the GitHub REST API using GITHUB_TOKEN.
 * Paths look like "/repos/owner/repo/issues" (or a full URL). Method defaults
 * to GET; body is sent as JSON for POST/PATCH/PUT/DELETE.
 */
export async function githubApi(
  config: NovaConfig,
  path: string,
  method = "GET",
  body?: Record<string, unknown>,
): Promise<string> {
  if (!config.githubToken) {
    return (
      "ERROR: GITHUB_TOKEN is not set. Add it to .env (or export it) to enable the github tool. " +
      "Create a token at https://github.com/settings/tokens (fine-grained tokens work best)."
    );
  }

  const url = path.startsWith("http")
    ? path
    : `https://api.github.com${path.startsWith("/") ? path : `/${path}`}`;

  try {
    const res = await fetch(url, {
      method,
      headers: {
        Authorization: `Bearer ${config.githubToken}`,
        Accept: "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
        "Content-Type": "application/json",
        "User-Agent": "koola10-nova-agent",
      },
      body: body && method !== "GET" ? JSON.stringify(body) : undefined,
      signal: AbortSignal.timeout(30_000),
    });
    const text = await res.text();
    return formatHttpResult(res.status, text);
  } catch (err) {
    return `ERROR: GitHub request failed — ${(err as Error).message}`;
  }
}

function formatHttpResult(status: number, text: string): string {
  let pretty = text;
  if (text) {
    try {
      pretty = JSON.stringify(JSON.parse(text), null, 2);
    } catch {
      /* keep raw text */
    }
  }
  const body = pretty ? `\n${pretty}` : "";
  return `HTTP ${status}${body}`;
}
