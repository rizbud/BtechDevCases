package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"btech-wallet/server"
)

type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	hits   map[string]*rateLimitEntry
}

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		hits:   make(map[string]*rateLimitEntry),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientKey(r)
		now := time.Now()

		rl.mu.Lock()
		entry, ok := rl.hits[key]
		if !ok || now.Sub(entry.windowStart) >= rl.window {
			entry = &rateLimitEntry{windowStart: now}
			rl.hits[key] = entry
		}
		entry.count++
		count := entry.count
		reset := entry.windowStart.Add(rl.window)
		rl.mu.Unlock()

		remaining := rl.limit - count
		if remaining < 0 {
			remaining = 0
		}
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		if count > rl.limit {
			retryAfter := int(time.Until(reset).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			server.ErrorResponseJSON(w, http.StatusTooManyRequests, "Too many requests, please slow down", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientKey(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok && userID != "" {
		return "user:" + userID
	}
	return "ip:" + clientIP(r)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
