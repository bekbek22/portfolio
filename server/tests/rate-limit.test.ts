import { describe, expect, test } from "bun:test";
import { Elysia } from "elysia";
import { errorHandler } from "../src/lib/errors";
import { createRateLimiter, rateLimit } from "../src/lib/rate-limit";

describe("createRateLimiter", () => {
  test("allows `limit` hits per window, then blocks until the window resets", () => {
    const limiter = createRateLimiter({ limit: 2, windowMs: 1000 });
    expect(limiter.consume("a", 0).allowed).toBe(true);
    expect(limiter.consume("a", 100).allowed).toBe(true);
    const blocked = limiter.consume("a", 200);
    expect(blocked).toEqual({ allowed: false, retryAfterSec: 1 });
    expect(limiter.consume("a", 1000).allowed).toBe(true); // new window
  });

  test("tracks keys independently", () => {
    const limiter = createRateLimiter({ limit: 1, windowMs: 1000 });
    expect(limiter.consume("a", 0).allowed).toBe(true);
    expect(limiter.consume("b", 0).allowed).toBe(true);
    expect(limiter.consume("a", 0).allowed).toBe(false);
  });
});

describe("rateLimit macro", () => {
  const app = new Elysia()
    .use(errorHandler)
    .use(rateLimit)
    .post("/limited", () => "ok", { rateLimit: { limit: 3, windowMs: 60_000 } })
    .get("/free", () => "ok");

  const post = () => app.handle(new Request("http://localhost/limited", { method: "POST" }));

  test("returns 429 with Retry-After and the error envelope once the limit is hit", async () => {
    for (let i = 0; i < 3; i++) expect((await post()).status).toBe(200);

    const res = await post();
    expect(res.status).toBe(429);
    expect(Number(res.headers.get("retry-after"))).toBeGreaterThan(0);
    expect(await res.json()).toEqual({
      error: { code: "RATE_LIMITED", message: "Too many requests, try again later" },
    });
  });

  test("does not affect routes without the option", async () => {
    const res = await app.handle(new Request("http://localhost/free"));
    expect(res.status).toBe(200);
  });
});
