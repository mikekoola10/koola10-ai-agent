import type { NovaConfig } from "../config.js";

/**
 * Stripe connector. Calls the Stripe REST API using STRIPE_SECRET_KEY.
 * Paths look like "/balance", "/customers", "/subscriptions", "/charges",
 * "/payment_intents" (the /v1 prefix is added automatically). POST/PATCH bodies
 * are form-encoded. Note: this tool can also create charges/refunds — the agent
 * must be careful with write operations in live mode.
 */
export async function stripeApi(
  config: NovaConfig,
  path: string,
  method = "GET",
  body?: Record<string, unknown>,
): Promise<string> {
  if (!config.stripeKey) {
    return (
      "ERROR: STRIPE_SECRET_KEY is not set. Add it to .env (or export it) to enable the stripe tool. " +
      "Test keys start with sk_test_ — never use sk_live_ unless you intend real charges."
    );
  }

  const apiPath = path.startsWith("/v1") ? path : path.startsWith("/") ? `/v1${path}` : `/v1/${path}`;
  const url = `https://api.stripe.com${apiPath}`;

  const form = new URLSearchParams();
  if (body) {
    for (const [k, v] of Object.entries(body)) {
      if (v == null) continue;
      if (typeof v === "object") form.append(k, JSON.stringify(v));
      else form.append(k, String(v));
    }
  }

  try {
    const res = await fetch(url, {
      method,
      headers: {
        Authorization: `Bearer ${config.stripeKey}`,
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: method !== "GET" && method !== "HEAD" ? form.toString() : undefined,
      signal: AbortSignal.timeout(30_000),
    });
    const text = await res.text();
    let pretty = text;
    if (text) {
      try {
        pretty = JSON.stringify(JSON.parse(text), null, 2);
      } catch {
        /* keep raw */
      }
    }
    return `HTTP ${res.status}${pretty ? `\n${pretty}` : ""}`;
  } catch (err) {
    return `ERROR: Stripe request failed — ${(err as Error).message}`;
  }
}
