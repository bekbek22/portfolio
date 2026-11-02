# syntax=docker/dockerfile:1
# Build from server/:  docker build -t portfolio-server server

# ── build: static binary, no cgo ──────────────────────────────────────────────
FROM golang:1.27.3-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ── runtime: distroless (no shell, no package manager), non-root ──────────────
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/api /api

ENV APP_ENV=production PORT=4000
EXPOSE 4000
USER nonroot:nonroot

# No shell or curl in distroless: the binary checks itself.
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD ["/api", "healthcheck"]

# CMD, not ENTRYPOINT: platforms can run `/api migrate up` as a one-off command in this image.
CMD ["/api", "serve"]
