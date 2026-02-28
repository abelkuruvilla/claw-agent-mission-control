CREATE TABLE events (
    id TEXT PRIMARY KEY,
    task_id TEXT REFERENCES tasks(id),
    agent_id TEXT REFERENCES agents(id),
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    details TEXT, -- JSON
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_task ON events(task_id);
CREATE INDEX idx_events_created ON events(created_at DESC);
