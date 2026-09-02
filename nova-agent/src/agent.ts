import type { NovaConfig } from "./config.js";
import { complete } from "./llm.js";
import type { AgentResult, ChatMessage } from "./types.js";
import { buildToolDefinitions, dispatchTool } from "./tools/index.js";
import { primeContext, memorySummary } from "./memory.js";

const SYSTEM_PROMPT = `You are Nova, an autonomous AI agent built for the koola10 team. You are given an open-ended task and you work it through to completion on your own, using the tools available to you.

How you operate:
1. Think step by step. Break the task into concrete steps and execute them in order. Do not ask for permission to proceed — act.
2. Prefer real actions over speculation. Read files before quoting them, run commands to verify, and search the web for up-to-date facts. Never invent file contents, command output, or search results you did not observe.
3. After each tool call you will see its output. Use it to decide the next step. When something fails, correct course: retry with a fix, try a different approach, or clearly report the blocker.
4. When the task is complete, produce a concise final report: what you did, the key results and artifacts (with file paths), and any caveats. Write long deliverables to files rather than dumping them in chat.
5. Keep intermediate narration brief. Tool arguments must always be valid JSON.
6. You have persistent memory. Use it: record what you learn, patterns you discover, and mistakes to avoid. Memory persists across sessions.

Formatting:
- The final message must start with the heading "## Final report".
- Reference artifact paths so the user can open them (e.g. output/report.md).`;

const MAX_CONTEXT_CHARS = 120_000;

export interface AgentCallbacks {
  onStep?: (info: { step: number; toolNames: string[]; elapsedMs: number }) => void;
  onTool?: (info: { name: string; output: string; elapsedMs: number }) => void;
  onError?: (err: Error) => void;
}

/** Runs the agent loop until the model answers without tool calls. */
/** Detect task type from the task text for memory priming. */
function detectTaskType(task: string): string {
  const lower = task.toLowerCase();
  if (/bounty.*sweep|sweep.*bounty|scan.*bounty/i.test(lower)) return "sweep";
  if (/bounty.*solve|solve.*bounty|clone.*fix|implement.*fix/i.test(lower)) return "solve";
  if (/report|daily|summary/i.test(lower)) return "report";
  if (/research|investigate|analyze/i.test(lower)) return "research";
  return "general";
}

export async function runAgent(
  task: string,
  config: NovaConfig,
  cbs: AgentCallbacks = {},
): Promise<AgentResult> {
  // Prime context from the memory vault before starting
  const taskType = detectTaskType(task);
  const memCtx = primeContext(taskType);
  const memSummary = memorySummary();

  const messages: ChatMessage[] = [
    { role: "system", content: SYSTEM_PROMPT },
    ...(memCtx ? [{ role: "user" as const, content: `[MEMORY CONTEXT — read this before starting]\n\n${memCtx}` }] : []),
    { role: "user", content: task },
  ];
  const tools = buildToolDefinitions();
  const started = Date.now();
  let steps = 0;
  let toolCalls = 0;

  const effectiveMaxSteps = taskType === "solve" ? config.solveSteps : config.maxSteps;
  for (let step = 0; step < effectiveMaxSteps; step++) {
    let reply: ChatMessage;
    try {
      reply = await complete(config, messages, tools);
    } catch (err) {
      const e = err as Error;
      cbs.onError?.(e);
      return {
        report: `Agent stopped due to an error: ${e.message}`,
        steps,
        toolCalls,
        durationMs: Date.now() - started,
      };
    }

    // No tool calls → this is the final answer.
    if (!reply.tool_calls || reply.tool_calls.length === 0) {
      const report = reply.content ?? "";
      return {
        report,
        steps: steps + 1,
        toolCalls,
        durationMs: Date.now() - started,
      };
    }

    steps += 1;
    cbs.onStep?.({
      step: steps,
      toolNames: reply.tool_calls.map((tc) => tc.name),
      elapsedMs: Date.now() - started,
    });

    messages.push({ role: "assistant", content: reply.content ?? "", tool_calls: reply.tool_calls });

    for (const tc of reply.tool_calls) {
      toolCalls += 1;
      const output = await dispatchTool(tc, config);
      cbs.onTool?.({ name: tc.name, output, elapsedMs: Date.now() - started });
      messages.push({
        role: "tool",
        tool_call_id: tc.id,
        content: output,
        name: tc.name,
      });
    }

    trimContext(messages);
  }

  return {
    report: `Reached the maximum of ${config.maxSteps} steps without finishing. Increase NOVA_MAX_STEPS for longer tasks.`,
    steps,
    toolCalls,
    durationMs: Date.now() - started,
  };
}

/** Keeps the conversation under a rough character budget by dropping oldest turns. */
function trimContext(messages: ChatMessage[]): void {
  let total = 0;
  for (const m of messages) {
    total += (m.content?.length ?? 0) + JSON.stringify(m.tool_calls ?? []).length;
  }
  while (total > MAX_CONTEXT_CHARS && messages.length > 4) {
    const idx = messages.findIndex((m, i) => i > 0 && m.role !== "system");
    if (idx === -1) break;
    const removed = messages.splice(idx, 1)[0]!;
    total -= (removed.content?.length ?? 0) + JSON.stringify(removed.tool_calls ?? []).length;

    // If we removed an assistant turn that made tool calls, drop its orphaned
    // tool-result messages too (otherwise the API rejects the sequence).
    if (removed.tool_calls?.length) {
      const ids = new Set(removed.tool_calls.map((tc) => tc.id));
      for (let i = messages.length - 1; i >= 0; i--) {
        const m = messages[i]!;
        if (m.role === "tool" && m.tool_call_id && ids.has(m.tool_call_id)) {
          total -= m.content?.length ?? 0;
          messages.splice(i, 1);
        }
      }
    }
  }
}
