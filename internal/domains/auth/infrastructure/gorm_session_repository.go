package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/guidewire-oss/fern-platform/internal/domains/auth/domain"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"gorm.io/gorm"
)

// GormSessionRepository implements SessionRepository using GORM
type GormSessionRepository struct {
	db *gorm.DB
}

// NewGormSessionRepository creates a new GORM-based session repository
func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

// Create creates a new session
func (r *GormSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	dbSession := &database.UserSession{
		UserID:       session.UserID,
		SessionID:    session.SessionID,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		IDToken:      session.IDToken,
		ExpiresAt:    session.ExpiresAt,
		IsActive:     session.IsActive,
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		LastActivity: session.LastActivity,
	}

	if err := r.db.WithContext(ctx).Create(dbSession).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// FindByID finds a session by ID
func (r *GormSessionRepository) FindByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	var dbSession database.UserSession
	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&dbSession).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return r.toDomainSession(&dbSession), nil
}

// FindActiveByID finds an active session by ID
func (r *GormSessionRepository) FindActiveByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	var dbSession database.UserSession
	if err := r.db.WithContext(ctx).Where("session_id = ? AND is_active = ? AND expires_at > ?",
		sessionID, true, time.Now()).First(&dbSession).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active session not found")
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return r.toDomainSession(&dbSession), nil
}

// UpdateActivity updates the session's last activity time
func (r *GormSessionRepository) UpdateActivity(ctx context.Context, sessionID string) error {
	result := r.db.WithContext(ctx).Model(&database.UserSession{}).
		Where("session_id = ?", sessionID).
		Update("last_activity", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to update session activity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// Invalidate marks a session as inactive
func (r *GormSessionRepository) Invalidate(ctx context.Context, sessionID string) error {
	result := r.db.WithContext(ctx).Model(&database.UserSession{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to invalidate session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// InvalidateAllForUser invalidates all sessions for a user
func (r *GormSessionRepository) InvalidateAllForUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&database.UserSession{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

// CleanupExpired removes expired sessions
func (r *GormSessionRepository) CleanupExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&database.UserSession{}).Error
}

// Helper method to convert database session to domain session
func (r *GormSessionRepository) toDomainSession(dbSession *database.UserSession) *domain.Session {
	return &domain.Session{
		SessionID:    dbSession.SessionID,
		UserID:       dbSession.UserID,
		AccessToken:  dbSession.AccessToken,
		RefreshToken: dbSession.RefreshToken,
		IDToken:      dbSession.IDToken,
		ExpiresAt:    dbSession.ExpiresAt,
		IsActive:     dbSession.IsActive,
		IPAddress:    dbSession.IPAddress,
		UserAgent:    dbSession.UserAgent,
		LastActivity: dbSession.LastActivity,
		CreatedAt:    dbSession.CreatedAt,
	}
}
