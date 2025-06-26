// Package service provides business logic for the fern-reporter service
package service

import (
	"fmt"
	"time"

	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
)

// TestRunService handles test run business logic
type TestRunService struct {
	testRunRepo   *repository.TestRunRepository
	suiteRunRepo  *repository.SuiteRunRepository
	specRunRepo   *repository.SpecRunRepository
	logger        *logging.Logger
}

// NewTestRunService creates a new test run service
func NewTestRunService(
	testRunRepo *repository.TestRunRepository,
	suiteRunRepo *repository.SuiteRunRepository,
	specRunRepo *repository.SpecRunRepository,
	logger *logging.Logger,
) *TestRunService {
	return &TestRunService{
		testRunRepo:  testRunRepo,
		suiteRunRepo: suiteRunRepo,
		specRunRepo:  specRunRepo,
		logger:       logger,
	}
}

// CreateTestRunInput represents input for creating a test run
type CreateTestRunInput struct {
	ProjectID     string            `json:"project_id" binding:"required"`
	RunID         string            `json:"run_id" binding:"required"`
	Branch        string            `json:"branch"`
	CommitSHA     string            `json:"commit_sha"`
	Environment   string            `json:"environment"`
	Metadata      database.JSONMap `json:"metadata,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	StartTime     *time.Time        `json:"start_time,omitempty"`
	EndTime       *time.Time        `json:"end_time,omitempty"`
	Duration      int64             `json:"duration,omitempty"`
	TotalTests    int               `json:"total_tests,omitempty"`
	PassedTests   int               `json:"passed_tests,omitempty"`
	FailedTests   int               `json:"failed_tests,omitempty"`
	SkippedTests  int               `json:"skipped_tests,omitempty"`
	SuiteRuns     []CreateSuiteRunInput `json:"suite_runs,omitempty"`
}

// CreateSuiteRunInput represents input for creating a suite run
type CreateSuiteRunInput struct {
	SuiteName string              `json:"suite_name" binding:"required"`
	StartTime string              `json:"start_time"`
	EndTime   string              `json:"end_time"`
	SpecRuns  []CreateSpecRunInput `json:"spec_runs,omitempty"`
}

// CreateSpecRunInput represents input for creating a spec run
type CreateSpecRunInput struct {
	SpecDescription string `json:"spec_description" binding:"required"`
	Status          string `json:"status" binding:"required"`
	Message         string `json:"message"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Tags            []struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
}

// CreateTestRun creates a new test run
func (s *TestRunService) CreateTestRun(input CreateTestRunInput) (*database.TestRun, error) {
	s.logger.WithTestRun(input.RunID, input.ProjectID).Info("Creating test run")

	// Check if test run with this run_id already exists
	existing, err := s.testRunRepo.GetTestRunByRunID(input.RunID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("test run with run_id %s already exists", input.RunID)
	}

	// Set start time
	startTime := time.Now().UTC()
	if input.StartTime != nil {
		startTime = *input.StartTime
	}

	testRun := &database.TestRun{
		ProjectID:     input.ProjectID,
		RunID:         input.RunID,
		Branch:        input.Branch,
		CommitSHA:     input.CommitSHA,
		Status:        "running",
		StartTime:     startTime,
		Environment:   input.Environment,
		TotalTests:    input.TotalTests,
		PassedTests:   input.PassedTests,
		FailedTests:   input.FailedTests,
		SkippedTests:  input.SkippedTests,
	}

	// Set end time if provided
	if input.EndTime != nil {
		testRun.EndTime = input.EndTime
		testRun.Status = "completed" // Mark as completed if end time is provided
	}

	// Set duration if provided
	if input.Duration > 0 {
		testRun.Duration = input.Duration
	} else if input.EndTime != nil {
		// Calculate duration from start and end time
		testRun.Duration = input.EndTime.Sub(startTime).Milliseconds()
	}

	// Handle metadata
	if len(input.Metadata) > 0 {
		testRun.Metadata = input.Metadata
	}

	if err := s.testRunRepo.CreateTestRun(testRun); err != nil {
		s.logger.WithTestRun(input.RunID, input.ProjectID).WithError(err).Error("Failed to create test run")
		return nil, fmt.Errorf("failed to create test run: %w", err)
	}

	s.logger.WithTestRun(input.RunID, input.ProjectID).
		WithField("test_run_id", testRun.ID).
		Info("Test run created successfully")

	return testRun, nil
}

// CreateTestRunWithSuites creates a complete test run with suite runs and spec runs
func (s *TestRunService) CreateTestRunWithSuites(input CreateTestRunInput) (*database.TestRun, error) {
	s.logger.WithTestRun(input.RunID, input.ProjectID).Info("Creating test run with suites and specs")

	// First create the test run
	testRun, err := s.CreateTestRun(input)
	if err != nil {
		return nil, err
	}

	// Create suite runs and spec runs
	for _, suiteInput := range input.SuiteRuns {
		// Parse suite start time
		var suiteStartTime time.Time
		if suiteInput.StartTime != "" {
			if parsedTime, err := time.Parse(time.RFC3339, suiteInput.StartTime); err == nil {
				suiteStartTime = parsedTime
			} else {
				suiteStartTime = testRun.StartTime
			}
		} else {
			suiteStartTime = testRun.StartTime
		}

		// Create suite run
		suiteRun, err := s.AddSuiteRun(testRun.ID, suiteInput.SuiteName, suiteStartTime)
		if err != nil {
			s.logger.WithError(err).Error("Failed to create suite run")
			continue
		}

		// Parse suite end time and update if provided
		if suiteInput.EndTime != "" {
			if parsedTime, err := time.Parse(time.RFC3339, suiteInput.EndTime); err == nil {
				if err := s.UpdateSuiteRunStatus(suiteRun.ID, "completed", &parsedTime); err != nil {
					s.logger.WithError(err).Warn("Failed to update suite run end time")
				}
			}
		}

		// Create spec runs
		for _, specInput := range suiteInput.SpecRuns {
			// Parse spec start and end times - these should always be provided
			specStartTime, err := time.Parse(time.RFC3339, specInput.StartTime)
			if err != nil {
				s.logger.WithField("spec", specInput.SpecDescription).WithField("start_time", specInput.StartTime).WithError(err).Error("Failed to parse spec start time")
				continue // Skip this spec if we can't parse the time
			}

			specEndTime, err := time.Parse(time.RFC3339, specInput.EndTime)
			if err != nil {
				s.logger.WithField("spec", specInput.SpecDescription).WithField("end_time", specInput.EndTime).WithError(err).Error("Failed to parse spec end time")
				continue // Skip this spec if we can't parse the time
			}

			// Create spec run with improved naming
			specName := specInput.SpecDescription
			if specName == "" || specName == "[Test Setup/Teardown]" {
				// Generate a more descriptive name for empty specs
				specName = fmt.Sprintf("[%s Setup/Teardown]", suiteInput.SuiteName)
			}
			
			_, err = s.AddSpecRun(
				suiteRun.ID,
				specName,
				specInput.Status,
				specStartTime,
				specEndTime,
				specInput.Message,
				"", // No stack trace in input
			)
			if err != nil {
				s.logger.WithError(err).Error("Failed to create spec run")
			}
		}
		
		// Update suite run statistics after all specs are created
		if err := s.suiteRunRepo.UpdateSuiteRunStats(suiteRun.ID); err != nil {
			s.logger.WithError(err).Warn("Failed to update suite run stats")
		}
	}

	// Recalculate test run statistics from the actual suite runs
	if err := s.calculateTestRunStats(testRun); err != nil {
		s.logger.WithError(err).Warn("Failed to calculate test run stats")
	}
	
	// Save the updated test run with recalculated stats
	if err := s.testRunRepo.UpdateTestRun(testRun); err != nil {
		s.logger.WithError(err).Warn("Failed to save updated test run stats")
	}

	s.logger.WithTestRun(input.RunID, input.ProjectID).
		WithField("test_run_id", testRun.ID).
		WithField("suite_count", len(input.SuiteRuns)).
		Info("Test run with suites created successfully")

	return testRun, nil
}

// UpdateTestRunStatus updates the status of a test run
func (s *TestRunService) UpdateTestRunStatus(runID, status string, endTime *time.Time) error {
	testRun, err := s.testRunRepo.GetTestRunByRunID(runID)
	if err != nil {
		return fmt.Errorf("test run not found: %w", err)
	}

	testRun.Status = status
	if endTime != nil {
		testRun.EndTime = endTime
		testRun.Duration = endTime.Sub(testRun.StartTime).Milliseconds()
	}

	// Recalculate statistics from suite runs
	if err := s.calculateTestRunStats(testRun); err != nil {
		s.logger.WithTestRun(runID, testRun.ProjectID).WithError(err).Warn("Failed to calculate test run stats")
	}

	if err := s.testRunRepo.UpdateTestRun(testRun); err != nil {
		return fmt.Errorf("failed to update test run status: %w", err)
	}

	s.logger.WithTestRun(runID, testRun.ProjectID).
		WithField("status", status).
		Info("Test run status updated")

	return nil
}

// AddSuiteRun adds a suite run to a test run
func (s *TestRunService) AddSuiteRun(testRunID uint, suiteName string, startTime time.Time) (*database.SuiteRun, error) {
	suiteRun := &database.SuiteRun{
		TestRunID: testRunID,
		SuiteName: suiteName,
		Status:    "running",
		StartTime: startTime,
	}

	if err := s.suiteRunRepo.CreateSuiteRun(suiteRun); err != nil {
		return nil, fmt.Errorf("failed to create suite run: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"test_run_id":  testRunID,
		"suite_run_id": suiteRun.ID,
		"suite_name":   suiteName,
	}).Info("Suite run created")

	return suiteRun, nil
}

// UpdateSuiteRunStatus updates the status of a suite run
func (s *TestRunService) UpdateSuiteRunStatus(suiteRunID uint, status string, endTime *time.Time) error {
	suiteRun, err := s.suiteRunRepo.GetSuiteRunByID(suiteRunID)
	if err != nil {
		return fmt.Errorf("suite run not found: %w", err)
	}

	suiteRun.Status = status
	if endTime != nil {
		suiteRun.EndTime = endTime
		suiteRun.Duration = endTime.Sub(suiteRun.StartTime).Milliseconds()
	}

	// Update suite run statistics
	if err := s.suiteRunRepo.UpdateSuiteRunStats(suiteRunID); err != nil {
		s.logger.WithField("suite_run_id", suiteRunID).WithError(err).Warn("Failed to update suite run stats")
	}

	if err := s.suiteRunRepo.UpdateSuiteRun(suiteRun); err != nil {
		return fmt.Errorf("failed to update suite run: %w", err)
	}

	return nil
}

// AddSpecRun adds a spec run to a suite run
func (s *TestRunService) AddSpecRun(suiteRunID uint, specName, status string, startTime, endTime time.Time, errorMessage, stackTrace string) (*database.SpecRun, error) {
	// Handle empty or generic spec names
	if specName == "" {
		specName = "[Unnamed Spec]"
	}
	
	specRun := &database.SpecRun{
		SuiteRunID:   suiteRunID,
		SpecName:     specName,
		Status:       status,
		StartTime:    startTime,
		EndTime:      &endTime,
		Duration:     endTime.Sub(startTime).Milliseconds(),
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
	}

	if err := s.specRunRepo.CreateSpecRun(specRun); err != nil {
		return nil, fmt.Errorf("failed to create spec run: %w", err)
	}

	// Update flaky test tracking if the spec failed or was retried
	// TODO: Re-enable flaky test tracking once FlakyTestRepository is properly imported
	/*
	suiteRun, err := s.suiteRunRepo.GetSuiteRunByID(suiteRunID)
	if err == nil {
		testRun, err := s.testRunRepo.GetTestRunByID(suiteRun.TestRunID)
		if err == nil {
			isFlaky := status == "failed" && specRun.RetryCount > 0
			if err := s.flakyTestRepo.UpdateFlakyTestExecution(
				testRun.ProjectID, specName, suiteRun.SuiteName, isFlaky, errorMessage,
			); err != nil {
				s.logger.WithError(err).Warn("Failed to update flaky test tracking")
			}
		}
	}
	*/

	s.logger.WithFields(map[string]interface{}{
		"suite_run_id": suiteRunID,
		"spec_run_id":  specRun.ID,
		"spec_name":    specName,
		"status":       status,
	}).Info("Spec run created")

	return specRun, nil
}

// GetTestRun retrieves a test run by ID
func (s *TestRunService) GetTestRun(id uint) (*database.TestRun, error) {
	return s.testRunRepo.GetTestRunByID(id)
}

// GetTestRunByRunID retrieves a test run by run ID
func (s *TestRunService) GetTestRunByRunID(runID string) (*database.TestRun, error) {
	return s.testRunRepo.GetTestRunByRunID(runID)
}

// ListTestRuns retrieves test runs with filtering
func (s *TestRunService) ListTestRuns(filter repository.ListTestRunsFilter) ([]*database.TestRun, int64, error) {
	return s.testRunRepo.ListTestRuns(filter)
}

// ListTestRunsWithProjects retrieves test runs with project names
func (s *TestRunService) ListTestRunsWithProjects(filter repository.ListTestRunsFilter) ([]*repository.TestRunWithProject, int64, error) {
	return s.testRunRepo.ListTestRunsWithProjects(filter)
}

// GetTestRunStats retrieves test run statistics
func (s *TestRunService) GetTestRunStats(projectID string, days int) (*repository.TestRunStats, error) {
	return s.testRunRepo.GetTestRunStats(projectID, days)
}

// DeleteTestRun deletes a test run
func (s *TestRunService) DeleteTestRun(id uint) error {
	testRun, err := s.testRunRepo.GetTestRunByID(id)
	if err != nil {
		return fmt.Errorf("test run not found: %w", err)
	}

	if err := s.testRunRepo.DeleteTestRun(id); err != nil {
		return fmt.Errorf("failed to delete test run: %w", err)
	}

	s.logger.WithTestRun(testRun.RunID, testRun.ProjectID).Info("Test run deleted")
	return nil
}

// calculateTestRunStats recalculates test run statistics from suite runs
func (s *TestRunService) calculateTestRunStats(testRun *database.TestRun) error {
	suiteRuns, err := s.suiteRunRepo.ListSuiteRunsByTestRun(testRun.ID)
	if err != nil {
		return err
	}

	var totalTests, passedTests, failedTests, skippedTests int
	for _, suite := range suiteRuns {
		totalTests += suite.TotalSpecs
		passedTests += suite.PassedSpecs
		failedTests += suite.FailedSpecs
		skippedTests += suite.SkippedSpecs
	}

	testRun.TotalTests = totalTests
	testRun.PassedTests = passedTests
	testRun.FailedTests = failedTests
	testRun.SkippedTests = skippedTests

	return nil
}

// GetRecentTestRuns returns recent test runs
func (s *TestRunService) GetRecentTestRuns(projectID string, limit int) ([]*database.TestRun, error) {
	return s.testRunRepo.GetRecentTestRuns(projectID, limit)
}

// AssignTagsToTestRun assigns tags to a test run
func (s *TestRunService) AssignTagsToTestRun(testRunID uint, tagIDs []uint) error {
	return s.testRunRepo.AssignTagsToTestRun(testRunID, tagIDs)
}

// UpdateSuiteRunStats updates the statistics for a suite run
func (s *TestRunService) UpdateSuiteRunStats(suiteRunID uint) error {
	return s.suiteRunRepo.UpdateSuiteRunStats(suiteRunID)
}

// RecalculateTestRunStats recalculates and saves test run statistics
func (s *TestRunService) RecalculateTestRunStats(testRunID uint) error {
	testRun, err := s.testRunRepo.GetTestRunByID(testRunID)
	if err != nil {
		return fmt.Errorf("test run not found: %w", err)
	}
	
	if err := s.calculateTestRunStats(testRun); err != nil {
		return fmt.Errorf("failed to calculate test run stats: %w", err)
	}
	
	if err := s.testRunRepo.UpdateTestRun(testRun); err != nil {
		return fmt.Errorf("failed to save test run stats: %w", err)
	}
	
	return nil
}