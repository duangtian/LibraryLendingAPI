package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/librarylendingapi/internal/auth"
)

type Authenticator interface { Parse(token string) (*auth.Claims, error) }

type contextKey string

const userClaimsKey contextKey = "userClaims"

func Auth(j Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				next.ServeHTTP(w, r) // allow anonymous; handlers can enforce auth
				return
			}
			token := strings.TrimPrefix(h, "Bearer ")
			claims, err := j.Parse(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(r *http.Request) *auth.Claims {
	if v := r.Context().Value(userClaimsKey); v != nil {
		if c, ok := v.(*auth.Claims); ok { return c }
	}
	return nil
}
