# Fern Platform Architecture

<div align="center">
  <img src="https://github.com/guidewire-oss/fern-reporter/blob/main/docs/images/logo-color.png" alt="Fern Platform" width="150"/>
</div>

## Project Structure

The Fern Platform follows standard Go project layout conventions:

```
fern-platform/
├── cmd/                          # Application entry points
│   └── fern-platform/           # Main platform binary
│       └── main.go
├── internal/                     # Private application code
│   ├── reporter/                # Test reporting module
│   │   ├── api/                 # REST API handlers
│   │   ├── graphql/             # GraphQL schema and resolvers
│   │   ├── repository/          # Data access layer
│   │   └── service/             # Business logic layer
│   ├── mycelium/                # AI analysis module (future)
│   └── ui/                      # UI components (future)
├── pkg/                         # Public library code
│   ├── auth/                    # Authentication utilities
│   ├── config/                  # Configuration management
│   ├── database/                # Database connections and models
│   ├── logging/                 # Logging utilities
│   ├── middleware/              # HTTP middleware
│   ├── types/                   # Shared type definitions
│   └── utils/                   # General utilities
├── web/                         # Web UI assets
│   └── index.html              # Single-page application
├── migrations/                  # Database migrations
├── deployments/                 # Kubernetes/KubeVela configurations
├── acceptance-go/               # Go-based acceptance tests
├── acceptance/                  # Legacy JavaScript tests
├── config/                      # Configuration files
└── docs/                        # Documentation
```

## Architecture Principles

### 1. Unified Monolithic Design

The Fern Platform consolidates multiple services into a single, cohesive application:

- **Single Binary**: One deployable artifact (`cmd/fern-platform`)
- **Modular Internal Structure**: Clear separation of concerns within `internal/`
- **Shared Infrastructure**: Common utilities and infrastructure in `pkg/`
- **Integrated APIs**: Both REST and GraphQL endpoints in one service

### 2. Layered Architecture

Each internal module follows a consistent layered approach:

```
internal/reporter/
├── api/           # HTTP handlers and routing
├── graphql/       # GraphQL schema and resolvers  
├── service/       # Business logic and domain operations
└── repository/    # Data access and persistence
```

**Benefits:**
- Clear separation of concerns
- Easy testing and mocking
- Consistent patterns across modules
- Simplified dependency management

### 3. Package Organization

#### `cmd/` - Application Entry Points
Contains the main application binaries. Currently includes:
- `fern-platform`: The main platform service

#### `internal/` - Private Application Code
Contains code that is specific to this application and should not be imported by other projects:

- **`reporter/`**: Test reporting and data collection
- **`mycelium/`**: AI-powered analysis (planned)
- **`ui/`**: Server-side UI components (planned)

#### `pkg/` - Public Library Code
Contains code that could potentially be imported by other projects:

- **`auth/`**: Authentication and authorization
- **`config/`**: Configuration management
- **`database/`**: Database models and connections
- **`logging/`**: Structured logging
- **`middleware/`**: HTTP middleware components
- **`types/`**: Shared data types
- **`utils/`**: General utility functions

### 4. Data Flow

```
Client Request
    ↓
HTTP Router (Gin)
    ↓
Middleware Chain
    ↓
API Handlers (REST/GraphQL)
    ↓
Service Layer
    ↓
Repository Layer
    ↓
Database (PostgreSQL)
```

### 5. Module Communication

Modules communicate through well-defined interfaces:

```go
// Service interfaces for cross-module communication
type TestRunService interface {
    CreateTestRun(ctx context.Context, testRun *TestRun) (*TestRun, error)
    GetTestRuns(ctx context.Context, opts *QueryOptions) ([]*TestRun, error)
}

// Repository interfaces for data access
type TestRunRepository interface {
    Create(ctx context.Context, testRun *TestRun) error
    FindByID(ctx context.Context, id string) (*TestRun, error)
}
```

## Configuration Management

Configuration is centralized in `pkg/config/` and supports:

- **Environment Variables**: For deployment-specific settings
- **Configuration Files**: YAML-based configuration
- **Default Values**: Sensible defaults for all settings
- **Validation**: Configuration validation at startup

Example configuration structure:
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  
database:
  host: "localhost"
  port: 5432
  name: "fern_platform"
  
logging:
  level: "info"
  format: "json"
```

## Database Design

The platform uses PostgreSQL with a normalized schema:

```sql
-- Core entities
test_runs       -- Test execution records
suite_runs      -- Test suite groupings
spec_runs       -- Individual test specifications
projects        -- Project configurations
tags           -- Test categorization
flaky_tests    -- Flaky test analysis
```

Migrations are managed in the `migrations/` directory using Go migrate.

## API Design

### REST API

RESTful endpoints following standard conventions:

```
GET    /api/v1/test-runs           # List test runs
POST   /api/v1/test-runs           # Create test run
GET    /api/v1/test-runs/{id}      # Get specific test run
PUT    /api/v1/test-runs/{id}      # Update test run
DELETE /api/v1/test-runs/{id}      # Delete test run
```

### GraphQL API (Planned)

GraphQL endpoint is planned for future release to provide rich querying capabilities:

```graphql
# Example of planned GraphQL query support
query {
  testRuns(
    projectId: "abc123"
    status: FAILED
    limit: 10
  ) {
    id
    status
    duration
    project {
      name
    }
    specRuns {
      description
      errorMessage
    }
  }
}
```

**Note:** GraphQL schema is defined in `internal/reporter/graphql/schema.graphql` but implementation is pending.

## Testing Strategy

### Unit Tests
- Located alongside source code
- Test individual functions and methods
- Mock external dependencies

### Integration Tests
- Test module interactions
- Use test database
- Validate business workflows

### Acceptance Tests
- End-to-end testing with real environment
- Both Go-based (`acceptance-go/`) and legacy JavaScript (`acceptance/`)
- Cover API, UI, and integration scenarios

## Deployment Architecture

### Local Development
- Single binary deployment
- Docker Compose for dependencies
- Hot reloading for development

### Kubernetes
- KubeVela applications for orchestration
- Separate components for scalability:
  - Application pods
  - PostgreSQL (CloudNativePG)
  - Redis for caching
  - Load balancer/ingress

### Build and Release
- Multi-stage Docker builds
- Semantic versioning
- Automated CI/CD pipeline

## Security Considerations

- **Authentication**: JWT-based with configurable providers
- **Authorization**: Role-based access control
- **Input Validation**: Request validation middleware
- **SQL Injection**: Parameterized queries with GORM
- **CORS**: Configurable cross-origin policies
- **TLS**: Optional HTTPS support

## Performance Optimizations

- **Database Indexing**: Strategic indexes on query patterns
- **Connection Pooling**: Configured database connection pools
- **Caching**: Redis for session and query caching
- **GraphQL**: DataLoader pattern to prevent N+1 queries
- **Pagination**: Cursor-based pagination for large datasets

## Monitoring and Observability

- **Health Checks**: `/health` endpoint with dependency checks
- **Metrics**: Prometheus-compatible metrics
- **Logging**: Structured JSON logging with correlation IDs
- **Tracing**: OpenTelemetry support (configurable)
- **Error Tracking**: Comprehensive error logging and alerting

## Future Enhancements

### Planned Modules

1. **Mycelium (`internal/mycelium/`)**
   - AI-powered test analysis
   - Flaky test detection
   - Performance regression analysis
   - Integration with multiple LLM providers

2. **Advanced UI (`internal/ui/`)**
   - Server-side rendering components
   - Real-time dashboards
   - Interactive analytics

3. **Plugin System**
   - Extensible architecture for custom analyzers
   - Third-party integrations
   - Custom reporting formats

### Scalability Considerations

While currently monolithic, the architecture supports future microservice extraction:

- Clear module boundaries
- Interface-based communication
- Separate database schemas per module
- Event-driven architecture foundation