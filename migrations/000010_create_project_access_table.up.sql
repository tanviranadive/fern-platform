-- Create project_access table for project-level permissions
CREATE TABLE project_access (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL,        -- References users(user_id)
    project_id VARCHAR(255) NOT NULL,     -- Project identifier
    role VARCHAR(50) NOT NULL,            -- viewer, editor, admin
    granted_by VARCHAR(255),              -- Who granted this access
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE   -- Optional expiration
);

-- Create indexes
CREATE INDEX idx_project_access_deleted_at ON project_access(deleted_at);
CREATE INDEX idx_project_access_user_id ON project_access(user_id);
CREATE INDEX idx_project_access_project_id ON project_access(project_id);
CREATE INDEX idx_project_access_role ON project_access(role);
CREATE UNIQUE INDEX idx_project_access_user_project ON project_access(user_id, project_id) WHERE deleted_at IS NULL;

-- Add foreign key constraint
ALTER TABLE project_access ADD CONSTRAINT fk_project_access_user_id 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;