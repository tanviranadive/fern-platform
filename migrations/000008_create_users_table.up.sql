-- Create users table for OAuth authentication with proper relational schema
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL UNIQUE,     -- OAuth provider user ID
    email VARCHAR(255) NOT NULL UNIQUE,       -- User email
    name VARCHAR(255) NOT NULL,               -- Display name
    role VARCHAR(50) NOT NULL DEFAULT 'user', -- user, admin
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, inactive
    last_login_at TIMESTAMP WITH TIME ZONE,
    profile_url VARCHAR(500),                 -- Avatar/profile picture URL
    first_name VARCHAR(255),                  -- First name from OAuth
    last_name VARCHAR(255),                   -- Last name from OAuth
    email_verified BOOLEAN DEFAULT false     -- Email verification status
);

-- Create indexes
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE UNIQUE INDEX idx_users_user_id ON users(user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- Create user_groups table for group membership (relational instead of JSONB)
CREATE TABLE user_groups (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL,        -- References users(user_id)
    group_name VARCHAR(255) NOT NULL      -- Group name from OAuth provider
);

-- Create indexes for user_groups table
CREATE INDEX idx_user_groups_deleted_at ON user_groups(deleted_at);
CREATE INDEX idx_user_groups_user_id ON user_groups(user_id);
CREATE INDEX idx_user_groups_group_name ON user_groups(group_name);
CREATE UNIQUE INDEX idx_user_groups_user_group ON user_groups(user_id, group_name) WHERE deleted_at IS NULL;

-- Add foreign key constraint for user_groups
ALTER TABLE user_groups ADD CONSTRAINT fk_user_groups_user_id 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;