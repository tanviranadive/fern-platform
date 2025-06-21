package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// KubeVelaManager handles KubeVela application deployment and management
type KubeVelaManager struct {
	namespace      string
	appName        string
	appFile        string
	kubeClient     kubernetes.Interface
	components     []string
	serviceBaseURL string
}

// NewKubeVelaManager creates a new KubeVela manager
func NewKubeVelaManager(namespace, appName, appFile string, kubeClient kubernetes.Interface, components []string) *KubeVelaManager {
	return &KubeVelaManager{
		namespace:  namespace,
		appName:    appName,
		appFile:    appFile,
		kubeClient: kubeClient,
		components: components,
	}
}

// DeployApplication deploys the KubeVela application
func (k *KubeVelaManager) DeployApplication(ctx context.Context) error {
	fmt.Printf("🚀 Deploying KubeVela application %s in namespace %s\n", k.appName, k.namespace)

	// Create namespace-specific application manifest
	appManifest, err := k.createNamespacedManifest()
	if err != nil {
		return fmt.Errorf("failed to create namespaced manifest: %w", err)
	}

	// Apply the KubeVela application
	cmd := exec.CommandContext(ctx, "vela", "up", "-f", "-", "-n", k.namespace)
	cmd.Stdin = appManifest
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy KubeVela application: %w, output: %s", err, string(output))
	}

	fmt.Printf("✅ KubeVela application deployed: %s\n", string(output))
	return nil
}

// WaitForApplicationReady waits for all components to be ready
func (k *KubeVelaManager) WaitForApplicationReady(ctx context.Context) error {
	fmt.Printf("⏳ Waiting for application components to be ready...\n")

	Eventually(func(g Gomega) {
		status, err := k.getApplicationStatus(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Phase).To(Equal("running"))
		
		// Check all components are healthy
		for _, comp := range status.Services {
			g.Expect(comp.Healthy).To(BeTrue(), "Component %s should be healthy", comp.Name)
		}
	}, 5*time.Minute, 10*time.Second).Should(Succeed())

	fmt.Printf("✅ All application components are ready\n")
	return nil
}

// GetServiceURLs returns the service URLs for the deployed application
func (k *KubeVelaManager) GetServiceURLs(ctx context.Context) (map[string]string, error) {
	urls := make(map[string]string)

	// Get service URLs from KubeVela application status
	status, err := k.getApplicationStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get application status: %w", err)
	}

	// Extract service URLs from application status
	for _, service := range status.Services {
		if len(service.URLs) > 0 {
			urls[service.Name] = service.URLs[0]
		} else {
			// Fallback to port-forward for local testing
			port, err := k.getServicePort(ctx, service.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to get port for service %s: %w", service.Name, err)
			}
			urls[service.Name] = fmt.Sprintf("http://localhost:%d", port)
		}
	}

	return urls, nil
}

// DeleteApplication removes the KubeVela application
func (k *KubeVelaManager) DeleteApplication(ctx context.Context) error {
	fmt.Printf("🧹 Deleting KubeVela application %s from namespace %s\n", k.appName, k.namespace)

	cmd := exec.CommandContext(ctx, "vela", "delete", k.appName, "-n", k.namespace, "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete KubeVela application: %w, output: %s", err, string(output))
	}

	fmt.Printf("✅ KubeVela application deleted\n")
	return nil
}

// ApplicationStatus represents the status of a KubeVela application
type ApplicationStatus struct {
	Phase    string `json:"phase"`
	Services []struct {
		Name    string   `json:"name"`
		Healthy bool     `json:"healthy"`
		URLs    []string `json:"urls,omitempty"`
	} `json:"services"`
}

func (k *KubeVelaManager) createNamespacedManifest() (*bytes.Reader, error) {
	// Read the original application file
	originalContent, err := os.ReadFile(k.appFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read application file: %w", err)
	}

	// Parse the YAML
	var appSpec map[string]interface{}
	if err := yaml.Unmarshal(originalContent, &appSpec); err != nil {
		return nil, fmt.Errorf("failed to parse application YAML: %w", err)
	}

	// Update namespace and app name for test isolation
	appSpec["metadata"].(map[string]interface{})["name"] = k.appName
	appSpec["metadata"].(map[string]interface{})["namespace"] = k.namespace

	// Marshal back to YAML
	updatedContent, err := yaml.Marshal(appSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated YAML: %w", err)
	}

	return bytes.NewReader(updatedContent), nil
}

func (k *KubeVelaManager) getApplicationStatus(ctx context.Context) (*ApplicationStatus, error) {
	cmd := exec.CommandContext(ctx, "vela", "status", k.appName, "-n", k.namespace, "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get application status: %w, output: %s", err, string(output))
	}

	var status ApplicationStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse application status: %w", err)
	}

	return &status, nil
}

func (k *KubeVelaManager) getServicePort(ctx context.Context, serviceName string) (int, error) {
	// This would implement port-forward logic or service discovery
	// For now, return default ports based on service name
	defaultPorts := map[string]int{
		"fern-reporter": 8080,
		"fern-mycelium": 8081,
		"fern-ui":       3000,
	}

	if port, exists := defaultPorts[serviceName]; exists {
		return port, nil
	}

	return 0, fmt.Errorf("unknown service: %s", serviceName)
}