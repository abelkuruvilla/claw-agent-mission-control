-- name: GetProject :one
SELECT * FROM projects WHERE id = ? LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: ListProjectsByStatus :many
SELECT * FROM projects WHERE status = ? ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (id, name, description, status, color, location, default_branch, local_exec_branch, remote_merge_branch)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects SET
    name = ?, 
    description = ?, 
    status = ?, 
    color = ?,
    location = ?,
    default_branch = ?,
    local_exec_branch = ?,
    remote_merge_branch = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? 
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- name: GetProjectTaskCount :one
SELECT COUNT(*) as count FROM tasks WHERE project_id = ?;

-- name: GetProjectDoneTaskCount :one
SELECT COUNT(*) as count FROM tasks WHERE project_id = ? AND status = 'done';
