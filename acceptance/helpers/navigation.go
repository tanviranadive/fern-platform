package helpers

import (
	. "github.com/onsi/gomega"
	"github.com/playwright-community/playwright-go"
)

// NavigationHelper helps with page navigation
type NavigationHelper struct {
	Page    playwright.Page
	BaseURL string
}

// NewNavigationHelper creates a new navigation helper
func NewNavigationHelper(page playwright.Page, baseURL string) *NavigationHelper {
	return &NavigationHelper{
		Page:    page,
		BaseURL: baseURL,
	}
}

// GoToTestSummaries navigates to the Test Summaries page
func (n *NavigationHelper) GoToTestSummaries() {
	// Try clicking nav button first (if already logged in)
	navButton := n.Page.Locator("button.nav-button:has-text('Test Summaries')")
	count, _ := navButton.Count()
	
	if count > 0 {
		err := navButton.Click()
		Expect(err).NotTo(HaveOccurred())
	} else {
		// Fallback to direct navigation
		_, err := n.Page.Goto(n.BaseURL + "/test-summaries")
		Expect(err).NotTo(HaveOccurred())
	}
	
	// Wait for page to load
	err := n.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	Expect(err).NotTo(HaveOccurred())
}

// GoToTestRuns navigates to the Test Runs page
func (n *NavigationHelper) GoToTestRuns() {
	_, err := n.Page.Goto(n.BaseURL + "/test-runs")
	Expect(err).NotTo(HaveOccurred())
	
	// Wait for page to load
	err = n.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	Expect(err).NotTo(HaveOccurred())
}

// GoToProjects navigates to the Projects page
func (n *NavigationHelper) GoToProjects() {
	_, err := n.Page.Goto(n.BaseURL + "/projects")
	Expect(err).NotTo(HaveOccurred())
	
	// Wait for page to load
	err = n.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	Expect(err).NotTo(HaveOccurred())
}

// ClickBreadcrumb clicks on a breadcrumb link
func (n *NavigationHelper) ClickBreadcrumb(text string) {
	breadcrumbs := n.Page.Locator("nav[aria-label='breadcrumb'], .breadcrumbs")
	err := breadcrumbs.Locator("text=" + text).Click()
	Expect(err).NotTo(HaveOccurred())
	
	// Wait for navigation
	err = n.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	Expect(err).NotTo(HaveOccurred())
}

// GetCurrentBreadcrumbs returns the current breadcrumb path
func (n *NavigationHelper) GetCurrentBreadcrumbs() []string {
	breadcrumbs := n.Page.Locator("nav[aria-label='breadcrumb'], .breadcrumbs").Locator("a, span")
	count, err := breadcrumbs.Count()
	Expect(err).NotTo(HaveOccurred())
	
	var path []string
	for i := 0; i < count; i++ {
		text, err := breadcrumbs.Nth(i).TextContent()
		Expect(err).NotTo(HaveOccurred())
		path = append(path, text)
	}
	
	return path
}