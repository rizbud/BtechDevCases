package auth

import (
	"testing"
	"time"
)

func TestJWTIssueAndVerify(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)

	token, err := m.Issue("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("claims.UserID = %q, want %q", claims.UserID, "user-1")
	}
	if claims.Email != "user@example.com" {
		t.Errorf("claims.Email = %q, want %q", claims.Email, "user@example.com")
	}
}

func TestJWTVerifyExpiredToken(t *testing.T) {
	m := NewJWTManager("test-secret", -time.Hour) // already expired

	token, err := m.Issue("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := m.Verify(token); err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestJWTVerifyWrongSecret(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, err := m.Issue("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	other := NewJWTManager("different-secret", time.Hour)
	if _, err := other.Verify(token); err == nil {
		t.Error("expected error for token signed with a different secret, got nil")
	}
}

func TestJWTVerifyMalformedToken(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	if _, err := m.Verify("not-a-valid-token"); err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}
