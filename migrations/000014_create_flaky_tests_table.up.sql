-- Create flaky_tests table for tracking flaky test detection
CREATE TABLE IF NOT EXISTS flaky_tests (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    project_id VARCHAR(255) NOT NULL,
    test_name TEXT NOT NULL,
    suite_name TEXT,
    flake_rate DOUBLE PRECISION DEFAULT 0,
    total_executions INTEGER DEFAULT 0,
    flaky_executions INTEGER DEFAULT 0,
    last_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    severity VARCHAR(50),
    last_error_message TEXT
);

-- Indexes for performance
CREATE INDEX idx_flaky_tests_deleted_at ON flaky_tests(deleted_at);
CREATE INDEX idx_flaky_tests_project_id ON flaky_tests(project_id);
CREATE INDEX idx_flaky_tests_test_name ON flaky_tests(test_name);
CREATE INDEX idx_flaky_tests_suite_name ON flaky_tests(suite_name);
CREATE INDEX idx_flaky_tests_status ON flaky_tests(status);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_flaky_tests_updated_at BEFORE UPDATE
    ON flaky_tests FOR EACH ROW EXECUTE PROCEDURE 
    update_updated_at_column();