# AWS Instance Benchmarks Makefile
.PHONY: build test clean lint fmt vet install help

# Build configuration
BINARY_NAME=aws-benchmark-collector
PACKAGE=./cmd
BUILD_DIR=./bin
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: fmt vet test build

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(PACKAGE)

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
benchmark:
	@echo "⚡ Running Go benchmarks..."
	go test -bench=. -benchmem ./...

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "🔍 Vetting code..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "🔎 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Install the binary
install: build
	@echo "📥 Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(PACKAGE)

# Generate documentation
docs:
	@echo "📚 Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server on http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Run security checks (requires gosec)
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Check for vulnerabilities (requires govulncheck)
vulncheck:
	@echo "🛡️  Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "⚠️  govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Run all quality checks
check: fmt vet lint test security vulncheck

# Development setup
dev-setup:
	@echo "🚀 Setting up development environment..."
	go mod download
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/godoc@latest

# Help
help:
	@echo "AWS Instance Benchmarks - Available targets:"
	@echo ""
	@echo "  build         Build the binary"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  benchmark     Run Go benchmarks"
	@echo "  fmt           Format code"
	@echo "  vet           Vet code"
	@echo "  lint          Lint code (requires golangci-lint)"
	@echo "  clean         Clean build artifacts"
	@echo "  deps          Install dependencies"
	@echo "  install       Install the binary"
	@echo "  docs          Start documentation server"
	@echo "  security      Run security checks (requires gosec)"
	@echo "  vulncheck     Check for vulnerabilities (requires govulncheck)"
	@echo "  check         Run all quality checks"
	@echo "  dev-setup     Set up development environment"
	@echo "  help          Show this help message"