package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createGoal = `-- name: CreateGoal :one
INSERT INTO goals (user_id, title, target_amount, deadline)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, title, target_amount, current_amount, deadline, status, created_at`

type CreateGoalParams struct {
	UserID       int64          `json:"user_id"`
	Title        string         `json:"title"`
	TargetAmount pgtype.Numeric `json:"target_amount"`
	Deadline     *pgtype.Date   `json:"deadline"`
}

func (q *Queries) CreateGoal(ctx context.Context, arg CreateGoalParams) (Goal, error) {
	row := q.db.QueryRow(ctx, createGoal, arg.UserID, arg.Title, arg.TargetAmount, arg.Deadline)
	var i Goal
	err := row.Scan(&i.ID, &i.UserID, &i.Title, &i.TargetAmount, &i.CurrentAmount, &i.Deadline, &i.Status, &i.CreatedAt)
	return i, err
}

const listGoalsByUser = `-- name: ListGoalsByUser :many
SELECT id, user_id, title, target_amount, current_amount, deadline, status, created_at FROM goals WHERE user_id = $1 ORDER BY created_at DESC`

func (q *Queries) ListGoalsByUser(ctx context.Context, userID int64) ([]Goal, error) {
	rows, err := q.db.Query(ctx, listGoalsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Goal
	for rows.Next() {
		var i Goal
		if err := rows.Scan(&i.ID, &i.UserID, &i.Title, &i.TargetAmount, &i.CurrentAmount, &i.Deadline, &i.Status, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const getGoalByID = `-- name: GetGoalByID :one
SELECT id, user_id, title, target_amount, current_amount, deadline, status, created_at FROM goals WHERE id = $1`

func (q *Queries) GetGoalByID(ctx context.Context, id int64) (Goal, error) {
	row := q.db.QueryRow(ctx, getGoalByID, id)
	var i Goal
	err := row.Scan(&i.ID, &i.UserID, &i.Title, &i.TargetAmount, &i.CurrentAmount, &i.Deadline, &i.Status, &i.CreatedAt)
	return i, err
}

const updateGoal = `-- name: UpdateGoal :one
UPDATE goals
SET title = $2, target_amount = $3, deadline = $4, status = $5
WHERE id = $1
RETURNING id, user_id, title, target_amount, current_amount, deadline, status, created_at`

type UpdateGoalParams struct {
	ID           int64          `json:"id"`
	Title        string         `json:"title"`
	TargetAmount pgtype.Numeric `json:"target_amount"`
	Deadline     *pgtype.Date   `json:"deadline"`
	Status       string         `json:"status"`
}

func (q *Queries) UpdateGoal(ctx context.Context, arg UpdateGoalParams) (Goal, error) {
	row := q.db.QueryRow(ctx, updateGoal, arg.ID, arg.Title, arg.TargetAmount, arg.Deadline, arg.Status)
	var i Goal
	err := row.Scan(&i.ID, &i.UserID, &i.Title, &i.TargetAmount, &i.CurrentAmount, &i.Deadline, &i.Status, &i.CreatedAt)
	return i, err
}

const deleteGoal = `-- name: DeleteGoal :exec
DELETE FROM goals WHERE id = $1`

func (q *Queries) DeleteGoal(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteGoal, id)
	return err
}

const updateGoalCurrentAmount = `-- name: UpdateGoalCurrentAmount :one
UPDATE goals
SET current_amount = current_amount + $2,
    status = CASE WHEN current_amount + $2 >= target_amount THEN 'completed' ELSE status END
WHERE id = $1
RETURNING id, user_id, title, target_amount, current_amount, deadline, status, created_at`

func (q *Queries) UpdateGoalCurrentAmount(ctx context.Context, id int64, delta pgtype.Numeric) (Goal, error) {
	row := q.db.QueryRow(ctx, updateGoalCurrentAmount, id, delta)
	var i Goal
	err := row.Scan(&i.ID, &i.UserID, &i.Title, &i.TargetAmount, &i.CurrentAmount, &i.Deadline, &i.Status, &i.CreatedAt)
	return i, err
}
