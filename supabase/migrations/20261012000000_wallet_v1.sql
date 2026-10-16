-- Migration: wallet_v1
-- Applies roles, wallet tables, constraints, security, and wallet/chapa functions.

-- ============================================================================
-- 1. Roles & Security Fixes
-- ============================================================================

ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE public.users
  ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'owner', 'customer'));

-- Prevent authenticated role self-escalation on role column
REVOKE UPDATE ON public.users FROM anon, authenticated;
GRANT UPDATE (full_name, avatar_url) ON public.users TO authenticated;

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
BEGIN
  INSERT INTO public.users (id, email, full_name, avatar_url, role)
  VALUES (
    new.id,
    new.email,
    new.raw_user_meta_data ->> 'full_name',
    new.raw_user_meta_data ->> 'avatar_url',
    CASE WHEN new.raw_user_meta_data ->> 'app' = 'wallet' THEN 'customer' ELSE 'owner' END
  )
  ON CONFLICT (id) DO NOTHING;
  RETURN new;
END;
$$;

CREATE OR REPLACE FUNCTION public.has_role(p_role text)
RETURNS boolean
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT EXISTS (SELECT 1 FROM public.users WHERE id = auth.uid() AND role = p_role);
$$;

CREATE OR REPLACE FUNCTION public.is_admin()
RETURNS boolean
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT EXISTS (SELECT 1 FROM public.users WHERE id = auth.uid() AND role = 'admin');
$$;

-- Protect owner tables from customer writes
DROP POLICY IF EXISTS "items_insert_own_or_admin" ON public.vending_machine_items;
DROP POLICY IF EXISTS "items_insert_owner_or_admin" ON public.vending_machine_items;
CREATE POLICY "items_insert_owner_or_admin"
  ON public.vending_machine_items FOR INSERT
  WITH CHECK ((created_by = auth.uid() AND public.has_role('owner')) OR public.is_admin());

DROP POLICY IF EXISTS "order_requests_insert_own" ON public.machine_order_requests;
DROP POLICY IF EXISTS "order_requests_insert_owner" ON public.machine_order_requests;
CREATE POLICY "order_requests_insert_owner"
  ON public.machine_order_requests FOR INSERT
  WITH CHECK (user_id = auth.uid() AND public.has_role('owner'));


-- ============================================================================
-- 2. Tables
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.wallets (
  user_id    UUID PRIMARY KEY REFERENCES public.users(id) ON DELETE RESTRICT,
  balance    NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.wallet_topups (
  id               UUID PRIMARY KEY,
  user_id          UUID NOT NULL REFERENCES public.wallets(user_id),
  deposit_id       TEXT NOT NULL UNIQUE,
  requested_amount NUMERIC(12,2) CHECK (requested_amount > 0),
  credited_amount  NUMERIC(12,2),
  status           TEXT NOT NULL CHECK (status IN (
                     'created','assigned','awaiting_transfer','reference_submitted',
                     'verification_pending','review_required',
                     'succeeded','failed','expired','cancelled')),
  checkout_url     TEXT,
  expires_at       TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS wallet_topups_user_created_idx ON public.wallet_topups (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS wallet_topups_open_idx ON public.wallet_topups (created_at)
  WHERE status NOT IN ('succeeded','failed','expired','cancelled');

CREATE TABLE IF NOT EXISTS public.wallet_ledger (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES public.wallets(user_id),
  kind          TEXT NOT NULL CHECK (kind IN ('topup','purchase')),
  amount        NUMERIC(12,2) NOT NULL,
  balance_after NUMERIC(12,2) NOT NULL CHECK (balance_after >= 0),
  topup_id      UUID UNIQUE REFERENCES public.wallet_topups(id),
  tex_ref       TEXT UNIQUE,
  machine_id    UUID REFERENCES public.vending_machine(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (
    (kind = 'topup'    AND amount > 0 AND topup_id IS NOT NULL AND tex_ref IS NULL) OR
    (kind = 'purchase' AND amount < 0 AND tex_ref IS NOT NULL AND topup_id IS NULL)
  )
);
CREATE INDEX IF NOT EXISTS wallet_ledger_user_created_idx ON public.wallet_ledger (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.machine_credentials (
  user_id         UUID PRIMARY KEY REFERENCES public.wallets(user_id) ON DELETE CASCADE,
  username        TEXT NOT NULL UNIQUE CHECK (username ~ '^[A-Za-z0-9]{4,6}$'),
  password_hash   TEXT NOT NULL,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_until    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.wallet_pay_result (
  status      TEXT NOT NULL,
  tex_ref     TEXT,
  total       NUMERIC, -- unconstrained: wallet_pay's early returns yield plain numeric
  balance     NUMERIC,
  motor_codes TEXT
);


-- ============================================================================
-- 3. Grants and RLS
-- ============================================================================

REVOKE ALL ON public.wallets, public.wallet_topups, public.wallet_ledger,
              public.machine_credentials, public.wallet_pay_result
  FROM anon, authenticated;

GRANT SELECT ON public.wallets, public.wallet_topups, public.wallet_ledger TO authenticated;
GRANT SELECT (user_id, username, locked_until, updated_at) ON public.machine_credentials TO authenticated;

ALTER TABLE public.wallets             ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.wallet_topups       ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.wallet_ledger       ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.machine_credentials ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.wallet_pay_result   ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS wallets_select_own ON public.wallets;
CREATE POLICY wallets_select_own ON public.wallets
  FOR SELECT TO authenticated USING (user_id = (SELECT auth.uid()));

DROP POLICY IF EXISTS wallet_topups_select_own ON public.wallet_topups;
CREATE POLICY wallet_topups_select_own ON public.wallet_topups
  FOR SELECT TO authenticated USING (user_id = (SELECT auth.uid()));

DROP POLICY IF EXISTS wallet_ledger_select_own ON public.wallet_ledger;
CREATE POLICY wallet_ledger_select_own ON public.wallet_ledger
  FOR SELECT TO authenticated USING (user_id = (SELECT auth.uid()));

DROP POLICY IF EXISTS machine_credentials_select_own ON public.machine_credentials;
CREATE POLICY machine_credentials_select_own ON public.machine_credentials
  FOR SELECT TO authenticated USING (user_id = (SELECT auth.uid()));


-- ============================================================================
-- 4. Database Functions
-- ============================================================================

-- wallet_onboard
CREATE OR REPLACE FUNCTION public.wallet_onboard(p_user_id uuid)
RETURNS SETOF public.wallets
LANGUAGE plpgsql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_role text;
BEGIN
  SELECT role INTO v_role FROM public.users WHERE id = p_user_id FOR UPDATE;

  IF v_role = 'owner'
     AND NOT EXISTS (SELECT 1 FROM public.vending_machine        WHERE owner_id   = p_user_id)
     AND NOT EXISTS (SELECT 1 FROM public.vending_machine_items  WHERE created_by = p_user_id)
     AND NOT EXISTS (SELECT 1 FROM public.machine_order_requests WHERE user_id    = p_user_id)
     AND NOT EXISTS (SELECT 1 FROM public.contact_support        WHERE user_id    = p_user_id)
     AND NOT EXISTS (SELECT 1 FROM public.user_phone_numbers     WHERE user_id    = p_user_id)
  THEN
    UPDATE public.users SET role = 'customer', updated_at = now() WHERE id = p_user_id;
    v_role := 'customer';
  END IF;

  IF v_role IS DISTINCT FROM 'customer' THEN
    RETURN;
  END IF;

  INSERT INTO public.wallets (user_id) VALUES (p_user_id) ON CONFLICT DO NOTHING;
  RETURN QUERY SELECT * FROM public.wallets WHERE user_id = p_user_id;
END;
$$;

-- wallet_apply_deposit
CREATE OR REPLACE FUNCTION public.wallet_apply_deposit(
  p_deposit_id text,
  p_user_id    uuid,
  p_status     text,
  p_amount     numeric
)
RETURNS SETOF public.wallet_topups
LANGUAGE plpgsql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_topup   public.wallet_topups;
  v_balance numeric(12,2);
BEGIN
  SELECT * INTO v_topup FROM public.wallet_topups WHERE deposit_id = p_deposit_id FOR UPDATE;

  IF NOT FOUND THEN
    INSERT INTO public.wallets (user_id) VALUES (p_user_id) ON CONFLICT DO NOTHING;
    INSERT INTO public.wallet_topups (id, user_id, deposit_id, status)
    VALUES (gen_random_uuid(), p_user_id, p_deposit_id, p_status)
    RETURNING * INTO v_topup;
  END IF;

  IF v_topup.user_id <> p_user_id THEN
    RAISE EXCEPTION 'deposit % belongs to a different user', p_deposit_id;
  END IF;

  IF v_topup.status = 'succeeded' THEN
    RETURN NEXT v_topup;
    RETURN;
  END IF;

  IF p_status = 'succeeded' THEN
    IF p_amount IS NULL OR p_amount <= 0 THEN
      RAISE EXCEPTION 'succeeded deposit % has invalid amount %', p_deposit_id, p_amount;
    END IF;

    UPDATE public.wallets
       SET balance = balance + p_amount, updated_at = now()
     WHERE user_id = p_user_id
    RETURNING balance INTO v_balance;

    INSERT INTO public.wallet_ledger (user_id, kind, amount, balance_after, topup_id)
    VALUES (p_user_id, 'topup', p_amount, v_balance, v_topup.id);

    UPDATE public.wallet_topups
       SET status = 'succeeded', credited_amount = p_amount, updated_at = now()
     WHERE id = v_topup.id
    RETURNING * INTO v_topup;
  ELSE
    UPDATE public.wallet_topups
       SET status = p_status, updated_at = now()
     WHERE id = v_topup.id
    RETURNING * INTO v_topup;
  END IF;

  RETURN NEXT v_topup;
END;
$$;

-- wallet_pay
CREATE OR REPLACE FUNCTION public.wallet_pay(
  p_user_id    uuid,
  p_machine_id uuid,
  p_items      jsonb,
  p_tex_ref    text
)
RETURNS SETOF public.wallet_pay_result
LANGUAGE plpgsql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_balance     numeric(12,2);
  v_total       numeric(12,2) := 0;
  v_motor_codes text := '';
  r             record;
  v_slot        record;
BEGIN
  IF p_items IS NULL OR jsonb_typeof(p_items) <> 'array' OR jsonb_array_length(p_items) = 0 THEN
    RETURN QUERY SELECT 'INVALID_ITEMS'::text, null::text, null::numeric, null::numeric, null::text;
    RETURN;
  END IF;

  SELECT balance INTO v_balance FROM public.wallets WHERE user_id = p_user_id FOR UPDATE;
  IF NOT FOUND THEN
    RETURN QUERY SELECT 'NO_WALLET'::text, null::text, null::numeric, null::numeric, null::text;
    RETURN;
  END IF;

  UPDATE public.machine_credentials
     SET failed_attempts = 0, locked_until = null, updated_at = now()
   WHERE user_id = p_user_id AND (failed_attempts <> 0 OR locked_until IS NOT NULL);

  FOR r IN
    SELECT (e ->> 'slot_id')::int AS slot_id, sum((e ->> 'quantity')::int) AS qty
      FROM jsonb_array_elements(p_items) e
     GROUP BY 1
     ORDER BY 1
  LOOP
    IF r.qty IS NULL OR r.qty <= 0 THEN
      RETURN QUERY SELECT 'INVALID_ITEMS'::text, null::text, null::numeric, null::numeric, null::text;
      RETURN;
    END IF;

    SELECT s.amount AS stock, i.price
      INTO v_slot
      FROM public.vending_machine_to_items s
      JOIN public.vending_machine_items i ON i.id = s.item_id
     WHERE s.id = r.slot_id AND s.machine_id = p_machine_id
       FOR UPDATE OF s;

    IF NOT FOUND THEN
      RETURN QUERY SELECT 'INVALID_ITEMS'::text, null::text, null::numeric, null::numeric, null::text;
      RETURN;
    END IF;
    IF v_slot.stock < r.qty THEN
      RETURN QUERY SELECT 'OUT_OF_STOCK'::text, null::text, null::numeric, null::numeric, null::text;
      RETURN;
    END IF;

    v_total := v_total + v_slot.price * r.qty;
  END LOOP;

  IF v_balance < v_total THEN
    RETURN QUERY SELECT 'INSUFFICIENT_BALANCE'::text, null::text, v_total, v_balance, null::text;
    RETURN;
  END IF;

  UPDATE public.wallets
     SET balance = balance - v_total, updated_at = now()
   WHERE user_id = p_user_id
  RETURNING balance INTO v_balance;

  INSERT INTO public.wallet_ledger (user_id, kind, amount, balance_after, tex_ref, machine_id)
  VALUES (p_user_id, 'purchase', -v_total, v_balance, p_tex_ref, p_machine_id);

  FOR r IN
    SELECT (e ->> 'slot_id')::int AS slot_id, sum((e ->> 'quantity')::int) AS qty
      FROM jsonb_array_elements(p_items) e
     GROUP BY 1
     ORDER BY 1
  LOOP
    INSERT INTO public.transactions (combination_id, price, amount, status, tex_ref, created_at, updated_at)
    SELECT s.id, i.price, r.qty, true, p_tex_ref, now(), now()
      FROM public.vending_machine_to_items s
      JOIN public.vending_machine_items i ON i.id = s.item_id
     WHERE s.id = r.slot_id;

    UPDATE public.vending_machine_to_items
       SET amount = amount - r.qty
     WHERE id = r.slot_id
    RETURNING motor_code INTO v_slot;

    IF coalesce(v_slot.motor_code, '') <> '' THEN
      v_motor_codes := concat_ws(',', nullif(v_motor_codes, ''),
                                 array_to_string(array_fill(v_slot.motor_code, array[r.qty::int]), ','));
    END IF;
  END LOOP;

  RETURN QUERY SELECT 'SUCCESS'::text, p_tex_ref, v_total, v_balance, v_motor_codes;
END;
$$;

-- machine_login_failed
CREATE OR REPLACE FUNCTION public.machine_login_failed(p_user_id uuid)
RETURNS SETOF public.machine_credentials
LANGUAGE sql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
  UPDATE public.machine_credentials
     SET failed_attempts = CASE WHEN failed_attempts + 1 >= 5 THEN 0 ELSE failed_attempts + 1 END,
         locked_until    = CASE WHEN failed_attempts + 1 >= 5 THEN now() + interval '15 minutes' ELSE locked_until END,
         updated_at      = now()
   WHERE user_id = p_user_id
  RETURNING *;
$$;

-- chapa_mark_paid
-- Note: decrement_item_amount trigger on public.transactions handles stock decrement
-- when status changes from false -> true, so we omit manual decrement here to prevent double-decrement.
CREATE OR REPLACE FUNCTION public.chapa_mark_paid(p_tex_ref text)
RETURNS SETOF public.transactions
LANGUAGE sql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
  UPDATE public.transactions
     SET status = true, updated_at = now()
   WHERE tex_ref = p_tex_ref AND status = false
  RETURNING *;
$$;

REVOKE EXECUTE ON FUNCTION
  public.wallet_onboard(uuid),
  public.wallet_apply_deposit(text, uuid, text, numeric),
  public.wallet_pay(uuid, uuid, jsonb, text),
  public.machine_login_failed(uuid),
  public.chapa_mark_paid(text)
FROM public, anon, authenticated;
