-- name: GetPhase :one
SELECT * FROM phases WHERE id = ? LIMIT 1;

-- name: ListPhasesByTask :many
SELECT * FROM phases WHERE task_id = ? ORDER BY sequence ASC;

-- name: CreatePhase :one
INSERT INTO phases (id, task_id, sequence, title, description, status)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePhase :one
UPDATE phases SET
    title = ?, description = ?, status = ?,
    context_md = ?, research_md = ?, plan_md = ?, summary_md = ?, uat_md = ?,
    verification_result = ?, session_key = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? RETURNING *;

-- name: UpdatePhaseStatus :exec
UPDATE phases SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeletePhase :exec
DELETE FROM phases WHERE id = ?;
