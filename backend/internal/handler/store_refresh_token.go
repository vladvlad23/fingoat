package handler

import (
	"context"

	"github.com/fingoat/api/internal/db"
)

// RefreshTokenStore defines refresh token database operations.
type RefreshTokenStore interface {
	CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id int64) error
	DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error
}
