import { Star } from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { ProjectCard as Project } from "../api";
import { filtersToHref } from "../filters";

export function ProjectCard({ project }: { project: Project }) {
  return (
    <Card className="relative h-full transition-colors hover:border-foreground/30">
      <CardHeader>
        <CardTitle className="flex items-start justify-between gap-2">
          {/* The stretched link makes the whole card clickable; tag links sit above it. */}
          <Link
            href={`/projects/${project.slug}`}
            className="after:absolute after:inset-0"
          >
            {project.title}
          </Link>
          {project.featured && (
            <Star
              className="size-4 shrink-0 fill-current text-amber-500"
              aria-label="Featured"
            />
          )}
        </CardTitle>
        <CardDescription>{project.summary}</CardDescription>
      </CardHeader>
      {project.tags.length > 0 && (
        <CardContent className="mt-auto">
          <ul className="relative flex flex-wrap gap-1.5">
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
        </CardContent>
      )}
    </Card>
  );
}

export function ProjectGrid({ projects }: { projects: Project[] }) {
  return (
    <ul aria-label="Projects" className="grid gap-4 sm:grid-cols-2">
      {projects.map((p) => (
        <li key={p.slug}>
          <ProjectCard project={p} />
        </li>
      ))}
    </ul>
  );
}
