# AutoSuq Wallet: deploy checklist

Do these in order. Nothing here has been applied to any live system yet.

## 0. Before anything: secrets already in git history

Rotate these and set them only through env vars:

- **MQTT (HiveMQ) password** for user `autosuq`. It was hardcoded in `utils/mqtt_client.go` (`DefaultMQTTConfig`, now deleted).
  The server now reads `MQTT_BROKER_URL`, `MQTT_USERNAME`, `MQTT_PASSWORD`, `MQTT_CLIENT_ID`.
  **If these aren't set on the Go host, the server falls back to `tcp://test.mosquitto.org:1883` and machines won't dispense.**
  Production value: `MQTT_BROKER_URL=mqtts://0b8c3235cf1145668f5d9841d803ed68.s1.eu.hivemq.cloud:8883`.
  Leave `MQTT_CLIENT_ID` empty or make it unique per instance (two clients with the same id kick each other off).
- **Neon connection string** (`my_cinema_owner:…@ep-crimson-sun-…neon.tech`), previously in `docker-compose.yml`.
- **Hasura JWT key** `oursupersecret…`, previously in `docker-compose.yml`. It's still the local-dev default (`HASURA_JWT_KEY`).

## 1. verify.et dashboard

1. Receiving account active.
2. API key with create + read deposit scopes, which becomes `VERIFY_API_KEY` (`vchk_…`).
3. Return origin = `WALLET_WEB_URL` (scheme, host and port must match exactly).
4. Webhook endpoint `https://<go-host>/v1/wallet/webhooks/verify`. Keep the default outcome events + `webhook.test`.
   Copy the one-time `whsec_…` secret into `VERIFY_WEBHOOK_SECRET`. Deploy Go first, then click "send test" (the first 2xx activates it).

## 2. Env vars

Go server (`server/.env.example` lists all of them):

| Var | Notes |
|---|---|
| `SUPABASE_URL` | `https://weunvxrpkfeznuodhhye.supabase.co` |
| `SUPABASE_JWT_SECRET` | Optional if JWKS is active (Supabase asymmetric ES256 keys enabled). Used as fallback for legacy HS256 tokens. |
| `VERIFY_API_BASE_URL` | default `https://checkoutapi.verify.et` |
| `VERIFY_API_KEY`, `VERIFY_WEBHOOK_SECRET` | from step 1 |
| `WALLET_WEB_URL` | e.g. `https://wallet.autosuq.com`, no trailing slash |
| `HASURA_ACTION_SECRET` | random 32+ chars, **same value** as Hasura's `ACTION_SECRET` |
| `CHAPA_WEBHOOK_SECRET` | must be the secret configured in the Chapa dashboard; webhooks without a valid `x-chapa-signature` now get 401 |
| `MQTT_*` | see step 0 |

The server refuses to start if a required wallet var is missing.

Hasura Cloud env: `ACTION_SECRET` (same as `HASURA_ACTION_SECRET`). Keep `query-log` off in production
(the `payWithWallet` request body contains the kiosk password).

## 3. Supabase Auth

Email + Google providers enabled. Add `<WALLET_WEB_URL>/auth/callback` and `<WALLET_WEB_URL>/auth/confirm` to Redirect URLs.

## 4. Order

1. DB migration `db/migrations/20261012_wallet_v1.sql` (Supabase branch first, then production), then `db/tests/wallet_v1_test.sql` on the branch.
2. Hasura metadata: run `scripts/sync-hasura-metadata.sh` into a scratch copy and diff it against `hasura/metadata`, so nothing changed live since the export gets clobbered. Then `hasura metadata apply`.
3. Go deploy.
4. Kiosk build.
5. Web app.

## 5. Smoke test

1. Sign up in the wallet app and top up 10 ETB for real. The balance shows 10.00.
2. Set machine credentials. Buy one item at the kiosk with "Pay with AutoSuq Wallet".
3. Check: one `wallet_ledger` row per action, a `transactions` row with `status = true` and the unit price, slot stock decremented.
4. Chapa: buy one item with Chapa. It dispenses once, and the basket rows store unit prices.

## Known gaps

- Chapa baskets created by the **old** server and paid after the deploy store the basket total on every row, so the
  amount check fails and they don't dispense (logged as `amount mismatch`). Deploy at a quiet time.
- No automatic refund if a wallet purchase is charged but MQTT dispense fails. Grep logs for `wallet dispense failed`.
- The failed-login limiter is in memory. It's fine for one Go instance.
