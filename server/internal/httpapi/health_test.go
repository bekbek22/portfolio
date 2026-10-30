package httpapi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bekbek101/portfolio/server/internal/httpapi/gen"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ping       fakePinger
		wantStatus int
		wantBody   gen.Health
	}{
		{
			name:       "database up",
			ping:       func(context.Context) error { return nil },
			wantStatus: http.StatusOK,
			wantBody:   gen.Health{Status: gen.HealthStatusOk, DB: gen.HealthDBUp},
		},
		{
			name:       "database down",
			ping:       func(context.Context) error { return errors.New("connection refused") },
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   gen.Health{Status: gen.HealthStatusDegraded, DB: gen.HealthDBDown},
		},
		{
			name: "ping receives a deadline",
			ping: func(ctx context.Context) error {
				if _, ok := ctx.Deadline(); !ok {
					return errors.New("no deadline on context")
				}
				return nil
			},
			wantStatus: http.StatusOK,
			wantBody:   gen.Health{Status: gen.HealthStatusOk, DB: gen.HealthDBUp},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp(t, DefaultConfig(), Deps{DB: tt.ping})

			res := do(t, app, http.MethodGet, "/api/v1/health", nil)

			if res.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if got := decode[gen.Health](t, res); got != tt.wantBody {
				t.Errorf("body = %+v, want %+v", got, tt.wantBody)
			}
		})
	}
}
