import { Elysia } from "elysia";
import { errorHandler } from "./lib/errors";
import { requestLogger } from "./lib/logger";
import { healthRoutes } from "./modules/health/health.routes";

export const app = new Elysia({ prefix: "/api/v1" })
  .use(requestLogger)
  .use(errorHandler)
  .use(healthRoutes);

export type App = typeof app;
