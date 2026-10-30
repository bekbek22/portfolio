package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	const dbURL = "postgres://u:p@localhost:5432/db"

	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr []string // substrings expected in the error
	}{
		{
			name: "defaults",
			env:  map[string]string{"DATABASE_URL": dbURL},
			want: Config{Env: EnvDevelopment, Port: 4000, DatabaseURL: dbURL},
		},
		{
			name: "explicit values",
			env:  map[string]string{"APP_ENV": "production", "PORT": "8080", "DATABASE_URL": dbURL, "TRUSTED_PROXY_HOPS": "2"},
			want: Config{Env: EnvProduction, Port: 8080, DatabaseURL: dbURL, TrustedProxyHops: 2},
		},
		{
			name:    "missing database url",
			env:     map[string]string{},
			wantErr: []string{"DATABASE_URL: required"},
		},
		{
			name:    "reports every invalid value",
			env:     map[string]string{"APP_ENV": "staging", "PORT": "99999", "TRUSTED_PROXY_HOPS": "-1"},
			wantErr: []string{"APP_ENV", "PORT", "DATABASE_URL", "TRUSTED_PROXY_HOPS"},
		},
		{
			name:    "non-numeric port",
			env:     map[string]string{"PORT": "abc", "DATABASE_URL": dbURL},
			wantErr: []string{`PORT: must be 1-65535, got "abc"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Load(func(k string) string { return tt.env[k] })

			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("expected error, got config %+v", got)
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q does not contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
