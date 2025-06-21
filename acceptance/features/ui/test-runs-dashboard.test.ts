/**
 * Test Runs Dashboard - UI Acceptance Tests
 * 
 * Tests the complete test runs dashboard functionality including:
 * - Data loading and pagination
 * - Filtering and search
 * - Row expansion and details
 * - Navigation and deep linking
 * - Performance and responsiveness
 * 
 * Based on analysis of fern-ui issues and features
 */

import { TestUtils, HttpUtils, performanceMonitor } from '@acceptance/setup/test-helpers';
import { TestRunsPage } from '@acceptance/utils/page-objects/test-runs-page';
import { ApiClient } from '@acceptance/utils/api-clients/fern-api-client';

describe('Test Runs Dashboard', () => {
  let testRunsPage: TestRunsPage;
  let apiClient: ApiClient;
  let context: any;

  beforeAll(async () => {
    context = TestUtils.getTestContext();
    apiClient = new ApiClient(context.baseUrls);
    testRunsPage = new TestRunsPage(context.baseUrls.ui);
    
    // Ensure services are ready
    await TestUtils.waitForCondition(
      async () => {
        try {
          await apiClient.healthCheck();
          return true;
        } catch {
          return false;
        }
      },
      60000 // 1 minute timeout
    );
  });

  describe('Data Loading and Display', () => {
    test('should load test runs with proper pagination', async () => {
      const endMeasurement = performanceMonitor.startMeasurement('test_runs_page_load');
      
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      const pageLoadTime = endMeasurement();
      
      // Performance requirement: Page should load within 2 seconds
      expect(pageLoadTime).toBeWithinTimeRange(0, 2000);
      
      // Verify test runs are displayed
      const testRuns = await testRunsPage.getVisibleTestRuns();
      expect(testRuns.length).toBeGreaterThan(0);
      expect(testRuns.length).toBeLessThanOrEqual(20); // Default page size
      
      // Verify pagination controls
      const paginationInfo = await testRunsPage.getPaginationInfo();
      expect(paginationInfo).toHaveProperty('currentPage');
      expect(paginationInfo).toHaveProperty('totalPages');
      expect(paginationInfo).toHaveProperty('totalItems');
    });

    test('should display test run information correctly', async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      const firstTestRun = await testRunsPage.getTestRunByIndex(0);
      
      // Verify all required fields are displayed
      expect(firstTestRun).toHaveProperty('id');
      expect(firstTestRun).toHaveProperty('projectName');
      expect(firstTestRun).toHaveProperty('status');
      expect(firstTestRun).toHaveProperty('duration');
      expect(firstTestRun).toHaveProperty('startTime');
      expect(firstTestRun).toHaveProperty('branch');
      
      // Verify status indicators
      expect(['passed', 'failed', 'skipped']).toContain(firstTestRun.status);
      
      // Verify duration format
      expect(firstTestRun.duration).toMatch(/\d+[smh]/); // Should be formatted like "5m 30s"
    });

    test('should handle empty state gracefully', async () => {
      // Create test scenario with no data
      await apiClient.clearTestData(context.namespace);
      
      await testRunsPage.navigate();
      await testRunsPage.waitForEmptyState();
      
      const emptyMessage = await testRunsPage.getEmptyStateMessage();
      expect(emptyMessage).toContain('No test runs found');
      
      // Restore test data
      await context.dataFixtures.setupTestData();
    });
  });

  describe('Filtering and Search', () => {
    beforeEach(async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
    });

    test('should filter by project', async () => {
      const projects = await testRunsPage.getAvailableProjects();
      expect(projects.length).toBeGreaterThan(1);
      
      const selectedProject = projects[0];
      await testRunsPage.filterByProject(selectedProject);
      await testRunsPage.waitForFilterResults();
      
      const filteredRuns = await testRunsPage.getVisibleTestRuns();
      expect(filteredRuns.length).toBeGreaterThan(0);
      
      // Verify all results match the filter
      filteredRuns.forEach(run => {
        expect(run.projectName).toBe(selectedProject);
      });
    });

    test('should filter by status', async () => {
      // Filter by failed tests
      await testRunsPage.filterByStatus('failed');
      await testRunsPage.waitForFilterResults();
      
      const failedRuns = await testRunsPage.getVisibleTestRuns();
      
      if (failedRuns.length > 0) {
        failedRuns.forEach(run => {
          expect(run.status).toBe('failed');
        });
      }
    });

    test('should filter by branch', async () => {
      const branches = await testRunsPage.getAvailableBranches();
      expect(branches.length).toBeGreaterThan(0);
      
      const mainBranch = branches.find(b => b === 'main') || branches[0];
      await testRunsPage.filterByBranch(mainBranch);
      await testRunsPage.waitForFilterResults();
      
      const branchRuns = await testRunsPage.getVisibleTestRuns();
      branchRuns.forEach(run => {
        expect(run.branch).toBe(mainBranch);
      });
    });

    test('should search by test run ID or description', async () => {
      const searchTerm = 'auth'; // Should match authentication-related tests
      
      await testRunsPage.searchTests(searchTerm);
      await testRunsPage.waitForSearchResults();
      
      const searchResults = await testRunsPage.getVisibleTestRuns();
      
      if (searchResults.length > 0) {
        searchResults.forEach(run => {
          const matchesSearch = 
            run.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
            run.projectName.toLowerCase().includes(searchTerm.toLowerCase()) ||
            run.description?.toLowerCase().includes(searchTerm.toLowerCase());
          
          expect(matchesSearch).toBe(true);
        });
      }
    });

    test('should combine multiple filters', async () => {
      // Apply multiple filters
      await testRunsPage.filterByStatus('passed');
      await testRunsPage.filterByBranch('main');
      await testRunsPage.waitForFilterResults();
      
      const filteredRuns = await testRunsPage.getVisibleTestRuns();
      
      filteredRuns.forEach(run => {
        expect(run.status).toBe('passed');
        expect(run.branch).toBe('main');
      });
    });

    test('should clear filters and restore full dataset', async () => {
      // Apply some filters
      await testRunsPage.filterByStatus('failed');
      await testRunsPage.waitForFilterResults();
      
      const filteredCount = (await testRunsPage.getVisibleTestRuns()).length;
      
      // Clear filters
      await testRunsPage.clearAllFilters();
      await testRunsPage.waitForFilterResults();
      
      const fullCount = (await testRunsPage.getVisibleTestRuns()).length;
      expect(fullCount).toBeGreaterThanOrEqual(filteredCount);
    });
  });

  describe('Row Expansion and Details', () => {
    beforeEach(async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
    });

    test('should expand test run to show spec details', async () => {
      const firstTestRun = await testRunsPage.getTestRunByIndex(0);
      
      // Expand the row
      await testRunsPage.expandTestRun(firstTestRun.id);
      await testRunsPage.waitForSpecDetailsToLoad(firstTestRun.id);
      
      const specRuns = await testRunsPage.getSpecRuns(firstTestRun.id);
      expect(specRuns.length).toBeGreaterThan(0);
      
      // Verify spec run structure
      specRuns.forEach(spec => {
        expect(spec).toHaveProperty('description');
        expect(spec).toHaveProperty('status');
        expect(spec).toHaveProperty('duration');
        expect(['passed', 'failed', 'skipped']).toContain(spec.status);
      });
    });

    test('should show error details for failed specs', async () => {
      // Find a failed test run
      await testRunsPage.filterByStatus('failed');
      await testRunsPage.waitForFilterResults();
      
      const failedRuns = await testRunsPage.getVisibleTestRuns();
      
      if (failedRuns.length > 0) {
        const failedRun = failedRuns[0];
        
        await testRunsPage.expandTestRun(failedRun.id);
        await testRunsPage.waitForSpecDetailsToLoad(failedRun.id);
        
        const failedSpecs = await testRunsPage.getFailedSpecs(failedRun.id);
        
        if (failedSpecs.length > 0) {
          const firstFailedSpec = failedSpecs[0];
          expect(firstFailedSpec).toHaveProperty('errorMessage');
          expect(firstFailedSpec.errorMessage).toBeTruthy();
        }
      }
    });

    test('should collapse expanded rows', async () => {
      const firstTestRun = await testRunsPage.getTestRunByIndex(0);
      
      // Expand
      await testRunsPage.expandTestRun(firstTestRun.id);
      await testRunsPage.waitForSpecDetailsToLoad(firstTestRun.id);
      
      expect(await testRunsPage.isTestRunExpanded(firstTestRun.id)).toBe(true);
      
      // Collapse
      await testRunsPage.collapseTestRun(firstTestRun.id);
      
      expect(await testRunsPage.isTestRunExpanded(firstTestRun.id)).toBe(false);
    });
  });

  describe('Navigation and Deep Linking', () => {
    test('should navigate to specific test run via URL', async () => {
      // Get a test run ID from the API
      const testRuns = await apiClient.getTestRuns({ limit: 1 });
      expect(testRuns.data.length).toBeGreaterThan(0);
      
      const testRunId = testRuns.data[0].id;
      
      // Navigate directly to the test run
      await testRunsPage.navigateToTestRun(testRunId);
      await testRunsPage.waitForTestRunsToLoad();
      
      // Verify the test run is highlighted/focused
      const highlightedRun = await testRunsPage.getHighlightedTestRun();
      expect(highlightedRun.id).toBe(testRunId);
    });

    test('should maintain filter state in URL parameters', async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      // Apply filters
      await testRunsPage.filterByStatus('passed');
      await testRunsPage.filterByBranch('main');
      await testRunsPage.waitForFilterResults();
      
      // Check URL contains filter parameters
      const currentUrl = await testRunsPage.getCurrentUrl();
      expect(currentUrl).toContain('status=passed');
      expect(currentUrl).toContain('branch=main');
      
      // Refresh page and verify filters persist
      await testRunsPage.refresh();
      await testRunsPage.waitForTestRunsToLoad();
      
      const statusFilter = await testRunsPage.getCurrentStatusFilter();
      const branchFilter = await testRunsPage.getCurrentBranchFilter();
      
      expect(statusFilter).toBe('passed');
      expect(branchFilter).toBe('main');
    });

    test('should handle navigation to non-existent test run gracefully', async () => {
      const nonExistentId = 'non-existent-test-run-id';
      
      await testRunsPage.navigateToTestRun(nonExistentId);
      
      // Should show appropriate error message
      const errorMessage = await testRunsPage.getErrorMessage();
      expect(errorMessage).toContain('Test run not found');
    });
  });

  describe('Pagination', () => {
    beforeEach(async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
    });

    test('should navigate between pages', async () => {
      const paginationInfo = await testRunsPage.getPaginationInfo();
      
      if (paginationInfo.totalPages > 1) {
        // Go to next page
        await testRunsPage.goToNextPage();
        await testRunsPage.waitForPageChange();
        
        const newPaginationInfo = await testRunsPage.getPaginationInfo();
        expect(newPaginationInfo.currentPage).toBe(paginationInfo.currentPage + 1);
        
        // Go back to previous page
        await testRunsPage.goToPreviousPage();
        await testRunsPage.waitForPageChange();
        
        const backPaginationInfo = await testRunsPage.getPaginationInfo();
        expect(backPaginationInfo.currentPage).toBe(paginationInfo.currentPage);
      }
    });

    test('should change page size', async () => {
      const originalRuns = await testRunsPage.getVisibleTestRuns();
      const originalPageSize = originalRuns.length;
      
      // Change to different page size
      const newPageSize = originalPageSize === 10 ? 20 : 10;
      await testRunsPage.changePageSize(newPageSize);
      await testRunsPage.waitForPageSizeChange();
      
      const newRuns = await testRunsPage.getVisibleTestRuns();
      
      // Should have different number of items (unless total is less than new page size)
      if (await testRunsPage.getTotalItemCount() >= newPageSize) {
        expect(newRuns.length).toBe(newPageSize);
      }
    });

    test('should handle pagination with filters', async () => {
      // Apply filter that results in multiple pages
      await testRunsPage.filterByBranch('main');
      await testRunsPage.waitForFilterResults();
      
      const filteredPagination = await testRunsPage.getPaginationInfo();
      
      if (filteredPagination.totalPages > 1) {
        await testRunsPage.goToNextPage();
        await testRunsPage.waitForPageChange();
        
        // Verify filter is still applied on next page
        const runsOnPage2 = await testRunsPage.getVisibleTestRuns();
        runsOnPage2.forEach(run => {
          expect(run.branch).toBe('main');
        });
      }
    });
  });

  describe('Performance and Responsiveness', () => {
    test('should load large datasets efficiently', async () => {
      // Ensure we have a reasonable amount of test data
      const totalItems = await testRunsPage.getTotalItemCount();
      expect(totalItems).toBeGreaterThan(50); // Should have substantial test data
      
      const endMeasurement = performanceMonitor.startMeasurement('large_dataset_load');
      
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      const loadTime = endMeasurement();
      
      // Should load efficiently even with large datasets
      expect(loadTime).toBeWithinTimeRange(0, 3000); // 3 second max
    });

    test('should handle rapid filter changes without performance degradation', async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      const endMeasurement = performanceMonitor.startMeasurement('rapid_filter_changes');
      
      // Rapidly change filters
      await testRunsPage.filterByStatus('passed');
      await testRunsPage.waitForFilterResults();
      
      await testRunsPage.filterByStatus('failed');
      await testRunsPage.waitForFilterResults();
      
      await testRunsPage.filterByBranch('main');
      await testRunsPage.waitForFilterResults();
      
      await testRunsPage.clearAllFilters();
      await testRunsPage.waitForFilterResults();
      
      const totalTime = endMeasurement();
      
      // Rapid filter changes should complete quickly
      expect(totalTime).toBeWithinTimeRange(0, 5000); // 5 second max for all changes
    });

    test('should handle concurrent user interactions smoothly', async () => {
      await testRunsPage.navigate();
      await testRunsPage.waitForTestRunsToLoad();
      
      // Simulate concurrent interactions
      const interactions = [
        testRunsPage.filterByStatus('passed'),
        testRunsPage.expandTestRun((await testRunsPage.getTestRunByIndex(0)).id),
        testRunsPage.searchTests('auth')
      ];
      
      // All interactions should complete without errors
      await expect(Promise.all(interactions)).resolves.toBeDefined();
    });
  });

  describe('Error Handling', () => {
    test('should handle API failures gracefully', async () => {
      // Simulate API failure by using wrong endpoint
      await testRunsPage.navigate();
      
      // Mock API failure
      await apiClient.simulateFailure();
      
      await testRunsPage.refresh();
      
      // Should show appropriate error message
      const errorMessage = await testRunsPage.getErrorMessage();
      expect(errorMessage).toContain('Unable to load test runs');
      
      // Should provide retry option
      const retryButton = await testRunsPage.getRetryButton();
      expect(retryButton).toBeTruthy();
      
      // Restore API
      await apiClient.restoreConnection();
    });

    test('should handle network timeouts', async () => {
      await testRunsPage.navigate();
      
      // Simulate slow network
      await apiClient.simulateSlowNetwork(10000); // 10 second delay
      
      const promise = testRunsPage.refresh();
      
      // Should show loading indicator
      expect(await testRunsPage.isLoading()).toBe(true);
      
      // Should eventually timeout and show error
      await expect(promise).toRespondWithinTimeout(15000);
    });
  });
});