import { Elysia } from "elysia";
import { healthRoutes } from "@/modules/health/health.routes";

// Error envelope shared by every endpoint: { error: { code, message, fieldErrors? } }
export const app = new Elysia({ prefix: "/api/v1" })
  .onError(({ code, error, set }) => {
    switch (code) {
      case "NOT_FOUND":
        set.status = 404;
        return { error: { code: "NOT_FOUND", message: "Route not found" } };
      case "VALIDATION":
      case "PARSE":
        set.status = 400;
        return { error: { code: "VALIDATION", message: error.message } };
      default:
        console.error(error);
        set.status = 500;
        return { error: { code: "INTERNAL", message: "Internal server error" } };
    }
  })
  .use(healthRoutes);

export type App = typeof app;
