// Package service provides business logic for projects
package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
)

// ProjectService handles project business logic
type ProjectService struct {
	projectRepo *repository.ProjectRepository
	logger      *logging.Logger
}

// NewProjectService creates a new project service
func NewProjectService(projectRepo *repository.ProjectRepository, logger *logging.Logger) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		logger:      logger,
	}
}

// CreateProjectInput represents input for creating a project
type CreateProjectInput struct {
	ProjectID     string                 `json:"project_id"`
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description,omitempty"`
	Repository    string                 `json:"repository,omitempty"`
	DefaultBranch string                 `json:"default_branch,omitempty"`
	Settings      map[string]interface{} `json:"settings,omitempty"`
	Team          string                 `json:"team,omitempty"`
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(input CreateProjectInput) (*database.ProjectDetails, error) {
	// Generate UUID if ProjectID is empty
	if input.ProjectID == "" {
		input.ProjectID = uuid.New().String()
		s.logger.WithFields(map[string]interface{}{
			"generated_project_id": input.ProjectID,
		}).Info("Generated new UUID for project")
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": input.ProjectID,
		"name":       input.Name,
		"team":       input.Team,
	}).Info("Creating project with details")

	// Check if project already exists
	existing, err := s.projectRepo.GetProjectByProjectID(input.ProjectID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("project with ID %s already exists", input.ProjectID)
	}

	project := &database.ProjectDetails{
		ProjectID:     input.ProjectID,
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

	if err := s.projectRepo.CreateProject(project); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"project_id": input.ProjectID,
		}).WithError(err).Error("Failed to create project")
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id":    project.ProjectID,
		"database_id":   project.ID,
		"name":          project.Name,
		"team":          project.Team,
		"input_team":    input.Team,
	}).Info("Project created successfully - verifying saved data")

	return project, nil
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(id uint) (*database.ProjectDetails, error) {
	return s.projectRepo.GetProjectByID(id)
}

// GetProjectByProjectID retrieves a project by project ID
func (s *ProjectService) GetProjectByProjectID(projectID string) (*database.ProjectDetails, error) {
	return s.projectRepo.GetProjectByProjectID(projectID)
}

// UpdateProjectInput represents input for updating a project
type UpdateProjectInput struct {
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Repository    string                 `json:"repository,omitempty"`
	DefaultBranch string                 `json:"default_branch,omitempty"`
	Settings      map[string]interface{} `json:"settings,omitempty"`
	Team          string                 `json:"team,omitempty"`
}

// UpdateProject updates an existing project by numeric ID
func (s *ProjectService) UpdateProject(id uint, input UpdateProjectInput) (*database.ProjectDetails, error) {
	project, err := s.projectRepo.GetProjectByID(id)
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

	if err := s.projectRepo.UpdateProject(project); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"project_id": project.ProjectID,
			"id":         project.ID,
		}).WithError(err).Error("Failed to update project")
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project updated successfully")

	return project, nil
}

// UpdateProjectByProjectID updates an existing project by project ID string
func (s *ProjectService) UpdateProjectByProjectID(projectID string, input UpdateProjectInput) (*database.ProjectDetails, error) {
	project, err := s.projectRepo.GetProjectByProjectID(projectID)
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

	if err := s.projectRepo.UpdateProject(project); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"project_id": project.ProjectID,
			"id":         project.ID,
		}).WithError(err).Error("Failed to update project")
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project updated successfully")

	return project, nil
}

// ListProjectsFilter represents filters for listing projects
type ListProjectsFilter struct {
	Search     string   `json:"search,omitempty"`
	ActiveOnly bool     `json:"active_only,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
	Teams      []string `json:"teams,omitempty"` // Filter by team names
}

// ListProjects retrieves projects with filtering
func (s *ProjectService) ListProjects(filter ListProjectsFilter) ([]*database.ProjectDetails, int64, error) {
	return s.projectRepo.ListProjects(filter.Search, filter.ActiveOnly, filter.Limit, filter.Offset, filter.Teams)
}

// DeleteProject deletes a project by numeric ID
func (s *ProjectService) DeleteProject(id uint) error {
	project, err := s.projectRepo.GetProjectByID(id)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if err := s.projectRepo.DeleteProject(id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project deleted")

	return nil
}

// DeleteProjectByProjectID deletes a project by project ID string
func (s *ProjectService) DeleteProjectByProjectID(projectID string) error {
	project, err := s.projectRepo.GetProjectByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	if err := s.projectRepo.DeleteProject(project.ID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": project.ProjectID,
		"id":         project.ID,
	}).Info("Project deleted")

	return nil
}

// DeactivateProject marks a project as inactive
func (s *ProjectService) DeactivateProject(projectID string) error {
	if err := s.projectRepo.DeactivateProject(projectID); err != nil {
		return fmt.Errorf("failed to deactivate project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": projectID,
	}).Info("Project deactivated")

	return nil
}

// ActivateProject marks a project as active
func (s *ProjectService) ActivateProject(projectID string) error {
	if err := s.projectRepo.ActivateProject(projectID); err != nil {
		return fmt.Errorf("failed to activate project: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"project_id": projectID,
	}).Info("Project activated")

	return nil
}

// GetProjectStats retrieves project statistics
func (s *ProjectService) GetProjectStats(projectID string) (*repository.ProjectStats, error) {
	return s.projectRepo.GetProjectStats(projectID)
}

// GetOrCreateProject gets a project or creates it if it doesn't exist
func (s *ProjectService) GetOrCreateProject(projectID, name, repository, defaultBranch string) (*database.ProjectDetails, error) {
	return s.projectRepo.GetOrCreateProject(projectID, name, repository, defaultBranch)
}