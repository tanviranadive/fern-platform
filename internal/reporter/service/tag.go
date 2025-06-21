// Package service provides business logic for tags
package service

import (
	"fmt"

	"github.com/guidewire-oss/fern-platform/pkg/database"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
)

// TagService handles tag business logic
type TagService struct {
	tagRepo *repository.TagRepository
	logger  *logging.Logger
}

// NewTagService creates a new tag service
func NewTagService(tagRepo *repository.TagRepository, logger *logging.Logger) *TagService {
	return &TagService{
		tagRepo: tagRepo,
		logger:  logger,
	}
}

// CreateTagInput represents input for creating a tag
type CreateTagInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// CreateTag creates a new tag
func (s *TagService) CreateTag(input CreateTagInput) (*database.Tag, error) {
	s.logger.WithFields(map[string]interface{}{
		"tag_name": input.Name,
	}).Info("Creating tag")

	// Check if tag already exists
	existing, err := s.tagRepo.GetTagByName(input.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tag with name %s already exists", input.Name)
	}

	tag := &database.Tag{
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
	}

	if err := s.tagRepo.CreateTag(tag); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"tag_name": input.Name,
		}).WithError(err).Error("Failed to create tag")
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"tag_name": input.Name,
		"tag_id":   tag.ID,
	}).Info("Tag created successfully")

	return tag, nil
}

// GetTag retrieves a tag by ID
func (s *TagService) GetTag(id uint) (*database.Tag, error) {
	return s.tagRepo.GetTagByID(id)
}

// GetTagByName retrieves a tag by name
func (s *TagService) GetTagByName(name string) (*database.Tag, error) {
	return s.tagRepo.GetTagByName(name)
}

// UpdateTagInput represents input for updating a tag
type UpdateTagInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// UpdateTag updates an existing tag
func (s *TagService) UpdateTag(id uint, input UpdateTagInput) (*database.Tag, error) {
	tag, err := s.tagRepo.GetTagByID(id)
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Update fields if provided
	if input.Name != "" {
		// Check if new name conflicts with existing tag
		if input.Name != tag.Name {
			existing, err := s.tagRepo.GetTagByName(input.Name)
			if err == nil && existing != nil {
				return nil, fmt.Errorf("tag with name %s already exists", input.Name)
			}
		}
		tag.Name = input.Name
	}
	if input.Description != "" {
		tag.Description = input.Description
	}
	if input.Color != "" {
		tag.Color = input.Color
	}

	if err := s.tagRepo.UpdateTag(tag); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"tag_id":   tag.ID,
			"tag_name": tag.Name,
		}).WithError(err).Error("Failed to update tag")
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"tag_id":   tag.ID,
		"tag_name": tag.Name,
	}).Info("Tag updated successfully")

	return tag, nil
}

// ListTagsFilter represents filters for listing tags
type ListTagsFilter struct {
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// ListTags retrieves tags with filtering
func (s *TagService) ListTags(filter ListTagsFilter) ([]*database.Tag, int64, error) {
	return s.tagRepo.ListTags(filter.Search, filter.Limit, filter.Offset)
}

// DeleteTag deletes a tag
func (s *TagService) DeleteTag(id uint) error {
	tag, err := s.tagRepo.GetTagByID(id)
	if err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	if err := s.tagRepo.DeleteTag(id); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"tag_id":   tag.ID,
		"tag_name": tag.Name,
	}).Info("Tag deleted")

	return nil
}

// GetTagsByTestRun retrieves tags associated with a test run
func (s *TagService) GetTagsByTestRun(testRunID uint) ([]*database.Tag, error) {
	return s.tagRepo.GetTagsByTestRun(testRunID)
}

// GetTagUsageStats retrieves tag usage statistics
func (s *TagService) GetTagUsageStats() ([]*repository.TagUsage, error) {
	return s.tagRepo.GetTagUsageStats()
}

// GetPopularTags retrieves most popular tags
func (s *TagService) GetPopularTags(limit int) ([]*repository.TagUsage, error) {
	return s.tagRepo.GetPopularTags(limit)
}

// GetOrCreateTag gets a tag by name or creates it if it doesn't exist
func (s *TagService) GetOrCreateTag(name, description, color string) (*database.Tag, error) {
	return s.tagRepo.GetOrCreateTag(name, description, color)
}

// CreateMultipleTags creates multiple tags from a list of names
func (s *TagService) CreateMultipleTags(tagNames []string) ([]*database.Tag, error) {
	var tags []*database.Tag

	for _, name := range tagNames {
		tag, err := s.GetOrCreateTag(name, "", "")
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"tag_name": name,
			}).WithError(err).Warn("Failed to create tag")
			continue
		}
		tags = append(tags, tag)
	}

	s.logger.WithFields(map[string]interface{}{
		"requested_count": len(tagNames),
		"created_count":   len(tags),
	}).Info("Multiple tags processed")

	return tags, nil
}