package httpapi

import (
	"context"
	"time"

	"github.com/bekbek101/portfolio/server/internal/httpapi/gen"
)

// Pinger is the slice of the database pool the health check needs.
type Pinger interface {
	Ping(ctx context.Context) error
}

// handlers implements the operations generated from api/openapi.yaml.
type handlers struct {
	db Pinger
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

// Projects and tags are specified in the contract; they are implemented in G3.

func (h *handlers) ListProjects(context.Context, gen.ListProjectsRequestObject) (gen.ListProjectsResponseObject, error) {
	return nil, ErrNotImplemented
}

func (h *handlers) GetProject(context.Context, gen.GetProjectRequestObject) (gen.GetProjectResponseObject, error) {
	return nil, ErrNotImplemented
}

func (h *handlers) ListTags(context.Context, gen.ListTagsRequestObject) (gen.ListTagsResponseObject, error) {
	return nil, ErrNotImplemented
}
