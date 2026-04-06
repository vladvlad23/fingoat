package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTransaction = `-- name: CreateTransaction :one
INSERT INTO transactions (user_id, goal_id, title, amount, type, category, date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, goal_id, title, amount, type, category, date, created_at`

type CreateTransactionParams struct {
	UserID   int64          `json:"user_id"`
	GoalID   *int64         `json:"goal_id"`
	Title    string         `json:"title"`
	Amount   pgtype.Numeric `json:"amount"`
	Type     string         `json:"type"`
	Category *string        `json:"category"`
	Date     pgtype.Date    `json:"date"`
}

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error) {
	row := q.db.QueryRow(ctx, createTransaction, arg.UserID, arg.GoalID, arg.Title, arg.Amount, arg.Type, arg.Category, arg.Date)
	var i Transaction
	err := row.Scan(&i.ID, &i.UserID, &i.GoalID, &i.Title, &i.Amount, &i.Type, &i.Category, &i.Date, &i.CreatedAt)
	return i, err
}

const listTransactions = `-- name: ListTransactions :many
SELECT id, user_id, goal_id, title, amount, type, category, date, created_at FROM transactions
WHERE user_id = $1
  AND ($2::bigint IS NULL OR goal_id = $2)
  AND ($3::text IS NULL OR type = $3)
  AND ($4::date IS NULL OR date >= $4)
  AND ($5::date IS NULL OR date <= $5)
ORDER BY date DESC, created_at DESC`

type ListTransactionsParams struct {
	UserID int64        `json:"user_id"`
	GoalID *int64       `json:"goal_id"`
	Type   *string      `json:"type"`
	From   *pgtype.Date `json:"from"`
	To     *pgtype.Date `json:"to"`
}

func (q *Queries) ListTransactions(ctx context.Context, arg ListTransactionsParams) ([]Transaction, error) {
	rows, err := q.db.Query(ctx, listTransactions, arg.UserID, arg.GoalID, arg.Type, arg.From, arg.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Transaction
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(&i.ID, &i.UserID, &i.GoalID, &i.Title, &i.Amount, &i.Type, &i.Category, &i.Date, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const getTransactionByID = `-- name: GetTransactionByID :one
SELECT id, user_id, goal_id, title, amount, type, category, date, created_at FROM transactions WHERE id = $1`

func (q *Queries) GetTransactionByID(ctx context.Context, id int64) (Transaction, error) {
	row := q.db.QueryRow(ctx, getTransactionByID, id)
	var i Transaction
	err := row.Scan(&i.ID, &i.UserID, &i.GoalID, &i.Title, &i.Amount, &i.Type, &i.Category, &i.Date, &i.CreatedAt)
	return i, err
}

const updateTransaction = `-- name: UpdateTransaction :one
UPDATE transactions
SET title = $2, amount = $3, type = $4, category = $5, date = $6, goal_id = $7
WHERE id = $1
RETURNING id, user_id, goal_id, title, amount, type, category, date, created_at`

type UpdateTransactionParams struct {
	ID       int64          `json:"id"`
	Title    string         `json:"title"`
	Amount   pgtype.Numeric `json:"amount"`
	Type     string         `json:"type"`
	Category *string        `json:"category"`
	Date     pgtype.Date    `json:"date"`
	GoalID   *int64         `json:"goal_id"`
}

func (q *Queries) UpdateTransaction(ctx context.Context, arg UpdateTransactionParams) (Transaction, error) {
	row := q.db.QueryRow(ctx, updateTransaction, arg.ID, arg.Title, arg.Amount, arg.Type, arg.Category, arg.Date, arg.GoalID)
	var i Transaction
	err := row.Scan(&i.ID, &i.UserID, &i.GoalID, &i.Title, &i.Amount, &i.Type, &i.Category, &i.Date, &i.CreatedAt)
	return i, err
}

const deleteTransaction = `-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1`

func (q *Queries) DeleteTransaction(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteTransaction, id)
	return err
}
