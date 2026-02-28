-- name: CreateEvent :one
INSERT INTO events (id, task_id, agent_id, type, message, details)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListEvents :many
SELECT * FROM events ORDER BY created_at DESC LIMIT ?;

-- name: ListEventsByTask :many
SELECT * FROM events WHERE task_id = ? ORDER BY created_at DESC LIMIT ?;

-- name: ListEventsByAgent :many  
SELECT * FROM events WHERE agent_id = ? ORDER BY created_at DESC LIMIT ?;
