-- name: CreateApplicationNote :one
INSERT INTO application_notes (
    application_id,
    user_id,
    body
)
SELECT
    a.id,
    a.user_id,
    sqlc.arg(body)::text
FROM applications a
WHERE a.id = sqlc.arg(application_id)::uuid
  AND a.user_id = sqlc.arg(user_id)::uuid
RETURNING id, application_id, user_id, body, created_at;

-- name: GetApplicationNotes :many
SELECT n.id, n.application_id, n.user_id, n.body, n.created_at
FROM application_notes n
JOIN applications a ON a.id = n.application_id
WHERE n.application_id = $1
  AND n.user_id = $2
  AND a.user_id = $2
ORDER BY n.created_at DESC;

-- name: DeleteApplicationNote :one
DELETE FROM application_notes
WHERE id = $1
  AND user_id = $2
RETURNING id;
