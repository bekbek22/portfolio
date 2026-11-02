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
cp .env.example .env                 # API_URL=http://localhost:4000
bun install
bun run dev                          # http://localhost:3000 (needs the API running)
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

**Image:** `docker build -t portfolio-server server` — distroless, non-root, ~18 MB.
Migrations are embedded: run `/api migrate up` as the pre-deploy command.

## Client

| Command              | What                                                               |
| -------------------- | ------------------------------------------------------------------ |
| `bun run dev`        | dev server; `/api/*` is forwarded to `API_URL`                     |
| `bun run build`      | production build; prerenders pages from the API, so it must be up  |
| `bun run test`       | unit tests (`bun test`)                                             |
| `bun run test:e2e`   | Playwright smoke tests against a production build + seeded API    |
| `bun run gen:api`    | regenerate `src/lib/api/schema.d.ts` from `server/api/openapi.yaml` |

Pages read data in Server Components through a typed `openapi-fetch` client, cached with
`"use cache"` + `cacheTag("projects")`. Project filters live in the URL
(`/projects?tag=go&tag=postgresql`), so every filtered view is a shareable link.

## Deployment

Neon (Postgres), Railway (`server/`) and Vercel (`client/`), all in Singapore. Deploy in that
order: `next build` prerenders pages from the API, so it must be live and seeded first.

| Service | Setting            | Value                                        |
| ------- | ------------------ | -------------------------------------------- |
| Railway | Root Directory     | `server`                                     |
| Railway | Pre-deploy command | `/api migrate up`                            |
| Railway | Healthcheck path   | `/api/v1/health`                             |
| Railway | Variables          | `APP_ENV=production`, `PORT=4000`, `DATABASE_URL` (Neon direct, not `-pooler`), `TRUSTED_PROXY_HOPS` |
| Vercel  | Root Directory     | `client`                                     |
| Vercel  | Function region    | `sin1`                                       |
| Vercel  | Variables          | `API_URL` (Railway URL, no trailing slash)   |

The image has no shell, so `seed` runs from a machine with the repo: set `DATABASE_URL` to the
Neon URL, then `go run ./cmd/api seed`.
