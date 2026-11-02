package httpapi

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/bekbek22/portfolio/server/internal/httpapi/gen"
)

func TestRequestValidation(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, DefaultConfig(), Deps{})

	tests := []struct {
		name         string
		target       string
		wantStatus   int
		wantFields   []string // keys expected in fieldErrors (422 only)
		wantContains string   // substring expected in one of the field messages
	}{
		// Valid requests reach the handler (fake service: empty list, unknown slug).
		{name: "no params", target: "/api/v1/projects", wantStatus: http.StatusOK},
		{name: "all params valid", target: "/api/v1/projects?tag=go&tag=postgresql&q=api&featured=true", wantStatus: http.StatusOK},
		{name: "valid slug", target: "/api/v1/projects/go-microservice", wantStatus: http.StatusNotFound},

		{name: "tag not a slug", target: "/api/v1/projects?tag=Not_Valid", wantStatus: http.StatusUnprocessableEntity, wantFields: []string{"tag"}},
		{name: "too many tags", target: "/api/v1/projects?" + strings.Repeat("tag=a&", 11), wantStatus: http.StatusUnprocessableEntity, wantFields: []string{"tag"}},
		{name: "q too long", target: "/api/v1/projects?q=" + strings.Repeat("x", 51), wantStatus: http.StatusUnprocessableEntity, wantFields: []string{"q"}, wantContains: "50"},
		{name: "featured not a boolean", target: "/api/v1/projects?featured=maybe", wantStatus: http.StatusUnprocessableEntity, wantFields: []string{"featured"}},
		{name: "every invalid param reported", target: "/api/v1/projects?tag=BAD&featured=maybe&q=" + strings.Repeat("x", 51), wantStatus: http.StatusUnprocessableEntity, wantFields: []string{"featured", "q", "tag"}},
		{name: "path slug invalid", target: "/api/v1/projects/Bad_Slug", wantStatus: http.StatusUnprocessableEntity, wantFields: []string{"slug"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := do(t, app, http.MethodGet, tt.target, nil)
			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusUnprocessableEntity {
				return
			}

			body := decode[gen.Error](t, res)
			if body.Error.Code != gen.ErrorCodeValidation || body.Error.FieldErrors == nil {
				t.Fatalf("unexpected error body: %+v", body.Error)
			}
			fields := *body.Error.FieldErrors
			keys := make([]string, 0, len(fields))
			for k := range fields {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			if !slices.Equal(keys, tt.wantFields) {
				t.Errorf("fieldErrors keys = %v, want %v (%v)", keys, tt.wantFields, fields)
			}
			if tt.wantContains != "" && !strings.Contains(strings.Join(fields[tt.wantFields[0]], " "), tt.wantContains) {
				t.Errorf("messages %v do not mention %q", fields[tt.wantFields[0]], tt.wantContains)
			}
		})
	}
}

func TestErrorEnvelope(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, DefaultConfig(), Deps{})
	app.Get("/api/v1/test-panic", func(fiber.Ctx) error { panic("secret internals") })

	tests := []struct {
		name        string
		method      string
		target      string
		wantStatus  int
		wantCode    gen.ErrorCode
		wantMessage string
	}{
		{"unknown route under prefix", http.MethodGet, "/api/v1/nope", http.StatusNotFound, gen.ErrorCodeNotFound, "Route not found"},
		{"unknown route outside prefix", http.MethodGet, "/", http.StatusNotFound, gen.ErrorCodeNotFound, "Route not found"},
		{"wrong method", http.MethodPost, "/api/v1/projects", http.StatusMethodNotAllowed, gen.ErrorCodeMethodNotAllowed, "Method not allowed"},
		{"panic does not leak", http.MethodGet, "/api/v1/test-panic", http.StatusInternalServerError, gen.ErrorCodeInternal, "Internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := do(t, app, tt.method, tt.target, nil)

			if res.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want JSON", ct)
			}
			body := decode[gen.Error](t, res)
			if body.Error.Code != tt.wantCode || body.Error.Message != tt.wantMessage {
				t.Errorf("error = %+v, want code %s message %q", body.Error, tt.wantCode, tt.wantMessage)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RateLimitMax = 2
	app := newTestApp(t, cfg, Deps{})

	for i := range 2 {
		if res := do(t, app, http.MethodGet, "/api/v1/projects", nil); res.StatusCode == http.StatusTooManyRequests {
			t.Fatalf("request %d limited too early", i+1)
		}
	}

	res := do(t, app, http.MethodGet, "/api/v1/projects", nil)
	if res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", res.StatusCode)
	}
	if res.Header.Get("Retry-After") == "" {
		t.Error("missing Retry-After header")
	}
	if body := decode[gen.Error](t, res); body.Error.Code != gen.ErrorCodeRateLimited {
		t.Errorf("code = %s, want RATE_LIMITED", body.Error.Code)
	}

	// Health checks are exempt so probes never fail because of traffic.
	if res := do(t, app, http.MethodGet, "/api/v1/health", nil); res.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want 200 while rate limited", res.StatusCode)
	}

	// The limit is per client IP.
	other := http.Header{"X-Forwarded-For": {"203.0.113.9"}}
	cfgProxy := cfg
	cfgProxy.TrustedProxyHops = 1
	proxied := newTestApp(t, cfgProxy, Deps{})
	for range 2 {
		do(t, proxied, http.MethodGet, "/api/v1/projects", other)
	}
	if res := do(t, proxied, http.MethodGet, "/api/v1/projects", http.Header{"X-Forwarded-For": {"198.51.100.7"}}); res.StatusCode == http.StatusTooManyRequests {
		t.Error("a different client IP was rate limited")
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, DefaultConfig(), Deps{})
	app.Get("/api/v1/test-context", func(c fiber.Ctx) error {
		deadline, ok := c.Context().Deadline()
		return c.JSON(fiber.Map{"hasDeadline": ok, "inSeconds": time.Until(deadline).Seconds(), "clientIP": ClientIP(c)})
	})

	t.Run("handler context has the request deadline", func(t *testing.T) {
		t.Parallel()
		got := decode[struct {
			HasDeadline bool    `json:"hasDeadline"`
			InSeconds   float64 `json:"inSeconds"`
		}](t, do(t, app, http.MethodGet, "/api/v1/test-context", nil))
		if !got.HasDeadline || got.InSeconds <= 0 || got.InSeconds > DefaultConfig().RequestTimeout.Seconds() {
			t.Errorf("deadline = %+v, want within %s", got, DefaultConfig().RequestTimeout)
		}
	})

	t.Run("request ID is generated", func(t *testing.T) {
		t.Parallel()
		if id := do(t, app, http.MethodGet, "/api/v1/health", nil).Header.Get("X-Request-ID"); id == "" {
			t.Error("missing X-Request-ID")
		}
	})

	t.Run("incoming request ID is echoed", func(t *testing.T) {
		t.Parallel()
		res := do(t, app, http.MethodGet, "/api/v1/health", http.Header{"X-Request-ID": {"trace-abc-123"}})
		if id := res.Header.Get("X-Request-ID"); id != "trace-abc-123" {
			t.Errorf("X-Request-ID = %q, want echoed value", id)
		}
	})

	t.Run("spec is served", func(t *testing.T) {
		t.Parallel()
		res := do(t, app, http.MethodGet, "/api/v1/openapi.json", nil)
		if got := decode[map[string]any](t, res); got["openapi"] != "3.0.3" {
			t.Errorf("openapi = %v, want 3.0.3", got["openapi"])
		}
	})
}

func TestResolveClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		forwardedFor string
		remoteIP     string
		hops         int
		want         string
	}{
		{"no trusted proxies ignores header", "6.6.6.6", "1.1.1.1", 0, "1.1.1.1"},
		{"one proxy: last entry", "6.6.6.6, 2.2.2.2", "10.0.0.2", 1, "2.2.2.2"},
		{"two proxies: second from right", "6.6.6.6, 2.2.2.2, 10.0.0.1", "10.0.0.2", 2, "2.2.2.2"},
		{"spoofed left entries never win", "1.2.3.4, 5.6.7.8, 2.2.2.2", "10.0.0.2", 1, "2.2.2.2"},
		{"missing header falls back", "", "1.1.1.1", 1, "1.1.1.1"},
		{"too few entries falls back", "2.2.2.2", "1.1.1.1", 3, "1.1.1.1"},
		{"empty entry falls back", "2.2.2.2, ", "1.1.1.1", 1, "1.1.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveClientIP(tt.forwardedFor, tt.remoteIP, tt.hops); got != tt.want {
				t.Errorf("resolveClientIP(%q, %q, %d) = %q, want %q", tt.forwardedFor, tt.remoteIP, tt.hops, got, tt.want)
			}
		})
	}
}
