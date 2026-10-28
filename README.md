# Portfolio

Decoupled full-stack portfolio.

| Path      | What                                                     |
| --------- | -------------------------------------------------------- |
| `server/` | REST API — Bun, ElysiaJS, Prisma, PostgreSQL (`/api/v1`) |
| `client/` | UI — Next.js (App Router), Tailwind CSS, TypeScript      |

## Local development

```bash
docker compose up -d                 # Postgres on localhost:5433

cd server
cp .env.example .env
bun install
bun run db:migrate                   # apply migrations + generate client
bun run dev                          # http://localhost:4000/api/v1/health

cd ../client
npm install
npm run dev                          # http://localhost:3000
```
