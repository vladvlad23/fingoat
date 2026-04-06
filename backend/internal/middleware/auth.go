package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/fingoat/api/internal/auth"
)

type contextKey string

const userIDKey contextKey = "userID"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := auth.ValidateToken(tokenStr, jwtSecret)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (int64, error) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, errors.New("no user in context")
	}
	id, ok := v.(int64)
	if !ok {
		return 0, errors.New("invalid user id in context")
	}
	return id, nil
}
