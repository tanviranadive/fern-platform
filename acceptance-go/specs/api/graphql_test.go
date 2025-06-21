package api_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/graphql"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/fixtures"
)

var _ = Describe("GraphQL API", func() {
	var (
		ctx           context.Context
		graphqlClient *graphql.Client
		testData      *fixtures.CreatedTestData
	)

	BeforeEach(func() {
		ctx = GetTestContext()
		graphqlClient = GetGraphQLClient()
		testData = GetTestData()
	})

	Describe("Schema and Introspection", func() {
		It("should expose valid GraphQL schema", func() {
			By("Performing introspection query")
			response, err := graphqlClient.Introspect(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			Expect(response.Data).NotTo(BeNil())
			
			// Verify schema structure
			schemaData, ok := response.Data.(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(schemaData).To(HaveKey("__schema"))
			
			schema := schemaData["__schema"].(map[string]interface{})
			Expect(schema).To(HaveKey("types"))
			
			types, ok := schema["types"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(len(types)).To(BeNumerically(">", 0))
		})

		It("should have required types defined", func() {
			By("Getting schema information")
			schema, err := graphqlClient.GetSchema(ctx)
			Expect(err).NotTo(HaveOccurred())
			
			requiredTypes := []string{
				"Query",
				"TestRun", 
				"SpecRun",
				"Project",
				"SuiteRun",
			}
			
			typeNames := make([]string, len(schema.Types))
			for i, t := range schema.Types {
				typeNames[i] = t.Name
			}
			
			By("Verifying all required types exist")
			for _, typeName := range requiredTypes {
				Expect(typeNames).To(ContainElement(typeName), 
					"Schema should contain type: %s", typeName)
			}
		})

		It("should have proper field definitions for TestRun type", func() {
			By("Getting TestRun type information")
			testRunType, err := graphqlClient.GetType(ctx, "TestRun")
			Expect(err).NotTo(HaveOccurred())
			Expect(testRunType).NotTo(BeNil())
			
			requiredFields := []string{
				"id",
				"projectId", 
				"suiteId",
				"status",
				"startTime",
				"endTime", 
				"duration",
				"branch",
				"tags",
				"specRuns",
			}
			
			fieldNames := make([]string, len(testRunType.Fields))
			for i, field := range testRunType.Fields {
				fieldNames[i] = field.Name
			}
			
			By("Verifying all required fields exist")
			for _, fieldName := range requiredFields {
				Expect(fieldNames).To(ContainElement(fieldName),
					"TestRun type should have field: %s", fieldName)
			}
		})

		It("should support proper enum types", func() {
			By("Getting TestStatus enum type")
			statusEnum, err := graphqlClient.GetType(ctx, "TestStatus")
			
			if err == nil && statusEnum != nil && len(statusEnum.EnumValues) > 0 {
				enumValues := make([]string, len(statusEnum.EnumValues))
				for i, value := range statusEnum.EnumValues {
					enumValues[i] = value.Name
				}
				
				Expect(enumValues).To(ContainElement("PASSED"))
				Expect(enumValues).To(ContainElement("FAILED"))
				Expect(enumValues).To(ContainElement("SKIPPED"))
			} else {
				Skip("TestStatus enum not implemented yet")
			}
		})
	})

	Describe("Test Run Queries", func() {
		It("should fetch test runs with basic query", func() {
			query := `
				query GetTestRuns {
					testRuns(limit: 10) {
						id
						projectId
						suiteId
						status
						startTime
						endTime
						duration
						branch
					}
				}
			`
			
			By("Executing basic test runs query")
			startTime := time.Now()
			response, err := graphqlClient.Query(ctx, query)
			queryDuration := time.Since(startTime)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			Expect(queryDuration).To(BeNumerically("<", 2*time.Second), 
				"Query should complete within 2 seconds")
			
			// Verify response structure
			Expect(response.Data).NotTo(BeNil())
			data, ok := response.Data.(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(data).To(HaveKey("testRuns"))
			
			testRuns, ok := data["testRuns"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(len(testRuns)).To(BeNumerically(">=", 0))
			
			// Verify test run structure if any exist
			if len(testRuns) > 0 {
				testRun := testRuns[0].(map[string]interface{})
				Expect(testRun).To(HaveKey("id"))
				Expect(testRun).To(HaveKey("projectId"))
				Expect(testRun).To(HaveKey("status"))
				Expect(testRun).To(HaveKey("startTime"))
			}
		})

		It("should support pagination with cursor-based navigation", func() {
			firstPageQuery := `
				query GetFirstPage {
					testRuns(limit: 5) {
						id
						startTime
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
					}
				}
			`
			
			By("Fetching first page of test runs")
			firstPage, err := graphqlClient.Query(ctx, firstPageQuery)
			Expect(err).NotTo(HaveOccurred())
			Expect(firstPage.Errors).To(BeEmpty())
			
			data := firstPage.Data.(map[string]interface{})
			testRuns := data["testRuns"].(map[string]interface{})
			
			if pageInfo, exists := testRuns["pageInfo"]; exists {
				pageInfoMap := pageInfo.(map[string]interface{})
				
				if hasNextPage, ok := pageInfoMap["hasNextPage"].(bool); ok && hasNextPage {
					endCursor := pageInfoMap["endCursor"].(string)
					
					secondPageQuery := fmt.Sprintf(`
						query GetSecondPage {
							testRuns(limit: 5, after: "%s") {
								id
								startTime
								pageInfo {
									hasNextPage
									hasPreviousPage
									startCursor
									endCursor
								}
							}
						}
					`, endCursor)
					
					By("Fetching second page using cursor")
					secondPage, err := graphqlClient.Query(ctx, secondPageQuery)
					Expect(err).NotTo(HaveOccurred())
					Expect(secondPage.Errors).To(BeEmpty())
					
					// Verify different results
					firstPageData := data["testRuns"].([]interface{})
					secondPageData := secondPage.Data.(map[string]interface{})["testRuns"].([]interface{})
					
					Expect(firstPageData).NotTo(Equal(secondPageData), 
						"Second page should have different results")
				}
			}
		})

		It("should filter test runs by project", func() {
			// Get a test project
			projects := testData.Projects
			Expect(len(projects)).To(BeNumerically(">", 0))
			testProject := projects[0]
			
			query := `
				query GetTestRunsByProject($projectId: ID!) {
					testRuns(projectId: $projectId, limit: 10) {
						id
						projectId
						status
					}
				}
			`
			
			variables := map[string]interface{}{
				"projectId": testProject.ID,
			}
			
			By(fmt.Sprintf("Filtering test runs by project: %s", testProject.Name))
			response, err := graphqlClient.Query(ctx, query, variables)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			
			data := response.Data.(map[string]interface{})
			testRuns, ok := data["testRuns"].([]interface{})
			Expect(ok).To(BeTrue())
			
			// Verify all results match the filter
			for _, tr := range testRuns {
				testRun := tr.(map[string]interface{})
				Expect(testRun["projectId"]).To(Equal(testProject.ID))
			}
		})

		It("should filter test runs by status", func() {
			query := `
				query GetFailedTestRuns {
					testRuns(status: FAILED, limit: 10) {
						id
						status
						projectId
					}
				}
			`
			
			By("Filtering test runs by failed status")
			response, err := graphqlClient.Query(ctx, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			
			data := response.Data.(map[string]interface{})
			testRuns, ok := data["testRuns"].([]interface{})
			Expect(ok).To(BeTrue())
			
			// Verify all results have failed status
			for _, tr := range testRuns {
				testRun := tr.(map[string]interface{})
				Expect(testRun["status"]).To(Equal("FAILED"))
			}
		})

		It("should filter test runs by date range", func() {
			endDate := time.Now()
			startDate := endDate.Add(-7 * 24 * time.Hour) // 7 days ago
			
			query := `
				query GetTestRunsByDateRange($startDate: DateTime!, $endDate: DateTime!) {
					testRuns(
						startTimeAfter: $startDate,
						startTimeBefore: $endDate,
						limit: 20
					) {
						id
						startTime
						endTime
					}
				}
			`
			
			variables := map[string]interface{}{
				"startDate": startDate.Format(time.RFC3339),
				"endDate":   endDate.Format(time.RFC3339),
			}
			
			By("Filtering test runs by date range")
			response, err := graphqlClient.Query(ctx, query, variables)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			
			data := response.Data.(map[string]interface{})
			testRuns, ok := data["testRuns"].([]interface{})
			Expect(ok).To(BeTrue())
			
			// Verify all results are within date range
			for _, tr := range testRuns {
				testRun := tr.(map[string]interface{})
				startTimeStr := testRun["startTime"].(string)
				testStartTime, err := time.Parse(time.RFC3339, startTimeStr)
				Expect(err).NotTo(HaveOccurred())
				
				Expect(testStartTime).To(BeTemporally(">=", startDate))
				Expect(testStartTime).To(BeTemporally("<=", endDate))
			}
		})
	})

	Describe("Complex Queries and Performance", func() {
		It("should handle complex nested queries efficiently", func() {
			complexQuery := `
				query GetComplexData {
					testRuns(limit: 5) {
						id
						projectId
						status
						duration
						project {
							id
							name
							tags
						}
						specRuns {
							id
							specDescription
							status
							duration
							errorMessage
						}
						suiteRun {
							id
							suiteName
							totalSpecs
							passedSpecs
							failedSpecs
						}
					}
				}
			`
			
			By("Executing complex nested query")
			startTime := time.Now()
			response, err := graphqlClient.Query(ctx, complexQuery)
			queryDuration := time.Since(startTime)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			Expect(queryDuration).To(BeNumerically("<", 5*time.Second),
				"Complex query should complete within 5 seconds")
			
			data := response.Data.(map[string]interface{})
			testRuns, ok := data["testRuns"].([]interface{})
			Expect(ok).To(BeTrue())
			
			// Verify nested data is populated
			for _, tr := range testRuns {
				testRun := tr.(map[string]interface{})
				
				if project, exists := testRun["project"]; exists && project != nil {
					projectMap := project.(map[string]interface{})
					Expect(projectMap).To(HaveKey("id"))
					Expect(projectMap).To(HaveKey("name"))
				}
				
				if specRuns, exists := testRun["specRuns"]; exists && specRuns != nil {
					specRunsArray := specRuns.([]interface{})
					Expect(len(specRunsArray)).To(BeNumerically(">=", 0))
				}
			}
		})

		It("should implement proper N+1 query prevention", func() {
			query := `
				query GetTestRunsWithProjects {
					testRuns(limit: 10) {
						id
						projectId
						project {
							id
							name
						}
					}
				}
			`
			
			By("Executing query that could cause N+1 problem")
			startTime := time.Now()
			response, err := graphqlClient.Query(ctx, query)
			queryDuration := time.Since(startTime)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Errors).To(BeEmpty())
			Expect(queryDuration).To(BeNumerically("<", 3*time.Second),
				"Query with nested projects should be fast (no N+1)")
			
			data := response.Data.(map[string]interface{})
			testRuns, ok := data["testRuns"].([]interface{})
			Expect(ok).To(BeTrue())
			
			// Verify project data is correctly joined
			for _, tr := range testRuns {
				testRun := tr.(map[string]interface{})
				if project, exists := testRun["project"]; exists && project != nil {
					projectMap := project.(map[string]interface{})
					Expect(projectMap["id"]).To(Equal(testRun["projectId"]))
					Expect(projectMap["name"]).To(BeAssignableToTypeOf(""))
				}
			}
		})
	})

	Describe("Error Handling and Validation", func() {
		It("should return proper error for invalid query syntax", func() {
			invalidQuery := `
				query InvalidSyntax {
					testRuns {
						id
						invalidField {
							nonExistentNestedField
						}
					}
				}
			`
			
			By("Executing query with invalid syntax")
			response, err := graphqlClient.Query(ctx, invalidQuery)
			Expect(err).NotTo(HaveOccurred()) // HTTP request should succeed
			
			Expect(response.Errors).NotTo(BeEmpty())
			Expect(response.Errors[0].Message).To(ContainSubstring("Cannot query field"))
		})

		It("should validate required arguments", func() {
			queryWithMissingArgs := `
				query MissingArgs {
					testRun {
						id
					}
				}
			`
			
			By("Executing query with missing required arguments")
			response, err := graphqlClient.Query(ctx, queryWithMissingArgs)
			Expect(err).NotTo(HaveOccurred())
			
			Expect(response.Errors).NotTo(BeEmpty())
			Expect(response.Errors[0].Message).To(ContainSubstring("required"))
		})

		It("should handle invalid variable types", func() {
			query := `
				query InvalidVariableType($limit: String!) {
					testRuns(limit: $limit) {
						id
					}
				}
			`
			
			variables := map[string]interface{}{
				"limit": "not_a_number",
			}
			
			By("Executing query with invalid variable type")
			response, err := graphqlClient.Query(ctx, query, variables)
			Expect(err).NotTo(HaveOccurred())
			
			Expect(response.Errors).NotTo(BeEmpty())
			Expect(response.Errors[0].Message).To(ContainSubstring("type"))
		})

		It("should return proper error for non-existent resources", func() {
			query := `
				query GetNonExistentTestRun($id: ID!) {
					testRun(id: $id) {
						id
						status
					}
				}
			`
			
			variables := map[string]interface{}{
				"id": "non-existent-id",
			}
			
			By("Querying for non-existent test run")
			response, err := graphqlClient.Query(ctx, query, variables)
			Expect(err).NotTo(HaveOccurred())
			
			// Should either return null or proper error
			if response.Errors == nil || len(response.Errors) == 0 {
				data := response.Data.(map[string]interface{})
				Expect(data["testRun"]).To(BeNil())
			}
		})

		It("should handle malformed JSON gracefully", func() {
			By("Sending malformed JSON request")
			_, err := graphqlClient.RawRequest(ctx, "{ invalid json }")
			
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON"))
		})
	})

	Describe("Security and Authorization", func() {
		It("should prevent SQL injection through GraphQL variables", func() {
			maliciousQuery := `
				query SQLInjectionAttempt($projectId: ID!) {
					testRuns(projectId: $projectId) {
						id
					}
				}
			`
			
			maliciousProjectId := "'; DROP TABLE test_runs; --"
			variables := map[string]interface{}{
				"projectId": maliciousProjectId,
			}
			
			By("Attempting SQL injection through GraphQL variables")
			response, err := graphqlClient.Query(ctx, maliciousQuery, variables)
			Expect(err).NotTo(HaveOccurred())
			
			// Should not cause SQL injection - either empty results or validation error
			if response.Errors == nil || len(response.Errors) == 0 {
				data := response.Data.(map[string]interface{})
				testRuns := data["testRuns"].([]interface{})
				Expect(len(testRuns)).To(Equal(0))
			}
		})
	})

	Describe("Real-time Features (Future)", func() {
		It("should support subscriptions for real-time updates", func() {
			By("Checking if subscription support is available")
			hasSubscriptionSupport, err := graphqlClient.HasSubscriptionSupport(ctx)
			Expect(err).NotTo(HaveOccurred())
			
			if hasSubscriptionSupport {
				// Test subscription functionality when implemented
				Skip("Subscription functionality not yet implemented")
			} else {
				Skip("Subscriptions not yet implemented")
			}
		})
	})
})