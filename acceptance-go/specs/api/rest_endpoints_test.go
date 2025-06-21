package api_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/reporter"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/fixtures"
)

var _ = Describe("REST API Endpoints", func() {
	var (
		ctx            context.Context
		reporterClient *reporter.Client
		testData       *fixtures.CreatedTestData
	)

	BeforeEach(func() {
		ctx = GetTestContext()
		reporterClient = GetReporterClient()
		testData = GetTestData()
	})

	Describe("Health Check Endpoint", func() {
		It("should return healthy status", func() {
			By("Checking health endpoint")
			err := reporterClient.HealthCheck(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should respond quickly", func() {
			By("Measuring health check response time")
			startTime := time.Now()
			err := reporterClient.HealthCheck(ctx)
			duration := time.Since(startTime)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 1*time.Second),
				"Health check should respond within 1 second")
		})
	})

	Describe("Test Runs API", func() {
		It("should retrieve test runs with default parameters", func() {
			By("Fetching test runs with default parameters")
			response, err := reporterClient.GetTestRuns(ctx, nil)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Data).NotTo(BeNil())
			Expect(len(response.Data)).To(BeNumerically(">=", 0))
			Expect(response.TotalCount).To(BeNumerically(">=", 0))
		})

		It("should respect limit parameter", func() {
			By("Fetching test runs with limit parameter")
			opts := &reporter.TestRunsOptions{
				Limit: 5,
			}
			
			response, err := reporterClient.GetTestRuns(ctx, opts)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Data)).To(BeNumerically("<=", 5))
		})

		It("should filter by project ID", func() {
			// Get a test project
			projects := testData.Projects
			Expect(len(projects)).To(BeNumerically(">", 0))
			testProject := projects[0]
			
			By(fmt.Sprintf("Filtering test runs by project: %s", testProject.Name))
			opts := &reporter.TestRunsOptions{
				ProjectID: testProject.ID,
				Limit:     20,
			}
			
			response, err := reporterClient.GetTestRuns(ctx, opts)
			
			Expect(err).NotTo(HaveOccurred())
			
			// Verify all results match the filter
			for _, testRun := range response.Data {
				Expect(testRun.ProjectID).To(Equal(testProject.ID))
			}
		})

		It("should filter by status", func() {
			By("Filtering test runs by failed status")
			opts := &reporter.TestRunsOptions{
				Status: "failed",
				Limit:  20,
			}
			
			response, err := reporterClient.GetTestRuns(ctx, opts)
			
			Expect(err).NotTo(HaveOccurred())
			
			// Verify all results have failed status
			for _, testRun := range response.Data {
				Expect(testRun.Status).To(Equal("failed"))
			}
		})

		It("should filter by branch", func() {
			By("Filtering test runs by main branch")
			opts := &reporter.TestRunsOptions{
				Branch: "main",
				Limit:  20,
			}
			
			response, err := reporterClient.GetTestRuns(ctx, opts)
			
			Expect(err).NotTo(HaveOccurred())
			
			// Verify all results match the branch filter
			for _, testRun := range response.Data {
				Expect(testRun.Branch).To(Equal("main"))
			}
		})

		It("should support pagination with offset", func() {
			By("Testing pagination with offset parameter")
			
			// Get first page
			firstPageOpts := &reporter.TestRunsOptions{
				Limit:  5,
				Offset: 0,
			}
			
			firstPage, err := reporterClient.GetTestRuns(ctx, firstPageOpts)
			Expect(err).NotTo(HaveOccurred())
			
			if firstPage.TotalCount > 5 {
				// Get second page
				secondPageOpts := &reporter.TestRunsOptions{
					Limit:  5,
					Offset: 5,
				}
				
				secondPage, err := reporterClient.GetTestRuns(ctx, secondPageOpts)
				Expect(err).NotTo(HaveOccurred())
				
				// Verify different results
				Expect(firstPage.Data).NotTo(Equal(secondPage.Data))
				
				// Verify offset behavior
				if len(firstPage.Data) > 0 && len(secondPage.Data) > 0 {
					Expect(firstPage.Data[0].ID).NotTo(Equal(secondPage.Data[0].ID))
				}
			}
		})

		It("should return proper response structure", func() {
			By("Verifying test run response structure")
			response, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{Limit: 1})
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(HaveField("Data", Not(BeNil())))
			Expect(response).To(HaveField("TotalCount", BeNumerically(">=", 0)))
			Expect(response).To(HaveField("Page", BeNumerically(">=", 0)))
			Expect(response).To(HaveField("PageSize", BeNumerically(">=", 0)))
			
			if len(response.Data) > 0 {
				testRun := response.Data[0]
				Expect(testRun).To(HaveField("ID", Not(BeEmpty())))
				Expect(testRun).To(HaveField("ProjectID", Not(BeEmpty())))
				Expect(testRun).To(HaveField("Status", Not(BeEmpty())))
				Expect(testRun).To(HaveField("StartTime", Not(BeZero())))
				Expect(testRun).To(HaveField("Duration", BeNumerically(">=", 0)))
				Expect(testRun).To(HaveField("Branch", Not(BeEmpty())))
				Expect(testRun).To(HaveField("Tags", Not(BeNil())))
			}
		})

		It("should handle invalid project ID gracefully", func() {
			By("Testing with invalid project ID")
			opts := &reporter.TestRunsOptions{
				ProjectID: "invalid-project-id",
				Limit:     10,
			}
			
			response, err := reporterClient.GetTestRuns(ctx, opts)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Data)).To(Equal(0))
		})
	})

	Describe("Individual Test Run API", func() {
		It("should retrieve specific test run by ID", func() {
			// Get a test run ID first
			testRuns := testData.TestRuns
			Expect(len(testRuns)).To(BeNumerically(">", 0))
			testRunID := testRuns[0].ID
			
			By(fmt.Sprintf("Fetching test run by ID: %s", testRunID))
			testRun, err := reporterClient.GetTestRun(ctx, testRunID)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(testRun).NotTo(BeNil())
			Expect(testRun.ID).To(Equal(testRunID))
		})

		It("should return 404 for non-existent test run", func() {
			By("Attempting to fetch non-existent test run")
			_, err := reporterClient.GetTestRun(ctx, "non-existent-id")
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should include spec runs in detailed view", func() {
			// Get a test run that has spec runs
			testRuns := testData.TestRuns
			Expect(len(testRuns)).To(BeNumerically(">", 0))
			
			var testRunWithSpecs *reporter.TestRun
			for _, tr := range testRuns {
				if len(tr.SpecRuns) > 0 {
					testRunWithSpecs = &tr
					break
				}
			}
			
			if testRunWithSpecs != nil {
				By(fmt.Sprintf("Fetching test run with spec runs: %s", testRunWithSpecs.ID))
				testRun, err := reporterClient.GetTestRun(ctx, testRunWithSpecs.ID)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(testRun.SpecRuns).NotTo(BeEmpty())
				
				// Verify spec run structure
				for _, specRun := range testRun.SpecRuns {
					Expect(specRun.ID).NotTo(BeEmpty())
					Expect(specRun.SpecDescription).NotTo(BeEmpty())
					Expect(specRun.Status).To(BeElementOf("passed", "failed", "skipped"))
				}
			}
		})
	})

	Describe("Projects API", func() {
		It("should retrieve all projects", func() {
			By("Fetching all projects")
			response, err := reporterClient.GetProjects(ctx)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Data).NotTo(BeEmpty())
			Expect(response.TotalCount).To(BeNumerically(">", 0))
		})

		It("should return proper project structure", func() {
			By("Verifying project response structure")
			response, err := reporterClient.GetProjects(ctx)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Data)).To(BeNumerically(">", 0))
			
			project := response.Data[0]
			Expect(project).To(HaveField("ID", Not(BeEmpty())))
			Expect(project).To(HaveField("Name", Not(BeEmpty())))
			Expect(project).To(HaveField("Description", BeAssignableToTypeOf("")))
			Expect(project).To(HaveField("Tags", Not(BeNil())))
			Expect(project).To(HaveField("CreatedAt", Not(BeZero())))
			
			Expect(project.ID).NotTo(BeEmpty())
			Expect(project.Name).NotTo(BeEmpty())
			Expect(project.Tags).NotTo(BeNil())
		})

		It("should include all created test projects", func() {
			By("Verifying all test projects are returned")
			response, err := reporterClient.GetProjects(ctx)
			
			Expect(err).NotTo(HaveOccurred())
			
			// Get expected project names from test data
			expectedProjects := testData.Projects
			returnedProjectIDs := make(map[string]bool)
			
			for _, project := range response.Data {
				returnedProjectIDs[project.ID] = true
			}
			
			// Verify all test projects are included
			for _, expectedProject := range expectedProjects {
				Expect(returnedProjectIDs).To(HaveKey(expectedProject.ID),
					"Project %s should be in the response", expectedProject.Name)
			}
		})
	})

	Describe("Data Creation APIs", func() {
		It("should create a new test run", func() {
			// Get a project to associate with
			projects := testData.Projects
			Expect(len(projects)).To(BeNumerically(">", 0))
			project := projects[0]
			
			newTestRun := &reporter.TestRun{
				ProjectID: project.ID,
				SuiteID:   "new-suite-id",
				Status:    "passed",
				StartTime: time.Now().Add(-5 * time.Minute),
				Duration:  300000, // 5 minutes
				Branch:    "test-branch",
				Tags:      []string{"acceptance", "api"},
			}
			
			endTime := newTestRun.StartTime.Add(time.Duration(newTestRun.Duration) * time.Millisecond)
			newTestRun.EndTime = &endTime
			
			By("Creating a new test run")
			createdTestRun, err := reporterClient.CreateTestRun(ctx, newTestRun)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTestRun).NotTo(BeNil())
			Expect(createdTestRun.ID).NotTo(BeEmpty())
			Expect(createdTestRun.ProjectID).To(Equal(project.ID))
			Expect(createdTestRun.Status).To(Equal("passed"))
			Expect(createdTestRun.Branch).To(Equal("test-branch"))
		})

		It("should create a new project", func() {
			newProject := &reporter.Project{
				Name:        fmt.Sprintf("api-test-project-%d", time.Now().Unix()),
				Description: "Project created during API acceptance tests",
				Tags:        []string{"api-test", "acceptance"},
			}
			
			By("Creating a new project")
			createdProject, err := reporterClient.CreateProject(ctx, newProject)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(createdProject).NotTo(BeNil())
			Expect(createdProject.ID).NotTo(BeEmpty())
			Expect(createdProject.Name).To(Equal(newProject.Name))
			Expect(createdProject.Description).To(Equal(newProject.Description))
			Expect(createdProject.Tags).To(Equal(newProject.Tags))
			Expect(createdProject.CreatedAt).NotTo(BeZero())
		})

		It("should validate required fields when creating test run", func() {
			invalidTestRun := &reporter.TestRun{
				// Missing required fields like ProjectID
				Status:    "passed",
				StartTime: time.Now(),
			}
			
			By("Attempting to create test run with missing required fields")
			_, err := reporterClient.CreateTestRun(ctx, invalidTestRun)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("400"))
		})

		It("should validate required fields when creating project", func() {
			invalidProject := &reporter.Project{
				// Missing required Name field
				Description: "Project with missing name",
				Tags:        []string{"test"},
			}
			
			By("Attempting to create project with missing required fields")
			_, err := reporterClient.CreateProject(ctx, invalidProject)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("400"))
		})
	})

	Describe("Performance and Load Testing", func() {
		It("should handle concurrent requests efficiently", func() {
			By("Making concurrent API requests")
			
			numConcurrentRequests := 10
			responses := make(chan *reporter.TestRunsResponse, numConcurrentRequests)
			errors := make(chan error, numConcurrentRequests)
			
			startTime := time.Now()
			
			// Launch concurrent requests
			for i := 0; i < numConcurrentRequests; i++ {
				go func() {
					opts := &reporter.TestRunsOptions{Limit: 5}
					response, err := reporterClient.GetTestRuns(ctx, opts)
					if err != nil {
						errors <- err
					} else {
						responses <- response
					}
				}()
			}
			
			// Collect results
			successCount := 0
			errorCount := 0
			
			for i := 0; i < numConcurrentRequests; i++ {
				select {
				case <-responses:
					successCount++
				case <-errors:
					errorCount++
				case <-time.After(10 * time.Second):
					Fail("Concurrent requests timed out")
				}
			}
			
			totalTime := time.Since(startTime)
			
			Expect(successCount).To(Equal(numConcurrentRequests))
			Expect(errorCount).To(Equal(0))
			Expect(totalTime).To(BeNumerically("<", 5*time.Second),
				"Concurrent requests should complete within 5 seconds")
		})

		It("should handle large result sets efficiently", func() {
			By("Requesting large result set")
			opts := &reporter.TestRunsOptions{
				Limit: 100,
			}
			
			startTime := time.Now()
			response, err := reporterClient.GetTestRuns(ctx, opts)
			duration := time.Since(startTime)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Data).NotTo(BeNil())
			Expect(duration).To(BeNumerically("<", 3*time.Second),
				"Large result set should be returned within 3 seconds")
		})

		It("should maintain consistent response times", func() {
			By("Testing response time consistency")
			
			numRequests := 5
			durations := make([]time.Duration, numRequests)
			
			for i := 0; i < numRequests; i++ {
				startTime := time.Now()
				_, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{Limit: 10})
				durations[i] = time.Since(startTime)
				
				Expect(err).NotTo(HaveOccurred())
			}
			
			// Calculate average and maximum duration
			var total time.Duration
			var maxDuration time.Duration
			
			for _, duration := range durations {
				total += duration
				if duration > maxDuration {
					maxDuration = duration
				}
			}
			
			avgDuration := total / time.Duration(numRequests)
			
			Expect(avgDuration).To(BeNumerically("<", 1*time.Second),
				"Average response time should be under 1 second")
			Expect(maxDuration).To(BeNumerically("<", 2*time.Second),
				"Maximum response time should be under 2 seconds")
		})
	})

	Describe("Error Handling", func() {
		It("should handle service unavailable gracefully", func() {
			// Create client with invalid URL to simulate service unavailable
			invalidClient, err := reporter.NewClient("http://invalid-url:9999")
			Expect(err).NotTo(HaveOccurred())
			
			By("Testing with unavailable service")
			_, err = invalidClient.GetTestRuns(ctx, nil)
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("request failed"))
		})

		It("should timeout appropriately for slow requests", func() {
			// Create client with very short timeout
			shortTimeoutClient := reporterClient.WithTimeout(100 * time.Millisecond)
			
			By("Testing request timeout")
			startTime := time.Now()
			_, err := shortTimeoutClient.GetTestRuns(ctx, &reporter.TestRunsOptions{Limit: 100})
			duration := time.Since(startTime)
			
			// Should either succeed quickly or timeout appropriately
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("timeout"))
				Expect(duration).To(BeNumerically(">=", 100*time.Millisecond))
				Expect(duration).To(BeNumerically("<", 5*time.Second))
			}
		})

		It("should handle malformed request data", func() {
			By("Testing with invalid test run data")
			invalidTestRun := &reporter.TestRun{
				ProjectID: "invalid-project",
				Status:    "invalid-status",
				StartTime: time.Time{}, // Invalid zero time
			}
			
			_, err := reporterClient.CreateTestRun(ctx, invalidTestRun)
			Expect(err).To(HaveOccurred())
		})
	})
})