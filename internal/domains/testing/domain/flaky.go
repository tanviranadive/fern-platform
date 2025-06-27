package domain

import (
	"errors"
	"time"
)

// FlakySeverity represents the severity level of test flakiness
type FlakySeverity string

const (
	FlakySeverityLow      FlakySeverity = "low"
	FlakySeverityMedium   FlakySeverity = "medium"
	FlakySeverityHigh     FlakySeverity = "high"
	FlakySeverityCritical FlakySeverity = "critical"
)

// FlakyStatus represents the status of a flaky test
type FlakyStatus string

const (
	FlakyStatusActive   FlakyStatus = "active"
	FlakyStatusResolved FlakyStatus = "resolved"
	FlakyStatusIgnored  FlakyStatus = "ignored"
)

// FlakyTest represents test flakiness analysis data
type FlakyTest struct {
	projectID        string
	testName         string
	suiteName        string
	flakeRate        float64
	totalExecutions  int
	flakyExecutions  int
	lastSeenAt       time.Time
	firstSeenAt      time.Time
	status           FlakyStatus
	severity         FlakySeverity
	lastErrorMessage string
}

// NewFlakyTest creates a new flaky test record
func NewFlakyTest(projectID, testName, suiteName string) (*FlakyTest, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}
	if testName == "" {
		return nil, errors.New("test name cannot be empty")
	}

	now := time.Now()
	return &FlakyTest{
		projectID:   projectID,
		testName:    testName,
		suiteName:   suiteName,
		status:      FlakyStatusActive,
		severity:    FlakySeverityLow,
		firstSeenAt: now,
		lastSeenAt:  now,
	}, nil
}

// RecordExecution records a test execution and updates flakiness metrics
func (ft *FlakyTest) RecordExecution(isFlaky bool, errorMessage string) {
	ft.totalExecutions++
	if isFlaky {
		ft.flakyExecutions++
		ft.lastErrorMessage = errorMessage
	}
	
	ft.lastSeenAt = time.Now()
	ft.updateFlakeRate()
	ft.updateSeverity()
}

// updateFlakeRate calculates the flake rate percentage
func (ft *FlakyTest) updateFlakeRate() {
	if ft.totalExecutions > 0 {
		ft.flakeRate = float64(ft.flakyExecutions) / float64(ft.totalExecutions) * 100
	}
}

// updateSeverity updates severity based on flake rate and frequency
func (ft *FlakyTest) updateSeverity() {
	// Severity based on flake rate and recent activity
	daysSinceFirst := time.Since(ft.firstSeenAt).Hours() / 24
	recentActivity := time.Since(ft.lastSeenAt).Hours() < 24

	switch {
	case ft.flakeRate > 50 && recentActivity:
		ft.severity = FlakySeverityCritical
	case ft.flakeRate > 30 || (ft.flakeRate > 20 && daysSinceFirst > 7):
		ft.severity = FlakySeverityHigh
	case ft.flakeRate > 10:
		ft.severity = FlakySeverityMedium
	default:
		ft.severity = FlakySeverityLow
	}
}

// Resolve marks the flaky test as resolved
func (ft *FlakyTest) Resolve() error {
	if ft.status != FlakyStatusActive {
		return errors.New("can only resolve active flaky tests")
	}
	ft.status = FlakyStatusResolved
	return nil
}

// Ignore marks the flaky test as ignored
func (ft *FlakyTest) Ignore() error {
	if ft.status == FlakyStatusResolved {
		return errors.New("cannot ignore resolved flaky tests")
	}
	ft.status = FlakyStatusIgnored
	return nil
}

// Reactivate marks the flaky test as active again
func (ft *FlakyTest) Reactivate() {
	ft.status = FlakyStatusActive
}

// Getters for read-only access
func (ft *FlakyTest) ProjectID() string        { return ft.projectID }
func (ft *FlakyTest) TestName() string         { return ft.testName }
func (ft *FlakyTest) SuiteName() string        { return ft.suiteName }
func (ft *FlakyTest) FlakeRate() float64       { return ft.flakeRate }
func (ft *FlakyTest) TotalExecutions() int     { return ft.totalExecutions }
func (ft *FlakyTest) FlakyExecutions() int     { return ft.flakyExecutions }
func (ft *FlakyTest) LastSeenAt() time.Time    { return ft.lastSeenAt }
func (ft *FlakyTest) FirstSeenAt() time.Time   { return ft.firstSeenAt }
func (ft *FlakyTest) Status() FlakyStatus      { return ft.status }
func (ft *FlakyTest) Severity() FlakySeverity  { return ft.severity }
func (ft *FlakyTest) LastErrorMessage() string { return ft.lastErrorMessage }