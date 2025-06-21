-- Drop test_runs table and related objects
DROP TRIGGER IF EXISTS update_test_runs_updated_at ON test_runs;
DROP TABLE IF EXISTS test_runs CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column();