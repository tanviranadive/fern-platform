package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/guidewire-oss/fern-platform/internal/domains/auth/domain"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GORM-based user repository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// Create creates a new user
func (r *GormUserRepository) Create(ctx context.Context, user *domain.User) error {
	dbUser := &database.User{
		UserID:        user.UserID,
		Email:         user.Email,
		Name:          user.Name,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Role:          string(user.Role),
		Status:        string(user.Status),
		ProfileURL:    user.ProfileURL,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
	}

	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update updates an existing user
func (r *GormUserRepository) Update(ctx context.Context, user *domain.User) error {
	updates := map[string]interface{}{
		"email":          user.Email,
		"name":           user.Name,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"role":           string(user.Role),
		"status":         string(user.Status),
		"profile_url":    user.ProfileURL,
		"email_verified": user.EmailVerified,
		"updated_at":     time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&database.User{}).Where("user_id = ?", user.UserID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// FindByID finds a user by ID
func (r *GormUserRepository) FindByID(ctx context.Context, userID string) (*domain.User, error) {
	var dbUser database.User
	if err := r.db.WithContext(ctx).Preload("UserGroups").Preload("UserScopes").Where("user_id = ?", userID).First(&dbUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return r.toDomainUser(&dbUser), nil
}

// FindByEmail finds a user by email
func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dbUser database.User
	if err := r.db.WithContext(ctx).Preload("UserGroups").Preload("UserScopes").Where("email = ?", email).First(&dbUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return r.toDomainUser(&dbUser), nil
}

// FindByIDOrEmail finds a user by ID or email
func (r *GormUserRepository) FindByIDOrEmail(ctx context.Context, userID, email string) (*domain.User, error) {
	var dbUser database.User
	if err := r.db.WithContext(ctx).Preload("UserGroups").Preload("UserScopes").Where("user_id = ? OR email = ?", userID, email).First(&dbUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return r.toDomainUser(&dbUser), nil
}

// UpdateLastLogin updates the user's last login time
func (r *GormUserRepository) UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error {
	return r.db.WithContext(ctx).Model(&database.User{}).Where("user_id = ?", userID).Update("last_login_at", loginTime).Error
}

// SetUserGroups sets the user's group memberships
func (r *GormUserRepository) SetUserGroups(ctx context.Context, userID string, groups []string) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()

	// Delete existing groups
	if err := tx.Where("user_id = ?", userID).Delete(&database.UserGroup{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete existing groups: %w", err)
	}

	// Add new groups
	for _, group := range groups {
		userGroup := &database.UserGroup{
			UserID:    userID,
			GroupName: group,
		}
		if err := tx.Create(userGroup).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create user group: %w", err)
		}
	}

	return tx.Commit().Error
}

// GetUserGroups gets the user's group memberships
func (r *GormUserRepository) GetUserGroups(ctx context.Context, userID string) ([]domain.UserGroup, error) {
	var dbGroups []database.UserGroup
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&dbGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	groups := make([]domain.UserGroup, len(dbGroups))
	for i, dbGroup := range dbGroups {
		groups[i] = domain.UserGroup{
			UserID:    dbGroup.UserID,
			GroupName: dbGroup.GroupName,
			CreatedAt: dbGroup.CreatedAt,
		}
	}

	return groups, nil
}

// GrantScope grants a scope to a user
func (r *GormUserRepository) GrantScope(ctx context.Context, scope domain.UserScope) error {
	dbScope := &database.UserScope{
		UserID:    scope.UserID,
		Scope:     scope.Scope,
		ExpiresAt: scope.ExpiresAt,
		GrantedBy: scope.GrantedBy,
	}

	return r.db.WithContext(ctx).Create(dbScope).Error
}

// RevokeScope revokes a scope from a user
func (r *GormUserRepository) RevokeScope(ctx context.Context, userID, scope string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND scope = ?", userID, scope).Delete(&database.UserScope{}).Error
}

// GetUserScopes gets the user's scopes
func (r *GormUserRepository) GetUserScopes(ctx context.Context, userID string) ([]domain.UserScope, error) {
	var dbScopes []database.UserScope
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&dbScopes).Error; err != nil {
		return nil, fmt.Errorf("failed to get user scopes: %w", err)
	}

	scopes := make([]domain.UserScope, len(dbScopes))
	for i, dbScope := range dbScopes {
		scopes[i] = domain.UserScope{
			UserID:    dbScope.UserID,
			Scope:     dbScope.Scope,
			ExpiresAt: dbScope.ExpiresAt,
			GrantedBy: dbScope.GrantedBy,
			GrantedAt: dbScope.CreatedAt,
		}
	}

	return scopes, nil
}

// Helper method to convert database user to domain user
func (r *GormUserRepository) toDomainUser(dbUser *database.User) *domain.User {
	user := &domain.User{
		UserID:        dbUser.UserID,
		Email:         dbUser.Email,
		Name:          dbUser.Name,
		FirstName:     dbUser.FirstName,
		LastName:      dbUser.LastName,
		Role:          domain.UserRole(dbUser.Role),
		Status:        domain.UserStatus(dbUser.Status),
		ProfileURL:    dbUser.ProfileURL,
		EmailVerified: dbUser.EmailVerified,
		LastLoginAt:   dbUser.LastLoginAt,
		CreatedAt:     dbUser.CreatedAt,
		UpdatedAt:     dbUser.UpdatedAt,
		Groups:        make([]domain.UserGroup, len(dbUser.UserGroups)),
		Scopes:        make([]domain.UserScope, len(dbUser.UserScopes)),
	}

	for i, group := range dbUser.UserGroups {
		user.Groups[i] = domain.UserGroup{
			UserID:    group.UserID,
			GroupName: group.GroupName,
			CreatedAt: group.CreatedAt,
		}
	}

	for i, scope := range dbUser.UserScopes {
		user.Scopes[i] = domain.UserScope{
			UserID:    scope.UserID,
			Scope:     scope.Scope,
			ExpiresAt: scope.ExpiresAt,
			GrantedBy: scope.GrantedBy,
			GrantedAt: scope.CreatedAt,
		}
	}

	return user
}
