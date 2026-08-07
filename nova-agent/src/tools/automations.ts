import type { NovaConfig } from "../config.js";

type AutomationProvider = "zapier" | "n8n" | "make";

type AutomationConfig = NovaConfig & {
  zapierWebhookUrl: string;
  n8nWebhookUrl: string;
  makeWebhookUrl: string;
};

export type AutomationCheck = {
  provider: AutomationProvider;
  configured: boolean;
  ok: boolean;
  status?: number;
  detail: string;
};

function formatResult(value: unknown, maxChars: number): string {
  const text = typeof value === "string" ? value : JSON.stringify(value, null, 2);
  if (!text) return "(empty response)";
  return text.length <= maxChars ? text : `${text.slice(0, maxChars)}\n… [truncated]`;
}

function configuredUrl(config: AutomationConfig, provider: AutomationProvider): string {
  return provider === "zapier"
    ? config.zapierWebhookUrl
    : provider === "n8n"
      ? config.n8nWebhookUrl
      : config.makeWebhookUrl;
}

function envName(provider: AutomationProvider): string {
  return provider === "zapier" ? "ZAPIER_WEBHOOK_URL" : provider === "n8n" ? "N8N_WEBHOOK_URL" : "MAKE_WEBHOOK_URL";
}

function headersFor(config: AutomationConfig, provider: AutomationProvider): Record<string, string> {
  const headers: Record<string, string> = { "content-type": "application/json" };
  if (provider === "n8n" && config.n8nApiKey) headers["x-n8n-api-key"] = config.n8nApiKey;
  return headers;
}

function pingPayload(): Record<string, unknown> {
  return { event: "nova_connector_ping", source: "nova-agent", ts: new Date().toISOString() };
}

/**
 * Live connectivity check for one automation provider. Sends a harmless ping
 * payload to the configured webhook and reports a structured result. This is
 * what powers `nova --verify-connectors` and the `ping` tool action.
 */
export async function checkAutomation(config: AutomationConfig, provider: AutomationProvider): Promise<AutomationCheck> {
  const url = configuredUrl(config, provider);
  if (!url) return { provider, configured: false, ok: false, detail: `${envName(provider)} is not set` };
  if (!/^https:\/\//i.test(url)) return { provider, configured: true, ok: false, detail: `${envName(provider)} must be an https:// URL` };
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: headersFor(config, provider),
      body: JSON.stringify(pingPayload()),
      signal: AbortSignal.timeout(config.toolTimeoutMs),
    });
    const text = (await response.text()).slice(0, 200);
    const hint = text ? ` — ${text}` : "";
    if (response.status === 401 || response.status === 403) {
      return {
        provider,
        configured: true,
        ok: false,
        status: response.status,
        detail: `authentication failed (${response.status}${hint}). Check the webhook auth settings / ${provider === "n8n" ? "N8N_API_KEY" : "access credentials"}.`,
      };
    }
    if (response.status === 404) {
      return {
        provider,
        configured: true,
        ok: false,
        status: 404,
        detail: `endpoint not found (404). The ${provider === "make" ? "scenario may be paused or not yet activated" : "workflow/zap may be paused"} or the URL is wrong.`,
      };
    }
    return { provider, configured: true, ok: response.ok, status: response.status, detail: `HTTP ${response.status}${hint}` };
  } catch (error) {
    return { provider, configured: true, ok: false, detail: `request failed — ${error instanceof Error ? error.message : String(error)}` };
  }
}

export async function automationTool(
  config: AutomationConfig,
  provider: AutomationProvider,
  action: string,
  args: Record<string, unknown> = {},
): Promise<string> {
  const url = configuredUrl(config, provider);
  if (!url) return `ERROR: ${envName(provider)} is not set. Configure the provider webhook URL in the server environment.`;
  if (!/^https:\/\//i.test(url)) return `ERROR: ${envName(provider)} must use an https:// URL.`;
  if (action !== "trigger" && action !== "ping")
    return `ERROR: ${provider} supports only the trigger and ping actions.`;

  // ping: harmless connectivity test (no real workflow side effects beyond the ping event)
  if (action === "ping") {
    return formatResult(await checkAutomation(config, provider), config.maxToolOutputChars);
  }

  const payload = args.payload && typeof args.payload === "object" ? args.payload : {};
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: headersFor(config, provider),
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(config.toolTimeoutMs),
    });
    const text = await response.text();
    let body: unknown = text;
    try {
      body = text ? JSON.parse(text) : "";
    } catch {
      /* plain-text webhook response */
    }
    return formatResult({ provider, action: "trigger", status: response.status, ok: response.ok, response: body }, config.maxToolOutputChars);
  } catch (error) {
    return `ERROR: ${provider} webhook request failed — ${error instanceof Error ? error.message : String(error)}`;
  }
}

/* ---------------- Daily report (Nova-generated email) ---------------- */

const DELIVERY_PROVIDERS = ["zapier", "n8n", "make"] as const;
export type DeliveryProvider = (typeof DELIVERY_PROVIDERS)[number];

/**
 * Choose where the daily report gets delivered. NOVA_REPORT_PROVIDER forces
 * one; otherwise Nova uses the first configured automation webhook
 * (Zapier → Make → n8n). Returns null when nothing is configured.
 */
export function reportDeliveryProvider(config: AutomationConfig): DeliveryProvider | null {
  const explicit = (process.env.NOVA_REPORT_PROVIDER ?? "").toLowerCase().trim() as DeliveryProvider;
  if ((DELIVERY_PROVIDERS as readonly string[]).includes(explicit)) return explicit;
  if (config.zapierWebhookUrl) return "zapier";
  if (config.makeWebhookUrl) return "make";
  if (config.n8nWebhookUrl) return "n8n";
  return null;
}

export type DailyReportInput = {
  checks: Array<{ provider: string; configured: boolean; ok: boolean; detail?: string }>;
  version?: string;
  taskStats?: { last24h?: number; done?: number; failed?: number; running?: number; total?: number };
};

export type DailyReport = {
  subject: string;
  body: string;
  payload: Record<string, unknown>;
};

/**
 * Build the daily report content. The `payload` is what gets POSTed to the
 * delivery webhook — the email action should map Subject → `subject` and
 * Body → `body`.
 */
export function buildDailyReport(input: DailyReportInput): DailyReport {
  const now = new Date();
  const dateLabel = now.toLocaleDateString(undefined, { weekday: "long", year: "numeric", month: "long", day: "numeric" });
  const configured = input.checks.filter((c) => c.configured);
  const okCount = configured.filter((c) => c.ok).length;
  const failCount = configured.length - okCount;
  const offCount = input.checks.length - configured.length;
  const lines = [
    `Nova daily report — ${dateLabel}`,
    `Version ${input.version ?? "0.4.0"} · generated ${now.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" })}`,
    "",
    `Connectors: ${okCount}/${configured.length} OK${failCount > 0 ? ` · ${failCount} FAILING` : ""}${offCount > 0 ? ` · ${offCount} not configured` : ""}`,
    ...input.checks.map((c) => `  ${c.ok ? "✔" : "✖"} ${c.provider.padEnd(10)} ${c.configured ? (c.ok ? "OK" : `FAIL — ${c.detail ?? ""}`) : "not configured"}`),
  ];
  if (input.taskStats) {
    const s = input.taskStats;
    lines.push("", `Tasks (last 24h): ${s.last24h ?? 0} run · ${s.done ?? 0} done · ${s.failed ?? 0} failed`);
  }
  const body = lines.join("\n");
  const subject = `Nova Daily Report — ${dateLabel}`;
  return {
    subject,
    body,
    payload: {
      event: "nova_daily_report",
      subject,
      body,
      generatedAt: now.toISOString(),
      version: input.version ?? "0.4.0",
      connectors: { ok: okCount, failing: failCount, notConfigured: offCount, total: input.checks.length },
      tasks: input.taskStats ?? null,
    },
  };
}
