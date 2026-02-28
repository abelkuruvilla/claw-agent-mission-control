-- name: GetComment :one
SELECT * FROM comments WHERE id = ? LIMIT 1;

-- name: ListCommentsByTask :many
SELECT * FROM comments WHERE task_id = ? ORDER BY created_at ASC;

-- name: CreateComment :one
INSERT INTO comments (id, task_id, author, content)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = ?;
