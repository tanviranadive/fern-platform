# Fern Platform Acceptance Tests

This directory contains comprehensive acceptance tests for the Fern Platform, implementing outer loop TDD best practices with a focus on user workflows and system integration testing.

## 🏗️ Architecture

The acceptance tests are organized using the **Page Object Model** and **Feature-driven** approach, ensuring maintainability and clear separation of concerns.

```
acceptance/
├── README.md                    # This file
├── jest.config.js              # Jest configuration for acceptance tests
├── setup/                      # Test environment setup and utilities
│   ├── test-environment.ts     # Custom Jest environment
│   ├── cluster-manager.ts      # Kubernetes cluster management
│   ├── data-fixtures.ts        # Test data generation and management
│   └── test-helpers.ts         # Common test utilities
├── fixtures/                   # Static test data and configurations
│   ├── test-data/              # SQL fixtures and seed data
│   ├── k8s-manifests/          # Kubernetes deployment manifests
│   └── config/                 # Environment configurations
├── features/                   # Feature-based test organization
│   ├── ui/                     # Frontend acceptance tests
│   ├── api/                    # Backend API acceptance tests
│   └── integration/            # Cross-system integration tests
└── utils/                      # Shared utilities and helpers
    ├── page-objects/           # Page Object Model implementations
    ├── api-clients/            # API client abstractions
    └── assertions/             # Custom matchers and assertions
```

## 🎯 Test Categories

### UI Feature Tests (`features/ui/`)
Testing the complete user experience across all frontend features:

- **Test Results Dashboard** - Viewing, filtering, and analyzing test runs
- **Test Summaries** - Historical view and trend analysis
- **User Preferences** - Customization and personalization
- **AI Chatbot** - Conversational test insights
- **Navigation & Routing** - SPA navigation and deep linking

### API Feature Tests (`features/api/`)
Testing backend services and their contracts:

- **Test Data Management** - CRUD operations for test runs and suites
- **GraphQL API** - Query validation and performance
- **REST Endpoints** - HTTP API contract testing
- **Authentication** - User management and authorization
- **Data Integrity** - Database consistency and migrations

### Integration Tests (`features/integration/`)
Testing cross-system workflows and data flows:

- **End-to-End Workflows** - Complete user journeys
- **Data Pipeline** - Test result ingestion and processing
- **AI Intelligence** - LLM integration and insights generation
- **Real-time Features** - WebSocket communication and live updates
- **Performance** - Load testing and scalability validation

## 🧪 Testing Philosophy

### Outer Loop TDD Approach
1. **Red**: Write failing acceptance test describing desired behavior
2. **Green**: Implement minimum code to make test pass
3. **Refactor**: Improve implementation while keeping tests green
4. **Repeat**: Continue cycle for next feature

### Test Pyramid Strategy
- **70% Unit Tests** (in individual service repositories)
- **20% Integration Tests** (API contracts and service interactions)
- **10% E2E Tests** (Critical user journeys only)

### Quality Gates
- All acceptance tests must pass before deployment
- Performance benchmarks must be met
- Security scans must pass
- API contract compliance validated

## 🚀 Running Tests

### Prerequisites
```bash
# Required tools
- Node.js 18+
- Docker & Docker Compose
- kubectl
- k3d (for local Kubernetes)
```

### Local Development
```bash
# Start test environment
npm run test:acceptance:setup

# Run all acceptance tests
npm run test:acceptance

# Run specific feature tests
npm run test:acceptance -- --testPathPattern=ui
npm run test:acceptance -- --testPathPattern=api
npm run test:acceptance -- --testPathPattern=integration

# Run with watch mode
npm run test:acceptance:watch

# Cleanup test environment
npm run test:acceptance:teardown
```

### CI/CD Pipeline
```bash
# Smoke tests (fast feedback)
npm run test:acceptance:smoke

# Full acceptance test suite
npm run test:acceptance:ci

# Performance regression tests
npm run test:acceptance:performance
```

## 📊 Test Coverage Requirements

### Functional Coverage
- **Critical User Paths**: 100% coverage
- **API Contracts**: 100% coverage
- **Error Scenarios**: 90% coverage
- **Integration Points**: 95% coverage

### Performance Benchmarks
- **Page Load Time**: < 2 seconds (95th percentile)
- **API Response Time**: < 500ms (95th percentile)
- **Database Queries**: < 100ms (95th percentile)
- **Memory Usage**: < 512MB per service

### Browser Compatibility
- **Chrome**: Latest 2 versions
- **Firefox**: Latest 2 versions
- **Safari**: Latest 2 versions
- **Edge**: Latest 2 versions

## 🔧 Configuration

### Environment Variables
```bash
# Test environment
TEST_ENV=acceptance
K8S_NAMESPACE=fern-acceptance-test
DATABASE_URL=postgresql://test:test@localhost:5432/fern_test

# Service endpoints
FERN_REPORTER_URL=http://localhost:8080
FERN_MYCELIUM_URL=http://localhost:8081
FERN_UI_URL=http://localhost:3000

# External integrations
ANTHROPIC_API_KEY=test-key
OPENAI_API_KEY=test-key
```

### Test Data Management
- **Isolation**: Each test suite runs in isolated namespace
- **Cleanup**: Automatic cleanup after test completion
- **Fixtures**: Repeatable test data generation
- **Snapshots**: Database state snapshots for debugging

## 🐛 Debugging

### Test Failures
```bash
# Run tests with debug output
DEBUG=fern:* npm run test:acceptance

# Generate test reports
npm run test:acceptance:report

# Screenshot capture on failure
CAPTURE_SCREENSHOTS=true npm run test:acceptance
```

### Performance Issues
```bash
# Profile test execution
npm run test:acceptance:profile

# Memory leak detection
npm run test:acceptance:memory-check
```

## 📈 Metrics and Reporting

### Test Execution Metrics
- Test execution time trends
- Flaky test identification
- Coverage reports
- Performance regression tracking

### Quality Metrics
- Bug escape rate
- Test maintenance overhead
- Feature delivery velocity
- System reliability scores

## 🔄 Continuous Improvement

### Test Health Monitoring
- Regular review of test execution times
- Flaky test identification and resolution
- Test coverage gap analysis
- Performance benchmark updates

### Best Practices Evolution
- Regular retrospectives on testing approach
- Tool and framework updates
- Team knowledge sharing
- Industry best practice adoption

## 📚 Resources

- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Testing Library](https://testing-library.com/)
- [Page Object Model](https://martinfowler.com/bliki/PageObject.html)
- [Kubernetes Testing](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-application/)