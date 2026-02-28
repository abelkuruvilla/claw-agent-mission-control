-- Add delegation_mode to tasks: 'auto' (default) or 'manual'
-- auto: orchestrator is notified immediately when subtasks complete
-- manual: subtask completion requires human approval before orchestrator is notified
ALTER TABLE tasks ADD COLUMN delegation_mode TEXT DEFAULT 'auto';
