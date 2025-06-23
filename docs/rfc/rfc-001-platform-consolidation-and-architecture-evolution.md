# RFC-001: Fern Platform Consolidation and Architecture Evolution

**Status:** Draft  
**Author:** Anoop Gopalakrishnan  
**Created:** June 2025  
**Last Updated:** June 2025  

## Abstract

This RFC proposes a comprehensive architectural evolution of the Fern test reporting ecosystem from the current multi-repository, tightly-coupled design to a unified platform architecture that enables scalable, maintainable, and extensible test intelligence capabilities. The proposal addresses critical operational challenges while preserving the core strengths that have made Fern successful in providing actionable test insights.

## Table of Contents

1. [Background and Current State](#background-and-current-state)
2. [Problem Statement](#problem-statement)
3. [Goals and Non-Goals](#goals-and-non-goals)
4. [Detailed Design](#detailed-design)
5. [Architecture Diagrams](#architecture-diagrams)
6. [Implementation Plan](#implementation-plan)
7. [Migration Strategy](#migration-strategy)
8. [Risks and Mitigation](#risks-and-mitigation)
9. [Success Metrics](#success-metrics)
10. [Open Questions](#open-questions)

## Background and Current State

### What Fern Does Well Today

Fern is a comprehensive test reporting solution with several key strengths:

**1. Multi-Framework Test Intelligence**
- **Ginkgo Integration**: Deep integration with Go's BDD testing framework through `fern-ginkgo-client`
- **JUnit Compatibility**: Universal support for JUnit XML reports via `fern-junit-client`
- **Rich Test Metadata**: Captures detailed test execution context including timing, git metadata, and CI/CD information

**2. Comprehensive Data Model**
- **Hierarchical Structure**: TestRun → SuiteRun → SpecRun provides granular insights
- **Performance Tracking**: Latency measurements at the individual test level
- **Tagging System**: Flexible categorization and filtering capabilities
- **Historical Context**: Long-term trend analysis and regression detection

**3. Multiple Access Patterns**
- **REST API**: Standard HTTP endpoints for data ingestion and retrieval
- **GraphQL Interface**: Flexible querying for complex dashboard requirements
- **gRPC Support**: High-performance binary protocol for service-to-service communication
- **Web Dashboard**: React-based UI for visual test report analysis

**4. Emerging AI Capabilities**
- **Flaky Test Detection**: Statistical analysis to identify unreliable tests
- **MCP Integration**: Model Context Protocol support for AI agent interactions
- **Test Intelligence**: Foundation for automated test analysis and recommendations

**5. Deployment Flexibility**
- **Containerized Services**: Docker support for consistent deployments
- **Kubernetes Ready**: Helm charts and KubeVela configurations
- **Local Development**: Docker Compose for rapid development iteration

### Current Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ fern-ginkgo-    │    │ fern-junit-     │    │                 │
│ client          │    │ client          │    │   Test Suite    │
│ (Go Library)    │    │ (CLI Tool)      │    │   Runners       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌──────────────────┐
                    │ fern-reporter    │
                    │ (Core Backend)   │
                    │ REST|GraphQL|gRPC│
                    └──────────────────┘
                                 │
                ┌────────────────────────────────┐
                │                                │
    ┌─────────────────┐              ┌─────────────────┐
    │ fern-ui         │              │ fern-mycelium   │
    │ (React Dashboard│              │ (AI Intelligence│
    │  & Analytics)   │              │  & MCP Server)  │
    └─────────────────┘              └─────────────────┘
```

## Problem Statement

Despite Fern's success, the current architecture faces significant challenges that limit its ability to meet growing demands:

### 1. Operational Complexity

**Multi-Repository Coordination:**
- Changes requiring updates across multiple services necessitate coordinated releases
- Version compatibility matrix between services becomes complex to manage
- Deployment orchestration requires deep understanding of inter-service dependencies

**Example Pain Point:**
```yaml
# Current deployment requires 3 separate KubeVela configurations
# with duplicated database and gateway definitions
fern-reporter/docs/kubevela/vela.yaml       # 180+ lines
fern-mycelium/docs/kubevela/vela.yaml       # 150+ lines  
fern-ui/kubevela/fern-complete-stack.yaml   # 400+ lines
```

### 2. Architectural Inconsistencies

**Mixed Responsibilities:**
The current `fern-reporter` service violates single responsibility principle:
- **Data Ingestion**: Handles test result submission from clients
- **Query Processing**: Serves dashboard and API requests
- **Analytics**: Computes insights and reports
- **HTML Rendering**: Serves embedded web views

**Data Model Duplication:**
- `fern-reporter/pkg/models/types.go` defines core data structures
- `fern-mycelium/internal/gql/models.go` duplicates similar structures
- No shared schema evolution or validation

### 3. Scalability Limitations

**Write vs. Read Optimization Conflict:**
- Single database optimized for neither high-throughput writes nor complex analytical queries
- No separation between operational data and analytical workloads
- Limited caching strategy for frequently accessed dashboard data

**Resource Scaling Challenges:**
- Cannot scale ingestion capacity independently from query serving
- Analytics processing competes with real-time data ingestion
- No horizontal scaling strategy for individual components

### 4. Developer Experience Friction

**Complex Local Development:**
```bash
# Old setup required multiple terminal sessions:
cd fern-reporter && docker-compose up -d && go run main.go
cd ../fern-mycelium && go run main.go  
cd ../fern-ui && npm install && npm run dev
# Plus manual database setup and service discovery configuration

# New unified platform:
make cluster-setup && kubectl apply -f deployments/fern-platform-kubevela.yaml
```

**Documentation Fragmentation:**
- 5 separate README files with overlapping setup instructions
- No centralized architecture documentation
- Deployment guides scattered across repositories

### 5. Future Capability Constraints

**AI Agent Integration Barriers:**
- No event-driven architecture for real-time test intelligence
- Limited context sharing between analytical and operational data
- No standardized plugin architecture for new test frameworks

**Enterprise Requirements:**
- No unified authentication/authorization strategy
- Limited audit trail and compliance capabilities
- No multi-tenancy support for large organizations

## Goals and Non-Goals

### Goals

**Primary Goals:**
1. **Operational Simplification**: Single-command development setup and unified deployment strategy
2. **Architectural Clarity**: Clear separation of concerns with well-defined service boundaries
3. **Enhanced Scalability**: Independent scaling of ingestion, storage, and query workloads
4. **Developer Productivity**: Faster iteration cycles and reduced cognitive overhead
5. **Future-Proofing**: Extensible architecture supporting AI agents and enterprise features

**Secondary Goals:**
1. **Performance Optimization**: Improved response times for dashboard queries
2. **Data Consistency**: Unified schema evolution and validation
3. **Observability**: Enhanced monitoring and debugging capabilities
4. **Security**: Consolidated authentication and authorization framework

### Non-Goals

**Explicitly Not Included:**
1. **Breaking API Changes**: Existing client integrations must continue working
2. **Framework Lock-in**: No dependency on specific cloud providers or proprietary technologies
3. **Big Bang Migration**: No requirement for simultaneous migration of all components
4. **Feature Removal**: All current capabilities must be preserved or enhanced

## Detailed Design

### Component Selection Strategy: Build vs. Leverage

**Philosophy: Maximize Open Source, Minimize Custom Development**

The proposed architecture follows a "hybrid approach" that leverages proven open source components where possible and builds custom solutions only for Fern-specific requirements.

#### Component Decision Matrix

| Component | Decision | Rationale | Selected Solution |
|-----------|----------|-----------|-------------------|
| **Authentication/Authorization** | Leverage | Well-solved, security-critical | Kong Gateway / Envoy Proxy |
| **Rate Limiting** | Leverage | Standard API gateway feature | Kong Gateway / Envoy Proxy |
| **Message Bus** | Leverage | Battle-tested, extensive ecosystem | Apache Kafka / Redis Streams |
| **Metrics/Observability** | Leverage | Industry standard | OpenTelemetry Collector |
| **Gateway Architecture** | Adapt | Proven pattern, customize for test data | OpenTelemetry Collector Pattern |
| **Test Format Parsers** | Build | Domain-specific, core differentiator | Custom Go Libraries |
| **Schema Validation** | Build | Test-specific validation rules | Custom Validation Engine |
| **Time-Series Storage** | Leverage | Mature solutions available | TimescaleDB / PostgreSQL |
| **Event Sourcing** | Leverage | Standard patterns | EventStore / PostgreSQL |

#### Selected Open Source Components

**Core Infrastructure:**
- **OpenTelemetry Collector**: Plugin architecture for data ingestion
- **Kong Gateway**: API management, auth, rate limiting
- **Apache Kafka**: Event streaming and message bus
- **TimescaleDB**: Time-series data storage
- **Redis**: Caching and lightweight messaging

**Rationale for Component Selection:**
1. **OpenTelemetry Collector**: CNCF graduated project, designed for high-volume data ingestion, excellent plugin architecture
2. **Kong Gateway**: CNCF project, mature ecosystem, battle-tested in production
3. **Apache Kafka**: Industry standard for event streaming, extensive connector ecosystem
4. **TimescaleDB**: PostgreSQL-based, familiar operations, excellent time-series performance

### Proposed Architecture: Event-Driven Microservices

```
                           ┌─────────────────────────────────┐
                           │        Test Execution           │
                           │   Ginkgo | JUnit | PyTest       │
                           └─────────────────────────────────┘
                                         │
                              ┌─────────────────┐
                              │  Fern Clients   │
                              │ (SDK Libraries) │
                              └─────────────────┘
                                         │
                           ┌─────────────────────────────────┐
                           │   Hybrid Ingestion Gateway      │
                           │ ┌─────────────────────────────┐ │
                           │ │ Kong/Envoy (Auth, Rate Limit)│ │
                           │ ├─────────────────────────────┤ │
                           │ │ OTel Collector Architecture │ │
                           │ │ • Custom Test Receivers     │ │
                           │ │ • Schema Validation         │ │
                           │ │ • Format Normalization      │ │
                           │ │ • Event Exporters           │ │
                           │ └─────────────────────────────┘ │
                           └─────────────────────────────────┘
                                         │
                           ┌─────────────────────────────────┐
                           │        Message Bus              │
                           │    Apache Kafka / Redis         │
                           │    (CNCF Standard Components)   │
                           └─────────────────────────────────┘
                          ┌─────────┼─────────┼──────────────┐
                          │         │         │              │
              ┌─────────────────┐ ┌──────────────┐ ┌─────────────────┐
              │   Raw Storage   │ │  Analytics   │ │  Intelligence   │
              │    Service      │ │   Service    │ │    Service      │
              │ • TimescaleDB   │ │ • Real-time  │ │ • AI Agents     │
              │ • Event Store   │ │ • Aggregation│ │ • MCP Server    │
              │ • PostgreSQL    │ │ • Metrics    │ │ • LLM Providers │
              └─────────────────┘ └──────────────┘ └─────────────────┘
                          │         │         │              │
                          └─────────┼─────────┼──────────────┘
                                   │         │
                           ┌─────────────────────────────────┐
                           │    Your Existing Gateway        │
                           │  Kong | Traefik | NGINX | Envoy │
                           │ • Or Optional Built-in Gateway  │
                           │ • GraphQL Federation (Optional) │
                           │ • Gateway-Agnostic Design       │
                           └─────────────────────────────────┘
                                         │
                           ┌─────────────────────────────────┐
                           │     Frontend Applications       │
                           │ • Web Dashboard                 │
                           │ • CLI Tools                     │
                           │ • AI Agent Interfaces           │
                           └─────────────────────────────────┘

Key Architectural Principles:
✅ Leverage CNCF graduated projects (OpenTelemetry, Kong, Kafka)
✅ Build only test-specific components (parsers, validation)
✅ Gateway-agnostic design (works with any existing gateway)
✅ Standard Kubernetes services (no vendor lock-in)
```

### Core Services Design

#### 1. Ingestion Gateway Service (Hybrid Open Source Approach)

**Architecture Decision: Leverage Existing CNCF Components**

Based on analysis of open source solutions, the ingestion gateway will use a **hybrid approach** leveraging proven components:

**Core Framework: OpenTelemetry Collector Architecture**
- **Receiver Pattern**: Custom receivers for test data formats (Ginkgo, JUnit)
- **Processor Pattern**: Schema validation and data transformation
- **Exporter Pattern**: Event publishing to message bus
- **Plugin Architecture**: Extensible for new test frameworks

**Integration Layer: Kong Gateway or Envoy Proxy**
- **Authentication/Authorization**: Leverage existing security features
- **Rate Limiting**: Use battle-tested traffic management
- **Metrics**: Built-in observability and monitoring
- **API Management**: Standard REST/GraphQL endpoint handling

**Implementation Architecture:**
```go
// Leverage OpenTelemetry Collector receiver pattern
type TestDataReceiver struct {
    config     *Config
    consumer   consumer.Traces  // OTel consumer interface
    server     *http.Server
    parsers    map[string]TestDataParser  // Custom parsers
}

// Custom test data processors
type TestDataProcessor struct {
    validators []TestDataValidator
    normalizer TestDataNormalizer
}

// Event publishing via standard exporters
type TestEventExporter struct {
    kafka      *kafka.Producer     // Or Redis Streams
    eventStore EventStore
}

// Test-specific data models (custom)
type TestExecutionEvent struct {
    EventID     string                 `json:"event_id"`
    EventType   string                 `json:"event_type"`
    ProjectID   string                 `json:"project_id"`
    TestRunID   string                 `json:"test_run_id"`
    Timestamp   time.Time              `json:"timestamp"`
    Payload     CanonicalTestRun       `json:"payload"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

**Rationale for Hybrid Approach:**
- **Proven Scalability**: OpenTelemetry Collector handles high-volume data ingestion
- **Reduced Development**: Leverage existing auth, rate limiting, metrics
- **CNCF Ecosystem**: Benefit from community support and best practices
- **Focused Innovation**: Build only test-specific components
- **Maintenance**: Reduced operational burden using mature components

#### 2. Raw Storage Service

**Responsibilities:**
- **Event Sourcing**: Immutable storage of all test execution events
- **Time-Series Data**: Optimized storage for temporal test metrics
- **Data Retention**: Configurable archival and cleanup policies

**Storage Strategy:**
```sql
-- Event Store Table
CREATE TABLE test_execution_events (
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    project_id UUID NOT NULL,
    test_run_id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Time-Series Metrics (TimescaleDB)
CREATE TABLE test_metrics (
    time TIMESTAMPTZ NOT NULL,
    project_id UUID NOT NULL,
    test_name TEXT NOT NULL,
    duration_ms INTEGER,
    status VARCHAR(20),
    tags JSONB
);

SELECT create_hypertable('test_metrics', 'time');
```

#### 3. Analytics Service

**Responsibilities:**
- **Real-time Aggregation**: Continuous computation of test metrics
- **Flaky Test Detection**: Statistical analysis of test reliability
- **Performance Trending**: Historical analysis of test execution times
- **Custom Dashboards**: User-defined metrics and visualizations

**Analytics Pipeline:**
```go
type AnalyticsProcessor struct {
    eventStream <-chan TestExecutionEvent
    aggregators []MetricAggregator
    storage     AnalyticsStorage
}

type FlakyTestAnalyzer struct {
    windowSize    time.Duration
    threshold     float64
    storage       TimeSeriesStorage
}

func (f *FlakyTestAnalyzer) AnalyzeFlakiness(testName string) FlakyTestScore {
    // Statistical analysis of pass/fail patterns
    // Confidence intervals and trend analysis
    // Historical context and seasonality detection
}
```

#### 4. Intelligence Service (Enhanced Mycelium)

**Responsibilities:**
- **AI Agent Framework**: Extensible platform for test intelligence agents
- **MCP Server**: Model Context Protocol implementation for LLM integration  
- **LLM Provider Integration**: Multi-provider support for external AI services
- **Automated Insights**: Context-aware recommendations and alerts

**LLM Provider Architecture:**
```go
type IntelligenceService struct {
    agents        map[string]Agent
    mcpServer     MCPServer
    llmProviders  map[string]LLMProvider
    knowledge     KnowledgeGraph
}

// LLM Provider abstraction for external services
type LLMProvider interface {
    Name() string
    GenerateCompletion(ctx context.Context, prompt string, options CompletionOptions) (*CompletionResponse, error)
    StreamCompletion(ctx context.Context, prompt string, options CompletionOptions) (<-chan CompletionChunk, error)
    ValidateCredentials(ctx context.Context) error
}

// Concrete implementations for major providers
type AnthropicProvider struct {
    apiKey string
    client *anthropic.Client
}

type OpenAIProvider struct {
    apiKey string
    client *openai.Client
}

type HuggingFaceProvider struct {
    apiKey string
    endpoint string
    client *huggingface.InferenceClient
}

// Future extensibility for self-hosted models
type OllamaProvider struct {
    endpoint string
    client   *ollama.Client
}

type Agent interface {
    Name() string
    Process(ctx context.Context, event TestExecutionEvent) ([]Insight, error)
    Subscribe() []EventType
    RequiredLLMCapabilities() []LLMCapability
}

// Example agents with LLM integration
type FlakyTestDetectorAgent struct {
    llmProvider LLMProvider
    promptTemplate string
}

type TestPerformanceCoachAgent struct {
    llmProvider LLMProvider
    analysisPrompts map[string]string
}

type PostmortemGeneratorAgent struct {
    llmProvider LLMProvider
    templateLibrary PostmortemTemplates
}
```

**LLM Integration Strategy:**
```go
type LLMCapability string

const (
    TextGeneration    LLMCapability = "text_generation"
    CodeAnalysis      LLMCapability = "code_analysis"
    DataAnalysis      LLMCapability = "data_analysis"
    ConversationalAI  LLMCapability = "conversational_ai"
)

type CompletionOptions struct {
    MaxTokens     int               `json:"max_tokens"`
    Temperature   float64           `json:"temperature"`
    SystemPrompt  string            `json:"system_prompt,omitempty"`
    Tools         []Tool            `json:"tools,omitempty"`
    ResponseFormat string           `json:"response_format,omitempty"`
}

// Provider factory with credential management
type LLMProviderFactory struct {
    credentials map[string]ProviderCredentials
}

func (f *LLMProviderFactory) CreateProvider(providerName string) (LLMProvider, error) {
    creds, exists := f.credentials[providerName]
    if !exists {
        return nil, fmt.Errorf("no credentials configured for provider: %s", providerName)
    }
    
    switch providerName {
    case "anthropic":
        return NewAnthropicProvider(creds.APIKey), nil
    case "openai":
        return NewOpenAIProvider(creds.APIKey), nil
    case "huggingface":
        return NewHuggingFaceProvider(creds.APIKey, creds.Endpoint), nil
    case "ollama":
        return NewOllamaProvider(creds.Endpoint), nil
    default:
        return nil, fmt.Errorf("unsupported provider: %s", providerName)
    }
}
```

#### 5. Gateway-Agnostic API Layer

**Responsibilities:**
- **Service Exposure**: Standard Kubernetes services for integration flexibility
- **GraphQL Federation**: Optional unified schema across all services
- **Legacy REST Support**: Backward compatibility for existing integrations
- **Authentication**: Pluggable auth with JWT and API keys
- **Caching**: Intelligent response caching for improved performance

**Integration Architecture:**
```
┌─────────────────────────────────────────────────┐
│          User's Existing Infrastructure         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │ Kong        │ │ NGINX       │ │ Traefik     │ │
│  │ Gateway     │ │ Ingress     │ │ Proxy       │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────┬───────────────────────────┘
                      │ (User configures routes)
              ┌───────▼───────┐
              │ Kubernetes    │
              │ Services      │
              └───────┬───────┘
                      │
    ┌─────────────────┼─────────────────┐
    │                 │                 │
    ▼                 ▼                 ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ Ingestion   │ │ Analytics   │ │Intelligence │
│ Service     │ │ Service     │ │ Service     │
└─────────────┘ └─────────────┘ └─────────────┘
```

**Optional GraphQL Federation:**
If users want unified GraphQL, we provide an optional federation service:

```graphql
# Optional Federation Service - deployed only if needed
# Base schema from Raw Storage Service
type TestRun @key(fields: "id") {
  id: ID!
  projectId: String!
  startTime: DateTime!
  endTime: DateTime!
  suiteRuns: [SuiteRun!]!
}

# Extended by Analytics Service  
extend type TestRun {
  passRate: Float!
  flakiness: Float!
  performanceTrend: PerformanceTrend!
}

# Extended by Intelligence Service
extend type TestRun {
  insights: [Insight!]!
  recommendations: [Recommendation!]!
}
```

**Service-First Design:**
```go
// Each service exposes standard REST and GraphQL endpoints
type ServiceConfig struct {
    Port        int    `yaml:"port"`
    EnableAuth  bool   `yaml:"enable_auth"`
    EnableCORS  bool   `yaml:"enable_cors"`
    Endpoints   []EndpointConfig `yaml:"endpoints"`
}

// Services are independently accessible
// Gateway integration is optional
func (s *Service) ExposeEndpoints() {
    // REST endpoints
    s.router.GET("/health", s.healthCheck)
    s.router.GET("/metrics", s.metrics)
    s.router.POST("/api/v1/resource", s.handleResource)
    
    // GraphQL endpoint
    s.router.POST("/graphql", s.graphqlHandler)
    
    // Service can work standalone or behind any gateway
}
```

### Data Flow Patterns

#### Write Path (Test Ingestion)
```
1. Test Client → User's Gateway/Load Balancer
   POST /api/v1/ingest/ginkgo
   
2. User's Gateway → Ingestion Service
   Standard Kubernetes service routing
   
3. Ingestion Service → Message Bus  
   TestExecutionEvent published
   
4. Message Bus → Storage Services
   Raw Storage: Persist event
   Analytics: Update aggregations
   Intelligence: Trigger analysis
   
5. Response to Client
   202 Accepted (async processing)
```

#### Read Path (Dashboard Queries)
**Option 1: Direct Service Access**
```
1. UI Client → User's Gateway
   GET /analytics/api/test-runs
   
2. User's Gateway → Analytics Service
   Direct service routing
   
3. Analytics Service → Cache Check
   Internal Redis cache lookup
   
4. Response
   JSON data or GraphQL response
```

**Option 2: Federation Service (Optional)**
```
1. UI Client → User's Gateway
   GraphQL query for dashboard data
   
2. User's Gateway → Federation Service
   Optional unified GraphQL endpoint
   
3. Federation Service → Multiple Services
   Query federation across services
   
4. Response Assembly
   Combine results, cache, return
```

### Schema Evolution Strategy

**Shared Protocol Definitions:**
```protobuf
// shared/proto/test_execution.proto
syntax = "proto3";

message TestRun {
  string id = 1;
  string project_id = 2;
  google.protobuf.Timestamp start_time = 3;
  google.protobuf.Timestamp end_time = 4;
  repeated SuiteRun suite_runs = 5;
  GitMetadata git_metadata = 6;
  CIMetadata ci_metadata = 7;
}

message SuiteRun {
  string id = 1;
  string name = 2;
  google.protobuf.Timestamp start_time = 3;
  google.protobuf.Timestamp end_time = 4;
  repeated SpecRun spec_runs = 5;
}

message SpecRun {
  string id = 1;
  string description = 2;
  TestStatus status = 3;
  google.protobuf.Duration duration = 4;
  repeated string tags = 5;
  string error_message = 6;
}
```

## Architecture Diagrams

### Current State Architecture
```
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ fern-ginkgo-    │ │ fern-junit-     │ │ Direct Test     │
│ client          │ │ client          │ │ Runners         │
│                 │ │                 │ │                 │
└─────────┬───────┘ └─────────┬───────┘ └─────────┬───────┘
          │                   │                   │
          └───────────────────┼───────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ fern-reporter   │
                    │ ┌─────────────┐ │
                    │ │ REST API    │ │
                    │ ├─────────────┤ │
                    │ │ GraphQL     │ │
                    │ ├─────────────┤ │
                    │ │ gRPC        │ │
                    │ ├─────────────┤ │
                    │ │ HTML Views  │ │
                    │ ├─────────────┤ │
                    │ │ Database    │ │
                    │ └─────────────┘ │
                    └─────────────────┘
                              │
                    ┌─────────┼─────────┐
                    │                   │
          ┌─────────▼───────┐ ┌─────────▼───────┐
          │ fern-ui         │ │ fern-mycelium   │
          │ React Dashboard │ │ AI Intelligence │
          │                 │ │ MCP Server      │
          └─────────────────┘ └─────────────────┘

ISSUES:
• Single point of failure in fern-reporter
• Mixed responsibilities (CRUD + Analytics + Serving)
• Cannot scale components independently  
• Tight coupling between services
• Complex deployment coordination
```

### Proposed Architecture
```
                    ┌──────────────────────────────────┐
                    │         Test Clients             │
                    │  ┌─────────┐ ┌─────────────────┐ │
                    │  │ Ginkgo  │ │ JUnit & Others  │ │
                    │  │ Library │ │ CLI Tools       │ │
                    │  └─────────┘ └─────────────────┘ │
                    └─────────────────┬────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────┐
                    │      Ingestion Gateway          │
                    │ ┌─────────────────────────────┐ │
                    │ │ • Format Detection          │ │
                    │ │ • Schema Validation         │ │
                    │ │ • Rate Limiting             │ │
                    │ │ • Event Publishing          │ │
                    │ └─────────────────────────────┘ │
                    └─────────────────┬───────────────┘
                                      │
                                      ▼
                    ┌──────────────────────────────────┐
                    │         Message Bus              │
                    │     (Event Stream/Queue)         │
                    │  ┌─────────────────────────────┐ │
                    │  │ • Async Event Delivery      │ │
                    │  │ • Guaranteed Processing     │ │
                    │  │ • Replay Capability         │ │
                    │  └─────────────────────────────┘ │
                    └──┬──────────────┬───────────────┬┘
                       │              │               │
           ┌───────────▼─┐ ┌─────────▼───────┐ ┌─────▼─────────┐
           │ Raw Storage │ │   Analytics     │ │ Intelligence  │
           │   Service   │ │    Service      │ │   Service     │
           │ ┌─────────┐ │ │ ┌─────────────┐ │ │ ┌───────────┐ │
           │ │ Events  │ │ │ │ Aggregation │ │ │ │AI Agents  │ │
           │ │ Archive │ │ │ │ Flaky Tests │ │ │ │MCP Server │ │
           │ │ Metrics │ │ │ │ Dashboards  │ │ │ │ML Models  │ │
           │ └─────────┘ │ │ └─────────────┘ │ │ └───────────┘ │
           └─────────────┘ └─────────────────┘ └───────────────┘
                       │              │               │
                       └──────────────┼───────────────┘
                                      │
                    ┌─────────────────────────────────┐
                    │       Unified API Gateway       │
                    │ ┌─────────────────────────────┐ │
                    │ │ • GraphQL Federation        │ │
                    │ │ • REST Compatibility        │ │
                    │ │ • Authentication/AuthZ      │ │
                    │ │ • Caching & Rate Limiting   │ │
                    │ └─────────────────────────────┘ │
                    └─────────────────┬───────────────┘
                                      │
                    ┌─────────────────────────────────┐
                    │    Frontend Applications        │
                    │ ┌─────────┐ ┌─────────────────┐ │
                    │ │Web      │ │CLI Tools &      │ │
                    │ │Dashboard│ │AI Interfaces    │ │
                    │ └─────────┘ └─────────────────┘ │
                    └─────────────────────────────────┘

BENEFITS:
• Independent scaling of each service
• Clear separation of concerns
• Event-driven real-time processing
• Unified API with backward compatibility
• Enhanced observability and debugging
```

### Service Communication Patterns
```
┌─────────────┐    HTTP/gRPC     ┌─────────────────┐
│   Clients   │ ────────────────▶│ Ingestion       │
│             │                  │ Gateway         │
└─────────────┘                  └─────────────────┘
                                           │ Async Events
                                           ▼
                                 ┌─────────────────┐
                                 │  Message Bus    │
                                 │                 │
                                 └─────────┬───────┘
                                           │ Fan-out
                          ┌────────────────┼────────────────┐
                          │                │                │
                          ▼                ▼                ▼
                ┌─────────────────┐ ┌──────────────┐ ┌─────────────┐
                │ Raw Storage     │ │ Analytics    │ │Intelligence │
                │                 │ │              │ │             │
                └─────────────────┘ └──────────────┘ └─────────────┘
                          │                │                │
                          └────────────────┼────────────────┘
                                           │ GraphQL Federation
                                           ▼
                                 ┌─────────────────┐
                                 │  API Gateway    │
                                 │                 │
                                 └─────────┬───────┘
                                           │ HTTP/GraphQL
                                           ▼
                                 ┌─────────────────┐
                                 │   Frontend      │
                                 │  Applications   │
                                 └─────────────────┘
```

## Implementation Activities

### Foundation Activities
**Objective**: Establish unified repository and shared infrastructure

**Core Activities:**
1. **Repository Consolidation**
   - Create `fern-platform` monorepo structure
   - Migrate existing services as subdirectories
   - Establish shared build and development tooling

2. **Shared Schema Definition**
   - Define Protocol Buffer schemas for all data types
   - Create shared Go types generation
   - Implement backward compatibility validation

3. **Development Infrastructure**
   - Unified k3d + KubeVela local development stack
   - Shared Makefile with common targets
   - Consolidated documentation structure

**Acceptance Criteria:**
- Single command (`make dev-setup`) starts entire local stack
- All existing functionality preserved
- Shared types eliminate duplication

### Event Infrastructure Activities
**Objective**: Implement event-driven communication backbone

**Core Activities:**
1. **Message Bus Integration**
   - Integrate event streaming platform (Redis Streams for local, Kafka/NATS for production)
   - Define event schemas and publishing patterns
   - Implement event replay and debugging capabilities

2. **Ingestion Gateway Development**
   - Extract ingestion logic from fern-reporter
   - Implement pluggable parser architecture
   - Add rate limiting and validation layers

3. **Event-Driven Storage**
   - Migrate to event sourcing patterns
   - Implement event replay for data recovery
   - Add audit trail capabilities

**Acceptance Criteria:**
- All test data flows through event system
- Zero data loss during ingestion spikes
- Event replay successfully reconstructs state

### Service Separation Activities
**Objective**: Split monolithic fern-reporter into focused services

**Core Activities:**
1. **Raw Storage Service**
   - Extract core data persistence logic
   - Implement time-series optimizations
   - Add data retention and archival policies

2. **Analytics Service Enhancement**
   - Extract analytics from fern-reporter
   - Implement real-time aggregation pipelines
   - Enhance flaky test detection algorithms

3. **API Gateway Implementation**
   - Implement GraphQL federation
   - Add authentication and authorization
   - Implement intelligent caching strategies

**Acceptance Criteria:**
- Services can scale independently
- API response times improve significantly
- Analytics processing doesn't impact ingestion

### Intelligence Platform Activities
**Objective**: Implement LLM-powered AI capabilities and agent framework

**Core Activities:**
1. **Enhanced Intelligence Service**
   - Migrate and enhance fern-mycelium capabilities
   - Implement multi-provider LLM integration (Anthropic, OpenAI, HuggingFace)
   - Add agent plugin architecture with LLM abstraction layer

2. **LLM-Powered Analytics**
   - Intelligent test failure analysis using external LLM APIs
   - Automated insight generation with natural language explanations
   - Context-aware recommendations and alerts

3. **Integration APIs**
   - Enhanced MCP server implementation for LLM workflows
   - Webhook and notification systems
   - Third-party tool integrations (Slack, GitHub, etc.)

**Acceptance Criteria:**
- AI agents provide actionable insights using external LLM APIs
- MCP integration enables seamless LLM workflows
- Cost-effective LLM usage with intelligent caching and batching
- Extensible architecture supports future self-hosted model integration

### Production Readiness Activities
**Objective**: Production readiness and performance optimization

**Core Activities:**
1. **Observability Stack**
   - Distributed tracing implementation
   - Comprehensive metrics and alerting
   - Performance monitoring dashboards

2. **Security Hardening**
   - Authentication and authorization audit
   - API security scanning and validation
   - Data encryption and compliance features

3. **Deployment Automation**
   - Production-ready Kubernetes manifests
   - Automated rollback and blue-green deployment
   - Disaster recovery procedures

**Acceptance Criteria:**
- High availability deployment capability
- Complete observability into system health
- Automated deployment and rollback procedures

## Migration Strategy

### Backward Compatibility Approach

**API Compatibility:**
```go
// Existing fern-reporter endpoints remain functional
// through API Gateway proxy layer
type LegacyAPIProxy struct {
    newGateway APIGateway
    translator EndpointTranslator
}

func (p *LegacyAPIProxy) CreateTestRun(c *gin.Context) {
    // Translate legacy request to new format
    newReq := p.translator.TranslateLegacyTestRun(c)
    
    // Forward to new ingestion gateway
    response := p.newGateway.IngestTestData(newReq)
    
    // Translate response back to legacy format
    legacyResp := p.translator.TranslateLegacyResponse(response)
    c.JSON(200, legacyResp)
}
```

**Database Migration:**
```sql
-- Phase 1: Run new and old systems in parallel
-- Phase 2: Dual-write to both old and new schemas
-- Phase 3: Validate data consistency
-- Phase 4: Switch reads to new system
-- Phase 5: Deprecate old schema

-- Example migration script
BEGIN TRANSACTION;

-- Create new event-sourced tables
CREATE TABLE test_execution_events (...);

-- Migrate existing data to event format
INSERT INTO test_execution_events (...)
SELECT generate_event_from_testrun(*)
FROM test_runs;

-- Validate data integrity
SELECT validate_migration_consistency();

COMMIT;
```

### Rollback Strategy

**Service-by-Service Rollback:**
1. **Traffic Routing**: Use API Gateway to route traffic back to old services
2. **Data Consistency**: Event replay ensures no data loss during rollback
3. **Feature Flags**: Gradually disable new features before rollback
4. **Monitoring**: Automated alerts trigger rollback on SLA violations

**Zero-Downtime Migration:**
```yaml
# Kubernetes deployment strategy
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  
  # Blue-green deployment with traffic splitting
  selector:
    version: v2
  
  # Canary analysis
  canaryAnalysis:
    threshold: 5
    metrics:
    - name: success-rate
      thresholdRange:
        min: 99
```

### Client Migration Approach

**Immediate:**
- All existing clients continue working unchanged
- New optional features available through enhanced APIs

**Short-term:**
- Client library updates with new features
- Backward compatibility maintained for extended period

**Long-term:**
- Gradual migration to new SDK patterns
- Legacy endpoint deprecation with appropriate notice period

## Risks and Mitigation

### High-Risk Areas

**1. Data Migration Complexity**
- **Risk**: Data corruption or loss during migration
- **Mitigation**: 
  - Comprehensive data validation pipelines
  - Parallel running of old and new systems
  - Automated rollback triggers on data inconsistency

**2. Performance Regression**
- **Risk**: New architecture performs worse than current system
- **Mitigation**:
  - Extensive load testing before migration
  - Performance benchmarking against current system
  - Gradual traffic shifting with monitoring

**3. Client Integration Breaking**
- **Risk**: Existing integrations stop working
- **Mitigation**:
  - Strict API compatibility testing
  - Extended deprecation timeline (12+ months)
  - Comprehensive integration test suite

### Medium-Risk Areas

**4. Operational Complexity**
- **Risk**: More services increase operational burden
- **Mitigation**:
  - Comprehensive observability from day one
  - Automated deployment and scaling
  - Runbook documentation and training

**5. Event System Reliability**
- **Risk**: Message bus becomes single point of failure
- **Mitigation**:
  - High-availability message bus deployment
  - Event persistence and replay capabilities
  - Graceful degradation when events are delayed

### Low-Risk Areas

**6. Configuration Complexity**
- **Risk**: More services increase configuration and deployment complexity
- **Mitigation**:
  - Unified configuration management via KubeVela applications
  - Comprehensive local development tooling
  - Automated configuration validation and testing

## Local Development Strategy

### Maintaining Dependency Simplicity

One of the key success factors of the current Fern architecture is its **dependency simplicity**:
- Single command setup (now `make cluster-setup`)
- Minimal external dependencies
- Fast iteration cycles
- Runs on any developer laptop with k3d

The new platform architecture preserves this simplicity through **intelligent local substitutions**:

### Infrastructure Component Substitutions

| Production Component | Local Development Alternative | Justification |
|---------------------|------------------------------|--------------|
| **Apache Kafka/NATS** | Redis Streams | Single process, persistent, full message bus features |
| **TimescaleDB Cluster** | PostgreSQL + JSONB | Reuse existing DB, sufficient performance for dev |
| **Redis Cluster** | Redis Single Node | Zero operational overhead, full feature parity |
| **Load Balancers** | k3d LoadBalancer | Built-in, zero configuration required |
| **Service Mesh** | Direct Communication | Eliminates complexity, maintains functionality |

### One-Command Development Setup

```bash
# Complete platform setup in under 5 minutes
git clone https://github.com/guidewire-oss/fern-platform.git
cd fern-platform
make dev-setup

# This single command:
# 1. Creates k3d cluster with local registry
# 2. Installs KubeVela for application orchestration  
# 3. Deploys lightweight infrastructure (PostgreSQL + Redis)
# 4. Builds and deploys all Fern services
# 5. Configures port forwarding and ingress
# 6. Runs health checks and displays access points
```

### Local Architecture

```
Developer Laptop (8GB RAM minimum)
├── k3d Cluster (~1.2GB RAM)
│   ├── PostgreSQL (single pod, 256MB)
│   ├── Redis (message bus + cache, 128MB) 
│   ├── Fern Services (6 services, 768MB total)
│   └── KubeVela (orchestration)
├── Local Registry (image storage)
└── Host Processes (test clients, IDE)
```

### Lightweight Message Bus Implementation

For local development, Redis Streams provides full message bus functionality:

```go
// Simplified message bus using Redis Streams
type LocalMessageBus struct {
    redis *redis.Client
}

func (mb *LocalMessageBus) Publish(topic string, event Event) error {
    return mb.redis.XAdd(ctx, &redis.XAddArgs{
        Stream: "fern:" + topic,
        Values: map[string]interface{}{
            "type": event.Type,
            "data": event.Data,
        },
    }).Err()
}

func (mb *LocalMessageBus) Subscribe(topic string, handler EventHandler) {
    // Consumer group provides exactly-once delivery
    mb.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
        Group:   "fern-dev",
        Consumer: "local",
        Streams: []string{"fern:" + topic, ">"},
    })
}
```

### Offline Development Support

The platform supports **completely offline development**:

```yaml
# Mock LLM provider for offline work
intelligence:
  env:
    - name: LLM_PROVIDERS
      value: "mock"
    - name: MOCK_RESPONSES_FILE  
      value: "/app/mock-llm-responses.json"
```

### Development Workflow

```bash
# Daily development cycle
make dev-status    # Check all services healthy
make dev-logs      # View aggregated logs

# Code change workflow
# 1. Edit code
# 2. make build-images
# 3. kubectl rollout restart deployment/service-name
# 4. Auto-restart completes in ~10 seconds

# Testing
make test-local    # Run tests against local environment
make dev-reset     # Reset to clean state

# Cleanup
make dev-destroy   # Remove entire cluster
```

### Resource Requirements

**Minimum laptop specs:**
- 8GB RAM (4GB for cluster, 4GB for development tools)
- 2 CPU cores (4 cores recommended)
- 10GB disk space
- No internet required (with mock providers)

**Service resource allocation:**
```yaml
# Each service configured for minimal resource usage
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

This approach ensures that complex microservices architecture remains as simple to develop with as the current monolithic design, while providing a realistic testing environment that closely mirrors production behavior.

## Success Metrics

### Operational Metrics

**Development Velocity:**
- **Setup Time**: Reduce local development setup from 30+ minutes to <5 minutes
- **Build Time**: Unified build completes in <3 minutes
- **Deployment Time**: Production deployment completes in <10 minutes

**System Reliability:**
- **Uptime**: Achieve 99.9% uptime SLA
- **Error Rate**: Maintain <0.1% error rate for all APIs
- **Recovery Time**: Mean time to recovery <5 minutes

### Performance Metrics

**Ingestion Performance:**
- **Throughput**: Handle 10,000+ test results per minute
- **Latency**: 95th percentile ingestion latency <100ms
- **Scalability**: Linear scaling up to 100,000 test results per minute

**Query Performance:**
- **Dashboard Load Time**: <2 seconds for standard dashboard queries
- **API Response Time**: 95th percentile <500ms for all read operations
- **Cache Hit Rate**: >80% cache hit rate for frequently accessed data

### User Experience Metrics

**Developer Satisfaction:**
- **Setup Complexity**: Single command setup success rate >95%
- **Documentation**: Complete API documentation coverage
- **Error Messages**: Actionable error messages with resolution guidance

**Feature Adoption:**
- **AI Insights**: >50% of users regularly use AI-generated insights
- **Custom Dashboards**: >25% of users create custom analytics views
- **API Usage**: >90% of current integrations migrate to new APIs within 6 months

## Open Questions

### Technical Decisions

**1. Message Bus Technology Choice:**
- **Options**: Apache Kafka, NATS Streaming, Redis Streams, AWS EventBridge
- **Considerations**: Operational complexity, scalability requirements, cloud vendor lock-in
- **Decision Timeline**: Phase 1 completion

**2. Time-Series Database Selection:**
- **Options**: TimescaleDB, InfluxDB, Prometheus + VictoriaMetrics
- **Considerations**: PostgreSQL compatibility, query performance, operational overhead
- **Decision Timeline**: Phase 2 planning

**3. LLM Provider Prioritization:**
- **Primary Focus**: Anthropic Claude, OpenAI GPT, HuggingFace Inference API
- **Secondary Support**: Azure OpenAI, Google Vertex AI, AWS Bedrock
- **Future Extensibility**: Ollama, local model serving, custom endpoints
- **Decision Timeline**: Phase 4 planning

### Technical Questions

**4. Backward Compatibility Timeline:**
- **Question**: How long should legacy APIs be maintained?
- **Options**: 6 months, 12 months, 18 months, indefinitely
- **Dependencies**: Client adoption rates, migration complexity

**5. Release Coordination:**
- **Question**: How should releases be coordinated across services?
- **Options**: Synchronized releases, independent releases, feature flag coordination
- **Dependencies**: CI/CD tooling, deployment automation maturity

### Future Architecture

**7. Multi-Tenancy Strategy:**
- **Question**: How should the platform support multiple organizations?
- **Options**: Database-per-tenant, schema-per-tenant, row-level security
- **Dependencies**: Enterprise sales requirements, compliance needs

**8. Event Retention Policy:**
- **Question**: How long should detailed events be retained?
- **Options**: 30 days, 90 days, 1 year, configurable per tenant
- **Dependencies**: Storage costs, compliance requirements, analytics needs

---

## Conclusion

This RFC proposes a comprehensive evolution of the Fern platform that addresses current operational challenges while positioning the system for future growth and innovation. The event-driven microservices architecture provides clear separation of concerns, independent scalability, and enhanced capabilities for AI-driven test intelligence.

The proposed implementation plan balances the need for architectural improvement with practical migration constraints, ensuring that existing users experience no disruption while new capabilities are developed and deployed.

We invite feedback from all stakeholders on this proposal, particularly regarding:
- Technical implementation details and alternative approaches
- Migration timeline feasibility and risk assessment  
- Open questions requiring community input
- Additional requirements or constraints not addressed

The success of this architectural evolution depends on community consensus and collaborative implementation. We look forward to your comments and contributions to refine this proposal into a robust implementation plan.

---

**Next Steps:**
1. Community review and feedback period (2 weeks)
2. Technical design sessions for open questions
3. Proof-of-concept implementation of critical components
4. Final RFC approval and implementation kickoff

**Contact:**
- RFC Discussion: [GitHub Issues](https://github.com/guidewire-oss/fern-platform/issues)
- Architecture Questions: [GitHub Discussions](https://github.com/guidewire-oss/fern-platform/discussions)