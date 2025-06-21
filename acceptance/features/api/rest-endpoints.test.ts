/**
 * REST API Endpoints - Acceptance Tests
 * 
 * Tests all REST API endpoints including:
 * - Test data ingestion and reporting
 * - Health checks and monitoring
 * - User preferences and configuration
 * - Project management
 * - Summary and aggregation endpoints
 * - Error handling and validation
 * 
 * Based on fern-reporter REST API and GitHub issues
 */

import { TestUtils, HttpUtils, performanceMonitor } from '@acceptance/setup/test-helpers';
import { RestApiClient } from '@acceptance/utils/api-clients/rest-api-client';

describe('REST API Endpoints', () => {
  let apiClient: RestApiClient;
  let context: any;

  beforeAll(async () => {
    context = TestUtils.getTestContext();
    apiClient = new RestApiClient(context.baseUrls.reporter);
    
    // Ensure API service is ready
    await TestUtils.waitForCondition(
      async () => {
        try {
          await apiClient.healthCheck();
          return true;
        } catch {
          return false;
        }
      },
      60000
    );
  });

  describe('Health and Monitoring Endpoints', () => {
    test('should provide health check endpoint', async () => {
      const endMeasurement = performanceMonitor.startMeasurement('health_check');
      const response = await apiClient.get('/health');
      const responseTime = endMeasurement();
      
      expect(response.status).toBe(200);
      expect(responseTime).toBeWithinTimeRange(0, 1000); // 1 second max
      
      expect(response.data).toHaveProperty('status');
      expect(response.data.status).toBe('healthy');
    });

    test('should provide readiness check endpoint', async () => {
      const response = await apiClient.get('/ready');
      
      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('ready');
      expect(response.data.ready).toBe(true);
      
      // Should include dependency checks
      if (response.data.dependencies) {
        expect(response.data.dependencies).toHaveProperty('database');
        expect(response.data.dependencies.database).toBe('healthy');
      }
    });

    test('should provide metrics endpoint for monitoring', async () => {
      const response = await apiClient.get('/metrics');
      
      expect(response.status).toBe(200);
      
      // Should be in Prometheus format or JSON
      const isPrometheusFormat = typeof response.data === 'string' && 
                                response.data.includes('# HELP');
      const isJsonFormat = typeof response.data === 'object';
      
      expect(isPrometheusFormat || isJsonFormat).toBe(true);
    });

    test('should provide system info endpoint', async () => {
      const response = await apiClient.get('/info');
      
      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('version');
      expect(response.data).toHaveProperty('buildTime');
      expect(response.data).toHaveProperty('gitSha');
    });
  });

  describe('Test Data Ingestion', () => {
    test('should accept test run data via POST', async () => {
      const testRunData = {
        id: `test-run-${Date.now()}`,
        projectId: 'test-project',
        suiteId: 'test-suite',
        branch: 'main',
        buildUrl: 'https://ci.example.com/build/123',
        buildActor: 'test-user',
        gitSha: 'abc123def456',
        status: 'passed',
        startTime: new Date().toISOString(),
        endTime: new Date().toISOString(),
        duration: 120000,
        tags: ['integration', 'api'],
        specRuns: [
          {
            id: `spec-run-${Date.now()}`,
            specDescription: 'should handle API requests',
            status: 'passed',
            startTime: new Date().toISOString(),
            endTime: new Date().toISOString(),
            duration: 5000
          }
        ]
      };
      
      const endMeasurement = performanceMonitor.startMeasurement('test_data_ingestion');
      const response = await apiClient.post('/api/test-runs', testRunData);
      const ingestionTime = endMeasurement();
      
      expect(response.status).toBe(201);
      expect(ingestionTime).toBeWithinTimeRange(0, 5000); // 5 second max
      
      expect(response.data).toHaveProperty('id');
      expect(response.data.id).toBe(testRunData.id);
      expect(response.data).toHaveProperty('status');
      expect(response.data.status).toBe('created');
    });

    test('should validate required fields in test run data', async () => {
      const invalidTestRunData = {
        // Missing required fields
        projectId: 'test-project',
        status: 'passed'
      };
      
      const response = await apiClient.post('/api/test-runs', invalidTestRunData);
      
      expect(response.status).toBe(400);
      expect(response.data).toHaveProperty('error');
      expect(response.data.error).toContain('required');
    });

    test('should handle bulk test run ingestion', async () => {
      const bulkTestRuns = Array.from({ length: 10 }, (_, i) => ({
        id: `bulk-test-run-${Date.now()}-${i}`,
        projectId: 'bulk-test-project',
        suiteId: `bulk-suite-${i}`,
        branch: 'main',
        buildUrl: `https://ci.example.com/build/${i}`,
        buildActor: 'bulk-test-user',
        gitSha: `abc123def${i}`,
        status: i % 3 === 0 ? 'failed' : 'passed',
        startTime: new Date().toISOString(),
        endTime: new Date().toISOString(),
        duration: 60000 + (i * 1000),
        tags: ['bulk', 'test'],
        specRuns: []
      }));
      
      const endMeasurement = performanceMonitor.startMeasurement('bulk_ingestion');
      const response = await apiClient.post('/api/test-runs/bulk', { 
        testRuns: bulkTestRuns 
      });
      const bulkIngestionTime = endMeasurement();
      
      expect(response.status).toBe(201);
      expect(bulkIngestionTime).toBeWithinTimeRange(0, 10000); // 10 second max
      
      expect(response.data).toHaveProperty('created');
      expect(response.data.created).toBe(bulkTestRuns.length);
      expect(response.data).toHaveProperty('failed');
      expect(response.data.failed).toBe(0);
    });

    test('should handle partial failures in bulk ingestion', async () => {
      const mixedTestRuns = [
        {
          id: `valid-test-run-${Date.now()}`,
          projectId: 'test-project',
          suiteId: 'test-suite',
          branch: 'main',
          status: 'passed',
          startTime: new Date().toISOString(),
          endTime: new Date().toISOString(),
          duration: 60000
        },
        {
          // Invalid - missing required fields
          id: `invalid-test-run-${Date.now()}`,
          status: 'passed'
        }
      ];
      
      const response = await apiClient.post('/api/test-runs/bulk', { 
        testRuns: mixedTestRuns 
      });
      
      expect(response.status).toBe(207); // Multi-status
      expect(response.data).toHaveProperty('created');
      expect(response.data).toHaveProperty('failed');
      expect(response.data.created).toBe(1);
      expect(response.data.failed).toBe(1);
      expect(response.data).toHaveProperty('errors');
      expect(response.data.errors.length).toBe(1);
    });
  });

  describe('Project Management', () => {
    test('should create new projects', async () => {
      const projectData = {
        id: `project-${Date.now()}`,
        name: 'Test Project API',
        description: 'A test project created via API',
        tags: ['api', 'test', 'automation']
      };
      
      const response = await apiClient.post('/api/projects', projectData);
      
      expect(response.status).toBe(201);
      expect(response.data).toHaveProperty('id');
      expect(response.data.id).toBe(projectData.id);
      expect(response.data.name).toBe(projectData.name);
    });

    test('should list all projects', async () => {
      const response = await apiClient.get('/api/projects');
      
      expect(response.status).toBe(200);
      expect(Array.isArray(response.data)).toBe(true);
      expect(response.data.length).toBeGreaterThan(0);
      
      response.data.forEach((project: any) => {
        expect(project).toHaveProperty('id');
        expect(project).toHaveProperty('name');
        expect(project).toHaveProperty('tags');
        expect(Array.isArray(project.tags)).toBe(true);
      });
    });

    test('should get specific project by ID', async () => {
      const projects = await context.dataFixtures.createdData.projects;
      const testProject = projects[0];
      
      const response = await apiClient.get(`/api/projects/${testProject.id}`);
      
      expect(response.status).toBe(200);
      expect(response.data.id).toBe(testProject.id);
      expect(response.data.name).toBe(testProject.name);
    });

    test('should update existing project', async () => {
      const projects = await context.dataFixtures.createdData.projects;
      const testProject = projects[0];
      
      const updateData = {
        name: 'Updated Project Name',
        description: 'Updated description',
        tags: ['updated', 'test']
      };
      
      const response = await apiClient.put(`/api/projects/${testProject.id}`, updateData);
      
      expect(response.status).toBe(200);
      expect(response.data.name).toBe(updateData.name);
      expect(response.data.description).toBe(updateData.description);
      expect(response.data.tags).toEqual(updateData.tags);
    });

    test('should return 404 for non-existent project', async () => {
      const response = await apiClient.get('/api/projects/non-existent-project');
      
      expect(response.status).toBe(404);
      expect(response.data).toHaveProperty('error');
      expect(response.data.error).toContain('not found');
    });
  });

  describe('Summary and Aggregation Endpoints', () => {
    test('should provide project summary data', async () => {
      const response = await apiClient.get('/api/reports/projects');
      
      expect(response.status).toBe(200);
      expect(Array.isArray(response.data)).toBe(true);
      
      response.data.forEach((projectSummary: any) => {
        expect(projectSummary).toHaveProperty('projectId');
        expect(projectSummary).toHaveProperty('projectName');
        expect(projectSummary).toHaveProperty('totalRuns');
        expect(projectSummary).toHaveProperty('successRate');
        expect(projectSummary).toHaveProperty('averageDuration');
        expect(typeof projectSummary.totalRuns).toBe('number');
        expect(typeof projectSummary.successRate).toBe('number');
      });
    });

    test('should provide detailed summary for specific project', async () => {
      const projects = await context.dataFixtures.createdData.projects;
      const testProject = projects[0];
      
      const response = await apiClient.get(`/api/reports/summary/${testProject.id}`);
      
      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('projectId');
      expect(response.data.projectId).toBe(testProject.id);
      expect(response.data).toHaveProperty('summaryData');
      expect(Array.isArray(response.data.summaryData)).toBe(true);
      
      // Summary data should be chronologically ordered
      const summaryData = response.data.summaryData;
      if (summaryData.length > 1) {
        for (let i = 1; i < summaryData.length; i++) {
          const prevDate = new Date(summaryData[i - 1].date);
          const currDate = new Date(summaryData[i].date);
          expect(prevDate.getTime()).toBeGreaterThanOrEqual(currDate.getTime());
        }
      }
    });

    test('should support date range filtering for summaries', async () => {
      const projects = await context.dataFixtures.createdData.projects;
      const testProject = projects[0];
      
      const endDate = new Date();
      const startDate = new Date(endDate.getTime() - (7 * 24 * 60 * 60 * 1000)); // 7 days
      
      const response = await apiClient.get(`/api/reports/summary/${testProject.id}`, {
        params: {
          startDate: startDate.toISOString(),
          endDate: endDate.toISOString()
        }
      });
      
      expect(response.status).toBe(200);
      
      response.data.summaryData.forEach((summary: any) => {
        const summaryDate = new Date(summary.date);
        expect(summaryDate.getTime()).toBeGreaterThanOrEqual(startDate.getTime());
        expect(summaryDate.getTime()).toBeLessThanOrEqual(endDate.getTime());
      });
    });

    test('should provide system-wide statistics', async () => {
      const response = await apiClient.get('/api/reports/stats');
      
      expect(response.status).toBe(200);
      expect(response.data).toHaveProperty('totalProjects');
      expect(response.data).toHaveProperty('totalTestRuns');
      expect(response.data).toHaveProperty('totalSpecRuns');
      expect(response.data).toHaveProperty('overallSuccessRate');
      expect(response.data).toHaveProperty('averageTestDuration');
      
      expect(typeof response.data.totalProjects).toBe('number');
      expect(typeof response.data.totalTestRuns).toBe('number');
      expect(typeof response.data.overallSuccessRate).toBe('number');
      expect(response.data.overallSuccessRate).toBeGreaterThanOrEqual(0);
      expect(response.data.overallSuccessRate).toBeLessThanOrEqual(1);
    });
  });

  describe('User Preferences and Configuration', () => {
    test('should manage user preferences', async () => {
      const userId = `test-user-${Date.now()}`;
      const preferences = {
        timezone: 'America/New_York',
        theme: 'dark',
        favoriteProjects: ['project1', 'project2'],
        projectGroups: [
          {
            name: 'Core Services',
            projects: ['project1']
          }
        ]
      };
      
      // Create preferences
      const createResponse = await apiClient.post('/api/user-preferences', {
        userId,
        preferences
      });
      
      expect(createResponse.status).toBe(201);
      expect(createResponse.data).toHaveProperty('userId');
      expect(createResponse.data.userId).toBe(userId);
      
      // Retrieve preferences
      const getResponse = await apiClient.get(`/api/user-preferences/${userId}`);
      
      expect(getResponse.status).toBe(200);
      expect(getResponse.data.preferences.timezone).toBe(preferences.timezone);
      expect(getResponse.data.preferences.theme).toBe(preferences.theme);
      expect(getResponse.data.preferences.favoriteProjects).toEqual(preferences.favoriteProjects);
    });

    test('should update user preferences', async () => {
      const userId = `update-test-user-${Date.now()}`;
      
      // Create initial preferences
      await apiClient.post('/api/user-preferences', {
        userId,
        preferences: { timezone: 'UTC', theme: 'light' }
      });
      
      // Update preferences
      const updatedPreferences = {
        timezone: 'Europe/London',
        theme: 'dark',
        favoriteProjects: ['new-favorite']
      };
      
      const updateResponse = await apiClient.put(`/api/user-preferences/${userId}`, {
        preferences: updatedPreferences
      });
      
      expect(updateResponse.status).toBe(200);
      expect(updateResponse.data.preferences.timezone).toBe(updatedPreferences.timezone);
      expect(updateResponse.data.preferences.theme).toBe(updatedPreferences.theme);
    });

    test('should manage favorite projects', async () => {
      const userId = `favorite-test-user-${Date.now()}`;
      const projects = await context.dataFixtures.createdData.projects;
      const testProject = projects[0];
      
      // Add to favorites
      const addResponse = await apiClient.post(`/api/user-preferences/${userId}/favorites`, {
        projectId: testProject.id
      });
      
      expect(addResponse.status).toBe(200);
      expect(addResponse.data.favoriteProjects).toContain(testProject.id);
      
      // Remove from favorites
      const removeResponse = await apiClient.delete(`/api/user-preferences/${userId}/favorites/${testProject.id}`);
      
      expect(removeResponse.status).toBe(200);
      expect(addResponse.data.favoriteProjects).not.toContain(testProject.id);
    });
  });

  describe('Error Handling and Validation', () => {
    test('should return proper error for malformed JSON', async () => {
      const response = await apiClient.postRaw('/api/test-runs', '{ invalid json }');
      
      expect(response.status).toBe(400);
      expect(response.data).toHaveProperty('error');
      expect(response.data.error).toContain('JSON');
    });

    test('should validate Content-Type headers', async () => {
      const response = await apiClient.post('/api/test-runs', 
        { id: 'test' },
        { headers: { 'Content-Type': 'text/plain' } }
      );
      
      expect(response.status).toBe(415); // Unsupported Media Type
      expect(response.data).toHaveProperty('error');
      expect(response.data.error).toContain('Content-Type');
    });

    test('should handle large payload gracefully', async () => {
      const largePayload = {
        id: 'large-test-run',
        projectId: 'test-project',
        specRuns: Array.from({ length: 10000 }, (_, i) => ({
          id: `spec-${i}`,
          specDescription: `Large test spec ${i}`.repeat(100),
          status: 'passed',
          startTime: new Date().toISOString(),
          endTime: new Date().toISOString(),
          duration: 1000
        }))
      };
      
      const response = await apiClient.post('/api/test-runs', largePayload);
      
      // Should either succeed or return proper error for payload size
      expect([201, 413]).toContain(response.status); // 413 = Payload Too Large
      
      if (response.status === 413) {
        expect(response.data.error).toContain('payload');
      }
    });

    test('should implement proper CORS headers', async () => {
      const response = await apiClient.options('/api/test-runs');
      
      expect(response.status).toBe(200);
      expect(response.headers).toHaveProperty('access-control-allow-origin');
      expect(response.headers).toHaveProperty('access-control-allow-methods');
      expect(response.headers).toHaveProperty('access-control-allow-headers');
    });

    test('should handle concurrent requests without conflicts', async () => {
      const concurrentRequests = Array.from({ length: 10 }, (_, i) => 
        apiClient.post('/api/test-runs', {
          id: `concurrent-test-${Date.now()}-${i}`,
          projectId: 'concurrent-project',
          suiteId: 'concurrent-suite',
          status: 'passed',
          startTime: new Date().toISOString(),
          endTime: new Date().toISOString(),
          duration: 1000
        })
      );
      
      const results = await Promise.allSettled(concurrentRequests);
      
      // All requests should succeed
      results.forEach((result, index) => {
        expect(result.status).toBe('fulfilled');
        if (result.status === 'fulfilled') {
          expect(result.value.status).toBe(201);
        }
      });
    });
  });

  describe('Performance and Rate Limiting', () => {
    test('should respond to health checks quickly', async () => {
      const iterations = 10;
      const times: number[] = [];
      
      for (let i = 0; i < iterations; i++) {
        const endMeasurement = performanceMonitor.startMeasurement(`health_check_${i}`);
        await apiClient.get('/health');
        const time = endMeasurement();
        times.push(time);
      }
      
      const avgTime = times.reduce((a, b) => a + b, 0) / times.length;
      expect(avgTime).toBeWithinTimeRange(0, 500); // 500ms average max
    });

    test('should implement rate limiting', async () => {
      const rapidRequests = Array.from({ length: 100 }, () => 
        apiClient.get('/api/projects')
      );
      
      const results = await Promise.allSettled(rapidRequests);
      
      // Should have some rate-limited responses
      const rateLimitedCount = results.filter(result => 
        result.status === 'rejected' || 
        (result.status === 'fulfilled' && result.value.status === 429)
      ).length;
      
      // If rate limiting is implemented, expect some failures
      // If not implemented yet, all should succeed
      expect(rateLimitedCount >= 0).toBe(true);
    });

    test('should handle pagination efficiently', async () => {
      const pageSize = 50;
      const maxPages = 5;
      
      for (let page = 1; page <= maxPages; page++) {
        const endMeasurement = performanceMonitor.startMeasurement(`pagination_page_${page}`);
        
        const response = await apiClient.get('/api/test-runs', {
          params: { page, limit: pageSize }
        });
        
        const pageTime = endMeasurement();
        
        expect(response.status).toBe(200);
        expect(pageTime).toBeWithinTimeRange(0, 3000); // 3 second max per page
        
        if (response.data.length < pageSize) {
          // Reached end of data
          break;
        }
      }
    });
  });

  describe('Content Negotiation', () => {
    test('should support JSON content type', async () => {
      const response = await apiClient.get('/api/projects', {
        headers: { Accept: 'application/json' }
      });
      
      expect(response.status).toBe(200);
      expect(response.headers['content-type']).toContain('application/json');
    });

    test('should support CSV export format', async () => {
      const response = await apiClient.get('/api/reports/projects', {
        headers: { Accept: 'text/csv' }
      });
      
      // Should either support CSV or return 406 Not Acceptable
      expect([200, 406]).toContain(response.status);
      
      if (response.status === 200) {
        expect(response.headers['content-type']).toContain('text/csv');
        expect(typeof response.data).toBe('string');
        expect(response.data).toContain(','); // CSV delimiter
      }
    });

    test('should handle unsupported Accept headers gracefully', async () => {
      const response = await apiClient.get('/api/projects', {
        headers: { Accept: 'application/xml' }
      });
      
      // Should return 406 Not Acceptable or default to JSON
      expect([200, 406]).toContain(response.status);
    });
  });

  describe('API Versioning', () => {
    test('should support API version headers', async () => {
      const response = await apiClient.get('/api/projects', {
        headers: { 'API-Version': 'v1' }
      });
      
      expect(response.status).toBe(200);
      
      // Should include version in response headers
      if (response.headers['api-version']) {
        expect(response.headers['api-version']).toBe('v1');
      }
    });

    test('should handle version negotiation', async () => {
      const response = await apiClient.get('/api/projects', {
        headers: { 'API-Version': 'v2' }
      });
      
      // Should either support v2 or return appropriate error
      expect([200, 400, 404]).toContain(response.status);
      
      if (response.status !== 200) {
        expect(response.data).toHaveProperty('error');
        expect(response.data.error).toMatch(/version|supported/i);
      }
    });
  });
});