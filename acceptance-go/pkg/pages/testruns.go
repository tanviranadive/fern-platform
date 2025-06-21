package pages

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	. "github.com/onsi/ginkgo/v2"
)

// TestRunsPage represents the test runs page
type TestRunsPage struct {
	*BasePage
}

// NewTestRunsPage creates a new test runs page
func NewTestRunsPage(baseURL string, browserCtx context.Context) *TestRunsPage {
	return &TestRunsPage{
		BasePage: NewBasePage(baseURL, browserCtx),
	}
}

// Navigate navigates to the test runs page
func (p *TestRunsPage) Navigate() error {
	GinkgoHelper()
	By("Navigating to test runs page")

	url := p.baseURL + "/test-runs"
	return chromedp.Run(p.ctx, chromedp.Navigate(url))
}

// NavigateToTestRun navigates to a specific test run
func (p *TestRunsPage) NavigateToTestRun(testRunID string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Navigating to test run: %s", testRunID))

	url := fmt.Sprintf("%s/test-runs?id=%s", p.baseURL, testRunID)
	return chromedp.Run(p.ctx, chromedp.Navigate(url))
}

// WaitForTestRunsToLoad waits for test runs to load
func (p *TestRunsPage) WaitForTestRunsToLoad() error {
	GinkgoHelper()
	By("Waiting for test runs to load")

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(`[data-testid="test-runs-table"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="test-run-row"]`, chromedp.ByQuery),
	)
}

// WaitForEmptyState waits for empty state to appear
func (p *TestRunsPage) WaitForEmptyState() error {
	GinkgoHelper()
	By("Waiting for empty state")

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(`[data-testid="empty-state"]`, chromedp.ByQuery),
	)
}

// GetEmptyStateMessage retrieves the empty state message
func (p *TestRunsPage) GetEmptyStateMessage() (string, error) {
	GinkgoHelper()
	By("Getting empty state message")

	return p.GetText(`[data-testid="empty-state-message"]`)
}

// GetVisibleTestRuns retrieves all visible test runs
func (p *TestRunsPage) GetVisibleTestRuns() ([]TestRunRow, error) {
	GinkgoHelper()
	By("Getting visible test runs")

	var testRuns []TestRunRow

	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[data-testid="test-run-row"]')).map(row => ({
				id: row.getAttribute('data-test-run-id'),
				projectName: row.querySelector('[data-testid="project-name"]').textContent,
				status: row.querySelector('[data-testid="status-badge"]').getAttribute('data-status'),
				duration: row.querySelector('[data-testid="duration"]').textContent,
				startTime: row.querySelector('[data-testid="start-time"]').textContent,
				branch: row.querySelector('[data-testid="branch"]').textContent,
				description: row.querySelector('[data-testid="description"]')?.textContent || ''
			}))
		`, &testRuns),
	)

	return testRuns, err
}

// GetTestRunByIndex retrieves a test run by its index
func (p *TestRunsPage) GetTestRunByIndex(index int) (*TestRunRow, error) {
	GinkgoHelper()
	By(fmt.Sprintf("Getting test run at index: %d", index))

	var testRun TestRunRow

	script := fmt.Sprintf(`
		(() => {
			const rows = document.querySelectorAll('[data-testid="test-run-row"]');
			if (rows.length > %d) {
				const row = rows[%d];
				return {
					id: row.getAttribute('data-test-run-id'),
					projectName: row.querySelector('[data-testid="project-name"]').textContent,
					status: row.querySelector('[data-testid="status-badge"]').getAttribute('data-status'),
					duration: row.querySelector('[data-testid="duration"]').textContent,
					startTime: row.querySelector('[data-testid="start-time"]').textContent,
					branch: row.querySelector('[data-testid="branch"]').textContent,
					description: row.querySelector('[data-testid="description"]')?.textContent || ''
				};
			}
			return null;
		})()
	`, index, index)

	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(script, &testRun),
	)

	return &testRun, err
}

// FilterByProject filters test runs by project
func (p *TestRunsPage) FilterByProject(projectName string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Filtering by project: %s", projectName))

	return chromedp.Run(p.ctx,
		chromedp.Click(`[data-testid="project-filter"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="project-filter-dropdown"]`, chromedp.ByQuery),
		chromedp.Click(fmt.Sprintf(`[data-testid="project-option"][data-value="%s"]`, projectName), chromedp.ByQuery),
	)
}

// FilterByStatus filters test runs by status
func (p *TestRunsPage) FilterByStatus(status string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Filtering by status: %s", status))

	return chromedp.Run(p.ctx,
		chromedp.Click(`[data-testid="status-filter"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="status-filter-dropdown"]`, chromedp.ByQuery),
		chromedp.Click(fmt.Sprintf(`[data-testid="status-option"][data-value="%s"]`, status), chromedp.ByQuery),
	)
}

// FilterByBranch filters test runs by branch
func (p *TestRunsPage) FilterByBranch(branch string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Filtering by branch: %s", branch))

	return chromedp.Run(p.ctx,
		chromedp.Click(`[data-testid="branch-filter"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="branch-filter-dropdown"]`, chromedp.ByQuery),
		chromedp.Click(fmt.Sprintf(`[data-testid="branch-option"][data-value="%s"]`, branch), chromedp.ByQuery),
	)
}

// SearchTests searches for test runs
func (p *TestRunsPage) SearchTests(searchTerm string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Searching for: %s", searchTerm))

	return p.TypeText(`[data-testid="search-input"]`, searchTerm)
}

// ClearAllFilters clears all applied filters
func (p *TestRunsPage) ClearAllFilters() error {
	GinkgoHelper()
	By("Clearing all filters")

	return p.ClickElement(`[data-testid="clear-filters-button"]`)
}

// WaitForFilterResults waits for filter results to load
func (p *TestRunsPage) WaitForFilterResults() error {
	GinkgoHelper()
	By("Waiting for filter results")

	return chromedp.Run(p.ctx,
		chromedp.WaitNotPresent(`[data-testid="loading-spinner"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Small delay for UI to update
	)
}

// WaitForSearchResults waits for search results to load
func (p *TestRunsPage) WaitForSearchResults() error {
	GinkgoHelper()
	By("Waiting for search results")

	return p.WaitForFilterResults()
}

// GetAvailableProjects retrieves available projects for filtering
func (p *TestRunsPage) GetAvailableProjects() ([]string, error) {
	GinkgoHelper()
	By("Getting available projects")

	var projects []string

	err := chromedp.Run(p.ctx,
		chromedp.Click(`[data-testid="project-filter"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="project-filter-dropdown"]`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[data-testid="project-option"]')).map(option => option.getAttribute('data-value'))
		`, &projects),
		chromedp.Click(`[data-testid="project-filter"]`, chromedp.ByQuery), // Close dropdown
	)

	return projects, err
}

// GetAvailableBranches retrieves available branches for filtering
func (p *TestRunsPage) GetAvailableBranches() ([]string, error) {
	GinkgoHelper()
	By("Getting available branches")

	var branches []string

	err := chromedp.Run(p.ctx,
		chromedp.Click(`[data-testid="branch-filter"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="branch-filter-dropdown"]`, chromedp.ByQuery),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('[data-testid="branch-option"]')).map(option => option.getAttribute('data-value'))
		`, &branches),
		chromedp.Click(`[data-testid="branch-filter"]`, chromedp.ByQuery), // Close dropdown
	)

	return branches, err
}

// ExpandTestRun expands a test run to show spec details
func (p *TestRunsPage) ExpandTestRun(testRunID string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Expanding test run: %s", testRunID))

	selector := fmt.Sprintf(`[data-test-run-id="%s"] [data-testid="expand-button"]`, testRunID)
	return p.ClickElement(selector)
}

// CollapseTestRun collapses an expanded test run
func (p *TestRunsPage) CollapseTestRun(testRunID string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Collapsing test run: %s", testRunID))

	selector := fmt.Sprintf(`[data-test-run-id="%s"] [data-testid="collapse-button"]`, testRunID)
	return p.ClickElement(selector)
}

// IsTestRunExpanded checks if a test run is expanded
func (p *TestRunsPage) IsTestRunExpanded(testRunID string) (bool, error) {
	GinkgoHelper()
	By(fmt.Sprintf("Checking if test run is expanded: %s", testRunID))

	selector := fmt.Sprintf(`[data-test-run-id="%s"][data-expanded="true"]`, testRunID)
	return p.IsElementVisible(selector)
}

// WaitForSpecDetailsToLoad waits for spec details to load
func (p *TestRunsPage) WaitForSpecDetailsToLoad(testRunID string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Waiting for spec details to load for test run: %s", testRunID))

	selector := fmt.Sprintf(`[data-test-run-id="%s"] [data-testid="spec-runs"]`, testRunID)
	return p.WaitForElement(selector)
}

// GetSpecRuns retrieves spec runs for a test run
func (p *TestRunsPage) GetSpecRuns(testRunID string) ([]SpecRunRow, error) {
	GinkgoHelper()
	By(fmt.Sprintf("Getting spec runs for test run: %s", testRunID))

	var specRuns []SpecRunRow

	script := fmt.Sprintf(`
		Array.from(document.querySelectorAll('[data-test-run-id="%s"] [data-testid="spec-run-row"]')).map(row => ({
			id: row.getAttribute('data-spec-run-id'),
			description: row.querySelector('[data-testid="spec-description"]').textContent,
			status: row.querySelector('[data-testid="spec-status"]').getAttribute('data-status'),
			duration: row.querySelector('[data-testid="spec-duration"]').textContent,
			errorMessage: row.querySelector('[data-testid="error-message"]')?.textContent || ''
		}))
	`, testRunID)

	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(script, &specRuns),
	)

	return specRuns, err
}

// GetFailedSpecs retrieves only failed spec runs for a test run
func (p *TestRunsPage) GetFailedSpecs(testRunID string) ([]SpecRunRow, error) {
	GinkgoHelper()
	By(fmt.Sprintf("Getting failed specs for test run: %s", testRunID))

	var failedSpecs []SpecRunRow

	script := fmt.Sprintf(`
		Array.from(document.querySelectorAll('[data-test-run-id="%s"] [data-testid="spec-run-row"][data-status="failed"]')).map(row => ({
			id: row.getAttribute('data-spec-run-id'),
			description: row.querySelector('[data-testid="spec-description"]').textContent,
			status: row.querySelector('[data-testid="spec-status"]').getAttribute('data-status'),
			duration: row.querySelector('[data-testid="spec-duration"]').textContent,
			errorMessage: row.querySelector('[data-testid="error-message"]')?.textContent || ''
		}))
	`, testRunID)

	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(script, &failedSpecs),
	)

	return failedSpecs, err
}

// Pagination methods
func (p *TestRunsPage) GetPaginationInfo() (*PaginationInfo, error) {
	GinkgoHelper()
	By("Getting pagination information")

	var info PaginationInfo

	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(`
			(() => {
				const pagination = document.querySelector('[data-testid="pagination"]');
				if (!pagination) return null;
				
				return {
					currentPage: parseInt(pagination.getAttribute('data-current-page')),
					totalPages: parseInt(pagination.getAttribute('data-total-pages')),
					totalItems: parseInt(pagination.getAttribute('data-total-items'))
				};
			})()
		`, &info),
	)

	return &info, err
}

// GoToNextPage navigates to the next page
func (p *TestRunsPage) GoToNextPage() error {
	GinkgoHelper()
	By("Going to next page")

	return p.ClickElement(`[data-testid="next-page-button"]`)
}

// GoToPreviousPage navigates to the previous page
func (p *TestRunsPage) GoToPreviousPage() error {
	GinkgoHelper()
	By("Going to previous page")

	return p.ClickElement(`[data-testid="previous-page-button"]`)
}

// WaitForPageChange waits for page change to complete
func (p *TestRunsPage) WaitForPageChange() error {
	GinkgoHelper()
	By("Waiting for page change")

	return p.WaitForFilterResults()
}

// ChangePageSize changes the page size
func (p *TestRunsPage) ChangePageSize(pageSize int) error {
	GinkgoHelper()
	By(fmt.Sprintf("Changing page size to: %d", pageSize))

	return chromedp.Run(p.ctx,
		chromedp.Click(`[data-testid="page-size-selector"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`[data-testid="page-size-dropdown"]`, chromedp.ByQuery),
		chromedp.Click(fmt.Sprintf(`[data-testid="page-size-option"][data-value="%d"]`, pageSize), chromedp.ByQuery),
	)
}

// WaitForPageSizeChange waits for page size change to complete
func (p *TestRunsPage) WaitForPageSizeChange() error {
	GinkgoHelper()
	By("Waiting for page size change")

	return p.WaitForFilterResults()
}

// GetTotalItemCount retrieves the total number of items
func (p *TestRunsPage) GetTotalItemCount() (int, error) {
	GinkgoHelper()
	By("Getting total item count")

	paginationInfo, err := p.GetPaginationInfo()
	if err != nil {
		return 0, err
	}

	return paginationInfo.TotalItems, nil
}

// GetHighlightedTestRun retrieves the currently highlighted test run
func (p *TestRunsPage) GetHighlightedTestRun() (*TestRunRow, error) {
	GinkgoHelper()
	By("Getting highlighted test run")

	var testRun TestRunRow

	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(`
			(() => {
				const highlightedRow = document.querySelector('[data-testid="test-run-row"][data-highlighted="true"]');
				if (!highlightedRow) return null;
				
				return {
					id: highlightedRow.getAttribute('data-test-run-id'),
					projectName: highlightedRow.querySelector('[data-testid="project-name"]').textContent,
					status: highlightedRow.querySelector('[data-testid="status-badge"]').getAttribute('data-status'),
					duration: highlightedRow.querySelector('[data-testid="duration"]').textContent,
					startTime: highlightedRow.querySelector('[data-testid="start-time"]').textContent,
					branch: highlightedRow.querySelector('[data-testid="branch"]').textContent
				};
			})()
		`, &testRun),
	)

	return &testRun, err
}

// Filter state methods
func (p *TestRunsPage) getCurrentStatusFilter() (string, error) {
	return p.GetAttribute(`[data-testid="status-filter"]`, "data-selected-value")
}

func (p *TestRunsPage) getCurrentBranchFilter() (string, error) {
	return p.GetAttribute(`[data-testid="branch-filter"]`, "data-selected-value")
}

func (p *TestRunsPage) GetCurrentStatusFilter() (string, error) {
	return p.getCurrentStatusFilter()
}

func (p *TestRunsPage) GetCurrentBranchFilter() (string, error) {
	return p.getCurrentBranchFilter()
}

// GetErrorMessage retrieves error message if any
func (p *TestRunsPage) GetErrorMessage() (string, error) {
	GinkgoHelper()
	By("Getting error message")

	return p.GetText(`[data-testid="error-message"]`)
}

// GetRetryButton retrieves the retry button element
func (p *TestRunsPage) GetRetryButton() (bool, error) {
	GinkgoHelper()
	By("Getting retry button")

	return p.IsElementVisible(`[data-testid="retry-button"]`)
}

// IsLoading checks if the page is in loading state
func (p *TestRunsPage) IsLoading() (bool, error) {
	GinkgoHelper()
	By("Checking if page is loading")

	return p.IsElementVisible(`[data-testid="loading-spinner"]`)
}

// Data structures
type TestRunRow struct {
	ID          string `json:"id"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	Duration    string `json:"duration"`
	StartTime   string `json:"startTime"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
}

type SpecRunRow struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	Duration     string `json:"duration"`
	ErrorMessage string `json:"errorMessage"`
}

type PaginationInfo struct {
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
	TotalItems  int `json:"totalItems"`
}