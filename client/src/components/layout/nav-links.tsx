"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const links = [
  { href: "/", label: "Home" },
  { href: "/projects", label: "Projects" },
] as const;

export function NavLinks() {
  const pathname = usePathname();

  return (
    <ul className="flex items-center gap-1">
      {links.map(({ href, label }) => {
        const active =
          href === "/" ? pathname === "/" : pathname.startsWith(href);
        return (
          <li key={href}>
            <Link
              href={href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "rounded-md px-3 py-2 text-sm transition-colors hover:text-foreground",
                active
                  ? "font-medium text-foreground"
                  : "text-muted-foreground",
              )}
            >
              {label}
            </Link>
          </li>
        );
      })}
    </ul>
  );
}
