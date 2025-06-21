package pages

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	. "github.com/onsi/ginkgo/v2"
)

// DashboardPage represents the dashboard page
type DashboardPage struct {
	*BasePage
}

// NewDashboardPage creates a new dashboard page
func NewDashboardPage(baseURL string, browserCtx context.Context) *DashboardPage {
	return &DashboardPage{
		BasePage: NewBasePage(baseURL, browserCtx),
	}
}

// Navigate navigates to the dashboard page
func (p *DashboardPage) Navigate() error {
	GinkgoHelper()
	By("Navigating to dashboard page")

	url := p.baseURL + "/"
	return chromedp.Run(p.ctx, chromedp.Navigate(url))
}

// WaitForDashboardToLoad waits for the dashboard to fully load
func (p *DashboardPage) WaitForDashboardToLoad() error {
	GinkgoHelper()
	By("Waiting for dashboard to load")

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(`[data-testid="dashboard-content"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="stats-cards"]`, chromedp.ByQuery),
	)
}

// GetStats retrieves the dashboard statistics
func (p *DashboardPage) GetStats() (*DashboardStats, error) {
	GinkgoHelper()
	By("Getting dashboard statistics")

	var stats DashboardStats

	err := chromedp.Run(p.ctx,
		chromedp.Text(`[data-testid="total-test-runs"] .stat-value`, &stats.TotalTestRuns, chromedp.ByQuery),
		chromedp.Text(`[data-testid="total-projects"] .stat-value`, &stats.TotalProjects, chromedp.ByQuery),
		chromedp.Text(`[data-testid="success-rate"] .stat-value`, &stats.SuccessRate, chromedp.ByQuery),
		chromedp.Text(`[data-testid="avg-duration"] .stat-value`, &stats.AvgDuration, chromedp.ByQuery),
	)

	return &stats, err
}

// GetRecentTestRuns retrieves the recent test runs from the dashboard
func (p *DashboardPage) GetRecentTestRuns() ([]DashboardTestRun, error) {
	GinkgoHelper()
	By("Getting recent test runs from dashboard")

	var runs []DashboardTestRun

	err := chromedp.Run(p.ctx,
		chromedp.WaitVisible(`[data-testid="recent-test-runs"]`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[data-testid="recent-test-runs"] .test-run-item')).map(item => ({
				id: item.getAttribute('data-test-run-id'),
				projectName: item.querySelector('.project-name').textContent,
				status: item.querySelector('.status-badge').getAttribute('data-status'),
				duration: item.querySelector('.duration').textContent,
				startTime: item.querySelector('.start-time').textContent
			}))
		`, &runs),
	)

	return runs, err
}

// GetProjectSummaries retrieves project summaries from the dashboard
func (p *DashboardPage) GetProjectSummaries() ([]ProjectSummary, error) {
	GinkgoHelper()
	By("Getting project summaries from dashboard")

	var summaries []ProjectSummary

	err := chromedp.Run(p.ctx,
		chromedp.WaitVisible(`[data-testid="project-summaries"]`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[data-testid="project-summaries"] .project-card')).map(card => ({
				id: card.getAttribute('data-project-id'),
				name: card.querySelector('.project-name').textContent,
				testRunsCount: parseInt(card.querySelector('.test-runs-count').textContent),
				successRate: parseFloat(card.querySelector('.success-rate').textContent),
				lastRunTime: card.querySelector('.last-run-time').textContent
			}))
		`, &summaries),
	)

	return summaries, err
}

// ClickProject clicks on a project in the dashboard
func (p *DashboardPage) ClickProject(projectID string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Clicking on project: %s", projectID))

	selector := fmt.Sprintf(`[data-project-id="%s"]`, projectID)
	return chromedp.Run(p.ctx,
		chromedp.Click(selector, chromedp.ByQuery),
	)
}

// WaitForNavigation waits for navigation to complete
func (p *DashboardPage) WaitForNavigation() error {
	GinkgoHelper()
	By("Waiting for navigation to complete")

	return chromedp.Run(p.ctx,
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Small delay to ensure navigation is complete
	)
}

// GetErrorMessage retrieves error message if any
func (p *DashboardPage) GetErrorMessage() (string, error) {
	GinkgoHelper()
	By("Getting error message")

	var errorMessage string
	err := chromedp.Run(p.ctx,
		chromedp.Text(`[data-testid="error-message"]`, &errorMessage, chromedp.ByQuery),
	)

	// If element not found, return empty string
	if err != nil {
		return "", nil
	}

	return errorMessage, nil
}

// IsLoading checks if the dashboard is in loading state
func (p *DashboardPage) IsLoading() (bool, error) {
	GinkgoHelper()
	By("Checking if dashboard is loading")

	var isVisible bool
	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(`
			document.querySelector('[data-testid="loading-spinner"]') !== null &&
			document.querySelector('[data-testid="loading-spinner"]').offsetParent !== null
		`, &isVisible),
	)

	return isVisible, err
}

// WaitForLoadingToComplete waits for loading to complete
func (p *DashboardPage) WaitForLoadingToComplete() error {
	GinkgoHelper()
	By("Waiting for loading to complete")

	return chromedp.Run(p.ctx,
		chromedp.WaitNotPresent(`[data-testid="loading-spinner"]`, chromedp.ByQuery),
	)
}

// GetTestHistory retrieves test history visualization data
func (p *DashboardPage) GetTestHistory() ([]TestHistoryItem, error) {
	GinkgoHelper()
	By("Getting test history visualization")

	var history []TestHistoryItem

	err := chromedp.Run(p.ctx,
		chromedp.WaitVisible(`[data-testid="test-history"]`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[data-testid="test-history"] .history-item')).map(item => ({
				date: item.getAttribute('data-date'),
				status: item.getAttribute('data-status'),
				count: parseInt(item.getAttribute('data-count'))
			}))
		`, &history),
	)

	return history, err
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalTestRuns string `json:"totalTestRuns"`
	TotalProjects string `json:"totalProjects"`
	SuccessRate   string `json:"successRate"`
	AvgDuration   string `json:"avgDuration"`
}

// DashboardTestRun represents a test run shown on the dashboard
type DashboardTestRun struct {
	ID          string `json:"id"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	Duration    string `json:"duration"`
	StartTime   string `json:"startTime"`
}

// ProjectSummary represents a project summary on the dashboard
type ProjectSummary struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	TestRunsCount int     `json:"testRunsCount"`
	SuccessRate   float64 `json:"successRate"`
	LastRunTime   string  `json:"lastRunTime"`
}

// TestHistoryItem represents an item in the test history visualization
type TestHistoryItem struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Count  int    `json:"count"`
}