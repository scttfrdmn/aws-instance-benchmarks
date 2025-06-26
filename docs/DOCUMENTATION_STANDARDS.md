# Documentation Standards

This document defines the comprehensive documentation requirements for the AWS Instance Benchmarks project to ensure excellent developer onboarding experience.

## 📋 Core Requirements

### **1. Every Exported Function Must Have Documentation**
```go
// GenerateArchitectureMappings creates a mapping between AWS instance families 
// and their corresponding optimized container architectures.
//
// This function processes a list of instance information and deduplicates by 
// instance family, ensuring each family maps to exactly one container architecture.
// The mapping enables efficient container selection during benchmark execution.
//
// Parameters:
//   - instances: Slice of InstanceInfo containing AWS instance type details
//
// Returns:
//   - map[string]ArchitectureMapping: Family name -> container architecture mapping
//
// Example:
//   instances := []InstanceInfo{{InstanceFamily: "m7i", Architecture: "x86_64"}}
//   mappings := discoverer.GenerateArchitectureMappings(instances)
//   // mappings["m7i"] -> {ContainerTag: "intel-icelake", ...}
func (d *InstanceDiscoverer) GenerateArchitectureMappings(instances []InstanceInfo) map[string]ArchitectureMapping {
```

### **2. Package-Level Documentation Required**
```go
// Package discovery provides AWS EC2 instance type discovery and architecture 
// mapping functionality for the AWS Instance Benchmarks project.
//
// This package handles automatic discovery of AWS instance types through the 
// EC2 API, extracts microarchitecture information, and generates mappings 
// between instance families and optimized benchmark container tags.
//
// Key Components:
//   - InstanceDiscoverer: Main service for AWS API interaction
//   - InstanceInfo: Data structure for instance metadata
//   - ArchitectureMapping: Container tag mapping configuration
//
// Usage:
//   discoverer, err := discovery.NewInstanceDiscoverer()
//   instances, err := discoverer.DiscoverAllInstanceTypes(ctx)
//   mappings := discoverer.GenerateArchitectureMappings(instances)
//
// The package automatically handles:
//   - AWS SDK v2 authentication via default profile
//   - Pagination for large instance type lists
//   - Architecture detection (Intel, AMD, Graviton)
//   - Container tag assignment based on microarchitecture
package discovery
```

### **3. Complex Function Documentation**
For functions with 3+ parameters or complex logic:

```go
// BuildContainer orchestrates the complete container build process for a specific
// architecture and benchmark suite combination.
//
// This function performs the following steps:
//   1. Creates isolated build directory structure
//   2. Generates architecture-optimized Dockerfile from templates
//   3. Copies required Spack configuration files
//   4. Executes Docker build with proper context
//   5. Tags container with registry-compatible naming
//
// The build process uses multi-stage Dockerfiles to minimize final image size
// while maintaining all necessary compilation artifacts for benchmark execution.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - config: BuildConfig containing:
//     * InstanceType: AWS instance type (e.g., "m7i.large")
//     * ContainerTag: Architecture tag (e.g., "intel-icelake")
//     * BenchmarkSuite: Benchmark to build (e.g., "stream")
//     * CompilerType: Compiler optimization ("intel", "amd", "gcc")
//     * OptimizationFlags: Architecture-specific compiler flags
//     * BaseImage: Container base image
//     * SpackConfig: Spack environment configuration file
//
// Returns:
//   - error: nil on success, detailed error on failure
//
// Build Directory Structure:
//   builds/
//   ├── {container-tag}/
//   │   └── {benchmark}/
//   │       ├── Dockerfile
//   │       └── spack-configs/
//
// Example:
//   config := BuildConfig{
//       ContainerTag: "intel-icelake",
//       BenchmarkSuite: "stream",
//       CompilerType: "intel",
//       OptimizationFlags: []string{"-O3", "-xCORE-AVX512"},
//   }
//   err := builder.BuildContainer(ctx, config)
//
// Common Errors:
//   - Docker daemon not running
//   - Insufficient disk space for build
//   - Network connectivity issues during dependency download
//   - Invalid Spack configuration syntax
func (b *Builder) BuildContainer(ctx context.Context, config BuildConfig) error {
```

### **4. Struct Documentation**
```go
// InstanceResult contains the complete execution results and metadata for a 
// benchmark run on a specific AWS EC2 instance.
//
// This structure captures both the operational aspects (instance lifecycle)
// and the benchmark data (performance metrics) from a single benchmark execution.
// Results are used for data aggregation, cost analysis, and performance ranking.
//
// Lifecycle States:
//   - StartTime set when benchmark request begins
//   - InstanceID populated after successful EC2 launch
//   - PublicIP/PrivateIP set when instance reaches running state
//   - BenchmarkData populated after benchmark execution
//   - EndTime set when instance is terminated
//   - Error set if any step fails
//
// Example Usage:
//   result, err := orchestrator.RunBenchmark(ctx, config)
//   if result.Error == nil {
//       duration := result.EndTime.Sub(result.StartTime)
//       bandwidth := result.BenchmarkData["stream"].(map[string]interface{})["triad"]
//   }
type InstanceResult struct {
    // InstanceID is the AWS EC2 instance identifier assigned during launch.
    // Format: "i-1234567890abcdef0"
    InstanceID string
    
    // InstanceType is the AWS instance type used for benchmarking.
    // Examples: "m7i.large", "c7g.xlarge", "r7a.2xlarge"
    InstanceType string
    
    // PublicIP is the internet-accessible IP address assigned to the instance.
    // May be empty if instance is in private subnet.
    PublicIP string
    
    // PrivateIP is the VPC-internal IP address of the instance.
    PrivateIP string
    
    // Status indicates the current state of the benchmark execution.
    // Values: "launching", "running", "completed", "failed"
    Status string
    
    // BenchmarkData contains the structured performance results from benchmark execution.
    // Format varies by benchmark suite:
    //   - STREAM: {"copy": {"bandwidth": 45.2, "unit": "GB/s"}, ...}
    //   - HPL: {"gflops": 123.4, "efficiency": 0.85, ...}
    BenchmarkData map[string]interface{}
    
    // Error contains any error encountered during benchmark execution.
    // Common errors: quota exceeded, instance launch failure, benchmark timeout
    Error error
    
    // StartTime marks when the benchmark request was initiated.
    StartTime time.Time
    
    // EndTime marks when the instance was terminated and results collected.
    EndTime time.Time
}
```

## 🔧 Enforcement Mechanisms

### **1. Automated Linting (Primary)**
- **golangci-lint** with revive rules enforces documentation
- **CI/CD integration** blocks merges without proper documentation
- **Custom rules** for project-specific requirements

### **2. Pre-commit Hooks**
- Validates documentation before commits
- Checks for TODO/FIXME without issue references
- Spell-checks comments and documentation

### **3. Documentation Coverage Metrics**
```bash
# Check documentation coverage
go run tools/doc-coverage.go
# Output: 87% of exported functions documented (target: 100%)
```

### **4. Reviewer Guidelines**
All code reviews must verify:
- [ ] All exported functions documented
- [ ] Complex functions have detailed explanations
- [ ] Comments explain "why" not just "what"
- [ ] Examples provided for non-trivial usage
- [ ] Error conditions documented

## 📝 Documentation Templates

### **Function Template**
```go
// FunctionName does X by performing Y and Z operations.
//
// Detailed explanation of the function's purpose, algorithm, or business logic.
// Include any important implementation details, assumptions, or constraints.
//
// Parameters:
//   - param1: Description of first parameter and expected values
//   - param2: Description of second parameter and constraints
//
// Returns:
//   - returnType: Description of return value and possible states
//   - error: Description of error conditions and types
//
// Example:
//   result, err := FunctionName(param1, param2)
//   if err != nil { ... }
//
// Common Errors:
//   - ErrInvalidInput: when param1 is nil
//   - ErrNetworkTimeout: when operation exceeds deadline
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error) {
```

### **Package Template**
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
//
// The package handles:
//   - Automatic retry logic with exponential backoff
//   - Connection pooling and resource management
//   - Error categorization and recovery strategies
//
// Configuration:
//   See Config struct for available options and defaults.
//   Most users can use packagename.NewDefault() for standard setups.
//
// Thread Safety:
//   All exported functions are safe for concurrent use unless noted.
//   Individual instances are not thread-safe and require external synchronization.
package packagename
```

## 🎯 Quality Metrics

### **Documentation Coverage Targets**
- **100%** of exported functions documented
- **100%** of packages have package-level documentation
- **90%** of complex functions (3+ params) have examples
- **85%** of error conditions documented

### **Comment Quality Indicators**
- Comments explain business logic, not just syntax
- Examples demonstrate real-world usage patterns
- Error conditions and recovery strategies documented
- Performance characteristics noted for critical paths

### **Automation Checks**
- Spell-check all comments and documentation
- Verify code examples actually compile
- Check that documentation stays synchronized with code
- Validate godoc generation produces clean output

## 🚀 Developer Onboarding Benefits

With comprehensive documentation:

1. **Faster Ramp-up**: New developers understand code purpose immediately
2. **Reduced Context Switching**: Less time asking questions, more time coding
3. **Better Code Reviews**: Reviewers can focus on logic, not understanding
4. **Maintainability**: Future changes consider documented contracts
5. **API Stability**: Documentation prevents accidental breaking changes

## 📊 Measuring Success

Track documentation quality with:
- Documentation coverage percentage
- Time for new developers to make first contribution
- Number of documentation-related questions in reviews
- godoc page views and user feedback
- Code maintainability scores over time