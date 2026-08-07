/**
 * Shared types for the Nova agent.
 */

export type Role = "system" | "user" | "assistant" | "tool";

/** A function call the model asked to invoke. */
export interface ToolCall {
  id: string;
  name: string;
  /** JSON-encoded arguments for the tool. */
  arguments: string;
}

/** One message in the agent conversation. */
export interface ChatMessage {
  role: Role;
  content: string | null;
  tool_calls?: ToolCall[];
  tool_call_id?: string;
  name?: string;
}

/** OpenAI-style tool definition sent to the model. */
export interface ToolDefinition {
  type: "function";
  function: {
    name: string;
    description: string;
    parameters: Record<string, unknown>;
  };
}

/** Result of running an agent to completion. */
export interface AgentResult {
  report: string;
  steps: number;
  toolCalls: number;
  durationMs: number;
}
