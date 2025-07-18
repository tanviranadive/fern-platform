package auth_test

import (
	"flag"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/playwright-community/playwright-go"
)

var (
	// Configuration flags
	baseURL     string
	headless    bool
	slowMo      float64
	username    string
	password    string
	recordVideo bool

	// Playwright objects
	pw *playwright.Playwright
)

func init() {
	flag.StringVar(&baseURL, "base-url", getEnvOrDefault("FERN_BASE_URL", "http://fern-platform.local:8080"), "Base URL for Fern Platform")
	flag.BoolVar(&headless, "headless", getEnvOrDefault("FERN_HEADLESS", "true") == "true", "Run browser in headless mode")
	flag.Float64Var(&slowMo, "slow-mo", 0, "Slow motion delay in milliseconds")
	flag.StringVar(&username, "username", getEnvOrDefault("FERN_USERNAME", "fern-user@fern.com"), "Username for authentication")
	flag.StringVar(&password, "password", getEnvOrDefault("FERN_PASSWORD", "test123"), "Password for authentication")
	flag.BoolVar(&recordVideo, "record-video", getEnvOrDefault("FERN_RECORD_VIDEO", "false") == "true", "Record videos of test runs")
}

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authentication Test Suite")
}

var _ = BeforeSuite(func() {
	var err error

	// Install playwright browsers if needed
	err = playwright.Install()
	Expect(err).NotTo(HaveOccurred())

	// Initialize playwright
	pw, err = playwright.Run()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if pw != nil {
		defer func() {
			recover()
		}()
		pw.Stop()
	}
})

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateBrowser creates a new browser instance
func CreateBrowser() playwright.Browser {
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		SlowMo:   playwright.Float(slowMo),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--disable-web-security",
			"--disable-features=IsolateOrigins,site-per-process",
		},
	})
	Expect(err).NotTo(HaveOccurred())
	return browser
}
