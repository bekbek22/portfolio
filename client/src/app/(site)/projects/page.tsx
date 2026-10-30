import type { Metadata } from "next";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { listProjects, listTags } from "@/features/projects/api";
import { ProjectGrid } from "@/features/projects/components/project-card";
import {
  EmptyState,
  FilterSummary,
  InvalidFilters,
  SearchForm,
  TagFilter,
} from "@/features/projects/components/project-filters";
import { parseFilters } from "@/features/projects/filters";

export const metadata: Metadata = {
  title: "Projects",
  description:
    "Backend, infrastructure and tooling projects. Filter by tag to see matching work.",
};

export default function ProjectsPage({ searchParams }: PageProps<"/projects">) {
  return (
    <div className="py-12">
      <h1 className="text-3xl font-bold tracking-tight">Projects</h1>
      <p className="mt-2 text-muted-foreground">
        Select tags to narrow the list. A project must have every selected tag.
      </p>
      {/* searchParams are request-time data: the static shell ships the fallback. */}
      <Suspense fallback={<ExplorerSkeleton />}>
        <ProjectExplorer searchParams={searchParams} />
      </Suspense>
    </div>
  );
}

async function ProjectExplorer({
  searchParams,
}: Pick<PageProps<"/projects">, "searchParams">) {
  const filters = parseFilters(await searchParams);
  const [tags, result] = await Promise.all([listTags(), listProjects(filters)]);

  return (
    <div className="mt-8 space-y-6">
      <TagFilter tags={tags} filters={filters} />
      <SearchForm filters={filters} />
      {!result.ok ? (
        <InvalidFilters />
      ) : (
        <>
          <FilterSummary count={result.projects.length} filters={filters} />
          {result.projects.length ? (
            <ProjectGrid projects={result.projects} />
          ) : (
            <EmptyState filters={filters} />
          )}
        </>
      )}
    </div>
  );
}

function ExplorerSkeleton() {
  return (
    <div
      className="mt-8 space-y-6"
      aria-busy="true"
      aria-label="Loading projects"
    >
      <Skeleton className="h-24 w-full" />
      <div className="grid gap-4 sm:grid-cols-2">
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-40" />
        ))}
      </div>
    </div>
  );
}
