package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/example/librarylendingapi/internal/auth"
)

type RateLimiter struct {
	mu sync.Mutex
	limits map[string]*tokenBucket
	limit int
	interval time.Duration
}

type tokenBucket struct {
	tokens int
	lastRefill time.Time
}

func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{limits: make(map[string]*tokenBucket), limit: limit, interval: interval}
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rl.keyForRequest(r)
			if !rl.allow(key) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) keyForRequest(r *http.Request) string {
	if c := r.Context().Value(userClaimsKey); c != nil {
		if claims, ok := c.(*auth.Claims); ok { return "user:" + fmt.Sprint(claims.UserID) }
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return "ip:" + ip
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.limits[key]
	now := time.Now()
	if !ok {
		b = &tokenBucket{tokens: rl.limit - 1, lastRefill: now}
		rl.limits[key] = b
		return true
	}
	if now.Sub(b.lastRefill) >= rl.interval {
		b.tokens = rl.limit - 1
		b.lastRefill = now
		return true
	}
	if b.tokens <= 0 { return false }
	b.tokens--
	return true
}
