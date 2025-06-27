package domain

import (
	"errors"
	"time"
)

// ProjectID represents a unique identifier for a project
type ProjectID string

// Team represents a team that owns projects
type Team string

// Project represents a project in the domain
type Project struct {
	id            uint
	projectID     ProjectID
	name          string
	description   string
	repository    string
	defaultBranch string
	team          Team
	isActive      bool
	settings      map[string]interface{}
	createdAt     time.Time
	updatedAt     time.Time
}

// NewProject creates a new project
func NewProject(projectID ProjectID, name string, team Team) (*Project, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}
	if team == "" {
		return nil, errors.New("team cannot be empty")
	}

	now := time.Now()
	return &Project{
		projectID:     projectID,
		name:          name,
		team:          team,
		defaultBranch: "main",
		isActive:      true,
		settings:      make(map[string]interface{}),
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// ID returns the internal database ID
func (p *Project) ID() uint {
	return p.id
}

// ProjectID returns the project ID
func (p *Project) ProjectID() ProjectID {
	return p.projectID
}

// Name returns the project name
func (p *Project) Name() string {
	return p.name
}

// Team returns the team that owns the project
func (p *Project) Team() Team {
	return p.team
}

// IsActive returns whether the project is active
func (p *Project) IsActive() bool {
	return p.isActive
}

// UpdateName updates the project name
func (p *Project) UpdateName(name string) error {
	if name == "" {
		return errors.New("project name cannot be empty")
	}
	p.name = name
	p.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the project description
func (p *Project) UpdateDescription(description string) {
	p.description = description
	p.updatedAt = time.Now()
}

// UpdateRepository updates the repository URL
func (p *Project) UpdateRepository(repository string) {
	p.repository = repository
	p.updatedAt = time.Now()
}

// UpdateDefaultBranch updates the default branch
func (p *Project) UpdateDefaultBranch(branch string) error {
	if branch == "" {
		return errors.New("default branch cannot be empty")
	}
	p.defaultBranch = branch
	p.updatedAt = time.Now()
	return nil
}

// UpdateTeam changes the team ownership
func (p *Project) UpdateTeam(team Team) error {
	if team == "" {
		return errors.New("team cannot be empty")
	}
	p.team = team
	p.updatedAt = time.Now()
	return nil
}

// Activate activates the project
func (p *Project) Activate() {
	p.isActive = true
	p.updatedAt = time.Now()
}

// Deactivate deactivates the project
func (p *Project) Deactivate() {
	p.isActive = false
	p.updatedAt = time.Now()
}

// SetSetting sets a project setting
func (p *Project) SetSetting(key string, value interface{}) {
	p.settings[key] = value
	p.updatedAt = time.Now()
}

// GetSetting retrieves a project setting
func (p *Project) GetSetting(key string) (interface{}, bool) {
	val, exists := p.settings[key]
	return val, exists
}

// ToSnapshot returns a read-only snapshot of the project
func (p *Project) ToSnapshot() ProjectSnapshot {
	return ProjectSnapshot{
		ID:            p.id,
		ProjectID:     p.projectID,
		Name:          p.name,
		Description:   p.description,
		Repository:    p.repository,
		DefaultBranch: p.defaultBranch,
		Team:          p.team,
		IsActive:      p.isActive,
		Settings:      p.settings,
		CreatedAt:     p.createdAt,
		UpdatedAt:     p.updatedAt,
	}
}

// ProjectSnapshot is a read-only view of a project
type ProjectSnapshot struct {
	ID            uint
	ProjectID     ProjectID
	Name          string
	Description   string
	Repository    string
	DefaultBranch string
	Team          Team
	IsActive      bool
	Settings      map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
}