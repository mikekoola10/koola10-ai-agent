# Supabase env vars — flexible naming

The Mikekoola10Org / Koola10 Command Supabase client libraries try these env-var names **in order**; the first match wins. This lets you paste values under whichever convention the Freebuff API Keys tab supports (next.js-style `NEXT_PUBLIC_*`, the brand-prefixed `MIKEKOOLA10ORG_*`, or plain names).

## Required for Day 2 (Authentication)

| Priority | Name | Purpose |
| --- | --- | --- |
| 1 | `NEXT_PUBLIC_SUPABASE_URL` | Project URL (client + server) |
| 2 | `MIKEKOOLA10ORG_SUPABASE_URL` | Brand-prefixed fallback |
| 3 | `SUPABASE_URL` | Plain fallback |

Same priority chain for `*_SUPABASE_ANON_KEY`.

## Required for Day 5 (Stripe webhook = server-only)

| Priority | Name |
| --- | --- |
| 1 | `SUPABASE_SERVICE_ROLE_KEY` |
| 2 | `MIKEKOOLA10ORG_SUPABASE_SERVICE_ROLE_KEY` |

NOTE: service-role key is **never** `NEXT_PUBLIC_*`. It bypasses RLS.

## Required for Day 3 (Stripe checkout)

| Priority | Name |
| --- | --- |
| 1 | `STRIPE_SECRET_KEY` |
| 2 | `MIKEKOOLA10ORG_STRIPE_SECRET_KEY` |
| 1 | `STRIPE_WEBHOOK_SECRET` |
| 2 | `MIKEKOOLA10ORG_STRIPE_WEBHOOK_SECRET` |
| 1 | `STRIPE_PRICE_ID` |
| 2 | `MIKEKOOLA10ORG_STRIPE_PRICE_ID` |

## How to paste these into Freebuff

1. Open the project's API Keys tab.
2. Add each key as `KEY` / `VALUE` pair. Freebuff injects `process.env.KEY` at runtime.
3. The code's fallback chain in `lib/supabase/{client,server,admin}.ts` resolves whichever name you used.
4. After saving, restart the dev server. The dashboard's "Auth not configured" warning will disappear and the "Sign in with GitHub" button will become active.

## Freebuff CLI equivalents (if you prefer)

```sh
freebuff-env set --file .env.local '{"NEXT_PUBLIC_SUPABASE_URL":"https://your-project.supabase.co"}'
freebuff-env set --file .env.local '{"NEXT_PUBLIC_SUPABASE_ANON_KEY":"eyJhbGc..."}'
freebuff-env set --file .env.local '{"SUPABASE_SERVICE_ROLE_KEY":"eyJhbGc..."}'
freebuff-env set --file .env.local '{"STRIPE_SECRET_KEY":"sk_test_..."}'
freebuff-env set --file .env.local '{"STRIPE_WEBHOOK_SECRET":"whsec_..."}'
freebuff-env set --file .env.local '{"STRIPE_PRICE_ID":"price_..."}'
```
