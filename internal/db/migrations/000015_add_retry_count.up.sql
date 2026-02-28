-- Add retry_count for watchdog: tracks re-notification attempts for stuck tasks
ALTER TABLE tasks ADD COLUMN retry_count INTEGER NOT NULL DEFAULT 0;
