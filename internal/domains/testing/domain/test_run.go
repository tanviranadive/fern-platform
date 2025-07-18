package domain

import (
	"time"
)

// TestRun represents a test execution instance
type TestRun struct {
	ID           uint                   `json:"id"`     // Database ID
	RunID        string                 `json:"run_id"` // Unique run identifier
	ProjectID    string                 `json:"project_id"`
	Name         string                 `json:"name"`
	Branch       string                 `json:"branch"`
	GitBranch    string                 `json:"git_branch"`
	GitCommit    string                 `json:"git_commit"`
	Status       string                 `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	TotalTests   int                    `json:"total_tests"`
	PassedTests  int                    `json:"passed_tests"`
	FailedTests  int                    `json:"failed_tests"`
	SkippedTests int                    `json:"skipped_tests"`
	Duration     time.Duration          `json:"duration"`
	Environment  string                 `json:"environment"`
	Source       string                 `json:"source"`
	SessionID    string                 `json:"session_id"`
	Metadata     map[string]interface{} `json:"metadata"`
	SuiteRuns    []SuiteRun             `json:"suite_runs"`
}

// SuiteRun represents a test suite execution
type SuiteRun struct {
	ID           uint          `json:"id"`
	TestRunID    uint          `json:"test_run_id"`
	Name         string        `json:"name"`
	PackageName  string        `json:"package_name"`
	ClassName    string        `json:"class_name"`
	Status       string        `json:"status"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time"`
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	SkippedTests int           `json:"skipped_tests"`
	Duration     time.Duration `json:"duration"`
	SpecRuns     []*SpecRun    `json:"spec_runs"`
}

// SpecRun represents a single test specification execution
type SpecRun struct {
	ID             uint          `json:"id"`
	SuiteRunID     uint          `json:"suite_run_id"`
	Name           string        `json:"name"`
	ClassName      string        `json:"class_name"`
	Status         string        `json:"status"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        *time.Time    `json:"end_time"`
	Duration       time.Duration `json:"duration"`
	ErrorMessage   string        `json:"error_message"`
	FailureMessage string        `json:"failure_message"`
	StackTrace     string        `json:"stack_trace"`
	RetryCount     int           `json:"retry_count"`
	IsFlaky        bool          `json:"is_flaky"`
}

// TestRunSummary represents aggregated test run statistics
type TestRunSummary struct {
	TotalRuns      int           `json:"total_runs"`
	PassedRuns     int           `json:"passed_runs"`
	FailedRuns     int           `json:"failed_runs"`
	AverageRunTime time.Duration `json:"average_run_time"`
	SuccessRate    float64       `json:"success_rate"`
}
