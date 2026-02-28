-- Add scheduled execution time for tasks.
ALTER TABLE tasks ADD COLUMN scheduled_at DATETIME;
ALTER TABLE tasks ADD COLUMN retry_at DATETIME;
