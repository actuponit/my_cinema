# AutoSuq Wallet: Architecture

Status: agreed design, 2026-10-12. Scope: v1 for one physical location (one or
more machines). This is a temporary side flow next to the existing Chapa
pay-per-purchase flow. It does not replace it.

## 1. What the product does

1. A customer creates an account in the **wallet web app** (email + password, or Google).
2. The account has exactly one **wallet** (an ETB balance).
3. The customer **tops up** the wallet through **Verify Checkout** (checkout.verify.et).
   They transfer money to AutoSuq's receiving account and Verify.et checks the
   transfer reference. We credit the amount Verify.et **confirms**, which may
   differ from the amount the customer picked in our UI.
4. The customer sets a **machine credential**: a username and a password, each
   4–6 alphanumeric characters, case-sensitive.
5. At the vending machine kiosk the customer picks items, chooses **Pay with
   wallet** and types the machine credential. The backend checks the
   credential and the balance. If the balance covers the order, the backend
   debits the wallet and tells the machine to dispense. Otherwise the kiosk
   shows "insufficient balance".

There is no refund or reversal flow. The machine dispenses only after a
successful debit.

## 2. Systems

```
                        ┌──────────────────────────────┐
  Customer phone ─────▶ │ Wallet web app (THIS REPO)    │
  (browser)             │ Next.js 16, Vercel            │
                        └───────┬───────────────┬──────┘
             reads (RLS, anon key)              │ writes (Bearer <Supabase access token>)
                                ▼               ▼
┌──────────────────────────────────┐   ┌───────────────────────────────────────┐
│ Supabase project "actuponit"      │   │ Go backend  (repo actuponit/my_cinema, │
│ - Auth (email/password, Google)   │◀──│ dir server/; Gin; already hosted)      │
│ - Postgres (shared with admin     │   │ - /v1/wallet/* endpoints               │
│   dashboard + kiosk)              │   │ - Chapa endpoints (existing)           │
└──────────────▲───────────────────┘   │ - verify.et client + webhook + poller  │
               │ SQL (as postgres)      │ - MQTT publisher (dispense)            │
┌──────────────┴───────────────────┐   └───┬───────────────▲────────────┬──────┘
│ Hasura Cloud                      │◀──────┘ GraphQL        │ Action     │ MQTT
│ pleasing-viper-86.hasura.app      │  (x-hasura-admin-secret)│ webhook    ▼
│ - tables + tracked SQL functions  │─────────────────────────┘    ┌────────────┐
│ - Actions: initiateTransaction,   │                              │ Machine    │
│   payWithWallet (new)             │                              │ controller │
└──────────────▲───────────────────┘                              │ (firmware) │
               │ GraphQL (role: anonymous)                         └────────────┘
┌──────────────┴───────────────────┐
│ Kiosk app (repo vending_machine_  │        ┌────────────────────────────────┐
│ app; Flutter, mounted on machine) │        │ Verify Checkout (verify.et)    │
└──────────────────────────────────┘        │ hosted checkout + webhooks      │
                                             └────────────────────────────────┘
```

| System | Repo / location | Role in wallet flow |
|---|---|---|
| Wallet web app | `autosuq-web-app` (this repo) | Sign-up/login, balance, history, top-up, machine credential |
| Go backend | `actuponit/my_cinema` → `server/` | All wallet writes, verify.et integration, kiosk payment, MQTT dispense |
| Hasura Cloud | `pleasing-viper-86.hasura.app`, metadata in `my_cinema/hasura/` | GraphQL over Postgres. Go uses it with the admin secret. The kiosk uses it as `anonymous` |
| Supabase | project `actuponit` | Auth plus the single Postgres DB, shared with the admin dashboard (`autosuq_admin`) |
| Kiosk app | `vending_machine_app` | New "Pay with wallet" path |
| Admin dashboard | `autosuq_admin` | Must reject `customer` accounts. No wallet UI in v1 |
| Machine firmware | not in these repos | Unchanged. It still receives comma-separated `motor_code`s on MQTT topic `chapa/webhook` |

### Answer: does Supabase need Edge Functions to talk to our backend?

No. Supabase here is Auth plus Postgres. Our existing Go backend reaches the
DB through Hasura, which connects to Supabase Postgres directly. Hasura
Actions call any HTTP endpoint we host, which is how `initiateTransaction`
already works. Postgres functions hold the money logic. Edge Functions would
only be useful for code that has to run inside Supabase's infrastructure. We
have no such code.

## 3. Key decisions (and why)

| # | Decision | Why |
|---|---|---|
| D1 | Supabase only for the web app. Remove Firebase from this repo | One auth system, shared with the admin dashboard |
| D2 | Same Supabase project, new role `customer` in `public.users.role` | Debit and `transactions` insert must happen in one DB transaction, so one DB |
| D3 | Wallet logic lives in the existing Go backend | It already owns payments, Hasura access and MQTT dispensing |
| D4 | Money changes only in Postgres functions (`wallet_pay`, `wallet_apply_deposit`). Go calls them through Hasura | Row locks and idempotency in one transaction. No double spend between two kiosks |
| D5 | Web app reads wallet data directly from Supabase (RLS: own rows only) and sends writes to Go with the Supabase access token. Go verifies the token | No extra Next.js API layer |
| D6 | Top-ups always create a variable Verify.et deposit. Preset buttons only suggest an amount | We credit what Verify.et confirms (`deposit.amount` on `succeeded`), even if it differs from the pick |
| D7 | The kiosk sends slot ids + quantities. The server prices the order from the catalog | The kiosk can't choose its own price |
| D8 | Machine credential: 4–6 alphanumeric chars, case-sensitive, username unique, password bcrypt-hashed in Go. 5 consecutive failures lock it for 15 min | The credential space is small. Lockout plus a global failure limiter limits guessing |
| D9 | No refund/reversal path | The machine only dispenses after the debit commits |
| D10 | Mismatched top-ups are explained in the app only (no email/SMS) | Smallest v1 |
| D11 | No admin UI for wallets in v1 | Use the Supabase dashboard / SQL |
| D12 | One account has one role. A pristine `owner` (no machines/items/requests/tickets) that logs into the wallet app is converted to `customer`. Any other owner is refused | Google OAuth can't carry signup metadata, so the onboard call decides |

## 4. Data model (new)

All amounts are `numeric(12,2)` in ETB.

```
public.users (existing)  role: 'admin' | 'owner' | 'customer'   (customer is new)
   │ 1
   │ 1
public.wallets            user_id PK/FK, balance >= 0
   │ 1                 │ 1
   │ *                 │ 0..1
public.wallet_ledger   public.machine_credentials
  kind topup|purchase    username UNIQUE, password_hash (bcrypt),
  amount (+/-)           failed_attempts, locked_until
  balance_after
  topup_id UNIQUE ─────▶ public.wallet_topups (deposit_id UNIQUE, requested_amount,
  tex_ref  UNIQUE ──┐                          credited_amount, status, checkout_url)
                    └─▶ public.transactions.tex_ref (existing; one basket = one tex_ref)

public.wallet_pay_result  (empty table; only the return type of wallet_pay for Hasura)
```

Invariants (enforced in the DB):

- `wallets.balance >= 0` (CHECK).
- `wallets.balance` changes only together with a `wallet_ledger` insert in the same function.
  So `balance = sum(ledger.amount)` always holds.
- A deposit is credited at most once (`wallet_ledger.topup_id UNIQUE` + `wallet_topups.deposit_id UNIQUE`).
- A basket is debited at most once (`wallet_ledger.tex_ref UNIQUE`).
- Each wallet purchase writes one `transactions` row per slot. That row holds the **unit**
  `price`, `amount` = quantity, `status = true` and the shared `tex_ref`. This matches the
  admin dashboard's revenue formula `sum(price * coalesce(amount,1))`.

The full SQL is in [backend-spec.md](backend-spec.md) §T2.

## 5. Flows

### 5.1 Sign-up / onboarding

```
Browser ──signUp(email,pw, data:{app:'wallet'})──▶ Supabase Auth ──trigger──▶ users.role='customer'
Browser ──Google OAuth──▶ Supabase ──▶ /auth/callback (Next.js) ──▶ users.role='owner' (default)
Browser (after any sign-in) ──POST /v1/wallet/onboard──▶ Go ──wallet_onboard(user_id)──▶
      role owner & pristine → customer;  customer → ensure wallet row;  else → 403 owner_account
```

### 5.2 Top-up

```
1. Web  POST /v1/wallet/topups {requested_amount?}            (Bearer token)
2. Go   topup_id := uuid
        POST verify.et /v1/deposits  (Idempotency-Key: topup_id)
             {merchant_customer_id: user_id, currency: ETB,
              return_url: WALLET_WEB_URL/topup/return?ref=topup_id}
        insert wallet_topups(id=topup_id, deposit_id, requested_amount, checkout_url, ...)
        → {topup_id, checkout_url}
3. Web  window.location = checkout_url
4. Payer transfers money and enters amount + reference on verify.et
5a. verify.et ──signed webhook──▶ Go /v1/wallet/webhooks/verify
        verify HMAC, then GET /v1/deposits/{id} (authoritative), then wallet_apply_deposit(...)
5b. verify.et redirects the browser to /topup/return?ref=topup_id
        Web polls GET /v1/wallet/topups/{topup_id}. Go refreshes from verify.et if non-terminal
5c. Go poller (every 60s) refreshes non-terminal top-ups younger than 48h
6. wallet_apply_deposit credits deposit.amount once, and only when status = succeeded
```

Credit triggers 5a, 5b and 5c all go through the same idempotent function, so
whichever arrives first credits the wallet and the others do nothing.

### 5.3 Purchase at the kiosk

```
Kiosk ──payWithWallet(machine_id, username, password, items[{item_id: SLOT id, amount}])──▶ Hasura
Hasura ──POST /v1/wallet/pay (x-action-secret)──▶ Go
Go:  validate input format
     global failure limiter open?                    no → TRY_LATER
     load machine_credentials by username (Hasura)
     unknown → bcrypt against a dummy hash           → INVALID_CREDENTIALS
     locked_until > now()                            → LOCKED
     bcrypt mismatch → machine_login_failed(user)    → INVALID_CREDENTIALS
     MQTT connected?                                 no → ERROR (nothing charged)
     tex_ref := uuid
     wallet_pay(user, machine, items, tex_ref)   ── one SQL transaction:
         lock wallet; lock slots (ordered, must belong to machine);
         stock check; total from catalog; balance check;
         debit + ledger; insert transactions (status=true, unit price); decrement stock
     status != SUCCESS                               → that status (INSUFFICIENT_BALANCE, OUT_OF_STOCK, ...)
     MQTT publish motor codes to "chapa/webhook"     (same payload as the Chapa path)
     → SUCCESS {txRef, total, balance}
Kiosk: SUCCESS → "Dispensing…", refetch stock, clear cart; otherwise show the message for that status
```

### 5.4 Existing Chapa flow (fixed, not redesigned)

The flow is unchanged except for these fixes ([backend-spec.md](backend-spec.md) §T6):
verify `x-chapa-signature` over the raw body, re-verify with Chapa's verify
API, mark the basket paid only on the `false → true` transition (no replay
dispense), store the unit price per row, and decrement stock in the same
statement.

## 6. Security model

| Surface | Who calls | Auth |
|---|---|---|
| Supabase PostgREST (web reads) | Browser | Supabase session. RLS: own rows only. Column grants hide `password_hash`. No write grants on wallet tables |
| Go `/v1/wallet/*` (web writes) | Browser | `Authorization: Bearer <Supabase access token>`, verified by Go (JWKS or the legacy HS256 secret) |
| Go `/v1/wallet/pay` | Hasura Action only | `x-action-secret` header, constant-time compare |
| Go `/v1/wallet/webhooks/verify` | verify.et | `VerifyCheckout-Signature` HMAC-SHA256 over `<timestamp>.<raw body>`, 300 s tolerance |
| Go `/payments/webhook` | Chapa | `x-chapa-signature` HMAC-SHA256 over the raw body + Chapa verify API |
| Wallet SQL functions | Hasura (postgres role) | `EXECUTE` revoked from `public, anon, authenticated`, so they can't be called through Supabase RPC |
| New tables in Hasura | Go (admin secret) | Tracked with **no** role permissions, so the kiosk's `anonymous` role can't see them |

Fixed alongside this work: `public.users` lets any signed-in user `UPDATE`
their own row, **including `role`**, so a customer could make themselves
`admin`. [backend-spec.md](backend-spec.md) §T1 restricts it to column-level
grants.

## 7. Environments and config

| Where | Variable | Purpose |
|---|---|---|
| Web (Vercel) | `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Supabase client |
| Web | `NEXT_PUBLIC_WALLET_API_URL` | Go backend base URL |
| Web | `NEXT_PUBLIC_SITE_URL` | OAuth redirect + verify.et return origin |
| Go | `SUPABASE_URL` | JWKS + issuer check |
| Go | `SUPABASE_JWT_SECRET` | Only if the project still signs with the legacy HS256 secret |
| Go | `VERIFY_API_BASE_URL` (`https://checkoutapi.verify.et`), `VERIFY_API_KEY` (`vchk_…`), `VERIFY_WEBHOOK_SECRET` (`whsec_…`) | verify.et |
| Go | `WALLET_WEB_URL` | Builds `return_url`. Must be a registered return origin in verify.et |
| Go | `HASURA_ACTION_SECRET` | Authenticates the `payWithWallet` action |
| Hasura | `ACTION_SECRET` (same value), `WEB_HOOK_URL` (exists) | Action header + handler base |

## 8. Out of scope for v1

Admin wallet UI, refunds/reversals, withdrawals, email/SMS notifications,
multi-currency, multi-location tenancy, Amharic UI strings (structure allows
adding them later), and the existing Chapa rows with `status=false` that the
admin dashboard reads as "failed" while they are still pending.
