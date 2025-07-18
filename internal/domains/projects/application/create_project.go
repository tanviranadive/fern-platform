package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/guidewire-oss/fern-platform/internal/domains/projects/domain"
)

// CreateProjectCommand represents the command to create a project
type CreateProjectCommand struct {
	ProjectID     string                 `json:"project_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Repository    string                 `json:"repository"`
	DefaultBranch string                 `json:"default_branch"`
	Team          string                 `json:"team"`
	Settings      map[string]interface{} `json:"settings"`
	CreatedBy     string                 `json:"created_by"`
}

// CreateProjectHandler handles the create project use case
type CreateProjectHandler struct {
	projectRepo    domain.ProjectRepository
	permissionRepo domain.ProjectPermissionRepository
}

// NewCreateProjectHandler creates a new handler
func NewCreateProjectHandler(
	projectRepo domain.ProjectRepository,
	permissionRepo domain.ProjectPermissionRepository,
) *CreateProjectHandler {
	return &CreateProjectHandler{
		projectRepo:    projectRepo,
		permissionRepo: permissionRepo,
	}
}

// Handle executes the use case
func (h *CreateProjectHandler) Handle(ctx context.Context, cmd CreateProjectCommand) (*domain.ProjectSnapshot, error) {
	// Validate command
	if err := h.validateCommand(cmd); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Generate project ID if not provided
	projectID := cmd.ProjectID
	if projectID == "" {
		projectID = uuid.New().String()
	}

	// Check if project already exists
	exists, err := h.projectRepo.ExistsByProjectID(ctx, domain.ProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if exists {
		return nil, errors.New("project already exists")
	}

	// Create new project
	project, err := domain.NewProject(
		domain.ProjectID(projectID),
		cmd.Name,
		domain.Team(cmd.Team),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Set additional properties
	if cmd.Description != "" {
		project.UpdateDescription(cmd.Description)
	}
	if cmd.Repository != "" {
		project.UpdateRepository(cmd.Repository)
	}
	if cmd.DefaultBranch != "" {
		project.UpdateDefaultBranch(cmd.DefaultBranch)
	}

	// Set settings
	for k, v := range cmd.Settings {
		project.SetSetting(k, v)
	}

	// Save project
	if err := h.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	// Grant creator admin permission
	if cmd.CreatedBy != "" {
		permission, err := domain.NewProjectPermission(
			project.ProjectID(),
			cmd.CreatedBy,
			domain.PermissionAdmin,
			cmd.CreatedBy,
		)
		if err != nil {
			// Log error but don't fail project creation
			fmt.Printf("failed to create creator permission: %v\n", err)
		} else {
			if err := h.permissionRepo.Save(ctx, permission); err != nil {
				// Log error but don't fail project creation
				fmt.Printf("failed to save creator permission: %v\n", err)
			}
		}
	}

	// Return snapshot
	snapshot := project.ToSnapshot()
	return &snapshot, nil
}

func (h *CreateProjectHandler) validateCommand(cmd CreateProjectCommand) error {
	if cmd.Name == "" {
		return errors.New("project name is required")
	}
	if cmd.Team == "" {
		return errors.New("team is required")
	}
	return nil
}
