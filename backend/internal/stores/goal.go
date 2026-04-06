package stores

import (
	"context"

	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// GoalStore defines goal-related database operations.
type GoalStore interface {
	CreateGoal(ctx context.Context, arg db.CreateGoalParams) (db.Goal, error)
	ListGoalsByUser(ctx context.Context, userID int64) ([]db.Goal, error)
	GetGoalByID(ctx context.Context, id int64) (db.Goal, error)
	UpdateGoal(ctx context.Context, arg db.UpdateGoalParams) (db.Goal, error)
	DeleteGoal(ctx context.Context, id int64) error
	UpdateGoalCurrentAmount(ctx context.Context, id int64, delta pgtype.Numeric) (db.Goal, error)
}
