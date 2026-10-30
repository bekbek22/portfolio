import { ArrowLeft, Code, ExternalLink } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getProject, listProjects } from "@/features/projects/api";
import { ProjectBody } from "@/features/projects/components/project-body";
import { filtersToHref } from "@/features/projects/filters";

// Prerender every published project; slugs published later render on first visit.
export async function generateStaticParams() {
  const result = await listProjects({ tags: [] });
  return result.ok ? result.projects.map((p) => ({ slug: p.slug })) : [];
}

export async function generateMetadata({
  params,
}: PageProps<"/projects/[slug]">): Promise<Metadata> {
  const project = await getProject((await params).slug);
  return project
    ? { title: project.title, description: project.summary }
    : { title: "Project not found" };
}

export default async function ProjectPage({
  params,
}: PageProps<"/projects/[slug]">) {
  const project = await getProject((await params).slug);
  if (!project) notFound();

  return (
    <article className="py-12">
      <Link
        href="/projects"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" /> All projects
      </Link>

      <h1 className="mt-6 text-3xl font-bold tracking-tight sm:text-4xl">
        {project.title}
      </h1>
      <p className="mt-3 text-lg text-muted-foreground">{project.summary}</p>

      {project.tags.length > 0 && (
        <ul className="mt-6 flex flex-wrap gap-1.5" aria-label="Tags">
          {project.tags.map((tag) => (
            <li key={tag.slug}>
              <Badge asChild variant="outline">
                <Link href={filtersToHref({ tags: [tag.slug] })}>
                  {tag.name}
                </Link>
              </Badge>
            </li>
          ))}
        </ul>
      )}

      {(project.repoUrl || project.liveUrl) && (
        <div className="mt-6 flex gap-3">
          {project.repoUrl && (
            <Button asChild variant="outline">
              <a href={project.repoUrl} target="_blank" rel="noreferrer">
                <Code /> Source
              </a>
            </Button>
          )}
          {project.liveUrl && (
            <Button asChild>
              <a href={project.liveUrl} target="_blank" rel="noreferrer">
                <ExternalLink /> Live demo
              </a>
            </Button>
          )}
        </div>
      )}

      <div className="mt-10 max-w-3xl">
        <ProjectBody markdown={project.body} />
      </div>
    </article>
  );
}
