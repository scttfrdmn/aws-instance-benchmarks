# Pull Request Submission Plan

## Overview
This document outlines the strategy for submitting the comprehensive improvements made to aws-instance-benchmarks as organized pull requests.

## PR Structure (Recommended Order)

### PR #1: Fix Dependencies and Core Issues 🔧
**Branch**: `fix/dependencies-and-core-issues`
**Priority**: High (foundational fixes)

**Changes:**
- Add missing `github.com/google/uuid v1.6.0` dependency
- Fix storage package test interface mismatches
- Resolve build system conflicts (main function conflicts)
- Clean up async components (disable incomplete implementations)

**Files:**
- `go.mod`, `go.sum`
- `pkg/storage/s3_test.go`
- Move/disable async files
- Move `cmd/analyze_price_performance.go`

### PR #2: Enhanced Configuration System 📋
**Branch**: `feature/enhanced-configuration-system`
**Priority**: Medium
**Depends on**: PR #1

**Changes:**
- Extract hardcoded S3 bucket references to configuration
- Add comprehensive instance family classification
- Expand default instance type coverage (27 → 35+ types)
- Enhance benchmark configuration options

**Files:**
- `configs/aws-infrastructure.json`
- `async_benchmark_launcher.go` → config integration
- `async_benchmark_collector.go` → config integration

### PR #3: Graviton4 Support and Architecture Optimization 🚀
**Branch**: `feature/graviton4-architecture-support`
**Priority**: Medium
**Depends on**: PR #1

**Changes:**
- Add Graviton4 (c8g, m8g, r8g) instance family support
- Update architecture mappings with processor-specific information
- Create optimized Graviton4 build toolchain
- Add architecture-specific compilation flags

**Files:**
- `configs/architecture-mappings.json`
- `builds/graviton4/stream/` (new directory)
- Architecture-specific optimizations

### PR #4: Expanded Instance Coverage 📊
**Branch**: `feature/expanded-instance-coverage`
**Priority**: Medium
**Depends on**: PR #2, PR #3

**Changes:**
- Add support for older generation instances (c4, m4, r4)
- Add specialized instance types (i3, i4g, z1d, t4g)
- Historical baseline support for performance comparison
- Enhanced instance family organization

**Files:**
- Configuration updates for new instance types
- Documentation updates

### PR #5: Documentation and Improvements Summary 📚
**Branch**: `docs/comprehensive-improvements-2025`
**Priority**: Low
**Depends on**: All previous PRs

**Changes:**
- Update README.md with new capabilities
- Add comprehensive improvements documentation
- Update architecture coverage information
- Add usage examples for new features

**Files:**
- `README.md`
- `IMPROVEMENTS_2025.md`
- Architecture documentation

## Git Commands for Each PR

### Setup Commands (Run First)
```bash
# Ensure we're on main and up to date
git checkout main
git pull origin main

# Create backup of current changes
git stash push -m "All improvements for PR preparation"
```

### PR #1: Dependencies and Core Fixes
```bash
git checkout -b fix/dependencies-and-core-issues

# Add dependency fixes
git add go.mod go.sum
git commit -m "fix: add missing github.com/google/uuid dependency

- Resolves build failures in async packages
- Updates go.mod and go.sum with correct dependency versions"

# Fix storage tests
git add pkg/storage/s3_test.go
git commit -m "fix: correct storage package test interface signatures

- Update NewS3Storage test calls to include required region parameter
- Ensures tests pass with proper AWS credential error handling"

# Clean up build conflicts
git add -A  # This will stage deletions and moves
git commit -m "fix: resolve main function conflicts and clean build system

- Move analyze_price_performance.go to separate tool
- Temporarily disable incomplete async implementations
- Ensures clean build of main CLI tool"

git push origin fix/dependencies-and-core-issues
```

### PR #2: Configuration System
```bash
git checkout main
git checkout -b feature/enhanced-configuration-system

# Apply configuration changes
git stash pop  # Get our changes back
# ... stage configuration files
git add configs/aws-infrastructure.json
git commit -m "feat: enhance configuration system with comprehensive instance coverage

- Add instance family classification (compute, memory, storage, accelerated)
- Expand default instance types from 27 to 35+ types
- Add support for current and previous generation instances
- Increase default iterations to 3 for statistical rigor"

# Configuration integration
# ... stage async tool updates (re-enable with config integration)
git commit -m "feat: replace hardcoded S3 bucket references with configuration-driven approach

- Load S3 bucket from aws-infrastructure.json config file
- Support multi-region bucket configuration
- Eliminate hardcoded bucket names in async tools"

git push origin feature/enhanced-configuration-system
```

### PR #3: Graviton4 Support
```bash
git checkout main
git checkout -b feature/graviton4-architecture-support

# Architecture mappings
git add configs/architecture-mappings.json
git commit -m "feat: add AWS Graviton4 processor support

- Add c8g, m8g, r8g instance families with Graviton4 optimization
- Update architecture mappings with processor-specific information
- Replace generic tags with precise processor generation labels"

# Build toolchain
git add builds/graviton4/
git commit -m "feat: create Graviton4-optimized build toolchain

- Add specialized Dockerfile for ARM Neoverse-V2 compilation
- Include SVE support and advanced optimization flags
- Create Spack configuration for architecture-specific builds
- Add NUMA-aware benchmark execution scripts"

git push origin feature/graviton4-architecture-support
```

### Continue for remaining PRs...

## PR Descriptions Templates

### PR #1 Description:
```markdown
## Fix Dependencies and Core Build Issues

### Summary
Resolves critical build and test failures that prevent the project from compiling and running tests successfully.

### Changes
- ✅ **Dependencies**: Add missing `github.com/google/uuid v1.6.0`
- ✅ **Tests**: Fix storage package interface mismatches
- ✅ **Build System**: Resolve main function conflicts
- ✅ **Cleanup**: Temporarily disable incomplete async implementations

### Impact
- Enables successful compilation of main CLI tool
- All core package tests now pass
- Clean foundation for future feature development

### Testing
```bash
go build ./cmd  # Now succeeds
go test ./pkg/storage  # All tests pass
```
```

## Next Steps
1. Review this plan
2. Execute git commands in order
3. Submit PRs one by one
4. Address review feedback
5. Merge in dependency order

This structured approach ensures:
- ✅ **Reviewable chunks**: Each PR has clear scope
- ✅ **Logical dependencies**: Build fixes first, features second
- ✅ **Clear impact**: Each PR solves specific problems
- ✅ **Maintainable**: Future updates can target specific areas