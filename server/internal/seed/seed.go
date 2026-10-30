// Package seed loads the portfolio's initial tags and projects. Running it repeatedly
// converges on the same data: rows are upserted by slug and project tags are replaced.
package seed

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bekbek101/portfolio/server/internal/database/sqlc"
)

// TxBeginner is satisfied by *pgxpool.Pool and by pgx.Tx (which nests via a savepoint).
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Run seeds everything in a single transaction: either all of it lands or none of it.
func Run(ctx context.Context, db TxBeginner) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("seed: begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	q := sqlc.New(tx)

	tagIDs := make(map[string]uuid.UUID, len(tagSeeds))
	for _, t := range tagSeeds {
		id, err := q.UpsertTag(ctx, sqlc.UpsertTagParams{Slug: t.slug, Name: t.name, Category: t.category})
		if err != nil {
			return fmt.Errorf("seed: tag %q: %w", t.slug, err)
		}
		tagIDs[t.slug] = id
	}

	for _, p := range projectSeeds {
		id, err := q.UpsertProject(ctx, sqlc.UpsertProjectParams{
			Slug: p.slug, Title: p.title, Summary: p.summary, Body: p.body,
			Featured: p.featured, Published: true, SortOrder: p.sortOrder,
		})
		if err != nil {
			return fmt.Errorf("seed: project %q: %w", p.slug, err)
		}

		ids := make([]uuid.UUID, 0, len(p.tags))
		for _, slug := range p.tags {
			tagID, ok := tagIDs[slug]
			if !ok {
				return fmt.Errorf("seed: project %q references unknown tag %q", p.slug, slug)
			}
			ids = append(ids, tagID)
		}
		if err := q.DeleteProjectTags(ctx, id); err != nil {
			return fmt.Errorf("seed: clear tags of %q: %w", p.slug, err)
		}
		if err := q.AddProjectTags(ctx, sqlc.AddProjectTagsParams{ProjectID: id, TagIDs: ids}); err != nil {
			return fmt.Errorf("seed: tag %q: %w", p.slug, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("seed: commit: %w", err)
	}
	return nil
}
