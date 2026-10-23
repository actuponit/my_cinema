-- review_fixes_test.sql
-- Asserts 20261022000000_review_fixes and rolls back.

BEGIN;

DO $$
DECLARE
  v_owner    uuid := gen_random_uuid();
  v_customer uuid := gen_random_uuid();
  v_archived uuid := gen_random_uuid();
  v_item     int;
  v_slot     int;
  v_rec      record;
  v_raised   boolean := false;
BEGIN
  INSERT INTO public.users (id, email, role) VALUES (v_owner, 'owner@review.test', 'owner');
  INSERT INTO public.users (id, email, role) VALUES (v_customer, 'cust@review.test', 'customer');
  INSERT INTO public.wallets (user_id, balance) VALUES (v_customer, 100);
  INSERT INTO public.vending_machine (id, name, location, latitude, longitude, owner_id, archived_at)
  VALUES (v_archived, 'Old', 'Gone', 9.0, 38.0, v_owner, now());
  INSERT INTO public.vending_machine_items (name, price) VALUES ('Soda', 10) RETURNING id INTO v_item;
  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code)
  VALUES (v_archived, v_item, 5, 'review_m1') RETURNING id INTO v_slot;

  -- 1. set_primary_phone is not callable by anon
  ASSERT NOT has_function_privilege('anon', 'public.set_primary_phone(uuid)', 'EXECUTE'), 'anon can run set_primary_phone';
  ASSERT has_function_privilege('authenticated', 'public.set_primary_phone(uuid)', 'EXECUTE'), 'authenticated lost set_primary_phone';

  -- 2. Archived machines sell nothing
  ASSERT NOT EXISTS (SELECT 1 FROM public.kiosk_machine_items(v_archived)), 'archived machine still lists items';

  SELECT * INTO v_rec FROM public.chapa_create_basket(v_archived, jsonb_build_array(jsonb_build_object('slot_id', v_slot, 'quantity', 1)), 'review-chapa');
  ASSERT v_rec.status = 'INVALID_ITEMS', 'chapa basket on archived machine: ' || v_rec.status;
  ASSERT NOT EXISTS (SELECT 1 FROM public.transactions WHERE tex_ref = 'review-chapa'), 'archived basket left rows';

  SELECT * INTO v_rec FROM public.wallet_pay(v_customer, v_archived, jsonb_build_array(jsonb_build_object('slot_id', v_slot, 'quantity', 1)), 'review-wallet');
  ASSERT v_rec.status = 'INVALID_ITEMS', 'wallet_pay on archived machine: ' || v_rec.status;
  ASSERT (SELECT balance FROM public.wallets WHERE user_id = v_customer) = 100, 'archived wallet_pay charged';

  -- 3. Support rate limit: 10 per machine per hour
  FOR i IN 1..10 LOOP
    PERFORM public.kiosk_submit_support(v_archived, 'subj', 'msg');
  END LOOP;
  BEGIN
    PERFORM public.kiosk_submit_support(v_archived, 'subj', 'msg');
  EXCEPTION WHEN SQLSTATE '54000' THEN
    v_raised := true;
  END;
  ASSERT v_raised, '11th support request was not rate limited';

  RAISE NOTICE 'review_fixes_test: all assertions passed';
END $$;

ROLLBACK;
