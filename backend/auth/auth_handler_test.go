package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"btech-wallet/server"
)

// These tests cover the handler's request parsing and validation paths,
// which return before any repository/DB call is made.

func newTestAuthHandler() *AuthHandler {
	return &AuthHandler{service: &AuthService{}}
}

func decodeValidationErrors(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp server.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	errs, ok := resp.Error.(map[string]any)
	if !ok {
		t.Fatalf("expected error field to be a map, got %v (%T)", resp.Error, resp.Error)
	}
	return errs
}

func TestHandleLogin_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()
	r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	h.handleLogin(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleLogin_ValidationErrors(t *testing.T) {
	h := newTestAuthHandler()
	body := `{"email":"not-an-email","password":"short"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.handleLogin(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	errs := decodeValidationErrors(t, w)
	if _, ok := errs["email"]; !ok {
		t.Error("expected email validation error")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("expected password validation error")
	}
}

func TestHandleRegister_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()
	r := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(""))
	w := httptest.NewRecorder()

	h.handleRegister(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleRegister_ValidationErrors(t *testing.T) {
	h := newTestAuthHandler()
	body := `{"email":"user@example.com","password":"password1","confirmPassword":"password2"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	w := httptest.NewRecorder()

	h.handleRegister(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	errs := decodeValidationErrors(t, w)
	if _, ok := errs["confirmPassword"]; !ok {
		t.Error("expected confirmPassword validation error")
	}
}

func TestHandleRefreshToken_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	h.handleRefreshToken(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleRefreshToken_MissingToken(t *testing.T) {
	h := newTestAuthHandler()
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh-token", strings.NewReader(`{"refresh_token":""}`))
	w := httptest.NewRecorder()

	h.handleRefreshToken(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	errs := decodeValidationErrors(t, w)
	if _, ok := errs["refresh_token"]; !ok {
		t.Error("expected refresh_token validation error")
	}
}
