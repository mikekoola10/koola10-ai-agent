'use client';

import { LiveAgentDisplay } from '@/components/dashboard/LiveAgentDisplay';
import { Cpu } from 'lucide-react';

export default function LivePage() {
  return (
    <main className="min-h-screen p-4 md:p-8 max-w-full mx-auto space-y-8 bg-[#0a0a0b]">
      {/* Header */}
      <header className="flex justify-between items-center bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-2xl">
        <div className="flex items-center gap-3">
          <div className="p-3 bg-amber-400 rounded-xl shadow-[0_0_20px_rgba(251,191,36,0.4)]">
            <Cpu className="h-8 w-8 text-black" />
          </div>
          <div>
            <h1 className="text-2xl font-black text-white tracking-tight">
              LIVE <span className="text-amber-400">AGENT</span> DISPLAY
            </h1>
            <p className="text-white/50 text-xs font-medium uppercase tracking-[0.2em]">
              Real-time workspace inspection
            </p>
          </div>
        </div>
      </header>

      <LiveAgentDisplay />
    </main>
  );
}
