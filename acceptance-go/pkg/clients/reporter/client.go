// Package reporter provides a client for the Fern Platform reporter API
package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

// Client represents the Fern Platform reporter API client
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Fern Platform reporter API client
func NewClient(baseURL string) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		baseURL: parsedURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// TestRun represents a test run
type TestRun struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	SuiteID   string    `json:"suiteId"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime,omitempty"`
	Duration  int64     `json:"duration"`
	Branch    string    `json:"branch"`
	Tags      []string  `json:"tags"`
	SpecRuns  []SpecRun `json:"specRuns,omitempty"`
}

// SpecRun represents a spec run
type SpecRun struct {
	ID              string    `json:"id"`
	TestRunID       string    `json:"testRunId"`
	SpecDescription string    `json:"specDescription"`
	Status          string    `json:"status"`
	StartTime       time.Time `json:"startTime"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	Duration        int64     `json:"duration"`
	ErrorMessage    string    `json:"errorMessage,omitempty"`
	StackTrace      string    `json:"stackTrace,omitempty"`
}

// Project represents a project
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TestRunsResponse represents the response from the test runs endpoint
type TestRunsResponse struct {
	Data       []TestRun `json:"data"`
	TotalCount int       `json:"totalCount"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
}

// ProjectsResponse represents the response from the projects endpoint
type ProjectsResponse struct {
	Data       []Project `json:"data"`
	TotalCount int       `json:"totalCount"`
}

// HealthCheck performs a health check on the reporter service
func (c *Client) HealthCheck(ctx context.Context) error {
	GinkgoHelper()
	
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL.JoinPath("/health").String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetTestRuns retrieves test runs with optional filters
func (c *Client) GetTestRuns(ctx context.Context, opts *TestRunsOptions) (*TestRunsResponse, error) {
	GinkgoHelper()
	
	endpoint := c.baseURL.JoinPath("/api/v1/test-runs")
	
	// Add query parameters
	if opts != nil {
		query := endpoint.Query()
		if opts.ProjectID != "" {
			query.Set("projectId", opts.ProjectID)
		}
		if opts.Status != "" {
			query.Set("status", opts.Status)
		}
		if opts.Branch != "" {
			query.Set("branch", opts.Branch)
		}
		if opts.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			query.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response TestRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetTestRun retrieves a specific test run by ID
func (c *Client) GetTestRun(ctx context.Context, testRunID string) (*TestRun, error) {
	GinkgoHelper()
	
	endpoint := c.baseURL.JoinPath("/api/v1/test-runs", testRunID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("test run not found: %s", testRunID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var testRun TestRun
	if err := json.NewDecoder(resp.Body).Decode(&testRun); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &testRun, nil
}

// GetProjects retrieves all projects
func (c *Client) GetProjects(ctx context.Context) (*ProjectsResponse, error) {
	GinkgoHelper()
	
	endpoint := c.baseURL.JoinPath("/api/v1/projects")
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CreateTestRun creates a new test run
func (c *Client) CreateTestRun(ctx context.Context, testRun *TestRun) (*TestRun, error) {
	GinkgoHelper()
	
	endpoint := c.baseURL.JoinPath("/api/v1/test-runs")
	
	body, err := json.Marshal(testRun)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal test run: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var createdTestRun TestRun
	if err := json.NewDecoder(resp.Body).Decode(&createdTestRun); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdTestRun, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, project *Project) (*Project, error) {
	GinkgoHelper()
	
	endpoint := c.baseURL.JoinPath("/api/v1/projects")
	
	body, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var createdProject Project
	if err := json.NewDecoder(resp.Body).Decode(&createdProject); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdProject, nil
}

// TestRunsOptions represents options for querying test runs
type TestRunsOptions struct {
	ProjectID string
	Status    string
	Branch    string
	Limit     int
	Offset    int
}

// WithTimeout sets a custom timeout for the HTTP client
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}