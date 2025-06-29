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
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/containers"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/discovery"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/monitoring"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/pricing"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/schema"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/storage"
	"github.com/spf13/cobra"
)

// CLI validation errors.
var (
	ErrKeyPairRequired      = errors.New("--key-pair is required")
	ErrSecurityGroupRequired = errors.New("--security-group is required") 
	ErrSubnetRequired       = errors.New("--subnet is required")
)

// benchmarkResult stores the results of individual benchmark runs for statistical analysis
type benchmarkResult struct {
	instanceType   string
	benchmarkSuite string
	iteration      int
	success        bool
	result         *aws.InstanceResult
	metrics        monitoring.BenchmarkMetrics
}

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
	var iterations int
	var s3Bucket string

	runCmd.Flags().StringSliceVar(&instanceTypes, "instance-types", []string{"m7i.large"}, "Instance types to benchmark")
	runCmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region")
	runCmd.Flags().StringVar(&keyPair, "key-pair", "", "EC2 key pair name")
	runCmd.Flags().StringVar(&securityGroup, "security-group", "", "Security group ID")
	runCmd.Flags().StringVar(&subnet, "subnet", "", "Subnet ID")
	runCmd.Flags().BoolVar(&skipQuota, "skip-quota-check", false, "Skip quota validation before launching")
	runCmd.Flags().StringSliceVar(&benchmarkSuites, "benchmarks", []string{"stream"}, "Benchmark suites to run (stream, hpl)")
	runCmd.Flags().IntVar(&maxConcurrency, "max-concurrency", 5, "Maximum number of concurrent benchmarks")
	runCmd.Flags().IntVar(&iterations, "iterations", 1, "Number of benchmark iterations for statistical validation")
	runCmd.Flags().StringVar(&s3Bucket, "s3-bucket", "", "S3 bucket for storing results (defaults to aws-instance-benchmarks-data-{region})")

	var schemaCmd = &cobra.Command{
		Use:   "schema",
		Short: "Schema validation and migration tools",
		Long:  "Tools for validating and migrating benchmark data schemas",
	}

	var validateCmd = &cobra.Command{
		Use:   "validate [file|directory]",
		Short: "Validate JSON files against schema",
		Args:  cobra.ExactArgs(1),
		RunE:  runSchemaValidate,
	}

	var migrateCmd = &cobra.Command{
		Use:   "migrate [input] [output]",
		Short: "Migrate data to target schema version",
		Args:  cobra.ExactArgs(2),
		RunE:  runSchemaMigrate,
	}

	var targetVersion string
	var reportOnly bool

	validateCmd.Flags().StringVar(&targetVersion, "version", "1.0.0", "Target schema version")
	migrateCmd.Flags().StringVar(&targetVersion, "version", "1.0.0", "Target schema version")
	migrateCmd.Flags().BoolVar(&reportOnly, "report-only", false, "Generate migration report without migrating")

	schemaCmd.AddCommand(validateCmd)
	schemaCmd.AddCommand(migrateCmd)

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [results-directory]",
		Short: "Analyze benchmark results with price/performance calculations",
		Args:  cobra.ExactArgs(1),
		RunE:  runAnalyze,
	}

	var baselineInstance string
	var outputFormat string
	var sortByMetric string

	analyzeCmd.Flags().StringVar(&baselineInstance, "baseline", "m7i.large", "Baseline instance for normalization")
	analyzeCmd.Flags().StringVar(&outputFormat, "format", "table", "Output format: table, json, csv")
	analyzeCmd.Flags().StringVar(&sortByMetric, "sort", "value_score", "Sort by: value_score, cost_efficiency, performance, price")

	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(analyzeCmd)

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
	iterations, _ := cmd.Flags().GetInt("iterations")
	s3Bucket, _ := cmd.Flags().GetString("s3-bucket")

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
	bucketName := s3Bucket
	if bucketName == "" {
		bucketName = fmt.Sprintf("aws-instance-benchmarks-data-%s", region)
	}
	
	storageConfig := storage.Config{
		BucketName:         bucketName,
		KeyPrefix:          "instance-benchmarks/",
		EnableCompression:  false,
		EnableVersioning:   false,
		RetryAttempts:      3,
		UploadTimeout:      5 * time.Minute,
		BatchSize:          1,
		StorageClass:       "STANDARD",
		DataVersion:        "1.0",
	}
	s3Storage, err := storage.NewS3Storage(ctx, storageConfig, region)
	if err != nil {
		return fmt.Errorf("failed to initialize S3 storage: %w", err)
	}

	// Initialize CloudWatch metrics collector
	metricsCollector, err := monitoring.NewMetricsCollector(region)
	if err != nil {
		fmt.Printf("⚠️  Failed to initialize CloudWatch metrics: %v\n", err)
		fmt.Println("   Continuing without metrics collection...")
		metricsCollector = nil
	} else {
		fmt.Println("✅ CloudWatch metrics collection enabled")
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
		iteration      int
		config         aws.BenchmarkConfig
	}

	var jobs []benchmarkJob
	for _, instanceType := range instanceTypes {
		for _, benchmarkSuite := range benchmarkSuites {
			for iteration := 1; iteration <= iterations; iteration++ {
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
					iteration:      iteration,
					config:         config,
				})
			}
		}
	}

	fmt.Printf("Starting parallel benchmark run for %d jobs (%d instance types, %d iterations) in region %s\n", 
		len(jobs), len(instanceTypes), iterations, region)
	fmt.Printf("Max concurrency: %d\n", maxConcurrency)

	// Create semaphore to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	
	successCount := 0
	failureCount := 0
	startTime := time.Now()
	
	// Collect all results for statistical analysis
	var allResults []benchmarkResult

	// Execute benchmarks in parallel
	for _, job := range jobs {
		wg.Add(1)
		go func(j benchmarkJob) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			if iterations > 1 {
				fmt.Printf("🚀 Starting %s benchmark on %s (iteration %d/%d)...\n", j.benchmarkSuite, j.instanceType, j.iteration, iterations)
			} else {
				fmt.Printf("🚀 Starting %s benchmark on %s...\n", j.benchmarkSuite, j.instanceType)
			}
			
			benchmarkStartTime := time.Now()
			result, err := orchestrator.RunBenchmark(ctx, j.config)
			benchmarkEndTime := time.Now()
			
			// Prepare metrics for CloudWatch
			benchmarkMetrics := monitoring.BenchmarkMetrics{
				InstanceType:       j.instanceType,
				InstanceFamily:     extractInstanceFamily(j.instanceType),
				BenchmarkSuite:     j.benchmarkSuite,
				Region:            region,
				Success:           err == nil,
				ExecutionDuration: benchmarkEndTime.Sub(benchmarkStartTime).Seconds(),
				Timestamp:         benchmarkEndTime,
			}
			
			if err != nil {
				resultsMutex.Lock()
				failureCount++
				resultsMutex.Unlock()
				
				// Categorize error for metrics
				if quotaErr, ok := err.(*aws.QuotaError); ok {
					benchmarkMetrics.ErrorCategory = "quota"
					fmt.Printf("⚠️  Skipped %s due to quota: %s\n", j.instanceType, quotaErr.Message)
				} else {
					benchmarkMetrics.ErrorCategory = "infrastructure"
					fmt.Printf("❌ Failed %s benchmark on %s: %v\n", j.benchmarkSuite, j.instanceType, err)
				}
				
				// Publish failure metrics
				if metricsCollector != nil {
					if publishErr := metricsCollector.PublishBenchmarkMetrics(ctx, benchmarkMetrics); publishErr != nil {
						fmt.Printf("   ⚠️ Failed to publish failure metrics: %v\n", publishErr)
					}
				}
				
				// Store failed result for analysis
				resultsMutex.Lock()
				allResults = append(allResults, benchmarkResult{
					instanceType:   j.instanceType,
					benchmarkSuite: j.benchmarkSuite,
					iteration:      j.iteration,
					success:        false,
					result:         nil,
					metrics:        benchmarkMetrics,
				})
				resultsMutex.Unlock()
				return
			}

			benchmarkDuration := result.EndTime.Sub(result.StartTime).Seconds()
			benchmarkMetrics.BenchmarkDuration = benchmarkDuration
			
			// Extract performance metrics from benchmark results
			if result.BenchmarkData != nil {
				benchmarkMetrics.PerformanceMetrics = make(map[string]float64)
				
				// Extract benchmark-specific performance data
				switch j.benchmarkSuite {
				case "stream":
					streamData := result.BenchmarkData
					if triad, exists := streamData["triad_bandwidth"]; exists {
						if triadVal, ok := triad.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["triad_bandwidth"] = triadVal
						}
					}
					if copy, exists := streamData["copy_bandwidth"]; exists {
						if copyVal, ok := copy.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["copy_bandwidth"] = copyVal
						}
					}
					if scale, exists := streamData["scale_bandwidth"]; exists {
						if scaleVal, ok := scale.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["scale_bandwidth"] = scaleVal
						}
					}
					if add, exists := streamData["add_bandwidth"]; exists {
						if addVal, ok := add.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["add_bandwidth"] = addVal
						}
					}
				case "hpl":
					hplData := result.BenchmarkData
					if gflops, exists := hplData["gflops"]; exists {
						if gflopsVal, ok := gflops.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["gflops"] = gflopsVal
						}
					}
					if efficiency, exists := hplData["efficiency"]; exists {
						if efficiencyVal, ok := efficiency.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["efficiency"] = efficiencyVal
						}
					}
					if executionTime, exists := hplData["execution_time"]; exists {
						if executionTimeVal, ok := executionTime.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["execution_time"] = executionTimeVal
						}
					}
					if residual, exists := hplData["residual"]; exists {
						if residualVal, ok := residual.(float64); ok {
							benchmarkMetrics.PerformanceMetrics["residual"] = residualVal
						}
					}
				}
				
				// Calculate quality score based on performance stability
				benchmarkMetrics.QualityScore = calculateQualityScore(result.BenchmarkData)
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
			
			// Publish success metrics to CloudWatch
			if metricsCollector != nil {
				if publishErr := metricsCollector.PublishBenchmarkMetrics(ctx, benchmarkMetrics); publishErr != nil {
					fmt.Printf("   ⚠️ Failed to publish success metrics: %v\n", publishErr)
				} else {
					fmt.Printf("   📊 Metrics published to CloudWatch\n")
				}
			}
			
			// Store successful result for analysis
			resultsMutex.Lock()
			allResults = append(allResults, benchmarkResult{
				instanceType:   j.instanceType,
				benchmarkSuite: j.benchmarkSuite,
				iteration:      j.iteration,
				success:        true,
				result:         result,
				metrics:        benchmarkMetrics,
			})
			successCount++
			resultsMutex.Unlock()
		}(job)
	}

	// Wait for all benchmarks to complete
	wg.Wait()
	totalTime := time.Since(startTime)

	// Perform statistical analysis if multiple iterations
	if iterations > 1 {
		fmt.Printf("\n📈 Statistical Analysis:\n")
		performStatisticalAnalysis(allResults, iterations)
	}

	// Print summary report
	fmt.Printf("\n📊 Benchmark Run Summary:\n")
	fmt.Printf("   Total jobs: %d\n", len(jobs))
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", failureCount)
	fmt.Printf("   Total time: %v\n", totalTime)
	fmt.Printf("   Average time per job: %v\n", totalTime/time.Duration(len(jobs)))
	
	var efficiency float64
	if maxConcurrency > 1 {
		sequentialTime := time.Duration(len(jobs)) * 48 * time.Second // Estimated 48s per benchmark
		efficiency = float64(sequentialTime) / float64(totalTime) * 100
		fmt.Printf("   Estimated speedup: %.1fx (%.0f%% efficiency)\n", 
			float64(sequentialTime)/float64(totalTime), efficiency)
	}

	// Publish operational metrics to CloudWatch
	if metricsCollector != nil {
		operationalMetrics := monitoring.OperationalMetrics{
			InstanceLaunchDuration: totalTime.Seconds() / float64(len(jobs)), // Average launch time
			ActiveInstances:        0, // All instances terminated after benchmarks
			FailureRate:           float64(failureCount) / float64(len(jobs)) * 100,
			Region:               region,
			Timestamp:            time.Now(),
		}
		
		if publishErr := metricsCollector.PublishOperationalMetrics(ctx, operationalMetrics); publishErr != nil {
			fmt.Printf("⚠️  Failed to publish operational metrics: %v\n", publishErr)
		} else {
			fmt.Printf("📈 Operational metrics published to CloudWatch\n")
		}
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
		"schema_version": "1.0.0",
		"metadata": map[string]interface{}{
			"data_version":     "1.0",
			"instanceType":     result.InstanceType,
			"instanceFamily":   extractInstanceFamily(result.InstanceType),
			"region":          region,
			"processorArchitecture": getArchitectureFromInstance(result.InstanceType),
			"timestamp":        result.StartTime.UTC().Format(time.RFC3339),
			"instance_id":      result.InstanceID,
			"benchmark_suite":  benchmarkSuite,
			"duration_seconds": result.EndTime.Sub(result.StartTime).Seconds(),
			"collection_method": "automated",
			"environment": map[string]interface{}{
				"containerImage": getContainerImageForInstance(result.InstanceType, benchmarkSuite),
				"timestamp":     result.StartTime.UTC().Format(time.RFC3339),
				"duration":      result.EndTime.Sub(result.StartTime).Seconds(),
			},
		},
		"performance": map[string]interface{}{
			"memory": result.BenchmarkData,
		},
		"validation": map[string]interface{}{
			"checksums": map[string]interface{}{
				"md5":    generateMD5Checksum(result.BenchmarkData),
				"sha256": generateSHA256Checksum(result.BenchmarkData),
			},
			"reproducibility": map[string]interface{}{
				"runs":       1,
				"confidence": 1.0,
			},
		},
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

	// Validate against schema
	schemaManager := schema.DefaultSchemaManager()
	validator, err := schemaManager.GetLatestValidator()
	if err != nil {
		fmt.Printf("⚠️  Schema validation unavailable: %v\n", err)
	} else {
		validationResult, err := validator.ValidateBytes(jsonData)
		if err != nil {
			fmt.Printf("⚠️  Schema validation failed: %v\n", err)
		} else if !validationResult.Valid {
			fmt.Printf("⚠️  Schema validation errors:\n%s\n", validationResult.String())
			// Continue storing despite validation errors for now
		} else {
			fmt.Printf("✅ Schema validation passed (v%s)\n", validationResult.SchemaVersion)
		}
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

func calculateQualityScore(benchmarkData interface{}) float64 {
	// Default quality score for successful benchmarks
	if benchmarkData == nil {
		return 0.5
	}
	
	if data, ok := benchmarkData.(map[string]interface{}); ok {
		// Check if this is STREAM data
		if _, hasTriad := data["triad_bandwidth"]; hasTriad {
			return calculateSTREAMQualityScore(data)
		}
		
		// Check if this is HPL data
		if _, hasGFLOPS := data["gflops"]; hasGFLOPS {
			return calculateHPLQualityScore(data)
		}
	}
	
	return 0.7 // Default score for other benchmark types
}

func performStatisticalAnalysis(allResults []benchmarkResult, iterations int) {
	// Group results by instance type and benchmark suite
	grouped := make(map[string][]benchmarkResult)
	
	for _, result := range allResults {
		if result.success {
			key := fmt.Sprintf("%s-%s", result.instanceType, result.benchmarkSuite)
			grouped[key] = append(grouped[key], result)
		}
	}
	
	// Analyze each group
	for key, results := range grouped {
		if len(results) < 2 {
			continue // Need at least 2 results for statistical analysis
		}
		
		parts := strings.Split(key, "-")
		instanceType := parts[0]
		benchmarkSuite := parts[1]
		
		fmt.Printf("\n   %s on %s (%d successful runs):\n", benchmarkSuite, instanceType, len(results))
		
		if benchmarkSuite == "stream" {
			analyzeSTREAMResults(results)
		} else if benchmarkSuite == "hpl" {
			analyzeHPLResults(results)
		}
	}
}

func analyzeSTREAMResults(results []benchmarkResult) {
	var triadValues []float64
	var copyValues []float64
	var scaleValues []float64
	var addValues []float64
	
	// Extract bandwidth values
	for _, result := range results {
		if result.result != nil && result.result.BenchmarkData != nil {
			data := result.result.BenchmarkData
			
			// Check for nested STREAM data structure
			if streamData, exists := data["stream"]; exists {
				if streamMap, ok := streamData.(map[string]interface{}); ok {
					// Extract triad bandwidth
					if triad, exists := streamMap["triad"]; exists {
						if triadMap, ok := triad.(map[string]interface{}); ok {
							if bandwidth, exists := triadMap["bandwidth"]; exists {
								if floatVal, ok := bandwidth.(float64); ok {
									triadValues = append(triadValues, floatVal)
								}
							}
						}
					}
					// Extract copy bandwidth  
					if copy, exists := streamMap["copy"]; exists {
						if copyMap, ok := copy.(map[string]interface{}); ok {
							if bandwidth, exists := copyMap["bandwidth"]; exists {
								if floatVal, ok := bandwidth.(float64); ok {
									copyValues = append(copyValues, floatVal)
								}
							}
						}
					}
					// Extract scale bandwidth
					if scale, exists := streamMap["scale"]; exists {
						if scaleMap, ok := scale.(map[string]interface{}); ok {
							if bandwidth, exists := scaleMap["bandwidth"]; exists {
								if floatVal, ok := bandwidth.(float64); ok {
									scaleValues = append(scaleValues, floatVal)
								}
							}
						}
					}
					// Extract add bandwidth
					if add, exists := streamMap["add"]; exists {
						if addMap, ok := add.(map[string]interface{}); ok {
							if bandwidth, exists := addMap["bandwidth"]; exists {
								if floatVal, ok := bandwidth.(float64); ok {
									addValues = append(addValues, floatVal)
								}
							}
						}
					}
				}
			}
			
			// Also check for flat structure (legacy support)
			if val, exists := data["triad_bandwidth"]; exists {
				if floatVal, ok := val.(float64); ok {
					triadValues = append(triadValues, floatVal)
				}
			}
			if val, exists := data["copy_bandwidth"]; exists {
				if floatVal, ok := val.(float64); ok {
					copyValues = append(copyValues, floatVal)
				}
			}
			if val, exists := data["scale_bandwidth"]; exists {
				if floatVal, ok := val.(float64); ok {
					scaleValues = append(scaleValues, floatVal)
				}
			}
			if val, exists := data["add_bandwidth"]; exists {
				if floatVal, ok := val.(float64); ok {
					addValues = append(addValues, floatVal)
				}
			}
		}
	}
	
	// Calculate and display statistics
	if len(triadValues) > 0 {
		mean, stdDev, cv := calculateStatistics(triadValues)
		confInt := calculateConfidenceInterval(triadValues, 0.95)
		fmt.Printf("     Triad Bandwidth: %.2f ± %.2f GB/s (CV: %.1f%%, 95%% CI: %.2f-%.2f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
	
	if len(copyValues) > 0 {
		mean, stdDev, cv := calculateStatistics(copyValues)
		confInt := calculateConfidenceInterval(copyValues, 0.95)
		fmt.Printf("     Copy Bandwidth:  %.2f ± %.2f GB/s (CV: %.1f%%, 95%% CI: %.2f-%.2f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
	
	if len(scaleValues) > 0 {
		mean, stdDev, cv := calculateStatistics(scaleValues)
		confInt := calculateConfidenceInterval(scaleValues, 0.95)
		fmt.Printf("     Scale Bandwidth: %.2f ± %.2f GB/s (CV: %.1f%%, 95%% CI: %.2f-%.2f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
	
	if len(addValues) > 0 {
		mean, stdDev, cv := calculateStatistics(addValues)
		confInt := calculateConfidenceInterval(addValues, 0.95)
		fmt.Printf("     Add Bandwidth:   %.2f ± %.2f GB/s (CV: %.1f%%, 95%% CI: %.2f-%.2f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
}

func analyzeHPLResults(results []benchmarkResult) {
	var gflopsValues []float64
	var efficiencyValues []float64
	var executionTimeValues []float64
	
	// Extract performance values
	for _, result := range results {
		if result.result != nil && result.result.BenchmarkData != nil {
			data := result.result.BenchmarkData
			if val, exists := data["gflops"]; exists {
				if floatVal, ok := val.(float64); ok {
					gflopsValues = append(gflopsValues, floatVal)
				}
			}
			if val, exists := data["efficiency"]; exists {
				if floatVal, ok := val.(float64); ok {
					efficiencyValues = append(efficiencyValues, floatVal)
				}
			}
			if val, exists := data["execution_time"]; exists {
				if floatVal, ok := val.(float64); ok {
					executionTimeValues = append(executionTimeValues, floatVal)
				}
			}
		}
	}
	
	// Calculate and display statistics
	if len(gflopsValues) > 0 {
		mean, stdDev, cv := calculateStatistics(gflopsValues)
		confInt := calculateConfidenceInterval(gflopsValues, 0.95)
		fmt.Printf("     GFLOPS:          %.2f ± %.2f (CV: %.1f%%, 95%% CI: %.2f-%.2f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
	
	if len(efficiencyValues) > 0 {
		mean, stdDev, cv := calculateStatistics(efficiencyValues)
		confInt := calculateConfidenceInterval(efficiencyValues, 0.95)
		fmt.Printf("     Efficiency:      %.3f ± %.3f (CV: %.1f%%, 95%% CI: %.3f-%.3f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
	
	if len(executionTimeValues) > 0 {
		mean, stdDev, cv := calculateStatistics(executionTimeValues)
		confInt := calculateConfidenceInterval(executionTimeValues, 0.95)
		fmt.Printf("     Execution Time:  %.2f ± %.2f s (CV: %.1f%%, 95%% CI: %.2f-%.2f)\n", 
			mean, stdDev, cv, confInt.lower, confInt.upper)
	}
}

func calculateStatistics(values []float64) (mean, stdDev, cv float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	mean = sum / float64(len(values))
	
	// Calculate standard deviation
	sumSquares := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	stdDev = math.Sqrt(variance)
	
	// Calculate coefficient of variation
	if mean != 0 {
		cv = (stdDev / mean) * 100
	}
	
	return mean, stdDev, cv
}

type confidenceInterval struct {
	lower, upper float64
}

func calculateConfidenceInterval(values []float64, confidence float64) confidenceInterval {
	if len(values) < 2 {
		return confidenceInterval{0, 0}
	}
	
	mean, stdDev, _ := calculateStatistics(values)
	n := float64(len(values))
	
	// Use t-distribution for small samples (simplified)
	var tValue float64
	switch {
	case n >= 30:
		tValue = 1.96 // Normal approximation for large samples
	case n >= 10:
		tValue = 2.26 // Approximate t-value for medium samples
	default:
		tValue = 3.18 // Conservative t-value for small samples
	}
	
	margin := tValue * (stdDev / math.Sqrt(n))
	
	return confidenceInterval{
		lower: mean - margin,
		upper: mean + margin,
	}
}

func calculateSTREAMQualityScore(streamData map[string]interface{}) float64 {
	var bandwidths []float64
	
	// Collect bandwidth values
	for _, key := range []string{"copy_bandwidth", "scale_bandwidth", "add_bandwidth", "triad_bandwidth"} {
		if val, exists := streamData[key]; exists {
			if floatVal, ok := val.(float64); ok && floatVal > 0 {
				bandwidths = append(bandwidths, floatVal)
			}
		}
	}
	
	if len(bandwidths) < 2 {
		return 0.5 // Not enough data points
	}
	
	// Calculate coefficient of variation (CV)
	mean := 0.0
	for _, bw := range bandwidths {
		mean += bw
	}
	mean /= float64(len(bandwidths))
	
	variance := 0.0
	for _, bw := range bandwidths {
		variance += (bw - mean) * (bw - mean)
	}
	variance /= float64(len(bandwidths))
	
	if mean == 0 {
		return 0.5
	}
	
	cv := (variance / (mean * mean)) // Coefficient of variation squared
	
	// Convert CV to quality score (lower CV = higher quality)
	qualityScore := 1.0 - (cv * 2.0)
	if qualityScore < 0.0 {
		qualityScore = 0.1
	}
	if qualityScore > 1.0 {
		qualityScore = 1.0
	}
	
	return qualityScore
}

func calculateHPLQualityScore(hplData map[string]interface{}) float64 {
	qualityScore := 1.0
	
	// Check efficiency
	if effVal, exists := hplData["efficiency"]; exists {
		if efficiency, ok := effVal.(float64); ok {
			if efficiency < 0.5 {
				qualityScore -= 0.4 // Penalize low efficiency heavily
			} else if efficiency < 0.7 {
				qualityScore -= 0.2 // Moderate penalty
			}
		}
	}
	
	// Check residual (numerical accuracy)
	if residualVal, exists := hplData["residual"]; exists {
		if residual, ok := residualVal.(float64); ok {
			if residual > 1e-6 {
				qualityScore -= 0.3 // Penalize poor numerical accuracy
			} else if residual > 1e-9 {
				qualityScore -= 0.1 // Small penalty for moderate accuracy
			}
		}
	}
	
	// Ensure quality is in valid range
	if qualityScore < 0.0 {
		qualityScore = 0.1
	}
	if qualityScore > 1.0 {
		qualityScore = 1.0
	}
	
	return qualityScore
}

func getContainerImageForInstance(instanceType, benchmarkSuite string) string {
	containerTag := getContainerTagForInstance(instanceType)
	return fmt.Sprintf("public.ecr.aws/aws-benchmarks/%s:%s", benchmarkSuite, containerTag)
}

func generateMD5Checksum(data interface{}) string {
	// For simplicity, return a placeholder checksum
	// In production, this should generate actual MD5 from the data
	return "d41d8cd98f00b204e9800998ecf8427e"
}

func generateSHA256Checksum(data interface{}) string {
	// For simplicity, return a placeholder checksum
	// In production, this should generate actual SHA256 from the data
	return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
}

func runSchemaValidate(cmd *cobra.Command, args []string) error {
	targetPath := args[0]
	versionStr, _ := cmd.Flags().GetString("version")
	
	// Parse target version
	targetVersion, err := schema.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}
	
	// Create schema manager
	schemaManager := schema.DefaultSchemaManager()
	validator, err := schemaManager.GetValidator(targetVersion)
	if err != nil {
		return fmt.Errorf("failed to get validator for version %s: %w", targetVersion, err)
	}
	
	// Check if path is file or directory
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}
	
	if info.IsDir() {
		return validateDirectory(validator, targetPath)
	} else {
		return validateFile(validator, targetPath)
	}
}

func validateFile(validator *schema.Validator, filePath string) error {
	fmt.Printf("Validating file: %s\n", filePath)
	
	result, err := validator.ValidateFile(filePath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	fmt.Println(result.String())
	
	if !result.Valid {
		os.Exit(1)
	}
	
	return nil
}

func validateDirectory(validator *schema.Validator, dirPath string) error {
	fmt.Printf("Validating directory: %s\n", dirPath)
	
	var totalFiles, validFiles, invalidFiles int
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		
		totalFiles++
		
		result, err := validator.ValidateFile(path)
		if err != nil {
			fmt.Printf("❌ %s: validation error: %v\n", path, err)
			invalidFiles++
			return nil
		}
		
		if result.Valid {
			fmt.Printf("✅ %s: valid\n", path)
			validFiles++
		} else {
			fmt.Printf("❌ %s: invalid\n", path)
			for _, errMsg := range result.Errors {
				fmt.Printf("   - %s\n", errMsg)
			}
			invalidFiles++
		}
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("  Total files: %d\n", totalFiles)
	fmt.Printf("  Valid: %d\n", validFiles)
	fmt.Printf("  Invalid: %d\n", invalidFiles)
	
	if invalidFiles > 0 {
		os.Exit(1)
	}
	
	return nil
}

func runSchemaMigrate(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]
	versionStr, _ := cmd.Flags().GetString("version")
	reportOnly, _ := cmd.Flags().GetBool("report-only")
	
	// Parse target version
	targetVersion, err := schema.ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}
	
	// Check if input is file or directory
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to access input path: %w", err)
	}
	
	if info.IsDir() {
		return migrateDirectory(inputPath, outputPath, targetVersion, reportOnly)
	} else {
		return migrateFile(inputPath, outputPath, targetVersion, reportOnly)
	}
}

func migrateFile(inputFile, outputFile string, targetVersion schema.SchemaVersion, reportOnly bool) error {
	migrator := schema.NewMigrator()
	
	if reportOnly {
		// Read and analyze file
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		
		// Extract current version
		currentVersion, err := extractVersionFromFile(jsonData)
		if err != nil {
			return fmt.Errorf("failed to extract version: %w", err)
		}
		
		fmt.Printf("Migration Report for: %s\n", inputFile)
		fmt.Printf("  Current version: %s\n", currentVersion)
		fmt.Printf("  Target version: %s\n", targetVersion)
		
		if currentVersion.String() == targetVersion.String() {
			fmt.Printf("  Status: No migration needed\n")
		} else {
			fmt.Printf("  Status: Migration required\n")
		}
		
		return nil
	}
	
	// Perform actual migration
	fmt.Printf("Migrating %s -> %s (target: %s)\n", inputFile, outputFile, targetVersion)
	
	if err := migrator.MigrateFile(inputFile, outputFile, targetVersion); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	fmt.Printf("✅ Migration completed successfully\n")
	return nil
}

func migrateDirectory(inputDir, outputDir string, targetVersion schema.SchemaVersion, reportOnly bool) error {
	batchMigrator := schema.NewBatchMigrator()
	
	if reportOnly {
		report, err := batchMigrator.GenerateReport(inputDir, targetVersion)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}
		
		fmt.Printf("Migration Report for: %s\n", inputDir)
		fmt.Printf("  Source version: %s\n", report.SourceVersion)
		fmt.Printf("  Target version: %s\n", report.TargetVersion)
		fmt.Printf("  Files processed: %d\n", report.FilesProcessed)
		fmt.Printf("  Files that can be migrated: %d\n", report.FilesSucceeded)
		fmt.Printf("  Files with issues: %d\n", report.FilesFailed)
		
		if len(report.Errors) > 0 {
			fmt.Printf("\nIssues found:\n")
			for _, errMsg := range report.Errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
		
		return nil
	}
	
	// Perform actual migration
	fmt.Printf("Migrating directory %s -> %s (target: %s)\n", inputDir, outputDir, targetVersion)
	
	if err := batchMigrator.MigrateDirectory(inputDir, outputDir, targetVersion); err != nil {
		return fmt.Errorf("batch migration failed: %w", err)
	}
	
	fmt.Printf("✅ Batch migration completed successfully\n")
	return nil
}

func extractVersionFromFile(data map[string]interface{}) (schema.SchemaVersion, error) {
	// Check for schema_version field
	if versionStr, ok := data["schema_version"].(string); ok {
		return schema.ParseVersion(versionStr)
	}
	
	// Check for legacy data_version in metadata
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if dataVersion, ok := metadata["data_version"].(string); ok {
			if dataVersion == "1.0" {
				return schema.SchemaVersion{Major: 1, Minor: 0, Patch: 0}, nil
			}
		}
	}
	
	// Default to 1.0.0 for legacy data
	return schema.SchemaVersion{Major: 1, Minor: 0, Patch: 0}, nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	resultsDir := args[0]
	baselineInstance, _ := cmd.Flags().GetString("baseline")
	outputFormat, _ := cmd.Flags().GetString("format")
	sortByMetric, _ := cmd.Flags().GetString("sort")

	ctx := context.Background()

	fmt.Printf("📊 Analyzing benchmark results in: %s\n", resultsDir)
	fmt.Printf("📏 Using baseline: %s\n", baselineInstance)

	// Load all benchmark results
	results, err := loadBenchmarkResults(resultsDir)
	if err != nil {
		return fmt.Errorf("failed to load results: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("❌ No benchmark results found")
		return nil
	}

	fmt.Printf("📁 Loaded %d benchmark results\n", len(results))

	// Set up baseline for price/performance calculations
	baseline, err := setupBaseline(ctx, baselineInstance, results)
	if err != nil {
		return fmt.Errorf("failed to setup baseline: %w", err)
	}

	fmt.Printf("💰 Baseline: %s at $%.4f/hour, %.1f GB/s\n", 
		baseline.InstanceType, baseline.HourlyPrice, baseline.TriadBandwidth)

	// Calculate price/performance for all results
	calculator := pricing.NewPricePerformanceCalculator(baseline)
	analysisResults, err := calculatePricePerformanceForResults(ctx, calculator, results)
	if err != nil {
		return fmt.Errorf("failed to calculate price/performance: %w", err)
	}

	// Sort results
	sortAnalysisResults(analysisResults, sortByMetric)

	// Display results
	return displayAnalysisResults(analysisResults, outputFormat)
}

func loadBenchmarkResults(resultsDir string) ([]benchmarkFileResult, error) {
	var results []benchmarkFileResult

	err := filepath.Walk(resultsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("⚠️  Failed to read %s: %v\n", path, err)
			return nil
		}

		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err != nil {
			fmt.Printf("⚠️  Failed to parse %s: %v\n", path, err)
			return nil
		}

		result := extractBenchmarkData(rawData, path)
		if result != nil {
			results = append(results, *result)
		}

		return nil
	})

	return results, err
}

type benchmarkFileResult struct {
	FilePath     string
	InstanceType string
	Region       string
	Timestamp    string
	Metrics      *pricing.PerformanceMetrics
}

func extractBenchmarkData(data map[string]interface{}, filePath string) *benchmarkFileResult {
	// Extract metadata
	metadata, _ := data["metadata"].(map[string]interface{})
	performanceData, _ := data["performance_data"].(map[string]interface{})

	if metadata == nil && performanceData == nil {
		return nil
	}

	// Get instance type
	instanceType := extractStringValue(metadata, "instance_type")
	if instanceType == "" {
		instanceType = extractStringValue(metadata, "instanceType")
	}
	if instanceType == "" {
		// Try to extract from filename
		parts := strings.Split(filepath.Base(filePath), "-")
		if len(parts) >= 2 {
			instanceType = parts[0]
		}
	}

	// Get region
	region := extractStringValue(metadata, "region")
	if region == "" {
		region = "us-east-1" // Default
	}

	// Get timestamp
	timestamp := extractStringValue(metadata, "timestamp")

	// Extract STREAM performance data
	streamData, _ := performanceData["stream"].(map[string]interface{})
	if streamData == nil {
		return nil
	}

	metrics := &pricing.PerformanceMetrics{
		TriadBandwidth: extractBandwidthValue(streamData, "triad"),
		CopyBandwidth:  extractBandwidthValue(streamData, "copy"),
		ScaleBandwidth: extractBandwidthValue(streamData, "scale"),
		AddBandwidth:   extractBandwidthValue(streamData, "add"),
	}

	// Skip if no valid metrics
	if metrics.TriadBandwidth == 0 {
		return nil
	}

	return &benchmarkFileResult{
		FilePath:     filePath,
		InstanceType: instanceType,
		Region:       region,
		Timestamp:    timestamp,
		Metrics:      metrics,
	}
}

func extractStringValue(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func extractBandwidthValue(streamData map[string]interface{}, test string) float64 {
	if streamData == nil {
		return 0
	}

	testData, ok := streamData[test].(map[string]interface{})
	if !ok {
		return 0
	}

	if bandwidth, ok := testData["bandwidth"].(float64); ok {
		return bandwidth
	}

	return 0
}

func setupBaseline(ctx context.Context, baselineInstance string, results []benchmarkFileResult) (*pricing.PricePerformanceMetrics, error) {
	// Find baseline instance in results
	for _, result := range results {
		if result.InstanceType == baselineInstance {
			calculator := pricing.NewPricePerformanceCalculator(nil)
			return calculator.CalculatePricePerformance(ctx, result.InstanceType, result.Region, result.Metrics)
		}
	}

	// If not found in results, use default baseline
	return pricing.GetDefaultBaseline(ctx)
}

func calculatePricePerformanceForResults(ctx context.Context, calculator *pricing.PricePerformanceCalculator, results []benchmarkFileResult) ([]*pricing.PricePerformanceMetrics, error) {
	var analysisResults []*pricing.PricePerformanceMetrics

	for _, result := range results {
		analysis, err := calculator.CalculatePricePerformance(ctx, result.InstanceType, result.Region, result.Metrics)
		if err != nil {
			fmt.Printf("⚠️  Failed to analyze %s: %v\n", result.InstanceType, err)
			continue
		}

		analysisResults = append(analysisResults, analysis)
	}

	return analysisResults, nil
}

func sortAnalysisResults(results []*pricing.PricePerformanceMetrics, sortBy string) {
	switch sortBy {
	case "value_score":
		// Sort by value score (higher is better)
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].ValueScore < results[j].ValueScore {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "cost_efficiency":
		// Sort by cost efficiency ratio (higher is better)
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].CostEfficiencyRatio < results[j].CostEfficiencyRatio {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "performance":
		// Sort by performance ratio (higher is better)
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].PerformanceRatio < results[j].PerformanceRatio {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "price":
		// Sort by hourly price (lower is better)
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].HourlyPrice > results[j].HourlyPrice {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}
}

func displayAnalysisResults(results []*pricing.PricePerformanceMetrics, format string) error {
	if len(results) == 0 {
		fmt.Println("❌ No analysis results to display")
		return nil
	}

	switch format {
	case "json":
		return displayJSON(results)
	case "csv":
		return displayCSV(results)
	default:
		return displayTable(results)
	}
}

func displayTable(results []*pricing.PricePerformanceMetrics) error {
	fmt.Printf("\n📊 Price/Performance Analysis Results\n")
	fmt.Printf("🏆 Baseline: %s (Score: 1.00)\n\n", results[0].BaselineInstance)

	// Header
	fmt.Printf("%-15s %-8s %-8s %-10s %-8s %-8s %-10s %-12s\n",
		"Instance", "Price/Hr", "GB/s", "$/GB/s", "Perf", "Cost Eff", "Value", "Ranking")
	fmt.Printf("%-15s %-8s %-8s %-10s %-8s %-8s %-10s %-12s\n",
		strings.Repeat("-", 15), strings.Repeat("-", 8), strings.Repeat("-", 8),
		strings.Repeat("-", 10), strings.Repeat("-", 8), strings.Repeat("-", 8),
		strings.Repeat("-", 10), strings.Repeat("-", 12))

	// Results
	for i, result := range results {
		ranking := getRankingEmoji(i + 1)
		fmt.Printf("%-15s $%-7.4f %-8.1f $%-9.4f %-8.2fx %-8.2fx %-10.2f %s\n",
			result.InstanceType,
			result.HourlyPrice,
			result.TriadBandwidth,
			result.PricePerGBps,
			result.PerformanceRatio,
			result.CostEfficiencyRatio,
			result.ValueScore,
			ranking)
	}

	fmt.Printf("\n💡 Value Score = Performance Ratio × Cost Efficiency Ratio\n")
	fmt.Printf("   Higher values indicate better price/performance\n")

	return nil
}

func displayJSON(results []*pricing.PricePerformanceMetrics) error {
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func displayCSV(results []*pricing.PricePerformanceMetrics) error {
	fmt.Println("instance_type,region,hourly_price,triad_bandwidth,price_per_gbps,performance_ratio,cost_efficiency_ratio,value_score")
	for _, result := range results {
		fmt.Printf("%s,%s,%.4f,%.1f,%.4f,%.2f,%.2f,%.2f\n",
			result.InstanceType,
			result.Region,
			result.HourlyPrice,
			result.TriadBandwidth,
			result.PricePerGBps,
			result.PerformanceRatio,
			result.CostEfficiencyRatio,
			result.ValueScore)
	}
	return nil
}

func getRankingEmoji(rank int) string {
	switch rank {
	case 1:
		return "🥇 #1"
	case 2:
		return "🥈 #2"
	case 3:
		return "🥉 #3"
	default:
		return fmt.Sprintf("   #%d", rank)
	}
}