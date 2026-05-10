'use client';

import { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { MessageSquare, X, Send, Loader2, Bot } from 'lucide-react';
import { postRequest } from '@/lib/api';
import { ScrollArea } from '@/components/ui/scroll-area';

interface Message {
  role: 'user' | 'nova';
  content: string;
}

export function NovaChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTo(0, scrollRef.current.scrollHeight);
    }
  }, [messages]);

  const handleSend = async (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || isLoading) return;

    const userMessage = input.trim();
    setInput('');
    setMessages((prev) => [...prev, { role: 'user', content: userMessage }]);
    setIsLoading(true);

    try {
      const data = await postRequest('/ai/chat', { prompt: userMessage }) as { response: string };
      setMessages((prev) => [...prev, { role: 'nova', content: data.response }]);
    } catch (error) {
      console.error('Nova chat error:', error);
      setMessages((prev) => [...prev, { role: 'nova', content: '⚠️ Error: System temporarily unreachable.' }]);
    } finally {
      setIsLoading(false);
    }
  };

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
                    <p className="text-sm text-white/40 italic">Ask me anything about the swarm.</p>
                  </div>
                )}
                {messages.map((m, i) => (
                  <div
                    key={i}
                    className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}
                  >
                    <div
                      className={`max-w-[80%] rounded-2xl px-3 py-2 text-sm ${
                        m.role === 'user'
                          ? 'bg-amber-400 text-black rounded-tr-none'
                          : 'bg-white/10 text-white border border-white/10 rounded-tl-none'
                      }`}
                    >
                      {m.content}
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
              <Button
                type="submit"
                disabled={isLoading || !input.trim()}
                className="bg-amber-400 hover:bg-amber-500 text-black px-3"
              >
                <Send className="h-4 w-4" />
              </Button>
            </form>
          </CardFooter>
        </Card>
      )}
    </div>
  );
}
