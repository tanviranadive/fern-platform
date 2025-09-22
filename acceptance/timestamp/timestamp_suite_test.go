package timestamp_test

import (
	"flag"
	"os"
	"runtime"
	"strings"
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
	flag.StringVar(&username, "username", getEnvOrDefault("FERN_USERNAME", "fern-manager@fern.com"), "Username for authentication")
	flag.StringVar(&password, "password", getEnvOrDefault("FERN_PASSWORD", "test123"), "Password for authentication")
	flag.BoolVar(&recordVideo, "record-video", getEnvOrDefault("FERN_RECORD_VIDEO", "false") == "true", "Record videos of test runs")
}

func TestTimestamp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Timestamp Component Suite")
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
			if r := recover(); r != nil {
				GinkgoWriter.Printf("Warning: panic during Playwright shutdown: %v\n", r)
			}
		}()
		if err := pw.Stop(); err != nil {
			GinkgoWriter.Printf("Warning: error stopping Playwright: %v\n", err)
		}
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
	args := []string{
		"--disable-blink-features=AutomationControlled",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-dev-shm-usage",
		"--disable-gpu",
		"--disable-web-security",
		"--disable-features=IsolateOrigins,site-per-process",
		"--disable-accelerated-2d-canvas",
		"--disable-audio-output",
	}

	// Add platform-specific args
	if runtime.GOOS == "darwin" {
		// Mac-specific: helps with TLS certificate issues
		args = append(args, "--single-process", "--no-zygote")
	} else if isRunningInDocker() {
		// Docker/CI-specific: additional stability flags
		args = append(args,
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
		)
	}

	// Allow CI to override with custom args
	if customArgs := os.Getenv("PLAYWRIGHT_CHROMIUM_ARGS"); customArgs != "" {
		extraArgs := parseShellArgs(customArgs)
		args = append(args, extraArgs...)
	}

	// Log browser launch args in verbose mode
	if os.Getenv("DEBUG") != "" {
		GinkgoWriter.Printf("Launching Chromium with args: %v\n", args)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		SlowMo:   playwright.Float(slowMo),
		Args:     args,
	})
	Expect(err).NotTo(HaveOccurred())
	return browser
}

// isRunningInDocker checks if we're running inside a Docker container
func isRunningInDocker() bool {
	// Check for .dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	// Check for Docker in cgroup
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker") || strings.Contains(string(data), "containerd")
	}
	return false
}

// parseShellArgs parses shell-style arguments, handling quotes and escapes properly
func parseShellArgs(input string) []string {
	var args []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune
	
	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		
		switch {
		case !inQuotes && (char == '"' || char == '\''):
			// Start quoted section
			inQuotes = true
			quoteChar = char
		case inQuotes && char == quoteChar:
			// End quoted section
			inQuotes = false
			quoteChar = 0
		case !inQuotes && (char == ' ' || char == '\t'):
			// Whitespace outside quotes - end current arg
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			// Skip consecutive whitespace
			for i+1 < len(runes) && (runes[i+1] == ' ' || runes[i+1] == '\t') {
				i++
			}
		case char == '\\' && i+1 < len(runes):
			// Escape sequence
			i++
			current.WriteRune(runes[i])
		default:
			// Regular character
			current.WriteRune(char)
		}
	}
	
	// Add final argument if any
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	
	return args
}
