package database_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bekbek22/portfolio/server/internal/database/sqlc"
	"github.com/bekbek22/portfolio/server/internal/database/testdb"
)

// The schema enforces its own invariants; these writes must be rejected by Postgres itself.
func TestSchemaConstraints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sql      string
		args     []any
		wantCode string
	}{
		{
			name:     "project slug must be kebab-case",
			sql:      "INSERT INTO projects (slug, title, summary) VALUES ($1, 'T', 'S')",
			args:     []any{"Not A Slug"},
			wantCode: pgerrcode.CheckViolation,
		},
		{
			name:     "project summary is at most 160 chars",
			sql:      "INSERT INTO projects (slug, title, summary) VALUES ('ok', 'T', $1)",
			args:     []any{strings.Repeat("x", 161)},
			wantCode: pgerrcode.CheckViolation,
		},
		{
			name:     "guestbook message is at most 280 chars",
			sql:      "INSERT INTO guestbook_entries (author_github_id, author_name, message) VALUES (1, 'a', $1)",
			args:     []any{strings.Repeat("x", 281)},
			wantCode: pgerrcode.CheckViolation,
		},
		{
			name:     "tag category must be a known enum value",
			sql:      "INSERT INTO tags (slug, name, category) VALUES ('x', 'X', $1)",
			args:     []any{"cooking"},
			wantCode: pgerrcode.InvalidTextRepresentation,
		},
		{
			name:     "project_tags requires an existing project",
			sql:      "INSERT INTO project_tags (project_id, tag_id) VALUES (gen_random_uuid(), gen_random_uuid())",
			wantCode: pgerrcode.ForeignKeyViolation,
		},
		{
			name:     "analytics country is an ISO 3166 alpha-2 code",
			sql:      "INSERT INTO analytics_events (type, path, device, visitor_hash, country) VALUES ('page_view', '/', 'desktop', 'h', $1)",
			args:     []any{"tha"},
			wantCode: pgerrcode.CheckViolation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := testdb.Tx(t)

			_, err := tx.Exec(t.Context(), tt.sql, tt.args...)

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf("expected Postgres error %s, got %v", tt.wantCode, err)
			}
			if pgErr.Code != tt.wantCode {
				t.Errorf("error code = %s (%s), want %s", pgErr.Code, pgErr.Message, tt.wantCode)
			}
		})
	}
}

func TestGetPublishedProjectBySlug(t *testing.T) {
	t.Parallel()
	tx := testdb.Tx(t)
	q := sqlc.New(tx)

	if _, err := tx.Exec(t.Context(), `
		INSERT INTO projects (slug, title, summary, repo_url, published) VALUES
			('live-project',  'Live',  'Published project', 'https://github.com/x/live', true),
			('draft-project', 'Draft', 'Unpublished draft', NULL,                         false)`); err != nil {
		t.Fatalf("insert fixtures: %v", err)
	}

	tests := []struct {
		slug      string
		wantFound bool
	}{
		{slug: "live-project", wantFound: true},
		{slug: "draft-project", wantFound: false}, // drafts are never public
		{slug: "missing", wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got, err := q.GetPublishedProjectBySlug(t.Context(), tt.slug)

			if !tt.wantFound {
				if !errors.Is(err, pgx.ErrNoRows) {
					t.Fatalf("want pgx.ErrNoRows, got row %+v, err %v", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Title != "Live" || got.RepoURL == nil || *got.RepoURL != "https://github.com/x/live" {
				t.Errorf("unexpected row: %+v", got)
			}
			if got.LiveURL != nil {
				t.Errorf("LiveURL = %q, want nil for NULL column", *got.LiveURL)
			}
		})
	}
}
