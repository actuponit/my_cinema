-- wallet_v1_test.sql
-- Self-contained test transaction that asserts all invariants and rolls back.

BEGIN;

DO $$
DECLARE
  v_user_id     uuid := gen_random_uuid();
  v_other_user  uuid := gen_random_uuid();
  v_machine_id  uuid := gen_random_uuid();
  v_other_mach  uuid := gen_random_uuid();
  v_item1_id    int;
  v_item2_id    int;
  v_slot1_id    int;
  v_slot2_id    int;
  v_other_slot  int;
  v_bal         numeric;
  v_ledger_cnt  int;
  v_topup_rec   record;
  v_pay_rec     record;
  v_chapa_rec   record;
  v_cred_rec    record;
  v_stock       int;
BEGIN
  -- Setup test data
  INSERT INTO public.users (id, email, full_name, role)
  VALUES (v_user_id, 'wallet_test@example.com', 'Wallet Tester', 'customer');

  INSERT INTO public.users (id, email, full_name, role)
  VALUES (v_other_user, 'other_test@example.com', 'Other Tester', 'owner');

  INSERT INTO public.vending_machine (id, name, location, latitude, longitude, owner_id)
  VALUES (v_machine_id, 'Machine Test', 'Lobby', 9.0, 38.0, v_other_user);

  INSERT INTO public.vending_machine (id, name, location, latitude, longitude, owner_id)
  VALUES (v_other_mach, 'Other Machine', 'Outside', 9.0, 38.0, v_other_user);

  INSERT INTO public.vending_machine_items (name, price)
  VALUES ('Item A', 10.00) RETURNING id INTO v_item1_id;

  INSERT INTO public.vending_machine_items (name, price)
  VALUES ('Item B', 25.50) RETURNING id INTO v_item2_id;

  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code, capacity)
  VALUES (v_machine_id, v_item1_id, 3, 'motor1', 10) RETURNING id INTO v_slot1_id;

  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code, capacity)
  VALUES (v_machine_id, v_item2_id, 1, 'motor2', 10) RETURNING id INTO v_slot2_id;

  INSERT INTO public.vending_machine_to_items (machine_id, item_id, amount, motor_code, capacity)
  VALUES (v_other_mach, v_item1_id, 5, 'motor3', 10) RETURNING id INTO v_other_slot;

  INSERT INTO public.wallets (user_id, balance) VALUES (v_user_id, 0);

  INSERT INTO public.machine_credentials (user_id, username, password_hash)
  VALUES (v_user_id, 'user01', '$2a$10$abcdefghijklmnopqrstuvwxyz123456');

  -- 1. wallet_apply_deposit('dep_1', u, 'awaiting_transfer', null): topup row exists, balance 0.
  PERFORM * FROM public.wallet_apply_deposit('dep_1', v_user_id, 'awaiting_transfer', null);
  SELECT balance INTO v_bal FROM public.wallets WHERE user_id = v_user_id;
  ASSERT v_bal = 0, 'Test 1 Failed: Balance should be 0';
  ASSERT EXISTS (SELECT 1 FROM public.wallet_topups WHERE deposit_id = 'dep_1' AND status = 'awaiting_transfer'),
    'Test 1 Failed: Topup row not found or status mismatch';

  -- 2. wallet_apply_deposit('dep_1', u, 'succeeded', 50) twice: balance 50.00 (credited once), one ledger row.
  PERFORM * FROM public.wallet_apply_deposit('dep_1', v_user_id, 'succeeded', 50.00);
  PERFORM * FROM public.wallet_apply_deposit('dep_1', v_user_id, 'succeeded', 50.00);
  SELECT balance INTO v_bal FROM public.wallets WHERE user_id = v_user_id;
  ASSERT v_bal = 50.00, 'Test 2 Failed: Balance should be 50.00';
  SELECT count(*) INTO v_ledger_cnt FROM public.wallet_ledger WHERE user_id = v_user_id AND kind = 'topup';
  ASSERT v_ledger_cnt = 1, 'Test 2 Failed: Should have exactly 1 ledger row';

  -- 3. wallet_apply_deposit('dep_1', u, 'failed', null) after success: status stays succeeded.
  PERFORM * FROM public.wallet_apply_deposit('dep_1', v_user_id, 'failed', null);
  SELECT * INTO v_topup_rec FROM public.wallet_topups WHERE deposit_id = 'dep_1';
  ASSERT v_topup_rec.status = 'succeeded', 'Test 3 Failed: Status should stay succeeded';

  -- 4. wallet_pay(u, m, '[{"slot_id":S1,"quantity":2}]', 'tx_a'): SUCCESS, total 20.00, balance 30.00,
  --    motor_codes has 2 entries, slot stock 1, transactions row with price = 10.00, amount = 2, status = true.
  SELECT * INTO v_pay_rec FROM public.wallet_pay(
    v_user_id,
    v_machine_id,
    jsonb_build_array(jsonb_build_object('slot_id', v_slot1_id, 'quantity', 2)),
    'tx_a'
  );
  ASSERT v_pay_rec.status = 'SUCCESS', 'Test 4 Failed: Expected SUCCESS, got ' || COALESCE(v_pay_rec.status, 'null');
  ASSERT v_pay_rec.total = 20.00, 'Test 4 Failed: Expected total 20.00';
  ASSERT v_pay_rec.balance = 30.00, 'Test 4 Failed: Expected balance 30.00';
  ASSERT v_pay_rec.motor_codes = 'motor1,motor1', 'Test 4 Failed: Expected motor1,motor1, got ' || COALESCE(v_pay_rec.motor_codes, 'null');

  SELECT amount INTO v_stock FROM public.vending_machine_to_items WHERE id = v_slot1_id;
  ASSERT v_stock = 1, 'Test 4 Failed: Slot stock should be 1';
  ASSERT EXISTS (SELECT 1 FROM public.transactions WHERE tex_ref = 'tx_a' AND price = 10.00 AND amount = 2 AND status = true),
    'Test 4 Failed: Transactions row not found with status true, amount 2';

  -- 5. wallet_pay(..., S2 qty 2) returns OUT_OF_STOCK with no changes.
  --    S2 qty 1 + S1 qty 1 (total 35.50 > 30) returns INSUFFICIENT_BALANCE with no changes.
  SELECT * INTO v_pay_rec FROM public.wallet_pay(
    v_user_id,
    v_machine_id,
    jsonb_build_array(jsonb_build_object('slot_id', v_slot2_id, 'quantity', 2)),
    'tx_b'
  );
  ASSERT v_pay_rec.status = 'OUT_OF_STOCK', 'Test 5 Failed: Expected OUT_OF_STOCK';

  SELECT * INTO v_pay_rec FROM public.wallet_pay(
    v_user_id,
    v_machine_id,
    jsonb_build_array(
      jsonb_build_object('slot_id', v_slot1_id, 'quantity', 1),
      jsonb_build_object('slot_id', v_slot2_id, 'quantity', 1)
    ),
    'tx_c'
  );
  ASSERT v_pay_rec.status = 'INSUFFICIENT_BALANCE', 'Test 5 Failed: Expected INSUFFICIENT_BALANCE';
  ASSERT v_pay_rec.total = 35.50, 'Test 5 Failed: Total should be 35.50';
  ASSERT v_pay_rec.balance = 30.00, 'Test 5 Failed: Balance should remain 30.00';

  -- 6. A slot of a different machine returns INVALID_ITEMS. Quantity 0 returns INVALID_ITEMS.
  SELECT * INTO v_pay_rec FROM public.wallet_pay(
    v_user_id,
    v_machine_id,
    jsonb_build_array(jsonb_build_object('slot_id', v_other_slot, 'quantity', 1)),
    'tx_d'
  );
  ASSERT v_pay_rec.status = 'INVALID_ITEMS', 'Test 6 Failed: Expected INVALID_ITEMS for different machine';

  SELECT * INTO v_pay_rec FROM public.wallet_pay(
    v_user_id,
    v_machine_id,
    jsonb_build_array(jsonb_build_object('slot_id', v_slot1_id, 'quantity', 0)),
    'tx_e'
  );
  ASSERT v_pay_rec.status = 'INVALID_ITEMS', 'Test 6 Failed: Expected INVALID_ITEMS for quantity 0';

  -- 7. 5 × machine_login_failed(u): locked_until is about now + 15 min, failed_attempts = 0.
  FOR i IN 1..5 LOOP
    PERFORM * FROM public.machine_login_failed(v_user_id);
  END LOOP;
  SELECT * INTO v_cred_rec FROM public.machine_credentials WHERE user_id = v_user_id;
  ASSERT v_cred_rec.failed_attempts = 0, 'Test 7 Failed: failed_attempts should be reset to 0';
  ASSERT v_cred_rec.locked_until > now() + interval '14 minutes', 'Test 7 Failed: locked_until not in ~15 mins';

  -- 8. chapa_mark_paid on a seeded status=false basket returns rows and decrements stock.
  --    A second call returns 0 rows.
  INSERT INTO public.transactions (combination_id, price, amount, status, tex_ref)
  VALUES (v_slot1_id, 10.00, 1, false, 'tx_chapa');

  SELECT count(*) INTO v_ledger_cnt FROM public.chapa_mark_paid('tx_chapa');
  ASSERT v_ledger_cnt = 1, 'Test 8 Failed: chapa_mark_paid first call should return 1 row';
  SELECT amount INTO v_stock FROM public.vending_machine_to_items WHERE id = v_slot1_id;
  ASSERT v_stock = 0, 'Test 8 Failed: Stock should be decremented to 0 by trigger';

  SELECT count(*) INTO v_ledger_cnt FROM public.chapa_mark_paid('tx_chapa');
  ASSERT v_ledger_cnt = 0, 'Test 8 Failed: chapa_mark_paid second call should return 0 rows';

  RAISE NOTICE 'All SQL wallet_v1 assertions PASSED!';
END;
$$;

ROLLBACK;
