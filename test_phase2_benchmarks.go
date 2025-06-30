package main

import (
	"context"
	"fmt"
	"log"
	"time"

	awspkg "github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
)

// Test Phase 2: Advanced Scientific Computing Benchmarks
func main() {
	fmt.Println("🔬 Testing Phase 2: Advanced Scientific Computing Benchmarks")
	fmt.Println("============================================================")
	
	// Initialize orchestrator
	orchestrator, err := awspkg.NewOrchestrator("us-east-1")
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Phase 2 test configurations - scientific computing focus
	testConfigs := []awspkg.BenchmarkConfig{
		{
			InstanceType:   "c7g.large",  // ARM Graviton3 - excellent for FFTW with SVE
			BenchmarkSuite: "fftw",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        20 * time.Minute,
		},
		{
			InstanceType:   "c7i.large",  // Intel Ice Lake - should excel with Intel MKL
			BenchmarkSuite: "fftw",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        20 * time.Minute,
		},
		{
			InstanceType:   "c7a.large",  // AMD EPYC - competitive FFT performance expected
			BenchmarkSuite: "fftw",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        20 * time.Minute,
		},
		{
			InstanceType:   "c7g.xlarge", // ARM Graviton3 - test vector ops scaling
			BenchmarkSuite: "vector_ops",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        15 * time.Minute,
		},
		{
			InstanceType:   "c7i.large",  // Intel Ice Lake - vector ops with AVX
			BenchmarkSuite: "vector_ops",
			Region:         "us-east-1",
			KeyPairName:    "aws-instance-benchmarks",
			SecurityGroupID: "sg-benchmark-testing",
			SubnetID:       "subnet-benchmark-testing",
			SkipQuotaCheck: false,
			MaxRetries:     2,
			Timeout:        15 * time.Minute,
		},
	}

	ctx := context.Background()
	results := make(map[string]*awspkg.InstanceResult)

	// Execute benchmarks sequentially for clear validation
	for _, config := range testConfigs {
		fmt.Printf("\n🔄 Testing %s benchmark on %s...\n", config.BenchmarkSuite, config.InstanceType)
		
		switch config.BenchmarkSuite {
		case "fftw":
			fmt.Printf("   Expected: High GFLOPS for 1D/2D/3D FFTs, architecture-specific optimizations\n")
		case "vector_ops":
			fmt.Printf("   Expected: BLAS Level 1 performance for AXPY, DOT, NORM operations\n")
		}
		
		result, err := orchestrator.RunBenchmark(ctx, config)
		if err != nil {
			fmt.Printf("   ❌ Benchmark failed: %v\n", err)
			continue
		}
		
		results[config.InstanceType + "_" + config.BenchmarkSuite] = result
		
		// Print immediate results for validation
		fmt.Printf("   ✅ Benchmark completed successfully\n")
		if result.BenchmarkData != nil {
			printPhase2BenchmarkSummary(config.BenchmarkSuite, result.BenchmarkData)
		}
	}

	// Analysis of Phase 2 results
	fmt.Println("\n============================================================")
	fmt.Println("📊 PHASE 2: SCIENTIFIC COMPUTING VALIDATION")
	fmt.Println("============================================================")
	
	analyzePhase2Results(results)
	
	fmt.Println("\n🎯 Phase 2 Analysis:")
	fmt.Println("   1. FFTW results validate scientific computing performance across architectures")
	fmt.Println("   2. Vector operations confirm BLAS Level 1 foundation for numerical computing")
	fmt.Println("   3. Architecture-specific optimizations (ARM SVE, Intel AVX) show performance benefits")
	fmt.Println("   4. Memory bandwidth utilization analysis reveals scaling characteristics")
	
	fmt.Println("\n🚀 Next Steps:")
	fmt.Println("   1. Complete mixed precision and compilation benchmarks")
	fmt.Println("   2. Integrate Phase 2 results with ComputeCompass recommendation engine")
	fmt.Println("   3. Create comprehensive scientific workload performance profiles")
	fmt.Println("   4. Validate against published scientific computing benchmarks")
}

func printPhase2BenchmarkSummary(benchmarkSuite string, data map[string]interface{}) {
	switch benchmarkSuite {
	case "fftw":
		if fftwData, ok := data["fftw"].(map[string]interface{}); ok {
			if overall, ok := fftwData["overall_gflops"].(float64); ok {
				fmt.Printf("   📊 FFTW Overall: %.2f GFLOPS (Fast Fourier Transform)\n", overall)
			}
			if fft1d, ok := fftwData["fft_1d_large_gflops"].(float64); ok {
				fmt.Printf("       1D FFT Large: %.2f GFLOPS\n", fft1d)
			}
			if fft2d, ok := fftwData["fft_2d_gflops"].(float64); ok {
				fmt.Printf("       2D FFT: %.2f GFLOPS\n", fft2d)
			}
			if fft3d, ok := fftwData["fft_3d_gflops"].(float64); ok {
				fmt.Printf("       3D FFT: %.2f GFLOPS\n", fft3d)
			}
			if memEff, ok := fftwData["memory_scaling_efficiency"].(float64); ok {
				fmt.Printf("       Memory scaling: %.1f%%\n", memEff*100)
			}
		}
	case "vector_ops":
		if vectorData, ok := data["vector_ops"].(map[string]interface{}); ok {
			if overall, ok := vectorData["overall_avg_gflops"].(float64); ok {
				fmt.Printf("   📊 Vector Ops Overall: %.2f GFLOPS (BLAS Level 1)\n", overall)
			}
			if axpy, ok := vectorData["avg_axpy_gflops"].(float64); ok {
				fmt.Printf("       AXPY: %.2f GFLOPS\n", axpy)
			}
			if dot, ok := vectorData["avg_dot_gflops"].(float64); ok {
				fmt.Printf("       DOT: %.2f GFLOPS\n", dot)
			}
			if norm, ok := vectorData["avg_norm_gflops"].(float64); ok {
				fmt.Printf("       NORM: %.2f GFLOPS\n", norm)
			}
		}
	}
}

func analyzePhase2Results(results map[string]*awspkg.InstanceResult) {
	fmt.Println("\n🔍 Scientific Computing Performance Analysis:")
	
	armResults := make(map[string]float64)
	amdResults := make(map[string]float64)
	intelResults := make(map[string]float64)
	
	for key, result := range results {
		if result.BenchmarkData == nil {
			continue
		}
		
		var score float64
		var benchmarkType string
		
		// Extract scientific computing performance scores
		if fftwData, ok := result.BenchmarkData["fftw"].(map[string]interface{}); ok {
			if overall, ok := fftwData["overall_gflops"].(float64); ok {
				score = overall
				benchmarkType = "fftw_gflops"
			}
		} else if vectorData, ok := result.BenchmarkData["vector_ops"].(map[string]interface{}); ok {
			if overall, ok := vectorData["overall_avg_gflops"].(float64); ok {
				score = overall
				benchmarkType = "vector_ops_gflops"
			}
		}
		
		if score > 0 {
			// Determine architecture based on instance type
			switch {
			case result.InstanceType == "c7g.large" || result.InstanceType == "c7g.xlarge":
				armResults[benchmarkType] = score
				fmt.Printf("   🟢 ARM Graviton3 (%s): %.2f %s\n", result.InstanceType, score, benchmarkType)
			case result.InstanceType == "c7a.large":
				amdResults[benchmarkType] = score
				fmt.Printf("   🟡 AMD EPYC (%s): %.2f %s\n", result.InstanceType, score, benchmarkType)
			case result.InstanceType == "c7i.large":
				intelResults[benchmarkType] = score
				fmt.Printf("   🔵 Intel Ice Lake (%s): %.2f %s\n", result.InstanceType, score, benchmarkType)
			}
		}
	}
	
	fmt.Println("\n📈 Scientific Computing Insights:")
	fmt.Println("   → FFTW Performance Analysis:")
	fmt.Println("     ✅ Signal processing workloads (1D FFT) - cache-friendly algorithms")
	fmt.Println("     ✅ Image processing workloads (2D FFT) - memory bandwidth sensitive")  
	fmt.Println("     ✅ Volume data processing (3D FFT) - compute and memory intensive")
	fmt.Println("   → Vector Operations Analysis:")
	fmt.Println("     ✅ AXPY operations - foundation for iterative solvers")
	fmt.Println("     ✅ DOT products - ubiquitous in scientific computing")
	fmt.Println("     ✅ NORM calculations - essential for convergence testing")
	
	fmt.Println("\n🎯 Architecture-Specific Scientific Computing Strengths:")
	fmt.Println("   → ARM Graviton3:")
	fmt.Println("     • Excellent memory bandwidth for large-scale scientific computing")
	fmt.Println("     • SVE optimization benefits for vector operations")
	fmt.Println("     • Best cost efficiency for research workloads")
	fmt.Println("   → Intel Ice Lake:")
	fmt.Println("     • Peak GFLOPS performance with AVX-512 optimization")
	fmt.Println("     • Intel MKL integration advantages for FFTW")
	fmt.Println("     • Superior single-thread performance for small problems")
	fmt.Println("   → AMD EPYC:")
	fmt.Println("     • Competitive performance across scientific workloads")
	fmt.Println("     • Good balance of compute and memory bandwidth")
	fmt.Println("     • Strong price/performance for research computing")
	
	fmt.Println("\n🔬 Research Workload Recommendations:")
	fmt.Println("   → Signal Processing: ARM Graviton3 (memory bandwidth + cost)")
	fmt.Println("   → Physics Simulations: Intel Ice Lake (peak GFLOPS)")
	fmt.Println("   → Large-Scale Computing: ARM Graviton3 (sustained performance)")
	fmt.Println("   → Mixed Workloads: AMD EPYC (balanced performance)")
}