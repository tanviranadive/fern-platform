# Fern Platform - Main Project Makefile

# Build configuration
BINARY_NAME=fern-platform
BUILD_DIR=bin
CMD_DIR=cmd/fern-platform
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(shell git rev-parse HEAD)

# Go build flags
GO_BUILD_FLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: help build build-linux test test-unit test-integration lint fmt clean dev run deps docker-build docker-run

help: ## Display this help message
	@echo "🌿 Fern Platform Build System"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

build: deps ## Build the platform binary
	@echo "🔨 Building Fern Platform..."
	mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: deps ## Build for Linux (useful for containers)
	@echo "🔨 Building Fern Platform for Linux..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)/main.go
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)-linux"

test: test-unit ## Run all tests

test-unit: ## Run unit tests
	@echo "🧪 Running unit tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Unit tests completed"

test-integration: ## Run integration tests (requires database)
	@echo "🧪 Running integration tests..."
	go test -v -tags=integration ./...
	@echo "✅ Integration tests completed"

test-acceptance: ## Run Go acceptance tests
	@echo "🧪 Running Go acceptance tests..."
	cd acceptance-go && make test
	@echo "✅ Acceptance tests completed"

coverage: test-unit ## Generate test coverage report
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

fmt: ## Format Go code
	@echo "🎨 Formatting Go code..."
	go fmt ./...
	@echo "✅ Code formatted"

lint: ## Run Go linting
	@echo "🔍 Linting Go code..."
	golangci-lint run
	@echo "✅ Linting completed"

vet: ## Run Go vet
	@echo "🔍 Running go vet..."
	go vet ./...
	@echo "✅ Go vet completed"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html
	@echo "✅ Clean completed"

dev: ## Run in development mode with live reload
	@echo "🔧 Starting development mode..."
	air -c .air.toml || go run $(CMD_DIR)/main.go -config config/config.yaml

run: build ## Build and run the platform
	@echo "🚀 Starting Fern Platform..."
	./$(BUILD_DIR)/$(BINARY_NAME) -config config/config.yaml

# Database operations
migrate-up: ## Run database migrations up
	@echo "📈 Running database migrations..."
	go run $(CMD_DIR)/main.go -config config/config.yaml -migrate up

migrate-down: ## Run database migrations down
	@echo "📉 Rolling back database migrations..."
	go run $(CMD_DIR)/main.go -config config/config.yaml -migrate down

migrate-status: ## Check migration status
	@echo "📊 Checking migration status..."
	go run $(CMD_DIR)/main.go -config config/config.yaml -migrate status

# Docker operations
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t fern-platform:$(VERSION) .
	docker tag fern-platform:$(VERSION) fern-platform:latest
	@echo "✅ Docker image built: fern-platform:$(VERSION)"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 fern-platform:latest

# Development tools
install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/vektra/mockery/v2@latest
	@echo "✅ Development tools installed"

# Generate code
generate: ## Generate code (mocks, etc.)
	@echo "🔧 Generating code..."
	go generate ./...
	@echo "✅ Code generation completed"

# K3D Cluster Management
k3d-create: ## Create k3d cluster for fern-platform
	@echo "🎯 Creating k3d cluster..."
	k3d cluster create fern-platform --port "8080:80@loadbalancer" --port "8443:443@loadbalancer"
	@echo "✅ k3d cluster 'fern-platform' created"

k3d-delete: ## Delete k3d cluster
	@echo "🧹 Deleting k3d cluster..."
	k3d cluster delete fern-platform
	@echo "✅ k3d cluster deleted"

k3d-status: ## Check k3d cluster status
	@echo "📊 Checking k3d cluster status..."
	kubectl cluster-info
	kubectl get nodes

# Kubernetes Prerequisites
install-kubevela: ## Install KubeVela operator
	@echo "📦 Installing KubeVela CLI and operator..."
	@if ! command -v vela &> /dev/null; then \
		echo "Installing KubeVela CLI..."; \
		curl -fsSl https://kubevela.io/script/install.sh | bash; \
	fi
	@echo "Installing KubeVela operator to cluster..."
	vela install --version v1.10.3
	@echo "⏳ Waiting for KubeVela operator to be ready..."
	kubectl wait --for=condition=Available deployment/kubevela-vela-core -n vela-system --timeout=300s
	@echo "✅ KubeVela operator installed"

install-cnpg: ## Install CloudNativePG operator
	@echo "📦 Installing CloudNativePG operator using Helm..."
	helm repo add cnpg https://cloudnative-pg.github.io/charts || true
	helm repo update
	helm install cnpg cnpg/cloudnative-pg --namespace postgres-operator --create-namespace
	@echo "⏳ Waiting for CNPG operator to be ready..."
	kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cloudnative-pg -n postgres-operator --timeout=300s
	@echo "✅ CloudNativePG operator installed"

setup-components: ## Install KubeVela component definitions
	@echo "🔧 Installing KubeVela component definitions..."
	vela addon enable velaux
	vela addon enable terraform
	vela addon enable fluxcd
	@echo "📋 Creating custom component definitions..."
	kubectl apply -f deployments/components/
	@echo "✅ Component definitions installed"

setup-prereqs: install-kubevela install-cnpg setup-components ## Install all Kubernetes prerequisites
	@echo "✅ All prerequisites installed and ready"

verify-cluster: ## Verify cluster prerequisites
	@echo "🔍 Verifying cluster prerequisites..."
	@echo "Checking KubeVela..."
	vela version
	@echo "Checking CNPG..."
	kubectl get pods -n cnpg-system
	@echo "Checking component definitions..."
	vela comp list
	@echo "✅ Cluster verification completed"

# Kubernetes/KubeVela operations
k8s-deploy: ## Deploy to Kubernetes using KubeVela
	@echo "☸️ Deploying to Kubernetes..."
	vela up -f deployments/fern-platform-local.yaml
	@echo "✅ Deployed to Kubernetes"

k8s-delete: ## Delete from Kubernetes
	@echo "☸️ Deleting from Kubernetes..."
	vela delete fern-platform
	@echo "✅ Deleted from Kubernetes"

k8s-status: ## Check Kubernetes deployment status
	@echo "☸️ Checking deployment status..."
	vela status fern-platform

# Complete cluster setup workflow
cluster-setup: k3d-create setup-prereqs verify-cluster ## Complete k3d cluster setup with prerequisites
	@echo "🎉 k3d cluster setup completed successfully!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Run 'make k8s-deploy' to deploy fern-platform"
	@echo "2. Run 'make test-acceptance' to run acceptance tests"

cluster-teardown: k8s-delete k3d-delete ## Complete cluster teardown
	@echo "🧹 Cluster teardown completed"

# Release operations
release-dry: ## Dry run release process
	@echo "🚀 Dry run release process..."
	goreleaser release --snapshot --rm-dist

release: ## Create a release
	@echo "🚀 Creating release..."
	goreleaser release --rm-dist

# CI/CD helpers
ci-test: deps test lint vet ## Run CI test pipeline
	@echo "🤖 CI test pipeline completed"

ci-build: deps build build-linux ## Run CI build pipeline
	@echo "🤖 CI build pipeline completed"

# Project information
info: ## Show project information
	@echo "🌿 Fern Platform Information"
	@echo "========================="
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(shell go version)"
	@echo "Binary: $(BUILD_DIR)/$(BINARY_NAME)"

# Local setup helpers
setup-local: ## Setup local development environment
	@echo "🔧 Setting up local development environment..."
	@$(MAKE) deps
	@$(MAKE) install-tools
	@echo "Starting local dependencies with Docker Compose..."
	docker-compose up -d postgres redis
	@echo "✅ Local development environment ready"
	@echo ""
	@echo "Next steps:"
	@echo "1. Run 'make migrate-up' to setup database"
	@echo "2. Run 'make dev' to start development server"

teardown-local: ## Teardown local development environment
	@echo "🧹 Tearing down local development environment..."
	docker-compose down -v
	@$(MAKE) clean
	@echo "✅ Local environment cleaned up"