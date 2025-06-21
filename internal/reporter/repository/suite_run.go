// Package repository provides data access layer for the fern-reporter service
package repository

import (
	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// SuiteRunRepository handles suite run data operations
type SuiteRunRepository struct {
	*database.BaseRepository
	db *gorm.DB
}

// NewSuiteRunRepository creates a new suite run repository
func NewSuiteRunRepository(db *gorm.DB) *SuiteRunRepository {
	return &SuiteRunRepository{
		BaseRepository: database.NewBaseRepository(db),
		db:             db,
	}
}

// CreateSuiteRun creates a new suite run
func (r *SuiteRunRepository) CreateSuiteRun(suiteRun *database.SuiteRun) error {
	return r.db.Create(suiteRun).Error
}

// GetSuiteRunByID retrieves a suite run by ID with all related data
func (r *SuiteRunRepository) GetSuiteRunByID(id uint) (*database.SuiteRun, error) {
	var suiteRun database.SuiteRun
	err := r.db.Preload("SpecRuns").First(&suiteRun, id).Error
	return &suiteRun, err
}

// UpdateSuiteRun updates an existing suite run
func (r *SuiteRunRepository) UpdateSuiteRun(suiteRun *database.SuiteRun) error {
	return r.db.Save(suiteRun).Error
}

// DeleteSuiteRun soft deletes a suite run
func (r *SuiteRunRepository) DeleteSuiteRun(id uint) error {
	return r.db.Delete(&database.SuiteRun{}, id).Error
}

// ListSuiteRunsByTestRun retrieves all suite runs for a test run
func (r *SuiteRunRepository) ListSuiteRunsByTestRun(testRunID uint) ([]*database.SuiteRun, error) {
	var suiteRuns []*database.SuiteRun
	err := r.db.Preload("SpecRuns").
		Where("test_run_id = ?", testRunID).
		Order("start_time ASC").
		Find(&suiteRuns).Error
	return suiteRuns, err
}

// GetSuiteRunsBySuiteName retrieves suite runs by suite name across test runs
func (r *SuiteRunRepository) GetSuiteRunsBySuiteName(suiteName string, limit int) ([]*database.SuiteRun, error) {
	var suiteRuns []*database.SuiteRun
	query := r.db.Where("suite_name = ?", suiteName).
		Order("start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&suiteRuns).Error
	return suiteRuns, err
}

// UpdateSuiteRunStats updates the suite run statistics
func (r *SuiteRunRepository) UpdateSuiteRunStats(suiteRunID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var stats struct {
			TotalSpecs   int `json:"total_specs"`
			PassedSpecs  int `json:"passed_specs"`
			FailedSpecs  int `json:"failed_specs"`
			SkippedSpecs int `json:"skipped_specs"`
		}

		// Calculate stats from spec runs
		err := tx.Model(&database.SpecRun{}).
			Select(`
				COUNT(*) as total_specs,
				COUNT(CASE WHEN status = 'passed' THEN 1 END) as passed_specs,
				COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_specs,
				COUNT(CASE WHEN status = 'skipped' THEN 1 END) as skipped_specs
			`).
			Where("suite_run_id = ?", suiteRunID).
			Scan(&stats).Error
		if err != nil {
			return err
		}

		// Update suite run with calculated stats
		return tx.Model(&database.SuiteRun{}).
			Where("id = ?", suiteRunID).
			Updates(map[string]interface{}{
				"total_specs":   stats.TotalSpecs,
				"passed_specs":  stats.PassedSpecs,
				"failed_specs":  stats.FailedSpecs,
				"skipped_specs": stats.SkippedSpecs,
			}).Error
	})
}