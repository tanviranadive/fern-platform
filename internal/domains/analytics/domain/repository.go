package domain

import (
	"context"
	"time"
)

// FlakyDetectionRepository defines the interface for flaky test persistence
type FlakyDetectionRepository interface {
	// Save or update a flaky test record
	SaveFlakyTest(ctx context.Context, flaky *FlakyTest) error

	// Get a flaky test by ID
	GetFlakyTest(ctx context.Context, testID string) (*FlakyTest, error)

	// Find flaky tests for a project
	FindFlakyTestsByProject(ctx context.Context, projectID string, status FlakyTestStatus) ([]*FlakyTest, error)

	// Update flaky test status
	UpdateFlakyTestStatus(ctx context.Context, testID string, status FlakyTestStatus) error

	// Record a test run analysis
	SaveTestRunAnalysis(ctx context.Context, analysis *TestRunAnalysis) error

	// Get test run history for flaky detection
	GetTestRunHistory(ctx context.Context, projectID string, testName string, since time.Time) ([]TestExecutionResult, error)

	// Get unique test names for a project
	GetUniqueTestNames(ctx context.Context, projectID string, since time.Time) ([]string, error)
}

// TestExecutionResult represents a single test execution result
type TestExecutionResult struct {
	TestRunID   string
	TestName    string
	SuiteName   string
	Status      string // passed, failed, skipped
	Duration    time.Duration
	ExecutedAt  time.Time
	Error       string
	Environment map[string]string
}
