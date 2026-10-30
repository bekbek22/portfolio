package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger func(ctx context.Context) error

func (f fakePinger) Ping(ctx context.Context) error { return f(ctx) }

func TestHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ping       fakePinger
		wantStatus int
		wantBody   healthResponse
	}{
		{
			name:       "database up",
			ping:       func(context.Context) error { return nil },
			wantStatus: http.StatusOK,
			wantBody:   healthResponse{Status: "ok", DB: "up"},
		},
		{
			name:       "database down",
			ping:       func(context.Context) error { return errors.New("connection refused") },
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   healthResponse{Status: "degraded", DB: "down"},
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
			wantBody:   healthResponse{Status: "ok", DB: "up"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := New(Deps{DB: tt.ping})
			res, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			var got healthResponse
			if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got != tt.wantBody {
				t.Errorf("body = %+v, want %+v", got, tt.wantBody)
			}
		})
	}
}
