-- Add git branch fields to projects schema
ALTER TABLE projects ADD COLUMN default_branch TEXT;
ALTER TABLE projects ADD COLUMN local_exec_branch TEXT;
ALTER TABLE projects ADD COLUMN remote_merge_branch TEXT;
