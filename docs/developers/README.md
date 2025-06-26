# Developer Documentation

Welcome to the Fern Platform developer documentation! This guide will help you get started with development, understand the APIs, and integrate with your systems.

## 🚀 Getting Started

### Prerequisites
Before you begin, ensure you have:
- Docker Desktop running
- k3d, kubectl, helm, and vela CLI installed
- /etc/hosts configured with required entries
- Go 1.23+ (for local development)

### Quick Links
- **[Quick Start Guide](quick-start.md)** - Get Fern Platform running in 15 minutes
- **[Networking & DNS Setup](networking-and-dns.md)** - Understand local DNS requirements
- **[Local k3d Installation](../installation/local-k3d.md)** - Detailed Kubernetes setup

## 🔧 API Documentation

Fern Platform provides comprehensive APIs for integration:

### REST API
- **[REST API Reference](api-reference.md)** - Complete REST endpoint documentation
- Authentication via OAuth 2.0 bearer tokens
- JSON request/response format
- Comprehensive error handling

### GraphQL API  
- **[GraphQL API Guide](../graphql-api.md)** - GraphQL schema and queries
- Single endpoint with flexible queries
- Real-time subscriptions (planned)
- Introspection enabled

## 🔐 Authentication & Security

- **[OAuth Configuration](../configuration/oauth.md)** - Set up OAuth providers
- Support for any OAuth 2.0/OpenID Connect provider
- Role-based access control (admin vs user)
- JWT token validation

## 🏗️ Architecture & Design

- **[Architecture Overview](../ARCHITECTURE.md)** - System design and components
- **[UI Enhancements](../UI_ENHANCEMENTS.md)** - Frontend features and design
- **[RFCs](../rfc/)** - Design proposals and future plans

## 💻 Development Workflow

### Local Development
1. Clone the repository
2. Install dependencies: `make deps`
3. Run locally: `make dev`
4. Run tests: `make test`

### Building & Deployment
1. Build binary: `make build`
2. Build Docker image: `make docker-build`
3. Deploy to k3d: `make deploy-all`

### Code Quality
- Run linting: `make lint`
- Format code: `make fmt`
- Run tests: `make test`

## 🧪 Testing

### Unit Tests
```bash
make test-unit
```

### Integration Tests
```bash
make test-integration
```

### Acceptance Tests
```bash
make test-acceptance
```

## 📝 Contributing

See our [Contributing Guide](../../CONTRIBUTING.md) for:
- Code style guidelines
- Pull request process
- Development best practices
- Community guidelines

## 🆘 Troubleshooting

### Common Issues

**OAuth redirect errors**
- Ensure /etc/hosts has entries for `fern-platform.local` and `keycloak`
- Access via http://fern-platform.local:8080, not localhost

**Pod startup failures**
- Check resource availability: `kubectl describe nodes`
- Review pod logs: `kubectl logs -n fern-platform <pod-name>`

**Component definition errors**
- Reinstall definitions: `vela def apply deployments/components/*.cue`

## 📚 Additional Resources

- [Project README](../../README.md)
- [Documentation Hub](../README.md)
- [GitHub Issues](https://github.com/guidewire-oss/fern-platform/issues)
- [Discussions](https://github.com/guidewire-oss/fern-platform/discussions)