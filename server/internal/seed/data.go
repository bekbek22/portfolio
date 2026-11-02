package seed

import "github.com/bekbek22/portfolio/server/internal/database/sqlc"

// Placeholder content based on the owner's real work. Repo/live URLs are deliberately
// nil so no invented links are published; fill them in once the repos are public.

type tagSeed struct {
	slug, name string
	category   sqlc.TagCategory
}

type projectSeed struct {
	slug, title, summary, body string
	featured                   bool
	sortOrder                  int32
	tags                       []string // tag slugs
}

var tagSeeds = []tagSeed{
	{"go", "Go", sqlc.TagCategoryLanguage},
	{"typescript", "TypeScript", sqlc.TagCategoryLanguage},
	{"luau", "Luau", sqlc.TagCategoryLanguage},

	{"gin", "Gin", sqlc.TagCategoryFramework},
	{"elysiajs", "ElysiaJS", sqlc.TagCategoryFramework},
	{"mongoose", "Mongoose", sqlc.TagCategoryFramework},

	{"postgresql", "PostgreSQL", sqlc.TagCategoryTool},
	{"mongodb", "MongoDB", sqlc.TagCategoryTool},
	{"prisma", "Prisma", sqlc.TagCategoryTool},
	{"bun", "Bun", sqlc.TagCategoryTool},
	{"docker", "Docker", sqlc.TagCategoryTool},
	{"kubernetes", "Kubernetes", sqlc.TagCategoryTool},
	{"gitlab-ci", "GitLab CI", sqlc.TagCategoryTool},
	{"rojo", "Rojo", sqlc.TagCategoryTool},
	{"lune", "Lune", sqlc.TagCategoryTool},

	{"microservices", "Microservices", sqlc.TagCategoryDomain},
	{"rest-api", "REST API", sqlc.TagCategoryDomain},
	{"devops", "DevOps", sqlc.TagCategoryDomain},
	{"data-pipelines", "Data Pipelines", sqlc.TagCategoryDomain},
	{"game-dev", "Game Dev", sqlc.TagCategoryDomain},
}

var projectSeeds = []projectSeed{
	{
		slug:      "go-cvm-tax-id-integration",
		title:     "Go Microservice: CVM Tax ID Integration",
		summary:   "High-performance Go microservice for transaction enrichment and external API integrations, deployed with a zero-downtime strategy.",
		featured:  true,
		sortOrder: 10,
		tags:      []string{"go", "gin", "postgresql", "microservices", "rest-api", "docker", "kubernetes"},
		body: `## Overview
A Go microservice that enriches transactions with tax ID data from an external CVM
integration and exposes the result to other services over a REST API.

## Highlights
- Gin HTTP layer with context-aware handlers and explicit timeouts on every upstream call
- Retries with backoff and idempotent writes so partial upstream failures are safe to replay
- Rolling deployments on Kubernetes with readiness probes and graceful shutdown
`,
	},
	{
		slug:      "elysia-bun-backend",
		title:     "ElysiaJS & Bun High-Speed Backend",
		summary:   "Modular TypeScript API server built on ElysiaJS, Prisma and PostgreSQL, with rate limiting and schema validation.",
		featured:  true,
		sortOrder: 20,
		tags:      []string{"typescript", "elysiajs", "bun", "prisma", "postgresql", "rest-api"},
		body: `## Overview
A modular API server on the Bun runtime with ElysiaJS, organised by feature module.

## Highlights
- Request validation at the edge with a single consistent error envelope
- Per-route in-memory rate limiting keyed by a proxy-aware client IP
- Prisma with PostgreSQL, end-to-end typed client via Eden Treaty
`,
	},
	{
		slug:      "k8s-zero-downtime-deployment",
		title:     "Kubernetes Zero-Downtime Deployment Configuration",
		summary:   "Kubernetes deployment setup with readiness/liveness probes, graceful shutdown and ingress rules for zero-downtime releases.",
		featured:  true,
		sortOrder: 30,
		tags:      []string{"kubernetes", "docker", "gitlab-ci", "devops"},
		body: `## Overview
Deployment configuration that lets services roll out without dropping requests.

## Highlights
- Readiness and liveness probes tuned separately so slow starts don't cause restarts
- preStop hook plus SIGTERM handling to drain in-flight requests before shutdown
- Rolling update strategy with surge capacity, and ingress rules for routing
- GitLab CI pipeline that builds, tags and deploys images
`,
	},
	{
		slug:      "costsheet-online-pipeline",
		title:     "Costsheet Online Backend Pipeline",
		summary:   "Aggregation pipeline across MongoDB and PostgreSQL that computes financial metrics and cost calculations.",
		sortOrder: 40,
		tags:      []string{"typescript", "mongodb", "mongoose", "postgresql", "data-pipelines"},
		body: `## Overview
Backend pipeline that turns raw cost entries into financial metrics for the Costsheet
Online product.

## Highlights
- Multi-stage MongoDB aggregation pipelines for grouping and rolling up cost data
- PostgreSQL for the relational, reporting-facing side of the data
- Calculations kept deterministic so reports can be re-run and compared
`,
	},
	{
		slug:      "roblox-custom-engine-utilities",
		title:     "Roblox Custom Engine Utilities",
		summary:   "Luau game systems with an automated build pipeline using Rojo and a Lune test harness.",
		sortOrder: 50,
		tags:      []string{"luau", "rojo", "lune", "game-dev"},
		body: `## Overview
Reusable Luau game systems developed outside Roblox Studio with a file-based workflow.

## Highlights
- Rojo project layout so game code lives in Git and builds reproducibly
- Lune-based test harness to run game logic tests from the command line
- Automated build pipeline that produces place files from source
`,
	},
}
