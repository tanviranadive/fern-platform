-- Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    color VARCHAR(7), -- For hex color codes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create test_run_tags junction table
CREATE TABLE IF NOT EXISTS test_run_tags (
    test_run_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (test_run_id, tag_id),
    CONSTRAINT fk_test_run_tags_test_run_id 
        FOREIGN KEY (test_run_id) 
        REFERENCES test_runs(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_test_run_tags_tag_id 
        FOREIGN KEY (tag_id) 
        REFERENCES tags(id) 
        ON DELETE CASCADE
);

-- Create indexes for tags
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON tags(deleted_at);

-- Create indexes for test_run_tags
CREATE INDEX IF NOT EXISTS idx_test_run_tags_test_run_id ON test_run_tags(test_run_id);
CREATE INDEX IF NOT EXISTS idx_test_run_tags_tag_id ON test_run_tags(tag_id);

-- Add updated_at trigger for tags
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();