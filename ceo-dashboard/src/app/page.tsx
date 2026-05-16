'use client';

import { HealthStatus } from '@/components/dashboard/HealthStatus';
import { FinancialHealth } from '@/components/dashboard/FinancialHealth';
import { RevenuePulse } from '@/components/dashboard/RevenuePulse';
import { AgentStatus } from '@/components/dashboard/AgentStatus';
import { CollaborationTimeline } from '@/components/dashboard/CollaborationTimeline';
import { DailyReport } from '@/components/dashboard/DailyReport';
import { QuickActions } from '@/components/dashboard/QuickActions';
import { NotificationPanel } from '@/components/dashboard/NotificationPanel';
import { VoiceListener } from '@/components/dashboard/VoiceListener';
import { ConversationPanel } from '@/components/dashboard/ConversationPanel';
import { NovaChat } from '@/components/chat/NovaChat';
import { Cpu } from 'lucide-react';
import { useEffect, useState } from 'react';

export default function Dashboard() {
  const [activeNotification, setActiveNotification] = useState<any>(null);
  const [isListening, setIsListening] = useState(false);
  const [conversation, setConversation] = useState<any[]>([]);

  useEffect(() => {
    const eventSource = new EventSource(`${process.env.NEXT_PUBLIC_API_URL || 'https://koola10.fly.dev'}/events/stream`);

    eventSource.onmessage = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      if (data.type === 'jarvis_notification') {
        setActiveNotification({
          alert_id: data.alert_id,
          title: data.title,
          body: data.message
        });
        setIsListening(true);
        setConversation([]);
      } else if (data.type === 'jarvis_analysis') {
        setConversation([{ role: 'assistant', content: data.analysis }]);
        setIsListening(true); // Keep listening for "handle it"
      }
    };

    return () => eventSource.close();
  }, []);

  const handleProceed = async (id: string) => {
    setActiveNotification(null);
    setConversation([{ role: 'assistant', content: 'Analyzing situation... one moment.' }]);

    // Trigger /voice/confirm with "proceed"
    await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'https://koola10.fly.dev'}/voice/confirm`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ alert_id: id, text: 'proceed' })
    });
  };

  const handleDismiss = async (id: string) => {
    setActiveNotification(null);
    setIsListening(false);
    setConversation([]);

    await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'https://koola10.fly.dev'}/voice/confirm`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ alert_id: id, text: 'dismiss' })
    });
  };

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
          <AgentStatus />
          <CollaborationTimeline />
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

      {/* Jarvis System */}
      <NotificationPanel
        notification={activeNotification}
        onProceed={handleProceed}
        onDismiss={handleDismiss}
      />
      <VoiceListener
        isActive={isListening}
        onTranscription={() => {}}
        onFinal={async (text) => {
          if (text === 'timeout') {
            setIsListening(false);
            if (activeNotification) {
              handleDismiss(activeNotification.alert_id);
            }
            return;
          }

          const lower = text.toLowerCase();
          const id = activeNotification?.alert_id || (conversation.length > 0 ? conversation[0].alert_id : null);

          if (activeNotification) {
            if (lower.includes('proceed') || lower.includes('yes') || lower.includes('tell me more')) {
              handleProceed(id);
            } else if (lower.includes('dismiss') || lower.includes('not now') || lower.includes('ignore')) {
              handleDismiss(id);
            }
          } else if (conversation.length > 0) {
            // Confirmation for handling
            if (lower.includes('handle it') || lower.includes('do it') || lower.includes('yes') || lower.includes('proceed')) {
              await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'https://koola10.fly.dev'}/voice/confirm`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ alert_id: id + '_handle', text: lower })
              });
              setConversation(prev => [...prev, { role: 'user', content: text }, { role: 'assistant', content: 'Understood. Executing.' }]);
              setTimeout(() => {
                setConversation([]);
                setIsListening(false);
              }, 3000);
            }
          }
        }}
      />
      <ConversationPanel
        messages={conversation}
        isActive={conversation.length > 0}
      />
    </main>
  );
}
