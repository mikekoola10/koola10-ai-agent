'use client';

import { useRevenue } from '@/hooks/use-dashboard-data';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { BarChart3, Loader2 } from 'lucide-react';

interface RevenueItem {
  vertical: string;
  revenue: number;
}

export function RevenuePulse() {
  const { data, error, isLoading } = useRevenue();

  const revenueData: RevenueItem[] = Array.isArray(data) ? data : [];
  const maxRevenue = Math.max(...revenueData.map((item) => item.revenue), 1);
  const topEarner = revenueData.length > 0
    ? revenueData.reduce((prev, current) => (prev.revenue > current.revenue) ? prev : current)
    : null;

  return (
    <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white">
      <CardHeader>
        <CardTitle className="text-sm font-medium flex items-center">
          <BarChart3 className="h-4 w-4 mr-2 text-amber-400" />
          Revenue Pulse (24h)
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-white/30" />
          </div>
        ) : error ? (
          <p className="text-sm text-red-400 py-8">Failed to load revenue data.</p>
        ) : revenueData.length === 0 ? (
          <div className="text-center py-12 text-white/40">
            <p className="italic">The ledger is quiet today.</p>
            <p className="text-xs mt-1">No recent revenue entries found.</p>
          </div>
        ) : (
          <div className="space-y-4 py-2">
            {revenueData.map((item) => (
              <div key={item.vertical} className="space-y-1">
                <div className="flex justify-between text-xs mb-1">
                  <span className="capitalize font-medium">{item.vertical}</span>
                  <span className="text-amber-400">${item.revenue.toFixed(2)}</span>
                </div>
                <div className="w-full bg-white/5 rounded-full h-2 overflow-hidden border border-white/5">
                  <div
                    className={`h-full transition-all duration-500 ease-out ${
                      topEarner?.vertical === item.vertical ? 'bg-amber-400' : 'bg-white/40'
                    }`}
                    style={{ width: `${(item.revenue / maxRevenue) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
