-- name: GetStory :one
SELECT * FROM stories WHERE id = ? LIMIT 1;

-- name: ListStoriesByTask :many
SELECT * FROM stories WHERE task_id = ? ORDER BY priority ASC, sequence ASC;

-- name: GetNextPendingStory :one
SELECT * FROM stories WHERE task_id = ? AND passes = FALSE ORDER BY priority ASC, sequence ASC LIMIT 1;

-- name: CreateStory :one
INSERT INTO stories (id, task_id, sequence, title, description, priority, acceptance_criteria)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateStory :one
UPDATE stories SET
    title = ?, description = ?, priority = ?, passes = ?,
    acceptance_criteria = ?, iterations = ?, last_error = ?,
    session_key = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? RETURNING *;

-- name: MarkStoryPassed :exec
UPDATE stories SET passes = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: MarkStoryFailed :exec
UPDATE stories SET passes = FALSE, last_error = ?, iterations = iterations + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteStory :exec
DELETE FROM stories WHERE id = ?;

-- name: CountPassedStories :one
SELECT COUNT(*) FROM stories WHERE task_id = ? AND passes = TRUE;

-- name: CountTotalStories :one
SELECT COUNT(*) FROM stories WHERE task_id = ?;
