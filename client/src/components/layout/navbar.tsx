import Link from "next/link";
import { profile } from "@/content/profile";
import { NavLinks } from "./nav-links";
import { ThemeToggle } from "./theme-toggle";

export function Navbar() {
  return (
    <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
      <nav
        aria-label="Main"
        className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4"
      >
        <Link href="/" className="font-semibold tracking-tight">
          {profile.name}
        </Link>
        <div className="flex items-center gap-2">
          <NavLinks />
          <ThemeToggle />
        </div>
      </nav>
    </header>
  );
}
