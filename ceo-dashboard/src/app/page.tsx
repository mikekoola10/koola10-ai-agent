"use client";

import { useEffect, useState } from "react";
import { HealthStatus } from "@/components/dashboard/HealthStatus";
import { FinancialHealth } from "@/components/dashboard/FinancialHealth";
import { RevenuePulse } from "@/components/dashboard/RevenuePulse";
import { AgentStatus } from "@/components/dashboard/AgentStatus";
import { DailyReport } from "@/components/dashboard/DailyReport";
import { QuickActions } from "@/components/dashboard/QuickActions";
import { NovaChat } from "@/components/chat/NovaChat";
import { Cpu, LogOut, Sparkles, Loader2 } from "lucide-react";
import { useAuth } from "./providers";

export default function Dashboard() {
  const { user, isPro, isLoading, signInWithGithub, signOut, authConfigured, refreshProStatus } = useAuth();
  const [upgrading, setUpgrading] = useState(false);
  const [upgradeError, setUpgradeError] = useState<string | null>(null);

  // Handle returning from Stripe Checkout (success banner). Runs client-side
  // only, so we read window.location instead of useSearchParams (which would
  // force a Suspense boundary at prerender time).
  useEffect(() => {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    if (params.get("checkout") === "success") {
      void refreshProStatus();
      const url = new URL(window.location.href);
      url.searchParams.delete("checkout");
      window.history.replaceState({}, "", url.toString());
    }
  }, [refreshProStatus]);

  const startCheckout = async () => {
    setUpgrading(true);
    setUpgradeError(null);
    try {
      const res = await fetch("/api/checkout", { method: "POST" });
      const data = await res.json();
      if (!res.ok || !data.url) {
        throw new Error(data.error ?? "Could not start checkout");
      }
      window.location.href = data.url;
    } catch (err) {
      setUpgradeError(
        err instanceof Error ? err.message : "Could not start checkout",
      );
      setUpgrading(false);
    }
  };

  if (isLoading) {
    return (
      <main className="min-h-screen flex items-center justify-center">
        <div className="text-white/50 font-mono text-xs uppercase tracking-[0.3em]">
          Authenticating…
        </div>
      </main>
    );
  }

  if (!user) {
    return (
      <main className="min-h-screen flex items-center justify-center p-4">
        <div className="max-w-md w-full text-center space-y-6 bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-12 shadow-2xl">
          <div className="flex justify-center">
            <div className="p-3 bg-amber-400 rounded-xl shadow-[0_0_20px_rgba(251,191,36,0.4)]">
              <Cpu className="h-10 w-10 text-black" />
            </div>
          </div>
          <div className="space-y-2">
            <h1 className="text-3xl font-black text-white tracking-tight">
              KOOLA10 <span className="text-amber-400">COMMAND</span>
            </h1>
            <p className="text-white/50 text-xs font-medium uppercase tracking-[0.2em]">
              Mikekoola10Org · Autonomous Swarm Intelligence
            </p>
          </div>
          <p className="text-white/70 text-sm">
            Sign in with GitHub to access the Autonomous Agent Swarm Control
            Center.
          </p>
          <button
            onClick={signInWithGithub}
            disabled={!authConfigured}
            className="w-full bg-amber-400 hover:bg-amber-300 disabled:bg-white/10 disabled:text-white/40 text-black font-bold py-3 px-6 rounded-xl transition-colors"
          >
            {authConfigured ? "Sign in with GitHub" : "Auth not configured"}
          </button>
          {!authConfigured && (
            <p className="text-white/30 text-xs leading-relaxed">
              Paste <code className="text-amber-300/60">SUPABASE_URL</code> &{" "}
              <code className="text-amber-300/60">SUPABASE_ANON_KEY</code> in
              the Freebuff API Keys tab to enable sign-in.
            </p>
          )}
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen p-4 md:p-8 max-w-7xl mx-auto space-y-8 relative">
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
              {user.email ?? "Mikekoola10Org"} ·{" "}
              {isPro ? "Pro Tier" : "Free Tier"}
            </p>
          </div>
        </div>
        <div className="flex flex-col items-stretch md:items-end gap-3">
          <div className="flex items-center gap-3">
            {!isPro && (
              <button
                onClick={startCheckout}
                disabled={upgrading}
                className="flex items-center gap-2 bg-amber-400 hover:bg-amber-300 disabled:opacity-60 text-black text-xs font-bold uppercase tracking-widest rounded-xl px-4 py-2 shadow-[0_0_16px_rgba(251,191,36,0.35)] transition-colors"
              >
                {upgrading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Sparkles className="h-4 w-4" />
                )}
                Upgrade to Pro
              </button>
            )}
            {isPro && (
              <span className="flex items-center gap-2 text-amber-300 text-xs font-bold uppercase tracking-widest bg-amber-400/10 border border-amber-400/30 rounded-xl px-4 py-2">
                <Sparkles className="h-4 w-4" />
                Pro Active
              </span>
            )}
            <QuickActions />
            <button
              onClick={signOut}
              className="flex items-center gap-2 text-white/60 hover:text-white text-xs font-medium uppercase tracking-widest bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl px-4 py-2 transition-colors"
            >
              <LogOut className="h-4 w-4" />
              Sign out
            </button>
          </div>
          {upgradeError && (
            <p className="text-red-400 text-xs max-w-xs md:text-right">
              {upgradeError}
            </p>
          )}
        </div>
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
