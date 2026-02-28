CREATE TABLE sub_agents (
    id TEXT PRIMARY KEY,
    orchestrator_id TEXT NOT NULL REFERENCES agents(id),
    task_id TEXT REFERENCES tasks(id),
    name TEXT NOT NULL,
    status TEXT DEFAULT 'running',
    session_key TEXT,
    session_label TEXT,
    purpose TEXT,
    iteration INTEGER DEFAULT 0,
    output TEXT,
    error TEXT,
    spawned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);
