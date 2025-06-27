# Fern Platform Architecture

<div align="center">
  <img src="https://github.com/guidewire-oss/fern-platform/blob/main/docs/images/logo-color.png" alt="Fern Platform" width="150"/>
</div>

## Project Structure

The Fern Platform follows Domain-Driven Design (DDD) principles with standard Go project layout conventions:

```
fern-platform/
├── cmd/                          # Application entry points
│   └── fern-platform/           # Main platform binary
│       └── main.go
├── internal/                     # Private application code
│   ├── domains/                 # Domain-driven design structure
│   │   ├── testing/             # Testing domain (test runs, suites, specs)
│   │   │   ├── domain/          # Core business entities and logic
│   │   │   ├── application/     # Use cases and application services
│   │   │   ├── infrastructure/  # External integrations (DB, APIs)
│   │   │   └── interfaces/      # Adapters (HTTP, GraphQL)
│   │   ├── projects/            # Projects domain (projects, permissions)
│   │   │   ├── domain/          # Core business entities
│   │   │   ├── application/     # Use cases
│   │   │   ├── infrastructure/  # Persistence
│   │   │   └── interfaces/      # API adapters
│   │   ├── auth/                # Authentication domain
│   │   └── analytics/           # Analytics domain (future)
│   ├── reporter/                # Legacy structure (being migrated)
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

### 2. Domain-Driven Design (DDD)

The platform is organized around business domains following hexagonal architecture:

```
internal/domains/{domain}/
├── domain/          # Core business logic (entities, value objects, domain services)
├── application/     # Use cases and application services
├── infrastructure/  # External integrations (database, external APIs)
└── interfaces/      # Adapters for inbound ports (HTTP, GraphQL, CLI)
```

**Key Domains:**
- **Testing**: Test execution tracking, suite management, flaky test detection
- **Projects**: Project configuration, team ownership, permissions
- **Auth**: User authentication, authorization, session management
- **Analytics**: Test insights, trends, AI-powered analysis (future)

**Benefits:**
- Business logic is isolated from infrastructure
- Easy to test domain logic without external dependencies
- Clear boundaries between domains
- Supports future microservices extraction if needed

### 3. Package Organization

#### `cmd/` - Application Entry Points
Contains the main application binaries. Currently includes:
- `fern-platform`: The main platform service

#### `internal/` - Private Application Code
Contains code that is specific to this application and should not be imported by other projects:

- **`domains/`**: Domain-driven design structure with business domains
  - **`testing/`**: Test execution, suites, specs, flaky test detection
  - **`projects/`**: Project management, permissions, team ownership
  - **`auth/`**: Authentication and authorization logic
  - **`analytics/`**: Test analytics and insights (planned)
- **`reporter/`**: Legacy structure being migrated to domains
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

### 4. Data Flow with DDD

```
Client Request
    ↓
HTTP Router (Gin)
    ↓
Middleware Chain (Auth, Logging, etc.)
    ↓
Interface Adapters (REST/GraphQL Handlers)
    ↓
Application Services (Use Cases)
    ↓
Domain Layer (Business Logic)
    ↓
Infrastructure Layer (Repositories)
    ↓
Database (PostgreSQL)
```

**Example: Creating a Test Run**
1. HTTP POST request to `/api/v1/test-runs`
2. Auth middleware validates JWT token
3. REST handler validates request format
4. `RecordTestRunHandler` (application service) orchestrates the use case
5. `TestRun` domain entity enforces business rules
6. `GormTestRunRepository` persists to PostgreSQL
7. Response returned through the same layers

### 5. Domain Communication

Domains communicate through well-defined interfaces and domain events:

```go
// Domain repository interfaces (in domain layer)
type TestRunRepository interface {
    Save(ctx context.Context, testRun *TestRun) error
    FindByID(ctx context.Context, id TestRunID) (*TestRun, error)
}

// Application service interfaces (use cases)
type RecordTestRunHandler interface {
    Handle(ctx context.Context, cmd RecordTestRunCommand) (*TestRunSnapshot, error)
}

// Domain entities enforce business rules
type TestRun struct {
    id        TestRunID
    projectID string
    status    TestRunStatus
    // ... other fields
}

func (tr *TestRun) Complete() error {
    if tr.status != TestRunStatusRunning {
        return errors.New("can only complete a running test")
    }
    // Business logic here
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

### GraphQL API

GraphQL endpoint at `/graphql` provides rich querying capabilities:

```graphql
# Example GraphQL query
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

GraphQL schema is defined in `internal/reporter/graphql/schema.graphql` with resolver implementations in `internal/reporter/graphql/`.

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