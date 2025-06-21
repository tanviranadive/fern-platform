-- Drop project_details table
DROP TRIGGER IF EXISTS update_project_details_updated_at ON project_details;
DROP TABLE IF EXISTS project_details CASCADE;