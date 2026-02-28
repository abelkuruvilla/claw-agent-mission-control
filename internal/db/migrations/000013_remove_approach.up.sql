-- Migration: Remove approach field from tasks and default_approach from settings
-- Reason: GSD and Ralph are now used together (GSD for planning, Ralph for execution)
-- All tasks automatically use both protocols based on their status

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the tables

-- Step 1: Recreate tasks table without approach column
CREATE TABLE tasks_new (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    agent_id TEXT REFERENCES agents(id) ON DELETE SET NULL,
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    parent_task_id TEXT REFERENCES tasks(id) ON DELETE SET NULL,
    status TEXT DEFAULT 'backlog',
    priority INTEGER DEFAULT 3,
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

-- Copy data (excluding approach and work_dir columns)
INSERT INTO tasks_new (
    id, title, description, agent_id, project_id, parent_task_id,
    status, priority, git_branch, project_md, requirements_md,
    roadmap_md, state_md, prd_json, progress_txt, quality_checks,
    created_at, updated_at, started_at, completed_at
)
SELECT 
    id, title, description, agent_id, project_id, parent_task_id,
    status, priority, git_branch, project_md, requirements_md,
    roadmap_md, state_md, prd_json, progress_txt, quality_checks,
    created_at, updated_at, started_at, completed_at
FROM tasks;

-- Drop old table and rename new one
DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

-- Recreate indexes for tasks
CREATE INDEX idx_tasks_agent_id ON tasks(agent_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);

-- Step 2: Recreate settings table without default_approach column
CREATE TABLE settings_new (
    id TEXT PRIMARY KEY DEFAULT 'default',
    openclaw_gateway_url TEXT,
    openclaw_gateway_token TEXT,
    default_model TEXT,
    max_parallel_executions INTEGER DEFAULT 3,
    default_project_directory TEXT,
    gsd_depth TEXT DEFAULT 'standard',
    gsd_mode TEXT DEFAULT 'interactive',
    gsd_research_enabled INTEGER DEFAULT 1,
    gsd_plan_check_enabled INTEGER DEFAULT 1,
    gsd_verifier_enabled INTEGER DEFAULT 1,
    ralph_max_iterations INTEGER DEFAULT 10,
    ralph_auto_commit INTEGER DEFAULT 1,
    theme TEXT DEFAULT 'dark',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data (excluding default_approach)
INSERT INTO settings_new (
    id, openclaw_gateway_url, openclaw_gateway_token, default_model,
    max_parallel_executions, default_project_directory,
    gsd_depth, gsd_mode, gsd_research_enabled, gsd_plan_check_enabled,
    gsd_verifier_enabled, ralph_max_iterations, ralph_auto_commit,
    theme, updated_at
)
SELECT 
    id, openclaw_gateway_url, openclaw_gateway_token, default_model,
    max_parallel_executions, default_project_directory,
    gsd_depth, gsd_mode, gsd_research_enabled, gsd_plan_check_enabled,
    gsd_verifier_enabled, ralph_max_iterations, ralph_auto_commit,
    theme, updated_at
FROM settings;

-- Drop old table and rename new one
DROP TABLE settings;
ALTER TABLE settings_new RENAME TO settings;
