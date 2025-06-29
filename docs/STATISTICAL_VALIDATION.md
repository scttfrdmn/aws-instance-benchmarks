# Statistical Validation and Multi-Run Analysis

This document describes the statistical validation capabilities implemented in AWS Instance Benchmarks for ensuring data quality and reproducibility.

## Overview

The statistical validation system provides:
- **Multiple iteration execution** for robust statistical analysis
- **Confidence interval calculations** using proper t-distribution
- **Coefficient of variation** analysis for performance consistency
- **Quality scoring** based on statistical stability
- **Outlier detection** and data quality assessment

## Usage

### Command Line Interface

```bash
# Run benchmarks with statistical validation
./aws-benchmark-collector run \
  --instance-types m7i.large,c7g.large \
  --benchmarks stream,hpl \
  --iterations 5 \
  --region us-east-1 \
  --key-pair my-key-pair \
  --security-group sg-12345678 \
  --subnet subnet-12345678

# Multiple iterations with parallel execution
./aws-benchmark-collector run \
  --instance-types m7i.large \
  --benchmarks stream \
  --iterations 10 \
  --max-concurrency 3
```

### Statistical Analysis Output

When `--iterations > 1`, the system automatically provides detailed statistical analysis:

```
📈 Statistical Analysis:

   stream on m7i.large (3 successful runs):
     Triad Bandwidth: 41.90 ± 0.00 GB/s (CV: 0.0%, 95% CI: 41.90-41.90)
     Copy Bandwidth:  45.20 ± 0.00 GB/s (CV: 0.0%, 95% CI: 45.20-45.20)
     Scale Bandwidth: 44.80 ± 0.00 GB/s (CV: 0.0%, 95% CI: 44.80-44.80)
     Add Bandwidth:   42.10 ± 0.00 GB/s (CV: 0.0%, 95% CI: 42.10-42.10)

   hpl on m7i.large (3 successful runs):
     GFLOPS:          156.50 ± 2.30 (CV: 1.5%, 95% CI: 154.20-158.80)
     Efficiency:      0.850 ± 0.020 (CV: 2.4%, 95% CI: 0.830-0.870)
     Execution Time:  120.50 ± 5.20 s (CV: 4.3%, 95% CI: 115.30-125.70)
```

## Statistical Metrics

### Key Statistical Measures

1. **Mean (μ)**: Average value across all successful runs
2. **Standard Deviation (σ)**: Measure of variability
3. **Coefficient of Variation (CV)**: Relative variability (σ/μ × 100%)
4. **95% Confidence Interval**: Range of plausible values for the true mean

### Confidence Interval Calculation

The system uses proper t-distribution for small samples:

```
CI = μ ± t(α/2, n-1) × (σ/√n)
```

Where:
- `t(α/2, n-1)` is the t-value for 95% confidence
- Sample size dependent: 1.96 (n≥30), 2.26 (n≥10), 3.18 (n<10)

### Quality Score Calculation

#### STREAM Benchmarks
Quality score based on bandwidth consistency:
- **CV ≤ 5%**: Excellent (score ≥ 0.9)
- **CV ≤ 10%**: Good (score ≥ 0.7)
- **CV > 10%**: Poor (score < 0.7)

#### HPL Benchmarks
Quality score considers multiple factors:
- **Efficiency ≥ 0.7**: No penalty
- **Efficiency 0.5-0.7**: -0.2 penalty
- **Efficiency < 0.5**: -0.4 penalty
- **Residual ≤ 1e-9**: No penalty
- **Residual > 1e-6**: -0.3 penalty

## Data Collection and Aggregation

### Result Storage Structure

Each iteration creates an individual `benchmarkResult`:

```go
type benchmarkResult struct {
    instanceType   string
    benchmarkSuite string
    iteration      int
    success        bool
    result         *aws.InstanceResult
    metrics        monitoring.BenchmarkMetrics
}
```

### Statistical Aggregation Process

1. **Data Collection**: All iterations stored in `allResults` slice
2. **Grouping**: Results grouped by `instanceType-benchmarkSuite` key
3. **Extraction**: Performance values extracted from nested data structures
4. **Analysis**: Statistical calculations performed on valid results
5. **Reporting**: Formatted output with quality assessment

## Implementation Details

### STREAM Data Extraction

The system handles nested STREAM data structures:

```json
{
  "performance_data": {
    "stream": {
      "triad": {"bandwidth": 41.9, "unit": "GB/s"},
      "copy": {"bandwidth": 45.2, "unit": "GB/s"},
      "scale": {"bandwidth": 44.8, "unit": "GB/s"},
      "add": {"bandwidth": 42.1, "unit": "GB/s"}
    }
  }
}
```

### HPL Data Extraction

HPL performance metrics extraction:

```json
{
  "performance_data": {
    "hpl": {
      "gflops": 156.7,
      "efficiency": 0.85,
      "execution_time": 120.5,
      "residual": 1e-12
    }
  }
}
```

### Error Handling and Recovery

The system handles various failure scenarios:

1. **Instance Launch Failures**: Categorized as "infrastructure" or "quota" errors
2. **Benchmark Execution Failures**: Marked as failed iterations
3. **Capacity Constraints**: Graceful degradation with partial results
4. **Data Parsing Errors**: Robust extraction with fallback mechanisms

## Best Practices

### Recommended Iteration Counts

- **Development/Testing**: 3-5 iterations
- **Production Data Collection**: 5-10 iterations
- **Research/Publication**: 10+ iterations
- **High-Precision Measurements**: 20+ iterations

### Quality Thresholds

- **Acceptable Quality**: CV ≤ 10% for STREAM, Efficiency ≥ 0.7 for HPL
- **Good Quality**: CV ≤ 5% for STREAM, Efficiency ≥ 0.8 for HPL
- **Excellent Quality**: CV ≤ 2% for STREAM, Efficiency ≥ 0.9 for HPL

### Statistical Significance

- **Minimum Sample Size**: 3 iterations for basic analysis
- **Recommended Sample Size**: 5+ iterations for reliable statistics
- **Confidence Level**: 95% (α = 0.05) for all calculations

## CloudWatch Integration

Statistical metrics are automatically published to CloudWatch:

- **Individual Run Metrics**: Each iteration publishes separate metrics
- **Quality Scores**: Calculated and published for trend analysis
- **Coefficient of Variation**: Tracked for performance stability monitoring
- **Confidence Intervals**: Lower and upper bounds tracked separately

## Automation and CI/CD

The statistical validation integrates with GitHub Actions:

```yaml
- name: Execute benchmarks with statistical validation
  run: |
    ./aws-benchmark-collector run \
      --instance-types $INSTANCE_TYPES \
      --benchmarks stream,hpl \
      --iterations 5 \
      --max-concurrency 3
```

## Troubleshooting

### Common Issues

1. **Insufficient Valid Runs**: Increase `--iterations` or check infrastructure
2. **High Coefficient of Variation**: Check for system load or configuration issues
3. **Statistical Analysis Missing**: Ensure `--iterations > 1` is specified
4. **Data Extraction Errors**: Verify benchmark data format consistency

### Debug Commands

```bash
# Test with single iteration first
./aws-benchmark-collector run --instance-types m7i.large --iterations 1

# Check data format
cat results/latest/*.json | jq '.performance_data'

# Validate statistical calculations
./scripts/validate_contribution.sh
```

## References

- [Student's t-distribution](https://en.wikipedia.org/wiki/Student%27s_t-distribution)
- [Coefficient of Variation](https://en.wikipedia.org/wiki/Coefficient_of_variation)
- [Confidence Intervals](https://en.wikipedia.org/wiki/Confidence_interval)
- [STREAM Benchmark Methodology](https://www.cs.virginia.edu/stream/)
- [HPL Benchmark Specification](https://www.netlib.org/benchmark/hpl/)