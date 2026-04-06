-- name: CreateTransaction :one
INSERT INTO transactions (user_id, goal_id, title, amount, type, category, date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
  AND ($2::bigint IS NULL OR goal_id = $2)
  AND ($3::text IS NULL OR type = $3)
  AND ($4::date IS NULL OR date >= $4)
  AND ($5::date IS NULL OR date <= $5)
ORDER BY date DESC, created_at DESC;

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: UpdateTransaction :one
UPDATE transactions
SET title = $2, amount = $3, type = $4, category = $5, date = $6, goal_id = $7
WHERE id = $1
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1;
