# go-auth

Authentication service for the personal-enterprise project. Handles user registration, login, and JWT issuance. All other services in the system validate tokens independently — this is the only service that signs them.

> Active development — endpoints and schema documented once stable.

---

## Stack

- **Language**: Go
- **Router**: [chi](https://github.com/go-chi/chi)
- **Database**: Postgres via [pgx](https://github.com/jackc/pgx)
- **Query generation**: [sqlc](https://sqlc.dev)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Tokens**: RSA-signed JWTs (access) + hashed refresh tokens (Postgres-backed, rotated on use)
- **Passwords**: argon2id
- **Logging**: `slog` with JSON output

---

## Running Locally

Copy `.env.example` to `.env.local` and fill in values.

```bash
# Start the server
go run ./cmd/server

# Run migrations
go run ./cmd/migrate up
go run ./cmd/migrate down   # rolls back one step

# Run tests
go test ./...
```

---

## Environment Variables

| Variable | Description |
|---|---|
| `PORT` | HTTP port (defaults to 8080) |
| `DATABASE_URL` | Postgres connection string (`postgres://user:pass@host/auth`) |
| `JWT_PRIVATE_KEY` | RSA private key PEM — used to sign access tokens |
| `JWT_PUBLIC_KEY` | RSA public key PEM — used to verify tokens on protected endpoints |
| `REFRESH_TOKEN_PEPPER` | Secret mixed into refresh token hashing |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins |

The RSA key pair is shared: `JWT_PUBLIC_KEY` is distributed to all other services so they can verify tokens without calling back to go-auth.
