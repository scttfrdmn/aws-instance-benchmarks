# Architecture-Aware Container Build Strategy

## Overview

This document outlines our strategy for building high-performance benchmark containers that are optimized for specific CPU architectures. Instead of cross-compiling, we build containers directly on target hardware using optimal toolchains to achieve maximum performance.

## Why Architecture-Aware Builds?

### Performance Impact

Building on the target architecture with optimal compilers can yield significant performance improvements:

- **Intel Ice Lake**: Intel OneAPI compiler with `-xHost` can be 15-30% faster than generic GCC
- **AMD Zen4**: AOCC (AMD Optimizing C/C++ Compiler) with `-march=znver4` optimizes for Zen4 microarchitecture
- **ARM Graviton**: Native ARM builds avoid emulation overhead and enable ARM-specific SIMD optimizations
- **Architecture-Specific Features**: AVX-512, SVE, NEON, and other SIMD instructions are properly utilized

### Benchmark Accuracy

- **True Performance**: Results reflect actual hardware performance, not emulated or cross-compiled approximations
- **Cache Optimization**: Compilers can optimize for specific L1/L2/L3 cache hierarchies
- **Memory Patterns**: Architecture-specific memory access patterns are optimized
- **Instruction Scheduling**: CPU-specific instruction scheduling optimizations

## Build Architecture

### Instance Type Mapping

| Architecture | Instance Type | AMI Type | Optimal Compiler |
|-------------|---------------|----------|------------------|
| Universal   | t3.medium     | x86_64   | GCC (conservative) |
| Intel Ice Lake | m7i.large   | x86_64   | Intel OneAPI ICC |
| AMD Zen4    | m7a.large     | x86_64   | AMD AOCC |
| Graviton3   | m7g.large     | arm64    | GCC ARM-optimized |
| Graviton4   | m8g.large     | arm64    | GCC ARM-optimized |

### Compiler Optimization Flags

#### Intel Ice Lake (OneAPI ICC)
```bash
export CC=icc
export CXX=icpc
export CFLAGS="-O3 -xHost -ipo -no-prec-div -fp-model fast=2"
export CXXFLAGS="-O3 -xHost -ipo -no-prec-div -fp-model fast=2"
```

- `-xHost`: Generate code optimized for the host processor
- `-ipo`: Inter-procedural optimization
- `-no-prec-div`: Faster division (acceptable for benchmarks)
- `-fp-model fast=2`: Aggressive floating-point optimizations

#### AMD Zen4 (AOCC)
```bash
export CC=clang
export CXX=clang++
export CFLAGS="-O3 -march=znver4 -mtune=znver4 -flto"
export CXXFLAGS="-O3 -march=znver4 -mtune=znver4 -flto"
```

- `-march=znver4`: Target Zen4 microarchitecture specifically
- `-mtune=znver4`: Tune instruction scheduling for Zen4
- `-flto`: Link-time optimization

#### ARM Graviton (GCC)
```bash
export CC=gcc
export CXX=g++
export CFLAGS="-O3 -march=native -mtune=native -mcpu=native -flto"
export CXXFLAGS="-O3 -march=native -mtune=native -mcpu=native -flto"
```

- `-mcpu=native`: Optimize for the specific ARM CPU
- `-march=native`: Use all available instruction set extensions
- `-mtune=native`: Tune for the specific CPU implementation

#### Universal (Conservative GCC)
```bash
export CC=gcc
export CXX=g++
export CFLAGS="-O3 -march=native -mtune=native"
export CXXFLAGS="-O3 -march=native -mtune=native"
```

- Safe optimizations that work across multiple architectures
- Still uses native optimizations for the build host

## Build Process

### 1. Launch Architecture-Specific Instances

The `build-on-target.sh` script launches EC2 instances of each target architecture:

```bash
# Build all architectures
./scripts/build-on-target.sh all stream

# Build specific architecture
./scripts/build-on-target.sh intel-icelake stream
```

### 2. Install Optimal Toolchains

Each instance installs the optimal compiler for its architecture:

- **Intel**: Downloads and installs Intel OneAPI Base + HPC Toolkit
- **AMD**: Downloads and installs AMD AOCC compiler suite  
- **ARM**: Uses optimized GCC with ARM-specific flags
- **Universal**: Uses system GCC with safe optimizations

### 3. Build with Architecture Awareness

Containers are built on the target hardware with:

- CPU-specific optimization flags
- Architecture-aware library linking
- Native instruction set utilization
- Optimal cache and memory patterns

### 4. Validation and Distribution

Built containers are:

1. **Tested**: Quick validation run to ensure functionality
2. **Packaged**: Saved as compressed tarballs
3. **Uploaded**: Stored in S3 for distribution
4. **Loaded**: Imported into local Podman/Docker registry
5. **Cleaned**: Build instances auto-terminate after completion

## Container Registry Strategy

### Local Development
```
localhost/aws-instance-benchmarks/stream:intel-icelake
localhost/aws-instance-benchmarks/stream:amd-zen4
localhost/aws-instance-benchmarks/stream:graviton3
localhost/aws-instance-benchmarks/stream:graviton4
localhost/aws-instance-benchmarks/stream:universal
```

### Production ECR
```
public.ecr.aws/aws-benchmarks/stream:intel-icelake
public.ecr.aws/aws-benchmarks/stream:amd-zen4
public.ecr.aws/aws-benchmarks/stream:graviton3
public.ecr.aws/aws-benchmarks/stream:graviton4
public.ecr.aws/aws-benchmarks/stream:universal
```

## Container Selection Logic

The orchestrator selects containers based on EC2 instance characteristics:

```go
func getContainerTagForInstance(instanceType string) string {
    // Intel Ice Lake instances
    if strings.HasPrefix(instanceType, "m7i") || strings.HasPrefix(instanceType, "c7i") || strings.HasPrefix(instanceType, "r7i") {
        return "intel-icelake"
    }
    
    // AMD Zen4 instances
    if strings.HasPrefix(instanceType, "m7a") || strings.HasPrefix(instanceType, "c7a") || strings.HasPrefix(instanceType, "r7a") {
        return "amd-zen4"
    }
    
    // Graviton3 instances
    if strings.HasPrefix(instanceType, "m7g") || strings.HasPrefix(instanceType, "c7g") || strings.HasPrefix(instanceType, "r7g") {
        return "graviton3"
    }
    
    // Graviton4 instances
    if strings.HasPrefix(instanceType, "m8g") || strings.HasPrefix(instanceType, "c8g") || strings.HasPrefix(instanceType, "r8g") {
        return "graviton4"
    }
    
    // Universal fallback
    return "universal"
}
```

## Performance Expectations

Based on architecture-specific optimizations, we expect:

### STREAM Memory Bandwidth
- **Intel OneAPI**: 10-20% improvement over generic GCC
- **AMD AOCC**: 15-25% improvement with Zen4 optimizations
- **ARM Native**: 20-30% improvement over x86 cross-compilation

### HPL Linear Algebra
- **Intel MKL**: 30-50% improvement with optimized BLAS/LAPACK
- **AMD LibM**: 20-35% improvement with AOCC math libraries
- **ARM Performance Libraries**: 25-40% improvement with Arm Compute Library

### Cost Efficiency
- Higher performance per dollar due to optimal hardware utilization
- Reduced benchmark runtime = lower EC2 costs
- More accurate price/performance comparisons

## Maintenance and Updates

### Compiler Updates
- **Quarterly**: Check for new compiler versions
- **Automated**: Update build scripts with new optimization flags
- **Tested**: Validate performance improvements/regressions

### Architecture Support
- **New Instances**: Add support for new EC2 instance families
- **New CPUs**: Update compiler targets for new microarchitectures
- **Performance Tuning**: Continuous optimization of build flags

### Container Lifecycle
- **Daily Builds**: Automated container builds for active development
- **Release Tags**: Versioned containers for stable benchmarks
- **Cleanup**: Regular cleanup of old container versions

## Security Considerations

### Build Environment
- **Isolated Instances**: Each build runs in a fresh EC2 instance
- **Auto-Termination**: Instances self-destruct after build completion
- **IAM Roles**: Minimal permissions for S3 upload and SSM communication

### Container Security
- **Base Images**: Use minimal, security-patched base images
- **Vulnerability Scanning**: Regular security scans of container images
- **Supply Chain**: Verify compiler and library sources

## Monitoring and Observability

### Build Metrics
- **Build Time**: Track compilation times per architecture
- **Success Rate**: Monitor build failure rates
- **Performance**: Compare benchmark results across compiler versions

### Cost Tracking
- **Instance Costs**: Monitor EC2 costs per container build
- **Storage Costs**: Track S3 storage for container images
- **Efficiency**: Cost per successful container build

This architecture-aware approach ensures our benchmark results accurately reflect real-world performance characteristics while providing optimal efficiency for cost-conscious research computing workloads.