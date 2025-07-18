package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/guidewire-oss/fern-platform/internal/domains/auth/domain"
)

// AuthorizationService handles permission checks
type AuthorizationService struct {
	userRepo domain.UserRepository
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(userRepo domain.UserRepository) *AuthorizationService {
	return &AuthorizationService{
		userRepo: userRepo,
	}
}

// CanAccessProject checks if a user can access a specific project
func (s *AuthorizationService) CanAccessProject(ctx context.Context, user *domain.User, projectID string, requiredAction string) (bool, error) {
	// Admin can do anything
	if user.IsAdmin() {
		return true, nil
	}

	// Check scopes
	scopes, err := s.userRepo.GetUserScopes(ctx, user.UserID)
	if err != nil {
		return false, fmt.Errorf("failed to get user scopes: %w", err)
	}

	// Check if user has matching scope
	for _, scope := range scopes {
		if scope.ExpiresAt != nil && scope.ExpiresAt.Before(time.Now()) {
			continue // Skip expired scopes
		}

		if s.matchProjectScope(scope.Scope, projectID, requiredAction) {
			return true, nil
		}
	}

	return false, nil
}

// CanManageTeam checks if a user can manage a specific team
func (s *AuthorizationService) CanManageTeam(ctx context.Context, user *domain.User, team string) bool {
	return user.IsManagerForTeam(team)
}

// GrantScope grants a scope to a user
func (s *AuthorizationService) GrantScope(ctx context.Context, scope domain.UserScope) error {
	return s.userRepo.GrantScope(ctx, scope)
}

// RevokeScope revokes a scope from a user
func (s *AuthorizationService) RevokeScope(ctx context.Context, userID, scope string) error {
	return s.userRepo.RevokeScope(ctx, userID, scope)
}

// matchProjectScope checks if a user scope matches the required project action
func (s *AuthorizationService) matchProjectScope(userScope, projectID, action string) bool {
	// Expected scope formats:
	// - project:read:PROJECT_ID
	// - project:write:PROJECT_ID
	// - project:*:PROJECT_ID (all actions)
	// - project:ACTION:* (action on all projects)
	// - project:*:* (all actions on all projects)

	parts := strings.Split(userScope, ":")
	if len(parts) != 3 || parts[0] != "project" {
		return false
	}

	scopeAction := parts[1]
	scopeProject := parts[2]

	// Check action match
	actionMatch := scopeAction == "*" || scopeAction == action

	// Check project match
	projectMatch := scopeProject == "*" || scopeProject == projectID

	return actionMatch && projectMatch
}
