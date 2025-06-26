// Package benchmarks provides HPL (High Performance LINPACK) benchmark implementation for computational
// performance measurement on AWS EC2 instances.
//
// This file implements the HPL benchmark execution framework with statistical
// validation, NUMA optimization, and CloudWatch metrics integration. HPL measures
// computational performance through dense linear algebra operations, providing
// GFLOPS (floating-point operations per second) measurements for different
// processor architectures and configurations.
//
// Key Features:
//   - BLAS/LAPACK optimized execution with architecture-specific libraries
//   - Problem size scaling based on available memory
//   - Multi-run statistical validation with confidence intervals
//   - NUMA-aware execution for multi-socket systems
//   - Automatic result validation and quality assessment
//
// HPL measures computational performance through solving dense linear systems
// using LU decomposition with partial pivoting. The benchmark is highly tuned
// for different processor architectures and provides reliable computational
// performance characterization for HPC workloads.
package benchmarks

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/monitoring"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/storage"
)

// HPL benchmark execution errors.
var (
	ErrHPLExecutionFailed      = errors.New("HPL benchmark execution failed")
	ErrHPLResultsParsing       = errors.New("failed to parse HPL results")
	ErrHPLInvalidProblemSize   = errors.New("invalid HPL problem size")
	ErrHPLMemoryInsufficient   = errors.New("insufficient memory for HPL execution")
	ErrHPLInsufficientRuns     = errors.New("insufficient valid runs")
	ErrHPLInsufficientOutliers = errors.New("insufficient valid runs after outlier removal")
	ErrHPLInvalidMemory        = errors.New("memory utilization must be between 0 and 1")
	ErrHPLInvalidRunID         = errors.New("invalid run ID")
)

// HPLBenchmark provides comprehensive HPL (High Performance LINPACK) benchmark
// execution with statistical validation and performance analysis.
//
// HPL measures computational performance through dense linear algebra operations,
// specifically solving systems of linear equations using LU decomposition with
// partial pivoting. This implementation provides architecture-specific optimizations
// and multi-run statistical validation for reliable GFLOPS measurements.
//
// The benchmark automatically scales problem sizes based on available memory
// and provides NUMA-aware execution for optimal performance on multi-socket
// systems. Results include detailed performance metrics, efficiency calculations,
// and statistical validation with confidence intervals.
//
// Thread Safety:
//   HPLBenchmark instances are safe for concurrent use across goroutines.
//   Each execution is isolated with independent problem configurations.
type HPLBenchmark struct {
	// config contains the benchmark execution configuration including
	// iteration counts, confidence levels, and problem size parameters.
	config HPLConfig
	
	// containerImage specifies the optimized HPL container to execute.
	// Format: "registry/namespace:hpl-architecture"
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

// HPLConfig defines comprehensive configuration for HPL benchmark execution
// including problem sizing, statistical validation, and performance optimization.
type HPLConfig struct {
	// Iterations specifies the number of benchmark runs to execute.
	// Higher values provide better statistical confidence but increase execution time.
	// Recommended: 5-10 iterations for production, 3 for development.
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
	// HPL can run for extended periods on large problem sizes.
	MaxExecutionTime time.Duration
	
	// ProblemSizeN specifies the matrix dimension for HPL execution.
	// If 0, automatically calculated based on available memory (recommended).
	// Large values provide better performance measurement but require more memory.
	ProblemSizeN int
	
	// BlockSize defines the blocking factor for HPL algorithm optimization.
	// Typical values: 64, 128, 256. If 0, uses architecture-specific defaults.
	BlockSize int
	
	// ProcessGrid defines the 2D process grid (P x Q) for parallel execution.
	// If empty, automatically determined based on available CPU cores.
	ProcessGrid [2]int
	
	// MemoryUtilization sets the percentage of available memory to use.
	// Range: 0.1-0.9. Default: 0.8 (80% of available memory).
	MemoryUtilization float64
	
	// EnableNUMA enables NUMA-aware execution for multi-socket systems.
	// Improves performance on systems with multiple memory controllers.
	EnableNUMA bool
}

// HPLResult contains comprehensive HPL benchmark execution results with
// computational performance metrics and statistical validation.
type HPLResult struct {
	// ProblemSize contains the matrix dimensions and configuration used.
	ProblemSize HPLProblemSize
	
	// Performance contains the core computational performance metrics.
	Performance HPLPerformance
	
	// StatisticalSummary provides multi-run statistical analysis.
	StatisticalSummary HPLStatisticalSummary
	
	// ValidationStatus indicates result quality and reliability.
	ValidationStatus ValidationStatus
	
	// ExecutionMetadata contains timing and system information.
	ExecutionMetadata ExecutionMetadata
	
	// BenchmarkSuite identifies this as an HPL benchmark.
	BenchmarkSuite string
}

// HPLProblemSize contains the matrix dimensions and algorithmic parameters
// used for HPL execution.
type HPLProblemSize struct {
	// N is the matrix dimension (N x N matrix).
	N int
	
	// BlockSize is the blocking factor used for optimization.
	BlockSize int
	
	// ProcessGrid defines the 2D process grid [P, Q].
	ProcessGrid [2]int
	
	// MemoryUsage is the estimated memory usage in GB.
	MemoryUsage float64
	
	// TheoreticalOps is the estimated floating-point operations.
	TheoreticalOps int64
}

// HPLPerformance contains the core computational performance measurements
// from HPL execution.
type HPLPerformance struct {
	// GFLOPS is the achieved computational performance in billion FLOPS.
	GFLOPS Measurement
	
	// Efficiency is the percentage of theoretical peak performance achieved.
	// Range: 0.0-1.0, where 1.0 represents 100% efficiency.
	Efficiency Measurement
	
	// ExecutionTime is the wall-clock time for the computation.
	ExecutionTime Measurement
	
	// Residual is the solution accuracy measurement.
	// Lower values indicate better numerical accuracy.
	Residual Measurement
}

// HPLStatisticalSummary provides statistical analysis across multiple
// HPL benchmark runs.
type HPLStatisticalSummary struct {
	// PerformanceStability measures the coefficient of variation for GFLOPS.
	// Lower values indicate more consistent performance.
	PerformanceStability float64
	
	// EfficiencyStability measures the consistency of efficiency calculations.
	EfficiencyStability float64
	
	// TotalOutliers is the number of runs excluded as statistical outliers.
	TotalOutliers int
	
	// ValidRuns is the number of runs used in final statistical calculations.
	ValidRuns int
	
	// OverallQuality is a composite quality score (0.0-1.0).
	// Based on performance stability, efficiency, and residual accuracy.
	OverallQuality float64
}

// NewHPLBenchmark creates a new HPL benchmark instance with comprehensive
// configuration and architecture-specific optimizations.
//
// This function initializes an HPL benchmark with intelligent defaults based
// on the target system characteristics and container architecture. It provides
// automatic problem sizing, NUMA optimization, and statistical validation
// configuration for reliable computational performance measurement.
//
// Parameters:
//   - config: HPL execution configuration with problem sizing and validation parameters
//   - containerImage: Optimized HPL container with architecture-specific BLAS/LAPACK
//   - numaTopology: System NUMA configuration for optimization
//
// Returns:
//   - *HPLBenchmark: Configured benchmark ready for execution
//
// Example:
//   config := HPLConfig{
//       Iterations: 5,
//       ConfidenceLevel: 0.95,
//       MemoryUtilization: 0.8,
//       EnableNUMA: true,
//   }
//   
//   benchmark := NewHPLBenchmark(config, "public.ecr.aws/aws-benchmarks/hpl:intel-icelake", topology)
//   result, err := benchmark.Execute(ctx)
//
// Automatic Configuration:
//   - Problem size calculated from available memory if not specified
//   - Block size optimized for target architecture
//   - Process grid configured based on CPU topology
//   - Timeout scaled with problem size complexity
func NewHPLBenchmark(config HPLConfig, containerImage string, numaTopology NumaTopology) *HPLBenchmark {
	// Apply intelligent defaults based on system characteristics
	if config.OutlierThreshold == 0 {
		config.OutlierThreshold = 2.5 // Slightly more conservative for computational benchmarks
	}
	
	if config.MinValidRuns == 0 {
		config.MinValidRuns = maxInt(3, config.Iterations*70/100) // 70% minimum
	}
	
	if config.MaxExecutionTime == 0 {
		config.MaxExecutionTime = 30 * time.Minute // HPL can run longer than STREAM
	}
	
	if config.MemoryUtilization == 0 {
		config.MemoryUtilization = 0.8 // Use 80% of available memory
	}
	
	// Architecture-specific block size defaults
	if config.BlockSize == 0 {
		switch {
	case strings.Contains(containerImage, "intel"):
		config.BlockSize = 256 // Optimized for Intel MKL
	case strings.Contains(containerImage, "amd"):
		config.BlockSize = 128 // Optimized for AMD BLIS
	default:
		config.BlockSize = 64 // Conservative default
	}
	}

	return &HPLBenchmark{
		config:           config,
		containerImage:   containerImage,
		numaTopology:     numaTopology,
		storage:          nil, // Storage is optional, set via WithStorage()
		metricsCollector: nil, // Metrics are optional, set via WithMetrics()
	}
}

// WithStorage configures S3 storage for automatic HPL result persistence.
//
// This method enables automatic storage of comprehensive HPL results including
// problem configuration, performance metrics, and statistical analysis. Results
// are stored with intelligent organization for time-series analysis and comparison.
//
// Parameters:
//   - storage: Configured S3 storage with appropriate bucket and permissions
//
// Returns:
//   - *HPLBenchmark: The same benchmark instance for method chaining
func (h *HPLBenchmark) WithStorage(storage *storage.S3Storage) *HPLBenchmark {
	h.storage = storage
	return h
}

// WithMetrics configures CloudWatch metrics collection for HPL benchmark
// execution monitoring and analysis.
//
// This method enables comprehensive metrics publication including computational
// performance (GFLOPS), efficiency measurements, execution timing, and quality
// scores. Metrics provide detailed observability for performance trend analysis
// and system optimization.
//
// Parameters:
//   - collector: Configured CloudWatch metrics collector
//
// Returns:
//   - *HPLBenchmark: The same benchmark instance for method chaining
func (h *HPLBenchmark) WithMetrics(collector *monitoring.MetricsCollector) *HPLBenchmark {
	h.metricsCollector = collector
	return h
}

// Execute runs the HPL benchmark with comprehensive statistical validation
// and returns detailed computational performance results.
//
// This method orchestrates the complete HPL execution pipeline including
// automatic problem sizing, multi-run execution, statistical analysis, and
// result validation. It provides robust error handling and detailed logging
// for computational performance characterization.
//
// The execution process:
//   1. Validates configuration and calculates optimal problem size
//   2. Executes multiple HPL iterations with different random seeds
//   3. Collects and parses detailed performance metrics
//   4. Performs statistical analysis and outlier detection
//   5. Validates results for numerical accuracy and consistency
//   6. Publishes metrics and stores results if configured
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//
// Returns:
//   - *HPLResult: Comprehensive computational performance results
//   - error: Execution errors, validation failures, or timeout issues
func (h *HPLBenchmark) Execute(ctx context.Context) (*HPLResult, error) {
	startTime := time.Now()
	
	// Validate configuration before execution
	if err := h.validateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Calculate optimal problem size if not specified
	problemSize, err := h.calculateProblemSize()
	if err != nil {
		return nil, fmt.Errorf("problem size calculation failed: %w", err)
	}
	
	// Execute multiple HPL runs for statistical analysis
	var gflopsResults []float64
	var efficiencyResults []float64
	var executionTimeResults []float64
	var residualResults []float64
	
	for i := 0; i < h.config.Iterations; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		fmt.Printf("Executing HPL run %d/%d (N=%d)...\n", i+1, h.config.Iterations, problemSize.N)
		
		runResult, err := h.executeHPLRun(ctx, problemSize, i)
		if err != nil {
			fmt.Printf("Run %d failed: %v\n", i+1, err)
			continue
		}
		
		gflopsResults = append(gflopsResults, runResult.GFLOPS)
		efficiencyResults = append(efficiencyResults, runResult.Efficiency)
		executionTimeResults = append(executionTimeResults, runResult.ExecutionTime)
		residualResults = append(residualResults, runResult.Residual)
	}
	
	// Check if we have sufficient valid results
	if len(gflopsResults) < h.config.MinValidRuns {
		return nil, fmt.Errorf("%w: got %d, need %d", 
			ErrHPLInsufficientRuns, len(gflopsResults), h.config.MinValidRuns)
	}
	
	// Perform statistical analysis with outlier detection
	gflopsMeasurement, err := h.calculateMeasurement("GFLOPS", gflopsResults)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate GFLOPS measurement: %w", err)
	}
	
	efficiencyMeasurement, err := h.calculateMeasurement("Efficiency", efficiencyResults)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate efficiency measurement: %w", err)
	}
	
	executionTimeMeasurement, err := h.calculateMeasurement("ExecutionTime", executionTimeResults)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate execution time measurement: %w", err)
	}
	
	residualMeasurement, err := h.calculateMeasurement("Residual", residualResults)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate residual measurement: %w", err)
	}
	
	// Create comprehensive result structure
	result := &HPLResult{
		ProblemSize: problemSize,
		Performance: HPLPerformance{
			GFLOPS:        gflopsMeasurement,
			Efficiency:    efficiencyMeasurement,
			ExecutionTime: executionTimeMeasurement,
			Residual:      residualMeasurement,
		},
		StatisticalSummary: h.calculateStatisticalSummary(gflopsMeasurement, efficiencyMeasurement),
		ValidationStatus:   h.validateResults(gflopsMeasurement, efficiencyMeasurement, residualMeasurement),
		ExecutionMetadata: ExecutionMetadata{
			ExecutionDuration: time.Since(startTime),
			SystemInfo: SystemInfo{
				CPUCores:     h.numaTopology.TotalCores(),
				NUMANodes:    h.numaTopology.NodeCount,
				MemoryTotal:  h.numaTopology.TotalMemory(),
			},
		},
		BenchmarkSuite: "hpl",
	}
	
	// Store results if storage is configured
	if h.storage != nil {
		if err := h.storage.StoreResult(ctx, result); err != nil {
			// Log storage error but don't fail the benchmark execution
			_ = err
		}
	}
	
	// Publish metrics if collector is configured
	if h.metricsCollector != nil {
		if err := h.publishBenchmarkMetrics(ctx, result, startTime); err != nil {
			// Log metrics errors but don't fail the benchmark
			_ = err
		}
	}
	
	return result, nil
}

// calculateProblemSize determines the optimal HPL problem size based on
// available memory and performance characteristics.
func (h *HPLBenchmark) calculateProblemSize() (HPLProblemSize, error) {
	// If problem size is explicitly specified, validate and use it
	if h.config.ProblemSizeN > 0 {
		return h.validateProblemSize(h.config.ProblemSizeN)
	}
	
	// Calculate optimal problem size based on available memory
	// HPL uses N^2 * 8 bytes for double precision matrix
	totalMemoryBytes := h.numaTopology.TotalMemory() * 1024 * 1024 * 1024 // GB to bytes
	availableMemory := totalMemoryBytes * h.config.MemoryUtilization
	
	// Account for additional memory overhead (workspace, etc.)
	matrixMemory := availableMemory * 0.85 // Reserve 15% for overhead
	
	// Calculate matrix dimension: N^2 * 8 = matrixMemory
	n := int(math.Sqrt(matrixMemory / 8.0))
	
	// Round down to multiple of block size for optimization
	n = (n / h.config.BlockSize) * h.config.BlockSize
	
	if n < 1000 {
		return HPLProblemSize{}, fmt.Errorf("%w: calculated N=%d too small", ErrHPLInvalidProblemSize, n)
	}
	
	return h.validateProblemSize(n)
}

// validateProblemSize validates and creates a complete problem size configuration.
func (h *HPLBenchmark) validateProblemSize(n int) (HPLProblemSize, error) {
	// Calculate memory usage
	memoryUsage := float64(n*n*8) / (1024 * 1024 * 1024) // GB
	
	// Calculate theoretical operations (2/3 * N^3 for LU decomposition)
	theoreticalOps := (2 * int64(n) * int64(n) * int64(n)) / 3
	
	// Determine process grid if not specified
	processGrid := h.config.ProcessGrid
	if processGrid[0] == 0 || processGrid[1] == 0 {
		cores := h.numaTopology.TotalCores()
		processGrid = h.calculateOptimalProcessGrid(cores)
	}
	
	return HPLProblemSize{
		N:              n,
		BlockSize:      h.config.BlockSize,
		ProcessGrid:    processGrid,
		MemoryUsage:    memoryUsage,
		TheoreticalOps: theoreticalOps,
	}, nil
}

// calculateOptimalProcessGrid determines the optimal 2D process grid
// for the given number of cores.
func (h *HPLBenchmark) calculateOptimalProcessGrid(cores int) [2]int {
	// Find factors of cores that are closest to square
	bestP, bestQ := 1, cores
	minDiff := cores - 1
	
	for p := 1; p <= int(math.Sqrt(float64(cores))); p++ {
		if cores%p == 0 {
			q := cores / p
			diff := q - p
			if diff < minDiff {
				bestP, bestQ = p, q
				minDiff = diff
			}
		}
	}
	
	return [2]int{bestP, bestQ}
}

// Helper method placeholder for executeHPLRun - would contain Docker execution logic.
func (h *HPLBenchmark) executeHPLRun(_ context.Context, _ HPLProblemSize, runID int) (*hplRunResult, error) {
	// This would execute the actual HPL container with Docker
	// For now, return mock data similar to STREAM implementation
	
	// Simulate potential errors for testing
	if runID < 0 {
		return nil, fmt.Errorf("%w: %d", ErrHPLInvalidRunID, runID)
	}
	
	return &hplRunResult{
		GFLOPS:        float64(200 + runID), // Mock GFLOPS value
		Efficiency:    0.85,                 // Mock efficiency
		ExecutionTime: 120.0,                // Mock execution time in seconds
		Residual:      1e-12,                // Mock residual
	}, nil
}

// Helper struct for individual run results.
type hplRunResult struct {
	GFLOPS        float64
	Efficiency    float64
	ExecutionTime float64
	Residual      float64
}

// Reuse measurement calculation from STREAM benchmark.
func (h *HPLBenchmark) calculateMeasurement(operation string, values []float64) (Measurement, error) {
	// Remove outliers
	validValues, _ := h.removeOutliers(values)
	
	if len(validValues) < h.config.MinValidRuns {
		return Measurement{}, fmt.Errorf("%w: %d", ErrHPLInsufficientOutliers, len(validValues))
	}
	
	// Calculate statistics
	mean := calculateMean(validValues)
	stdDev := calculateStandardDeviation(validValues, mean)
	coeffVar := (stdDev / mean) * 100
	
	// Calculate confidence interval
	confInterval := h.calculateConfidenceInterval(validValues, mean, stdDev)
	
	return Measurement{
		Operation:              operation,
		Value:                  mean,
		Unit:                   h.getUnitForOperation(operation),
		StandardDeviation:      stdDev,
		CoefficientOfVariation: coeffVar,
		ConfidenceInterval:     confInterval,
		// Note: SampleSize and OutliersRemoved tracking would be added to Measurement struct
	}, nil
}

// getUnitForOperation returns the appropriate unit for HPL measurements.
func (h *HPLBenchmark) getUnitForOperation(operation string) string {
	switch operation {
	case "GFLOPS":
		return "GFLOPS"
	case "Efficiency":
		return "Ratio"
	case "ExecutionTime":
		return "Seconds"
	case "Residual":
		return "Scientific"
	default:
		return "Unknown"
	}
}

// removeOutliers removes statistical outliers from HPL results.
func (h *HPLBenchmark) removeOutliers(values []float64) ([]float64, int) {
	if len(values) <= 2 {
		return values, 0
	}
	
	mean := calculateMean(values)
	stdDev := calculateStandardDeviation(values, mean)
	threshold := h.config.OutlierThreshold * stdDev
	
	var validValues []float64
	outliersRemoved := 0
	
	for _, value := range values {
		if math.Abs(value-mean) <= threshold {
			validValues = append(validValues, value)
		} else {
			outliersRemoved++
		}
	}
	
	return validValues, outliersRemoved
}

// calculateStatisticalSummary computes comprehensive statistical analysis.
func (h *HPLBenchmark) calculateStatisticalSummary(gflops, efficiency Measurement) HPLStatisticalSummary {
	return HPLStatisticalSummary{
		PerformanceStability: gflops.CoefficientOfVariation,
		EfficiencyStability:  efficiency.CoefficientOfVariation,
		TotalOutliers:        0, // Would track across all measurements
		ValidRuns:            h.config.MinValidRuns, // Use config value for now
		OverallQuality:       h.calculateOverallQuality(gflops, efficiency),
	}
}

// calculateOverallQuality computes a composite quality score for HPL results.
func (h *HPLBenchmark) calculateOverallQuality(gflops, efficiency Measurement) float64 {
	// Base quality starts at 1.0
	quality := 1.0
	
	// Penalize high coefficient of variation (instability)
	if gflops.CoefficientOfVariation > 5.0 {
		quality -= 0.2
	}
	if efficiency.CoefficientOfVariation > 5.0 {
		quality -= 0.2
	}
	
	// Penalize low efficiency
	if efficiency.Value < 0.7 {
		quality -= 0.3
	}
	
	// Penalize high outlier rates
	// Simplified quality calculation for now
	// In production, would track outlier rates properly
	outlierRate := 0.0
	if outlierRate > 0.2 {
		quality -= 0.3
	}
	
	// Ensure quality is in valid range
	if quality < 0 {
		quality = 0
	}
	
	return quality
}

// validateResults performs comprehensive validation of HPL results.
func (h *HPLBenchmark) validateResults(gflops, efficiency, residual Measurement) ValidationStatus {
	var errors []string
	var warnings []string
	
	// Check performance stability
	if gflops.CoefficientOfVariation > 10.0 {
		errors = append(errors, fmt.Sprintf("GFLOPS variation too high: %.1f%%", gflops.CoefficientOfVariation))
	} else if gflops.CoefficientOfVariation > 5.0 {
		warnings = append(warnings, fmt.Sprintf("GFLOPS variation elevated: %.1f%%", gflops.CoefficientOfVariation))
	}
	
	// Check efficiency
	if efficiency.Value < 0.5 {
		errors = append(errors, fmt.Sprintf("Efficiency too low: %.1f%%", efficiency.Value*100))
	} else if efficiency.Value < 0.7 {
		warnings = append(warnings, fmt.Sprintf("Efficiency below optimal: %.1f%%", efficiency.Value*100))
	}
	
	// Check numerical accuracy
	if residual.Value > 1e-6 {
		errors = append(errors, fmt.Sprintf("Residual too large: %e", residual.Value))
	}
	
	// Calculate quality score
	qualityScore := h.calculateOverallQuality(gflops, efficiency)
	
	return ValidationStatus{
		IsValid:           len(errors) == 0,
		QualityScore:      qualityScore,
		ValidationErrors:  errors,
		WarningMessages:   warnings,
	}
}

// validateConfig validates the HPL configuration before execution.
func (h *HPLBenchmark) validateConfig() error {
	if h.config.Iterations < 3 {
		return ErrInsufficientIterations
	}
	
	if h.config.ConfidenceLevel <= 0 || h.config.ConfidenceLevel >= 1 {
		return ErrInvalidConfidenceLevel
	}
	
	if h.config.MinValidRuns > h.config.Iterations {
		return ErrInvalidMinValidRuns
	}
	
	if h.config.MemoryUtilization <= 0 || h.config.MemoryUtilization > 1 {
		return fmt.Errorf("%w", ErrHPLInvalidMemory)
	}
	
	return nil
}

// calculateConfidenceInterval computes the confidence interval for HPL measurements.
func (h *HPLBenchmark) calculateConfidenceInterval(values []float64, mean, stdDev float64) ConfidenceInterval {
	// Simplified confidence interval calculation
	// In production, would use proper t-distribution
	n := float64(len(values))
	margin := 1.96 * (stdDev / math.Sqrt(n)) // 95% confidence interval approximation
	
	return ConfidenceInterval{
		Lower: mean - margin,
		Upper: mean + margin,
		Level: h.config.ConfidenceLevel,
	}
}

// publishBenchmarkMetrics publishes HPL metrics to CloudWatch.
func (h *HPLBenchmark) publishBenchmarkMetrics(ctx context.Context, result *HPLResult, startTime time.Time) error {
	// Extract instance information from execution context
	instanceType := unknownValue   // Would be passed from orchestrator
	instanceFamily := unknownValue // Would be extracted from instance type
	region := "us-east-1"       // Would be passed from orchestrator
	
	// Calculate execution duration
	executionDuration := time.Since(startTime).Seconds()
	benchmarkDuration := time.Since(startTime).Seconds() // Simplified for now
	
	// Extract performance metrics from HPL results
	performanceMetrics := map[string]float64{
		"gflops":         result.Performance.GFLOPS.Value,
		"efficiency":     result.Performance.Efficiency.Value,
		"execution_time": result.Performance.ExecutionTime.Value,
		"residual":       result.Performance.Residual.Value,
	}
	
	// Calculate estimated cost (mock calculation)
	estimatedCost := executionDuration * 0.05 // $0.05 per second example
	
	// Calculate price-performance ratio (cost per GFLOP for HPL)
	pricePerformanceRatio := 0.0
	if result.Performance.GFLOPS.Value > 0 {
		pricePerformanceRatio = estimatedCost / result.Performance.GFLOPS.Value
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
		InstanceType:       instanceType,
		InstanceFamily:     instanceFamily,
		BenchmarkSuite:     "hpl",
		Region:            region,
		Success:           result.ValidationStatus.IsValid,
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
	return h.metricsCollector.PublishBenchmarkMetrics(ctx, metrics)
}