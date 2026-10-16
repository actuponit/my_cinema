-- cutover_test.sql
-- Asserts the supabase_cutover functions and rolls back.

BEGIN;

DO $$
DECLARE
  v_owner    uuid := gen_random_uuid();
  v_customer uuid := gen_random_uuid();
  v_machine  uuid := gen_random_uuid();
  v_archived uuid := gen_random_uuid();
  v_other    uuid := gen_random_uuid();
  v_item     int;
  v_slot     int;
  v_slot2    int;
  v_foreign  int;
  v_rec      record;
  v_cnt      int;
BEGIN
  INSERT INTO public.users (id, email, role) VALUES (v_owner, 'owner@test', 'owner');
  INSERT INTO public.users (id, email, role) VALUES (v_customer, 'cust@test', 'customer');
  INSERT INTO public.vending_machine (id, name, location, latitude, longitude, owner_id)
  VALUES (v_machine, 'M1', 'Lobby', 9.0, 38.0, v_owner),
         (v_other, 'M2', 'Gate', 9.1, 38.1, v_owner);
  INSERT INTO public.vending_machine (id, name, location, latitude, longitude, owner_id, archived_at)
  VALUES (v_archived, 'Old', 'Gone', 9.0, 38.0, v_owner, now());
  INSERT INTO public.vending_machine_items (name, price) VALUES ('Soda', 12.50) RETURNING id INTO v_item;
  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code)
  VALUES (v_machine, v_item, 3, 'm1') RETURNING id INTO v_slot;
  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code)
  VALUES (v_machine, v_item, 1, 'm2') RETURNING id INTO v_slot2;
  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code)
  VALUES (v_other, v_item, 9, 'm3') RETURNING id INTO v_foreign;
  INSERT INTO public.user_phone_numbers (user_id, phone, is_primary) VALUES (v_owner, '+251911000000', true);

  -- 1. chapa_create_basket: duplicate slot lines are summed, priced from the DB, stored unpaid.
  SELECT * INTO v_rec FROM public.chapa_create_basket(v_machine,
    jsonb_build_array(jsonb_build_object('slot_id', v_slot, 'quantity', 1),
                      jsonb_build_object('slot_id', v_slot, 'quantity', 1),
                      jsonb_build_object('slot_id', v_slot2, 'quantity', 1)), 'chapa_1');
  ASSERT v_rec.status = 'OK' AND v_rec.total = 37.50, 'Test 1: expected OK 37.50, got ' || v_rec.status || ' ' || coalesce(v_rec.total::text, 'null');
  SELECT count(*) INTO v_cnt FROM public.transactions WHERE tex_ref = 'chapa_1' AND status = false;
  ASSERT v_cnt = 2, 'Test 1: expected 2 unpaid rows';
  ASSERT EXISTS (SELECT 1 FROM public.transactions WHERE tex_ref = 'chapa_1' AND combination_id = v_slot AND amount = 2 AND price = 12.50),
    'Test 1: slot row should have amount 2 at unit price';

  -- 2. Bad baskets leave no rows.
  SELECT * INTO v_rec FROM public.chapa_create_basket(v_machine,
    jsonb_build_array(jsonb_build_object('slot_id', v_slot, 'quantity', 1),
                      jsonb_build_object('slot_id', v_foreign, 'quantity', 1)), 'chapa_2');
  ASSERT v_rec.status = 'INVALID_ITEMS', 'Test 2: other machine slot should be INVALID_ITEMS';
  SELECT * INTO v_rec FROM public.chapa_create_basket(v_machine,
    jsonb_build_array(jsonb_build_object('slot_id', v_slot2, 'quantity', 2)), 'chapa_3');
  ASSERT v_rec.status = 'OUT_OF_STOCK', 'Test 2: expected OUT_OF_STOCK';
  ASSERT NOT EXISTS (SELECT 1 FROM public.transactions WHERE tex_ref IN ('chapa_2', 'chapa_3')), 'Test 2: no rows expected';

  -- 3. kiosk_payment_status: waiting, then paid via chapa_mark_paid; unknown is null.
  ASSERT public.kiosk_payment_status('chapa_1') = false, 'Test 3: should be waiting';
  PERFORM * FROM public.chapa_mark_paid('chapa_1');
  ASSERT public.kiosk_payment_status('chapa_1') = true, 'Test 3: should be paid';
  ASSERT public.kiosk_payment_status('nope') IS NULL, 'Test 3: unknown should be null';

  -- 4. wallet_pay throttle trips at 30 failures in a minute.
  ASSERT NOT public.wallet_pay_throttled(), 'Test 4: should start open';
  FOR i IN 1..30 LOOP PERFORM public.wallet_pay_record_failure(); END LOOP;
  ASSERT public.wallet_pay_throttled(), 'Test 4: should be throttled after 30';

  -- 5. wallet_onboard_me: customer gets a wallet, owner with machines gets nothing.
  PERFORM set_config('request.jwt.claims', json_build_object('sub', v_customer)::text, true);
  SELECT * INTO v_rec FROM public.wallet_onboard_me();
  ASSERT v_rec.balance = 0 AND v_rec.username IS NULL, 'Test 5: customer should get a 0 wallet';
  PERFORM set_config('request.jwt.claims', json_build_object('sub', v_owner)::text, true);
  SELECT count(*) INTO v_cnt FROM public.wallet_onboard_me();
  ASSERT v_cnt = 0, 'Test 5: owner should get no wallet';

  -- 6. Kiosk reads.
  ASSERT NOT EXISTS (SELECT 1 FROM public.kiosk_machines() WHERE id = v_archived), 'Test 6: archived machine hidden';
  ASSERT EXISTS (SELECT 1 FROM public.kiosk_machines() WHERE id = v_machine), 'Test 6: machine listed';
  SELECT count(*) INTO v_cnt FROM public.kiosk_machine_items(v_machine);
  ASSERT v_cnt = 2, 'Test 6: expected 2 slots';
  SELECT count(*) INTO v_cnt FROM public.kiosk_support_phones(v_machine);
  ASSERT v_cnt = 1, 'Test 6: expected owner phone';

  -- 7. kiosk_submit_support validates and inserts anonymously.
  PERFORM public.kiosk_submit_support(v_machine, 'Stuck', 'Item did not drop');
  ASSERT EXISTS (SELECT 1 FROM public.contact_support WHERE vending_machine_id = v_machine AND user_id IS NULL), 'Test 7: ticket inserted';
  BEGIN
    PERFORM public.kiosk_submit_support(v_machine, '', 'x');
    ASSERT false, 'Test 7: empty subject should raise';
  EXCEPTION WHEN invalid_parameter_value THEN NULL;
  END;

  RAISE NOTICE 'All SQL cutover assertions PASSED!';
END;
$$;

ROLLBACK;
