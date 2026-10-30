# Portfolio

Decoupled full-stack portfolio.

| Path      | What                                                       |
| --------- | ---------------------------------------------------------- |
| `server/` | REST API — Go, Fiber v3, pgx, PostgreSQL (`/api/v1`)       |
| `client/` | UI — Next.js (App Router), Tailwind CSS, TypeScript (Bun)  |

## Local development

```bash
docker compose up -d                 # Postgres on 127.0.0.1:5433 (dbs: portfolio, portfolio_test)

cd server
cp .env.example .env
go run ./cmd/api migrate up          # apply embedded SQL migrations
go run ./cmd/api seed                # load placeholder projects (safe to re-run)
go run ./cmd/api                     # http://localhost:4000/api/v1/health

cd ../client
bun install
bun run dev                          # http://localhost:3000
```

## Server

| Command                                   | What                                                    |
| ----------------------------------------- | ------------------------------------------------------- |
| `go run ./cmd/api migrate up`             | apply pending migrations                                |
| `go run ./cmd/api migrate down [n]`       | roll back `n` migrations (default 1)                    |
| `go run ./cmd/api migrate version`        | show the current schema version                         |
| `go run ./cmd/api seed`                   | upsert the 5 placeholder projects and their tags (idempotent) |
| `sqlc generate`                           | regenerate `internal/database/sqlc` from `queries/*.sql` |
| `go generate ./...`                       | regenerate the Fiber strict server from `api/openapi.yaml` |

**API contract.** `server/api/openapi.yaml` is the source of truth (spec-first). The Go
server interface (`internal/httpapi/gen`) and the client types (`client/src/lib/api/schema.d.ts`,
`bun run gen:api`) are generated from it and committed; CI fails if either is stale. The live
spec is served at `GET /api/v1/openapi.json`.

**Tests.** Unit tests need nothing. Database integration tests use `TEST_DATABASE_URL`
and run each test in its own rolled-back transaction:

```bash
export TEST_DATABASE_URL="postgres://portfolio:portfolio@127.0.0.1:5433/portfolio_test?sslmode=disable"
go test ./...
```

Without `TEST_DATABASE_URL` they are skipped. Plain `go test ./...` only prints `ok` for
passing packages, so use `go test -v ./...` to see the skip warnings. In CI (`CI` set) a
missing `TEST_DATABASE_URL` fails the build instead.

**sqlc** is pinned to **v1.31.1** (CI and `sqlc.yaml`). Install the release binary from
https://github.com/sqlc-dev/sqlc/releases/tag/v1.31.1 (building from source needs cgo).

**Image:** `docker build -t portfolio-server server` — distroless, non-root, ~16 MB.
Migrations are embedded: run `/api migrate up` as the pre-deploy command.
