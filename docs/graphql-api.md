# GraphQL API Documentation

The Fern Platform provides a GraphQL API for efficient data fetching and real-time subscriptions. This API is designed to optimize UI performance by allowing clients to request exactly the data they need in a single query.

## Overview

The GraphQL API is available at `/query` and provides:

- **Efficient data fetching**: Request multiple resources in a single query
- **Type safety**: Strongly typed schema with automatic code generation
- **Performance optimizations**: Built-in DataLoader for N+1 query prevention
- **Real-time updates**: Subscription support for live data
- **Automatic persisted queries**: Reduced bandwidth usage for repeated queries

## Authentication

The GraphQL API uses the same OAuth 2.0 authentication as the REST API. Include your session cookie with all requests:

```javascript
const response = await fetch('/query', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({
        query: '...',
        variables: {}
    })
});
```

## Schema

### Queries

#### Get Current User

```graphql
query GetCurrentUser {
    currentUser {
        id
        email
        name
        firstName
        lastName
        role
        profileUrl
    }
}
```

#### Get Dashboard Data

This query efficiently fetches all data needed for the dashboard in a single request:

```graphql
query GetDashboardData {
    dashboardSummary {
        health {
            status
            service
            timestamp
            version
        }
        projectCount
        activeProjectCount
        totalTestRuns
        recentTestRuns
        overallPassRate
        totalTestsExecuted
        averageTestDuration
    }
    
    projects(first: 10) {
        edges {
            node {
                id
                projectId
                name
                description
                isActive
                stats {
                    totalTestRuns
                    successRate
                    averageDuration
                    lastRunTime
                }
            }
        }
        totalCount
    }
    
    recentTestRuns(limit: 10) {
        id
        runId
        projectId
        branch
        status
        startTime
        duration
        totalTests
        passedTests
        failedTests
    }
}
```

#### Get Treemap Data

Fetch hierarchical data for treemap visualization:

```graphql
query GetTreemapData($projectId: String, $days: Int) {
    treemapData(projectId: $projectId, days: $days) {
        projects {
            project {
                id
                projectId
                name
            }
            suites {
                suite {
                    id
                    suiteName
                    status
                }
                totalDuration
                totalSpecs
                passRate
            }
            totalDuration
            totalTests
            passRate
        }
        totalDuration
        totalTests
        overallPassRate
    }
}
```

#### Get Project Details

```graphql
query GetProjectDetails($projectId: String!) {
    projectByProjectId(projectId: $projectId) {
        id
        projectId
        name
        description
        repository
        defaultBranch
        isActive
        stats {
            totalTestRuns
            successRate
            averageDuration
            flakyTestCount
        }
        recentRuns {
            id
            runId
            branch
            status
            startTime
            duration
            totalTests
            passedTests
            failedTests
        }
    }
    
    testRunStats(projectId: $projectId) {
        totalRuns
        statusCounts {
            status
            count
            percentage
        }
        averageDuration
        successRate
        trendsOverTime {
            date
            totalRuns
            passRate
            averageDuration
        }
    }
    
    flakyTests(filter: { projectId: $projectId }, first: 10) {
        edges {
            node {
                id
                testName
                suiteName
                flakeRate
                totalExecutions
                lastSeenAt
                severity
                status
            }
        }
    }
}
```

### Mutations

Currently, mutations are not implemented. All write operations should continue using the REST API endpoints.

### Subscriptions

Real-time subscriptions are planned for future releases:

```graphql
subscription TestRunUpdates($projectId: String) {
    testRunCreated(projectId: $projectId) {
        id
        runId
        status
        startTime
    }
    
    testRunStatusChanged(projectId: $projectId) {
        id
        runId
        status
        endTime
    }
}
```

## JavaScript Client

The web interface includes a GraphQL client for easy integration:

```javascript
// The client is available globally
const data = await graphqlClient.query(GRAPHQL_QUERIES.GET_DASHBOARD_DATA);

// With variables
const projectData = await graphqlClient.query(
    GRAPHQL_QUERIES.GET_PROJECT_DETAILS,
    { projectId: 'my-project' }
);
```

## Performance Features

### DataLoader Integration

The GraphQL server automatically batches and caches database queries using DataLoader. This prevents N+1 query problems when fetching related data:

- Projects and their stats are fetched in batches
- Test runs for multiple projects are loaded efficiently
- User data is cached per request

### Query Complexity Limits

To prevent abuse, queries have complexity limits:
- Maximum query depth: 10
- Maximum complexity score: 1000
- Rate limiting: 100 requests per minute per user

### Automatic Persisted Queries

The server supports APQ (Automatic Persisted Queries) to reduce bandwidth:

1. Client sends query hash first
2. If server has the query cached, it executes without needing the full query text
3. If not cached, client sends full query and server caches it

This significantly reduces payload size for repeated queries.

## Migration from REST

The GraphQL API is designed to coexist with the REST API. Key differences:

1. **Single Request**: Instead of multiple REST calls, fetch all dashboard data in one GraphQL query
2. **Flexible Fields**: Request only the fields you need
3. **Nested Data**: Get related data (projects with their stats) without additional requests
4. **Type Safety**: The schema provides compile-time type checking with generated types

### Example Migration

REST approach (3 requests):
```javascript
const health = await fetch('/health');
const projects = await fetch('/api/v1/projects');
const testRuns = await fetch('/api/v1/test-runs');
```

GraphQL approach (1 request):
```javascript
const data = await graphqlClient.query(GRAPHQL_QUERIES.GET_DASHBOARD_DATA);
// data contains health, projects, and testRuns
```

## GraphQL Playground

A GraphQL playground is available at `/graphql` when authenticated. This provides:
- Interactive query builder
- Schema documentation
- Query history
- Variable editor
- Performance tracing

## Best Practices

1. **Use Fragments**: Define reusable fragments for common fields
2. **Limit Query Depth**: Avoid deeply nested queries for performance
3. **Use Variables**: Pass dynamic values as variables, not string concatenation
4. **Error Handling**: Check for both network errors and GraphQL errors
5. **Caching**: Leverage the built-in caching for repeated queries

## Error Handling

GraphQL errors are returned in a standard format:

```json
{
    "errors": [
        {
            "message": "User not authenticated",
            "path": ["currentUser"],
            "extensions": {
                "code": "UNAUTHENTICATED"
            }
        }
    ],
    "data": null
}
```

Common error codes:
- `UNAUTHENTICATED`: User not logged in
- `FORBIDDEN`: User lacks permissions
- `NOT_FOUND`: Resource not found
- `BAD_REQUEST`: Invalid query or variables
- `INTERNAL_ERROR`: Server error

## Future Enhancements

Planned features for the GraphQL API:

1. **Mutations**: Create/update projects and test runs
2. **Subscriptions**: Real-time updates via WebSocket
3. **File Uploads**: GraphQL multipart for test artifacts
4. **Field-level Permissions**: Fine-grained access control
5. **Query Whitelisting**: Production-only query whitelist for security