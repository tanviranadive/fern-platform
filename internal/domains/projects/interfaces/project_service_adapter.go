package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/internal/domains/projects/application"
	"github.com/guidewire-oss/fern-platform/internal/domains/projects/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
)

// ProjectServiceAdapter adapts the new domain to the existing ProjectService interface
// This ensures complete backward compatibility with REST and GraphQL APIs
type ProjectServiceAdapter struct {
	createProjectHandler *application.CreateProjectHandler
	updateProjectHandler *application.UpdateProjectHandler
	projectRepo         domain.ProjectRepository
	legacyRepo          *repository.ProjectRepository
	logger              *logging.Logger
}

// NewProjectServiceAdapter creates a new adapter that implements the existing service interface
func NewProjectServiceAdapter(
	createHandler *application.CreateProjectHandler,
	updateHandler *application.UpdateProjectHandler,
	domainRepo domain.ProjectRepository,
	legacyRepo *repository.ProjectRepository,
	logger *logging.Logger,
) *ProjectServiceAdapter {
	return &ProjectServiceAdapter{
		createProjectHandler: createHandler,
		updateProjectHandler: updateHandler,
		projectRepo:         domainRepo,
		legacyRepo:          legacyRepo,
		logger:              logger,
	}
}

// CreateProject implements the existing service interface method
func (a *ProjectServiceAdapter) CreateProject(input service.CreateProjectInput) (*database.ProjectDetails, error) {
	// Always generate a new UUID for the project ID
	projectID := uuid.New().String()
	
	a.logger.WithFields(map[string]interface{}{
		"project_id":         projectID,
		"name":               input.Name,
		"team":               input.Team,
		"input_project_id":   input.ProjectID, // Log if one was provided
	}).Info("Creating project with generated UUID")

	// Note: We're not checking for duplicate names here since project names 
	// don't need to be unique across the system. The UUID ensures uniqueness.

	// For now, we'll use the legacy repository directly to ensure complete compatibility
	// The domain handlers will be integrated gradually
	project := &database.ProjectDetails{
		ProjectID:     projectID, // Always use the generated UUID
		Name:          input.Name,
		Description:   input.Description,
		Repository:    input.Repository,
		DefaultBranch: input.DefaultBranch,
		Settings:      "{}",
		IsActive:      true,
		Team:          input.Team,
	}

	if input.DefaultBranch == "" {
		project.DefaultBranch = "main"
	}
	
	// Handle settings if provided
	if input.Settings != nil {
		settingsJSON, err := json.Marshal(input.Settings)
		if err == nil {
			project.Settings = string(settingsJSON)
		}
	}

	if err := a.legacyRepo.CreateProject(project); err != nil {
		a.logger.WithFields(map[string]interface{}{
			"project_id": input.ProjectID,
			"name":       input.Name,
			"error_type": fmt.Sprintf("%T", err),
			"error_msg":  err.Error(),
		}).WithError(err).Error("Failed to create project in database")
		
		// Check if it's a duplicate key error
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("project with ID %s already exists", input.ProjectID)
		}
		
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id":    project.ProjectID,
		"database_id":   project.ID,
		"name":          project.Name,
		"team":          project.Team,
	}).Info("Project created successfully")

	return project, nil
}

// GetProject retrieves a project by ID
func (a *ProjectServiceAdapter) GetProject(id uint) (*database.ProjectDetails, error) {
	return a.legacyRepo.GetProjectByID(id)
}

// GetProjectByProjectID retrieves a project by project ID
func (a *ProjectServiceAdapter) GetProjectByProjectID(projectID string) (*database.ProjectDetails, error) {
	return a.legacyRepo.GetProjectByProjectID(projectID)
}

// UpdateProject updates an existing project by numeric ID
func (a *ProjectServiceAdapter) UpdateProject(id uint, input service.UpdateProjectInput) (*database.ProjectDetails, error) {
	project, err := a.legacyRepo.GetProjectByID(id)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Update fields if provided
	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}
	if input.Repository != "" {
		project.Repository = input.Repository
	}
	if input.DefaultBranch != "" {
		project.DefaultBranch = input.DefaultBranch
	}
	if input.Team != "" {
		project.Team = input.Team
	}
	if input.Settings != nil {
		// Marshal settings to JSON string
		settingsJSON, err := json.Marshal(input.Settings)
		if err == nil {
			project.Settings = string(settingsJSON)
		}
	}

	if err := a.legacyRepo.UpdateProject(project); err != nil {
		a.logger.WithFields(map[string]interface{}{
			"project_id": project.ProjectID,
			"id":         project.ID,
		}).WithError(err).Error("Failed to update project")
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project updated successfully")

	return project, nil
}

// UpdateProjectByProjectID updates an existing project by project ID string
func (a *ProjectServiceAdapter) UpdateProjectByProjectID(projectID string, input service.UpdateProjectInput) (*database.ProjectDetails, error) {
	project, err := a.legacyRepo.GetProjectByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Update fields if provided
	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}
	if input.Repository != "" {
		project.Repository = input.Repository
	}
	if input.DefaultBranch != "" {
		project.DefaultBranch = input.DefaultBranch
	}
	if input.Team != "" {
		project.Team = input.Team
	}
	if input.Settings != nil {
		// Marshal settings to JSON string
		settingsJSON, err := json.Marshal(input.Settings)
		if err == nil {
			project.Settings = string(settingsJSON)
		}
	}

	if err := a.legacyRepo.UpdateProject(project); err != nil {
		a.logger.WithFields(map[string]interface{}{
			"project_id": project.ProjectID,
			"id":         project.ID,
		}).WithError(err).Error("Failed to update project")
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project updated successfully")

	return project, nil
}

// ListProjects retrieves projects with filtering
func (a *ProjectServiceAdapter) ListProjects(filter service.ListProjectsFilter) ([]*database.ProjectDetails, int64, error) {
	return a.legacyRepo.ListProjects(filter.Search, filter.ActiveOnly, filter.Limit, filter.Offset, filter.Teams)
}

// DeleteProject deletes a project by numeric ID
func (a *ProjectServiceAdapter) DeleteProject(id uint) error {
	project, err := a.legacyRepo.GetProjectByID(id)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if err := a.legacyRepo.DeleteProject(id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project deleted")

	return nil
}

// DeleteProjectByProjectID deletes a project by project ID string
func (a *ProjectServiceAdapter) DeleteProjectByProjectID(projectID string) error {
	project, err := a.legacyRepo.GetProjectByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if err := a.legacyRepo.DeleteProject(project.ID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project deleted")

	return nil
}

// DeactivateProject marks a project as inactive
func (a *ProjectServiceAdapter) DeactivateProject(projectID string) error {
	if err := a.legacyRepo.DeactivateProject(projectID); err != nil {
		return fmt.Errorf("failed to deactivate project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id": projectID,
	}).Info("Project deactivated")

	return nil
}

// ActivateProject marks a project as active
func (a *ProjectServiceAdapter) ActivateProject(projectID string) error {
	if err := a.legacyRepo.ActivateProject(projectID); err != nil {
		return fmt.Errorf("failed to activate project: %w", err)
	}

	a.logger.WithFields(map[string]interface{}{
		"project_id": projectID,
	}).Info("Project activated")

	return nil
}

// GetProjectStats retrieves project statistics
func (a *ProjectServiceAdapter) GetProjectStats(projectID string) (*repository.ProjectStats, error) {
	return a.legacyRepo.GetProjectStats(projectID)
}

// GetOrCreateProject gets a project or creates it if it doesn't exist
func (a *ProjectServiceAdapter) GetOrCreateProject(projectID, name, repository, defaultBranch string) (*database.ProjectDetails, error) {
	// If projectID is provided, try to get the existing project
	if projectID != "" {
		existing, err := a.legacyRepo.GetProjectByProjectID(projectID)
		if err == nil && existing != nil {
			return existing, nil
		}
		// If not found with the provided ID, we'll create a new one with a generated ID
	}
	
	// Create a new project with a generated UUID
	newProjectID := uuid.New().String()
	
	project := &database.ProjectDetails{
		ProjectID:     newProjectID,
		Name:          name,
		Description:   "",
		Repository:    repository,
		DefaultBranch: defaultBranch,
		Settings:      "{}",
		IsActive:      true,
		Team:          "",
	}
	
	if defaultBranch == "" {
		project.DefaultBranch = "main"
	}
	
	if err := a.legacyRepo.CreateProject(project); err != nil {
		// If creation fails due to a race condition, try to get the project again
		if strings.Contains(err.Error(), "duplicate key") {
			// Another process might have created it, try to fetch again
			if projectID != "" {
				return a.legacyRepo.GetProjectByProjectID(projectID)
			}
			return nil, fmt.Errorf("failed to create project: %w", err)
		}
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	return project, nil
}