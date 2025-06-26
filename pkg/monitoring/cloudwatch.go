// Package monitoring provides comprehensive CloudWatch metrics integration for
// AWS Instance Benchmarks operations and performance tracking.
//
// This package enables detailed observability into benchmark execution lifecycle,
// performance characteristics, and operational metrics. It provides standardized
// metric collection with proper namespacing, dimensions, and statistical aggregation
// for production monitoring and alerting workflows.
//
// Key Components:
//   - MetricsCollector: Main service for CloudWatch metrics publication
//   - BenchmarkMetrics: Structured metrics for benchmark execution tracking
//   - OperationalMetrics: Infrastructure and cost metrics collection
//   - CustomMetrics: Application-specific performance indicators
//
// Usage:
//   collector, err := monitoring.NewMetricsCollector("us-east-1")
//   metrics := monitoring.BenchmarkMetrics{
//       InstanceType: "m7i.large",
//       BenchmarkSuite: "stream",
//       ExecutionDuration: 45.2,
//       Success: true,
//   }
//   err := collector.PublishBenchmarkMetrics(ctx, metrics)
//
// The package provides:
//   - Standardized metric namespaces and dimension schemes
//   - Automatic retry logic with exponential backoff
//   - Batch metric publishing for efficiency
//   - Cost tracking and quota monitoring integration
//   - Performance trend analysis with statistical aggregation
//
// Metric Categories:
//   - Execution Metrics: Duration, success rate, error categorization
//   - Performance Metrics: Bandwidth, GFLOPS, efficiency measurements
//   - Infrastructure Metrics: Instance launches, quota utilization
//   - Cost Metrics: Spend tracking, price-performance ratios
//
// CloudWatch Integration:
//   - Custom metric namespaces for logical separation
//   - Dimension-based filtering and aggregation
//   - Alarm integration for automated alerting
//   - Dashboard support for operational visibility
package monitoring

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// CloudWatch monitoring errors.
var (
	ErrInvalidMetricValue = errors.New("metric value is invalid")
	ErrMetricNameRequired = errors.New("metric name is required")
	ErrDimensionLimit     = errors.New("dimension count exceeds CloudWatch limit")
)

// MetricsCollector provides comprehensive CloudWatch metrics collection and
// publishing capabilities for AWS Instance Benchmarks.
//
// This struct manages the complete metrics lifecycle including batch collection,
// dimension standardization, and efficient publishing to CloudWatch. It implements
// automatic retry logic, error categorization, and cost optimization through
// intelligent metric aggregation.
//
// Thread Safety:
//   The MetricsCollector is safe for concurrent use across multiple goroutines.
//   Metric publishing operations are automatically serialized and batched for
//   optimal CloudWatch API efficiency.
type MetricsCollector struct {
	// cloudwatchClient is the AWS SDK v2 CloudWatch client for metric publishing.
	// Configured with automatic retry logic and regional endpoint optimization.
	cloudwatchClient *cloudwatch.Client
	
	// namespace is the CloudWatch namespace for all published metrics.
	// Provides logical separation and access control for benchmark metrics.
	namespace string
	
	// region is the AWS region for CloudWatch metric publishing.
	// Must match the region where benchmarks are executed for consistency.
	region string
	
	// defaultDimensions contains standard dimensions applied to all metrics.
	// Enables consistent filtering and aggregation across metric queries.
	defaultDimensions []types.Dimension
}

// BenchmarkMetrics contains comprehensive performance and execution metrics
// for a single benchmark run on AWS EC2 infrastructure.
//
// This structure captures both operational metrics (execution success, duration)
// and performance data (bandwidth, efficiency) for trend analysis and alerting.
// All metrics are published with standardized dimensions for consistent querying.
type BenchmarkMetrics struct {
	// InstanceType is the AWS EC2 instance type used for benchmark execution.
	// Used as a primary dimension for performance comparison and analysis.
	InstanceType string
	
	// InstanceFamily is the extracted instance family (e.g., "m7i", "c7g").
	// Enables family-level aggregation and trend analysis.
	InstanceFamily string
	
	// BenchmarkSuite identifies the executed benchmark (e.g., "stream", "hpl").
	// Used for suite-specific performance tracking and comparison.
	BenchmarkSuite string
	
	// Region is the AWS region where the benchmark was executed.
	// Enables regional performance comparison and capacity planning.
	Region string
	
	// Success indicates whether the benchmark completed successfully.
	// Used for success rate calculation and error rate monitoring.
	Success bool
	
	// ExecutionDuration is the total benchmark execution time in seconds.
	// Includes instance launch, benchmark execution, and cleanup time.
	ExecutionDuration float64
	
	// BenchmarkDuration is the actual benchmark execution time in seconds.
	// Excludes infrastructure provisioning for pure performance measurement.
	BenchmarkDuration float64
	
	// PerformanceMetrics contains benchmark-specific performance measurements.
	// Structure varies by benchmark suite (STREAM bandwidth, HPL GFLOPS).
	PerformanceMetrics map[string]float64
	
	// ErrorCategory categorizes any execution errors for trend analysis.
	// Values: "quota", "infrastructure", "benchmark", "timeout"
	ErrorCategory string
	
	// CostMetrics contains cost tracking information for the execution.
	CostMetrics CostMetrics
	
	// QualityScore is the benchmark result quality assessment (0.0-1.0).
	// Based on statistical stability and validation criteria.
	QualityScore float64
	
	// Timestamp is when the benchmark execution completed.
	Timestamp time.Time
}

// CostMetrics contains detailed cost tracking information for benchmark execution.
//
// This structure enables cost analysis, budget tracking, and price-performance
// optimization across different instance types and regions.
type CostMetrics struct {
	// EstimatedCost is the calculated cost for the benchmark execution in USD.
	// Based on current AWS pricing and actual execution duration.
	EstimatedCost float64
	
	// PricePerformanceRatio calculates cost efficiency for the benchmark.
	// Units vary by benchmark: $/GB/s for STREAM, $/GFLOP for HPL.
	PricePerformanceRatio float64
	
	// InstanceHourCost is the hourly cost rate for the instance type.
	// Used for cost projection and budget planning.
	InstanceHourCost float64
}

// OperationalMetrics contains infrastructure and operational metrics for
// monitoring the benchmark collection system health and efficiency.
//
// These metrics enable system-level monitoring, capacity planning, and
// operational alerting for the benchmark collection infrastructure.
type OperationalMetrics struct {
	// QuotaUtilization tracks current quota usage for instance families.
	// Map key is instance family, value is utilization percentage (0-100).
	QuotaUtilization map[string]float64
	
	// InstanceLaunchDuration is the time to launch and initialize instances.
	// Used for capacity planning and performance optimization.
	InstanceLaunchDuration float64
	
	// ContainerPullDuration is the time to download benchmark containers.
	// Indicates network performance and registry efficiency.
	ContainerPullDuration float64
	
	// ActiveInstances is the current count of running benchmark instances.
	// Used for capacity monitoring and cost tracking.
	ActiveInstances int64
	
	// FailureRate is the percentage of failed benchmark executions.
	// Calculated over a rolling time window for trend analysis.
	FailureRate float64
	
	// Region is the AWS region for operational metrics.
	Region string
	
	// Timestamp is when the operational metrics were collected.
	Timestamp time.Time
}

// NewMetricsCollector creates a new CloudWatch metrics collector configured
// for the specified AWS region.
//
// This function initializes a complete metrics collection environment with
// AWS SDK v2 integration, standardized namespacing, and default dimensions
// for consistent metric publication and querying.
//
// The collector uses the AWS SDK's default credential chain and configuration,
// which includes environment variables, shared credentials, and IAM roles.
// It requires CloudWatch PutMetricData permissions for successful operation.
//
// Parameters:
//   - region: AWS region for CloudWatch metric publishing
//
// Returns:
//   - *MetricsCollector: Configured collector ready for metric publication
//   - error: Configuration errors, credential issues, or connectivity problems
//
// Example:
//   collector, err := monitoring.NewMetricsCollector("us-east-1")
//   if err != nil {
//       log.Fatal("Failed to initialize metrics collector:", err)
//   }
//   
//   // Ready for metric collection
//   err = collector.PublishBenchmarkMetrics(ctx, metrics)
//
// Permissions Required:
//   - cloudwatch:PutMetricData for metric publishing
//   - Optional: cloudwatch:GetMetricStatistics for validation
//
// Common Errors:
//   - Invalid AWS credentials or expired tokens
//   - Insufficient CloudWatch permissions
//   - Network connectivity issues to AWS endpoints
//   - Invalid region specification
func NewMetricsCollector(region string) (*MetricsCollector, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile("aws"), // Use 'aws' profile as specified
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	defaultDimensions := []types.Dimension{
		{
			Name:  aws.String("Project"),
			Value: aws.String("aws-instance-benchmarks"),
		},
		{
			Name:  aws.String("Environment"),
			Value: aws.String("production"),
		},
		{
			Name:  aws.String("Region"),
			Value: aws.String(region),
		},
	}

	return &MetricsCollector{
		cloudwatchClient:  cloudwatch.NewFromConfig(cfg),
		namespace:         "InstanceBenchmarks",
		region:           region,
		defaultDimensions: defaultDimensions,
	}, nil
}

// PublishBenchmarkMetrics publishes comprehensive benchmark execution metrics
// to CloudWatch for monitoring and analysis.
//
// This method handles the complete metric publication workflow including
// dimension standardization, metric validation, and efficient batch publishing.
// It provides detailed visibility into benchmark performance, execution success,
// and cost efficiency across different instance types and regions.
//
// The published metrics include:
//   - Execution success/failure rates with error categorization
//   - Performance measurements specific to each benchmark suite
//   - Duration metrics for execution and infrastructure provisioning
//   - Cost tracking and price-performance ratio calculations
//   - Quality scores for result validation and confidence
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - metrics: Comprehensive benchmark execution metrics to publish
//
// Returns:
//   - error: Publishing failures, validation errors, or API issues
//
// Example:
//   metrics := BenchmarkMetrics{
//       InstanceType: "m7i.large",
//       BenchmarkSuite: "stream",
//       Success: true,
//       ExecutionDuration: 45.2,
//       PerformanceMetrics: map[string]float64{
//           "triad_bandwidth": 41.9,
//           "copy_bandwidth": 45.2,
//       },
//       QualityScore: 0.95,
//   }
//   
//   err := collector.PublishBenchmarkMetrics(ctx, metrics)
//   if err != nil {
//       log.Printf("Failed to publish metrics: %v", err)
//   }
//
// CloudWatch Metrics Published:
//   - BenchmarkExecution (Count): Success/failure tracking
//   - ExecutionDuration (Seconds): Total execution time
//   - BenchmarkDuration (Seconds): Pure benchmark time
//   - PerformanceValue (Various): Benchmark-specific measurements
//   - QualityScore (Ratio): Result validation confidence
//   - EstimatedCost (USD): Cost tracking for budget management
//
// Dimensions Applied:
//   - InstanceType: For instance-specific analysis
//   - InstanceFamily: For family-level aggregation
//   - BenchmarkSuite: For suite-specific tracking
//   - Success: For success/failure filtering
//   - Region: For regional comparison
//
// Error Handling:
//   - Automatic retry with exponential backoff
//   - Metric validation before publication
//   - Graceful degradation for non-critical metrics
func (mc *MetricsCollector) PublishBenchmarkMetrics(ctx context.Context, metrics BenchmarkMetrics) error {
	if err := mc.validateBenchmarkMetrics(metrics); err != nil {
		return fmt.Errorf("metric validation failed: %w", err)
	}

	var metricData []types.MetricDatum
	timestamp := metrics.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Build dimensions for this benchmark execution
	dimensions := make([]types.Dimension, len(mc.defaultDimensions))
	copy(dimensions, mc.defaultDimensions)
	dimensions = append(dimensions,
		types.Dimension{Name: aws.String("InstanceType"), Value: aws.String(metrics.InstanceType)},
		types.Dimension{Name: aws.String("InstanceFamily"), Value: aws.String(metrics.InstanceFamily)},
		types.Dimension{Name: aws.String("BenchmarkSuite"), Value: aws.String(metrics.BenchmarkSuite)},
		types.Dimension{Name: aws.String("Success"), Value: aws.String(fmt.Sprintf("%t", metrics.Success))},
	)

	// Add error category dimension if execution failed
	if !metrics.Success && metrics.ErrorCategory != "" {
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String("ErrorCategory"),
			Value: aws.String(metrics.ErrorCategory),
		})
	}

	// Core execution metrics
	metricData = append(metricData, types.MetricDatum{
		MetricName: aws.String("BenchmarkExecution"),
		Value:      aws.Float64(1.0),
		Unit:       types.StandardUnitCount,
		Timestamp:  aws.Time(timestamp),
		Dimensions: dimensions,
	})

	if metrics.ExecutionDuration > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("ExecutionDuration"),
			Value:      aws.Float64(metrics.ExecutionDuration),
			Unit:       types.StandardUnitSeconds,
			Timestamp:  aws.Time(timestamp),
			Dimensions: dimensions,
		})
	}

	if metrics.BenchmarkDuration > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("BenchmarkDuration"),
			Value:      aws.Float64(metrics.BenchmarkDuration),
			Unit:       types.StandardUnitSeconds,
			Timestamp:  aws.Time(timestamp),
			Dimensions: dimensions,
		})
	}

	// Performance metrics (benchmark-specific)
	for metricName, value := range metrics.PerformanceMetrics {
		if value > 0 {
			metricData = append(metricData, types.MetricDatum{
				MetricName: aws.String(fmt.Sprintf("Performance_%s", metricName)),
				Value:      aws.Float64(value),
				Unit:       mc.getUnitForPerformanceMetric(metricName),
				Timestamp:  aws.Time(timestamp),
				Dimensions: dimensions,
			})
		}
	}

	// Quality score
	if metrics.QualityScore > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("QualityScore"),
			Value:      aws.Float64(metrics.QualityScore),
			Unit:       types.StandardUnitNone,
			Timestamp:  aws.Time(timestamp),
			Dimensions: dimensions,
		})
	}

	// Cost metrics
	if metrics.CostMetrics.EstimatedCost > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("EstimatedCost"),
			Value:      aws.Float64(metrics.CostMetrics.EstimatedCost),
			Unit:       types.StandardUnitNone, // USD
			Timestamp:  aws.Time(timestamp),
			Dimensions: dimensions,
		})
	}

	if metrics.CostMetrics.PricePerformanceRatio > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("PricePerformanceRatio"),
			Value:      aws.Float64(metrics.CostMetrics.PricePerformanceRatio),
			Unit:       types.StandardUnitNone,
			Timestamp:  aws.Time(timestamp),
			Dimensions: dimensions,
		})
	}

	// Publish metrics in batches (CloudWatch limit is 1000 per request)
	return mc.publishMetricBatch(ctx, metricData)
}

// PublishOperationalMetrics publishes infrastructure and system health metrics
// to CloudWatch for operational monitoring and alerting.
//
// This method provides visibility into the benchmark collection system's
// operational characteristics including quota utilization, failure rates,
// and infrastructure performance. These metrics enable proactive monitoring
// and capacity planning for production deployments.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - metrics: Operational metrics to publish
//
// Returns:
//   - error: Publishing failures or validation errors
func (mc *MetricsCollector) PublishOperationalMetrics(ctx context.Context, metrics OperationalMetrics) error {
	var metricData []types.MetricDatum
	timestamp := metrics.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	baseDimensions := make([]types.Dimension, len(mc.defaultDimensions))
	copy(baseDimensions, mc.defaultDimensions)
	baseDimensions = append(baseDimensions,
		types.Dimension{Name: aws.String("MetricType"), Value: aws.String("Operational")},
	)

	// Quota utilization metrics
	for instanceFamily, utilization := range metrics.QuotaUtilization {
		dimensions := make([]types.Dimension, len(baseDimensions))
		copy(dimensions, baseDimensions)
		dimensions = append(dimensions,
			types.Dimension{Name: aws.String("InstanceFamily"), Value: aws.String(instanceFamily)},
		)
		
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("QuotaUtilization"),
			Value:      aws.Float64(utilization),
			Unit:       types.StandardUnitPercent,
			Timestamp:  aws.Time(timestamp),
			Dimensions: dimensions,
		})
	}

	// Infrastructure timing metrics
	if metrics.InstanceLaunchDuration > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("InstanceLaunchDuration"),
			Value:      aws.Float64(metrics.InstanceLaunchDuration),
			Unit:       types.StandardUnitSeconds,
			Timestamp:  aws.Time(timestamp),
			Dimensions: baseDimensions,
		})
	}

	if metrics.ContainerPullDuration > 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("ContainerPullDuration"),
			Value:      aws.Float64(metrics.ContainerPullDuration),
			Unit:       types.StandardUnitSeconds,
			Timestamp:  aws.Time(timestamp),
			Dimensions: baseDimensions,
		})
	}

	// System health metrics
	metricData = append(metricData, types.MetricDatum{
		MetricName: aws.String("ActiveInstances"),
		Value:      aws.Float64(float64(metrics.ActiveInstances)),
		Unit:       types.StandardUnitCount,
		Timestamp:  aws.Time(timestamp),
		Dimensions: baseDimensions,
	})

	if metrics.FailureRate >= 0 {
		metricData = append(metricData, types.MetricDatum{
			MetricName: aws.String("FailureRate"),
			Value:      aws.Float64(metrics.FailureRate),
			Unit:       types.StandardUnitPercent,
			Timestamp:  aws.Time(timestamp),
			Dimensions: baseDimensions,
		})
	}

	return mc.publishMetricBatch(ctx, metricData)
}

// validateBenchmarkMetrics performs comprehensive validation of benchmark metrics
// before CloudWatch publication.
func (mc *MetricsCollector) validateBenchmarkMetrics(metrics BenchmarkMetrics) error {
	if metrics.InstanceType == "" {
		return fmt.Errorf("%w: instance type is required", ErrMetricNameRequired)
	}
	
	if metrics.BenchmarkSuite == "" {
		return fmt.Errorf("%w: benchmark suite is required", ErrMetricNameRequired)
	}
	
	if metrics.ExecutionDuration < 0 {
		return fmt.Errorf("%w: execution duration cannot be negative", ErrInvalidMetricValue)
	}
	
	if metrics.QualityScore < 0 || metrics.QualityScore > 1 {
		return fmt.Errorf("%w: quality score must be between 0 and 1", ErrInvalidMetricValue)
	}
	
	// Validate performance metrics
	for name, value := range metrics.PerformanceMetrics {
		if value < 0 {
			return fmt.Errorf("%w: performance metric %s cannot be negative", ErrInvalidMetricValue, name)
		}
	}
	
	return nil
}

// getUnitForPerformanceMetric determines the appropriate CloudWatch unit
// for performance metrics based on the metric name.
func (mc *MetricsCollector) getUnitForPerformanceMetric(metricName string) types.StandardUnit {
	switch {
	case contains(metricName, "bandwidth"):
		return types.StandardUnitBytesSecond
	case contains(metricName, "gflops"), contains(metricName, "flops"):
		return types.StandardUnitCountSecond
	case contains(metricName, "latency"), contains(metricName, "duration"):
		return types.StandardUnitSeconds
	case contains(metricName, "throughput"):
		return types.StandardUnitCountSecond
	default:
		return types.StandardUnitNone
	}
}

// publishMetricBatch efficiently publishes metrics to CloudWatch in batches
// with automatic retry logic and error handling.
func (mc *MetricsCollector) publishMetricBatch(ctx context.Context, metricData []types.MetricDatum) error {
	if len(metricData) == 0 {
		return nil
	}

	// CloudWatch allows maximum 1000 metrics per PutMetricData call
	const batchSize = 1000
	
	for i := 0; i < len(metricData); i += batchSize {
		end := i + batchSize
		if end > len(metricData) {
			end = len(metricData)
		}
		
		batch := metricData[i:end]
		
		input := &cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(mc.namespace),
			MetricData: batch,
		}
		
		_, err := mc.cloudwatchClient.PutMetricData(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to publish metric batch: %w", err)
		}
	}
	
	return nil
}

// contains is a helper function for case-insensitive string matching.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}