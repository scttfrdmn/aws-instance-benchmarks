# CloudWatch Integration and Monitoring

This document describes the comprehensive CloudWatch metrics integration for monitoring AWS Instance Benchmarks execution and performance.

## Overview

The CloudWatch integration provides:
- **Real-time metrics** for benchmark execution success/failure rates
- **Performance tracking** with statistical analysis
- **Operational monitoring** for system health and efficiency
- **Cost tracking** and price-performance analysis
- **Quality assessment** with confidence intervals and stability metrics

## Metrics Categories

### 1. Benchmark Execution Metrics

#### Core Metrics
- **`BenchmarkExecution`** (Count): Total benchmark runs with success/failure dimensions
- **`ExecutionDuration`** (Seconds): Total time including infrastructure provisioning
- **`BenchmarkDuration`** (Seconds): Pure benchmark execution time
- **`QualityScore`** (Ratio): Statistical quality assessment (0.0-1.0)

#### Performance Metrics
- **`Performance_triad_bandwidth`** (Bytes/Second): STREAM triad bandwidth
- **`Performance_copy_bandwidth`** (Bytes/Second): STREAM copy bandwidth  
- **`Performance_scale_bandwidth`** (Bytes/Second): STREAM scale bandwidth
- **`Performance_add_bandwidth`** (Bytes/Second): STREAM add bandwidth
- **`Performance_gflops`** (Count/Second): HPL computational performance
- **`Performance_efficiency`** (None): HPL efficiency ratio
- **`Performance_execution_time`** (Seconds): HPL execution time
- **`Performance_residual`** (None): HPL numerical accuracy

#### Cost Metrics
- **`EstimatedCost`** (None): USD cost for benchmark execution
- **`PricePerformanceRatio`** (None): Cost efficiency ($/GFLOP or $/GB/s)

### 2. Operational Metrics

#### System Health
- **`ActiveInstances`** (Count): Currently running benchmark instances
- **`FailureRate`** (Percent): Percentage of failed benchmark executions
- **`InstanceLaunchDuration`** (Seconds): Time to launch and initialize instances
- **`ContainerPullDuration`** (Seconds): Container download time

#### Capacity Management
- **`QuotaUtilization`** (Percent): Per-instance-family quota usage

## Metric Dimensions

### Standard Dimensions (All Metrics)
- **`Project`**: "aws-instance-benchmarks"
- **`Environment`**: "production"
- **`Region`**: AWS region where benchmark executed

### Benchmark-Specific Dimensions
- **`InstanceType`**: Specific instance type (e.g., "m7i.large")
- **`InstanceFamily`**: Instance family (e.g., "m7i")
- **`BenchmarkSuite`**: Benchmark type ("stream" or "hpl")
- **`Success`**: Execution success status ("true" or "false")

### Error Categorization
- **`ErrorCategory`**: For failed executions
  - `"quota"`: AWS quota or capacity limitations
  - `"infrastructure"`: Instance launch or networking issues
  - `"benchmark"`: Benchmark execution failures
  - `"timeout"`: Execution timeout
  - `"validation"`: Result validation failures

## CloudWatch Dashboards

### Executive Dashboard
Key performance indicators:
- Overall success rate trends
- Cost per benchmark execution
- Regional performance comparison
- Instance family efficiency ranking

### Operational Dashboard
System health monitoring:
- Active instance count
- Failure rate by error category
- Infrastructure provisioning time
- Quota utilization by family

### Performance Dashboard
Benchmark result analysis:
- Performance trends by instance type
- Statistical quality scores
- Coefficient of variation tracking
- Confidence interval bounds

## Alerting and Automation

### Critical Alerts

#### High Failure Rate
```yaml
MetricName: FailureRate
Threshold: 20%
Period: 15 minutes
Statistic: Average
ComparisonOperator: GreaterThanThreshold
```

#### Quota Utilization
```yaml
MetricName: QuotaUtilization  
Threshold: 80%
Period: 5 minutes
Statistic: Maximum
ComparisonOperator: GreaterThanThreshold
Dimensions:
  - Name: InstanceFamily
    Value: "*"
```

#### Cost Anomaly Detection
```yaml
MetricName: EstimatedCost
AnomalyDetector: ML-based cost anomaly detection
Threshold: 2 standard deviations
```

### Warning Alerts

#### Performance Degradation
```yaml
MetricName: QualityScore
Threshold: 0.7
Period: 30 minutes
Statistic: Average
ComparisonOperator: LessThanThreshold
```

#### Infrastructure Delays
```yaml
MetricName: InstanceLaunchDuration
Threshold: 300 seconds
Period: 10 minutes
Statistic: Average
ComparisonOperator: GreaterThanThreshold
```

## Implementation Details

### Metrics Collection Architecture

```go
// Initialize CloudWatch metrics collector
metricsCollector, err := monitoring.NewMetricsCollector(region)
if err != nil {
    log.Printf("CloudWatch metrics unavailable: %v", err)
    metricsCollector = nil // Continue without metrics
}

// Publish benchmark metrics
metrics := monitoring.BenchmarkMetrics{
    InstanceType:       instanceType,
    InstanceFamily:     extractInstanceFamily(instanceType),
    BenchmarkSuite:     benchmarkSuite,
    Region:            region,
    Success:           err == nil,
    ExecutionDuration: totalTime.Seconds(),
    BenchmarkDuration: benchmarkTime.Seconds(),
    PerformanceMetrics: extractedMetrics,
    QualityScore:      calculateQuality(results),
    Timestamp:         time.Now(),
}

err = metricsCollector.PublishBenchmarkMetrics(ctx, metrics)
```

### Error Categorization Logic

```go
// Categorize errors for better monitoring
var errorCategory string
if quotaErr, ok := err.(*aws.QuotaError); ok {
    errorCategory = "quota"
    fmt.Printf("⚠️ Quota constraint: %s\n", quotaErr.Message)
} else if strings.Contains(err.Error(), "timeout") {
    errorCategory = "timeout"
} else if strings.Contains(err.Error(), "capacity") {
    errorCategory = "infrastructure"
} else {
    errorCategory = "benchmark"
}

benchmarkMetrics.ErrorCategory = errorCategory
```

### Statistical Metrics Integration

```go
// Calculate and publish statistical quality metrics
for metricName, values := range collectedMetrics {
    mean, stdDev, cv := calculateStatistics(values)
    confInterval := calculateConfidenceInterval(values, 0.95)
    
    // Publish statistical aggregations
    publisher.PublishStatistic("Mean_" + metricName, mean)
    publisher.PublishStatistic("StdDev_" + metricName, stdDev)
    publisher.PublishStatistic("CV_" + metricName, cv)
    publisher.PublishStatistic("CI_Lower_" + metricName, confInterval.Lower)
    publisher.PublishStatistic("CI_Upper_" + metricName, confInterval.Upper)
}
```

## Query Examples

### CloudWatch Insights Queries

#### Success Rate Analysis
```sql
fields @timestamp, InstanceType, BenchmarkSuite, Success
| filter MetricName = "BenchmarkExecution"
| stats count() as TotalRuns, 
        sum(case when Success = "true" then 1 else 0 end) as SuccessfulRuns
        by InstanceType, BenchmarkSuite
| eval SuccessRate = SuccessfulRuns / TotalRuns * 100
| sort SuccessRate desc
```

#### Performance Trends
```sql
fields @timestamp, InstanceType, Value
| filter MetricName = "Performance_triad_bandwidth"
| stats avg(Value) as AvgBandwidth, 
        stddev(Value) as StdDevBandwidth,
        count() as Samples
        by InstanceType
| eval CV = StdDevBandwidth / AvgBandwidth * 100
| sort AvgBandwidth desc
```

#### Cost Analysis
```sql
fields @timestamp, InstanceType, Region, Value as Cost
| filter MetricName = "EstimatedCost"
| stats sum(Cost) as TotalCost,
        avg(Cost) as AvgCostPerRun,
        count() as Executions
        by InstanceType, Region
| eval CostPerExecution = TotalCost / Executions
| sort TotalCost desc
```

### CloudWatch API Queries

#### Recent Performance Data
```bash
aws cloudwatch get-metric-statistics \
  --namespace InstanceBenchmarks \
  --metric-name Performance_triad_bandwidth \
  --dimensions Name=InstanceType,Value=m7i.large \
  --start-time 2024-06-29T00:00:00Z \
  --end-time 2024-06-29T23:59:59Z \
  --period 3600 \
  --statistics Average,Maximum,Minimum
```

#### Failure Rate Monitoring
```bash
aws cloudwatch get-metric-statistics \
  --namespace InstanceBenchmarks \
  --metric-name FailureRate \
  --dimensions Name=Region,Value=us-east-1 \
  --start-time $(date -d '1 hour ago' -u +%Y-%m-%dT%H:%M:%SZ) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --period 300 \
  --statistics Average
```

## Data Retention and Cost Management

### Metric Retention Policies
- **High-frequency metrics** (execution counts): 15 months
- **Performance metrics**: 15 months with detailed resolution
- **Operational metrics**: 15 months
- **Statistical aggregations**: 15 months

### Cost Optimization
- **Metric publishing batching**: Up to 1000 metrics per API call
- **Intelligent aggregation**: Pre-aggregate statistical measures
- **Selective publishing**: Only publish when metrics exist
- **Error filtering**: Avoid publishing redundant error metrics

## Troubleshooting

### Common Issues

#### Missing Metrics
```bash
# Check AWS credentials and permissions
aws sts get-caller-identity

# Verify CloudWatch permissions
aws cloudwatch list-metrics --namespace InstanceBenchmarks

# Test metric publishing
aws cloudwatch put-metric-data \
  --namespace TestNamespace \
  --metric-data MetricName=TestMetric,Value=1.0
```

#### High CloudWatch Costs
```bash
# Check metric count
aws cloudwatch list-metrics --namespace InstanceBenchmarks | jq '.Metrics | length'

# Analyze metric dimensions
aws cloudwatch list-metrics --namespace InstanceBenchmarks | \
  jq -r '.Metrics[].Dimensions | length' | sort -n | uniq -c
```

### Debug Commands

```bash
# Enable verbose CloudWatch logging
export AWS_SDK_LOAD_CONFIG=1
export AWS_LOG_LEVEL=debug

# Test metrics collection locally
./aws-benchmark-collector run \
  --instance-types m7i.large \
  --benchmarks stream \
  --iterations 1

# Check CloudWatch namespace
aws cloudwatch list-metrics --namespace InstanceBenchmarks --max-records 50
```

## Security Considerations

### IAM Permissions

Required CloudWatch permissions:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:PutMetricData",
                "cloudwatch:GetMetricStatistics",
                "cloudwatch:ListMetrics"
            ],
            "Resource": "*"
        }
    ]
}
```

### Data Privacy
- **No sensitive data**: Only performance metrics and metadata
- **Instance IDs**: Temporary and automatically cleaned up
- **Regional isolation**: Metrics stored in benchmark execution region
- **Access control**: IAM-based access to CloudWatch metrics

## Integration with External Systems

### ComputeCompass Integration
CloudWatch metrics can be consumed by ComputeCompass for:
- Real-time performance monitoring
- Historical trend analysis
- Automated performance alerts
- Cost optimization recommendations

### Third-Party Monitoring
Standard CloudWatch APIs enable integration with:
- Grafana dashboards
- Datadog monitoring
- New Relic APM
- Custom monitoring solutions

## Future Enhancements

### Planned Features
- **Custom metric streams** for real-time analysis
- **Machine learning insights** for performance prediction
- **Automated anomaly detection** for quality regression
- **Cross-region aggregation** for global performance views

### Advanced Analytics
- **Trend prediction** using historical data
- **Performance regression detection** with statistical significance
- **Cost optimization insights** based on price-performance ratios
- **Capacity planning** recommendations based on usage patterns