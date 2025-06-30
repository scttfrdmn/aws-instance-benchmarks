package main

import (
	"context"
	"fmt"
	"log"
	"time"

	awspkg "github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
)

// Async benchmark collector - checks S3 for completed benchmarks
func main() {
	fmt.Println("🔍 ASYNC BENCHMARK COLLECTOR")
	fmt.Println("=============================")
	fmt.Println("Checking S3 for completed benchmark results")
	fmt.Println("=============================")

	ctx := context.Background()

	// Initialize collector
	collector, err := awspkg.NewAsyncCollector("us-west-2")
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	// S3 bucket where results are stored
	s3Bucket := "aws-benchmark-results-bucket" // Replace with your S3 bucket

	fmt.Printf("\n🔍 Scanning S3 bucket: %s\n", s3Bucket)
	fmt.Printf("   Region: us-west-2\n")
	fmt.Printf("   Looking for benchmark results...\n\n")

	// Check all benchmarks
	results, err := collector.CheckAllBenchmarks(ctx, s3Bucket)
	if err != nil {
		log.Fatalf("Failed to check benchmarks: %v", err)
	}

	// Display detailed results
	fmt.Printf("\n📊 DETAILED RESULTS\n")
	fmt.Printf("===================\n")

	if len(results.Completed) > 0 {
		fmt.Printf("\n✅ COMPLETED BENCHMARKS (%d):\n", len(results.Completed))
		fmt.Printf("================================\n")
		for i, result := range results.Completed {
			fmt.Printf("%d. %s on %s\n", i+1, 
				result.Job.BenchmarkConfig.BenchmarkSuite,
				result.Job.BenchmarkConfig.InstanceType)
			fmt.Printf("   Benchmark ID: %s\n", result.Job.BenchmarkID)
			fmt.Printf("   Instance ID: %s\n", result.Job.InstanceID)
			fmt.Printf("   Execution Time: %v\n", result.ExecutionTime)
			fmt.Printf("   Cost: $%.4f\n", result.Job.EstimatedCost)
			
			// Display key results
			if benchmarkData, ok := result.BenchmarkData["results"].(map[string]interface{}); ok {
				fmt.Printf("   Results:\n")
				for key, value := range benchmarkData {
					fmt.Printf("     %s: %v\n", key, value)
				}
			}
			fmt.Printf("\n")
		}
	}

	if len(results.Failed) > 0 {
		fmt.Printf("❌ FAILED BENCHMARKS (%d):\n", len(results.Failed))
		fmt.Printf("=============================\n")
		for i, result := range results.Failed {
			fmt.Printf("%d. %s on %s\n", i+1,
				result.Job.BenchmarkConfig.BenchmarkSuite,
				result.Job.BenchmarkConfig.InstanceType)
			fmt.Printf("   Benchmark ID: %s\n", result.Job.BenchmarkID)
			fmt.Printf("   Error: %s\n", result.Error)
			fmt.Printf("   Cost: $%.4f\n", result.Job.EstimatedCost)
			fmt.Printf("\n")
		}
	}

	if len(results.InProgress) > 0 {
		fmt.Printf("🔄 IN PROGRESS BENCHMARKS (%d):\n", len(results.InProgress))
		fmt.Printf("=================================\n")
		for i, job := range results.InProgress {
			fmt.Printf("%d. %s on %s\n", i+1,
				job.BenchmarkConfig.BenchmarkSuite,
				job.BenchmarkConfig.InstanceType)
			fmt.Printf("   Benchmark ID: %s\n", job.BenchmarkID)
			fmt.Printf("   Instance ID: %s\n", job.InstanceID)
			fmt.Printf("   Running since: %s\n", job.LaunchedAt.Format("15:04:05"))
			fmt.Printf("   Est. cost so far: $%.4f\n", job.EstimatedCost)
			fmt.Printf("\n")
		}
	}

	if len(results.TimedOut) > 0 {
		fmt.Printf("⏰ TIMED OUT BENCHMARKS (%d):\n", len(results.TimedOut))
		fmt.Printf("==============================\n")
		for i, job := range results.TimedOut {
			fmt.Printf("%d. %s on %s\n", i+1,
				job.BenchmarkConfig.BenchmarkSuite,
				job.BenchmarkConfig.InstanceType)
			fmt.Printf("   Benchmark ID: %s\n", job.BenchmarkID)
			fmt.Printf("   Instance ID: %s\n", job.InstanceID)
			fmt.Printf("   ⚠️  REQUIRES MANUAL CLEANUP\n")
			fmt.Printf("\n")
		}
	}

	// Performance analysis for completed benchmarks
	if len(results.Completed) > 0 {
		fmt.Printf("🏆 PERFORMANCE ANALYSIS\n")
		fmt.Printf("=======================\n")
		
		performanceData := analyzePerformance(results.Completed)
		for arch, data := range performanceData {
			fmt.Printf("🔬 %s Architecture:\n", arch)
			for benchmark, perf := range data {
				fmt.Printf("   %s: %s\n", benchmark, perf)
			}
			fmt.Printf("\n")
		}
	}

	// Recommendations
	fmt.Printf("📋 RECOMMENDATIONS\n")
	fmt.Printf("==================\n")
	
	if len(results.InProgress) > 0 {
		fmt.Printf("🔄 %d benchmarks still running\n", len(results.InProgress))
		fmt.Printf("   Run collector again later to check progress\n")
		fmt.Printf("   Monitor instances in AWS console\n")
	}
	
	if len(results.TimedOut) > 0 {
		fmt.Printf("⚠️  %d benchmarks timed out\n", len(results.TimedOut))
		fmt.Printf("   Check instances manually in AWS console\n")
		fmt.Printf("   May need manual termination\n")
	}
	
	if results.Summary.SuccessRate < 80 {
		fmt.Printf("⚠️  Low success rate (%.1f%%)\n", results.Summary.SuccessRate)
		fmt.Printf("   Check AWS quotas and permissions\n")
		fmt.Printf("   Verify instance types available in region\n")
	}

	fmt.Printf("\n🎯 COLLECTION COMPLETE\n")
	fmt.Printf("======================\n")
	fmt.Printf("   Total jobs checked: %d\n", results.Summary.TotalJobs)
	fmt.Printf("   Success rate: %.1f%%\n", results.Summary.SuccessRate)
	fmt.Printf("   Total cost: $%.4f\n", results.Summary.TotalCost)
	
	if len(results.Completed) > 0 {
		fmt.Printf("   ✅ Results ready for analysis\n")
		fmt.Printf("   📊 Data available for ComputeCompass integration\n")
	}
	
	if len(results.InProgress) > 0 {
		fmt.Printf("   🔄 Check again later for remaining results\n")
	}
}

// analyzePerformance provides basic performance analysis
func analyzePerformance(results []*awspkg.AsyncBenchmarkResult) map[string]map[string]string {
	performance := make(map[string]map[string]string)
	
	for _, result := range results {
		instanceType := result.Job.BenchmarkConfig.InstanceType
		benchmarkSuite := result.Job.BenchmarkConfig.BenchmarkSuite
		
		// Determine architecture
		var arch string
		if contains(instanceType, "c7g") {
			arch = "ARM Graviton3"
		} else if contains(instanceType, "c7i") {
			arch = "Intel Ice Lake"
		} else if contains(instanceType, "c7a") {
			arch = "AMD EPYC"
		} else {
			arch = "Unknown"
		}
		
		if performance[arch] == nil {
			performance[arch] = make(map[string]string)
		}
		
		// Extract performance metrics
		if benchmarkData, ok := result.BenchmarkData["results"].(map[string]interface{}); ok {
			switch benchmarkSuite {
			case "stream":
				if triad, ok := benchmarkData["triad_bandwidth_mbps"].(float64); ok {
					performance[arch]["STREAM Triad"] = fmt.Sprintf("%.1f MB/s", triad)
				}
			case "hpl":
				if gflops, ok := benchmarkData["peak_gflops"].(float64); ok {
					performance[arch]["HPL LINPACK"] = fmt.Sprintf("%.1f GFLOPS", gflops)
				}
			case "fftw":
				performance[arch]["FFTW Scientific"] = "Completed successfully"
			}
		}
	}
	
	return performance
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}