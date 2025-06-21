-- Create project_details table
CREATE TABLE IF NOT EXISTS project_details (
    id BIGSERIAL PRIMARY KEY,
    project_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    repository VARCHAR(500),
    default_branch VARCHAR(255) DEFAULT 'main',
    settings JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for project_details
CREATE INDEX IF NOT EXISTS idx_project_details_project_id ON project_details(project_id);
CREATE INDEX IF NOT EXISTS idx_project_details_name ON project_details(name);
CREATE INDEX IF NOT EXISTS idx_project_details_is_active ON project_details(is_active);
CREATE INDEX IF NOT EXISTS idx_project_details_deleted_at ON project_details(deleted_at);

-- Add updated_at trigger
CREATE TRIGGER update_project_details_updated_at BEFORE UPDATE ON project_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();