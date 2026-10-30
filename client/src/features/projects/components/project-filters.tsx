import { Search, X } from "lucide-react";
import Form from "next/form";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { TagWithCount } from "../api";
import {
  filtersToHref,
  hasFilters,
  toggleFeatured,
  toggleTag,
  type ProjectFilters,
} from "../filters";

const categoryLabels: Record<TagWithCount["category"], string> = {
  language: "Languages",
  framework: "Frameworks",
  tool: "Tools",
  domain: "Domains",
};

/** Tag chips grouped by category. Each chip is a link that toggles its tag in the URL. */
export function TagFilter({
  tags,
  filters,
}: {
  tags: TagWithCount[];
  filters: ProjectFilters;
}) {
  const groups = Object.groupBy(tags, (t) => t.category);

  return (
    <div className="space-y-3">
      {(Object.keys(categoryLabels) as TagWithCount["category"][]).map(
        (category) => {
          const group = groups[category];
          if (!group?.length) return null;
          return (
            <div key={category} className="flex flex-wrap items-baseline gap-2">
              <span className="w-24 shrink-0 text-xs font-medium text-muted-foreground">
                {categoryLabels[category]}
              </span>
              <ul className="flex flex-wrap gap-1.5">
                {group.map((tag) => {
                  const selected = filters.tags.includes(tag.slug);
                  return (
                    <li key={tag.slug}>
                      <Badge asChild variant={selected ? "default" : "outline"}>
                        <Link
                          href={filtersToHref(toggleTag(filters, tag.slug))}
                          aria-label={`${selected ? "Remove" : "Add"} filter: ${tag.name}`}
                          data-selected={selected}
                          scroll={false}
                        >
                          {tag.name}
                          <span className="opacity-60">{tag.projectCount}</span>
                        </Link>
                      </Badge>
                    </li>
                  );
                })}
              </ul>
            </div>
          );
        },
      )}
    </div>
  );
}

/** GET form: works without JavaScript; next/form upgrades it to a client navigation. */
export function SearchForm({ filters }: { filters: ProjectFilters }) {
  return (
    <Form action="/projects" className="flex gap-2" role="search">
      {filters.tags.map((tag) => (
        <input key={tag} type="hidden" name="tag" value={tag} />
      ))}
      {filters.featured && <input type="hidden" name="featured" value="true" />}
      <Input
        name="q"
        type="search"
        defaultValue={filters.q}
        maxLength={50}
        placeholder="Search title or summary"
        aria-label="Search projects"
        className="max-w-xs"
      />
      <Button type="submit" variant="outline" size="icon" aria-label="Search">
        <Search />
      </Button>
    </Form>
  );
}

export function FilterSummary({
  count,
  filters,
}: {
  count: number;
  filters: ProjectFilters;
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
      <p aria-live="polite">
        {count} {count === 1 ? "project" : "projects"}
      </p>
      <div className="flex items-center gap-2">
        <Button
          asChild
          variant={filters.featured ? "default" : "outline"}
          size="sm"
        >
          <Link
            href={filtersToHref(toggleFeatured(filters))}
            scroll={false}
            data-selected={Boolean(filters.featured)}
          >
            Featured only
          </Link>
        </Button>
        {hasFilters(filters) && (
          <Button asChild variant="ghost" size="sm">
            <Link href="/projects" scroll={false}>
              <X /> Clear filters
            </Link>
          </Button>
        )}
      </div>
    </div>
  );
}

export function EmptyState({ filters }: { filters: ProjectFilters }) {
  const lastTag = filters.tags.at(-1);
  return (
    <div className="rounded-lg border border-dashed p-10 text-center">
      <p className="font-medium">No projects match these filters.</p>
      <p className="mt-1 text-sm text-muted-foreground">
        Projects must have every selected tag.
      </p>
      <div className="mt-4 flex justify-center gap-2">
        {lastTag && (
          <Button asChild variant="outline" size="sm">
            <Link
              href={filtersToHref(toggleTag(filters, lastTag))}
              scroll={false}
            >
              Remove last tag
            </Link>
          </Button>
        )}
        <Button asChild variant="ghost" size="sm">
          <Link href="/projects">Clear filters</Link>
        </Button>
      </div>
    </div>
  );
}

export function InvalidFilters() {
  return (
    <div
      role="alert"
      className="rounded-lg border border-dashed p-10 text-center"
    >
      <p className="font-medium">
        Some filters in this link aren&apos;t valid.
      </p>
      <Button asChild variant="outline" size="sm" className="mt-4">
        <Link href="/projects">Show all projects</Link>
      </Button>
    </div>
  );
}
