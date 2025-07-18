package helpers

import (
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/playwright-community/playwright-go"
)

// LoginHelper handles authentication flows
type LoginHelper struct {
	Page     playwright.Page
	BaseURL  string
	Username string
	Password string
}

// NewLoginHelper creates a new login helper
func NewLoginHelper(page playwright.Page, baseURL, username, password string) *LoginHelper {
	return &LoginHelper{
		Page:     page,
		BaseURL:  baseURL,
		Username: username,
		Password: password,
	}
}

// Login performs the OAuth login flow
func (l *LoginHelper) Login() {
	// Navigate to the platform
	_, err := l.Page.Goto(l.BaseURL)
	Expect(err).NotTo(HaveOccurred())

	// Should be redirected to login page
	Eventually(func() string {
		return l.Page.URL()
	}, 5*time.Second).Should(Equal(l.BaseURL + "/auth/login"))

	// Click "Sign in with OAuth" button or link
	signInElement := l.Page.Locator("button:has-text('Sign in'), a:has-text('Sign in')")
	err = signInElement.Click()
	Expect(err).NotTo(HaveOccurred())

	// Should be redirected to Keycloak
	Eventually(func() bool {
		url := l.Page.URL()
		return strings.Contains(url, "keycloak") && strings.Contains(url, "/realms/fern-platform/protocol/openid-connect/auth")
	}, 10*time.Second).Should(BeTrue(), "Should be redirected to Keycloak")

	// Fill in Keycloak credentials
	// Keycloak might use different field names/IDs
	usernameField := l.Page.Locator("input#username, input[name='username']")
	passwordField := l.Page.Locator("input#password, input[name='password']")

	err = usernameField.Fill(l.Username)
	Expect(err).NotTo(HaveOccurred())

	err = passwordField.Fill(l.Password)
	Expect(err).NotTo(HaveOccurred())

	// Click sign in on Keycloak
	signInButton := l.Page.Locator("input[type='submit'], button[type='submit'], input[value='Sign In'], button:has-text('Sign In')")
	err = signInButton.Click()
	Expect(err).NotTo(HaveOccurred())

	// Wait for redirect back to platform (through /auth/callback)
	Eventually(func() bool {
		url := l.Page.URL()
		// Either we're on the main page or still going through callback
		return url == l.BaseURL+"/" || strings.Contains(url, "/auth/callback")
	}, 30*time.Second).Should(BeTrue(), "Should redirect back to platform")

	// Final check - should be on main dashboard
	Eventually(func() string {
		return l.Page.URL()
	}, 10*time.Second).Should(Equal(l.BaseURL + "/"))

	// Verify we're logged in by checking for user menu
	Eventually(func() bool {
		// Check for user menu element (div.user-menu shows "FFern User")
		userMenu := l.Page.Locator("div.user-menu")
		count, _ := userMenu.Count()
		if count > 0 {
			text, _ := userMenu.TextContent()
			// The UI shows "FFern User" with double F
			if strings.Contains(text, "User") || strings.Contains(text, "Fern") {
				return true
			}
		}

		// Also check for user menu trigger button
		userMenuTrigger := l.Page.Locator("button.user-menu-trigger")
		count, _ = userMenuTrigger.Count()
		if count > 0 {
			return true
		}

		return false
	}, 20*time.Second).Should(BeTrue(), "User menu should appear after login")
}

// Logout performs the logout flow
func (l *LoginHelper) Logout() {
	// Click on user menu trigger button
	userMenuTrigger := l.Page.Locator("button.user-menu-trigger")
	err := userMenuTrigger.Click()
	Expect(err).NotTo(HaveOccurred())

	// Wait a moment for dropdown to appear
	time.Sleep(500 * time.Millisecond)

	// Click logout option - the UI shows "Sign Out" with capital O
	logoutOption := l.Page.Locator("text=Sign Out")
	err = logoutOption.Click()
	Expect(err).NotTo(HaveOccurred())

	// Confirm logout if needed
	confirmBtn := l.Page.Locator("button:has-text('Confirm Logout'), button:has-text('Confirm')")
	count, _ := confirmBtn.Count()
	if count > 0 {
		err = confirmBtn.Click()
		Expect(err).NotTo(HaveOccurred())
	}

	// Should be redirected to login
	Eventually(func() string {
		return l.Page.URL()
	}, 10*time.Second).Should(ContainSubstring("/auth/login"))
}

// IsLoggedIn checks if user is currently logged in
func (l *LoginHelper) IsLoggedIn() bool {
	// Check for user menu or user menu trigger button
	userMenu := l.Page.Locator("div.user-menu, button.user-menu-trigger")
	count, _ := userMenu.Count()
	return count > 0
}
