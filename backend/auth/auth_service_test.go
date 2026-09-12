package auth

import "testing"

func TestValidatePassword(t *testing.T) {
	s := &AuthService{}
	cases := []struct {
		password string
		want     bool
	}{
		{"short", false},
		{"12345678", true},
		{"", false},
		{"exactly8", true},
	}
	for _, c := range cases {
		if got := s.validatePassword(c.password); got != c.want {
			t.Errorf("validatePassword(%q) = %v, want %v", c.password, got, c.want)
		}
	}
}

func TestValidateConfirmPassword(t *testing.T) {
	s := &AuthService{}
	if !s.validateConfirmPassword("abc123", "abc123") {
		t.Error("expected matching passwords to be valid")
	}
	if s.validateConfirmPassword("abc123", "different") {
		t.Error("expected mismatched passwords to be invalid")
	}
}

func TestValidateLoginRequest(t *testing.T) {
	s := &AuthService{}

	errs := s.validateLoginRequest("invalid-email", "short")
	if _, ok := errs["email"]; !ok {
		t.Error("expected email validation error")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("expected password validation error")
	}

	errs = s.validateLoginRequest("user@example.com", "longenoughpassword")
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidateRegisterRequest(t *testing.T) {
	s := &AuthService{}

	errs := s.validateRegisterRequest("user@example.com", "password1", "password2")
	if _, ok := errs["confirmPassword"]; !ok {
		t.Error("expected confirmPassword validation error for mismatched passwords")
	}

	errs = s.validateRegisterRequest("user@example.com", "password1", "password1")
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}
