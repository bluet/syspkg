.PHONY: all build build-all-arch test lint format fmt check install-tools test-docker test-docker-all test-fixtures

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

test-docker-fedora:
	@echo "Running Fedora DNF tests..."
	docker-compose -f testing/docker/docker-compose.test.yml up fedora-dnf-test --abort-on-container-exit

test-docker-alpine:
	@echo "Running Alpine APK tests..."
	docker-compose -f testing/docker/docker-compose.test.yml up alpine-apk-test --abort-on-container-exit

test-docker-all: test-docker

# Generate test fixtures from different OS
test-fixtures:
	@echo "Generating test fixtures from multiple OS..."
	@mkdir -p testing/fixtures/{apt,yum,dnf,apk}
	docker-compose -f testing/docker/docker-compose.test.yml up --abort-on-container-exit
	@echo "Test fixtures generated in testing/fixtures/"

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
