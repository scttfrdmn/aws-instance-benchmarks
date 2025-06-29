# Batch Scheduling System

## Overview

The AWS Instance Benchmarks project includes a sophisticated batch scheduling system that enables systematic benchmark execution across large numbers of AWS EC2 instances over time. This system is designed to avoid quota limits, optimize costs, and gather comprehensive microarchitectural performance data.

## Key Features

### 1. Time-Based Distribution
- **Weekly Planning**: Distributes benchmarks across a 7-day period
- **Daily Windows**: Multiple execution windows per day (morning, afternoon, evening)
- **Quota Management**: Respects AWS service quotas by spreading load over time
- **Cost Optimization**: Uses spot instances and off-peak execution when beneficial

### 2. Microarchitecture-Aware Benchmarking
- **Architecture Detection**: Automatically detects Intel, AMD, and Graviton architectures
- **Specialized Benchmarks**: Expands base benchmarks with architecture-specific variants
- **Performance Focus**: Gathers data critical for domain-specific workload optimization

### 3. Instance Size Wave Grouping
- **Physical Node Avoidance**: Groups instances by size to minimize same-hardware conflicts
- **Wave Distribution**: `large` → `xlarge` → `2xlarge` → `4xlarge/8xlarge`
- **Benchmark Isolation**: Reduces interference between concurrent benchmark executions

## Architecture

### Core Components

```
pkg/scheduler/
├── batch_scheduler.go      # Main scheduling logic
├── job_queue.go           # Job prioritization and queue management
├── time_windows.go        # Execution window definitions
└── progress_tracker.go    # Execution monitoring and reporting
```

### Data Structures

#### BatchScheduler
```go
type BatchScheduler struct {
    config          Config
    jobQueue        *JobQueue
    progressTracker *ProgressTracker
    timeWindows     []TimeWindow
    benchmarkRunner BenchmarkRunner
}
```

#### BenchmarkJob
```go
type BenchmarkJob struct {
    ID               string
    InstanceType     string
    BenchmarkSuite   string
    Region           string
    Priority         int
    EstimatedDuration time.Duration
    EstimatedCost    float64
    PreferSpotInstance bool
    Tags             map[string]string
}
```

## Benchmark Types

### Base Benchmarks
- **STREAM**: Memory bandwidth testing
- **HPL**: CPU floating-point performance (LINPACK)

### Microarchitecture Extensions

#### Memory Subsystem
- `stream-numa`: NUMA-aware memory access patterns
- `stream-cache`: Cache hierarchy analysis (L1/L2/L3)
- `stream-prefetch`: Hardware prefetcher evaluation

#### Architecture-Specific Memory
- `stream-avx512`: Intel AVX-512 memory bandwidth
- `stream-avx2`: AMD-optimized AVX2 memory access
- `stream-neon`: ARM Neon SIMD memory operations

#### CPU Microarchitecture
- `hpl-single`: Single-threaded performance analysis
- `hpl-vector`: Vectorization efficiency testing
- `hpl-branch`: Branch prediction analysis

#### Optimized Libraries
- `hpl-mkl`: Intel Math Kernel Library optimizations
- `hpl-blis`: AMD Basic Linear Algebra Subprograms
- `hpl-sve`: ARM Scalable Vector Extensions
- `hpl-neoverse`: ARM Neoverse core optimizations

#### Microbenchmarks
- `micro-latency`: Memory latency analysis
- `micro-ipc`: Instructions per cycle measurement
- `micro-tlb`: Translation Lookaside Buffer performance
- `micro-cache`: Cache miss pattern analysis

## Usage

### Weekly Scheduling Command

Execute a comprehensive weekly benchmark plan:

```bash
./aws-benchmark-collector schedule weekly \
    --instance-families m7i,c7g,r7a \
    --region us-east-1 \
    --max-daily-jobs 30 \
    --max-concurrent 5 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --benchmark-rotation \
    --instance-size-waves \
    --enable-spot
```

### Plan Generation Command

Generate a plan without executing:

```bash
./aws-benchmark-collector schedule plan \
    --instance-types m7i.large,c7g.large,r7a.large \
    --benchmarks stream,hpl \
    --output weekly-plan.json
```

### Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `--instance-families` | Instance families to benchmark | `m7i,c7g,r7a` |
| `--max-daily-jobs` | Maximum jobs per day | `30` |
| `--max-concurrent` | Maximum concurrent executions | `5` |
| `--benchmark-rotation` | Rotate benchmark types across windows | `true` |
| `--instance-size-waves` | Group by size to avoid physical conflicts | `true` |
| `--enable-spot` | Use spot instances for cost optimization | `true` |

## Time Window Strategy

### Daily Schedule

#### Morning Window (8 AM - 12 PM)
- **Focus**: Memory and cache benchmarks
- **Benchmarks**: `stream`, `stream-cache`, `stream-numa`, `micro-cache`
- **Priority**: Cost-optimized execution
- **Duration**: 4 hours

#### Afternoon Window (2 PM - 8 PM)
- **Focus**: Compute and vectorization benchmarks
- **Benchmarks**: `hpl`, `hpl-vector`, `stream-avx512`, `stream-neon`
- **Priority**: Balanced execution
- **Duration**: 6 hours

#### Evening Window (8 PM - 12 AM)
- **Focus**: Architecture-specific and optimized library tests
- **Benchmarks**: `hpl-mkl`, `hpl-blis`, `micro-latency`, `micro-ipc`
- **Priority**: Spot instance optimized
- **Duration**: 4 hours

## Instance Size Wave Grouping

When `--instance-size-waves` is enabled, instances are grouped to minimize physical node conflicts:

### Wave 1: Small Instances
- Instance types: `*.large`
- Execution priority: High
- Physical isolation: Maximum

### Wave 2: Medium Instances
- Instance types: `*.xlarge`
- Execution priority: Medium-High
- Physical isolation: Good

### Wave 3: Large Instances
- Instance types: `*.2xlarge`
- Execution priority: Medium
- Physical isolation: Moderate

### Wave 4: Very Large Instances
- Instance types: `*.4xlarge`, `*.8xlarge`
- Execution priority: Lower
- Physical isolation: Minimal concern (dedicated hardware)

## Architecture-Aware Scheduling

### Intel x86_64 Focus
- **Vectorization**: AVX-512, AVX2 testing
- **Libraries**: Intel MKL optimization
- **Features**: Intel-specific microarchitecture analysis

### AMD x86_64 Focus
- **Vectorization**: AVX2 optimization
- **Libraries**: AMD BLIS optimization
- **Features**: Zen4 architecture analysis

### AWS Graviton ARM64 Focus
- **Vectorization**: Neon SIMD, SVE testing
- **Libraries**: ARM-optimized implementations
- **Features**: Neoverse core analysis

## Cost Optimization

### Spot Instance Strategy
- **Large Instances**: Automatically use spot instances for `xlarge+`
- **Cost Savings**: Typically 50-70% cost reduction
- **Availability**: Monitor spot pricing and availability
- **Fallback**: Automatic fallback to on-demand if spot unavailable

### Off-Peak Execution
- **Evening Windows**: Prefer spot instances during off-peak hours
- **Regional Pricing**: Consider regional pricing differences
- **Long-Running Jobs**: Schedule expensive jobs during cost-optimal windows

## Progress Tracking

### Real-Time Monitoring
- **Job Status**: Track pending, running, completed, failed jobs
- **Progress Metrics**: Completion percentage and estimated time remaining
- **Error Handling**: Automatic retry logic with exponential backoff

### Reporting
- **Execution Summary**: Total jobs, success rate, execution time
- **Cost Analysis**: Actual vs estimated costs
- **Performance Insights**: Statistical analysis of benchmark results

## Integration with ComputeCompass

The batch scheduling system generates comprehensive microarchitectural data that enables ComputeCompass to make intelligent, domain-specific instance recommendations:

### Memory-Intensive Workloads
- **NUMA Topology**: Optimal memory access patterns
- **Cache Hierarchy**: L1/L2/L3 performance characteristics
- **Memory Bandwidth**: Peak and sustained throughput

### Compute-Intensive Workloads
- **Vectorization**: AVX-512 vs AVX2 vs Neon capabilities
- **IPC Analysis**: Instructions per cycle efficiency
- **Library Performance**: MKL vs BLIS vs generic implementations

### Domain-Specific Optimization
- **HPC**: Focus on vectorization and memory bandwidth
- **ML/AI**: Emphasize matrix operations and memory hierarchy
- **Scientific Computing**: Balance compute and memory subsystem performance

## Example Execution Plan

A typical weekly execution plan for 3 instance families across 4 sizes would generate:

```
Instance Types: 12 (3 families × 4 sizes)
Base Benchmarks: 2 (stream, hpl)
Extended Benchmarks: ~20 (microarchitecture variants)
Total Jobs: ~240 (12 × 20)
Execution Windows: 21 (7 days × 3 windows)
Jobs per Window: ~11-12
Estimated Duration: 7 days
Estimated Cost: $150-300 (with spot optimization)
```

## Best Practices

### 1. Regional Selection
- Choose regions with good spot availability
- Consider quota limits in target regions
- Account for regional pricing differences

### 2. Time Planning
- Allow buffer time for failed/retried jobs
- Consider AWS maintenance windows
- Plan around known high-demand periods

### 3. Cost Management
- Monitor actual vs estimated costs
- Use spot instances for non-critical large instances
- Implement cost alerts and limits

### 4. Performance Analysis
- Collect multiple iterations for statistical validity
- Monitor coefficient of variation for result quality
- Compare results across architectural variants

## Troubleshooting

### Common Issues

#### Quota Exceeded Errors
```bash
# Reduce concurrent executions
--max-concurrent 3

# Spread jobs over more days
--max-daily-jobs 15
```

#### Spot Instance Unavailability
```bash
# Disable spot instances
--enable-spot=false

# Use smaller instance types
--instance-families m6i,c6g,r6a
```

#### Long Execution Times
```bash
# Reduce benchmark scope
--benchmarks stream

# Focus on specific architectures
--instance-families m7i  # Intel only
```

## Future Enhancements

### Planned Features
- **Multi-Region Execution**: Parallel execution across multiple AWS regions
- **Custom Benchmark Integration**: Support for user-defined benchmark suites
- **Advanced Cost Optimization**: ML-based spot pricing prediction
- **Real-Time Adaptation**: Dynamic scheduling based on current AWS conditions

### Integration Opportunities
- **CI/CD Integration**: Automated benchmark execution on infrastructure changes
- **Performance Regression Detection**: Continuous monitoring of instance performance
- **Capacity Planning**: Historical performance trends for resource planning