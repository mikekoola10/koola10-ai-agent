/**
 * Browser-side Supabase client (uses anon key, safe to expose).
 * Reads env vars in priority order: NEXT_PUBLIC_* → MIKEKOOLA10ORG_* → plain *.
 */

import { createBrowserClient } from "@supabase/ssr";

export function getSupabaseEnv() {
  const url =
    process.env.NEXT_PUBLIC_SUPABASE_URL ||
    process.env.MIKEKOOLA10ORG_SUPABASE_URL ||
    process.env.SUPABASE_URL ||
    "";

  const anonKey =
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY ||
    process.env.MIKEKOOLA10ORG_SUPABASE_ANON_KEY ||
    process.env.SUPABASE_ANON_KEY ||
    "";

  if (!url || !anonKey) {
    if (typeof window !== "undefined") {
      // eslint-disable-next-line no-console
      console.warn(
        "[Supabase] missing NEXT_PUBLIC_SUPABASE_URL / NEXT_PUBLIC_SUPABASE_ANON_KEY (or MIKEKOOLA10ORG_* equivalents). Auth disabled until configured."
      );
    }
  }

  return { url, anonKey };
}

export function createClient() {
  const { url, anonKey } = getSupabaseEnv();
  if (!url || !anonKey) {
    // No keys configured yet. createBrowserClient throws on empty strings,
    // which would crash SSR/prerender during builds without env vars. The
    // AuthProvider gates every call behind `authConfigured`, so this client
    // is never actually used until real keys exist.
    return createBrowserClient(
      "https://placeholder.supabase.co",
      "placeholder-anon-key",
    );
  }
  return createBrowserClient(url, anonKey);
}
