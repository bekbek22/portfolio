# syntax=docker/dockerfile:1
# Build from the repo root (the Bun workspace lockfile lives there):
#   docker build -f server/Dockerfile -t portfolio-server .
# Migrations are not run by this image; apply them from CI/deploy with `bun run db:deploy`.

# ── build: install, generate the Prisma client, bundle to a single file ───────
FROM oven/bun:1.4-slim AS build
WORKDIR /app
COPY package.json bun.lock ./
COPY server/package.json server/
COPY client/package.json client/
RUN bun install --frozen-lockfile --filter portfolio-server

COPY server ./server
WORKDIR /app/server
# `prisma generate` doesn't connect to the DB, but the config requires the var to exist.
RUN DATABASE_URL="postgresql://build:build@localhost:5432/build" bun --bun prisma generate \
 && bun run build

# ── runtime: just Bun + the bundle (no node_modules) ──────────────────────────
FROM oven/bun:1.4-slim AS runtime
WORKDIR /app
ENV NODE_ENV=production PORT=4000
COPY --from=build --chown=bun:bun /app/server/dist ./dist

USER bun
EXPOSE 4000
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD ["bun", "-e", "fetch('http://localhost:' + process.env.PORT + '/api/v1/health').then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"]

CMD ["bun", "dist/index.js"]
