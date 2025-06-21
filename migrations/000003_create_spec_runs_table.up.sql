-- Create spec_runs table
CREATE TABLE IF NOT EXISTS spec_runs (
    id BIGSERIAL PRIMARY KEY,
    suite_run_id BIGINT NOT NULL,
    spec_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT DEFAULT 0,
    error_message TEXT,
    stack_trace TEXT,
    retry_count INTEGER DEFAULT 0,
    is_flaky BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_spec_runs_suite_run_id 
        FOREIGN KEY (suite_run_id) 
        REFERENCES suite_runs(id) 
        ON DELETE CASCADE
);

-- Create indexes for spec_runs
CREATE INDEX IF NOT EXISTS idx_spec_runs_suite_run_id ON spec_runs(suite_run_id);
CREATE INDEX IF NOT EXISTS idx_spec_runs_spec_name ON spec_runs(spec_name);
CREATE INDEX IF NOT EXISTS idx_spec_runs_status ON spec_runs(status);
CREATE INDEX IF NOT EXISTS idx_spec_runs_start_time ON spec_runs(start_time);
CREATE INDEX IF NOT EXISTS idx_spec_runs_is_flaky ON spec_runs(is_flaky);
CREATE INDEX IF NOT EXISTS idx_spec_runs_deleted_at ON spec_runs(deleted_at);

-- Add updated_at trigger
CREATE TRIGGER update_spec_runs_updated_at BEFORE UPDATE ON spec_runs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();