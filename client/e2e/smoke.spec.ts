import { expect, test } from "@playwright/test";

// Runs against the seeded data from `api seed` (server/internal/seed/data.go).

test("home shows the hero and the three featured projects", async ({
  page,
}) => {
  await page.goto("/");

  await expect(page.getByRole("heading", { level: 1 })).toContainText("Koon");
  const featured = page.getByRole("region", { name: "Featured projects" });
  const cards = featured
    .getByRole("list", { name: "Projects" })
    .locator(":scope > li");
  await expect(cards).toHaveCount(3);
  await expect(featured).toContainText(
    "Go Microservice: CVM Tax ID Integration",
  );
});

test("selecting tags narrows projects to the intersection", async ({
  page,
}) => {
  await page.goto("/projects");
  await expect(page.getByText(/^5 projects$/)).toBeVisible();

  await page.getByRole("link", { name: "Add filter: PostgreSQL" }).click();
  await expect(page).toHaveURL(/\/projects\?tag=postgresql$/);
  await expect(page.getByText(/^3 projects$/)).toBeVisible();

  await page.getByRole("link", { name: "Add filter: Go" }).click();
  await expect(page).toHaveURL(/\/projects\?tag=postgresql&tag=go$/);
  await expect(page.getByText(/^1 project$/)).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Go Microservice: CVM Tax ID Integration" }),
  ).toBeVisible();

  // Deselecting a tag widens the result again; Back restores the previous filter.
  await page.getByRole("link", { name: "Remove filter: Go" }).click();
  await expect(page.getByText(/^3 projects$/)).toBeVisible();
  await page.goBack();
  await expect(page.getByText(/^1 project$/)).toBeVisible();
});

test("search form filters by title or summary", async ({ page }) => {
  await page.goto("/projects");
  await page
    .getByRole("searchbox", { name: "Search projects" })
    .fill("kubernetes");
  await page.getByRole("button", { name: "Search" }).click();

  await expect(page).toHaveURL(/q=kubernetes/);
  await expect(page.getByText(/^1 project$/)).toBeVisible();
});

test("project page renders markdown and links tags back to the list", async ({
  page,
}) => {
  await page.goto("/projects/roblox-custom-engine-utilities");

  await expect(page).toHaveTitle("Roblox Custom Engine Utilities · Koon");
  await expect(
    page.getByRole("heading", { level: 2, name: "Overview" }),
  ).toBeVisible();

  await page
    .getByRole("list", { name: "Tags" })
    .getByRole("link", { name: "Luau" })
    .click();
  await expect(page).toHaveURL(/\/projects\?tag=luau$/);
  await expect(page.getByText(/^1 project$/)).toBeVisible();
});

test("unknown project shows the not-found page and is not indexable", async ({
  page,
}) => {
  await page.goto("/projects/does-not-exist");

  await expect(
    page.getByRole("heading", { name: "Page not found" }),
  ).toBeVisible();
  await expect(page.locator('meta[name="robots"]')).toHaveAttribute(
    "content",
    /noindex/,
  );
});
