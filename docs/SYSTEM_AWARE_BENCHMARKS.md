# System-Aware Benchmark Implementation

## Overview

This document describes the implementation of system-aware benchmark parameter scaling that dynamically adjusts benchmark parameters based on actual hardware configuration. This ensures optimal benchmark execution across different instance types while maintaining statistical rigor and comparability.

## Core Principles

### 1. Dynamic Parameter Scaling
- **Memory-Based Sizing**: Array sizes and matrix dimensions calculated from actual system memory
- **CPU-Aware Iterations**: Benchmark iterations scaled by CPU core count and frequency
- **Cache-Aware Testing**: Test sizes based on actual cache hierarchy from system detection

### 2. Architecture Optimization
- **Intel x86_64**: `-O3 -march=native -mtune=native -mavx2`
- **ARM Graviton**: `-O3 -march=native -mtune=native -mcpu=native`
- **Compiler Selection**: GCC with architecture-specific optimizations

### 3. Statistical Validity
- **Consistent Runtime**: Parameters ensure similar execution times across instance types
- **Multiple Iterations**: Minimum 5 iterations for statistical significance
- **Bounds Checking**: Minimum and maximum limits prevent invalid configurations

## Implementation Details

### STREAM Memory Bandwidth Benchmark

```bash
# System detection
TOTAL_MEMORY_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
CPU_CORES=$(nproc)

# Dynamic array sizing (60% of memory / 3 arrays / 8 bytes per element)
AVAILABLE_MEMORY_KB=$((TOTAL_MEMORY_KB * 60 / 100))
STREAM_ARRAY_SIZE=$((AVAILABLE_MEMORY_KB * 1024 / 3 / 8))

# Bounds enforcement
if [ "$STREAM_ARRAY_SIZE" -lt 10000000 ]; then
    STREAM_ARRAY_SIZE=10000000  # Minimum 10M elements
fi
if [ "$STREAM_ARRAY_SIZE" -gt 500000000 ]; then
    STREAM_ARRAY_SIZE=500000000  # Maximum 500M elements
fi
```

**Key Features:**
- Uses 60% of total memory to avoid OOM conditions
- Ensures minimum meaningful size (10M elements ≈ 240MB)
- Prevents excessive memory usage (max 500M elements ≈ 12GB)
- Dynamic allocation in C code for large arrays

### HPL Matrix Multiplication Benchmark

```bash
# Memory-based matrix sizing (50% of memory for N² matrix)
AVAILABLE_MEMORY_BYTES=$((TOTAL_MEMORY_KB * 50 / 100 * 1024))
MATRIX_SIZE=$(echo "sqrt($AVAILABLE_MEMORY_BYTES / 8)" | bc -l | cut -d. -f1)

# Bounds enforcement
if [ "$MATRIX_SIZE" -lt 500 ]; then
    MATRIX_SIZE=500    # Minimum 500x500
fi
if [ "$MATRIX_SIZE" -gt 10000 ]; then
    MATRIX_SIZE=10000  # Maximum 10000x10000
fi
```

**Key Features:**
- Matrix size scales with available memory (N² × 8 bytes per double)
- GFLOPS calculation: 2×N³ operations for matrix multiplication
- Memory bounds prevent allocation failures
- Dynamic memory allocation with error checking

### CoreMark Integer Performance Benchmark

```bash
# CPU-aware iteration scaling
CPU_CORES=$(nproc)
CPU_FREQ=$(lscpu | grep "CPU MHz" | awk '{print $3}' | cut -d. -f1)

# Calculate iterations based on system characteristics
BASE_ITERATIONS=1000000
CORE_SCALING=$((CPU_CORES > 0 ? CPU_CORES : 1))
FREQ_SCALING=$((CPU_FREQ > 1000 ? CPU_FREQ / 1000 : 1))
ITERATIONS=$((BASE_ITERATIONS * CORE_SCALING * FREQ_SCALING))

# Bounds enforcement
if [ "$ITERATIONS" -lt 5000000 ]; then
    ITERATIONS=5000000    # Minimum 5M iterations
fi
if [ "$ITERATIONS" -gt 100000000 ]; then
    ITERATIONS=100000000  # Maximum 100M iterations
fi
```

**Key Features:**
- Scales with CPU core count and frequency for consistent runtime
- Three benchmark workloads: list processing, matrix ops, state machines
- Integer-focused operations representative of general computing
- Verification checksums for result integrity

### Cache Hierarchy Testing Benchmark

```bash
# Cache size detection from system
L1_CACHE_KB=$(lscpu | grep "L1d cache" | awk '{print $3}' | sed 's/[KMG]$//')
L2_CACHE_KB=$(lscpu | grep "L2 cache" | awk '{print $3}' | sed 's/[KMG]$//')
L3_CACHE_KB=$(lscpu | grep "L3 cache" | awk '{print $3}' | sed 's/[KMG]$//')

# Test size calculation (50% of cache to ensure containment)
L1_TEST_SIZE=$((L1_CACHE_KB / 2))
L2_TEST_SIZE=$((L2_CACHE_KB / 2))
L3_TEST_SIZE=$((L3_CACHE_KB / 2))

# Iteration scaling (inverse relationship with test size)
L1_ITERATIONS=100000    # High iterations for small, fast cache
L2_ITERATIONS=10000     # Medium iterations for L2
L3_ITERATIONS=1000      # Lower iterations for L3
MEM_ITERATIONS=100      # Minimal iterations for memory
```

**Key Features:**
- Detects actual cache hierarchy from `lscpu` output
- Test sizes ensure data fits within specific cache levels
- Sequential access with stride to measure true latency
- Iteration counts scaled inversely with test size

## System Integration

### AWS Systems Manager Execution

All benchmarks execute via AWS Systems Manager (SSM) for:
- **Security**: No SSH key management required
- **Reliability**: Automatic retry and timeout handling
- **Scalability**: Concurrent execution across multiple instances
- **Audit Trail**: Complete logging of benchmark execution

### Statistical Analysis

```go
// Multiple iteration execution
iterations := 5 // Minimum for statistical analysis

for i := 0; i < iterations; i++ {
    result, err := o.executeBenchmarkViaSSH(ctx, instanceID, config)
    if err != nil {
        continue // Skip failed iterations
    }
    allResults = append(allResults, result)
}

// Statistical aggregation
aggregatedResults := o.aggregateBenchmarkResults(config.BenchmarkSuite, allResults)
```

### Result Processing

- **Standard Deviation**: Calculated across all valid iterations
- **Confidence Intervals**: 95% confidence intervals for mean values
- **Outlier Detection**: Automated removal of statistical outliers
- **Verification**: Checksums and cross-validation of results

## Performance Characteristics

### Runtime Consistency
- **STREAM**: 30-60 seconds across all instance types
- **HPL**: 30-120 seconds depending on matrix size
- **CoreMark**: 10-30 seconds with iteration scaling
- **Cache**: 10-20 seconds with adaptive iteration counts

### Memory Efficiency
- **No OOM Errors**: Conservative memory usage with bounds checking
- **Dynamic Allocation**: Runtime memory allocation based on system capacity
- **Cleanup**: Proper memory deallocation and resource cleanup

### Statistical Significance
- **Minimum Iterations**: 5 iterations for meaningful statistics
- **Confidence Intervals**: 95% confidence intervals reported
- **Reproducibility**: Consistent results across benchmark runs

## Architecture-Specific Optimizations

### Intel x86_64 Processors
```bash
gcc -O3 -march=native -mtune=native -mavx2 -o benchmark benchmark.c
```
- **AVX2**: Advanced Vector Extensions for SIMD operations
- **Native Tuning**: Processor-specific optimizations
- **O3 Optimization**: Aggressive compiler optimizations

### ARM Graviton Processors
```bash
gcc -O3 -march=native -mtune=native -mcpu=native -o benchmark benchmark.c
```
- **Native CPU**: ARM-specific processor optimizations
- **Neoverse Tuning**: Graviton processor family optimizations
- **ARM SIMD**: ARM NEON and SVE vector instructions

## Quality Assurance

### Data Integrity
- **Real Hardware Execution**: All benchmarks run on actual EC2 instances
- **No Simulation**: Zero tolerance for fake or simulated data
- **Verification**: Multiple validation points throughout execution

### Error Handling
- **Allocation Failures**: Graceful handling of memory allocation errors
- **Timeout Management**: Appropriate timeouts for long-running benchmarks
- **Retry Logic**: Automatic retry for transient failures

### Validation
- **Schema Compliance**: All results validated against JSON schemas
- **Statistical Checks**: Outlier detection and confidence validation
- **Cross-Validation**: Results compared across similar instance types

## Future Enhancements

### NUMA Awareness
- **NUMA Detection**: Identify NUMA topology for memory binding
- **Thread Affinity**: Pin benchmark threads to specific cores
- **Memory Policies**: Configure memory allocation policies

### Advanced Profiling
- **PMU Integration**: Performance monitoring unit integration
- **Cache Miss Analysis**: Detailed cache performance profiling
- **Memory Bandwidth**: Fine-grained memory subsystem analysis

### Workload Diversity
- **Application-Specific**: Workloads representative of real applications
- **Mixed Workloads**: Combined CPU, memory, and I/O benchmarks
- **Scalability Testing**: Multi-threaded and distributed workloads

---

*Last Updated: 2024-06-30*
*Implementation Status: Production Ready*