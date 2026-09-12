package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func withJWTSecret(t *testing.T, secret string) {
	t.Helper()
	prev, had := os.LookupEnv("JWT_SECRET")
	os.Setenv("JWT_SECRET", secret)
	t.Cleanup(func() {
		if had {
			os.Setenv("JWT_SECRET", prev)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	})
}

func signToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	withJWTSecret(t, "test-secret")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if called {
		t.Error("next handler should not be called")
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	withJWTSecret(t, "test-secret")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Basic sometoken")
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	withJWTSecret(t, "test-secret")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	withJWTSecret(t, "test-secret")

	token := signToken(t, "a-different-secret", jwt.MapClaims{
		"user_id": "user-1",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	withJWTSecret(t, "test-secret")

	token := signToken(t, "test-secret", jwt.MapClaims{
		"user_id": "user-1",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_MissingUserIDClaim(t *testing.T) {
	withJWTSecret(t, "test-secret")

	token := signToken(t, "test-secret", jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	withJWTSecret(t, "test-secret")

	token := signToken(t, "test-secret", jwt.MapClaims{
		"user_id": "user-42",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	var gotUserID any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Context().Value("user_id")
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotUserID != "user-42" {
		t.Errorf("context user_id = %v, want %q", gotUserID, "user-42")
	}
}
