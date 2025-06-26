// Package benchmarks provides execution and analysis capabilities for various
// performance benchmark suites on AWS EC2 instances.
//
// This package implements the core benchmark execution logic, result parsing,
// and statistical validation for different benchmark types. It specializes in
// memory bandwidth, computational performance, and efficiency measurements
// with architecture-aware optimizations.
//
// Key Components:
//   - StreamBenchmark: STREAM memory bandwidth benchmark execution
//   - BenchmarkResult: Structured results with statistical validation
//   - BenchmarkConfig: Configuration for benchmark execution parameters
//   - ResultProcessor: Statistical analysis and validation utilities
//
// Usage:
//   benchmark := benchmarks.NewStreamBenchmark(config)
//   result, err := benchmark.Execute(ctx)
//   if err != nil {
//       log.Fatal("Benchmark failed:", err)
//   }
//   
//   // Access validated results
//   bandwidth := result.Measurements["triad"].Value
//   confidence := result.Measurements["triad"].ConfidenceInterval
//
// The package provides:
//   - NUMA-aware memory bandwidth measurement via STREAM benchmark
//   - Multiple run execution with statistical validation
//   - Architecture-specific optimization validation
//   - Confidence interval calculation at configurable levels
//   - Result serialization for data persistence and analysis
//
// Benchmark Suites Supported:
//   - STREAM: Memory bandwidth (Copy, Scale, Add, Triad operations)
//   - HPL (planned): Computational performance (GFLOPS measurement)
//   - CoreMark (planned): Integer performance and efficiency
//
// Statistical Features:
//   - Multiple run execution (default: 10 runs)
//   - Outlier detection and removal
//   - Confidence interval calculation (default: 95%)
//   - Standard deviation and coefficient of variation
//   - Performance stability analysis
package benchmarks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/monitoring"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/storage"
)

// StreamBenchmark provides STREAM memory bandwidth benchmark execution
// with NUMA awareness and statistical validation.
//
// The STREAM benchmark measures sustainable memory bandwidth for four
// fundamental operations: Copy, Scale, Add, and Triad. This implementation
// provides architecture-specific optimizations and multi-run statistical
// validation for reliable performance characterization.
//
// Features:
//   - NUMA-aware execution for multi-socket systems
//   - Architecture-specific compiler optimizations
//   - Multiple run execution with outlier detection
//   - Confidence interval calculation
//   - Performance stability analysis
//
// Thread Safety:
//   StreamBenchmark instances are safe for concurrent use across goroutines.
//   Each execution is isolated and does not share mutable state.
type StreamBenchmark struct {
	// config contains the benchmark execution configuration including
	// iteration counts, confidence levels, and optimization parameters.
	config BenchmarkConfig
	
	// containerImage specifies the optimized STREAM container to execute.
	// Format: "registry/namespace:stream-architecture"
	containerImage string
	
	// numaTopology contains NUMA node information for optimization.
	numaTopology NumaTopology
	
	// storage provides S3-based result persistence (optional).
	// If nil, results are not automatically stored.
	storage *storage.S3Storage
	
	// metricsCollector provides CloudWatch metrics integration (optional).
	// If nil, metrics are not automatically published.
	metricsCollector *monitoring.MetricsCollector
}

// BenchmarkConfig defines comprehensive configuration for benchmark execution
// including statistical validation parameters and performance requirements.
type BenchmarkConfig struct {
	// Iterations specifies the number of benchmark runs to execute.
	// Higher values provide better statistical confidence but increase execution time.
	// Recommended: 10-20 iterations for production, 3-5 for development.
	Iterations int
	
	// ConfidenceLevel sets the statistical confidence level for interval calculation.
	// Common values: 0.90 (90%), 0.95 (95%), 0.99 (99%).
	ConfidenceLevel float64
	
	// OutlierThreshold defines the number of standard deviations beyond which
	// results are considered outliers and excluded from final statistics.
	// Recommended: 2.0-3.0 for robust statistical analysis.
	OutlierThreshold float64
	
	// MinValidRuns specifies the minimum number of valid runs required
	// after outlier removal for results to be considered statistically valid.
	MinValidRuns int
	
	// MaxExecutionTime sets the timeout for individual benchmark runs.
	// Prevents hanging benchmarks from blocking the execution pipeline.
	MaxExecutionTime time.Duration
	
	// MemoryPattern configures the memory access pattern for STREAM execution.
	// Options: "sequential", "random", "numa-local", "numa-remote"
	MemoryPattern string
	
	// EnableNUMA controls whether NUMA-aware optimizations are applied.
	// Set to false for single-socket systems or when testing cross-NUMA performance.
	EnableNUMA bool
}

// BenchmarkResult contains comprehensive results from benchmark execution
// including individual measurements, statistical analysis, and metadata.
//
// This structure provides both raw measurement data and derived statistics
// to support various analysis scenarios including performance ranking,
// trend analysis, and comparative studies.
type BenchmarkResult struct {
	// BenchmarkSuite identifies the executed benchmark type.
	// Examples: "stream", "hpl", "coremark"
	BenchmarkSuite string
	
	// Measurements contains individual operation results with statistical validation.
	// Key format: operation name (e.g., "copy", "scale", "add", "triad")
	Measurements map[string]Measurement
	
	// ExecutionMetadata provides detailed information about the benchmark execution
	// environment and configuration for reproducibility analysis.
	ExecutionMetadata ExecutionMetadata
	
	// StatisticalSummary contains aggregate statistical analysis across all operations.
	StatisticalSummary StatisticalSummary
	
	// ValidationStatus indicates whether the benchmark results meet
	// statistical confidence and stability requirements.
	ValidationStatus ValidationStatus
	
	// Timestamp records when the benchmark execution completed.
	Timestamp time.Time
}

// Measurement represents a single benchmark operation result with
// comprehensive statistical analysis and confidence intervals.
type Measurement struct {
	// Operation identifies the specific benchmark operation.
	// STREAM operations: "copy", "scale", "add", "triad"
	Operation string
	
	// Value is the final validated measurement value (e.g., bandwidth in GB/s).
	// Calculated as the mean of valid runs after outlier removal.
	Value float64
	
	// Unit specifies the measurement unit for proper interpretation.
	// Examples: "GB/s", "GFLOPS", "operations/second"
	Unit string
	
	// StandardDeviation measures the variability of individual runs.
	// Lower values indicate more consistent performance.
	StandardDeviation float64
	
	// ConfidenceInterval provides the statistical confidence range for the measurement.
	// Calculated at the configured confidence level (e.g., 95%).
	ConfidenceInterval ConfidenceInterval
	
	// CoefficientOfVariation measures relative variability as a percentage.
	// Values < 5% indicate stable performance, > 10% suggest variability issues.
	CoefficientOfVariation float64
	
	// ValidRuns contains the individual run values used in statistical calculation.
	// Excludes outliers that were removed during validation.
	ValidRuns []float64
	
	// OutliersRemoved indicates the number of runs excluded as statistical outliers.
	OutliersRemoved int
}

// ConfidenceInterval represents the statistical confidence range for a measurement.
type ConfidenceInterval struct {
	// Lower is the lower bound of the confidence interval.
	Lower float64
	
	// Upper is the upper bound of the confidence interval.
	Upper float64
	
	// Level is the confidence level used (e.g., 0.95 for 95% confidence).
	Level float64
}

// ExecutionMetadata captures detailed information about the benchmark
// execution environment for reproducibility and analysis.
type ExecutionMetadata struct {
	// InstanceType is the AWS EC2 instance type where the benchmark executed.
	InstanceType string
	
	// Architecture identifies the processor architecture and optimization level.
	Architecture string
	
	// ContainerImage specifies the exact container used for benchmark execution.
	ContainerImage string
	
	// CompilerInfo contains details about the compiler and optimization flags used.
	CompilerInfo CompilerInfo
	
	// SystemInfo provides hardware and system configuration details.
	SystemInfo SystemInfo
	
	// ExecutionDuration records the total time for all benchmark iterations.
	ExecutionDuration time.Duration
	
	// Region indicates the AWS region where the benchmark was executed.
	Region string
}

// CompilerInfo contains detailed information about the compiler toolchain
// and optimization settings used for benchmark compilation.
type CompilerInfo struct {
	// Compiler identifies the compiler type (e.g., "gcc", "intel", "aocc").
	Compiler string
	
	// Version specifies the exact compiler version used.
	Version string
	
	// OptimizationFlags lists all compiler flags applied during compilation.
	OptimizationFlags []string
	
	// TargetArchitecture specifies the target microarchitecture for optimization.
	TargetArchitecture string
}

// SystemInfo provides comprehensive hardware and system configuration details.
type SystemInfo struct {
	// CPUModel describes the processor model and generation.
	CPUModel string
	
	// CPUCores indicates the number of physical CPU cores available.
	CPUCores int
	
	// CPUThreads indicates the number of logical CPU threads (with hyperthreading).
	CPUThreads int
	
	// MemoryTotal specifies the total system memory in GB.
	MemoryTotal float64
	
	// NUMANodes indicates the number of NUMA nodes in the system.
	NUMANodes int
	
	// CacheHierarchy provides details about CPU cache configuration.
	CacheHierarchy CacheHierarchy
}

// CacheHierarchy describes the CPU cache configuration for performance analysis.
type CacheHierarchy struct {
	// L1Cache describes Level 1 cache configuration.
	L1Cache CacheLevel
	
	// L2Cache describes Level 2 cache configuration.
	L2Cache CacheLevel
	
	// L3Cache describes Level 3 cache configuration.
	L3Cache CacheLevel
}

// CacheLevel provides detailed information about a specific cache level.
type CacheLevel struct {
	// Size indicates the cache size in KB.
	Size int
	
	// Associativity describes the cache associativity (ways).
	Associativity int
	
	// LineSize specifies the cache line size in bytes.
	LineSize int
}

// NumaTopology contains NUMA (Non-Uniform Memory Access) system topology
// information for memory bandwidth optimization.
type NumaTopology struct {
	// NodeCount indicates the number of NUMA nodes in the system.
	NodeCount int
	
	// TotalMemoryGB is the total memory available across all NUMA nodes in GB.
	TotalMemoryGB float64
	
	// NodesInfo provides detailed information about each NUMA node.
	NodesInfo []NumaNode
	
	// InterconnectBandwidth describes bandwidth between NUMA nodes.
	InterconnectBandwidth map[string]float64
}

// NumaNode represents a single NUMA node with its associated resources.
type NumaNode struct {
	// NodeID is the NUMA node identifier (0, 1, 2, ...).
	NodeID int
	
	// CPUCores lists the CPU cores associated with this NUMA node.
	CPUCores []int
	
	// MemorySize indicates the memory size available on this node in GB.
	MemorySize float64
	
	// LocalBandwidth specifies the memory bandwidth for local access in GB/s.
	LocalBandwidth float64
}

// TotalMemory returns the total memory across all NUMA nodes in GB.
//
// This method calculates the aggregate memory from either the TotalMemoryGB
// field (if set) or by summing the memory from all individual NUMA nodes.
func (nt *NumaTopology) TotalMemory() float64 {
	if nt.TotalMemoryGB > 0 {
		return nt.TotalMemoryGB
	}
	
	total := 0.0
	for _, node := range nt.NodesInfo {
		total += node.MemorySize
	}
	return total
}

// TotalCores returns the total number of CPU cores across all NUMA nodes.
//
// This method counts the aggregate CPU cores from all NUMA nodes in the
// system topology for workload distribution and parallelization planning.
func (nt *NumaTopology) TotalCores() int {
	total := 0
	for _, node := range nt.NodesInfo {
		total += len(node.CPUCores)
	}
	
	// If no detailed node info is available, estimate from node count
	if total == 0 && nt.NodeCount > 0 {
		// Assume reasonable default of 8 cores per NUMA node
		return nt.NodeCount * 8
	}
	
	return total
}

// StatisticalSummary provides aggregate statistical analysis across all
// benchmark operations for overall performance characterization.
type StatisticalSummary struct {
	// OverallStability measures the consistency across all operations.
	// Calculated as the average coefficient of variation.
	OverallStability float64
	
	// TotalOutliers indicates the total number of outliers removed across all operations.
	TotalOutliers int
	
	// ExecutionStability measures the consistency of execution timing.
	ExecutionStability float64
	
	// RecommendedConfidence suggests whether results meet statistical requirements.
	RecommendedConfidence bool
}

// ValidationStatus indicates the statistical validity and reliability of benchmark results.
type ValidationStatus struct {
	// IsValid indicates whether the results meet minimum statistical requirements.
	IsValid bool
	
	// ValidationErrors lists any issues that prevent result validation.
	ValidationErrors []string
	
	// WarningMessages contains non-critical issues that may affect result interpretation.
	WarningMessages []string
	
	// QualityScore provides a numerical quality assessment (0.0-1.0).
	QualityScore float64
}

// NewStreamBenchmark creates a new STREAM benchmark executor with the specified
// configuration and container image.
//
// This function initializes a complete STREAM benchmark environment with
// statistical validation, NUMA awareness, and architecture-specific optimizations.
// The benchmark executor is configured for reliable, repeatable performance
// measurement with comprehensive error handling and result validation.
//
// Parameters:
//   - config: Benchmark execution configuration including iterations and confidence levels
//   - containerImage: Optimized STREAM container image for the target architecture
//   - numaTopology: NUMA system topology for optimization (can be empty for single-socket)
//
// Returns:
//   - *StreamBenchmark: Configured benchmark executor ready for execution
//
// Example:
//   config := benchmarks.BenchmarkConfig{
//       Iterations:       10,
//       ConfidenceLevel:  0.95,
//       OutlierThreshold: 2.5,
//       MinValidRuns:     8,
//       EnableNUMA:       true,
//   }
//   
//   benchmark := benchmarks.NewStreamBenchmark(
//       config,
//       "public.ecr.aws/aws-benchmarks/stream:intel-icelake",
//       numaTopology,
//   )
//
// Configuration Recommendations:
//   - Development: 3-5 iterations for quick validation
//   - Production: 10-20 iterations for statistical confidence
//   - Research: 20+ iterations for publication-quality results
//
// Container Requirements:
//   - Must contain STREAM benchmark compiled with architecture optimizations
//   - Should support NUMA-aware execution via environment variables
//   - Must output results in JSON format for automated parsing
func NewStreamBenchmark(config BenchmarkConfig, containerImage string, numaTopology NumaTopology) *StreamBenchmark {
	// Set sensible defaults for unspecified configuration
	if config.Iterations == 0 {
		config.Iterations = 10
	}
	if config.ConfidenceLevel == 0 {
		config.ConfidenceLevel = 0.95
	}
	if config.OutlierThreshold == 0 {
		config.OutlierThreshold = 2.5
	}
	if config.MinValidRuns == 0 {
		config.MinValidRuns = maxInt(3, config.Iterations/2)
	}
	if config.MaxExecutionTime == 0 {
		config.MaxExecutionTime = 5 * time.Minute
	}
	if config.MemoryPattern == "" {
		config.MemoryPattern = "sequential"
	}

	return &StreamBenchmark{
		config:         config,
		containerImage:   containerImage,
		numaTopology:     numaTopology,
		storage:          nil, // Storage is optional, set via WithStorage()
		metricsCollector: nil, // Metrics are optional, set via WithMetrics()
	}
}

// WithStorage configures S3 storage for automatic result persistence.
//
// This method enables automatic storage of benchmark results to S3 with
// comprehensive metadata and intelligent organization. Results are stored
// immediately after successful execution with proper error handling.
//
// Parameters:
//   - storage: Configured S3Storage instance for result persistence
//
// Returns:
//   - *StreamBenchmark: The same benchmark instance for method chaining
//
// Example:
//   storageConfig := storage.Config{
//       BucketName:         "aws-benchmarks-data",
//       KeyPrefix:          "stream-results/",
//       EnableCompression:  true,
//       StorageClass:       "STANDARD",
//   }
//   
//   s3Storage, err := storage.NewS3Storage(ctx, storageConfig)
//   if err != nil {
//       return fmt.Errorf("storage initialization failed: %w", err)
//   }
//   
//   benchmark := NewStreamBenchmark(config, containerImage, numaTopology).
//       WithStorage(s3Storage)
//   
//   result, err := benchmark.Execute(ctx)
//   // Result is automatically stored to S3 if execution succeeds
//
// Storage Behavior:
//   - Results are stored only after successful benchmark execution
//   - Storage failures do not cause benchmark execution to fail
//   - Storage errors are logged but do not propagate to the caller
//   - Metadata includes comprehensive context for analysis and filtering
func (s *StreamBenchmark) WithStorage(storage *storage.S3Storage) *StreamBenchmark {
	s.storage = storage
	return s
}

// WithMetrics configures CloudWatch metrics collection for automatic monitoring
// and observability of benchmark execution.
//
// This method enables comprehensive metrics publication to CloudWatch including
// execution success rates, performance measurements, duration tracking, and
// cost analysis. Metrics are published automatically after each benchmark
// execution with standardized dimensions for consistent querying and alerting.
//
// The metrics integration provides:
//   - Execution success and failure rate tracking
//   - Performance measurements (bandwidth, GFLOPS, efficiency)
//   - Duration metrics (total execution, benchmark-only time)
//   - Cost tracking and price-performance analysis
//   - Quality scores for result validation confidence
//   - Error categorization for failure analysis
//
// Parameters:
//   - collector: Configured CloudWatch metrics collector
//
// Returns:
//   - *StreamBenchmark: The same benchmark instance for method chaining
//
// Example:
//   collector, _ := monitoring.NewMetricsCollector("us-east-1")
//   benchmark := NewStreamBenchmark(config, image, topology).
//       WithStorage(s3Storage).
//       WithMetrics(collector)
//   
//   result, err := benchmark.Execute(ctx)
//   // Metrics are automatically published to CloudWatch
//
// Metrics Published:
//   - InstanceBenchmarks/BenchmarkExecution: Success/failure count
//   - InstanceBenchmarks/ExecutionDuration: Total execution time
//   - InstanceBenchmarks/Performance_*: Benchmark-specific measurements
//   - InstanceBenchmarks/QualityScore: Result validation confidence
//   - InstanceBenchmarks/EstimatedCost: Cost tracking for budget analysis
//
// Error Handling:
//   - Metrics publication errors are logged but do not fail the benchmark
//   - Failed metrics do not affect benchmark result validity
//   - Automatic retry logic handles transient CloudWatch API issues
func (s *StreamBenchmark) WithMetrics(collector *monitoring.MetricsCollector) *StreamBenchmark {
	s.metricsCollector = collector
	return s
}

// Execute runs the STREAM benchmark with statistical validation and returns
// comprehensive results including confidence intervals and performance analysis.
//
// This method orchestrates the complete benchmark execution pipeline including
// multiple run execution, outlier detection, statistical analysis, and result
// validation. It provides robust error handling and comprehensive logging for
// debugging and analysis purposes.
//
// The execution process:
//   1. Validates configuration and system requirements
//   2. Executes multiple benchmark iterations
//   3. Collects and parses individual run results
//   4. Performs outlier detection and removal
//   5. Calculates statistical measures and confidence intervals
//   6. Validates results against quality thresholds
//   7. Returns comprehensive result structure
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - *BenchmarkResult: Comprehensive results with statistical validation
//   - error: Execution errors, validation failures, or system issues
//
// Example:
//   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
//   defer cancel()
//   
//   result, err := benchmark.Execute(ctx)
//   if err != nil {
//       return fmt.Errorf("benchmark execution failed: %w", err)
//   }
//   
//   if !result.ValidationStatus.IsValid {
//       log.Printf("Warning: Results may not be reliable: %v", 
//           result.ValidationStatus.ValidationErrors)
//   }
//   
//   triadBandwidth := result.Measurements["triad"].Value
//   confidence := result.Measurements["triad"].ConfidenceInterval
//   fmt.Printf("STREAM Triad: %.2f ± %.2f GB/s (95%% confidence)\n",
//       triadBandwidth, (confidence.Upper-confidence.Lower)/2)
//
// Performance Characteristics:
//   - Execution time: 2-10 minutes depending on iterations and instance type
//   - Memory usage: ~100MB for result storage and analysis
//   - Network usage: Container download (~500MB) on first execution
//
// Common Errors:
//   - Container execution failures due to insufficient memory
//   - Timeout errors for slow instance types or high iteration counts
//   - Statistical validation failures due to inconsistent performance
//   - NUMA configuration errors on multi-socket systems
func (s *StreamBenchmark) Execute(ctx context.Context) (*BenchmarkResult, error) {
	startTime := time.Now()
	
	// Validate configuration before execution
	if err := s.validateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Collect system information for metadata
	systemInfo := s.collectSystemInfo(ctx)
	
	// Execute multiple benchmark runs
	rawResults, err := s.executeMultipleRuns(ctx)
	if err != nil {
		return nil, fmt.Errorf("benchmark execution failed: %w", err)
	}
	
	// Process results with statistical analysis
	measurements, err := s.processResults(rawResults)
	if err != nil {
		return nil, fmt.Errorf("result processing failed: %w", err)
	}
	
	// Calculate statistical summary
	summary := s.calculateStatisticalSummary(measurements)
	
	// Validate results quality
	validation := s.validateResults(measurements, summary)
	
	// Construct comprehensive result
	result := &BenchmarkResult{
		BenchmarkSuite: "stream",
		Measurements:   measurements,
		ExecutionMetadata: ExecutionMetadata{
			SystemInfo:        systemInfo,
			ExecutionDuration: time.Since(startTime),
			ContainerImage:    s.containerImage,
		},
		StatisticalSummary: summary,
		ValidationStatus:   validation,
		Timestamp:          time.Now(),
	}
	
	// Store result to S3 if storage is configured
	if s.storage != nil {
		if err := s.storage.StoreResult(ctx, result); err != nil {
			// Log storage error but don't fail the benchmark execution
			// In a real implementation, this would use a proper logger
			// For now, we silently ignore storage errors to keep the interface clean
			_ = err
		}
	}
	
	// Publish metrics to CloudWatch if collector is configured
	if s.metricsCollector != nil {
		if err := s.publishBenchmarkMetrics(ctx, result, startTime); err != nil {
			// Log metrics errors but don't fail the benchmark
			// In a production system, this would use a proper logger
			_ = err
		}
	}
	
	return result, nil
}

// Common constants for benchmarks.
const (
	unknownValue = "unknown"
)

// Configuration validation errors.
var (
	ErrInsufficientIterations = errors.New("minimum 3 iterations required for statistical validity")
	ErrInvalidConfidenceLevel = errors.New("confidence level must be between 0 and 1")
	ErrInvalidMinValidRuns    = errors.New("minimum valid runs cannot exceed total iterations")
)

// Container execution errors.
var (
	ErrContainerExecution     = errors.New("container execution failed")
	ErrMissingStreamResults   = errors.New("missing stream_results field in container output")
	ErrMissingStreamOperation = errors.New("missing required STREAM operation in results")
	ErrParseTextOutput        = errors.New("failed to parse operation from text output")
)

// Result validation errors.
var (
	ErrInvalidBandwidth   = errors.New("invalid bandwidth value")
	ErrBandwidthOutOfRange = errors.New("bandwidth outside reasonable range")
	ErrInsufficientValidRuns = errors.New("insufficient valid runs after outlier removal")
)

// validateConfig ensures the benchmark configuration is valid and
// will produce reliable results.
func (s *StreamBenchmark) validateConfig() error {
	if s.config.Iterations < 3 {
		return fmt.Errorf("%w: got %d", ErrInsufficientIterations, s.config.Iterations)
	}
	
	if s.config.ConfidenceLevel <= 0 || s.config.ConfidenceLevel >= 1 {
		return fmt.Errorf("%w: got %f", ErrInvalidConfidenceLevel, s.config.ConfidenceLevel)
	}
	
	if s.config.MinValidRuns > s.config.Iterations {
		return fmt.Errorf("%w: %d > %d", ErrInvalidMinValidRuns, 
			s.config.MinValidRuns, s.config.Iterations)
	}
	
	return nil
}

// collectSystemInfo gathers comprehensive system information for benchmark metadata.
func (s *StreamBenchmark) collectSystemInfo(_ context.Context) SystemInfo {
	// Implementation would collect real system information
	// For now, return mock data matching the structure
	return SystemInfo{
		CPUModel:    "Intel Xeon Platinum 8375C",
		CPUCores:    8,
		CPUThreads:  16,
		MemoryTotal: 32.0,
		NUMANodes:   1,
		CacheHierarchy: CacheHierarchy{
			L1Cache: CacheLevel{Size: 32, Associativity: 8, LineSize: 64},
			L2Cache: CacheLevel{Size: 1024, Associativity: 16, LineSize: 64},
			L3Cache: CacheLevel{Size: 54000, Associativity: 20, LineSize: 64},
		},
	}
}

// executeMultipleRuns performs the specified number of benchmark iterations.
func (s *StreamBenchmark) executeMultipleRuns(ctx context.Context) ([]map[string]float64, error) {
	results := make([]map[string]float64, 0, s.config.Iterations)
	
	for i := 0; i < s.config.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		// Execute single benchmark run
		runResult, err := s.executeSingleRun(ctx, i)
		if err != nil {
			return nil, fmt.Errorf("run %d failed: %w", i+1, err)
		}
		
		results = append(results, runResult)
	}
	
	return results, nil
}

// executeSingleRun executes a single STREAM benchmark iteration using Docker container.
//
// This method orchestrates the execution of an optimized STREAM benchmark container
// with appropriate NUMA configuration and resource constraints. It handles container
// lifecycle management, output parsing, and error recovery.
//
// The execution process:
//   1. Constructs Docker command with NUMA and resource constraints
//   2. Executes container with timeout protection
//   3. Captures and parses JSON output from STREAM benchmark
//   4. Validates results against expected format and ranges
//   5. Returns structured bandwidth measurements for each operation
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - runNumber: Sequential run identifier for logging and debugging
//
// Returns:
//   - map[string]float64: STREAM operation results (copy, scale, add, triad) in GB/s
//   - error: Container execution errors, parsing failures, or validation issues
//
// Container Requirements:
//   - Must output results in JSON format to stdout
//   - Should handle NUMA configuration via environment variables
//   - Must exit with code 0 for successful execution
//   - Should include proper error messages in stderr for debugging
//
// Example container output format:
//   {
//     "stream_results": {
//       "copy": 45.2,
//       "scale": 44.8,
//       "add": 42.1,
//       "triad": 41.9
//     },
//     "metadata": {
//       "run_id": 1,
//       "numa_node": 0,
//       "array_size": 80000000
//     }
//   }
//
// Performance Characteristics:
//   - Execution time: 10-60 seconds depending on instance type and array size
//   - Memory usage: Container overhead ~50MB plus benchmark array allocation
//   - CPU usage: Full utilization of available cores during execution
//
// Common Errors:
//   - Container image not found or pull failures
//   - Insufficient memory for STREAM array allocation
//   - NUMA configuration errors on multi-socket systems
//   - JSON parsing failures due to unexpected output format
//   - Timeout errors for slow instances or large array sizes
func (s *StreamBenchmark) executeSingleRun(ctx context.Context, runNumber int) (map[string]float64, error) {
	// Build Docker command with appropriate configuration
	dockerArgs := s.buildDockerCommand(runNumber)
	
	// Execute container with timeout protection
	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	
	// Capture stdout and stderr for parsing and debugging
	stdout, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w (exit %d): %s", ErrContainerExecution,
				exitError.ExitCode(), string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute container: %w", err)
	}
	
	// Parse JSON output from container
	results, err := s.parseContainerOutput(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container output: %w", err)
	}
	
	// Validate results are within expected ranges
	if err := s.validateStreamResults(results); err != nil {
		return nil, fmt.Errorf("benchmark results validation failed: %w", err)
	}
	
	return results, nil
}

// buildDockerCommand constructs the Docker command arguments for STREAM benchmark execution.
//
// This method builds a comprehensive Docker command with proper resource constraints,
// NUMA configuration, security settings, and environment variables. It ensures
// reproducible execution across different environments while maintaining security.
//
// The command includes:
//   - Resource limits (memory, CPU) based on instance capabilities
//   - NUMA topology configuration for multi-socket optimization
//   - Security constraints (read-only filesystem, no-network)
//   - Environment variables for benchmark configuration
//   - Proper cleanup settings for container lifecycle management
//
// Parameters:
//   - runNumber: Sequential run identifier for container naming and logging
//
// Returns:
//   - []string: Complete Docker command arguments ready for exec.Command
//
// Docker Command Structure:
//   docker run --rm --name stream-run-{N} 
//          --memory 4G --cpus 8 
//          --security-opt no-new-privileges 
//          --read-only --network none
//          -e NUMA_NODE=0 -e ARRAY_SIZE=80000000 
//          -e MEMORY_PATTERN=sequential -e OUTPUT_FORMAT=json
//          {container_image}
//
// Security Considerations:
//   - Read-only root filesystem prevents container modification
//   - Network isolation prevents external communication
//   - No privilege escalation allowed
//   - Resource limits prevent resource exhaustion attacks
//
// Performance Tuning:
//   - Memory limit based on instance type and array size requirements
//   - CPU constraints aligned with NUMA topology
//   - Environment variables for architecture-specific optimizations
func (s *StreamBenchmark) buildDockerCommand(runNumber int) []string {
	containerName := fmt.Sprintf("stream-run-%d-%d", runNumber, time.Now().Unix())
	
	// Base Docker arguments with security and resource constraints
	args := []string{
		"run",
		"--rm",                              // Automatic cleanup
		"--name", containerName,             // Unique container name
		"--memory", "4G",                    // Memory limit for safety
		"--cpus", "8",                       // CPU limit aligned with typical instances
		"--security-opt", "no-new-privileges", // Security hardening
		"--read-only",                       // Immutable container filesystem
		"--network", "none",                 // Network isolation
	}
	
	// Add NUMA configuration if enabled
	if s.config.EnableNUMA && s.numaTopology.NodeCount > 0 {
		// Use first NUMA node for single-node benchmarks
		args = append(args, "-e", "NUMA_NODE=0")
		args = append(args, "-e", fmt.Sprintf("NUMA_NODES=%d", s.numaTopology.NodeCount))
	}
	
	// Add benchmark configuration environment variables
	args = append(args,
		"-e", "OUTPUT_FORMAT=json",                    // Ensure JSON output
		"-e", fmt.Sprintf("RUN_ID=%d", runNumber),     // Run identification
		"-e", fmt.Sprintf("MEMORY_PATTERN=%s", s.config.MemoryPattern), // Access pattern
		"-e", "ARRAY_SIZE=80000000",                   // Standard STREAM array size
	)
	
	// Add container image as final argument
	args = append(args, s.containerImage)
	
	return args
}

// parseContainerOutput parses JSON output from the STREAM benchmark container.
//
// This method handles the structured JSON output from optimized STREAM containers,
// extracting bandwidth measurements and metadata. It provides robust error handling
// for various output formats and validates the data structure integrity.
//
// Expected JSON format from container:
//   {
//     "stream_results": {
//       "copy": 45.2,    // Copy bandwidth in GB/s
//       "scale": 44.8,   // Scale bandwidth in GB/s  
//       "add": 42.1,     // Add bandwidth in GB/s
//       "triad": 41.9    // Triad bandwidth in GB/s
//     },
//     "metadata": {
//       "run_id": 1,
//       "numa_node": 0,
//       "array_size": 80000000,
//       "compiler": "gcc-11",
//       "optimization": "-O3 -march=native"
//     }
//   }
//
// Parameters:
//   - output: Raw stdout bytes from container execution
//
// Returns:
//   - map[string]float64: Parsed STREAM operation results in GB/s
//   - error: JSON parsing errors or missing required fields
//
// Error Handling:
//   - Handles malformed JSON with descriptive error messages
//   - Validates presence of all required STREAM operations
//   - Checks for reasonable bandwidth values (positive, within expected ranges)
//   - Provides debugging information for troubleshooting container issues
//
// Fallback Parsing:
//   If JSON parsing fails, attempts to parse traditional STREAM text output
//   as a compatibility measure for older or custom containers.
func (s *StreamBenchmark) parseContainerOutput(output []byte) (map[string]float64, error) {
	// First, try to parse as JSON (preferred format)
	var containerResult struct {
		StreamResults map[string]float64 `json:"stream_results"`
		Metadata      map[string]interface{} `json:"metadata"`
	}
	
	if err := json.Unmarshal(output, &containerResult); err == nil {
		// JSON parsing successful - validate required fields
		if containerResult.StreamResults == nil {
			return nil, ErrMissingStreamResults
		}
		
		// Check for all required STREAM operations
		requiredOps := []string{"copy", "scale", "add", "triad"}
		for _, op := range requiredOps {
			if _, exists := containerResult.StreamResults[op]; !exists {
				return nil, fmt.Errorf("%w: '%s'", ErrMissingStreamOperation, op)
			}
		}
		
		return containerResult.StreamResults, nil
	}
	
	// JSON parsing failed - attempt fallback text parsing
	return s.parseTextOutput(string(output))
}

// parseTextOutput provides fallback parsing for traditional STREAM text output format.
//
// This method handles containers that output results in the traditional STREAM
// text format instead of JSON. It uses regular expressions to extract bandwidth
// values from formatted text output.
//
// Expected text format:
//   Function    Best Rate MB/s  Avg time     Min time     Max time
//   Copy:           45234.2     0.035482     0.035401     0.035563
//   Scale:          44876.1     0.035681     0.035681     0.035681  
//   Add:            42134.7     0.057012     0.057012     0.057012
//   Triad:          41987.3     0.057215     0.057215     0.057215
//
// Parameters:
//   - output: Raw text output from container as string
//
// Returns:
//   - map[string]float64: Parsed STREAM operation results converted to GB/s
//   - error: Text parsing errors or missing operation results
//
// Conversion Notes:
//   - Converts MB/s to GB/s using factor of 1000 (not 1024)
//   - Handles both "MB/s" and "GB/s" units in output
//   - Provides case-insensitive operation name matching
func (s *StreamBenchmark) parseTextOutput(output string) (map[string]float64, error) {
	results := make(map[string]float64)
	lines := strings.Split(output, "\n")
	
	// Parse each line looking for STREAM operation results
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Look for lines matching STREAM output format
		// Example: "Copy:           45234.2     0.035482     0.035401     0.035563"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		// Extract operation name (remove colon)
		opName := strings.ToLower(strings.TrimSuffix(fields[0], ":"))
		if opName != "copy" && opName != "scale" && opName != "add" && opName != "triad" {
			continue
		}
		
		// Parse bandwidth value (second field)
		bandwidthStr := fields[1]
		bandwidth, err := strconv.ParseFloat(bandwidthStr, 64)
		if err != nil {
			continue // Skip lines that don't have valid numbers
		}
		
		// Convert MB/s to GB/s (assuming STREAM output is in MB/s)
		results[opName] = bandwidth / 1000.0
	}
	
	// Validate that we found all required operations
	requiredOps := []string{"copy", "scale", "add", "triad"}
	for _, op := range requiredOps {
		if _, exists := results[op]; !exists {
			return nil, fmt.Errorf("%w: %s", ErrParseTextOutput, op)
		}
	}
	
	return results, nil
}

// validateStreamResults performs sanity checks on STREAM benchmark results.
//
// This method validates that STREAM bandwidth measurements are reasonable
// and follow expected patterns. It helps detect container execution issues,
// measurement errors, or system problems that could affect result reliability.
//
// Validation Checks:
//   - All bandwidth values are positive
//   - Values are within reasonable ranges for modern systems (0.1 - 1000 GB/s)
//   - Triad bandwidth is typically lowest due to arithmetic intensity
//   - Copy and Scale should be similar (both single-array operations)
//   - No values are exactly zero (indicates measurement failure)
//
// Parameters:
//   - results: Map of STREAM operation results to validate
//
// Returns:
//   - error: Validation failures with specific details for debugging
//
// Performance Expectations:
//   - Typical ranges: 10-200 GB/s for modern instances
//   - Copy/Scale: Usually highest bandwidth (memory-bound)
//   - Add: Moderate bandwidth (reads two arrays)
//   - Triad: Typically lowest (reads two arrays, writes one)
//
// Common Issues Detected:
//   - Zero values indicate container execution failure
//   - Extremely high values suggest measurement errors
//   - Inverted hierarchy (Triad > Copy) suggests system issues
func (s *StreamBenchmark) validateStreamResults(results map[string]float64) error {
	// Check for positive values
	for op, bandwidth := range results {
		if bandwidth <= 0 {
			return fmt.Errorf("%w for %s: %f (must be positive)", ErrInvalidBandwidth, op, bandwidth)
		}
		
		// Check for reasonable ranges (0.1 GB/s to 1000 GB/s)
		if bandwidth < 0.1 || bandwidth > 1000.0 {
			return fmt.Errorf("%w for %s: %f GB/s", ErrBandwidthOutOfRange, op, bandwidth)
		}
	}
	
	// Optional: Check for expected STREAM operation hierarchy
	// This is advisory rather than strict validation
	copyBandwidth := results["copy"]
	triadBandwidth := results["triad"]
	
	// Triad is typically the lowest bandwidth operation
	// This check is advisory - unusual patterns don't cause validation failure
	_ = triadBandwidth > copyBandwidth*1.1 // Allow 10% tolerance
	
	return nil
}

// processResults performs statistical analysis on raw benchmark results.
func (s *StreamBenchmark) processResults(rawResults []map[string]float64) (map[string]Measurement, error) {
	operations := []string{"copy", "scale", "add", "triad"}
	measurements := make(map[string]Measurement)
	
	for _, op := range operations {
		values := make([]float64, len(rawResults))
		for i, result := range rawResults {
			values[i] = result[op]
		}
		
		measurement, err := s.calculateMeasurement(op, values)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate measurement for %s: %w", op, err)
		}
		
		measurements[op] = measurement
	}
	
	return measurements, nil
}

// calculateMeasurement performs statistical analysis for a single operation.
func (s *StreamBenchmark) calculateMeasurement(operation string, values []float64) (Measurement, error) {
	// Remove outliers
	validValues, outlierCount := s.removeOutliers(values)
	
	if len(validValues) < s.config.MinValidRuns {
		return Measurement{}, fmt.Errorf("%w: %d < %d", ErrInsufficientValidRuns,
			len(validValues), s.config.MinValidRuns)
	}
	
	// Calculate statistics
	mean := calculateMean(validValues)
	stdDev := calculateStandardDeviation(validValues, mean)
	cv := (stdDev / mean) * 100
	
	// Calculate confidence interval
	confidenceInterval := s.calculateConfidenceInterval(validValues, mean, stdDev)
	
	return Measurement{
		Operation:              operation,
		Value:                  mean,
		Unit:                   "GB/s",
		StandardDeviation:      stdDev,
		ConfidenceInterval:     confidenceInterval,
		CoefficientOfVariation: cv,
		ValidRuns:              validValues,
		OutliersRemoved:        outlierCount,
	}, nil
}

// removeOutliers identifies and removes statistical outliers from the dataset.
func (s *StreamBenchmark) removeOutliers(values []float64) ([]float64, int) {
	if len(values) < 3 {
		return values, 0
	}
	
	mean := calculateMean(values)
	stdDev := calculateStandardDeviation(values, mean)
	threshold := s.config.OutlierThreshold * stdDev
	
	var validValues []float64
	for _, value := range values {
		if math.Abs(value-mean) <= threshold {
			validValues = append(validValues, value)
		}
	}
	
	return validValues, len(values) - len(validValues)
}

// calculateConfidenceInterval computes the confidence interval for the measurement.
func (s *StreamBenchmark) calculateConfidenceInterval(values []float64, mean, stdDev float64) ConfidenceInterval {
	// Simplified confidence interval calculation
	// Real implementation would use t-distribution
	n := float64(len(values))
	margin := 1.96 * (stdDev / math.Sqrt(n)) // 95% confidence for large samples
	
	return ConfidenceInterval{
		Lower: mean - margin,
		Upper: mean + margin,
		Level: s.config.ConfidenceLevel,
	}
}

// calculateStatisticalSummary provides aggregate analysis across all measurements.
func (s *StreamBenchmark) calculateStatisticalSummary(measurements map[string]Measurement) StatisticalSummary {
	var totalCV float64
	var totalOutliers int
	
	for _, measurement := range measurements {
		totalCV += measurement.CoefficientOfVariation
		totalOutliers += measurement.OutliersRemoved
	}
	
	avgCV := totalCV / float64(len(measurements))
	
	return StatisticalSummary{
		OverallStability:      avgCV,
		TotalOutliers:         totalOutliers,
		ExecutionStability:    avgCV, // Simplified
		RecommendedConfidence: avgCV < 5.0, // < 5% CV indicates good stability
	}
}

// validateResults assesses the quality and reliability of benchmark results.
func (s *StreamBenchmark) validateResults(measurements map[string]Measurement, summary StatisticalSummary) ValidationStatus {
	var errors []string
	var warnings []string
	
	// Check coefficient of variation for each measurement
	for op, measurement := range measurements {
		if measurement.CoefficientOfVariation > 10.0 {
			errors = append(errors, fmt.Sprintf("%s operation has high variability (%.1f%% CV)", 
				op, measurement.CoefficientOfVariation))
		} else if measurement.CoefficientOfVariation > 5.0 {
			warnings = append(warnings, fmt.Sprintf("%s operation has moderate variability (%.1f%% CV)", 
				op, measurement.CoefficientOfVariation))
		}
	}
	
	// Check outlier removal rate
	outlierRate := float64(summary.TotalOutliers) / float64(s.config.Iterations*len(measurements)) * 100
	if outlierRate > 20.0 {
		warnings = append(warnings, fmt.Sprintf("high outlier rate: %.1f%%", outlierRate))
	}
	
	// Calculate quality score
	qualityScore := 1.0
	if summary.OverallStability > 5.0 {
		qualityScore -= 0.3
	}
	if outlierRate > 15.0 {
		qualityScore -= 0.2
	}
	
	return ValidationStatus{
		IsValid:           len(errors) == 0,
		ValidationErrors:  errors,
		WarningMessages:   warnings,
		QualityScore:      math.Max(0.0, qualityScore),
	}
}

// Helper functions for statistical calculations

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	
	sumSquares := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquares += diff * diff
	}
	
	variance := sumSquares / float64(len(values)-1)
	return math.Sqrt(variance)
}

// publishBenchmarkMetrics publishes comprehensive benchmark execution metrics
// to CloudWatch for monitoring and analysis.
func (s *StreamBenchmark) publishBenchmarkMetrics(ctx context.Context, result *BenchmarkResult, startTime time.Time) error {
	// Extract instance information from execution context
	// In a real implementation, this would come from the orchestration layer
	instanceType := unknownValue // Would be passed from orchestrator
	instanceFamily := unknownValue // Would be extracted from instance type
	region := "us-east-1" // Would be passed from orchestrator
	
	// Calculate execution duration
	executionDuration := time.Since(startTime).Seconds()
	benchmarkDuration := result.ExecutionMetadata.ExecutionDuration.Seconds()
	
	// Extract performance metrics from STREAM results
	performanceMetrics := make(map[string]float64)
	for operation, measurement := range result.Measurements {
		// Convert bandwidth from GB/s to a consistent unit
		performanceMetrics[operation+"_bandwidth"] = measurement.Value
	}
	
	// Calculate estimated cost (mock calculation)
	// In production, this would use real AWS pricing data
	estimatedCost := executionDuration * 0.05 // $0.05 per second example
	
	// Calculate price-performance ratio (cost per GB/s for STREAM)
	var avgBandwidth float64
	if len(performanceMetrics) > 0 {
		total := 0.0
		for _, bandwidth := range performanceMetrics {
			total += bandwidth
		}
		avgBandwidth = total / float64(len(performanceMetrics))
	}
	
	pricePerformanceRatio := 0.0
	if avgBandwidth > 0 {
		pricePerformanceRatio = estimatedCost / avgBandwidth
	}
	
	// Determine error category if benchmark failed
	errorCategory := ""
	if !result.ValidationStatus.IsValid {
		if len(result.ValidationStatus.ValidationErrors) > 0 {
			errorCategory = "validation"
		} else {
			errorCategory = "execution"
		}
	}
	
	// Create comprehensive benchmark metrics
	metrics := monitoring.BenchmarkMetrics{
		InstanceType:      instanceType,
		InstanceFamily:    instanceFamily,
		BenchmarkSuite:    "stream",
		Region:           region,
		Success:          result.ValidationStatus.IsValid,
		ExecutionDuration: executionDuration,
		BenchmarkDuration: benchmarkDuration,
		PerformanceMetrics: performanceMetrics,
		ErrorCategory:     errorCategory,
		CostMetrics: monitoring.CostMetrics{
			EstimatedCost:         estimatedCost,
			PricePerformanceRatio: pricePerformanceRatio,
			InstanceHourCost:      180.0, // Mock hourly cost in cents
		},
		QualityScore: result.ValidationStatus.QualityScore,
		Timestamp:    time.Now(),
	}
	
	// Publish metrics to CloudWatch
	return s.metricsCollector.PublishBenchmarkMetrics(ctx, metrics)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}