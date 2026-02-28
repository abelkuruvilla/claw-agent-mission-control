-- name: GetAgent :one
SELECT * FROM agents WHERE id = ? LIMIT 1;

-- name: ListAgents :many
SELECT * FROM agents ORDER BY created_at DESC;

-- name: CreateAgent :one
INSERT INTO agents (id, name, description, status, workspace_path, agent_dir_path, model, mention_patterns, soul_md, agents_md, identity_md, user_md, tools_md, heartbeat_md, memory_md)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateAgent :one
UPDATE agents SET 
    name = ?, description = ?, status = ?, model = ?, mention_patterns = ?,
    soul_md = ?, agents_md = ?, identity_md = ?, user_md = ?, tools_md = ?, heartbeat_md = ?,
    active_session_key = ?, current_task_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? RETURNING *;

-- name: DeleteAgent :exec
DELETE FROM agents WHERE id = ?;

-- name: UpdateAgentStatus :exec
UPDATE agents SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;
