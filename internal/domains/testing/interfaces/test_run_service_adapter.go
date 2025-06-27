package interfaces

import (
	"fmt"
	"time"

	"github.com/guidewire-oss/fern-platform/internal/domains/testing/application"
	"github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
)

// TestRunServiceAdapter adapts the new domain to the existing TestRunService interface
// This ensures complete backward compatibility with REST and GraphQL APIs
type TestRunServiceAdapter struct {
	recordTestRunHandler   *application.RecordTestRunHandler
	completeTestRunHandler *application.CompleteTestRunHandler
	testRunRepo           domain.TestRunRepository
	legacyRepo            *repository.TestRunRepository
	suiteRunRepo          *repository.SuiteRunRepository
	specRunRepo           *repository.SpecRunRepository
	logger                *logging.Logger
}

// NewTestRunServiceAdapter creates a new adapter that implements the existing service interface
func NewTestRunServiceAdapter(
	recordHandler *application.RecordTestRunHandler,
	completeHandler *application.CompleteTestRunHandler,
	domainRepo domain.TestRunRepository,
	legacyRepo *repository.TestRunRepository,
	suiteRunRepo *repository.SuiteRunRepository,
	specRunRepo *repository.SpecRunRepository,
	logger *logging.Logger,
) *TestRunServiceAdapter {
	return &TestRunServiceAdapter{
		recordTestRunHandler:   recordHandler,
		completeTestRunHandler: completeHandler,
		testRunRepo:           domainRepo,
		legacyRepo:            legacyRepo,
		suiteRunRepo:          suiteRunRepo,
		specRunRepo:           specRunRepo,
		logger:                logger,
	}
}

// CreateTestRun implements the existing service interface method
func (a *TestRunServiceAdapter) CreateTestRun(input service.CreateTestRunInput) (*database.TestRun, error) {
	a.logger.WithTestRun(input.RunID, input.ProjectID).Info("Creating test run")

	// Check if test run already exists using legacy repo for compatibility
	existing, err := a.legacyRepo.GetTestRunByRunID(input.RunID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("test run with run_id %s already exists", input.RunID)
	}

	// For now, we'll use the legacy repository directly to ensure data consistency
	// The domain handlers can be integrated gradually once domain repositories are fully implemented
	testRun := &database.TestRun{
		ProjectID:    input.ProjectID,
		RunID:        input.RunID,
		Branch:       input.Branch,
		CommitSHA:    input.CommitSHA,
		Status:       "running",
		StartTime:    time.Now(),
		Environment:  input.Environment,
		Metadata:     input.Metadata,
		TotalTests:   input.TotalTests,
		PassedTests:  input.PassedTests,
		FailedTests:  input.FailedTests,
		SkippedTests: input.SkippedTests,
		Duration:     input.Duration,
	}

	// Create in database
	if err := a.legacyRepo.CreateTestRun(testRun); err != nil {
		return nil, fmt.Errorf("failed to create test run: %w", err)
	}

	created := testRun

	// Update with provided data that domain doesn't handle yet
	if input.StartTime != nil {
		created.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		created.EndTime = input.EndTime
		created.Status = "completed"
	}
	if input.Duration > 0 {
		created.Duration = input.Duration
	}
	created.TotalTests = input.TotalTests
	created.PassedTests = input.PassedTests
	created.FailedTests = input.FailedTests
	created.SkippedTests = input.SkippedTests

	// Update in database
	if err := a.legacyRepo.UpdateTestRun(created); err != nil {
		return nil, fmt.Errorf("failed to update test run details: %w", err)
	}

	a.logger.WithTestRun(input.RunID, input.ProjectID).
		WithField("test_run_id", created.ID).
		Info("Test run created successfully")

	return created, nil
}

// CreateTestRunWithSuites implements the existing service interface method
func (a *TestRunServiceAdapter) CreateTestRunWithSuites(input service.CreateTestRunInput) (*database.TestRun, error) {
	a.logger.WithTestRun(input.RunID, input.ProjectID).Info("Creating test run with suites and specs")

	// First create the test run
	testRun, err := a.CreateTestRun(input)
	if err != nil {
		return nil, err
	}

	// Create suite runs and spec runs using legacy repositories
	for _, suiteInput := range input.SuiteRuns {
		// Parse suite times
		var suiteStartTime, suiteEndTime *time.Time
		if suiteInput.StartTime != "" {
			if t, err := time.Parse(time.RFC3339, suiteInput.StartTime); err == nil {
				suiteStartTime = &t
			}
		}
		if suiteInput.EndTime != "" {
			if t, err := time.Parse(time.RFC3339, suiteInput.EndTime); err == nil {
				suiteEndTime = &t
			}
		}

		// Create suite run
		suiteRun := &database.SuiteRun{
			TestRunID:  testRun.ID,
			SuiteName:  suiteInput.SuiteName,
			Status:     "completed",
			StartTime:  time.Now(),
		}
		
		if suiteStartTime != nil {
			suiteRun.StartTime = *suiteStartTime
		}
		if suiteEndTime != nil {
			suiteRun.EndTime = suiteEndTime
		}

		// Save suite run
		if err := a.suiteRunRepo.CreateSuiteRun(suiteRun); err != nil {
			a.logger.WithError(err).Error("Failed to create suite run")
			continue
		}

		// Create spec runs
		totalSpecs := len(suiteInput.SpecRuns)
		passedSpecs := 0
		failedSpecs := 0
		skippedSpecs := 0

		for _, specInput := range suiteInput.SpecRuns {
			// Parse spec times
			var specStartTime, specEndTime *time.Time
			if specInput.StartTime != "" {
				if t, err := time.Parse(time.RFC3339, specInput.StartTime); err == nil {
					specStartTime = &t
				}
			}
			if specInput.EndTime != "" {
				if t, err := time.Parse(time.RFC3339, specInput.EndTime); err == nil {
					specEndTime = &t
				}
			}

			specRun := &database.SpecRun{
				SuiteRunID: suiteRun.ID,
				SpecName:   specInput.SpecDescription,
				Status:     specInput.Status,
				StartTime:  time.Now(),
			}

			if specStartTime != nil {
				specRun.StartTime = *specStartTime
			}
			if specEndTime != nil {
				specRun.EndTime = specEndTime
			}

			// Set error message if failed
			if specInput.Status == "failed" && specInput.Message != "" {
				specRun.ErrorMessage = specInput.Message
			}

			// Count status
			switch specInput.Status {
			case "passed":
				passedSpecs++
			case "failed":
				failedSpecs++
			case "skipped":
				skippedSpecs++
			}

			// Save spec run
			if err := a.specRunRepo.CreateSpecRun(specRun); err != nil {
				a.logger.WithError(err).Error("Failed to create spec run")
				continue
			}
		}

		// Update suite run stats
		suiteRun.TotalSpecs = totalSpecs
		suiteRun.PassedSpecs = passedSpecs
		suiteRun.FailedSpecs = failedSpecs
		suiteRun.SkippedSpecs = skippedSpecs
		if suiteEndTime != nil && suiteStartTime != nil {
			suiteRun.Duration = suiteEndTime.Sub(*suiteStartTime).Milliseconds()
		}

		if err := a.suiteRunRepo.UpdateSuiteRun(suiteRun); err != nil {
			a.logger.WithError(err).Error("Failed to update suite run stats")
		}
	}

	// Update test run status based on results
	if input.EndTime != nil {
		testRun.Status = "completed"
		testRun.EndTime = input.EndTime
		if err := a.legacyRepo.UpdateTestRun(testRun); err != nil {
			a.logger.WithError(err).Error("Failed to update test run status")
		}
	}

	return testRun, nil
}

// GetTestRun retrieves a test run by ID
func (a *TestRunServiceAdapter) GetTestRun(id uint) (*database.TestRun, error) {
	return a.legacyRepo.GetTestRunByID(id)
}

// GetTestRunByRunID retrieves a test run by run ID
func (a *TestRunServiceAdapter) GetTestRunByRunID(runID string) (*database.TestRun, error) {
	return a.legacyRepo.GetTestRunByRunID(runID)
}

// UpdateTestRun updates a test run
func (a *TestRunServiceAdapter) UpdateTestRun(testRun *database.TestRun) error {
	return a.legacyRepo.UpdateTestRun(testRun)
}

// CompleteTestRun marks a test run as completed
func (a *TestRunServiceAdapter) CompleteTestRun(runID string, status string, endTime time.Time) error {
	// Use legacy repo directly for consistency
	testRun, err := a.legacyRepo.GetTestRunByRunID(runID)
	if err != nil {
		return fmt.Errorf("failed to find test run: %w", err)
	}

	testRun.Status = status
	testRun.EndTime = &endTime
	testRun.Duration = endTime.Sub(testRun.StartTime).Milliseconds()

	return a.legacyRepo.UpdateTestRun(testRun)
}

// ListTestRuns lists test runs with filtering
func (a *TestRunServiceAdapter) ListTestRuns(filter repository.ListTestRunsFilter) ([]*database.TestRun, int64, error) {
	return a.legacyRepo.ListTestRuns(filter)
}

// GetTestRunsForProject gets test runs for a specific project
func (a *TestRunServiceAdapter) GetTestRunsForProject(projectID string, limit int) ([]*database.TestRun, error) {
	filter := repository.ListTestRunsFilter{
		ProjectID: projectID,
		Limit:     limit,
		OrderBy:   "start_time",
		Order:     "DESC",
	}
	
	runs, _, err := a.legacyRepo.ListTestRuns(filter)
	return runs, err
}

// GetRecentTestRuns gets recent test runs for a project
func (a *TestRunServiceAdapter) GetRecentTestRuns(projectID string, limit int) ([]*database.TestRun, error) {
	filter := repository.ListTestRunsFilter{
		ProjectID: projectID,
		Limit:     limit,
		OrderBy:   "start_time",
		Order:     "DESC",
	}
	
	runs, _, err := a.legacyRepo.ListTestRuns(filter)
	return runs, err
}

// GetTestRunsInTimeRange gets test runs within a time range
func (a *TestRunServiceAdapter) GetTestRunsInTimeRange(projectID string, startTime, endTime time.Time) ([]*database.TestRun, error) {
	filter := repository.ListTestRunsFilter{
		ProjectID: projectID,
		StartTime: &startTime,
		EndTime:   &endTime,
		OrderBy:   "start_time",
		Order:     "DESC",
	}
	
	runs, _, err := a.legacyRepo.ListTestRuns(filter)
	return runs, err
}

// UpdateTestRunStatus updates the status of a test run
func (a *TestRunServiceAdapter) UpdateTestRunStatus(runID, status string, endTime *time.Time) error {
	testRun, err := a.legacyRepo.GetTestRunByRunID(runID)
	if err != nil {
		return fmt.Errorf("test run not found: %w", err)
	}

	testRun.Status = status
	if endTime != nil {
		testRun.EndTime = endTime
		testRun.Duration = endTime.Sub(testRun.StartTime).Milliseconds()
	}

	return a.legacyRepo.UpdateTestRun(testRun)
}

// ListTestRunsWithProjects retrieves test runs with project names
func (a *TestRunServiceAdapter) ListTestRunsWithProjects(filter repository.ListTestRunsFilter) ([]*repository.TestRunWithProject, int64, error) {
	return a.legacyRepo.ListTestRunsWithProjects(filter)
}

// GetTestRunStats retrieves test run statistics
func (a *TestRunServiceAdapter) GetTestRunStats(projectID string, days int) (*repository.TestRunStats, error) {
	return a.legacyRepo.GetTestRunStats(projectID, days)
}

// DeleteTestRun deletes a test run
func (a *TestRunServiceAdapter) DeleteTestRun(id uint) error {
	return a.legacyRepo.DeleteTestRun(id)
}

// AssignTagsToTestRun assigns tags to a test run
func (a *TestRunServiceAdapter) AssignTagsToTestRun(testRunID uint, tagIDs []uint) error {
	return a.legacyRepo.AssignTagsToTestRun(testRunID, tagIDs)
}

// AddSuiteRun adds a suite run to a test run
func (a *TestRunServiceAdapter) AddSuiteRun(testRunID uint, suiteName string, startTime time.Time) (*database.SuiteRun, error) {
	suiteRun := &database.SuiteRun{
		TestRunID: testRunID,
		SuiteName: suiteName,
		Status:    "running",
		StartTime: startTime,
	}
	
	if err := a.suiteRunRepo.CreateSuiteRun(suiteRun); err != nil {
		return nil, fmt.Errorf("failed to create suite run: %w", err)
	}
	
	return suiteRun, nil
}

// UpdateSuiteRunStatus updates the status of a suite run
func (a *TestRunServiceAdapter) UpdateSuiteRunStatus(suiteRunID uint, status string, endTime *time.Time) error {
	suiteRun, err := a.suiteRunRepo.GetSuiteRunByID(suiteRunID)
	if err != nil {
		return fmt.Errorf("suite run not found: %w", err)
	}

	suiteRun.Status = status
	if endTime != nil {
		suiteRun.EndTime = endTime
		suiteRun.Duration = endTime.Sub(suiteRun.StartTime).Milliseconds()
	}

	return a.suiteRunRepo.UpdateSuiteRun(suiteRun)
}

// UpdateSuiteRunStats updates the statistics for a suite run
func (a *TestRunServiceAdapter) UpdateSuiteRunStats(suiteRunID uint) error {
	// Get all spec runs for this suite
	specRuns, err := a.specRunRepo.ListSpecRunsBySuite(suiteRunID)
	if err != nil {
		return fmt.Errorf("failed to get spec runs: %w", err)
	}

	// Calculate stats
	var totalSpecs, passedSpecs, failedSpecs, skippedSpecs int
	for _, spec := range specRuns {
		totalSpecs++
		switch spec.Status {
		case "passed":
			passedSpecs++
		case "failed":
			failedSpecs++
		case "skipped":
			skippedSpecs++
		}
	}

	// Update suite run
	suiteRun, err := a.suiteRunRepo.GetSuiteRunByID(suiteRunID)
	if err != nil {
		return fmt.Errorf("failed to get suite run: %w", err)
	}

	suiteRun.TotalSpecs = totalSpecs
	suiteRun.PassedSpecs = passedSpecs
	suiteRun.FailedSpecs = failedSpecs
	suiteRun.SkippedSpecs = skippedSpecs

	return a.suiteRunRepo.UpdateSuiteRun(suiteRun)
}

// AddSpecRun adds a spec run to a suite run
func (a *TestRunServiceAdapter) AddSpecRun(suiteRunID uint, specName, status string, startTime, endTime time.Time, errorMessage, stackTrace string) (*database.SpecRun, error) {
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
	
	if err := a.specRunRepo.CreateSpecRun(specRun); err != nil {
		return nil, fmt.Errorf("failed to create spec run: %w", err)
	}
	
	return specRun, nil
}

// RecalculateTestRunStats recalculates and saves test run statistics
func (a *TestRunServiceAdapter) RecalculateTestRunStats(testRunID uint) error {
	testRun, err := a.legacyRepo.GetTestRunByID(testRunID)
	if err != nil {
		return fmt.Errorf("test run not found: %w", err)
	}

	// Get all suite runs
	suiteRuns, err := a.suiteRunRepo.ListSuiteRunsByTestRun(testRunID)
	if err != nil {
		return fmt.Errorf("failed to get suite runs: %w", err)
	}

	// Calculate totals
	var totalTests, passedTests, failedTests, skippedTests int
	for _, suite := range suiteRuns {
		totalTests += suite.TotalSpecs
		passedTests += suite.PassedSpecs
		failedTests += suite.FailedSpecs
		skippedTests += suite.SkippedSpecs
	}

	// Update test run
	testRun.TotalTests = totalTests
	testRun.PassedTests = passedTests
	testRun.FailedTests = failedTests
	testRun.SkippedTests = skippedTests

	// Update status based on results
	if failedTests > 0 {
		testRun.Status = "failed"
	} else if passedTests == totalTests {
		testRun.Status = "passed"
	}

	return a.legacyRepo.UpdateTestRun(testRun)
}