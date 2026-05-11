'use client';

import { useState, useEffect, useRef } from 'react';
import { Card } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Video, Terminal, RefreshCw, AlertCircle } from 'lucide-react';
import axios from 'axios';

interface Event {
  type: string;
  data: string | Record<string, unknown>;
  timestamp: string;
}

export function LiveAgentDisplay() {
  const [screenshot, setScreenshot] = useState<string | null>(null);
  const [events, setEvents] = useState<Event[]>([]);
  const [isPolling, setIsPolling] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Poll screenshot every 3 seconds
  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isPolling) {
      const fetchScreenshot = async () => {
        try {
          const res = await axios.get('https://koola10-browser.fly.dev/browser/live-screenshot');
          if (res.data.screenshot) {
            setScreenshot(`data:image/png;base64,${res.data.screenshot}`);
            setError(null);
          }
        } catch (err) {
          console.error('Failed to fetch screenshot:', err);
          setError('Live agent disconnected. Retrying...');
        }
      };

      fetchScreenshot();
      interval = setInterval(fetchScreenshot, 3000);
    }
    return () => clearInterval(interval);
  }, [isPolling]);

  // Listen to SSE events
  useEffect(() => {
    const eventSource = new EventSource('https://koola10.fly.dev/events/stream');

    eventSource.onmessage = (event) => {
      try {
        const parsedEvent = JSON.parse(event.data);
        if (parsedEvent.type === 'connected') return;

        setEvents((prev) => [...prev, parsedEvent].slice(-50));
      } catch (err) {
        console.error('Failed to parse SSE event:', err);
      }
    };

    eventSource.onerror = (err) => {
      console.error('SSE connection error:', err);
      eventSource.close();
    };

    return () => eventSource.close();
  }, []);

  // Auto-scroll event log
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [events]);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 h-[calc(100vh-8rem)]">
      {/* Screenshot Panel */}
      <Card className="lg:col-span-3 bg-white/5 backdrop-blur-xl border-white/10 overflow-hidden flex flex-col relative group">
        <div className="absolute top-4 left-4 z-10 flex items-center gap-2 px-3 py-1.5 bg-black/60 backdrop-blur-md rounded-full border border-white/10">
          <div className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
          <span className="text-[10px] font-bold uppercase tracking-widest text-white/90">Live Feed</span>
        </div>

        {error && (
          <div className="absolute inset-0 z-20 flex items-center justify-center bg-black/40 backdrop-blur-sm">
            <div className="flex flex-col items-center gap-3 p-6 bg-black/80 rounded-2xl border border-white/10">
              <AlertCircle className="h-8 w-8 text-amber-400" />
              <p className="text-white font-medium">{error}</p>
              <RefreshCw className="h-4 w-4 text-white/40 animate-spin" />
            </div>
          </div>
        )}

        <div className="flex-1 bg-black/40 flex items-center justify-center overflow-hidden p-4">
          {screenshot ? (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={screenshot}
              alt="Live Agent Workspace"
              className="max-w-full max-h-full object-contain rounded-lg shadow-2xl transition-all duration-500 group-hover:scale-[1.01]"
            />
          ) : (
            <div className="flex flex-col items-center gap-4 text-white/20">
              <Video className="h-16 w-16" />
              <p className="font-medium animate-pulse">Initializing Agent Workspace...</p>
            </div>
          )}
        </div>

        <div className="p-4 bg-black/20 border-t border-white/5 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <Video className="h-4 w-4 text-amber-400" />
            <span className="text-xs font-medium text-white/60 tracking-tight">
              Browser Instance: <span className="text-white">koola10-browser</span>
            </span>
          </div>
          <div className="flex items-center gap-3">
             <button
                onClick={() => setIsPolling(!isPolling)}
                className="text-[10px] font-bold uppercase tracking-widest text-white/40 hover:text-white transition-colors"
             >
               {isPolling ? 'Pause Stream' : 'Resume Stream'}
             </button>
          </div>
        </div>
      </Card>

      {/* Event Log Panel */}
      <Card className="lg:col-span-1 bg-white/5 backdrop-blur-xl border-white/10 flex flex-col overflow-hidden">
        <div className="p-4 border-b border-white/5 flex items-center gap-2 bg-white/5">
          <Terminal className="h-4 w-4 text-amber-400" />
          <h2 className="text-sm font-bold uppercase tracking-widest text-white/90">Agent Activity</h2>
        </div>

        <ScrollArea className="flex-1 p-4" ref={scrollRef}>
          <div className="space-y-4">
            {events.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-white/20 gap-3">
                <Terminal className="h-8 w-8" />
                <p className="text-xs font-medium italic">Waiting for events...</p>
              </div>
            ) : (
              events.map((e, idx) => (
                <div key={idx} className="group animate-in fade-in slide-in-from-right-2 duration-300">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-[10px] font-mono text-amber-400/60">
                      {new Date(e.timestamp).toLocaleTimeString([], { hour12: false })}
                    </span>
                    <span className="px-1.5 py-0.5 rounded text-[9px] font-black uppercase tracking-tighter bg-white/10 text-white/80 group-hover:bg-amber-400 group-hover:text-black transition-colors">
                      {e.type}
                    </span>
                  </div>
                  <p className="text-xs font-medium text-white/70 leading-relaxed break-words">
                    {typeof e.data === 'string' ? e.data : JSON.stringify(e.data)}
                  </p>
                </div>
              ))
            )}
          </div>
        </ScrollArea>

        <div className="p-3 bg-black/40 border-t border-white/5 text-center">
          <span className="text-[9px] font-bold text-white/20 uppercase tracking-[0.3em]">
            Real-time SSE Tunnel Active
          </span>
        </div>
      </Card>
    </div>
  );
}
