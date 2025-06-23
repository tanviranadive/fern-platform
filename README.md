<div align="center">
  <img src="https://github.com/guidewire-oss/fern-reporter/blob/main/docs/images/logo-color.png" alt="Fern Platform" width="200"/>
  
  # Fern Platform

  A unified platform for test reporting, analysis, and AI-powered insights that consolidates the Fern ecosystem into a modern, scalable architecture.

  [![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
  [![Coverage](https://img.shields.io/badge/coverage-85%25-green.svg)](#)
</div>

## Overview

Fern Platform brings together the capabilities of multiple Fern projects into a single, cohesive platform:

- **fern-reporter**: Test run data collection and reporting
- **fern-mycelium**: AI-powered test analysis and insights  
- **fern-ui**: Modern React-based user interface built with Refine.dev

## Architecture

The platform follows a unified monolithic architecture with modular components:

- **Shared Infrastructure Layer**: Common database models, logging, configuration, and middleware in `pkg/`
- **Core Modules**: Modular components (reporter, mycelium, ui) in `internal/`
- **API Layer**: GraphQL and REST APIs with standardized patterns
- **Deployment Layer**: KubeVela-based orchestration for local and production environments
- **Command Layer**: Main application entry point in `cmd/`

### Technology Stack

- **Backend**: Go with Gin framework, GORM ORM, GraphQL (gqlgen)
- **Frontend**: React with Refine.dev framework, TypeScript
- **Database**: PostgreSQL with CloudNativePG (CNPG) operator
- **Caching**: Redis for sessions and message bus
- **Testing**: Ginkgo/Gomega for Go, Jest for frontend
- **Deployment**: Kubernetes with KubeVela application management
- **AI Integration**: Anthropic Claude, OpenAI, HuggingFace, Ollama support

## Quick Start

### Prerequisites

- **Go 1.21+** - For building the platform
- **Docker** - For dependencies
- **kubectl** - Kubernetes CLI
- **k3d** - Local Kubernetes cluster (optional)
- **PostgreSQL** - Database (deployed via k3d cluster)

### Local Development Setup

For comprehensive local development setup with k3d, KubeVela, and all dependencies, please follow the detailed instructions in our [Contributing Guide](CONTRIBUTING.md#local-deployment-with-kubevela).

**Quick Overview:**
1. **Clone the repository**
2. **Set up k3d cluster** with KubeVela and CloudNativePG
3. **Deploy the platform** using KubeVela
4. **Access via port-forward** at http://localhost:8080

### Development Workflow

See [Contributing Guide](CONTRIBUTING.md#development-workflow) for detailed development workflow including:
- Building and testing the platform
- Running acceptance tests in k3d
- Code quality checks
- Pull request process

## Configuration

Configuration is managed through YAML files and environment variables following the twelve-factor app methodology.

### Environment Variables

Key environment variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=fern_platform

# Authentication (optional)
AUTH_ENABLED=false
JWT_SECRET=your_jwt_secret

# LLM Providers
ANTHROPIC_API_KEY=your_anthropic_key
OPENAI_API_KEY=your_openai_key

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## API Documentation

### REST API

The platform provides comprehensive REST APIs:

- **Test Runs**: `/api/v1/test-runs/*`
- **Projects**: `/api/v1/projects/*`
- **Tags**: `/api/v1/tags/*`
- **Health**: `/health`

### GraphQL API

GraphQL endpoint available at `/graphql` with playground at `/graphql`.

Key types:
- `TestRun`: Test execution data
- `Project`: Project configuration
- `Tag`: Test categorization
- `FlakyTest`: Flaky test analysis

## Database Schema

The platform uses PostgreSQL with the following main tables:

- `test_runs`: Test execution records
- `suite_runs`: Test suite executions  
- `spec_runs`: Individual test specs
- `projects`: Project configurations
- `tags`: Test categorization
- `flaky_tests`: Flaky test analysis

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run service-specific tests
make test-reporter
```

### Acceptance Tests

Comprehensive acceptance tests using Jest and custom Kubernetes test environment:

```bash
# Run acceptance tests
make test-acceptance
```

### Test Coverage

The acceptance tests cover:
- UI functionality and user workflows
- API endpoints and data integrity
- Integration between services
- Performance and error scenarios

## Deployment

### Local Development

For complete local deployment instructions with k3d and KubeVela, see our [Contributing Guide](CONTRIBUTING.md#local-deployment-with-kubevela).

Quick commands:
```bash
# Complete k3d cluster setup
make cluster-setup

# Deploy application
kubectl apply -f deployments/fern-platform-kubevela.yaml
```

### Production

Deploy to production Kubernetes cluster:

```bash
VERSION=v1.0.0 make prod-deploy
```

### KubeVela Applications

The platform uses KubeVela applications for:
- Infrastructure orchestration (PostgreSQL, Redis)
- Service deployment and configuration
- Environment-specific policies
- Workflow management with dependencies

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for detailed instructions on:

- 🚀 Setting up your development environment
- ☸️ Local deployment with KubeVela and k3d
- 🧪 Running acceptance tests
- 📝 Code style guidelines
- 🔄 Pull request process

### Quick Start for Contributors

1. **Fork** and clone the repository
2. **Follow** the [local deployment guide](CONTRIBUTING.md#local-deployment-with-kubevela)
3. **Make** your changes with tests
4. **Submit** a pull request

For questions, open a [GitHub Discussion](../../discussions) or [Issue](../../issues).

## Architecture Decisions

### Layered Architecture

The platform uses a layered architecture for extensibility:

1. **Infrastructure Layer**: Shared utilities, database, logging
2. **Repository Layer**: Data access with GORM
3. **Service Layer**: Business logic and domain operations
4. **API Layer**: GraphQL and REST endpoints
5. **Presentation Layer**: React UI with Refine.dev

### Modular Design

Each internal module has clear responsibilities:
- **reporter**: Data ingestion and reporting (`internal/reporter/`)
- **mycelium**: AI analysis and insights (`internal/mycelium/`)
- **ui**: User interface components (`internal/ui/`)

### Event-Driven Design

Modules communicate through:
- Redis Streams for real-time events
- Database-level triggers for data consistency
- Internal Go interfaces for synchronous operations

## Monitoring and Observability

- **Health Checks**: `/health` endpoint on all services
- **Metrics**: Prometheus-compatible metrics
- **Logging**: Structured JSON logging with correlation IDs
- **Tracing**: OpenTelemetry support (configurable)

## Security

- **Authentication**: JWT-based with JWKS support
- **Authorization**: Role-based access control
- **Network**: Service mesh compatible
- **Secrets**: Kubernetes secrets integration

## License

[License information]

## Support

For issues and questions:
- GitHub Issues: [repository-url]/issues
- Documentation: [docs-url]
- Community: [community-url]