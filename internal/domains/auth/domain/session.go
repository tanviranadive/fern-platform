package domain

import (
	"time"
)

// Session represents an authenticated user session
type Session struct {
	SessionID    string
	UserID       string
	User         *User
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresAt    time.Time
	IsActive     bool
	IPAddress    string
	UserAgent    string
	LastActivity time.Time
	CreatedAt    time.Time
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (active and not expired)
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

// Invalidate marks the session as inactive
func (s *Session) Invalidate() {
	s.IsActive = false
}
