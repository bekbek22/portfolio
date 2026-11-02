package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bekbek22/portfolio/server/internal/httpapi/gen"
	"github.com/bekbek22/portfolio/server/internal/project"
)

func ptr[T any](v T) *T { return &v }

func TestListProjectsPassesFilterToService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		query string
		want  project.Filter
	}{
		{"", project.Filter{}},
		{"?tag=go&tag=postgresql", project.Filter{Tags: []string{"go", "postgresql"}}},
		{"?q=api", project.Filter{Search: "api"}},
		{"?featured=false", project.Filter{Featured: ptr(false)}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			t.Parallel()
			var got project.Filter
			app := newTestApp(t, DefaultConfig(), Deps{Projects: fakeProjects{
				list: func(_ context.Context, f project.Filter) ([]project.Card, error) {
					got = f
					return nil, nil
				},
			}})

			if res := do(t, app, http.MethodGet, "/api/v1/projects"+tt.query, nil); res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", res.StatusCode)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filter = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestListProjectsResponse(t *testing.T) {
	t.Parallel()

	t.Run("maps domain cards", func(t *testing.T) {
		t.Parallel()
		app := newTestApp(t, DefaultConfig(), Deps{Projects: fakeProjects{
			list: func(context.Context, project.Filter) ([]project.Card, error) {
				return []project.Card{{
					Slug: "a", Title: "A", Summary: "S", Featured: true,
					Tags: []project.Tag{{Slug: "go", Name: "Go"}},
				}}, nil
			},
		}})

		body := decode[gen.ListProjects200JSONResponse](t, do(t, app, http.MethodGet, "/api/v1/projects", nil))
		want := []gen.ProjectCard{{
			Slug: "a", Title: "A", Summary: "S", Featured: true, Tags: []gen.Tag{{Slug: "go", Name: "Go"}},
		}}
		if !reflect.DeepEqual(body.Data, want) {
			t.Errorf("data = %+v, want %+v", body.Data, want)
		}
	})

	t.Run("empty result is [] not null", func(t *testing.T) {
		t.Parallel()
		app := newTestApp(t, DefaultConfig(), Deps{})
		raw, _ := io.ReadAll(do(t, app, http.MethodGet, "/api/v1/projects", nil).Body)
		if got := strings.TrimSpace(string(raw)); got != `{"data":[]}` {
			t.Errorf("body = %s", got)
		}
	})

	t.Run("service error is a 500 without detail", func(t *testing.T) {
		t.Parallel()
		app := newTestApp(t, DefaultConfig(), Deps{Projects: fakeProjects{
			list: func(context.Context, project.Filter) ([]project.Card, error) {
				return nil, errors.New("pq: connection reset by peer")
			},
		}})
		res := do(t, app, http.MethodGet, "/api/v1/projects", nil)
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", res.StatusCode)
		}
		if body := decode[gen.Error](t, res); body.Error.Message != "Internal server error" {
			t.Errorf("message = %q leaks detail", body.Error.Message)
		}
	})
}

func TestGetProject(t *testing.T) {
	t.Parallel()
	// Stored in a non-UTC zone: the API must still answer in UTC.
	updated := time.Date(2026, 10, 30, 19, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	app := newTestApp(t, DefaultConfig(), Deps{Projects: fakeProjects{
		get: func(_ context.Context, slug string) (project.Detail, error) {
			if slug != "known" {
				return project.Detail{}, project.ErrNotFound
			}
			return project.Detail{
				Card:      project.Card{Slug: "known", Title: "Known", Summary: "S", Tags: []project.Tag{}},
				Body:      "# Hi",
				RepoURL:   ptr("https://github.com/x/y"),
				UpdatedAt: updated,
			}, nil
		},
	}})

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		res := do(t, app, http.MethodGet, "/api/v1/projects/known", nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", res.StatusCode)
		}
		raw, _ := io.ReadAll(res.Body)
		for _, want := range []string{
			`"body":"# Hi"`, `"repoUrl":"https://github.com/x/y"`,
			`"liveUrl":null`, `"coverImageUrl":null`, `"updatedAt":"2026-10-30T12:00:00Z"`, `"tags":[]`,
		} {
			if !strings.Contains(string(raw), want) {
				t.Errorf("body %s missing %s", raw, want)
			}
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		res := do(t, app, http.MethodGet, "/api/v1/projects/unknown", nil)
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", res.StatusCode)
		}
		if body := decode[gen.Error](t, res); body.Error.Code != gen.ErrorCodeNotFound || body.Error.Message != "Project not found" {
			t.Errorf("error = %+v", body.Error)
		}
	})
}

func TestListTags(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, DefaultConfig(), Deps{Projects: fakeProjects{
		tags: func(context.Context) ([]project.TagCount, error) {
			return []project.TagCount{{Tag: project.Tag{Slug: "go", Name: "Go"}, Category: project.TagCategoryLanguage, ProjectCount: 2}}, nil
		},
	}})

	body := decode[gen.ListTags200JSONResponse](t, do(t, app, http.MethodGet, "/api/v1/tags", nil))
	want := []gen.TagWithCount{{Slug: "go", Name: "Go", Category: gen.TagCategoryLanguage, ProjectCount: 2}}
	if !reflect.DeepEqual(body.Data, want) {
		t.Errorf("data = %+v, want %+v", body.Data, want)
	}
}
