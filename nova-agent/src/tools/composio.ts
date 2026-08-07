import type { NovaConfig } from "../config.js";

let composioPromise: Promise<any> | null = null;

function formatResult(result: unknown, maxChars: number): string {
  const serialized = typeof result === "string" ? result : JSON.stringify(result, null, 2);
  if (!serialized) return "(no result)";
  return serialized.length <= maxChars
    ? serialized
    : `${serialized.slice(0, maxChars)}\n… [truncated: ${serialized.length} chars total]`;
}

async function getComposio(apiKey: string): Promise<any> {
  if (!composioPromise) {
    composioPromise = import("@composio/core")
      .then(({ Composio }) => new Composio({ apiKey }))
      .catch((error: unknown) => {
        composioPromise = null;
        throw new Error(`Composio SDK could not load: ${error instanceof Error ? error.message : String(error)}`);
      });
  }
  return composioPromise;
}

/**
 * Live authentication check for Composio: creates a session for the configured
 * user and lists available tools. This is what powers `nova --verify-connectors`
 * and the `ping` tool action.
 */
export async function checkComposio(config: NovaConfig): Promise<{
  provider: "composio";
  configured: boolean;
  ok: boolean;
  detail: string;
}> {
  if (!config.composioApiKey) {
    return { provider: "composio", configured: false, ok: false, detail: "COMPOSIO_API_KEY is not set" };
  }
  try {
    const composio = await getComposio(config.composioApiKey);
    const session = await composio.create(config.composioUserId);
    const tools = await session.tools();
    const count = Array.isArray(tools)
      ? tools.length
      : typeof tools === "object" && tools !== null
        ? Object.keys(tools).length
        : "n/a";
    return {
      provider: "composio",
      configured: true,
      ok: true,
      detail: `authenticated as "${config.composioUserId}" — ${count} tools available`,
    };
  } catch (error) {
    return {
      provider: "composio",
      configured: true,
      ok: false,
      detail: `authentication or request failed — ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}

export async function composioTool(
  config: NovaConfig,
  action: string,
  args: Record<string, unknown> = {},
): Promise<string> {
  if (!config.composioApiKey) {
    return "ERROR: COMPOSIO_API_KEY is not set. Add it to the Nova environment to enable the composio tool.";
  }

  try {
    const composio = await getComposio(config.composioApiKey);
    const userId = typeof args.userId === "string" && args.userId.trim()
      ? args.userId.trim()
      : config.composioUserId;
    const toolkit = typeof args.toolkit === "string" && args.toolkit.trim() ? args.toolkit.trim() : undefined;
    const session = await composio.create(userId, toolkit ? { toolkits: [toolkit] } : undefined);

    if (action === "ping") {
      return formatResult(await checkComposio(config), config.maxToolOutputChars);
    }

    if (action === "list") {
      return formatResult(await session.tools(), config.maxToolOutputChars);
    }

    if (action === "execute") {
      const tool = typeof args.tool === "string" ? args.tool.trim() : "";
      if (!tool) return "ERROR: composio execute requires a tool name.";
      const toolArgs = args.arguments && typeof args.arguments === "object" ? args.arguments : {};
      return formatResult(await composio.tools.execute(tool, { userId, arguments: toolArgs }), config.maxToolOutputChars);
    }

    return `ERROR: unknown composio action "${action}". Use "list", "execute", or "ping".`;
  } catch (error) {
    return `ERROR: Composio request failed — ${error instanceof Error ? error.message : String(error)}`;
  }
}
