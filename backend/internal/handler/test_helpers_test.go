package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/fingoat/api/internal/auth"
	"github.com/fingoat/api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// hashForTest creates a bcrypt hash for testing. Fails the test on error.
func hashForTest(t *testing.T, password string) string {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword(%q): %v", password, err)
	}
	return hash
}

// requestWithUserID creates an HTTP request with the user ID set in the context,
// simulating what the auth middleware does.
func requestWithUserID(r *http.Request, userID int64) *http.Request {
	ctx := middleware.SetUserIDForTest(r.Context(), userID)
	return r.WithContext(ctx)
}

// requestWithChiURLParam creates an HTTP request with a chi URL param set.
func requestWithChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}
