.PHONY: help test test-verbose test-coverage test-race lint fmt vet build clean run

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(DATE)"

# Default target
help:
	@echo "Available targets:"
	@echo "  make test          - Run all tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-race     - Run tests with race detector"
	@echo "  make lint          - Run golangci-lint (requires golangci-lint installed)"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make build         - Build the binary with version info"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make run           - Run the exporter (requires INCIDENTIO_API_KEY)"
	@echo ""
	@echo "Version info:"
	@echo "  VERSION: $(VERSION)"
	@echo "  COMMIT:  $(COMMIT)"
	@echo "  DATE:    $(DATE)"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -count=1

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test ./... -v -count=1

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test ./... -race -count=1

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Build binary with version information
build:
	@echo "Building binary..."
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"
	@go build $(LDFLAGS) -o bin/incidentio-exporter ./cmd/exporter

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the exporter
run:
	@echo "Running exporter..."
	@go run $(LDFLAGS) ./cmd/exporter
