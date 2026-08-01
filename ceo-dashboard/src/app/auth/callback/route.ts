/**
 * OAuth callback handler — exchanges the GitHub OAuth ?code for a session cookie.
 * After Supabase sets the cookies, redirect to the home page.
 */

import { NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get("code");
  const next = searchParams.get("next") ?? "/";

  if (code) {
    const supabase = await createClient();
    const { error } = await supabase.auth.exchangeCodeForSession(code);
    if (!error) {
      return NextResponse.redirect(`${origin}${next}`);
    }
    // eslint-disable-next-line no-console
    console.error("[auth/callback] exchangeCodeForSession failed", error);
  }

  // Bad code or error — bounce back to home with a flag.
  return NextResponse.redirect(`${origin}/?auth_error=1`);
}
