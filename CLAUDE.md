# go-auth — Claude Context

## What This Service Is

The authentication service for the personal-enterprise project. It is the **only** service that owns JWT issuance — all other services validate JWTs independently using only the public key. This service manages users, issues access + refresh token pairs, and handles token rotation and revocation.

Built from `go-service-template`. The structure, patterns, and tooling are inherited from that template — this file documents what is specific to go-auth on top of that foundation.

---

## Architecture

```
cmd/
  server/main.go     ← wiring: config, DB, stores, token service, server.New()
  migrate/main.go    ← golang-migrate runner
app/
  app.go             ← Application struct: UserStore, TokenStore, TokenService
server/
  server.go          ← chi router, global middleware
  routes.go          ← route registration
config/
  config.go          ← Config struct — extended with auth-specific fields
db_connection/
  db.go              ← pgxpool setup
db/
  schema.sql         ← users + refresh_tokens tables
  sqlc.yaml
  queries/           ← user and token queries
  migrations/
  generated/
health/
  handler.go
middleware/
  auth.go            ← RequireAuth — used for protected endpoints (e.g. /me, /logout)
  logging.go
  requestid.go
user/
  user_model.go
  user_store.go
  user_handler.go    ← register, login
  user_routes.go
token/
  token_model.go
  token_store.go     ← refresh token persistence (Postgres)
  token_service.go   ← JWT issuance, verification, rotation, revocation
  token_handler.go   ← refresh, logout
  token_routes.go
```

---

## Key Differences from Other Go Services

### Token Service

go-auth has a `token/token_service.go` — a service layer between handler and store. This is the only domain in any Go service that warrants a service layer (per the architecture: service layer only where there's real logic beyond CRUD).

The token service owns:
- Minting access JWTs (RSA-signed, 15 min TTL)
- Minting refresh tokens (random bytes, stored as SHA-256 hash + pepper)
- Token rotation on refresh (revoke old, issue new pair)
- Revocation on logout

### Auth Middleware Usage

Unlike other services which apply `middleware.RequireAuth` broadly, go-auth applies it selectively:
- **Unprotected**: `POST /users/register`, `POST /users/login`, `POST /token/refresh`
- **Protected**: `GET /users/me`, `POST /token/logout`

### Both RSA Keys Required

Other services only need `JWT_PUBLIC_KEY`. go-auth needs both:
- `JWT_PRIVATE_KEY` — for signing access tokens
- `JWT_PUBLIC_KEY` — for verifying access tokens on protected endpoints
- `REFRESH_TOKEN_PEPPER` — mixed into refresh token hashing

### Password Hashing

argon2id via `golang.org/x/crypto`. Never bcrypt.

---

## Patterns Carried Over from Template

### Domain Structure

Four-file pattern per domain: `<domain>_model.go`, `_store.go`, `_handler.go`, `_routes.go`. Token adds a fifth: `token_service.go`.

### Database

- sqlc for all queries — no raw SQL strings in handlers or stores
- golang-migrate for migrations: `go run ./cmd/migrate [up|down]`
- Migration files: `000001_create_users.up.sql` / `000001_create_users.down.sql`
- `db/generated/` is committed

### Logging

`slog.SetDefault(logger)` in main. JSON to stdout. All packages use `slog` directly.

### Testing

Integration tests via testcontainers — real Postgres, no mocks. `TestMain` handles container lifecycle.

### Conventions

- File naming: `user_handler.go`, `token_store.go`, etc.
- Receiver names: `h` for handlers, `s` for stores, `svc` for the token service
- Errors: log with `slog.Error` server-side, generic message to client
- Routes function: `Routes(...) chi.Router`
- Auth applied at mount point in `server/routes.go`

---

## Environment Variables

| Variable | Description |
|---|---|
| `PORT` | HTTP port (defaults to 8080) |
| `DATABASE_URL` | Postgres connection string (`postgres://user:pass@host/db`) |
| `JWT_PRIVATE_KEY` | RSA private key PEM for signing access tokens |
| `JWT_PUBLIC_KEY` | RSA public key PEM for verifying access tokens on protected endpoints |
| `REFRESH_TOKEN_PEPPER` | Secret mixed into refresh token hashing |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins |

Copy `.env.example` to `.env.local` for local dev. Never commit `.env.local`.

---

## Decided

- **User IDs**: UUIDv7 — time-ordered (good for index locality), privacy-safe in JWTs, generated via `uuid.NewV7()` from `github.com/google/uuid`
- **No `/token/public-key` endpoint** — `JWT_PUBLIC_KEY` is an env var injected via GCP Secret Manager at deploy time. An endpoint that serves key material is unnecessary attack surface and introduces a startup dependency.
- **Schema**: work out when writing `db/schema.sql` — details depend on what sqlc queries and the token service actually need.
