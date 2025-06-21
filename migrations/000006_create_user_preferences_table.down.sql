-- Drop user_preferences table
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP TABLE IF EXISTS user_preferences CASCADE;