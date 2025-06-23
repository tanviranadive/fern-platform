package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	fern "github.com/guidewire-oss/fern-ginkgo-client/pkg/client"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/graphql"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/clients/reporter"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/fixtures"
	"github.com/guidewire-oss/fern-platform/acceptance-go/pkg/k8s"
)

// Test suite variables
var (
	clusterManager  *k8s.ClusterManager
	kubeVelaManager *k8s.KubeVelaManager
	reporterClient  *reporter.Client
	graphqlClient   *graphql.Client
	testDataManager *fixtures.TestDataManager

	testNamespace string
	serviceURLs   map[string]string
	suiteCtx      context.Context
	suiteCancel   context.CancelFunc

	// Configuration flags
	// Set useExistingPlatform to:
	//   - true: Use deployed platform at existingPlatformURL (faster, for sending reports to fern-platform)
	//   - false: Deploy fresh platform in k3d cluster (full isolation, for testing platform itself)
	useExistingPlatform = true
	existingPlatformURL = "http://localhost:8080"
)

func TestIntegrationAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	// Configure fern-ginkgo-client to report to the deployed platform
	fernApiClient := fern.New("d04c8eed-6b99-40af-bf74-adeda2f2fd7b", fern.WithBaseURL("http://localhost:8080"))

	// Register the fern reporter with correct signature for Ginkgo v2
	ReportAfterSuite("Fern Platform Reporter", func(report types.Report) {
		err := fernApiClient.Report(report)
		if err != nil {
			GinkgoLogr.Error(err, "Failed to send test report to fern-platform")
		}
	})

	RunSpecs(t, "Integration Acceptance Test Suite")
}

var _ = BeforeSuite(func() {
	By("Setting up integration acceptance test suite")

	// Create suite context with timeout
	if useExistingPlatform {
		suiteCtx, suiteCancel = context.WithTimeout(context.Background(), 5*time.Minute)
	} else {
		suiteCtx, suiteCancel = context.WithTimeout(context.Background(), 15*time.Minute)
	}

	// Generate unique test identifier for this suite execution
	testID := GinkgoRandomSeed()
	testNamespace = fmt.Sprintf("fern-integration-test-%d-%d", testID, GinkgoParallelProcess())

	if useExistingPlatform {
		By("Connecting to existing deployed fern-platform for integration tests")

		// Use the existing deployed platform
		serviceURLs = map[string]string{
			"fern-reporter": existingPlatformURL,
		}

		// Initialize API clients
		var err error
		reporterClient, err = reporter.NewClient(serviceURLs["fern-reporter"])
		Expect(err).NotTo(HaveOccurred())

		graphqlClient, err = graphql.NewClient(serviceURLs["fern-reporter"])
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

		// Initialize KubeVela manager with all services for integration testing
		kubeVelaManager = k8s.NewKubeVelaManager(
			testNamespace,
			"fern-platform-integration-test",
			"../../../deployments/fern-platform-local.yaml",
			clusterManager.GetKubeClient(),
			[]string{"postgres", "redis", "fern-reporter", "fern-mycelium", "fern-ui"},
		)

		// Deploy KubeVela application
		By("Deploying complete KubeVela application for integration testing")
		Expect(kubeVelaManager.DeployApplication(suiteCtx)).To(Succeed())

		// Wait for all services to be ready
		By("Waiting for all services to be ready")
		Expect(kubeVelaManager.WaitForApplicationReady(suiteCtx)).To(Succeed())

		// Get service URLs
		serviceURLs, err = kubeVelaManager.GetServiceURLs(suiteCtx)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceURLs).To(HaveKey("fern-reporter"))
		Expect(serviceURLs).To(HaveKey("fern-ui"))

		// Initialize API clients
		reporterClient, err = reporter.NewClient(serviceURLs["fern-reporter"])
		Expect(err).NotTo(HaveOccurred())

		graphqlClient, err = graphql.NewClient(serviceURLs["fern-reporter"])
		Expect(err).NotTo(HaveOccurred())

		// Wait for services to be responsive
		By("Waiting for all services to be responsive")
		Eventually(func() error {
			return reporterClient.HealthCheck(suiteCtx)
		}, 3*time.Minute, 10*time.Second).Should(Succeed())
	}

	// Initialize test data manager
	testDataManager = fixtures.NewTestDataManager(reporterClient, testNamespace, fmt.Sprintf("%d", testID))

	if !useExistingPlatform {
		// Create comprehensive test data for integration testing
		By("Setting up comprehensive test data for integration testing")
		Expect(testDataManager.SetupTestData(suiteCtx)).To(Succeed())
	} else {
		By("Using existing platform data for integration tests (no new test data created)")
		// Initialize with empty test data to avoid nil pointer issues
		_ = testDataManager.InitializeWithExistingData(suiteCtx)
	}

	By("✅ Integration acceptance test suite setup complete")
})

var _ = AfterSuite(func() {
	By("Cleaning up integration acceptance test suite")

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

	By("✅ Integration acceptance test suite cleanup complete")
})

// Helper functions for common operations
func GetReporterClient() *reporter.Client {
	GinkgoHelper()
	Expect(reporterClient).NotTo(BeNil(), "Reporter client not initialized")
	return reporterClient
}

func GetGraphQLClient() *graphql.Client {
	GinkgoHelper()
	Expect(graphqlClient).NotTo(BeNil(), "GraphQL client not initialized")
	return graphqlClient
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

func GetServiceURLs() map[string]string {
	GinkgoHelper()
	Expect(serviceURLs).NotTo(BeNil(), "Service URLs not initialized")
	return serviceURLs
}
