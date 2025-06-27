package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/guidewire-oss/fern-platform/internal/domains/projects/domain"
)

// UpdateProjectCommand represents the command to update a project
type UpdateProjectCommand struct {
	ID            uint                   `json:"id"`
	Name          *string                `json:"name"`
	Description   *string                `json:"description"`
	Repository    *string                `json:"repository"`
	DefaultBranch *string                `json:"default_branch"`
	Team          *string                `json:"team"`
	Settings      map[string]interface{} `json:"settings"`
	UpdatedBy     string                 `json:"updated_by"`
}

// UpdateProjectHandler handles the update project use case
type UpdateProjectHandler struct {
	projectRepo    domain.ProjectRepository
	permissionRepo domain.ProjectPermissionRepository
}

// NewUpdateProjectHandler creates a new handler
func NewUpdateProjectHandler(
	projectRepo domain.ProjectRepository,
	permissionRepo domain.ProjectPermissionRepository,
) *UpdateProjectHandler {
	return &UpdateProjectHandler{
		projectRepo:    projectRepo,
		permissionRepo: permissionRepo,
	}
}

// Handle executes the use case
func (h *UpdateProjectHandler) Handle(ctx context.Context, cmd UpdateProjectCommand) (*domain.ProjectSnapshot, error) {
	if cmd.ID == 0 {
		return nil, errors.New("project ID is required")
	}
	if cmd.UpdatedBy == "" {
		return nil, errors.New("updated by is required")
	}

	// Find the project
	project, err := h.projectRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Check permissions
	hasPermission, err := h.checkWritePermission(ctx, project.ProjectID(), cmd.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("insufficient permissions to update project")
	}

	// Update fields if provided
	if cmd.Name != nil {
		if err := project.UpdateName(*cmd.Name); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}
	
	if cmd.Description != nil {
		project.UpdateDescription(*cmd.Description)
	}
	
	if cmd.Repository != nil {
		project.UpdateRepository(*cmd.Repository)
	}
	
	if cmd.DefaultBranch != nil {
		if err := project.UpdateDefaultBranch(*cmd.DefaultBranch); err != nil {
			return nil, fmt.Errorf("failed to update default branch: %w", err)
		}
	}
	
	if cmd.Team != nil {
		if err := project.UpdateTeam(domain.Team(*cmd.Team)); err != nil {
			return nil, fmt.Errorf("failed to update team: %w", err)
		}
	}
	
	// Update settings
	for k, v := range cmd.Settings {
		project.SetSetting(k, v)
	}

	// Save the updated project
	if err := h.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Return snapshot
	snapshot := project.ToSnapshot()
	return &snapshot, nil
}

// checkWritePermission checks if a user has write permission on a project
func (h *UpdateProjectHandler) checkWritePermission(ctx context.Context, projectID domain.ProjectID, userID string) (bool, error) {
	permissions, err := h.permissionRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.CanWrite() {
			return true, nil
		}
	}

	return false, nil
}