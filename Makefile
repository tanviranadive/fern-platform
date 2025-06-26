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
	GOOS=linux GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)/main.go
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
	docker build -t fern-platform:latest .
	@echo "✅ Docker image built: fern-platform:latest"

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
	@echo "📦 Checking component definitions..."
	@echo "🔧 Installing component definitions..."
	@vela def apply cnpg.cue || true
	@vela def apply gateway.cue || true
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

# Complete deployment automation
deploy-all: ## Complete automated deployment (k3d + prerequisites + build + deploy)
	@echo "🚀 Starting complete automated deployment of Fern Platform..."
	@echo ""
	@echo "This will:"
	@echo "1. Check/create k3d cluster"
	@echo "2. Install prerequisites (KubeVela, CNPG)"
	@echo "3. Build and load Docker image"
	@echo "4. Deploy application with KubeVela"
	@echo "5. Resume workflow and verify deployment"
	@echo ""
	@$(MAKE) check-or-create-cluster
	@$(MAKE) check-and-install-prerequisites
	@$(MAKE) build-and-load-image
	@$(MAKE) deploy-and-verify
	@echo ""
	@echo "🎉 Fern Platform deployment completed successfully!"
	@echo ""
	@echo "🌐 Application is now accessible at: http://localhost:8080"
	@echo "📡 Port forwarding is running in the background"
	@echo ""
	@echo "📊 Useful commands:"
	@echo "   make k8s-status          - Check deployment status"
	@echo "   make verify-cluster      - Verify all components"
	@echo "   make stop-port-forward   - Stop port forwarding"
	@echo "   make k8s-delete          - Delete deployment"
	@echo "   make k3d-delete          - Delete entire cluster"

check-or-create-cluster: ## Check if k3d cluster exists, create if not
	@echo "🔍 Checking k3d cluster status..."
	@if k3d cluster list | grep -q "fern-platform.*running"; then \
		echo "✅ k3d cluster 'fern-platform' already exists and is running"; \
		kubectl cluster-info --context k3d-fern-platform > /dev/null 2>&1 || (echo "❌ Cluster not accessible, recreating..." && k3d cluster delete fern-platform && k3d cluster create fern-platform --port "8080:80@loadbalancer" --agents 2); \
	else \
		echo "📦 Creating new k3d cluster 'fern-platform'..."; \
		k3d cluster create fern-platform --port "8080:80@loadbalancer" --agents 2; \
		echo "✅ k3d cluster created successfully"; \
	fi
	@echo "🔗 Setting kubectl context..."
	@kubectl config use-context k3d-fern-platform
	@sleep 10
	@echo "✅ Cluster ready"


check-and-install-prerequisites: ## Check and install KubeVela and CNPG if not present
	@echo "🔍 Checking and installing prerequisites..."
	@$(MAKE) check-install-kubevela
	@$(MAKE) check-install-cnpg
	@$(MAKE) check-install-components
	@echo "✅ All prerequisites ready"

check-install-kubevela: ## Check if KubeVela is installed, install if not
	@echo "📦 Checking KubeVela installation..."
	@if kubectl get deployment kubevela-vela-core -n vela-system > /dev/null 2>&1; then \
		echo "✅ KubeVela already installed"; \
		if kubectl get deployment kubevela-vela-core -n vela-system -o jsonpath='{.status.readyReplicas}' | grep -q "1"; then \
			echo "✅ KubeVela is ready"; \
		else \
			echo "⏳ Waiting for KubeVela to be ready..."; \
			kubectl wait --for=condition=Available deployment/kubevela-vela-core -n vela-system --timeout=300s; \
		fi \
	else \
		echo "🔧 Installing KubeVela..."; \
		if ! command -v vela &> /dev/null; then \
			echo "📥 Installing KubeVela CLI..."; \
			curl -fsSl https://kubevela.io/script/install.sh | bash; \
		fi; \
		echo "📦 Installing KubeVela operator..."; \
		helm repo add kubevela https://kubevela.github.io/charts > /dev/null 2>&1 || true; \
		helm repo update > /dev/null 2>&1; \
		helm install --create-namespace -n vela-system kubevela kubevela/vela-core --wait --timeout=10m; \
		echo "✅ KubeVela installed successfully"; \
	fi

check-install-cnpg: ## Check if CloudNativePG is installed, install if not
	@echo "📦 Checking CloudNativePG installation..."
	@if kubectl get deployment cnpg-controller-manager -n cnpg-system > /dev/null 2>&1; then \
		echo "✅ CloudNativePG already installed"; \
		if kubectl get deployment cnpg-controller-manager -n cnpg-system -o jsonpath='{.status.readyReplicas}' | grep -q "1"; then \
			echo "✅ CloudNativePG is ready"; \
		else \
			echo "⏳ Waiting for CloudNativePG to be ready..."; \
			kubectl wait --for=condition=Available deployment/cnpg-controller-manager -n cnpg-system --timeout=300s; \
		fi \
	else \
		echo "🔧 Installing CloudNativePG..."; \
		helm repo add cnpg https://cloudnative-pg.github.io/charts > /dev/null 2>&1 || true; \
		helm repo update > /dev/null 2>&1; \
		helm upgrade --install cnpg --namespace cnpg-system --create-namespace cnpg/cloudnative-pg --wait --timeout=10m; \
		echo "✅ CloudNativePG installed successfully"; \
	fi

check-install-components: ## Check and install component definitions
	@echo "📦 Checking component definitions..."
	@if kubectl get componentdefinition cloud-native-postgres > /dev/null 2>&1; then \
		echo "✅ Component definitions already installed"; \
	else \
		echo "🔧 Installing component definitions..."; \
		cd deployments/components && vela def apply cnpg.cue > /dev/null 2>&1 || true; \
		cd deployments/components && vela def apply gateway.cue > /dev/null 2>&1 || true; \
		echo "✅ Component definitions installed"; \
	fi

build-and-load-image: ## Build Docker image and load into k3d cluster
	@echo "🐳 Building and loading Docker image..."
	@$(MAKE) docker-build
	@echo "📥 Loading image into k3d cluster..."
	@k3d image import fern-platform:latest -c fern-platform
	@echo "✅ Image loaded successfully"

deploy-and-verify: ## Deploy application and verify it's running
	@echo "☸️ Deploying Fern Platform application..."
	@echo "📁 Creating namespace..."
	@kubectl create namespace fern-platform > /dev/null 2>&1 || echo "✅ Namespace already exists"
	@echo "🚀 Applying KubeVela application..."
	@kubectl apply -f deployments/fern-platform-kubevela.yaml
	@echo "⏳ Waiting for initial deployment (60s)..."
	@sleep 60
	@echo "▶️ Resuming workflow..."
	@vela workflow resume fern-platform -n fern-platform > /dev/null 2>&1 || echo "⚠️ Workflow may already be running"
	@echo "⏳ Waiting for deployment to be ready..."
	@timeout=300; \
	while [ $$timeout -gt 0 ]; do \
		if kubectl get pods -n fern-platform | grep fern-platform | grep -q "Running"; then \
			echo "✅ Application is running!"; \
			break; \
		fi; \
		echo "⏳ Still waiting for pods to be ready... ($$timeout seconds remaining)"; \
		sleep 10; \
		timeout=$$((timeout-10)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "⚠️ Deployment may still be in progress. Check status with: vela status fern-platform -n fern-platform"; \
	fi
	@echo "📊 Final status check..."
	@kubectl get pods -n fern-platform
	@echo ""
	@echo "🌐 Application should be accessible via:"
	@echo "   kubectl port-forward -n fern-platform svc/fern-platform 8080:8080"

start-port-forward-and-open: ## Start port forwarding and open browser
	@echo "📡 Starting port forward and opening browser..."
	@echo "⏳ Waiting a moment for service to be ready..."
	@sleep 5
	@echo "🔗 Starting port forward in background..."
	@pkill -f "kubectl.*port-forward.*fern-platform.*8080:8080" > /dev/null 2>&1 || true
	@kubectl port-forward -n fern-platform svc/fern-platform 8080:8080 > /dev/null 2>&1 &
	@echo "⏳ Waiting for port forward to establish..."
	@sleep 3
	@echo "🏥 Checking application health..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Application is healthy and responding!"; \
			break; \
		fi; \
		echo "⏳ Waiting for application to respond... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "⚠️ Application may not be responding yet. Check logs with: kubectl logs -n fern-platform deployment/fern-platform"; \
	fi
	@echo "🌐 Opening browser..."
	@if command -v open >/dev/null 2>&1; then \
		open http://localhost:8080; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		xdg-open http://localhost:8080; \
	elif command -v wslview >/dev/null 2>&1; then \
		wslview http://localhost:8080; \
	else \
		echo "⚠️ Could not detect how to open browser. Please manually open: http://localhost:8080"; \
	fi
	@echo "✅ Port forwarding started (PID: $$(pgrep -f 'kubectl.*port-forward.*fern-platform.*8080:8080' | head -1))"
	@echo "📝 To stop port forwarding: make stop-port-forward"

stop-port-forward: ## Stop port forwarding
	@echo "🛑 Stopping port forward..."
	@pkill -f "kubectl.*port-forward.*fern-platform.*8080:8080" > /dev/null 2>&1 || echo "⚠️ No port forward found"
	@echo "✅ Port forwarding stopped"

# Quick deployment for development (assumes cluster exists)
deploy-quick: build-and-load-image deploy-and-verify ## Quick deployment (assumes cluster and prerequisites exist)
	@echo "🎉 Quick deployment completed!"
	@echo "📌 Access the application at http://fern-platform.local"

# Local setup helpers
setup-local: ## Setup local development environment
	@echo "🔧 Setting up local development environment..."
	@$(MAKE) deps
	@$(MAKE) install-tools
	@echo "✅ Local development environment ready"
	@echo ""
	@echo "Next steps:"
	@echo "1. Run 'make deploy-all' for complete automated deployment"
	@echo "2. Or follow CONTRIBUTING.md for manual k3d cluster setup"

teardown-local: ## Teardown local development environment
	@echo "🧹 Tearing down local development environment..."
	@$(MAKE) cluster-teardown
	@$(MAKE) clean
	@echo "✅ Local environment cleaned up"