package infrastructure

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/internal/domains/projects/domain"
)

// ProjectPermissionDB represents the database model for project permissions
type ProjectPermissionDB struct {
	ID         uint      `gorm:"primaryKey"`
	ProjectID  string    `gorm:"index:idx_project_user_permission,unique"`
	UserID     string    `gorm:"index:idx_project_user_permission,unique"`
	Permission string    `gorm:"index:idx_project_user_permission,unique"`
	GrantedBy  string
	GrantedAt  time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName specifies the table name
func (ProjectPermissionDB) TableName() string {
	return "project_permissions"
}

// GormProjectPermissionRepository is a GORM implementation of ProjectPermissionRepository
type GormProjectPermissionRepository struct {
	db *gorm.DB
}

// NewGormProjectPermissionRepository creates a new GORM project permission repository
func NewGormProjectPermissionRepository(db *gorm.DB) *GormProjectPermissionRepository {
	return &GormProjectPermissionRepository{db: db}
}

// Save persists a project permission
func (r *GormProjectPermissionRepository) Save(ctx context.Context, permission *domain.ProjectPermission) error {
	dbPermission := &ProjectPermissionDB{
		ProjectID:  string(permission.ProjectID()),
		UserID:     permission.UserID(),
		Permission: string(permission.Permission()),
		GrantedBy:  permission.UserID(), // Note: domain model doesn't expose GrantedBy
		GrantedAt:  time.Now(),
	}

	// Check if permission is expired and set expiration
	// Note: This is a limitation of the current domain model which doesn't expose expiration directly

	if err := r.db.WithContext(ctx).Create(dbPermission).Error; err != nil {
		return fmt.Errorf("failed to save project permission: %w", err)
	}

	return nil
}

// FindByProjectAndUser retrieves permissions for a specific project and user
func (r *GormProjectPermissionRepository) FindByProjectAndUser(ctx context.Context, projectID domain.ProjectID, userID string) ([]*domain.ProjectPermission, error) {
	var dbPermissions []ProjectPermissionDB
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", string(projectID), userID).
		Find(&dbPermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find permissions: %w", err)
	}

	permissions := make([]*domain.ProjectPermission, 0, len(dbPermissions))
	for _, dbPerm := range dbPermissions {
		perm, err := r.toDomainModel(&dbPerm)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// FindByUser retrieves all permissions for a user
func (r *GormProjectPermissionRepository) FindByUser(ctx context.Context, userID string) ([]*domain.ProjectPermission, error) {
	var dbPermissions []ProjectPermissionDB
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&dbPermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find user permissions: %w", err)
	}

	permissions := make([]*domain.ProjectPermission, 0, len(dbPermissions))
	for _, dbPerm := range dbPermissions {
		perm, err := r.toDomainModel(&dbPerm)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// FindByProject retrieves all permissions for a project
func (r *GormProjectPermissionRepository) FindByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.ProjectPermission, error) {
	var dbPermissions []ProjectPermissionDB
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", string(projectID)).
		Find(&dbPermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find project permissions: %w", err)
	}

	permissions := make([]*domain.ProjectPermission, 0, len(dbPermissions))
	for _, dbPerm := range dbPermissions {
		perm, err := r.toDomainModel(&dbPerm)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// Delete removes a permission
func (r *GormProjectPermissionRepository) Delete(ctx context.Context, projectID domain.ProjectID, userID string, permission domain.PermissionType) error {
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ? AND permission = ?", string(projectID), userID, string(permission)).
		Delete(&ProjectPermissionDB{}).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// DeleteExpired removes all expired permissions
func (r *GormProjectPermissionRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&ProjectPermissionDB{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired permissions: %w", err)
	}
	return nil
}

// toDomainModel converts a database model to a domain model
func (r *GormProjectPermissionRepository) toDomainModel(dbPerm *ProjectPermissionDB) (*domain.ProjectPermission, error) {
	perm, err := domain.NewProjectPermission(
		domain.ProjectID(dbPerm.ProjectID),
		dbPerm.UserID,
		domain.PermissionType(dbPerm.Permission),
		dbPerm.GrantedBy,
	)
	if err != nil {
		return nil, err
	}

	// Set expiration if exists
	if dbPerm.ExpiresAt != nil {
		if err := perm.SetExpiration(*dbPerm.ExpiresAt); err != nil {
			// Ignore error if expiration is in the past
			// The domain model will handle this appropriately
		}
	}

	return perm, nil
}