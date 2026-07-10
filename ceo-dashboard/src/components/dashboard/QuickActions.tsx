'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Play, RefreshCcw, Loader2, CheckCircle } from 'lucide-react';
import { postRequest } from '@/lib/api';
import { useHealth, useRevenue, useAgentStatus, useFinancialHealth, useDailyReport } from '@/hooks/use-dashboard-data';

export function QuickActions() {
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [monitorSuccess, setMonitorSuccess] = useState(false);

  const health = useHealth();
  const revenue = useRevenue();
  const agents = useAgentStatus();
  const financial = useFinancialHealth();
  const report = useDailyReport();

  const handleRunMonitor = async () => {
    setIsMonitoring(true);
    setMonitorSuccess(false);
    try {
      await postRequest('/grants/monitor');
      setMonitorSuccess(true);
      setTimeout(() => setMonitorSuccess(false), 3000);
    } catch (error) {
      console.error('Failed to run monitor:', error);
    } finally {
      setIsMonitoring(false);
    }
  };

  const handleRefreshAll = () => {
    health.mutate();
    revenue.mutate();
    agents.mutate();
    financial.mutate();
    report.mutate();
  };

  return (
    <div className="flex gap-4">
      <Button
        onClick={handleRunMonitor}
        disabled={isMonitoring}
        className="bg-amber-400 hover:bg-amber-500 text-black font-bold"
      >
        {isMonitoring ? (
          <Loader2 className="h-4 w-4 animate-spin mr-2" />
        ) : monitorSuccess ? (
          <CheckCircle className="h-4 w-4 mr-2" />
        ) : (
          <Play className="h-4 w-4 mr-2 fill-current" />
        )}
        {monitorSuccess ? 'Monitor Started' : 'Run Daily Monitor'}
      </Button>
      <Button
        variant="outline"
        onClick={handleRefreshAll}
        className="border-white/20 bg-white/5 hover:bg-white/10 text-white"
      >
        <RefreshCcw className="h-4 w-4 mr-2" />
        Refresh All
      </Button>
    </div>
  );
}
