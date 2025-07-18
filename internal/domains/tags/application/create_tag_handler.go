package application

import (
	"context"

	"github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
)

// CreateTagCommand represents the command to create a tag
type CreateTagCommand struct {
	Name string
}

// CreateTagHandler handles tag creation
type CreateTagHandler struct {
	tagRepo domain.TagRepository
}

// NewCreateTagHandler creates a new handler
func NewCreateTagHandler(tagRepo domain.TagRepository) *CreateTagHandler {
	return &CreateTagHandler{
		tagRepo: tagRepo,
	}
}

// Handle creates a new tag
func (h *CreateTagHandler) Handle(ctx context.Context, cmd CreateTagCommand) (*domain.TagSnapshot, error) {
	// Check if tag already exists
	existing, err := h.tagRepo.FindByName(ctx, cmd.Name)
	if err == nil && existing != nil {
		// Return existing tag
		snapshot := existing.ToSnapshot()
		return &snapshot, nil
	}

	// Create new tag
	tag, err := domain.NewTag(cmd.Name)
	if err != nil {
		return nil, err
	}

	// Save tag
	if err := h.tagRepo.Save(ctx, tag); err != nil {
		return nil, err
	}

	snapshot := tag.ToSnapshot()
	return &snapshot, nil
}
