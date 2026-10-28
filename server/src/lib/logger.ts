import { Elysia } from "elysia";
import { env } from "../config/env";

const startedAt = new WeakMap<Request, number>();

/** One line per request: method, path, status, duration. Silent under `bun test`. */
export const requestLogger = new Elysia({ name: "request-logger" })
  .onRequest(({ request }) => {
    startedAt.set(request, performance.now());
  })
  .onAfterResponse({ as: "global" }, ({ request, set }) => {
    if (env.NODE_ENV === "test") return;
    const ms = performance.now() - (startedAt.get(request) ?? performance.now());
    console.log(
      `${request.method} ${new URL(request.url).pathname} ${set.status ?? 200} ${ms.toFixed(1)}ms`,
    );
  });
