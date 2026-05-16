'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Activity, Radio } from 'lucide-react';

export default function LiveDisplay() {
  const [events, setEvents] = useState<any[]>([]);

  useEffect(() => {
    const eventSource = new EventSource('https://koola10.fly.dev/events/stream');

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setEvents((prev) => [data, ...prev].slice(0, 50));
    };

    return () => eventSource.close();
  }, []);

  return (
    <main className="min-h-screen bg-slate-950 p-4 md:p-8 text-white">
      <header className="mb-8 flex items-center justify-between bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-2xl">
        <div className="flex items-center gap-3">
          <div className="p-3 bg-amber-400 rounded-xl">
            <Radio className="h-8 w-8 text-black animate-pulse" />
          </div>
          <div>
            <h1 className="text-2xl font-black tracking-tight">
              KOOLA10 <span className="text-amber-400">LIVE</span>
            </h1>
            <p className="text-white/50 text-xs font-medium uppercase tracking-[0.2em]">
              Real-time Swarm Intelligence Stream
            </p>
          </div>
        </div>
      </header>

      <Card className="bg-slate-900/50 border-slate-800 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="text-lg font-medium flex items-center gap-2">
            <Activity className="h-5 w-5 text-green-400" />
            System Events
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[70vh] pr-4">
            <div className="space-y-4">
              {events.map((event, i) => (
                <div key={i} className="flex gap-3 text-sm border-b border-white/5 pb-3">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold capitalize text-amber-400">{event.type}</span>
                      <span className="text-xs text-slate-500">
                        {event.timestamp || new Date().toLocaleTimeString()}
                      </span>
                    </div>
                    <pre className="text-slate-400 whitespace-pre-wrap font-mono text-xs">
                      {JSON.stringify(event.data || event, null, 2)}
                    </pre>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
    </main>
  );
}
