'use client';

import { useState, useRef, useEffect } from 'react';
import { useChat } from '@ai-sdk/react';
import { TextStreamChatTransport } from 'ai';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { MessageSquare, X, Send, Loader2, Bot, RefreshCw } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';

/** Concatenate the text parts of a UI message (streaming-safe). */
function messageText(parts: ReadonlyArray<{ type: string; text?: string }>): string {
  return parts
    .filter((p) => p.type === 'text' && typeof p.text === 'string')
    .map((p) => p.text)
    .join('');
}

export function NovaChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [input, setInput] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const { messages, sendMessage, status, error, regenerate, stop } = useChat({
    transport: new TextStreamChatTransport({ api: '/api/chat' }),
    messages: [],
  });

  const isLoading = status === 'submitted' || status === 'streaming';

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTo(0, scrollRef.current.scrollHeight);
    }
  }, [messages, isLoading]);

  const handleSend = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || isLoading) return;
    sendMessage({ text: input.trim() });
    setInput('');
  };

  const displayError =
    error instanceof Error
      ? error.message
      : error
        ? 'System temporarily unreachable.'
        : null;

  return (
    <div className="fixed bottom-6 right-6 z-50">
      {!isOpen ? (
        <Button
          onClick={() => setIsOpen(true)}
          className="h-14 w-14 rounded-full bg-amber-400 hover:bg-amber-500 text-black shadow-lg"
        >
          <MessageSquare className="h-6 w-6" />
        </Button>
      ) : (
        <Card className="w-80 sm:w-96 h-[500px] flex flex-col bg-[#1a1033]/95 backdrop-blur-xl border-white/20 shadow-2xl overflow-hidden">
          <CardHeader className="bg-white/5 border-b border-white/10 flex flex-row items-center justify-between p-4 shrink-0">
            <CardTitle className="text-sm font-bold flex items-center text-white">
              <Bot className="h-4 w-4 mr-2 text-amber-400" />
              Nova Assistant
            </CardTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsOpen(false)}
              className="text-white/50 hover:text-white h-8 w-8 p-0"
            >
              <X className="h-4 w-4" />
            </Button>
          </CardHeader>

          <CardContent className="grow overflow-hidden p-0 flex flex-col">
            <ScrollArea className="flex-1 p-4">
              <div className="space-y-4">
                {messages.length === 0 && (
                  <div className="text-center py-8">
                    <Bot className="h-8 w-8 mx-auto text-amber-400/20 mb-2" />
                    <p className="text-sm text-white/40 italic">
                      Ask me anything about the swarm.
                    </p>
                  </div>
                )}
                {messages.map((m, i) => (
                  <div
                    key={i}
                    className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}
                  >
                    <div
                      className={`max-w-[80%] rounded-2xl px-3 py-2 text-sm whitespace-pre-wrap ${
                        m.role === 'user'
                          ? 'bg-amber-400 text-black rounded-tr-none'
                          : 'bg-white/10 text-white border border-white/10 rounded-tl-none'
                      }`}
                    >
                      {messageText(m.parts) || (isLoading && m.role === 'assistant' ? '…' : '')}
                    </div>
                  </div>
                ))}
                {isLoading && (
                  <div className="flex justify-start">
                    <div className="bg-white/10 text-white border border-white/10 rounded-2xl rounded-tl-none px-3 py-2 text-sm">
                      <Loader2 className="h-4 w-4 animate-spin" />
                    </div>
                  </div>
                )}
                {displayError && (
                  <div className="flex justify-center">
                    <div className="bg-red-500/10 text-red-300 border border-red-500/30 rounded-xl px-3 py-2 text-xs text-center max-w-full">
                      {displayError}
                    </div>
                  </div>
                )}
              </div>
            </ScrollArea>
          </CardContent>

          <CardFooter className="p-4 bg-white/5 border-t border-white/10 shrink-0">
            <form onSubmit={handleSend} className="flex w-full gap-2">
              <Input
                placeholder="Type a message..."
                value={input}
                onChange={(e) => setInput(e.target.value)}
                className="bg-black/20 border-white/10 text-white placeholder:text-white/30"
              />
              {isLoading ? (
                <Button
                  type="button"
                  onClick={() => stop()}
                  className="bg-white/10 hover:bg-white/20 text-white px-3"
                  title="Stop"
                >
                  <RefreshCw className="h-4 w-4" />
                </Button>
              ) : (
                <Button
                  type="submit"
                  disabled={!input.trim()}
                  className="bg-amber-400 hover:bg-amber-500 text-black px-3"
                >
                  <Send className="h-4 w-4" />
                </Button>
              )}
              {error && (
                <Button
                  type="button"
                  onClick={() => regenerate()}
                  className="bg-white/10 hover:bg-white/20 text-white px-3"
                  title="Retry"
                >
                  <RefreshCw className="h-4 w-4" />
                </Button>
              )}
            </form>
          </CardFooter>
        </Card>
      )}
    </div>
  );
}
