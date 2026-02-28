CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'idle',
    workspace_path TEXT,
    agent_dir_path TEXT,
    model TEXT,
    mention_patterns TEXT, -- JSON array
    soul_md TEXT,
    agents_md TEXT,
    identity_md TEXT,
    user_md TEXT,
    tools_md TEXT,
    heartbeat_md TEXT,
    memory_md TEXT,
    active_session_key TEXT,
    current_task_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
