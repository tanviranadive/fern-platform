package domain

import (
	"errors"
	"time"
)

// PermissionType represents the type of permission
type PermissionType string

const (
	PermissionRead   PermissionType = "read"
	PermissionWrite  PermissionType = "write"
	PermissionDelete PermissionType = "delete"
	PermissionAdmin  PermissionType = "admin"
)

// ProjectPermission represents a permission granted to a user for a project
type ProjectPermission struct {
	projectID  ProjectID
	userID     string
	permission PermissionType
	grantedBy  string
	grantedAt  time.Time
	expiresAt  *time.Time
}

// NewProjectPermission creates a new project permission
func NewProjectPermission(projectID ProjectID, userID string, permission PermissionType, grantedBy string) (*ProjectPermission, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if grantedBy == "" {
		return nil, errors.New("granted by cannot be empty")
	}
	if !isValidPermission(permission) {
		return nil, errors.New("invalid permission type")
	}

	return &ProjectPermission{
		projectID:  projectID,
		userID:     userID,
		permission: permission,
		grantedBy:  grantedBy,
		grantedAt:  time.Now(),
	}, nil
}

// ProjectID returns the project ID
func (pp *ProjectPermission) ProjectID() ProjectID {
	return pp.projectID
}

// UserID returns the user ID
func (pp *ProjectPermission) UserID() string {
	return pp.userID
}

// Permission returns the permission type
func (pp *ProjectPermission) Permission() PermissionType {
	return pp.permission
}

// SetExpiration sets the expiration time
func (pp *ProjectPermission) SetExpiration(expiresAt time.Time) error {
	if expiresAt.Before(time.Now()) {
		return errors.New("expiration time must be in the future")
	}
	pp.expiresAt = &expiresAt
	return nil
}

// IsExpired checks if the permission has expired
func (pp *ProjectPermission) IsExpired() bool {
	if pp.expiresAt == nil {
		return false
	}
	return time.Now().After(*pp.expiresAt)
}

// CanRead checks if the permission allows reading
func (pp *ProjectPermission) CanRead() bool {
	if pp.IsExpired() {
		return false
	}
	return pp.permission == PermissionRead || pp.permission == PermissionWrite || 
	       pp.permission == PermissionDelete || pp.permission == PermissionAdmin
}

// CanWrite checks if the permission allows writing
func (pp *ProjectPermission) CanWrite() bool {
	if pp.IsExpired() {
		return false
	}
	return pp.permission == PermissionWrite || pp.permission == PermissionDelete || 
	       pp.permission == PermissionAdmin
}

// CanDelete checks if the permission allows deletion
func (pp *ProjectPermission) CanDelete() bool {
	if pp.IsExpired() {
		return false
	}
	return pp.permission == PermissionDelete || pp.permission == PermissionAdmin
}

// CanAdmin checks if the permission allows administration
func (pp *ProjectPermission) CanAdmin() bool {
	if pp.IsExpired() {
		return false
	}
	return pp.permission == PermissionAdmin
}

// isValidPermission validates the permission type
func isValidPermission(p PermissionType) bool {
	switch p {
	case PermissionRead, PermissionWrite, PermissionDelete, PermissionAdmin:
		return true
	default:
		return false
	}
}