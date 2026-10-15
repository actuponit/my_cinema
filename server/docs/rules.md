# Rules

These are kept short on purpose. They apply to every repo touched by the wallet work.

## Clean architecture

Dependencies point inward. Each outer layer knows the layer inside it, and an
inner layer never knows the layers outside it.

```
delivery (HTTP handlers, routes, pages)  →  services (use cases)  →  domain (types, rules)
                                   repositories (DB / Hasura / external APIs) implement what services need
```

| Layer | Go backend (`my_cinema/server`) | Web app (this repo) | Allowed to know about |
|---|---|---|---|
| domain | `domain/` | `src/domain/` | nothing (no Gin, no Hasura, no React, no fetch) |
| services | `services/` | `src/services/` | domain + the interfaces it declares for repositories |
| repositories | `repositories/` (Hasura, verify.et, Chapa clients) | `src/lib/` (Supabase client, wallet API client) | domain, SDKs |
| delivery | `delivery/controllers`, `delivery/routes`, `delivery/main.go` | `src/app/` (routes, components) | services, domain |

1. Handlers parse and validate input, call **one** service method, and map the
   result to HTTP. They contain no business logic.
2. Services hold the use case and talk to repositories through small
   interfaces declared next to the service. That is the existing pattern, and
   it is what makes services testable with fakes.
3. Repositories return domain types, never raw GraphQL/HTTP structs.
4. `delivery/main.go` (Go) and the route files (web) are the only places that
   wire concrete implementations together.
5. Don't add a layer, interface or package unless something uses it today.

## Money

1. **Every balance change goes through a Postgres function** (`wallet_pay`,
   `wallet_apply_deposit`). Never update `wallets.balance` from Go or the web
   app directly.
2. Money is `numeric(12,2)` in SQL and a decimal **string** in JSON/GraphQL
   (`"12.50"`). Go doesn't do wallet arithmetic in `float64`, because the
   totals are computed in SQL.
3. Every credit and debit is idempotent through a unique key: `deposit_id`
   for top-ups, `tex_ref` for purchases.
4. Trust external payment state only after re-fetching it from the provider's
   API (verify.et `GET /v1/deposits/{id}`, Chapa verify). A webhook body or a
   browser redirect only means "go check".
5. Never log passwords, password hashes, API keys, webhook secrets or access tokens.

## Security

1. Secrets live in environment variables. Commit only `.env.example`.
2. Every webhook and action endpoint checks a signature or shared secret with
   a constant-time compare before it reads the payload.
3. New tables: enable RLS, grant only what the web app reads, and add no
   Hasura role permissions unless the kiosk needs them.
4. New SQL functions: `security definer set search_path = public`, plus
   `revoke execute ... from public, anon, authenticated`.

## Code style

1. Follow the conventions of the repo you're in. Go: `gofmt`, errors wrapped
   with `fmt.Errorf("...: %w", err)`. Web: TypeScript strict, server components
   by default, `"use client"` only where needed.
2. No comments that restate code. Comment only on a non-obvious *why*.
3. Each non-trivial piece of logic (money, auth, signature checks) ships with
   at least one test that fails if the logic breaks.
4. Don't commit, push or deploy unless asked. Report what changed and what's left.
