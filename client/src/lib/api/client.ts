import "server-only";
import createClient from "openapi-fetch";
import type { components, paths } from "./schema";

// Typed client for Server Components. Types come from ../server/api/openapi.yaml
// (`bun run gen:api`). Browser calls will go through the /api rewrite (added later).
export const api = createClient<paths>({
  baseUrl: `${process.env.API_URL ?? "http://localhost:4000"}/api/v1`,
});

export type ApiError = components["schemas"]["Error"];
