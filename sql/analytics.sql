-- name: CountApplicationsByUser :one
SELECT COUNT(*)::bigint AS total
FROM applications
WHERE user_id = $1;

-- name: CountApplicationsByStatus :many
SELECT status, COUNT(*)::bigint AS total
FROM applications
WHERE user_id = $1
GROUP BY status
ORDER BY status;

-- name: CountApplicationsBySource :many
SELECT source, COUNT(*)::bigint AS total
FROM applications
WHERE user_id = $1
GROUP BY source
ORDER BY source;

-- name: CountApplicationsByCategory :many
SELECT COALESCE(NULLIF(category, ''), 'Uncategorized')::text AS category, COUNT(*)::bigint AS total
FROM applications
WHERE user_id = $1
GROUP BY COALESCE(NULLIF(category, ''), 'Uncategorized')::text
ORDER BY category;

-- name: CountApplicationsBySavedMonth :many
SELECT to_char(date_trunc('month', created_at), 'YYYY-MM') AS month, COUNT(*)::bigint AS total
FROM applications
WHERE user_id = $1
GROUP BY date_trunc('month', created_at)
ORDER BY date_trunc('month', created_at);
