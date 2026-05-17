'use client';

import React from 'react';
import { Card } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Bot, User, Sparkles } from 'lucide-react';

interface Message {
  role: 'assistant' | 'user';
  content: string;
}

interface ConversationPanelProps {
  messages: Message[];
  isActive: boolean;
}

export function ConversationPanel({ messages, isActive }: ConversationPanelProps) {
  if (!isActive && messages.length === 0) return null;

  return (
    <Card className="fixed bottom-32 right-4 w-96 max-h-[500px] bg-indigo-950/95 border-amber-400/30 backdrop-blur-xl shadow-2xl flex flex-col z-40 animate-in slide-up">
      <div className="p-4 border-b border-white/10 flex items-center gap-2">
        <Sparkles className="h-5 w-5 text-amber-400" />
        <h3 className="text-white font-bold tracking-tight">JARVIS ANALYSIS</h3>
      </div>
      <ScrollArea className="flex-1 p-4">
        <div className="space-y-4">
          {messages.map((m, i) => (
            <div key={i} className={`flex gap-3 ${m.role === 'user' ? 'flex-row-reverse' : ''}`}>
              <div className={`p-2 rounded-lg shrink-0 ${m.role === 'assistant' ? 'bg-amber-400' : 'bg-white/10'}`}>
                {m.role === 'assistant' ? <Bot className="h-4 w-4 text-black" /> : <User className="h-4 w-4 text-white" />}
              </div>
              <div className={`p-3 rounded-2xl text-sm ${
                m.role === 'assistant'
                  ? 'bg-white/5 text-white border border-white/10'
                  : 'bg-amber-400 text-black font-medium'
              }`}>
                {m.content}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </Card>
  );
}
