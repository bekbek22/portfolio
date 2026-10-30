"use client";

import { Button } from "@/components/ui/button";

export default function SiteError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div role="alert" className="py-24">
      <h1 className="text-2xl font-semibold tracking-tight">
        Something went wrong
      </h1>
      <p className="mt-2 text-muted-foreground">
        This content couldn&apos;t be loaded right now. Please try again in a
        moment.
      </p>
      <Button className="mt-6" onClick={reset}>
        Try again
      </Button>
    </div>
  );
}
