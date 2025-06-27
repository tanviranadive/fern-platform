package domain

import (
	"context"
)

// ProjectRepository defines the interface for project persistence
type ProjectRepository interface {
	// Save persists a project
	Save(ctx context.Context, project *Project) error
	
	// FindByID retrieves a project by its internal ID
	FindByID(ctx context.Context, id uint) (*Project, error)
	
	// FindByProjectID retrieves a project by its project ID
	FindByProjectID(ctx context.Context, projectID ProjectID) (*Project, error)
	
	// FindByTeam retrieves all projects for a team
	FindByTeam(ctx context.Context, team Team) ([]*Project, error)
	
	// FindAll retrieves all projects with pagination
	FindAll(ctx context.Context, limit, offset int) ([]*Project, int64, error)
	
	// Update updates an existing project
	Update(ctx context.Context, project *Project) error
	
	// Delete deletes a project
	Delete(ctx context.Context, id uint) error
	
	// ExistsByProjectID checks if a project exists with the given project ID
	ExistsByProjectID(ctx context.Context, projectID ProjectID) (bool, error)
}

// ProjectPermissionRepository defines the interface for project permission persistence
type ProjectPermissionRepository interface {
	// Save persists a project permission
	Save(ctx context.Context, permission *ProjectPermission) error
	
	// FindByProjectAndUser retrieves permissions for a specific project and user
	FindByProjectAndUser(ctx context.Context, projectID ProjectID, userID string) ([]*ProjectPermission, error)
	
	// FindByUser retrieves all permissions for a user
	FindByUser(ctx context.Context, userID string) ([]*ProjectPermission, error)
	
	// FindByProject retrieves all permissions for a project
	FindByProject(ctx context.Context, projectID ProjectID) ([]*ProjectPermission, error)
	
	// Delete removes a permission
	Delete(ctx context.Context, projectID ProjectID, userID string, permission PermissionType) error
	
	// DeleteExpired removes all expired permissions
	DeleteExpired(ctx context.Context) error
}