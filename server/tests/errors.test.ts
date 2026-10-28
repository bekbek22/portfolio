import { describe, expect, test } from "bun:test";
import { Elysia } from "elysia";
import { z } from "zod";
import { errorHandler, HttpError } from "../src/lib/errors";

const app = new Elysia()
  .use(errorHandler)
  .post("/echo", ({ body }) => body, {
    body: z.object({
      name: z.string().min(2),
      email: z.email(),
      nested: z.object({ n: z.number() }),
    }),
  })
  .get("/teapot", () => {
    throw new HttpError(418, "NOT_FOUND", "custom");
  })
  .get("/boom", () => {
    throw new Error("secret internals");
  });

const json = (path: string, body: string) =>
  app.handle(
    new Request(`http://localhost${path}`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body,
    }),
  );

describe("errorHandler", () => {
  test("validation failures → 422 with per-field errors", async () => {
    const res = await json(
      "/echo",
      JSON.stringify({ name: "a", email: "bad", nested: { n: "x" } }),
    );
    expect(res.status).toBe(422);
    const { error } = await res.json();
    expect(error.code).toBe("VALIDATION");
    expect(Object.keys(error.fieldErrors).sort()).toEqual(["email", "name", "nested.n"]);
  });

  test("valid body passes through", async () => {
    const body = { name: "Koon", email: "a@b.dev", nested: { n: 1 } };
    const res = await json("/echo", JSON.stringify(body));
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual(body);
  });

  test("malformed JSON → 400", async () => {
    const res = await json("/echo", "{not json");
    expect(res.status).toBe(400);
    expect((await res.json()).error.code).toBe("VALIDATION");
  });

  test("unknown route → 404 envelope", async () => {
    const res = await app.handle(new Request("http://localhost/nope"));
    expect(res.status).toBe(404);
    expect(await res.json()).toEqual({ error: { code: "NOT_FOUND", message: "Route not found" } });
  });

  test("HttpError keeps its status, code and message", async () => {
    const res = await app.handle(new Request("http://localhost/teapot"));
    expect(res.status).toBe(418);
    expect(await res.json()).toEqual({ error: { code: "NOT_FOUND", message: "custom" } });
  });

  test("unexpected errors → 500 without leaking the message", async () => {
    const res = await app.handle(new Request("http://localhost/boom"));
    expect(res.status).toBe(500);
    expect(await res.json()).toEqual({
      error: { code: "INTERNAL", message: "Internal server error" },
    });
  });
});
