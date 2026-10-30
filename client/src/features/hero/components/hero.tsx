import { ArrowRight, Code, FileText, Mail } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { profile } from "@/content/profile";

export function Hero() {
  return (
    <section aria-labelledby="hero-title" className="py-16 sm:py-24">
      <p className="text-sm text-muted-foreground">Hi, I&apos;m</p>
      <h1
        id="hero-title"
        className="mt-2 text-4xl font-bold tracking-tight sm:text-6xl"
      >
        {profile.name}{" "}
        <span className="text-muted-foreground">({profile.nameLocal})</span>
      </h1>
      <p className="mt-6 max-w-2xl text-lg text-muted-foreground sm:text-xl">
        {profile.headline}
      </p>

      <div className="mt-8 flex flex-wrap gap-3">
        <Button asChild size="lg">
          <Link href="/projects">
            View projects <ArrowRight />
          </Link>
        </Button>
        <Button asChild size="lg" variant="outline">
          <a href={profile.githubUrl} target="_blank" rel="noreferrer">
            <Code /> GitHub
          </a>
        </Button>
        <Button asChild size="lg" variant="outline">
          <a href={`mailto:${profile.email}`}>
            <Mail /> Email
          </a>
        </Button>
        {profile.resumeUrl && (
          <Button asChild size="lg" variant="outline">
            <a href={profile.resumeUrl} download>
              <FileText /> Resume
            </a>
          </Button>
        )}
        {profile.linkedinUrl && (
          <Button asChild size="lg" variant="outline">
            <a href={profile.linkedinUrl} target="_blank" rel="noreferrer">
              LinkedIn
            </a>
          </Button>
        )}
      </div>
    </section>
  );
}
