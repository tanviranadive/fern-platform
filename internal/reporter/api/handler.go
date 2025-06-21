// Package api provides REST API handlers for the fern-reporter service
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/guidewire-oss/fern-platform/pkg/middleware"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
)

// Handler provides REST API handlers
type Handler struct {
	testRunService *service.TestRunService
	projectService *service.ProjectService
	tagService     *service.TagService
	authMiddleware *middleware.AuthMiddleware
	logger         *logging.Logger
}

// NewHandler creates a new API handler
func NewHandler(
	testRunService *service.TestRunService,
	projectService *service.ProjectService,
	tagService *service.TagService,
	authMiddleware *middleware.AuthMiddleware,
	logger *logging.Logger,
) *Handler {
	return &Handler{
		testRunService: testRunService,
		projectService: projectService,
		tagService:     tagService,
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// RegisterRoutes registers API routes with the Gin router
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Static file serving for web interface
	router.Static("/web", "./web")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/web/index.html")
	})

	v1 := router.Group("/api/v1")

	// Public routes (no authentication required)
	public := v1.Group("")
	{
		// Health check
		public.GET("/health", h.healthCheck)
		
		// Test runs - read operations
		public.GET("/test-runs", h.listTestRuns)
		public.GET("/test-runs/count", h.countTestRuns)
		public.GET("/test-runs/:id", h.getTestRun)
		public.GET("/test-runs/by-run-id/:runId", h.getTestRunByRunID)
		public.GET("/test-runs/stats", h.getTestRunStats)
		public.GET("/test-runs/recent", h.getRecentTestRuns)
		
		// Projects - read operations
		public.GET("/projects", h.listProjects)
		public.GET("/projects/:id", h.getProject)
		public.GET("/projects/by-project-id/:projectId", h.getProjectByProjectID)
		public.GET("/projects/stats/:projectId", h.getProjectStats)
		
		// Tags - read operations
		public.GET("/tags", h.listTags)
		public.GET("/tags/:id", h.getTag)
		public.GET("/tags/by-name/:name", h.getTagByName)
		public.GET("/tags/usage-stats", h.getTagUsageStats)
		public.GET("/tags/popular", h.getPopularTags)
	}

	// Protected routes (authentication required)
	protected := v1.Group("")
	protected.Use(h.authMiddleware.OptionalAuth()) // Use optional auth for now
	{
		// Test runs - write operations
		protected.POST("/test-runs", h.createTestRun)
		protected.PUT("/test-runs/:runId/status", h.updateTestRunStatus)
		protected.DELETE("/test-runs/:id", h.deleteTestRun)
		protected.POST("/test-runs/:id/tags", h.assignTagsToTestRun)
		
		// Projects - write operations
		protected.POST("/projects", h.createProject)
		protected.PUT("/projects/:id", h.updateProject)
		protected.DELETE("/projects/:id", h.deleteProject)
		protected.POST("/projects/:projectId/activate", h.activateProject)
		protected.POST("/projects/:projectId/deactivate", h.deactivateProject)
		
		// Tags - write operations
		protected.POST("/tags", h.createTag)
		protected.PUT("/tags/:id", h.updateTag)
		protected.DELETE("/tags/:id", h.deleteTag)
	}

	// fern-reporter compatible API endpoints
	api := router.Group("/api")
	{
		// Project endpoints compatible with fern-ginkgo-client
		api.POST("/project", h.createFernProject)
		api.GET("/project/:uuid", h.getFernProject)
		api.GET("/projects", h.listFernProjects)
		
		// Test reports endpoints
		api.POST("/reports/testrun", h.createFernTestReport)
		api.GET("/reports/testruns", h.listFernTestReports)
		api.GET("/reports/testrun/:uuid", h.getFernTestReport)
		
		// Additional endpoints that fern-ginkgo-client might expect
		api.POST("/testrun", h.createFernTestReport) // Alias for test run creation
	}
}

// healthCheck returns the service health status
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "fern-reporter",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Test Run Handlers

func (h *Handler) createTestRun(c *gin.Context) {
	var input struct {
		ID        string   `json:"id"`
		ProjectID string   `json:"projectId" binding:"required"`
		SuiteID   string   `json:"suiteId"`
		Status    string   `json:"status"`
		StartTime *time.Time `json:"startTime"`
		EndTime   *time.Time `json:"endTime,omitempty"`
		Duration  int64    `json:"duration"`
		Branch    string   `json:"branch"`
		Tags      []string `json:"tags"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service input, mapping fields correctly
	serviceInput := service.CreateTestRunInput{
		ProjectID:   input.ProjectID,
		RunID:       input.ID,
		Branch:      input.Branch,
		CommitSHA:   "",
		Environment: "test",
		Tags:        input.Tags,
	}

	testRun, err := h.testRunService.CreateTestRun(serviceInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response in format expected by client
	response := map[string]interface{}{
		"id":        testRun.RunID,
		"projectId": testRun.ProjectID,
		"suiteId":   testRun.ProjectID, // Use project ID as suite ID for now
		"status":    testRun.Status,
		"startTime": testRun.StartTime,
		"endTime":   testRun.EndTime,
		"duration":  testRun.Duration,
		"branch":    testRun.Branch,
		"tags":      []string{}, // Empty for now
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) getTestRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test run ID"})
		return
	}

	testRun, err := h.testRunService.GetTestRun(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test run not found"})
		return
	}

	c.JSON(http.StatusOK, testRun)
}

func (h *Handler) getTestRunByRunID(c *gin.Context) {
	runID := c.Param("runId")
	
	testRun, err := h.testRunService.GetTestRunByRunID(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test run not found"})
		return
	}

	c.JSON(http.StatusOK, testRun)
}

func (h *Handler) listTestRuns(c *gin.Context) {
	filter := repository.ListTestRunsFilter{
		ProjectID:   c.Query("project_id"),
		Branch:      c.Query("branch"),
		Status:      c.Query("status"),
		Environment: c.Query("environment"),
		OrderBy:     c.DefaultQuery("order_by", "start_time"),
		Order:       c.DefaultQuery("order", "DESC"),
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

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	testRuns, total, err := h.testRunService.ListTestRuns(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, gin.H{
		"data":  testRuns,
		"total": total,
	})
}

func (h *Handler) countTestRuns(c *gin.Context) {
	filter := repository.ListTestRunsFilter{
		ProjectID:   c.Query("project_id"),
		Branch:      c.Query("branch"),
		Status:      c.Query("status"),
		Environment: c.Query("environment"),
	}

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	_, total, err := h.testRunService.ListTestRuns(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
	})
}

func (h *Handler) updateTestRunStatus(c *gin.Context) {
	runID := c.Param("runId")
	
	var input struct {
		Status  string     `json:"status" binding:"required"`
		EndTime *time.Time `json:"end_time,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.testRunService.UpdateTestRunStatus(runID, input.Status, input.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	testRun, err := h.testRunService.GetTestRunByRunID(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, testRun)
}

func (h *Handler) deleteTestRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test run ID"})
		return
	}

	if err := h.testRunService.DeleteTestRun(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test run deleted successfully"})
}

func (h *Handler) getTestRunStats(c *gin.Context) {
	projectID := c.Query("project_id")
	days := 30 // default
	
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	stats, err := h.testRunService.GetTestRunStats(projectID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) getRecentTestRuns(c *gin.Context) {
	projectID := c.Query("project_id")
	limit := 10 // default
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	testRuns, err := h.testRunService.GetRecentTestRuns(projectID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, testRuns)
}

func (h *Handler) assignTagsToTestRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test run ID"})
		return
	}

	var input struct {
		TagIDs []uint `json:"tag_ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.testRunService.AssignTagsToTestRun(uint(id), input.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	testRun, err := h.testRunService.GetTestRun(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, testRun)
}

// Fern-reporter compatible API handlers

// createFernProject creates a new project compatible with fern-ginkgo-client
func (h *Handler) createFernProject(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		TeamName string `json:"team_name,omitempty"`
		Comment  string `json:"comment,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to internal service input
	projectInput := service.CreateProjectInput{
		ProjectID:   fmt.Sprintf("%s-%d", input.Name, time.Now().Unix()),
		Name:        input.Name,
		Description: input.Comment,
	}

	project, err := h.projectService.CreateProject(projectInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response compatible with fern-reporter API
	response := gin.H{
		"uuid":       project.ProjectID,
		"name":        project.Name,
		"team_name":   input.TeamName,
		"comment":     project.Description,
		"created_at":  project.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// getFernProject retrieves a project by UUID
func (h *Handler) getFernProject(c *gin.Context) {
	uuid := c.Param("uuid")
	
	project, err := h.projectService.GetProjectByProjectID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	response := gin.H{
		"uuid":       project.ProjectID,
		"name":       project.Name,
		"team_name":  "",
		"comment":    project.Description,
		"created_at": project.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// listFernProjects lists all projects in fern-reporter compatible format
func (h *Handler) listFernProjects(c *gin.Context) {
	// Use the existing project service to get all projects
	filter := service.ListProjectsFilter{
		Limit:  1000, // Get all projects
		Offset: 0,
	}

	projects, total, err := h.projectService.ListProjects(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to fern-reporter compatible format
	fernProjects := make([]gin.H, len(projects))
	for i, project := range projects {
		fernProjects[i] = gin.H{
			"uuid":       project.ProjectID,
			"name":       project.Name,
			"team_name":  "",
			"comment":    project.Description,
			"created_at": project.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": fernProjects,
		"total":    total,
	})
}

// createFernTestReport creates a test report compatible with fern-ginkgo-client
func (h *Handler) createFernTestReport(c *gin.Context) {
	// First, log the raw request body for debugging
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	
	h.logger.Info("fern-ginkgo-client raw request body", "body", string(bodyBytes))
	
	// Try to parse as generic JSON first
	var rawInput map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	
	h.logger.Info("fern-ginkgo-client parsed JSON", "data", rawInput)
	
	// Now try to parse with the actual structure fern-ginkgo-client sends
	var input struct {
		ID                 uint64 `json:"id"`
		TestProjectName    string `json:"test_project_name"`
		TestProjectID      string `json:"test_project_id"`
		TestSeed           uint64 `json:"test_seed"`
		StartTime          string `json:"start_time"`
		EndTime            string `json:"end_time"`
		GitBranch          string `json:"git_branch"`
		GitSHA             string `json:"git_sha"`
		BuildTriggerActor  string `json:"build_trigger_actor"`
		BuildURL           string `json:"build_url"`
		SuiteRuns       []struct {
			ID        uint64 `json:"id"`
			TestRunID uint64 `json:"test_run_id"`
			SuiteName string `json:"suite_name"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
			SpecRuns  []struct {
				ID              uint64 `json:"id"`
				SuiteID         uint64 `json:"suite_id"`
				SpecDescription string `json:"spec_description"`
				Status          string `json:"status"`
				Message         string `json:"message"`
				Tags            []struct {
					ID   uint64 `json:"id"`
					Name string `json:"name"`
				} `json:"tags"`
				StartTime string `json:"start_time"`
				EndTime   string `json:"end_time"`
			} `json:"spec_runs"`
		} `json:"suite_runs"`
	}
	
	if err := json.Unmarshal(bodyBytes, &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert the test run data to our internal format
	// Use test_project_id which contains the actual project UUID
	projectUUID := input.TestProjectID
	if projectUUID == "" {
		// Fallback to test_project_name if test_project_id is empty
		projectUUID = input.TestProjectName
	}
	
	if projectUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "test_project_id or test_project_name is required"})
		return
	}

	// Verify the project exists
	project, err := h.projectService.GetProjectByProjectID(projectUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("project not found: %s", projectUUID)})
		return
	}

	projectID := project.ProjectID

	// Parse start time
	var startTime time.Time
	if input.StartTime != "" {
		if parsedTime, err := time.Parse(time.RFC3339, input.StartTime); err == nil {
			startTime = parsedTime
		} else {
			startTime = time.Now()
		}
	} else {
		startTime = time.Now()
	}

	// Parse end time
	var endTime *time.Time
	if input.EndTime != "" {
		if parsedTime, err := time.Parse(time.RFC3339, input.EndTime); err == nil {
			endTime = &parsedTime
		}
	}

	// Calculate duration
	var duration int64
	if endTime != nil {
		duration = endTime.Sub(startTime).Milliseconds()
	}

	// Calculate test statistics from suite runs
	var totalTests, passedTests, failedTests, skippedTests int
	for _, suiteRun := range input.SuiteRuns {
		for _, specRun := range suiteRun.SpecRuns {
			totalTests++
			switch specRun.Status {
			case "passed":
				passedTests++
			case "failed":
				failedTests++
			case "skipped", "pending":
				skippedTests++
			}
		}
	}

	// Convert the entire input to metadata for storage
	metadata := map[string]interface{}{
		"test_seed":     input.TestSeed,
		"suite_runs":    input.SuiteRuns,
		"client_type":   "fern-ginkgo-client",
		"git_branch":    input.GitBranch,
		"git_sha":       input.GitSHA,
		"build_url":     input.BuildURL,
		"trigger_actor": input.BuildTriggerActor,
	}

	// Use git information from fern-ginkgo-client
	branch := input.GitBranch
	if branch == "" {
		branch = "main" // Default
	}
	
	commitSHA := input.GitSHA

	testRunInput := service.CreateTestRunInput{
		ProjectID:     projectID,
		RunID:         fmt.Sprintf("ginkgo-run-%d-%d", input.TestSeed, time.Now().Unix()),
		Branch:        branch,
		CommitSHA:     commitSHA,
		Environment:   "test",
		Metadata:      metadata,
		Tags:          []string{"ginkgo"},
		StartTime:     &startTime,
		EndTime:       endTime,
		Duration:      duration,
		TotalTests:    totalTests,
		PassedTests:   passedTests,
		FailedTests:   failedTests,
		SkippedTests:  skippedTests,
	}

	testRun, err := h.testRunService.CreateTestRun(testRunInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response in the format fern-ginkgo-client expects
	response := gin.H{
		"id":               testRun.ID,
		"test_project_name": projectID,
		"test_seed":        input.TestSeed,
		"start_time":       startTime.Format(time.RFC3339),
		"status":           testRun.Status,
		"created_at":       testRun.CreatedAt,
	}
	if endTime != nil {
		response["end_time"] = endTime.Format(time.RFC3339)
	}

	c.JSON(http.StatusCreated, response)
}

// listFernTestReports lists test reports with pagination
func (h *Handler) listFernTestReports(c *gin.Context) {
	filter := repository.ListTestRunsFilter{
		ProjectID: c.Query("project_uuid"),
		OrderBy:   c.DefaultQuery("order_by", "start_time"),
		Order:     c.DefaultQuery("order", "DESC"),
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

	testRuns, total, err := h.testRunService.ListTestRuns(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to fern-reporter compatible format
	reports := make([]gin.H, len(testRuns))
	for i, tr := range testRuns {
		reports[i] = gin.H{
			"uuid":         tr.RunID,
			"project_uuid": tr.ProjectID,
			"status":       tr.Status,
			"created_at":   tr.CreatedAt,
			"duration":     tr.Duration,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"total":   total,
	})
}

// getFernTestReport retrieves a specific test report by UUID
func (h *Handler) getFernTestReport(c *gin.Context) {
	uuid := c.Param("uuid")
	
	testRun, err := h.testRunService.GetTestRunByRunID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test report not found"})
		return
	}

	response := gin.H{
		"uuid":         testRun.RunID,
		"project_uuid": testRun.ProjectID,
		"status":       testRun.Status,
		"created_at":   testRun.CreatedAt,
		"duration":     testRun.Duration,
		"data":         testRun.Metadata,
	}

	c.JSON(http.StatusOK, response)
}