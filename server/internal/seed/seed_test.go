package seed_test

import (
	"context"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/bekbek101/portfolio/server/internal/database/sqlc"
	"github.com/bekbek101/portfolio/server/internal/database/testdb"
	"github.com/bekbek101/portfolio/server/internal/project"
	"github.com/bekbek101/portfolio/server/internal/seed"
)

type counts struct{ projects, tags, links int }

func countSeeded(t *testing.T, tx pgx.Tx) counts {
	t.Helper()
	var c counts
	err := tx.QueryRow(context.Background(), `
		SELECT
			(SELECT count(*) FROM projects),
			(SELECT count(*) FROM tags),
			(SELECT count(*) FROM project_tags)`).Scan(&c.projects, &c.tags, &c.links)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	return c
}

func TestRunIsIdempotent(t *testing.T) {
	tx := testdb.Tx(t)
	before := countSeeded(t, tx)

	if err := seed.Run(t.Context(), tx); err != nil {
		t.Fatalf("first run: %v", err)
	}
	first := countSeeded(t, tx)

	if err := seed.Run(t.Context(), tx); err != nil {
		t.Fatalf("second run: %v", err)
	}
	second := countSeeded(t, tx)

	added := counts{first.projects - before.projects, first.tags - before.tags, first.links - before.links}
	if want := (counts{projects: 5, tags: 20, links: 26}); added != want {
		t.Errorf("first run added %+v, want %+v", added, want)
	}
	if second != first {
		t.Errorf("second run changed counts: %+v -> %+v", first, second)
	}

	// The seeded data is served correctly through the real service.
	svc := project.NewService(sqlc.New(tx))
	featured := true
	tests := []struct {
		name   string
		filter project.Filter
		want   []string
	}{
		{"go AND postgresql", project.Filter{Tags: []string{"go", "postgresql"}}, []string{"go-cvm-tax-id-integration"}},
		{"postgresql", project.Filter{Tags: []string{"postgresql"}}, []string{"go-cvm-tax-id-integration", "elysia-bun-backend", "costsheet-online-pipeline"}},
		{"featured", project.Filter{Featured: &featured}, []string{"go-cvm-tax-id-integration", "elysia-bun-backend", "k8s-zero-downtime-deployment"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards, err := svc.List(t.Context(), tt.filter)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			var got []string
			for _, c := range cards {
				got = append(got, c.Slug)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
