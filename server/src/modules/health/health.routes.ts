import { Elysia } from "elysia";
import { prisma } from "@/lib/prisma";

export const healthRoutes = new Elysia({ prefix: "/health" }).get("/", async ({ set }) => {
  try {
    await prisma.$queryRaw`SELECT 1`;
    return { status: "ok", db: "up" } as const;
  } catch {
    set.status = 503;
    return { status: "degraded", db: "down" } as const;
  }
});
