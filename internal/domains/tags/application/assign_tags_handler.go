package application

import (
	"context"
	"fmt"

	"github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
)

// AssignTagsCommand represents the command to assign tags to a test run
type AssignTagsCommand struct {
	TestRunID uint
	TagNames  []string
}

// AssignTagsHandler handles tag assignment to test runs
type AssignTagsHandler struct {
	tagRepo domain.TagRepository
}

// NewAssignTagsHandler creates a new handler
func NewAssignTagsHandler(tagRepo domain.TagRepository) *AssignTagsHandler {
	return &AssignTagsHandler{
		tagRepo: tagRepo,
	}
}

// Handle assigns tags to a test run
func (h *AssignTagsHandler) Handle(ctx context.Context, cmd AssignTagsCommand) error {
	tagIDs := make([]domain.TagID, 0, len(cmd.TagNames))

	// Create or find tags
	for _, tagName := range cmd.TagNames {
		// Check if tag exists
		tag, err := h.tagRepo.FindByName(ctx, tagName)
		if err != nil || tag == nil {
			// Create new tag
			newTag, err := domain.NewTag(tagName)
			if err != nil {
				return err
			}

			if err := h.tagRepo.Save(ctx, newTag); err != nil {
				return err
			}

			tagIDs = append(tagIDs, newTag.ID())
		} else {
			tagIDs = append(tagIDs, tag.ID())
		}
	}

	// Assign tags to test run
	return h.tagRepo.AssignToTestRun(ctx, fmt.Sprintf("%d", cmd.TestRunID), tagIDs)
}
