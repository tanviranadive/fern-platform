package domain

import (
	"context"
	"time"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// User operations
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	FindByID(ctx context.Context, userID string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByIDOrEmail(ctx context.Context, userID, email string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error

	// Group operations
	SetUserGroups(ctx context.Context, userID string, groups []string) error
	GetUserGroups(ctx context.Context, userID string) ([]UserGroup, error)

	// Scope operations
	GrantScope(ctx context.Context, scope UserScope) error
	RevokeScope(ctx context.Context, userID, scope string) error
	GetUserScopes(ctx context.Context, userID string) ([]UserScope, error)
}

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, sessionID string) (*Session, error)
	FindActiveByID(ctx context.Context, sessionID string) (*Session, error)
	UpdateActivity(ctx context.Context, sessionID string) error
	Invalidate(ctx context.Context, sessionID string) error
	InvalidateAllForUser(ctx context.Context, userID string) error
	CleanupExpired(ctx context.Context) error
}
