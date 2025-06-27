package interfaces

import (
	"fmt"

	"github.com/guidewire-oss/fern-platform/internal/domains/tags/application"
	"github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
)

// TagServiceAdapter adapts the new domain to the existing TagService interface
type TagServiceAdapter struct {
	createTagHandler  *application.CreateTagHandler
	assignTagsHandler *application.AssignTagsHandler
	tagRepo          domain.TagRepository
	legacyRepo       *repository.TagRepository
	logger           *logging.Logger
}

// NewTagServiceAdapter creates a new adapter
func NewTagServiceAdapter(
	createHandler *application.CreateTagHandler,
	assignHandler *application.AssignTagsHandler,
	domainRepo domain.TagRepository,
	legacyRepo *repository.TagRepository,
	logger *logging.Logger,
) *TagServiceAdapter {
	return &TagServiceAdapter{
		createTagHandler:  createHandler,
		assignTagsHandler: assignHandler,
		tagRepo:          domainRepo,
		legacyRepo:       legacyRepo,
		logger:           logger,
	}
}

// CreateTag creates a new tag
func (a *TagServiceAdapter) CreateTag(input service.CreateTagInput) (*database.Tag, error) {
	// Check if tag already exists
	existing, err := a.legacyRepo.GetTagByName(input.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tag with name %s already exists", input.Name)
	}

	tag := &database.Tag{
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
	}

	if err := a.legacyRepo.CreateTag(tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// GetTag retrieves a tag by ID
func (a *TagServiceAdapter) GetTag(id uint) (*database.Tag, error) {
	return a.legacyRepo.GetTagByID(id)
}

// GetTagByName retrieves a tag by name
func (a *TagServiceAdapter) GetTagByName(name string) (*database.Tag, error) {
	return a.legacyRepo.GetTagByName(name)
}

// UpdateTag updates an existing tag
func (a *TagServiceAdapter) UpdateTag(id uint, input service.UpdateTagInput) (*database.Tag, error) {
	tag, err := a.legacyRepo.GetTagByID(id)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Update fields if provided
	if input.Name != "" {
		tag.Name = input.Name
	}
	if input.Description != "" {
		tag.Description = input.Description
	}
	if input.Color != "" {
		tag.Color = input.Color
	}

	if err := a.legacyRepo.UpdateTag(tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

// ListTags retrieves all tags
func (a *TagServiceAdapter) ListTags(filter service.ListTagsFilter) ([]*database.Tag, int64, error) {
	return a.legacyRepo.ListTags(filter.Search, filter.Limit, filter.Offset)
}

// DeleteTag deletes a tag
func (a *TagServiceAdapter) DeleteTag(id uint) error {
	return a.legacyRepo.DeleteTag(id)
}

// GetTagsByTestRun retrieves tags associated with a test run
func (a *TagServiceAdapter) GetTagsByTestRun(testRunID uint) ([]*database.Tag, error) {
	return a.legacyRepo.GetTagsByTestRun(testRunID)
}

// GetTagUsageStats retrieves tag usage statistics
func (a *TagServiceAdapter) GetTagUsageStats() ([]*repository.TagUsage, error) {
	return a.legacyRepo.GetTagUsageStats()
}

// GetPopularTags retrieves most popular tags
func (a *TagServiceAdapter) GetPopularTags(limit int) ([]*repository.TagUsage, error) {
	return a.legacyRepo.GetPopularTags(limit)
}

// GetOrCreateTag gets a tag by name or creates it if it doesn't exist
func (a *TagServiceAdapter) GetOrCreateTag(name, description, color string) (*database.Tag, error) {
	return a.legacyRepo.GetOrCreateTag(name, description, color)
}

// CreateMultipleTags creates multiple tags from a list of names
func (a *TagServiceAdapter) CreateMultipleTags(tagNames []string) ([]*database.Tag, error) {
	var tags []*database.Tag

	for _, name := range tagNames {
		tag, err := a.GetOrCreateTag(name, "", "")
		if err != nil {
			a.logger.WithFields(map[string]interface{}{
				"tag_name": name,
			}).WithError(err).Warn("Failed to create tag")
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
}