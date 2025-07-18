// Package service provides input/output types for service operations
package service

import (
	"time"

	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// CreateTestRunInput represents input for creating a test run
type CreateTestRunInput struct {
	ProjectID    string           `json:"project_id" binding:"required"`
	RunID        string           `json:"run_id" binding:"required"`
	Branch       string           `json:"branch"`
	CommitSHA    string           `json:"commit_sha"`
	Environment  string           `json:"environment"`
	Metadata     database.JSONMap `json:"metadata,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
	StartTime    *time.Time       `json:"start_time,omitempty"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	Duration     int64            `json:"duration,omitempty"`
	TotalTests   int              `json:"total_tests,omitempty"`
	PassedTests  int              `json:"passed_tests,omitempty"`
	FailedTests  int              `json:"failed_tests,omitempty"`
	SkippedTests int              `json:"skipped_tests,omitempty"`
	SuiteRuns    []SuiteRunInput  `json:"suite_runs,omitempty"`
}

// SuiteRunInput represents input for creating a suite run
type SuiteRunInput struct {
	SuiteName string         `json:"suite_name" binding:"required"`
	StartTime string         `json:"start_time,omitempty"`
	EndTime   string         `json:"end_time,omitempty"`
	SpecRuns  []SpecRunInput `json:"spec_runs,omitempty"`
}

// SpecRunInput represents input for creating a spec run
type SpecRunInput struct {
	SpecDescription string `json:"spec_description" binding:"required"`
	Status          string `json:"status" binding:"required,oneof=passed failed skipped pending"`
	StartTime       string `json:"start_time,omitempty"`
	EndTime         string `json:"end_time,omitempty"`
	Message         string `json:"message,omitempty"`
	StackTrace      string `json:"stack_trace,omitempty"`
}

// CreateProjectInput represents input for creating a project
type CreateProjectInput struct {
	ProjectID     string           `json:"project_id,omitempty"` // Deprecated: ProjectID is now auto-generated
	Name          string           `json:"name" binding:"required"`
	Description   string           `json:"description,omitempty"`
	Repository    string           `json:"repository,omitempty"`
	DefaultBranch string           `json:"default_branch,omitempty"`
	Team          string           `json:"team,omitempty"`
	Settings      database.JSONMap `json:"settings,omitempty"`
}

// UpdateProjectInput represents input for updating a project
type UpdateProjectInput struct {
	Name          string           `json:"name,omitempty"`
	Description   string           `json:"description,omitempty"`
	Repository    string           `json:"repository,omitempty"`
	DefaultBranch string           `json:"default_branch,omitempty"`
	Team          string           `json:"team,omitempty"`
	Settings      database.JSONMap `json:"settings,omitempty"`
}

// CreateSuiteRunInput represents input for creating a suite run directly
type CreateSuiteRunInput struct {
	TestRunID uint       `json:"test_run_id" binding:"required"`
	SuiteName string     `json:"suite_name" binding:"required"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Status    string     `json:"status,omitempty"`
}

// CreateSpecRunInput represents input for creating a spec run directly
type CreateSpecRunInput struct {
	SuiteRunID   uint       `json:"suite_run_id" binding:"required"`
	SpecName     string     `json:"spec_name" binding:"required"`
	Status       string     `json:"status" binding:"required,oneof=passed failed skipped pending"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StackTrace   string     `json:"stack_trace,omitempty"`
}

// ListProjectsFilter represents filters for listing projects
type ListProjectsFilter struct {
	Search     string   `json:"search,omitempty"`
	ActiveOnly bool     `json:"active_only,omitempty"`
	Teams      []string `json:"teams,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

// CreateTagInput represents input for creating a tag
type CreateTagInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// UpdateTagInput represents input for updating a tag
type UpdateTagInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// ListTagsFilter represents filters for listing tags
type ListTagsFilter struct {
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
