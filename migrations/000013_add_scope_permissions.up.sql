-- Add scope-based permissions tables

-- Create user_scopes table
CREATE TABLE IF NOT EXISTS user_scopes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL,
    scope VARCHAR(255) NOT NULL,
    granted_by VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT idx_user_scope UNIQUE (user_id, scope)
);

-- Create indexes for user_scopes
CREATE INDEX IF NOT EXISTS idx_user_scopes_user_id ON user_scopes(user_id);
CREATE INDEX IF NOT EXISTS idx_user_scopes_deleted_at ON user_scopes(deleted_at);

-- Create project_permissions table
CREATE TABLE IF NOT EXISTS project_permissions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    project_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    permission VARCHAR(50) NOT NULL,
    granted_by VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT idx_project_user_perm UNIQUE (project_id, user_id, permission)
);

-- Create indexes for project_permissions
CREATE INDEX IF NOT EXISTS idx_project_permissions_user_id ON project_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_project_permissions_project_id ON project_permissions(project_id);
CREATE INDEX IF NOT EXISTS idx_project_permissions_deleted_at ON project_permissions(deleted_at);

-- Add team column to project_details if not exists
ALTER TABLE project_details ADD COLUMN IF NOT EXISTS team VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_project_details_team ON project_details(team);

-- Create user_groups table
CREATE TABLE IF NOT EXISTS user_groups (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    group_name VARCHAR(255) NOT NULL
);

-- Create indexes for user_groups
CREATE INDEX IF NOT EXISTS idx_user_groups_user_id ON user_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_user_groups_group_name ON user_groups(group_name);