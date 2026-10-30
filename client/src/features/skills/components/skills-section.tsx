import { Badge } from "@/components/ui/badge";
import { skillGroups } from "../data";

export function SkillsSection() {
  return (
    <section aria-labelledby="skills-title" className="py-12">
      <h2 id="skills-title" className="text-2xl font-semibold tracking-tight">
        Skills
      </h2>
      <div className="mt-6 grid gap-8 sm:grid-cols-2">
        {skillGroups.map((group) => (
          <div key={group.title}>
            <h3 className="text-sm font-medium text-muted-foreground">
              {group.title}
            </h3>
            <ul className="mt-3 flex flex-wrap gap-2">
              {group.skills.map((skill) => (
                <li key={skill}>
                  <Badge variant="secondary">{skill}</Badge>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </section>
  );
}
