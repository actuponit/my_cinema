# Backend Spec: AutoSuq Wallet

Read [architecture.md](architecture.md) first for the why, and [rules.md](rules.md) for how code is written.
This file is the **what**: ordered tasks with acceptance criteria.

---

## Prompt for the implementing agent

> You are implementing the AutoSuq Wallet backend. Read `docs/architecture.md`,
> `docs/rules.md` and this file completely before changing anything.
>
> Repos (all under `~/Development/projects/`):
> - `my_cinema/server`: Go 1.22 / Gin backend (module `cinema-server`, entry `delivery/main.go`). Most of the work happens here.
> - `my_cinema/hasura`: Hasura metadata for the cloud instance `pleasing-viper-86.hasura.app`.
> - `vending_machine_app`: Flutter kiosk app (task T8).
> - `autosuq_admin`: owner admin dashboard (task T9, small).
> - `autosuq-web-app`: this repo. It holds the docs and is the web client, not part of the backend work.
>
> The database is the Supabase project **actuponit**. Use the Supabase MCP server for
> schema inspection and for applying migrations. Apply migrations to a **Supabase branch first** if
> branching is available, otherwise ask the user before touching production.
>
> Work through the tasks **in order** (T0 → T9). Each task lists acceptance criteria. Don't move
> on until they hold. Where this spec says **CONFIRM**, check the fact (live DB, provider docs) and
> adapt. If the fact contradicts the spec in a way that changes behavior, stop and ask the user.
>
> Constraints: don't commit, push or deploy. Don't rotate or print secrets. Don't change the
> MQTT topic or payload format (machine firmware depends on it). Keep the existing Chapa flow
> working. When done, report per task: what changed (files), test results, and open CONFIRM items.

---

## Glossary

| Term | Meaning |
|---|---|
| slot | Row in `vending_machine_to_items` (one motor). `ItemRequestInput.item_id` from the kiosk is a **slot id** |
| catalog item | Row in `vending_machine_items` (name, unit price) |
| basket | All `transactions` rows sharing one `tex_ref` |
| machine credential | Short username + password typed at the kiosk |
| deposit | A verify.et deposit (`data.id`), one per top-up attempt |

---

## T0. Preflight (read-only)

1. **Live schema vs reference.** Using the Supabase MCP, compare the live DB with
   `autosuq_admin/docs/01-schema-and-rls.sql` (§0 lists the base tables). Record:
   - columns and types of `users`, `vending_machine`, `vending_machine_items`,
     `vending_machine_to_items`, `transactions`;
   - whether `transactions.id` has a default (expected: `gen_random_uuid()`);
   - the name of the role CHECK constraint on `public.users` (expected `users_role_check`);
   - existing triggers on `transactions` and `vending_machine_to_items`
     (`select tgname, tgrelid::regclass from pg_trigger where not tgisinternal;`).
     If a trigger already decrements stock, **skip every "decrement stock" step below** and tell the user.
   - whether `pgcrypto` is installed (not required. Hashing happens in Go).
2. **Supabase JWT signing.** Fetch `https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json`.
   Non-empty `keys` means asymmetric signing keys, so use JWKS. Empty means a legacy HS256 secret, so use `SUPABASE_JWT_SECRET`.
3. **Hasura metadata drift.** `my_cinema/hasura/metadata/actions.graphql` defines
   `InitiateTransactionInput { id: Int! }`, but the kiosk schema (`vending_machine_app/lib/app/schema.graphql`)
   shows `{ items: [ItemRequestInput!]! }`, so the repo metadata is stale. Run
   `hasura metadata export --endpoint https://pleasing-viper-86.hasura.app --admin-secret $HASURA_GRAPHQL_ADMIN_SECRET`
   into `my_cinema/hasura` **before** any metadata change, so applying later can't clobber live config.
4. **verify.et contract (CONFIRM).** The public docs pages give the following. Read the API reference
   (`https://checkout.verify.et/docs/api`) or install their agent skill (see `https://checkout.verify.et/docs`)
   and confirm:
   - `POST https://checkoutapi.verify.et/v1/deposits`, headers `Authorization: Bearer vchk_…`,
     `Idempotency-Key` (required), `VerifyCheckout-Version: 2026-06-01`.
   - body: `merchant_customer_id`, `currency: "ETB"`, `return_url`, optional `amount` (omit = variable),
     optional `payment_method`. **CONFIRM** whether `merchant_order_id` / `metadata` are settable. If so, send `merchant_order_id = topup_id`.
   - response `data`: `id`, `status`, `amount` (string, e.g. `"250.00"`), `currency`, `checkout_url`,
     `expires_at`, `merchant_customer_id`.
   - statuses: `awaiting_transfer`, `verification_pending`, `succeeded`, `failed`, `expired`, `review_required` (and possibly `cancelled`).
   - **CONFIRM** that for a variable deposit `data.amount` on `succeeded` is the verified amount actually paid.
     **CONFIRM** what happens when the payer types an amount different from what they transferred
     (failed vs review_required vs corrected amount). The product rule is "credit what was actually
     received". If verify.et can't express that, report it to the user.
   - webhook: headers `VerifyCheckout-Signature: v1=<hex>`, `VerifyCheckout-Timestamp`, `VerifyCheckout-Event-Id`.
     **CONFIRM** the JSON path of the deposit id in the payload. Event `webhook.test` exists.
5. **Chapa verify endpoint (CONFIRM):** `GET https://api.chapa.co/v1/transaction/verify/{tx_ref}`
   with `Authorization: Bearer CHASECK…`, response `data.status == "success"`, `data.amount`, `data.currency`.

**Acceptance:** a short preflight note listing each finding. No writes yet.

---

## T1. P0 security fix: role self-escalation on `public.users`

The current policy `users_update_own` lets any authenticated user update **any column** of their own
row, including `role`, through Supabase's REST API (Supabase grants `UPDATE` on public tables to
`authenticated` by default). A wallet customer could set `role = 'admin'`.

Before applying, grep `autosuq_admin/src` for updates to `users` and list the columns it writes
(expected: `full_name`, `avatar_url`). Include all of them in the grant.

```sql
revoke update on public.users from anon, authenticated;
grant update (full_name, avatar_url) on public.users to authenticated;
```

**Acceptance:**
- As an authenticated non-admin (use a test user via the MCP or `set local role authenticated; set local request.jwt.claims = '{"sub":"<uuid>"}'`),
  `update public.users set role='admin' where id = auth.uid()` fails with *permission denied*.
- Updating `full_name` still works. The admin dashboard profile edit still works.

---

## T2. Database migration: roles, wallet tables, functions

Apply as **one migration** named `wallet_v1` through the Supabase MCP `apply_migration`, and save the same
SQL to `my_cinema/db/migrations/<timestamp>_wallet_v1.sql` for review. Adjust names that T0 found to differ.

### T2.1 Roles

```sql
alter table public.users drop constraint if exists users_role_check;
alter table public.users
  add constraint users_role_check check (role in ('admin', 'owner', 'customer'));

-- Email sign-ups from the wallet app pass options.data.app = 'wallet'.
-- Google OAuth can't pass metadata. Those users are converted in wallet_onboard().
create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer set search_path = public
as $$
begin
  insert into public.users (id, email, full_name, avatar_url, role)
  values (
    new.id,
    new.email,
    new.raw_user_meta_data ->> 'full_name',
    new.raw_user_meta_data ->> 'avatar_url',
    case when new.raw_user_meta_data ->> 'app' = 'wallet' then 'customer' else 'owner' end
  )
  on conflict (id) do nothing;
  return new;
end;
$$;

create or replace function public.has_role(p_role text)
returns boolean
language sql stable
security definer set search_path = public
as $$
  select exists (select 1 from public.users where id = auth.uid() and role = p_role);
$$;

-- Customers must not write to owner-facing tables.
drop policy if exists "items_insert_own_or_admin" on public.vending_machine_items;
create policy "items_insert_owner_or_admin"
  on public.vending_machine_items for insert
  with check ((created_by = auth.uid() and public.has_role('owner')) or public.is_admin());

drop policy if exists "order_requests_insert_own" on public.machine_order_requests;
create policy "order_requests_insert_owner"
  on public.machine_order_requests for insert
  with check (user_id = auth.uid() and public.has_role('owner'));
```

### T2.2 Tables

```sql
create table public.wallets (
  user_id    uuid primary key references public.users(id) on delete restrict,
  balance    numeric(12,2) not null default 0 check (balance >= 0),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table public.wallet_topups (
  id               uuid primary key,                       -- generated by Go before calling verify.et
  user_id          uuid not null references public.wallets(user_id),
  deposit_id       text not null unique,                   -- verify.et data.id
  requested_amount numeric(12,2) check (requested_amount > 0), -- what the user picked in our UI; a hint only
  credited_amount  numeric(12,2),                          -- set once, on succeeded
  status           text not null check (status in (
                     'awaiting_transfer','verification_pending','review_required',
                     'succeeded','failed','expired','cancelled')),
  checkout_url     text,                                   -- null when adopted from a webhook (see wallet_apply_deposit)
  expires_at       timestamptz,
  created_at       timestamptz not null default now(),
  updated_at       timestamptz not null default now()
);
create index wallet_topups_user_created_idx on public.wallet_topups (user_id, created_at desc);
create index wallet_topups_open_idx on public.wallet_topups (created_at)
  where status in ('awaiting_transfer','verification_pending','review_required');

create table public.wallet_ledger (
  id            uuid primary key default gen_random_uuid(),
  user_id       uuid not null references public.wallets(user_id),
  kind          text not null check (kind in ('topup','purchase')),
  amount        numeric(12,2) not null,                    -- + credit, - debit
  balance_after numeric(12,2) not null check (balance_after >= 0),
  topup_id      uuid unique references public.wallet_topups(id),
  tex_ref       text unique,                               -- = transactions.tex_ref of the basket
  machine_id    uuid references public.vending_machine(id),
  created_at    timestamptz not null default now(),
  check (
    (kind = 'topup'    and amount > 0 and topup_id is not null and tex_ref is null) or
    (kind = 'purchase' and amount < 0 and tex_ref is not null and topup_id is null)
  )
);
create index wallet_ledger_user_created_idx on public.wallet_ledger (user_id, created_at desc);

create table public.machine_credentials (
  user_id         uuid primary key references public.wallets(user_id) on delete cascade,
  username        text not null unique check (username ~ '^[A-Za-z0-9]{4,6}$'),
  password_hash   text not null,                           -- bcrypt, produced by Go
  failed_attempts int not null default 0,
  locked_until    timestamptz,
  created_at      timestamptz not null default now(),
  updated_at      timestamptz not null default now()
);

-- Return type for wallet_pay (Hasura v2 functions must return a tracked table). Never holds rows.
create table public.wallet_pay_result (
  status      text not null,
  tex_ref     text,
  total       numeric(12,2),
  balance     numeric(12,2),
  motor_codes text            -- comma-separated, one entry per unit, same format as the Chapa path
);
```

### T2.3 Grants and RLS

```sql
revoke all on public.wallets, public.wallet_topups, public.wallet_ledger,
              public.machine_credentials, public.wallet_pay_result
  from anon, authenticated;

grant select on public.wallets, public.wallet_topups, public.wallet_ledger to authenticated;
grant select (user_id, username, locked_until, updated_at) on public.machine_credentials to authenticated;

alter table public.wallets             enable row level security;
alter table public.wallet_topups       enable row level security;
alter table public.wallet_ledger       enable row level security;
alter table public.machine_credentials enable row level security;
alter table public.wallet_pay_result   enable row level security;   -- no policies: unreadable via API

create policy wallets_select_own on public.wallets
  for select to authenticated using (user_id = (select auth.uid()));
create policy wallet_topups_select_own on public.wallet_topups
  for select to authenticated using (user_id = (select auth.uid()));
create policy wallet_ledger_select_own on public.wallet_ledger
  for select to authenticated using (user_id = (select auth.uid()));
create policy machine_credentials_select_own on public.machine_credentials
  for select to authenticated using (user_id = (select auth.uid()));
```

### T2.4 Functions

All are `security definer`, `set search_path = public`, and not executable by API roles.

```sql
-- Called after every wallet-app sign-in. Returns the wallet, or no rows if the
-- account may not use the wallet (a real owner/admin).
create or replace function public.wallet_onboard(p_user_id uuid)
returns setof public.wallets
language plpgsql volatile
security definer set search_path = public
as $$
declare
  v_role text;
begin
  select role into v_role from public.users where id = p_user_id for update;

  if v_role = 'owner'
     and not exists (select 1 from public.vending_machine        where owner_id   = p_user_id)
     and not exists (select 1 from public.vending_machine_items  where created_by = p_user_id)
     and not exists (select 1 from public.machine_order_requests where user_id    = p_user_id)
     and not exists (select 1 from public.contact_support        where user_id    = p_user_id)
     and not exists (select 1 from public.user_phone_numbers     where user_id    = p_user_id)
  then
    update public.users set role = 'customer', updated_at = now() where id = p_user_id;
    v_role := 'customer';
  end if;

  if v_role is distinct from 'customer' then
    return;
  end if;

  insert into public.wallets (user_id) values (p_user_id) on conflict do nothing;
  return query select * from public.wallets where user_id = p_user_id;
end;
$$;

-- Idempotent: safe to call any number of times for the same deposit, in any order.
-- p_status / p_amount must come from GET /v1/deposits/{id}, never from a webhook body.
create or replace function public.wallet_apply_deposit(
  p_deposit_id text,
  p_user_id    uuid,
  p_status     text,
  p_amount     numeric
)
returns setof public.wallet_topups
language plpgsql volatile
security definer set search_path = public
as $$
declare
  v_topup   public.wallet_topups;
  v_balance numeric(12,2);
begin
  select * into v_topup from public.wallet_topups where deposit_id = p_deposit_id for update;

  if not found then
    -- Deposit exists at verify.et but our insert after creating it was lost.
    insert into public.wallets (user_id) values (p_user_id) on conflict do nothing;
    insert into public.wallet_topups (id, user_id, deposit_id, status)
    values (gen_random_uuid(), p_user_id, p_deposit_id, p_status)
    returning * into v_topup;
  end if;

  if v_topup.user_id <> p_user_id then
    raise exception 'deposit % belongs to a different user', p_deposit_id;
  end if;

  if v_topup.status = 'succeeded' then
    return next v_topup;
    return;
  end if;

  if p_status = 'succeeded' then
    if p_amount is null or p_amount <= 0 then
      raise exception 'succeeded deposit % has invalid amount %', p_deposit_id, p_amount;
    end if;

    update public.wallets
       set balance = balance + p_amount, updated_at = now()
     where user_id = p_user_id
    returning balance into v_balance;

    insert into public.wallet_ledger (user_id, kind, amount, balance_after, topup_id)
    values (p_user_id, 'topup', p_amount, v_balance, v_topup.id);

    update public.wallet_topups
       set status = 'succeeded', credited_amount = p_amount, updated_at = now()
     where id = v_topup.id
    returning * into v_topup;
  else
    update public.wallet_topups
       set status = p_status, updated_at = now()
     where id = v_topup.id
    returning * into v_topup;
  end if;

  return next v_topup;
end;
$$;

-- One atomic purchase. p_items = [{"slot_id": 12, "quantity": 2}, ...]
-- Returns exactly one row. status: SUCCESS | INVALID_ITEMS | OUT_OF_STOCK | INSUFFICIENT_BALANCE | NO_WALLET
-- On any non-SUCCESS status nothing is changed except resetting failed login attempts.
create or replace function public.wallet_pay(
  p_user_id    uuid,
  p_machine_id uuid,
  p_items      jsonb,
  p_tex_ref    text
)
returns setof public.wallet_pay_result
language plpgsql volatile
security definer set search_path = public
as $$
declare
  v_balance     numeric(12,2);
  v_total       numeric(12,2) := 0;
  v_motor_codes text := '';
  r             record;
  v_slot        record;
begin
  if p_items is null or jsonb_typeof(p_items) <> 'array' or jsonb_array_length(p_items) = 0 then
    return query select 'INVALID_ITEMS'::text, null::text, null::numeric, null::numeric, null::text;
    return;
  end if;

  -- Serializes concurrent payments from the same wallet (two kiosks, same credential).
  select balance into v_balance from public.wallets where user_id = p_user_id for update;
  if not found then
    return query select 'NO_WALLET'::text, null::text, null::numeric, null::numeric, null::text;
    return;
  end if;

  -- The credential was valid (Go checked it), so reset the lockout counter.
  update public.machine_credentials
     set failed_attempts = 0, locked_until = null, updated_at = now()
   where user_id = p_user_id and (failed_attempts <> 0 or locked_until is not null);

  -- Pass 1: validate, lock slots in id order (deadlock-safe), price from catalog.
  for r in
    select (e ->> 'slot_id')::int as slot_id, sum((e ->> 'quantity')::int) as qty
      from jsonb_array_elements(p_items) e
     group by 1
     order by 1
  loop
    if r.qty is null or r.qty <= 0 then
      return query select 'INVALID_ITEMS'::text, null::text, null::numeric, null::numeric, null::text;
      return;
    end if;

    select s.amount as stock, i.price
      into v_slot
      from public.vending_machine_to_items s
      join public.vending_machine_items i on i.id = s.item_id
     where s.id = r.slot_id and s.machine_id = p_machine_id
       for update of s;

    if not found then
      return query select 'INVALID_ITEMS'::text, null::text, null::numeric, null::numeric, null::text;
      return;
    end if;
    if v_slot.stock < r.qty then
      return query select 'OUT_OF_STOCK'::text, null::text, null::numeric, null::numeric, null::text;
      return;
    end if;

    v_total := v_total + v_slot.price * r.qty;
  end loop;

  if v_balance < v_total then
    return query select 'INSUFFICIENT_BALANCE'::text, null::text, v_total, v_balance, null::text;
    return;
  end if;

  -- Pass 2: apply. Rows are already locked.
  update public.wallets
     set balance = balance - v_total, updated_at = now()
   where user_id = p_user_id
  returning balance into v_balance;

  insert into public.wallet_ledger (user_id, kind, amount, balance_after, tex_ref, machine_id)
  values (p_user_id, 'purchase', -v_total, v_balance, p_tex_ref, p_machine_id);

  for r in
    select (e ->> 'slot_id')::int as slot_id, sum((e ->> 'quantity')::int) as qty
      from jsonb_array_elements(p_items) e
     group by 1
     order by 1
  loop
    insert into public.transactions (combination_id, price, amount, status, tex_ref, created_at, updated_at)
    select s.id, i.price, r.qty, true, p_tex_ref, now(), now()
      from public.vending_machine_to_items s
      join public.vending_machine_items i on i.id = s.item_id
     where s.id = r.slot_id;

    update public.vending_machine_to_items
       set amount = amount - r.qty
     where id = r.slot_id
    returning motor_code into v_slot;

    if coalesce(v_slot.motor_code, '') <> '' then
      v_motor_codes := concat_ws(',', nullif(v_motor_codes, ''),
                                 array_to_string(array_fill(v_slot.motor_code, array[r.qty::int]), ','));
    end if;
  end loop;

  return query select 'SUCCESS'::text, p_tex_ref, v_total, v_balance, v_motor_codes;
end;
$$;

-- Records one failed machine login. 5 consecutive failures lock the credential for 15 minutes.
create or replace function public.machine_login_failed(p_user_id uuid)
returns setof public.machine_credentials
language sql volatile
security definer set search_path = public
as $$
  update public.machine_credentials
     set failed_attempts = case when failed_attempts + 1 >= 5 then 0 else failed_attempts + 1 end,
         locked_until    = case when failed_attempts + 1 >= 5 then now() + interval '15 minutes' else locked_until end,
         updated_at      = now()
   where user_id = p_user_id
  returning *;
$$;

-- Chapa: mark a basket paid exactly once and decrement stock in the same statement.
-- Returns the rows that flipped false -> true. Empty result = already processed (don't dispense again).
create or replace function public.chapa_mark_paid(p_tex_ref text)
returns setof public.transactions
language sql volatile
security definer set search_path = public
as $$
  with paid as (
    update public.transactions
       set status = true, updated_at = now()
     where tex_ref = p_tex_ref and status = false
    returning *
  ), dec as (
    update public.vending_machine_to_items s
       set amount = greatest(s.amount - p.qty, 0)
      from (select combination_id, sum(coalesce(amount, 1))::int as qty from paid group by combination_id) p
     where s.id = p.combination_id
  )
  select * from paid;
$$;

revoke execute on function
  public.wallet_onboard(uuid),
  public.wallet_apply_deposit(text, uuid, text, numeric),
  public.wallet_pay(uuid, uuid, jsonb, text),
  public.machine_login_failed(uuid),
  public.chapa_mark_paid(text)
from public, anon, authenticated;
```

> If T0 found that stock is already decremented elsewhere, remove the `update ... vending_machine_to_items`
> statements from `wallet_pay` and the `dec` CTE from `chapa_mark_paid`.

### T2.5 SQL test script

Create `my_cinema/db/tests/wallet_v1_test.sql`. It runs inside `begin; … rollback;` against the branch DB
and raises if an assertion fails. Seed a test user, wallet, machine, two slots (price 10.00 stock 3,
price 25.50 stock 1) and assert:

1. `wallet_apply_deposit('dep_1', u, 'awaiting_transfer', null)`: topup row exists, balance 0.
2. `wallet_apply_deposit('dep_1', u, 'succeeded', 50)` twice: balance **50.00** (credited once), one ledger row.
3. `wallet_apply_deposit('dep_1', u, 'failed', null)` after success: status stays `succeeded`.
4. `wallet_pay(u, m, '[{"slot_id":S1,"quantity":2}]', 'tx_a')`: `SUCCESS`, total 20.00, balance 30.00,
   motor_codes has 2 entries, slot stock 1, `transactions` row with `price = 10.00`, `amount = 2`, `status = true`.
5. `wallet_pay(..., S2 qty 2)` returns `OUT_OF_STOCK` with no changes. `S2 qty 1` + `S1 qty 1` (total 35.50 > 30) returns `INSUFFICIENT_BALANCE` with no changes.
6. A slot of a different machine returns `INVALID_ITEMS`. Quantity 0 returns `INVALID_ITEMS`.
7. 5 × `machine_login_failed(u)`: `locked_until` is about now + 15 min, `failed_attempts` = 0.
8. `chapa_mark_paid` on a seeded `status=false` basket returns rows and decrements stock. A second call returns 0 rows.
9. As `authenticated`, `select public.wallet_pay(...)` fails with permission denied, `select password_hash from machine_credentials` fails, and `update wallets` fails.

**T2 acceptance:** the migration applies cleanly on the branch, the test script passes, and Supabase Advisors
(`get_advisors` security + performance) show no new warnings for these objects.

---

## T3. Hasura metadata

Work in the exported metadata (T0.3), then `hasura metadata apply` (ask the user before applying to the cloud instance).

1. **Track** tables `wallets`, `wallet_topups`, `wallet_ledger`, `machine_credentials`, `wallet_pay_result`.
   Add **no** select/insert/update/delete permissions for any role. Only the admin secret (Go) can use them.
2. Add object relationship `wallet_topups.user_id → wallets` (optional; not used by the code below).
3. **Track functions** as mutations (`exposed_as: mutation`), no role permissions:
   `wallet_onboard`, `wallet_apply_deposit`, `wallet_pay`, `machine_login_failed`, `chapa_mark_paid`.
   Hasura then exposes e.g. `wallet_pay(args: {p_user_id:…, p_machine_id:…, p_items:…, p_tex_ref:…}) { status … }`.
   `chapa_mark_paid` returns `transactions`, so callers can select `vending_machine_to_item { motor_code }` through the existing relationship.
4. **New action** `payWithWallet`:

```graphql
# actions.graphql
type Mutation {
  payWithWallet(arg: PayWithWalletInput!): PayWithWalletOutput!
}

input PayWithWalletInput {
  machine_id: String!
  username: String!
  password: String!
  items: [ItemRequestInput!]!   # existing type; item_id = SLOT id (vending_machine_to_items.id)
}

type PayWithWalletOutput {
  status: String!     # SUCCESS | INVALID_CREDENTIALS | LOCKED | INSUFFICIENT_BALANCE | OUT_OF_STOCK | INVALID_ITEMS | TRY_LATER | ERROR
  message: String!    # English fallback text; the kiosk maps status to its own copy
  txRef: String
  total: String       # decimal string, e.g. "35.50"
  balance: String     # decimal string; set on SUCCESS and INSUFFICIENT_BALANCE
}
```

```yaml
# actions.yaml
- name: payWithWallet
  definition:
    kind: synchronous
    handler: '{{WEB_HOOK_URL}}/v1/wallet/pay'
    timeout: 15
    forward_client_headers: false
    headers:
      - name: x-action-secret
        value_from_env: ACTION_SECRET
    request_transform:
      body:
        action: transform
        template: |-
          {
            "machine_id": {{$body.input.arg.machine_id}},
            "username": {{$body.input.arg.username}},
            "password": {{$body.input.arg.password}},
            "items": {{$body.input.arg.items}}
          }
      template_engine: Kriti
      version: 2
  permissions:
    - role: anonymous
```

Add `PayWithWalletInput` / `PayWithWalletOutput` to `custom_types` and set `ACTION_SECRET` in the Hasura Cloud env.

5. Make sure Hasura logs don't persist action request bodies, which contain the password.
   Hasura Cloud: check the project's logging/observability settings. Don't enable `query-log` in production.

**Acceptance:** from the Hasura console as `anonymous`, `payWithWallet` is visible. `wallets`, `wallet_pay`
etc. are **not** visible. With the admin secret they are.

---

## T4. Go: foundation

All paths are relative to `my_cinema/server`.

### T4.1 Config

`config/wallet_config.go`:

```go
type WalletConfig struct {
    SupabaseURL         string // SUPABASE_URL
    SupabaseJWTSecret   string // SUPABASE_JWT_SECRET (only for legacy HS256 projects)
    VerifyAPIBaseURL    string // VERIFY_API_BASE_URL, default https://checkoutapi.verify.et
    VerifyAPIKey        string // VERIFY_API_KEY
    VerifyWebhookSecret string // VERIFY_WEBHOOK_SECRET
    WalletWebURL        string // WALLET_WEB_URL, e.g. https://wallet.autosuq.com (no trailing slash)
    ActionSecret        string // HASURA_ACTION_SECRET
}
func NewWalletConfig() (*WalletConfig, error) // error if any required value is empty
```

`main.go` fails fast on a config error. Add the new variables to `server/.env.example` with placeholder values.

### T4.2 Supabase JWT middleware

`utils/supabase_auth.go`, middleware `SupabaseAuth(cfg) gin.HandlerFunc`:

- Read `Authorization: Bearer <token>`. If missing or malformed, return `401 {"code":"unauthorized"}`.
- Verify with `github.com/golang-jwt/jwt/v5` (already in go.sum):
  - JWKS mode (T0.2): use `github.com/MicahParks/keyfunc/v3` against `SUPABASE_URL/auth/v1/.well-known/jwks.json`
    (cached, refreshed in the background). Allowed algs: `ES256`, `RS256`.
  - Legacy mode: HS256 with `SupabaseJWTSecret`.
  - Require `exp` valid, `iss == SUPABASE_URL + "/auth/v1"`, `aud == "authenticated"`, `sub` a UUID.
- `c.Set("supabase_user_id", sub)`.

Don't reuse `utils/AuthMiddleware`. It validates the cinema app's own JWT.

### T4.3 MQTT: persistent connection

Today `VendingMachineService.SendCommand` connects and disconnects for every publish. Change it to:

- Connect once at startup in `main.go` (auto-reconnect is already on). Log the error and keep running if the broker is down.
- `IsConnected()` is exposed through the service interface.
- `SendCommand(motorCodes string) error` publishes to topic **`chapa/webhook`**, payload = the comma-separated
  motor codes (unchanged format), QoS 1, waits on the token with a 5 s timeout, and retries up to 3 times (200 ms, 400 ms, 800 ms).
- The unused `vendingMachineID int` parameter, `StartMotor` and `StopMotor` can go if nothing else uses them (grep first).

### T4.4 Action secret middleware

`utils/action_secret.go`: `ActionSecret(secret string) gin.HandlerFunc` compares the `x-action-secret` header with
`subtle.ConstantTimeCompare`. On mismatch it returns `401`.

---

## T5. Go: wallet feature

### T5.1 Domain (`domain/wallet.go`)

```go
type Wallet struct { UserID string; Balance string }                 // money as decimal strings
type Topup struct {
    ID, UserID, DepositID, Status string
    RequestedAmount, CreditedAmount *string
    CheckoutURL *string
    ExpiresAt *time.Time
}
type MachineCredential struct {
    UserID, Username, PasswordHash string
    FailedAttempts int
    LockedUntil *time.Time
}
type PayItem struct { SlotID int `json:"slot_id"`; Quantity int `json:"quantity"` }
type PayResult struct { Status, TexRef, Total, Balance, MotorCodes string }
type Deposit struct { ID, Status, Amount, CustomerID, CheckoutURL string; ExpiresAt *time.Time } // from verify.et

const (
    PaySuccess = "SUCCESS"; PayInvalidCredentials = "INVALID_CREDENTIALS"; PayLocked = "LOCKED"
    PayInsufficientBalance = "INSUFFICIENT_BALANCE"; PayOutOfStock = "OUT_OF_STOCK"
    PayInvalidItems = "INVALID_ITEMS"; PayTryLater = "TRY_LATER"; PayError = "ERROR"
)
var credentialPattern = regexp.MustCompile(`^[A-Za-z0-9]{4,6}$`)
func ValidCredentialPart(s string) bool
func IsTerminalDepositStatus(s string) bool // succeeded, failed, expired, cancelled
```

### T5.2 Repositories

`repositories/wallet_repository.go` uses Hasura with the admin secret and the same `*graphql.Client` as `PaymentRepository`.
Hasura returns `numeric` as strings (`HASURA_GRAPHQL_STRINGIFY_NUMERIC_TYPES=true`), so keep them as strings.

```go
type WalletRepositoryInterface interface {
    Onboard(userID string) (*domain.Wallet, error)                          // wallet_onboard; nil = not allowed
    GetWallet(userID string) (*domain.Wallet, error)                        // wallets_by_pk; nil if none
    InsertTopup(t domain.Topup) error                                       // insert_wallet_topups_one
    GetTopup(id string) (*domain.Topup, error)                              // wallet_topups_by_pk
    ListOpenTopups(olderThan, newerThan time.Time, limit int) ([]domain.Topup, error)
    ApplyDeposit(depositID, userID, status string, amount *string) (*domain.Topup, error) // wallet_apply_deposit
    GetCredentialByUsername(username string) (*domain.MachineCredential, error)  // exact, case-sensitive
    UpsertCredential(userID, username, passwordHash string) error            // on_conflict machine_credentials_pkey; update username, password_hash, failed_attempts=0, locked_until=null, updated_at
    RecordLoginFailure(userID string) error                                  // machine_login_failed
    Pay(userID, machineID string, items []domain.PayItem, texRef string) (domain.PayResult, error) // wallet_pay
}
var ErrUsernameTaken = errors.New("username taken") // map the unique violation on machine_credentials_username_key
```

`repositories/verify_checkout_client.go`:

```go
type VerifyCheckoutClientInterface interface {
    CreateDeposit(idempotencyKey, customerID, returnURL string) (domain.Deposit, error)
    GetDeposit(depositID string) (domain.Deposit, error)
}
```

- `POST {base}/v1/deposits` with headers `Authorization: Bearer {key}`, `Idempotency-Key`, `VerifyCheckout-Version: 2026-06-01`,
  `Content-Type: application/json`. Body: `{"merchant_customer_id": userID, "currency": "ETB", "return_url": …}`
  (+ `merchant_order_id` if T0.4 confirmed it). **No `amount`**, because deposits are always variable.
- `GET {base}/v1/deposits/{id}`.
- Parse `{data, error, meta}`. On error return `fmt.Errorf("verify.et %s: %s (request %s)", error.code, …, meta.requestId)`.
- Use an `http.Client` with a 10 s timeout.

### T5.3 Service (`services/wallet_service.go`)

```go
type WalletServiceInterface interface {
    Onboard(userID string) (*domain.Wallet, error)                       // ErrNotCustomer when the repo returns nil
    StartTopup(userID string, requestedAmount *string) (domain.Topup, error)
    GetTopup(userID, topupID string) (domain.Topup, *domain.Wallet, error) // refreshes from verify.et if non-terminal
    HandleVerifyWebhook(depositID string) error
    ReconcileOpenTopups() error
    SetMachineCredentials(userID, username, password string) error
    Pay(machineID, username, password string, items []domain.ItemRequest) domain.PayResult
}
```

Dependencies (interfaces): `WalletRepositoryInterface`, `VerifyCheckoutClientInterface`,
`VendingMachineServiceInterface` (T4.3), `*config.WalletConfig`, a clock `func() time.Time` (for tests).

**StartTopup**
1. `GetWallet(userID)`. If nil, return `ErrNotCustomer`.
2. If `requestedAmount` is set, it must parse as a decimal with `> 0`, `<= 10000.00` and at most 2 decimals. Otherwise return `ErrInvalidInput`.
3. `topupID := uuid.New()`. `returnURL := cfg.WalletWebURL + "/topup/return?ref=" + topupID`.
4. `CreateDeposit(topupID, userID, returnURL)`, then `InsertTopup{ID: topupID, DepositID: d.ID, Status: d.Status, RequestedAmount, CheckoutURL, ExpiresAt}`.
   If the insert fails, log the error with the deposit id and return it. A later webhook adopts the deposit through `wallet_apply_deposit`.

**GetTopup**
1. `GetTopup(id)`. If not found or `UserID != userID`, return `ErrNotFound`.
2. If the status is not terminal, call `GetDeposit`, then `ApplyDeposit(d.ID, d.CustomerID, d.Status, amountIfSucceeded)`.
   Pass the amount only when the status is `succeeded`. A verify.et failure is non-fatal: log it and return the stored row.
3. Return the topup + current wallet.

**HandleVerifyWebhook(depositID)**: `GetDeposit`, then `ApplyDeposit(d.ID, d.CustomerID, d.Status, amount)`. Return any error,
so the controller answers 5xx and verify.et retries.

**ReconcileOpenTopups**: `ListOpenTopups(now-2m, now-48h, 50)` and, for each, the same steps as the webhook. Errors are logged per item.
Started in `main.go` as a goroutine with `time.NewTicker(60 * time.Second)`.

**SetMachineCredentials**
1. Wallet must exist (`ErrNotCustomer`).
2. `ValidCredentialPart(username) && ValidCredentialPart(password)`, else `ErrInvalidInput`.
3. `utils.HashPassword(password)` (bcrypt, existing), then `UpsertCredential`. `ErrUsernameTaken` is passed through.

**Pay**: order matters. Never return before the bcrypt compare on the unknown-username path (timing).
1. `machineID` must be a UUID. 1–20 items, each `item_id > 0` and `1 <= amount <= 10`. Otherwise `INVALID_ITEMS`.
2. Format check on username/password. Invalid format returns `INVALID_CREDENTIALS`.
3. Global failure limiter (`sync.Mutex` + sliding 1-minute window): if there were ≥ 30 failed logins in the last
   minute, return `TRY_LATER`. Add `// ponytail: in-memory, per process; move to Redis/DB if the backend runs >1 instance`.
4. `GetCredentialByUsername`. If not found, run `bcrypt.CompareHashAndPassword(dummyHash, pw)`, record a limiter failure, return `INVALID_CREDENTIALS`.
5. If `LockedUntil` is after now, return `LOCKED` (message includes minutes remaining).
6. If bcrypt doesn't match, `RecordLoginFailure(userID)` + limiter failure, return `INVALID_CREDENTIALS`.
7. If `!vending.IsConnected()`, return `ERROR` ("Machine is offline. You were not charged.").
8. `texRef := uuid.New()`, then `repo.Pay(userID, machineID, items→[{slot_id, quantity}], texRef)`.
9. If the result is not `SUCCESS`, return it as is (`INSUFFICIENT_BALANCE` carries total + balance).
10. `vending.SendCommand(result.MotorCodes)`. On error, log `ERROR wallet dispense failed tex_ref=… user=…`
    and still return `SUCCESS`. The DB is the source of truth and support resolves it manually.
    Add `// ponytail: no automatic refund (D9); add wallet_refund(tex_ref) if dispense failures show up in logs`.
11. Return `SUCCESS` with txRef, total, balance.

Error mapping for unexpected repo errors: log with context, then return `ERROR`.

### T5.4 Controller and routes

`delivery/controllers/wallet_controller.go` and `delivery/routes/wallet_routes.go`:

| Method | Path | Auth | Request | Responses |
|---|---|---|---|---|
| POST | `/v1/wallet/onboard` | SupabaseAuth | – | `200 {balance, username\|null}` · `403 {code:"owner_account"}` |
| POST | `/v1/wallet/topups` | SupabaseAuth | `{requested_amount?: "50.00"}` | `201 {topup_id, checkout_url, expires_at}` · `400 invalid_input` · `403 not_customer` · `502 provider_error` |
| GET | `/v1/wallet/topups/:id` | SupabaseAuth | – | `200 {topup_id, status, requested_amount, credited_amount, balance}` · `404` |
| PUT | `/v1/wallet/machine-credentials` | SupabaseAuth | `{username, password}` | `200 {username}` · `400 invalid_input` · `409 username_taken` · `403 not_customer` |
| POST | `/v1/wallet/pay` | ActionSecret | `{machine_id, username, password, items:[{item_id, amount}]}` | always `200` with `PayWithWalletOutput` JSON (see T3) |
| POST | `/v1/wallet/webhooks/verify` | signature (below) | raw verify.et event | `200` on success or duplicate · `401` bad signature · `500` processing error (retry) |

Error body shape: `{"code": "<snake_case>", "message": "<human text>"}`.

Status messages for `/pay` (English, `message` field):

| status | message |
|---|---|
| SUCCESS | `Payment successful. Dispensing your items.` |
| INVALID_CREDENTIALS | `Wrong username or password.` |
| LOCKED | `Too many wrong attempts. Try again in N minutes.` |
| INSUFFICIENT_BALANCE | `Not enough balance. You have X ETB, this costs Y ETB. Top up in the AutoSuq Wallet app.` |
| OUT_OF_STOCK | `Some items just sold out. Please update your cart.` |
| INVALID_ITEMS | `These items are not available on this machine.` |
| TRY_LATER | `Too many attempts right now. Please wait a minute.` |
| ERROR | `Something went wrong. You were not charged.` (or the machine-offline text) |

**verify.et webhook handler**
1. `raw, _ := io.ReadAll(c.Request.Body)`. Don't bind JSON first.
2. `utils.VerifyCheckoutSignature(raw, ts, sig, secret, now)`: parse `v1=<hex>`. Reject if `|now - ts| > 300s`.
   `expected = hex(HMAC_SHA256(secret, ts + "." + raw))`, compared with `hmac.Equal`.
3. Parse the event. If it's `webhook.test`, return 200. Extract the deposit id (path confirmed in T0.4). If it's missing, return 200 and log (nothing to do).
4. `HandleVerifyWebhook(depositID)`. On error return 500 (verify.et retries). Otherwise 200.

Register routes in `delivery/main.go`. The reconcile goroutine also starts there.

### T5.5 Go tests

- `utils/verify_checkout_signature_test.go`: valid signature, tampered body, wrong secret, stale timestamp, malformed header.
- `services/wallet_service_test.go`, with fakes for the repo, verify client and vending service:
  1. Pay success: `repo.Pay` called once with summed items, `SendCommand` called with the motor codes, status `SUCCESS`.
  2. Unknown username: `INVALID_CREDENTIALS`, `repo.Pay` not called, bcrypt still ran.
  3. Wrong password: `RecordLoginFailure` called once, `INVALID_CREDENTIALS`.
  4. Locked: `LOCKED`, no bcrypt compare needed, `repo.Pay` not called.
  5. MQTT disconnected: `ERROR`, `repo.Pay` **not** called (nothing charged).
  6. Repo returns `INSUFFICIENT_BALANCE`: passed through, `SendCommand` not called.
  7. 30 failures in a minute: the next call returns `TRY_LATER`.
  8. `GetTopup` with a non-terminal status refreshes and applies. With a terminal status it makes no provider call.
  9. `SetMachineCredentials` rejects `abc`, `abcdefg`, `ab_12`, and accepts `Ab12`, `A1b2C3`.

`go test ./...` passes. `go vet ./...` is clean.

---

## T6. Go: fix the existing Chapa path

In `payment_controller.go`, `payment_service.go`, `payment_repository.go` and `utils/chapa_webhook.go`:

1. **Signature.** Read the raw body first. Verify `x-chapa-signature == hex(HMAC_SHA256(CHAPA_WEBHOOK_SECRET, raw))`
   with `hmac.Equal`. On mismatch return 401. Remove the debug `fmt.Println` lines that print the secret and body.
   Fix `VerifyChapaWebhookSignature`, which currently HMACs the secret instead of the body, and update its test.
2. **Re-verify.** In `HandleWebhook`, call Chapa's verify endpoint for `tx_ref` (T0.5). Dispense only if
   `status == "success"`, `currency == "ETB"` and the amount equals `sum(price * coalesce(amount,1))` of the basket's rows.
3. **Exactly once.** Replace `UpdateStatusMultiple` with a call to `chapa_mark_paid(p_tex_ref)` that selects
   `combination_id amount vending_machine_to_item { motor_code }`. With 0 rows, return 200 and don't dispense (duplicate webhook).
4. **Unit price.** In `CreateTransactions`, set `Price` to each slot's own unit price, not `totalPrice`.
   Remove the `totalPrice` parameter.
5. Dispense through the updated `SendCommand(motorCodes)` (T4.3).

**Acceptance:** a test for the signature (valid/invalid). Replaying the same webhook twice dispenses once (verified by a service
test with fakes). A new basket's rows store unit prices.

---

## T7. Deployment checklist (done by the user; the agent prepares it)

Produce `my_cinema/server/WALLET_DEPLOY.md` covering:

1. verify.et dashboard: receiving account active; API key with create + read deposit scopes; return origin =
   `WALLET_WEB_URL`; webhook endpoint `https://<go-host>/v1/wallet/webhooks/verify` with outcome events;
   copy `whsec_…`.
2. Go env vars (T4.1). Hasura env `ACTION_SECRET`.
3. Supabase Auth: Email + Google providers enabled (already, per the admin docs). Add the wallet app's URL to
   **Redirect URLs** (`<WALLET_WEB_URL>/auth/callback`, `<WALLET_WEB_URL>/auth/confirm`).
4. Order: migration (T1+T2), then Hasura metadata (T3), then Go deploy, then kiosk build, then web app.
5. Smoke test: sign up, top up 10 ETB for real, confirm the balance, set credentials, buy one item at the kiosk,
   check `wallet_ledger`, `transactions` and stock.

Also flag to the user: `my_cinema/docker-compose.yml` contains a plaintext Neon connection string and a Hasura
JWT key. Rotate them if they're live and move them to env vars.

---

## T8. Kiosk app (`vending_machine_app`, Flutter)

1. **GraphQL op** `lib/app/data/graphql/pay_with_wallet.graphql`:
   ```graphql
   mutation PayWithWallet($machineId: String!, $username: String!, $password: String!, $items: [ItemRequestInput!]!) {
     payWithWallet(arg: {machine_id: $machineId, username: $username, password: $password, items: $items}) {
       status message txRef total balance
     }
   }
   ```
   Refresh `schema.graphql` from Hasura, then `dart run build_runner build --delete-conflicting-outputs`.
2. **Repository**: `TransactionRepository.payWithWallet(machineId, username, password, items)`.
3. **Bug fix (existing):** `ItemsController.initiatePayment` sends `vending_machine_item.id` (catalog id). The backend
   expects the **slot** id. Use `selectedItem.value!.id`, as `checkoutCart` already does.
4. **UI**: in `PaymentModal` and `CheckoutBottomSheet`, add a secondary button **"Pay with AutoSuq Wallet"** under "Pay Now".
   It opens `WalletLoginSheet`:
   - Username and Password fields: `maxLength: 6`, input formatter `[A-Za-z0-9]`, no autocorrect/suggestions,
     password obscured with a show/hide toggle, a large on-screen-keyboard-friendly layout, and the order total shown.
   - "Pay X Birr" button, disabled until both fields are 4–6 chars. While the request runs it shows a spinner and the inputs are disabled.
   - Result handling by `status`:
     - `SUCCESS`: close the sheet, clear the cart, refetch items, show "Payment successful. Dispensing…" (same snackbar style as Chapa success).
     - `INSUFFICIENT_BALANCE`: stay on the sheet and show "Balance {balance} Birr, needed {total} Birr. Top up at {WALLET_WEB_URL}".
     - `INVALID_CREDENTIALS` / `LOCKED` / `TRY_LATER`: inline error. Clear the password field only.
     - `OUT_OF_STOCK` / `INVALID_ITEMS`: close the sheet, refetch items, show the message.
     - `ERROR` / network error: inline error with a "Get help" action (existing `ContactSupportSheet`).
   - Clear both fields whenever the sheet closes, when the cart-expiry dialog fires, and when the screensaver appears.
     Don't keep credentials in any controller field after the request.
5. `machineId` comes from the existing route parameter / `MACHINE_ID` dart-define. If it's empty, hide the wallet button.
6. Optional (P2): a QR code on the sheet linking to `WALLET_WEB_URL` (needs the `qr_flutter` package; ask the user first).

**Acceptance:** with a seeded credential + balance on the branch DB, both the single-item and cart paths succeed.
Wrong password ×5 shows LOCKED. Low balance shows the message with both amounts.

---

## T9. Admin dashboard guard (`autosuq_admin`)

After sign-in (wherever the session/user is loaded), read `public.users.role`. If it's `customer`, sign out and show:
"This account is an AutoSuq Wallet account and can't access the owner dashboard." No other changes.

**Acceptance:** a customer test account can't get past login. Owners are unaffected.

---

## Definition of done

- T0 note written. T1–T6 implemented, T8–T9 implemented, T7 doc written.
- SQL test script passes on the branch. `go test ./...` passes. Flutter builds.
- No secrets in code or logs. No MQTT format change. The Chapa flow still works end-to-end.
- Final report to the user: files changed per repo, test output, open CONFIRM items, and anything that needs their action.
