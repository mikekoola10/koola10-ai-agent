/**
 * Browser-side Supabase client (uses anon key, safe to expose).
 * Reads env vars in priority order: NEXT_PUBLIC_* → MIKEKOOLA10ORG_* → plain *.
 */

import { createBrowserClient } from "@supabase/ssr";

/**
 * supabase-js throws "Invalid supabaseUrl: Must be a valid HTTP or HTTPS URL"
 * when handed anything that isn't a real http(s) URL. A misconfigured env
 * value (placeholder, typo, missing protocol) would therefore crash build-time
 * prerendering — so we validate before ever constructing a client.
 */
export function isValidSupabaseUrl(url: string | undefined): boolean {
  if (!url) return false;
  try {
    const parsed = new URL(url);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

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

  const configured = isValidSupabaseUrl(url) && anonKey.length > 0;

  if (!configured) {
    if (typeof window !== "undefined") {
      // eslint-disable-next-line no-console
      console.warn(
        "[Supabase] missing or invalid NEXT_PUBLIC_SUPABASE_URL / NEXT_PUBLIC_SUPABASE_ANON_KEY (or MIKEKOOLA10ORG_* equivalents). Auth disabled until configured."
      );
    }
  }

  // Treat invalid values as unconfigured so createBrowserClient never receives
  // a non-http(s) string (which would crash SSR/prerender during builds).
  return { url: configured ? url : "", anonKey: configured ? anonKey : "" };
}

export function isSupabaseConfigured(): boolean {
  const { url, anonKey } = getSupabaseEnv();
  return Boolean(url && anonKey);
}

export function createClient() {
  const { url, anonKey } = getSupabaseEnv();
  if (!url || !anonKey) {
    // No valid keys configured yet. createBrowserClient throws on empty
    // strings, which would crash SSR/prerender during builds without env
    // vars. The AuthProvider gates every call behind `authConfigured`, so
    // this client is never actually used until real keys exist.
    return createBrowserClient(
      "https://placeholder.supabase.co",
      "placeholder-anon-key",
    );
  }
  return createBrowserClient(url, anonKey);
}
