'use client';

import { HealthStatus } from '@/components/dashboard/HealthStatus';
import { FinancialHealth } from '@/components/dashboard/FinancialHealth';
import { RevenuePulse } from '@/components/dashboard/RevenuePulse';
import { AgentStatus } from '@/components/dashboard/AgentStatus';
import { DailyReport } from '@/components/dashboard/DailyReport';
import { QuickActions } from '@/components/dashboard/QuickActions';
import SystemEvolution from '@/components/dashboard/SystemEvolution';
import { NovaChat } from '@/components/chat/NovaChat';
import { Cpu } from 'lucide-react';

export default function Dashboard() {
  return (
    <main className="min-h-screen p-4 md:p-8 max-w-7xl mx-auto space-y-8">
      {/* Header */}
      <header className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-2xl">
        <div className="flex items-center gap-3">
          <div className="p-3 bg-amber-400 rounded-xl shadow-[0_0_20px_rgba(251,191,36,0.4)]">
            <Cpu className="h-8 w-8 text-black" />
          </div>
          <div>
            <h1 className="text-2xl font-black text-white tracking-tight">
              KOOLA10 <span className="text-amber-400">COMMAND</span>
            </h1>
            <p className="text-white/50 text-xs font-medium uppercase tracking-[0.2em]">
              Autonomous Swarm Intelligence
            </p>
          </div>
        </div>
        <QuickActions />
      </header>

      {/* Top Row: Health & Stats */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-1">
          <HealthStatus />
        </div>
        <div className="lg:col-span-3">
          <FinancialHealth />
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Revenue & Agents */}
        <div className="space-y-6">
          <RevenuePulse />
          <SystemEvolution />
          <AgentStatus />
        </div>

        {/* Right Column: Daily Report */}
        <div className="lg:col-span-2">
          <DailyReport />
        </div>
      </div>

      {/* Footer */}
      <footer className="pt-8 pb-4 text-center">
        <p className="text-white/20 text-sm font-medium tracking-widest uppercase">
          Powered by <span className="text-white/40">Koola10</span>
        </p>
      </footer>

      {/* Floating Nova Chat */}
      <NovaChat />
    </main>
  );
}
