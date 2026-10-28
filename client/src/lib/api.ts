import "server-only";
import { treaty } from "@elysiajs/eden";
import type { App } from "portfolio-server/app";

// Typed client for Server Components. Browser calls will go through the /api rewrite (added later).
export const api = treaty<App>(process.env.API_URL ?? "http://localhost:4000");
