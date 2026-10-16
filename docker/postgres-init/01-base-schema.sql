-- 01-base-schema.sql
-- Base schema reflecting the live AutoSuq / my_cinema Postgres database

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Local-only stand-ins for what Supabase provides (roles, auth.uid(), auth.users).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN CREATE ROLE anon NOLOGIN; END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN CREATE ROLE authenticated NOLOGIN; END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN CREATE ROLE service_role NOLOGIN BYPASSRLS; END IF;
END $$;
CREATE SCHEMA IF NOT EXISTS auth;
GRANT USAGE ON SCHEMA auth, public TO anon, authenticated, service_role;
CREATE TABLE IF NOT EXISTS auth.users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT,
  raw_user_meta_data JSONB
);
CREATE OR REPLACE FUNCTION auth.uid() RETURNS uuid LANGUAGE sql STABLE AS $$
  SELECT nullif(current_setting('request.jwt.claims', true)::jsonb ->> 'sub', '')::uuid;
$$;
-- Supabase's default: API roles get table privileges on public; RLS does the filtering.
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO anon, authenticated, service_role;

-- Users
CREATE TABLE IF NOT EXISTS public.users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  full_name TEXT,
  avatar_url TEXT,
  role TEXT NOT NULL DEFAULT 'owner' CHECK (role IN ('admin', 'owner')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Vending machines
CREATE TABLE IF NOT EXISTS public.vending_machine (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  location TEXT NOT NULL,
  latitude NUMERIC NOT NULL,
  longitude NUMERIC NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  owner_id UUID REFERENCES public.users(id),
  archived_at TIMESTAMPTZ
);

-- Vending machine items (catalog)
CREATE TABLE IF NOT EXISTS public.vending_machine_items (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  price NUMERIC NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  image_url TEXT,
  created_by UUID REFERENCES public.users(id)
);

-- Slots (machine to items)
CREATE TABLE IF NOT EXISTS public.vending_machine_to_items (
  id SERIAL PRIMARY KEY,
  machine_id UUID NOT NULL REFERENCES public.vending_machine(id) ON DELETE CASCADE,
  item_id INT NOT NULL REFERENCES public.vending_machine_items(id) ON DELETE CASCADE,
  amount INT NOT NULL,
  motor_code TEXT NOT NULL DEFAULT 'motor1',
  capacity INT NOT NULL DEFAULT 10
);

-- Transactions
CREATE TABLE IF NOT EXISTS public.transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  combination_id INT NOT NULL REFERENCES public.vending_machine_to_items(id),
  price NUMERIC NOT NULL,
  status BOOLEAN NOT NULL DEFAULT false,
  tex_ref TEXT NOT NULL,
  amount INT DEFAULT 1,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- User phone numbers
CREATE TABLE IF NOT EXISTS public.user_phone_numbers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
  phone TEXT NOT NULL CHECK (phone ~ '^\+251(9|7)\d{8}$'),
  label TEXT,
  is_primary BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, phone)
);

-- Contact support
CREATE TABLE IF NOT EXISTS public.contact_support (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.users(id) ON DELETE CASCADE,
  vending_machine_id UUID REFERENCES public.vending_machine(id) ON UPDATE SET NULL ON DELETE CASCADE,
  subject TEXT NOT NULL,
  message TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Machine order requests
CREATE TABLE IF NOT EXISTS public.machine_order_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
  notes TEXT,
  contact_phone TEXT,
  status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'contacted', 'closed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Helper functions and triggers
CREATE OR REPLACE FUNCTION public.set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  new.updated_at = now();
  RETURN new;
END;
$$;

DROP TRIGGER IF EXISTS set_users_updated_at ON public.users;
CREATE TRIGGER set_users_updated_at
  BEFORE UPDATE ON public.users
  FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

DROP TRIGGER IF EXISTS set_vending_machine_updated_at ON public.vending_machine;
CREATE TRIGGER set_vending_machine_updated_at
  BEFORE UPDATE ON public.vending_machine
  FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

DROP TRIGGER IF EXISTS set_contact_support_updated_at ON public.contact_support;
CREATE TRIGGER set_contact_support_updated_at
  BEFORE UPDATE ON public.contact_support
  FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();

-- Existing decrement_item_amount trigger on transactions (as verified in live DB)
CREATE OR REPLACE FUNCTION public.decrement_item_amount()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path TO 'public'
AS $$
BEGIN
  IF OLD.status IS FALSE AND NEW.status IS TRUE THEN
    UPDATE public.vending_machine_to_items
    SET amount = amount - COALESCE(NEW.amount, 1)
    WHERE id = NEW.combination_id;
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS update_item_amount_on_success ON public.transactions;
CREATE TRIGGER update_item_amount_on_success
  AFTER UPDATE ON public.transactions
  FOR EACH ROW
  WHEN (((old.status IS DISTINCT FROM new.status) AND (new.status IS TRUE)))
  EXECUTE FUNCTION public.decrement_item_amount();
