/**
 * End-to-End Workflows - Integration Tests
 * 
 * Tests complete user workflows across the entire Fern platform:
 * - Test data ingestion → processing → visualization
 * - AI analysis → insights → recommendations
 * - User interactions → data persistence → retrieval
 * - Error scenarios → recovery → data consistency
 * 
 * These tests validate the integration between all services
 */

import { TestUtils, HttpUtils, performanceMonitor } from '@acceptance/setup/test-helpers';
import { ApiClient } from '@acceptance/utils/api-clients/fern-api-client';
import { TestRunsPage } from '@acceptance/utils/page-objects/test-runs-page';
import { TestSummariesPage } from '@acceptance/utils/page-objects/test-summaries-page';
import { ChatbotPage } from '@acceptance/utils/page-objects/chatbot-page';

describe('End-to-End Workflows', () => {
  let apiClient: ApiClient;
  let testRunsPage: TestRunsPage;
  let testSummariesPage: TestSummariesPage;
  let chatbotPage: ChatbotPage;
  let context: any;

  beforeAll(async () => {
    context = TestUtils.getTestContext();
    apiClient = new ApiClient(context.baseUrls);
    testRunsPage = new TestRunsPage(context.baseUrls.ui);
    testSummariesPage = new TestSummariesPage(context.baseUrls.ui);
    chatbotPage = new ChatbotPage(context.baseUrls.ui);
    
    // Ensure all services are ready
    await TestUtils.waitForCondition(
      async () => {
        try {
          return await apiClient.healthCheck();
        } catch {
          return false;
        }
      },
      120000 // 2 minutes for full stack readiness
    );
  });

  describe('Test Data Lifecycle', () => {
    test('should handle complete test data ingestion to visualization workflow', async () => {
      const workflowId = `e2e-workflow-${Date.now()}`;
      
      // Step 1: Create a new project
      const projectData = {
        name: `E2E Test Project ${workflowId}`,
        description: 'End-to-end workflow test project',
        tags: ['e2e', 'workflow', 'integration']
      };
      
      const projectResponse = await apiClient.createProject(projectData);
      expect(projectResponse.status).toBe(201);
      const projectId = projectResponse.data.id;
      
      // Step 2: Ingest test data
      const testRunData = {
        projectId,
        suiteId: `e2e-suite-${workflowId}`,
        branch: 'main',
        buildUrl: `https://ci.example.com/build/${workflowId}`,
        buildActor: 'e2e-test-user',
        gitSha: `abc${workflowId}def`,
        status: 'passed' as const,
        startTime: new Date().toISOString(),
        endTime: new Date(Date.now() + 120000).toISOString(),
        duration: 120000,
        tags: ['e2e', 'integration'],
        specRuns: [
          {
            specDescription: 'should complete user registration',
            status: 'passed' as const,
            startTime: new Date().toISOString(),
            endTime: new Date(Date.now() + 30000).toISOString(),
            duration: 30000
          },
          {
            specDescription: 'should handle payment processing',
            status: 'passed' as const,
            startTime: new Date(Date.now() + 30000).toISOString(),
            endTime: new Date(Date.now() + 60000).toISOString(),
            duration: 30000
          }
        ]
      };
      
      const testRunResponse = await apiClient.createTestRun(testRunData);
      expect(testRunResponse.status).toBe(201);
      const testRunId = testRunResponse.data.id;
      
      // Step 3: Wait for data processing
      await TestUtils.waitForCondition(
        async () => {
          const retrievedTestRun = await apiClient.getTestRunById(testRunId);
          return retrievedTestRun.data?.testRun?.id === testRunId;
        },
        30000
      );
      
      // Step 4: Verify data appears in UI
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      // Filter to our project
      const projects = await testRunsPage.getAvailableProjects();
      expect(projects).toContain(projectData.name);
      
      await testRunsPage.filterByProject(projectData.name);
      await testRunsPage.waitForFilterResults();
      
      const visibleRuns = await testRunsPage.getVisibleTestRuns();
      const ourTestRun = visibleRuns.find(run => run.id === testRunId);
      expect(ourTestRun).toBeDefined();
      expect(ourTestRun?.status).toBe('passed');
      
      // Step 5: Verify data in summaries
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const projectRows = await testSummariesPage.getProjectRows();
      const ourProjectRow = projectRows.find(row => row.projectName === projectData.name);
      expect(ourProjectRow).toBeDefined();
      expect(ourProjectRow?.summaryBoxes.length).toBeGreaterThan(0);
      
      // Step 6: Test AI analysis
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      await chatbotPage.sendMessage(`Analyze the test results for ${projectData.name}`);
      await chatbotPage.waitForResponse();
      
      const aiResponse = await chatbotPage.getLastBotMessage();
      expect(aiResponse.text).toContainTestInsight();
      expect(aiResponse.text).toContain(projectData.name);
    });

    test('should handle failed test workflow with error details', async () => {
      const workflowId = `failed-workflow-${Date.now()}`;
      
      // Create project for failed tests
      const projectResponse = await apiClient.createProject({
        name: `Failed Test Project ${workflowId}`,
        description: 'Failed test workflow',
        tags: ['failed', 'errors']
      });
      
      const projectId = projectResponse.data.id;
      
      // Ingest failed test data
      const failedTestRun = {
        projectId,
        suiteId: `failed-suite-${workflowId}`,
        branch: 'feature/buggy-code',
        status: 'failed' as const,
        startTime: new Date().toISOString(),
        endTime: new Date(Date.now() + 90000).toISOString(),
        duration: 90000,
        tags: ['failed'],
        specRuns: [
          {
            specDescription: 'should validate user input',
            status: 'failed' as const,
            startTime: new Date().toISOString(),
            endTime: new Date(Date.now() + 45000).toISOString(),
            duration: 45000,
            errorMessage: 'ValidationError: Email format is invalid',
            stackTrace: `at validateEmail (/app/validators.js:42:13)
                        at processUser (/app/user-service.js:156:8)
                        at /app/test/user.test.js:89:12`
          },
          {
            specDescription: 'should handle database connection',
            status: 'failed' as const,
            startTime: new Date(Date.now() + 45000).toISOString(),
            endTime: new Date(Date.now() + 90000).toISOString(),
            duration: 45000,
            errorMessage: 'ConnectionTimeoutError: Database connection timed out',
            stackTrace: `at Connection.timeout (/app/db/connection.js:78:15)
                        at Database.connect (/app/db/database.js:234:21)`
          }
        ]
      };
      
      const testRunResponse = await apiClient.createTestRun(failedTestRun);
      const testRunId = testRunResponse.data.id;
      
      // Wait for processing
      await TestUtils.waitForCondition(
        async () => {
          const retrievedTestRun = await apiClient.getTestRunById(testRunId);
          return retrievedTestRun.data?.testRun?.id === testRunId;
        },
        30000
      );
      
      // Verify failed test appears in UI with error details
      await testRunsPage.navigate();
      await testRunsPage.filterByStatus('failed');
      await testRunsPage.waitForFilterResults();
      
      const failedRuns = await testRunsPage.getVisibleTestRuns();
      const ourFailedRun = failedRuns.find(run => run.id === testRunId);
      expect(ourFailedRun).toBeDefined();
      expect(ourFailedRun?.status).toBe('failed');
      
      // Expand to see spec details
      await testRunsPage.expandTestRun(testRunId);
      await testRunsPage.waitForSpecDetailsToLoad(testRunId);
      
      const failedSpecs = await testRunsPage.getFailedSpecs(testRunId);
      expect(failedSpecs.length).toBe(2);
      
      failedSpecs.forEach(spec => {
        expect(spec.errorMessage).toBeTruthy();
        expect(spec.stackTrace).toBeTruthy();
      });
      
      // Test AI analysis of failures
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      await chatbotPage.sendMessage('What are the main causes of my recent test failures?');
      await chatbotPage.waitForResponse();
      
      const failureAnalysis = await chatbotPage.getLastBotMessage();
      expect(failureAnalysis.text).toContainTestInsight();
      expect(failureAnalysis.text.toLowerCase()).toMatch(/validation|database|connection|timeout/);
    });

    test('should handle flaky test pattern detection workflow', async () => {
      const workflowId = `flaky-workflow-${Date.now()}`;
      
      // Create project for flaky tests
      const projectResponse = await apiClient.createProject({
        name: `Flaky Test Project ${workflowId}`,
        description: 'Flaky test detection workflow',
        tags: ['flaky', 'unreliable']
      });
      
      const projectId = projectResponse.data.id;
      
      // Simulate flaky test pattern - multiple runs with inconsistent results
      const flakyTestRuns = [];
      
      for (let i = 0; i < 10; i++) {
        const isPass = Math.random() > 0.3; // 70% pass rate (flaky)
        
        const testRun = {
          projectId,
          suiteId: `flaky-suite-${workflowId}`,
          branch: 'main',
          status: isPass ? 'passed' : 'failed' as const,
          startTime: new Date(Date.now() - (i * 3600000)).toISOString(), // Hourly runs
          endTime: new Date(Date.now() - (i * 3600000) + 60000).toISOString(),
          duration: 60000,
          tags: ['flaky'],
          specRuns: [
            {
              specDescription: 'should handle concurrent user sessions',
              status: isPass ? 'passed' : 'failed' as const,
              startTime: new Date(Date.now() - (i * 3600000)).toISOString(),
              endTime: new Date(Date.now() - (i * 3600000) + 60000).toISOString(),
              duration: 60000,
              errorMessage: isPass ? undefined : 'Race condition detected in session manager'
            }
          ]
        };
        
        const response = await apiClient.createTestRun(testRun);
        flakyTestRuns.push(response.data.id);
      }
      
      // Wait for all data to be processed
      await TestUtils.waitForCondition(
        async () => {
          const testRuns = await apiClient.getTestRuns({ projectId });
          return testRuns.data?.testRuns?.length >= 10;
        },
        60000
      );
      
      // Test flaky detection via AI
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      await chatbotPage.sendMessage(`Which tests are flaky in ${projectResponse.data.name}?`);
      await chatbotPage.waitForResponse();
      
      const flakyAnalysis = await chatbotPage.getLastBotMessage();
      expect(flakyAnalysis.text).toContainTestInsight();
      expect(flakyAnalysis.text.toLowerCase()).toMatch(/flaky|intermittent|unstable|race condition/);
      
      // Test flaky detection via API
      const flakyTests = await apiClient.getFlakyTests(projectId);
      expect(flakyTests.data).toBeDefined();
      
      if (flakyTests.data.data?.flakyTests?.length > 0) {
        const flakyTest = flakyTests.data.data.flakyTests[0];
        expect(flakyTest.passRate).toBeLessThan(0.9); // Less than 90% pass rate
        expect(flakyTest.failureRate).toBeGreaterThan(0.1); // More than 10% failure rate
      }
    });
  });

  describe('User Preference and Personalization Workflow', () => {
    test('should persist user preferences across sessions', async () => {
      const userId = `e2e-user-${Date.now()}`;
      const projects = await context.dataFixtures.createdData.projects;
      
      // Step 1: Set user preferences via API
      const preferences = {
        timezone: 'America/Los_Angeles',
        theme: 'dark',
        favoriteProjects: [projects[0].id, projects[1].id],
        projectGroups: [
          {
            name: 'Core Services',
            projects: [projects[0].id]
          },
          {
            name: 'Frontend',
            projects: [projects[1].id]
          }
        ]
      };
      
      await apiClient.setUserPreferences(userId, preferences);
      
      // Step 2: Navigate to preferences page and verify settings
      await testSummariesPage.navigate();
      await testSummariesPage.setCurrentUser(userId);
      await testSummariesPage.waitForSummariesLoad();
      
      // Should show only favorite projects when filtered
      await testSummariesPage.filterByFavorites(true);
      await testSummariesPage.waitForFilterUpdate();
      
      const visibleProjects = await testSummariesPage.getVisibleProjects();
      expect(visibleProjects.length).toBe(2);
      
      visibleProjects.forEach(project => {
        expect(preferences.favoriteProjects).toContain(project.projectId);
      });
      
      // Step 3: Modify preferences in UI
      const newFavoriteProject = projects[2];
      await testSummariesPage.toggleFavorite(newFavoriteProject.id);
      await testSummariesPage.waitForFavoriteUpdate();
      
      // Step 4: Verify changes persist via API
      const updatedPreferences = await apiClient.getUserPreferences(userId);
      expect(updatedPreferences.preferences.favoriteProjects).toContain(newFavoriteProject.id);
      
      // Step 5: Verify preferences work across different pages
      await testRunsPage.navigate();
      await testRunsPage.setCurrentUser(userId);
      await testRunsPage.waitForTestRunsToLoad();
      
      // Theme should be applied
      const currentTheme = await testRunsPage.getCurrentTheme();
      expect(currentTheme).toBe('dark');
    });

    test('should handle favorite project management workflow', async () => {
      const userId = `favorite-user-${Date.now()}`;
      const projects = await context.dataFixtures.createdData.projects;
      
      // Start with no favorites
      await apiClient.setUserPreferences(userId, { favoriteProjects: [] });
      
      await testSummariesPage.navigate();
      await testSummariesPage.setCurrentUser(userId);
      await testSummariesPage.waitForSummariesLoad();
      
      // Initially show all projects
      const allProjects = await testSummariesPage.getVisibleProjects();
      expect(allProjects.length).toBeGreaterThan(2);
      
      // Add projects to favorites one by one
      const projectsToFavorite = projects.slice(0, 3);
      
      for (const project of projectsToFavorite) {
        await testSummariesPage.setFavorite(project.id, true);
        await testSummariesPage.waitForFavoriteUpdate();
        
        // Verify favorite icon changed
        const isFavorite = await testSummariesPage.isFavorite(project.id);
        expect(isFavorite).toBe(true);
      }
      
      // Filter to favorites only
      await testSummariesPage.filterByFavorites(true);
      await testSummariesPage.waitForFilterUpdate();
      
      const favoriteProjects = await testSummariesPage.getVisibleProjects();
      expect(favoriteProjects.length).toBe(3);
      
      // Remove one favorite
      await testSummariesPage.setFavorite(projectsToFavorite[0].id, false);
      await testSummariesPage.waitForFavoriteUpdate();
      
      const updatedFavorites = await testSummariesPage.getVisibleProjects();
      expect(updatedFavorites.length).toBe(2);
      
      // Verify persistence across page refresh
      await testSummariesPage.refresh();
      await testSummariesPage.waitForSummariesLoad();
      
      const persistedFavorites = await testSummariesPage.getVisibleProjects();
      expect(persistedFavorites.length).toBe(2);
    });
  });

  describe('AI-Powered Analysis Workflow', () => {
    test('should provide comprehensive test insights through conversation', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Start with general question
      await chatbotPage.sendMessage('How are my tests performing overall?');
      await chatbotPage.waitForResponse();
      
      const overallAnalysis = await chatbotPage.getLastBotMessage();
      expect(overallAnalysis.text).toContainTestInsight();
      
      // Ask for specific project analysis
      const projects = await context.dataFixtures.createdData.projects;
      const authProject = projects.find(p => p.name.includes('Authentication'));
      
      if (authProject) {
        await chatbotPage.sendMessage(`Tell me about the ${authProject.name} project`);
        await chatbotPage.waitForResponse();
        
        const projectAnalysis = await chatbotPage.getLastBotMessage();
        expect(projectAnalysis.text).toContainTestInsight();
        expect(projectAnalysis.text).toContain(authProject.name);
      }
      
      // Ask for actionable recommendations
      await chatbotPage.sendMessage('What should I prioritize fixing first?');
      await chatbotPage.waitForResponse();
      
      const recommendations = await chatbotPage.getLastBotMessage();
      expect(recommendations.text).toContainTestInsight();
      expect(recommendations.text.toLowerCase()).toMatch(/recommend|suggest|priorit|fix|improv/);
      
      // Ask about trends
      await chatbotPage.sendMessage('Are my test results getting better or worse over time?');
      await chatbotPage.waitForResponse();
      
      const trendAnalysis = await chatbotPage.getLastBotMessage();
      expect(trendAnalysis.text).toContainTestInsight();
      expect(trendAnalysis.text.toLowerCase()).toMatch(/trend|improv|wors|time|week|day/);
      
      // Verify conversation context is maintained
      await chatbotPage.sendMessage('Can you give me more details about that?');
      await chatbotPage.waitForResponse();
      
      const contextualResponse = await chatbotPage.getLastBotMessage();
      expect(contextualResponse.text).toBeTruthy();
      expect(contextualResponse.text.length).toBeGreaterThan(20);
    });

    test('should handle complex multi-step analysis workflow', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Complex analysis request
      const complexQuery = `I need help with a comprehensive analysis:
      1. Identify my most problematic tests
      2. Find patterns in the failures
      3. Estimate the impact on development velocity
      4. Provide a prioritized action plan`;
      
      await chatbotPage.sendMessage(complexQuery);
      await chatbotPage.waitForResponse(30000); // Allow more time for complex analysis
      
      const comprehensiveAnalysis = await chatbotPage.getLastBotMessage();
      expect(comprehensiveAnalysis.text).toContainTestInsight();
      expect(comprehensiveAnalysis.text.length).toBeGreaterThan(200); // Substantial response
      
      // Should address multiple aspects
      const analysisText = comprehensiveAnalysis.text.toLowerCase();
      const addressesProblems = analysisText.includes('problem') || analysisText.includes('issue');
      const addressesPatterns = analysisText.includes('pattern') || analysisText.includes('trend');
      const addressesImpact = analysisText.includes('impact') || analysisText.includes('velocity');
      const addressesActions = analysisText.includes('action') || analysisText.includes('recommend');
      
      expect(addressesProblems || addressesPatterns || addressesImpact || addressesActions).toBe(true);
      
      // Follow up with specific questions
      await chatbotPage.sendMessage('Which specific test should I fix first?');
      await chatbotPage.waitForResponse();
      
      const specificRecommendation = await chatbotPage.getLastBotMessage();
      expect(specificRecommendation.text).toContainTestInsight();
      
      // Ask for implementation guidance
      await chatbotPage.sendMessage('How should I fix this test?');
      await chatbotPage.waitForResponse();
      
      const implementationGuidance = await chatbotPage.getLastBotMessage();
      expect(implementationGuidance.text).toBeTruthy();
      expect(implementationGuidance.text.toLowerCase()).toMatch(/fix|improv|chang|updat|modif/);
    });
  });

  describe('Error Handling and Recovery Workflows', () => {
    test('should handle service failures gracefully across the platform', async () => {
      // Start with normal operation
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      const initialRunCount = (await testRunsPage.getVisibleTestRuns()).length;
      expect(initialRunCount).toBeGreaterThan(0);
      
      // Simulate API failure
      await apiClient.simulateFailure();
      
      // UI should show error state
      await testRunsPage.refresh();
      
      const errorMessage = await testRunsPage.getErrorMessage();
      expect(errorMessage).toBeTruthy();
      expect(errorMessage.toLowerCase()).toMatch(/error|unable|failed/);
      
      // Should provide retry mechanism
      const retryButton = await testRunsPage.getRetryButton();
      expect(retryButton).toBeTruthy();
      
      // Restore service
      await apiClient.restoreConnection();
      
      // Retry should work
      await testRunsPage.clickRetryButton();
      await testRunsPage.waitForTestRunsToLoad();
      
      const recoveredRunCount = (await testRunsPage.getVisibleTestRuns()).length;
      expect(recoveredRunCount).toBe(initialRunCount);
    });

    test('should handle network interruptions during chat sessions', async () => {
      await chatbotPage.navigate();
      await chatbotPage.openChatbot();
      
      // Start normal conversation
      await chatbotPage.sendMessage('Hello, can you help me?');
      await chatbotPage.waitForResponse();
      
      const initialResponse = await chatbotPage.getLastBotMessage();
      expect(initialResponse.text).toBeTruthy();
      
      // Simulate network disconnection
      await apiClient.simulateNetworkDisconnection();
      
      // Try to send message during disconnection
      await chatbotPage.sendMessage('This message should be queued');
      
      // Should show connection status
      const connectionStatus = await chatbotPage.getConnectionStatus();
      expect(['disconnected', 'reconnecting']).toContain(connectionStatus);
      
      // Should queue the message
      const queuedCount = await chatbotPage.getQueuedMessageCount();
      expect(queuedCount).toBeGreaterThan(0);
      
      // Restore connection
      await apiClient.restoreNetworkConnection();
      
      // Should reconnect and send queued messages
      await TestUtils.waitForCondition(
        async () => (await chatbotPage.getConnectionStatus()) === 'connected',
        15000
      );
      
      await TestUtils.waitForCondition(
        async () => (await chatbotPage.getQueuedMessageCount()) === 0,
        10000
      );
      
      // Should receive response to queued message
      const queuedResponse = await chatbotPage.getLastBotMessage();
      expect(queuedResponse.text).toBeTruthy();
    });

    test('should maintain data consistency during partial failures', async () => {
      const workflowId = `consistency-${Date.now()}`;
      
      // Create test data
      const projectResponse = await apiClient.createProject({
        name: `Consistency Test ${workflowId}`,
        description: 'Data consistency test',
        tags: ['consistency']
      });
      
      const projectId = projectResponse.data.id;
      
      // Simulate partial failure during bulk data ingestion
      const bulkTestRuns = Array.from({ length: 5 }, (_, i) => ({
        projectId,
        suiteId: `consistency-suite-${i}`,
        branch: 'main',
        status: 'passed' as const,
        startTime: new Date().toISOString(),
        endTime: new Date(Date.now() + 60000).toISOString(),
        duration: 60000,
        tags: ['consistency']
      }));
      
      // Inject one invalid run to test partial failure handling
      bulkTestRuns[2] = {
        ...bulkTestRuns[2],
        // @ts-ignore - intentionally invalid for testing
        status: 'invalid-status'
      };
      
      const bulkResponse = await apiClient.post('/api/test-runs/bulk', {
        testRuns: bulkTestRuns
      });
      
      // Should handle partial success
      expect(bulkResponse.status).toBe(207); // Multi-status
      expect(bulkResponse.data.created).toBe(4); // 4 valid runs
      expect(bulkResponse.data.failed).toBe(1); // 1 invalid run
      
      // Verify only valid data is stored
      const retrievedRuns = await apiClient.getTestRuns({ projectId });
      expect(retrievedRuns.data?.testRuns?.length).toBe(4);
      
      // All retrieved runs should be valid
      retrievedRuns.data?.testRuns?.forEach(run => {
        expect(['passed', 'failed', 'skipped']).toContain(run.status);
      });
      
      // UI should show only valid data
      await testRunsPage.navigate();
      await testRunsPage.filterByProject(`Consistency Test ${workflowId}`);
      await testRunsPage.waitForFilterResults();
      
      const visibleRuns = await testRunsPage.getVisibleTestRuns();
      expect(visibleRuns.length).toBe(4);
    });
  });

  describe('Performance and Scalability Workflows', () => {
    test('should handle large dataset visualization efficiently', async () => {
      const workflowId = `perf-${Date.now()}`;
      
      // Create project with substantial test history
      const projectResponse = await apiClient.createProject({
        name: `Performance Test ${workflowId}`,
        description: 'Performance testing project',
        tags: ['performance', 'large-dataset']
      });
      
      const projectId = projectResponse.data.id;
      
      // Create large amount of test data (in smaller batches)
      const batchSize = 20;
      const totalRuns = 100;
      const batches = Math.ceil(totalRuns / batchSize);
      
      for (let batch = 0; batch < batches; batch++) {
        const batchRuns = Array.from({ length: Math.min(batchSize, totalRuns - batch * batchSize) }, (_, i) => {
          const runIndex = batch * batchSize + i;
          return {
            projectId,
            suiteId: `perf-suite-${runIndex}`,
            branch: runIndex % 5 === 0 ? 'develop' : 'main',
            status: (runIndex % 10 === 0 ? 'failed' : 'passed') as const,
            startTime: new Date(Date.now() - (runIndex * 3600000)).toISOString(), // Hourly runs
            endTime: new Date(Date.now() - (runIndex * 3600000) + 120000).toISOString(),
            duration: 120000 + (runIndex * 1000), // Varying durations
            tags: ['performance', `batch-${batch}`]
          };
        });
        
        await apiClient.post('/api/test-runs/bulk', { testRuns: batchRuns });
      }
      
      // Wait for all data to be processed
      await TestUtils.waitForCondition(
        async () => {
          const runs = await apiClient.getTestRuns({ projectId, limit: 100 });
          return runs.data?.testRuns?.length >= totalRuns;
        },
        120000 // 2 minutes
      );
      
      // Test UI performance with large dataset
      const endMeasurement = performanceMonitor.startMeasurement('large_dataset_ui_load');
      
      await testRunsPage.navigate();
      await testRunsPage.filterByProject(`Performance Test ${workflowId}`);
      await testRunsPage.waitForFilterResults();
      
      const uiLoadTime = endMeasurement();
      expect(uiLoadTime).toBeWithinTimeRange(0, 10000); // 10 second max
      
      // Test pagination performance
      const firstPageRuns = await testRunsPage.getVisibleTestRuns();
      expect(firstPageRuns.length).toBeLessThanOrEqual(20); // Page size limit
      
      // Navigate through pages
      const paginationInfo = await testRunsPage.getPaginationInfo();
      if (paginationInfo.totalPages > 1) {
        const pageNavTime = performanceMonitor.startMeasurement('page_navigation');
        
        await testRunsPage.goToNextPage();
        await testRunsPage.waitForPageChange();
        
        const navTime = pageNavTime();
        expect(navTime).toBeWithinTimeRange(0, 3000); // 3 second max
      }
      
      // Test summary visualization performance
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const summaryLoadTime = performanceMonitor.startMeasurement('summary_large_dataset');
      await testSummariesPage.refresh();
      await testSummariesPage.waitForSummariesLoad();
      const summaryTime = summaryLoadTime();
      
      expect(summaryTime).toBeWithinTimeRange(0, 8000); // 8 second max
    });

    test('should handle concurrent user interactions efficiently', async () => {
      // Simulate multiple concurrent user sessions
      const concurrentSessions = 5;
      const sessionPromises = [];
      
      for (let i = 0; i < concurrentSessions; i++) {
        const sessionPromise = (async () => {
          const sessionId = `concurrent-session-${i}`;
          
          // Each session performs different operations
          const operations = [
            () => testRunsPage.navigate(),
            () => testSummariesPage.navigate(),
            () => chatbotPage.navigate()
          ];
          
          const operation = operations[i % operations.length];
          await operation();
          
          // Perform some interactions
          if (i % 3 === 0) {
            // Test runs page interactions
            await testRunsPage.waitForTestRunsToLoad();
            await testRunsPage.filterByStatus('passed');
            await testRunsPage.waitForFilterResults();
          } else if (i % 3 === 1) {
            // Summaries page interactions
            await testSummariesPage.waitForSummariesLoad();
            const projects = await testSummariesPage.getProjectRows();
            if (projects.length > 0) {
              await testSummariesPage.toggleFavorite(projects[0].projectId);
            }
          } else {
            // Chatbot interactions
            await chatbotPage.openChatbot();
            await chatbotPage.sendMessage(`Hello from session ${i}`);
            await chatbotPage.waitForResponse();
          }
          
          return sessionId;
        })();
        
        sessionPromises.push(sessionPromise);
      }
      
      // All sessions should complete without errors
      const results = await Promise.allSettled(sessionPromises);
      
      const successfulSessions = results.filter(result => result.status === 'fulfilled');
      expect(successfulSessions.length).toBe(concurrentSessions);
    });
  });
});