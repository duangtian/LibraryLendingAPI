package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestRateLimiter(t *testing.T) {
    rl := NewRateLimiter(2, time.Minute)
    calls := 0
    h := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(200) }))
    rr1 := httptest.NewRecorder(); h.ServeHTTP(rr1, httptest.NewRequest("GET","/",nil))
    if rr1.Code != 200 { t.Fatalf("expected 200") }
    rr2 := httptest.NewRecorder(); h.ServeHTTP(rr2, httptest.NewRequest("GET","/",nil))
    if rr2.Code != 200 { t.Fatalf("expected 200 second") }
    rr3 := httptest.NewRecorder(); h.ServeHTTP(rr3, httptest.NewRequest("GET","/",nil))
    if rr3.Code != http.StatusTooManyRequests { t.Fatalf("expected 429 got %d", rr3.Code) }
    if calls != 2 { t.Fatalf("expected 2 calls allowed, got %d", calls) }
}
