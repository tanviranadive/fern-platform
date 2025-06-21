// Package repository provides data access layer for the fern-reporter service
package repository

import (
	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// SpecRunRepository handles spec run data operations
type SpecRunRepository struct {
	*database.BaseRepository
	db *gorm.DB
}

// NewSpecRunRepository creates a new spec run repository
func NewSpecRunRepository(db *gorm.DB) *SpecRunRepository {
	return &SpecRunRepository{
		BaseRepository: database.NewBaseRepository(db),
		db:             db,
	}
}

// CreateSpecRun creates a new spec run
func (r *SpecRunRepository) CreateSpecRun(specRun *database.SpecRun) error {
	return r.db.Create(specRun).Error
}

// GetSpecRunByID retrieves a spec run by ID
func (r *SpecRunRepository) GetSpecRunByID(id uint) (*database.SpecRun, error) {
	var specRun database.SpecRun
	err := r.db.First(&specRun, id).Error
	return &specRun, err
}

// UpdateSpecRun updates an existing spec run
func (r *SpecRunRepository) UpdateSpecRun(specRun *database.SpecRun) error {
	return r.db.Save(specRun).Error
}

// DeleteSpecRun soft deletes a spec run
func (r *SpecRunRepository) DeleteSpecRun(id uint) error {
	return r.db.Delete(&database.SpecRun{}, id).Error
}

// ListSpecRunsBySuite retrieves all spec runs for a suite run
func (r *SpecRunRepository) ListSpecRunsBySuite(suiteRunID uint) ([]*database.SpecRun, error) {
	var specRuns []*database.SpecRun
	err := r.db.Where("suite_run_id = ?", suiteRunID).
		Order("start_time ASC").
		Find(&specRuns).Error
	return specRuns, err
}

// GetSpecRunsByName retrieves spec runs by name across suite runs
func (r *SpecRunRepository) GetSpecRunsByName(specName string, limit int) ([]*database.SpecRun, error) {
	var specRuns []*database.SpecRun
	query := r.db.Where("spec_name = ?", specName).
		Order("start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&specRuns).Error
	return specRuns, err
}

// GetFailedSpecRuns retrieves failed spec runs with optional filtering
func (r *SpecRunRepository) GetFailedSpecRuns(projectID string, limit int) ([]*database.SpecRun, error) {
	query := r.db.Joins("JOIN suite_runs ON spec_runs.suite_run_id = suite_runs.id").
		Joins("JOIN test_runs ON suite_runs.test_run_id = test_runs.id").
		Where("spec_runs.status = 'failed'").
		Order("spec_runs.start_time DESC")
	
	if projectID != "" {
		query = query.Where("test_runs.project_id = ?", projectID)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	var specRuns []*database.SpecRun
	err := query.Find(&specRuns).Error
	return specRuns, err
}

// GetFlakySpecRuns retrieves flaky spec runs (specs that have been marked as flaky)
func (r *SpecRunRepository) GetFlakySpecRuns(projectID string, limit int) ([]*database.SpecRun, error) {
	query := r.db.Joins("JOIN suite_runs ON spec_runs.suite_run_id = suite_runs.id").
		Joins("JOIN test_runs ON suite_runs.test_run_id = test_runs.id").
		Where("spec_runs.is_flaky = true").
		Order("spec_runs.start_time DESC")
	
	if projectID != "" {
		query = query.Where("test_runs.project_id = ?", projectID)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	var specRuns []*database.SpecRun
	err := query.Find(&specRuns).Error
	return specRuns, err
}

// MarkSpecAsFlaky marks a spec as flaky
func (r *SpecRunRepository) MarkSpecAsFlaky(id uint) error {
	return r.db.Model(&database.SpecRun{}).
		Where("id = ?", id).
		Update("is_flaky", true).Error
}

// GetSpecExecutionHistory returns execution history for a specific spec
func (r *SpecRunRepository) GetSpecExecutionHistory(specName, suiteName string, limit int) ([]*database.SpecRun, error) {
	query := r.db.Joins("JOIN suite_runs ON spec_runs.suite_run_id = suite_runs.id").
		Where("spec_runs.spec_name = ?", specName)
	
	if suiteName != "" {
		query = query.Where("suite_runs.suite_name = ?", suiteName)
	}
	
	query = query.Order("spec_runs.start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	var specRuns []*database.SpecRun
	err := query.Find(&specRuns).Error
	return specRuns, err
}