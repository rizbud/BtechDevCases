package user

import "testing"

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		email string
		want  bool
	}{
		{"user@example.com", true},
		{"user.name+tag@sub.example.co", true},
		{"user_name-1@example-domain.com", true},
		{"missing-at.example.com", false},
		{"missing-domain@", false},
		{"@missing-local.com", false},
		{"no-tld@example", false},
		{"", false},
		{"spaces in@example.com", false},
	}

	for _, c := range cases {
		if got := ValidateEmail(c.email); got != c.want {
			t.Errorf("ValidateEmail(%q) = %v, want %v", c.email, got, c.want)
		}
	}
}
