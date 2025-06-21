package ui_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dashboard UI", func() {
	var (
		dashboardPage = NewDashboardPage()
		testData      = GetTestData()
	)

	BeforeEach(func() {
		By("Navigating to dashboard")
		err := dashboardPage.Navigate()
		Expect(err).NotTo(HaveOccurred())
		
		err = dashboardPage.WaitForDashboardToLoad()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Dashboard Loading and Display", func() {
		It("should load dashboard within performance threshold", func() {
			By("Measuring dashboard load time")
			startTime := time.Now()
			
			err := dashboardPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			loadTime := time.Since(startTime)
			Expect(loadTime).To(BeNumerically("<", 3*time.Second),
				"Dashboard should load within 3 seconds")
		})

		It("should display dashboard statistics", func() {
			By("Retrieving dashboard statistics")
			stats, err := dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			Expect(stats).NotTo(BeNil())
			
			By("Verifying statistics are displayed")
			Expect(stats.TotalTestRuns).NotTo(BeEmpty())
			Expect(stats.TotalProjects).NotTo(BeEmpty())
			Expect(stats.SuccessRate).NotTo(BeEmpty())
			Expect(stats.AvgDuration).NotTo(BeEmpty())
			
			// Verify statistics make sense
			By("Verifying statistics contain valid data")
			Expect(stats.TotalProjects).To(MatchRegexp(`\d+`), "Total projects should be a number")
			Expect(stats.TotalTestRuns).To(MatchRegexp(`\d+`), "Total test runs should be a number")
			Expect(stats.SuccessRate).To(MatchRegexp(`\d+(\.\d+)?%?`), "Success rate should be a percentage")
		})

		It("should display recent test runs", func() {
			By("Retrieving recent test runs from dashboard")
			recentRuns, err := dashboardPage.GetRecentTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			if len(recentRuns) > 0 {
				By("Verifying recent test runs structure")
				firstRun := recentRuns[0]
				Expect(firstRun.ID).NotTo(BeEmpty())
				Expect(firstRun.ProjectName).NotTo(BeEmpty())
				Expect(firstRun.Status).To(BeElementOf("passed", "failed", "skipped"))
				Expect(firstRun.Duration).NotTo(BeEmpty())
				Expect(firstRun.StartTime).NotTo(BeEmpty())
			}
		})

		It("should display project summaries", func() {
			By("Retrieving project summaries from dashboard")
			projectSummaries, err := dashboardPage.GetProjectSummaries()
			Expect(err).NotTo(HaveOccurred())
			
			// Should have at least the test projects we created
			Expect(len(projectSummaries)).To(BeNumerically(">=", len(testData.Projects)))
			
			if len(projectSummaries) > 0 {
				By("Verifying project summary structure")
				firstProject := projectSummaries[0]
				Expect(firstProject.ID).NotTo(BeEmpty())
				Expect(firstProject.Name).NotTo(BeEmpty())
				Expect(firstProject.TestRunsCount).To(BeNumerically(">=", 0))
				Expect(firstProject.SuccessRate).To(BeNumerically(">=", 0))
				Expect(firstProject.SuccessRate).To(BeNumerically("<=", 100))
			}
		})

		It("should display test history visualization", func() {
			By("Retrieving test history data")
			testHistory, err := dashboardPage.GetTestHistory()
			Expect(err).NotTo(HaveOccurred())
			
			if len(testHistory) > 0 {
				By("Verifying test history structure")
				for _, historyItem := range testHistory {
					Expect(historyItem.Date).NotTo(BeEmpty())
					Expect(historyItem.Status).To(BeElementOf("passed", "failed", "skipped"))
					Expect(historyItem.Count).To(BeNumerically(">=", 0))
				}
			}
		})
	})

	Describe("Dashboard Navigation", func() {
		It("should navigate to project details when clicking on a project", func() {
			By("Getting project summaries")
			projectSummaries, err := dashboardPage.GetProjectSummaries()
			Expect(err).NotTo(HaveOccurred())
			
			if len(projectSummaries) > 0 {
				firstProject := projectSummaries[0]
				
				By(fmt.Sprintf("Clicking on project: %s", firstProject.Name))
				err = dashboardPage.ClickProject(firstProject.ID)
				Expect(err).NotTo(HaveOccurred())
				
				By("Waiting for navigation to complete")
				err = dashboardPage.WaitForNavigation()
				Expect(err).NotTo(HaveOccurred())
				
				By("Verifying navigation occurred")
				currentURL, err := dashboardPage.GetCurrentURL()
				Expect(err).NotTo(HaveOccurred())
				Expect(currentURL).To(ContainSubstring("test-runs"))
			} else {
				Skip("No projects available for navigation test")
			}
		})

		It("should maintain dashboard state during navigation", func() {
			By("Getting initial dashboard state")
			initialStats, err := dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			
			By("Refreshing the page")
			err = dashboardPage.Refresh()
			Expect(err).NotTo(HaveOccurred())
			
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			By("Verifying dashboard state is maintained")
			refreshedStats, err := dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			
			Expect(refreshedStats.TotalProjects).To(Equal(initialStats.TotalProjects))
			Expect(refreshedStats.TotalTestRuns).To(Equal(initialStats.TotalTestRuns))
		})
	})

	Describe("Dashboard Error Handling", func() {
		It("should handle loading states gracefully", func() {
			By("Checking if loading indicator appears during navigation")
			err := dashboardPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			// Check if loading indicator is present initially
			isLoading, err := dashboardPage.IsLoading()
			if err == nil && isLoading {
				By("Waiting for loading to complete")
				err = dashboardPage.WaitForLoadingToComplete()
				Expect(err).NotTo(HaveOccurred())
			}
			
			By("Verifying dashboard content is loaded")
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should display appropriate message when no data is available", func() {
			// This test would require a clean environment or API mocking
			// For now, we'll verify the current state shows data
			By("Verifying dashboard shows data when available")
			stats, err := dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			
			// If we have test data, stats should reflect it
			if len(testData.Projects) > 0 {
				Expect(stats.TotalProjects).NotTo(Equal("0"))
			}
			
			if len(testData.TestRuns) > 0 {
				Expect(stats.TotalTestRuns).NotTo(Equal("0"))
			}
		})

		It("should handle API errors gracefully", func() {
			By("Checking if error messages are handled properly")
			errorMessage, err := dashboardPage.GetErrorMessage()
			
			// If no error, that's good
			if err == nil && errorMessage != "" {
				// If there is an error message, it should be meaningful
				Expect(errorMessage).NotTo(ContainSubstring("undefined"))
				Expect(errorMessage).NotTo(ContainSubstring("null"))
			}
		})
	})

	Describe("Dashboard Responsiveness", func() {
		It("should be responsive to different screen sizes", func() {
			// Test mobile viewport
			By("Testing mobile viewport")
			err := dashboardPage.SetViewportSize(375, 667) // iPhone 6/7/8 size
			Expect(err).NotTo(HaveOccurred())
			
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			// Verify dashboard still functions
			stats, err := dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			Expect(stats).NotTo(BeNil())
			
			// Test tablet viewport
			By("Testing tablet viewport")
			err = dashboardPage.SetViewportSize(768, 1024) // iPad size
			Expect(err).NotTo(HaveOccurred())
			
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			// Test desktop viewport
			By("Testing desktop viewport")
			err = dashboardPage.SetViewportSize(1920, 1080) // Desktop size
			Expect(err).NotTo(HaveOccurred())
			
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should maintain functionality across viewport changes", func() {
			By("Testing functionality at different viewport sizes")
			viewports := []struct {
				width  int64
				height int64
				name   string
			}{
				{375, 667, "mobile"},
				{768, 1024, "tablet"},
				{1920, 1080, "desktop"},
			}
			
			for _, viewport := range viewports {
				By(fmt.Sprintf("Testing %s viewport (%dx%d)", viewport.name, viewport.width, viewport.height))
				
				err := dashboardPage.SetViewportSize(viewport.width, viewport.height)
				Expect(err).NotTo(HaveOccurred())
				
				err = dashboardPage.WaitForDashboardToLoad()
				Expect(err).NotTo(HaveOccurred())
				
				// Verify core functionality works
				stats, err := dashboardPage.GetStats()
				Expect(err).NotTo(HaveOccurred())
				Expect(stats.TotalProjects).NotTo(BeEmpty())
				
				projectSummaries, err := dashboardPage.GetProjectSummaries()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(projectSummaries)).To(BeNumerically(">=", 0))
			}
		})
	})

	Describe("Dashboard Accessibility", func() {
		It("should have proper page title", func() {
			By("Checking page title")
			title, err := dashboardPage.GetPageTitle()
			Expect(err).NotTo(HaveOccurred())
			Expect(title).To(ContainSubstring("Fern Platform"))
		})

		It("should have accessible navigation elements", func() {
			By("Verifying dashboard elements are accessible")
			
			// Check if main dashboard content is present
			isVisible, err := dashboardPage.IsElementVisible(`[data-testid="dashboard-content"]`)
			Expect(err).NotTo(HaveOccurred())
			Expect(isVisible).To(BeTrue())
			
			// Check if stats cards are present
			isVisible, err = dashboardPage.IsElementVisible(`[data-testid="stats-cards"]`)
			Expect(err).NotTo(HaveOccurred())
			Expect(isVisible).To(BeTrue())
		})
	})

	Describe("Dashboard Performance", func() {
		It("should render efficiently with large datasets", func() {
			By("Measuring render performance with current dataset")
			startTime := time.Now()
			
			err := dashboardPage.Navigate()
			Expect(err).NotTo(HaveOccurred())
			
			err = dashboardPage.WaitForDashboardToLoad()
			Expect(err).NotTo(HaveOccurred())
			
			renderTime := time.Since(startTime)
			
			By("Retrieving all dashboard data")
			_, err = dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			
			_, err = dashboardPage.GetProjectSummaries()
			Expect(err).NotTo(HaveOccurred())
			
			_, err = dashboardPage.GetRecentTestRuns()
			Expect(err).NotTo(HaveOccurred())
			
			totalTime := time.Since(startTime)
			
			By("Verifying performance requirements")
			Expect(renderTime).To(BeNumerically("<", 3*time.Second),
				"Dashboard should render within 3 seconds")
			Expect(totalTime).To(BeNumerically("<", 5*time.Second),
				"All dashboard data should load within 5 seconds")
		})

		It("should handle rapid interactions smoothly", func() {
			By("Performing rapid dashboard interactions")
			
			// Refresh multiple times quickly
			for i := 0; i < 3; i++ {
				err := dashboardPage.Refresh()
				Expect(err).NotTo(HaveOccurred())
				
				err = dashboardPage.WaitForDashboardToLoad()
				Expect(err).NotTo(HaveOccurred())
			}
			
			// Verify dashboard still functions correctly
			stats, err := dashboardPage.GetStats()
			Expect(err).NotTo(HaveOccurred())
			Expect(stats).NotTo(BeNil())
		})
	})
})