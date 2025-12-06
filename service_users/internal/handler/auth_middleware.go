package handler

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func (c *UserController) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error": "invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		authInfo, err := c.service.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, authInfo.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// helpers
func getUserIDFromContext(ctx context.Context) (int, bool) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return 0, false
	}
	id, ok := val.(int)
	return id, ok
}