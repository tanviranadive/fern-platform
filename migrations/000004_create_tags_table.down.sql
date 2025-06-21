-- Drop tags and test_run_tags tables
DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;
DROP TABLE IF EXISTS test_run_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;