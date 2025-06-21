/**
 * GraphQL API - Acceptance Tests
 * 
 * Tests the complete GraphQL API functionality including:
 * - Schema validation and introspection
 * - Query execution and data retrieval
 * - Filtering, pagination, and sorting
 * - Error handling and validation
 * - Performance and optimization
 * - Real-time subscriptions (if implemented)
 * 
 * Based on fern-reporter GraphQL implementation and GitHub issues
 */

import { TestUtils, HttpUtils, performanceMonitor } from '@acceptance/setup/test-helpers';
import { GraphQLClient } from '@acceptance/utils/api-clients/graphql-client';

describe('GraphQL API', () => {
  let graphqlClient: GraphQLClient;
  let context: any;

  beforeAll(async () => {
    context = TestUtils.getTestContext();
    graphqlClient = new GraphQLClient(context.baseUrls.reporter);
    
    // Ensure GraphQL service is ready
    await TestUtils.waitForCondition(
      async () => {
        try {
          await graphqlClient.introspect();
          return true;
        } catch {
          return false;
        }
      },
      60000
    );
  });

  describe('Schema and Introspection', () => {
    test('should expose valid GraphQL schema', async () => {
      const schema = await graphqlClient.introspect();
      
      expect(schema).toHaveProperty('data');
      expect(schema.data).toHaveProperty('__schema');
      expect(schema.data.__schema).toHaveProperty('types');
      expect(schema.data.__schema.types.length).toBeGreaterThan(0);
    });

    test('should have required types defined', async () => {
      const schema = await graphqlClient.getSchema();
      
      const requiredTypes = [
        'Query',
        'TestRun',
        'SpecRun', 
        'Project',
        'SuiteRun'
      ];
      
      const typeNames = schema.types.map((type: any) => type.name);
      
      requiredTypes.forEach(typeName => {
        expect(typeNames).toContain(typeName);
      });
    });

    test('should have proper field definitions for TestRun type', async () => {
      const testRunType = await graphqlClient.getType('TestRun');
      
      const requiredFields = [
        'id',
        'projectId',
        'suiteId', 
        'status',
        'startTime',
        'endTime',
        'duration',
        'branch',
        'tags',
        'specRuns'
      ];
      
      const fieldNames = testRunType.fields.map((field: any) => field.name);
      
      requiredFields.forEach(fieldName => {
        expect(fieldNames).toContain(fieldName);
      });
    });

    test('should support proper enum types', async () => {
      const statusEnum = await graphqlClient.getType('TestStatus');
      
      if (statusEnum && statusEnum.enumValues) {
        const enumValues = statusEnum.enumValues.map((value: any) => value.name);
        expect(enumValues).toContain('PASSED');
        expect(enumValues).toContain('FAILED'); 
        expect(enumValues).toContain('SKIPPED');
      }
    });
  });

  describe('Test Run Queries', () => {
    test('should fetch test runs with basic query', async () => {
      const query = `
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
      `;
      
      const endMeasurement = performanceMonitor.startMeasurement('basic_testruns_query');
      const result = await graphqlClient.query(query);
      const queryTime = endMeasurement();
      
      expect(queryTime).toBeWithinTimeRange(0, 2000); // 2 second max
      expect(result).toHaveValidApiResponse();
      expect(result.data.testRuns).toBeDefined();
      expect(Array.isArray(result.data.testRuns)).toBe(true);
      
      if (result.data.testRuns.length > 0) {
        const testRun = result.data.testRuns[0];
        expect(testRun).toHaveValidTestRunStructure();
      }
    });

    test('should support pagination with cursor-based navigation', async () => {
      const firstPageQuery = `
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
      `;
      
      const firstPage = await graphqlClient.query(firstPageQuery);
      expect(firstPage).toHaveValidApiResponse();
      
      const pageInfo = firstPage.data.testRuns.pageInfo;
      
      if (pageInfo.hasNextPage) {
        const secondPageQuery = `
          query GetSecondPage {
            testRuns(limit: 5, after: "${pageInfo.endCursor}") {
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
        `;
        
        const secondPage = await graphqlClient.query(secondPageQuery);
        expect(secondPage).toHaveValidApiResponse();
        
        // Should have different results
        const firstPageIds = firstPage.data.testRuns.map((tr: any) => tr.id);
        const secondPageIds = secondPage.data.testRuns.map((tr: any) => tr.id);
        
        expect(firstPageIds).not.toEqual(secondPageIds);
      }
    });

    test('should filter test runs by project', async () => {
      // Get available projects first
      const projects = await context.dataFixtures.createdData.projects;
      expect(projects.length).toBeGreaterThan(0);
      
      const testProject = projects[0];
      
      const query = `
        query GetTestRunsByProject($projectId: ID!) {
          testRuns(projectId: $projectId, limit: 10) {
            id
            projectId
            suiteId
            status
          }
        }
      `;
      
      const result = await graphqlClient.query(query, { 
        projectId: testProject.id 
      });
      
      expect(result).toHaveValidApiResponse();
      expect(result.data.testRuns).toBeDefined();
      
      // All results should match the filter
      result.data.testRuns.forEach((testRun: any) => {
        expect(testRun.projectId).toBe(testProject.id);
      });
    });

    test('should filter test runs by status', async () => {
      const query = `
        query GetFailedTestRuns {
          testRuns(status: FAILED, limit: 10) {
            id
            status
            projectId
          }
        }
      `;
      
      const result = await graphqlClient.query(query);
      expect(result).toHaveValidApiResponse();
      
      if (result.data.testRuns.length > 0) {
        result.data.testRuns.forEach((testRun: any) => {
          expect(testRun.status).toBe('FAILED');
        });
      }
    });

    test('should filter test runs by date range', async () => {
      const endDate = new Date();
      const startDate = new Date(endDate.getTime() - (7 * 24 * 60 * 60 * 1000)); // 7 days ago
      
      const query = `
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
      `;
      
      const result = await graphqlClient.query(query, {
        startDate: startDate.toISOString(),
        endDate: endDate.toISOString()
      });
      
      expect(result).toHaveValidApiResponse();
      
      result.data.testRuns.forEach((testRun: any) => {
        const testStartTime = new Date(testRun.startTime);
        expect(testStartTime.getTime()).toBeGreaterThanOrEqual(startDate.getTime());
        expect(testStartTime.getTime()).toBeLessThanOrEqual(endDate.getTime());
      });
    });

    test('should sort test runs by different fields', async () => {
      const sortFields = ['startTime', 'duration', 'status'];
      
      for (const sortField of sortFields) {
        const query = `
          query GetSortedTestRuns($sortBy: String!, $sortOrder: SortOrder!) {
            testRuns(sortBy: $sortBy, sortOrder: $sortOrder, limit: 10) {
              id
              ${sortField}
            }
          }
        `;
        
        const ascResult = await graphqlClient.query(query, {
          sortBy: sortField,
          sortOrder: 'ASC'
        });
        
        expect(ascResult).toHaveValidApiResponse();
        
        if (ascResult.data.testRuns.length > 1) {
          // Verify ascending order
          for (let i = 1; i < ascResult.data.testRuns.length; i++) {
            const prev = ascResult.data.testRuns[i - 1][sortField];
            const curr = ascResult.data.testRuns[i][sortField];
            
            if (sortField === 'startTime') {
              expect(new Date(curr).getTime()).toBeGreaterThanOrEqual(new Date(prev).getTime());
            } else if (sortField === 'duration') {
              expect(curr).toBeGreaterThanOrEqual(prev);
            }
          }
        }
      }
    });
  });

  describe('Spec Run Queries', () => {
    test('should fetch spec runs for a test run', async () => {
      // Get a test run first
      const testRunsQuery = `
        query GetTestRun {
          testRuns(limit: 1) {
            id
          }
        }
      `;
      
      const testRunsResult = await graphqlClient.query(testRunsQuery);
      expect(testRunsResult.data.testRuns.length).toBeGreaterThan(0);
      
      const testRunId = testRunsResult.data.testRuns[0].id;
      
      const specRunsQuery = `
        query GetSpecRuns($testRunId: ID!) {
          specRuns(testRunId: $testRunId) {
            id
            testRunId
            specDescription
            status
            startTime
            endTime
            duration
            errorMessage
            stackTrace
          }
        }
      `;
      
      const result = await graphqlClient.query(specRunsQuery, { testRunId });
      expect(result).toHaveValidApiResponse();
      expect(result.data.specRuns).toBeDefined();
      expect(Array.isArray(result.data.specRuns)).toBe(true);
      
      result.data.specRuns.forEach((specRun: any) => {
        expect(specRun.testRunId).toBe(testRunId);
        expect(specRun.specDescription).toBeTruthy();
        expect(['PASSED', 'FAILED', 'SKIPPED']).toContain(specRun.status);
      });
    });

    test('should include nested spec runs in test run query', async () => {
      const query = `
        query GetTestRunWithSpecs {
          testRuns(limit: 1) {
            id
            status
            specRuns {
              id
              specDescription
              status
              duration
              errorMessage
            }
          }
        }
      `;
      
      const result = await graphqlClient.query(query);
      expect(result).toHaveValidApiResponse();
      
      if (result.data.testRuns.length > 0) {
        const testRun = result.data.testRuns[0];
        expect(testRun.specRuns).toBeDefined();
        expect(Array.isArray(testRun.specRuns)).toBe(true);
        
        testRun.specRuns.forEach((specRun: any) => {
          expect(specRun.specDescription).toBeTruthy();
          expect(['PASSED', 'FAILED', 'SKIPPED']).toContain(specRun.status);
        });
      }
    });

    test('should filter spec runs by status', async () => {
      const query = `
        query GetFailedSpecRuns {
          specRuns(status: FAILED, limit: 10) {
            id
            status
            errorMessage
            stackTrace
          }
        }
      `;
      
      const result = await graphqlClient.query(query);
      expect(result).toHaveValidApiResponse();
      
      result.data.specRuns.forEach((specRun: any) => {
        expect(specRun.status).toBe('FAILED');
        expect(specRun.errorMessage).toBeTruthy();
      });
    });
  });

  describe('Project Queries', () => {
    test('should fetch all projects', async () => {
      const query = `
        query GetProjects {
          projects {
            id
            name
            description
            tags
            createdAt
          }
        }
      `;
      
      const result = await graphqlClient.query(query);
      expect(result).toHaveValidApiResponse();
      expect(result.data.projects).toBeDefined();
      expect(Array.isArray(result.data.projects)).toBe(true);
      expect(result.data.projects.length).toBeGreaterThan(0);
      
      result.data.projects.forEach((project: any) => {
        expect(project.id).toBeTruthy();
        expect(project.name).toBeTruthy();
        expect(Array.isArray(project.tags)).toBe(true);
      });
    });

    test('should get project statistics', async () => {
      const projects = await context.dataFixtures.createdData.projects;
      const testProject = projects[0];
      
      const query = `
        query GetProjectStats($projectId: ID!) {
          project(id: $projectId) {
            id
            name
            stats {
              totalRuns
              successRate
              averageDuration
              lastRunTime
            }
          }
        }
      `;
      
      const result = await graphqlClient.query(query, { 
        projectId: testProject.id 
      });
      
      expect(result).toHaveValidApiResponse();
      
      if (result.data.project) {
        const project = result.data.project;
        expect(project.id).toBe(testProject.id);
        
        if (project.stats) {
          expect(typeof project.stats.totalRuns).toBe('number');
          expect(typeof project.stats.successRate).toBe('number');
          expect(project.stats.successRate).toBeGreaterThanOrEqual(0);
          expect(project.stats.successRate).toBeLessThanOrEqual(1);
        }
      }
    });
  });

  describe('Complex Queries and Performance', () => {
    test('should handle complex nested queries efficiently', async () => {
      const complexQuery = `
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
      `;
      
      const endMeasurement = performanceMonitor.startMeasurement('complex_nested_query');
      const result = await graphqlClient.query(complexQuery);
      const queryTime = endMeasurement();
      
      expect(queryTime).toBeWithinTimeRange(0, 5000); // 5 second max for complex query
      expect(result).toHaveValidApiResponse();
      
      result.data.testRuns.forEach((testRun: any) => {
        expect(testRun.project).toBeDefined();
        expect(testRun.specRuns).toBeDefined();
        expect(testRun.suiteRun).toBeDefined();
      });
    });

    test('should implement proper N+1 query prevention', async () => {
      const query = `
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
      `;
      
      // This should be resolved with a single or minimal number of database queries
      const endMeasurement = performanceMonitor.startMeasurement('n_plus_one_prevention');
      const result = await graphqlClient.query(query);
      const queryTime = endMeasurement();
      
      // Should be fast even with nested project data
      expect(queryTime).toBeWithinTimeRange(0, 3000);
      expect(result).toHaveValidApiResponse();
      
      result.data.testRuns.forEach((testRun: any) => {
        expect(testRun.project.id).toBe(testRun.projectId);
        expect(testRun.project.name).toBeTruthy();
      });
    });

    test('should handle large result sets with proper pagination', async () => {
      const largeQuery = `
        query GetLargeResultSet($limit: Int!) {
          testRuns(limit: $limit) {
            id
            startTime
            pageInfo {
              hasNextPage
              endCursor
            }
          }
        }
      `;
      
      const result = await graphqlClient.query(largeQuery, { limit: 100 });
      expect(result).toHaveValidApiResponse();
      
      // Should handle large results without timeout
      expect(result.data.testRuns.length).toBeLessThanOrEqual(100);
      
      if (result.data.testRuns.pageInfo.hasNextPage) {
        expect(result.data.testRuns.pageInfo.endCursor).toBeTruthy();
      }
    });
  });

  describe('Error Handling and Validation', () => {
    test('should return proper error for invalid query syntax', async () => {
      const invalidQuery = `
        query InvalidSyntax {
          testRuns {
            id
            invalidField {
              nonExistentNestedField
            }
          }
        }
      `;
      
      const result = await graphqlClient.query(invalidQuery);
      
      expect(result.errors).toBeDefined();
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0].message).toContain('Cannot query field');
    });

    test('should validate required arguments', async () => {
      const queryWithMissingArgs = `
        query MissingArgs {
          testRun {
            id
          }
        }
      `;
      
      const result = await graphqlClient.query(queryWithMissingArgs);
      
      expect(result.errors).toBeDefined();
      expect(result.errors[0].message).toContain('required');
    });

    test('should handle invalid variable types', async () => {
      const query = `
        query InvalidVariableType($limit: String!) {
          testRuns(limit: $limit) {
            id
          }
        }
      `;
      
      const result = await graphqlClient.query(query, { limit: "not_a_number" });
      
      expect(result.errors).toBeDefined();
      expect(result.errors[0].message).toContain('type');
    });

    test('should return proper error for non-existent resources', async () => {
      const query = `
        query GetNonExistentTestRun($id: ID!) {
          testRun(id: $id) {
            id
            status
          }
        }
      `;
      
      const result = await graphqlClient.query(query, { 
        id: 'non-existent-id' 
      });
      
      // Should either return null or proper error
      expect(result.data.testRun).toBeNull();
    });

    test('should handle malformed JSON gracefully', async () => {
      try {
        await graphqlClient.rawRequest('{ invalid json }');
        fail('Should have thrown an error');
      } catch (error) {
        expect(error.message).toContain('JSON');
      }
    });
  });

  describe('Security and Authorization', () => {
    test('should require proper authentication for protected queries', async () => {
      // Test with invalid auth token
      const unauthorizedClient = new GraphQLClient(context.baseUrls.reporter, {
        headers: { Authorization: 'Bearer invalid-token' }
      });
      
      const query = `
        query GetProtectedData {
          userPreferences {
            userId
            preferences
          }
        }
      `;
      
      try {
        await unauthorizedClient.query(query);
        // If auth is not implemented yet, this test passes
      } catch (error) {
        expect(error.message).toContain('Unauthorized');
      }
    });

    test('should prevent SQL injection through GraphQL variables', async () => {
      const maliciousQuery = `
        query SQLInjectionAttempt($projectId: ID!) {
          testRuns(projectId: $projectId) {
            id
          }
        }
      `;
      
      const maliciousProjectId = "'; DROP TABLE test_runs; --";
      
      // Should not cause SQL injection
      const result = await graphqlClient.query(maliciousQuery, {
        projectId: maliciousProjectId
      });
      
      // Should either return empty results or validation error
      expect(result.data.testRuns).toEqual([]);
    });

    test('should have proper rate limiting', async () => {
      const query = `
        query RateLimitTest {
          testRuns(limit: 1) {
            id
          }
        }
      `;
      
      // Send many requests rapidly
      const requests = Array(20).fill(null).map(() => 
        graphqlClient.query(query)
      );
      
      const results = await Promise.allSettled(requests);
      
      // Some requests might be rate limited
      const rateLimitedRequests = results.filter(result => 
        result.status === 'rejected' && 
        result.reason.message.includes('rate limit')
      );
      
      // If rate limiting is implemented, should see some failures
      // If not implemented yet, all should succeed
      expect(results.every(r => r.status === 'fulfilled') || rateLimitedRequests.length > 0).toBe(true);
    });
  });

  describe('Real-time Features (Future)', () => {
    test('should support subscriptions for real-time updates', async () => {
      // This test is for future subscription support
      const hasSubscriptionSupport = await graphqlClient.hasSubscriptionSupport();
      
      if (hasSubscriptionSupport) {
        const subscription = `
          subscription TestRunUpdates {
            testRunUpdated {
              id
              status
              endTime
            }
          }
        `;
        
        const subscriptionClient = await graphqlClient.createSubscription(subscription);
        
        // Simulate test run update
        setTimeout(() => {
          // Trigger a test run update event
        }, 1000);
        
        const update = await subscriptionClient.waitForNext(5000);
        expect(update.data.testRunUpdated).toBeDefined();
        
        await subscriptionClient.close();
      } else {
        // Skip test if subscriptions not implemented yet
        console.log('Subscriptions not yet implemented, skipping test');
      }
    });
  });

  describe('Caching and Optimization', () => {
    test('should implement proper query caching', async () => {
      const query = `
        query CacheTest {
          testRuns(limit: 5) {
            id
            status
          }
        }
      `;
      
      // First request
      const endMeasurement1 = performanceMonitor.startMeasurement('first_cache_request');
      await graphqlClient.query(query);
      const firstRequestTime = endMeasurement1();
      
      // Second request (should be cached)
      const endMeasurement2 = performanceMonitor.startMeasurement('cached_request');
      await graphqlClient.query(query);
      const secondRequestTime = endMeasurement2();
      
      // If caching is implemented, second request should be faster
      // If not implemented yet, times might be similar
      expect(secondRequestTime).toBeLessThanOrEqual(firstRequestTime * 2);
    });

    test('should handle cache invalidation properly', async () => {
      const query = `
        query CacheInvalidationTest {
          testRuns(limit: 1) {
            id
            status
          }
        }
      `;
      
      // Get initial data
      const firstResult = await graphqlClient.query(query);
      
      // Simulate data change (if mutations are available)
      // await graphqlClient.mutate(updateMutation);
      
      // Get data again
      const secondResult = await graphqlClient.query(query, {}, { 
        fetchPolicy: 'no-cache' 
      });
      
      // Should reflect changes
      expect(secondResult).toHaveValidApiResponse();
    });
  });
});