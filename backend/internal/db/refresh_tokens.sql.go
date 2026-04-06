package db

import (
	"context"
	"time"
)

type CreateRefreshTokenParams struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, created_at
	`
	row := q.db.QueryRow(ctx, query, arg.UserID, arg.TokenHash, arg.ExpiresAt)
	var rt RefreshToken
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return RefreshToken{}, err
	}
	return rt, nil
}

func (q *Queries) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	row := q.db.QueryRow(ctx, query, tokenHash)
	var rt RefreshToken
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt); err != nil {
		return RefreshToken{}, err
	}
	return rt, nil
}

func (q *Queries) DeleteRefreshToken(ctx context.Context, id int64) error {
	const query = `DELETE FROM refresh_tokens WHERE id = $1`
	_, err := q.db.Exec(ctx, query, id)
	return err
}

func (q *Queries) DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error {
	const query = `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := q.db.Exec(ctx, query, userID)
	return err
}
