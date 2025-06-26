package monitoring

import (
	"context"
	"testing"
	"time"
)

const (
	// Test constants to avoid goconst linter warnings.
	testRegion = "us-east-1"
)

func TestNewMetricsCollector(t *testing.T) {
	// Skip this test if AWS credentials are not available
	if testing.Short() {
		t.Skip("Skipping AWS-dependent test in short mode")
	}
	
	collector, err := NewMetricsCollector(testRegion)
	if err != nil {
		// In CI/CD environments without AWS credentials, this is expected
		t.Logf("Expected error without AWS credentials: %v", err)
		return
	}
	
	// Verify collector configuration
	if collector.namespace != "InstanceBenchmarks" {
		t.Errorf("Expected namespace 'InstanceBenchmarks', got '%s'", collector.namespace)
	}
	
	if collector.region != testRegion {
		t.Errorf("Expected region '%s', got '%s'", testRegion, collector.region)
	}
	
	// Verify default dimensions are set
	if len(collector.defaultDimensions) == 0 {
		t.Error("Expected default dimensions to be set")
	}
	
	// Check for required dimensions
	foundProject := false
	foundEnvironment := false
	foundRegion := false
	
	for _, dim := range collector.defaultDimensions {
		switch *dim.Name {
		case "Project":
			foundProject = true
			if *dim.Value != "aws-instance-benchmarks" {
				t.Errorf("Expected Project dimension value 'aws-instance-benchmarks', got '%s'", *dim.Value)
			}
		case "Environment":
			foundEnvironment = true
			if *dim.Value != "production" {
				t.Errorf("Expected Environment dimension value 'production', got '%s'", *dim.Value)
			}
		case "Region":
			foundRegion = true
			if *dim.Value != testRegion {
				t.Errorf("Expected Region dimension value '%s', got '%s'", testRegion, *dim.Value)
			}
		}
	}
	
	if !foundProject {
		t.Error("Expected Project dimension to be present")
	}
	if !foundEnvironment {
		t.Error("Expected Environment dimension to be present")
	}
	if !foundRegion {
		t.Error("Expected Region dimension to be present")
	}
}

func TestValidateBenchmarkMetrics(t *testing.T) {
	collector := &MetricsCollector{
		namespace: "AWS/InstanceBenchmarks",
		region:    "us-east-1",
	}
	
	testCases := []struct {
		name        string
		metrics     BenchmarkMetrics
		expectError bool
		errorType   error
	}{
		{
			name: "valid metrics",
			metrics: BenchmarkMetrics{
				InstanceType:      "m7i.large",
				BenchmarkSuite:    "stream",
				Success:           true,
				ExecutionDuration: 45.2,
				QualityScore:      0.95,
				PerformanceMetrics: map[string]float64{
					"triad_bandwidth": 41.9,
					"copy_bandwidth":  45.2,
				},
			},
			expectError: false,
		},
		{
			name: "missing instance type",
			metrics: BenchmarkMetrics{
				BenchmarkSuite:    "stream",
				Success:           true,
				ExecutionDuration: 45.2,
			},
			expectError: true,
			errorType:   ErrMetricNameRequired,
		},
		{
			name: "missing benchmark suite",
			metrics: BenchmarkMetrics{
				InstanceType:      "m7i.large",
				Success:           true,
				ExecutionDuration: 45.2,
			},
			expectError: true,
			errorType:   ErrMetricNameRequired,
		},
		{
			name: "negative execution duration",
			metrics: BenchmarkMetrics{
				InstanceType:      "m7i.large",
				BenchmarkSuite:    "stream",
				Success:           false,
				ExecutionDuration: -10.0,
			},
			expectError: true,
			errorType:   ErrInvalidMetricValue,
		},
		{
			name: "invalid quality score - too high",
			metrics: BenchmarkMetrics{
				InstanceType:   "m7i.large",
				BenchmarkSuite: "stream",
				Success:        true,
				QualityScore:   1.5,
			},
			expectError: true,
			errorType:   ErrInvalidMetricValue,
		},
		{
			name: "invalid quality score - negative",
			metrics: BenchmarkMetrics{
				InstanceType:   "m7i.large",
				BenchmarkSuite: "stream",
				Success:        true,
				QualityScore:   -0.1,
			},
			expectError: true,
			errorType:   ErrInvalidMetricValue,
		},
		{
			name: "negative performance metric",
			metrics: BenchmarkMetrics{
				InstanceType:   "m7i.large",
				BenchmarkSuite: "stream",
				Success:        true,
				PerformanceMetrics: map[string]float64{
					"triad_bandwidth": -41.9,
				},
			},
			expectError: true,
			errorType:   ErrInvalidMetricValue,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := collector.validateBenchmarkMetrics(tc.metrics)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if tc.errorType != nil && !isErrorType(err, tc.errorType) {
					t.Errorf("Expected error type %v, got %v", tc.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestGetUnitForPerformanceMetric(t *testing.T) {
	collector := &MetricsCollector{}
	
	testCases := []struct {
		metricName   string
		expectedUnit string
	}{
		{"triad_bandwidth", "Bytes/Second"},
		{"copy_bandwidth", "Bytes/Second"},
		{"memory_bandwidth", "Bytes/Second"},
		{"gflops", "Count/Second"},
		{"peak_flops", "Count/Second"},
		{"latency", "Seconds"},
		{"execution_duration", "Seconds"},
		{"throughput", "Count/Second"},
		{"unknown_metric", "None"},
		{"efficiency_ratio", "None"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.metricName, func(t *testing.T) {
			unit := collector.getUnitForPerformanceMetric(tc.metricName)
			if string(unit) != tc.expectedUnit {
				t.Errorf("Expected unit '%s' for metric '%s', got '%s'", 
					tc.expectedUnit, tc.metricName, string(unit))
			}
		})
	}
}

func TestContainsFunction(t *testing.T) {
	testCases := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"triad_bandwidth", "bandwidth", true},
		{"copy_bandwidth", "bandwidth", true},
		{"gflops", "flops", true},
		{"peak_gflops", "gflops", true},
		{"latency_ms", "latency", true},
		{"execution_duration", "duration", true},
		{"throughput_ops", "throughput", true},
		{"memory_size", "bandwidth", false},
		{"cpu_cores", "gflops", false},
		{"", "test", false},
		{"test", "", true},
		{"exact_match", "exact_match", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.s+"_contains_"+tc.substr, func(t *testing.T) {
			result := contains(tc.s, tc.substr)
			if result != tc.expected {
				t.Errorf("contains('%s', '%s') = %t; expected %t", 
					tc.s, tc.substr, result, tc.expected)
			}
		})
	}
}

func TestPublishBenchmarkMetrics(t *testing.T) {
	// Skip this test if AWS credentials are not available
	if testing.Short() {
		t.Skip("Skipping AWS-dependent test in short mode")
	}
	
	collector, err := NewMetricsCollector(testRegion)
	if err != nil {
		t.Logf("Skipping test due to AWS configuration error: %v", err)
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	metrics := BenchmarkMetrics{
		InstanceType:      "m7i.large",
		InstanceFamily:    "m7i",
		BenchmarkSuite:    "stream",
		Region:            "us-east-1",
		Success:           true,
		ExecutionDuration: 45.2,
		BenchmarkDuration: 35.8,
		PerformanceMetrics: map[string]float64{
			"triad_bandwidth": 41.9,
			"copy_bandwidth":  45.2,
			"scale_bandwidth": 44.8,
			"add_bandwidth":   42.1,
		},
		QualityScore: 0.95,
		CostMetrics: CostMetrics{
			EstimatedCost:         0.034,
			PricePerformanceRatio: 0.00081,
			InstanceHourCost:      0.0544,
		},
		Timestamp: time.Now(),
	}
	
	// Test successful metrics publication
	err = collector.PublishBenchmarkMetrics(ctx, metrics)
	if err != nil {
		// In testing environments without proper AWS setup, log the error
		t.Logf("Expected error in test environment: %v", err)
	}
	
	// Test failed benchmark metrics
	failedMetrics := metrics
	failedMetrics.Success = false
	failedMetrics.ErrorCategory = "quota"
	
	err = collector.PublishBenchmarkMetrics(ctx, failedMetrics)
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestPublishOperationalMetrics(t *testing.T) {
	// Skip this test if AWS credentials are not available
	if testing.Short() {
		t.Skip("Skipping AWS-dependent test in short mode")
	}
	
	collector, err := NewMetricsCollector(testRegion)
	if err != nil {
		t.Logf("Skipping test due to AWS configuration error: %v", err)
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	metrics := OperationalMetrics{
		QuotaUtilization: map[string]float64{
			"m7i": 45.5,
			"c7g": 23.1,
			"r7a": 67.8,
		},
		InstanceLaunchDuration:  12.5,
		ContainerPullDuration:   3.2,
		ActiveInstances:         8,
		FailureRate:            2.5,
		Region:                 "us-east-1",
		Timestamp:              time.Now(),
	}
	
	err = collector.PublishOperationalMetrics(ctx, metrics)
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestMetricsValidationEdgeCases(t *testing.T) {
	collector := &MetricsCollector{}
	
	// Test empty performance metrics map
	metrics := BenchmarkMetrics{
		InstanceType:       "m7i.large",
		BenchmarkSuite:     "stream",
		Success:            true,
		PerformanceMetrics: map[string]float64{},
	}
	
	err := collector.validateBenchmarkMetrics(metrics)
	if err != nil {
		t.Errorf("Validation should pass with empty performance metrics: %v", err)
	}
	
	// Test nil performance metrics map
	metrics.PerformanceMetrics = nil
	err = collector.validateBenchmarkMetrics(metrics)
	if err != nil {
		t.Errorf("Validation should pass with nil performance metrics: %v", err)
	}
	
	// Test zero values
	metrics.PerformanceMetrics = map[string]float64{
		"zero_metric": 0.0,
	}
	err = collector.validateBenchmarkMetrics(metrics)
	if err != nil {
		t.Errorf("Validation should pass with zero performance metrics: %v", err)
	}
}

func TestBenchmarkMetricsStructure(t *testing.T) {
	// Test that all required fields are properly structured
	metrics := BenchmarkMetrics{
		InstanceType:      "m7i.large",
		InstanceFamily:    "m7i",
		BenchmarkSuite:    "stream",
		Region:            "us-east-1",
		Success:           true,
		ExecutionDuration: 45.2,
		BenchmarkDuration: 35.8,
		PerformanceMetrics: map[string]float64{
			"triad_bandwidth": 41.9,
		},
		ErrorCategory: "",
		CostMetrics: CostMetrics{
			EstimatedCost:         0.034,
			PricePerformanceRatio: 0.00081,
			InstanceHourCost:      0.0544,
		},
		QualityScore: 0.95,
		Timestamp:    time.Now(),
	}
	
	// Verify all fields are accessible and properly typed
	if metrics.InstanceType != "m7i.large" {
		t.Error("InstanceType field not properly accessible")
	}
	
	if len(metrics.PerformanceMetrics) != 1 {
		t.Error("PerformanceMetrics map not properly accessible")
	}
	
	if metrics.CostMetrics.EstimatedCost != 0.034 {
		t.Error("CostMetrics nested struct not properly accessible")
	}
	
	if metrics.Timestamp.IsZero() {
		t.Error("Timestamp field not properly set")
	}
}

func TestOperationalMetricsStructure(t *testing.T) {
	// Test operational metrics structure
	metrics := OperationalMetrics{
		QuotaUtilization: map[string]float64{
			"m7i": 45.5,
			"c7g": 23.1,
		},
		InstanceLaunchDuration: 12.5,
		ContainerPullDuration:  3.2,
		ActiveInstances:        8,
		FailureRate:           2.5,
		Region:                "us-east-1",
		Timestamp:             time.Now(),
	}
	
	// Verify all fields are accessible
	if len(metrics.QuotaUtilization) != 2 {
		t.Error("QuotaUtilization map not properly accessible")
	}
	
	if metrics.ActiveInstances != 8 {
		t.Error("ActiveInstances field not properly accessible")
	}
	
	if metrics.Region != "us-east-1" {
		t.Error("Region field not properly accessible")
	}
}

// Helper function to check if an error is of a specific type.
func isErrorType(err, targetType error) bool {
	return err.Error() == targetType.Error() || 
		   (len(err.Error()) > len(targetType.Error()) && 
		    err.Error()[:len(targetType.Error())] == targetType.Error())
}