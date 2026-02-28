-- name: GetSubAgent :one
SELECT * FROM sub_agents WHERE id = ? LIMIT 1;

-- name: ListSubAgentsByOrchestrator :many
SELECT * FROM sub_agents WHERE orchestrator_id = ? ORDER BY spawned_at DESC;

-- name: ListSubAgentsByTask :many
SELECT * FROM sub_agents WHERE task_id = ? ORDER BY spawned_at DESC;

-- name: CreateSubAgent :one
INSERT INTO sub_agents (id, orchestrator_id, task_id, name, status, session_key, session_label, purpose, iteration)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSubAgentStatus :exec
UPDATE sub_agents SET status = ?, output = ?, error = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteSubAgent :exec
DELETE FROM sub_agents WHERE id = ?;
