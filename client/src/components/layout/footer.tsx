import { profile } from "@/content/profile";

export function Footer() {
  return (
    <footer className="border-t">
      <div className="mx-auto flex max-w-5xl flex-col gap-2 px-4 py-8 text-sm text-muted-foreground sm:flex-row sm:justify-between">
        <p>© {profile.name}</p>
        <ul className="flex gap-4">
          <li>
            <a
              href={profile.githubUrl}
              className="hover:text-foreground"
              rel="noreferrer"
              target="_blank"
            >
              GitHub
            </a>
          </li>
          <li>
            <a
              href={`mailto:${profile.email}`}
              className="hover:text-foreground"
            >
              {profile.email}
            </a>
          </li>
        </ul>
      </div>
    </footer>
  );
}
