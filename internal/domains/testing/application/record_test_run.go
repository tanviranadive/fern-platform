package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
)

// RecordTestRunCommand represents the command to record a test run
type RecordTestRunCommand struct {
	RunID        string                 `json:"run_id"`
	ProjectID    string                 `json:"project_id"`
	Branch       string                 `json:"branch"`
	CommitSHA    string                 `json:"commit_sha"`
	Environment  string                 `json:"environment"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// RecordTestRunHandler handles the record test run use case
type RecordTestRunHandler struct {
	testRunRepo domain.TestRunRepository
}

// NewRecordTestRunHandler creates a new handler
func NewRecordTestRunHandler(testRunRepo domain.TestRunRepository) *RecordTestRunHandler {
	return &RecordTestRunHandler{
		testRunRepo: testRunRepo,
	}
}

// Handle executes the use case
func (h *RecordTestRunHandler) Handle(ctx context.Context, cmd RecordTestRunCommand) (*domain.TestRun, error) {
	// Validate command
	if err := h.validateCommand(cmd); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Create new test run
	testRun := &domain.TestRun{
		RunID:       cmd.RunID,
		ProjectID:   cmd.ProjectID,
		Branch:      cmd.Branch,
		GitCommit:   cmd.CommitSHA,
		Environment: cmd.Environment,
		Metadata:    cmd.Metadata,
		Status:      "running",
		StartTime:   time.Now(),
	}

	// Save to repository
	if err := h.testRunRepo.Create(ctx, testRun); err != nil {
		return nil, fmt.Errorf("failed to save test run: %w", err)
	}

	return testRun, nil
}

func (h *RecordTestRunHandler) validateCommand(cmd RecordTestRunCommand) error {
	if cmd.RunID == "" {
		return errors.New("run ID is required")
	}
	if cmd.ProjectID == "" {
		return errors.New("project ID is required")
	}
	if cmd.Branch == "" {
		return errors.New("branch is required")
	}
	return nil
}