import { describe, expect, test } from "bun:test";
import {
  filtersToHref,
  hasFilters,
  parseFilters,
  toggleFeatured,
  toggleTag,
  type ProjectFilters,
} from "./filters";

type SearchParams = Record<string, string | string[] | undefined>;

describe("parseFilters", () => {
  const cases: [SearchParams, ProjectFilters][] = [
    [{}, { tags: [] }],
    [{ tag: "go" }, { tags: ["go"] }],
    [{ tag: ["go", "postgresql"] }, { tags: ["go", "postgresql"] }],
    [{ tag: ["go", "go"] }, { tags: ["go"] }],
    [{ q: "  api  " }, { tags: [], q: "api" }],
    [{ q: "   " }, { tags: [] }],
    [{ featured: "true" }, { tags: [], featured: true }],
    [{ featured: "false" }, { tags: [] }],
    [{ q: ["first", "second"] }, { tags: [], q: "first" }],
  ];
  test.each(cases)("%j", (input, want) => {
    expect(parseFilters(input)).toEqual(want);
  });
});

describe("filtersToHref", () => {
  const cases: [ProjectFilters, string][] = [
    [{ tags: [] }, "/projects"],
    [{ tags: ["go", "postgresql"] }, "/projects?tag=go&tag=postgresql"],
    [{ tags: [], q: "100% c#" }, "/projects?q=100%25+c%23"],
    [
      { tags: ["go"], q: "api", featured: true },
      "/projects?tag=go&q=api&featured=true",
    ],
  ];
  test.each(cases)("%j → %s", (f, want) => {
    expect(filtersToHref(f)).toBe(want);
  });

  test("round-trips through parseFilters", () => {
    const f: ProjectFilters = {
      tags: ["go", "rest-api"],
      q: "a&b=c",
      featured: true,
    };
    const params = new URL(filtersToHref(f), "http://x").searchParams;
    expect(
      parseFilters({
        tag: params.getAll("tag"),
        q: params.get("q") ?? undefined,
        featured: params.get("featured") ?? undefined,
      }),
    ).toEqual(f);
  });
});

describe("toggleTag", () => {
  const base: ProjectFilters = { tags: ["go"], q: "api", featured: true };

  test("adds a tag and keeps other filters", () => {
    expect(toggleTag(base, "postgresql")).toEqual({
      tags: ["go", "postgresql"],
      q: "api",
      featured: true,
    });
  });

  test("removes a selected tag", () => {
    expect(toggleTag(base, "go")).toEqual({
      tags: [],
      q: "api",
      featured: true,
    });
  });

  test("does not mutate its input", () => {
    toggleTag(base, "postgresql");
    expect(base.tags).toEqual(["go"]);
  });
});

describe("toggleFeatured / hasFilters", () => {
  test("toggles featured on and off", () => {
    const on = toggleFeatured({ tags: [] });
    expect(on).toEqual({ tags: [], featured: true });
    expect(toggleFeatured(on)).toEqual({ tags: [] });
  });

  const cases: [ProjectFilters, boolean][] = [
    [{ tags: [] }, false],
    [{ tags: ["go"] }, true],
    [{ tags: [], q: "x" }, true],
    [{ tags: [], featured: true }, true],
  ];
  test.each(cases)("%j → %s", (f, want) => {
    expect(hasFilters(f)).toBe(want);
  });
});
