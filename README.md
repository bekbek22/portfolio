# Portfolio

Decoupled full-stack portfolio.

| Path      | What                                                       |
| --------- | ---------------------------------------------------------- |
| `server/` | REST API — Go, Fiber v3, pgx, PostgreSQL (`/api/v1`)       |
| `client/` | UI — Next.js (App Router), Tailwind CSS, TypeScript (Bun)  |

## Local development

```bash
docker compose up -d                 # Postgres on 127.0.0.1:5433 (+ portfolio_test DB)

cd server
cp .env.example .env
go run ./cmd/api                     # http://localhost:4000/api/v1/health
go test ./...

cd ../client
bun install
bun run dev                          # http://localhost:3000
```

Server image: `docker build -t portfolio-server server` (distroless, non-root, ~14 MB).
