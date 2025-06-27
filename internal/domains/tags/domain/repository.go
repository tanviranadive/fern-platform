package domain

import "context"

// TagRepository defines the interface for tag persistence
type TagRepository interface {
	// Save persists a tag
	Save(ctx context.Context, tag *Tag) error
	
	// FindByID retrieves a tag by ID
	FindByID(ctx context.Context, id TagID) (*Tag, error)
	
	// FindByName retrieves a tag by name
	FindByName(ctx context.Context, name string) (*Tag, error)
	
	// FindAll retrieves all tags
	FindAll(ctx context.Context) ([]*Tag, error)
	
	// Delete removes a tag
	Delete(ctx context.Context, id TagID) error
	
	// AssignToTestRun assigns tags to a test run
	AssignToTestRun(ctx context.Context, testRunID string, tagIDs []TagID) error
}