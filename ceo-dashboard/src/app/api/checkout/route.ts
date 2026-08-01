/**
 * POST /api/checkout
 * Creates a Stripe Checkout Session for the $29/mo Pro subscription.
 * Requires a signed-in user (session cookie). Returns { url } to redirect to.
 */

import { NextResponse } from "next/server";
import Stripe from "stripe";
import { createClient } from "@/lib/supabase/server";

function getStripeSecretKey() {
  return (
    process.env.STRIPE_SECRET_KEY ||
    process.env.MIKEKOOLA10ORG_STRIPE_SECRET_KEY ||
    ""
  );
}

function getAppUrl() {
  return process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000";
}

export async function POST() {
  const secretKey = getStripeSecretKey();
  if (!secretKey) {
    return NextResponse.json(
      { error: "Stripe is not configured. Add STRIPE_SECRET_KEY to the project keys." },
      { status: 500 },
    );
  }

  // 1. Make sure the caller is signed in.
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user?.email) {
    return NextResponse.json(
      { error: "You must be signed in to upgrade." },
      { status: 401 },
    );
  }

  // 2. Also record the Stripe customer id on the users row (idempotent).
  await supabase
    .from("users")
    .update({ email: user.email })
    .eq("id", user.id);

  // 3. Create the Checkout Session for a recurring monthly subscription.
  const stripe = new Stripe(secretKey);
  const appUrl = getAppUrl();

  try {
    const session = await stripe.checkout.sessions.create({
      mode: "subscription",
      customer_email: user.email,
      client_reference_id: user.id,
      line_items: [
        {
          price_data: {
            currency: "usd",
            unit_amount: 2900, // $29.00
            recurring: { interval: "month" },
            product_data: {
              name: "Koola10 Command Pro",
              description: "Unlock the full Autonomous Swarm Intelligence dashboard",
            },
          },
          quantity: 1,
        },
      ],
      metadata: { userId: user.id },
      success_url: `${appUrl}/?checkout=success`,
      cancel_url: `${appUrl}/?checkout=canceled`,
      subscription_data: { metadata: { userId: user.id } },
    });

    return NextResponse.json({ url: session.url });
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error("[checkout] Stripe session creation failed", err);
    return NextResponse.json(
      { error: "Could not start checkout. Check STRIPE_SECRET_KEY validity." },
      { status: 500 },
    );
  }
}
