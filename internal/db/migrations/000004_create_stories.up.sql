CREATE TABLE stories (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 3,
    passes BOOLEAN DEFAULT FALSE,
    acceptance_criteria TEXT, -- JSON array
    iterations INTEGER DEFAULT 0,
    last_error TEXT,
    session_key TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
