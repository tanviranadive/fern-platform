// Package repository provides data access layer for projects
package repository

import (
	"time"
	
	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// ProjectRepository handles project data operations
type ProjectRepository struct {
	*database.BaseRepository
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		BaseRepository: database.NewBaseRepository(db),
		db:             db,
	}
}

// CreateProject creates a new project
func (r *ProjectRepository) CreateProject(project *database.ProjectDetails) error {
	return r.db.Create(project).Error
}

// GetProjectByID retrieves a project by ID
func (r *ProjectRepository) GetProjectByID(id uint) (*database.ProjectDetails, error) {
	var project database.ProjectDetails
	err := r.db.First(&project, id).Error
	return &project, err
}

// GetProjectByProjectID retrieves a project by project_id
func (r *ProjectRepository) GetProjectByProjectID(projectID string) (*database.ProjectDetails, error) {
	var project database.ProjectDetails
	err := r.db.Where("project_id = ?", projectID).First(&project).Error
	return &project, err
}

// UpdateProject updates an existing project
func (r *ProjectRepository) UpdateProject(project *database.ProjectDetails) error {
	return r.db.Save(project).Error
}

// DeleteProject soft deletes a project
func (r *ProjectRepository) DeleteProject(id uint) error {
	return r.db.Delete(&database.ProjectDetails{}, id).Error
}

// ListProjects retrieves all projects with optional filtering
func (r *ProjectRepository) ListProjects(search string, activeOnly bool, limit, offset int) ([]*database.ProjectDetails, int64, error) {
	query := r.db.Model(&database.ProjectDetails{})
	
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR project_id ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	
	if activeOnly {
		query = query.Where("is_active = true")
	}
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	query = query.Order("name ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	var projects []*database.ProjectDetails
	err := query.Find(&projects).Error
	return projects, total, err
}

// GetOrCreateProject gets a project by project_id or creates it if it doesn't exist
func (r *ProjectRepository) GetOrCreateProject(projectID, name, repository, defaultBranch string) (*database.ProjectDetails, error) {
	var project database.ProjectDetails
	err := r.db.Where("project_id = ?", projectID).First(&project).Error
	
	if err == gorm.ErrRecordNotFound {
		// Project doesn't exist, create it
		project = database.ProjectDetails{
			ProjectID:     projectID,
			Name:          name,
			Repository:    repository,
			DefaultBranch: defaultBranch,
			IsActive:      true,
		}
		if err := r.db.Create(&project).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &project, nil
}

// DeactivateProject marks a project as inactive
func (r *ProjectRepository) DeactivateProject(projectID string) error {
	return r.db.Model(&database.ProjectDetails{}).
		Where("project_id = ?", projectID).
		Update("is_active", false).Error
}

// ActivateProject marks a project as active
func (r *ProjectRepository) ActivateProject(projectID string) error {
	return r.db.Model(&database.ProjectDetails{}).
		Where("project_id = ?", projectID).
		Update("is_active", true).Error
}

// GetProjectStats returns statistics for a project
func (r *ProjectRepository) GetProjectStats(projectID string) (*ProjectStats, error) {
	var stats ProjectStats
	
	// Get total test runs
	err := r.db.Model(&database.TestRun{}).
		Where("project_id = ?", projectID).
		Count(&stats.TotalTestRuns).Error
	if err != nil {
		return nil, err
	}
	
	// Get recent activity (last 30 days)
	err = r.db.Model(&database.TestRun{}).
		Where("project_id = ? AND start_time >= NOW() - INTERVAL '30 days'", projectID).
		Count(&stats.RecentTestRuns).Error
	if err != nil {
		return nil, err
	}
	
	// Get unique branches
	err = r.db.Model(&database.TestRun{}).
		Where("project_id = ?", projectID).
		Distinct("branch").
		Count(&stats.UniqueBranches).Error
	if err != nil {
		return nil, err
	}
	
	// Get success rate for last 30 days
	var successfulRuns int64
	err = r.db.Model(&database.TestRun{}).
		Where("project_id = ? AND status = 'passed' AND start_time >= NOW() - INTERVAL '30 days'", projectID).
		Count(&successfulRuns).Error
	if err != nil {
		return nil, err
	}
	
	if stats.RecentTestRuns > 0 {
		stats.SuccessRate = float64(successfulRuns) / float64(stats.RecentTestRuns) * 100
	}
	
	// Get average duration from recent test runs
	var avgDuration float64
	err = r.db.Model(&database.TestRun{}).
		Where("project_id = ? AND duration_ms > 0 AND start_time >= NOW() - INTERVAL '30 days'", projectID).
		Select("AVG(duration_ms)").
		Scan(&avgDuration).Error
	if err != nil {
		return nil, err
	}
	stats.AverageDuration = int64(avgDuration)
	
	// Get last run time
	var lastRun database.TestRun
	err = r.db.Model(&database.TestRun{}).
		Where("project_id = ?", projectID).
		Order("start_time DESC").
		First(&lastRun).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err != gorm.ErrRecordNotFound {
		stats.LastRunTime = &lastRun.StartTime
	}
	
	return &stats, nil
}

// ProjectStats represents project statistics
type ProjectStats struct {
	TotalTestRuns   int64      `json:"total_test_runs"`
	RecentTestRuns  int64      `json:"recent_test_runs"`  // Last 30 days
	UniqueBranches  int64      `json:"unique_branches"`
	SuccessRate     float64    `json:"success_rate"`      // Last 30 days
	AverageDuration int64      `json:"average_duration"`  // Average duration in milliseconds
	LastRunTime     *time.Time `json:"last_run_time"`     // Most recent test run time
}