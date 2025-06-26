# Contributing to AWS Instance Benchmarks

Thank you for your interest in contributing to the AWS Instance Benchmarks project! This document provides comprehensive guidelines for contributing code, documentation, and data to help maintain our high-quality standards.

## 🎯 Project Mission

We're building an open, community-driven database of comprehensive performance benchmarks for AWS EC2 instances that enables data-driven instance selection for research computing workloads.

## 📋 How to Contribute

### **Types of Contributions Welcome**
- 🐛 **Bug Reports**: Issues with benchmark execution or data accuracy
- ✨ **Feature Requests**: New benchmark suites or analysis capabilities
- 📝 **Documentation**: Improvements to guides, examples, and API docs
- 🔧 **Code Contributions**: Implementation of new features or bug fixes
- 📊 **Benchmark Data**: Submission of benchmark results for new instance types
- 🧪 **Testing**: Additional test cases and validation scenarios

## 🚀 Getting Started

### **Prerequisites**
- Go 1.21+ for code contributions
- AWS CLI v2 configured with appropriate permissions
- Docker/Podman for container operations
- Git with SSH keys configured for GitHub

### **Development Setup**
```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/aws-instance-benchmarks.git
cd aws-instance-benchmarks

# 3. Add upstream remote
git remote add upstream https://github.com/scttfrdmn/aws-instance-benchmarks.git

# 4. Install dependencies
go mod tidy

# 5. Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
pip install pre-commit

# 6. Set up pre-commit hooks
pre-commit install

# 7. Verify setup
go test ./... -v
./scripts/check-function-docs.sh
golangci-lint run
```

### **Development Workflow**
```bash
# 1. Create feature branch
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name

# 2. Make changes following our standards
# (See Documentation and Testing sections below)

# 3. Run quality checks
go test ./... -v -race -coverprofile=coverage.out
./scripts/check-function-docs.sh
golangci-lint run

# 4. Commit changes
git add .
git commit -m "feat: add STREAM benchmark execution

- Implement NUMA-aware STREAM benchmark
- Add statistical validation with confidence intervals
- Integrate S3 result storage
- Add comprehensive test coverage

Fixes #123"

# 5. Push and create PR
git push origin feature/your-feature-name
# Create PR on GitHub with detailed description
```

## 📖 Code Standards

### **Documentation Requirements (MANDATORY)**
Every exported function must have comprehensive documentation:

```go
// FunctionName performs X operation by implementing Y algorithm.
//
// Detailed explanation of purpose, algorithm, and business logic.
// Include assumptions, constraints, and important implementation details.
//
// Parameters:
//   - param1: Description with expected values and constraints
//   - param2: Description with validation requirements
//
// Returns:
//   - returnType: Description of return value and states
//   - error: Description of error conditions and types
//
// Example:
//   result, err := FunctionName(param1, param2)
//   if err != nil {
//       return fmt.Errorf("operation failed: %w", err)
//   }
//
// Performance Notes:
//   - Time complexity: O(n) where n is input size
//   - Memory usage: ~2MB for typical operations
//
// Common Errors:
//   - ErrInvalidInput: when param1 is nil
//   - ErrNetworkTimeout: when operation exceeds deadline
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error) {
```

### **Testing Requirements**
- **85%+ test coverage** for all packages
- **100% coverage** for exported functions
- **Unit tests** for all public APIs
- **Integration tests** for AWS interactions
- **Error path testing** for resilience

```go
func TestFunctionName(t *testing.T) {
    testCases := []struct {
        name        string
        input       InputType
        expected    ExpectedType
        expectError bool
    }{
        {
            name:     "successful operation",
            input:    validInput,
            expected: expectedOutput,
        },
        {
            name:        "invalid input error",
            input:       invalidInput,
            expectError: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := FunctionName(tc.input)
            
            if tc.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### **Code Quality Standards**
- **golangci-lint** must pass with project configuration
- **Pre-commit hooks** must pass for all changes
- **Error handling** must be comprehensive with proper wrapping
- **Performance** considerations for critical paths
- **Security** validation for all inputs and AWS operations

## 🐛 Bug Reports

### **Before Submitting**
- Search existing issues to avoid duplicates
- Test with the latest version
- Gather relevant system information

### **Bug Report Template**
```markdown
## Bug Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., macOS 13.0, Ubuntu 22.04]
- Go version: [output of `go version`]
- AWS Region: [e.g., us-east-1]
- Tool version: [output of `./aws-benchmark-collector --version`]

## Additional Context
- Error messages and logs
- Configuration files (with sensitive data removed)
- Relevant AWS account limits or constraints
```

## ✨ Feature Requests

### **Feature Request Template**
```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed? What problem does it solve?

## Proposed Solution
Detailed description of how you envision this working

## Alternatives Considered
Other approaches you've considered

## Implementation Notes
Technical considerations, dependencies, or constraints

## Additional Context
Screenshots, mockups, or examples
```

## 📊 Contributing Benchmark Data

### **Data Submission Guidelines**
- **Reproducible methodology** with documented configuration
- **Multiple runs** with statistical validation (minimum 10 runs)
- **Confidence intervals** at 95% level
- **Architecture optimization** appropriate for instance type
- **Complete metadata** including region, AMI, and environment

### **Data Format Requirements**
```json
{
  "metadata": {
    "instanceType": "m7i.large",
    "region": "us-east-1",
    "availabilityZone": "us-east-1a",
    "amiId": "ami-12345678",
    "timestamp": "2024-01-15T10:30:00Z",
    "submitter": "your-github-username"
  },
  "benchmark": {
    "suite": "stream",
    "version": "5.10",
    "compiler": "gcc-11",
    "optimizationFlags": ["-O3", "-march=native"],
    "runs": 10,
    "confidence": 0.95
  },
  "results": {
    "stream": {
      "copy": {"bandwidth": 45.2, "stddev": 0.8, "unit": "GB/s"},
      "scale": {"bandwidth": 44.8, "stddev": 0.7, "unit": "GB/s"},
      "add": {"bandwidth": 42.1, "stddev": 0.9, "unit": "GB/s"},
      "triad": {"bandwidth": 41.9, "stddev": 0.8, "unit": "GB/s"}
    }
  }
}
```

### **Validation Process**
1. **Automated validation** against JSON schema
2. **Statistical analysis** for outlier detection
3. **Peer review** by community members
4. **Integration testing** with existing dataset
5. **Approval** by project maintainers

## 🔍 Code Review Process

### **Pull Request Guidelines**
- **Descriptive title** summarizing the change
- **Detailed description** explaining motivation and implementation
- **Link to issues** that are addressed
- **Breaking changes** clearly documented
- **Testing** evidence provided

### **PR Description Template**
```markdown
## Summary
Brief description of changes

## Motivation
Why is this change needed?

## Changes Made
- Detailed list of changes
- New features or improvements
- Bug fixes

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing performed
- [ ] Documentation updated

## Breaking Changes
List any breaking changes and migration steps

## Checklist
- [ ] Code follows project standards
- [ ] Documentation coverage at 100%
- [ ] Tests pass with 85%+ coverage
- [ ] Pre-commit hooks pass
- [ ] No security vulnerabilities introduced
```

### **Review Criteria**
- **Functionality**: Does the code work as intended?
- **Documentation**: Are all functions properly documented?
- **Testing**: Is test coverage adequate and meaningful?
- **Performance**: Are there any performance implications?
- **Security**: Are there any security concerns?
- **Maintainability**: Is the code readable and well-structured?

## 🏗️ Development Guidelines

### **Package Organization**
- **Single responsibility** per package
- **Clear interfaces** between packages
- **Minimal dependencies** between internal packages
- **Consistent naming** following Go conventions

### **Error Handling**
```go
// Define package-specific error types
var (
    ErrInvalidInput = errors.New("invalid input provided")
    ErrQuotaExceeded = errors.New("AWS quota exceeded")
)

// Use error wrapping for context
func ProcessData(data []byte) error {
    if len(data) == 0 {
        return fmt.Errorf("ProcessData: %w", ErrInvalidInput)
    }
    return nil
}
```

### **Configuration Management**
- **Environment variables** for runtime configuration
- **YAML files** for complex configuration
- **Sensible defaults** for all optional settings
- **Validation** for all configuration values

## 🚦 CI/CD and Quality Gates

### **Automated Checks**
All pull requests must pass:
- **Unit tests** with 85%+ coverage
- **Integration tests** (if applicable)
- **golangci-lint** with project configuration
- **Documentation coverage** at 100%
- **Security scanning** for vulnerabilities
- **Build verification** for all supported platforms

### **Manual Review Requirements**
- **Code review** by at least one maintainer
- **Architecture review** for significant changes
- **Security review** for AWS integration changes
- **Performance review** for critical path modifications

## 📞 Getting Help

### **Community Support**
- **GitHub Discussions**: General questions and brainstorming
- **GitHub Issues**: Bug reports and feature requests
- **Documentation**: Comprehensive guides in `/docs`
- **Examples**: Working examples in repository

### **Maintainer Contact**
- **GitHub**: @scttfrdmn for project-related questions
- **Email**: benchmarks@computecompass.dev for private inquiries

### **Response Times**
- **Bug reports**: 48-72 hours for initial response
- **Feature requests**: 1 week for evaluation
- **Pull requests**: 3-5 days for review
- **Security issues**: 24 hours for acknowledgment

## 🎉 Recognition

### **Contributor Recognition**
- **Contributors file** listing all contributors
- **Release notes** crediting significant contributions
- **GitHub badges** for various contribution types
- **Conference presentations** acknowledging community contributions

### **Long-term Collaboration**
Outstanding contributors may be invited to become:
- **Project collaborators** with enhanced permissions
- **Maintainers** helping guide project direction
- **Advisory board members** for strategic decisions

## 📜 Code of Conduct

We are committed to providing a welcoming and inclusive experience for everyone. Please review our [Code of Conduct](CODE_OF_CONDUCT.md) for details on our community standards.

## 📄 License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (MIT License for code, CC BY 4.0 for data).

---

**Thank you for contributing to AWS Instance Benchmarks!** 🚀

Your contributions help the research computing community make better instance selection decisions based on real performance data.