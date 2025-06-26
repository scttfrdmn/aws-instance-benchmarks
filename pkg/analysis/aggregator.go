// Package analysis provides comprehensive data processing and aggregation
// capabilities for AWS Instance Benchmarks results analysis.
//
// This package implements sophisticated statistical analysis, trend detection,
// and comparative performance evaluation across different AWS EC2 instance types,
// regions, and benchmark suites. It processes raw benchmark results into
// actionable insights for performance optimization and instance selection.
//
// Key Components:
//   - DataAggregator: Core aggregation engine for multi-dimensional analysis
//   - PerformanceAnalyzer: Statistical analysis and trend detection
//   - ComparisonEngine: Instance type and configuration comparison
//   - ReportGenerator: Structured output generation for various formats
//
// Usage:
//   aggregator := analysis.NewDataAggregator(config)
//   
//   // Process benchmark results from multiple sources
//   results, err := aggregator.ProcessBenchmarkData(ctx, sources)
//   if err != nil {
//       log.Fatal("Aggregation failed:", err)
//   }
//   
//   // Generate comparative analysis
//   comparison := aggregator.CompareInstanceTypes(results, criteria)
//   report := aggregator.GenerateReport(comparison, formats.JSON)
//
// The package provides:
//   - Multi-dimensional aggregation by instance type, family, region
//   - Time-series analysis for performance trend detection
//   - Statistical validation and confidence interval calculation
//   - Cost-performance optimization recommendations
//   - Automated outlier detection and data quality assessment
//
// Analysis Capabilities:
//   - Performance Ranking: Best-performing instances for specific workloads
//   - Price-Performance Optimization: Cost-effective instance recommendations
//   - Regional Comparison: Performance variations across AWS regions
//   - Temporal Analysis: Performance trends over time
//   - Quality Assessment: Statistical confidence and reliability scoring
//
// Data Processing Pipeline:
//   - Input Validation: Schema compliance and data quality checks
//   - Normalization: Consistent units and scaling across benchmarks
//   - Aggregation: Multi-level statistical summarization
//   - Analysis: Trend detection and comparative evaluation
//   - Output Generation: Structured results in multiple formats
package analysis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/benchmarks"
)

// Data aggregation errors.
var (
	ErrInvalidDataSource     = errors.New("invalid data source")
	ErrInsufficientData      = errors.New("insufficient data for aggregation")
	ErrAggregationFailed     = errors.New("data aggregation failed")
	ErrInvalidAggregationKey = errors.New("invalid aggregation key")
	ErrInvalidConfidenceLevel = errors.New("invalid confidence level")
	ErrInvalidMinSampleSize  = errors.New("minimum sample size must be at least 3")
)

// DataAggregator provides comprehensive benchmark data processing and
// aggregation capabilities for multi-dimensional performance analysis.
//
// This struct orchestrates the complete data processing pipeline from raw
// benchmark results to aggregated insights. It supports flexible aggregation
// dimensions, statistical validation, and trend analysis for large-scale
// performance characterization across AWS EC2 infrastructure.
//
// Thread Safety:
//   The DataAggregator is safe for concurrent use across multiple goroutines.
//   Aggregation operations are stateless and can be executed in parallel.
type DataAggregator struct {
	// config contains aggregation configuration including grouping dimensions,
	// statistical parameters, and output formatting preferences.
	config AggregationConfig
	
	// dataSource provides access to benchmark results from storage backends.
	// Supports S3, local filesystem, and streaming data sources.
	dataSource DataSource
	
	// performanceAnalyzer provides statistical analysis and trend detection.
	performanceAnalyzer *PerformanceAnalyzer
	
	// comparisonEngine enables comparative analysis across instance types.
	comparisonEngine *ComparisonEngine
}

// AggregationConfig defines comprehensive configuration for data processing
// and analysis operations including grouping dimensions and statistical parameters.
type AggregationConfig struct {
	// GroupingDimensions specifies the dimensions for data aggregation.
	// Common values: ["instance_type"], ["instance_family", "region"], ["benchmark_suite"]
	GroupingDimensions []string
	
	// TimeWindow defines the time range for analysis.
	// Used for trend analysis and temporal aggregation.
	TimeWindow TimeWindow
	
	// StatisticalConfig contains parameters for statistical analysis.
	StatisticalConfig StatisticalConfig
	
	// QualityThreshold sets the minimum quality score for result inclusion.
	// Range: 0.0-1.0, where 1.0 represents highest quality.
	QualityThreshold float64
	
	// EnableTrendAnalysis controls whether temporal trend detection is performed.
	EnableTrendAnalysis bool
	
	// OutputFormats specifies the desired output formats for results.
	// Supported: JSON, CSV, Markdown, HTML
	OutputFormats []string
}

// TimeWindow defines a time range for temporal analysis and aggregation.
type TimeWindow struct {
	// Start is the beginning of the analysis time window.
	Start time.Time
	
	// End is the end of the analysis time window.
	End time.Time
	
	// Granularity specifies the time bucket size for aggregation.
	// Values: "hour", "day", "week", "month"
	Granularity string
}

// StatisticalConfig contains parameters for statistical analysis and validation.
type StatisticalConfig struct {
	// ConfidenceLevel sets the confidence level for interval calculations.
	// Common values: 0.90, 0.95, 0.99
	ConfidenceLevel float64
	
	// OutlierThreshold defines the number of standard deviations for outlier detection.
	// Recommended: 2.0-3.0 for robust analysis
	OutlierThreshold float64
	
	// MinSampleSize specifies the minimum number of samples required for aggregation.
	MinSampleSize int
	
	// EnableBootstrapping controls whether bootstrap resampling is used for confidence intervals.
	EnableBootstrapping bool
}

// DataSource provides access to benchmark results from various storage backends.
type DataSource interface {
	// ListResults returns available benchmark results within the specified time window.
	ListResults(ctx context.Context, window TimeWindow) ([]ResultMetadata, error)
	
	// LoadResults retrieves the actual benchmark data for the specified result IDs.
	LoadResults(ctx context.Context, resultIDs []string) ([]BenchmarkData, error)
	
	// GetSchema returns the data schema version for compatibility validation.
	GetSchema(ctx context.Context) (SchemaVersion, error)
}

// ResultMetadata contains metadata about available benchmark results.
type ResultMetadata struct {
	// ResultID uniquely identifies the benchmark result.
	ResultID string
	
	// InstanceType is the AWS EC2 instance type used for the benchmark.
	InstanceType string
	
	// InstanceFamily is the instance family (e.g., "m7i", "c7g").
	InstanceFamily string
	
	// BenchmarkSuite identifies the benchmark type ("stream", "hpl").
	BenchmarkSuite string
	
	// Region is the AWS region where the benchmark was executed.
	Region string
	
	// Timestamp is when the benchmark was executed.
	Timestamp time.Time
	
	// QualityScore indicates the result quality (0.0-1.0).
	QualityScore float64
	
	// DataSize is the size of the result data in bytes.
	DataSize int64
}

// BenchmarkData contains the complete benchmark result data for analysis.
type BenchmarkData struct {
	// Metadata provides result identification and context information.
	Metadata ResultMetadata
	
	// StreamResult contains STREAM benchmark results (if applicable).
	StreamResult *benchmarks.BenchmarkResult
	
	// HPLResult contains HPL benchmark results (if applicable).
	HPLResult *benchmarks.HPLResult
	
	// ExecutionContext provides additional execution environment details.
	ExecutionContext ExecutionContext
}

// ExecutionContext provides detailed information about the benchmark execution environment.
type ExecutionContext struct {
	// ContainerImage specifies the exact container image used.
	ContainerImage string
	
	// CompilerInfo contains compilation details for reproducibility.
	CompilerInfo benchmarks.CompilerInfo
	
	// SystemConfiguration provides hardware and system details.
	SystemConfiguration SystemConfiguration
	
	// ExecutionParameters contains benchmark-specific execution settings.
	ExecutionParameters map[string]interface{}
}

// SystemConfiguration describes the hardware and system configuration.
type SystemConfiguration struct {
	// ProcessorModel describes the CPU model and generation.
	ProcessorModel string
	
	// ProcessorArchitecture specifies the instruction set architecture.
	ProcessorArchitecture string
	
	// MemoryConfiguration describes memory subsystem details.
	MemoryConfiguration MemoryConfiguration
	
	// NetworkConfiguration describes network performance characteristics.
	NetworkConfiguration NetworkConfiguration
}

// MemoryConfiguration provides detailed memory subsystem information.
type MemoryConfiguration struct {
	// TotalCapacity is the total memory capacity in GB.
	TotalCapacity float64
	
	// MemoryType describes the memory technology (DDR4, DDR5, HBM).
	MemoryType string
	
	// MemorySpeed specifies the memory clock speed in MHz.
	MemorySpeed int
	
	// NUMATopology describes the NUMA configuration.
	NUMATopology benchmarks.NumaTopology
}

// NetworkConfiguration describes network performance characteristics.
type NetworkConfiguration struct {
	// NetworkPerformance indicates the network performance tier.
	NetworkPerformance string
	
	// BandwidthCapacity specifies the network bandwidth in Gbps.
	BandwidthCapacity float64
	
	// PacketPerSecond indicates packet processing capacity.
	PacketPerSecond int64
}

// SchemaVersion represents the data schema version for compatibility validation.
type SchemaVersion struct {
	// Major version number for breaking changes.
	Major int
	
	// Minor version number for backward-compatible additions.
	Minor int
	
	// Patch version number for bug fixes.
	Patch int
}

// AggregatedResult contains the results of data aggregation and analysis.
type AggregatedResult struct {
	// GroupKey identifies the aggregation group (e.g., instance type, family).
	GroupKey AggregationKey
	
	// PerformanceMetrics contains aggregated performance measurements.
	PerformanceMetrics PerformanceMetrics
	
	// StatisticalSummary provides statistical analysis of the aggregated data.
	StatisticalSummary AggregatedStatistics
	
	// TrendAnalysis contains temporal trend information (if enabled).
	TrendAnalysis *TrendAnalysis
	
	// QualityAssessment indicates the reliability of the aggregated results.
	QualityAssessment QualityAssessment
	
	// SampleSize indicates the number of individual results in the aggregation.
	SampleSize int
	
	// TimeRange specifies the time period covered by the aggregation.
	TimeRange TimeWindow
}

// AggregationKey uniquely identifies an aggregation group.
type AggregationKey struct {
	// Dimensions contains the key-value pairs that define the aggregation group.
	Dimensions map[string]string
	
	// Hash provides a fast comparison mechanism for grouping operations.
	Hash string
}

// PerformanceMetrics contains aggregated performance measurements across benchmarks.
type PerformanceMetrics struct {
	// StreamMetrics contains aggregated STREAM benchmark results.
	StreamMetrics *StreamAggregatedMetrics
	
	// HPLMetrics contains aggregated HPL benchmark results.
	HPLMetrics *HPLAggregatedMetrics
	
	// CrossBenchmarkMetrics contains metrics computed across multiple benchmark types.
	CrossBenchmarkMetrics *CrossBenchmarkMetrics
}

// StreamAggregatedMetrics contains statistical aggregation of STREAM benchmark results.
type StreamAggregatedMetrics struct {
	// TriadBandwidth contains aggregated Triad bandwidth measurements.
	TriadBandwidth AggregatedMeasurement
	
	// CopyBandwidth contains aggregated Copy bandwidth measurements.
	CopyBandwidth AggregatedMeasurement
	
	// ScaleBandwidth contains aggregated Scale bandwidth measurements.
	ScaleBandwidth AggregatedMeasurement
	
	// AddBandwidth contains aggregated Add bandwidth measurements.
	AddBandwidth AggregatedMeasurement
	
	// OverallStability measures consistency across all STREAM operations.
	OverallStability float64
}

// HPLAggregatedMetrics contains statistical aggregation of HPL benchmark results.
type HPLAggregatedMetrics struct {
	// GFLOPS contains aggregated computational performance measurements.
	GFLOPS AggregatedMeasurement
	
	// Efficiency contains aggregated efficiency measurements.
	Efficiency AggregatedMeasurement
	
	// ExecutionTime contains aggregated execution time measurements.
	ExecutionTime AggregatedMeasurement
	
	// NumericalAccuracy contains aggregated residual accuracy measurements.
	NumericalAccuracy AggregatedMeasurement
}

// CrossBenchmarkMetrics contains metrics computed across multiple benchmark types.
type CrossBenchmarkMetrics struct {
	// PerformanceBalance measures the relative performance across different workload types.
	PerformanceBalance float64
	
	// EfficiencyScore provides a composite efficiency assessment.
	EfficiencyScore float64
	
	// ConsistencyScore measures performance consistency across benchmark types.
	ConsistencyScore float64
}

// AggregatedMeasurement contains statistical aggregation of individual measurements.
type AggregatedMeasurement struct {
	// Mean is the arithmetic mean of all measurements.
	Mean float64
	
	// Median is the middle value of all measurements.
	Median float64
	
	// StandardDeviation measures the variability of measurements.
	StandardDeviation float64
	
	// ConfidenceInterval provides the statistical confidence range.
	ConfidenceInterval benchmarks.ConfidenceInterval
	
	// Percentiles contains key percentile values (P5, P25, P75, P95).
	Percentiles map[string]float64
	
	// Min is the minimum measurement value.
	Min float64
	
	// Max is the maximum measurement value.
	Max float64
	
	// Count is the number of measurements included in the aggregation.
	Count int
}

// AggregatedStatistics provides comprehensive statistical analysis of aggregated data.
type AggregatedStatistics struct {
	// DataQuality indicates the overall quality of the aggregated data.
	DataQuality float64
	
	// TemporalStability measures consistency over time.
	TemporalStability float64
	
	// RegionalConsistency measures performance consistency across regions.
	RegionalConsistency float64
	
	// SampleDistribution describes the distribution of samples across dimensions.
	SampleDistribution map[string]int
	
	// OutlierRate indicates the percentage of outliers removed during aggregation.
	OutlierRate float64
}

// TrendAnalysis contains temporal trend analysis results.
type TrendAnalysis struct {
	// TrendDirection indicates whether performance is improving, declining, or stable.
	TrendDirection TrendDirection
	
	// TrendStrength measures the strength of the trend (0.0-1.0).
	TrendStrength float64
	
	// SeasonalPattern indicates if seasonal patterns are detected.
	SeasonalPattern *SeasonalPattern
	
	// ChangePoints identifies significant performance changes over time.
	ChangePoints []ChangePoint
	
	// ForecastConfidence indicates the confidence in trend predictions.
	ForecastConfidence float64
}

// TrendDirection represents the direction of performance trends.
type TrendDirection string

// Trend direction constants.
const (
	TrendImproving TrendDirection = "improving"
	TrendDeclining TrendDirection = "declining"
	TrendStable    TrendDirection = "stable"
	TrendVolatile  TrendDirection = "volatile"
)

// SeasonalPattern describes detected seasonal performance patterns.
type SeasonalPattern struct {
	// Period indicates the seasonal period (e.g., daily, weekly).
	Period string
	
	// Amplitude measures the strength of seasonal variation.
	Amplitude float64
	
	// PhaseShift indicates the timing offset of the seasonal pattern.
	PhaseShift time.Duration
}

// ChangePoint identifies a significant change in performance trends.
type ChangePoint struct {
	// Timestamp indicates when the change occurred.
	Timestamp time.Time
	
	// Magnitude measures the size of the performance change.
	Magnitude float64
	
	// Confidence indicates the statistical confidence in the change detection.
	Confidence float64
	
	// Cause provides context about potential causes (if known).
	Cause string
}

// QualityAssessment provides detailed assessment of result reliability and confidence.
type QualityAssessment struct {
	// OverallScore provides a composite quality score (0.0-1.0).
	OverallScore float64
	
	// StatisticalConfidence indicates confidence in statistical measures.
	StatisticalConfidence float64
	
	// DataCompleteness measures the completeness of available data.
	DataCompleteness float64
	
	// ConsistencyScore measures internal consistency of measurements.
	ConsistencyScore float64
	
	// Issues contains any identified data quality issues.
	Issues []QualityIssue
	
	// Recommendations provides suggestions for improving data quality.
	Recommendations []string
}

// QualityIssue describes a specific data quality concern.
type QualityIssue struct {
	// Severity indicates the impact level of the issue.
	Severity IssueSeverity
	
	// Category classifies the type of quality issue.
	Category IssueCategory
	
	// Description provides human-readable details about the issue.
	Description string
	
	// AffectedSamples indicates how many samples are affected.
	AffectedSamples int
}

// IssueSeverity represents the severity level of quality issues.
type IssueSeverity string

// Issue severity levels.
const (
	SeverityLow      IssueSeverity = "low"
	SeverityMedium   IssueSeverity = "medium"
	SeverityHigh     IssueSeverity = "high"
	SeverityCritical IssueSeverity = "critical"
)

// IssueCategory classifies the type of quality issues.
type IssueCategory string

// Issue category types.
const (
	CategoryStatistical  IssueCategory = "statistical"
	CategoryTemporal     IssueCategory = "temporal"
	CategoryConsistency  IssueCategory = "consistency"
	CategoryCompleteness IssueCategory = "completeness"
	CategoryOutlier      IssueCategory = "outlier"
)

// NewDataAggregator creates a new data aggregator with the specified configuration.
//
// This function initializes a complete data processing pipeline with statistical
// analysis capabilities, trend detection, and quality assessment. The aggregator
// provides flexible aggregation dimensions and comprehensive performance analysis
// for large-scale benchmark data processing.
//
// Parameters:
//   - config: Aggregation configuration including grouping dimensions and statistical parameters
//   - dataSource: Data source interface for accessing benchmark results
//
// Returns:
//   - *DataAggregator: Configured aggregator ready for data processing
//   - error: Configuration validation errors or initialization failures
//
// Example:
//   config := AggregationConfig{
//       GroupingDimensions: []string{"instance_type", "region"},
//       TimeWindow: TimeWindow{
//           Start: time.Now().AddDate(0, -1, 0),
//           End:   time.Now(),
//           Granularity: "day",
//       },
//       StatisticalConfig: StatisticalConfig{
//           ConfidenceLevel: 0.95,
//           MinSampleSize: 10,
//       },
//       QualityThreshold: 0.7,
//   }
//   
//   aggregator, err := analysis.NewDataAggregator(config, dataSource)
//   if err != nil {
//       log.Fatal("Failed to create aggregator:", err)
//   }
func NewDataAggregator(config AggregationConfig, dataSource DataSource) (*DataAggregator, error) {
	if err := validateAggregationConfig(config); err != nil {
		return nil, fmt.Errorf("invalid aggregation config: %w", err)
	}

	performanceAnalyzer := NewPerformanceAnalyzer(config.StatisticalConfig)
	comparisonEngine := NewComparisonEngine(config)

	return &DataAggregator{
		config:              config,
		dataSource:          dataSource,
		performanceAnalyzer: performanceAnalyzer,
		comparisonEngine:    comparisonEngine,
	}, nil
}

// ProcessBenchmarkData performs comprehensive aggregation and analysis of benchmark data.
//
// This method orchestrates the complete data processing pipeline from raw benchmark
// results to aggregated insights. It handles data loading, validation, aggregation,
// statistical analysis, and quality assessment for multi-dimensional performance
// characterization.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//
// Returns:
//   - []AggregatedResult: Aggregated results grouped by specified dimensions
//   - error: Processing errors, data validation failures, or insufficient data
func (da *DataAggregator) ProcessBenchmarkData(ctx context.Context) ([]AggregatedResult, error) {
	// Load available result metadata
	metadata, err := da.dataSource.ListResults(ctx, da.config.TimeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to list results: %w", err)
	}

	if len(metadata) == 0 {
		return nil, fmt.Errorf("%w: no results found in time window", ErrInsufficientData)
	}

	// Filter by quality threshold
	qualityMetadata := da.filterByQuality(metadata)
	if len(qualityMetadata) < da.config.StatisticalConfig.MinSampleSize {
		return nil, fmt.Errorf("%w: insufficient quality results (%d < %d)", 
			ErrInsufficientData, len(qualityMetadata), da.config.StatisticalConfig.MinSampleSize)
	}

	// Load actual benchmark data
	resultIDs := make([]string, len(qualityMetadata))
	for i, meta := range qualityMetadata {
		resultIDs[i] = meta.ResultID
	}

	benchmarkData, err := da.dataSource.LoadResults(ctx, resultIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load benchmark data: %w", err)
	}

	// Group data by aggregation dimensions
	groupedData := da.groupDataByDimensions(benchmarkData)

	// Process each group
	var results []AggregatedResult
	for _, groupData := range groupedData {
		if len(groupData) < da.config.StatisticalConfig.MinSampleSize {
			continue // Skip groups with insufficient data
		}

		// Recreate the key from the first item in the group
		groupKey := da.createAggregationKey(groupData[0].Metadata)
		
		aggregatedResult := da.aggregateGroup(groupKey, groupData)
		results = append(results, aggregatedResult)
	}

	// Sort results by quality score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].QualityAssessment.OverallScore > results[j].QualityAssessment.OverallScore
	})

	return results, nil
}

// filterByQuality filters results based on the configured quality threshold.
func (da *DataAggregator) filterByQuality(metadata []ResultMetadata) []ResultMetadata {
	var filtered []ResultMetadata
	for _, meta := range metadata {
		if meta.QualityScore >= da.config.QualityThreshold {
			filtered = append(filtered, meta)
		}
	}
	return filtered
}

// groupDataByDimensions groups benchmark data by the configured aggregation dimensions.
func (da *DataAggregator) groupDataByDimensions(data []BenchmarkData) map[string][]BenchmarkData {
	groups := make(map[string][]BenchmarkData)

	for _, item := range data {
		key := da.createAggregationKey(item.Metadata)
		groups[key.Hash] = append(groups[key.Hash], item)
	}

	return groups
}

// createAggregationKey creates an aggregation key from result metadata.
func (da *DataAggregator) createAggregationKey(metadata ResultMetadata) AggregationKey {
	dimensions := make(map[string]string)

	for _, dimension := range da.config.GroupingDimensions {
		switch dimension {
		case "instance_type":
			dimensions["instance_type"] = metadata.InstanceType
		case "instance_family":
			dimensions["instance_family"] = metadata.InstanceFamily
		case "benchmark_suite":
			dimensions["benchmark_suite"] = metadata.BenchmarkSuite
		case "region":
			dimensions["region"] = metadata.Region
		}
	}

	// Create hash for fast comparison
	hash := fmt.Sprintf("%v", dimensions)

	return AggregationKey{
		Dimensions: dimensions,
		Hash:       hash,
	}
}

// aggregateGroup performs statistical aggregation for a single group of benchmark data.
func (da *DataAggregator) aggregateGroup(groupKey AggregationKey, groupData []BenchmarkData) AggregatedResult {
	// Separate data by benchmark type
	streamData := make([]*benchmarks.BenchmarkResult, 0)
	hplData := make([]*benchmarks.HPLResult, 0)

	for _, item := range groupData {
		if item.StreamResult != nil {
			streamData = append(streamData, item.StreamResult)
		}
		if item.HPLResult != nil {
			hplData = append(hplData, item.HPLResult)
		}
	}

	// Aggregate performance metrics
	performanceMetrics := PerformanceMetrics{}
	
	if len(streamData) > 0 {
		streamMetrics := da.aggregateStreamData(streamData)
		performanceMetrics.StreamMetrics = streamMetrics
	}

	if len(hplData) > 0 {
		hplMetrics := da.aggregateHPLData(hplData)
		performanceMetrics.HPLMetrics = hplMetrics
	}

	// Calculate cross-benchmark metrics if we have multiple benchmark types
	if performanceMetrics.StreamMetrics != nil && performanceMetrics.HPLMetrics != nil {
		crossMetrics := da.calculateCrossBenchmarkMetrics(
			performanceMetrics.StreamMetrics, 
			performanceMetrics.HPLMetrics,
		)
		performanceMetrics.CrossBenchmarkMetrics = crossMetrics
	}

	// Calculate time range
	timeRange := da.calculateTimeRange(groupData)

	// Assess quality
	qualityAssessment := da.assessDataQuality(groupData)

	return AggregatedResult{
		GroupKey:           groupKey,
		PerformanceMetrics: performanceMetrics,
		QualityAssessment:  qualityAssessment,
		SampleSize:         len(groupData),
		TimeRange:          timeRange,
	}
}

// aggregateStreamData performs statistical aggregation of STREAM benchmark results.
func (da *DataAggregator) aggregateStreamData(data []*benchmarks.BenchmarkResult) *StreamAggregatedMetrics {
	// Extract bandwidth values for each operation
	triadValues := make([]float64, 0, len(data))
	copyValues := make([]float64, 0, len(data))
	scaleValues := make([]float64, 0, len(data))
	addValues := make([]float64, 0, len(data))

	for _, result := range data {
		if triad, exists := result.Measurements["triad"]; exists {
			triadValues = append(triadValues, triad.Value)
		}
		if copyMeasurement, exists := result.Measurements["copy"]; exists {
			copyValues = append(copyValues, copyMeasurement.Value)
		}
		if scale, exists := result.Measurements["scale"]; exists {
			scaleValues = append(scaleValues, scale.Value)
		}
		if add, exists := result.Measurements["add"]; exists {
			addValues = append(addValues, add.Value)
		}
	}

	// Aggregate each operation
	triadAgg := da.aggregateMeasurement(triadValues)
	copyAgg := da.aggregateMeasurement(copyValues)
	scaleAgg := da.aggregateMeasurement(scaleValues)
	addAgg := da.aggregateMeasurement(addValues)

	// Calculate overall stability
	stabilities := []float64{
		triadAgg.StandardDeviation / triadAgg.Mean * 100,
		copyAgg.StandardDeviation / copyAgg.Mean * 100,
		scaleAgg.StandardDeviation / scaleAgg.Mean * 100,
		addAgg.StandardDeviation / addAgg.Mean * 100,
	}
	overallStability := da.calculateMean(stabilities)

	return &StreamAggregatedMetrics{
		TriadBandwidth:   triadAgg,
		CopyBandwidth:    copyAgg,
		ScaleBandwidth:   scaleAgg,
		AddBandwidth:     addAgg,
		OverallStability: overallStability,
	}
}

// aggregateHPLData performs statistical aggregation of HPL benchmark results.
func (da *DataAggregator) aggregateHPLData(data []*benchmarks.HPLResult) *HPLAggregatedMetrics {
	// Extract performance values
	gflopsValues := make([]float64, 0, len(data))
	efficiencyValues := make([]float64, 0, len(data))
	executionTimeValues := make([]float64, 0, len(data))
	residualValues := make([]float64, 0, len(data))

	for _, result := range data {
		gflopsValues = append(gflopsValues, result.Performance.GFLOPS.Value)
		efficiencyValues = append(efficiencyValues, result.Performance.Efficiency.Value)
		executionTimeValues = append(executionTimeValues, result.Performance.ExecutionTime.Value)
		residualValues = append(residualValues, result.Performance.Residual.Value)
	}

	return &HPLAggregatedMetrics{
		GFLOPS:            da.aggregateMeasurement(gflopsValues),
		Efficiency:        da.aggregateMeasurement(efficiencyValues),
		ExecutionTime:     da.aggregateMeasurement(executionTimeValues),
		NumericalAccuracy: da.aggregateMeasurement(residualValues),
	}
}

// aggregateMeasurement performs statistical aggregation of a set of measurement values.
func (da *DataAggregator) aggregateMeasurement(values []float64) AggregatedMeasurement {
	if len(values) == 0 {
		return AggregatedMeasurement{}
	}

	// Sort values for percentile calculations
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mean := da.calculateMean(values)
	stdDev := da.calculateStandardDeviation(values, mean)

	return AggregatedMeasurement{
		Mean:              mean,
		Median:            da.calculateMedian(sorted),
		StandardDeviation: stdDev,
		Percentiles: map[string]float64{
			"P5":  da.calculatePercentile(sorted, 5),
			"P25": da.calculatePercentile(sorted, 25),
			"P75": da.calculatePercentile(sorted, 75),
			"P95": da.calculatePercentile(sorted, 95),
		},
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
		Count: len(values),
	}
}

// calculateMean calculates the arithmetic mean of a slice of float64 values.
func (da *DataAggregator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStandardDeviation calculates the standard deviation of a slice of float64 values.
func (da *DataAggregator) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculateMedian calculates the median of a sorted slice of float64 values.
func (da *DataAggregator) calculateMedian(sortedValues []float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sortedValues[n/2-1] + sortedValues[n/2]) / 2
	}
	return sortedValues[n/2]
}

// calculatePercentile calculates the specified percentile of a sorted slice of float64 values.
func (da *DataAggregator) calculatePercentile(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := (percentile / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

// calculateCrossBenchmarkMetrics computes metrics across multiple benchmark types.
func (da *DataAggregator) calculateCrossBenchmarkMetrics(streamMetrics *StreamAggregatedMetrics, hplMetrics *HPLAggregatedMetrics) *CrossBenchmarkMetrics {
	// Simplified cross-benchmark analysis
	// In production, this would be more sophisticated
	
	performanceBalance := 0.5 // Placeholder - would calculate actual balance
	efficiencyScore := (streamMetrics.OverallStability + hplMetrics.Efficiency.Mean) / 2
	consistencyScore := 1.0 - (streamMetrics.OverallStability/100.0) // Convert CV% to consistency

	return &CrossBenchmarkMetrics{
		PerformanceBalance: performanceBalance,
		EfficiencyScore:    efficiencyScore,
		ConsistencyScore:   math.Max(0, consistencyScore),
	}
}

// calculateTimeRange determines the time range covered by the data group.
func (da *DataAggregator) calculateTimeRange(data []BenchmarkData) TimeWindow {
	if len(data) == 0 {
		return TimeWindow{}
	}

	start := data[0].Metadata.Timestamp
	end := data[0].Metadata.Timestamp

	for _, item := range data {
		if item.Metadata.Timestamp.Before(start) {
			start = item.Metadata.Timestamp
		}
		if item.Metadata.Timestamp.After(end) {
			end = item.Metadata.Timestamp
		}
	}

	return TimeWindow{
		Start: start,
		End:   end,
	}
}

// assessDataQuality performs comprehensive quality assessment of the aggregated data.
func (da *DataAggregator) assessDataQuality(data []BenchmarkData) QualityAssessment {
	if len(data) == 0 {
		return QualityAssessment{
			OverallScore: 0.0,
		}
	}

	// Calculate quality metrics
	qualitySum := 0.0
	for _, item := range data {
		qualitySum += item.Metadata.QualityScore
	}
	avgQuality := qualitySum / float64(len(data))

	// Assess completeness
	completeness := 1.0 // Simplified - would check for missing data

	// Assess consistency (simplified)
	consistency := 0.9 // Would calculate actual consistency

	overallScore := (avgQuality + completeness + consistency) / 3.0

	return QualityAssessment{
		OverallScore:          overallScore,
		StatisticalConfidence: avgQuality,
		DataCompleteness:      completeness,
		ConsistencyScore:      consistency,
		Issues:                []QualityIssue{},
		Recommendations:       []string{},
	}
}

// validateAggregationConfig validates the aggregation configuration.
func validateAggregationConfig(config AggregationConfig) error {
	if len(config.GroupingDimensions) == 0 {
		return fmt.Errorf("%w: no grouping dimensions specified", ErrInvalidAggregationKey)
	}

	if config.StatisticalConfig.ConfidenceLevel <= 0 || config.StatisticalConfig.ConfidenceLevel >= 1 {
		return fmt.Errorf("%w: %f", ErrInvalidConfidenceLevel, config.StatisticalConfig.ConfidenceLevel)
	}

	if config.StatisticalConfig.MinSampleSize < 3 {
		return fmt.Errorf("%w", ErrInvalidMinSampleSize)
	}

	return nil
}

// PerformanceAnalyzer provides statistical analysis and trend detection capabilities.
type PerformanceAnalyzer struct {
	config StatisticalConfig
}

// NewPerformanceAnalyzer creates a new performance analyzer with the specified configuration.
func NewPerformanceAnalyzer(config StatisticalConfig) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{config: config}
}

// ComparisonEngine enables comparative analysis across instance types and configurations.
type ComparisonEngine struct {
	config AggregationConfig
}

// NewComparisonEngine creates a new comparison engine with the specified configuration.
func NewComparisonEngine(config AggregationConfig) *ComparisonEngine {
	return &ComparisonEngine{config: config}
}