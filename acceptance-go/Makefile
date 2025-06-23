# Fern Platform - Go Acceptance Tests Makefile

# Configuration
GINKGO := go run github.com/onsi/ginkgo/v2/ginkgo
KUBECTL := kubectl
VELA := vela
TEST_NAMESPACE_PREFIX := fern-acceptance
K8S_CONTEXT := k3d-fern-platform

# Test execution settings
PARALLEL_PROCESSES := 4
TEST_TIMEOUT := 30m
VERBOSE := false

# Platform configuration
# Set to true to use existing deployed platform instead of creating k3d environment
USE_EXISTING_PLATFORM := true
EXISTING_PLATFORM_URL := http://localhost:8080

# Cluster and deployment settings
KUBEVELA_APP_FILE := ../deployments/fern-platform-local.yaml
CLUSTER_SETUP_TIMEOUT := 10m

.PHONY: help deps test test-api test-integration test-ui test-all clean setup verify-cluster

help: ## Display this help message
	@echo "🌿 Fern Platform Go Acceptance Tests"
	@echo ""
	@echo "📚 Test Execution Modes:"
	@echo "  USE_EXISTING_PLATFORM=true   🚀 Fast mode: test against existing platform"
	@echo "  USE_EXISTING_PLATFORM=false  🐌 Full mode: deploy fresh k3d environment"
	@echo ""
	@echo "🎯 Recommended Usage:"
	@echo "  make test-existing           # Quick tests against deployed platform"
	@echo "  make test-isolated           # Full isolation with fresh deployment" 
	@echo "  make test                    # Smart mode (respects USE_EXISTING_PLATFORM)"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install test dependencies
	@echo "Installing Go test dependencies..."
	go mod download
	go mod tidy
	@echo "Installing Ginkgo CLI..."
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@echo "✅ Dependencies installed"

verify-cluster: ## Verify cluster prerequisites (KubeVela, CNPG)
	@echo "🔍 Verifying cluster prerequisites..."
	@echo "Checking kubectl connection..."
	$(KUBECTL) cluster-info --context=$(K8S_CONTEXT) || (echo "❌ kubectl not connected to k3d cluster" && exit 1)
	@echo "Checking KubeVela installation..."
	$(KUBECTL) get deployment vela-core -n vela-system --context=$(K8S_CONTEXT) || (echo "❌ KubeVela not installed" && exit 1)
	@echo "Checking CNPG operator..."
	$(KUBECTL) get deployment cnpg-controller-manager -n cnpg-system --context=$(K8S_CONTEXT) || (echo "❌ CNPG operator not installed" && exit 1)
	@echo "Verifying KubeVela component definitions..."
	$(VELA) comp list | grep -q "postgres" || (echo "❌ PostgreSQL component definition not found" && exit 1)
	$(VELA) comp list | grep -q "gateway" || (echo "❌ Gateway component definition not found" && exit 1)
	@echo "✅ All cluster prerequisites verified"

setup: deps verify-cluster ## Setup test environment and dependencies
	@echo "🚀 Setting up test environment..."
	@echo "Creating test configuration..."
	@mkdir -p ./test-results
	@echo "✅ Test environment setup complete"

test-api: setup ## Run API acceptance tests only
	@echo "🧪 Running API acceptance tests..."
	$(GINKGO) run \
		--procs=$(PARALLEL_PROCESSES) \
		--timeout=$(TEST_TIMEOUT) \
		--output-dir=./test-results \
		--json-report=api-test-results.json \
		--junit-report=api-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/api/
	@echo "✅ API acceptance tests completed"

# Targets for running tests against existing platform (fast)
test-api-existing: deps ## Run API acceptance tests against existing deployed platform
	@echo "🧪 Running API acceptance tests against existing platform..."
	@mkdir -p ./test-results
	@echo "Using existing platform at: $(EXISTING_PLATFORM_URL)"
	$(GINKGO) run \
		--procs=$(PARALLEL_PROCESSES) \
		--timeout=10m \
		--output-dir=./test-results \
		--json-report=api-test-results.json \
		--junit-report=api-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/api/
	@echo "✅ API acceptance tests completed"

test-integration-existing: deps ## Run integration acceptance tests against existing deployed platform
	@echo "🧪 Running integration acceptance tests against existing platform..."
	@mkdir -p ./test-results
	@echo "Using existing platform at: $(EXISTING_PLATFORM_URL)"
	$(GINKGO) run \
		--procs=$(PARALLEL_PROCESSES) \
		--timeout=10m \
		--output-dir=./test-results \
		--json-report=integration-test-results.json \
		--junit-report=integration-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/integration/
	@echo "✅ Integration acceptance tests completed"

test-ui-existing: deps ## Run UI acceptance tests against existing deployed platform
	@echo "🧪 Running UI acceptance tests against existing platform..."
	@mkdir -p ./test-results
	@echo "Using existing platform at: $(EXISTING_PLATFORM_URL)"
	$(GINKGO) run \
		--procs=2 \
		--timeout=15m \
		--output-dir=./test-results \
		--json-report=ui-test-results.json \
		--junit-report=ui-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/ui/
	@echo "✅ UI acceptance tests completed"

test-all-existing: deps ## Run all acceptance tests against existing deployed platform  
	@echo "🧪 Running all acceptance tests against existing platform..."
	@mkdir -p ./test-results
	@echo "Using existing platform at: $(EXISTING_PLATFORM_URL)"
	$(GINKGO) run \
		--procs=$(PARALLEL_PROCESSES) \
		--timeout=15m \
		--output-dir=./test-results \
		--json-report=all-test-results.json \
		--junit-report=all-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/...
	@echo "✅ All acceptance tests completed"

test-integration: setup ## Run integration acceptance tests only
	@echo "🧪 Running integration acceptance tests..."
	$(GINKGO) run \
		--procs=$(PARALLEL_PROCESSES) \
		--timeout=$(TEST_TIMEOUT) \
		--output-dir=./test-results \
		--json-report=integration-test-results.json \
		--junit-report=integration-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/integration/
	@echo "✅ Integration acceptance tests completed"

test-ui: setup ## Run UI acceptance tests only
	@echo "🧪 Running UI acceptance tests..."
	$(GINKGO) run \
		--procs=2 \
		--timeout=$(TEST_TIMEOUT) \
		--output-dir=./test-results \
		--json-report=ui-test-results.json \
		--junit-report=ui-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/ui/
	@echo "✅ UI acceptance tests completed"

test-all: setup ## Run all acceptance tests
	@echo "🧪 Running all acceptance tests..."
	$(GINKGO) run \
		--procs=$(PARALLEL_PROCESSES) \
		--timeout=$(TEST_TIMEOUT) \
		--output-dir=./test-results \
		--json-report=all-test-results.json \
		--junit-report=all-test-results.xml \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/...
	@echo "✅ All acceptance tests completed"

# Smart test targets that respect USE_EXISTING_PLATFORM flag
test-smart-api: ## Run API tests (respects USE_EXISTING_PLATFORM flag)
ifeq ($(USE_EXISTING_PLATFORM),true)
	@$(MAKE) test-api-existing
else
	@$(MAKE) test-api
endif

test-smart-integration: ## Run integration tests (respects USE_EXISTING_PLATFORM flag)  
ifeq ($(USE_EXISTING_PLATFORM),true)
	@$(MAKE) test-integration-existing
else
	@$(MAKE) test-integration
endif

test-smart-ui: ## Run UI tests (respects USE_EXISTING_PLATFORM flag)
ifeq ($(USE_EXISTING_PLATFORM),true)
	@$(MAKE) test-ui-existing
else
	@$(MAKE) test-ui
endif

test-smart-all: ## Run all tests (respects USE_EXISTING_PLATFORM flag)
ifeq ($(USE_EXISTING_PLATFORM),true)
	@$(MAKE) test-all-existing
else
	@$(MAKE) test-all
endif

test: test-smart-all ## Alias for smart test execution

# Mode switching convenience targets
test-existing: ## Run tests in fast mode (against existing platform)
	@$(MAKE) test USE_EXISTING_PLATFORM=true

test-isolated: ## Run tests in isolated mode (fresh k3d deployment)
	@$(MAKE) test USE_EXISTING_PLATFORM=false

test-verbose: ## Run all tests with verbose output
	@$(MAKE) test VERBOSE=true

test-fast: ## Run tests with minimal parallelism for quick feedback
	@echo "🧪 Running fast acceptance tests..."
	$(GINKGO) run \
		--procs=2 \
		--timeout=15m \
		--focus="should load|should display|should navigate" \
		--output-dir=./test-results \
		--json-report=fast-test-results.json \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/...
	@echo "✅ Fast acceptance tests completed"

test-smoke: ## Run smoke tests only
	@echo "🧪 Running smoke tests..."
	$(GINKGO) run \
		--procs=2 \
		--timeout=10m \
		--focus="should.*health|should.*load.*dashboard|should.*basic.*query" \
		--output-dir=./test-results \
		--json-report=smoke-test-results.json \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/...
	@echo "✅ Smoke tests completed"

test-performance: ## Run performance-focused tests
	@echo "🧪 Running performance tests..."
	$(GINKGO) run \
		--procs=1 \
		--timeout=$(TEST_TIMEOUT) \
		--focus="Performance|performance|load time|response time" \
		--output-dir=./test-results \
		--json-report=performance-test-results.json \
		$(if $(filter true,$(VERBOSE)),-v) \
		./specs/...
	@echo "✅ Performance tests completed"

clean-namespaces: ## Clean up test namespaces
	@echo "🧹 Cleaning up test namespaces..."
	@for ns in $$($(KUBECTL) get namespaces --context=$(K8S_CONTEXT) -o name | grep $(TEST_NAMESPACE_PREFIX)); do \
		echo "Deleting namespace: $$ns"; \
		$(KUBECTL) delete $$ns --context=$(K8S_CONTEXT) --timeout=60s || true; \
	done
	@echo "✅ Test namespaces cleaned up"

clean-artifacts: ## Clean up test artifacts
	@echo "🧹 Cleaning up test artifacts..."
	rm -rf ./test-results/*
	@echo "✅ Test artifacts cleaned up"

clean: clean-namespaces clean-artifacts ## Clean up everything

watch-api: ## Watch and run API tests on file changes
	$(GINKGO) watch \
		--procs=2 \
		--timeout=10m \
		./specs/api/

watch-ui: ## Watch and run UI tests on file changes
	$(GINKGO) watch \
		--procs=1 \
		--timeout=15m \
		./specs/ui/

debug-api: ## Run API tests with debug output
	@echo "🐛 Running API tests in debug mode..."
	$(GINKGO) run \
		--procs=1 \
		--verbose \
		--trace \
		--timeout=$(TEST_TIMEOUT) \
		./specs/api/

debug-ui: ## Run UI tests with debug output
	@echo "🐛 Running UI tests in debug mode..."
	$(GINKGO) run \
		--procs=1 \
		--verbose \
		--trace \
		--timeout=$(TEST_TIMEOUT) \
		./specs/ui/

lint: ## Run Go linting
	@echo "🔍 Running Go linting..."
	go fmt ./...
	go vet ./...
	@echo "✅ Linting completed"

generate: ## Generate any code (if needed)
	@echo "🔧 Generating code..."
	go generate ./...
	@echo "✅ Code generation completed"

build: ## Build test binaries
	@echo "🔨 Building test binaries..."
	$(GINKGO) build ./specs/...
	@echo "✅ Test binaries built"

install-tools: ## Install additional testing tools
	@echo "🔧 Installing additional testing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	@echo "✅ Additional tools installed"

check-deps: ## Check if all dependencies are available
	@echo "🔍 Checking dependencies..."
	@command -v kubectl >/dev/null 2>&1 || (echo "❌ kubectl is required but not installed" && exit 1)
	@command -v vela >/dev/null 2>&1 || (echo "❌ vela CLI is required but not installed" && exit 1)
	@command -v go >/dev/null 2>&1 || (echo "❌ Go is required but not installed" && exit 1)
	@echo "✅ All dependencies are available"

status: ## Show test environment status
	@echo "📊 Test Environment Status"
	@echo "=========================="
	@echo "Kubernetes Context: $(K8S_CONTEXT)"
	@echo "KubeVela Status:"
	@$(KUBECTL) get pods -n vela-system --context=$(K8S_CONTEXT) 2>/dev/null || echo "  ❌ KubeVela not accessible"
	@echo "CNPG Status:"
	@$(KUBECTL) get pods -n cnpg-system --context=$(K8S_CONTEXT) 2>/dev/null || echo "  ❌ CNPG not accessible"
	@echo "Test Namespaces:"
	@$(KUBECTL) get namespaces --context=$(K8S_CONTEXT) | grep $(TEST_NAMESPACE_PREFIX) || echo "  No active test namespaces"
	@echo ""
	@echo "Test Results:"
	@ls -la ./test-results/ 2>/dev/null || echo "  No test results available"

env: ## Show environment variables for testing
	@echo "📝 Test Environment Variables"
	@echo "============================"
	@echo "KUBECONFIG: $${KUBECONFIG:-default}"
	@echo "K8S_CONTEXT: $(K8S_CONTEXT)"
	@echo "TEST_NAMESPACE_PREFIX: $(TEST_NAMESPACE_PREFIX)"
	@echo "PARALLEL_PROCESSES: $(PARALLEL_PROCESSES)"
	@echo "TEST_TIMEOUT: $(TEST_TIMEOUT)"
	@echo "KUBEVELA_APP_FILE: $(KUBEVELA_APP_FILE)"
	@echo ""
	@echo "Platform Configuration:"
	@echo "USE_EXISTING_PLATFORM: $(USE_EXISTING_PLATFORM)"
	@echo "EXISTING_PLATFORM_URL: $(EXISTING_PLATFORM_URL)"
	@echo ""
	@echo "Test Mode:"
ifeq ($(USE_EXISTING_PLATFORM),true)
	@echo "  🚀 FAST MODE: Using existing platform (no k3d deployment)"
else
	@echo "  🐌 FULL MODE: Will create isolated k3d environment"
endif

# CI/CD targets
ci-setup: check-deps verify-cluster ## Setup for CI environment
	@echo "🤖 Setting up CI environment..."
	@$(MAKE) setup

ci-test: ci-setup ## Run tests in CI environment
	@echo "🤖 Running tests in CI environment..."
	@$(MAKE) test-all PARALLEL_PROCESSES=2 VERBOSE=true

ci-cleanup: ## Cleanup after CI run
	@echo "🤖 Cleaning up CI environment..."
	@$(MAKE) clean

# Docker targets (if running in containers)
docker-test: ## Run tests in Docker container
	@echo "🐳 Running tests in Docker..."
	docker run --rm \
		-v $(PWD):/workspace \
		-v $(HOME)/.kube:/root/.kube \
		-w /workspace \
		golang:1.21 \
		make test-all

# Help with common issues
troubleshoot: ## Show troubleshooting information
	@echo "🔧 Troubleshooting Guide"
	@echo "======================="
	@echo ""
	@echo "Common Issues:"
	@echo "1. Cluster not ready:"
	@echo "   - Run: make verify-cluster"
	@echo "   - Check: kubectl cluster-info"
	@echo ""
	@echo "2. KubeVela not working:"
	@echo "   - Check: vela version"
	@echo "   - Check: kubectl get pods -n vela-system"
	@echo ""
	@echo "3. Tests timing out:"
	@echo "   - Increase TEST_TIMEOUT: make test TEST_TIMEOUT=45m"
	@echo "   - Reduce parallelism: make test PARALLEL_PROCESSES=1"
	@echo ""
	@echo "4. UI tests failing:"
	@echo "   - Ensure Chrome/Chromium is available"
	@echo "   - Check browser dependencies"
	@echo ""
	@echo "For more help, check the README.md file"