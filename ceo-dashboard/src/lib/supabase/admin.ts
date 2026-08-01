/**
 * Supabase admin client — uses SERVICE_ROLE key, bypasses RLS.
 * Server-only. Treats secrets as STRICTLY confidential.
 * Use only from server-side code paths (Next.js API routes, Server Actions).
 */

import { createClient } from "@supabase/supabase-js";

function getServiceKey() {
  return (
    process.env.SUPABASE_SERVICE_ROLE_KEY ||
    process.env.MIKEKOOLA10ORG_SUPABASE_SERVICE_ROLE_KEY ||
    ""
  );
}

function getUrl() {
  return (
    process.env.NEXT_PUBLIC_SUPABASE_URL ||
    process.env.MIKEKOOLA10ORG_SUPABASE_URL ||
    process.env.SUPABASE_URL ||
    ""
  );
}

export function createAdminClient() {
  const url = getUrl();
  const serviceKey = getServiceKey();

  if (!url || !serviceKey) {
    throw new Error(
      "[Supabase admin] missing SUPABASE_SERVICE_ROLE_KEY or MIKEKOOLA10ORG_SUPABASE_SERVICE_ROLE_KEY"
    );
  }

  return createClient(url, serviceKey, {
    auth: { persistSession: false, autoRefreshToken: false },
  });
}
