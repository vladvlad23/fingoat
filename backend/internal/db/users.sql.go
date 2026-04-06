package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, monthly_income, payment_day, created_at`

type CreateUserParams struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Email, arg.PasswordHash)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash, &i.MonthlyIncome, &i.PaymentDay, &i.CreatedAt)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, password_hash, monthly_income, payment_day, created_at FROM users WHERE email = $1`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash, &i.MonthlyIncome, &i.PaymentDay, &i.CreatedAt)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, password_hash, monthly_income, payment_day, created_at FROM users WHERE id = $1`

func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash, &i.MonthlyIncome, &i.PaymentDay, &i.CreatedAt)
	return i, err
}

const updateUserIncome = `-- name: UpdateUserIncome :one
UPDATE users SET monthly_income = $2 WHERE id = $1 RETURNING id, email, password_hash, monthly_income, payment_day, created_at`

func (q *Queries) UpdateUserIncome(ctx context.Context, id int64, monthlyIncome pgtype.Numeric) (User, error) {
	row := q.db.QueryRow(ctx, updateUserIncome, id, monthlyIncome)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash, &i.MonthlyIncome, &i.PaymentDay, &i.CreatedAt)
	return i, err
}

const updateUserSettings = `-- name: UpdateUserSettings :one
UPDATE users SET monthly_income = $2, payment_day = $3 WHERE id = $1
RETURNING id, email, password_hash, monthly_income, payment_day, created_at`

type UpdateUserSettingsParams struct {
	ID            int64
	MonthlyIncome pgtype.Numeric
	PaymentDay    *int32
}

func (q *Queries) UpdateUserSettings(ctx context.Context, arg UpdateUserSettingsParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserSettings, arg.ID, arg.MonthlyIncome, arg.PaymentDay)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.PasswordHash, &i.MonthlyIncome, &i.PaymentDay, &i.CreatedAt)
	return i, err
}
