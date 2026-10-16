-- Migration: supabase_cutover
-- Replaces what Hasura + the Go server did: Chapa basket pricing, wallet-pay
-- throttling, wallet onboarding, and the kiosk's anonymous read/write access.

-- ============================================================================
-- 1. Chapa: price the basket server-side and store it unpaid
-- ============================================================================

-- Stores one unpaid transactions row per slot (unit price, summed quantity).
-- chapa-webhook later re-totals these rows against Chapa and marks them paid.
CREATE OR REPLACE FUNCTION public.chapa_create_basket(
  p_machine_id uuid,
  p_items      jsonb,
  p_tex_ref    text
)
RETURNS TABLE (status text, total numeric)
LANGUAGE plpgsql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_total numeric := 0;
  r       record;
  v_slot  record;
BEGIN
  IF p_items IS NULL OR jsonb_typeof(p_items) <> 'array' OR jsonb_array_length(p_items) = 0 THEN
    RETURN QUERY SELECT 'INVALID_ITEMS'::text, null::numeric;
    RETURN;
  END IF;

  -- Validate everything before inserting, so a bad basket leaves no rows.
  FOR r IN
    SELECT (e ->> 'slot_id')::int AS slot_id, sum((e ->> 'quantity')::int) AS qty
      FROM jsonb_array_elements(p_items) e
     GROUP BY 1
  LOOP
    IF r.qty IS NULL OR r.qty <= 0 THEN
      RETURN QUERY SELECT 'INVALID_ITEMS'::text, null::numeric;
      RETURN;
    END IF;

    SELECT s.amount AS stock, i.price
      INTO v_slot
      FROM public.vending_machine_to_items s
      JOIN public.vending_machine_items i ON i.id = s.item_id
     WHERE s.id = r.slot_id AND s.machine_id = p_machine_id;

    IF NOT FOUND THEN
      RETURN QUERY SELECT 'INVALID_ITEMS'::text, null::numeric;
      RETURN;
    END IF;
    IF v_slot.stock < r.qty THEN
      RETURN QUERY SELECT 'OUT_OF_STOCK'::text, null::numeric;
      RETURN;
    END IF;

    v_total := v_total + v_slot.price * r.qty;
  END LOOP;

  INSERT INTO public.transactions (combination_id, price, amount, status, tex_ref, created_at, updated_at)
  SELECT s.id, i.price, g.qty, false, p_tex_ref, now(), now()
    FROM (SELECT (e ->> 'slot_id')::int AS slot_id, sum((e ->> 'quantity')::int) AS qty
            FROM jsonb_array_elements(p_items) e
           GROUP BY 1) g
    JOIN public.vending_machine_to_items s ON s.id = g.slot_id
    JOIN public.vending_machine_items i ON i.id = s.item_id;

  RETURN QUERY SELECT 'OK'::text, v_total;
END;
$$;


-- ============================================================================
-- 2. wallet-pay: global wrong-credential throttle (was in-memory in Go)
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.wallet_pay_failures (
  at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS wallet_pay_failures_at_idx ON public.wallet_pay_failures (at);
REVOKE ALL ON public.wallet_pay_failures FROM anon, authenticated;
ALTER TABLE public.wallet_pay_failures ENABLE ROW LEVEL SECURITY;

-- ponytail: one global bucket like the Go version (30/min); per-machine buckets
-- if a single attacker locking out every kiosk for a minute becomes a problem.
CREATE OR REPLACE FUNCTION public.wallet_pay_throttled()
RETURNS boolean
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT count(*) >= 30 FROM public.wallet_pay_failures WHERE at > now() - interval '1 minute';
$$;

CREATE OR REPLACE FUNCTION public.wallet_pay_record_failure()
RETURNS void
LANGUAGE sql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
  DELETE FROM public.wallet_pay_failures WHERE at <= now() - interval '1 minute';
  INSERT INTO public.wallet_pay_failures DEFAULT VALUES;
$$;


-- ============================================================================
-- 3. Wallet onboarding, called by the signed-in wallet app via supabase.rpc()
-- ============================================================================

-- Zero rows = owner account (can't use the wallet). Otherwise the wallet and
-- the kiosk username, if one is set.
CREATE OR REPLACE FUNCTION public.wallet_onboard_me()
RETURNS TABLE (balance numeric, username text)
LANGUAGE sql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT w.balance, c.username
    FROM public.wallet_onboard(auth.uid()) w
    LEFT JOIN public.machine_credentials c ON c.user_id = w.user_id;
$$;


-- ============================================================================
-- 4. Kiosk API (anon). Hasura's "anonymous" role read whole tables, including
--    every owner's phone numbers and every transaction. These expose only what
--    the kiosk screens need.
-- ============================================================================

CREATE OR REPLACE FUNCTION public.kiosk_machines()
RETURNS TABLE (id uuid, name text, location text, latitude numeric, longitude numeric)
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT m.id, m.name, m.location, m.latitude, m.longitude
    FROM public.vending_machine m
   WHERE m.archived_at IS NULL
   ORDER BY m.name;
$$;

CREATE OR REPLACE FUNCTION public.kiosk_machine_items(p_machine_id uuid)
RETURNS TABLE (slot_id int, item_id int, name text, price numeric, image_url text, stock int, capacity int)
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT s.id, i.id, i.name, i.price, i.image_url, s.amount, s.capacity
    FROM public.vending_machine_to_items s
    JOIN public.vending_machine_items i ON i.id = s.item_id
   WHERE s.machine_id = p_machine_id
   ORDER BY i.name;
$$;

-- Replaces GetMachineOwner + GetSupportPhoneNumbers: only this machine's owner's numbers.
CREATE OR REPLACE FUNCTION public.kiosk_support_phones(p_machine_id uuid)
RETURNS TABLE (phone text, label text, is_primary boolean)
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT p.phone, p.label, p.is_primary
    FROM public.vending_machine m
    JOIN public.user_phone_numbers p ON p.user_id = m.owner_id
   WHERE m.id = p_machine_id
   ORDER BY p.is_primary DESC, p.created_at;
$$;

-- user_id stays null (filed anonymously at the kiosk), same as under Hasura.
CREATE OR REPLACE FUNCTION public.kiosk_submit_support(p_machine_id uuid, p_subject text, p_message text)
RETURNS uuid
LANGUAGE plpgsql VOLATILE
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_id uuid;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM public.vending_machine WHERE id = p_machine_id) THEN
    RAISE EXCEPTION 'unknown machine' USING ERRCODE = '22023';
  END IF;
  IF length(trim(coalesce(p_subject, ''))) NOT BETWEEN 1 AND 200
     OR length(trim(coalesce(p_message, ''))) NOT BETWEEN 1 AND 2000 THEN
    RAISE EXCEPTION 'subject must be 1-200 and message 1-2000 characters' USING ERRCODE = '22023';
  END IF;

  INSERT INTO public.contact_support (vending_machine_id, subject, message)
  VALUES (p_machine_id, trim(p_subject), trim(p_message))
  RETURNING id INTO v_id;
  RETURN v_id;
END;
$$;

-- true = paid, false = waiting, null = unknown tx_ref.
CREATE OR REPLACE FUNCTION public.kiosk_payment_status(p_tex_ref text)
RETURNS boolean
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT bool_and(t.status) FROM public.transactions t WHERE t.tex_ref = p_tex_ref;
$$;


-- ============================================================================
-- 5. Execute grants (Postgres grants EXECUTE to PUBLIC by default)
-- ============================================================================

REVOKE EXECUTE ON FUNCTION
  public.chapa_create_basket(uuid, jsonb, text),
  public.wallet_pay_throttled(),
  public.wallet_pay_record_failure()
FROM public, anon, authenticated;

REVOKE EXECUTE ON FUNCTION public.wallet_onboard_me() FROM public, anon;
GRANT  EXECUTE ON FUNCTION public.wallet_onboard_me() TO authenticated;

REVOKE EXECUTE ON FUNCTION
  public.kiosk_machines(),
  public.kiosk_machine_items(uuid),
  public.kiosk_support_phones(uuid),
  public.kiosk_submit_support(uuid, text, text),
  public.kiosk_payment_status(text)
FROM public;
GRANT EXECUTE ON FUNCTION
  public.kiosk_machines(),
  public.kiosk_machine_items(uuid),
  public.kiosk_support_phones(uuid),
  public.kiosk_submit_support(uuid, text, text),
  public.kiosk_payment_status(text)
TO anon, authenticated;
