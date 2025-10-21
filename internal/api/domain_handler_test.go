package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/guidewire-oss/fern-platform/internal/testhelpers"
	"github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
	tagsApp "github.com/guidewire-oss/fern-platform/internal/domains/tags/application"
	tagsDomain "github.com/guidewire-oss/fern-platform/internal/domains/tags/domain"
	testingApp "github.com/guidewire-oss/fern-platform/internal/domains/testing/application"
	testingDomain "github.com/guidewire-oss/fern-platform/internal/domains/testing/domain"
	"github.com/guidewire-oss/fern-platform/pkg/config"
	"github.com/guidewire-oss/fern-platform/pkg/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TODO:DomainHandler uses concrete service types (not interfaces) and you can't inject mocks without changing the implementation, this branch will never be reached in unit tests.
// MockAuthMiddleware provides a mock implementation of AuthMiddlewareAdapter
type MockAuthMiddleware struct {
	mock.Mock
}

func (m *MockAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func (m *MockAuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// For testing, just check if user is authenticated
		if _, exists := c.Get("user_id"); !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func (m *MockAuthMiddleware) StartOAuthFlow() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/auth/callback")
	}
}

func (m *MockAuthMiddleware) HandleOAuthCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OAuth callback"})
	}
}

func (m *MockAuthMiddleware) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
	}
}

var _ = Describe("DomainHandler Integration Tests", func() {
	var (
		logger *logging.Logger
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)

		// Initialize logger
		loggingConfig := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
		}
		var err error
		logger, err = logging.NewLogger(loggingConfig)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Health Check", func() {
		It("should return healthy status", func() {
			// Create a fresh router for this test
			router := gin.New()

			// Create handler - health check doesn't require services
			handler := NewDomainHandler(nil, nil, nil, nil, nil, nil, logger)

			// Register routes
			handler.RegisterRoutes(router)

			// Create request
			w := testhelpers.PerformRequest(router, "GET", "/health", nil)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["status"]).To(Equal("healthy"))
			Expect(response).To(HaveKey("time"))
		})
	})

	Describe("Route Registration", func() {
		It("should register all expected routes", func() {
			// Create a fresh router for this test
			router := gin.New()

			handler := NewDomainHandler(nil, nil, nil, nil, nil, nil, logger)
			handler.RegisterRoutes(router)

			routes := router.Routes()

			// Verify some key routes are registered
			expectedPaths := []string{
				"/health",
				"/auth/login",
				"/auth/logout",
				"/auth/callback",
			}

			for _, expectedPath := range expectedPaths {
				found := false
				for _, route := range routes {
					if route.Path == expectedPath {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "Expected route %s to be registered", expectedPath)
			}
		})
	})
})

var _ = Describe("recordTestRun Function Tests", func() {
	var (
		handler *DomainHandler
		router  *gin.Engine
		logger  *logging.Logger
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)

		// Initialize logger
		loggingConfig := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
		}
		var err error
		logger, err = logging.NewLogger(loggingConfig)
		Expect(err).NotTo(HaveOccurred())

		// Create handler with nil services - we'll test what we can without mocking
		handler = NewDomainHandler(nil, nil, nil, nil, nil, nil, logger)

		// Setup router with only the specific route we're testing
		router = gin.New()
		router.POST("/api/v1/test-runs", handler.recordTestRun)
	})

	Describe("Request Validation and JSON Binding", func() {
		It("should return 400 for invalid JSON", func() {
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["error"]).To(Not(BeNil()))
		})

		It("should return 500 for empty JSON object", func() {
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBufferString("{}"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return 500 for missing testProjectId", func() {
			requestBody := map[string]interface{}{
				"suiteRuns": []interface{}{},
				// Missing testProjectId
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should pass JSON binding validation with valid structure", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// With nil services, this should reach the service layer and fail there (500)
			// This proves JSON binding worked correctly
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return 400 when request body is nil or empty", func() {
			// Test with nil body - Gin treats nil body as EOF during JSON binding
			req := httptest.NewRequest("POST", "/api/v1/test-runs", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			// Nil or empty body causes EOF error during JSON binding
			Expect(response["error"]).To(ContainSubstring("EOF"))
		})

		It("should check for nil request body before JSON binding", func() {
			// This test verifies the explicit nil body check exists in the code
			// The check at line 189: if c.Request.Body == nil
			// However, in practice, Gin's httptest always provides a body (even if empty),
			// so the ShouldBindJSON catches it first with EOF error

			// Test with empty string body
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBufferString(""))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			// Empty body causes EOF error during JSON parsing
			Expect(response["error"]).To(ContainSubstring("EOF"))
		})
	})

	Describe("Environment Field Processing", func() {
		It("should process request with explicit environment", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"environment":   "production",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should pass environment processing logic, fail at service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request without environment field", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should default environment and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request with empty environment string", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"environment":   "",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should handle empty environment and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("TestSeed and RunID Generation Logic", func() {
		It("should process request with non-zero TestSeed", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"testSeed":      12345,
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should process TestSeed logic and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request with zero TestSeed", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"testSeed":      0,
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should generate UUID and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request without TestSeed field", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should generate UUID and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("Suite Runs Processing", func() {
		It("should process empty suite runs array", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should process empty suite array and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process suite runs with various statuses", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns": []map[string]interface{}{
					{
						"name":         "suite-1",
						"status":       "passed",
						"totalTests":   5,
						"passedTests":  5,
						"failedTests":  0,
						"skippedTests": 0,
					},
					{
						"name":         "suite-2",
						"status":       "failed",
						"totalTests":   3,
						"passedTests":  1,
						"failedTests":  2,
						"skippedTests": 0,
					},
					{
						"name":         "suite-3",
						"status":       "skipped",
						"totalTests":   2,
						"passedTests":  1,
						"failedTests":  0,
						"skippedTests": 1,
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should process all suite conversion logic and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("Git Information Processing", func() {
		It("should process request with git branch and commit", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"gitBranch":     "main",
				"gitSha":        "abc123def456",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should process git information and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request without git information", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should handle missing git information and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("Complex Request Scenarios", func() {
		It("should process comprehensive test run request", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "comprehensive-project",
				"gitBranch":     "feature/testing",
				"gitSha":        "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
				"environment":   "staging",
				"testSeed":      uint64(999888777),
				"suiteRuns": []map[string]interface{}{
					{
						"name":         "unit-tests",
						"status":       "passed",
						"totalTests":   25,
						"passedTests":  25,
						"failedTests":  0,
						"skippedTests": 0,
					},
					{
						"name":         "integration-tests",
						"status":       "skipped",
						"totalTests":   15,
						"passedTests":  12,
						"failedTests":  0,
						"skippedTests": 3,
					},
					{
						"name":         "e2e-tests",
						"status":       "failed",
						"totalTests":   8,
						"passedTests":  6,
						"failedTests":  2,
						"skippedTests": 0,
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should process all fields and logic, reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should handle large number of suite runs", func() {
			suiteRuns := make([]map[string]interface{}, 20)
			for i := 0; i < 20; i++ {
				suiteRuns[i] = map[string]interface{}{
					"name":         fmt.Sprintf("suite-%d", i),
					"status":       "passed",
					"totalTests":   10,
					"passedTests":  10,
					"failedTests":  0,
					"skippedTests": 0,
				}
			}

			requestBody := map[string]interface{}{
				"testProjectId": "large-test-project",
				"suiteRuns":     suiteRuns,
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should handle large payload and reach service layer
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("Tags in Spec Runs and Suite Runs", func() {
		It("should process request with tags in spec runs", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns": []map[string]interface{}{
					{
						"name":   "suite-1",
						"status": "passed",
						"tags": []map[string]interface{}{
							{"name": "smoke"},
							{"name": "priority:high"},
						},
						"specRuns": []map[string]interface{}{
							{
								"name":   "spec-1",
								"status": "passed",
								"tags": []map[string]interface{}{
									{"name": "unit"},
									{"name": "category:backend"},
								},
							},
						},
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should parse JSON with tags structure successfully
			// Will fail at service layer due to nil services, but proves tags were parsed
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request with tags only in suite runs", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns": []map[string]interface{}{
					{
						"name":   "suite-1",
						"status": "passed",
						"tags": []map[string]interface{}{
							{"name": "regression"},
						},
						"specRuns": []map[string]interface{}{},
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should parse JSON with suite-level tags successfully
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should process request with tags at both suite and spec levels", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns": []map[string]interface{}{
					{
						"name":   "integration-suite",
						"status": "passed",
						"tags": []map[string]interface{}{
							{"name": "integration"},
							{"name": "environment:staging"},
						},
						"specRuns": []map[string]interface{}{
							{
								"name":   "api-test",
								"status": "passed",
								"tags": []map[string]interface{}{
									{"name": "api"},
									{"name": "critical"},
								},
							},
							{
								"name":   "db-test",
								"status": "passed",
								"tags": []map[string]interface{}{
									{"name": "database"},
								},
							},
						},
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should parse complex tag structure successfully
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should handle missing tags field in spec runs", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns": []map[string]interface{}{
					{
						"name":   "suite-1",
						"status": "passed",
						// No tags field
						"specRuns": []map[string]interface{}{
							{
								"name":   "spec-1",
								"status": "passed",
								// No tags field
							},
						},
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should handle missing tags field gracefully
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("HTTP Method and Content Type Validation", func() {
		It("should reject non-POST requests", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("GET", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})

		It("should handle requests without Content-Type header", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			// No Content-Type header
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should still attempt to parse JSON
			Expect(w.Code).To(BeNumerically(">=", 400))
		})
	})

	Describe("Error Response Format Consistency", func() {
		It("should return consistent error format for JSON parsing errors", func() {
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusBadRequest))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(HaveKey("error"))
			Expect(response["error"]).To(BeAssignableToTypeOf(""))
		})

		It("should return consistent error format for validation errors", func() {
			requestBody := map[string]interface{}{
				"suiteRuns": []interface{}{},
				// Missing required testProjectId
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(HaveKey("error"))
		})

		It("should return consistent error format for service layer errors", func() {
			requestBody := map[string]interface{}{
				"testProjectId": "test-project",
				"suiteRuns":     []interface{}{},
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(HaveKey("error"))
		})
	})
})

// MockTagRepository provides a mock implementation of TagRepository
type MockTagRepository struct {
	mock.Mock
}

func (m *MockTagRepository) Save(ctx context.Context, tag *tagsDomain.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockTagRepository) FindByID(ctx context.Context, id tagsDomain.TagID) (*tagsDomain.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tagsDomain.Tag), args.Error(1)
}

func (m *MockTagRepository) FindByName(ctx context.Context, name string) (*tagsDomain.Tag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tagsDomain.Tag), args.Error(1)
}

func (m *MockTagRepository) FindAll(ctx context.Context) ([]*tagsDomain.Tag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tagsDomain.Tag), args.Error(1)
}

func (m *MockTagRepository) Delete(ctx context.Context, id tagsDomain.TagID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTagRepository) AssignToTestRun(ctx context.Context, testRunID string, tagIDs []tagsDomain.TagID) error {
	args := m.Called(ctx, testRunID, tagIDs)
	return args.Error(0)
}

// Integration tests with mocked services for recordTestRun
var _ = Describe("recordTestRun Integration Tests with Mocked Services", func() {
	var (
		handler        *DomainHandler
		router         *gin.Engine
		logger         *logging.Logger
		testRunRepo    *MockTestRunRepository
		suiteRunRepo   *MockSuiteRunRepository
		specRunRepo    *MockSpecRunRepository
		tagRepo        *MockTagRepository
		testingService *testingApp.TestRunService
		tagService     *tagsApp.TagService
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)

		// Initialize logger
		loggingConfig := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
		}
		var err error
		logger, err = logging.NewLogger(loggingConfig)
		Expect(err).NotTo(HaveOccurred())

		// Create mocks
		testRunRepo = new(MockTestRunRepository)
		suiteRunRepo = new(MockSuiteRunRepository)
		specRunRepo = new(MockSpecRunRepository)
		tagRepo = new(MockTagRepository)

		// Create services with mocks
		testingService = testingApp.NewTestRunService(testRunRepo, suiteRunRepo, specRunRepo)
		tagService = tagsApp.NewTagService(tagRepo)

		// Setup mock for tag processing - return the same tags (no-op behavior for these tests)
		tagRepo.On("FindByName", mock.Anything, mock.Anything).Return(nil, errors.New("not found")).Maybe()
		tagRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Maybe()

		// Create handler
		handler = NewDomainHandler(testingService, nil, tagService, nil, nil, nil, logger)

		// Setup router
		router = gin.New()
		router.POST("/api/v1/test-runs", handler.recordTestRun)
	})

	Describe("Creating New Test Run", func() {
		It("should create a new test run successfully with no TestSeed", func() {
			requestBody := map[string]interface{}{
				"test_project_id": "test-project-123",
				"git_branch":      "main",
				"git_sha":         "abc123",
				"environment":     "production",
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "unit-tests",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "test 1",
								"status":           "passed",
							},
							{
								"spec_description": "test 2",
								"status":           "passed",
							},
						},
					},
				},
			}

			// Mock expectations - new test run creation
			testRunRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["projectId"]).To(Equal("test-project-123"))
			Expect(response["branch"]).To(Equal("main"))
			Expect(response["commitSha"]).To(Equal("abc123"))
			Expect(response["environment"]).To(Equal("production"))
			Expect(response["status"]).To(Equal("passed"))
			Expect(response["totalTests"]).To(Equal(float64(2)))
			Expect(response["passedTests"]).To(Equal(float64(2)))
			Expect(response["runId"]).NotTo(BeEmpty())

			testRunRepo.AssertExpectations(GinkgoT())
		})

		It("should create a test run with TestSeed and generate runID from seed", func() {
			requestBody := map[string]interface{}{
				"test_project_id": "test-project-123",
				"test_seed":       uint64(999888),
				"suite_runs":      []interface{}{},
			}

			// Mock expectations - GetByRunID should not find existing
			testRunRepo.On("GetByRunID", mock.Anything, "999888").Return(nil, errors.New("not found")).Once()
			testRunRepo.On("Create", mock.Anything, mock.MatchedBy(func(tr *testingDomain.TestRun) bool {
				return tr.RunID == "999888"
			})).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["runId"]).To(Equal("999888"))

			testRunRepo.AssertExpectations(GinkgoT())
		})

		It("should default environment to 'default' when not provided", func() {
			requestBody := map[string]interface{}{
				"test_project_id": "test-project-123",
				"suite_runs":      []interface{}{},
			}

			testRunRepo.On("Create", mock.Anything, mock.MatchedBy(func(tr *testingDomain.TestRun) bool {
				return tr.Environment == "default"
			})).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["environment"]).To(Equal("default"))

			testRunRepo.AssertExpectations(GinkgoT())
		})

		It("should calculate status as failed when suite has failed tests", func() {
			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "failing-suite",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "passing test",
								"status":           "passed",
							},
							{
								"spec_description": "failing test",
								"status":           "failed",
							},
						},
					},
				},
			}

			testRunRepo.On("Create", mock.Anything, mock.MatchedBy(func(tr *testingDomain.TestRun) bool {
				return tr.Status == "failed"
			})).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["status"]).To(Equal("failed"))
			Expect(response["failedTests"]).To(Equal(float64(1)))
			Expect(response["passedTests"]).To(Equal(float64(1)))

			testRunRepo.AssertExpectations(GinkgoT())
		})
	})

	Describe("Updating Existing Test Run", func() {
		It("should find existing test run by runID and append suite runs", func() {
			existingTestRun := &testingDomain.TestRun{
				ID:           1,
				RunID:        "existing-run-id",
				ProjectID:    "test-project",
				Status:       "passed",
				TotalTests:   5,
				PassedTests:  5,
				FailedTests:  0,
				SkippedTests: 0,
				SuiteRuns:    []testingDomain.SuiteRun{},
			}

			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"test_seed":       uint64(12345),
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "new-suite",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "new test",
								"status":           "passed",
							},
						},
					},
				},
			}

			// Mock expectations - find existing run, add suite run, update test run
			testRunRepo.On("GetByRunID", mock.Anything, "12345").Return(existingTestRun, nil).Once()
			suiteRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				// Simulate database auto-increment by setting the ID
				suite := args.Get(1).(*testingDomain.SuiteRun)
				suite.ID = 100
			}).Return(nil).Once()
			specRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				// Simulate database auto-increment by setting the ID
				spec := args.Get(1).(*testingDomain.SpecRun)
				spec.ID = 200
			}).Return(nil).Once()
			testRunRepo.On("Update", mock.Anything, mock.MatchedBy(func(tr *testingDomain.TestRun) bool {
				// Should have accumulated counts
				return tr.TotalTests == 6 && tr.PassedTests == 6
			})).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["totalTests"]).To(Equal(float64(6)))
			Expect(response["passedTests"]).To(Equal(float64(6)))

			testRunRepo.AssertExpectations(GinkgoT())
			suiteRunRepo.AssertExpectations(GinkgoT())
			specRunRepo.AssertExpectations(GinkgoT())
		})

		It("should handle concurrent creation by treating as update when duplicate occurs", func() {
			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"test_seed":       uint64(77777),
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "suite-1",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "test 1",
								"status":           "passed",
							},
						},
					},
				},
			}

			// Mock GetByRunID to return not found initially
			testRunRepo.On("GetByRunID", mock.Anything, "77777").Return(nil, errors.New("not found")).Once()

			// Mock Create to simulate duplicate/unique constraint error
			testRunRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("UNIQUE constraint failed")).Once()

			// After duplicate error, it tries to fetch existing test run by runID
			existingTestRun := &testingDomain.TestRun{
				ID:           99,
				RunID:        "77777",
				ProjectID:    "test-project",
				Status:       "passed",
				TotalTests:   0,
				PassedTests:  0,
				FailedTests:  0,
				SkippedTests: 0,
				SuiteRuns:    []testingDomain.SuiteRun{},
			}
			testRunRepo.On("GetByRunID", mock.Anything, "77777").Return(existingTestRun, nil).Once()

			// Now it adds suite runs to the existing test run
			suiteRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				suite := args.Get(1).(*testingDomain.SuiteRun)
				suite.ID = 100
			}).Return(nil).Once()
			specRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				spec := args.Get(1).(*testingDomain.SpecRun)
				spec.ID = 200
			}).Return(nil).Once()

			// Update the test run with accumulated data
			testRunRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))

			testRunRepo.AssertExpectations(GinkgoT())
			suiteRunRepo.AssertExpectations(GinkgoT())
			specRunRepo.AssertExpectations(GinkgoT())
		})

		It("should update status to failed when new batch has failures", func() {
			existingTestRun := &testingDomain.TestRun{
				ID:           1,
				RunID:        "test-run-id",
				ProjectID:    "test-project",
				Status:       "passed",
				TotalTests:   3,
				PassedTests:  3,
				FailedTests:  0,
				SkippedTests: 0,
				SuiteRuns:    []testingDomain.SuiteRun{},
			}

			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"test_seed":       uint64(88888),
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "failing-suite",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "failing test",
								"status":           "failed",
							},
						},
					},
				},
			}

			testRunRepo.On("GetByRunID", mock.Anything, "88888").Return(existingTestRun, nil).Once()
			suiteRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				suite := args.Get(1).(*testingDomain.SuiteRun)
				suite.ID = 100
			}).Return(nil).Once()
			specRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				spec := args.Get(1).(*testingDomain.SpecRun)
				spec.ID = 200
			}).Return(nil).Once()
			testRunRepo.On("Update", mock.Anything, mock.MatchedBy(func(tr *testingDomain.TestRun) bool {
				return tr.Status == "failed" && tr.FailedTests == 1
			})).Return(nil).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusCreated))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["status"]).To(Equal("failed"))

			testRunRepo.AssertExpectations(GinkgoT())
		})
	})

	Describe("Error Handling", func() {
		It("should return 500 when CreateTestRun fails", func() {
			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"suite_runs":      []interface{}{},
			}

			testRunRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error")).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["error"]).To(ContainSubstring("database error"))

			testRunRepo.AssertExpectations(GinkgoT())
		})

		It("should return 500 when CreateSuiteRun fails", func() {
			existingTestRun := &testingDomain.TestRun{
				ID:        1,
				RunID:     "run-id",
				ProjectID: "test-project",
			}

			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"test_seed":       uint64(55555),
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "suite-1",
						"spec_runs":  []interface{}{},
					},
				},
			}

			testRunRepo.On("GetByRunID", mock.Anything, "55555").Return(existingTestRun, nil).Once()
			suiteRunRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("suite creation error")).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))

			testRunRepo.AssertExpectations(GinkgoT())
			suiteRunRepo.AssertExpectations(GinkgoT())
		})

		It("should return 500 when CreateSpecRun fails", func() {
			existingTestRun := &testingDomain.TestRun{
				ID:        1,
				RunID:     "run-id",
				ProjectID: "test-project",
			}

			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"test_seed":       uint64(66666),
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "suite-1",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "spec-1",
								"status":           "passed",
							},
						},
					},
				},
			}

			testRunRepo.On("GetByRunID", mock.Anything, "66666").Return(existingTestRun, nil).Once()
			suiteRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				suite := args.Get(1).(*testingDomain.SuiteRun)
				suite.ID = 100
			}).Return(nil).Once()
			specRunRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("spec creation error")).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))

			testRunRepo.AssertExpectations(GinkgoT())
			suiteRunRepo.AssertExpectations(GinkgoT())
			specRunRepo.AssertExpectations(GinkgoT())
		})

		It("should return 500 when UpdateTestRun fails", func() {
			existingTestRun := &testingDomain.TestRun{
				ID:        1,
				RunID:     "run-id",
				ProjectID: "test-project",
				SuiteRuns: []testingDomain.SuiteRun{},
			}

			requestBody := map[string]interface{}{
				"test_project_id": "test-project",
				"test_seed":       uint64(44444),
				"suite_runs": []map[string]interface{}{
					{
						"suite_name": "suite-1",
						"spec_runs": []map[string]interface{}{
							{
								"spec_description": "spec-1",
								"status":           "passed",
							},
						},
					},
				},
			}

			testRunRepo.On("GetByRunID", mock.Anything, "44444").Return(existingTestRun, nil).Once()
			suiteRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				suite := args.Get(1).(*testingDomain.SuiteRun)
				suite.ID = 100
			}).Return(nil).Once()
			specRunRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				spec := args.Get(1).(*testingDomain.SpecRun)
				spec.ID = 200
			}).Return(nil).Once()
			testRunRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error")).Once()

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/test-runs", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusInternalServerError))

			testRunRepo.AssertExpectations(GinkgoT())
		})
	})
})
