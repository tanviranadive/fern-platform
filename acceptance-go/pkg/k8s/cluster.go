package k8s

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterManager handles Kubernetes cluster operations for acceptance tests
type ClusterManager struct {
	kubeClient    kubernetes.Interface
	config        *rest.Config
	baseNamespace string
}

// NewClusterManager creates a new cluster manager
func NewClusterManager() (*ClusterManager, error) {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		// Fallback to in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &ClusterManager{
		kubeClient:    kubeClient,
		config:        config,
		baseNamespace: "fern-acceptance",
	}, nil
}

// CreateTestNamespace creates an isolated namespace for test execution
func (c *ClusterManager) CreateTestNamespace(ctx context.Context, testID string) (string, error) {
	namespaceName := fmt.Sprintf("%s-%s", c.baseNamespace, testID)

	GinkgoHelper()
	By(fmt.Sprintf("Creating test namespace: %s", namespaceName))

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"test-id":     testID,
				"test-suite":  "fern-acceptance",
				"managed-by":  "ginkgo",
			},
		},
	}

	_, err := c.kubeClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create namespace %s: %w", namespaceName, err)
	}

	fmt.Printf("✅ Created test namespace: %s\n", namespaceName)
	return namespaceName, nil
}

// DeleteTestNamespace removes the test namespace and all resources
func (c *ClusterManager) DeleteTestNamespace(ctx context.Context, namespaceName string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Deleting test namespace: %s", namespaceName))

	err := c.kubeClient.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", namespaceName, err)
	}

	// Wait for namespace to be fully deleted
	Eventually(func(g Gomega) {
		_, err := c.kubeClient.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
		g.Expect(err).To(HaveOccurred())
		g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
	}, 2*time.Minute, 5*time.Second).Should(Succeed())

	fmt.Printf("✅ Deleted test namespace: %s\n", namespaceName)
	return nil
}

// WaitForNamespaceReady waits for namespace to be in Active phase
func (c *ClusterManager) WaitForNamespaceReady(ctx context.Context, namespaceName string) error {
	GinkgoHelper()
	By(fmt.Sprintf("Waiting for namespace %s to be ready", namespaceName))

	Eventually(func(g Gomega) {
		ns, err := c.kubeClient.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ns.Status.Phase).To(Equal(corev1.NamespaceActive))
	}, 30*time.Second, 2*time.Second).Should(Succeed())

	return nil
}

// GetKubeClient returns the Kubernetes client
func (c *ClusterManager) GetKubeClient() kubernetes.Interface {
	return c.kubeClient
}

// GetConfig returns the Kubernetes REST config
func (c *ClusterManager) GetConfig() *rest.Config {
	return c.config
}

// VerifyClusterPrerequisites checks that required components are available
func (c *ClusterManager) VerifyClusterPrerequisites(ctx context.Context) error {
	GinkgoHelper()
	By("Verifying cluster prerequisites")

	// Check if KubeVela is installed
	if err := c.verifyKubeVela(ctx); err != nil {
		return fmt.Errorf("KubeVela verification failed: %w", err)
	}

	// Check if CNPG operator is available
	if err := c.verifyCNPG(ctx); err != nil {
		return fmt.Errorf("CNPG verification failed: %w", err)
	}

	fmt.Printf("✅ All cluster prerequisites verified\n")
	return nil
}

func (c *ClusterManager) verifyKubeVela(ctx context.Context) error {
	// Check if KubeVela CRDs exist
	discoveryClient := c.kubeClient.Discovery()
	
	resources, err := discoveryClient.ServerResourcesForGroupVersion("core.oam.dev/v1beta1")
	if err != nil {
		return fmt.Errorf("KubeVela CRDs not found: %w", err)
	}

	// Check for Application CRD
	found := false
	for _, resource := range resources.APIResources {
		if resource.Kind == "Application" {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("KubeVela Application CRD not found")
	}

	return nil
}

func (c *ClusterManager) verifyCNPG(ctx context.Context) error {
	// Check if CNPG operator is running
	deployments, err := c.kubeClient.AppsV1().Deployments("cnpg-system").List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cloudnative-pg",
	})
	if err != nil {
		return fmt.Errorf("failed to check CNPG operator: %w", err)
	}

	if len(deployments.Items) == 0 {
		return fmt.Errorf("CNPG operator not found in cnpg-system namespace")
	}

	// Check if operator is ready
	for _, deployment := range deployments.Items {
		if deployment.Status.ReadyReplicas == 0 {
			return fmt.Errorf("CNPG operator is not ready")
		}
	}

	return nil
}