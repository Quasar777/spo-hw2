package handler

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// AuthMiddleware — проверяет Authorization: Bearer <token>,
// если токен валиден — кладёт userID в контекст.
func (c *UserController) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			c.writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "missing Authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid Authorization header format",
			})
			return
		}

		tokenStr := parts[1]

		authInfo, err := c.service.ParseToken(tokenStr)
		if err != nil {
			c.writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
			return
		}

		// Кладём userID в контекст
		ctx := context.WithValue(r.Context(), userIDContextKey, authInfo.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// helper для хендлеров
func getUserIDFromContext(ctx context.Context) (int, bool) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return 0, false
	}
	id, ok := val.(int)
	return id, ok
}