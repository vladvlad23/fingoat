-- name: GetTransactionSummary :many
SELECT type, COALESCE(category, '') AS category, SUM(amount) AS total
FROM transactions
WHERE user_id = $1
  AND ($2::date IS NULL OR date >= $2)
  AND ($3::date IS NULL OR date <= $3)
GROUP BY type, category
ORDER BY type, category;
