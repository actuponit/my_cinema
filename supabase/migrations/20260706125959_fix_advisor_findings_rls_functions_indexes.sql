-- =============================================================
-- 1. Function hardening
-- =============================================================

-- handle_new_user is an auth trigger; nobody should call it via the API.
revoke execute on function public.handle_new_user() from public, anon, authenticated;

-- is_admin is needed by RLS policies evaluated as `authenticated`,
-- but anon never needs it (all policies below are scoped TO authenticated).
revoke execute on function public.is_admin() from public, anon;
grant execute on function public.is_admin() to authenticated;

-- Pin search_path on flagged functions (role-mutable search_path lint).
alter function public.set_current_timestamp_updated_at() set search_path = public;
alter function public.set_updated_at() set search_path = public;
alter function public.decrement_item_amount() set search_path = public;

-- =============================================================
-- 2. RLS policies: wrap auth.uid()/is_admin() in scalar subqueries
--    (initplan fix), scope TO authenticated, split FOR ALL policies,
--    merge duplicate UPDATE policies on users.
-- =============================================================

-- ---------- users ----------
drop policy users_select_own_or_admin on public.users;
drop policy users_update_admin on public.users;
drop policy users_update_own on public.users;

create policy users_select_own_or_admin on public.users
  for select to authenticated
  using (id = (select auth.uid()) or (select public.is_admin()));

create policy users_update_own_or_admin on public.users
  for update to authenticated
  using (id = (select auth.uid()) or (select public.is_admin()))
  with check (id = (select auth.uid()) or (select public.is_admin()));

-- ---------- user_phone_numbers ----------
drop policy phones_select_own_or_admin on public.user_phone_numbers;
drop policy phones_modify_own on public.user_phone_numbers;

create policy phones_select_own_or_admin on public.user_phone_numbers
  for select to authenticated
  using (user_id = (select auth.uid()) or (select public.is_admin()));

create policy phones_insert_own on public.user_phone_numbers
  for insert to authenticated
  with check (user_id = (select auth.uid()));

create policy phones_update_own on public.user_phone_numbers
  for update to authenticated
  using (user_id = (select auth.uid()))
  with check (user_id = (select auth.uid()));

create policy phones_delete_own on public.user_phone_numbers
  for delete to authenticated
  using (user_id = (select auth.uid()));

-- ---------- vending_machine ----------
drop policy machines_select_own_or_admin on public.vending_machine;
drop policy machines_insert_own_or_admin on public.vending_machine;
drop policy machines_update_own_or_admin on public.vending_machine;
drop policy machines_delete_admin_only on public.vending_machine;

create policy machines_select_own_or_admin on public.vending_machine
  for select to authenticated
  using (owner_id = (select auth.uid()) or (select public.is_admin()));

create policy machines_insert_own_or_admin on public.vending_machine
  for insert to authenticated
  with check (owner_id = (select auth.uid()) or (select public.is_admin()));

create policy machines_update_own_or_admin on public.vending_machine
  for update to authenticated
  using (owner_id = (select auth.uid()) or (select public.is_admin()));

create policy machines_delete_admin_only on public.vending_machine
  for delete to authenticated
  using ((select public.is_admin()));

-- ---------- vending_machine_items ----------
drop policy items_select_authenticated on public.vending_machine_items;
drop policy items_insert_own_or_admin on public.vending_machine_items;
drop policy items_update_own_or_admin on public.vending_machine_items;
drop policy items_delete_own_or_admin on public.vending_machine_items;

create policy items_select_authenticated on public.vending_machine_items
  for select to authenticated
  using (true);

create policy items_insert_own_or_admin on public.vending_machine_items
  for insert to authenticated
  with check (created_by = (select auth.uid()) or (select public.is_admin()));

create policy items_update_own_or_admin on public.vending_machine_items
  for update to authenticated
  using (created_by = (select auth.uid()) or (select public.is_admin()))
  with check (created_by = (select auth.uid()) or (select public.is_admin()));

create policy items_delete_own_or_admin on public.vending_machine_items
  for delete to authenticated
  using (created_by = (select auth.uid()) or (select public.is_admin()));

-- ---------- vending_machine_to_items (slots) ----------
drop policy slots_select_own_or_admin on public.vending_machine_to_items;
drop policy slots_write_own_or_admin on public.vending_machine_to_items;

create policy slots_select_own_or_admin on public.vending_machine_to_items
  for select to authenticated
  using (
    (select public.is_admin())
    or exists (
      select 1 from public.vending_machine vm
      where vm.id = vending_machine_to_items.machine_id
        and vm.owner_id = (select auth.uid())
    )
  );

create policy slots_insert_own_or_admin on public.vending_machine_to_items
  for insert to authenticated
  with check (
    (select public.is_admin())
    or exists (
      select 1 from public.vending_machine vm
      where vm.id = vending_machine_to_items.machine_id
        and vm.owner_id = (select auth.uid())
    )
  );

create policy slots_update_own_or_admin on public.vending_machine_to_items
  for update to authenticated
  using (
    (select public.is_admin())
    or exists (
      select 1 from public.vending_machine vm
      where vm.id = vending_machine_to_items.machine_id
        and vm.owner_id = (select auth.uid())
    )
  )
  with check (
    (select public.is_admin())
    or exists (
      select 1 from public.vending_machine vm
      where vm.id = vending_machine_to_items.machine_id
        and vm.owner_id = (select auth.uid())
    )
  );

create policy slots_delete_own_or_admin on public.vending_machine_to_items
  for delete to authenticated
  using (
    (select public.is_admin())
    or exists (
      select 1 from public.vending_machine vm
      where vm.id = vending_machine_to_items.machine_id
        and vm.owner_id = (select auth.uid())
    )
  );

-- ---------- transactions ----------
drop policy transactions_select_own_or_admin on public.transactions;

create policy transactions_select_own_or_admin on public.transactions
  for select to authenticated
  using (
    (select public.is_admin())
    or exists (
      select 1
      from public.vending_machine_to_items vmi
      join public.vending_machine vm on vm.id = vmi.machine_id
      where vmi.id = transactions.combination_id
        and vm.owner_id = (select auth.uid())
    )
  );

-- ---------- contact_support ----------
drop policy support_insert_own on public.contact_support;
drop policy support_select_own_or_admin on public.contact_support;
drop policy support_update_own_or_admin on public.contact_support;

create policy support_insert_own on public.contact_support
  for insert to authenticated
  with check (user_id = (select auth.uid()));

create policy support_select_own_or_admin on public.contact_support
  for select to authenticated
  using (user_id = (select auth.uid()) or (select public.is_admin()));

create policy support_update_own_or_admin on public.contact_support
  for update to authenticated
  using (user_id = (select auth.uid()) or (select public.is_admin()))
  with check (user_id = (select auth.uid()) or (select public.is_admin()));

-- ---------- machine_order_requests ----------
drop policy order_requests_insert_own on public.machine_order_requests;
drop policy order_requests_select_own_or_admin on public.machine_order_requests;
drop policy order_requests_update_admin_only on public.machine_order_requests;

create policy order_requests_insert_own on public.machine_order_requests
  for insert to authenticated
  with check (user_id = (select auth.uid()));

create policy order_requests_select_own_or_admin on public.machine_order_requests
  for select to authenticated
  using (user_id = (select auth.uid()) or (select public.is_admin()));

create policy order_requests_update_admin_only on public.machine_order_requests
  for update to authenticated
  using ((select public.is_admin()));

-- =============================================================
-- 3. Covering indexes for unindexed foreign keys
-- =============================================================
create index if not exists idx_contact_support_user_id on public.contact_support (user_id);
create index if not exists idx_machine_order_requests_user_id on public.machine_order_requests (user_id);
create index if not exists idx_transactions_combination_id on public.transactions (combination_id);
create index if not exists idx_vending_machine_owner_id on public.vending_machine (owner_id);
create index if not exists idx_vmti_item_id on public.vending_machine_to_items (item_id);
create index if not exists idx_vmti_machine_id on public.vending_machine_to_items (machine_id);;
