package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToLimitThenBlocks(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/wallet/transfer", nil)
	req.RemoteAddr = "1.2.3.4:5555"

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding limit, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}
}

func TestRateLimiterKeysByUserIDWhenAuthenticated(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	baseReq := httptest.NewRequest(http.MethodPost, "/wallet/transfer", nil)
	baseReq.RemoteAddr = "1.2.3.4:5555"

	userAReq := baseReq.WithContext(context.WithValue(baseReq.Context(), "user_id", "user-a"))
	userBReq := baseReq.WithContext(context.WithValue(baseReq.Context(), "user_id", "user-b"))

	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, userAReq)
	if recA.Code != http.StatusOK {
		t.Fatalf("user A first request: expected 200, got %d", recA.Code)
	}

	recAAgain := httptest.NewRecorder()
	handler.ServeHTTP(recAAgain, userAReq)
	if recAAgain.Code != http.StatusTooManyRequests {
		t.Fatalf("user A second request: expected 429, got %d", recAAgain.Code)
	}

	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, userBReq)
	if recB.Code != http.StatusOK {
		t.Fatalf("user B (different key, same IP) should not be blocked by user A's limit, got %d", recB.Code)
	}
}
