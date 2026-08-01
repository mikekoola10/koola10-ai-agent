/**
 * POST /api/chat
 * Vercel AI SDK streaming chat endpoint backing the Nova Assistant widget.
 * Requires a signed-in user (session cookie) and an OpenAI API key.
 */

import { NextResponse } from "next/server";
import { createOpenAI } from "@ai-sdk/openai";
import {
  convertToModelMessages,
  createUIMessageStreamResponse,
  streamText,
  toUIMessageStream,
  type UIMessage,
} from "ai";
import { createClient } from "@/lib/supabase/server";

export const maxDuration = 30;

function getOpenAIKey() {
  return (
    process.env.OPENAI_API_KEY ||
    process.env.MIKEKOOLA10ORG_OPENAI_API_KEY ||
    ""
  );
}

const SYSTEM_PROMPT = `You are NOVA, the autonomous swarm intelligence assistant for Koola10 Command (Mikekoola10Org).

You help the operator (the business owner) run their autonomous agent empire. You are concise, sharp, and direct — think a battle-tested ops commander, not a chatbot.

You can help with:
- Reading and interpreting the dashboard: revenue, affiliate, bounty, content, product, and grant verticals
- Agent swarm status and operations
- Financial health, cash flow, and business strategy
- Turning vague ideas into concrete next actions

Rules:
- Be direct and practical. No fluff, no disclaimers, no "As an AI...".
- When asked about numbers or status, give the best answer you can from general knowledge and say clearly what the operator should verify in the dashboard.
- Keep answers tight unless the operator asks for depth.`;

export async function POST(req: Request) {
  const apiKey = getOpenAIKey();
  if (!apiKey) {
    return NextResponse.json(
      {
        error:
          "AI is not configured. Add OPENAI_API_KEY (or MIKEKOOLA10ORG_OPENAI_API_KEY) to the project keys.",
      },
      { status: 500 },
    );
  }

  // Only signed-in users may talk to Nova.
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    return NextResponse.json(
      { error: "You must be signed in to chat with Nova." },
      { status: 401 },
    );
  }

  const { messages }: { messages: UIMessage[] } = await req.json();

  const openai = createOpenAI({ apiKey });
  const result = streamText({
    model: openai("gpt-4o-mini"),
    system: SYSTEM_PROMPT,
    messages: await convertToModelMessages(messages),
  });

  return createUIMessageStreamResponse({
    stream: toUIMessageStream({ stream: result.stream }),
  });
}
