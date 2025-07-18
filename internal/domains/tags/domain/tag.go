package domain

import (
	"errors"
	"strings"
	"time"
)

// TagID represents a unique identifier for a tag
type TagID string

// Tag represents a label that can be applied to test runs
type Tag struct {
	id        TagID
	name      string
	createdAt time.Time
}

// TagSnapshot represents a point-in-time view of a tag
type TagSnapshot struct {
	ID        TagID     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTag creates a new tag
func NewTag(name string) (*Tag, error) {
	if name == "" {
		return nil, errors.New("tag name cannot be empty")
	}

	// Normalize tag name
	normalizedName := strings.TrimSpace(strings.ToLower(name))
	if normalizedName == "" {
		return nil, errors.New("tag name cannot be empty after normalization")
	}

	return &Tag{
		id:        TagID(""), // ID will be assigned by database
		name:      normalizedName,
		createdAt: time.Now(),
	}, nil
}

// ID returns the tag ID
func (t *Tag) ID() TagID {
	return t.id
}

// Name returns the tag name
func (t *Tag) Name() string {
	return t.name
}

// CreatedAt returns the creation time
func (t *Tag) CreatedAt() time.Time {
	return t.createdAt
}

// ToSnapshot returns a snapshot of the tag
func (t *Tag) ToSnapshot() TagSnapshot {
	return TagSnapshot{
		ID:        t.id,
		Name:      t.name,
		CreatedAt: t.createdAt,
	}
}
