// Package api provides REST API handlers for projects
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	"github.com/guidewire-oss/fern-platform/pkg/middleware"
)

// Project Handlers

func (h *Handler) createProject(c *gin.Context) {
	var input struct {
		ID          string   `json:"id"`
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Team        string   `json:"team"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's teams from context
	userTeams := middleware.GetUserTeams(c)
	
	// If team is provided, verify user can manage it
	if input.Team != "" {
		if !middleware.IsManagerForTeam(c, input.Team) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to create projects for this team"})
			return
		}
	} else if len(userTeams) > 0 {
		// If no team specified but user has teams, use the first one they can manage
		for _, team := range userTeams {
			if middleware.IsManagerForTeam(c, team) {
				input.Team = team
				break
			}
		}
	}

	// Convert to service input, using ID as ProjectID
	serviceInput := service.CreateProjectInput{
		ProjectID:   input.ID,
		Name:        input.Name,
		Description: input.Description,
		Team:        input.Team,
	}

	project, err := h.projectService.CreateProject(serviceInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response in format expected by client
	response := map[string]interface{}{
		"id":          project.ProjectID, // Use project_id as the id field
		"name":        project.Name,
		"description": project.Description,
		"tags":        []string{}, // Empty for now since we don't store tags separately
		"team":        project.Team,
		"createdAt":   project.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) getProject(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.projectService.GetProjectByProjectID(projectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) getProjectByProjectID(c *gin.Context) {
	projectID := c.Param("projectId")
	
	project, err := h.projectService.GetProjectByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) listProjects(c *gin.Context) {
	filter := service.ListProjectsFilter{
		Search:     c.Query("search"),
		ActiveOnly: c.Query("active_only") == "true",
	}

	// Parse limit and offset
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// Filter by user's teams unless admin
	if !middleware.IsAdmin(c) {
		userTeams := middleware.GetUserTeams(c)
		if len(userTeams) > 0 {
			filter.Teams = userTeams
		}
	}

	projects, total, err := h.projectService.ListProjects(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to format expected by client with string IDs
	projectData := make([]map[string]interface{}, len(projects))
	for i, project := range projects {
		// Check if user can manage this project
		canManage := middleware.IsAdmin(c) || middleware.IsManagerForTeam(c, project.Team)
		
		projectData[i] = map[string]interface{}{
			"id":             project.ProjectID, // Use project_id as the id field
			"name":           project.Name,
			"description":    project.Description,
			"is_active":      project.IsActive,
			"repository":     project.Repository,
			"default_branch": project.DefaultBranch,
			"tags":           []string{}, // Empty for now since we don't store tags separately
			"team":           project.Team,
			"can_manage":     canManage,
			"createdAt":      project.CreatedAt,
		}
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, gin.H{
		"data":       projectData,
		"total":      total,
		"totalCount": total, // For compatibility with client expectations
	})
}

func (h *Handler) updateProject(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get existing project to check team
	existing, err := h.projectService.GetProjectByProjectID(projectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if user can manage this project
	if !middleware.IsAdmin(c) && !middleware.IsManagerForTeam(c, existing.Team) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this project"})
		return
	}

	var input service.UpdateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If changing team, verify user can manage the new team
	if input.Team != "" && input.Team != existing.Team {
		if !middleware.IsAdmin(c) && !middleware.IsManagerForTeam(c, input.Team) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to move project to this team"})
			return
		}
	}

	project, err := h.projectService.UpdateProjectByProjectID(projectId, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) deleteProject(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get existing project to check team
	existing, err := h.projectService.GetProjectByProjectID(projectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if user can manage this project
	if !middleware.IsAdmin(c) && !middleware.IsManagerForTeam(c, existing.Team) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this project"})
		return
	}

	if err := h.projectService.DeleteProjectByProjectID(projectId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (h *Handler) activateProject(c *gin.Context) {
	projectID := c.Param("projectId")

	if err := h.projectService.ActivateProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.GetProjectByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) deactivateProject(c *gin.Context) {
	projectID := c.Param("projectId")

	if err := h.projectService.DeactivateProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.GetProjectByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) getProjectStats(c *gin.Context) {
	projectID := c.Param("projectId")

	stats, err := h.projectService.GetProjectStats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}