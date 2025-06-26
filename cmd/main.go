// Package main provides the command-line interface for AWS Instance Benchmarks.
//
// This package implements a comprehensive CLI tool for executing performance
// benchmarks across AWS EC2 instance types with intelligent orchestration,
// container management, and result storage capabilities.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/containers"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/discovery"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/storage"
	"github.com/spf13/cobra"
)

// CLI validation errors.
var (
	ErrKeyPairRequired      = errors.New("--key-pair is required")
	ErrSecurityGroupRequired = errors.New("--security-group is required") 
	ErrSubnetRequired       = errors.New("--subnet is required")
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aws-benchmark-collector",
		Short: "AWS EC2 instance benchmark collection tool",
		Long:  "Comprehensive performance benchmark collection for AWS EC2 instances",
	}

	var discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "Discover AWS instance types and their architectures",
		RunE:  runDiscover,
	}

	var updateContainers bool
	var dryRun bool

	discoverCmd.Flags().BoolVar(&updateContainers, "update-containers", false, "Update container architecture mappings")
	discoverCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build architecture-optimized benchmark containers",
		RunE:  runBuild,
	}

	var architectures []string
	var benchmarks []string
	var registry string
	var namespace string
	var pushFlag bool

	buildCmd.Flags().StringSliceVar(&architectures, "architectures", []string{"intel-icelake", "amd-zen4", "graviton3"}, "Architecture tags to build")
	buildCmd.Flags().StringSliceVar(&benchmarks, "benchmarks", []string{"stream"}, "Benchmark suites to build")
	buildCmd.Flags().StringVar(&registry, "registry", "public.ecr.aws", "Container registry URL")
	buildCmd.Flags().StringVar(&namespace, "namespace", "aws-benchmarks", "Registry namespace")
	buildCmd.Flags().BoolVar(&pushFlag, "push", false, "Push containers after building")

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run benchmarks on AWS EC2 instances",
		RunE:  runBenchmarkCmd,
	}

	var instanceTypes []string
	var region string
	var keyPair string
	var securityGroup string
	var subnet string
	var skipQuota bool
	var benchmarkSuites []string
	var maxConcurrency int

	runCmd.Flags().StringSliceVar(&instanceTypes, "instance-types", []string{"m7i.large"}, "Instance types to benchmark")
	runCmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region")
	runCmd.Flags().StringVar(&keyPair, "key-pair", "", "EC2 key pair name")
	runCmd.Flags().StringVar(&securityGroup, "security-group", "", "Security group ID")
	runCmd.Flags().StringVar(&subnet, "subnet", "", "Subnet ID")
	runCmd.Flags().BoolVar(&skipQuota, "skip-quota-check", false, "Skip quota validation before launching")
	runCmd.Flags().StringSliceVar(&benchmarkSuites, "benchmarks", []string{"stream"}, "Benchmark suites to run")
	runCmd.Flags().IntVar(&maxConcurrency, "max-concurrency", 5, "Maximum number of concurrent benchmarks")

	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runDiscover(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	
	discoverer, err := discovery.NewInstanceDiscoverer()
	if err != nil {
		return fmt.Errorf("failed to create discoverer: %w", err)
	}

	updateContainers, _ := cmd.Flags().GetBool("update-containers")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if dryRun {
		fmt.Println("DRY RUN: Would discover instance types and architectures")
	}

	instances, err := discoverer.DiscoverAllInstanceTypes(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover instances: %w", err)
	}

	fmt.Printf("Discovered %d instance types\n", len(instances))
	
	if updateContainers {
		mappings := discoverer.GenerateArchitectureMappings(instances)
		fmt.Printf("Generated mappings for %d instance families\n", len(mappings))
		
		if !dryRun {
			if err := discoverer.UpdateMappingsFile(mappings); err != nil {
				return fmt.Errorf("failed to update mappings: %w", err)
			}
			fmt.Println("Updated architecture mappings file")
		}
	}

	return nil
}

func runBuild(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	
	architectures, _ := cmd.Flags().GetStringSlice("architectures")
	benchmarks, _ := cmd.Flags().GetStringSlice("benchmarks")
	registry, _ := cmd.Flags().GetString("registry")
	namespace, _ := cmd.Flags().GetString("namespace")
	pushFlag, _ := cmd.Flags().GetBool("push")

	builder := containers.NewBuilder(registry, namespace)

	for _, arch := range architectures {
		for _, benchmark := range benchmarks {
			fmt.Printf("Building %s container for %s architecture...\n", benchmark, arch)
			
			config := containers.BuildConfig{
				Architecture:      arch,
				ContainerTag:      arch,
				BenchmarkSuite:    benchmark,
				CompilerType:      getCompilerType(arch),
				OptimizationFlags: builder.GetOptimizationFlags(arch, getCompilerType(arch)),
				BaseImage:         getBaseImage(arch),
				SpackConfig:       fmt.Sprintf("%s.yaml", arch),
			}

			if err := builder.BuildContainer(ctx, config); err != nil {
				return fmt.Errorf("failed to build container for %s/%s: %w", arch, benchmark, err)
			}

			if pushFlag {
				fmt.Printf("Pushing %s container for %s architecture...\n", benchmark, arch)
				if err := builder.PushContainer(ctx, config); err != nil {
					return fmt.Errorf("failed to push container for %s/%s: %w", arch, benchmark, err)
				}
			}
		}
	}

	fmt.Println("Container build process completed successfully")
	return nil
}

func getCompilerType(architecture string) string {
	if strings.Contains(architecture, "intel") {
		return "intel"
	}
	if strings.Contains(architecture, "amd") {
		return "amd"
	}
	return "gcc"
}

func getBaseImage(architecture string) string {
	if strings.Contains(architecture, "arm") || strings.Contains(architecture, "graviton") {
		return "arm64v8/ubuntu:22.04"  // ARM64 base
	}
	return "ubuntu:22.04"  // x86_64 base
}

func runBenchmarkCmd(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	
	instanceTypes, _ := cmd.Flags().GetStringSlice("instance-types")
	region, _ := cmd.Flags().GetString("region")
	keyPair, _ := cmd.Flags().GetString("key-pair")
	securityGroup, _ := cmd.Flags().GetString("security-group")
	subnet, _ := cmd.Flags().GetString("subnet")
	skipQuota, _ := cmd.Flags().GetBool("skip-quota-check")
	benchmarkSuites, _ := cmd.Flags().GetStringSlice("benchmarks")
	maxConcurrency, _ := cmd.Flags().GetInt("max-concurrency")

	// Validate required parameters
	if keyPair == "" {
		return ErrKeyPairRequired
	}
	if securityGroup == "" {
		return ErrSecurityGroupRequired
	}
	if subnet == "" {
		return ErrSubnetRequired
	}

	orchestrator, err := aws.NewOrchestrator(region)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Initialize S3 storage for results
	storageConfig := storage.Config{
		BucketName:         "aws-instance-benchmarks-data",
		KeyPrefix:          "instance-benchmarks/",
		EnableCompression:  false,
		EnableVersioning:   false,
		RetryAttempts:      3,
		UploadTimeout:      5 * time.Minute,
		BatchSize:          1,
		StorageClass:       "STANDARD",
		DataVersion:        "1.0",
	}
	s3Storage, err := storage.NewS3Storage(ctx, storageConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize S3 storage: %w", err)
	}

	registry, _ := cmd.Parent().PersistentFlags().GetString("registry")
	namespace, _ := cmd.Parent().PersistentFlags().GetString("namespace")
	if registry == "" {
		registry = "public.ecr.aws"
	}
	if namespace == "" {
		namespace = "aws-benchmarks"
	}

	// Create benchmark jobs for parallel execution
	type benchmarkJob struct {
		instanceType   string
		benchmarkSuite string
		config         aws.BenchmarkConfig
	}

	var jobs []benchmarkJob
	for _, instanceType := range instanceTypes {
		for _, benchmarkSuite := range benchmarkSuites {
			containerImage := fmt.Sprintf("%s/%s:%s-%s", registry, namespace, benchmarkSuite, 
				getContainerTagForInstance(instanceType))

			config := aws.BenchmarkConfig{
				InstanceType:    instanceType,
				ContainerImage:  containerImage,
				BenchmarkSuite:  benchmarkSuite,
				Region:          region,
				KeyPairName:     keyPair,
				SecurityGroupID: securityGroup,
				SubnetID:        subnet,
				SkipQuotaCheck:  skipQuota,
				MaxRetries:      3,
				Timeout:         10 * time.Minute,
			}
			
			jobs = append(jobs, benchmarkJob{
				instanceType:   instanceType,
				benchmarkSuite: benchmarkSuite,
				config:         config,
			})
		}
	}

	fmt.Printf("Starting parallel benchmark run for %d jobs (%d instance types) in region %s\n", 
		len(jobs), len(instanceTypes), region)
	fmt.Printf("Max concurrency: %d\n", maxConcurrency)

	// Create semaphore to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	
	successCount := 0
	failureCount := 0
	startTime := time.Now()

	// Execute benchmarks in parallel
	for _, job := range jobs {
		wg.Add(1)
		go func(j benchmarkJob) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			fmt.Printf("🚀 Starting %s benchmark on %s...\n", j.benchmarkSuite, j.instanceType)
			
			result, err := orchestrator.RunBenchmark(ctx, j.config)
			if err != nil {
				resultsMutex.Lock()
				failureCount++
				resultsMutex.Unlock()
				
				if quotaErr, ok := err.(*aws.QuotaError); ok {
					fmt.Printf("⚠️  Skipped %s due to quota: %s\n", j.instanceType, quotaErr.Message)
					return
				}
				fmt.Printf("❌ Failed %s benchmark on %s: %v\n", j.benchmarkSuite, j.instanceType, err)
				return
			}

			fmt.Printf("✅ Completed %s benchmark on %s (took %v)\n", 
				j.benchmarkSuite, j.instanceType, result.EndTime.Sub(result.StartTime))
			fmt.Printf("   Instance: %s, Public IP: %s\n", result.InstanceID, result.PublicIP)

			// Store results to S3 and locally
			if err := storeResults(ctx, s3Storage, result, j.benchmarkSuite, region); err != nil {
				fmt.Printf("⚠️  Failed to store results for %s: %v\n", j.instanceType, err)
			} else {
				fmt.Printf("   Results stored successfully for %s\n", j.instanceType)
			}
			
			resultsMutex.Lock()
			successCount++
			resultsMutex.Unlock()
		}(job)
	}

	// Wait for all benchmarks to complete
	wg.Wait()
	totalTime := time.Since(startTime)

	// Print summary report
	fmt.Printf("\n📊 Benchmark Run Summary:\n")
	fmt.Printf("   Total jobs: %d\n", len(jobs))
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", failureCount)
	fmt.Printf("   Total time: %v\n", totalTime)
	fmt.Printf("   Average time per job: %v\n", totalTime/time.Duration(len(jobs)))
	
	if maxConcurrency > 1 {
		sequentialTime := time.Duration(len(jobs)) * 48 * time.Second // Estimated 48s per benchmark
		efficiency := float64(sequentialTime) / float64(totalTime) * 100
		fmt.Printf("   Estimated speedup: %.1fx (%.0f%% efficiency)\n", 
			float64(sequentialTime)/float64(totalTime), efficiency)
	}

	fmt.Println("\n✅ Parallel benchmark execution completed!")
	return nil
}

func getContainerTagForInstance(instanceType string) string {
	// Extract family and map to container tag
	family := extractInstanceFamily(instanceType)
	
	// Simple mapping - in real implementation would use the mappings file
	if strings.Contains(family, "7i") || strings.Contains(family, "7") && strings.Contains(instanceType, "i") {
		return "intel-icelake"
	}
	if strings.Contains(family, "7a") || strings.Contains(family, "7") && strings.Contains(instanceType, "a") {
		return "amd-zen4"
	}
	if strings.Contains(family, "7g") || strings.Contains(family, "7") && strings.Contains(instanceType, "g") {
		return "graviton3"
	}
	return "intel-skylake" // Default fallback
}

func storeResults(ctx context.Context, s3Storage *storage.S3Storage, result *aws.InstanceResult, benchmarkSuite string, region string) error {
	// Create comprehensive result structure for JSON storage following ComputeCompass integration format
	resultData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":        result.StartTime.UTC().Format(time.RFC3339),
			"instance_type":    result.InstanceType,
			"instance_id":      result.InstanceID,
			"benchmark_suite":  benchmarkSuite,
			"region":          region,
			"duration_seconds": result.EndTime.Sub(result.StartTime).Seconds(),
			"data_version":    "1.0",
			"collection_method": "automated",
		},
		"performance_data": result.BenchmarkData,
		"system_info": map[string]interface{}{
			"public_ip":   result.PublicIP,
			"private_ip":  result.PrivateIP,
			"status":      result.Status,
			"architecture": getArchitectureFromInstance(result.InstanceType),
			"instance_family": extractInstanceFamily(result.InstanceType),
		},
		"execution_context": map[string]interface{}{
			"container_runtime": "docker",
			"benchmark_version": "latest",
			"compiler_optimizations": getCompilerOptimizations(result.InstanceType),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(resultData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Generate filename with timestamp
	timestamp := result.StartTime.UTC().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s-%s.json", result.InstanceType, benchmarkSuite, timestamp)
	
	// Store locally
	localDir := filepath.Join("results", result.StartTime.UTC().Format("2006-01-02"))
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}
	
	localPath := filepath.Join(localDir, filename)
	if err := os.WriteFile(localPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	// Store to S3
	if err := s3Storage.StoreResult(ctx, resultData); err != nil {
		return fmt.Errorf("failed to store to S3: %w", err)
	}

	fmt.Printf("   Local:  %s\n", localPath)
	fmt.Printf("   S3:     Stored to S3 with structured key\n")
	
	return nil
}

func extractInstanceFamily(instanceType string) string {
	// Simple extraction - get everything before the first dot
	parts := strings.Split(instanceType, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return instanceType
}

func getArchitectureFromInstance(instanceType string) string {
	// Determine architecture based on instance type
	if strings.Contains(instanceType, "g.") || strings.HasSuffix(instanceType, "g") {
		if strings.HasPrefix(instanceType, "m") || strings.HasPrefix(instanceType, "c") || 
			strings.HasPrefix(instanceType, "r") || strings.HasPrefix(instanceType, "t") {
			return "arm64" // Graviton instances
		}
	}
	return "x86_64" // Intel/AMD instances
}

func getCompilerOptimizations(instanceType string) string {
	arch := getArchitectureFromInstance(instanceType)
	if arch == "arm64" {
		return "-O3 -march=native -mtune=native -mcpu=neoverse-v1"
	}
	
	// Detect Intel vs AMD for x86_64
	family := extractInstanceFamily(instanceType)
	if strings.Contains(family, "a") {
		return "-O3 -march=native -mtune=native -mprefer-avx128" // AMD optimizations
	}
	return "-O3 -march=native -mtune=native -mavx2" // Intel optimizations
}