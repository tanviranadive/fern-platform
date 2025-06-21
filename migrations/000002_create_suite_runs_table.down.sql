-- Drop suite_runs table
DROP TRIGGER IF EXISTS update_suite_runs_updated_at ON suite_runs;
DROP TABLE IF EXISTS suite_runs CASCADE;