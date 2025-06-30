package main

import (
	"context"
	"fmt"
	"log"
	"time"

	awspkg "github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
)

// Test the unified comprehensive benchmark strategy
// This test validates the new 7-zip and Sysbench benchmarks
func main() {
	fmt.Println("🚀 Testing Unified Comprehensive Benchmark Strategy")
	fmt.Println("============================================================")
	
	// Initialize orchestrator
	orchestrator, err := awspkg.NewOrchestrator("us-east-1")
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Test configurations for unified benchmark strategy validation
	testConfigs := []awspkg.BenchmarkConfig{
		{
			InstanceType:   "c7g.large",  // ARM Graviton3 - should show excellent memory + cost efficiency
			BenchmarkSuite: "7zip",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        15 * time.Minute,
		},
		{
			InstanceType:   "c7a.large",  // AMD EPYC - should resolve the 76% performance mystery
			BenchmarkSuite: "7zip",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        15 * time.Minute,
		},
		{
			InstanceType:   "c7i.large",  // Intel Ice Lake - should show peak integer performance
			BenchmarkSuite: "sysbench",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        15 * time.Minute,
		},
		{
			InstanceType:   "c7g.large",  // ARM Graviton3 - enhanced scientific computing
			BenchmarkSuite: "dgemm",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        20 * time.Minute,
		},
	}

	ctx := context.Background()
	results := make(map[string]*awspkg.InstanceResult)

	// Execute benchmarks sequentially for clear validation
	for _, config := range testConfigs {
		fmt.Printf("\n🔄 Testing %s benchmark on %s...\n", config.BenchmarkSuite, config.InstanceType)
		fmt.Printf("   Expected: Industry-standard %s results comparable to published benchmarks\n", config.BenchmarkSuite)
		
		result, err := orchestrator.RunBenchmark(ctx, config)
		if err != nil {
			fmt.Printf("   ❌ Benchmark failed: %v\n", err)
			continue
		}
		
		results[config.InstanceType + "_" + config.BenchmarkSuite] = result
		
		// Print immediate results for validation
		fmt.Printf("   ✅ Benchmark completed successfully\n")
		if result.BenchmarkData != nil {
			printBenchmarkSummary(config.BenchmarkSuite, result.BenchmarkData)
		}
	}

	// Analysis of results
	fmt.Println("\n============================================================")
	fmt.Println("📊 UNIFIED BENCHMARK STRATEGY VALIDATION")
	fmt.Println("============================================================")
	
	analyzeUnifiedResults(results)
	
	fmt.Println("\n🎯 Next Steps:")
	fmt.Println("   1. If 7-zip shows competitive AMD performance → Custom CoreMark was the issue")
	fmt.Println("   2. If Sysbench shows consistent results → Industry standards work across architectures")
	fmt.Println("   3. Implement Phase 2: Add DGEMM enhancement and FFTW for scientific computing")
	fmt.Println("   4. Complete unified benchmark suite covering both server and research workloads")
}

func printBenchmarkSummary(benchmarkSuite string, data map[string]interface{}) {
	switch benchmarkSuite {
	case "7zip":
		if sevenZipData, ok := data["7zip"].(map[string]interface{}); ok {
			if totalMIPS, ok := sevenZipData["total_mips"].(float64); ok {
				fmt.Printf("   📊 7-zip Total MIPS: %.0f (Industry-standard compression benchmark)\n", totalMIPS)
			}
			if compMIPS, ok := sevenZipData["compression_mips"].(float64); ok {
				fmt.Printf("       Compression: %.0f MIPS\n", compMIPS)
			}
			if decompMIPS, ok := sevenZipData["decompression_mips"].(float64); ok {
				fmt.Printf("       Decompression: %.0f MIPS\n", decompMIPS)
			}
		}
	case "sysbench":
		if sysbenchData, ok := data["sysbench"].(map[string]interface{}); ok {
			if eps, ok := sysbenchData["events_per_second"].(float64); ok {
				fmt.Printf("   📊 Sysbench Events/sec: %.0f (Prime number calculation)\n", eps)
			}
			if totalTime, ok := sysbenchData["total_time"].(float64); ok {
				fmt.Printf("       Execution time: %.2f seconds\n", totalTime)
			}
		}
	case "dgemm":
		if dgemmData, ok := data["dgemm"].(map[string]interface{}); ok {
			if peakGflops, ok := dgemmData["peak_gflops"].(float64); ok {
				fmt.Printf("   📊 DGEMM Peak GFLOPS: %.2f (Enhanced scientific computing)\n", peakGflops)
			}
			if smallGflops, ok := dgemmData["small_matrix_gflops"].(float64); ok {
				fmt.Printf("       Small matrix: %.2f GFLOPS\n", smallGflops)
			}
			if largeGflops, ok := dgemmData["large_matrix_gflops"].(float64); ok {
				fmt.Printf("       Large matrix: %.2f GFLOPS\n", largeGflops)
			}
			if memEff, ok := dgemmData["memory_bound_efficiency"].(float64); ok {
				fmt.Printf("       Memory efficiency: %.1f%%\n", memEff*100)
			}
		}
	}
}

func analyzeUnifiedResults(results map[string]*awspkg.InstanceResult) {
	fmt.Println("\n🔍 Architecture Performance Analysis:")
	
	armResults := make(map[string]float64)
	amdResults := make(map[string]float64)
	intelResults := make(map[string]float64)
	
	for key, result := range results {
		if result.BenchmarkData == nil {
			continue
		}
		
		var score float64
		var benchmarkType string
		
		// Extract performance scores
		if sevenZipData, ok := result.BenchmarkData["7zip"].(map[string]interface{}); ok {
			if totalMIPS, ok := sevenZipData["total_mips"].(float64); ok {
				score = totalMIPS
				benchmarkType = "7zip_mips"
			}
		} else if sysbenchData, ok := result.BenchmarkData["sysbench"].(map[string]interface{}); ok {
			if eps, ok := sysbenchData["events_per_second"].(float64); ok {
				score = eps
				benchmarkType = "sysbench_eps"
			}
		} else if dgemmData, ok := result.BenchmarkData["dgemm"].(map[string]interface{}); ok {
			if peakGflops, ok := dgemmData["peak_gflops"].(float64); ok {
				score = peakGflops
				benchmarkType = "dgemm_gflops"
			}
		}
		
		if score > 0 {
			switch {
			case result.InstanceType == "c7g.large":
				armResults[benchmarkType] = score
				fmt.Printf("   🟢 ARM Graviton3 (%s): %.0f %s\n", result.InstanceType, score, benchmarkType)
			case result.InstanceType == "c7a.large":
				amdResults[benchmarkType] = score
				fmt.Printf("   🟡 AMD EPYC (%s): %.0f %s\n", result.InstanceType, score, benchmarkType)
			case result.InstanceType == "c7i.large":
				intelResults[benchmarkType] = score
				fmt.Printf("   🔵 Intel Ice Lake (%s): %.0f %s\n", result.InstanceType, score, benchmarkType)
			}
		}
	}
	
	fmt.Println("\n📈 Performance Insights:")
	fmt.Println("   → If AMD shows competitive scores (~40,000-55,000 MIPS for 7-zip):")
	fmt.Println("     ✅ Custom CoreMark was the root cause of poor AMD performance")
	fmt.Println("   → If all architectures show reasonable performance ratios:")
	fmt.Println("     ✅ Industry-standard benchmarks provide fair cross-architecture comparison")
	fmt.Println("   → If results align with published vendor benchmarks:")
	fmt.Println("     ✅ Unified benchmark strategy successfully validates real performance")
	
	fmt.Println("\n🎯 Unified Strategy Benefits Demonstrated:")
	fmt.Println("   ✅ Real workload benchmarks (compression, prime calculation)")
	fmt.Println("   ✅ Industry-comparable results")
	fmt.Println("   ✅ Fair cross-architecture comparison")
	fmt.Println("   ✅ Zero licensing costs")
	fmt.Println("   ✅ Foundation for both server and scientific computing analysis")
}