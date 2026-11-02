package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/bekbek22/portfolio/server/internal/project"
)

type fakePinger func(ctx context.Context) error

func (f fakePinger) Ping(ctx context.Context) error { return f(ctx) }

func healthyDB(context.Context) error { return nil }

// fakeProjects implements ProjectService; unset funcs return empty results.
type fakeProjects struct {
	list func(context.Context, project.Filter) ([]project.Card, error)
	get  func(context.Context, string) (project.Detail, error)
	tags func(context.Context) ([]project.TagCount, error)
}

func (f fakeProjects) List(ctx context.Context, flt project.Filter) ([]project.Card, error) {
	if f.list == nil {
		return nil, nil
	}
	return f.list(ctx, flt)
}

func (f fakeProjects) Get(ctx context.Context, slug string) (project.Detail, error) {
	if f.get == nil {
		return project.Detail{}, project.ErrNotFound
	}
	return f.get(ctx, slug)
}

func (f fakeProjects) Tags(ctx context.Context) ([]project.TagCount, error) {
	if f.tags == nil {
		return nil, nil
	}
	return f.tags(ctx)
}

// newTestApp builds the app with a healthy fake DB and empty fake projects unless
// deps overrides them.
func newTestApp(t *testing.T, cfg Config, deps Deps) *fiber.App {
	t.Helper()
	if deps.DB == nil {
		deps.DB = fakePinger(healthyDB)
	}
	if deps.Projects == nil {
		deps.Projects = fakeProjects{}
	}
	deps.Logger = slog.New(slog.DiscardHandler)
	app, err := New(deps, cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return app
}

func do(t *testing.T, app *fiber.App, method, target string, header http.Header) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

func decode[T any](t *testing.T, res *http.Response) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		t.Fatalf("decode %T: %v", v, err)
	}
	return v
}
