<div align="center">
  <img src="https://github.com/guidewire-oss/fern-reporter/blob/main/docs/images/logo-color.png" alt="Fern Platform" width="150"/>
  
  # Fern Platform - Go Acceptance Tests

  A comprehensive Go-based acceptance test suite for the Fern Platform, built with Ginkgo/Gomega.
</div>

## Overview

This test suite provides end-to-end acceptance testing for the Fern Platform, covering:

- **API Testing**: REST and GraphQL API validation
- **Integration Testing**: Cross-service workflows and data consistency
- **UI Testing**: Browser-based functional testing with Chromedp
- **Performance Testing**: Response time and scalability validation

## Architecture

The test suite follows a layered architecture:

```
acceptance-go/
├── specs/                 # Test specifications
│   ├── api/              # API acceptance tests
│   ├── integration/      # End-to-end integration tests
│   └── ui/               # UI acceptance tests
├── pkg/                  # Shared packages
│   ├── clients/          # API clients (REST, GraphQL)
│   ├── pages/            # Page objects for UI testing
│   ├── fixtures/         # Test data management
│   ├── k8s/              # Kubernetes and KubeVela utilities
│   └── utils/            # Test utilities and custom matchers
└── config/               # Test configuration
```

## Prerequisites

### Required Software

- **Go 1.21+**
- **kubectl** - Kubernetes CLI
- **vela** - KubeVela CLI
- **k3d** - Local Kubernetes cluster
- **Chrome/Chromium** - For UI testing

### Cluster Requirements

The tests assume a pre-configured k3d cluster with:

- **KubeVela** installed and configured
- **CloudNativePG (CNPG)** operator installed
- **Component definitions** available:
  - `postgres` - PostgreSQL database
  - `gateway` - HTTP gateway/ingress
  - Additional components as defined in KubeVela application

## Quick Start

### 1. Install Dependencies

```bash
make deps
```

### 2. Verify Cluster Prerequisites

```bash
make verify-cluster
```

### 3. Run All Tests

```bash
make test
```

### 4. Run Specific Test Suites

```bash
# API tests only
make test-api

# Integration tests only
make test-integration

# UI tests only
make test-ui
```

## Test Execution Options

### Parallel Execution

The test suite supports parallel execution across multiple processes:

```bash
# Run with 4 parallel processes (default)
make test

# Run with custom parallelism
make test PARALLEL_PROCESSES=2

# Run UI tests with reduced parallelism (recommended)
make test-ui PARALLEL_PROCESSES=2
```

### Focused Testing

```bash
# Smoke tests for quick feedback
make test-smoke

# Performance tests only
make test-performance

# Fast subset of tests
make test-fast

# Run with specific focus
ginkgo run --focus="GraphQL API" ./specs/api/
```

### Debug Mode

```bash
# Debug API tests
make debug-api

# Debug UI tests
make debug-ui

# Verbose output
make test-verbose
```

## Configuration

### Test Configuration

Configuration is managed through `config/test-config.yaml`:

```yaml
test:
  timeout: 300s
  parallel_processes: 4
  
k8s:
  namespace_prefix: "fern-acceptance"
  wait_timeout: 5m
  
kubevela:
  app_name: "fern-platform-test"
  app_file: "../deployments/fern-platform-local.yaml"
  
services:
  reporter:
    port: 8080
    health_path: "/health"
  ui:
    port: 3000
    health_path: "/"
```

### Environment Variables

Key environment variables for customization:

```bash
# Kubernetes context
export KUBECONFIG=/path/to/kubeconfig
export K8S_CONTEXT=k3d-k3s-default

# Test execution
export TEST_TIMEOUT=30m
export PARALLEL_PROCESSES=4
export VERBOSE=true

# Browser settings (UI tests)
export HEADLESS=true
export VIEWPORT_WIDTH=1920
export VIEWPORT_HEIGHT=1080
```

## Test Suites

### API Acceptance Tests

Located in `specs/api/`, these tests validate:

#### REST API Tests (`rest_endpoints_test.go`)
- Health check endpoints
- CRUD operations for projects and test runs
- Filtering and pagination
- Error handling and validation
- Performance and concurrency
- Security validation

#### GraphQL API Tests (`graphql_test.go`)
- Schema introspection and validation
- Query execution and filtering
- Complex nested queries
- Error handling and validation
- Performance optimization (N+1 prevention)
- Security (SQL injection prevention)

### Integration Tests

Located in `specs/integration/`, these tests cover:

#### End-to-End Workflows (`end_to_end_workflows_test.go`)
- Complete test run lifecycle
- Cross-service data consistency
- Transaction integrity
- Performance under load
- Error recovery and resilience

### UI Acceptance Tests

Located in `specs/ui/`, these tests validate:

#### Dashboard Tests (`dashboard_test.go`)
- Dashboard loading and display
- Statistics and visualizations
- Navigation and responsiveness
- Error handling
- Accessibility compliance

#### Test Runs Page Tests (`testruns_test.go`)
- Data loading and pagination
- Filtering and search functionality
- Row expansion and spec details
- Navigation and deep linking
- Performance and error handling

## Test Environment Management

### Isolated Namespaces

Each test suite execution creates isolated Kubernetes namespaces:

```
fern-api-test-<random-id>-<process-id>
fern-integration-test-<random-id>-<process-id>
fern-ui-test-<random-id>-<process-id>
```

### Application Deployment

Tests use KubeVela applications deployed from `../deployments/fern-platform-local.yaml`:

1. **BeforeSuite**: Deploy application, wait for readiness
2. **Test Execution**: Run tests against deployed services
3. **AfterSuite**: Clean up application and namespace

### Test Data Management

The `fixtures` package provides comprehensive test data:

- **Projects**: 5 test projects with varied configurations
- **Test Runs**: 15-25 test runs per project with realistic data
- **Spec Runs**: 5-15 specs per test run with errors and timing
- **Data Relationships**: Proper foreign key relationships

## Best Practices

### Test Independence

Following Ginkgo best practices, all tests are independent:

- No shared state between tests
- Each test creates its own data
- Proper cleanup in `AfterEach`/`DeferCleanup`
- Parallel execution safe

### Performance Testing

Built-in performance monitoring:

```go
endMeasurement := performanceMonitor.StartMeasurement("api_response")
// ... perform operation
duration := endMeasurement()
Expect(duration).To(BeNumerically("<", 2*time.Second))
```

### Custom Matchers

Comprehensive custom Gomega matchers:

```go
Expect(response).To(HaveValidApiResponse())
Expect(testRun).To(HaveValidTestRunStructure())
Expect(duration).To(BeWithinTimeRange(100*time.Millisecond, 2*time.Second))
```

### Error Handling

Robust error handling and recovery:

- Service availability validation
- Graceful handling of temporary failures
- Meaningful error messages
- Cleanup on test failures

## CI/CD Integration

### Continuous Integration

```bash
# Setup CI environment
make ci-setup

# Run tests in CI
make ci-test

# Cleanup after CI
make ci-cleanup
```

### Test Reports

Multiple output formats supported:

- **JSON Reports**: `test-results/*.json`
- **JUnit XML**: `test-results/*.xml`
- **Console Output**: Colored and formatted

### Docker Support

```bash
# Run tests in Docker container
make docker-test
```

## Troubleshooting

### Common Issues

#### Cluster Not Ready
```bash
# Verify cluster status
make verify-cluster
kubectl cluster-info
```

#### KubeVela Issues
```bash
# Check KubeVela installation
vela version
kubectl get pods -n vela-system
```

#### Test Timeouts
```bash
# Increase timeout
make test TEST_TIMEOUT=45m

# Reduce parallelism
make test PARALLEL_PROCESSES=1
```

#### UI Test Failures
```bash
# Check browser dependencies
# Ensure Chrome/Chromium is available
# Run UI tests with debug output
make debug-ui
```

### Getting Help

```bash
# Show troubleshooting guide
make troubleshoot

# Show environment status
make status

# Show environment variables
make env
```

## Development

### Adding New Tests

1. **API Tests**: Add to `specs/api/`
2. **Integration Tests**: Add to `specs/integration/`
3. **UI Tests**: Add to `specs/ui/`

### Test Structure

Follow Ginkgo best practices:

```go
var _ = Describe("Feature Name", func() {
    var (
        ctx    context.Context
        client = GetClient()
    )

    BeforeEach(func() {
        ctx = GetTestContext()
    })

    Describe("Functionality Group", func() {
        It("should do something specific", func() {
            By("Step description")
            // Test implementation
            Expect(result).To(BeTrue())
        })
    })
})
```

### Adding Page Objects

For UI tests, create page objects in `pkg/pages/`:

```go
type NewPage struct {
    *BasePage
}

func NewNewPage(baseURL string, browserCtx context.Context) *NewPage {
    return &NewPage{
        BasePage: NewBasePage(baseURL, browserCtx),
    }
}
```

### Custom Matchers

Add custom matchers in `pkg/utils/assertions.go`:

```go
func HaveCustomProperty() types.GomegaMatcher {
    return &customPropertyMatcher{}
}
```

## Contributing

1. Follow Go best practices
2. Write descriptive test names
3. Use proper `By()` steps for complex tests
4. Add performance assertions where appropriate
5. Update documentation for new features
6. Ensure tests are independent and parallel-safe

## License

This project follows the same license as the Fern Platform.