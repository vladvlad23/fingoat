-- name: CreateUser :one
INSERT INTO users (email, password_hash, monthly_income)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserIncome :one
UPDATE users SET monthly_income = $2 WHERE id = $1 RETURNING *;
