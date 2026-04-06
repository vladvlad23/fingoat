-- name: CreateGoal :one
INSERT INTO goals (user_id, title, target_amount, deadline)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListGoalsByUser :many
SELECT * FROM goals WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetGoalByID :one
SELECT * FROM goals WHERE id = $1;

-- name: UpdateGoal :one
UPDATE goals
SET title = $2, target_amount = $3, deadline = $4, status = $5
WHERE id = $1
RETURNING *;

-- name: DeleteGoal :exec
DELETE FROM goals WHERE id = $1;

-- name: UpdateGoalCurrentAmount :one
UPDATE goals
SET current_amount = current_amount + $2,
    status = CASE WHEN current_amount + $2 >= target_amount THEN 'completed' ELSE status END
WHERE id = $1
RETURNING *;
