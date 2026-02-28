-- Remove indexes
DROP INDEX idx_tasks_parent_task_id;
DROP INDEX idx_tasks_project_id;

-- Note: SQLite doesn't support DROP COLUMN directly
-- This would require recreating the table without these columns
-- For now, we'll leave the columns but document the limitation
-- ALTER TABLE tasks DROP COLUMN parent_task_id;
-- ALTER TABLE tasks DROP COLUMN project_id;
