'use client';

import { useHealth } from '@/hooks/use-dashboard-data';
import { Card, CardContent } from '@/components/ui/card';
import { AlertCircle, CheckCircle2, Loader2, RefreshCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';

export function HealthStatus() {
  const { data, error, isLoading, mutate } = useHealth();

  if (isLoading) {
    return (
      <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white">
        <CardContent className="pt-6 flex items-center justify-center">
          <Loader2 className="h-4 w-4 animate-spin mr-2" />
          <span>Checking systems...</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white">
        <CardContent className="pt-6 flex items-center justify-between">
          <div className="flex items-center text-red-400">
            <AlertCircle className="h-5 w-5 mr-2" />
            <span>Connection error</span>
          </div>
          <Button variant="ghost" size="sm" onClick={() => mutate()} className="text-white hover:bg-white/10">
            <RefreshCcw className="h-4 w-4" />
          </Button>
        </CardContent>
      </Card>
    );
  }

  const isOk = data?.status === 'ok';

  return (
    <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white">
      <CardContent className="pt-6 flex items-center justify-between">
        <div className="flex items-center">
          {isOk ? (
            <CheckCircle2 className="h-5 w-5 mr-2 text-green-400" />
          ) : (
            <AlertCircle className="h-5 w-5 mr-2 text-amber-400" />
          )}
          <span className="font-medium">
            {isOk ? '✅ All systems operational' : '⚠️ Systems unavailable'}
          </span>
        </div>
        <Button variant="ghost" size="sm" onClick={() => mutate()} className="text-white hover:bg-white/10">
          <RefreshCcw className="h-4 w-4" />
        </Button>
      </CardContent>
    </Card>
  );
}
