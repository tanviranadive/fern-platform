package pages

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	. "github.com/onsi/ginkgo/v2"
)

// BasePage provides common functionality for all page objects
type BasePage struct {
	baseURL string
	ctx     context.Context
}

// NewBasePage creates a new base page
func NewBasePage(baseURL string, browserCtx context.Context) *BasePage {
	return &BasePage{
		baseURL: baseURL,
		ctx:     browserCtx,
	}
}

// WaitForElement waits for an element to be visible
func (p *BasePage) WaitForElement(selector string, timeout ...time.Duration) error {
	GinkgoHelper()
	
	to := 30 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}

	ctx, cancel := context.WithTimeout(p.ctx, to)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Sleep(100*time.Millisecond), // Small delay to ensure element is ready
	)
}

// WaitForElementWithTimeout waits for an element with a specific timeout
func (p *BasePage) WaitForElementWithTimeout(selector string, timeout time.Duration) error {
	GinkgoHelper()

	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
	)
}

// ClickElement clicks on an element
func (p *BasePage) ClickElement(selector string) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Click(selector, chromedp.ByQuery),
	)
}

// TypeText types text into an input field
func (p *BasePage) TypeText(selector, text string) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Clear(selector, chromedp.ByQuery),
		chromedp.SendKeys(selector, text, chromedp.ByQuery),
	)
}

// GetText retrieves text from an element
func (p *BasePage) GetText(selector string) (string, error) {
	GinkgoHelper()

	var text string
	err := chromedp.Run(p.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	return text, err
}

// GetAttribute retrieves an attribute from an element
func (p *BasePage) GetAttribute(selector, attribute string) (string, error) {
	GinkgoHelper()

	var value string
	err := chromedp.Run(p.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.AttributeValue(selector, attribute, &value, nil, chromedp.ByQuery),
	)
	return value, err
}

// IsElementVisible checks if an element is visible
func (p *BasePage) IsElementVisible(selector string) (bool, error) {
	GinkgoHelper()

	var visible bool
	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			(() => {
				const element = document.querySelector('%s');
				return element !== null && element.offsetParent !== null;
			})()
		`, selector), &visible),
	)
	return visible, err
}

// IsElementPresent checks if an element exists in the DOM
func (p *BasePage) IsElementPresent(selector string) (bool, error) {
	GinkgoHelper()

	var present bool
	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			document.querySelector('%s') !== null
		`, selector), &present),
	)
	return present, err
}

// WaitForPageLoad waits for the page to finish loading
func (p *BasePage) WaitForPageLoad() error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`document.readyState === 'complete'`, nil),
	)
}

// ScrollToElement scrolls an element into view
func (p *BasePage) ScrollToElement(selector string) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Evaluate(fmt.Sprintf(`
			document.querySelector('%s').scrollIntoView({behavior: 'smooth', block: 'center'})
		`, selector), nil),
		chromedp.Sleep(500*time.Millisecond), // Wait for scroll animation
	)
}

// GetCurrentURL retrieves the current page URL
func (p *BasePage) GetCurrentURL() (string, error) {
	GinkgoHelper()

	var url string
	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(`window.location.href`, &url),
	)
	return url, err
}

// Refresh refreshes the current page
func (p *BasePage) Refresh() error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.Reload(),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
}

// TakeScreenshot takes a screenshot of the current page
func (p *BasePage) TakeScreenshot() ([]byte, error) {
	GinkgoHelper()

	var screenshot []byte
	err := chromedp.Run(p.ctx,
		chromedp.CaptureScreenshot(&screenshot),
	)
	return screenshot, err
}

// WaitForElementToDisappear waits for an element to disappear
func (p *BasePage) WaitForElementToDisappear(selector string, timeout ...time.Duration) error {
	GinkgoHelper()

	to := 30 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}

	ctx, cancel := context.WithTimeout(p.ctx, to)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitNotPresent(selector, chromedp.ByQuery),
	)
}

// SelectDropdownOption selects an option from a dropdown
func (p *BasePage) SelectDropdownOption(dropdownSelector, optionValue string) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(dropdownSelector, chromedp.ByQuery),
		chromedp.Click(dropdownSelector, chromedp.ByQuery),
		chromedp.WaitVisible(fmt.Sprintf(`%s option[value="%s"]`, dropdownSelector, optionValue), chromedp.ByQuery),
		chromedp.SetValue(dropdownSelector, optionValue, chromedp.ByQuery),
	)
}

// WaitForAjaxToComplete waits for AJAX requests to complete
func (p *BasePage) WaitForAjaxToComplete() error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.Evaluate(`
			new Promise((resolve) => {
				if (typeof jQuery !== 'undefined' && jQuery.active === 0) {
					resolve();
				} else {
					// Fallback: wait for network to be idle
					setTimeout(resolve, 1000);
				}
			})
		`, nil),
	)
}

// ExecuteScript executes custom JavaScript
func (p *BasePage) ExecuteScript(script string) (interface{}, error) {
	GinkgoHelper()

	var result interface{}
	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(script, &result),
	)
	return result, err
}

// WaitForCondition waits for a custom JavaScript condition to be true
func (p *BasePage) WaitForCondition(condition string, timeout time.Duration) error {
	GinkgoHelper()

	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	script := fmt.Sprintf(`
		new Promise((resolve, reject) => {
			const checkCondition = () => {
				if (%s) {
					resolve(true);
				} else {
					setTimeout(checkCondition, 100);
				}
			};
			checkCondition();
		})
	`, condition)

	return chromedp.Run(ctx,
		chromedp.Evaluate(script, nil),
	)
}

// GetElementCount returns the number of elements matching the selector
func (p *BasePage) GetElementCount(selector string) (int, error) {
	GinkgoHelper()

	var count int
	err := chromedp.Run(p.ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.querySelectorAll('%s').length`, selector), &count),
	)
	return count, err
}

// HighlightElement highlights an element (useful for debugging)
func (p *BasePage) HighlightElement(selector string) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Evaluate(fmt.Sprintf(`
			(() => {
				const element = document.querySelector('%s');
				if (element) {
					element.style.border = '3px solid red';
					element.style.backgroundColor = 'yellow';
				}
			})()
		`, selector), nil),
		chromedp.Sleep(1*time.Second),
	)
}

// RemoveHighlight removes highlighting from an element
func (p *BasePage) RemoveHighlight(selector string) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			(() => {
				const element = document.querySelector('%s');
				if (element) {
					element.style.border = '';
					element.style.backgroundColor = '';
				}
			})()
		`, selector), nil),
	)
}

// SetViewportSize sets the browser viewport size
func (p *BasePage) SetViewportSize(width, height int64) error {
	GinkgoHelper()

	return chromedp.Run(p.ctx,
		chromedp.EmulateViewport(width, height),
	)
}

// GetPageTitle retrieves the page title
func (p *BasePage) GetPageTitle() (string, error) {
	GinkgoHelper()

	var title string
	err := chromedp.Run(p.ctx,
		chromedp.Title(&title),
	)
	return title, err
}

// WaitForTextToAppear waits for specific text to appear in an element
func (p *BasePage) WaitForTextToAppear(selector, expectedText string, timeout time.Duration) error {
	GinkgoHelper()

	condition := fmt.Sprintf(`
		(() => {
			const element = document.querySelector('%s');
			return element && element.textContent.includes('%s');
		})()
	`, selector, expectedText)

	return p.WaitForCondition(condition, timeout)
}