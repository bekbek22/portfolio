import { ArrowRight } from "lucide-react";
import Link from "next/link";
import { Hero } from "@/features/hero/components/hero";
import { listProjects } from "@/features/projects/api";
import { ProjectGrid } from "@/features/projects/components/project-card";
import { SkillsSection } from "@/features/skills/components/skills-section";

export default async function HomePage() {
  const featured = await listProjects({ tags: [], featured: true });

  return (
    <>
      <Hero />
      <SkillsSection />
      <section aria-labelledby="featured-title" className="py-12">
        <div className="flex items-baseline justify-between gap-4">
          <h2
            id="featured-title"
            className="text-2xl font-semibold tracking-tight"
          >
            Featured projects
          </h2>
          <Link
            href="/projects"
            className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
          >
            All projects <ArrowRight className="size-4" />
          </Link>
        </div>
        <div className="mt-6">
          {featured.ok && <ProjectGrid projects={featured.projects} />}
        </div>
      </section>
    </>
  );
}
