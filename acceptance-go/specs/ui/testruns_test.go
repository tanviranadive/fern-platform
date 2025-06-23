package ui_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Runs UI", func() {
	var (
		testRunsPage = NewTestRunsPage()
	)

	BeforeEach(func() {
		By("Navigating to test runs page")
		err := testRunsPage.Navigate()
		Expect(err).NotTo(HaveOccurred())
		
		err = testRunsPage.WaitForTestRunsToLoad()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Data Loading and Display", func() {
		It("should load test runs with proper pagination", func() {
			By("Measuring test runs page load time")
			startTime := time.Now()
			
			err := testRunsPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			loadTime := time.Since(startTime)
			Expect(loadTime).To(BeNumerically("<", 3*time.Second),
				"Test runs page should load within 3 seconds")
			
			By("Verifying test runs are displayed")
			testRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(testRuns)).To(BeNumerically(">", 0))
			Expect(len(testRuns)).To(BeNumerically("<=", 20)) // Default page size
			
			By("Verifying pagination controls")
			paginationInfo, err := testRunsPage.GetPaginationInfo()
			Expect(err).NotTo(HaveOccurred())
			Expect(paginationInfo).NotTo(BeNil())
			Expect(paginationInfo.CurrentPage).To(BeNumerically(">=", 1))
			Expect(paginationInfo.TotalPages).To(BeNumerically(">=", 1))
			Expect(paginationInfo.TotalItems).To(BeNumerically(">=", len(testRuns)))
		})

		It("should display test run information correctly", func() {
			By("Getting the first test run")
			firstTestRun, err := testRunsPage.GetTestRunByIndex(0)
			Expect(err).NotTo(HaveOccurred())
			Expect(firstTestRun).NotTo(BeNil())
			
			By("Verifying test run structure")
			Expect(firstTestRun.ID).NotTo(BeEmpty())
			Expect(firstTestRun.ProjectName).NotTo(BeEmpty())
			Expect(firstTestRun.Status).To(BeElementOf("passed", "failed", "skipped"))
			Expect(firstTestRun.Duration).NotTo(BeEmpty())
			Expect(firstTestRun.StartTime).NotTo(BeEmpty())
			Expect(firstTestRun.Branch).NotTo(BeEmpty())
			
			By("Verifying status indicators are properly formatted")
			Expect(firstTestRun.Status).To(MatchRegexp("^(passed|failed|skipped)$"))
			
			By("Verifying duration format")
			Expect(firstTestRun.Duration).To(MatchRegexp(`\d+(\.\d+)?\s*(ms|s|m|h)`))
		})

		It("should handle empty state gracefully", func() {
			// This test simulates an empty state by applying filters that return no results
			By("Applying filters that should return no results")
			err := testRunsPage.SearchTests("non-existent-search-term-12345")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForSearchResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Checking if empty state is handled")
			testRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			if len(testRuns) == 0 {
				// Verify empty state message if no results
				emptyMessage, err := testRunsPage.GetEmptyStateMessage()
				if err == nil && emptyMessage != "" {
					Expect(emptyMessage).To(ContainSubstring("No test runs found"))
				}
			}
			
			By("Clearing search to restore data")
			err = testRunsPage.ClearAllFilters()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Filtering and Search", func() {
		BeforeEach(func() {
			// Ensure we're on a clean test runs page
			err := testRunsPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should filter by project", func() {
			By("Getting available projects")
			projects, err := testRunsPage.GetAvailableProjects()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(projects)).To(BeNumerically(">", 0))
			
			selectedProject := projects[0]
			
			By(fmt.Sprintf("Filtering by project: %s", selectedProject))
			err = testRunsPage.FilterByProject(selectedProject)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying filtered results")
			filteredRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			// All results should match the filter
			for _, run := range filteredRuns {
				Expect(run.ProjectName).To(Equal(selectedProject))
			}
		})

		It("should filter by status", func() {
			By("Filtering by failed status")
			err := testRunsPage.FilterByStatus("failed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying filtered results")
			failedRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			// All results should be failed
			for _, run := range failedRuns {
				Expect(run.Status).To(Equal("failed"))
			}
		})

		It("should filter by branch", func() {
			By("Getting available branches")
			branches, err := testRunsPage.GetAvailableBranches()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(branches)).To(BeNumerically(">", 0))
			
			mainBranch := "main"
			// Use main branch if available, otherwise use first available
			for _, branch := range branches {
				if branch == "main" {
					mainBranch = branch
					break
				}
			}
			if mainBranch != "main" {
				mainBranch = branches[0]
			}
			
			By(fmt.Sprintf("Filtering by branch: %s", mainBranch))
			err = testRunsPage.FilterByBranch(mainBranch)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying filtered results")
			branchRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			for _, run := range branchRuns {
				Expect(run.Branch).To(Equal(mainBranch))
			}
		})

		It("should search by test run content", func() {
			By("Performing search for 'test'")
			searchTerm := "test"
			
			err := testRunsPage.SearchTests(searchTerm)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForSearchResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying search results")
			searchResults, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			// Results should contain the search term (case-insensitive)
			for _, run := range searchResults {
				containsSearchTerm := 
					contains(run.ID, searchTerm) ||
					contains(run.ProjectName, searchTerm) ||
					contains(run.Description, searchTerm) ||
					contains(run.Branch, searchTerm)
				
				Expect(containsSearchTerm).To(BeTrue(), 
					"Test run should contain search term: %s", searchTerm)
			}
		})

		It("should combine multiple filters", func() {
			By("Applying multiple filters")
			err := testRunsPage.FilterByStatus("passed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.FilterByBranch("main")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying combined filter results")
			filteredRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			for _, run := range filteredRuns {
				Expect(run.Status).To(Equal("passed"))
				Expect(run.Branch).To(Equal("main"))
			}
		})

		It("should clear filters and restore full dataset", func() {
			By("Applying filters")
			err := testRunsPage.FilterByStatus("failed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			filteredCount := len(func() []interface{} {
				runs, _ := testRunsPage.GetVisibleTestRuns()
				result := make([]interface{}, len(runs))
				for i, run := range runs {
					result[i] = run
				}
				return result
			}())
			
			By("Clearing all filters")
			err = testRunsPage.ClearAllFilters()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying full dataset is restored")
			fullRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			fullCount := len(fullRuns)
			Expect(fullCount).To(BeNumerically(">=", filteredCount))
		})
	})

	Describe("Row Expansion and Details", func() {
		BeforeEach(func() {
			err := testRunsPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should expand test run to show spec details", func() {
			By("Getting the first test run")
			firstTestRun, err := testRunsPage.GetTestRunByIndex(0)
			Expect(err).NotTo(HaveOccurred())
			
			By(fmt.Sprintf("Expanding test run: %s", firstTestRun.ID))
			err = testRunsPage.ExpandTestRun(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForSpecDetailsToLoad(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying expansion state")
			isExpanded, err := testRunsPage.IsTestRunExpanded(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(isExpanded).To(BeTrue())
			
			By("Getting spec runs")
			specRuns, err := testRunsPage.GetSpecRuns(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(specRuns)).To(BeNumerically(">=", 0))
			
			// Verify spec run structure if any exist
			for _, spec := range specRuns {
				Expect(spec.ID).NotTo(BeEmpty())
				Expect(spec.Description).NotTo(BeEmpty())
				Expect(spec.Status).To(BeElementOf("passed", "failed", "skipped"))
				Expect(spec.Duration).NotTo(BeEmpty())
			}
		})

		It("should show error details for failed specs", func() {
			By("Looking for failed test runs")
			err := testRunsPage.FilterByStatus("failed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			failedRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			if len(failedRuns) > 0 {
				failedRun := failedRuns[0]
				
				By(fmt.Sprintf("Expanding failed test run: %s", failedRun.ID))
				err = testRunsPage.ExpandTestRun(failedRun.ID)
				Expect(err).NotTo(HaveOccurred())
				
				err = testRunsPage.WaitForSpecDetailsToLoad(failedRun.ID)
				Expect(err).NotTo(HaveOccurred())
				
				By("Getting failed specs")
				failedSpecs, err := testRunsPage.GetFailedSpecs(failedRun.ID)
				Expect(err).NotTo(HaveOccurred())
				
				// If there are failed specs, they should have error messages
				for _, spec := range failedSpecs {
					Expect(spec.Status).To(Equal("failed"))
					Expect(spec.ErrorMessage).NotTo(BeEmpty())
				}
			} else {
				Skip("No failed test runs available for error details test")
			}
		})

		It("should collapse expanded rows", func() {
			By("Getting a test run to expand")
			firstTestRun, err := testRunsPage.GetTestRunByIndex(0)
			Expect(err).NotTo(HaveOccurred())
			
			By("Expanding the test run")
			err = testRunsPage.ExpandTestRun(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForSpecDetailsToLoad(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			
			isExpanded, err := testRunsPage.IsTestRunExpanded(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(isExpanded).To(BeTrue())
			
			By("Collapsing the test run")
			err = testRunsPage.CollapseTestRun(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying collapse state")
			isExpanded, err = testRunsPage.IsTestRunExpanded(firstTestRun.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(isExpanded).To(BeFalse())
		})
	})

	Describe("Pagination", func() {
		BeforeEach(func() {
			err := testRunsPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should navigate between pages", func() {
			By("Getting pagination information")
			paginationInfo, err := testRunsPage.GetPaginationInfo()
			Expect(err).NotTo(HaveOccurred())
			
			if paginationInfo.TotalPages > 1 {
				originalPage := paginationInfo.CurrentPage
				
				By("Going to next page")
				err = testRunsPage.GoToNextPage()
				Expect(err).NotTo(HaveOccurred())
				
				err = testRunsPage.WaitForPageChange()
				Expect(err).NotTo(HaveOccurred())
				
				By("Verifying page change")
				newPaginationInfo, err := testRunsPage.GetPaginationInfo()
				Expect(err).NotTo(HaveOccurred())
				Expect(newPaginationInfo.CurrentPage).To(Equal(originalPage + 1))
				
				By("Going back to previous page")
				err = testRunsPage.GoToPreviousPage()
				Expect(err).NotTo(HaveOccurred())
				
				err = testRunsPage.WaitForPageChange()
				Expect(err).NotTo(HaveOccurred())
				
				By("Verifying return to original page")
				backPaginationInfo, err := testRunsPage.GetPaginationInfo()
				Expect(err).NotTo(HaveOccurred())
				Expect(backPaginationInfo.CurrentPage).To(Equal(originalPage))
			} else {
				Skip("Only one page available, skipping pagination test")
			}
		})

		It("should change page size", func() {
			By("Getting current test runs")
			originalRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			originalPageSize := len(originalRuns)
			
			By("Changing page size")
			newPageSize := 10
			if originalPageSize == 10 {
				newPageSize = 20
			}
			
			err = testRunsPage.ChangePageSize(newPageSize)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForPageSizeChange()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying page size change")
			newRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			totalItems, err := testRunsPage.GetTotalItemCount()
			Expect(err).NotTo(HaveOccurred())
			
			// Should have new page size items (unless total is less than new page size)
			if totalItems >= newPageSize {
				Expect(len(newRuns)).To(Equal(newPageSize))
			} else {
				Expect(len(newRuns)).To(Equal(totalItems))
			}
		})
	})

	Describe("Navigation and Deep Linking", func() {
		It("should navigate to specific test run via URL", func() {
			By("Getting a test run ID")
			testRuns, err := testRunsPage.GetVisibleTestRuns()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(testRuns)).To(BeNumerically(">", 0))
			
			testRunID := testRuns[0].ID
			
			By(fmt.Sprintf("Navigating directly to test run: %s", testRunID))
			err = testRunsPage.NavigateToTestRun(testRunID)
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying test run is highlighted or focused")
			highlightedRun, err := testRunsPage.GetHighlightedTestRun()
			if err == nil && highlightedRun != nil {
				Expect(highlightedRun.ID).To(Equal(testRunID))
			}
		})

		It("should maintain filter state in URL parameters", func() {
			By("Applying filters")
			err := testRunsPage.FilterByStatus("passed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.FilterByBranch("main")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			By("Checking URL contains filter parameters")
			currentURL, err := testRunsPage.GetCurrentURL()
			Expect(err).NotTo(HaveOccurred())
			Expect(currentURL).To(ContainSubstring("status=passed"))
			Expect(currentURL).To(ContainSubstring("branch=main"))
			
			By("Refreshing page and verifying filters persist")
			err = testRunsPage.Refresh()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			statusFilter, err := testRunsPage.GetCurrentStatusFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(statusFilter).To(Equal("passed"))
			
			branchFilter, err := testRunsPage.GetCurrentBranchFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(branchFilter).To(Equal("main"))
		})
	})

	Describe("Performance and Error Handling", func() {
		It("should handle rapid filter changes without performance degradation", func() {
			By("Performing rapid filter changes")
			startTime := time.Now()
			
			// Rapidly change filters
			err := testRunsPage.FilterByStatus("passed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.FilterByStatus("failed")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.FilterByBranch("main")
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.ClearAllFilters()
			Expect(err).NotTo(HaveOccurred())
			
			err = testRunsPage.WaitForFilterResults()
			Expect(err).NotTo(HaveOccurred())
			
			totalTime := time.Since(startTime)
			
			By("Verifying rapid filter changes complete within reasonable time")
			Expect(totalTime).To(BeNumerically("<", 10*time.Second),
				"Rapid filter changes should complete within 10 seconds")
		})

		It("should handle loading states properly", func() {
			By("Checking loading states during navigation")
			err := testRunsPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			// Check if loading indicator appears
			isLoading, err := testRunsPage.IsLoading()
			if err == nil && isLoading {
				// Wait for loading to complete
				Eventually(func() bool {
					loading, _ := testRunsPage.IsLoading()
					return !loading
				}, 10*time.Second, 500*time.Millisecond).Should(BeTrue())
			}
			
			err = testRunsPage.WaitForTestRunsToLoad()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// Helper function to check if a string contains another string (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(substr) == 0 || 
		 (len(s) > 0 && len(substr) > 0 && 
		  strings.Contains(strings.ToLower(s), strings.ToLower(substr))))
}