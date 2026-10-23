-- Fixes from the kiosk/web review (2026-10-22).
-- 1. set_primary_phone: no anon execute (it needs auth.uid(); anon always failed anyway).
-- 2. Archived machines stop taking payments, even on a kiosk that cached their id.
-- 3. kiosk_submit_support: max 10 tickets per machine per hour (anon can call it).

-- ============================================================================
-- 1. set_primary_phone
-- ============================================================================
REVOKE EXECUTE ON FUNCTION public.set_primary_phone(uuid) FROM PUBLIC, anon;
GRANT  EXECUTE ON FUNCTION public.set_primary_phone(uuid) TO authenticated;

-- ============================================================================
-- 2. Archived machines
-- ============================================================================
CREATE OR REPLACE FUNCTION public.kiosk_machine_items(p_machine_id uuid)
RETURNS TABLE (slot_id int, item_id int, name text, price numeric, image_url text, stock int, capacity int)
LANGUAGE sql STABLE
SECURITY DEFINER SET search_path = public
AS $$
  SELECT s.id, i.id, i.name, i.price, i.image_url, s.amount, s.capacity
    FROM public.vending_machine_to_items s
    JOIN public.vending_machine_items i ON i.id = s.item_id
    JOIN public.vending_machine m ON m.id = s.machine_id AND m.archived_at IS NULL
   WHERE s.machine_id = p_machine_id
   ORDER BY i.name;
$$;

CREATE OR REPLACE FUNCTION public.chapa_create_basket(p_machine_id uuid, p_items jsonb, p_tex_ref text)
RETURNS TABLE(status text, total numeric)
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_total numeric := 0;
  r       record;
  v_slot  record;
BEGIN
  IF p_items IS NULL OR jsonb_typeof(p_items) <> 'array' OR jsonb_array_length(p_items) = 0
     OR NOT EXISTS (SELECT 1 FROM public.vending_machine WHERE id = p_machine_id AND archived_at IS NULL) THEN
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

CREATE OR REPLACE FUNCTION public.wallet_pay(p_user_id uuid, p_machine_id uuid, p_items jsonb, p_tex_ref text)
RETURNS SETOF wallet_pay_result
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  v_balance     numeric(12,2);
  v_total       numeric(12,2) := 0;
  v_motor_codes text := '';
  r             record;
  v_slot        record;
BEGIN
  IF p_items IS NULL OR jsonb_typeof(p_items) <> 'array' OR jsonb_array_length(p_items) = 0
     OR NOT EXISTS (SELECT 1 FROM public.vending_machine WHERE id = p_machine_id AND archived_at IS NULL) THEN
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

-- ============================================================================
-- 3. kiosk_submit_support rate limit
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_contact_support_machine_created
  ON public.contact_support (vending_machine_id, created_at);

CREATE OR REPLACE FUNCTION public.kiosk_submit_support(p_machine_id uuid, p_subject text, p_message text)
RETURNS uuid
LANGUAGE plpgsql
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
  -- ponytail: per-machine bucket; anyone with the anon key can still fill one machine's quota.
  IF (SELECT count(*) FROM public.contact_support
       WHERE vending_machine_id = p_machine_id AND created_at > now() - interval '1 hour') >= 10 THEN
    RAISE EXCEPTION 'too many support requests from this machine, try again later' USING ERRCODE = '54000';
  END IF;

  INSERT INTO public.contact_support (vending_machine_id, subject, message)
  VALUES (p_machine_id, trim(p_subject), trim(p_message))
  RETURNING id INTO v_id;
  RETURN v_id;
END;
$$;
