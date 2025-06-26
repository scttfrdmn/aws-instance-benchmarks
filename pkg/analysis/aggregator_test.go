package analysis

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/benchmarks"
)

// createTestAggregator creates a data aggregator with valid test configuration.
func createTestAggregator() (*DataAggregator, error) {
	config := AggregationConfig{
		GroupingDimensions: []string{"instance_type"},
		StatisticalConfig: StatisticalConfig{
			ConfidenceLevel: 0.95,
			MinSampleSize:   3,
		},
		QualityThreshold: 0.7,
	}
	dataSource := NewMockDataSource()
	return NewDataAggregator(config, dataSource)
}

// MockDataSource implements the DataSource interface for testing.
type MockDataSource struct {
	results []ResultMetadata
	data    map[string]BenchmarkData
}

func NewMockDataSource() *MockDataSource {
	return &MockDataSource{
		results: []ResultMetadata{},
		data:    make(map[string]BenchmarkData),
	}
}

func (m *MockDataSource) AddResult(metadata ResultMetadata, data BenchmarkData) {
	m.results = append(m.results, metadata)
	m.data[metadata.ResultID] = data
}

func (m *MockDataSource) ListResults(_ context.Context, _ TimeWindow) ([]ResultMetadata, error) {
	return m.results, nil
}

func (m *MockDataSource) LoadResults(_ context.Context, resultIDs []string) ([]BenchmarkData, error) {
	var data []BenchmarkData
	for _, id := range resultIDs {
		if result, exists := m.data[id]; exists {
			data = append(data, result)
		}
	}
	return data, nil
}

func (m *MockDataSource) GetSchema(_ context.Context) (SchemaVersion, error) {
	return SchemaVersion{Major: 1, Minor: 0, Patch: 0}, nil
}

func TestNewDataAggregator(t *testing.T) {
	config := AggregationConfig{
		GroupingDimensions: []string{"instance_type"},
		StatisticalConfig: StatisticalConfig{
			ConfidenceLevel: 0.95,
			MinSampleSize:   3,
		},
		QualityThreshold: 0.7,
	}

	dataSource := NewMockDataSource()
	aggregator, err := NewDataAggregator(config, dataSource)

	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	if aggregator == nil {
		t.Fatal("Expected non-nil aggregator")
	}

	if aggregator.config.QualityThreshold != 0.7 {
		t.Errorf("Expected quality threshold 0.7, got %f", aggregator.config.QualityThreshold)
	}
}

func TestValidateAggregationConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      AggregationConfig
		expectError bool
	}{
		{
			name: "valid configuration",
			config: AggregationConfig{
				GroupingDimensions: []string{"instance_type"},
				StatisticalConfig: StatisticalConfig{
					ConfidenceLevel: 0.95,
					MinSampleSize:   5,
				},
			},
			expectError: false,
		},
		{
			name: "missing grouping dimensions",
			config: AggregationConfig{
				GroupingDimensions: []string{},
				StatisticalConfig: StatisticalConfig{
					ConfidenceLevel: 0.95,
					MinSampleSize:   5,
				},
			},
			expectError: true,
		},
		{
			name: "invalid confidence level",
			config: AggregationConfig{
				GroupingDimensions: []string{"instance_type"},
				StatisticalConfig: StatisticalConfig{
					ConfidenceLevel: 1.5,
					MinSampleSize:   5,
				},
			},
			expectError: true,
		},
		{
			name: "insufficient sample size",
			config: AggregationConfig{
				GroupingDimensions: []string{"instance_type"},
				StatisticalConfig: StatisticalConfig{
					ConfidenceLevel: 0.95,
					MinSampleSize:   2,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAggregationConfig(tc.config)
			if tc.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestCreateAggregationKey(t *testing.T) {
	config := AggregationConfig{
		GroupingDimensions: []string{"instance_type", "region"},
		StatisticalConfig: StatisticalConfig{
			ConfidenceLevel: 0.95,
			MinSampleSize:   3,
		},
	}
	dataSource := NewMockDataSource()
	aggregator, err := NewDataAggregator(config, dataSource)
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	metadata := ResultMetadata{
		InstanceType: "m7i.large",
		Region:       "us-east-1",
		BenchmarkSuite: "stream",
	}

	key := aggregator.createAggregationKey(metadata)

	expectedDimensions := map[string]string{
		"instance_type": "m7i.large",
		"region":        "us-east-1",
	}

	if len(key.Dimensions) != 2 {
		t.Errorf("Expected 2 dimensions, got %d", len(key.Dimensions))
	}

	for k, v := range expectedDimensions {
		if key.Dimensions[k] != v {
			t.Errorf("Expected %s=%s, got %s=%s", k, v, k, key.Dimensions[k])
		}
	}

	if key.Hash == "" {
		t.Error("Expected non-empty hash")
	}
}

func TestFilterByQuality(t *testing.T) {
	config := AggregationConfig{
		GroupingDimensions: []string{"instance_type"},
		QualityThreshold: 0.8,
		StatisticalConfig: StatisticalConfig{
			ConfidenceLevel: 0.95,
			MinSampleSize:   3,
		},
	}
	dataSource := NewMockDataSource()
	aggregator, err := NewDataAggregator(config, dataSource)
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	metadata := []ResultMetadata{
		{ResultID: "1", QualityScore: 0.9},
		{ResultID: "2", QualityScore: 0.7},
		{ResultID: "3", QualityScore: 0.85},
		{ResultID: "4", QualityScore: 0.6},
	}

	filtered := aggregator.filterByQuality(metadata)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered results, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.QualityScore < 0.8 {
			t.Errorf("Filtered result has quality score %f below threshold 0.8", result.QualityScore)
		}
	}
}

func TestCalculateMean(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	testCases := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "simple average",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "single value",
			values:   []float64{42.0},
			expected: 42.0,
		},
		{
			name:     "empty slice",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "negative values",
			values:   []float64{-1.0, -2.0, -3.0},
			expected: -2.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := aggregator.calculateMean(tc.values)
			if result != tc.expected {
				t.Errorf("Expected mean %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestCalculateStandardDeviation(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	testCases := []struct {
		name      string
		values    []float64
		mean      float64
		expected  float64
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
			result := aggregator.calculateStandardDeviation(tc.values, tc.mean)
			if abs(result-tc.expected) > tc.tolerance {
				t.Errorf("Expected standard deviation %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	testCases := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "odd number of values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "even number of values",
			values:   []float64{1.0, 2.0, 3.0, 4.0},
			expected: 2.5,
		},
		{
			name:     "single value",
			values:   []float64{42.0},
			expected: 42.0,
		},
		{
			name:     "empty slice",
			values:   []float64{},
			expected: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := aggregator.calculateMedian(tc.values)
			if result != tc.expected {
				t.Errorf("Expected median %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestCalculatePercentile(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}

	testCases := []struct {
		percentile float64
		expected   float64
		tolerance  float64
	}{
		{percentile: 0, expected: 1.0, tolerance: 0.01},
		{percentile: 50, expected: 5.5, tolerance: 0.01},
		{percentile: 100, expected: 10.0, tolerance: 0.01},
		{percentile: 25, expected: 3.25, tolerance: 0.01},
		{percentile: 75, expected: 7.75, tolerance: 0.01},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("P%.0f", tc.percentile), func(t *testing.T) {
			result := aggregator.calculatePercentile(values, tc.percentile)
			if abs(result-tc.expected) > tc.tolerance {
				t.Errorf("Expected P%.0f = %f, got %f", tc.percentile, tc.expected, result)
			}
		})
	}
}

func TestAggregateMeasurement(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	values := []float64{45.0, 45.2, 44.8, 45.1, 44.9, 45.3, 45.0, 44.7}

	result := aggregator.aggregateMeasurement(values)

	if result.Count != len(values) {
		t.Errorf("Expected count %d, got %d", len(values), result.Count)
	}

	if result.Mean <= 0 {
		t.Error("Expected positive mean")
	}

	if result.StandardDeviation < 0 {
		t.Error("Expected non-negative standard deviation")
	}

	if result.Min > result.Max {
		t.Error("Expected min <= max")
	}

	if result.Median < result.Min || result.Median > result.Max {
		t.Error("Expected median between min and max")
	}

	// Check percentiles
	if result.Percentiles["P25"] > result.Percentiles["P75"] {
		t.Error("Expected P25 <= P75")
	}

	if result.Percentiles["P5"] > result.Percentiles["P95"] {
		t.Error("Expected P5 <= P95")
	}
}

func TestAggregateStreamData(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	// Create mock STREAM benchmark results
	streamResults := []*benchmarks.BenchmarkResult{
		{
			BenchmarkSuite: "stream",
			Measurements: map[string]benchmarks.Measurement{
				"triad": {Value: 45.0, Unit: "GB/s"},
				"copy":  {Value: 46.0, Unit: "GB/s"},
				"scale": {Value: 44.0, Unit: "GB/s"},
				"add":   {Value: 43.0, Unit: "GB/s"},
			},
		},
		{
			BenchmarkSuite: "stream",
			Measurements: map[string]benchmarks.Measurement{
				"triad": {Value: 45.2, Unit: "GB/s"},
				"copy":  {Value: 46.1, Unit: "GB/s"},
				"scale": {Value: 44.1, Unit: "GB/s"},
				"add":   {Value: 43.2, Unit: "GB/s"},
			},
		},
	}

	result := aggregator.aggregateStreamData(streamResults)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify aggregated measurements exist
	if result.TriadBandwidth.Count != 2 {
		t.Errorf("Expected 2 triad measurements, got %d", result.TriadBandwidth.Count)
	}

	if result.CopyBandwidth.Count != 2 {
		t.Errorf("Expected 2 copy measurements, got %d", result.CopyBandwidth.Count)
	}

	// Verify means are reasonable
	if result.TriadBandwidth.Mean < 44.0 || result.TriadBandwidth.Mean > 46.0 {
		t.Errorf("Unexpected triad bandwidth mean: %f", result.TriadBandwidth.Mean)
	}

	// Verify overall stability is calculated
	if result.OverallStability < 0 {
		t.Errorf("Expected non-negative overall stability, got %f", result.OverallStability)
	}
}

func TestAggregateHPLData(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	// Create mock HPL benchmark results
	hplResults := []*benchmarks.HPLResult{
		{
			BenchmarkSuite: "hpl",
			Performance: benchmarks.HPLPerformance{
				GFLOPS:        benchmarks.Measurement{Value: 200.0},
				Efficiency:    benchmarks.Measurement{Value: 0.85},
				ExecutionTime: benchmarks.Measurement{Value: 120.0},
				Residual:      benchmarks.Measurement{Value: 1e-12},
			},
		},
		{
			BenchmarkSuite: "hpl",
			Performance: benchmarks.HPLPerformance{
				GFLOPS:        benchmarks.Measurement{Value: 202.0},
				Efficiency:    benchmarks.Measurement{Value: 0.87},
				ExecutionTime: benchmarks.Measurement{Value: 118.0},
				Residual:      benchmarks.Measurement{Value: 1e-13},
			},
		},
	}

	result := aggregator.aggregateHPLData(hplResults)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify aggregated measurements exist
	if result.GFLOPS.Count != 2 {
		t.Errorf("Expected 2 GFLOPS measurements, got %d", result.GFLOPS.Count)
	}

	if result.Efficiency.Count != 2 {
		t.Errorf("Expected 2 efficiency measurements, got %d", result.Efficiency.Count)
	}

	// Verify means are reasonable
	if result.GFLOPS.Mean < 200.0 || result.GFLOPS.Mean > 205.0 {
		t.Errorf("Unexpected GFLOPS mean: %f", result.GFLOPS.Mean)
	}

	if result.Efficiency.Mean < 0.8 || result.Efficiency.Mean > 0.9 {
		t.Errorf("Unexpected efficiency mean: %f", result.Efficiency.Mean)
	}
}

func TestProcessBenchmarkDataInsufficientData(t *testing.T) {
	config := AggregationConfig{
		GroupingDimensions: []string{"instance_type"},
		StatisticalConfig: StatisticalConfig{
			ConfidenceLevel: 0.95,
			MinSampleSize:   5, // Require 5 samples
		},
		QualityThreshold: 0.7,
	}

	dataSource := NewMockDataSource()
	aggregator, err := NewDataAggregator(config, dataSource)
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	// Add insufficient data (only 2 results)
	for i := 0; i < 2; i++ {
		metadata := ResultMetadata{
			ResultID:     fmt.Sprintf("result-%d", i),
			InstanceType: "m7i.large",
			QualityScore: 0.8,
		}
		data := BenchmarkData{Metadata: metadata}
		dataSource.AddResult(metadata, data)
	}

	ctx := context.Background()
	_, err = aggregator.ProcessBenchmarkData(ctx)

	if err == nil {
		t.Error("Expected insufficient data error")
	}

	if !errors.Is(err, ErrInsufficientData) {
		t.Errorf("Expected ErrInsufficientData, got %v", err)
	}
}

func TestCalculateTimeRange(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	baseTime := time.Now()
	data := []BenchmarkData{
		{Metadata: ResultMetadata{Timestamp: baseTime}},
		{Metadata: ResultMetadata{Timestamp: baseTime.Add(time.Hour)}},
		{Metadata: ResultMetadata{Timestamp: baseTime.Add(-time.Hour)}},
	}

	timeRange := aggregator.calculateTimeRange(data)

	expectedStart := baseTime.Add(-time.Hour)
	expectedEnd := baseTime.Add(time.Hour)

	if !timeRange.Start.Equal(expectedStart) {
		t.Errorf("Expected start time %v, got %v", expectedStart, timeRange.Start)
	}

	if !timeRange.End.Equal(expectedEnd) {
		t.Errorf("Expected end time %v, got %v", expectedEnd, timeRange.End)
	}
}

func TestAssessDataQuality(t *testing.T) {
	aggregator, err := createTestAggregator()
	if err != nil {
		t.Fatalf("Failed to create aggregator: %v", err)
	}

	data := []BenchmarkData{
		{Metadata: ResultMetadata{QualityScore: 0.9}},
		{Metadata: ResultMetadata{QualityScore: 0.8}},
		{Metadata: ResultMetadata{QualityScore: 0.85}},
	}

	assessment := aggregator.assessDataQuality(data)

	if assessment.OverallScore <= 0 {
		t.Error("Expected positive overall score")
	}

	if assessment.OverallScore > 1.0 {
		t.Error("Expected overall score <= 1.0")
	}

	if assessment.StatisticalConfidence < 0.8 {
		t.Errorf("Expected statistical confidence >= 0.8, got %f", assessment.StatisticalConfidence)
	}
}

// Helper function for floating point comparison.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}