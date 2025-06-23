package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/reporter"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/fixtures"
)

var _ = Describe("End-to-End Workflows", func() {
	var (
		ctx            context.Context
		reporterClient *reporter.Client
		testData       *fixtures.CreatedTestData
		serviceURLs    map[string]string
	)

	BeforeEach(func() {
		ctx = GetTestContext()
		reporterClient = GetReporterClient()
		testData = GetTestData()
		serviceURLs = GetServiceURLs()
	})

	Describe("Complete Test Run Lifecycle", func() {
		It("should support the full test run creation to analysis workflow", func() {
			By("Creating a new project")
			project := &reporter.Project{
				Name:        fmt.Sprintf("e2e-test-project-%d", time.Now().Unix()),
				Description: "End-to-end test project for workflow validation",
				Tags:        []string{"e2e", "workflow", "integration"},
			}
			
			createdProject, err := reporterClient.CreateProject(ctx, project)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdProject.ID).NotTo(BeEmpty())
			
			By("Creating test runs for the project")
			testRuns := make([]*reporter.TestRun, 3)
			
			for i := 0; i < 3; i++ {
				testRun := &reporter.TestRun{
					ProjectID: createdProject.ID,
					SuiteID:   fmt.Sprintf("suite-%d", i+1),
					Status:    []string{"passed", "failed", "passed"}[i],
					StartTime: time.Now().Add(-time.Duration(i+1) * time.Hour),
					Duration:  int64((i + 1) * 60000), // 1-3 minutes
					Branch:    "main",
					Tags:      []string{"e2e", fmt.Sprintf("run-%d", i+1)},
				}
				
				endTime := testRun.StartTime.Add(time.Duration(testRun.Duration) * time.Millisecond)
				testRun.EndTime = &endTime
				
				createdTestRun, err := reporterClient.CreateTestRun(ctx, testRun)
				Expect(err).NotTo(HaveOccurred())
				testRuns[i] = createdTestRun
			}
			
			By("Verifying test runs appear in API queries")
			Eventually(func() int {
				response, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{
					ProjectID: createdProject.ID,
					Limit:     10,
				})
				Expect(err).NotTo(HaveOccurred())
				return len(response.Data)
			}, 30*time.Second, 2*time.Second).Should(BeNumerically(">=", 3))
			
			By("Verifying test runs appear in GraphQL queries")
			query := `
				query GetProjectTestRuns($projectId: ID!) {
					testRuns(projectId: $projectId, limit: 10) {
						id
						status
						projectId
						project {
							id
							name
						}
					}
				}
			`
			
			Eventually(func() int {
				response, err := graphqlClient.Query(ctx, query, map[string]interface{}{
					"projectId": createdProject.ID,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Errors).To(BeEmpty())
				
				data := response.Data.(map[string]interface{})
				testRuns := data["testRuns"].([]interface{})
				return len(testRuns)
			}, 30*time.Second, 2*time.Second).Should(BeNumerically(">=", 3))
			
			By("Verifying project statistics are calculated correctly")
			projectsResponse, err := reporterClient.GetProjects(ctx)
			Expect(err).NotTo(HaveOccurred())
			
			var targetProject *reporter.Project
			for _, p := range projectsResponse.Data {
				if p.ID == createdProject.ID {
					targetProject = &p
					break
				}
			}
			
			Expect(targetProject).NotTo(BeNil())
			Expect(targetProject.Name).To(Equal(createdProject.Name))
		})

		It("should maintain data consistency across REST and GraphQL APIs", func() {
			By("Getting test data from REST API")
			restResponse, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{
				Limit: 10,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(restResponse.Data)).To(BeNumerically(">", 0))
			
			firstTestRunRest := restResponse.Data[0]
			
			By("Getting the same test run via GraphQL")
			query := `
				query GetTestRun($id: ID!) {
					testRun(id: $id) {
						id
						projectId
						status
						startTime
						endTime
						duration
						branch
						tags
					}
				}
			`
			
			graphqlResponse, err := graphqlClient.Query(ctx, query, map[string]interface{}{
				"id": firstTestRunRest.ID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(graphqlResponse.Errors).To(BeEmpty())
			
			data := graphqlResponse.Data.(map[string]interface{})
			testRunGraphQL := data["testRun"].(map[string]interface{})
			
			By("Verifying data consistency between APIs")
			Expect(testRunGraphQL["id"]).To(Equal(firstTestRunRest.ID))
			Expect(testRunGraphQL["projectId"]).To(Equal(firstTestRunRest.ProjectID))
			Expect(testRunGraphQL["status"]).To(Equal(firstTestRunRest.Status))
			Expect(testRunGraphQL["branch"]).To(Equal(firstTestRunRest.Branch))
		})
	})

	Describe("Cross-Service Integration", func() {
		It("should handle service dependencies correctly", func() {
			By("Verifying all services are running and healthy")
			
			// Check reporter service
			err := reporterClient.HealthCheck(ctx)
			Expect(err).NotTo(HaveOccurred())
			
			// Check if UI service is accessible (if available)
			if uiURL, exists := serviceURLs["fern-ui"]; exists {
				By("Verifying UI service is accessible")
				
				// Simple HTTP check to UI service
				resp, err := http.Get(uiURL)
				if err == nil {
					defer resp.Body.Close()
					Expect(resp.StatusCode).To(BeNumerically("<", 500))
				}
			}
			
			// Check if mycelium service is accessible (if available)
			if myceliumURL, exists := serviceURLs["fern-mycelium"]; exists {
				By("Verifying Mycelium service is accessible")
				
				// Simple HTTP check to mycelium service
				resp, err := http.Get(myceliumURL + "/health")
				if err == nil {
					defer resp.Body.Close()
					Expect(resp.StatusCode).To(BeNumerically("<", 500))
				}
			}
		})

		It("should handle database transactions correctly", func() {
			By("Creating test data that requires transactional consistency")
			
			project := &reporter.Project{
				Name:        fmt.Sprintf("transaction-test-%d", time.Now().Unix()),
				Description: "Project for testing transactional consistency",
				Tags:        []string{"transaction", "test"},
			}
			
			createdProject, err := reporterClient.CreateProject(ctx, project)
			Expect(err).NotTo(HaveOccurred())
			
			// Create multiple test runs in quick succession
			testRunIDs := make([]string, 5)
			for i := 0; i < 5; i++ {
				testRun := &reporter.TestRun{
					ProjectID: createdProject.ID,
					SuiteID:   fmt.Sprintf("concurrent-suite-%d", i),
					Status:    "passed",
					StartTime: time.Now().Add(-time.Duration(i) * time.Minute),
					Duration:  60000, // 1 minute
					Branch:    "main",
					Tags:      []string{"concurrent"},
				}
				
				endTime := testRun.StartTime.Add(time.Duration(testRun.Duration) * time.Millisecond)
				testRun.EndTime = &endTime
				
				createdTestRun, err := reporterClient.CreateTestRun(ctx, testRun)
				Expect(err).NotTo(HaveOccurred())
				testRunIDs[i] = createdTestRun.ID
			}
			
			By("Verifying all test runs are retrievable and consistent")
			Eventually(func() bool {
				response, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{
					ProjectID: createdProject.ID,
					Limit:     10,
				})
				if err != nil {
					return false
				}
				
				// Verify all test runs are present
				foundIDs := make(map[string]bool)
				for _, tr := range response.Data {
					foundIDs[tr.ID] = true
				}
				
				for _, id := range testRunIDs {
					if !foundIDs[id] {
						return false
					}
				}
				
				return len(response.Data) >= 5
			}, 30*time.Second, 2*time.Second).Should(BeTrue())
		})
	})

	Describe("Performance and Scalability", func() {
		It("should handle high volume of concurrent requests", func() {
			By("Generating concurrent load on the system")
			
			numConcurrentUsers := 20
			numRequestsPerUser := 5
			
			results := make(chan error, numConcurrentUsers*numRequestsPerUser)
			
			// Simulate concurrent users
			for user := 0; user < numConcurrentUsers; user++ {
				go func(userID int) {
					defer GinkgoRecover()
					
					for req := 0; req < numRequestsPerUser; req++ {
						// Mix of different API calls
						switch req % 3 {
						case 0:
							_, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{
								Limit: 10,
							})
							results <- err
						case 1:
							_, err := reporterClient.GetProjects(ctx)
							results <- err
						case 2:
							query := `query { testRuns(limit: 5) { id status } }`
							_, err := graphqlClient.Query(ctx, query)
							results <- err
						}
					}
				}(user)
			}
			
			By("Collecting results from concurrent requests")
			successCount := 0
			errorCount := 0
			totalRequests := numConcurrentUsers * numRequestsPerUser
			
			for i := 0; i < totalRequests; i++ {
				select {
				case err := <-results:
					if err != nil {
						errorCount++
					} else {
						successCount++
					}
				case <-time.After(30 * time.Second):
					Fail(fmt.Sprintf("Timeout waiting for concurrent requests. Processed %d/%d", i, totalRequests))
				}
			}
			
			By("Verifying system handled concurrent load")
			successRate := float64(successCount) / float64(totalRequests)
			Expect(successRate).To(BeNumerically(">=", 0.95), 
				"At least 95%% of requests should succeed under load")
			
			fmt.Printf("Load test results: %d/%d requests succeeded (%.1f%%)\n", 
				successCount, totalRequests, successRate*100)
		})

		It("should maintain response times under load", func() {
			By("Measuring response times under sustained load")
			
			// Warm up the system
			for i := 0; i < 5; i++ {
				_, _ = reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{Limit: 10})
			}
			
			// Measure response times
			numRequests := 10
			responseTimes := make([]time.Duration, numRequests)
			
			for i := 0; i < numRequests; i++ {
				startTime := time.Now()
				_, err := reporterClient.GetTestRuns(ctx, &reporter.TestRunsOptions{
					Limit: 20,
				})
				responseTimes[i] = time.Since(startTime)
				
				Expect(err).NotTo(HaveOccurred())
			}
			
			// Calculate statistics
			var total time.Duration
			var maxTime time.Duration
			
			for _, duration := range responseTimes {
				total += duration
				if duration > maxTime {
					maxTime = duration
				}
			}
			
			avgTime := total / time.Duration(numRequests)
			
			By("Verifying response time requirements")
			Expect(avgTime).To(BeNumerically("<", 1*time.Second),
				"Average response time should be under 1 second")
			Expect(maxTime).To(BeNumerically("<", 3*time.Second),
				"Maximum response time should be under 3 seconds")
			
			fmt.Printf("Response time stats: avg=%.0fms, max=%.0fms\n", 
				float64(avgTime.Nanoseconds())/1e6, float64(maxTime.Nanoseconds())/1e6)
		})
	})

	Describe("Data Integrity and Validation", func() {
		It("should enforce referential integrity", func() {
			By("Attempting to create test run with invalid project ID")
			invalidTestRun := &reporter.TestRun{
				ProjectID: "non-existent-project-id",
				SuiteID:   "test-suite",
				Status:    "passed",
				StartTime: time.Now(),
				Duration:  60000,
				Branch:    "main",
			}
			
			_, err := reporterClient.CreateTestRun(ctx, invalidTestRun)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("400"))
		})

		It("should validate data consistency across operations", func() {
			By("Creating a project and verifying it appears in all queries")
			project := &reporter.Project{
				Name:        fmt.Sprintf("consistency-test-%d", time.Now().Unix()),
				Description: "Project for testing data consistency",
				Tags:        []string{"consistency", "validation"},
			}
			
			createdProject, err := reporterClient.CreateProject(ctx, project)
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying project appears in REST API")
			Eventually(func() bool {
				response, err := reporterClient.GetProjects(ctx)
				if err != nil {
					return false
				}
				
				for _, p := range response.Data {
					if p.ID == createdProject.ID {
						return true
					}
				}
				return false
			}, 10*time.Second, 1*time.Second).Should(BeTrue())
			
			By("Verifying project appears in GraphQL API")
			query := `
				query GetProjects {
					projects {
						id
						name
					}
				}
			`
			
			Eventually(func() bool {
				response, err := graphqlClient.Query(ctx, query)
				if err != nil || len(response.Errors) > 0 {
					return false
				}
				
				data := response.Data.(map[string]interface{})
				projects := data["projects"].([]interface{})
				
				for _, p := range projects {
					project := p.(map[string]interface{})
					if project["id"] == createdProject.ID {
						return true
					}
				}
				return false
			}, 10*time.Second, 1*time.Second).Should(BeTrue())
		})
	})

	Describe("Error Recovery and Resilience", func() {
		It("should handle partial failures gracefully", func() {
			By("Testing system behavior with edge case data")
			
			// Test with various edge cases
			edgeCases := []*reporter.TestRun{
				{
					ProjectID: testData.Projects[0].ID,
					SuiteID:   "edge-case-1",
					Status:    "passed",
					StartTime: time.Now().Add(-24 * time.Hour), // Very old timestamp
					Duration:  1, // Very short duration
					Branch:    "feature/with-special-chars!@#$%",
					Tags:      []string{"edge", "special-chars"},
				},
				{
					ProjectID: testData.Projects[0].ID,
					SuiteID:   "edge-case-2",
					Status:    "failed",
					StartTime: time.Now(),
					Duration:  3600000, // Very long duration (1 hour)
					Branch:    "main",
					Tags:      []string{}, // Empty tags
				},
			}
			
			for i, testRun := range edgeCases {
				endTime := testRun.StartTime.Add(time.Duration(testRun.Duration) * time.Millisecond)
				testRun.EndTime = &endTime
				
				By(fmt.Sprintf("Testing edge case %d", i+1))
				createdTestRun, err := reporterClient.CreateTestRun(ctx, testRun)
				
				if err != nil {
					// If creation fails, ensure it's for a valid reason
					Expect(err.Error()).To(ContainSubstring("400"))
				} else {
					// If creation succeeds, ensure data is retrievable
					retrievedTestRun, err := reporterClient.GetTestRun(ctx, createdTestRun.ID)
					Expect(err).NotTo(HaveOccurred())
					Expect(retrievedTestRun.ID).To(Equal(createdTestRun.ID))
				}
			}
		})

		It("should maintain service availability during data operations", func() {
			By("Performing data operations while monitoring service health")
			
			// Start background health checks
			healthCheckDone := make(chan bool)
			healthCheckErrors := make(chan error, 10)
			
			go func() {
				defer GinkgoRecover()
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				
				for {
					select {
					case <-ticker.C:
						err := reporterClient.HealthCheck(ctx)
						if err != nil {
							healthCheckErrors <- err
						}
					case <-healthCheckDone:
						return
					}
				}
			}()
			
			// Perform intensive data operations
			By("Creating multiple projects and test runs")
			for i := 0; i < 10; i++ {
				project := &reporter.Project{
					Name:        fmt.Sprintf("availability-test-%d-%d", i, time.Now().Unix()),
					Description: fmt.Sprintf("Project %d for availability testing", i),
					Tags:        []string{"availability", "stress"},
				}
				
				createdProject, err := reporterClient.CreateProject(ctx, project)
				Expect(err).NotTo(HaveOccurred())
				
				// Create test runs for each project
				for j := 0; j < 3; j++ {
					testRun := &reporter.TestRun{
						ProjectID: createdProject.ID,
						SuiteID:   fmt.Sprintf("availability-suite-%d-%d", i, j),
						Status:    []string{"passed", "failed", "skipped"}[j%3],
						StartTime: time.Now().Add(-time.Duration(j) * time.Minute),
						Duration:  int64((j + 1) * 30000),
						Branch:    "main",
						Tags:      []string{"availability"},
					}
					
					endTime := testRun.StartTime.Add(time.Duration(testRun.Duration) * time.Millisecond)
					testRun.EndTime = &endTime
					
					_, err := reporterClient.CreateTestRun(ctx, testRun)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			
			// Stop health checks
			close(healthCheckDone)
			
			// Verify no health check failures occurred
			select {
			case err := <-healthCheckErrors:
				Fail(fmt.Sprintf("Service became unhealthy during data operations: %v", err))
			default:
				// No health check errors - good!
			}
		})
	})
})