/**
 * POST /api/webhook
 * Stripe webhook receiver. Verifies the signature, then keeps the
 * public.subscriptions table in sync so the dashboard shows Pro access.
 *
 * Events handled:
 *  - checkout.session.completed       → create subscription row (status active/trialing)
 *  - customer.subscription.updated    → update status / period end / cancel_at_period_end
 *  - customer.subscription.deleted    → mark canceled
 */

import { NextResponse } from "next/server";
import Stripe from "stripe";
import { createAdminClient } from "@/lib/supabase/admin";

function getStripeSecretKey() {
  return (
    process.env.STRIPE_SECRET_KEY ||
    process.env.MIKEKOOLA10ORG_STRIPE_SECRET_KEY ||
    ""
  );
}

function getWebhookSecret() {
  return (
    process.env.STRIPE_WEBHOOK_SECRET ||
    process.env.MIKEKOOLA10ORG_STRIPE_WEBHOOK_SECRET ||
    ""
  );
}

type SubRecord = {
  user_id: string;
  stripe_subscription_id: string;
  status: string;
  price_id?: string | null;
  current_period_end?: string | null;
  cancel_at_period_end?: boolean;
};

async function upsertSubscription(record: SubRecord) {
  const admin = createAdminClient();
  const { data: existing } = await admin
    .from("subscriptions")
    .select("id")
    .eq("stripe_subscription_id", record.stripe_subscription_id)
    .maybeSingle();

  if (existing?.id) {
    await admin
      .from("subscriptions")
      .update({
        status: record.status,
        price_id: record.price_id ?? null,
        current_period_end: record.current_period_end ?? null,
        cancel_at_period_end: record.cancel_at_period_end ?? false,
        updated_at: new Date().toISOString(),
      })
      .eq("id", existing.id);
  } else {
    await admin.from("subscriptions").insert({
      user_id: record.user_id,
      stripe_subscription_id: record.stripe_subscription_id,
      status: record.status,
      price_id: record.price_id ?? null,
      current_period_end: record.current_period_end ?? null,
      cancel_at_period_end: record.cancel_at_period_end ?? false,
    });
  }
}

export async function POST(request: Request) {
  const secretKey = getStripeSecretKey();
  const webhookSecret = getWebhookSecret();

  if (!secretKey || !webhookSecret) {
    // eslint-disable-next-line no-console
    console.error("[webhook] missing STRIPE_SECRET_KEY / STRIPE_WEBHOOK_SECRET");
    return NextResponse.json(
      { error: "Webhook not configured" },
      { status: 500 },
    );
  }

  const stripe = new Stripe(secretKey);
  const rawBody = await request.text();
  const signature = request.headers.get("stripe-signature");

  let event: Stripe.Event;
  try {
    event = stripe.webhooks.constructEvent(
      rawBody,
      signature ?? "",
      webhookSecret,
    );
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error("[webhook] signature verification failed", err);
    return NextResponse.json(
      { error: "Invalid signature" },
      { status: 400 },
    );
  }

  try {
    switch (event.type) {
      case "checkout.session.completed": {
        const session = event.data.object as Stripe.Checkout.Session;
        const userId = session.metadata?.userId ?? session.client_reference_id;
        const subId = String(session.subscription ?? "");
        if (userId && subId) {
          await upsertSubscription({
            user_id: userId,
            stripe_subscription_id: subId,
            status: "active",
          });
        }
        break;
      }

      case "customer.subscription.updated":
      case "customer.subscription.created": {
        const sub = event.data.object as Stripe.Subscription &
          { current_period_end?: number | null };
        const userId = sub.metadata?.userId;
        if (userId) {
          await upsertSubscription({
            user_id: userId,
            stripe_subscription_id: sub.id,
            status: sub.status, // active | trialing | past_due | canceled ...
            price_id: sub.items.data[0]?.price.id ?? null,
            current_period_end: sub.current_period_end
              ? new Date(sub.current_period_end * 1000).toISOString()
              : null,
            cancel_at_period_end: sub.cancel_at_period_end,
          });
        }
        break;
      }

      case "customer.subscription.deleted": {
        const sub = event.data.object as Stripe.Subscription;
        const userId = sub.metadata?.userId;
        if (userId) {
          await upsertSubscription({
            user_id: userId,
            stripe_subscription_id: sub.id,
            status: "canceled",
          });
        }
        break;
      }

      default:
        // Acknowledge unhandled events so Stripe stops retrying.
        break;
    }

    return NextResponse.json({ received: true });
  } catch (err) {
    // eslint-disable-next-line no-console
    console.error("[webhook] handler error", err);
    return NextResponse.json(
      { error: "Handler failure" },
      { status: 500 },
    );
  }
}
