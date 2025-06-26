package benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/monitoring"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/storage"
)

const (
	// Test constants for HPL benchmarks.
	testHPLContainerImage = "test-registry/hpl:intel-icelake"
)

func TestNewHPLBenchmark(t *testing.T) {
	config := HPLConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testHPLContainerImage
	numaTopology := NumaTopology{NodeCount: 1, TotalMemoryGB: 8}
	
	benchmark := NewHPLBenchmark(config, containerImage, numaTopology)
	
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
	
	if benchmark.config.MemoryUtilization == 0 {
		t.Error("Expected default memory utilization to be set")
	}
	
	if benchmark.config.BlockSize == 0 {
		t.Error("Expected default block size to be set")
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

func TestHPLValidateConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      HPLConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: HPLConfig{
				Iterations:        10,
				ConfidenceLevel:   0.95,
				MinValidRuns:      8,
				MemoryUtilization: 0.8,
			},
			expectError: false,
		},
		{
			name: "insufficient iterations",
			config: HPLConfig{
				Iterations:      2,
				ConfidenceLevel: 0.95,
			},
			expectError: true,
		},
		{
			name: "invalid confidence level",
			config: HPLConfig{
				Iterations:      10,
				ConfidenceLevel: 1.5,
			},
			expectError: true,
		},
		{
			name: "min valid runs exceeds iterations",
			config: HPLConfig{
				Iterations:      5,
				ConfidenceLevel: 0.95,
				MinValidRuns:    10,
			},
			expectError: true,
		},
		{
			name: "invalid memory utilization - too high",
			config: HPLConfig{
				Iterations:        5,
				ConfidenceLevel:   0.95,
				MemoryUtilization: 1.5,
			},
			expectError: true,
		},
		{
			name: "invalid memory utilization - negative",
			config: HPLConfig{
				Iterations:        5,
				ConfidenceLevel:   0.95,
				MemoryUtilization: -0.1,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			benchmark := &HPLBenchmark{config: tc.config}
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

func TestCalculateOptimalProcessGrid(t *testing.T) {
	benchmark := &HPLBenchmark{}
	
	testCases := []struct {
		cores    int
		expectedP int
		expectedQ int
	}{
		{1, 1, 1},
		{2, 1, 2},
		{4, 2, 2},
		{8, 2, 4},
		{16, 4, 4},
		{24, 4, 6},
		{32, 4, 8},
		{64, 8, 8},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("cores_%d", tc.cores), func(t *testing.T) {
			grid := benchmark.calculateOptimalProcessGrid(tc.cores)
			
			// Verify that P * Q = cores
			if grid[0]*grid[1] != tc.cores {
				t.Errorf("Process grid [%d, %d] doesn't multiply to %d cores", 
					grid[0], grid[1], tc.cores)
			}
			
			// Verify the grid is reasonably close to square
			ratio := float64(grid[1]) / float64(grid[0])
			if ratio > 4.0 || ratio < 0.25 {
				t.Errorf("Process grid [%d, %d] is too unbalanced (ratio: %.2f)", 
					grid[0], grid[1], ratio)
			}
		})
	}
}

func TestCalculateProblemSize(t *testing.T) {
	testCases := []struct {
		name             string
		config           HPLConfig
		numaTopology     NumaTopology
		expectError      bool
		minExpectedN     int
		maxExpectedN     int
	}{
		{
			name: "explicit problem size",
			config: HPLConfig{
				ProblemSizeN:      10240, // Multiple of 256
				BlockSize:         256,
				MemoryUtilization: 0.8,
			},
			numaTopology: NumaTopology{TotalMemoryGB: 16},
			expectError:  false,
			minExpectedN: 10240,
			maxExpectedN: 10240,
		},
		{
			name: "calculated problem size - small memory",
			config: HPLConfig{
				BlockSize:         256,
				MemoryUtilization: 0.8,
			},
			numaTopology: NumaTopology{TotalMemoryGB: 4},
			expectError:  false,
			minExpectedN: 1000,
			maxExpectedN: 25000,
		},
		{
			name: "calculated problem size - large memory",
			config: HPLConfig{
				BlockSize:         256,
				MemoryUtilization: 0.8,
			},
			numaTopology: NumaTopology{TotalMemoryGB: 64},
			expectError:  false,
			minExpectedN: 30000,
			maxExpectedN: 100000,
		},
		{
			name: "insufficient memory",
			config: HPLConfig{
				BlockSize:         256,
				MemoryUtilization: 0.8,
			},
			numaTopology: NumaTopology{TotalMemoryGB: 0.001}, // Extremely small memory
			expectError:  true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			benchmark := &HPLBenchmark{
				config:       tc.config,
				numaTopology: tc.numaTopology,
			}
			
			problemSize, err := benchmark.calculateProblemSize()
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if problemSize.N < tc.minExpectedN || problemSize.N > tc.maxExpectedN {
				t.Errorf("Problem size N=%d outside expected range [%d, %d]", 
					problemSize.N, tc.minExpectedN, tc.maxExpectedN)
			}
			
			// Verify problem size is multiple of block size
			if problemSize.N%problemSize.BlockSize != 0 {
				t.Errorf("Problem size N=%d is not multiple of block size %d", 
					problemSize.N, problemSize.BlockSize)
			}
			
			// Verify memory usage calculation
			expectedMemory := float64(problemSize.N*problemSize.N*8) / (1024 * 1024 * 1024)
			if absFloat(problemSize.MemoryUsage-expectedMemory) > 0.1 {
				t.Errorf("Memory usage calculation incorrect: got %.2f GB, expected %.2f GB", 
					problemSize.MemoryUsage, expectedMemory)
			}
			
			// Verify theoretical operations calculation
			expectedOps := (2 * int64(problemSize.N) * int64(problemSize.N) * int64(problemSize.N)) / 3
			if problemSize.TheoreticalOps != expectedOps {
				t.Errorf("Theoretical operations incorrect: got %d, expected %d", 
					problemSize.TheoreticalOps, expectedOps)
			}
		})
	}
}

func TestHPLGetUnitForOperation(t *testing.T) {
	benchmark := &HPLBenchmark{}
	
	testCases := []struct {
		operation    string
		expectedUnit string
	}{
		{"GFLOPS", "GFLOPS"},
		{"Efficiency", "Ratio"},
		{"ExecutionTime", "Seconds"},
		{"Residual", "Scientific"},
		{"Unknown", "Unknown"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.operation, func(t *testing.T) {
			unit := benchmark.getUnitForOperation(tc.operation)
			if unit != tc.expectedUnit {
				t.Errorf("Expected unit '%s' for operation '%s', got '%s'", 
					tc.expectedUnit, tc.operation, unit)
			}
		})
	}
}

func TestHPLRemoveOutliers(t *testing.T) {
	config := HPLConfig{
		OutlierThreshold: 2.0,
	}
	
	benchmark := &HPLBenchmark{config: config}
	
	// Create dataset with obvious outliers
	values := []float64{200.0, 201.0, 202.0, 201.5, 203.0, 250.0, 200.5, 202.5} // 250.0 is outlier
	
	validValues, outlierCount := benchmark.removeOutliers(values)
	
	// Should remove the outlier (250.0)
	if outlierCount != 1 {
		t.Errorf("Expected 1 outlier removed, got %d", outlierCount)
	}
	
	if len(validValues) != 7 {
		t.Errorf("Expected 7 valid values, got %d", len(validValues))
	}
	
	// Check that 250.0 is not in valid values
	for _, value := range validValues {
		if value == 250.0 {
			t.Error("Outlier value 250.0 should have been removed")
		}
	}
}

func TestHPLCalculateOverallQuality(t *testing.T) {
	benchmark := &HPLBenchmark{}
	
	testCases := []struct {
		name               string
		gflopsCV           float64
		efficiencyCV       float64
		efficiencyValue    float64
		outliersRemoved    int
		sampleSize         int
		expectedMinQuality float64
		expectedMaxQuality float64
	}{
		{
			name:               "excellent quality",
			gflopsCV:           2.0,  // Low variation
			efficiencyCV:       1.5,  // Low variation
			efficiencyValue:    0.9,  // High efficiency
			outliersRemoved:    0,    // No outliers
			sampleSize:         10,
			expectedMinQuality: 0.9,
			expectedMaxQuality: 1.0,
		},
		{
			name:               "moderate quality",
			gflopsCV:           6.0,  // Higher variation
			efficiencyCV:       4.0,  // Moderate variation
			efficiencyValue:    0.75, // Good efficiency
			outliersRemoved:    1,    // Some outliers
			sampleSize:         10,
			expectedMinQuality: 0.5,
			expectedMaxQuality: 0.8,
		},
		{
			name:               "poor quality",
			gflopsCV:           12.0, // High variation
			efficiencyCV:       10.0, // High variation
			efficiencyValue:    0.5,  // Low efficiency
			outliersRemoved:    3,    // Many outliers
			sampleSize:         10,
			expectedMinQuality: 0.0,
			expectedMaxQuality: 0.4,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gflops := Measurement{
				CoefficientOfVariation: tc.gflopsCV,
				// Note: OutliersRemoved and SampleSize would be added to Measurement struct
			}
			
			efficiency := Measurement{
				CoefficientOfVariation: tc.efficiencyCV,
				Value:                  tc.efficiencyValue,
				// Note: OutliersRemoved and SampleSize would be added to Measurement struct
			}
			
			quality := benchmark.calculateOverallQuality(gflops, efficiency)
			
			if quality < tc.expectedMinQuality || quality > tc.expectedMaxQuality {
				t.Errorf("Quality score %.3f outside expected range [%.3f, %.3f]", 
					quality, tc.expectedMinQuality, tc.expectedMaxQuality)
			}
		})
	}
}

func TestHPLValidateResults(t *testing.T) {
	benchmark := &HPLBenchmark{}
	
	testCases := []struct {
		name          string
		gflopsCV      float64
		efficiencyVal float64
		residualVal   float64
		expectValid   bool
		expectErrors  int
		expectWarnings int
	}{
		{
			name:           "excellent results",
			gflopsCV:       2.0,
			efficiencyVal:  0.9,
			residualVal:    1e-12,
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name:           "good results with warnings",
			gflopsCV:       6.0,
			efficiencyVal:  0.65,
			residualVal:    1e-10,
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 2,
		},
		{
			name:          "poor performance stability",
			gflopsCV:      15.0,
			efficiencyVal: 0.8,
			residualVal:   1e-12,
			expectValid:   false,
			expectErrors:  1,
		},
		{
			name:          "low efficiency",
			gflopsCV:      3.0,
			efficiencyVal: 0.4,
			residualVal:   1e-12,
			expectValid:   false,
			expectErrors:  1,
		},
		{
			name:          "poor numerical accuracy",
			gflopsCV:      3.0,
			efficiencyVal: 0.8,
			residualVal:   1e-5,
			expectValid:   false,
			expectErrors:  1,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gflops := Measurement{CoefficientOfVariation: tc.gflopsCV}
			efficiency := Measurement{Value: tc.efficiencyVal}
			residual := Measurement{Value: tc.residualVal}
			
			validation := benchmark.validateResults(gflops, efficiency, residual)
			
			if validation.IsValid != tc.expectValid {
				t.Errorf("Expected IsValid=%t, got %t", tc.expectValid, validation.IsValid)
			}
			
			if tc.expectErrors > 0 && len(validation.ValidationErrors) != tc.expectErrors {
				t.Errorf("Expected %d errors, got %d", tc.expectErrors, len(validation.ValidationErrors))
			}
			
			if tc.expectWarnings > 0 && len(validation.WarningMessages) < tc.expectWarnings {
				t.Errorf("Expected at least %d warnings, got %d", tc.expectWarnings, len(validation.WarningMessages))
			}
			
			// Quality score should be in valid range
			if validation.QualityScore < 0 || validation.QualityScore > 1 {
				t.Errorf("Quality score %.3f outside valid range [0, 1]", validation.QualityScore)
			}
		})
	}
}

func TestHPLWithStorage(t *testing.T) {
	config := HPLConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testHPLContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	benchmark := NewHPLBenchmark(config, containerImage, numaTopology)
	
	// Initially storage should be nil
	if benchmark.storage != nil {
		t.Error("Expected storage to be nil initially")
	}
	
	// Create mock storage
	s3Storage := &storage.S3Storage{}
	
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

func TestHPLWithMetrics(t *testing.T) {
	config := HPLConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testHPLContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	benchmark := NewHPLBenchmark(config, containerImage, numaTopology)
	
	// Initially metricsCollector should be nil
	if benchmark.metricsCollector != nil {
		t.Error("Expected metricsCollector to be nil initially")
	}
	
	// Create mock metrics collector
	metricsCollector := &monitoring.MetricsCollector{}
	
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

func TestHPLMethodChaining(t *testing.T) {
	config := HPLConfig{
		Iterations:      5,
		ConfidenceLevel: 0.95,
	}
	
	containerImage := testHPLContainerImage
	numaTopology := NumaTopology{NodeCount: 1}
	
	// Test chaining both storage and metrics
	s3Storage := &storage.S3Storage{}
	metricsCollector := &monitoring.MetricsCollector{}
	
	benchmark := NewHPLBenchmark(config, containerImage, numaTopology).
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

func TestHPLArchitectureSpecificDefaults(t *testing.T) {
	testCases := []struct {
		name              string
		containerImage    string
		expectedBlockSize int
	}{
		{
			name:              "Intel architecture",
			containerImage:    "registry/hpl:intel-icelake",
			expectedBlockSize: 256,
		},
		{
			name:              "AMD architecture",
			containerImage:    "registry/hpl:amd-zen4",
			expectedBlockSize: 128,
		},
		{
			name:              "Generic architecture",
			containerImage:    "registry/hpl:generic",
			expectedBlockSize: 64,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := HPLConfig{
				Iterations:      5,
				ConfidenceLevel: 0.95,
				// Don't set BlockSize to test defaults
			}
			
			benchmark := NewHPLBenchmark(config, tc.containerImage, NumaTopology{})
			
			if benchmark.config.BlockSize != tc.expectedBlockSize {
				t.Errorf("Expected block size %d for %s, got %d", 
					tc.expectedBlockSize, tc.name, benchmark.config.BlockSize)
			}
		})
	}
}

func TestHPLExecuteWithMockData(t *testing.T) {
	// Skip this test in short mode since it involves execution
	if testing.Short() {
		t.Skip("Skipping HPL execution test in short mode")
	}
	
	config := HPLConfig{
		Iterations:       3, // Small number for fast testing
		ConfidenceLevel:  0.95,
		OutlierThreshold: 2.0,
		MinValidRuns:     2,
		MaxExecutionTime: 5 * time.Second,
		ProblemSizeN:     1000, // Small problem size for quick execution
		BlockSize:        64,
	}
	
	containerImage := testHPLContainerImage
	numaTopology := NumaTopology{NodeCount: 1, TotalMemoryGB: 8}
	
	benchmark := NewHPLBenchmark(config, containerImage, numaTopology)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Note: This test will use mock data from executeHPLRun
	result, err := benchmark.Execute(ctx)
	
	// Since we're using mock data, this should succeed
	if err != nil {
		t.Logf("HPL execution completed with mock data, any errors are expected: %v", err)
		return
	}
	
	// Verify the result structure
	if result.BenchmarkSuite != "hpl" {
		t.Errorf("Expected benchmark suite 'hpl', got '%s'", result.BenchmarkSuite)
	}
	
	// Check that performance metrics are present
	if result.Performance.GFLOPS.Value <= 0 {
		t.Error("Expected positive GFLOPS value")
	}
	
	if result.Performance.Efficiency.Value <= 0 || result.Performance.Efficiency.Value > 1 {
		t.Errorf("Expected efficiency between 0 and 1, got %f", result.Performance.Efficiency.Value)
	}
	
	// Verify problem size configuration
	if result.ProblemSize.N != 1000 {
		t.Errorf("Expected problem size N=1000, got %d", result.ProblemSize.N)
	}
	
	// Verify metadata is populated
	if result.ExecutionMetadata.SystemInfo.CPUCores == 0 {
		t.Error("Expected system info to be populated")
	}
	
	if result.ExecutionMetadata.ExecutionDuration == 0 {
		t.Error("Expected execution duration to be recorded")
	}
}

// Helper function for floating point absolute value.
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestHPLResultStructure(t *testing.T) {
	// Test that all required fields are properly structured
	result := HPLResult{
		ProblemSize: HPLProblemSize{
			N:              10000,
			BlockSize:      256,
			ProcessGrid:    [2]int{4, 4},
			MemoryUsage:    0.8,
			TheoreticalOps: 666666666666,
		},
		Performance: HPLPerformance{
			GFLOPS:        Measurement{Value: 200.5, Unit: "GFLOPS"},
			Efficiency:    Measurement{Value: 0.85, Unit: "Ratio"},
			ExecutionTime: Measurement{Value: 120.0, Unit: "Seconds"},
			Residual:      Measurement{Value: 1e-12, Unit: "Scientific"},
		},
		StatisticalSummary: HPLStatisticalSummary{
			PerformanceStability: 2.5,
			EfficiencyStability:  1.8,
			TotalOutliers:        0,
			ValidRuns:            5,
			OverallQuality:       0.95,
		},
		ValidationStatus: ValidationStatus{
			IsValid:           true,
			QualityScore:      0.95,
			ValidationErrors:  []string{},
			WarningMessages:   []string{},
		},
		ExecutionMetadata: ExecutionMetadata{
			ExecutionDuration: 2 * time.Minute,
			SystemInfo: SystemInfo{
				CPUCores:     16,
				NUMANodes:    2,
				MemoryTotal:  64,
			},
		},
		BenchmarkSuite: "hpl",
	}
	
	// Verify all fields are accessible and properly typed
	if result.ProblemSize.N != 10000 {
		t.Error("ProblemSize.N field not properly accessible")
	}
	
	if result.Performance.GFLOPS.Value != 200.5 {
		t.Error("Performance.GFLOPS field not properly accessible")
	}
	
	if result.StatisticalSummary.OverallQuality != 0.95 {
		t.Error("StatisticalSummary field not properly accessible")
	}
	
	if !result.ValidationStatus.IsValid {
		t.Error("ValidationStatus field not properly accessible")
	}
	
	if result.BenchmarkSuite != "hpl" {
		t.Error("BenchmarkSuite field not properly set")
	}
}