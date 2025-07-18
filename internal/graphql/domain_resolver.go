// Package graphql provides domain-based GraphQL resolvers
package graphql

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	analyticsApp "github.com/guidewire-oss/fern-platform/internal/domains/analytics/application"
	authInterfaces "github.com/guidewire-oss/fern-platform/internal/domains/auth/interfaces"
	projectsApp "github.com/guidewire-oss/fern-platform/internal/domains/projects/application"
	projectsDomain "github.com/guidewire-oss/fern-platform/internal/domains/projects/domain"
	tagsApp "github.com/guidewire-oss/fern-platform/internal/domains/tags/application"
	tagsDomain "github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
	"github.com/guidewire-oss/fern-platform/internal/domains/testing/application"
	testingDomain "github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/dataloader"
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/generated"
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/model"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	"gorm.io/gorm"
)

// DomainResolver is the root GraphQL resolver using domain services
type DomainResolver struct {
	testingService        *application.TestRunService
	projectService        *projectsApp.ProjectService
	tagService            *tagsApp.TagService
	flakyDetectionService *analyticsApp.FlakyDetectionService
	loaders               *dataloader.Loaders
	db                    *gorm.DB
	logger                *logging.Logger
}

// NewDomainResolver creates a new GraphQL resolver using domain services
func NewDomainResolver(
	testingService *application.TestRunService,
	projectService *projectsApp.ProjectService,
	tagService *tagsApp.TagService,
	flakyDetectionService *analyticsApp.FlakyDetectionService,
	db *gorm.DB,
	logger *logging.Logger,
) *DomainResolver {
	return &DomainResolver{
		testingService:        testingService,
		projectService:        projectService,
		tagService:            tagService,
		flakyDetectionService: flakyDetectionService,
		loaders:               dataloader.NewLoaders(db),
		db:                    db,
		logger:                logger,
	}
}

// Mutation returns generated.MutationResolver implementation.
func (r *DomainResolver) Mutation() generated.MutationResolver {
	return &mutationDomainResolver{r}
}

// Query returns generated.QueryResolver implementation.
func (r *DomainResolver) Query() generated.QueryResolver {
	return &queryDomainResolver{r}
}

// Project returns generated.ProjectResolver implementation.
func (r *DomainResolver) Project() generated.ProjectResolver {
	return &projectDomainResolver{r}
}

// TestRun returns generated.TestRunResolver implementation.
func (r *DomainResolver) TestRun() generated.TestRunResolver {
	return &testRunDomainResolver{r}
}

// SuiteRun returns generated.SuiteRunResolver implementation.
func (r *DomainResolver) SuiteRun() generated.SuiteRunResolver {
	return &suiteRunDomainResolver{r}
}

type mutationDomainResolver struct{ *DomainResolver }

// CreateTestRun creates a new test run
func (r *mutationDomainResolver) CreateTestRun(ctx context.Context, input model.CreateTestRunInput) (*model.TestRun, error) {
	// Create domain test run
	testRun := &testingDomain.TestRun{
		ProjectID:   input.ProjectID,
		RunID:       input.RunID,
		Name:        fmt.Sprintf("Test Run %s", time.Now().Format("2006-01-02 15:04:05")),
		Branch:      ptrString(input.Branch),
		GitBranch:   ptrString(input.Branch), // Use branch as git branch
		GitCommit:   ptrString(input.CommitSha),
		Environment: ptrString(input.Environment),
		Status:      "running",
	}

	if err := r.testingService.CreateTestRun(ctx, testRun); err != nil {
		return nil, fmt.Errorf("failed to create test run: %w", err)
	}

	// Convert to GraphQL model
	return r.convertTestRunToGraphQL(testRun), nil
}

// UpdateTestRunStatus updates test run status
func (r *mutationDomainResolver) UpdateTestRunStatus(ctx context.Context, runID string, status string, endTime *time.Time) (*model.TestRun, error) {
	// TODO: Implement update status in domain service
	return nil, fmt.Errorf("update test run status not yet implemented")
}

// DeleteTestRun deletes a test run
func (r *mutationDomainResolver) DeleteTestRun(ctx context.Context, id string) (bool, error) {
	// TODO: Implement delete in domain service
	return false, fmt.Errorf("delete test run not yet implemented")
}

// AssignTagsToTestRun assigns tags to a test run
func (r *mutationDomainResolver) AssignTagsToTestRun(ctx context.Context, testRunID string, tagIds []string) (*model.TestRun, error) {
	// Convert tag IDs to domain format
	domainTagIDs := make([]tagsDomain.TagID, len(tagIds))
	for i, id := range tagIds {
		domainTagIDs[i] = tagsDomain.TagID(id)
	}

	if err := r.tagService.AssignTagsToTestRun(ctx, testRunID, domainTagIDs); err != nil {
		return nil, fmt.Errorf("failed to assign tags: %w", err)
	}

	// Get updated test run
	testRunIDUint, err := strconv.ParseUint(testRunID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid test run ID")
	}

	testRun, err := r.testingService.GetTestRun(ctx, uint(testRunIDUint))
	if err != nil {
		return nil, err
	}

	return r.convertTestRunToGraphQL(testRun), nil
}

// CreateProject creates a new project
func (r *mutationDomainResolver) CreateProject(ctx context.Context, input model.CreateProjectInput) (*model.Project, error) {
	// Get current user for creator ID
	user, _ := authInterfaces.GetAuthUser(ctx.Value("gin_context").(*gin.Context))
	creatorUserID := ""
	if user != nil {
		creatorUserID = user.UserID
	}

	// Generate project ID if not provided
	projectID := input.ProjectID
	if projectID == "" {
		projectID = uuid.New().String()
	}

	// Create project
	project, err := r.projectService.CreateProject(
		ctx,
		projectsDomain.ProjectID(projectID),
		input.Name,
		projectsDomain.Team(ptrString(input.Team)),
		creatorUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Update additional fields if provided
	if input.Description != nil || input.Repository != nil || input.DefaultBranch != nil {
		updates := projectsApp.UpdateProjectRequest{
			Description:   input.Description,
			Repository:    input.Repository,
			DefaultBranch: input.DefaultBranch,
		}
		r.projectService.UpdateProject(ctx, project.ProjectID(), updates)
	}

	return r.convertProjectToGraphQL(project), nil
}

// UpdateProject updates a project
func (r *mutationDomainResolver) UpdateProject(ctx context.Context, id string, input model.UpdateProjectInput) (*model.Project, error) {
	// Build update request
	updates := projectsApp.UpdateProjectRequest{
		Name:          input.Name,
		Description:   input.Description,
		Repository:    input.Repository,
		DefaultBranch: input.DefaultBranch,
	}

	if input.Team != nil {
		team := projectsDomain.Team(*input.Team)
		updates.Team = &team
	}

	// Update project
	if err := r.projectService.UpdateProject(ctx, projectsDomain.ProjectID(id), updates); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Get updated project
	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(id))
	if err != nil {
		return nil, err
	}

	return r.convertProjectToGraphQL(project), nil
}

// DeleteProject deletes a project
func (r *mutationDomainResolver) DeleteProject(ctx context.Context, id string) (bool, error) {
	if err := r.projectService.DeleteProject(ctx, projectsDomain.ProjectID(id)); err != nil {
		return false, fmt.Errorf("failed to delete project: %w", err)
	}
	return true, nil
}

// ActivateProject activates a project
func (r *mutationDomainResolver) ActivateProject(ctx context.Context, id string) (*model.Project, error) {
	if err := r.projectService.ActivateProject(ctx, projectsDomain.ProjectID(id)); err != nil {
		return nil, fmt.Errorf("failed to activate project: %w", err)
	}

	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(id))
	if err != nil {
		return nil, err
	}

	return r.convertProjectToGraphQL(project), nil
}

// DeactivateProject deactivates a project
func (r *mutationDomainResolver) DeactivateProject(ctx context.Context, id string) (*model.Project, error) {
	if err := r.projectService.DeactivateProject(ctx, projectsDomain.ProjectID(id)); err != nil {
		return nil, fmt.Errorf("failed to deactivate project: %w", err)
	}

	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(id))
	if err != nil {
		return nil, err
	}

	return r.convertProjectToGraphQL(project), nil
}

// CreateTag creates a new tag
func (r *mutationDomainResolver) CreateTag(ctx context.Context, input model.CreateTagInput) (*model.Tag, error) {
	tag, err := r.tagService.CreateTag(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return r.convertTagToGraphQL(tag), nil
}

// UpdateTag updates a tag (no-op for domain tags which are immutable)
func (r *mutationDomainResolver) UpdateTag(ctx context.Context, id string, input model.UpdateTagInput) (*model.Tag, error) {
	// Tags are immutable in the domain, so just return the existing tag
	tag, err := r.tagService.GetTag(ctx, tagsDomain.TagID(id))
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	return r.convertTagToGraphQL(tag), nil
}

// DeleteTag deletes a tag
func (r *mutationDomainResolver) DeleteTag(ctx context.Context, id string) (bool, error) {
	if err := r.tagService.DeleteTag(ctx, tagsDomain.TagID(id)); err != nil {
		return false, fmt.Errorf("failed to delete tag: %w", err)
	}
	return true, nil
}

type queryDomainResolver struct{ *DomainResolver }

// TestRun retrieves a test run by ID
func (r *queryDomainResolver) TestRun(ctx context.Context, id string) (*model.TestRun, error) {
	testRunID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid test run ID")
	}

	testRun, err := r.testingService.GetTestRun(ctx, uint(testRunID))
	if err != nil {
		return nil, fmt.Errorf("test run not found")
	}

	return r.convertTestRunToGraphQL(testRun), nil
}

// TestRuns retrieves test runs with optional filtering
func (r *queryDomainResolver) TestRuns(ctx context.Context, filter *model.TestRunFilter, first *int, after *string, orderBy *string, orderDirection *model.OrderDirection) (*model.TestRunConnection, error) {
	// Default values
	limitVal := 50
	if first != nil {
		limitVal = *first
	}

	// Get test runs from domain service
	var testRuns []*testingDomain.TestRun
	var err error

	if filter != nil && filter.ProjectID != nil {
		testRuns, err = r.testingService.GetProjectTestRuns(ctx, *filter.ProjectID, limitVal)
	} else {
		// TODO: Implement general list method in domain service
		testRuns = []*testingDomain.TestRun{}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get test runs: %w", err)
	}

	// Filter by additional criteria if needed
	filteredRuns := testRuns
	if filter != nil && (filter.Branch != nil || filter.Status != nil) {
		filteredRuns = make([]*testingDomain.TestRun, 0)
		for _, tr := range testRuns {
			if filter.Branch != nil && tr.Branch != *filter.Branch {
				continue
			}
			if filter.Status != nil && tr.Status != *filter.Status {
				continue
			}
			filteredRuns = append(filteredRuns, tr)
		}
	}

	// Convert to GraphQL models
	edges := make([]*model.TestRunEdge, len(filteredRuns))
	for i, tr := range filteredRuns {
		edges[i] = &model.TestRunEdge{
			Node:   r.convertTestRunToGraphQL(tr),
			Cursor: fmt.Sprintf("%d", i),
		}
	}

	return &model.TestRunConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     len(filteredRuns) == limitVal,
			HasPreviousPage: after != nil && *after != "",
		},
		TotalCount: len(filteredRuns),
	}, nil
}

// Project retrieves a project by ID
func (r *queryDomainResolver) Project(ctx context.Context, id string) (*model.Project, error) {
	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(id))
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}

	return r.convertProjectToGraphQL(project), nil
}

// Projects retrieves projects with optional filtering
func (r *queryDomainResolver) Projects(ctx context.Context, filter *model.ProjectFilter, first *int, after *string) (*model.ProjectConnection, error) {
	// Default values
	limitVal := 20
	offsetVal := 0

	if first != nil {
		limitVal = *first
	}

	// Get projects from domain service
	projects, total, err := r.projectService.ListProjects(ctx, limitVal, offsetVal)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	// Filter by additional criteria if needed
	filteredProjects := projects
	if filter != nil && (filter.Search != nil || filter.ActiveOnly != nil) {
		filteredProjects = make([]*projectsDomain.Project, 0)
		for _, p := range projects {
			snapshot := p.ToSnapshot()

			// Filter by search term
			if filter.Search != nil && *filter.Search != "" {
				searchLower := strings.ToLower(*filter.Search)
				projectText := strings.ToLower(fmt.Sprintf("%s %s", snapshot.Name, snapshot.Description))
				if !strings.Contains(projectText, searchLower) {
					continue
				}
			}

			// Filter by active status
			if filter.ActiveOnly != nil && *filter.ActiveOnly && !snapshot.IsActive {
				continue
			}

			filteredProjects = append(filteredProjects, p)
		}
	}

	// Convert to GraphQL models
	edges := make([]*model.ProjectEdge, len(filteredProjects))
	for i, p := range filteredProjects {
		edges[i] = &model.ProjectEdge{
			Node:   r.convertProjectToGraphQL(p),
			Cursor: fmt.Sprintf("%d", offsetVal+i),
		}
	}

	return &model.ProjectConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     int64(offsetVal+limitVal) < total,
			HasPreviousPage: offsetVal > 0,
		},
		TotalCount: len(filteredProjects),
	}, nil
}

// Tag retrieves a tag by ID
func (r *queryDomainResolver) Tag(ctx context.Context, id string) (*model.Tag, error) {
	tag, err := r.tagService.GetTag(ctx, tagsDomain.TagID(id))
	if err != nil {
		return nil, fmt.Errorf("tag not found")
	}

	return r.convertTagToGraphQL(tag), nil
}

// Tags retrieves all tags
func (r *queryDomainResolver) Tags(ctx context.Context, filter *model.TagFilter, first *int, after *string) (*model.TagConnection, error) {
	tags, err := r.tagService.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	// Filter by search if provided
	filteredTags := tags
	if filter != nil && filter.Search != nil && *filter.Search != "" {
		searchLower := strings.ToLower(*filter.Search)
		filteredTags = make([]*tagsDomain.Tag, 0)
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(tag.Name()), searchLower) {
				filteredTags = append(filteredTags, tag)
			}
		}
	}

	// Apply pagination
	start := 0

	end := len(filteredTags)
	if first != nil && *first < end {
		end = *first
	}

	paginatedTags := filteredTags[start:end]

	// Convert to GraphQL models
	edges := make([]*model.TagEdge, len(paginatedTags))
	for i, tag := range paginatedTags {
		edges[i] = &model.TagEdge{
			Node:   r.convertTagToGraphQL(tag),
			Cursor: fmt.Sprintf("%d", start+i),
		}
	}

	return &model.TagConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     end < len(filteredTags),
			HasPreviousPage: start > 0,
		},
		TotalCount: len(filteredTags),
	}, nil
}

// FlakyTests retrieves flaky tests
func (r *queryDomainResolver) FlakyTests(ctx context.Context, filter *model.FlakyTestFilter, first *int, after *string, orderBy *string, orderDirection *model.OrderDirection) (*model.FlakyTestConnection, error) {
	// TODO: Implement flaky tests when analytics service is ready
	return &model.FlakyTestConnection{
		Edges: []*model.FlakyTestEdge{},
		PageInfo: &model.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		},
		TotalCount: 0,
	}, nil
}

type projectDomainResolver struct{ *DomainResolver }

func (r *projectDomainResolver) TestRuns(ctx context.Context, obj *model.Project, limit *int, offset *int) ([]*model.TestRun, error) {
	limitVal := 50
	if limit != nil {
		limitVal = *limit
	}

	testRuns, err := r.testingService.GetProjectTestRuns(ctx, obj.ProjectID, limitVal)
	if err != nil {
		return nil, err
	}

	// Convert to GraphQL models
	result := make([]*model.TestRun, len(testRuns))
	for i, tr := range testRuns {
		result[i] = r.convertTestRunToGraphQL(tr)
	}

	return result, nil
}

func (r *projectDomainResolver) Tags(ctx context.Context, obj *model.Project) ([]*model.Tag, error) {
	// TODO: Implement project tags relationship
	return []*model.Tag{}, nil
}

func (r *projectDomainResolver) Stats(ctx context.Context, obj *model.Project) (*model.ProjectStats, error) {
	// TODO: Implement project stats
	return &model.ProjectStats{
		TotalTestRuns:   0,
		RecentTestRuns:  0,
		UniqueBranches:  0,
		SuccessRate:     0,
		AverageDuration: 0,
		LastRunTime:     nil,
	}, nil
}

type testRunDomainResolver struct{ *DomainResolver }

func (r *testRunDomainResolver) Project(ctx context.Context, obj *model.TestRun) (*model.Project, error) {
	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(obj.ProjectID))
	if err != nil {
		return nil, err
	}
	return r.convertProjectToGraphQL(project), nil
}

func (r *testRunDomainResolver) SuiteRuns(ctx context.Context, obj *model.TestRun) ([]*model.SuiteRun, error) {
	// TODO: Load suite runs from dataloader or service
	return []*model.SuiteRun{}, nil
}

func (r *testRunDomainResolver) Tags(ctx context.Context, obj *model.TestRun) ([]*model.Tag, error) {
	// TODO: Load tags for test run
	return []*model.Tag{}, nil
}

type suiteRunDomainResolver struct{ *DomainResolver }

func (r *suiteRunDomainResolver) TestRun(ctx context.Context, obj *model.SuiteRun) (*model.TestRun, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (r *suiteRunDomainResolver) SpecRuns(ctx context.Context, obj *model.SuiteRun) ([]*model.SpecRun, error) {
	// TODO: Load spec runs
	return []*model.SpecRun{}, nil
}

// Helper functions

func (r *DomainResolver) convertTestRunToGraphQL(tr *testingDomain.TestRun) *model.TestRun {
	var endTime *time.Time
	if tr.EndTime != nil && !tr.EndTime.IsZero() {
		endTime = tr.EndTime
	}

	// Handle optional fields
	var branch, commitSha, environment *string
	if tr.GitBranch != "" {
		branch = &tr.GitBranch
	}
	if tr.GitCommit != "" {
		commitSha = &tr.GitCommit
	}
	if tr.Environment != "" {
		environment = &tr.Environment
	}

	return &model.TestRun{
		ID:           strconv.FormatUint(uint64(tr.ID), 10),
		ProjectID:    tr.ProjectID,
		RunID:        tr.RunID,
		Branch:       branch,
		CommitSha:    commitSha,
		Status:       tr.Status,
		StartTime:    tr.StartTime,
		EndTime:      endTime,
		TotalTests:   tr.TotalTests,
		PassedTests:  tr.PassedTests,
		FailedTests:  tr.FailedTests,
		SkippedTests: tr.SkippedTests,
		Duration:     int(tr.Duration.Milliseconds()),
		Environment:  environment,
		Metadata:     tr.Metadata,
		Tags:         []*model.Tag{},      // TODO: Load tags
		SuiteRuns:    []*model.SuiteRun{}, // TODO: Load suite runs
		CreatedAt:    tr.StartTime,
		UpdatedAt:    tr.StartTime,
	}
}

func (r *DomainResolver) convertProjectToGraphQL(p *projectsDomain.Project) *model.Project {
	snapshot := p.ToSnapshot()
	desc := &snapshot.Description
	repo := &snapshot.Repository
	team := string(snapshot.Team)

	return &model.Project{
		ID:            strconv.FormatUint(uint64(snapshot.ID), 10),
		ProjectID:     string(snapshot.ProjectID),
		Name:          snapshot.Name,
		Description:   desc,
		Repository:    repo,
		DefaultBranch: snapshot.DefaultBranch,
		Team:          &team,
		IsActive:      snapshot.IsActive,
		Settings:      snapshot.Settings,
		CanManage:     false, // TODO: Implement permission check
		Stats:         nil,   // TODO: Implement stats calculation
		CreatedAt:     snapshot.CreatedAt,
		UpdatedAt:     snapshot.UpdatedAt,
	}
}

func (r *DomainResolver) convertTagToGraphQL(t *tagsDomain.Tag) *model.Tag {
	emptyStr := ""
	return &model.Tag{
		ID:          string(t.ID()),
		Name:        t.Name(),
		Description: &emptyStr,
		Color:       &emptyStr,
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.CreatedAt(),
	}
}

func ptrString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// Missing methods for MutationResolver interface

func (r *mutationDomainResolver) MarkFlakyTestResolved(ctx context.Context, id string) (*model.FlakyTest, error) {
	// TODO: Implement flaky test resolution
	return nil, fmt.Errorf("not implemented")
}

func (r *mutationDomainResolver) MarkSpecAsFlaky(ctx context.Context, specRunID string) (*model.SpecRun, error) {
	// TODO: Implement mark spec as flaky
	return nil, fmt.Errorf("not implemented")
}

func (r *mutationDomainResolver) UpdateUserPreferences(ctx context.Context, input model.UpdateUserPreferencesInput) (*model.UserPreferences, error) {
	// TODO: Implement update user preferences
	return &model.UserPreferences{
		ID:          "1",
		UserID:      "user-1",
		Theme:       input.Theme,
		Timezone:    input.Timezone,
		Language:    input.Language,
		Favorites:   input.Favorites,
		Preferences: input.Preferences,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (r *mutationDomainResolver) ToggleProjectFavorite(ctx context.Context, projectID string) (*model.UserPreferences, error) {
	// TODO: Implement toggle project favorite
	return &model.UserPreferences{
		ID:        "1",
		UserID:    "user-1",
		Favorites: []string{projectID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// Missing methods for QueryResolver interface

func (r *queryDomainResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	// Get user from context (set by auth middleware)
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if user, exists := authInterfaces.GetAuthUser(ginCtx); exists {
			return &model.User{
				ID:     user.UserID,
				UserID: user.UserID,
				Email:  user.Email,
				Name:   user.Name,
				Role:   string(user.Role),
			}, nil
		}
	}
	return nil, fmt.Errorf("user not found in context")
}

func (r *queryDomainResolver) DashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	// Get project counts
	projects, totalProjects, err := r.projectService.ListProjects(ctx, 1000, 0)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get projects for dashboard")
	}

	activeProjectCount := 0
	for _, p := range projects {
		if p.ToSnapshot().IsActive {
			activeProjectCount++
		}
	}

	// Get test run statistics
	totalTestRuns := 0
	recentTestRuns := 0
	overallPassRate := 0.0
	totalTestsExecuted := 0
	averageTestDuration := 0

	// Get recent test runs across all projects to calculate stats
	recentRuns, err := r.testingService.GetRecentTestRuns(ctx, 100)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get recent test runs for dashboard")
	} else {
		r.logger.Info("Got recent test runs for dashboard", "count", len(recentRuns))
		totalTestRuns = len(recentRuns)
		recentTestRuns = len(recentRuns)

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
			averageTestDuration = int(totalDuration / int64(len(recentRuns)))
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
		ProjectCount:        int(totalProjects),
		ActiveProjectCount:  activeProjectCount,
		TotalTestRuns:       totalTestRuns,
		RecentTestRuns:      recentTestRuns,
		OverallPassRate:     overallPassRate,
		TotalTestsExecuted:  totalTestsExecuted,
		AverageTestDuration: averageTestDuration,
	}, nil
}

func (r *queryDomainResolver) FlakyTest(ctx context.Context, id string) (*model.FlakyTest, error) {
	// TODO: Implement get flaky test by ID
	return nil, fmt.Errorf("not implemented")
}

func (r *queryDomainResolver) FlakyTestStats(ctx context.Context, projectID *string) (*model.FlakyTestStats, error) {
	// TODO: Implement flaky test stats
	return &model.FlakyTestStats{
		TotalFlakyTests:  0,
		SeverityCounts:   []*model.SeverityCount{},
		AverageFlakeRate: 0,
		MostFlakyTest:    nil,
	}, nil
}

func (r *queryDomainResolver) Health(ctx context.Context) (*model.HealthStatus, error) {
	version := "1.0.0"
	return &model.HealthStatus{
		Status:    "healthy",
		Service:   "fern-platform",
		Timestamp: time.Now(),
		Version:   &version,
	}, nil
}

func (r *queryDomainResolver) SystemConfig(ctx context.Context) (*model.SystemConfig, error) {
	// TODO: Implement system config
	return &model.SystemConfig{
		RoleGroups: &model.RoleGroupConfig{
			AdminGroup:   "admin",
			ManagerGroup: "manager",
			UserGroup:    "user",
		},
	}, nil
}

func (r *queryDomainResolver) TreemapData(ctx context.Context, projectID *string, days *int) (*model.TreemapData, error) {
	// TODO: Implement treemap data
	return &model.TreemapData{
		Projects:        []*model.ProjectTreemapNode{},
		TotalDuration:   0,
		TotalTests:      0,
		OverallPassRate: 0,
	}, nil
}

func (r *queryDomainResolver) TagByName(ctx context.Context, name string) (*model.Tag, error) {
	// TODO: Implement get tag by name
	return nil, fmt.Errorf("not implemented")
}

func (r *queryDomainResolver) TagUsageStats(ctx context.Context) ([]*model.TagUsage, error) {
	// TODO: Implement tag usage stats
	return []*model.TagUsage{}, nil
}

func (r *queryDomainResolver) PopularTags(ctx context.Context, limit *int) ([]*model.TagUsage, error) {
	// TODO: Implement popular tags
	return []*model.TagUsage{}, nil
}

func (r *queryDomainResolver) RecentlyAddedFlakyTests(ctx context.Context, projectID *string, days *int, limit *int) ([]*model.FlakyTest, error) {
	// TODO: Implement recently added flaky tests
	return []*model.FlakyTest{}, nil
}

func (r *queryDomainResolver) UserPreferences(ctx context.Context) (*model.UserPreferences, error) {
	// TODO: Implement user preferences retrieval
	return &model.UserPreferences{
		ID:          "1",
		UserID:      "user-1",
		Theme:       nil,
		Timezone:    nil,
		Language:    nil,
		Favorites:   []string{},
		Preferences: nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (r *queryDomainResolver) TestRunStats(ctx context.Context, projectID *string, days *int) (*model.TestRunStats, error) {
	// TODO: Implement test run stats
	return &model.TestRunStats{
		TotalRuns:       0,
		StatusCounts:    []*model.StatusCount{},
		AverageDuration: 0,
		SuccessRate:     0,
	}, nil
}

func (r *queryDomainResolver) RecentTestRuns(ctx context.Context, projectID *string, limit *int) ([]*model.TestRun, error) {
	limitVal := 10
	if limit != nil {
		limitVal = *limit
	}

	r.logger.Info("RecentTestRuns query called", "projectID", projectID, "limit", limitVal)

	var testRuns []*testingDomain.TestRun
	var err error

	if projectID != nil {
		r.logger.Info("Getting test runs for specific project", "projectID", *projectID)
		testRuns, err = r.testingService.GetProjectTestRuns(ctx, *projectID, limitVal)
	} else {
		r.logger.Info("Getting recent test runs across all projects")
		testRuns, err = r.testingService.GetRecentTestRuns(ctx, limitVal)
	}

	if err != nil {
		r.logger.WithError(err).Error("Failed to get recent test runs")
		return nil, fmt.Errorf("failed to get recent test runs: %w", err)
	}

	r.logger.Info("Retrieved test runs", "count", len(testRuns))

	// Convert to GraphQL models
	result := make([]*model.TestRun, len(testRuns))
	for i, tr := range testRuns {
		result[i] = r.convertTestRunToGraphQL(tr)
		r.logger.Debug("Converted test run", "index", i, "id", tr.ID, "runId", tr.RunID)
	}

	return result, nil
}

func (r *queryDomainResolver) TestRunByRunID(ctx context.Context, runID string) (*model.TestRun, error) {
	testRun, err := r.testingService.GetTestRunByRunID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("test run not found")
	}
	return r.convertTestRunToGraphQL(testRun), nil
}

func (r *queryDomainResolver) ProjectByProjectID(ctx context.Context, projectID string) (*model.Project, error) {
	project, err := r.projectService.GetProject(ctx, projectsDomain.ProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}
	return r.convertProjectToGraphQL(project), nil
}

// Missing methods for ProjectResolver interface

func (r *projectDomainResolver) CanManage(ctx context.Context, obj *model.Project) (bool, error) {
	// TODO: Implement permission checking based on user and project
	return true, nil
}

// Missing Subscription resolver

func (r *DomainResolver) Subscription() generated.SubscriptionResolver {
	return &subscriptionDomainResolver{r}
}

type subscriptionDomainResolver struct{ *DomainResolver }

// Add empty subscription methods as needed by the interface
func (r *subscriptionDomainResolver) TestRunCreated(ctx context.Context, projectID *string) (<-chan *model.TestRun, error) {
	// TODO: Implement real-time subscriptions
	ch := make(chan *model.TestRun)
	close(ch)
	return ch, nil
}

func (r *subscriptionDomainResolver) TestRunUpdated(ctx context.Context, projectID *string) (<-chan *model.TestRun, error) {
	// TODO: Implement real-time subscriptions
	ch := make(chan *model.TestRun)
	close(ch)
	return ch, nil
}

func (r *subscriptionDomainResolver) TestRunStatusChanged(ctx context.Context, projectID *string) (<-chan *model.TestRun, error) {
	// TODO: Implement real-time subscriptions
	ch := make(chan *model.TestRun)
	close(ch)
	return ch, nil
}

func (r *subscriptionDomainResolver) FlakyTestDetected(ctx context.Context, projectID *string) (<-chan *model.FlakyTest, error) {
	// TODO: Implement real-time subscriptions
	ch := make(chan *model.FlakyTest)
	close(ch)
	return ch, nil
}

// Ensure DomainResolver implements the GraphQL resolver interface
var _ generated.ResolverRoot = (*DomainResolver)(nil)
