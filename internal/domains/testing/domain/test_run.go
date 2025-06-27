package domain

import (
	"errors"
	"time"
)

// TestRunID represents a unique identifier for a test run
type TestRunID string

// TestRunStatus represents the status of a test run
type TestRunStatus string

const (
	TestRunStatusRunning TestRunStatus = "running"
	TestRunStatusPassed  TestRunStatus = "passed"
	TestRunStatusFailed  TestRunStatus = "failed"
	TestRunStatusError   TestRunStatus = "error"
)

// TestRun represents a test execution instance in the domain
type TestRun struct {
	id           TestRunID
	projectID    string
	branch       string
	commitSHA    string
	status       TestRunStatus
	startTime    time.Time
	endTime      *time.Time
	totalTests   int
	passedTests  int
	failedTests  int
	skippedTests int
	environment  string
	metadata     map[string]interface{}
	suites       []SuiteRun
}

// NewTestRun creates a new test run
func NewTestRun(runID TestRunID, projectID string, branch string) (*TestRun, error) {
	if runID == "" {
		return nil, errors.New("test run ID cannot be empty")
	}
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	return &TestRun{
		id:        runID,
		projectID: projectID,
		branch:    branch,
		status:    TestRunStatusRunning,
		startTime: time.Now(),
		metadata:  make(map[string]interface{}),
		suites:    make([]SuiteRun, 0),
	}, nil
}

// ID returns the test run ID
func (tr *TestRun) ID() TestRunID {
	return tr.id
}

// ProjectID returns the project ID
func (tr *TestRun) ProjectID() string {
	return tr.projectID
}

// Status returns the current status
func (tr *TestRun) Status() TestRunStatus {
	return tr.status
}

// Complete marks the test run as completed
func (tr *TestRun) Complete() error {
	if tr.status != TestRunStatusRunning {
		return errors.New("can only complete a running test")
	}

	now := time.Now()
	tr.endTime = &now

	// Determine final status based on test results
	if tr.failedTests > 0 {
		tr.status = TestRunStatusFailed
	} else if tr.totalTests == tr.passedTests+tr.skippedTests {
		tr.status = TestRunStatusPassed
	} else {
		tr.status = TestRunStatusError
	}

	return nil
}

// AddSuite adds a suite run to the test run
func (tr *TestRun) AddSuite(suite SuiteRun) error {
	if tr.status != TestRunStatusRunning {
		return errors.New("cannot add suite to completed test run")
	}

	tr.suites = append(tr.suites, suite)
	
	// Update test counts
	tr.totalTests += suite.TotalSpecs()
	tr.passedTests += suite.PassedSpecs()
	tr.failedTests += suite.FailedSpecs()
	tr.skippedTests += suite.SkippedSpecs()

	return nil
}

// Duration returns the duration of the test run in milliseconds
func (tr *TestRun) Duration() int64 {
	if tr.endTime == nil {
		return int64(time.Since(tr.startTime).Milliseconds())
	}
	return int64(tr.endTime.Sub(tr.startTime).Milliseconds())
}

// SetMetadata sets a metadata key-value pair
func (tr *TestRun) SetMetadata(key string, value interface{}) {
	tr.metadata[key] = value
}

// GetMetadata returns metadata for a given key
func (tr *TestRun) GetMetadata(key string) (interface{}, bool) {
	val, exists := tr.metadata[key]
	return val, exists
}

// Validate ensures the test run is in a valid state
func (tr *TestRun) Validate() error {
	if tr.id == "" {
		return errors.New("test run must have an ID")
	}
	if tr.projectID == "" {
		return errors.New("test run must have a project ID")
	}
	if tr.totalTests < 0 || tr.passedTests < 0 || tr.failedTests < 0 || tr.skippedTests < 0 {
		return errors.New("test counts cannot be negative")
	}
	if tr.passedTests+tr.failedTests+tr.skippedTests > tr.totalTests {
		return errors.New("sum of test results exceeds total tests")
	}
	return nil
}

// ToSnapshot returns a read-only snapshot of the test run state
func (tr *TestRun) ToSnapshot() TestRunSnapshot {
	return TestRunSnapshot{
		ID:           tr.id,
		ProjectID:    tr.projectID,
		Branch:       tr.branch,
		CommitSHA:    tr.commitSHA,
		Status:       tr.status,
		StartTime:    tr.startTime,
		EndTime:      tr.endTime,
		TotalTests:   tr.totalTests,
		PassedTests:  tr.passedTests,
		FailedTests:  tr.failedTests,
		SkippedTests: tr.skippedTests,
		Environment:  tr.environment,
		Metadata:     tr.metadata,
		Duration:     tr.Duration(),
	}
}

// TestRunSnapshot is a read-only view of a test run
type TestRunSnapshot struct {
	ID           TestRunID
	ProjectID    string
	Branch       string
	CommitSHA    string
	Status       TestRunStatus
	StartTime    time.Time
	EndTime      *time.Time
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Environment  string
	Metadata     map[string]interface{}
	Duration     int64
}