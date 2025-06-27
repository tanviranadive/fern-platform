package infrastructure

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	
	"github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// GormTestRunRepository implements the TestRunRepository using GORM
type GormTestRunRepository struct {
	db *gorm.DB
	legacyRepo *repository.TestRunRepository
}

// NewGormTestRunRepository creates a new GORM-based repository
func NewGormTestRunRepository(db *gorm.DB) *GormTestRunRepository {
	return &GormTestRunRepository{
		db: db,
		legacyRepo: repository.NewTestRunRepository(db),
	}
}

// Save persists a test run
func (r *GormTestRunRepository) Save(ctx context.Context, testRun *domain.TestRun) error {
	// Convert domain model to database model
	dbModel := r.toDBModel(testRun)
	
	// Save to database
	if err := r.db.WithContext(ctx).Create(&dbModel).Error; err != nil {
		return fmt.Errorf("failed to save test run: %w", err)
	}
	
	return nil
}

// FindByID retrieves a test run by ID
func (r *GormTestRunRepository) FindByID(ctx context.Context, id domain.TestRunID) (*domain.TestRun, error) {
	var dbModel database.TestRun
	
	err := r.db.WithContext(ctx).
		Where("run_id = ?", string(id)).
		Preload("SuiteRuns").
		Preload("SuiteRuns.SpecRuns").
		First(&dbModel).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find test run: %w", err)
	}
	
	return r.toDomainModel(&dbModel)
}

// FindByProjectID retrieves test runs for a project
func (r *GormTestRunRepository) FindByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]*domain.TestRun, error) {
	var dbModels []database.TestRun
	
	query := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("start_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&dbModels).Error; err != nil {
		return nil, fmt.Errorf("failed to find test runs: %w", err)
	}
	
	// Convert to domain models
	domainModels := make([]*domain.TestRun, len(dbModels))
	for i, dbModel := range dbModels {
		dm, err := r.toDomainModel(&dbModel)
		if err != nil {
			return nil, err
		}
		domainModels[i] = dm
	}
	
	return domainModels, nil
}

// Update updates an existing test run
func (r *GormTestRunRepository) Update(ctx context.Context, testRun *domain.TestRun) error {
	snapshot := testRun.ToSnapshot()
	
	// Update only specific fields
	updates := map[string]interface{}{
		"status":        string(snapshot.Status),
		"end_time":      snapshot.EndTime,
		"total_tests":   snapshot.TotalTests,
		"passed_tests":  snapshot.PassedTests,
		"failed_tests":  snapshot.FailedTests,
		"skipped_tests": snapshot.SkippedTests,
		"duration_ms":   snapshot.Duration,
		"metadata":      snapshot.Metadata,
	}
	
	if err := r.db.WithContext(ctx).
		Model(&database.TestRun{}).
		Where("run_id = ?", string(snapshot.ID)).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update test run: %w", err)
	}
	
	return nil
}

// toDBModel converts domain model to database model
func (r *GormTestRunRepository) toDBModel(testRun *domain.TestRun) *database.TestRun {
	snapshot := testRun.ToSnapshot()
	
	return &database.TestRun{
		ProjectID:    snapshot.ProjectID,
		RunID:        string(snapshot.ID),
		Branch:       snapshot.Branch,
		CommitSHA:    snapshot.CommitSHA,
		Status:       string(snapshot.Status),
		StartTime:    snapshot.StartTime,
		EndTime:      snapshot.EndTime,
		TotalTests:   snapshot.TotalTests,
		PassedTests:  snapshot.PassedTests,
		FailedTests:  snapshot.FailedTests,
		SkippedTests: snapshot.SkippedTests,
		Duration:     snapshot.Duration,
		Environment:  snapshot.Environment,
		Metadata:     database.JSONMap(snapshot.Metadata),
	}
}

// toDomainModel converts database model to domain model
func (r *GormTestRunRepository) toDomainModel(dbModel *database.TestRun) (*domain.TestRun, error) {
	// For now, we'll create a simple conversion
	// In a full implementation, this would reconstruct the full domain object
	testRun, err := domain.NewTestRun(
		domain.TestRunID(dbModel.RunID),
		dbModel.ProjectID,
		dbModel.Branch,
	)
	if err != nil {
		return nil, err
	}
	
	// Set metadata
	for k, v := range dbModel.Metadata {
		testRun.SetMetadata(k, v)
	}
	
	// Note: This is a simplified version. In a full implementation,
	// we would need to properly reconstruct the test run with all its
	// suite runs and spec runs from the database
	
	return testRun, nil
}

// Implement remaining methods...
func (r *GormTestRunRepository) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.TestRun, error) {
	// Implementation here
	return nil, nil
}

func (r *GormTestRunRepository) Count(ctx context.Context, filter domain.TestRunFilter) (int64, error) {
	// Implementation here
	return 0, nil
}