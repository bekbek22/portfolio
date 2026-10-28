# syntax=docker/dockerfile:1

# ── deps: full install (incl. prisma CLI) + generate the Prisma client ────────
FROM oven/bun:1.4-slim AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY prisma ./prisma
COPY prisma7.config.ts ./
# `prisma generate` doesn't connect to the DB, but the config requires the var to exist.
RUN DATABASE_URL="postgresql://build:build@localhost:5432/build" bun --bun prisma generate

# ── prod-deps: runtime dependencies only ──────────────────────────────────────
FROM oven/bun:1.4-slim AS prod-deps
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --production

# ── runtime ───────────────────────────────────────────────────────────────────
FROM oven/bun:1.4-slim AS runtime
WORKDIR /app
ENV NODE_ENV=production PORT=4000

COPY --from=prod-deps --chown=bun:bun /app/node_modules ./node_modules
COPY --from=build --chown=bun:bun /app/generated ./generated
COPY --chown=bun:bun package.json tsconfig.json ./
COPY --chown=bun:bun src ./src

USER bun
EXPOSE 4000
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD ["bun", "-e", "fetch('http://localhost:' + process.env.PORT + '/api/v1/health').then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"]

CMD ["bun", "src/index.ts"]
