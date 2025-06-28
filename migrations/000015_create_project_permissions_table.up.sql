CREATE TABLE IF NOT EXISTS project_permissions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    project_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    permission VARCHAR(50) NOT NULL,
    granted_by VARCHAR(255),
    expires_at TIMESTAMP
);

-- Create indexes for efficient queries
CREATE INDEX idx_project_permissions_deleted_at ON project_permissions(deleted_at);
CREATE INDEX idx_project_permissions_project_id ON project_permissions(project_id);
CREATE INDEX idx_project_permissions_user_id ON project_permissions(user_id);
CREATE UNIQUE INDEX idx_project_user_perm ON project_permissions(project_id, user_id, permission) WHERE deleted_at IS NULL;