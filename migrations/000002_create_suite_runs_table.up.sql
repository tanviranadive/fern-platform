-- Create suite_runs table
CREATE TABLE IF NOT EXISTS suite_runs (
    id BIGSERIAL PRIMARY KEY,
    test_run_id BIGINT NOT NULL,
    suite_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'running',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    total_specs INTEGER DEFAULT 0,
    passed_specs INTEGER DEFAULT 0,
    failed_specs INTEGER DEFAULT 0,
    skipped_specs INTEGER DEFAULT 0,
    duration_ms BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_suite_runs_test_run_id 
        FOREIGN KEY (test_run_id) 
        REFERENCES test_runs(id) 
        ON DELETE CASCADE
);

-- Create indexes for suite_runs
CREATE INDEX IF NOT EXISTS idx_suite_runs_test_run_id ON suite_runs(test_run_id);
CREATE INDEX IF NOT EXISTS idx_suite_runs_suite_name ON suite_runs(suite_name);
CREATE INDEX IF NOT EXISTS idx_suite_runs_status ON suite_runs(status);
CREATE INDEX IF NOT EXISTS idx_suite_runs_start_time ON suite_runs(start_time);
CREATE INDEX IF NOT EXISTS idx_suite_runs_deleted_at ON suite_runs(deleted_at);

-- Add updated_at trigger
CREATE TRIGGER update_suite_runs_updated_at BEFORE UPDATE ON suite_runs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();