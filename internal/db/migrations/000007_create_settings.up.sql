CREATE TABLE settings (
    id TEXT PRIMARY KEY DEFAULT 'default',
    openclaw_gateway_url TEXT,
    openclaw_gateway_token TEXT,
    default_approach TEXT DEFAULT 'gsd',
    default_model TEXT,
    max_parallel_executions INTEGER DEFAULT 3,
    gsd_depth TEXT DEFAULT 'standard',
    gsd_mode TEXT DEFAULT 'interactive',
    gsd_research_enabled BOOLEAN DEFAULT TRUE,
    gsd_plan_check_enabled BOOLEAN DEFAULT TRUE,
    gsd_verifier_enabled BOOLEAN DEFAULT TRUE,
    ralph_max_iterations INTEGER DEFAULT 10,
    ralph_auto_commit BOOLEAN DEFAULT TRUE,
    theme TEXT DEFAULT 'dark',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO settings (id) VALUES ('default');
