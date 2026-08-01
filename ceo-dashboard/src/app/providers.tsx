"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import type { SupabaseClient, User } from "@supabase/supabase-js";
import { createClient } from "@/lib/supabase/client";

type AuthState = {
  user: User | null;
  isPro: boolean;
  isLoading: boolean;
  authConfigured: boolean;
  signInWithGithub: () => Promise<void>;
  signOut: () => Promise<void>;
  refreshProStatus: () => Promise<void>;
};

const AuthContext = createContext<AuthState | undefined>(undefined);

async function fetchProStatus(supabase: SupabaseClient, userId: string) {
  const { data, error } = await supabase
    .from("subscriptions")
    .select("status, current_period_end")
    .eq("user_id", userId)
    .in("status", ["active", "trialing"])
    .order("current_period_end", { ascending: false })
    .maybeSingle();

  if (error) {
    // eslint-disable-next-line no-console
    console.warn("[auth] pro-status query failed (RLS? no table yet?)", error);
    return false;
  }
  if (!data) return false;
  return new Date(data.current_period_end).getTime() > Date.now();
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isPro, setIsPro] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const supabase = useMemo(() => createClient(), []);
  const authConfigured = Boolean(
    process.env.NEXT_PUBLIC_SUPABASE_URL ||
      process.env.MIKEKOOLA10ORG_SUPABASE_URL ||
      process.env.SUPABASE_URL,
  );

  const refreshProStatus = useCallback(async () => {
    if (!user) {
      setIsPro(false);
      return;
    }
    setIsPro(await fetchProStatus(supabase, user.id));
  }, [supabase, user]);

  useEffect(() => {
    if (!authConfigured) {
      setIsLoading(false);
      return;
    }

    let mounted = true;

    const init = async () => {
      const {
        data: { user: current },
      } = await supabase.auth.getUser();

      if (!mounted) return;
      setUser(current);
      if (current) {
        setIsPro(await fetchProStatus(supabase, current.id));
      }
      setIsLoading(false);
    };

    init();

    const { data: sub } = supabase.auth.onAuthStateChange((_event, session) => {
      setUser(session?.user ?? null);
      if (session?.user) {
        setIsPro(false); // optimistic; refresh below
        void fetchProStatus(supabase, session.user.id).then(setIsPro);
      } else {
        setIsPro(false);
      }
    });

    return () => {
      mounted = false;
      sub.subscription.unsubscribe();
    };
  }, [supabase, authConfigured]);

  const signInWithGithub = useCallback(async () => {
    if (!authConfigured) return;
    const origin =
      typeof window !== "undefined" ? window.location.origin : "";
    await supabase.auth.signInWithOAuth({
      provider: "github",
      options: { redirectTo: `${origin}/auth/callback` },
    });
  }, [supabase, authConfigured]);

  const signOut = useCallback(async () => {
    if (!authConfigured) return;
    await supabase.auth.signOut();
  }, [supabase, authConfigured]);

  const value = useMemo<AuthState>(
    () => ({
      user,
      isPro,
      isLoading,
      authConfigured,
      signInWithGithub,
      signOut,
      refreshProStatus,
    }),
    [user, isPro, isLoading, authConfigured, signInWithGithub, signOut, refreshProStatus],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used inside <AuthProvider>");
  }
  return ctx;
}
