-- Add id_token column to user_sessions table for proper logout support
ALTER TABLE user_sessions ADD COLUMN id_token TEXT DEFAULT '';