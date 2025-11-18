.PHONY: help test test-verbose test-coverage test-e2e test-e2e-verbose test-e2e-coverage fmt fmt-check vet lint lint-install check validate build run-example-basic run-example-advanced clean

# Default target
.DEFAULT_GOAL := help

# Show help message
help:
	@echo "Available targets:"
	@echo "  make help                 - Show this help message"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                  - Format code"
	@echo "  make fmt-check            - Check if code is formatted (fails if not)"
	@echo "  make vet                  - Run go vet"
	@echo "  make lint                 - Run golangci-lint (install with: make lint-install)"
	@echo "  make lint-install         - Install golangci-lint"
	@echo "  make lint-strict          - Run strict linting (may have warnings)"
	@echo "  make check                - Run fmt-check and vet (basic checks)"
	@echo "  make validate             - Run fmt-check, vet, and tests (CI-friendly)"
	@echo ""
	@echo "Testing:"
	@echo "  make test                 - Run all tests"
	@echo "  make test-verbose         - Run all tests with verbose output"
	@echo "  make test-coverage        - Run all tests with coverage report"
	@echo "  make test-e2e             - Run e2e tests"
	@echo "  make test-e2e-verbose     - Run e2e tests with verbose output"
	@echo "  make test-e2e-coverage    - Run e2e tests with coverage"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build                - Build all examples"
	@echo "  make run-example-basic    - Run basic example"
	@echo "  make run-example-advanced - Run advanced example"
	@echo ""
	@echo "Maintenance:"
	@echo "  make mod                  - Update dependencies (go mod tidy)"
	@echo "  make clean                - Clean test artifacts"

# Run all tests
test:
	go test ./...

# Run all tests with verbose output
test-verbose:
	go test -v ./...

# Run all tests with coverage
test-coverage:
	go test -cover ./...

# Run e2e tests
test-e2e:
	go test ./tests/e2e/...

# Run e2e tests with verbose output
test-e2e-verbose:
	go test -v ./tests/e2e/...

# Run e2e tests with coverage
test-e2e-coverage:
	go test -cover ./tests/e2e/...

# Format code
fmt:
	go fmt ./...

# Check if code is formatted (fails if not)
fmt-check:
	@echo "Checking code formatting..."
	@if [ $$(gofmt -l . | grep -v vendor | grep -v ".history" | grep -v ".git" | wc -l) -ne 0 ]; then \
		echo "Error: Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -d . | grep -v vendor | grep -v ".history" | grep -v ".git"; \
		exit 1; \
	fi
	@echo "✓ Code is properly formatted"

# Run go vet
vet:
	go vet ./...

# Install golangci-lint
lint-install:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run golangci-lint
lint: lint-install
	@echo "Running golangci-lint..."
	@golangci-lint run ./... || (echo "Linting failed. Install with: make lint-install" && exit 1)

# Run all code quality checks (lint is optional, use lint-strict for full linting)
check: fmt-check vet
	@echo "✓ Basic code quality checks passed (fmt, vet)"
	@echo "  Run 'make lint' for additional linting checks"

# Run strict linting (may fail on some issues)
lint-strict: lint-install
	@echo "Running strict golangci-lint..."
	@golangci-lint run ./...

# Validate: run checks + tests (CI-friendly, lint is optional)
validate: fmt-check vet test
	@echo "✓ All validation checks passed (fmt, vet, tests)"
	@echo "  Run 'make lint' for additional linting checks"

# Update dependencies
mod:
	go mod tidy

# Build all examples
build:
	go build ./examples/basic/main.go
	go build ./examples/advanced/main.go

# Run basic example
run-example-basic:
	go run ./examples/basic/main.go

# Run advanced example
run-example-advanced:
	go run ./examples/advanced/main.go

# Clean test artifacts
clean:
	go clean -testcache
