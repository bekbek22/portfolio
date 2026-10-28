# Portfolio

Decoupled full-stack portfolio.

| Path      | What                                                     |
| --------- | -------------------------------------------------------- |
| `server/` | REST API — Bun, ElysiaJS, Prisma, PostgreSQL (`/api/v1`) |
| `client/` | UI — Next.js (App Router), Tailwind CSS, TypeScript      |

## Local development

```bash
bun install                          # installs client + server (Bun workspace)
docker compose up -d                 # Postgres on 127.0.0.1:5433

cd server
cp .env.example .env
bun run db:migrate                   # apply migrations + generate client
bun run dev                          # http://localhost:4000/api/v1/health
bun test

cd ../client
bun run dev                          # http://localhost:3000
```

Server image (build from the repo root): `docker build -f server/Dockerfile -t portfolio-server .`
