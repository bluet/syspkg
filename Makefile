.PHONY: all build build-all-arch test lint format fmt check install-tools

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

# TODO: Add Docker testing targets when Dockerfiles are implemented
# TODO: Add unit/integration test targets when build tags are added to test files
