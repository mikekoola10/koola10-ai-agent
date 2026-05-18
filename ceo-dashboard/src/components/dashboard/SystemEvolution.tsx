'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useEffect, useState } from 'react';

interface Optimization {
  vertical: string;
  optimization: string;
  justification: string;
}

export default function SystemEvolution() {
  const [skillCount, setSkillCount] = useState(0);
  const [recentOptimizations, setRecentOptimizations] = useState<Optimization[]>([]);

  useEffect(() => {
    // In a real scenario, we would fetch this from the Go agent
    // For now, we simulate or listen to SSE events
    setSkillCount(12); // Simulated starting count
    setRecentOptimizations([
      { vertical: 'grant', optimization: 'ROI-first filtering', justification: 'High token waste on low-probability grants' }
    ]);
  }, []);

  return (
    <Card className="bg-black/40 border-amber-500/30 text-white backdrop-blur-md">
      <CardHeader>
        <CardTitle className="text-amber-400 font-mono text-lg">SYSTEM_EVOLUTION</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex justify-between items-center">
            <span className="text-gray-400 font-mono">DISTILLED_SKILLS:</span>
            <span className="text-2xl font-bold text-amber-500">{skillCount}</span>
          </div>

          <div className="space-y-2">
            <h4 className="text-sm font-semibold text-gray-300 font-mono border-b border-gray-700 pb-1">RECENT_OPTIMIZATIONS</h4>
            {recentOptimizations.map((opt, i) => (
              <div key={i} className="text-xs bg-amber-500/10 p-2 rounded border border-amber-500/20">
                <div className="font-bold text-amber-400">[{opt.vertical.toUpperCase()}]</div>
                <div className="text-gray-200 mt-1">{opt.optimization}</div>
              </div>
            ))}
          </div>

          <div className="space-y-2">
            <h4 className="text-sm font-semibold text-gray-300 font-mono border-b border-gray-700 pb-1">WEEKLY_BENCHMARKS (ADARUBRIC)</h4>
            <div className="text-xs space-y-1">
              <div className="flex justify-between">
                <span>Task Autonomy</span>
                <span className="text-amber-400">8.5/10</span>
              </div>
              <div className="flex justify-between">
                <span>Error Recovery</span>
                <span className="text-amber-400">9.2/10</span>
              </div>
              <div className="flex justify-between">
                <span>Skill Retention</span>
                <span className="text-amber-400">7.8/10</span>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
