package project

import "testing"

func TestEscapeLike(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"100%", `100\%`},
		{"snake_case", `snake\_case`},
		{`back\slash`, `back\\slash`},
		{`%_\`, `\%\_\\`},
		{"", ""},
	}
	for _, tt := range tests {
		if got := escapeLike(tt.in); got != tt.want {
			t.Errorf("escapeLike(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
