package domain

import (
	"context"
	"time"
)

// TestRunRepository defines the interface for test run persistence
type TestRunRepository interface {
	// Save persists a test run
	Save(ctx context.Context, testRun *TestRun) error
	
	// FindByID retrieves a test run by ID
	FindByID(ctx context.Context, id TestRunID) (*TestRun, error)
	
	// FindByProjectID retrieves test runs for a project
	FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*TestRun, error)
	
	// FindByTimeRange retrieves test runs within a time range
	FindByTimeRange(ctx context.Context, start, end time.Time) ([]*TestRun, error)
	
	// Update updates an existing test run
	Update(ctx context.Context, testRun *TestRun) error
	
	// Count returns the total number of test runs matching criteria
	Count(ctx context.Context, filter TestRunFilter) (int64, error)
}

// TestRunFilter represents filtering criteria for test runs
type TestRunFilter struct {
	ProjectID   string
	Branch      string
	Status      TestRunStatus
	Environment string
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int
	Offset      int
}

// FlakyTestRepository defines the interface for flaky test persistence
type FlakyTestRepository interface {
	// Save persists flaky test data
	Save(ctx context.Context, flakyTest *FlakyTest) error
	
	// FindByProject retrieves flaky tests for a project
	FindByProject(ctx context.Context, projectID string) ([]*FlakyTest, error)
	
	// FindByTestName retrieves flaky test history
	FindByTestName(ctx context.Context, projectID, testName string) (*FlakyTest, error)
	
	// Update updates flaky test statistics
	Update(ctx context.Context, flakyTest *FlakyTest) error
}