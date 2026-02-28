-- name: GetTask :one
SELECT * FROM tasks WHERE id = ? LIMIT 1;

-- name: ListTasks :many
SELECT * FROM tasks ORDER BY priority ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks WHERE status = ? ORDER BY priority ASC, created_at DESC;

-- name: ListTasksByAgent :many
SELECT * FROM tasks WHERE agent_id = ? ORDER BY created_at DESC;

-- name: CreateTask :one
INSERT INTO tasks (id, title, description, agent_id, project_id, parent_task_id, status, priority, quality_checks, delegation_mode, scheduled_at, git_branch)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTaskWithStoryCounts :one
SELECT 
    t.*,
    (SELECT COUNT(*) FROM stories WHERE task_id = t.id) as stories_total,
    (SELECT COUNT(*) FROM stories WHERE task_id = t.id AND passes = 1) as stories_passed
FROM tasks t WHERE t.id = ? LIMIT 1;

-- name: ListTasksWithStoryCounts :many
SELECT 
    t.*,
    (SELECT COUNT(*) FROM stories WHERE task_id = t.id) as stories_total,
    (SELECT COUNT(*) FROM stories WHERE task_id = t.id AND passes = 1) as stories_passed
FROM tasks t ORDER BY t.priority ASC, t.created_at DESC;

-- name: UpdateTask :one
UPDATE tasks SET
    title = ?, description = ?, agent_id = ?, project_id = ?, status = ?, priority = ?,
    project_md = ?, requirements_md = ?, roadmap_md = ?, state_md = ?,
    prd_json = ?, progress_txt = ?, git_branch = ?, quality_checks = ?,
    delegation_mode = ?, scheduled_at = ?, retry_at = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? RETURNING *;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- name: ListTasksByProject :many
SELECT * FROM tasks WHERE project_id = ? ORDER BY priority ASC, created_at DESC;

-- name: ListSubtasks :many
SELECT * FROM tasks WHERE parent_task_id = ? ORDER BY created_at ASC;

-- name: ListQueuedTasksByAgent :many
SELECT * FROM tasks WHERE agent_id = ? AND status = 'queued' ORDER BY priority ASC, created_at ASC;

-- name: CountActiveTasksByAgent :one
SELECT COUNT(*) FROM tasks WHERE agent_id = ? AND status IN ('executing', 'planning', 'discussing', 'verifying');

-- name: ListStaleTasks :many
SELECT * FROM tasks
WHERE status IN ('executing', 'planning', 'discussing', 'verifying')
  AND (updated_at IS NULL OR updated_at < ?)
ORDER BY updated_at ASC;

-- name: IncrementTaskRetryCount :exec
UPDATE tasks SET retry_count = retry_count + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ResetStuckTask :exec
UPDATE tasks SET status = 'backlog', agent_id = NULL, retry_count = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ResetTaskRetryCount :exec
UPDATE tasks SET retry_count = 0 WHERE id = ?;

-- name: AppendProgressTxt :exec
UPDATE tasks SET progress_txt = COALESCE(progress_txt || char(10), '') || ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetTaskScheduledAt :exec
UPDATE tasks SET scheduled_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetTaskRetryAt :exec
UPDATE tasks SET retry_at = ?, status = 'backlog', updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ClearTaskScheduledAt :exec
UPDATE tasks SET scheduled_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ClearTaskRetryAt :exec
UPDATE tasks SET retry_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: ListScheduledDueTasks :many
SELECT * FROM tasks
WHERE scheduled_at IS NOT NULL
  AND scheduled_at <= CURRENT_TIMESTAMP
  AND status = 'backlog'
ORDER BY scheduled_at ASC;

-- name: ListRetryDueTasks :many
SELECT * FROM tasks
WHERE retry_at IS NOT NULL
  AND retry_at <= CURRENT_TIMESTAMP
  AND status = 'backlog'
ORDER BY retry_at ASC;
