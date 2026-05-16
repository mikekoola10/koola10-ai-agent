'use client';

import { useAgentStatus } from '@/hooks/use-dashboard-data';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, RefreshCcw, Users } from 'lucide-react';
import { Button } from '@/components/ui/button';

export function AgentStatus() {
  const { data, error, isLoading, mutate } = useAgentStatus();

  return (
    <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium flex items-center">
          <Users className="h-4 w-4 mr-2 text-amber-400" />
          Agent Swarm Status
        </CardTitle>
        <Button variant="ghost" size="sm" onClick={() => mutate()} className="text-white hover:bg-white/10 h-8 w-8 p-0">
          <RefreshCcw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
        </Button>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex justify-center py-4">
            <Loader2 className="h-6 w-6 animate-spin text-white/50" />
          </div>
        ) : error ? (
          <p className="text-sm text-red-400 py-4">Failed to load agents.</p>
        ) : !data?.verticals || data.verticals.length === 0 ? (
          <p className="text-sm text-white/50 py-4 italic">No active agents found.</p>
        ) : (
          <div className="grid grid-cols-2 gap-3 py-2">
            {data.verticals.map((vertical: string) => (
              <div
                key={vertical}
                className="flex items-center justify-between bg-white/5 rounded-lg px-3 py-2 border border-white/10"
              >
                <span className="text-sm capitalize">{vertical}</span>
                <div className="flex items-center">
                  <div className="h-2 w-2 rounded-full bg-green-400 mr-2 animate-pulse" />
                  <span className="text-[10px] uppercase font-bold text-green-400">active</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
