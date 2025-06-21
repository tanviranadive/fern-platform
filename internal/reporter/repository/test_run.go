// Package repository provides data access layer for the fern-reporter service
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// TestRunRepository handles test run data operations
type TestRunRepository struct {
	*database.BaseRepository
	db *gorm.DB
}

// NewTestRunRepository creates a new test run repository
func NewTestRunRepository(db *gorm.DB) *TestRunRepository {
	return &TestRunRepository{
		BaseRepository: database.NewBaseRepository(db),
		db:             db,
	}
}

// CreateTestRun creates a new test run
func (r *TestRunRepository) CreateTestRun(testRun *database.TestRun) error {
	return r.db.Create(testRun).Error
}

// GetTestRunByID retrieves a test run by ID with all related data
func (r *TestRunRepository) GetTestRunByID(id uint) (*database.TestRun, error) {
	var testRun database.TestRun
	err := r.db.Preload("Tags").
		Preload("SuiteRuns").
		Preload("SuiteRuns.SpecRuns").
		First(&testRun, id).Error
	return &testRun, err
}

// GetTestRunByRunID retrieves a test run by run_id with all related data
func (r *TestRunRepository) GetTestRunByRunID(runID string) (*database.TestRun, error) {
	var testRun database.TestRun
	err := r.db.Preload("Tags").
		Preload("SuiteRuns").
		Preload("SuiteRuns.SpecRuns").
		Where("run_id = ?", runID).
		First(&testRun).Error
	return &testRun, err
}

// UpdateTestRun updates an existing test run
func (r *TestRunRepository) UpdateTestRun(testRun *database.TestRun) error {
	return r.db.Save(testRun).Error
}

// DeleteTestRun soft deletes a test run
func (r *TestRunRepository) DeleteTestRun(id uint) error {
	return r.db.Delete(&database.TestRun{}, id).Error
}

// ListTestRunsFilter represents filters for listing test runs
type ListTestRunsFilter struct {
	ProjectID   string
	Branch      string
	Status      string
	Environment string
	StartTime   *time.Time
	EndTime     *time.Time
	Tags        []string
	Limit       int
	Offset      int
	OrderBy     string
	Order       string // ASC or DESC
}

// ListTestRuns retrieves test runs with filtering and pagination
func (r *TestRunRepository) ListTestRuns(filter ListTestRunsFilter) ([]*database.TestRun, int64, error) {
	query := r.db.Model(&database.TestRun{})

	// Apply filters
	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Branch != "" {
		query = query.Where("branch = ?", filter.Branch)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Environment != "" {
		query = query.Where("environment = ?", filter.Environment)
	}
	if filter.StartTime != nil {
		query = query.Where("start_time >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("start_time <= ?", filter.EndTime)
	}

	// Filter by tags if provided
	if len(filter.Tags) > 0 {
		query = query.Joins("JOIN test_run_tags ON test_runs.id = test_run_tags.test_run_id").
			Joins("JOIN tags ON test_run_tags.tag_id = tags.id").
			Where("tags.name IN ?", filter.Tags).
			Group("test_runs.id").
			Having("COUNT(DISTINCT tags.id) = ?", len(filter.Tags))
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering
	orderBy := "start_time"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	order := "DESC"
	if filter.Order != "" {
		order = filter.Order
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, order))

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Preload related data and exclude metadata to avoid scanning issues
	query = query.Preload("Tags").Select("id, created_at, updated_at, deleted_at, project_id, run_id, branch, commit_sha, status, start_time, end_time, total_tests, passed_tests, failed_tests, skipped_tests, duration_ms, environment")

	var testRuns []*database.TestRun
	err := query.Find(&testRuns).Error
	return testRuns, total, err
}

// GetTestRunStats returns aggregated statistics for test runs
func (r *TestRunRepository) GetTestRunStats(projectID string, days int) (*TestRunStats, error) {
	var stats TestRunStats
	
	// Base query for the time range
	timeFilter := time.Now().AddDate(0, 0, -days)
	baseQuery := r.db.Model(&database.TestRun{}).
		Where("start_time >= ?", timeFilter)
	
	if projectID != "" {
		baseQuery = baseQuery.Where("project_id = ?", projectID)
	}

	// Total runs
	if err := baseQuery.Count(&stats.TotalRuns).Error; err != nil {
		return nil, err
	}

	// Runs by status
	var statusCounts []StatusCount
	err := baseQuery.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error
	if err != nil {
		return nil, err
	}
	
	stats.StatusCounts = statusCounts

	// Average duration
	var avgDuration float64
	err = baseQuery.Select("AVG(duration_ms) as avg_duration").
		Where("status = 'completed'").
		Scan(&avgDuration).Error
	if err != nil {
		return nil, err
	}
	stats.AverageDuration = int64(avgDuration)

	// Success rate
	var successfulRuns int64
	err = baseQuery.Where("status = 'passed'").Count(&successfulRuns).Error
	if err != nil {
		return nil, err
	}
	
	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(successfulRuns) / float64(stats.TotalRuns) * 100
	}

	return &stats, nil
}

// TestRunStats represents aggregated test run statistics
type TestRunStats struct {
	TotalRuns       int64         `json:"total_runs"`
	StatusCounts    []StatusCount `json:"status_counts"`
	AverageDuration int64         `json:"average_duration_ms"`
	SuccessRate     float64       `json:"success_rate"`
}

// StatusCount represents count by status
type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// GetRecentTestRuns returns the most recent test runs for a project
func (r *TestRunRepository) GetRecentTestRuns(projectID string, limit int) ([]*database.TestRun, error) {
	var testRuns []*database.TestRun
	query := r.db.Preload("Tags").
		Order("start_time DESC").
		Limit(limit)
	
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	
	err := query.Find(&testRuns).Error
	return testRuns, err
}

// AssignTagsToTestRun assigns tags to a test run
func (r *TestRunRepository) AssignTagsToTestRun(testRunID uint, tagIDs []uint) error {
	// Remove existing associations
	if err := r.db.Exec("DELETE FROM test_run_tags WHERE test_run_id = ?", testRunID).Error; err != nil {
		return err
	}

	// Add new associations
	if len(tagIDs) > 0 {
		for _, tagID := range tagIDs {
			if err := r.db.Exec("INSERT INTO test_run_tags (test_run_id, tag_id) VALUES (?, ?)", testRunID, tagID).Error; err != nil {
				return err
			}
		}
	}

	return nil
}