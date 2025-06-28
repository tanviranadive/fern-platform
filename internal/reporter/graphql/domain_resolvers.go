package graphql

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/model"
	authDomain "github.com/guidewire-oss/fern-platform/internal/domains/auth/domain"
	testingDomain "github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	projectsDomain "github.com/guidewire-oss/fern-platform/internal/domains/projects/domain"
	projectsApp "github.com/guidewire-oss/fern-platform/internal/domains/projects/application"
	tagsDomain "github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
)

// convertTestRunToGraphQL converts a domain test run to GraphQL model
func (r *Resolver) convertTestRunToGraphQL(testRun *testingDomain.TestRun) *model.TestRun {
	return &model.TestRun{
		ID:           strconv.FormatUint(uint64(testRun.ID), 10),
		RunID:        testRun.RunID,
		ProjectID:    testRun.ProjectID,
		Branch:       convertStringPtr(testRun.Branch),
		CommitSha:    convertStringPtr(testRun.GitCommit),
		Status:       testRun.Status,
		StartTime:    testRun.StartTime,
		EndTime:      testRun.EndTime,
		TotalTests:   testRun.TotalTests,
		PassedTests:  testRun.PassedTests,
		FailedTests:  testRun.FailedTests,
		SkippedTests: testRun.SkippedTests,
		Duration:     int(testRun.Duration.Milliseconds()),
		Environment:  convertStringPtr(testRun.Environment),
		CreatedAt:    testRun.StartTime, // Use StartTime as CreatedAt
		UpdatedAt:    testRun.StartTime, // Use StartTime as UpdatedAt
	}
}

// RecentTestRuns_domain retrieves recent test runs using domain service
func (r *queryResolver) RecentTestRuns_domain(ctx context.Context, projectID *string, limit *int) ([]*model.TestRun, error) {
	limitVal := 10
	if limit != nil {
		limitVal = *limit
	}

	var testRuns []*testingDomain.TestRun
	var err error
	
	if projectID != nil {
		// Get test runs for specific project
		testRuns, err = r.testingService.GetProjectTestRuns(ctx, *projectID, limitVal)
	} else {
		// Get recent test runs across all projects
		testRuns, err = r.testingService.GetRecentTestRuns(ctx, limitVal)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get recent test runs: %w", err)
	}

	// Convert to GraphQL models
	result := make([]*model.TestRun, len(testRuns))
	for i, tr := range testRuns {
		result[i] = r.convertTestRunToGraphQL(tr)
	}

	return result, nil
}

// convertProjectToGraphQL converts a domain project to GraphQL model
func (r *Resolver) convertProjectToGraphQL(project *projectsDomain.Project) *model.Project {
	snapshot := project.ToSnapshot()
	return &model.Project{
		ID:            strconv.FormatUint(uint64(snapshot.ID), 10),
		ProjectID:     string(snapshot.ProjectID),
		Name:          snapshot.Name,
		Description:   convertStringPtr(snapshot.Description),
		Repository:    convertStringPtr(snapshot.Repository),
		DefaultBranch: snapshot.DefaultBranch,
		IsActive:      snapshot.IsActive,
		Team:          convertStringPtr(string(snapshot.Team)),
		CreatedAt:     snapshot.CreatedAt,
		UpdatedAt:     snapshot.UpdatedAt,
	}
}

// convertTagToGraphQL converts a domain tag to GraphQL model
func (r *Resolver) convertTagToGraphQL(tag *tagsDomain.Tag) *model.Tag {
	// Convert TagID string to numeric ID for GraphQL
	// For now, use a simple hash
	id := "0"
	if tagID := tag.ID(); tagID != "" {
		// Simple conversion - in production you'd want a proper mapping
		id = fmt.Sprintf("%d", len(string(tagID)))
	}
	
	return &model.Tag{
		ID:          id,
		Name:        tag.Name(),
		Description: nil, // Not available in domain model
		Color:       nil, // Not available in domain model
		CreatedAt:   time.Now(), // TODO: Add timestamps to domain model
		UpdatedAt:   time.Now(), // TODO: Add timestamps to domain model
	}
}

// Helper function to convert duration to int pointer (milliseconds)
func convertDurationPtr(d time.Duration) *int {
	ms := int(d.Milliseconds())
	return &ms
}

// GetTestRun implementation using domain service
func (r *queryResolver) GetTestRun_domain(ctx context.Context, id string) (*model.TestRun, error) {
	// Try to parse as uint ID
	if numID, err := strconv.ParseUint(id, 10, 32); err == nil {
		testRun, err := r.testingService.GetTestRun(ctx, uint(numID))
		if err != nil {
			return nil, err
		}
		return r.convertTestRunToGraphQL(testRun), nil
	}

	// Otherwise return error - GetTestRunByRunID not implemented
	return nil, fmt.Errorf("test run not found")
}

// GetProject implementation using domain service
func (r *queryResolver) GetProject_domain(ctx context.Context, projectID string) (*model.Project, error) {
	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(projectID))
	if err != nil {
		return nil, err
	}
	return r.convertProjectToGraphQL(project), nil
}

// ListProjects implementation using domain service
func (r *queryResolver) ListProjects_domain(ctx context.Context, limit *int, offset *int, team *string) ([]*model.Project, error) {
	// Default pagination
	pageLimit := 50
	pageOffset := 0
	if limit != nil {
		pageLimit = *limit
	}
	if offset != nil {
		pageOffset = *offset
	}

	// Get projects based on team filter
	var projects []*projectsDomain.Project
	var err error

	// For now just list all projects - team filtering not implemented
	projects, _, err = r.projectService.ListProjects(ctx, pageLimit, pageOffset)

	if err != nil {
		return nil, err
	}

	// Convert to GraphQL models
	result := make([]*model.Project, len(projects))
	for i, project := range projects {
		result[i] = r.convertProjectToGraphQL(project)
	}

	return result, nil
}

// ListTags implementation using domain service
func (r *queryResolver) ListTags_domain(ctx context.Context) ([]*model.Tag, error) {
	// GetAllTags not implemented - use ListTags
	tags, err := r.tagService.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to GraphQL models
	result := make([]*model.Tag, len(tags))
	for i, tag := range tags {
		result[i] = r.convertTagToGraphQL(tag)
	}

	return result, nil
}

// CreateProject implementation using domain service
func (r *mutationResolver) CreateProject_domain(ctx context.Context, input model.CreateProjectInput) (*model.Project, error) {
	// Get current user for team assignment
	user, err := getCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Check if user has permission to create projects
	// Only admins and managers can create projects
	roleGroups := getRoleGroupNamesFromContext(ctx)
	canCreate := user.Role == authDomain.RoleAdmin || hasManagerRole(user, roleGroups)
	
	if !canCreate {
		return nil, fmt.Errorf("insufficient permissions to create project")
	}

	// Generate project ID if not provided
	projectID := input.ProjectID
	if projectID == "" {
		projectID = uuid.New().String()
	}

	// Determine team from user groups
	team := projectsDomain.Team("default") // Default team
	if len(user.Groups) > 0 {
		team = projectsDomain.Team(user.Groups[0].GroupName)
	}

	// Create project with creator user ID
	project, err := r.projectService.CreateProject(ctx, projectsDomain.ProjectID(projectID), input.Name, team, user.UserID)
	if err != nil {
		return nil, err
	}

	// Update optional fields
	if input.Description != nil {
		project.UpdateDescription(*input.Description)
	}
	if input.Repository != nil {
		project.UpdateRepository(*input.Repository)
	}
	if input.DefaultBranch != nil {
		project.UpdateDefaultBranch(*input.DefaultBranch)
	}

	// Prepare update request
	updateReq := projectsApp.UpdateProjectRequest{}
	if input.Description != nil {
		updateReq.Description = input.Description
	}
	if input.Repository != nil {
		updateReq.Repository = input.Repository
	}
	if input.DefaultBranch != nil {
		updateReq.DefaultBranch = input.DefaultBranch
	}

	// Save updates if any fields were set
	if updateReq.Description != nil || updateReq.Repository != nil || updateReq.DefaultBranch != nil {
		if err := r.projectService.UpdateProject(ctx, projectsDomain.ProjectID(input.ProjectID), updateReq); err != nil {
			return nil, err
		}
		
		// Fetch updated project
		project, err = r.projectService.GetProject(ctx, projectsDomain.ProjectID(input.ProjectID))
		if err != nil {
			return nil, err
		}
	}

	return r.convertProjectToGraphQL(project), nil
}

// CreateTag implementation using domain service
func (r *mutationResolver) CreateTag_domain(ctx context.Context, input model.CreateTagInput) (*model.Tag, error) {
	// CreateTag only takes a name parameter
	createdTag, err := r.tagService.CreateTag(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	return r.convertTagToGraphQL(createdTag), nil
}

// Helper to get string value from pointer
func getStringValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// UpdateProject implementation using domain service
func (r *mutationResolver) UpdateProject_domain(ctx context.Context, id string, input model.UpdateProjectInput) (*model.Project, error) {
	// Check user permissions
	user, err := getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// For domain service, we need to get the project by its ProjectID, not database ID
	// First get the current project to get its ProjectID
	// Since we don't have a way to get by database ID in domain service, 
	// we'll need to list projects and find the one with matching ID
	// This is a temporary workaround - ideally domain service should support GetByID
	projects, _, err := r.projectService.ListProjects(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var existingProject *projectsDomain.Project
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	for _, p := range projects {
		if p.ID() == uint(idUint) {
			existingProject = p
			break
		}
	}

	if existingProject == nil {
		return nil, fmt.Errorf("project not found")
	}

	snapshot := existingProject.ToSnapshot()

	// Check if user can update this project
	canUpdate := false
	if user.Role == authDomain.RoleAdmin {
		canUpdate = true
	} else {
		// Get role group names from context
		roleGroups := getRoleGroupNamesFromContext(ctx)

		// Check if user has team + manager group combination
		if snapshot.Team != "" {
			hasTeamGroup := false
			hasManagerGroup := false

			for _, group := range user.Groups {
				groupName := strings.TrimPrefix(group.GroupName, "/")
				if groupName == string(snapshot.Team) {
					hasTeamGroup = true
				}
				if groupName == roleGroups.ManagerGroup {
					hasManagerGroup = true
				}
			}

			// If user is in both team and manager groups, they can update
			if hasTeamGroup && hasManagerGroup {
				canUpdate = true
			}
		}

		// If not via groups, check scopes
		if !canUpdate {
			// Check scopes for update permission on this project
			requiredScopes := []string{
				fmt.Sprintf("project:write:%s", snapshot.ProjectID),
				fmt.Sprintf("project:*:%s", snapshot.ProjectID),
			}

			// If project has a team, also check team-based scopes
			if snapshot.Team != "" {
				requiredScopes = append(requiredScopes,
					fmt.Sprintf("project:write:%s:*", snapshot.Team),
					fmt.Sprintf("project:*:%s:*", snapshot.Team),
				)
			}

			scopes := getUserScopesFromContext(ctx)
			for _, scope := range scopes {
				for _, required := range requiredScopes {
					if matchScope(scope, required) {
						canUpdate = true
						break
					}
				}
				if canUpdate {
					break
				}
			}
		}
	}

	if !canUpdate {
		return nil, fmt.Errorf("insufficient permissions to update project")
	}

	// Prepare update request
	updateReq := projectsApp.UpdateProjectRequest{}
	if input.Name != nil {
		updateReq.Name = input.Name
	}
	if input.Description != nil {
		updateReq.Description = input.Description
	}
	if input.Repository != nil {
		updateReq.Repository = input.Repository
	}
	if input.DefaultBranch != nil {
		updateReq.DefaultBranch = input.DefaultBranch
	}
	if input.Team != nil {
		team := projectsDomain.Team(*input.Team)
		updateReq.Team = &team
	}

	// Update the project
	if err := r.projectService.UpdateProject(ctx, snapshot.ProjectID, updateReq); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Fetch updated project
	updatedProject, err := r.projectService.GetProject(ctx, snapshot.ProjectID)
	if err != nil {
		return nil, err
	}

	return r.convertProjectToGraphQL(updatedProject), nil
}

// DeleteProject implementation using domain service
func (r *mutationResolver) DeleteProject_domain(ctx context.Context, id string) (bool, error) {
	// Check user permissions
	user, err := getCurrentUser(ctx)
	if err != nil {
		return false, err
	}

	// Same workaround as UpdateProject - find project by ID
	projects, _, err := r.projectService.ListProjects(ctx, 1000, 0)
	if err != nil {
		return false, fmt.Errorf("failed to list projects: %w", err)
	}

	var existingProject *projectsDomain.Project
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return false, fmt.Errorf("invalid project ID: %w", err)
	}

	for _, p := range projects {
		if p.ID() == uint(idUint) {
			existingProject = p
			break
		}
	}

	if existingProject == nil {
		return false, fmt.Errorf("project not found")
	}

	snapshot := existingProject.ToSnapshot()

	// Check if user can delete this project
	canDelete := false
	if user.Role == authDomain.RoleAdmin {
		canDelete = true
	} else {
		// Get role group names from context
		roleGroups := getRoleGroupNamesFromContext(ctx)

		// Check if user has team + manager group combination
		if snapshot.Team != "" {
			hasTeamGroup := false
			hasManagerGroup := false

			for _, group := range user.Groups {
				groupName := strings.TrimPrefix(group.GroupName, "/")
				if groupName == string(snapshot.Team) {
					hasTeamGroup = true
				}
				if groupName == roleGroups.ManagerGroup {
					hasManagerGroup = true
				}
			}

			// If user is in both team and manager groups, they can delete
			if hasTeamGroup && hasManagerGroup {
				canDelete = true
			}
		}

		// If not via groups, check scopes
		if !canDelete {
			// Check scopes for delete permission on this project
			requiredScopes := []string{
				fmt.Sprintf("project:delete:%s", snapshot.ProjectID),
				fmt.Sprintf("project:*:%s", snapshot.ProjectID),
			}

			// If project has a team, also check team-based scopes
			if snapshot.Team != "" {
				requiredScopes = append(requiredScopes,
					fmt.Sprintf("project:delete:%s:*", snapshot.Team),
					fmt.Sprintf("project:*:%s:*", snapshot.Team),
				)
			}

			scopes := getUserScopesFromContext(ctx)
			for _, scope := range scopes {
				for _, required := range requiredScopes {
					if matchScope(scope, required) {
						canDelete = true
						break
					}
				}
				if canDelete {
					break
				}
			}
		}
	}

	if !canDelete {
		return false, fmt.Errorf("insufficient permissions to delete project")
	}

	// Delete project
	err = r.projectService.DeleteProject(ctx, snapshot.ProjectID)
	if err != nil {
		return false, fmt.Errorf("failed to delete project: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"project_id": snapshot.ProjectID,
		"deleted_by": user.UserID,
	}).Info("Project deleted")

	return true, nil
}

// Project implementation using domain service
func (r *queryResolver) Project_domain(ctx context.Context, id string) (*model.Project, error) {
	// Parse ID and find project
	projects, _, err := r.projectService.ListProjects(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	for _, p := range projects {
		if p.ID() == uint(idUint) {
			return r.convertProjectToGraphQL(p), nil
		}
	}

	return nil, fmt.Errorf("project not found")
}

// ProjectByProjectID implementation using domain service
func (r *queryResolver) ProjectByProjectID_domain(ctx context.Context, projectID string) (*model.Project, error) {
	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return r.convertProjectToGraphQL(project), nil
}

// Projects implementation using domain service with pagination
func (r *queryResolver) Projects_domain(ctx context.Context, filter *model.ProjectFilter, first *int, after *string) (*model.ProjectConnection, error) {
	// Get current user
	user, err := getCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not authenticated")
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

	// Get projects with pagination
	projects, totalCount, err := r.projectService.ListProjects(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Filter projects based on user's team access
	var filteredProjects []*projectsDomain.Project
	if user.Role == authDomain.RoleAdmin {
		// Admins can see all projects
		filteredProjects = projects
	} else {
		// Get user's teams
		userTeams := getUserTeamsFromContext(ctx)
		teamMap := make(map[string]bool)
		for _, team := range userTeams {
			teamMap[team] = true
		}
		
		// Filter projects by team
		for _, project := range projects {
			snapshot := project.ToSnapshot()
			if snapshot.Team != "" && teamMap[string(snapshot.Team)] {
				filteredProjects = append(filteredProjects, project)
			}
		}
		
		// Update total count to reflect filtered results
		totalCount = int64(len(filteredProjects))
	}

	// Apply search filter if provided
	if filter != nil && filter.Search != nil && *filter.Search != "" {
		searchLower := strings.ToLower(*filter.Search)
		searchFiltered := make([]*projectsDomain.Project, 0)
		for _, project := range filteredProjects {
			snapshot := project.ToSnapshot()
			if strings.Contains(strings.ToLower(snapshot.Name), searchLower) ||
				strings.Contains(strings.ToLower(string(snapshot.ProjectID)), searchLower) {
				searchFiltered = append(searchFiltered, project)
			}
		}
		filteredProjects = searchFiltered
		totalCount = int64(len(filteredProjects))
	}

	// Apply active filter if provided
	if filter != nil && filter.ActiveOnly != nil && *filter.ActiveOnly {
		activeFiltered := make([]*projectsDomain.Project, 0)
		for _, project := range filteredProjects {
			if project.ToSnapshot().IsActive {
				activeFiltered = append(activeFiltered, project)
			}
		}
		filteredProjects = activeFiltered
		totalCount = int64(len(filteredProjects))
	}

	hasMore := offset + len(filteredProjects) < int(totalCount)

	// Build edges
	edges := make([]*model.ProjectEdge, len(filteredProjects))
	for i, project := range filteredProjects {
		edges[i] = &model.ProjectEdge{
			Node:   r.convertProjectToGraphQL(project),
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

// DashboardSummary implementation using domain service
func (r *queryResolver) DashboardSummary_domain(ctx context.Context) (*model.DashboardSummary, error) {
	// Get all projects to count them
	projects, totalProjects, err := r.projectService.ListProjects(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get project count: %w", err)
	}
	
	// Count active projects
	activeProjects := int64(0)
	for _, project := range projects {
		if project.ToSnapshot().IsActive {
			activeProjects++
		}
	}

	// Get test run statistics
	recentRuns, err := r.testingService.GetRecentTestRuns(ctx, 100)
	if err != nil {
		// Log error but don't fail the whole query
		r.logger.WithError(err).Error("Failed to get recent test runs for dashboard")
	}
	
	totalTestRuns := len(recentRuns)
	recentTestRuns := len(recentRuns)
	overallPassRate := float64(0)
	totalTestsExecuted := 0
	avgDuration := 0
	
	if len(recentRuns) > 0 {
		var totalTests, passedTests int
		var totalDuration int64
		
		for _, tr := range recentRuns {
			totalTests += tr.TotalTests
			passedTests += tr.PassedTests
			totalTestsExecuted += tr.TotalTests
			totalDuration += tr.Duration.Milliseconds()
		}
		
		if totalTests > 0 {
			overallPassRate = float64(passedTests) / float64(totalTests) * 100
		}
		
		if len(recentRuns) > 0 {
			avgDuration = int(totalDuration / int64(len(recentRuns)))
		}
	}

	version := "1.0.0"
	return &model.DashboardSummary{
		Health: &model.HealthStatus{
			Status:    "healthy",
			Service:   "fern-platform",
			Timestamp: time.Now(),
			Version:   &version,
		},
		ProjectCount:         int(totalProjects),
		ActiveProjectCount:   int(activeProjects),
		TotalTestRuns:        totalTestRuns,
		RecentTestRuns:       recentTestRuns,
		OverallPassRate:      overallPassRate,
		TotalTestsExecuted:   totalTestsExecuted,
		AverageTestDuration:  avgDuration,
	}, nil
}

// TreemapData implementation using domain service
func (r *queryResolver) TreemapData_domain(ctx context.Context, projectID *string, days *int) (*model.TreemapData, error) {
	// TODO: Implement treemap data using domain services
	// Need to integrate with testing service to get test runs
	return &model.TreemapData{
		Projects:        []*model.ProjectTreemapNode{},
		TotalDuration:   0,
		TotalTests:      0,
		OverallPassRate: 0,
	}, nil
}

// TestRuns implementation using domain service with pagination
func (r *queryResolver) TestRuns_domain(ctx context.Context, filter *model.TestRunFilter, first *int, after *string, orderBy *string, orderDirection *model.OrderDirection) (*model.TestRunConnection, error) {
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

	// Get project ID from filter if provided
	projectID := ""
	if filter != nil && filter.ProjectID != nil {
		projectID = *filter.ProjectID
	}

	// Get test runs with pagination
	testRuns, totalCount, err := r.testingService.ListTestRuns(ctx, projectID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list test runs: %w", err)
	}

	hasMore := offset + len(testRuns) < int(totalCount)

	// Build edges
	edges := make([]*model.TestRunEdge, len(testRuns))
	for i, run := range testRuns {
		edges[i] = &model.TestRunEdge{
			Node:   r.convertTestRunToGraphQL(run),
			Cursor: fmt.Sprintf("%d", offset+i), // Simple cursor
		}
	}

	// Build page info
	pageInfo := &model.PageInfo{
		HasNextPage:     hasMore,
		HasPreviousPage: offset > 0,
	}
	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.TestRunConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: int(totalCount),
	}, nil
}

// Tags implementation using domain service with pagination
func (r *queryResolver) Tags_domain(ctx context.Context, filter *model.TagFilter, first *int, after *string) (*model.TagConnection, error) {
	// Get all tags - domain doesn't support pagination yet
	tags, err := r.tagService.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	// Apply filter if provided
	filteredTags := tags
	if filter != nil && filter.Search != nil && *filter.Search != "" {
		filtered := make([]*tagsDomain.Tag, 0)
		searchLower := strings.ToLower(*filter.Search)
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(tag.Name()), searchLower) {
				filtered = append(filtered, tag)
			}
		}
		filteredTags = filtered
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

	// Slice tags for pagination
	start := offset
	end := offset + pageSize
	if start > len(filteredTags) {
		start = len(filteredTags)
	}
	if end > len(filteredTags) {
		end = len(filteredTags)
	}

	paginatedTags := filteredTags[start:end]
	hasMore := end < len(filteredTags)

	// Build edges
	edges := make([]*model.TagEdge, len(paginatedTags))
	for i, tag := range paginatedTags {
		edges[i] = &model.TagEdge{
			Node:   r.convertTagToGraphQL(tag),
			Cursor: fmt.Sprintf("%d", start+i),
		}
	}

	// Build page info
	pageInfo := &model.PageInfo{
		HasNextPage:     hasMore,
		HasPreviousPage: offset > 0,
	}
	if len(edges) > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[len(edges)-1].Cursor
	}

	return &model.TagConnection{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: len(filteredTags),
	}, nil
}