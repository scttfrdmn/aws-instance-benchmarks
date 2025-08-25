package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/pricing"
)

// BenchmarkResult represents a parsed benchmark result file
type BenchmarkResult struct {
	InstanceType string
	Region       string
	FilePath     string
	Performance  struct {
		Memory struct {
			Stream *struct {
				Copy struct {
					Bandwidth float64 `json:"bandwidth"`
				} `json:"copy"`
				Scale struct {
					Bandwidth float64 `json:"bandwidth"`
				} `json:"scale"`
				Add struct {
					Bandwidth float64 `json:"bandwidth"`
				} `json:"add"`
				Triad struct {
					Bandwidth float64 `json:"bandwidth"`
				} `json:"triad"`
			} `json:"stream"`
			CoreMark *struct {
				Score float64 `json:"score"`
			} `json:"coremark"`
			HPL *struct {
				GFLOPS float64 `json:"gflops"`
			} `json:"hpl"`
		} `json:"memory"`
	} `json:"performance"`
	Metadata struct {
		InstanceType string `json:"instanceType"`
		Region       string `json:"region"`
	} `json:"metadata"`
}

// PricePerformanceAnalysis contains the analysis results
type PricePerformanceAnalysis struct {
	GeneratedAt time.Time                        `json:"generated_at"`
	Region      string                           `json:"region"`
	Summary     PricePerformanceSummary         `json:"summary"`
	Details     []InstancePricePerformance      `json:"instance_details"`
	Rankings    PricePerformanceRankings        `json:"rankings"`
}

type PricePerformanceSummary struct {
	TotalInstances      int     `json:"total_instances"`
	BestValueInstance   string  `json:"best_value_instance"`
	BestValueScore      float64 `json:"best_value_score"`
	WorstValueInstance  string  `json:"worst_value_instance"`
	WorstValueScore     float64 `json:"worst_value_score"`
	AverageCostPerGBps  float64 `json:"average_cost_per_gbps"`
}

type InstancePricePerformance struct {
	InstanceType        string  `json:"instance_type"`
	HourlyPrice         float64 `json:"hourly_price"`
	TriadBandwidth      float64 `json:"triad_bandwidth_gbps"`
	CostPerGBps         float64 `json:"cost_per_gbps_per_hour"`
	CoreMarkScore       float64 `json:"coremark_score,omitempty"`
	CostPerMOps         float64 `json:"cost_per_mops_per_hour,omitempty"`
	ValueScore          float64 `json:"value_score"`
	EfficiencyRating    string  `json:"efficiency_rating"`
}

type PricePerformanceRankings struct {
	ByMemoryEfficiency []InstanceRanking `json:"by_memory_efficiency"`
	ByComputeEfficiency []InstanceRanking `json:"by_compute_efficiency"`
	ByOverallValue     []InstanceRanking `json:"by_overall_value"`
}

type InstanceRanking struct {
	Rank         int     `json:"rank"`
	InstanceType string  `json:"instance_type"`
	Score        float64 `json:"score"`
	Metric       string  `json:"metric"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/analyze_price_performance.go <results_directory>")
		fmt.Println("Example: go run cmd/analyze_price_performance.go results/2025-06-30")
		os.Exit(1)
	}

	resultsDir := os.Args[1]
	
	ctx := context.Background()
	analysis, err := analyzePricePerformance(ctx, resultsDir)
	if err != nil {
		fmt.Printf("Error analyzing price/performance: %v\n", err)
		os.Exit(1)
	}

	// Output analysis as JSON
	output, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling analysis: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func analyzePricePerformance(ctx context.Context, resultsDir string) (*PricePerformanceAnalysis, error) {
	// Parse all benchmark result files
	results, err := parseBenchmarkResults(resultsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse benchmark results: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no benchmark results found in %s", resultsDir)
	}

	// Create pricing service
	pricingService := pricing.NewPricingService()

	// Calculate price/performance for each instance
	var instanceAnalyses []InstancePricePerformance
	
	for _, result := range results {
		instanceType := result.Metadata.InstanceType
		region := result.Metadata.Region
		
		// Get pricing data
		pricingData, err := pricingService.GetInstancePricing(ctx, instanceType, region)
		if err != nil {
			fmt.Printf("Warning: Could not get pricing for %s: %v\n", instanceType, err)
			continue
		}

		analysis := InstancePricePerformance{
			InstanceType: instanceType,
			HourlyPrice:  pricingData.OnDemand,
		}

		// Extract STREAM performance if available
		if result.Performance.Memory.Stream != nil {
			bandwidth := result.Performance.Memory.Stream.Triad.Bandwidth
			analysis.TriadBandwidth = bandwidth
			analysis.CostPerGBps = pricingData.OnDemand / bandwidth
		}

		// Extract CoreMark performance if available
		if result.Performance.Memory.CoreMark != nil {
			score := result.Performance.Memory.CoreMark.Score
			analysis.CoreMarkScore = score
			scoreMOps := score / 1000000.0 // Convert to millions of ops/sec
			analysis.CostPerMOps = pricingData.OnDemand / scoreMOps
		}

		// Calculate overall value score (lower cost per unit = higher value)
		if analysis.TriadBandwidth > 0 {
			// Inverse of cost per GB/s, normalized to 0-100 scale
			analysis.ValueScore = 100.0 / analysis.CostPerGBps
			analysis.EfficiencyRating = getEfficiencyRating(analysis.CostPerGBps)
		}

		instanceAnalyses = append(instanceAnalyses, analysis)
	}

	if len(instanceAnalyses) == 0 {
		return nil, fmt.Errorf("no valid price/performance analyses could be calculated")
	}

	// Calculate summary statistics
	summary := calculateSummary(instanceAnalyses)

	// Generate rankings
	rankings := generateRankings(instanceAnalyses)

	// Determine region from first result
	region := "unknown"
	if len(results) > 0 {
		region = results[0].Metadata.Region
	}

	return &PricePerformanceAnalysis{
		GeneratedAt: time.Now(),
		Region:      region,
		Summary:     summary,
		Details:     instanceAnalyses,
		Rankings:    rankings,
	}, nil
}

func parseBenchmarkResults(resultsDir string) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	err := filepath.Walk(resultsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var result BenchmarkResult
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		result.FilePath = path
		results = append(results, result)
		return nil
	})

	return results, err
}

func calculateSummary(analyses []InstancePricePerformance) PricePerformanceSummary {
	if len(analyses) == 0 {
		return PricePerformanceSummary{}
	}

	var totalCostPerGBps float64
	var validCount int
	bestValue := analyses[0]
	worstValue := analyses[0]

	for _, analysis := range analyses {
		if analysis.CostPerGBps > 0 {
			totalCostPerGBps += analysis.CostPerGBps
			validCount++

			if analysis.ValueScore > bestValue.ValueScore {
				bestValue = analysis
			}
			if analysis.ValueScore < worstValue.ValueScore {
				worstValue = analysis
			}
		}
	}

	avgCostPerGBps := float64(0)
	if validCount > 0 {
		avgCostPerGBps = totalCostPerGBps / float64(validCount)
	}

	return PricePerformanceSummary{
		TotalInstances:     len(analyses),
		BestValueInstance:  bestValue.InstanceType,
		BestValueScore:     bestValue.ValueScore,
		WorstValueInstance: worstValue.InstanceType,
		WorstValueScore:    worstValue.ValueScore,
		AverageCostPerGBps: avgCostPerGBps,
	}
}

func generateRankings(analyses []InstancePricePerformance) PricePerformanceRankings {
	// Sort by memory efficiency (lower cost per GB/s = better)
	memoryRanking := make([]InstancePricePerformance, len(analyses))
	copy(memoryRanking, analyses)
	sort.Slice(memoryRanking, func(i, j int) bool {
		return memoryRanking[i].CostPerGBps < memoryRanking[j].CostPerGBps
	})

	var memoryRankings []InstanceRanking
	for i, analysis := range memoryRanking {
		if analysis.CostPerGBps > 0 {
			memoryRankings = append(memoryRankings, InstanceRanking{
				Rank:         i + 1,
				InstanceType: analysis.InstanceType,
				Score:        analysis.CostPerGBps,
				Metric:       "Cost per GB/s/hour",
			})
		}
	}

	// Sort by compute efficiency (lower cost per MOps = better)
	computeRanking := make([]InstancePricePerformance, len(analyses))
	copy(computeRanking, analyses)
	sort.Slice(computeRanking, func(i, j int) bool {
		if computeRanking[i].CostPerMOps == 0 && computeRanking[j].CostPerMOps == 0 {
			return false
		}
		if computeRanking[i].CostPerMOps == 0 {
			return false
		}
		if computeRanking[j].CostPerMOps == 0 {
			return true
		}
		return computeRanking[i].CostPerMOps < computeRanking[j].CostPerMOps
	})

	var computeRankings []InstanceRanking
	for i, analysis := range computeRanking {
		if analysis.CostPerMOps > 0 {
			computeRankings = append(computeRankings, InstanceRanking{
				Rank:         i + 1,
				InstanceType: analysis.InstanceType,
				Score:        analysis.CostPerMOps,
				Metric:       "Cost per MOps/s/hour",
			})
		}
	}

	// Sort by overall value (higher value score = better)
	valueRanking := make([]InstancePricePerformance, len(analyses))
	copy(valueRanking, analyses)
	sort.Slice(valueRanking, func(i, j int) bool {
		return valueRanking[i].ValueScore > valueRanking[j].ValueScore
	})

	var valueRankings []InstanceRanking
	for i, analysis := range valueRanking {
		if analysis.ValueScore > 0 {
			valueRankings = append(valueRankings, InstanceRanking{
				Rank:         i + 1,
				InstanceType: analysis.InstanceType,
				Score:        analysis.ValueScore,
				Metric:       "Value Score",
			})
		}
	}

	return PricePerformanceRankings{
		ByMemoryEfficiency:  memoryRankings,
		ByComputeEfficiency: computeRankings,
		ByOverallValue:      valueRankings,
	}
}

func getEfficiencyRating(costPerGBps float64) string {
	if costPerGBps < 0.0015 {
		return "Excellent"
	} else if costPerGBps < 0.002 {
		return "Very Good"
	} else if costPerGBps < 0.003 {
		return "Good"
	} else if costPerGBps < 0.004 {
		return "Fair"
	} else {
		return "Poor"
	}
}