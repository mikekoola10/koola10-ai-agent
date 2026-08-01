-- ============================================================
-- Mikekoola10Org / Koola10 Command: Supabase Schema
-- Run this in the Supabase SQL Editor (Project → SQL Editor → New query)
-- ============================================================

-- 1. public.users table — mirrors auth.users with our custom columns.
--    The trigger below auto-creates a row on each GitHub sign-up.
create table if not exists public.users (
  id                uuid references auth.users (id) on delete cascade primary key,
  email             text,
  stripe_customer_id text,
  created_at        timestamp with time zone default timezone('utc'::text, now()) not null
);

-- 2. public.subscriptions table — one row per Stripe subscription per user.
create table if not exists public.subscriptions (
  id                    uuid default uuid_generate_v4() primary key,
  user_id               uuid references public.users (id) on delete cascade not null,
  stripe_subscription_id text unique,
  status                text,        -- 'active' | 'trialing' | 'past_due' | 'canceled' | 'incomplete'
  price_id              text,
  current_period_end    timestamp with time zone,
  cancel_at_period_end  boolean default false,
  created_at            timestamp with time zone default timezone('utc'::text, now()) not null,
  updated_at            timestamp with time zone default timezone('utc'::text, now()) not null
);

create index if not exists subscriptions_user_id_idx on public.subscriptions (user_id);
create index if not exists subscriptions_status_idx on public.subscriptions (status);

-- 3. Trigger: when a user signs in via Supabase Auth, ensure they have
--    a public.users row. Idempotent — safe to re-run.
create or replace function public.handle_new_user()
returns trigger
security definer
set search_path = public
as $$
begin
  insert into public.users (id, email)
  values (new.id, new.email)
  on conflict (id) do update set email = excluded.email;
  return new;
end;
$$ language plpgsql;

drop trigger if exists on_auth_user_created on auth.users;
create trigger on_auth_user_created
  after insert on auth.users
  for each row execute procedure public.handle_new_user();

-- Make sure every existing auth.users has a public.users row.
insert into public.users (id, email)
select id, email from auth.users
on conflict (id) do nothing;

-- 4. Row Level Security — users can READ their own data only.
--    Writes are done exclusively by the service-role admin client
--    (the Day 5 Stripe webhook uses that).
alter table public.users enable row level security;
alter table public.subscriptions enable row level security;

drop policy if exists "users can read own profile" on public.users;
create policy "users can read own profile"
  on public.users for select
  using (auth.uid() = id);

drop policy if exists "users can read own subscription" on public.subscriptions;
create policy "users can read own subscription"
  on public.subscriptions for select
  using (auth.uid() = user_id);

-- ============================================================
-- Done. After this runs, the Day 2 providers.tsx will be able
-- to query subscriptions.isPro based on active+trialing rows
-- whose current_period_end is in the future.
-- ============================================================
