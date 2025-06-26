# AWS Instance Benchmarks - Developer Guide

## 🚀 Quick Start for New Developers

### **Prerequisites**
- Go 1.21+ installed
- Docker or Podman for container operations
- AWS CLI v2 configured with 'aws' profile
- Git with SSH keys configured

### **Repository Setup**
```bash
# Clone the repository
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks

# Install Go dependencies
go mod tidy

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
pip install pre-commit

# Set up pre-commit hooks
pre-commit install

# Verify setup
go test ./... -v
./scripts/check-function-docs.sh
golangci-lint run
```

### **Basic Development Workflow**
```bash
# 1. Create feature branch
git checkout -b feature/stream-benchmark

# 2. Make changes with proper documentation
# (See Documentation Standards below)

# 3. Run quality checks
go test ./... -v
./scripts/check-function-docs.sh
golangci-lint run

# 4. Build and test CLI
go build -o aws-benchmark-collector ./cmd
./aws-benchmark-collector --help

# 5. Commit with pre-commit validation
git add .
git commit -m "Add STREAM benchmark execution"

# 6. Push and create PR
git push origin feature/stream-benchmark
```

## 📖 Documentation Standards (MANDATORY)

### **Function Documentation Template**
```go
// FunctionName performs X operation by implementing Y algorithm.
//
// Detailed explanation of the function's purpose, algorithm, and business logic.
// Include important implementation details, assumptions, and constraints.
//
// Parameters:
//   - param1: Description of first parameter with expected values/constraints
//   - param2: Description of second parameter and validation requirements
//
// Returns:
//   - returnType: Description of return value and possible states
//   - error: Description of error conditions and error types
//
// Example:
//   result, err := FunctionName(param1, param2)
//   if err != nil {
//       log.Fatal("Function failed:", err)
//   }
//   fmt.Printf("Result: %v\n", result)
//
// Performance Notes:
//   - Time complexity: O(n) where n is input size
//   - Memory usage: Allocates ~X MB for processing
//   - Network calls: Makes Y API requests
//
// Common Errors:
//   - ErrInvalidInput: when param1 is nil or empty
//   - ErrNetworkTimeout: when operation exceeds 30s deadline
//   - ErrQuotaExceeded: when AWS limits are reached
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error) {
```

### **Package Documentation Template**
```go
// Package packagename provides functionality for X, enabling users to Y.
//
// This package implements Z algorithm/protocol/service and handles W concerns.
// It is designed for use cases involving V and integrates with U systems.
//
// Key Components:
//   - MainStruct: Primary interface for X operations
//   - HelperStruct: Supporting functionality for Y
//   - Constants: Configuration values and defaults
//
// Usage:
//   service := packagename.New(config)
//   result, err := service.DoSomething(params)
//   if err != nil {
//       log.Fatal("Operation failed:", err)
//   }
//
// The package handles:
//   - Automatic retry logic with exponential backoff
//   - Connection pooling and resource management
//   - Error categorization and recovery strategies
//
// Thread Safety:
//   All exported functions are safe for concurrent use unless noted.
//   Individual instances require external synchronization.
//
// Performance Characteristics:
//   - Typical operation latency: 10-100ms
//   - Memory overhead: 2-5MB per instance
//   - Connection limits: 100 concurrent operations
package packagename
```

### **Struct Documentation Template**
```go
// StructName represents X and provides Y functionality.
//
// This structure encapsulates Z behavior and maintains W state.
// It is designed for X use cases and integrates with Y systems.
//
// Lifecycle:
//   1. Create with NewStructName()
//   2. Configure with SetOptions()
//   3. Use with DoOperation()
//   4. Cleanup with Close()
//
// Thread Safety:
//   This struct is [safe|unsafe] for concurrent use.
//   [Additional thread safety notes]
//
// Example:
//   s := NewStructName(config)
//   defer s.Close()
//   
//   result, err := s.DoOperation(input)
//   if err != nil {
//       return fmt.Errorf("operation failed: %w", err)
//   }
type StructName struct {
    // fieldName is the purpose of this field.
    // Additional constraints or usage notes.
    fieldName string
    
    // privateField contains internal state for X.
    // Not exported to maintain Y invariant.
    privateField int
}
```

### **Documentation Enforcement**
All documentation is automatically validated by:
- **golangci-lint** with revive rules
- **Pre-commit hooks** checking coverage
- **Custom validation script** (`./scripts/check-function-docs.sh`)
- **GitHub Actions** blocking merges without proper docs

## 🧪 Testing Standards

### **Test Coverage Requirements**
- **85%+ overall coverage** across all packages
- **100% coverage** for exported functions
- **Integration tests** for AWS API interactions
- **Error path testing** for resilience validation

### **Test Structure Template**
```go
func TestFunctionName(t *testing.T) {
    // Test cases covering different scenarios
    testCases := []struct {
        name        string
        input       InputType
        expected    ExpectedType
        expectError bool
        errorType   error
    }{
        {
            name:     "valid input success case",
            input:    validInput,
            expected: expectedOutput,
        },
        {
            name:        "invalid input error case",
            input:       invalidInput,
            expectError: true,
            errorType:   ErrInvalidInput,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := FunctionName(tc.input)
            
            if tc.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                if !errors.Is(err, tc.errorType) {
                    t.Errorf("Expected error type %T, got %T", tc.errorType, err)
                }
                return
            }
            
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }
            
            if !reflect.DeepEqual(result, tc.expected) {
                t.Errorf("Expected %v, got %v", tc.expected, result)
            }
        })
    }
}
```

### **Integration Test Patterns**
```go
func TestAWSIntegration(t *testing.T) {
    // Skip if not in integration test mode
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup test environment
    ctx := context.Background()
    orchestrator, err := aws.NewOrchestrator("us-east-1")
    if err != nil {
        t.Fatalf("Failed to create orchestrator: %v", err)
    }
    
    // Test with minimal resources
    config := aws.BenchmarkConfig{
        InstanceType: "t3.nano",  // Use smallest instance for testing
        // ... other config
    }
    
    result, err := orchestrator.RunBenchmark(ctx, config)
    if err != nil {
        t.Fatalf("Benchmark failed: %v", err)
    }
    
    // Validate results
    if result.InstanceID == "" {
        t.Error("Expected instance ID to be set")
    }
}
```

## 🏗️ Architecture Guidelines

### **Package Organization**
```
pkg/
├── discovery/          # AWS instance discovery and mapping
├── aws/               # AWS orchestration and lifecycle management  
├── containers/        # Container build and optimization
├── benchmarks/        # Benchmark execution (future)
├── results/           # Result processing and validation (future)
└── storage/           # Data persistence (future)
```

### **Error Handling Patterns**
```go
// Define package-specific error types
var (
    ErrInvalidInput    = errors.New("invalid input provided")
    ErrQuotaExceeded   = errors.New("AWS quota exceeded")
    ErrNetworkTimeout  = errors.New("network operation timed out")
)

// Use error wrapping for context
func ProcessData(data []byte) error {
    if len(data) == 0 {
        return fmt.Errorf("ProcessData: %w", ErrInvalidInput)
    }
    
    if err := validateData(data); err != nil {
        return fmt.Errorf("data validation failed: %w", err)
    }
    
    return nil
}

// Handle errors with appropriate detail
result, err := ProcessData(input)
if err != nil {
    if errors.Is(err, ErrInvalidInput) {
        // Handle invalid input specifically
        return fmt.Errorf("input validation failed: %w", err)
    }
    // Handle other errors generically
    return fmt.Errorf("processing failed: %w", err)
}
```

### **Configuration Management**
```go
// Use structured configuration
type Config struct {
    AWS struct {
        Region          string        `yaml:"region"`
        Profile         string        `yaml:"profile"`
        Timeout         time.Duration `yaml:"timeout"`
    } `yaml:"aws"`
    
    Benchmarks struct {
        Suites          []string      `yaml:"suites"`
        Iterations      int           `yaml:"iterations"`
        ConfidenceLevel float64       `yaml:"confidence_level"`
    } `yaml:"benchmarks"`
}

// Provide sensible defaults
func DefaultConfig() *Config {
    return &Config{
        AWS: struct {
            Region  string        `yaml:"region"`
            Profile string        `yaml:"profile"`
            Timeout time.Duration `yaml:"timeout"`
        }{
            Region:  "us-east-1",
            Profile: "aws",
            Timeout: 10 * time.Minute,
        },
        Benchmarks: struct {
            Suites          []string `yaml:"suites"`
            Iterations      int      `yaml:"iterations"`
            ConfidenceLevel float64  `yaml:"confidence_level"`
        }{
            Suites:          []string{"stream"},
            Iterations:      10,
            ConfidenceLevel: 0.95,
        },
    }
}
```

## 🔧 Development Tools

### **Useful Commands**
```bash
# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Check documentation coverage
./scripts/check-function-docs.sh

# Run linting
golangci-lint run --config .golangci.yml

# Build all architectures
GOOS=linux GOARCH=amd64 go build -o build/aws-benchmark-collector-linux-amd64 ./cmd
GOOS=linux GOARCH=arm64 go build -o build/aws-benchmark-collector-linux-arm64 ./cmd
GOOS=darwin GOARCH=amd64 go build -o build/aws-benchmark-collector-darwin-amd64 ./cmd
GOOS=darwin GOARCH=arm64 go build -o build/aws-benchmark-collector-darwin-arm64 ./cmd

# Test container builds
./aws-benchmark-collector build --architectures intel-icelake --benchmarks stream --registry localhost:5000

# Test AWS integration (requires valid credentials)
./aws-benchmark-collector discover --dry-run
./aws-benchmark-collector run --help
```

### **IDE Configuration**

#### **VS Code Settings (.vscode/settings.json)**
```json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--config", ".golangci.yml"],
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,128,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "go.generateTestsFlags": ["-exported", "-parallel"]
}
```

#### **GoLand/IntelliJ Setup**
1. Install Go plugin
2. Configure golangci-lint as external tool
3. Set up run configurations for tests
4. Enable coverage highlighting

### **Git Hooks Configuration**
The pre-commit hooks automatically:
- Run `go fmt` and `goimports`
- Execute `golangci-lint` with project config
- Validate documentation coverage
- Check for secrets in code
- Verify commit message format

## 🐛 Debugging Guide

### **Common Issues & Solutions**

#### **AWS Authentication**
```bash
# Verify AWS configuration
aws sts get-caller-identity --profile aws

# Check regional access
aws ec2 describe-instance-types --profile aws --region us-east-1 --max-items 5

# Validate IAM permissions
aws iam simulate-principal-policy \
    --policy-source-arn arn:aws:iam::ACCOUNT:user/USERNAME \
    --action-names ec2:DescribeInstanceTypes \
    --profile aws
```

#### **Container Build Issues**
```bash
# Check Docker daemon
docker info

# Test basic container operations
docker run --rm hello-world

# Debug container builds
docker build --progress=plain --no-cache -t test-build -f Dockerfile .

# Check registry authentication
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws
```

#### **Go Module Issues**
```bash
# Clean module cache
go clean -modcache

# Verify module consistency
go mod tidy
go mod verify

# Check for version conflicts
go mod graph | grep module-name
```

### **Profiling & Performance**
```go
// Add profiling to main function
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Run with profiling
go run ./cmd &
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## 🔒 Security Guidelines

### **Credential Management**
- **Never hardcode** AWS credentials or API keys
- Use **IAM roles** for production environments
- Store secrets in **AWS Secrets Manager** or **Parameter Store**
- Validate **input parameters** to prevent injection attacks

### **Container Security**
```dockerfile
# Use minimal base images
FROM gcr.io/distroless/base-debian11

# Run as non-root user
RUN addgroup --system --gid 1001 benchmarks
RUN adduser --system --uid 1001 --gid 1001 benchmarks
USER benchmarks

# Copy only necessary files
COPY --chown=benchmarks:benchmarks ./binary /usr/local/bin/
```

### **Network Security**
- Use **VPC endpoints** for AWS API access
- Implement **security groups** with minimal required access
- Enable **VPC Flow Logs** for network monitoring
- Use **HTTPS only** for all external communications

## 📊 Performance Guidelines

### **Optimization Strategies**
- **Concurrent execution** for independent operations
- **Connection pooling** for AWS API clients
- **Caching** for frequently accessed data
- **Batch operations** where possible

### **Resource Management**
```go
// Use context for cancellation and timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// Implement proper cleanup
func (s *Service) Close() error {
    // Close connections, clean up resources
    return s.cleanup()
}

// Use resource pools for expensive operations
type ConnectionPool struct {
    pool chan *Connection
}

func (p *ConnectionPool) Get() *Connection {
    select {
    case conn := <-p.pool:
        return conn
    default:
        return p.newConnection()
    }
}
```

## 🚀 Contribution Workflow

### **Feature Development Process**
1. **Create issue** describing the feature/bug
2. **Create feature branch** from main
3. **Implement changes** following standards
4. **Add comprehensive tests** (85%+ coverage)
5. **Update documentation** (100% coverage)
6. **Run quality checks** (all must pass)
7. **Create pull request** with detailed description
8. **Address review feedback** promptly
9. **Merge after approval** and CI success

### **Code Review Checklist**
- [ ] All functions properly documented
- [ ] Test coverage above 85%
- [ ] Error handling comprehensive
- [ ] Performance considerations addressed
- [ ] Security implications reviewed
- [ ] Breaking changes documented
- [ ] Examples work correctly

This developer guide ensures consistent, high-quality contributions to the AWS Instance Benchmarks project while maintaining excellent documentation and testing standards.