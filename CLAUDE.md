# go-auth — Claude Context

## What This Service Is

The authentication service for the personal-enterprise project. It is the **only** service that owns JWT issuance — all other services validate JWTs independently using only the public key. This service manages users, issues access + refresh token pairs, and handles token rotation and revocation.

Built from `go-service-template`. The structure, patterns, and tooling are inherited from that template — this file documents what is specific to go-auth on top of that foundation.

---

## Architecture

```
cmd/
  server/main.go     ← wiring: config, DB, stores, token service, server.New()
app/
  app.go             ← Application struct: UserStore, TokenService
server/
  server.go          ← chi router, global middleware
  routes.go          ← route registration
config/
  config.go          ← Config struct — extended with auth-specific fields
db_connection/
  db.go              ← MongoDB Connect() — returns (*mongo.Client, *mongo.Database), creates indexes
health/
  handler.go
middleware/
  auth.go            ← RequireAuth — used for protected endpoints (e.g. /me, /logout)
  logging.go
  requestid.go
user/
  user_model.go      ← User struct (domain), PublicUser, CreateUserRequest, LoginResponse
  user_store.go      ← MongoDB store; internal userDoc with bson tags
  user_handler.go    ← register, login
  user_routes.go
token/
  token_model.go     ← ExchangeResult, RefreshToken (domain)
  token_store.go     ← MongoDB store; internal refreshTokenDoc with bson tags
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

## Database

MongoDB Atlas. Database name: `auth`. Collections: `users`, `refresh_tokens`.

### Store pattern

Stores define a private doc struct with `bson` tags for MongoDB serialization, and a domain struct (in `_model.go`) with plain Go types for everything outside the store. The store's methods convert between the two. This keeps bson tags out of the domain layer and handlers unchanged if the DB ever changes again.

```
userDoc (bson)  ←→  User (domain)         — in user/
refreshTokenDoc (bson)  ←→  RefreshToken (domain)  — in token/
```

IDs are stored as strings (`uuid.UUID.String()`), parsed back to `uuid.UUID` on read.

### Indexes (created at startup in db_connection/db.go)

- `users.email` — unique
- `refresh_tokens.hash` — for active token lookup
- `refresh_tokens.userId` — for RevokeAllForUser

### Active token query

`revokedAt` is stored as `null` (not omitted) on new tokens so the filter `{revokedAt: null, expiresAt: {$gt: now}}` works correctly. Do not add `omitempty` to `refreshTokenDoc.RevokedAt`.

### Connection string

`DATABASE_URL` is a MongoDB URI (`mongodb+srv://user:pass@cluster.mongodb.net/`). No database name in the URI — the database is selected in code via `client.Database("auth")`. Secret Manager secret: `db-url-auth`. **Trailing newlines in the secret value will cause auth failures** — Secret Manager doesn't make them visible in the UI.

---

## Patterns Carried Over from Template

### Domain Structure

Four-file pattern per domain: `<domain>_model.go`, `_store.go`, `_handler.go`, `_routes.go`. Token adds a fifth: `token_service.go`.

### Logging

`slog.SetDefault(logger)` in main. JSON to stdout. All packages use `slog` directly.

### Testing

Integration tests via testcontainers — `mongo:6` container, no mocks. `TestMain` handles container lifecycle.

### Conventions

- File naming: `user_handler.go`, `token_store.go`, etc.
- Receiver names: `h` for handlers, `s` for stores, `svc` for the token service
- Errors: log with `slog.Error` server-side, generic message to client
- Routes function signatures:
  - `user.Routes(store *user.Store, svc *token.TokenService) chi.Router` — login needs to issue tokens
  - `token.Routes(svc *token.TokenService) chi.Router` — no store needed directly; service owns store
- Auth applied at mount point in `server/routes.go`

---

## Environment Variables

| Variable | Description |
|---|---|
| `PORT` | HTTP port (defaults to 8080) |
| `DATABASE_URL` | MongoDB URI (`mongodb+srv://user:pass@cluster.mongodb.net/`) |
| `JWT_PRIVATE_KEY` | RSA private key PEM for signing access tokens |
| `JWT_PUBLIC_KEY` | RSA public key PEM for verifying access tokens on protected endpoints |
| `REFRESH_TOKEN_PEPPER` | Secret mixed into refresh token hashing |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins |

Copy `.env.example` to `.env.local` for local dev. Never commit `.env.local`.

---

## Current State

- Fully implemented and deployed to dev
- MongoDB migration complete — no Postgres, no sqlc, no golang-migrate
- User registration, login, JWT issuance, token rotation, and revocation all working
- CI/CD: `--add-cloudsql-instances` removed from deploy steps (no longer needed)

---

## Decided

- **User IDs**: UUIDv7 — time-ordered, privacy-safe in JWTs, generated via `uuid.NewV7()` from `github.com/google/uuid`
- **No `/token/public-key` endpoint** — `JWT_PUBLIC_KEY` is an env var injected via GCP Secret Manager at deploy time. An endpoint that serves key material is unnecessary attack surface and introduces a startup dependency.
- **No database name in URI** — database selected in code (`client.Database("auth")`), not in the connection string
