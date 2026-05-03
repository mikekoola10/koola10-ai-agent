'use client';

import { useDailyReport } from '@/hooks/use-dashboard-data';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { FileText, Loader2, RefreshCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';

export function DailyReport() {
  const { data, error, isLoading, mutate } = useDailyReport();

  const reportText = typeof data === 'string'
    ? data
    : data && typeof data === 'object'
      ? JSON.stringify(data, null, 2)
      : null;

  return (
    <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white h-[400px] flex flex-col">
      <CardHeader className="flex flex-row items-center justify-between pb-2 shrink-0">
        <CardTitle className="text-sm font-medium flex items-center">
          <FileText className="h-4 w-4 mr-2 text-amber-400" />
          Daily Swarm Report
        </CardTitle>
        <Button variant="ghost" size="sm" onClick={() => mutate()} className="text-white hover:bg-white/10 h-8 w-8 p-0">
          <RefreshCcw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
        </Button>
      </CardHeader>
      <CardContent className="grow overflow-hidden pt-2">
        {isLoading ? (
          <div className="flex justify-center items-center h-full">
            <Loader2 className="h-8 w-8 animate-spin text-white/30" />
          </div>
        ) : error ? (
          <div className="text-center py-12 text-red-400">
            <p>Failed to load the daily report.</p>
          </div>
        ) : !reportText ? (
          <div className="text-center py-12 text-white/40 italic">
            <p>No report available for today.</p>
          </div>
        ) : (
          <ScrollArea className="h-full pr-4">
            <pre className="font-mono text-xs leading-relaxed whitespace-pre-wrap text-white/90">
              {reportText}
            </pre>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  );
}
