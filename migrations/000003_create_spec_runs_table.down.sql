-- Drop spec_runs table
DROP TRIGGER IF EXISTS update_spec_runs_updated_at ON spec_runs;
DROP TABLE IF EXISTS spec_runs CASCADE;