/**
 * Server-side Supabase client with cookie-based session (Next.js App Router).
 * Use this in Server Components, route handlers, and Server Actions.
 */

import { createServerClient } from "@supabase/ssr";
import type { CookieOptions } from "@supabase/ssr";
import { cookies } from "next/headers";

/** Same validation as the browser client — see lib/supabase/client.ts. */
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

  return { url, anonKey };
}

export async function createClient() {
  const { url, anonKey } = getSupabaseEnv();
  const cookieStore = await cookies();

  // Never hand supabase-js an empty/invalid URL — it throws "Invalid
  // supabaseUrl". Fall back to a placeholder so construction always succeeds;
  // calls on it will fail auth (i.e. "not signed in"), which routes handle
  // gracefully via the authConfigured gate.
  const configured = isValidSupabaseUrl(url) && anonKey.length > 0;
  const resolvedUrl = configured ? url : "https://placeholder.supabase.co";
  const resolvedKey = configured ? anonKey : "placeholder-anon-key";

  return createServerClient(resolvedUrl, resolvedKey, {
    cookies: {
      getAll() {
        return cookieStore.getAll();
      },
      setAll(toSet: { name: string; value: string; options: CookieOptions }[]) {
        try {
          toSet.forEach(({ name, value, options }) => {
            cookieStore.set(name, value, options);
          });
        } catch {
          // Server Components cannot mutate cookies; middleware handles refresh.
        }
      },
    },
  });
}
