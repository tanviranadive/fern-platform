// Package api provides domain-based REST API handlers
package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	tagsApp "github.com/guidewire-oss/fern-platform/internal/domains/tags/application"
	tagsDomain "github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
)

// TagHandler handles tag related endpoints
type TagHandler struct {
	*BaseHandler
	tagService tagsApp.TagService
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService *tagsApp.TagService, logger *logging.Logger) *TagHandler {
	return &TagHandler{
		BaseHandler: NewBaseHandler(logger),
		tagService:  *tagService,
	}
}

// ProcessTestRunTags processes tags for all suite runs and spec runs in a test run request
func ProcessTestRunTags(c context.Context, tagService *tagsApp.TagService, req *TestRunRequest) error {
	// Process run-level tags first
	if len(req.Tags) > 0 {
		processedTags, err := processTagList(c, tagService, req.Tags)
		if err != nil {
			return err
		}
		req.Tags = processedTags
	}

	// Process suite and spec tags
	for i := range req.SuiteRuns {
		// Process suite-level tags
		if len(req.SuiteRuns[i].Tags) > 0 {
			processedTags, err := processTagList(c, tagService, req.SuiteRuns[i].Tags)
			if err != nil {
				return err
			}
			req.SuiteRuns[i].Tags = processedTags
		}

		// Process spec-level tags
		for j := range req.SuiteRuns[i].SpecRuns {
			if len(req.SuiteRuns[i].SpecRuns[j].Tags) > 0 {
				processedTags, err := processTagList(c, tagService, req.SuiteRuns[i].SpecRuns[j].Tags)
				if err != nil {
					return err
				}
				req.SuiteRuns[i].SpecRuns[j].Tags = processedTags
			}
		}
	}
	return nil
}

// processTagList processes a list of tags using the tag service
func processTagList(c context.Context, tagService *tagsApp.TagService, tags []Tag) ([]Tag, error) {
	result := make([]Tag, 0, len(tags))
	for _, t := range tags {
		// Use the tag service to get or create the tag
		domainTag, err := tagService.GetOrCreateTag(c, t.Name)
		if err != nil {
			return nil, err
		}

		// Convert domain tag ID to uint
		tagIDStr := string(domainTag.ID())
		tagID, err := strconv.ParseUint(tagIDStr, 10, 64)
		if err != nil {
			return nil, err
		}

		result = append(result, Tag{
			ID:       tagID,
			Name:     domainTag.Name(),
			Category: domainTag.Category(),
			Value:    domainTag.Value(),
		})
	}
	return result, nil
}

// createTag handles POST /api/v1/admin/tags
func (h *TagHandler) createTag(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.tagService.CreateTag(c.Request.Context(), input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to API response format with additional fields
	c.JSON(http.StatusCreated, gin.H{
		"id":          tag.ID(),
		"name":        tag.Name(),
		"category":    tag.Category(),
		"value":       tag.Value(),
		"description": input.Description,
		"color":       input.Color,
		"createdAt":   tag.CreatedAt(),
	})
}

// getTag handles GET /api/v1/tags/:id
func (h *TagHandler) getTag(c *gin.Context) {
	idStr := c.Param("id")

	tag, err := h.tagService.GetTag(c.Request.Context(), tagsDomain.TagID(idStr))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, h.convertTagToAPI(tag))
}

// getTagByName handles GET /api/v1/tags/by-name/:name
func (h *TagHandler) getTagByName(c *gin.Context) {
	name := c.Param("name")

	tag, err := h.tagService.GetTagByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, h.convertTagToAPI(tag))
}

// updateTag handles PUT /api/v1/admin/tags/:id
func (h *TagHandler) updateTag(c *gin.Context) {
	idStr := c.Param("id")

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Tags in the domain are immutable except for deletion
	// For backward compatibility, we'll return success but not actually update
	tag, err := h.tagService.GetTag(c.Request.Context(), tagsDomain.TagID(idStr))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	// Return the tag with potentially updated metadata (though not actually persisted)
	c.JSON(http.StatusOK, gin.H{
		"id":          tag.ID(),
		"name":        tag.Name(),
		"category":    tag.Category(),
		"value":       tag.Value(),
		"description": input.Description,
		"color":       input.Color,
		"updatedAt":   time.Now(),
	})
}

// deleteTag handles DELETE /api/v1/admin/tags/:id
func (h *TagHandler) deleteTag(c *gin.Context) {
	idStr := c.Param("id")

	if err := h.tagService.DeleteTag(c.Request.Context(), tagsDomain.TagID(idStr)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// listTags handles GET /api/v1/tags
func (h *TagHandler) listTags(c *gin.Context) {
	tags, err := h.tagService.ListTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to API response format
	apiTags := make([]interface{}, len(tags))
	for i, tag := range tags {
		apiTags[i] = h.convertTagToAPI(tag)
	}

	c.JSON(http.StatusOK, apiTags)
}

// getTagUsageStats handles GET /api/v1/tags/usage-stats
func (h *TagHandler) getTagUsageStats(c *gin.Context) {
	// TODO: Implement tag usage stats in domain service
	c.JSON(http.StatusOK, []gin.H{})
}

// getPopularTags handles GET /api/v1/tags/popular
func (h *TagHandler) getPopularTags(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else if l <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be greater than 0"})
			return
		}
	}

	// For now, return top N tags from all tags
	tags, err := h.tagService.ListTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Limit results
	if len(tags) > limit {
		tags = tags[:limit]
	}

	// Convert to usage format
	popularTags := make([]gin.H, len(tags))
	for i, tag := range tags {
		popularTags[i] = gin.H{
			"tag":        h.convertTagToAPI(tag),
			"usageCount": 0, // TODO: Implement usage counting
		}
	}

	c.JSON(http.StatusOK, popularTags)
}

// convertTagToAPI converts a domain tag to API response format
func (h *TagHandler) convertTagToAPI(t *tagsDomain.Tag) gin.H {
	return gin.H{
		"id":          string(t.ID()),
		"name":        t.Name(),
		"category":    t.Category(),
		"value":       t.Value(),
		"description": "", // Domain tags don't have descriptions
		"color":       "", // Domain tags don't have colors
		"createdAt":   t.CreatedAt(),
		"updatedAt":   t.CreatedAt(), // Domain tags are immutable
	}
}

// RegisterRoutes registers tag routes
func (h *TagHandler) RegisterRoutes(userGroup, adminGroup *gin.RouterGroup) {
	// User routes (read operations)
	userGroup.GET("/tags", h.listTags)
	userGroup.GET("/tags/:id", h.getTag)
	userGroup.GET("/tags/by-name/:name", h.getTagByName)
	userGroup.GET("/tags/usage-stats", h.getTagUsageStats)
	userGroup.GET("/tags/popular", h.getPopularTags)

	// Admin routes (create/update/delete)
	adminGroup.POST("/tags", h.createTag)
	adminGroup.PUT("/tags/:id", h.updateTag)
	adminGroup.DELETE("/tags/:id", h.deleteTag)
}
