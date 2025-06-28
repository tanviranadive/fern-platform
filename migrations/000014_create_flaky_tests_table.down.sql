-- Drop flaky_tests table
DROP TRIGGER IF EXISTS update_flaky_tests_updated_at ON flaky_tests;
DROP TABLE IF EXISTS flaky_tests;