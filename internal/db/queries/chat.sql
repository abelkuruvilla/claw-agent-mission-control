-- ============ Chat Sessions ============

-- name: CreateChatSession :one
INSERT INTO chat_sessions (id, agent_id, openclaw_session_key, status)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetChatSession :one
SELECT * FROM chat_sessions WHERE id = ? LIMIT 1;

-- name: ListChatSessionsByAgent :many
SELECT * FROM chat_sessions 
WHERE agent_id = ? 
ORDER BY started_at DESC;

-- name: EndChatSession :exec
UPDATE chat_sessions 
SET status = 'ended', ended_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: UpdateMessageCount :exec
UPDATE chat_sessions 
SET message_count = message_count + 1 
WHERE id = ?;

-- ============ Chat Messages ============

-- name: CreateChatMessage :one
INSERT INTO chat_messages (id, session_id, role, content)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListMessagesBySession :many
SELECT * FROM chat_messages 
WHERE session_id = ? 
ORDER BY created_at ASC;
