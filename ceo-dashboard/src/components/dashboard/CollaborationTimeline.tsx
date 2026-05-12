'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { MessageSquare, CheckCircle, AlertTriangle, Lightbulb } from 'lucide-react';

interface CollabEvent {
  id: string;
  type: string;
  timestamp: string;
  data: any;
}

export function CollaborationTimeline() {
  const [events, setEvents] = useState<CollabEvent[]>([]);

  useEffect(() => {
    const fetchTimeline = async () => {
      try {
        const res = await fetch('https://koola10.fly.dev/collaborate/timeline');
        const data = await res.json();
        setEvents(data.slice(0, 20));
      } catch (err) {
        console.error('Failed to fetch timeline', err);
      }
    };

    fetchTimeline();
    const interval = setInterval(fetchTimeline, 30000);
    return () => clearInterval(interval);
  }, []);

  const getIcon = (type: string) => {
    switch (type) {
      case 'decision': return <CheckCircle className="h-4 w-4 text-green-400" />;
      case 'advisor_note': return <Lightbulb className="h-4 w-4 text-amber-400" />;
      case 'alert': return <AlertTriangle className="h-4 w-4 text-red-400" />;
      default: return <MessageSquare className="h-4 w-4 text-blue-400" />;
    }
  };

  return (
    <Card className="bg-slate-900/50 border-slate-800 backdrop-blur-sm">
      <CardHeader>
        <CardTitle className="text-lg font-medium flex items-center gap-2">
          <MessageSquare className="h-5 w-5 text-purple-400" />
          Collaboration Timeline
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px] pr-4">
          <div className="space-y-4">
            {events.map((event) => (
              <div key={event.id} className="flex gap-3 text-sm">
                <div className="mt-1">{getIcon(event.type)}</div>
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold capitalize text-slate-200">{event.type.replace('_', ' ')}</span>
                    <span className="text-xs text-slate-500">
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <p className="text-slate-400 leading-relaxed">
                    {event.type === 'decision' && `${event.data.decision}: ${event.data.rationale}`}
                    {event.type === 'advisor_note' && event.data.note}
                    {event.type === 'alert' && `${event.data.title}: ${event.data.message}`}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
