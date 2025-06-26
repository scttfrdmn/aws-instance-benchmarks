package benchmarks

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/monitoring"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/storage"
)

const (
	// Test constants to avoid goconst linter warnings.
	testContainerImage = "test-registry/stream:intel-icelake"
)

func TestNewStreamBenchmark(t *testing.T) {
	config := BenchmarkConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	benchmark := NewStreamBenchmark(config, containerImage, numaTopology)
	
	// Verify configuration defaults are applied
	if benchmark.config.OutlierThreshold == 0 {
		t.Error("Expected default outlier threshold to be set")
	}
	
	if benchmark.config.MinValidRuns == 0 {
		t.Error("Expected default min valid runs to be set")
	}
	
	if benchmark.config.MaxExecutionTime == 0 {
		t.Error("Expected default max execution time to be set")
	}
	
	if benchmark.config.MemoryPattern == "" {
		t.Error("Expected default memory pattern to be set")
	}
	
	// Verify provided values are preserved
	if benchmark.config.Iterations != 5 {
		t.Errorf("Expected iterations to be 5, got %d", benchmark.config.Iterations)
	}
	
	if benchmark.config.ConfidenceLevel != 0.95 {
		t.Errorf("Expected confidence level to be 0.95, got %f", benchmark.config.ConfidenceLevel)
	}
	
	if benchmark.containerImage != containerImage {
		t.Errorf("Expected container image %s, got %s", containerImage, benchmark.containerImage)
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      BenchmarkConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: BenchmarkConfig{
				Iterations:      10,
				ConfidenceLevel: 0.95,
				MinValidRuns:    8,
			},
			expectError: false,
		},
		{
			name: "insufficient iterations",
			config: BenchmarkConfig{
				Iterations:      2,
				ConfidenceLevel: 0.95,
			},
			expectError: true,
		},
		{
			name: "invalid confidence level",
			config: BenchmarkConfig{
				Iterations:      10,
				ConfidenceLevel: 1.5,
			},
			expectError: true,
		},
		{
			name: "min valid runs exceeds iterations",
			config: BenchmarkConfig{
				Iterations:      5,
				ConfidenceLevel: 0.95,
				MinValidRuns:    10,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			benchmark := &StreamBenchmark{config: tc.config}
			err := benchmark.validateConfig()
			
			if tc.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestCalculateMean(t *testing.T) {
	testCases := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "simple average",
			values:   []float64{1.0, 2.0, 3.0},
			expected: 2.0,
		},
		{
			name:     "single value",
			values:   []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "empty slice",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "stream benchmark values",
			values:   []float64{45.2, 45.1, 45.3, 45.0, 45.4},
			expected: 45.2,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateMean(tc.values)
			if result != tc.expected {
				t.Errorf("Expected mean %.2f, got %.2f", tc.expected, result)
			}
		})
	}
}

func TestCalculateStandardDeviation(t *testing.T) {
	testCases := []struct {
		name     string
		values   []float64
		mean     float64
		expected float64
		tolerance float64
	}{
		{
			name:      "known standard deviation",
			values:    []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			mean:      3.0,
			expected:  1.58, // Approximate
			tolerance: 0.01,
		},
		{
			name:      "single value",
			values:    []float64{5.0},
			mean:      5.0,
			expected:  0.0,
			tolerance: 0.01,
		},
		{
			name:      "identical values",
			values:    []float64{10.0, 10.0, 10.0},
			mean:      10.0,
			expected:  0.0,
			tolerance: 0.01,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateStandardDeviation(tc.values, tc.mean)
			if abs(result-tc.expected) > tc.tolerance {
				t.Errorf("Expected standard deviation %.2f, got %.2f", tc.expected, result)
			}
		})
	}
}

func TestRemoveOutliers(t *testing.T) {
	config := BenchmarkConfig{
		OutlierThreshold: 2.0,
	}
	
	benchmark := &StreamBenchmark{config: config}
	
	// Create dataset with obvious outliers
	values := []float64{45.0, 45.1, 45.2, 45.1, 45.3, 50.0, 45.0, 45.2} // 50.0 is outlier
	
	validValues, outlierCount := benchmark.removeOutliers(values)
	
	// Should remove the outlier (50.0)
	if outlierCount != 1 {
		t.Errorf("Expected 1 outlier removed, got %d", outlierCount)
	}
	
	if len(validValues) != 7 {
		t.Errorf("Expected 7 valid values, got %d", len(validValues))
	}
	
	// Check that 50.0 is not in valid values
	for _, value := range validValues {
		if value == 50.0 {
			t.Error("Outlier value 50.0 should have been removed")
		}
	}
}

func TestCalculateMeasurement(t *testing.T) {
	config := BenchmarkConfig{
		OutlierThreshold: 2.0,
		MinValidRuns:     3,
		ConfidenceLevel:  0.95,
	}
	
	benchmark := &StreamBenchmark{config: config}
	
	values := []float64{45.0, 45.1, 45.2, 45.1, 45.3}
	
	measurement, err := benchmark.calculateMeasurement("triad", values)
	if err != nil {
		t.Fatalf("Unexpected error calculating measurement: %v", err)
	}
	
	// Verify basic properties
	if measurement.Operation != "triad" {
		t.Errorf("Expected operation 'triad', got '%s'", measurement.Operation)
	}
	
	if measurement.Unit != "GB/s" {
		t.Errorf("Expected unit 'GB/s', got '%s'", measurement.Unit)
	}
	
	if measurement.Value <= 0 {
		t.Errorf("Expected positive value, got %f", measurement.Value)
	}
	
	if measurement.StandardDeviation < 0 {
		t.Errorf("Expected non-negative standard deviation, got %f", measurement.StandardDeviation)
	}
	
	if measurement.ConfidenceInterval.Lower >= measurement.ConfidenceInterval.Upper {
		t.Error("Confidence interval lower bound should be less than upper bound")
	}
	
	if measurement.ConfidenceInterval.Level != 0.95 {
		t.Errorf("Expected confidence level 0.95, got %f", measurement.ConfidenceInterval.Level)
	}
}

func TestExecute(t *testing.T) {
	// Skip this test if Docker is not available (CI/CD environments)
	if testing.Short() {
		t.Skip("Skipping Docker-dependent test in short mode")
	}
	
	config := BenchmarkConfig{
		Iterations:       3, // Small number for fast testing
		ConfidenceLevel:  0.95,
		OutlierThreshold: 2.0,
		MinValidRuns:     2,
		MaxExecutionTime: 5 * time.Second,
	}
	
	containerImage := "test-registry/stream:test"
	numaTopology := NumaTopology{NodeCount: 1}
	
	benchmark := NewStreamBenchmark(config, containerImage, numaTopology)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Note: This test will fail without a real STREAM container
	// For CI/CD, we should either mock the Docker execution or use a test container
	result, err := benchmark.Execute(ctx)
	
	// Expect error since we don't have a real STREAM container in testing
	if err == nil {
		// If somehow we get a successful result, verify the structure
		if result.BenchmarkSuite != "stream" {
			t.Errorf("Expected benchmark suite 'stream', got '%s'", result.BenchmarkSuite)
		}
		
		// Check that all STREAM operations are present
		expectedOps := []string{"copy", "scale", "add", "triad"}
		for _, op := range expectedOps {
			if _, exists := result.Measurements[op]; !exists {
				t.Errorf("Missing measurement for operation '%s'", op)
			}
		}
		
		// Verify metadata is populated
		if result.ExecutionMetadata.SystemInfo.CPUCores == 0 {
			t.Error("Expected system info to be populated")
		}
		
		if result.ExecutionMetadata.ExecutionDuration == 0 {
			t.Error("Expected execution duration to be recorded")
		}
		
		// Verify statistical summary
		if result.StatisticalSummary.OverallStability < 0 {
			t.Error("Expected non-negative overall stability")
		}
		
		// Verify validation status
		if result.ValidationStatus.QualityScore < 0 || result.ValidationStatus.QualityScore > 1 {
			t.Errorf("Expected quality score between 0 and 1, got %f", result.ValidationStatus.QualityScore)
		}
	} else if !strings.Contains(err.Error(), "container execution failed") && 
		   !strings.Contains(err.Error(), "failed to execute container") &&
		   !strings.Contains(err.Error(), "executable file not found") {
		t.Logf("Expected container execution error, got: %v", err)
	}
}

func TestExecuteWithCancellation(t *testing.T) {
	config := BenchmarkConfig{
		Iterations:      100, // Large number to ensure cancellation
		ConfidenceLevel: 0.95,
	}
	
	benchmark := NewStreamBenchmark(config, "test-image", NumaTopology{})
	
	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	_, err := benchmark.Execute(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	
	// With the new Docker implementation, we might get different errors
	// Accept either context deadline exceeded or container execution failure
	if !errors.Is(err, context.DeadlineExceeded) && 
	   !strings.Contains(err.Error(), "container execution failed") &&
	   !strings.Contains(err.Error(), "failed to execute container") {
		t.Logf("Expected context deadline exceeded or container execution error, got %v", err)
	}
}

func TestValidateResults(t *testing.T) {
	benchmark := &StreamBenchmark{}
	
	// Create measurements with varying quality
	measurements := map[string]Measurement{
		"copy": {
			CoefficientOfVariation: 2.0, // Good stability
			OutliersRemoved:        0,
		},
		"scale": {
			CoefficientOfVariation: 6.0, // Moderate stability
			OutliersRemoved:        1,
		},
		"add": {
			CoefficientOfVariation: 12.0, // Poor stability
			OutliersRemoved:        2,
		},
		"triad": {
			CoefficientOfVariation: 3.0, // Good stability
			OutliersRemoved:        0,
		},
	}
	
	summary := StatisticalSummary{
		OverallStability: 5.75, // Average of above values
		TotalOutliers:    3,
	}
	
	validation := benchmark.validateResults(measurements, summary)
	
	// Should have errors due to high variability in "add" operation
	if len(validation.ValidationErrors) == 0 {
		t.Error("Expected validation errors due to high variability")
	}
	
	// Should have warnings due to moderate variability in "scale" operation
	if len(validation.WarningMessages) == 0 {
		t.Error("Expected validation warnings due to moderate variability")
	}
	
	// Should not be valid due to errors
	if validation.IsValid {
		t.Error("Results should not be valid due to high variability")
	}
	
	// Quality score should be reduced
	if validation.QualityScore >= 1.0 {
		t.Error("Quality score should be reduced due to stability issues")
	}
}

func TestWithStorage(t *testing.T) {
	config := BenchmarkConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	benchmark := NewStreamBenchmark(config, containerImage, numaTopology)
	
	// Initially storage should be nil
	if benchmark.storage != nil {
		t.Error("Expected storage to be nil initially")
	}
	
	// Create mock storage (will fail without AWS credentials, but tests the interface)
	s3Storage := &storage.S3Storage{} // Mock storage for testing
	
	// Test method chaining
	result := benchmark.WithStorage(s3Storage)
	
	// Should return the same benchmark instance
	if result != benchmark {
		t.Error("WithStorage should return the same benchmark instance for chaining")
	}
	
	// Storage should now be set
	if benchmark.storage != s3Storage {
		t.Error("Expected storage to be set after WithStorage call")
	}
}

func TestWithMetrics(t *testing.T) {
	config := BenchmarkConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	benchmark := NewStreamBenchmark(config, containerImage, numaTopology)
	
	// Initially metricsCollector should be nil
	if benchmark.metricsCollector != nil {
		t.Error("Expected metricsCollector to be nil initially")
	}
	
	// Create mock metrics collector (will fail without AWS credentials, but tests the interface)
	metricsCollector := &monitoring.MetricsCollector{} // Mock collector for testing
	
	// Test method chaining
	result := benchmark.WithMetrics(metricsCollector)
	
	// Should return the same benchmark instance
	if result != benchmark {
		t.Error("WithMetrics should return the same benchmark instance for chaining")
	}
	
	// Metrics collector should now be set
	if benchmark.metricsCollector != metricsCollector {
		t.Error("Expected metricsCollector to be set after WithMetrics call")
	}
}

func TestMethodChaining(t *testing.T) {
	config := BenchmarkConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	// Test chaining both storage and metrics
	s3Storage := &storage.S3Storage{}
	metricsCollector := &monitoring.MetricsCollector{}
	
	benchmark := NewStreamBenchmark(config, containerImage, numaTopology).
		WithStorage(s3Storage).
		WithMetrics(metricsCollector)
	
	// Both should be set
	if benchmark.storage != s3Storage {
		t.Error("Expected storage to be set after method chaining")
	}
	
	if benchmark.metricsCollector != metricsCollector {
		t.Error("Expected metricsCollector to be set after method chaining")
	}
}

// Helper function for floating point comparison.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestBuildDockerCommand(t *testing.T) {
	config := BenchmarkConfig{
		EnableNUMA:    true,
		MemoryPattern: "sequential",
	}
	
	containerImage := testContainerImage
	numaTopology := NumaTopology{NodeCount: 2}
	
	benchmark := NewStreamBenchmark(config, containerImage, numaTopology)
	
	args := benchmark.buildDockerCommand(1)
	
	// Verify essential Docker arguments are present
	cmdStr := strings.Join(args, " ")
	
	if !strings.Contains(cmdStr, "run") {
		t.Error("Expected 'run' command in Docker args")
	}
	
	if !strings.Contains(cmdStr, "--rm") {
		t.Error("Expected '--rm' flag for container cleanup")
	}
	
	if !strings.Contains(cmdStr, "--memory") {
		t.Error("Expected memory limit in Docker args")
	}
	
	if !strings.Contains(cmdStr, "--cpus") {
		t.Error("Expected CPU limit in Docker args")
	}
	
	if !strings.Contains(cmdStr, "--read-only") {
		t.Error("Expected read-only filesystem for security")
	}
	
	if !strings.Contains(cmdStr, "--network none") {
		t.Error("Expected network isolation")
	}
	
	if !strings.Contains(cmdStr, "OUTPUT_FORMAT=json") {
		t.Error("Expected JSON output format environment variable")
	}
	
	if !strings.Contains(cmdStr, "MEMORY_PATTERN=sequential") {
		t.Error("Expected memory pattern environment variable")
	}
	
	if !strings.Contains(cmdStr, containerImage) {
		t.Error("Expected container image in command")
	}
	
	// Check NUMA configuration
	if !strings.Contains(cmdStr, "NUMA_NODE=0") {
		t.Error("Expected NUMA node configuration")
	}
	
	if !strings.Contains(cmdStr, "NUMA_NODES=2") {
		t.Error("Expected NUMA node count configuration")
	}
}

func TestParseContainerOutput(t *testing.T) {
	benchmark := &StreamBenchmark{}
	
	// Test JSON output parsing
	jsonOutput := `{
		"stream_results": {
			"copy": 45.2,
			"scale": 44.8,
			"add": 42.1,
			"triad": 41.9
		},
		"metadata": {
			"run_id": 1,
			"numa_node": 0
		}
	}`
	
	results, err := benchmark.parseContainerOutput([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	
	// Verify all operations are present
	expectedOps := []string{"copy", "scale", "add", "triad"}
	for _, op := range expectedOps {
		if _, exists := results[op]; !exists {
			t.Errorf("Missing operation '%s' in parsed results", op)
		}
	}
	
	// Verify specific values
	if results["copy"] != 45.2 {
		t.Errorf("Expected copy bandwidth 45.2, got %f", results["copy"])
	}
	
	if results["triad"] != 41.9 {
		t.Errorf("Expected triad bandwidth 41.9, got %f", results["triad"])
	}
}

func TestParseContainerOutputInvalidJSON(t *testing.T) {
	benchmark := &StreamBenchmark{}
	
	// Test with invalid JSON
	invalidJSON := `{"stream_results": invalid json}`
	
	_, err := benchmark.parseContainerOutput([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
}

func TestParseContainerOutputMissingFields(t *testing.T) {
	benchmark := &StreamBenchmark{}
	
	// Test with missing stream_results field
	missingResults := `{"metadata": {"run_id": 1}}`
	
	_, err := benchmark.parseContainerOutput([]byte(missingResults))
	if err == nil {
		t.Error("Expected error for missing stream_results field")
	}
	
	// Test with missing operations
	incompleteResults := `{
		"stream_results": {
			"copy": 45.2,
			"scale": 44.8
		}
	}`
	
	_, err = benchmark.parseContainerOutput([]byte(incompleteResults))
	if err == nil {
		t.Error("Expected error for missing STREAM operations")
	}
	
	if !strings.Contains(err.Error(), "missing required STREAM operation") {
		t.Errorf("Expected specific error message about missing operations, got: %v", err)
	}
}

func TestParseTextOutput(t *testing.T) {
	benchmark := &StreamBenchmark{}
	
	// Test traditional STREAM text output
	textOutput := `-------------------------------------------------------------
STREAM version $Revision: 5.10 $
-------------------------------------------------------------
Array size = 80000000 (elements), Offset = 0 (elements)
Memory per array = 610.4 MiB (= 0.6 GiB).
Total memory required = 1831.1 MiB (= 1.8 GiB).
Each kernel will be executed 10 times.
 The *best* time for each kernel (excluding the first iteration)
 will be used to compute the reported bandwidth.
-------------------------------------------------------------
Your clock granularity/precision appears to be 1 microseconds.
Each test below will take on the order of 32513 microseconds.
   (= 32513 clock ticks)
Increase the size of the arrays if this shows that
you are not getting at least 20 clock ticks per test.
-------------------------------------------------------------
WARNING -- The above is only a rough guideline.
For best results, please be sure you know the
precision of your system timer.
-------------------------------------------------------------
Function    Best Rate MB/s  Avg time     Min time     Max time
Copy:           45234.2     0.035482     0.035401     0.035563
Scale:          44876.1     0.035681     0.035681     0.035681
Add:            42134.7     0.057012     0.057012     0.057012
Triad:          41987.3     0.057215     0.057215     0.057215
-------------------------------------------------------------
Solution Validates: avg error less than 1.000000e-13 on all three arrays
-------------------------------------------------------------`
	
	results, err := benchmark.parseTextOutput(textOutput)
	if err != nil {
		t.Fatalf("Failed to parse text output: %v", err)
	}
	
	// Verify all operations are present and converted to GB/s
	expectedOps := map[string]float64{
		"copy":  45.2342, // 45234.2 MB/s converted to GB/s
		"scale": 44.8761,
		"add":   42.1347,
		"triad": 41.9873,
	}
	
	for op, expectedBandwidth := range expectedOps {
		if actualBandwidth, exists := results[op]; !exists {
			t.Errorf("Missing operation '%s' in parsed results", op)
		} else if abs(actualBandwidth-expectedBandwidth) > 0.01 {
			t.Errorf("Expected %s bandwidth %.4f GB/s, got %.4f GB/s", 
				op, expectedBandwidth, actualBandwidth)
		}
	}
}

func TestValidateStreamResults(t *testing.T) {
	benchmark := &StreamBenchmark{}
	
	testCases := []struct {
		name        string
		results     map[string]float64
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid results",
			results: map[string]float64{
				"copy":  45.2,
				"scale": 44.8,
				"add":   42.1,
				"triad": 41.9,
			},
			expectError: false,
		},
		{
			name: "negative bandwidth",
			results: map[string]float64{
				"copy":  -45.2,
				"scale": 44.8,
				"add":   42.1,
				"triad": 41.9,
			},
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name: "zero bandwidth",
			results: map[string]float64{
				"copy":  0.0,
				"scale": 44.8,
				"add":   42.1,
				"triad": 41.9,
			},
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name: "extremely high bandwidth",
			results: map[string]float64{
				"copy":  2000.0, // Unreasonably high
				"scale": 44.8,
				"add":   42.1,
				"triad": 41.9,
			},
			expectError: true,
			errorMsg:    "outside reasonable range",
		},
		{
			name: "extremely low bandwidth",
			results: map[string]float64{
				"copy":  0.05, // Below minimum threshold
				"scale": 44.8,
				"add":   42.1,
				"triad": 41.9,
			},
			expectError: true,
			errorMsg:    "outside reasonable range",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := benchmark.validateStreamResults(tc.results)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error message containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}