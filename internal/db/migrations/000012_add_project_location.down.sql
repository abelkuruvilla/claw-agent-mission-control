-- SQLite doesn't support DROP COLUMN in older versions
-- For newer SQLite (3.35+), we can use:
ALTER TABLE settings DROP COLUMN default_project_directory;
ALTER TABLE projects DROP COLUMN location;
