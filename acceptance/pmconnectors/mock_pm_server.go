package pmconnectors_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// MockPMServer simulates JIRA/Aha/other PM tool APIs for testing
type MockPMServer struct {
	server *httptest.Server
	// Track requests for assertions
	RequestLog []MockRequest
	// Configurable responses
	Responses map[string]MockResponse
}

type MockRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
	Time    time.Time
}

type MockResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// NewMockPMServer creates a new mock PM server
func NewMockPMServer() *MockPMServer {
	m := &MockPMServer{
		RequestLog: make([]MockRequest, 0),
		Responses:  make(map[string]MockResponse),
	}

	// Set up default JIRA-like responses
	m.setupDefaultResponses()

	// Create test server
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))

	return m
}

// URL returns the mock server URL
func (m *MockPMServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockPMServer) Close() {
	m.server.Close()
}

// handler processes all requests
func (m *MockPMServer) handler(w http.ResponseWriter, r *http.Request) {
	// Log the request
	body := make([]byte, 0)
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body.Close()
	}

	m.RequestLog = append(m.RequestLog, MockRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header,
		Body:    body,
		Time:    time.Now(),
	})

	// Route to appropriate handler based on path
	switch {
	case r.URL.Path == "/rest/api/2/myself" && r.Method == "GET":
		m.handleAuthTest(w, r)
	case r.URL.Path == "/rest/api/2/search" && r.Method == "GET":
		m.handleJiraSearch(w, r)
	case r.URL.Path == "/rest/api/2/issue" && r.Method == "POST":
		m.handleJiraCreateIssue(w, r)
	case r.URL.Path == "/api/v1/features" && r.Method == "GET":
		m.handleAhaFeatures(w, r)
	default:
		// Check for configured responses
		key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if resp, ok := m.Responses[key]; ok {
			m.sendResponse(w, resp)
			return
		}

		// Default 404
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Not found",
		})
	}
}

// handleAuthTest simulates JIRA authentication test
func (m *MockPMServer) handleAuthTest(w http.ResponseWriter, r *http.Request) {
	// Check for authentication header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Authentication required",
		})
		return
	}

	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":          "test-user",
		"name":         "Test User",
		"emailAddress": "test@example.com",
		"displayName":  "Test User",
		"active":       true,
	})
}

// handleJiraSearch simulates JIRA issue search
func (m *MockPMServer) handleJiraSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	jql := r.URL.Query().Get("jql")
	startAt := r.URL.Query().Get("startAt")
	maxResults := r.URL.Query().Get("maxResults")

	// Generate mock issues
	issues := m.generateMockJiraIssues(jql, startAt, maxResults)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
}

// generateMockJiraIssues creates mock JIRA issues
func (m *MockPMServer) generateMockJiraIssues(jql, startAt, maxResults string) map[string]interface{} {
	start := 0
	if startAt != "" {
		fmt.Sscanf(startAt, "%d", &start)
	}

	max := 50
	if maxResults != "" {
		fmt.Sscanf(maxResults, "%d", &max)
	}

	// Generate issues
	issues := make([]map[string]interface{}, 0)
	for i := start; i < start+max && i < 100; i++ {
		issues = append(issues, map[string]interface{}{
			"id":  fmt.Sprintf("1000%d", i),
			"key": fmt.Sprintf("TEST-%d", i+1),
			"fields": map[string]interface{}{
				"summary":     fmt.Sprintf("Test Issue %d", i+1),
				"description": fmt.Sprintf("This is test issue number %d", i+1),
				"status": map[string]interface{}{
					"name": []string{"To Do", "In Progress", "Done"}[i%3],
				},
				"issuetype": map[string]interface{}{
					"name": []string{"Story", "Bug", "Task"}[i%3],
				},
				"priority": map[string]interface{}{
					"name": []string{"High", "Medium", "Low"}[i%3],
				},
				"assignee": map[string]interface{}{
					"displayName": fmt.Sprintf("User %d", (i%5)+1),
				},
				"fixVersions": []map[string]interface{}{
					{
						"name": fmt.Sprintf("Release %d.%d", (i/10)+1, (i%10)+1),
					},
				},
				"customfield_10001": fmt.Sprintf("Team %d", (i%3)+1), // Team field
				"created":           time.Now().Add(-time.Duration(i*24) * time.Hour).Format(time.RFC3339),
				"updated":           time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
			},
		})
	}

	return map[string]interface{}{
		"startAt":    start,
		"maxResults": max,
		"total":      100,
		"issues":     issues,
	}
}

// handleJiraCreateIssue simulates JIRA issue creation
func (m *MockPMServer) handleJiraCreateIssue(w http.ResponseWriter, r *http.Request) {
	// Check for authentication
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid authentication",
		})
		return
	}

	// Parse the request body
	var createReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Create a mock response
	issueID := fmt.Sprintf("TEST-%d", time.Now().Unix()%10000)
	response := map[string]interface{}{
		"id":   fmt.Sprintf("100%d", time.Now().Unix()%10000),
		"key":  issueID,
		"self": fmt.Sprintf("http://mock-jira:8080/rest/api/2/issue/%s", issueID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAhaFeatures simulates Aha! features API
func (m *MockPMServer) handleAhaFeatures(w http.ResponseWriter, r *http.Request) {
	// Check for authentication
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid authentication",
		})
		return
	}

	// Generate mock features
	features := m.generateMockAhaFeatures()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(features)
}

// generateMockAhaFeatures creates mock Aha! features
func (m *MockPMServer) generateMockAhaFeatures() map[string]interface{} {
	features := make([]map[string]interface{}, 0)

	for i := 0; i < 20; i++ {
		features = append(features, map[string]interface{}{
			"id":            fmt.Sprintf("FEAT-%d", i+1),
			"reference_num": fmt.Sprintf("FEAT-%d", i+1),
			"name":          fmt.Sprintf("Feature %d", i+1),
			"description": map[string]interface{}{
				"body": fmt.Sprintf("This is feature number %d", i+1),
			},
			"workflow_status": map[string]interface{}{
				"name": []string{"Under consideration", "In development", "Shipped"}[i%3],
			},
			"score": []int{100, 80, 60}[i%3],
			"release": map[string]interface{}{
				"name": fmt.Sprintf("Release %d.%d", (i/5)+1, (i%5)+1),
			},
			"assigned_to_user": map[string]interface{}{
				"name": fmt.Sprintf("Product Manager %d", (i%3)+1),
			},
			"team": map[string]interface{}{
				"name": fmt.Sprintf("Team %d", (i%2)+1),
			},
			"created_at": time.Now().Add(-time.Duration(i*48) * time.Hour).Format(time.RFC3339),
			"updated_at": time.Now().Add(-time.Duration(i*2) * time.Hour).Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		"features": features,
		"pagination": map[string]interface{}{
			"total_records": len(features),
			"total_pages":   1,
			"current_page":  1,
		},
	}
}

// setupDefaultResponses configures default mock responses
func (m *MockPMServer) setupDefaultResponses() {
	// Health check endpoint
	m.Responses["GET /health"] = MockResponse{
		StatusCode: http.StatusOK,
		Body: map[string]string{
			"status": "healthy",
		},
	}

	// JIRA version endpoint
	m.Responses["GET /rest/api/2/serverInfo"] = MockResponse{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"version":        "8.20.0",
			"versionNumbers": []int{8, 20, 0},
			"deploymentType": "Cloud",
			"buildNumber":    820000,
		},
	}
}

// SetResponse configures a custom response for a specific endpoint
func (m *MockPMServer) SetResponse(method, path string, response MockResponse) {
	key := fmt.Sprintf("%s %s", method, path)
	m.Responses[key] = response
}

// sendResponse writes a mock response
func (m *MockPMServer) sendResponse(w http.ResponseWriter, resp MockResponse) {
	// Set headers
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	// Set status code
	if resp.StatusCode != 0 {
		w.WriteHeader(resp.StatusCode)
	}

	// Write body
	if resp.Body != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Body)
	}
}

// GetRequests returns all logged requests
func (m *MockPMServer) GetRequests() []MockRequest {
	return m.RequestLog
}

// GetRequestsForPath returns requests for a specific path
func (m *MockPMServer) GetRequestsForPath(method, path string) []MockRequest {
	var filtered []MockRequest
	for _, req := range m.RequestLog {
		if req.Method == method && req.Path == path {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// Reset clears the request log
func (m *MockPMServer) Reset() {
	m.RequestLog = make([]MockRequest, 0)
}

// SimulateError configures an error response
func (m *MockPMServer) SimulateError(method, path string, statusCode int, message string) {
	m.SetResponse(method, path, MockResponse{
		StatusCode: statusCode,
		Body: map[string]string{
			"error": message,
		},
	})
}

// SimulateTimeout simulates a timeout by delaying the response
func (m *MockPMServer) SimulateTimeout(method, path string, delay time.Duration) {
	key := fmt.Sprintf("%s %s", method, path)
	// Store original response
	original := m.Responses[key]

	// Create delayed handler
	m.Responses[key] = MockResponse{
		StatusCode: http.StatusOK,
		Body: func() interface{} {
			time.Sleep(delay)
			return original.Body
		}(),
	}
}
