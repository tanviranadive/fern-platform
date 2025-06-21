// Package repository provides data access layer for flaky tests
package repository

import (
	"time"

	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// FlakyTestRepository handles flaky test data operations
type FlakyTestRepository struct {
	*database.BaseRepository
	db *gorm.DB
}

// NewFlakyTestRepository creates a new flaky test repository
func NewFlakyTestRepository(db *gorm.DB) *FlakyTestRepository {
	return &FlakyTestRepository{
		BaseRepository: database.NewBaseRepository(db),
		db:             db,
	}
}

// CreateFlakyTest creates a new flaky test record
func (r *FlakyTestRepository) CreateFlakyTest(flakyTest *database.FlakyTest) error {
	return r.db.Create(flakyTest).Error
}

// GetFlakyTestByID retrieves a flaky test by ID
func (r *FlakyTestRepository) GetFlakyTestByID(id uint) (*database.FlakyTest, error) {
	var flakyTest database.FlakyTest
	err := r.db.First(&flakyTest, id).Error
	return &flakyTest, err
}

// UpdateFlakyTest updates an existing flaky test
func (r *FlakyTestRepository) UpdateFlakyTest(flakyTest *database.FlakyTest) error {
	return r.db.Save(flakyTest).Error
}

// DeleteFlakyTest soft deletes a flaky test
func (r *FlakyTestRepository) DeleteFlakyTest(id uint) error {
	return r.db.Delete(&database.FlakyTest{}, id).Error
}

// FlakyTestFilter represents filters for listing flaky tests
type FlakyTestFilter struct {
	ProjectID     string
	Severity      string
	Status        string
	MinFlakeRate  float64
	MaxFlakeRate  float64
	Limit         int
	Offset        int
	OrderBy       string
	Order         string // ASC or DESC
}

// ListFlakyTests retrieves flaky tests with filtering and pagination
func (r *FlakyTestRepository) ListFlakyTests(filter FlakyTestFilter) ([]*database.FlakyTest, int64, error) {
	query := r.db.Model(&database.FlakyTest{})

	// Apply filters
	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.MinFlakeRate > 0 {
		query = query.Where("flake_rate >= ?", filter.MinFlakeRate)
	}
	if filter.MaxFlakeRate > 0 {
		query = query.Where("flake_rate <= ?", filter.MaxFlakeRate)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering
	orderBy := "flake_rate"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	order := "DESC"
	if filter.Order != "" {
		order = filter.Order
	}
	query = query.Order(orderBy + " " + order)

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var flakyTests []*database.FlakyTest
	err := query.Find(&flakyTests).Error
	return flakyTests, total, err
}

// GetOrCreateFlakyTest gets or creates a flaky test record
func (r *FlakyTestRepository) GetOrCreateFlakyTest(projectID, testName, suiteName string) (*database.FlakyTest, error) {
	var flakyTest database.FlakyTest
	err := r.db.Where("project_id = ? AND test_name = ? AND suite_name = ?", 
		projectID, testName, suiteName).First(&flakyTest).Error

	if err == gorm.ErrRecordNotFound {
		// Flaky test doesn't exist, create it
		now := time.Now()
		flakyTest = database.FlakyTest{
			ProjectID:        projectID,
			TestName:         testName,
			SuiteName:        suiteName,
			FlakeRate:        0.0,
			TotalExecutions:  0,
			FlakyExecutions:  0,
			FirstSeenAt:      now,
			LastSeenAt:       now,
			Status:           "active",
			Severity:         "low",
		}
		if err := r.db.Create(&flakyTest).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &flakyTest, nil
}

// UpdateFlakyTestExecution updates flaky test execution statistics
func (r *FlakyTestRepository) UpdateFlakyTestExecution(projectID, testName, suiteName string, isFlaky bool, errorMessage string) error {
	flakyTest, err := r.GetOrCreateFlakyTest(projectID, testName, suiteName)
	if err != nil {
		return err
	}

	// Update execution counts
	flakyTest.TotalExecutions++
	flakyTest.LastSeenAt = time.Now()

	if isFlaky {
		flakyTest.FlakyExecutions++
		if errorMessage != "" {
			flakyTest.LastErrorMessage = errorMessage
		}
	}

	// Calculate flake rate
	if flakyTest.TotalExecutions > 0 {
		flakyTest.FlakeRate = float64(flakyTest.FlakyExecutions) / float64(flakyTest.TotalExecutions)
	}

	// Update severity based on flake rate
	flakyTest.Severity = r.calculateSeverity(flakyTest.FlakeRate, flakyTest.TotalExecutions)

	return r.UpdateFlakyTest(flakyTest)
}

// calculateSeverity determines severity based on flake rate and total executions
func (r *FlakyTestRepository) calculateSeverity(flakeRate float64, totalExecutions int) string {
	// Need minimum executions to be reliable
	if totalExecutions < 5 {
		return "low"
	}

	if flakeRate >= 0.5 {
		return "critical"
	} else if flakeRate >= 0.3 {
		return "high"
	} else if flakeRate >= 0.1 {
		return "medium"
	}
	return "low"
}

// GetFlakyTestsByProject returns flaky tests for a specific project
func (r *FlakyTestRepository) GetFlakyTestsByProject(projectID string, severity string, limit int) ([]*database.FlakyTest, error) {
	query := r.db.Where("project_id = ? AND status = 'active'", projectID)

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	query = query.Order("flake_rate DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var flakyTests []*database.FlakyTest
	err := query.Find(&flakyTests).Error
	return flakyTests, err
}

// GetFlakyTestStats returns aggregated flaky test statistics
func (r *FlakyTestRepository) GetFlakyTestStats(projectID string) (*FlakyTestStats, error) {
	var stats FlakyTestStats

	query := r.db.Model(&database.FlakyTest{}).Where("status = 'active'")
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// Total flaky tests
	if err := query.Count(&stats.TotalFlakyTests).Error; err != nil {
		return nil, err
	}

	// Count by severity
	var severityCounts []SeverityCount
	err := query.Select("severity, COUNT(*) as count").
		Group("severity").
		Scan(&severityCounts).Error
	if err != nil {
		return nil, err
	}
	stats.SeverityCounts = severityCounts

	// Average flake rate
	var avgFlakeRate float64
	err = query.Select("AVG(flake_rate) as avg_flake_rate").
		Where("total_executions >= 5"). // Only consider tests with enough data
		Scan(&avgFlakeRate).Error
	if err != nil {
		return nil, err
	}
	stats.AverageFlakeRate = avgFlakeRate

	// Most flaky test
	var mostFlaky database.FlakyTest
	err = query.Order("flake_rate DESC").First(&mostFlaky).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		stats.MostFlakyTest = &mostFlaky
	}

	return &stats, nil
}

// FlakyTestStats represents aggregated flaky test statistics
type FlakyTestStats struct {
	TotalFlakyTests   int64                  `json:"total_flaky_tests"`
	SeverityCounts    []SeverityCount        `json:"severity_counts"`
	AverageFlakeRate  float64                `json:"average_flake_rate"`
	MostFlakyTest     *database.FlakyTest    `json:"most_flaky_test,omitempty"`
}

// SeverityCount represents count by severity
type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

// MarkFlakyTestResolved marks a flaky test as resolved
func (r *FlakyTestRepository) MarkFlakyTestResolved(id uint) error {
	return r.db.Model(&database.FlakyTest{}).
		Where("id = ?", id).
		Update("status", "resolved").Error
}

// GetRecentlyAddedFlakyTests returns recently added flaky tests
func (r *FlakyTestRepository) GetRecentlyAddedFlakyTests(projectID string, days int, limit int) ([]*database.FlakyTest, error) {
	query := r.db.Where("first_seen_at >= ?", time.Now().AddDate(0, 0, -days))

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	query = query.Order("first_seen_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var flakyTests []*database.FlakyTest
	err := query.Find(&flakyTests).Error
	return flakyTests, err
}