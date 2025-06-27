-- Drop scope-based permissions tables

DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS project_permissions;
DROP TABLE IF EXISTS user_scopes;

-- Remove team column from project_details
ALTER TABLE project_details DROP COLUMN IF EXISTS team;