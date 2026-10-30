import "server-only";
import { cacheLife, cacheTag } from "next/cache";
import { api } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";
import type { ProjectFilters } from "./filters";

export type ProjectCard = components["schemas"]["ProjectCard"];
export type ProjectDetail = components["schemas"]["ProjectDetail"];
export type TagWithCount = components["schemas"]["TagWithCount"];

export type ProjectListResult =
  | { ok: true; projects: ProjectCard[] }
  | { ok: false; fieldErrors: Record<string, string[]> }; // invalid filter values (422)

// Every fetcher is cached and prerendered; the "projects" tag lets a future admin
// change refresh them on demand with revalidateTag.
const CACHE_TAG = "projects";

export async function listProjects(
  filters: ProjectFilters,
): Promise<ProjectListResult> {
  "use cache";
  cacheLife("hours");
  cacheTag(CACHE_TAG);

  const { data, error, response } = await api.GET("/projects", {
    params: {
      query: {
        tag: filters.tags.length ? filters.tags : undefined,
        q: filters.q,
        featured: filters.featured,
      },
    },
  });
  if (data) return { ok: true, projects: data.data };
  if (response.status === 422)
    return { ok: false, fieldErrors: error?.error.fieldErrors ?? {} };
  throw apiError("list projects", response.status, error);
}

export async function getProject(slug: string): Promise<ProjectDetail | null> {
  "use cache";
  cacheLife("hours");
  cacheTag(CACHE_TAG);

  const { data, error, response } = await api.GET("/projects/{slug}", {
    params: { path: { slug } },
  });
  if (data) return data.data;
  // Unknown or malformed slugs are both "no such project" for a page.
  if (response.status === 404 || response.status === 422) return null;
  throw apiError(`get project ${slug}`, response.status, error);
}

export async function listTags(): Promise<TagWithCount[]> {
  "use cache";
  cacheLife("hours");
  cacheTag(CACHE_TAG);

  const { data, error, response } = await api.GET("/tags");
  if (data) return data.data;
  throw apiError("list tags", response.status, error);
}

function apiError(op: string, status: number, error: unknown): Error {
  return new Error(`API ${op} failed with ${status}: ${JSON.stringify(error)}`);
}
