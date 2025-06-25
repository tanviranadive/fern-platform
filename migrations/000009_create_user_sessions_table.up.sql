-- Create user_sessions table for OAuth session management
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id VARCHAR(255) NOT NULL,        -- References users(user_id)
    session_id VARCHAR(255) NOT NULL UNIQUE,
    access_token TEXT,                    -- OAuth access token
    refresh_token TEXT,                   -- OAuth refresh token
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    ip_address VARCHAR(45),               -- IPv4/IPv6 address
    user_agent TEXT,
    last_activity TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_user_sessions_deleted_at ON user_sessions(deleted_at);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX idx_user_sessions_last_activity ON user_sessions(last_activity);
CREATE UNIQUE INDEX idx_user_sessions_session_id ON user_sessions(session_id) WHERE deleted_at IS NULL;

-- Add foreign key constraint
ALTER TABLE user_sessions ADD CONSTRAINT fk_user_sessions_user_id 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;