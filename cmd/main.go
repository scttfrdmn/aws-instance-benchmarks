// Package main provides the command-line interface for AWS Instance Benchmarks.
//
// This package implements a comprehensive CLI tool for executing performance
// benchmarks across AWS EC2 instance types with intelligent orchestration,
// container management, and result storage capabilities.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/containers"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/discovery"
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

	runCmd.Flags().StringSliceVar(&instanceTypes, "instance-types", []string{"m7i.large"}, "Instance types to benchmark")
	runCmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region")
	runCmd.Flags().StringVar(&keyPair, "key-pair", "", "EC2 key pair name")
	runCmd.Flags().StringVar(&securityGroup, "security-group", "", "Security group ID")
	runCmd.Flags().StringVar(&subnet, "subnet", "", "Subnet ID")
	runCmd.Flags().BoolVar(&skipQuota, "skip-quota-check", false, "Skip quota validation before launching")
	runCmd.Flags().StringSliceVar(&benchmarkSuites, "benchmarks", []string{"stream"}, "Benchmark suites to run")

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

	registry, _ := cmd.Parent().PersistentFlags().GetString("registry")
	namespace, _ := cmd.Parent().PersistentFlags().GetString("namespace")
	if registry == "" {
		registry = "public.ecr.aws"
	}
	if namespace == "" {
		namespace = "aws-benchmarks"
	}

	fmt.Printf("Starting benchmark run for %d instance types in region %s\n", len(instanceTypes), region)

	for _, instanceType := range instanceTypes {
		for _, benchmarkSuite := range benchmarkSuites {
			fmt.Printf("\nRunning %s benchmark on %s...\n", benchmarkSuite, instanceType)
			
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

			result, err := orchestrator.RunBenchmark(ctx, config)
			if err != nil {
				if quotaErr, ok := err.(*aws.QuotaError); ok {
					fmt.Printf("⚠️  Skipping %s due to quota: %s\n", instanceType, quotaErr.Message)
					continue
				}
				fmt.Printf("❌ Failed to run benchmark on %s: %v\n", instanceType, err)
				continue
			}

			fmt.Printf("✅ Completed %s benchmark on %s (took %v)\n", 
				benchmarkSuite, instanceType, result.EndTime.Sub(result.StartTime))
			fmt.Printf("   Instance: %s, Public IP: %s\n", result.InstanceID, result.PublicIP)
		}
	}

	fmt.Println("\nBenchmark run completed!")
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

func extractInstanceFamily(instanceType string) string {
	// Simple extraction - get everything before the first dot
	parts := strings.Split(instanceType, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return instanceType
}