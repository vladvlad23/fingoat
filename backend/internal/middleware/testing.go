package middleware

import "context"

// SetUserIDForTest sets the user ID in the context. This is exported only
// for use in handler tests to simulate authenticated requests.
func SetUserIDForTest(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
