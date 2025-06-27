# Domain-Driven Design Structure

This directory contains the domain-driven implementation of Fern Platform, organized by business domains rather than technical layers.

## Structure

Each domain follows a hexagonal architecture pattern with these layers:

```
domain/
├── domain/          # Core business logic and entities
├── application/     # Use cases and application services
├── infrastructure/  # External integrations (DB, APIs, etc.)
└── interfaces/      # Adapters (HTTP handlers, GraphQL resolvers)
```

## Domains

### Testing Domain (`/testing`)
Core domain responsible for test execution tracking and analysis.

**Entities:**
- TestRun
- SuiteRun
- SpecRun
- FlakyTest

**Use Cases:**
- Record test run results
- Analyze test flakiness
- Query test history
- Generate test reports

### Projects Domain (`/projects`)
Manages project configuration and permissions.

**Entities:**
- Project
- ProjectPermission
- Team

**Use Cases:**
- Create/update projects
- Manage project permissions
- Associate projects with teams
- Query project statistics

### Auth Domain (`/auth`)
Handles authentication and authorization.

**Entities:**
- User
- UserGroup
- UserScope
- Session

**Use Cases:**
- OAuth authentication
- Permission checking
- Session management
- User profile management

### Analytics Domain (`/analytics`)
Future domain for advanced analytics and AI features.

**Planned Entities:**
- TestTrend
- FailurePattern
- Recommendation

**Planned Use Cases:**
- Analyze test trends
- Detect failure patterns
- Generate insights
- AI-powered recommendations

## Design Principles

1. **Domain Isolation**: Each domain is self-contained with its own models and logic
2. **Dependency Rule**: Dependencies point inward (interfaces → application → domain)
3. **Interface Segregation**: Each layer exposes only what's needed by outer layers
4. **Backward Compatibility**: All changes maintain existing API contracts

## Migration Strategy

The existing code is being gradually migrated to this structure:
- Phase 1: Create domain structure (current)
- Phase 2: Extract domain entities and value objects
- Phase 3: Move business logic to domain services
- Phase 4: Implement application services (use cases)
- Phase 5: Adapt existing handlers/resolvers as interfaces

During migration, the old structure remains functional to ensure zero downtime and backward compatibility.