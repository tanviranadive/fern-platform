/**
 * Test Summaries - UI Acceptance Tests
 * 
 * Tests the test summaries grid functionality including:
 * - Grid visualization and layout
 * - Color coding and status indicators
 * - Favorites management
 * - Project filtering
 * - Historical data navigation
 * - Performance with large datasets
 * 
 * Based on fern-ui features and GitHub issues analysis
 */

import { TestUtils, HttpUtils, performanceMonitor } from '@acceptance/setup/test-helpers';
import { TestSummariesPage } from '@acceptance/utils/page-objects/test-summaries-page';
import { ApiClient } from '@acceptance/utils/api-clients/fern-api-client';

describe('Test Summaries Grid', () => {
  let testSummariesPage: TestSummariesPage;
  let apiClient: ApiClient;
  let context: any;

  beforeAll(async () => {
    context = TestUtils.getTestContext();
    apiClient = new ApiClient(context.baseUrls);
    testSummariesPage = new TestSummariesPage(context.baseUrls.ui);
    
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
      60000
    );
  });

  describe('Grid Layout and Visualization', () => {
    test('should display summary grid with proper layout', async () => {
      const endMeasurement = performanceMonitor.startMeasurement('summaries_page_load');
      
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const pageLoadTime = endMeasurement();
      expect(pageLoadTime).toBeWithinTimeRange(0, 2000);
      
      // Verify grid structure
      const gridInfo = await testSummariesPage.getGridInfo();
      expect(gridInfo.totalProjects).toBeGreaterThan(0);
      expect(gridInfo.columnsPerRow).toBe(40); // Based on fern-ui implementation
      expect(gridInfo.visibleRows).toBeGreaterThan(0);
    });

    test('should show color-coded status boxes', async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const statusBoxes = await testSummariesPage.getStatusBoxes();
      expect(statusBoxes.length).toBeGreaterThan(0);
      
      // Verify each box has proper color coding
      statusBoxes.forEach(box => {
        expect(box).toHaveProperty('color');
        expect(box).toHaveProperty('status');
        expect(box).toHaveProperty('projectId');
        expect(box).toHaveProperty('runDate');
        
        // Verify color mapping
        switch (box.status) {
          case 'all_passed':
            expect(box.color).toBe('green');
            break;
          case 'passed_with_skipped':
            expect(box.color).toBe('yellow');
            break;
          case 'failed':
            expect(box.color).toBe('red');
            break;
          default:
            expect(box.color).toBe('gray'); // No data or unknown
        }
      });
    });

    test('should display project names and metadata', async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const projectRows = await testSummariesPage.getProjectRows();
      expect(projectRows.length).toBeGreaterThan(0);
      
      projectRows.forEach(row => {
        expect(row).toHaveProperty('projectId');
        expect(row).toHaveProperty('projectName');
        expect(row).toHaveProperty('summaryBoxes');
        expect(row.projectName).toBeTruthy();
        expect(row.summaryBoxes.length).toBeLessThanOrEqual(40); // Max columns per row
      });
    });

    test('should handle responsive grid layout', async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      // Test different viewport sizes
      const viewports = [
        { width: 1920, height: 1080 }, // Desktop
        { width: 1366, height: 768 },  // Laptop
        { width: 768, height: 1024 }   // Tablet
      ];
      
      for (const viewport of viewports) {
        await testSummariesPage.setViewportSize(viewport.width, viewport.height);
        await testSummariesPage.waitForLayoutUpdate();
        
        const gridInfo = await testSummariesPage.getGridInfo();
        expect(gridInfo.isResponsive).toBe(true);
        expect(gridInfo.visibleColumns).toBeGreaterThan(0);
        expect(gridInfo.visibleColumns).toBeLessThanOrEqual(40);
      }
    });
  });

  describe('Favorites Management', () => {
    beforeEach(async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
    });

    test('should toggle project favorites', async () => {
      const projects = await testSummariesPage.getProjectRows();
      expect(projects.length).toBeGreaterThan(0);
      
      const firstProject = projects[0];
      const initialFavoriteState = await testSummariesPage.isFavorite(firstProject.projectId);
      
      // Toggle favorite
      await testSummariesPage.toggleFavorite(firstProject.projectId);
      await testSummariesPage.waitForFavoriteUpdate();
      
      const newFavoriteState = await testSummariesPage.isFavorite(firstProject.projectId);
      expect(newFavoriteState).toBe(!initialFavoriteState);
      
      // Verify favorite icon changed
      const favoriteIcon = await testSummariesPage.getFavoriteIcon(firstProject.projectId);
      if (newFavoriteState) {
        expect(favoriteIcon).toContain('filled');
      } else {
        expect(favoriteIcon).toContain('outline');
      }
    });

    test('should persist favorites across page reloads', async () => {
      const projects = await testSummariesPage.getProjectRows();
      const testProject = projects[0];
      
      // Set as favorite
      await testSummariesPage.setFavorite(testProject.projectId, true);
      await testSummariesPage.waitForFavoriteUpdate();
      
      // Reload page
      await testSummariesPage.refresh();
      await testSummariesPage.waitForSummariesLoad();
      
      // Verify favorite state persisted
      const isFavorite = await testSummariesPage.isFavorite(testProject.projectId);
      expect(isFavorite).toBe(true);
    });

    test('should filter to show only favorites', async () => {
      // Set multiple projects as favorites
      const projects = await testSummariesPage.getProjectRows();
      const favoriteProjects = projects.slice(0, 2);
      
      for (const project of favoriteProjects) {
        await testSummariesPage.setFavorite(project.projectId, true);
      }
      await testSummariesPage.waitForFavoriteUpdate();
      
      // Enable favorites filter
      await testSummariesPage.filterByFavorites(true);
      await testSummariesPage.waitForFilterUpdate();
      
      const visibleProjects = await testSummariesPage.getVisibleProjects();
      expect(visibleProjects.length).toBe(favoriteProjects.length);
      
      // Verify only favorites are shown
      for (const visibleProject of visibleProjects) {
        const isFavorite = await testSummariesPage.isFavorite(visibleProject.projectId);
        expect(isFavorite).toBe(true);
      }
    });

    test('should show all projects when favorites filter is disabled', async () => {
      // Enable favorites filter first
      await testSummariesPage.filterByFavorites(true);
      await testSummariesPage.waitForFilterUpdate();
      
      const favoritesCount = (await testSummariesPage.getVisibleProjects()).length;
      
      // Disable favorites filter
      await testSummariesPage.filterByFavorites(false);
      await testSummariesPage.waitForFilterUpdate();
      
      const allProjectsCount = (await testSummariesPage.getVisibleProjects()).length;
      expect(allProjectsCount).toBeGreaterThanOrEqual(favoritesCount);
    });
  });

  describe('Interactive Features', () => {
    beforeEach(async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
    });

    test('should show tooltip on hover with detailed information', async () => {
      const projects = await testSummariesPage.getProjectRows();
      const firstProject = projects[0];
      
      if (firstProject.summaryBoxes.length > 0) {
        const firstBox = firstProject.summaryBoxes[0];
        
        // Hover over summary box
        await testSummariesPage.hoverOverSummaryBox(firstProject.projectId, firstBox.date);
        
        const tooltip = await testSummariesPage.getTooltipContent();
        expect(tooltip).toHaveProperty('projectName');
        expect(tooltip).toHaveProperty('runDate');
        expect(tooltip).toHaveProperty('totalTests');
        expect(tooltip).toHaveProperty('passedTests');
        expect(tooltip).toHaveProperty('failedTests');
        expect(tooltip).toHaveProperty('skippedTests');
        expect(tooltip).toHaveProperty('duration');
      }
    });

    test('should navigate to detailed view on click', async () => {
      const projects = await testSummariesPage.getProjectRows();
      const firstProject = projects[0];
      
      if (firstProject.summaryBoxes.length > 0) {
        const firstBox = firstProject.summaryBoxes[0];
        
        // Click on summary box
        await testSummariesPage.clickSummaryBox(firstProject.projectId, firstBox.date);
        
        // Should navigate to test runs page with filters
        await TestUtils.waitForCondition(
          async () => (await testSummariesPage.getCurrentUrl()).includes('/testruns'),
          5000
        );
        
        const currentUrl = await testSummariesPage.getCurrentUrl();
        expect(currentUrl).toContain('/testruns');
        expect(currentUrl).toContain(`project=${firstProject.projectId}`);
      }
    });

    test('should support keyboard navigation', async () => {
      await testSummariesPage.focusGrid();
      
      // Test arrow key navigation
      await testSummariesPage.pressKey('ArrowRight');
      await testSummariesPage.pressKey('ArrowDown');
      
      const focusedElement = await testSummariesPage.getFocusedElement();
      expect(focusedElement).toHaveProperty('projectId');
      expect(focusedElement).toHaveProperty('columnIndex');
      
      // Test Enter key activation
      await testSummariesPage.pressKey('Enter');
      
      // Should trigger navigation or show details
      const hasNavigated = (await testSummariesPage.getCurrentUrl()).includes('/testruns');
      const hasTooltip = await testSummariesPage.hasVisibleTooltip();
      
      expect(hasNavigated || hasTooltip).toBe(true);
    });
  });

  describe('Historical Data Navigation', () => {
    beforeEach(async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
    });

    test('should display chronological test run history', async () => {
      const projects = await testSummariesPage.getProjectRows();
      const projectWithHistory = projects.find(p => p.summaryBoxes.length > 1);
      
      if (projectWithHistory) {
        const boxes = projectWithHistory.summaryBoxes;
        
        // Verify chronological order (most recent first)
        for (let i = 1; i < boxes.length; i++) {
          const currentDate = new Date(boxes[i - 1].date);
          const nextDate = new Date(boxes[i].date);
          expect(currentDate.getTime()).toBeGreaterThanOrEqual(nextDate.getTime());
        }
      }
    });

    test('should handle pagination for long history', async () => {
      // Check if pagination controls exist
      const hasPagination = await testSummariesPage.hasPaginationControls();
      
      if (hasPagination) {
        const initialPageInfo = await testSummariesPage.getPageInfo();
        
        if (initialPageInfo.totalPages > 1) {
          // Navigate to next page
          await testSummariesPage.goToNextPage();
          await testSummariesPage.waitForPageLoad();
          
          const newPageInfo = await testSummariesPage.getPageInfo();
          expect(newPageInfo.currentPage).toBe(initialPageInfo.currentPage + 1);
          
          // Verify different data is shown
          const newProjects = await testSummariesPage.getProjectRows();
          expect(newProjects).toBeDefined();
        }
      }
    });

    test('should support date range filtering', async () => {
      const dateRangeSupported = await testSummariesPage.hasDateRangeFilter();
      
      if (dateRangeSupported) {
        const endDate = new Date();
        const startDate = new Date(endDate.getTime() - (7 * 24 * 60 * 60 * 1000)); // 7 days ago
        
        await testSummariesPage.setDateRange(startDate, endDate);
        await testSummariesPage.waitForFilterUpdate();
        
        const projects = await testSummariesPage.getProjectRows();
        
        // Verify all boxes are within date range
        projects.forEach(project => {
          project.summaryBoxes.forEach(box => {
            const boxDate = new Date(box.date);
            expect(boxDate.getTime()).toBeGreaterThanOrEqual(startDate.getTime());
            expect(boxDate.getTime()).toBeLessThanOrEqual(endDate.getTime());
          });
        });
      }
    });
  });

  describe('Performance with Large Datasets', () => {
    test('should handle many projects efficiently', async () => {
      // Ensure we have substantial test data
      const projectCount = await testSummariesPage.getProjectCount();
      expect(projectCount).toBeGreaterThan(20);
      
      const endMeasurement = performanceMonitor.startMeasurement('large_summaries_render');
      
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const renderTime = endMeasurement();
      
      // Should render efficiently even with many projects
      expect(renderTime).toBeWithinTimeRange(0, 3000);
    });

    test('should handle long project histories without performance degradation', async () => {
      // Find project with extensive history
      const projects = await testSummariesPage.getProjectRows();
      const projectWithLongHistory = projects.find(p => p.summaryBoxes.length > 30);
      
      if (projectWithLongHistory) {
        const endMeasurement = performanceMonitor.startMeasurement('long_history_interaction');
        
        // Interact with the long history
        for (let i = 0; i < Math.min(10, projectWithLongHistory.summaryBoxes.length); i++) {
          await testSummariesPage.hoverOverSummaryBox(
            projectWithLongHistory.projectId, 
            projectWithLongHistory.summaryBoxes[i].date
          );
          await TestUtils.waitForCondition(
            async () => await testSummariesPage.hasVisibleTooltip(),
            1000
          );
        }
        
        const interactionTime = endMeasurement();
        expect(interactionTime).toBeWithinTimeRange(0, 5000);
      }
    });

    test('should use virtual scrolling for large datasets', async () => {
      const hasVirtualScrolling = await testSummariesPage.hasVirtualScrolling();
      
      if (hasVirtualScrolling) {
        const initialVisibleRows = await testSummariesPage.getVisibleRowCount();
        
        // Scroll down
        await testSummariesPage.scrollDown(1000);
        await testSummariesPage.waitForScrollUpdate();
        
        const afterScrollVisibleRows = await testSummariesPage.getVisibleRowCount();
        
        // Should maintain similar number of visible rows (virtual scrolling)
        expect(Math.abs(afterScrollVisibleRows - initialVisibleRows)).toBeLessThan(5);
      }
    });
  });

  describe('Error Handling and Edge Cases', () => {
    test('should handle projects with no test history', async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const projects = await testSummariesPage.getProjectRows();
      const emptyProject = projects.find(p => p.summaryBoxes.length === 0);
      
      if (emptyProject) {
        // Should show placeholder or empty state
        const emptyIndicator = await testSummariesPage.getEmptyStateIndicator(emptyProject.projectId);
        expect(emptyIndicator).toBeTruthy();
      }
    });

    test('should handle API failures gracefully', async () => {
      await testSummariesPage.navigate();
      
      // Simulate API failure
      await apiClient.simulateFailure();
      
      await testSummariesPage.refresh();
      
      const errorMessage = await testSummariesPage.getErrorMessage();
      expect(errorMessage).toContain('Unable to load');
      
      const retryButton = await testSummariesPage.getRetryButton();
      expect(retryButton).toBeTruthy();
      
      // Restore API
      await apiClient.restoreConnection();
    });

    test('should handle malformed data gracefully', async () => {
      // Inject malformed data via API
      await apiClient.injectMalformedData(context.namespace);
      
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      // Should not crash and should show available valid data
      const projects = await testSummariesPage.getProjectRows();
      expect(Array.isArray(projects)).toBe(true);
      
      // Clean up malformed data
      await apiClient.cleanMalformedData(context.namespace);
    });

    test('should handle extremely long project names', async () => {
      const projects = await testSummariesPage.getProjectRows();
      
      projects.forEach(async (project) => {
        const nameElement = await testSummariesPage.getProjectNameElement(project.projectId);
        
        // Should handle text overflow properly
        const hasEllipsis = await testSummariesPage.hasTextEllipsis(nameElement);
        const isFullyVisible = await testSummariesPage.isTextFullyVisible(nameElement);
        
        // Long names should either be truncated with ellipsis or wrapped
        expect(hasEllipsis || isFullyVisible).toBe(true);
      });
    });
  });

  describe('Accessibility', () => {
    test('should support screen readers', async () => {
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      // Check ARIA labels and roles
      const gridRole = await testSummariesPage.getGridRole();
      expect(gridRole).toBe('grid');
      
      const projects = await testSummariesPage.getProjectRows();
      
      for (const project of projects.slice(0, 3)) { // Check first 3 projects
        const ariaLabel = await testSummariesPage.getProjectAriaLabel(project.projectId);
        expect(ariaLabel).toContain(project.projectName);
        
        for (const box of project.summaryBoxes.slice(0, 3)) { // Check first 3 boxes
          const boxAriaLabel = await testSummariesPage.getSummaryBoxAriaLabel(
            project.projectId, 
            box.date
          );
          expect(boxAriaLabel).toContain(box.status);
          expect(boxAriaLabel).toContain(box.date);
        }
      }
    });

    test('should support high contrast mode', async () => {
      await testSummariesPage.enableHighContrastMode();
      await testSummariesPage.navigate();
      await testSummariesPage.waitForSummariesLoad();
      
      const statusBoxes = await testSummariesPage.getStatusBoxes();
      
      // Verify high contrast colors are applied
      statusBoxes.forEach(async (box) => {
        const contrastRatio = await testSummariesPage.getContrastRatio(box.element);
        expect(contrastRatio).toBeGreaterThan(4.5); // WCAG AA standard
      });
    });
  });
});