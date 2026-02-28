-- Add default project directory to settings
ALTER TABLE settings ADD COLUMN default_project_directory TEXT;

-- Add location to projects
ALTER TABLE projects ADD COLUMN location TEXT;
