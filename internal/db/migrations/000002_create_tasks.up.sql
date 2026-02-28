CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    agent_id TEXT REFERENCES agents(id),
    approach TEXT DEFAULT 'gsd', -- 'gsd' or 'ralph'
    status TEXT DEFAULT 'backlog',
    priority INTEGER DEFAULT 3,
    work_dir TEXT,
    git_branch TEXT,
    project_md TEXT,
    requirements_md TEXT,
    roadmap_md TEXT,
    state_md TEXT,
    prd_json TEXT,
    progress_txt TEXT,
    quality_checks TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME
);
