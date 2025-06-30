package main

import (
	"context"
	"fmt"
	"log"
	"time"

	awspkg "github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
)

// Async benchmark launcher - fire and forget with S3 tracking
func main() {
	fmt.Println("🚀 ASYNC BENCHMARK LAUNCHER")
	fmt.Println("============================")
	fmt.Println("Fire-and-forget benchmark execution with S3 sentinel tracking")
	fmt.Println("============================")

	ctx := context.Background()

	// Initialize async launcher
	launcher, err := awspkg.NewAsyncLauncher("us-west-2")
	if err != nil {
		log.Fatalf("Failed to create async launcher: %v", err)
	}

	// Configure benchmark jobs
	configs := []awspkg.BenchmarkConfig{
		// ARM Graviton3 - STREAM memory bandwidth
		{
			InstanceType:    "c7g.large",
			BenchmarkSuite:  "stream", 
			Region:          "us-west-2",
			KeyPairName:     "aws-benchmark-test",
			SecurityGroupID: "sg-06feaa8214edbfdbf",
			SubnetID:        "subnet-06a8cff8a4457b4a7",
			SkipQuotaCheck:  false,
			MaxRetries:      3,
		},
		// Intel Ice Lake - HPL LINPACK
		{
			InstanceType:    "c7i.large",
			BenchmarkSuite:  "hpl",
			Region:          "us-west-2", 
			KeyPairName:     "aws-benchmark-test",
			SecurityGroupID: "sg-06feaa8214edbfdbf",
			SubnetID:        "subnet-06a8cff8a4457b4a7",
			SkipQuotaCheck:  false,
			MaxRetries:      3,
		},
		// AMD EPYC - FFTW scientific computing
		{
			InstanceType:    "c7a.large",
			BenchmarkSuite:  "fftw",
			Region:          "us-west-2",
			KeyPairName:     "aws-benchmark-test", 
			SecurityGroupID: "sg-06feaa8214edbfdbf",
			SubnetID:        "subnet-06a8cff8a4457b4a7",
			SkipQuotaCheck:  false,
			MaxRetries:      3,
		},
	}

	// Create launch request
	request := &awspkg.LaunchRequest{
		Configs:       configs,
		S3Bucket:      "aws-benchmark-results-bucket", // Replace with your S3 bucket
		JobNamePrefix: "phase2-async-test",
		MaxRuntime:    4 * time.Hour, // 4 hours max per benchmark
		Tags: map[string]string{
			"Project":     "AWS-Instance-Benchmarks",
			"Phase":       "2",
			"LaunchType":  "Async",
			"Environment": "Testing",
		},
	}

	fmt.Printf("\n🎯 LAUNCH CONFIGURATION\n")
	fmt.Printf("=======================\n")
	fmt.Printf("   Benchmarks: %d\n", len(configs))
	fmt.Printf("   S3 Bucket: %s\n", request.S3Bucket)
	fmt.Printf("   Max Runtime: %v per benchmark\n", request.MaxRuntime)
	fmt.Printf("   Job Prefix: %s\n", request.JobNamePrefix)
	fmt.Printf("=======================\n\n")

	// Print expected performance for each benchmark
	fmt.Printf("📊 EXPECTED PERFORMANCE\n")
	fmt.Printf("=======================\n")
	for i, config := range configs {
		fmt.Printf("   %d. %s on %s:\n", i+1, config.BenchmarkSuite, config.InstanceType)
		printExpectedPerformance(config.InstanceType, config.BenchmarkSuite)
		fmt.Printf("\n")
	}

	// Launch benchmarks asynchronously
	response, err := launcher.LaunchBenchmarks(ctx, request)
	if err != nil {
		log.Fatalf("Failed to launch benchmarks: %v", err)
	}

	// Display launch results
	fmt.Printf("🎉 LAUNCH COMPLETE!\n")
	fmt.Printf("===================\n")
	fmt.Printf("   Successfully launched: %d/%d benchmarks\n", 
		response.LaunchedCount, len(configs))
	fmt.Printf("   Failed launches: %d\n", response.FailedCount)
	
	if len(response.Errors) > 0 {
		fmt.Printf("   Errors:\n")
		for _, err := range response.Errors {
			fmt.Printf("     - %s\n", err)
		}
	}

	fmt.Printf("\n📍 TRACKING INFORMATION\n")
	fmt.Printf("=======================\n")
	
	var totalEstimatedCost float64
	for i, job := range response.Jobs {
		fmt.Printf("   %d. Benchmark ID: %s\n", i+1, job.BenchmarkID)
		fmt.Printf("      Instance: %s (%s)\n", job.InstanceID, job.BenchmarkConfig.InstanceType)
		fmt.Printf("      Benchmark: %s\n", job.BenchmarkConfig.BenchmarkSuite)
		fmt.Printf("      S3 Path: s3://%s/%s\n", job.S3Bucket, job.S3Prefix)
		fmt.Printf("      Est. Cost: $%.4f\n", job.EstimatedCost)
		fmt.Printf("      Launched: %s\n", job.LaunchedAt.Format("15:04:05"))
		fmt.Printf("\n")
		totalEstimatedCost += job.EstimatedCost
	}

	fmt.Printf("💰 TOTAL ESTIMATED COST: $%.4f\n\n", totalEstimatedCost)

	fmt.Printf("🔍 MONITORING INSTRUCTIONS\n")
	fmt.Printf("===========================\n")
	fmt.Printf("   1. Benchmarks are running independently on AWS instances\n")
	fmt.Printf("   2. Each instance will self-terminate when complete\n")
	fmt.Printf("   3. Results are automatically uploaded to S3\n")
	fmt.Printf("   4. Use the collector tool to check progress:\n")
	fmt.Printf("      go run async_benchmark_collector.go\n")
	fmt.Printf("   5. Monitor S3 bucket for sentinel files:\n")
	fmt.Printf("      aws s3 ls s3://%s/benchmarks/ --recursive\n", request.S3Bucket)

	fmt.Printf("\n📈 NEXT STEPS\n")
	fmt.Printf("=============\n")
	fmt.Printf("   ✅ Benchmarks launched successfully\n")
	fmt.Printf("   🔄 Instances executing independently (no timeouts!)\n")
	fmt.Printf("   📊 Results will appear in S3 as they complete\n")
	fmt.Printf("   🎯 Use collector tool to gather results when ready\n")
	fmt.Printf("   💰 Instances will self-terminate to minimize cost\n")

	if response.LaunchedCount > 0 {
		fmt.Printf("\n🎉 ASYNC BENCHMARK LAUNCH: ✅ SUCCESSFUL\n")
		fmt.Printf("All benchmarks are now running independently!\n")
	} else {
		fmt.Printf("\n❌ ASYNC BENCHMARK LAUNCH: FAILED\n")
		fmt.Printf("No benchmarks were launched successfully.\n")
	}
}

func printExpectedPerformance(instanceType, benchmark string) {
	switch instanceType {
	case "c7g.large":
		fmt.Printf("      🟢 ARM Graviton3 - Excellent efficiency\n")
		if benchmark == "stream" {
			fmt.Printf("      Copy: ~40-50 GB/s, Scale: ~40-50 GB/s\n")
			fmt.Printf("      Add: ~45-55 GB/s, Triad: ~45-55 GB/s\n")
		}
	case "c7i.large":
		fmt.Printf("      🔵 Intel Ice Lake - Peak GFLOPS performance\n")
		if benchmark == "hpl" {
			fmt.Printf("      Peak LINPACK: ~80-120 GFLOPS\n")
			fmt.Printf("      Sustained: ~70-100 GFLOPS\n")
		}
	case "c7a.large":
		fmt.Printf("      🟡 AMD EPYC 9R14 - Balanced performance\n")
		if benchmark == "fftw" {
			fmt.Printf("      1D FFT: ~75-95 GFLOPS\n")
			fmt.Printf("      2D FFT: ~60-80 GFLOPS\n") 
			fmt.Printf("      3D FFT: ~42-62 GFLOPS\n")
		}
	}
}