package database

import "testing"

func TestToPgx5URL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"postgres://u:p@h:5432/db?sslmode=disable", "pgx5://u:p@h:5432/db?sslmode=disable"},
		{"postgresql://u:p@h/db", "pgx5://u:p@h/db"},
		{"pgx5://u:p@h/db", "pgx5://u:p@h/db"},
	}
	for _, tt := range tests {
		if got := toPgx5URL(tt.in); got != tt.want {
			t.Errorf("toPgx5URL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
