import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function NotFound() {
  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col items-start justify-center px-4 py-24">
      <p className="text-sm text-muted-foreground">404</p>
      <h1 className="mt-2 text-3xl font-bold tracking-tight">Page not found</h1>
      <p className="mt-2 text-muted-foreground">
        The page or project you&apos;re looking for doesn&apos;t exist.
      </p>
      <div className="mt-6 flex gap-3">
        <Button asChild>
          <Link href="/">Home</Link>
        </Button>
        <Button asChild variant="outline">
          <Link href="/projects">All projects</Link>
        </Button>
      </div>
    </main>
  );
}
