package ui_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	fern "github.com/guidewire-oss/fern-ginkgo-client/pkg/client"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/reporter"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/fixtures"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/k8s"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/pages"
)

// Test suite variables
var (
	clusterManager  *k8s.ClusterManager
	kubeVelaManager *k8s.KubeVelaManager
	reporterClient  *reporter.Client
	testDataManager *fixtures.TestDataManager

	testNamespace string
	serviceURLs   map[string]string
	suiteCtx      context.Context
	suiteCancel   context.CancelFunc

	// Browser context for UI testing
	browserCtx    context.Context
	browserCancel context.CancelFunc

	// Configuration flags
	// Set useExistingPlatform to:
	//   - true: Use deployed platform at existingPlatformURL (faster, for sending reports to fern-platform)
	//   - false: Deploy fresh platform in k3d cluster (full isolation, for testing platform itself)
	useExistingPlatform = true
	existingPlatformURL = "http://localhost:8080"
)

func TestUIAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	// Configure fern-ginkgo-client to report to the deployed platform
	fernApiClient := fern.New("8a02b62f-1bb4-408a-ad2d-1dca8c1f1449", fern.WithBaseURL("http://localhost:8080"))

	// Register the fern reporter with correct signature for Ginkgo v2
	ReportAfterSuite("Fern Platform Reporter", func(report types.Report) {
		err := fernApiClient.Report(report)
		if err != nil {
			GinkgoLogr.Error(err, "Failed to send test report to fern-platform")
		}
	})

	RunSpecs(t, "UI Acceptance Test Suite")
}

var _ = BeforeSuite(func() {
	By("Setting up UI acceptance test suite")

	// Create suite context with timeout
	if useExistingPlatform {
		suiteCtx, suiteCancel = context.WithTimeout(context.Background(), 10*time.Minute)
	} else {
		suiteCtx, suiteCancel = context.WithTimeout(context.Background(), 20*time.Minute)
	}

	// Generate unique test identifier for this suite execution
	testID := GinkgoRandomSeed()
	testNamespace = fmt.Sprintf("fern-ui-test-%d-%d", testID, GinkgoParallelProcess())

	if useExistingPlatform {
		By("Connecting to existing deployed fern-platform for UI tests")

		// Use the existing deployed platform
		serviceURLs = map[string]string{
			"fern-reporter": existingPlatformURL,
			"fern-ui":       existingPlatformURL, // UI is served from the same platform
		}

		// Initialize API client
		var err error
		reporterClient, err = reporter.NewClient(serviceURLs["fern-reporter"])
		Expect(err).NotTo(HaveOccurred())

		// Wait for services to be responsive
		By("Waiting for existing platform to be responsive")
		Eventually(func() error {
			return reporterClient.HealthCheck(suiteCtx)
		}, 2*time.Minute, 5*time.Second).Should(Succeed())

	} else {
		By(fmt.Sprintf("Creating isolated test environment: %s", testNamespace))

		// Initialize cluster manager
		var err error
		clusterManager, err = k8s.NewClusterManager()
		Expect(err).NotTo(HaveOccurred(), "Failed to create cluster manager")

		// Verify cluster prerequisites (KubeVela, CNPG)
		Expect(clusterManager.VerifyClusterPrerequisites(suiteCtx)).To(Succeed())

		// Create isolated namespace for this test suite
		_, err = clusterManager.CreateTestNamespace(suiteCtx, fmt.Sprintf("%d-%d", testID, GinkgoParallelProcess()))
		Expect(err).NotTo(HaveOccurred())

		// Wait for namespace to be ready
		Expect(clusterManager.WaitForNamespaceReady(suiteCtx, testNamespace)).To(Succeed())

		// Initialize KubeVela manager with all services including UI
		kubeVelaManager = k8s.NewKubeVelaManager(
			testNamespace,
			"fern-platform-ui-test",
			"../../../deployments/fern-platform-local.yaml",
			clusterManager.GetKubeClient(),
			[]string{"postgres", "redis", "fern-reporter", "fern-mycelium", "fern-ui"},
		)

		// Deploy KubeVela application
		By("Deploying complete KubeVela application including UI")
		Expect(kubeVelaManager.DeployApplication(suiteCtx)).To(Succeed())

		// Wait for all services to be ready
		By("Waiting for all services including UI to be ready")
		Expect(kubeVelaManager.WaitForApplicationReady(suiteCtx)).To(Succeed())

		// Get service URLs
		serviceURLs, err = kubeVelaManager.GetServiceURLs(suiteCtx)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceURLs).To(HaveKey("fern-reporter"))
		Expect(serviceURLs).To(HaveKey("fern-ui"))

		// Initialize API client for test data setup
		reporterClient, err = reporter.NewClient(serviceURLs["fern-reporter"])
		Expect(err).NotTo(HaveOccurred())

		// Wait for services to be responsive
		By("Waiting for API and UI services to be responsive")
		Eventually(func() error {
			return reporterClient.HealthCheck(suiteCtx)
		}, 3*time.Minute, 10*time.Second).Should(Succeed())
	}

	// Initialize test data manager
	testDataManager = fixtures.NewTestDataManager(reporterClient, testNamespace, fmt.Sprintf("%d", testID))

	if !useExistingPlatform {
		// Create comprehensive test data for UI testing
		By("Setting up comprehensive test data for UI testing")
		Expect(testDataManager.SetupTestData(suiteCtx)).To(Succeed())
	} else {
		By("Using existing platform data for UI tests (no new test data created)")
		// Initialize with empty test data to avoid nil pointer issues
		_ = testDataManager.InitializeWithExistingData(suiteCtx)
	}

	// Setup browser context for UI testing
	By("Setting up browser context for UI testing")
	setupBrowserContext()

	By("✅ UI acceptance test suite setup complete")
})

var _ = AfterSuite(func() {
	By("Cleaning up UI acceptance test suite")

	// Cleanup browser context first
	if browserCancel != nil {
		browserCancel()
	}

	defer suiteCancel()

	// Cleanup test data
	if testDataManager != nil {
		_ = testDataManager.CleanupTestData(suiteCtx)
	}

	if !useExistingPlatform {
		// Delete KubeVela application
		if kubeVelaManager != nil {
			By("Deleting KubeVela application")
			_ = kubeVelaManager.DeleteApplication(suiteCtx)
		}

		// Delete test namespace
		if clusterManager != nil && testNamespace != "" {
			By("Deleting test namespace")
			_ = clusterManager.DeleteTestNamespace(suiteCtx, testNamespace)
		}
	}

	By("✅ UI acceptance test suite cleanup complete")
})

func setupBrowserContext() {
	// Setup Chrome browser context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, _ := chromedp.NewExecAllocator(suiteCtx, opts...)
	browserCtx, browserCancel = chromedp.NewContext(allocCtx)

	// Initialize browser
	err := chromedp.Run(browserCtx, chromedp.Navigate("about:blank"))
	Expect(err).NotTo(HaveOccurred())
}

// Helper functions for common operations
func GetReporterClient() *reporter.Client {
	GinkgoHelper()
	Expect(reporterClient).NotTo(BeNil(), "Reporter client not initialized")
	return reporterClient
}

func GetTestData() *fixtures.CreatedTestData {
	GinkgoHelper()
	Expect(testDataManager).NotTo(BeNil(), "Test data manager not initialized")
	return testDataManager.GetCreatedData()
}

func GetTestContext() context.Context {
	GinkgoHelper()
	Expect(suiteCtx).NotTo(BeNil(), "Suite context not initialized")
	return suiteCtx
}

func GetBrowserContext() context.Context {
	GinkgoHelper()
	Expect(browserCtx).NotTo(BeNil(), "Browser context not initialized")
	return browserCtx
}

func GetServiceURLs() map[string]string {
	GinkgoHelper()
	Expect(serviceURLs).NotTo(BeNil(), "Service URLs not initialized")
	return serviceURLs
}

func GetUIBaseURL() string {
	GinkgoHelper()
	urls := GetServiceURLs()
	Expect(urls).To(HaveKey("fern-ui"))
	return urls["fern-ui"]
}

// Page object factory functions
func NewDashboardPage() *pages.DashboardPage {
	return pages.NewDashboardPage(GetUIBaseURL(), GetBrowserContext())
}

func NewTestRunsPage() *pages.TestRunsPage {
	return pages.NewTestRunsPage(GetUIBaseURL(), GetBrowserContext())
}
