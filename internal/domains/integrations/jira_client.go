package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DefaultJiraClient implements the JiraClient interface
type DefaultJiraClient struct {
	httpClient *http.Client
}

// NewDefaultJiraClient creates a new default JIRA client
func NewDefaultJiraClient() *DefaultJiraClient {
	return &DefaultJiraClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TestConnection tests the JIRA connection
func (c *DefaultJiraClient) TestConnection(ctx context.Context, url, username, credential string, authType AuthenticationType) error {
	// Test by getting current user info
	endpoint := fmt.Sprintf("%s/rest/api/2/myself", url)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	c.setAuthHeader(req, username, credential, authType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to JIRA: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JIRA authentication failed: status %d", resp.StatusCode)
	}

	return nil
}

// GetProject retrieves a JIRA project by key
func (c *DefaultJiraClient) GetProject(ctx context.Context, url, projectKey, username, credential string, authType AuthenticationType) (*JiraProject, error) {
	endpoint := fmt.Sprintf("%s/rest/api/2/project/%s", url, projectKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	c.setAuthHeader(req, username, credential, authType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to JIRA: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project '%s' not found", projectKey)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get project: status %d", resp.StatusCode)
	}

	var project JiraProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to parse project response: %w", err)
	}

	return &project, nil
}

// setAuthHeader sets the appropriate authentication header
func (c *DefaultJiraClient) setAuthHeader(req *http.Request, username, credential string, authType AuthenticationType) {
	switch authType {
	case AuthTypeAPIToken:
		// For API token, use Basic auth with email:token
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, credential)))
		req.Header.Set("Authorization", "Basic "+auth)
	case AuthTypeOAuth:
		// For OAuth, use Bearer token
		req.Header.Set("Authorization", "Bearer "+credential)
	case AuthTypePersonalAccessToken:
		// For PAT, use Bearer token
		req.Header.Set("Authorization", "Bearer "+credential)
	}
	req.Header.Set("Accept", "application/json")
}