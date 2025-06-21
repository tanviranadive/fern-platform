-- Create flaky_tests table
CREATE TABLE IF NOT EXISTS flaky_tests (
    id BIGSERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    test_name VARCHAR(500) NOT NULL,
    suite_name VARCHAR(255),
    flake_rate DECIMAL(5,4) DEFAULT 0.0000, -- Percentage as decimal (0.0000 to 1.0000)
    total_executions INTEGER DEFAULT 0,
    flaky_executions INTEGER DEFAULT 0,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    severity VARCHAR(20), -- low, medium, high, critical
    last_error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(project_id, test_name, suite_name)
);

-- Create indexes for flaky_tests
CREATE INDEX IF NOT EXISTS idx_flaky_tests_project_id ON flaky_tests(project_id);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_test_name ON flaky_tests(test_name);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_suite_name ON flaky_tests(suite_name);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_flake_rate ON flaky_tests(flake_rate);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_status ON flaky_tests(status);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_severity ON flaky_tests(severity);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_last_seen_at ON flaky_tests(last_seen_at);
CREATE INDEX IF NOT EXISTS idx_flaky_tests_deleted_at ON flaky_tests(deleted_at);

-- Add updated_at trigger
CREATE TRIGGER update_flaky_tests_updated_at BEFORE UPDATE ON flaky_tests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();