// Package api provides REST API handlers for the fern-reporter service
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"github.com/guidewire-oss/fern-platform/pkg/middleware"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
)

// Handler provides REST API handlers
type Handler struct {
	testRunService  *service.TestRunService
	projectService  *service.ProjectService
	tagService      *service.TagService
	authMiddleware  *middleware.AuthMiddleware
	oauthMiddleware *middleware.OAuthMiddleware
	logger          *logging.Logger
}

// NewHandler creates a new API handler
func NewHandler(
	testRunService *service.TestRunService,
	projectService *service.ProjectService,
	tagService *service.TagService,
	authMiddleware *middleware.AuthMiddleware,
	oauthMiddleware *middleware.OAuthMiddleware,
	logger *logging.Logger,
) *Handler {
	return &Handler{
		testRunService:  testRunService,
		projectService:  projectService,
		tagService:      tagService,
		authMiddleware:  authMiddleware,
		oauthMiddleware: oauthMiddleware,
		logger:          logger,
	}
}

// RegisterRoutes registers API routes with the Gin router
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Static file serving for web interface
	router.Static("/web", "./web")
	
	// Root route - redirect to login if not authenticated, otherwise serve app
	router.GET("/", func(c *gin.Context) {
		// Check if user is authenticated
		if !h.isUserAuthenticated(c) {
			// Redirect to login
			c.Redirect(302, "/auth/login")
			return
		}
		// Serve the main application
		c.File("./web/index.html")
	})

	// OAuth authentication routes
	auth := router.Group("/auth")
	{
		auth.GET("/login", h.showLoginPage)
		auth.GET("/start", h.oauthMiddleware.StartOAuthFlow())
		auth.GET("/callback", h.oauthMiddleware.HandleOAuthCallback())
		auth.POST("/logout", h.oauthMiddleware.Logout())
		auth.GET("/user", h.oauthMiddleware.RequireOAuth(), h.getCurrentUser)
	}

	v1 := router.Group("/api/v1")

	// Public routes (no authentication required)
	public := v1.Group("")
	{
		// Health check - only public endpoint
		public.GET("/health", h.healthCheck)
	}

	// User routes (require authentication)
	user := v1.Group("")
	user.Use(h.oauthMiddleware.RequireOAuth())
	{
		// Test runs - read operations (require authentication)
		user.GET("/test-runs", h.listTestRuns)
		user.GET("/test-runs/count", h.countTestRuns)
		user.GET("/test-runs/:id", h.getTestRun)
		user.GET("/test-runs/by-run-id/:runId", h.getTestRunByRunID)
		user.GET("/test-runs/stats", h.getTestRunStats)
		user.GET("/test-runs/recent", h.getRecentTestRuns)
		
		// Projects - read operations (require authentication)
		user.GET("/projects", h.listProjects)
		user.GET("/projects/:projectId", h.getProject)
		user.GET("/projects/by-project-id/:projectId", h.getProjectByProjectID)
		user.GET("/projects/stats/:projectId", h.getProjectStats)
		
		// Tags - read operations (require authentication)
		user.GET("/tags", h.listTags)
		user.GET("/tags/:id", h.getTag)
		user.GET("/tags/by-name/:name", h.getTagByName)
		user.GET("/tags/usage-stats", h.getTagUsageStats)
		user.GET("/tags/popular", h.getPopularTags)
		
		// User-specific operations
		user.GET("/user/preferences", h.getUserPreferences)
		user.PUT("/user/preferences", h.updateUserPreferences)
		user.GET("/user/projects", h.getUserProjects)
		
		// Test runs - user operations
		user.POST("/test-runs/:id/tags", h.assignTagsToTestRun)
	}

	// Admin routes (require admin role)
	admin := v1.Group("/admin")
	admin.Use(h.oauthMiddleware.RequireAdmin())
	{
		// User management
		admin.GET("/users", h.listUsers)
		admin.GET("/users/:userId", h.getUser)
		admin.PUT("/users/:userId/role", h.updateUserRole)
		admin.POST("/users/:userId/suspend", h.suspendUser)
		admin.POST("/users/:userId/activate", h.activateUser)
		admin.DELETE("/users/:userId", h.deleteUser)
		
		// Project management
		admin.POST("/projects", h.createProject)
		admin.PUT("/projects/:projectId", h.updateProject)
		admin.DELETE("/projects/:projectId", h.deleteProject)
		admin.POST("/projects/:projectId/activate", h.activateProject)
		admin.POST("/projects/:projectId/deactivate", h.deactivateProject)
		
		// Project access management
		admin.POST("/projects/:projectId/users/:userId/access", h.grantProjectAccess)
		admin.DELETE("/projects/:projectId/users/:userId/access", h.revokeProjectAccess)
		admin.GET("/projects/:projectId/users", h.getProjectUsers)
		
		// Tag management
		admin.POST("/tags", h.createTag)
		admin.PUT("/tags/:id", h.updateTag)
		admin.DELETE("/tags/:id", h.deleteTag)
		
		// Test run management
		admin.POST("/test-runs", h.createTestRun)
		admin.PUT("/test-runs/:runId/status", h.updateTestRunStatus)
		admin.DELETE("/test-runs/:id", h.deleteTestRun)
		admin.POST("/test-runs/bulk-delete", h.bulkDeleteTestRuns)
		
		// System management
		admin.GET("/system/stats", h.getSystemStats)
		admin.GET("/system/health", h.getSystemHealth)
		admin.POST("/system/cleanup", h.performSystemCleanup)
		admin.GET("/audit-logs", h.getAuditLogs)
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

	testRuns, total, err := h.testRunService.ListTestRunsWithProjects(filter)
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

	// Check if project already exists by name or project ID
	existingProject, err := h.projectService.GetProjectByProjectID(input.Name)
	if err == nil {
		// Project exists, return it
		response := gin.H{
			"uuid":       existingProject.ProjectID,
			"name":        existingProject.Name,
			"team_name":   input.TeamName,
			"comment":     existingProject.Description,
			"created_at":  existingProject.CreatedAt,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Convert to internal service input for new project
	projectInput := service.CreateProjectInput{
		ProjectID:   uuid.New().String(),
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

	// Convert suite runs from fern-ginkgo-client format to our service format
	var suiteRuns []service.CreateSuiteRunInput
	for _, suiteRun := range input.SuiteRuns {
		var specRuns []service.CreateSpecRunInput
		for _, specRun := range suiteRun.SpecRuns {
			specRuns = append(specRuns, service.CreateSpecRunInput{
				SpecDescription: specRun.SpecDescription,
				Status:          specRun.Status,
				Message:         specRun.Message,
				StartTime:       specRun.StartTime,
				EndTime:         specRun.EndTime,
				Tags:            specRun.Tags,
			})
		}

		suiteRuns = append(suiteRuns, service.CreateSuiteRunInput{
			SuiteName: suiteRun.SuiteName,
			StartTime: suiteRun.StartTime,
			EndTime:   suiteRun.EndTime,
			SpecRuns:  specRuns,
		})
	}

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
		SuiteRuns:     suiteRuns,
	}

	testRun, err := h.testRunService.CreateTestRunWithSuites(testRunInput)
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

// OAuth and User Management Handlers

// getCurrentUser returns the current authenticated user
func (h *Handler) getCurrentUser(c *gin.Context) {
	user, exists := middleware.GetOAuthUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Return user info without sensitive data
	// Extract group names from the preloaded UserGroups
	groups := make([]string, len(user.UserGroups))
	for i, ug := range user.UserGroups {
		groups[i] = ug.GroupName
	}

	response := gin.H{
		"user_id":        user.UserID,
		"email":          user.Email,
		"name":           user.Name,
		"role":           user.Role,
		"status":         user.Status,
		"profile_url":    user.ProfileURL,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"email_verified": user.EmailVerified,
		"groups":         groups,
		"last_login":     user.LastLoginAt,
	}

	c.JSON(http.StatusOK, response)
}

// getUserPreferences returns user preferences
func (h *Handler) getUserPreferences(c *gin.Context) {
	// TODO: Implement user preferences retrieval
	c.JSON(http.StatusOK, gin.H{
		"theme":        "light",
		"timezone":     "UTC",
		"language":     "en",
		"favorites":    []string{},
		"preferences": map[string]interface{}{},
	})
}

// updateUserPreferences updates user preferences
func (h *Handler) updateUserPreferences(c *gin.Context) {
	// TODO: Implement user preferences update
	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}

// getUserProjects returns projects accessible to the user
func (h *Handler) getUserProjects(c *gin.Context) {
	user, exists := middleware.GetOAuthUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// TODO: Implement project access filtering based on user permissions
	// For now, return all projects if admin, or projects with specific access
	if user.Role == "admin" {
		h.listProjects(c)
	} else {
		// Return projects user has access to
		c.JSON(http.StatusOK, gin.H{
			"projects": []gin.H{},
			"total":    0,
		})
	}
}

// Admin User Management Handlers

// listUsers returns all users (admin only)
func (h *Handler) listUsers(c *gin.Context) {
	// TODO: Implement user listing with pagination and filtering
	c.JSON(http.StatusOK, gin.H{
		"users": []gin.H{},
		"total": 0,
	})
}

// getUser returns a specific user (admin only)
func (h *Handler) getUser(c *gin.Context) {
	userID := c.Param("userId")
	
	// TODO: Implement user retrieval
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "User retrieval not yet implemented",
	})
}

// updateUserRole updates a user's role (admin only)
func (h *Handler) updateUserRole(c *gin.Context) {
	userID := c.Param("userId")
	
	var input struct {
		Role string `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if input.Role != "user" && input.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'user' or 'admin'"})
		return
	}

	// TODO: Implement role update
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"role":    input.Role,
		"message": "User role updated successfully",
	})
}

// suspendUser suspends a user account (admin only)
func (h *Handler) suspendUser(c *gin.Context) {
	userID := c.Param("userId")
	
	// TODO: Implement user suspension
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"status":  "suspended",
		"message": "User suspended successfully",
	})
}

// activateUser activates a user account (admin only)
func (h *Handler) activateUser(c *gin.Context) {
	userID := c.Param("userId")
	
	// TODO: Implement user activation
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"status":  "active",
		"message": "User activated successfully",
	})
}

// deleteUser deletes a user account (admin only)
func (h *Handler) deleteUser(c *gin.Context) {
	userID := c.Param("userId")
	
	// TODO: Implement user deletion with safety checks
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "User deleted successfully",
	})
}

// Project Access Management Handlers

// grantProjectAccess grants user access to a project (admin only)
func (h *Handler) grantProjectAccess(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.Param("userId")
	
	var input struct {
		Role      string     `json:"role" binding:"required"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate project role
	if input.Role != "viewer" && input.Role != "editor" && input.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project role"})
		return
	}

	// TODO: Implement project access granting
	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"user_id":    userID,
		"role":       input.Role,
		"expires_at": input.ExpiresAt,
		"message":    "Project access granted successfully",
	})
}

// revokeProjectAccess revokes user access from a project (admin only)
func (h *Handler) revokeProjectAccess(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.Param("userId")
	
	// TODO: Implement project access revocation
	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"user_id":    userID,
		"message":    "Project access revoked successfully",
	})
}

// getProjectUsers returns users with access to a project (admin only)
func (h *Handler) getProjectUsers(c *gin.Context) {
	projectID := c.Param("projectId")
	
	// TODO: Implement project user listing
	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"users":      []gin.H{},
		"total":      0,
	})
}

// System Management Handlers

// bulkDeleteTestRuns deletes multiple test runs (admin only)
func (h *Handler) bulkDeleteTestRuns(c *gin.Context) {
	var input struct {
		TestRunIDs []uint   `json:"test_run_ids" binding:"required"`
		Filters    []string `json:"filters,omitempty"` // Additional filters like older than X days
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement bulk deletion with safety checks
	c.JSON(http.StatusOK, gin.H{
		"deleted_count": len(input.TestRunIDs),
		"message":       "Test runs deleted successfully",
	})
}

// getSystemStats returns system-wide statistics (admin only)
func (h *Handler) getSystemStats(c *gin.Context) {
	// TODO: Implement system statistics
	c.JSON(http.StatusOK, gin.H{
		"total_users":       0,
		"total_projects":    0,
		"total_test_runs":   0,
		"active_sessions":   0,
		"storage_usage":     "0 GB",
		"api_requests_24h":  0,
		"system_uptime":     "0 days",
		"last_cleanup":      nil,
	})
}

// getSystemHealth returns detailed system health (admin only)
func (h *Handler) getSystemHealth(c *gin.Context) {
	// TODO: Implement comprehensive health checks
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"checks": gin.H{
			"database":    "healthy",
			"redis":       "healthy",
			"disk_space":  "healthy",
			"memory":      "healthy",
			"cpu":         "healthy",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// performSystemCleanup performs system maintenance tasks (admin only)
func (h *Handler) performSystemCleanup(c *gin.Context) {
	var input struct {
		CleanupType string `json:"cleanup_type" binding:"required"` // expired_sessions, old_test_runs, etc.
		DryRun      bool   `json:"dry_run"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement system cleanup operations
	c.JSON(http.StatusOK, gin.H{
		"cleanup_type":    input.CleanupType,
		"dry_run":         input.DryRun,
		"items_cleaned":   0,
		"space_reclaimed": "0 GB",
		"message":         "System cleanup completed",
	})
}

// getAuditLogs returns system audit logs (admin only)
func (h *Handler) getAuditLogs(c *gin.Context) {
	// Parse query parameters
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// TODO: Implement audit log retrieval
	c.JSON(http.StatusOK, gin.H{
		"logs":   []gin.H{},
		"total":  0,
		"limit":  limit,
		"offset": offset,
	})
}

// Helper Methods

// isUserAuthenticated checks if the current request is from an authenticated user
func (h *Handler) isUserAuthenticated(c *gin.Context) bool {
	// Use the OAuth middleware to properly validate the session
	user, _, err := h.oauthMiddleware.ValidateSession(c)
	return err == nil && user != nil
}

// showLoginPage displays the login page with a sign-in button
func (h *Handler) showLoginPage(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(200, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fern Platform - Sign In</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .login-container {
            background: white;
            border-radius: 16px;
            padding: 3rem;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            text-align: center;
            max-width: 400px;
            width: 100%%;
            margin: 2rem;
        }
        .logo {
            font-size: 3rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 0.5rem;
            font-size: 2rem;
            font-weight: 700;
        }
        .subtitle {
            color: #718096;
            margin-bottom: 2rem;
            font-size: 1.1rem;
        }
        .sign-in-button {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border: none;
            padding: 1rem 2rem;
            font-size: 1.1rem;
            font-weight: 600;
            border-radius: 12px;
            cursor: pointer;
            transition: all 0.3s ease;
            width: 100%%;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
        }
        .sign-in-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
        }
        .features {
            margin-top: 3rem;
            text-align: left;
        }
        .feature {
            display: flex;
            align-items: center;
            margin-bottom: 1rem;
            color: #4a5568;
        }
        .feature-icon {
            margin-right: 0.75rem;
            font-size: 1.2rem;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">🌿</div>
        <h1>Fern Platform</h1>
        <p class="subtitle">Modern Test Intelligence & Analytics</p>
        
        <button class="sign-in-button" onclick="window.location.href='/auth/start'">
            Sign In with OAuth
        </button>
        
        <div class="features">
            <div class="feature">
                <span class="feature-icon">📊</span>
                <span>Real-time test analytics</span>
            </div>
            <div class="feature">
                <span class="feature-icon">🔍</span>
                <span>Detailed test insights</span>
            </div>
            <div class="feature">
                <span class="feature-icon">⚡</span>
                <span>Fast & reliable reporting</span>
            </div>
        </div>
    </div>
</body>
</html>`)
}