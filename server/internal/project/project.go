// Package project is the read side of the portfolio's projects: filtering, search and
// tag counts. It owns its domain types; the HTTP layer maps them to API types.
package project

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/bekbek22/portfolio/server/internal/database/sqlc"
)

var ErrNotFound = errors.New("project not found")

type Tag struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type TagCategory string

const (
	TagCategoryLanguage  TagCategory = "language"
	TagCategoryFramework TagCategory = "framework"
	TagCategoryTool      TagCategory = "tool"
	TagCategoryDomain    TagCategory = "domain"
)

type TagCount struct {
	Tag
	Category     TagCategory
	ProjectCount int
}

// Card is the listing view of a published project.
type Card struct {
	Slug          string
	Title         string
	Summary       string
	CoverImageURL *string
	Featured      bool
	Tags          []Tag
}

// Detail is the full view of a published project.
type Detail struct {
	Card
	Body      string // markdown
	RepoURL   *string
	LiveURL   *string
	UpdatedAt time.Time
}

// Filter narrows List. The zero value lists every published project.
type Filter struct {
	Tags     []string // a project must have ALL of these tag slugs
	Search   string   // case-insensitive substring of title or summary
	Featured *bool
}

// Store is the subset of the generated queries the service needs.
type Store interface {
	ListPublishedProjects(ctx context.Context, arg sqlc.ListPublishedProjectsParams) ([]sqlc.ListPublishedProjectsRow, error)
	GetPublishedProjectBySlug(ctx context.Context, slug string) (sqlc.GetPublishedProjectBySlugRow, error)
	ListTagsWithCounts(ctx context.Context) ([]sqlc.ListTagsWithCountsRow, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, f Filter) ([]Card, error) {
	params := sqlc.ListPublishedProjectsParams{
		Featured: f.Featured,
		// Duplicates would make the HAVING count unreachable (?tag=go&tag=go).
		TagSlugs: slices.Compact(slices.Sorted(slices.Values(f.Tags))),
	}
	if f.Search != "" {
		search := escapeLike(f.Search)
		params.Search = &search
	}

	rows, err := s.store.ListPublishedProjects(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	cards := make([]Card, 0, len(rows))
	for _, r := range rows {
		tags, err := decodeTags(r.Tags)
		if err != nil {
			return nil, fmt.Errorf("list projects: project %q: %w", r.Slug, err)
		}
		cards = append(cards, Card{
			Slug: r.Slug, Title: r.Title, Summary: r.Summary,
			CoverImageURL: r.CoverImageURL, Featured: r.Featured, Tags: tags,
		})
	}
	return cards, nil
}

func (s *Service) Get(ctx context.Context, slug string) (Detail, error) {
	r, err := s.store.GetPublishedProjectBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Detail{}, ErrNotFound
	}
	if err != nil {
		return Detail{}, fmt.Errorf("get project %q: %w", slug, err)
	}

	tags, err := decodeTags(r.Tags)
	if err != nil {
		return Detail{}, fmt.Errorf("get project %q: %w", slug, err)
	}
	return Detail{
		Card: Card{
			Slug: r.Slug, Title: r.Title, Summary: r.Summary,
			CoverImageURL: r.CoverImageURL, Featured: r.Featured, Tags: tags,
		},
		Body: r.Body, RepoURL: r.RepoURL, LiveURL: r.LiveURL, UpdatedAt: r.UpdatedAt,
	}, nil
}

func (s *Service) Tags(ctx context.Context) ([]TagCount, error) {
	rows, err := s.store.ListTagsWithCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	out := make([]TagCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, TagCount{
			Tag:          Tag{Slug: r.Slug, Name: r.Name},
			Category:     TagCategory(r.Category),
			ProjectCount: int(r.ProjectCount),
		})
	}
	return out, nil
}

// decodeTags parses the json_agg column produced by the queries.
func decodeTags(raw []byte) ([]Tag, error) {
	var tags []Tag
	if err := json.Unmarshal(raw, &tags); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}
	return tags, nil
}

// likeEscaper makes user input match literally inside an ILIKE ... ESCAPE '\' pattern.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func escapeLike(s string) string {
	return likeEscaper.Replace(s)
}
