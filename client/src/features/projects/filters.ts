// Project list filters live in the URL (/projects?tag=go&tag=postgresql&q=api&featured=true)
// so every filtered view is linkable and works without client-side JavaScript.

export type ProjectFilters = {
  tags: string[]; // a project must have ALL of these
  q?: string;
  featured?: true;
};

type SearchParams = Record<string, string | string[] | undefined>;

export function parseFilters(sp: SearchParams): ProjectFilters {
  const tags = [sp.tag ?? []].flat().filter(Boolean);
  const q = first(sp.q)?.trim();
  return {
    tags: [...new Set(tags)],
    ...(q ? { q } : {}),
    ...(first(sp.featured) === "true" ? { featured: true as const } : {}),
  };
}

export function filtersToHref(f: ProjectFilters): string {
  const params = new URLSearchParams();
  for (const tag of f.tags) params.append("tag", tag);
  if (f.q) params.set("q", f.q);
  if (f.featured) params.set("featured", "true");
  const qs = params.toString();
  return qs ? `/projects?${qs}` : "/projects";
}

/** Adds the tag if absent, removes it if present; other filters are kept. */
export function toggleTag(f: ProjectFilters, slug: string): ProjectFilters {
  const tags = f.tags.includes(slug)
    ? f.tags.filter((t) => t !== slug)
    : [...f.tags, slug];
  return { ...f, tags };
}

export function toggleFeatured(f: ProjectFilters): ProjectFilters {
  const { featured, ...rest } = f;
  return featured ? rest : { ...rest, featured: true };
}

export function hasFilters(f: ProjectFilters): boolean {
  return f.tags.length > 0 || Boolean(f.q) || Boolean(f.featured);
}

function first(v: string | string[] | undefined): string | undefined {
  return Array.isArray(v) ? v[0] : v;
}
