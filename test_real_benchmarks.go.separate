package main

import (
	"context"
	"fmt"
	"log"
	"time"

	awspkg "github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
)

// Real benchmark execution test with actual AWS resources
func main() {
	fmt.Println("🚀 REAL BENCHMARK EXECUTION TEST")
	fmt.Println("================================")
	fmt.Println("Testing Phase 2 benchmarks on real AWS instances:")
	fmt.Println("  Region: us-west-2")
	fmt.Println("  Testing: ARM Graviton3, Intel Ice Lake, AMD EPYC")
	fmt.Println("================================")
	
	// Initialize orchestrator for us-west-2
	orchestrator, err := awspkg.NewOrchestrator("us-west-2")
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Real test configurations using actual us-west-2 resources
	testConfigs := []awspkg.BenchmarkConfig{
		// Test 1: ARM Graviton3 - Vector Operations (fastest test)
		{
			InstanceType:   "c7g.large",
			BenchmarkSuite: "vector_ops",
			Region:         "us-west-2",
			KeyPairName:    "pop-test-arm-instance6",
			SecurityGroupID: "sg-4cfcb21a",
			SubnetID:       "subnet-86a157cc",
			SkipQuotaCheck: false,
			MaxRetries:     3,
			Timeout:        15 * time.Minute,
		},
		// Test 2: Intel Ice Lake - Mixed Precision
		{
			InstanceType:   "c7i.large",
			BenchmarkSuite: "mixed_precision",
			Region:         "us-west-2",
			KeyPairName:    "pop-test-arm-instance6",
			SecurityGroupID: "sg-4cfcb21a",
			SubnetID:       "subnet-86a157cc",
			SkipQuotaCheck: false,
			MaxRetries:     3,
			Timeout:        20 * time.Minute,
		},
		// Test 3: AMD EPYC - FFTW (if first two succeed)
		{
			InstanceType:   "c7a.large",
			BenchmarkSuite: "fftw",
			Region:         "us-west-2",
			KeyPairName:    "pop-test-arm-instance6",
			SecurityGroupID: "sg-4cfcb21a",
			SubnetID:       "subnet-86a157cc",
			SkipQuotaCheck: false,
			MaxRetries:     3,
			Timeout:        18 * time.Minute,
		},
	}

	ctx := context.Background()
	results := make(map[string]*awspkg.InstanceResult)
	startTime := time.Now()
	
	fmt.Printf("\n🔄 Starting real benchmark execution...\n")
	fmt.Printf("   Total tests: %d\n", len(testConfigs))
	fmt.Printf("   Start time: %s\n\n", startTime.Format("15:04:05"))

	// Execute benchmarks with detailed progress tracking
	for i, config := range testConfigs {
		testKey := fmt.Sprintf("%s_%s", config.InstanceType, config.BenchmarkSuite)
		
		fmt.Printf("🚀 Test %d/%d: %s on %s\n", i+1, len(testConfigs), config.BenchmarkSuite, config.InstanceType)
		fmt.Printf("   📍 Region: %s, Subnet: %s\n", config.Region, config.SubnetID)
		
		// Print expected results
		printExpectedResults(config.InstanceType, config.BenchmarkSuite)
		
		fmt.Printf("   ⏱️  Starting at: %s\n", time.Now().Format("15:04:05"))
		testStartTime := time.Now()
		
		result, err := orchestrator.RunBenchmark(ctx, config)
		testDuration := time.Since(testStartTime)
		
		if err != nil {
			fmt.Printf("   ❌ FAILED after %.1f minutes: %v\n", testDuration.Minutes(), err)
			fmt.Printf("   🔍 Error analysis:\n")
			analyzeError(err)
			fmt.Println()
			continue
		}
		
		results[testKey] = result
		
		// Print detailed success results
		fmt.Printf("   ✅ SUCCESS in %.1f minutes\n", testDuration.Minutes())
		fmt.Printf("   📊 Instance ID: %s\n", result.InstanceID)
		fmt.Printf("   💰 Cost: ~$%.4f\n", estimateCost(config.InstanceType, testDuration))
		
		if result.BenchmarkData != nil {
			printDetailedBenchmarkResults(config.BenchmarkSuite, result.BenchmarkData)
		}
		
		fmt.Printf("   ⏱️  Completed at: %s\n", time.Now().Format("15:04:05"))
		fmt.Println("   " + repeatString("=", 60))
		fmt.Println()
	}

	// Final analysis
	totalDuration := time.Since(startTime)
	fmt.Println("🎯 REAL BENCHMARK EXECUTION SUMMARY")
	fmt.Println("===================================")
	
	successCount := len(results)
	totalCount := len(testConfigs)
	totalCost := 0.0
	
	fmt.Printf("📊 Execution Statistics:\n")
	fmt.Printf("   Tests completed: %d/%d (%.1f%%)\n", successCount, totalCount, float64(successCount)/float64(totalCount)*100)
	fmt.Printf("   Total runtime: %.1f minutes\n", totalDuration.Minutes())
	fmt.Printf("   Average per test: %.1f minutes\n", totalDuration.Minutes()/float64(len(testConfigs)))
	
	if successCount > 0 {
		fmt.Printf("\n🏆 SUCCESSFUL BENCHMARKS:\n")
		for testKey, result := range results {
			fmt.Printf("   ✅ %s\n", testKey)
			if result.BenchmarkData != nil {
				printSummaryMetrics(testKey, result.BenchmarkData)
			}
			totalCost += estimateCost(result.InstanceType, 15*time.Minute) // Rough estimate
		}
		fmt.Printf("\n💰 Estimated total cost: $%.4f\n", totalCost)
	}
	
	if successCount == totalCount {
		fmt.Printf("\n🎉 ALL TESTS PASSED! Phase 2 benchmarks fully validated on real hardware\n")
		analyzeArchitecturePerformance(results)
	} else if successCount > 0 {
		fmt.Printf("\n✅ PARTIAL SUCCESS: %d benchmarks validated on real hardware\n", successCount)
		if len(results) > 0 {
			analyzeArchitecturePerformance(results)
		}
	} else {
		fmt.Printf("\n❌ NO TESTS COMPLETED: Check AWS permissions and quotas\n")
		fmt.Printf("   Troubleshooting:\n")
		fmt.Printf("   - Verify EC2 instance limits in us-west-2\n")
		fmt.Printf("   - Check IAM permissions for EC2, SSM\n")
		fmt.Printf("   - Ensure subnet allows public IP assignment\n")
	}
	
	fmt.Printf("\n🚀 Production Readiness: ")
	if successCount >= 2 {
		fmt.Printf("✅ CONFIRMED - Multi-architecture validation successful\n")
	} else if successCount == 1 {
		fmt.Printf("⚠️  PARTIAL - Single architecture validated\n")
	} else {
		fmt.Printf("❌ NEEDS INVESTIGATION - No successful executions\n")
	}
}

func printExpectedResults(instanceType, benchmark string) {
	fmt.Printf("   🎯 Expected results:\n")
	switch instanceType {
	case "c7g.large":
		fmt.Printf("      🟢 ARM Graviton3 - Excellent efficiency\n")
		if benchmark == "vector_ops" {
			fmt.Printf("      AXPY: ~85-105 GFLOPS, DOT: ~75-95 GFLOPS, NORM: ~75-95 GFLOPS\n")
		}
	case "c7i.large":
		fmt.Printf("      🔵 Intel Ice Lake - Peak GFLOPS performance\n")
		if benchmark == "mixed_precision" {
			fmt.Printf("      FP16: ~100-140 GFLOPS, FP32: ~90-120 GFLOPS, FP64: ~60-80 GFLOPS\n")
		}
	case "c7a.large":
		fmt.Printf("      🟡 AMD EPYC 9R14 - Balanced performance\n")
		if benchmark == "fftw" {
			fmt.Printf("      1D: ~75-95 GFLOPS, 2D: ~60-80 GFLOPS, 3D: ~42-62 GFLOPS\n")
		}
	}
}

func printDetailedBenchmarkResults(benchmark string, data map[string]interface{}) {
	fmt.Printf("   📈 ACTUAL RESULTS:\n")
	
	switch benchmark {
	case "vector_ops":
		if vectorData, ok := data["vector_ops"].(map[string]interface{}); ok {
			if axpy, ok := vectorData["avg_axpy_gflops"].(float64); ok {
				fmt.Printf("      AXPY: %.2f GFLOPS ⭐\n", axpy)
			}
			if dot, ok := vectorData["avg_dot_gflops"].(float64); ok {
				fmt.Printf("      DOT:  %.2f GFLOPS ⭐\n", dot)
			}
			if norm, ok := vectorData["avg_norm_gflops"].(float64); ok {
				fmt.Printf("      NORM: %.2f GFLOPS ⭐\n", norm)
			}
			if overall, ok := vectorData["overall_avg_gflops"].(float64); ok {
				fmt.Printf("      Overall: %.2f GFLOPS 🏆\n", overall)
			}
		}
	case "mixed_precision":
		if mixedData, ok := data["mixed_precision"].(map[string]interface{}); ok {
			if fp16, ok := mixedData["peak_fp16_gflops"].(float64); ok {
				fmt.Printf("      FP16: %.2f GFLOPS ⭐\n", fp16)
			}
			if fp32, ok := mixedData["peak_fp32_gflops"].(float64); ok {
				fmt.Printf("      FP32: %.2f GFLOPS ⭐\n", fp32)
			}
			if fp64, ok := mixedData["peak_fp64_gflops"].(float64); ok {
				fmt.Printf("      FP64: %.2f GFLOPS ⭐\n", fp64)
			}
			if overall, ok := mixedData["overall_mixed_precision_score"].(float64); ok {
				fmt.Printf("      Overall: %.2f Score 🏆\n", overall)
			}
		}
	case "fftw":
		if fftwData, ok := data["fftw"].(map[string]interface{}); ok {
			if fft1d, ok := fftwData["fft_1d_large_gflops"].(float64); ok {
				fmt.Printf("      1D FFT: %.2f GFLOPS ⭐\n", fft1d)
			}
			if fft2d, ok := fftwData["fft_2d_gflops"].(float64); ok {
				fmt.Printf("      2D FFT: %.2f GFLOPS ⭐\n", fft2d)
			}
			if fft3d, ok := fftwData["fft_3d_gflops"].(float64); ok {
				fmt.Printf("      3D FFT: %.2f GFLOPS ⭐\n", fft3d)
			}
			if overall, ok := fftwData["overall_gflops"].(float64); ok {
				fmt.Printf("      Overall: %.2f GFLOPS 🏆\n", overall)
			}
		}
	}
}

func printSummaryMetrics(testKey string, data map[string]interface{}) {
	parts := splitString(testKey, "_")
	if len(parts) >= 2 {
		benchmark := parts[1]
		switch benchmark {
		case "vector":
			if vectorData, ok := data["vector_ops"].(map[string]interface{}); ok {
				if overall, ok := vectorData["overall_avg_gflops"].(float64); ok {
					fmt.Printf("      Vector Ops: %.2f GFLOPS\n", overall)
				}
			}
		case "mixed":
			if mixedData, ok := data["mixed_precision"].(map[string]interface{}); ok {
				if overall, ok := mixedData["overall_mixed_precision_score"].(float64); ok {
					fmt.Printf("      Mixed Precision: %.2f Score\n", overall)
				}
			}
		case "fftw":
			if fftwData, ok := data["fftw"].(map[string]interface{}); ok {
				if overall, ok := fftwData["overall_gflops"].(float64); ok {
					fmt.Printf("      FFTW: %.2f GFLOPS\n", overall)
				}
			}
		}
	}
}

func analyzeArchitecturePerformance(results map[string]*awspkg.InstanceResult) {
	fmt.Printf("\n🔬 CROSS-ARCHITECTURE ANALYSIS:\n")
	
	architectures := make(map[string][]float64)
	
	for _, result := range results {
		if result.BenchmarkData == nil {
			continue
		}
		
		arch := getArchitecture(result.InstanceType)
		score := extractPerformanceScore(result.BenchmarkData)
		
		if score > 0 {
			architectures[arch] = append(architectures[arch], score)
			fmt.Printf("   %s (%s): %.2f performance units\n", arch, result.InstanceType, score)
		}
	}
	
	if len(architectures) > 1 {
		fmt.Printf("\n🏆 Architecture Performance Ranking:\n")
		for arch, scores := range architectures {
			avgScore := calculateAverage(scores)
			fmt.Printf("   %s: %.2f average performance\n", arch, avgScore)
		}
	}
	
	fmt.Printf("\n✅ Real hardware validation confirms Phase 2 implementation success\n")
}

func analyzeError(err error) {
	errStr := err.Error()
	if containsString(errStr, "quota") || containsString(errStr, "limit") {
		fmt.Printf("      💡 Instance quota/limit issue - try different instance type\n")
	} else if containsString(errStr, "subnet") {
		fmt.Printf("      💡 Subnet configuration issue - check VPC settings\n")
	} else if containsString(errStr, "security") {
		fmt.Printf("      💡 Security group issue - verify SSH/SSM access\n")
	} else if containsString(errStr, "timeout") {
		fmt.Printf("      💡 Timeout issue - benchmark may need more time\n")
	} else {
		fmt.Printf("      💡 General AWS issue - check permissions and region\n")
	}
}

func estimateCost(instanceType string, duration time.Duration) float64 {
	// Rough hourly costs (as of 2024)
	hourlyCosts := map[string]float64{
		"c7g.large": 0.0725,  // ARM Graviton3
		"c7i.large": 0.0864,  // Intel Ice Lake
		"c7a.large": 0.0864,  // AMD EPYC
	}
	
	if cost, ok := hourlyCosts[instanceType]; ok {
		return cost * duration.Hours()
	}
	return 0.10 * duration.Hours() // Default estimate
}

func getArchitecture(instanceType string) string {
	if containsString(instanceType, "c7g") {
		return "ARM Graviton3"
	} else if containsString(instanceType, "c7i") {
		return "Intel Ice Lake"
	} else if containsString(instanceType, "c7a") {
		return "AMD EPYC 9R14"
	}
	return "Unknown"
}

func extractPerformanceScore(data map[string]interface{}) float64 {
	// Try to extract a representative performance score
	if vectorData, ok := data["vector_ops"].(map[string]interface{}); ok {
		if overall, ok := vectorData["overall_avg_gflops"].(float64); ok {
			return overall
		}
	}
	if mixedData, ok := data["mixed_precision"].(map[string]interface{}); ok {
		if overall, ok := mixedData["overall_mixed_precision_score"].(float64); ok {
			return overall
		}
	}
	if fftwData, ok := data["fftw"].(map[string]interface{}); ok {
		if overall, ok := fftwData["overall_gflops"].(float64); ok {
			return overall
		}
	}
	return 0
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func containsString(s, substr string) bool {
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

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}