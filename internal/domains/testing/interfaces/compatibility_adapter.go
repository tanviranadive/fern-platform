package interfaces

import (
	"context"
	"fmt"

	"github.com/guidewire-oss/fern-platform/internal/domains/testing/application"
	"github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// CompatibilityAdapter adapts the new domain-driven design to the existing service interface
// This ensures backward compatibility while we migrate to DDD
type CompatibilityAdapter struct {
	recordTestRunHandler   *application.RecordTestRunHandler
	completeTestRunHandler *application.CompleteTestRunHandler
	testRunRepo           domain.TestRunRepository
}

// NewCompatibilityAdapter creates a new adapter
func NewCompatibilityAdapter(
	recordHandler *application.RecordTestRunHandler,
	completeHandler *application.CompleteTestRunHandler,
	repo domain.TestRunRepository,
) *CompatibilityAdapter {
	return &CompatibilityAdapter{
		recordTestRunHandler:   recordHandler,
		completeTestRunHandler: completeHandler,
		testRunRepo:           repo,
	}
}

// CreateTestRun adapts the existing service method to use the new domain
func (a *CompatibilityAdapter) CreateTestRun(input service.CreateTestRunInput) (*database.TestRun, error) {
	// Convert to domain command
	cmd := application.RecordTestRunCommand{
		RunID:       input.RunID,
		ProjectID:   input.ProjectID,
		Branch:      input.Branch,
		CommitSHA:   input.CommitSHA,
		Environment: input.Environment,
		Metadata:    input.Metadata,
	}
	
	// Execute use case
	snapshot, err := a.recordTestRunHandler.Handle(context.Background(), cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create test run: %w", err)
	}
	
	// Convert back to database model for compatibility
	return &database.TestRun{
		ProjectID:    snapshot.ProjectID,
		RunID:        string(snapshot.ID),
		Branch:       snapshot.Branch,
		CommitSHA:    snapshot.CommitSHA,
		Status:       string(snapshot.Status),
		StartTime:    snapshot.StartTime,
		EndTime:      snapshot.EndTime,
		TotalTests:   snapshot.TotalTests,
		PassedTests:  snapshot.PassedTests,
		FailedTests:  snapshot.FailedTests,
		SkippedTests: snapshot.SkippedTests,
		Duration:     snapshot.Duration,
		Environment:  snapshot.Environment,
		Metadata:     database.JSONMap(snapshot.Metadata),
	}, nil
}

// CompleteTestRun adapts the existing service method
func (a *CompatibilityAdapter) CompleteTestRun(runID string) error {
	cmd := application.CompleteTestRunCommand{
		RunID: runID,
	}
	
	return a.completeTestRunHandler.Handle(context.Background(), cmd)
}

// GetTestRun retrieves a test run by ID
func (a *CompatibilityAdapter) GetTestRun(runID string) (*database.TestRun, error) {
	testRun, err := a.testRunRepo.FindByID(context.Background(), domain.TestRunID(runID))
	if err != nil {
		return nil, err
	}
	if testRun == nil {
		return nil, nil
	}
	
	// Convert to database model
	snapshot := testRun.ToSnapshot()
	return &database.TestRun{
		ProjectID:    snapshot.ProjectID,
		RunID:        string(snapshot.ID),
		Branch:       snapshot.Branch,
		CommitSHA:    snapshot.CommitSHA,
		Status:       string(snapshot.Status),
		StartTime:    snapshot.StartTime,
		EndTime:      snapshot.EndTime,
		TotalTests:   snapshot.TotalTests,
		PassedTests:  snapshot.PassedTests,
		FailedTests:  snapshot.FailedTests,
		SkippedTests: snapshot.SkippedTests,
		Duration:     snapshot.Duration,
		Environment:  snapshot.Environment,
		Metadata:     database.JSONMap(snapshot.Metadata),
	}, nil
}