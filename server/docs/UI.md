# UI: AutoSuq Wallet Web App

The customer-facing web app in this repo. Mobile-first and responsive, used mostly on phones standing near
the machine. Read [architecture.md](architecture.md) for the flows and [backend-spec.md](backend-spec.md) §T5.4
for the API it calls.

## 1. Brand (taken from `../auto-suq-landingpage`)

| Token | Light | Dark | Use |
|---|---|---|---|
| `--background` | `#f7fafc` | `#0a0f1c` | page |
| `--foreground` | `#0f172a` | `#e7edf5` | text |
| `--card` | `#ffffff` | `#101828` | cards, sheets |
| `--card-border` | `#e2e8f0` | `#1f2b40` | borders |
| `--muted` | `#5b6b81` | `#93a1b8` | secondary text |
| `--muted-bg` | `#eef2f7` | `#131d30` | inputs, chips |
| `--primary` | `#5b9bd5` | `#5b9bd5` | buttons, links, balance accent |
| `--primary-strong` | `#3d7cb8` | `#4a8bc9` | button hover/pressed, primary text on light bg |
| `--primary-tint` | `#e3eef9` | `#16283d` | selected preset, info banners |
| `--ring` | `#5b9bd5` | `#5b9bd5` | focus ring |
| `--warn` | `#b5650f` | `#f0ac3c` | pending / mismatch notices (landing's `--stage-warm-text`) |
| `--danger` | `#c2410c` | `#f87171` | errors (new; keep contrast ≥ 4.5:1) |
| `--success` | `#15803d` | `#4ade80` | credited (new) |

- Copy the token block (`:root` + dark) and the `@theme inline` mapping from
  `auto-suq-landingpage/src/app/globals.css` into `src/app/globals.css`. Then add `--warn`, `--danger`, `--success`.
- Dark mode follows the system: put the dark tokens under `@media (prefers-color-scheme: dark)`. No theme toggle in v1.
- Body text on `--primary` buttons is white. Check contrast. If `#5b9bd5` with white text fails at body size, use
  `--primary-strong` as the button background.
- **Fonts:** `Archivo` (Latin) + `Noto_Sans_Ethiopic` (Ge'ez fallback for names) via `next/font/google`, the same
  setup as `auto-suq-landingpage/src/app/[lang]/layout.tsx`. `--font-sans: var(--font-archivo), var(--font-ethiopic), ui-sans-serif, system-ui`.
  Headlines use `font-stretch: 106%` (landing's `.font-display`).
- **Logo:** copy `auto-suq-landingpage/public/images/logo-light.png` and `logo-dark.png` to `public/images/`. Use
  `logo-light.png` on light backgrounds and `logo-dark.png` on dark (`<picture>` with `prefers-color-scheme`), height 28–32 px in the header.
  Copy the landing `favicon.ico` to `src/app/favicon.ico`.
- Name: **AutoSuq Wallet**. Tagline (optional, auth pages): "Convenience at your fingertips!"
- Language: English only in v1. Keep all user-facing strings in `src/content/strings.ts` so Amharic can be added later.
- Money is always shown as `1,250.00 ETB` (`Intl.NumberFormat('en-ET', {minimumFractionDigits: 2})` + " ETB").
  Times are shown in Africa/Addis_Ababa.

## 2. Layout

- Single column, `max-width: 28rem`, centered, `16px` side padding, no horizontal scroll down to 320 px width.
- Header: logo left, account icon right (→ `/account`). Header is hidden on auth pages; those show the logo centered instead.
- Primary actions are full-width buttons, min height 48 px. Touch targets are ≥ 44 px.
- Cards: `--card` bg, 1 px `--card-border`, radius 16 px, padding 20 px.
- Loading uses skeletons for cards and a spinner inside buttons. Never show a blank screen.
- Accessibility: every input has a visible label, errors are linked with `aria-describedby`, focus rings
  are visible (`--ring`), and status changes are announced (`role="status"` on the top-up result).

## 3. Screens

### `/login` and `/signup`
- Logo, title ("Sign in" / "Create your wallet"), and **Continue with Google** (secondary button with the Google mark).
- Divider "or", then Email and Password fields (signup: password ≥ 8 chars, show/hide toggle), primary button.
- Signup calls `supabase.auth.signUp({ email, password, options: { data: { app: 'wallet' }, emailRedirectTo: SITE_URL + '/auth/confirm' } })`,
  then shows the "Check your email to confirm" state.
- Google calls `signInWithOAuth({ provider: 'google', options: { redirectTo: SITE_URL + '/auth/callback' } })`.
- Errors are shown inline: wrong credentials, email not confirmed (with a resend link), network.
- Links to switch between sign in and sign up.

### `/auth/callback` and `/auth/confirm` (route handlers, no UI)
- `callback`: `exchangeCodeForSession(code)`, then redirect to `/onboard`.
- `confirm`: `verifyOtp({ type, token_hash })`, then redirect to `/onboard`.

### `/onboard` (transient)
- Spinner "Setting up your wallet…". Calls `POST /v1/wallet/onboard`.
- `200` redirects to `/wallet`. If `username` is null, the wallet home shows the "Set your machine login" card first.
- `403 owner_account`: full-screen message: "This email belongs to an AutoSuq machine owner account. Please use a
  different email for the wallet." Button **Sign out**.

### `/wallet` (home)
1. **Balance card**: label "Balance", amount in large type (`--foreground`, 40 px, tabular numerals), primary button **Top up**.
2. **Machine login card**:
   - Not set: "Set the username and password you'll type at the vending machine." Button **Set machine login**.
   - Set: "Machine username: **Ab12**" + **Change**. If `locked_until > now` show a `--warn` line
     "Locked after too many wrong attempts. Unlocks at 14:35."
3. **Recent activity**: last 10 rows from `wallet_ledger`, newest first.
   Row: icon (↓ top-up / ↑ purchase), title ("Top-up" / "Vending purchase"), time, and amount (`+50.00 ETB` in `--success`, `−20.00 ETB` in `--foreground`).
   Pending/failed top-ups from `wallet_topups` (status not `succeeded`, last 7 days) appear above the list as
   chips: "Top-up verifying…" (`--warn`) / "Top-up failed" (`--danger`) → tap opens `/topup/return?ref=<id>`.
   Footer link "See all" → `/history`.
- Empty state: "No activity yet. Top up to start buying from the machine."

### `/topup`
- Title "Top up". Preset chips in a 2×2 grid: **50**, **100**, **200**, **500** ETB, plus **Other amount**, which opens a numeric input
  (`inputmode="decimal"`, 1–10,000, up to 2 decimals).
- Info box (`--primary-tint`):
  "You'll pay on the secure Verify checkout. Enter **{amount} ETB** there and send exactly that amount.
  We credit whatever amount Verify confirms."
- Primary button **Continue to payment**, disabled until an amount is chosen. On click: `POST /v1/wallet/topups { requested_amount }`
  → `window.location.assign(checkout_url)`.
- Errors: `502 provider_error` shows "Payment service unavailable, try again shortly". `403` redirects to `/onboard`.

### `/topup/return?ref=<topup_id>`
Polls `GET /v1/wallet/topups/{ref}` every 3 s for up to 2 min. Stops when the status is terminal or the user leaves.

| Status | UI |
|---|---|
| `awaiting_transfer`, `verification_pending` | Spinner + "Verifying your payment… this usually takes a few seconds." |
| `review_required` | `--warn` icon + "Your payment is being reviewed. We'll add it to your balance when it's approved. You can close this page." |
| `succeeded`, credited = requested (or no requested) | `--success` check + "**{credited} ETB** added. New balance {balance} ETB." Button **Back to wallet** |
| `succeeded`, credited ≠ requested | `--success` check + `--warn` notice: "You selected {requested} ETB, but we received **{credited} ETB**. We added {credited} ETB to your balance." |
| `failed` | `--danger` + "We couldn't verify this payment. If money left your account, contact support with reference {deposit/support ref}." Buttons **Try again** / **Back to wallet** |
| `expired` / `cancelled` | "This payment link expired. No money was taken through it." **Try again** |
| 2 min elapsed, still non-terminal | "Still verifying. Your balance will update automatically. You can close this page." **Back to wallet** |

Never infer success from the redirect alone. Only the API status counts.

### `/machine-login`
- Title "Machine login". Explainer: "Type these on the vending machine screen to pay from your wallet."
- **Username** and **Password** fields: `maxLength=6`, `pattern="[A-Za-z0-9]{4,6}"`, `autocapitalize="off"`,
  `autocomplete="off"`, `spellcheck=false`. Live helper text: "4–6 letters or numbers. Capital letters matter. 4 is easiest to type."
  Password has a show/hide toggle. Confirm password field.
- Save calls `PUT /v1/wallet/machine-credentials`. `409` shows "That username is taken" on the username field. On success, a toast
  "Machine login saved" and back to `/wallet`.
- Note under the form: "After 5 wrong tries the machine login locks for 15 minutes. Changing your password here unlocks it."

### `/history`
- All `wallet_ledger` rows, paginated 20 per page (Supabase `.range()`), grouped by day (Addis Ababa time).
- Tapping a top-up row shows requested vs credited amount. Tapping a purchase row shows the date/time and machine name (if `machine_id` is set, join on
  `vending_machine.name` server-side. RLS hides machines from customers, so show just "Vending purchase" if that's unavailable).

### `/account`
- Email (read-only), sign-in method, **Sign out**. Support contact: Telegram @Isoazm, +251 93 932 5619 (from the landing page).

## 4. Data access in the web app

| Data | Source | How |
|---|---|---|
| Session | Supabase Auth | `@supabase/ssr` (`src/lib/supabase/*`, already present) |
| Balance | `wallets` | `select balance` where `user_id = me` (RLS) |
| Activity/history | `wallet_ledger`, `wallet_topups` | RLS-filtered selects, ordered by `created_at desc` |
| Machine username / lock | `machine_credentials` | **explicit columns** `select('username, locked_until')`. `select('*')` fails by design (column grants) |
| Onboard, top-up, top-up status, set credentials | Go backend | `src/lib/wallet-api.ts`: `fetch(NEXT_PUBLIC_WALLET_API_URL + path, { headers: { Authorization: 'Bearer ' + session.access_token } })`, typed responses, throws `WalletApiError {status, code, message}` |

## 5. Build tasks for this repo (web app)

1. Remove Firebase: `npm uninstall firebase firebase-admin`, delete `src/lib/firebase/`, remove the Firebase vars from `.env.example`,
   and remove the `firebase` server from `.mcp.json` if it isn't used elsewhere (ask the user).
2. `.env.example`: `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY`, `NEXT_PUBLIC_WALLET_API_URL`, `NEXT_PUBLIC_SITE_URL`.
3. Session refresh: add `src/proxy.ts` (Next.js 16's replacement for `middleware.ts`; check the current Next docs) calling
   `updateSession` from `src/lib/supabase/middleware.ts`. Redirect unauthenticated users to `/login` for every route except
   `/login`, `/signup`, `/auth/*`. Use `supabase.auth.getClaims()` (or `getUser()`) server-side, never `getSession()`, for auth decisions.
4. Structure (see [rules.md](rules.md)):
   ```
   src/
     app/
       (auth)/login/page.tsx  (auth)/signup/page.tsx
       auth/callback/route.ts  auth/confirm/route.ts
       onboard/page.tsx
       (app)/layout.tsx  (app)/wallet/page.tsx  (app)/topup/page.tsx  (app)/topup/return/page.tsx
       (app)/machine-login/page.tsx  (app)/history/page.tsx  (app)/account/page.tsx
     components/     balance-card.tsx, activity-list.tsx, amount-picker.tsx, credential-form.tsx, ...
     domain/         money.ts (format/parse/validate), credentials.ts (same regex as the backend)
     lib/            supabase/*, wallet-api.ts
     content/        strings.ts
   ```
5. Server components fetch Supabase reads. Client components handle forms, polling and the redirect to checkout.
6. Tests (minimum): `domain/money.ts` and `domain/credentials.ts` unit tests (valid/invalid amounts and credential strings).
7. Before calling it done: run `npm run dev` and walk through, at 375 px width and in dark mode: signup, onboard, set machine login,
   top up (use the verify.et test mode if available), return page states (mock the API responses to see each state), history, sign out.

## 6. Prompt for the implementing agent (frontend)

> Build the AutoSuq Wallet web app in this repo (`autosuq-web-app`, Next.js 16, React 19, Tailwind 4, Supabase).
> Read `docs/architecture.md`, `docs/rules.md` and `docs/UI.md` fully. The API contract is `docs/backend-spec.md` §T5.4
> (the Go backend may not be deployed yet; build against the contract and mock it with a small in-file fake where needed).
> Do §5 tasks in order. Match the brand in §1 exactly, using the landing page repo at `../auto-suq-landingpage` as the visual reference.
> Check current Next.js 16 and @supabase/ssr docs before writing auth/proxy code. Don't commit or deploy. When done,
> report the files changed, what you tested in the browser (widths, themes, states) and anything blocked on the backend.
