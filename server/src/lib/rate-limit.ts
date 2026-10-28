import { Elysia } from "elysia";
import { env } from "../config/env";
import { getClientIp } from "./client-ip";
import { HttpError } from "./errors";

type Options = { limit: number; windowMs: number };

/** Fixed-window counter kept in memory. Fine for a single server instance. */
export function createRateLimiter({ limit, windowMs }: Options) {
  const hits = new Map<string, { count: number; resetAt: number }>();

  const sweep = setInterval(() => {
    const now = Date.now();
    for (const [key, entry] of hits) if (entry.resetAt <= now) hits.delete(key);
  }, windowMs);
  sweep.unref();

  return {
    consume(key: string, now = Date.now()) {
      let entry = hits.get(key);
      if (!entry || entry.resetAt <= now) {
        entry = { count: 0, resetAt: now + windowMs };
        hits.set(key, entry);
      }
      entry.count++;
      return {
        allowed: entry.count <= limit,
        retryAfterSec: Math.ceil((entry.resetAt - now) / 1000),
      };
    },
  };
}

/** Per-route limit keyed by client IP: `{ rateLimit: { limit: 3, windowMs: 3_600_000 } }`. */
export const rateLimit = new Elysia({ name: "rate-limit" }).macro({
  rateLimit(options: Options) {
    const limiter = createRateLimiter(options);
    return {
      beforeHandle({ request, server, set }) {
        const ip = getClientIp(
          request.headers.get("x-forwarded-for"),
          server?.requestIP(request)?.address,
          env.TRUSTED_PROXY_HOPS,
        );
        const result = limiter.consume(ip);
        if (!result.allowed) {
          set.headers["retry-after"] = String(result.retryAfterSec);
          throw new HttpError(429, "RATE_LIMITED", "Too many requests, try again later");
        }
      },
    };
  },
});
