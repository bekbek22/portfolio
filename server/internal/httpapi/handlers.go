package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/bekbek22/portfolio/server/internal/httpapi/gen"
	"github.com/bekbek22/portfolio/server/internal/project"
)

// Pinger is the slice of the database pool the health check needs.
type Pinger interface {
	Ping(ctx context.Context) error
}

// ProjectService is what the project endpoints need from the domain layer.
type ProjectService interface {
	List(ctx context.Context, f project.Filter) ([]project.Card, error)
	Get(ctx context.Context, slug string) (project.Detail, error)
	Tags(ctx context.Context) ([]project.TagCount, error)
}

// handlers implements the operations generated from api/openapi.yaml.
type handlers struct {
	db       Pinger
	projects ProjectService
}

var _ gen.StrictServerInterface = (*handlers)(nil)

const healthDBTimeout = 2 * time.Second

func (h *handlers) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, healthDBTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return gen.GetHealth503JSONResponse{Status: gen.HealthStatusDegraded, DB: gen.HealthDBDown}, nil
	}
	return gen.GetHealth200JSONResponse{Status: gen.HealthStatusOk, DB: gen.HealthDBUp}, nil
}

func (h *handlers) ListProjects(ctx context.Context, req gen.ListProjectsRequestObject) (gen.ListProjectsResponseObject, error) {
	f := project.Filter{Featured: req.Params.Featured}
	if req.Params.Tag != nil {
		f.Tags = *req.Params.Tag
	}
	if req.Params.Q != nil {
		f.Search = *req.Params.Q
	}

	cards, err := h.projects.List(ctx, f)
	if err != nil {
		return nil, err
	}
	data := make([]gen.ProjectCard, 0, len(cards))
	for _, c := range cards {
		data = append(data, toProjectCard(c))
	}
	return gen.ListProjects200JSONResponse{Data: data}, nil
}

func (h *handlers) GetProject(ctx context.Context, req gen.GetProjectRequestObject) (gen.GetProjectResponseObject, error) {
	d, err := h.projects.Get(ctx, req.Slug)
	if errors.Is(err, project.ErrNotFound) {
		return gen.GetProject404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{
			Error: gen.ErrorDetail{Code: gen.ErrorCodeNotFound, Message: "Project not found"},
		}}, nil
	}
	if err != nil {
		return nil, err
	}

	card := toProjectCard(d.Card)
	return gen.GetProject200JSONResponse{Data: gen.ProjectDetail{
		Slug: card.Slug, Title: card.Title, Summary: card.Summary,
		CoverImageURL: card.CoverImageURL, Featured: card.Featured, Tags: card.Tags,
		Body: d.Body, RepoURL: d.RepoURL, LiveURL: d.LiveURL, UpdatedAt: d.UpdatedAt.UTC(),
	}}, nil
}

func (h *handlers) ListTags(ctx context.Context, _ gen.ListTagsRequestObject) (gen.ListTagsResponseObject, error) {
	tags, err := h.projects.Tags(ctx)
	if err != nil {
		return nil, err
	}
	data := make([]gen.TagWithCount, 0, len(tags))
	for _, t := range tags {
		data = append(data, gen.TagWithCount{
			Slug: t.Slug, Name: t.Name,
			Category: gen.TagCategory(t.Category), ProjectCount: t.ProjectCount,
		})
	}
	return gen.ListTags200JSONResponse{Data: data}, nil
}

func toProjectCard(c project.Card) gen.ProjectCard {
	tags := make([]gen.Tag, 0, len(c.Tags))
	for _, t := range c.Tags {
		tags = append(tags, gen.Tag{Slug: t.Slug, Name: t.Name})
	}
	return gen.ProjectCard{
		Slug: c.Slug, Title: c.Title, Summary: c.Summary,
		CoverImageURL: c.CoverImageURL, Featured: c.Featured, Tags: tags,
	}
}
