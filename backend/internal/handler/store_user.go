package handler

import (
	"context"

	"github.com/fingoat/api/internal/db"
)

// UserStore defines user-related database operations.
type UserStore interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	UpdateUserSettings(ctx context.Context, arg db.UpdateUserSettingsParams) (db.User, error)
}
