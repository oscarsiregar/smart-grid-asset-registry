package querysafety

import "testing"

func TestSafeOrderClause(t *testing.T) {
	tests := []struct {
		name    string
		sort    string
		order   string
		allowed []string
		want    string
	}{
		{"valid identifier and direction, order normalized", "created_at", "desc", nil, "created_at DESC"},
		{"already-upper direction", "updated_at", "ASC", nil, "updated_at ASC"},
		{"sql injection in sort falls back", "id; DROP TABLE users;--", "ASC", nil, "id ASC"},
		{"whitespace in sort falls back", "id desc, (select 1)", "ASC", nil, "id ASC"},
		{"empty sort falls back", "", "ASC", nil, "id ASC"},
		{"invalid order falls back", "created_at", "ASC; DROP TABLE users;--", nil, "created_at DESC"},
		{"empty order falls back", "created_at", "", nil, "created_at DESC"},
		{"sort in allow-list passes", "name", "asc", []string{"name", "code"}, "name ASC"},
		{"valid identifier outside allow-list falls back", "password_hash", "asc", []string{"name", "code"}, "id ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeOrderClause(tt.sort, tt.order, "id", "DESC", tt.allowed...)
			if got != tt.want {
				t.Errorf("SafeOrderClause(%q, %q) = %q, want %q", tt.sort, tt.order, got, tt.want)
			}
		})
	}
}

func TestContainsPattern(t *testing.T) {
	tests := map[string]string{
		"widget":  "%widget%",
		"":        "%%",
		"50%":     `%50\%%`,
		"a_b":     `%a\_b%`,
		`back\sl`: `%back\\sl%`,
	}
	for in, want := range tests {
		if got := ContainsPattern(in); got != want {
			t.Errorf("ContainsPattern(%q) = %q, want %q", in, got, want)
		}
	}
}
