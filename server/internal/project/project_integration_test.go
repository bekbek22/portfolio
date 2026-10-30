package project_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/bekbek101/portfolio/server/internal/database/sqlc"
	"github.com/bekbek101/portfolio/server/internal/database/testdb"
	"github.com/bekbek101/portfolio/server/internal/project"
)

type fixture struct {
	slug, title, summary string
	published, featured  bool
	sortOrder            int32
	tags                 []string
}

// Tag slugs are prefixed so these rows never contend with the seed test's rows when
// test packages run concurrently against the same database.
var fixtureTags = []sqlc.UpsertTagParams{
	{Slug: "fx-go", Name: "Go", Category: sqlc.TagCategoryLanguage},
	{Slug: "fx-typescript", Name: "TypeScript", Category: sqlc.TagCategoryLanguage},
	{Slug: "fx-rust", Name: "Rust", Category: sqlc.TagCategoryLanguage}, // only on a draft
	{Slug: "fx-postgresql", Name: "PostgreSQL", Category: sqlc.TagCategoryTool},
}

var fixtures = []fixture{
	{slug: "fx-zeta-notes", title: "Zeta Notes", summary: "No tags at all", published: true, sortOrder: 1},
	{slug: "fx-alpha-service", title: "Alpha Go Service", summary: "Go and PostgreSQL", published: true, featured: true, sortOrder: 2, tags: []string{"fx-go", "fx-postgresql"}},
	{slug: "fx-beta-tool", title: "Beta Tool", summary: "100% coverage_report generator", published: true, sortOrder: 3, tags: []string{"fx-go"}},
	{slug: "fx-gamma-web", title: "Gamma Web", summary: "TypeScript frontend", published: true, sortOrder: 4, tags: []string{"fx-postgresql", "fx-typescript"}},
	// A draft that matches almost every filter: it must never be visible.
	{slug: "fx-alpha-draft", title: "Alpha Draft", summary: "Go and PostgreSQL 100%", featured: true, sortOrder: 0, tags: []string{"fx-go", "fx-postgresql", "fx-rust"}},
}

func newFixtureService(t *testing.T) *project.Service {
	t.Helper()
	tx := testdb.Tx(t)
	ctx := context.Background()
	q := sqlc.New(tx)

	tagIDs := map[string]uuid.UUID{}
	for _, tag := range fixtureTags {
		id, err := q.UpsertTag(ctx, tag)
		if err != nil {
			t.Fatalf("insert tag %s: %v", tag.Slug, err)
		}
		tagIDs[tag.Slug] = id
	}
	for _, f := range fixtures {
		id, err := q.UpsertProject(ctx, sqlc.UpsertProjectParams{
			Slug: f.slug, Title: f.title, Summary: f.summary, Body: "body of " + f.slug,
			Featured: f.featured, Published: f.published, SortOrder: f.sortOrder,
		})
		if err != nil {
			t.Fatalf("insert project %s: %v", f.slug, err)
		}
		ids := make([]uuid.UUID, 0, len(f.tags))
		for _, s := range f.tags {
			ids = append(ids, tagIDs[s])
		}
		if err := q.AddProjectTags(ctx, sqlc.AddProjectTagsParams{ProjectID: id, TagIDs: ids}); err != nil {
			t.Fatalf("tag project %s: %v", f.slug, err)
		}
	}
	return project.NewService(q)
}

func slugs(cards []project.Card) []string {
	out := make([]string, 0, len(cards))
	for _, c := range cards {
		out = append(out, c.Slug)
	}
	return out
}

func TestList(t *testing.T) {
	svc := newFixtureService(t)
	yes, no := true, false

	// Subtests share one transaction, so they run sequentially.
	tests := []struct {
		name   string
		filter project.Filter
		want   []string
	}{
		{"no filter: all published, by sort order", project.Filter{}, []string{"fx-zeta-notes", "fx-alpha-service", "fx-beta-tool", "fx-gamma-web"}},

		{"one tag", project.Filter{Tags: []string{"fx-go"}}, []string{"fx-alpha-service", "fx-beta-tool"}},
		{"two tags: intersection", project.Filter{Tags: []string{"fx-go", "fx-postgresql"}}, []string{"fx-alpha-service"}},
		{"two tags: order irrelevant", project.Filter{Tags: []string{"fx-postgresql", "fx-go"}}, []string{"fx-alpha-service"}},
		{"duplicate tag counts once", project.Filter{Tags: []string{"fx-go", "fx-go"}}, []string{"fx-alpha-service", "fx-beta-tool"}},
		{"three tags: no project has all", project.Filter{Tags: []string{"fx-go", "fx-postgresql", "fx-typescript"}}, []string{}},
		{"tag only on a draft", project.Filter{Tags: []string{"fx-rust"}}, []string{}},
		{"unknown tag", project.Filter{Tags: []string{"fx-unknown"}}, []string{}},

		{"featured", project.Filter{Featured: &yes}, []string{"fx-alpha-service"}},
		{"not featured", project.Filter{Featured: &no}, []string{"fx-zeta-notes", "fx-beta-tool", "fx-gamma-web"}},

		{"search is case-insensitive", project.Filter{Search: "ALPHA"}, []string{"fx-alpha-service"}},
		{"search matches summary", project.Filter{Search: "frontend"}, []string{"fx-gamma-web"}},
		{"search % is literal", project.Filter{Search: "%"}, []string{"fx-beta-tool"}},
		{"search _ is literal", project.Filter{Search: "_"}, []string{"fx-beta-tool"}},
		{"search with literal % and _", project.Filter{Search: "100% coverage_"}, []string{"fx-beta-tool"}},
		{"search backslash is literal", project.Filter{Search: `\`}, []string{}},

		{"tag + search + featured", project.Filter{Tags: []string{"fx-postgresql"}, Search: "go", Featured: &yes}, []string{"fx-alpha-service"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards, err := svc.List(t.Context(), tt.filter)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if got := slugs(cards); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("cards carry tags sorted by name", func(t *testing.T) {
		cards, err := svc.List(t.Context(), project.Filter{Tags: []string{"fx-go", "fx-postgresql"}})
		if err != nil || len(cards) != 1 {
			t.Fatalf("List: %v %v", cards, err)
		}
		want := []project.Tag{{Slug: "fx-go", Name: "Go"}, {Slug: "fx-postgresql", Name: "PostgreSQL"}}
		if !slices.Equal(cards[0].Tags, want) {
			t.Errorf("tags = %v, want %v", cards[0].Tags, want)
		}
	})

	t.Run("untagged project has an empty, non-nil tag list", func(t *testing.T) {
		cards, _ := svc.List(t.Context(), project.Filter{Search: "Zeta"})
		if len(cards) != 1 || cards[0].Tags == nil || len(cards[0].Tags) != 0 {
			t.Errorf("tags = %#v, want empty slice", cards)
		}
	})
}

func TestGet(t *testing.T) {
	svc := newFixtureService(t)

	t.Run("published", func(t *testing.T) {
		d, err := svc.Get(t.Context(), "fx-alpha-service")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if d.Title != "Alpha Go Service" || d.Body != "body of fx-alpha-service" || len(d.Tags) != 2 || d.UpdatedAt.IsZero() {
			t.Errorf("unexpected detail: %+v", d)
		}
	})

	for _, slug := range []string{"fx-alpha-draft", "fx-missing"} {
		t.Run(slug+" is not found", func(t *testing.T) {
			if _, err := svc.Get(t.Context(), slug); !errors.Is(err, project.ErrNotFound) {
				t.Errorf("err = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestTags(t *testing.T) {
	svc := newFixtureService(t)

	got, err := svc.Tags(t.Context())
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	// Only fixture tags; rust is excluded because its only project is a draft, and the
	// draft does not add to the go/postgresql counts. Order: category enum, then name.
	var fx []project.TagCount
	for _, tc := range got {
		if len(tc.Slug) > 3 && tc.Slug[:3] == "fx-" {
			fx = append(fx, tc)
		}
	}
	want := []project.TagCount{
		{Tag: project.Tag{Slug: "fx-go", Name: "Go"}, Category: project.TagCategoryLanguage, ProjectCount: 2},
		{Tag: project.Tag{Slug: "fx-typescript", Name: "TypeScript"}, Category: project.TagCategoryLanguage, ProjectCount: 1},
		{Tag: project.Tag{Slug: "fx-postgresql", Name: "PostgreSQL"}, Category: project.TagCategoryTool, ProjectCount: 2},
	}
	if !slices.Equal(fx, want) {
		t.Errorf("tags = %+v\nwant   %+v", fx, want)
	}
}
