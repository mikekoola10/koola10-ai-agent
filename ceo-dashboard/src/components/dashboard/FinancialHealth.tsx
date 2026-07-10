'use client';

import { useFinancialHealth } from '@/hooks/use-dashboard-data';
import { Card, CardContent } from '@/components/ui/card';
import { Wallet, TrendingUp, TrendingDown, Loader2 } from 'lucide-react';

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount);
};

export function FinancialHealth() {
  const { data, error, isLoading } = useFinancialHealth();

  const metrics = [
    {
      label: 'Balance',
      value: data?.balance ?? 0,
      icon: Wallet,
      color: 'text-amber-400',
    },
    {
      label: 'Earned (30D)',
      value: data?.earned ?? 0,
      icon: TrendingUp,
      color: 'text-green-400',
    },
    {
      label: 'Spent (30D)',
      value: data?.spent ?? 0,
      icon: TrendingDown,
      color: 'text-red-400',
    },
  ];

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="bg-white/10 backdrop-blur-md border-white/20 text-white">
            <CardContent className="pt-6 flex items-center justify-center h-24">
              <Loader2 className="h-6 w-6 animate-spin text-white/30" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Card className="bg-white/10 backdrop-blur-md border-white/20 text-white w-full">
        <CardContent className="pt-6 text-center text-red-400">
          Failed to load financial data.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {metrics.map((metric) => (
        <Card key={metric.label} className="bg-white/10 backdrop-blur-md border-white/20 text-white">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-white/60 uppercase tracking-wider">
                  {metric.label}
                </p>
                <p className="text-2xl font-bold mt-1">
                  {formatCurrency(metric.value)}
                </p>
              </div>
              <div className={`p-3 rounded-full bg-white/5 ${metric.color}`}>
                <metric.icon className="h-5 w-5" />
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
