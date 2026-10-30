export type SkillGroup = { title: string; skills: string[] };

export const skillGroups: SkillGroup[] = [
  {
    title: "Backend",
    skills: [
      "Go (Gin/GORM)",
      "ElysiaJS",
      "Bun",
      "Node.js/TypeScript",
      "Python",
      "RESTful API Design",
      "Microservices",
      "CVM Integration",
    ],
  },
  {
    title: "Database & Storage",
    skills: [
      "PostgreSQL",
      "MongoDB (Aggregation Pipelines, Mongoose)",
      "MinIO / AWS S3",
    ],
  },
  {
    title: "DevOps & Infrastructure",
    skills: [
      "Docker",
      "Kubernetes (Zero-Downtime Deployments)",
      "Git / GitLab CI/CD",
      "Linux Server Management",
    ],
  },
  {
    title: "Frontend",
    skills: ["Next.js", "React", "Tailwind CSS"],
  },
];
