# Microarchitecture Benchmarks

## Overview

The AWS Instance Benchmarks project includes comprehensive microarchitecture-specific benchmarks designed to provide deep insights into CPU and memory subsystem performance characteristics. These benchmarks enable domain-specific workload optimization and intelligent instance selection for specialized computing tasks.

## Benchmark Categories

### 1. Memory Subsystem Benchmarks

#### STREAM Variants

##### `stream` (Base)
- **Purpose**: Standard memory bandwidth measurement
- **Operations**: Copy, Scale, Add, Triad
- **Output**: Peak memory bandwidth (GB/s)
- **Architectures**: All (Intel, AMD, Graviton)

##### `stream-numa`
- **Purpose**: NUMA topology and cross-socket memory access analysis
- **Operations**: Local vs remote memory bandwidth measurement
- **Output**: NUMA efficiency ratios, bandwidth by memory region
- **Critical for**: Large multi-socket instances (>8xlarge)

##### `stream-cache`
- **Purpose**: Cache hierarchy performance analysis
- **Operations**: Working set size variation (L1/L2/L3 cache sizes)
- **Output**: Cache bandwidth, latency by level, cache efficiency
- **Critical for**: Memory-sensitive workloads, cache-optimized algorithms

##### `stream-prefetch`
- **Purpose**: Hardware prefetcher effectiveness evaluation
- **Operations**: Sequential vs random access patterns
- **Output**: Prefetcher hit rates, bandwidth with/without prefetching
- **Critical for**: Scientific computing, data streaming applications

#### Architecture-Specific Memory Tests

##### `stream-avx512` (Intel)
- **Purpose**: AVX-512 memory bandwidth utilization
- **Operations**: 512-bit vector memory operations
- **Output**: Peak vectorized memory throughput
- **Target Instances**: Intel 3rd gen (Ice Lake): m6i, c6i, r6i and newer

##### `stream-avx2` (AMD)
- **Purpose**: AMD-optimized AVX2 memory access
- **Operations**: 256-bit vector operations optimized for Zen architecture
- **Output**: Zen-optimized memory bandwidth
- **Target Instances**: AMD instances: m6a, c6a, r6a, m7a, c7a, r7a

##### `stream-neon` (Graviton)
- **Purpose**: ARM Neon SIMD memory operations
- **Operations**: 128-bit ARM vector instructions
- **Output**: ARM SIMD memory throughput
- **Target Instances**: Graviton instances: m6g, c6g, r6g, m7g, c7g, r7g

### 2. CPU Microarchitecture Benchmarks

#### HPL Variants

##### `hpl` (Base)
- **Purpose**: Standard CPU floating-point performance (LINPACK)
- **Operations**: Dense matrix solve, peak GFLOPS measurement
- **Output**: Theoretical peak performance percentage
- **Architectures**: All

##### `hpl-single`
- **Purpose**: Single-threaded performance analysis
- **Operations**: Single-core LINPACK solve
- **Output**: Per-core performance, single-thread efficiency
- **Critical for**: Serial workloads, single-threaded bottlenecks

##### `hpl-vector`
- **Purpose**: Vectorization efficiency measurement
- **Operations**: Vector vs scalar performance comparison
- **Output**: Vectorization speedup ratios
- **Critical for**: Compute-intensive scientific applications

##### `hpl-branch`
- **Purpose**: Branch prediction analysis
- **Operations**: Conditional computation patterns
- **Output**: Branch misprediction rates, conditional performance
- **Critical for**: Control-flow intensive applications

#### Architecture-Specific Compute Tests

##### `hpl-mkl` (Intel)
- **Purpose**: Intel Math Kernel Library optimization analysis
- **Operations**: MKL-optimized BLAS routines
- **Output**: MKL vs generic performance ratios
- **Target Instances**: Intel instances with MKL support

##### `hpl-avx512-fma` (Intel)
- **Purpose**: AVX-512 Fused Multiply-Add performance
- **Operations**: FMA-intensive computations
- **Output**: Peak FMA throughput, instruction-level parallelism
- **Target Instances**: Intel Ice Lake and newer (3rd+ gen)

##### `hpl-blis` (AMD)
- **Purpose**: AMD Basic Linear Algebra Subprograms optimization
- **Operations**: BLIS-optimized matrix operations
- **Output**: BLIS vs generic performance comparisons
- **Target Instances**: AMD Zen3/Zen4 instances

##### `hpl-zen4` (AMD)
- **Purpose**: AMD Zen4 architecture-specific features
- **Operations**: Zen4-optimized instruction sequences
- **Output**: Zen4 architectural efficiency metrics
- **Target Instances**: m7a, c7a, r7a (4th gen AMD)

##### `hpl-sve` (Graviton)
- **Purpose**: ARM Scalable Vector Extensions performance
- **Operations**: SVE vector computations
- **Output**: SVE vectorization efficiency
- **Target Instances**: Graviton3+ with SVE support

##### `hpl-neoverse` (Graviton)
- **Purpose**: ARM Neoverse core optimization analysis
- **Operations**: Neoverse-specific instruction scheduling
- **Output**: Neoverse architectural efficiency
- **Target Instances**: All Graviton instances

### 3. Microbenchmarks

#### `micro-latency`
- **Purpose**: Memory latency characterization
- **Operations**: Pointer chasing, random access patterns
- **Output**: Latency by memory distance (L1/L2/L3/DRAM/NUMA)
- **Critical for**: Latency-sensitive applications, cache optimization

#### `micro-ipc`
- **Purpose**: Instructions per cycle measurement
- **Operations**: Various instruction mixes and dependency patterns
- **Output**: IPC by instruction type, dependency analysis
- **Critical for**: Compiler optimization, instruction scheduling

#### `micro-tlb`
- **Purpose**: Translation Lookaside Buffer performance
- **Operations**: Virtual memory access patterns, page size variations
- **Output**: TLB miss rates, page size efficiency
- **Critical for**: Memory-intensive applications, large working sets

#### `micro-cache`
- **Purpose**: Cache miss pattern analysis
- **Operations**: Access patterns designed to stress cache hierarchy
- **Output**: Miss rates by access pattern, cache conflict analysis
- **Critical for**: Cache-aware algorithm development

## Benchmark Selection by Instance Family

### General Purpose (M-series)

#### Intel (m6i, m7i)
```bash
Primary: stream, hpl, stream-cache, hpl-vector
Intel-specific: stream-avx512, hpl-mkl, hpl-avx512-fma
Micro: micro-latency, micro-ipc
```

#### AMD (m6a, m7a)
```bash
Primary: stream, hpl, stream-cache, hpl-vector
AMD-specific: stream-avx2, hpl-blis, hpl-zen4
Micro: micro-latency, micro-cache
```

#### Graviton (m6g, m7g)
```bash
Primary: stream, hpl, stream-cache, hpl-vector
ARM-specific: stream-neon, hpl-sve, hpl-neoverse
Micro: micro-latency, micro-tlb
```

### Compute Optimized (C-series)

#### Focus Areas
- **Vectorization**: `hpl-vector`, `stream-avx512`, `stream-neon`
- **Single-thread**: `hpl-single`, `micro-ipc`
- **Optimization**: `hpl-mkl`, `hpl-blis`, `hpl-sve`

### Memory Optimized (R-series)

#### Focus Areas
- **Memory bandwidth**: `stream`, `stream-numa`, `stream-prefetch`
- **Cache hierarchy**: `stream-cache`, `micro-cache`
- **NUMA**: `stream-numa` (especially for large instances)

## Container Architecture Mapping

### Intel Ice Lake (3rd Gen)
```dockerfile
# Optimized for m6i, c6i, r6i instances
FROM intel/oneapi-hpckit:latest
RUN spack install stream+avx512 %intel
RUN spack install hpl+mkl+avx512 %intel
```

### AMD Zen4 (4th Gen)
```dockerfile
# Optimized for m7a, c7a, r7a instances
FROM amd/aocc:latest
RUN spack install stream+avx2 %aocc
RUN spack install hpl+blis %aocc
```

### Graviton3 (ARM64)
```dockerfile
# Optimized for m7g, c7g, r7g instances
FROM arm64v8/ubuntu:22.04
RUN spack install stream+neon %gcc@11
RUN spack install hpl+sve %gcc@11
```

## Performance Data Structure

### Benchmark Result Schema

```json
{
  "schema_version": "1.1.0",
  "metadata": {
    "instance_type": "m7i.xlarge",
    "architecture": "x86_64",
    "microarchitecture": "intel-icelake",
    "benchmark_suite": "stream-avx512"
  },
  "performance": {
    "memory": {
      "stream_avx512": {
        "copy_bandwidth": 51.2,
        "scale_bandwidth": 50.8,
        "add_bandwidth": 48.3,
        "triad_bandwidth": 47.9,
        "vectorization_efficiency": 0.92,
        "instruction_throughput": 2.1
      }
    },
    "microarchitecture": {
      "vector_utilization": 0.89,
      "cache_efficiency": 0.94,
      "prefetcher_effectiveness": 0.87
    }
  }
}
```

## Integration with ComputeCompass

### Domain-Specific Recommendations

#### High-Performance Computing (HPC)
```json
{
  "workload_type": "hpc",
  "priority_metrics": [
    "hpl_vector_efficiency",
    "stream_triad_bandwidth", 
    "micro_ipc",
    "vectorization_speedup"
  ],
  "recommendations": {
    "intel": "Prefer instances with AVX-512 and high MKL efficiency",
    "amd": "Focus on Zen4 instances with BLIS optimization",
    "graviton": "Leverage SVE capabilities for vector workloads"
  }
}
```

#### Machine Learning/AI
```json
{
  "workload_type": "ml_ai",
  "priority_metrics": [
    "hpl_mkl_efficiency",
    "stream_cache_bandwidth",
    "micro_cache_efficiency",
    "memory_latency"
  ],
  "recommendations": {
    "matrix_operations": "Prioritize MKL/BLIS optimized instances",
    "memory_intensive": "Focus on cache hierarchy performance",
    "inference": "Balance single-thread and memory latency"
  }
}
```

#### Scientific Computing
```json
{
  "workload_type": "scientific",
  "priority_metrics": [
    "stream_numa_efficiency",
    "hpl_branch_performance", 
    "micro_tlb_efficiency",
    "prefetcher_effectiveness"
  ],
  "recommendations": {
    "large_datasets": "Prioritize NUMA-aware instances",
    "irregular_access": "Focus on TLB and prefetcher performance",
    "conditional_code": "Emphasize branch prediction efficiency"
  }
}
```

## Benchmark Execution Priority

### High Priority (Foundation Metrics)
1. `stream` - Base memory bandwidth
2. `hpl` - Base CPU performance
3. `stream-cache` - Cache hierarchy
4. `micro-latency` - Memory latency

### Medium Priority (Architecture-Specific)
1. `stream-avx512`/`stream-neon`/`stream-avx2` - Vector memory
2. `hpl-mkl`/`hpl-blis`/`hpl-sve` - Optimized compute
3. `stream-numa` - Multi-socket performance
4. `micro-ipc` - Instruction efficiency

### Lower Priority (Specialized Analysis)
1. `hpl-branch` - Control flow analysis
2. `micro-tlb` - Virtual memory performance
3. `stream-prefetch` - Prefetcher analysis
4. `micro-cache` - Cache conflict analysis

## Quality Assurance

### Statistical Validation
- **Multiple Runs**: 3-5 iterations per benchmark
- **Confidence Intervals**: 95% confidence level
- **Outlier Detection**: Coefficient of variation < 5%
- **Reproducibility**: Cross-validation across instances

### Performance Thresholds
- **Memory Bandwidth**: Within 10% of theoretical peak
- **CPU Performance**: >70% of theoretical GFLOPS
- **Cache Efficiency**: >85% for L1, >80% for L2
- **Vectorization**: >60% theoretical vector peak

## Future Enhancements

### Planned Benchmark Extensions
- **GPU Compute**: CUDA/ROCm/OpenCL microbenchmarks
- **Network Performance**: InfiniBand/SR-IOV analysis
- **Storage I/O**: NVMe, EBS optimization analysis
- **Power Efficiency**: Performance per watt metrics

### Advanced Analysis
- **Workload Modeling**: Real-world application performance prediction
- **Thermal Analysis**: Performance under sustained load
- **Compiler Optimization**: Impact of different optimization levels
- **Memory Pattern Analysis**: Application-specific access pattern modeling