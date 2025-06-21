package fixtures

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"

	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/reporter"
)

// TestDataManager manages test data fixtures
type TestDataManager struct {
	reporterClient *reporter.Client
	namespace      string
	testID         string
	createdData    *CreatedTestData
}

// CreatedTestData holds references to created test data
type CreatedTestData struct {
	Projects []reporter.Project  `json:"projects"`
	TestRuns []reporter.TestRun  `json:"testRuns"`
	SpecRuns []reporter.SpecRun  `json:"specRuns"`
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager(reporterClient *reporter.Client, namespace, testID string) *TestDataManager {
	return &TestDataManager{
		reporterClient: reporterClient,
		namespace:      namespace,
		testID:         testID,
		createdData:    &CreatedTestData{},
	}
}

// SetupTestData creates comprehensive test data for acceptance tests
func (t *TestDataManager) SetupTestData(ctx context.Context) error {
	GinkgoHelper()
	By("Setting up test data fixtures")

	// Create projects
	if err := t.createProjects(ctx); err != nil {
		return fmt.Errorf("failed to create projects: %w", err)
	}

	// Create test runs
	if err := t.createTestRuns(ctx); err != nil {
		return fmt.Errorf("failed to create test runs: %w", err)
	}

	fmt.Printf("✅ Created test data: %d projects, %d test runs\n", 
		len(t.createdData.Projects), len(t.createdData.TestRuns))
	
	return nil
}

// GetCreatedData returns the created test data
func (t *TestDataManager) GetCreatedData() *CreatedTestData {
	return t.createdData
}

// CleanupTestData removes all created test data
func (t *TestDataManager) CleanupTestData(ctx context.Context) error {
	GinkgoHelper()
	By("Cleaning up test data fixtures")

	// Note: In a real implementation, you would have delete endpoints
	// For now, we assume data is cleaned up when the namespace is deleted
	
	fmt.Printf("✅ Test data cleanup completed for namespace: %s\n", t.namespace)
	return nil
}

func (t *TestDataManager) createProjects(ctx context.Context) error {
	projectTemplates := []struct {
		name        string
		description string
		tags        []string
	}{
		{
			name:        "fern-platform-api",
			description: "Core API service acceptance tests",
			tags:        []string{"api", "core", "backend"},
		},
		{
			name:        "fern-platform-ui",
			description: "Frontend React application tests",
			tags:        []string{"ui", "frontend", "react"},
		},
		{
			name:        "fern-integration",
			description: "End-to-end integration tests",
			tags:        []string{"integration", "e2e"},
		},
		{
			name:        "fern-mycelium",
			description: "AI analysis service tests",
			tags:        []string{"ai", "analysis", "mycelium"},
		},
		{
			name:        "fern-performance",
			description: "Performance and load testing",
			tags:        []string{"performance", "load"},
		},
	}

	for _, template := range projectTemplates {
		project := &reporter.Project{
			ID:          uuid.New().String(),
			Name:        fmt.Sprintf("%s-%s", template.name, t.testID),
			Description: template.description,
			Tags:        template.tags,
			CreatedAt:   time.Now(),
		}

		createdProject, err := t.reporterClient.CreateProject(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to create project %s: %w", project.Name, err)
		}

		t.createdData.Projects = append(t.createdData.Projects, *createdProject)
	}

	return nil
}

func (t *TestDataManager) createTestRuns(ctx context.Context) error {
	branches := []string{"main", "develop", "feature/auth", "feature/dashboard", "hotfix/bug-123"}
	statuses := []string{"passed", "failed", "skipped"}
	
	rand.Seed(time.Now().UnixNano())

	for _, project := range t.createdData.Projects {
		// Create 15-25 test runs per project for good test data variety
		numTestRuns := 15 + rand.Intn(10)
		
		for i := 0; i < numTestRuns; i++ {
			testRun := t.generateTestRun(project.ID, branches, statuses, i)
			
			createdTestRun, err := t.reporterClient.CreateTestRun(ctx, testRun)
			if err != nil {
				return fmt.Errorf("failed to create test run for project %s: %w", project.Name, err)
			}

			t.createdData.TestRuns = append(t.createdData.TestRuns, *createdTestRun)
		}
	}

	return nil
}

func (t *TestDataManager) generateTestRun(projectID string, branches, statuses []string, index int) *reporter.TestRun {
	// Generate realistic test run data
	startTime := time.Now().Add(-time.Duration(rand.Intn(7*24)) * time.Hour) // Within last week
	
	// Duration between 30 seconds and 20 minutes
	durationMs := int64(30000 + rand.Intn(1170000))
	endTime := startTime.Add(time.Duration(durationMs) * time.Millisecond)
	
	// Status distribution: 70% passed, 20% failed, 10% skipped
	var status string
	statusRand := rand.Float32()
	switch {
	case statusRand < 0.7:
		status = "passed"
	case statusRand < 0.9:
		status = "failed"
	default:
		status = "skipped"
	}

	// Branch distribution: 50% main, 20% develop, 30% feature branches
	var branch string
	branchRand := rand.Float32()
	switch {
	case branchRand < 0.5:
		branch = "main"
	case branchRand < 0.7:
		branch = "develop"
	default:
		branch = branches[2+rand.Intn(len(branches)-2)] // Feature branches
	}

	testRun := &reporter.TestRun{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		SuiteID:   uuid.New().String(),
		Status:    status,
		StartTime: startTime,
		EndTime:   &endTime,
		Duration:  durationMs,
		Branch:    branch,
		Tags:      t.generateTags(),
		SpecRuns:  t.generateSpecRuns(status),
	}

	return testRun
}

func (t *TestDataManager) generateTags() []string {
	allTags := []string{"smoke", "regression", "api", "ui", "integration", "unit", "auth", "database", "security"}
	
	// Select 1-4 random tags
	numTags := 1 + rand.Intn(4)
	selectedTags := make([]string, 0, numTags)
	
	// Shuffle and select
	rand.Shuffle(len(allTags), func(i, j int) {
		allTags[i], allTags[j] = allTags[j], allTags[i]
	})
	
	for i := 0; i < numTags && i < len(allTags); i++ {
		selectedTags = append(selectedTags, allTags[i])
	}
	
	return selectedTags
}

func (t *TestDataManager) generateSpecRuns(testRunStatus string) []reporter.SpecRun {
	specTemplates := []string{
		"should authenticate user with valid credentials",
		"should reject invalid login attempts",
		"should load dashboard within performance threshold",
		"should display test runs with proper pagination",
		"should filter test runs by project",
		"should handle API errors gracefully",
		"should validate input parameters",
		"should maintain session state",
		"should log user activities",
		"should enforce rate limiting",
	}

	// Generate 5-15 specs per test run
	numSpecs := 5 + rand.Intn(10)
	specRuns := make([]reporter.SpecRun, 0, numSpecs)

	for i := 0; i < numSpecs; i++ {
		spec := t.generateSpecRun(specTemplates, testRunStatus, i)
		specRuns = append(specRuns, spec)
	}

	return specRuns
}

func (t *TestDataManager) generateSpecRun(templates []string, testRunStatus string, index int) reporter.SpecRun {
	startTime := time.Now().Add(-time.Duration(rand.Intn(60)) * time.Minute)
	
	// Spec duration between 100ms and 30 seconds
	durationMs := int64(100 + rand.Intn(29900))
	endTime := startTime.Add(time.Duration(durationMs) * time.Millisecond)

	// Determine spec status based on test run status
	var status string
	var errorMessage string
	
	if testRunStatus == "failed" && rand.Float32() < 0.3 { // 30% of specs fail in failed test runs
		status = "failed"
		errorMessage = t.generateErrorMessage()
	} else if testRunStatus == "skipped" && rand.Float32() < 0.2 { // 20% of specs skipped in skipped test runs
		status = "skipped"
	} else {
		status = "passed"
	}

	// Select random description template
	description := templates[rand.Intn(len(templates))]

	return reporter.SpecRun{
		ID:              uuid.New().String(),
		TestRunID:       "", // Will be set by the parent test run
		SpecDescription: description,
		Status:          status,
		StartTime:       startTime,
		EndTime:         &endTime,
		Duration:        durationMs,
		ErrorMessage:    errorMessage,
		StackTrace:      t.generateStackTrace(errorMessage),
	}
}

func (t *TestDataManager) generateErrorMessage() string {
	errorTemplates := []string{
		"Expected status code 200, but got 404",
		"Timeout waiting for element to appear",
		"Database connection failed",
		"Invalid JSON response from API",
		"Authentication token expired",
		"Network request timeout after 30 seconds",
		"Assertion failed: expected 'true' but got 'false'",
		"Resource not found: /api/v1/test-runs/123",
		"Validation error: email field is required",
		"Permission denied: insufficient privileges",
	}
	
	return errorTemplates[rand.Intn(len(errorTemplates))]
}

func (t *TestDataManager) generateStackTrace(errorMessage string) string {
	if errorMessage == "" {
		return ""
	}

	stackTraces := []string{
		`at Object.exports.default (/app/tests/api.test.js:45:23)
    at /app/node_modules/jest/lib/jasmine.js:117:18
    at tryCallOne (/app/node_modules/promise/lib/core.js:37:12)`,
		
		`at TestRunner.handleError (/app/src/test-runner.go:123)
    at TestSuite.executeSpec (/app/src/test-suite.go:89)
    at main.runTests (/app/main.go:67)`,
		
		`at BrowserContext.evaluate (/app/node_modules/chromedp/index.js:234:15)
    at Page.screenshot (/app/tests/ui/dashboard.test.js:89:31)
    at TestCase.run (/app/tests/ui/base.test.js:156:12)`,
	}
	
	return stackTraces[rand.Intn(len(stackTraces))]
}

// CreateMinimalTestData creates a minimal set of test data for quick tests
func (t *TestDataManager) CreateMinimalTestData(ctx context.Context) error {
	GinkgoHelper()
	By("Creating minimal test data")

	// Create one project
	project := &reporter.Project{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("minimal-test-%s", t.testID),
		Description: "Minimal test project for quick tests",
		Tags:        []string{"test"},
		CreatedAt:   time.Now(),
	}

	createdProject, err := t.reporterClient.CreateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to create minimal project: %w", err)
	}

	t.createdData.Projects = append(t.createdData.Projects, *createdProject)

	// Create a few test runs
	for i := 0; i < 3; i++ {
		testRun := t.generateTestRun(
			createdProject.ID, 
			[]string{"main"}, 
			[]string{"passed", "failed", "skipped"}, 
			i,
		)
		
		createdTestRun, err := t.reporterClient.CreateTestRun(ctx, testRun)
		if err != nil {
			return fmt.Errorf("failed to create minimal test run: %w", err)
		}

		t.createdData.TestRuns = append(t.createdData.TestRuns, *createdTestRun)
	}

	fmt.Printf("✅ Created minimal test data: 1 project, 3 test runs\n")
	return nil
}

// InitializeWithExistingData initializes the test data manager without creating new data
// This is used when testing against an existing platform to avoid data pollution
func (t *TestDataManager) InitializeWithExistingData(ctx context.Context) error {
	GinkgoHelper()
	By("Initializing with existing platform data")
	
	// Initialize with empty data structures to avoid nil pointer issues
	t.createdData = &CreatedTestData{
		Projects: []reporter.Project{},
		TestRuns: []reporter.TestRun{},
		SpecRuns: []reporter.SpecRun{},
	}
	
	fmt.Printf("✅ Initialized test data manager for existing platform\n")
	return nil
}