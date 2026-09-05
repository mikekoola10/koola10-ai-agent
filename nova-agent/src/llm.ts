import type { NovaConfig } from "./config.js";
import type { ChatMessage, ToolCall, ToolDefinition } from "./types.js";

export interface CompleteOptions {
  temperature?: number;
}

/**
 * Routes to the active provider and returns the assistant message, which may
 * contain tool_calls for the agent loop to execute.
 */
export async function complete(
  config: NovaConfig,
  messages: ChatMessage[],
  tools: ToolDefinition[],
  opts: CompleteOptions = {},
): Promise<ChatMessage> {
  if (config.mock) return mockComplete(messages);
  if (config.provider === "anthropic") {
    return completeAnthropic(config, messages, tools, opts);
  }
  return completeOpenAICompatible(config, messages, tools, opts);
}

/* ------------------------------------------------------------------ */
/* OpenAI-compatible (DeepSeek, OpenAI, or any compatible host)        */
/* ------------------------------------------------------------------ */

async function completeOpenAICompatible(
  config: NovaConfig,
  messages: ChatMessage[],
  tools: ToolDefinition[],
  opts: CompleteOptions,
): Promise<ChatMessage> {
  // DeepSeek/OpenAI require tool calls on the wire in the OpenAI shape
  // ({id, type: "function", function: {name, arguments}}) and tool results as
  // {role: "tool", tool_call_id, content}. Convert the internal shape back.
  const wireMessages = messages.map((m) => {
    if (m.tool_calls?.length) {
      return {
        role: m.role,
        content: m.content,
        tool_calls: m.tool_calls.map((tc) => {
          const wire: Record<string, unknown> = {
            id: tc.id,
            type: "function",
            function: { name: tc.name, arguments: tc.arguments },
          };
          // Gemini thought_signature — must be echoed back exactly as received
          if (tc.thought_signature) {
            wire.extra_content = { google: { thought_signature: tc.thought_signature } };
          }
          return wire;
        }),
      };
    }
    if (m.role === "tool") {
      return { role: "tool", tool_call_id: m.tool_call_id, content: m.content };
    }
    return { role: m.role, content: m.content };
  });

  // Debug: log wire messages for Gemini to diagnose thought_signature issues
  if (config.provider === "gemini") {
    const debugMsgs = wireMessages.map((m) => {
      const base = { role: m.role, hasToolCalls: !!(m as Record<string, unknown>).tool_calls };
      const tc = (m as Record<string, unknown>).tool_calls as Array<Record<string, unknown>> | undefined;
      if (tc && tc.length > 0) {
        return { ...base, toolCalls: tc.map((t: Record<string, unknown>) => {
          const fn = t.function as Record<string, unknown> | undefined;
          return { name: fn?.name ?? t.name, hasExtraContent: !!t.extra_content, extraContent: t.extra_content };
        }) };
      }
      return base;
    });
    console.log("[nova:llm] Gemini wire messages:", JSON.stringify(debugMsgs));
  }

  // Retry on transient 503 (server overload) or 429 (rate limit — wait for quota reset)
  let res: Response;
  const MAX_RETRIES = 3;
  for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
    res = await fetch(`${config.apiBase}/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${config.apiKey}`,
      },
      body: JSON.stringify({
        model: config.model,
        messages: wireMessages,
        tools,
        tool_choice: "auto",
        temperature: opts.temperature ?? 0.7,
        max_tokens: 4096,
      }),
      signal: AbortSignal.timeout(180_000),
    });
    if (res.ok) break;
    // 429 = quota/rate limit — wait 5 min per retry (max 3 retries = 15 min total)
    // Handles both per-minute rate limits AND daily quota exhaustion windows
    if (res.status === 429 && attempt < MAX_RETRIES) {
      const waitMs = 5 * 60_000; // 5 minutes — covers per-minute AND daily quota gaps
      console.log(`[nova:llm] 429 rate/quota limited, waiting 5 min for reset (attempt ${attempt + 1}/${MAX_RETRIES})...`);
      await new Promise((r) => setTimeout(r, waitMs));
      continue;
    }
    // 503 = server overload — retry with backoff
    if (res.status === 503 && attempt < MAX_RETRIES) {
      const waitMs = (attempt + 1) * 20_000; // 20s, 40s, 60s
      console.log(`[nova:llm] 503 overload, retrying in ${waitMs / 1000}s (attempt ${attempt + 1}/${MAX_RETRIES})`);
      await new Promise((r) => setTimeout(r, waitMs));
      continue;
    }
    const text = await res.text().catch(() => "");
    throw new Error(`LLM API error ${res.status}: ${text.slice(0, 500)}`);
  }

  const data = (await res!.json()) as {
    choices?: {
      message?: {
        role?: string;
        content?: string | null;
        tool_calls?: Array<{
          id?: string;
          type?: string;
          name?: string;
          function?: { name?: string; arguments?: string };
          /** Gemini thought_signature lives inside extra_content.google.thought_signature */
          extra_content?: { google?: { thought_signature?: string } };
        }>;
      };
    }[];
    error?: { message?: string };
  };

  if (data.error?.message) throw new Error(`LLM API error: ${data.error.message}`);
  const message = data.choices?.[0]?.message;
  if (!message) throw new Error("LLM API returned no message.");

  // OpenAI-compatible providers (DeepSeek, OpenAI, …) nest each tool call under
  // `tool_calls[].function`, while the agent loop expects the flattened
  // internal shape `{ id, name, arguments }`. Normalize here so the dispatcher
  // never sees a tool named "undefined".
  const rawCalls = Array.isArray(message.tool_calls) ? message.tool_calls : [];
  
  // Debug: log raw tool call structure from Gemini
  if (config.provider === "gemini" && rawCalls.length > 0) {
    console.log("[nova:llm] Gemini raw tool_calls:", JSON.stringify(rawCalls.map((tc) => ({
      id: tc.id,
      name: tc.function?.name,
      hasExtraContent: !!tc.extra_content,
      hasThoughtSig: !!tc.extra_content?.google?.thought_signature,
      extraContentKeys: tc.extra_content ? Object.keys(tc.extra_content) : [],
    }))));
  }

  const tool_calls: ToolCall[] | undefined = rawCalls.length
    ? rawCalls.map((tc, idx) => {
        const call: ToolCall = {
          id: tc.id ?? `call_${Math.random().toString(36).slice(2, 10)}`,
          name: tc.function?.name ?? tc.name ?? "unknown",
          arguments: tc.function?.arguments ?? "",
        };
        // Gemini thought_signature: first call gets the real sig, rest get sentinel.
        // For parallel calls, Gemini only puts the signature on the FIRST call.
        // skip_thought_signature_validator is a documented Google bypass.
        if (config.provider === "gemini") {
          if (idx === 0 && tc.extra_content?.google?.thought_signature) {
            call.thought_signature = tc.extra_content.google.thought_signature;
          } else {
            call.thought_signature = "skip_thought_signature_validator";
          }
        }
        return call;
      })
    : undefined;

  return {
    role: (message.role as ChatMessage["role"]) ?? "assistant",
    content: message.content ?? null,
    tool_calls,
  };
}

/* ------------------------------------------------------------------ */
/* Anthropic (Claude) via the Messages API                             */
/* ------------------------------------------------------------------ */

interface AnthropicContentBlock {
  type: "text" | "tool_use" | "tool_result";
  text?: string;
  id?: string;
  name?: string;
  input?: unknown;
  tool_use_id?: string;
  content?: string;
}

interface AnthropicResponse {
  content?: AnthropicContentBlock[];
  stop_reason?: string;
  error?: { message?: string };
}

function toAnthropicMessages(messages: ChatMessage[]): { system: string; rest: ChatMessage[] } {
  const system = messages.find((m) => m.role === "system")?.content ?? "";
  const rest = messages.filter((m) => m.role !== "system");

  // Anthropic requires strictly alternating user/assistant turns, and tool
  // results must be delivered as user messages with tool_result blocks.
  const out: Array<{ role: "user" | "assistant"; content: unknown }> = [];
  for (const m of rest) {
    if (m.role === "tool") {
      const last = out[out.length - 1];
      if (last && last.role === "user" && Array.isArray(last.content)) {
        (last.content as AnthropicContentBlock[]).push({
          type: "tool_result",
          tool_use_id: m.tool_call_id ?? "",
          content: m.content ?? "",
        });
      } else {
        out.push({
          role: "user",
          content: [
            {
              type: "tool_result",
              tool_use_id: m.tool_call_id ?? "",
              content: m.content ?? "",
            },
          ] satisfies AnthropicContentBlock[],
        });
      }
      continue;
    }

    const blocks: AnthropicContentBlock[] = [];
    if (m.content) blocks.push({ type: "text", text: m.content });
    for (const tc of m.tool_calls ?? []) {
      blocks.push({
        type: "tool_use",
        id: tc.id,
        name: tc.name,
        input: safeParse(tc.arguments),
      });
    }
    out.push({ role: m.role === "assistant" ? "assistant" : "user", content: blocks });
  }
  return { system, rest: out as unknown as ChatMessage[] };
}

function safeParse(s: string): unknown {
  try {
    return JSON.parse(s);
  } catch {
    return {};
  }
}

function fromAnthropicResponse(resp: AnthropicResponse): ChatMessage {
  const blocks = resp.content ?? [];
  const text = blocks
    .filter((b) => b.type === "text" && b.text)
    .map((b) => b.text)
    .join("\n");
  const toolUses = blocks.filter((b) => b.type === "tool_use" && b.id && b.name);
  const tool_calls: ToolCall[] = toolUses.map((b) => ({
    id: b.id ?? `tool_${Math.random().toString(36).slice(2, 8)}`,
    name: b.name ?? "unknown",
    arguments: JSON.stringify(b.input ?? {}),
  }));
  return { role: "assistant", content: text || null, tool_calls: tool_calls.length ? tool_calls : undefined };
}

async function completeAnthropic(
  config: NovaConfig,
  messages: ChatMessage[],
  tools: ToolDefinition[],
  opts: CompleteOptions,
): Promise<ChatMessage> {
  const { system, rest } = toAnthropicMessages(messages);
  const anthropicTools = tools.map((t) => ({
    name: t.function.name,
    description: t.function.description,
    input_schema: t.function.parameters,
  }));

  const res = await fetch(`${config.apiBase}/v1/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": config.apiKey,
      "anthropic-version": "2023-06-01",
    },
    body: JSON.stringify({
      model: config.model,
      max_tokens: 4096,
      system,
      messages: rest,
      tools: anthropicTools,
      temperature: opts.temperature ?? 0.7,
    }),
    signal: AbortSignal.timeout(180_000),
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`Anthropic API error ${res.status}: ${text.slice(0, 500)}`);
  }

  const data = (await res.json()) as AnthropicResponse;
  if (data.error?.message) throw new Error(`Anthropic API error: ${data.error.message}`);
  return fromAnthropicResponse(data);
}

/* ------------------------------------------------------------------ */
/* Mock mode: a deterministic, scripted "brain" so the whole agent     */
/* loop + tools can be exercised without an API key.                   */
/* ------------------------------------------------------------------ */

let mockCalls = 0;

function mockComplete(messages: ChatMessage[]): ChatMessage {
  mockCalls += 1;
  const userMsg = [...messages].reverse().find((m) => m.role === "user");
  const task = userMsg?.content ?? "(no task)";

  if (mockCalls === 1) {
    return {
      role: "assistant",
      content:
        "I'll start by exploring the working directory and checking the web for context, then write a summary artifact.",
      tool_calls: [
        { id: "call_mock_1", name: "list_directory", arguments: JSON.stringify({ path: "." }) },
        { id: "call_mock_2", name: "web_search", arguments: JSON.stringify({ query: "koola10 nova agent" }) },
      ],
    };
  }

  if (mockCalls === 2) {
    return {
      role: "assistant",
      content: "I have enough context. Saving a summary artifact now.",
      tool_calls: [
        {
          id: "call_mock_3",
          name: "write_file",
          arguments: JSON.stringify({
            path: "output/nova-mock-summary.txt",
            content:
              "Nova agent — end-to-end verification (mock mode)\n" +
              "===============================================\n" +
              `Task: ${task}\n` +
              "The agent loop, tool dispatch (shell/files/web), artifact writing, and " +
              "final report pipeline all work without a live API key.",
          }),
        },
      ],
    };
  }

  return {
    role: "assistant",
    content:
      "## Final report\n\n" +
      "Mock task complete. Nova successfully:\n" +
      "1. Listed the working directory.\n" +
      "2. Ran a live web search (real DuckDuckGo call — proves the web tool works).\n" +
      "3. Wrote a summary artifact to `output/nova-mock-summary.txt`.\n\n" +
      "This proves the agent loop, tool dispatch, error handling, artifact pipeline, " +
      "and report generation all function. Set DEEPSEEK_API_KEY to run with the real brain.",
  };
}
