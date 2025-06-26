package graphql

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/model"
	"github.com/guidewire-oss/fern-platform/internal/reporter/repository"
	"github.com/guidewire-oss/fern-platform/internal/reporter/service"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// Implement Query resolvers

// CurrentUser returns the current authenticated user
func (r *queryResolver) CurrentUser_impl(ctx context.Context) (*model.User, error) {
	user, err := getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:          user.UserID,
		UserID:      user.UserID,
		Email:       user.Email,
		Name:        user.Name,
		FirstName:   convertStringPtr(user.FirstName),
		LastName:    convertStringPtr(user.LastName),
		Role:        user.Role,
		ProfileURL:  convertStringPtr(user.ProfileURL),
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
	}, nil
}

// Health returns the service health status
func (r *queryResolver) Health_impl(ctx context.Context) (*model.HealthStatus, error) {
	return &model.HealthStatus{
		Status:    "healthy",
		Service:   "fern-platform",
		Timestamp: time.Now(),
		Version:   convertStringPtr("1.0.0"), // TODO: Get from build info
	}, nil
}

// DashboardSummary returns aggregated dashboard data
func (r *queryResolver) DashboardSummary_impl(ctx context.Context) (*model.DashboardSummary, error) {
	// Get project count - more efficient to just get counts
	_, totalProjects, err := r.projectService.ListProjects(service.ListProjectsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get project count: %w", err)
	}
	
	// Get active project count
	_, activeProjects, err := r.projectService.ListProjects(service.ListProjectsFilter{ActiveOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get active project count: %w", err)
	}

	// Get test run stats
	testRuns, _, err := r.testRunService.ListTestRuns(repository.ListTestRunsFilter{
		StartTime: func() *time.Time { t := time.Now().AddDate(0, 0, -30); return &t }(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get test runs: %w", err)
	}

	// Calculate stats
	totalTests := 0
	passedTests := 0
	totalDuration := 0

	for _, run := range testRuns {
		totalTests += run.TotalTests
		passedTests += run.PassedTests
		totalDuration += int(run.Duration)
	}

	overallPassRate := float64(0)
	avgDuration := 0
	if totalTests > 0 {
		overallPassRate = float64(passedTests) / float64(totalTests) * 100
	}
	if len(testRuns) > 0 {
		avgDuration = totalDuration / len(testRuns)
	}

	return &model.DashboardSummary{
		Health: &model.HealthStatus{
			Status:    "healthy",
			Service:   "fern-platform",
			Timestamp: time.Now(),
			Version:   convertStringPtr("1.0.0"),
		},
		ProjectCount:         int(totalProjects),
		ActiveProjectCount:   int(activeProjects),
		TotalTestRuns:        len(testRuns),
		RecentTestRuns:       len(testRuns),
		OverallPassRate:      overallPassRate,
		TotalTestsExecuted:   totalTests,
		AverageTestDuration:  avgDuration,
	}, nil
}

// TreemapData returns hierarchical data for treemap visualization
func (r *queryResolver) TreemapData_impl(ctx context.Context, projectID *string, days *int) (*model.TreemapData, error) {
	daysFilter := 7
	if days != nil {
		daysFilter = *days
	}

	startTime := time.Now().AddDate(0, 0, -daysFilter)
	
	filter := repository.ListTestRunsFilter{
		StartTime: &startTime,
	}
	if projectID != nil {
		filter.ProjectID = *projectID
	}

	// Get test runs
	testRuns, _, err := r.testRunService.ListTestRuns(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get test runs: %w", err)
	}

	// Get all projects
	projects, _, err := r.projectService.ListProjects(service.ListProjectsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	// Build project map
	projectMap := make(map[string]*database.ProjectDetails)
	for i := range projects {
		projectMap[projects[i].ProjectID] = projects[i]
	}

	// Group test runs by project
	projectNodes := make(map[string]*model.ProjectTreemapNode)
	totalDuration := 0
	totalTests := 0
	totalPassed := 0

	for _, run := range testRuns {
		project, exists := projectMap[run.ProjectID]
		if !exists {
			continue
		}

		node, exists := projectNodes[run.ProjectID]
		if !exists {
			node = &model.ProjectTreemapNode{
				Project: convertProject(project),
				Suites:  []*model.SuiteTreemapNode{},
			}
			projectNodes[run.ProjectID] = node
		}

		// Add to totals
		node.TotalDuration += int(run.Duration)
		node.TotalTests += run.TotalTests
		node.PassedTests += run.PassedTests
		node.FailedTests += run.FailedTests

		totalDuration += int(run.Duration)
		totalTests += run.TotalTests
		totalPassed += run.PassedTests

		// TODO: Load suite runs for each test run
		// This would need to be optimized with DataLoader
	}

	// Calculate pass rates
	for _, node := range projectNodes {
		if node.TotalTests > 0 {
			node.PassRate = float64(node.PassedTests) / float64(node.TotalTests)
		}
	}

	// Convert map to slice
	projectNodeList := make([]*model.ProjectTreemapNode, 0, len(projectNodes))
	for _, node := range projectNodes {
		projectNodeList = append(projectNodeList, node)
	}

	overallPassRate := float64(0)
	if totalTests > 0 {
		overallPassRate = float64(totalPassed) / float64(totalTests)
	}

	return &model.TreemapData{
		Projects:        projectNodeList,
		TotalDuration:   totalDuration,
		TotalTests:      totalTests,
		OverallPassRate: overallPassRate,
	}, nil
}

// TestRun returns a single test run by ID
func (r *queryResolver) TestRun_impl(ctx context.Context, id string) (*model.TestRun, error) {
	loaders := getLoaders(ctx)
	
	testRun, err := loaders.TestRunByID.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	return convertTestRun(testRun), nil
}

// Projects returns paginated projects
func (r *queryResolver) Projects_impl(ctx context.Context, filter *model.ProjectFilter, first *int, after *string) (*model.ProjectConnection, error) {
	// Build filter
	repoFilter := service.ListProjectsFilter{}
	if filter != nil {
		if filter.Search != nil {
			repoFilter.Search = *filter.Search
		}
		if filter.ActiveOnly != nil && *filter.ActiveOnly {
			repoFilter.ActiveOnly = true
		}
	}

	// Apply pagination
	pageSize := 20
	if first != nil && *first > 0 && *first <= 100 {
		pageSize = *first
	}
	
	offset := 0
	if after != nil && *after != "" {
		// Simple cursor: just the index
		if idx, err := strconv.Atoi(*after); err == nil && idx >= 0 {
			offset = idx + 1
		}
	}
	
	// Set pagination params
	repoFilter.Limit = pageSize
	repoFilter.Offset = offset
	
	// Get projects with pagination
	projects, totalCount, err := r.projectService.ListProjects(repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	
	hasMore := offset + len(projects) < int(totalCount)

	// Build edges
	edges := make([]*model.ProjectEdge, len(projects))
	for i, project := range projects {
		edges[i] = &model.ProjectEdge{
			Node:   convertProject(project),
			Cursor: fmt.Sprintf("%d", offset+i), // Simple cursor
		}
	}

	// Build page info
	pageInfo := &model.PageInfo{
		HasNextPage: hasMore,
		HasPreviousPage: offset > 0,
	}
	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.ProjectConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: int(totalCount),
	}, nil
}

// Helper functions to convert between repository and GraphQL models

func convertProject(p *database.ProjectDetails) *model.Project {
	return &model.Project{
		ID:            fmt.Sprintf("%d", p.ID),
		ProjectID:     p.ProjectID,
		Name:          p.Name,
		Description:   convertStringPtr(p.Description),
		Repository:    convertStringPtr(p.Repository),
		DefaultBranch: p.DefaultBranch,
		Settings:      nil, // TODO: Parse JSON settings
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func convertTestRun(tr *database.TestRun) *model.TestRun {
	return &model.TestRun{
		ID:            fmt.Sprintf("%d", tr.ID),
		ProjectID:     tr.ProjectID,
		RunID:         tr.RunID,
		Branch:        convertStringPtr(tr.Branch),
		CommitSha:     convertStringPtr(tr.CommitSHA),
		Status:        tr.Status,
		StartTime:     tr.StartTime,
		EndTime:       tr.EndTime,
		TotalTests:    tr.TotalTests,
		PassedTests:   tr.PassedTests,
		FailedTests:   tr.FailedTests,
		SkippedTests:  tr.SkippedTests,
		Duration:      int(tr.Duration),
		Environment:   convertStringPtr(tr.Environment),
		Metadata:      tr.Metadata,
		CreatedAt:     tr.CreatedAt,
		UpdatedAt:     tr.UpdatedAt,
	}
}

// Mutation implementations

// CreateProject creates a new project
func (r *mutationResolver) CreateProject_impl(ctx context.Context, input model.CreateProjectInput) (*model.Project, error) {
	// Check user permissions
	user, err := getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	if user.Role != string(database.RoleAdmin) {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Create project
	projectInput := service.CreateProjectInput{
		ProjectID:     input.ProjectID,
		Name:          input.Name,
		Description:   convertPtrString(input.Description),
		Repository:    convertPtrString(input.Repository),
		DefaultBranch: convertPtrString(input.DefaultBranch),
		Settings:      input.Settings,
	}

	project, err := r.projectService.CreateProject(projectInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return convertProject(project), nil
}