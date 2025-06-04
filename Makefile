.PHONY: all build build-all-arch test lint format fmt check install-tools \
	test-docker test-docker-all test-docker-ubuntu test-docker-rocky test-docker-alma \
	test-docker-clean test-fixtures test-fixtures-legacy test-fixtures-apt test-fixtures-yum \
	test-fixtures-validate test-unit test-integration test-env help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOINSTALL=$(GOCMD) install
BINARY_NAME=syspkg
BINARY_OUTPUT=bin/$(BINARY_NAME)

# Determine platform
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	GOOS=linux
endif
ifeq ($(UNAME_S),Darwin)
	GOOS=darwin
endif
ifeq ($(UNAME_S),Windows_NT)
	GOOS=windows
endif

# Determine architecture
UNAME_P := $(shell uname -m)
ifeq ($(UNAME_P),x86_64)
	GOARCH=amd64
endif
ifeq ($(UNAME_P),aarch64)
	GOARCH=arm64
endif

all: test build

build: lint install-tools
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -o $(BINARY_OUTPUT) ./cmd/syspkg

build-all-arch: lint install-tools
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_OUTPUT)_linux_amd64 ./cmd/syspkg
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINARY_OUTPUT)_linux_arm64 ./cmd/syspkg
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_OUTPUT)_darwin_amd64 ./cmd/syspkg
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_OUTPUT)_darwin_arm64 ./cmd/syspkg
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_OUTPUT)_windows_amd64.exe ./cmd/syspkg

test:
	$(GOTEST) -v ./...
	$(GOTEST) -v -run ExampleGetOSInfo ./osinfo

lint:
	go mod tidy
	golangci-lint run
	gofmt -s -w .

format fmt:
	@echo "Running gofmt..."
	@gofmt -s -w .
	@echo "Running goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@goimports -w -local github.com/bluet/syspkg .
	@echo "Formatting complete!"

check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files need formatting:"; \
		gofmt -l .; \
		echo ""; \
		echo "Run 'make format' to fix formatting"; \
		exit 1; \
	fi
	@echo "Checking go mod tidy..."
	@go mod tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum needs updating"; \
		echo "Run 'go mod tidy' and commit the changes"; \
		exit 1; \
	fi
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running golangci-lint..."
	@golangci-lint run
	@echo "All checks passed!"

install-tools:
	$(GOINSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker testing targets
test-docker:
	@echo "Running tests in Docker containers..."
	docker-compose -f testing/docker/docker-compose.test.yml up --abort-on-container-exit --remove-orphans

test-docker-ubuntu:
	@echo "Running Ubuntu APT tests..."
	docker-compose -f testing/docker/docker-compose.test.yml up ubuntu-apt-test --abort-on-container-exit

test-docker-rocky:
	@echo "Running Rocky Linux YUM tests..."
	docker-compose -f testing/docker/docker-compose.test.yml up rockylinux-yum-test --abort-on-container-exit

test-docker-alma:
	@echo "Running AlmaLinux YUM tests..."
	docker-compose -f testing/docker/docker-compose.test.yml up almalinux-yum-test --abort-on-container-exit

# TODO: Enable when DNF support is implemented
# test-docker-fedora:
# 	@echo "Running Fedora DNF tests..."
# 	docker-compose -f testing/docker/docker-compose.test.yml up fedora-dnf-test --abort-on-container-exit

# TODO: Enable when APK support is implemented
# test-docker-alpine:
# 	@echo "Running Alpine APK tests..."
# 	docker-compose -f testing/docker/docker-compose.test.yml up alpine-apk-test --abort-on-container-exit

test-docker-all: test-docker

# Generate test fixtures from different OS using persistent containers for realistic scenarios
test-fixtures:
	@echo "Generating comprehensive test fixtures with realistic system states..."
	@mkdir -p testing/fixtures/{apt,yum,dnf,apk}
	@echo "Using persistent container approach for authentic real-world scenarios..."
	./testing/generate-fixtures.sh
	@echo "Comprehensive test fixtures generated in testing/fixtures/"

# Legacy fixture generation using docker-compose (basic commands only)
test-fixtures-legacy:
	@echo "Generating basic test fixtures from multiple OS..."
	@mkdir -p testing/fixtures/{apt,yum,dnf,apk}
	docker-compose -f testing/docker/docker-compose.test.yml up --abort-on-container-exit
	@echo "Basic test fixtures generated in testing/fixtures/"

# Generate fixtures for specific package managers
test-fixtures-apt:
	@echo "Generating comprehensive APT fixtures only..."
	@mkdir -p testing/fixtures/apt
	@echo "This will create a persistent Ubuntu container with realistic system state..."
	GENERATE_ONLY=apt ./testing/generate-fixtures.sh

test-fixtures-yum:
	@echo "Generating comprehensive YUM fixtures only..."
	@mkdir -p testing/fixtures/yum
	@echo "This will create a persistent Rocky Linux container with realistic system state..."
	GENERATE_ONLY=yum ./testing/generate-fixtures.sh

# Validate generated fixtures
test-fixtures-validate:
	@echo "Validating generated fixtures..."
	@for fixture in testing/fixtures/apt/*.txt; do \
		if [ -f "$$fixture" ]; then \
			lines=$$(wc -l < "$$fixture" 2>/dev/null || echo "0"); \
			echo "  $$(basename $$fixture): $$lines lines"; \
			if [ "$$lines" -eq 0 ]; then \
				echo "    ⚠️  Empty fixture detected"; \
			fi; \
		fi; \
	done
	@echo "Run 'make test' to verify fixtures work with parsers"

# Clean up Docker resources
test-docker-clean:
	@echo "Cleaning up Docker test resources..."
	docker-compose -f testing/docker/docker-compose.test.yml down --volumes --remove-orphans
	docker system prune -f --filter "label=com.docker.compose.project=syspkg-test"

# Unit tests only (no integration/system tests)
test-unit:
	$(GOTEST) -v -tags=unit ./...

# Integration tests (requires appropriate OS/package managers)
test-integration:
	$(GOTEST) -v -tags=integration ./...

# Environment-aware testing
test-env:
	@echo "Running environment-aware tests..."
	@echo "OS: $$(go run ./testing/testenv/cmd/detect-env || echo 'Unknown')"
	$(GOTEST) -v -tags="unit,integration" ./...

# Show help information
help:
	@echo "SysPkg Build & Test Targets:"
	@echo ""
	@echo "🏗️  Build Targets:"
	@echo "  make build              - Build for current platform"
	@echo "  make build-all-arch     - Build for all supported platforms"
	@echo "  make all                - Run tests and build"
	@echo ""
	@echo "🧪 Test Targets:"
	@echo "  make test               - Run all tests"
	@echo "  make test-unit          - Run unit tests only"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make test-env           - Environment-aware testing"
	@echo ""
	@echo "🐳 Docker Test Targets:"
	@echo "  make test-docker        - Run tests in all Docker containers"
	@echo "  make test-docker-ubuntu - APT tests in Ubuntu container"
	@echo "  make test-docker-rocky  - YUM tests in Rocky Linux container"
	@echo "  make test-docker-clean  - Clean up Docker test resources"
	@echo ""
	@echo "📋 Fixture Generation (Comprehensive Approach):"
	@echo "  make test-fixtures      - Generate comprehensive fixtures for all package managers"
	@echo "  make test-fixtures-apt  - Generate comprehensive APT fixtures only"
	@echo "  make test-fixtures-yum  - Generate comprehensive YUM fixtures only"
	@echo "  make test-fixtures-validate - Validate generated fixtures"
	@echo "  make test-fixtures-legacy   - Generate basic fixtures (old approach)"
	@echo ""
	@echo "💡 Key Features of Fixture Generation:"
	@echo "   • Uses persistent containers for realistic system states"
	@echo "   • Captures error scenarios (not found, already installed)"
	@echo "   • Includes complex operations (autoremove, dependencies)"
	@echo "   • Generates authentic real-world command outputs"
	@echo "   • Safe: All operations run in disposable Docker containers"
	@echo ""
	@echo "🔧 Code Quality:"
	@echo "  make lint               - Run linting and formatting"
	@echo "  make format             - Format code with gofmt and goimports"
	@echo "  make check              - Run all code quality checks"
	@echo "  make install-tools      - Install required development tools"
	@echo ""
	@echo "📖 Examples:"
	@echo "  make test-fixtures-apt  # Generate realistic APT fixtures"
	@echo "  make test               # Run tests with new fixtures"
	@echo "  head testing/fixtures/apt/autoremove-apt.txt  # View fixture"
