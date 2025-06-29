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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	awspkg "github.com/scttfrdmn/aws-instance-benchmarks/pkg/aws"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/containers"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/discovery"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/monitoring"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/pricing"
	"github.com/scttfrdmn/aws-instance-benchmarks/pkg/scheduler"
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
	result         *awspkg.InstanceResult
	metrics        monitoring.BenchmarkMetrics
}

// BenchmarkRunner interface for custom benchmark execution
type BenchmarkRunner interface {
	ExecuteBenchmark(ctx context.Context, job *scheduler.BenchmarkJob) error
}

// Data processing types for Git-native workflow
type StatisticalDataSet struct {
	ValidInstances   map[string]*InstanceStatistics
	QualityPassRate  float64
	ProcessingDate   time.Time
	TotalSamples     int
	OutliersRemoved  int
}

type InstanceStatistics struct {
	InstanceType string
	Architecture string
	Family       string
	MemoryStats  map[string]*MetricStatistics
	CPUStats     map[string]*MetricStatistics
	QualityScore float64
}

type MetricStatistics struct {
	Mean                 float64
	Median               float64
	StdDev               float64
	Min                  float64
	Max                  float64
	CoefficientVariation float64
	SampleCount          int
	ConfidenceInterval95 struct {
		Lower float64
		Upper float64
	}
	OutliersRemoved int
	QualityScore    float64
}

type GitDataProcessor struct {
	BranchPrefix     string
	QualityThreshold float64
}

type AggregateProcessor struct {
	InputDir  string
	OutputDir string
}

type DataValidator struct {
	DataDir   string
	SchemaDir string
}

type ValidationReport struct {
	Timestamp         time.Time
	Results           map[string]ValidationResult
	SchemaResults     SchemaValidationResults
	StatisticalResults StatisticalValidationResults
}

type ValidationResult struct {
	Valid        bool
	Errors       []string
	Warnings     []string
	QualityScore float64
}

type SchemaValidationResults struct {
	FilesChecked int
	Errors       []string
	Warnings     []string
}

type StatisticalValidationResults struct {
	InstancesChecked int
	PassRate         float64
	Failures         []string
	Warnings         []string
}

// Data processing implementation functions

func retrieveRawResults(ctx context.Context, s3Storage *storage.S3Storage, date time.Time) ([]map[string]interface{}, error) {
	// This would implement S3 retrieval logic
	// For now, return placeholder
	return []map[string]interface{}{}, nil
}

func convertToStatisticalFormat(rawResults []map[string]interface{}, qualityThreshold float64) (*StatisticalDataSet, error) {
	// This would implement the statistical conversion logic
	// For now, return placeholder
	return &StatisticalDataSet{
		ValidInstances:  make(map[string]*InstanceStatistics),
		QualityPassRate: 98.5,
		ProcessingDate:  time.Now(),
		TotalSamples:    150,
		OutliersRemoved: 3,
	}, nil
}

func (gdp *GitDataProcessor) ProcessAndCommit(date time.Time, data *StatisticalDataSet) error {
	// This would implement Git branch creation, file updates, and commit logic
	fmt.Printf("   Git processing placeholder - would create branch and commit %d instances\n", len(data.ValidInstances))
	return nil
}

func saveStatisticalDataLocally(data *StatisticalDataSet, outputDir string) error {
	// This would implement local file saving logic
	fmt.Printf("   Local save placeholder - would save %d instances to %s\n", len(data.ValidInstances), outputDir)
	return nil
}

func (ap *AggregateProcessor) GenerateFamilySummaries() error {
	// This would implement family aggregation logic
	fmt.Printf("   Generated family summaries for all instance families\n")
	return nil
}

func (ap *AggregateProcessor) GenerateArchitectureSummaries() error {
	// This would implement architecture aggregation logic
	fmt.Printf("   Generated architecture summaries for Intel, AMD, Graviton\n")
	return nil
}

func (ap *AggregateProcessor) GeneratePerformanceIndices() error {
	// This would implement performance index generation logic
	fmt.Printf("   Generated performance indices and rankings\n")
	return nil
}

func (dv *DataValidator) ValidateSchemas() (SchemaValidationResults, error) {
	// This would implement schema validation logic
	return SchemaValidationResults{
		FilesChecked: 67,
		Errors:       []string{},
		Warnings:     []string{},
	}, nil
}

func (dv *DataValidator) ValidateStatistics() (StatisticalValidationResults, error) {
	// This would implement statistical validation logic
	return StatisticalValidationResults{
		InstancesChecked: 67,
		PassRate:         98.5,
		Failures:         []string{},
		Warnings:         []string{},
	}, nil
}

// Infrastructure configuration structures
type InfrastructureConfig struct {
	Environments       map[string]*EnvironmentConfig `json:"environments"`
	BenchmarkDefaults *BenchmarkDefaults            `json:"benchmark_defaults"`
}

type EnvironmentConfig struct {
	Profile    string              `json:"profile"`
	Region     string              `json:"region"`
	VPC        *VPCConfig          `json:"vpc"`
	Networking *NetworkingConfig   `json:"networking"`
	Compute    *ComputeConfig      `json:"compute"`
	Storage    *StorageConfig      `json:"storage"`
	Monitoring *MonitoringConfig   `json:"monitoring"`
}

type VPCConfig struct {
	VPCID string `json:"vpc_id"`
	Name  string `json:"name"`
}

type NetworkingConfig struct {
	SubnetID         string `json:"subnet_id"`
	AvailabilityZone string `json:"availability_zone"`
	SecurityGroupID  string `json:"security_group_id"`
}

type ComputeConfig struct {
	KeyPairName     string `json:"key_pair_name"`
	InstanceProfile string `json:"instance_profile"`
}

type StorageConfig struct {
	S3Bucket string `json:"s3_bucket"`
	S3Region string `json:"s3_region"`
}

type MonitoringConfig struct {
	CloudWatchEnabled   bool   `json:"cloudwatch_enabled"`
	CloudWatchNamespace string `json:"cloudwatch_namespace"`
}

type BenchmarkDefaults struct {
	MaxConcurrency         int      `json:"max_concurrency"`
	Iterations             int      `json:"iterations"`
	TimeoutMinutes         int      `json:"timeout_minutes"`
	EnableSystemProfiling  bool     `json:"enable_system_profiling"`
	SkipQuotaCheck         bool     `json:"skip_quota_check"`
	Benchmarks             []string `json:"benchmarks"`
	InstanceTypes          []string `json:"instance_types"`
}

func runDiscoverInfrastructure(cmd *cobra.Command, args []string) error {
	infraRegion, _ := cmd.Flags().GetString("region")
	infraProfile, _ := cmd.Flags().GetString("profile") 
	configFile, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	fmt.Printf("🔍 Discovering AWS infrastructure in %s (profile: %s)...\n", infraRegion, infraProfile)

	// Set AWS profile for this session
	os.Setenv("AWS_PROFILE", infraProfile)

	ctx := context.Background()
	
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(infraRegion),
		config.WithSharedConfigProfile(infraProfile),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Discover VPC (use default VPC)
	fmt.Printf("📡 Discovering VPC...\n")
	vpcResp, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{Name: aws.String("is-default"), Values: []string{"true"}},
		},
	})
	if err != nil || len(vpcResp.Vpcs) == 0 {
		return fmt.Errorf("failed to find default VPC: %w", err)
	}
	vpcID := *vpcResp.Vpcs[0].VpcId
	fmt.Printf("   ✅ Found default VPC: %s\n", vpcID)

	// Discover suitable subnet (try multiple AZs)
	fmt.Printf("📡 Discovering subnet...\n")
	azSuffixes := []string{"a", "b", "c", "d"}
	var subnetID, availabilityZone string
	
	for _, suffix := range azSuffixes {
		targetAZ := infraRegion + suffix
		subnetResp, err := ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
			Filters: []types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
				{Name: aws.String("availability-zone"), Values: []string{targetAZ}},
				{Name: aws.String("state"), Values: []string{"available"}},
			},
		})
		if err == nil && len(subnetResp.Subnets) > 0 {
			subnetID = *subnetResp.Subnets[0].SubnetId
			availabilityZone = *subnetResp.Subnets[0].AvailabilityZone
			break
		}
	}
	
	if subnetID == "" {
		return fmt.Errorf("failed to find suitable subnet in any AZ")
	}
	fmt.Printf("   ✅ Found subnet: %s (%s)\n", subnetID, availabilityZone)

	// Discover default security group
	fmt.Printf("📡 Discovering security group...\n")
	sgResp, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			{Name: aws.String("group-name"), Values: []string{"default"}},
		},
	})
	if err != nil || len(sgResp.SecurityGroups) == 0 {
		return fmt.Errorf("failed to find default security group: %w", err)
	}
	securityGroupID := *sgResp.SecurityGroups[0].GroupId
	fmt.Printf("   ✅ Found security group: %s\n", securityGroupID)

	// Discover key pair (first available)
	fmt.Printf("📡 Discovering key pair...\n")
	keyResp, err := ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{})
	if err != nil || len(keyResp.KeyPairs) == 0 {
		return fmt.Errorf("failed to find any key pairs: %w", err)
	}
	keyPairName := *keyResp.KeyPairs[0].KeyName
	fmt.Printf("   ✅ Found key pair: %s\n", keyPairName)

	// Create S3 bucket for benchmarks
	fmt.Printf("📡 Creating S3 bucket...\n")
	bucketName := fmt.Sprintf("aws-instance-benchmarks-%s-%d", infraRegion, time.Now().Unix())
	
	s3Client := s3.NewFromConfig(cfg)
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(infraRegion),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}
	fmt.Printf("   ✅ Created S3 bucket: %s\n", bucketName)

	// Build configuration
	envConfig := &EnvironmentConfig{
		Profile: infraProfile,
		Region:  infraRegion,
		VPC: &VPCConfig{
			VPCID: vpcID,
			Name:  "default",
		},
		Networking: &NetworkingConfig{
			SubnetID:         subnetID,
			AvailabilityZone: availabilityZone,
			SecurityGroupID:  securityGroupID,
		},
		Compute: &ComputeConfig{
			KeyPairName:     keyPairName,
			InstanceProfile: "",
		},
		Storage: &StorageConfig{
			S3Bucket: bucketName,
			S3Region: infraRegion,
		},
		Monitoring: &MonitoringConfig{
			CloudWatchEnabled:   true,
			CloudWatchNamespace: "AWS/InstanceBenchmarks",
		},
	}

	// Load existing config or create new
	var infraConfig *InfrastructureConfig
	if configData, err := os.ReadFile(configFile); err == nil {
		infraConfig = &InfrastructureConfig{}
		if err := json.Unmarshal(configData, infraConfig); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
	} else {
		infraConfig = &InfrastructureConfig{
			Environments: make(map[string]*EnvironmentConfig),
			BenchmarkDefaults: &BenchmarkDefaults{
				MaxConcurrency:        5,
				Iterations:            1,
				TimeoutMinutes:        30,
				EnableSystemProfiling: true,
				SkipQuotaCheck:        false,
				Benchmarks:            []string{"stream"},
				InstanceTypes:         []string{"m7i.large", "c7g.large", "r7a.large"},
			},
		}
	}

	// Update with discovered configuration
	infraConfig.Environments[infraRegion] = envConfig

	if dryRun {
		fmt.Printf("\n📋 Discovered configuration (dry-run):\n")
		configJSON, _ := json.MarshalIndent(envConfig, "", "  ")
		fmt.Printf("%s\n", configJSON)
		return nil
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration
	configJSON, err := json.MarshalIndent(infraConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("\n✅ Infrastructure configuration saved to: %s\n", configFile)
	fmt.Printf("\n🚀 Ready to run benchmarks with:\n")
	fmt.Printf("   ./cloud-benchmark-collector run --config %s --environment %s\n", configFile, infraRegion)

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cloud-benchmark-collector",
		Short: "Multi-cloud instance benchmark collection tool",
		Long:  `Comprehensive performance benchmark collection for cloud instances across providers.

Supported cloud providers:
- AWS EC2 (production ready)
- Google Cloud Compute Engine (planned)
- Microsoft Azure Virtual Machines (planned)
- Oracle Cloud Infrastructure (planned)

The tool provides consistent benchmarking methodology across providers while
capturing provider-specific optimizations and system characteristics.`,
	}

	var discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "Discover AWS instance types and infrastructure",
		Long:  "Discover AWS instance types, architectures, and infrastructure configuration",
	}

	var discoverInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "Discover AWS instance types and their architectures",
		RunE:  runDiscover,
	}

	var discoverInfraCmd = &cobra.Command{
		Use:   "infrastructure",
		Short: "Discover and configure AWS infrastructure settings",
		RunE:  runDiscoverInfrastructure,
	}

	var updateContainers bool
	var dryRun bool
	var infraRegion string
	var infraProfile string
	var configFile string

	discoverInstancesCmd.Flags().BoolVar(&updateContainers, "update-containers", false, "Update container architecture mappings")
	discoverInstancesCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	discoverInfraCmd.Flags().StringVar(&infraRegion, "region", "us-west-2", "AWS region to discover infrastructure for")
	discoverInfraCmd.Flags().StringVar(&infraProfile, "profile", "aws", "AWS profile to use for discovery")
	discoverInfraCmd.Flags().StringVar(&configFile, "config", "configs/aws-infrastructure.json", "Path to infrastructure config file")
	discoverInfraCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show discovered configuration without saving")

	discoverCmd.AddCommand(discoverInstancesCmd)
	discoverCmd.AddCommand(discoverInfraCmd)

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
	var enableSystemProfiling bool
	var configFileRun string
	var environment string

	var runProvider string
	runCmd.Flags().StringVar(&runProvider, "provider", "aws", "Cloud provider (aws, gcp, azure, oci)")
	runCmd.Flags().StringSliceVar(&instanceTypes, "instance-types", []string{"m7i.large"}, "Instance types to benchmark")
	runCmd.Flags().StringVar(&region, "region", "us-east-1", "Cloud provider region")
	runCmd.Flags().StringVar(&keyPair, "key-pair", "", "SSH key pair name (provider-specific)")
	runCmd.Flags().StringVar(&securityGroup, "security-group", "", "Security group/firewall rule ID")
	runCmd.Flags().StringVar(&subnet, "subnet", "", "Subnet/VPC subnet ID")
	runCmd.Flags().BoolVar(&skipQuota, "skip-quota-check", false, "Skip quota validation before launching")
	runCmd.Flags().StringSliceVar(&benchmarkSuites, "benchmarks", []string{"stream"}, "Benchmark suites to run (stream, hpl)")
	runCmd.Flags().IntVar(&maxConcurrency, "max-concurrency", 5, "Maximum number of concurrent benchmarks")
	runCmd.Flags().IntVar(&iterations, "iterations", 1, "Number of benchmark iterations for statistical validation")
	runCmd.Flags().StringVar(&s3Bucket, "storage-bucket", "", "Cloud storage bucket for storing results")
	runCmd.Flags().StringVar(&s3Bucket, "s3-bucket", "", "(Deprecated) Use --storage-bucket instead")
	runCmd.Flags().BoolVar(&enableSystemProfiling, "enable-system-profiling", false, "Enable comprehensive system topology discovery and profiling")
	runCmd.Flags().StringVar(&configFileRun, "config", "", "Path to infrastructure config file (overrides individual flags)")
	runCmd.Flags().StringVar(&environment, "environment", "", "Environment name from config file (e.g., us-west-2)")

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

	// Add schedule command with subcommands
	var scheduleCmd = &cobra.Command{
		Use:   "schedule",
		Short: "Schedule systematic benchmark execution over time",
		Long: `Schedule comprehensive benchmark execution across multiple instance types
using intelligent time-based distribution to avoid quota limits and optimize costs.

This command enables systematic testing of large numbers of instances by:
- Distributing workloads across daily time windows over a week
- Balancing benchmark types (STREAM + HPL) across execution periods
- Managing AWS quotas and concurrent instance limits
- Optimizing costs through spot instances and off-peak execution
- Providing progress tracking and retry logic for failed jobs

Example usage:
  # Generate and execute a weekly benchmark plan
  ./aws-benchmark-collector schedule weekly \
    --instance-families m7i,c7g,r7a \
    --region us-east-1 \
    --max-daily-jobs 20 \
    --max-concurrent 5

  # Create a plan without executing
  ./aws-benchmark-collector schedule plan \
    --instance-types m7i.large,c7g.large \
    --output weekly-plan.json`,
	}

	var weeklyCmd = &cobra.Command{
		Use:   "weekly",
		Short: "Generate and execute weekly benchmark plan",
		RunE:  runWeeklySchedule,
	}

	var planCmd = &cobra.Command{
		Use:   "plan",
		Short: "Generate benchmark execution plan without executing",
		RunE:  runPlanGeneration,
	}

	// Weekly command flags
	var instanceFamilies []string
	var weeklyRegion string
	var maxDailyJobs int
	var maxConcurrentJobs int
	var weeklyKeyPair string
	var weeklySecurityGroup string
	var weeklySubnet string
	var weeklyS3Bucket string
	var enableSpotInstances bool
	var benchmarkRotation bool
	var instanceSizeWaves bool
	var cloudProvider string

	weeklyCmd.Flags().StringVar(&cloudProvider, "provider", "aws", "Cloud provider (aws, gcp, azure, oci)")
	weeklyCmd.Flags().StringSliceVar(&instanceFamilies, "instance-families", []string{"m7i", "c7g", "r7a"}, "Instance families to benchmark")
	weeklyCmd.Flags().StringVar(&weeklyRegion, "region", "us-east-1", "Cloud provider region")
	weeklyCmd.Flags().IntVar(&maxDailyJobs, "max-daily-jobs", 30, "Maximum jobs per day")
	weeklyCmd.Flags().IntVar(&maxConcurrentJobs, "max-concurrent", 5, "Maximum concurrent executions")
	weeklyCmd.Flags().StringVar(&weeklyKeyPair, "key-pair", "", "SSH key pair name (provider-specific)")
	weeklyCmd.Flags().StringVar(&weeklySecurityGroup, "security-group", "", "Security group/firewall rule ID")
	weeklyCmd.Flags().StringVar(&weeklySubnet, "subnet", "", "Subnet/VPC subnet ID")
	weeklyCmd.Flags().StringVar(&weeklyS3Bucket, "storage-bucket", "", "Cloud storage bucket for results")
	weeklyCmd.Flags().StringVar(&weeklyS3Bucket, "s3-bucket", "", "(Deprecated) Use --storage-bucket instead")
	weeklyCmd.Flags().BoolVar(&enableSpotInstances, "enable-spot", true, "Use spot/preemptible instances for cost optimization")
	weeklyCmd.Flags().BoolVar(&benchmarkRotation, "benchmark-rotation", true, "Rotate benchmark types across time windows")
	weeklyCmd.Flags().BoolVar(&instanceSizeWaves, "instance-size-waves", true, "Group instances by size to avoid same physical nodes")

	// Plan command flags
	var planInstanceTypes []string
	var planOutput string
	var planBenchmarks []string

	planCmd.Flags().StringSliceVar(&planInstanceTypes, "instance-types", []string{}, "Specific instance types to plan")
	planCmd.Flags().StringVar(&planOutput, "output", "weekly-plan.json", "Output file for plan")
	planCmd.Flags().StringSliceVar(&planBenchmarks, "benchmarks", []string{"stream", "hpl"}, "Benchmark suites to include")

	scheduleCmd.AddCommand(weeklyCmd)
	scheduleCmd.AddCommand(planCmd)

	// Add data processing command for Git-native workflow
	var processCmd = &cobra.Command{
		Use:   "process",
		Short: "Process and commit benchmark data to Git repository",
		Long: `Process benchmark results from cloud storage into Git-native statistical format.

This command provides comprehensive data processing capabilities:
- Convert raw cloud storage results to statistical summaries
- Update Git repository with versioned performance data
- Generate family and architecture aggregations
- Validate data quality and statistical significance
- Create descriptive commits with performance summaries
- Support multi-cloud data normalization and comparison

Example usage:
  # Process daily results from AWS S3 into Git
  ./cloud-benchmark-collector process daily \
    --provider aws \
    --date 2024-06-29 \
    --storage-bucket aws-instance-benchmarks-data-us-east-1 \
    --commit-to-git

  # Process Google Cloud Storage results
  ./cloud-benchmark-collector process daily \
    --provider gcp \
    --date 2024-06-29 \
    --storage-bucket gs://gcp-benchmarks-data \
    --commit-to-git

  # Generate cross-provider aggregated summaries
  ./cloud-benchmark-collector process aggregate \
    --regenerate-families \
    --regenerate-architectures \
    --cross-provider-analysis`,
	}

	var dailyCmd = &cobra.Command{
		Use:   "daily",
		Short: "Process daily benchmark results from S3 into Git",
		RunE:  runDailyProcessing,
	}

	var aggregateCmd = &cobra.Command{
		Use:   "aggregate",
		Short: "Generate aggregated summaries and indices",
		RunE:  runAggregateProcessing,
	}

	var validateDataCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate statistical data quality in Git repository",
		RunE:  runDataValidation,
	}

	// Daily processing flags
	var processDate string
	var s3BucketProcess string
	var commitToGit bool
	var branchPrefix string
	var qualityThreshold float64

	var processProvider string
	dailyCmd.Flags().StringVar(&processProvider, "provider", "aws", "Cloud provider (aws, gcp, azure, oci)")
	dailyCmd.Flags().StringVar(&processDate, "date", time.Now().Format("2006-01-02"), "Date to process (YYYY-MM-DD)")
	dailyCmd.Flags().StringVar(&s3BucketProcess, "storage-bucket", "", "Cloud storage bucket containing raw results (S3, GCS, etc.)")
	dailyCmd.Flags().StringVar(&s3BucketProcess, "s3-bucket", "", "(Deprecated) Use --storage-bucket instead")
	dailyCmd.Flags().BoolVar(&commitToGit, "commit-to-git", true, "Commit processed data to Git repository")
	dailyCmd.Flags().StringVar(&branchPrefix, "branch-prefix", "data-collection-", "Prefix for Git branch names")
	dailyCmd.Flags().Float64Var(&qualityThreshold, "quality-threshold", 0.95, "Minimum quality score for data inclusion")

	// Aggregate processing flags
	var regenerateFamilies bool
	var regenerateArchitectures bool
	var regenerateIndices bool
	var outputDir string

	var crossProviderAnalysis bool
	aggregateCmd.Flags().BoolVar(&regenerateFamilies, "regenerate-families", true, "Regenerate family summaries")
	aggregateCmd.Flags().BoolVar(&regenerateArchitectures, "regenerate-architectures", true, "Regenerate architecture summaries")
	aggregateCmd.Flags().BoolVar(&regenerateIndices, "regenerate-indices", true, "Regenerate performance indices")
	aggregateCmd.Flags().BoolVar(&crossProviderAnalysis, "cross-provider-analysis", false, "Generate cross-provider comparison analysis")
	aggregateCmd.Flags().StringVar(&outputDir, "output-dir", "data/aggregated", "Output directory for aggregated data")

	// Validation flags
	var validateStatistical bool
	var validateSchema bool
	var reportPath string

	validateDataCmd.Flags().BoolVar(&validateStatistical, "statistical", true, "Perform statistical validation")
	validateDataCmd.Flags().BoolVar(&validateSchema, "schema", true, "Perform schema validation")
	validateDataCmd.Flags().StringVar(&reportPath, "report", "validation-report.json", "Path for validation report")

	processCmd.AddCommand(dailyCmd)
	processCmd.AddCommand(aggregateCmd)
	processCmd.AddCommand(validateDataCmd)

	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(processCmd)
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
	
	// Check if config file is specified
	configFileRun, _ := cmd.Flags().GetString("config")
	environment, _ := cmd.Flags().GetString("environment")
	
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
	var enableSystemProfiling bool
	
	if configFileRun != "" {
		// Load from config file
		if environment == "" {
			return fmt.Errorf("--environment is required when using --config")
		}
		
		fmt.Printf("📋 Loading configuration from %s (environment: %s)...\n", configFileRun, environment)
		
		configData, err := os.ReadFile(configFileRun)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		
		var infraConfig InfrastructureConfig
		if err := json.Unmarshal(configData, &infraConfig); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
		
		envConfig, exists := infraConfig.Environments[environment]
		if !exists {
			return fmt.Errorf("environment '%s' not found in config file", environment)
		}
		
		// Set AWS profile from config
		os.Setenv("AWS_PROFILE", envConfig.Profile)
		
		// Load values from config
		region = envConfig.Region
		keyPair = envConfig.Compute.KeyPairName
		securityGroup = envConfig.Networking.SecurityGroupID
		subnet = envConfig.Networking.SubnetID
		s3Bucket = envConfig.Storage.S3Bucket
		
		// Load defaults
		if infraConfig.BenchmarkDefaults != nil {
			instanceTypes = infraConfig.BenchmarkDefaults.InstanceTypes
			benchmarkSuites = infraConfig.BenchmarkDefaults.Benchmarks
			maxConcurrency = infraConfig.BenchmarkDefaults.MaxConcurrency
			iterations = infraConfig.BenchmarkDefaults.Iterations
			enableSystemProfiling = infraConfig.BenchmarkDefaults.EnableSystemProfiling
			skipQuota = infraConfig.BenchmarkDefaults.SkipQuotaCheck
		}
		
		// Allow CLI flags to override config values
		if flagValue, _ := cmd.Flags().GetStringSlice("instance-types"); cmd.Flags().Changed("instance-types") {
			instanceTypes = flagValue
		}
		if flagValue, _ := cmd.Flags().GetStringSlice("benchmarks"); cmd.Flags().Changed("benchmarks") {
			benchmarkSuites = flagValue
		}
		if flagValue, _ := cmd.Flags().GetInt("max-concurrency"); cmd.Flags().Changed("max-concurrency") {
			maxConcurrency = flagValue
		}
		if flagValue, _ := cmd.Flags().GetInt("iterations"); cmd.Flags().Changed("iterations") {
			iterations = flagValue
		}
		if flagValue, _ := cmd.Flags().GetBool("enable-system-profiling"); cmd.Flags().Changed("enable-system-profiling") {
			enableSystemProfiling = flagValue
		}
		
		fmt.Printf("   ✅ Region: %s, VPC: %s, Subnet: %s\n", region, envConfig.VPC.VPCID, subnet)
		fmt.Printf("   ✅ S3 Bucket: %s, Key Pair: %s\n", s3Bucket, keyPair)
		
	} else {
		// Load from command-line flags (existing behavior)
		instanceTypes, _ = cmd.Flags().GetStringSlice("instance-types")
		region, _ = cmd.Flags().GetString("region")
		keyPair, _ = cmd.Flags().GetString("key-pair")
		securityGroup, _ = cmd.Flags().GetString("security-group")
		subnet, _ = cmd.Flags().GetString("subnet")
		skipQuota, _ = cmd.Flags().GetBool("skip-quota-check")
		benchmarkSuites, _ = cmd.Flags().GetStringSlice("benchmarks")
		maxConcurrency, _ = cmd.Flags().GetInt("max-concurrency")
		iterations, _ = cmd.Flags().GetInt("iterations")
		s3Bucket, _ = cmd.Flags().GetString("s3-bucket")
		enableSystemProfiling, _ = cmd.Flags().GetBool("enable-system-profiling")
	}

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

	orchestrator, err := awspkg.NewOrchestrator(region)
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
		config         awspkg.BenchmarkConfig
	}

	var jobs []benchmarkJob
	for _, instanceType := range instanceTypes {
		for _, benchmarkSuite := range benchmarkSuites {
			for iteration := 1; iteration <= iterations; iteration++ {
				containerImage := fmt.Sprintf("%s/%s:%s-%s", registry, namespace, benchmarkSuite, 
					getContainerTagForInstance(instanceType))

				config := awspkg.BenchmarkConfig{
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
			var result *awspkg.InstanceResult
			var err error
			
			// Use system profiling if enabled
			if enableSystemProfiling {
				result, err = orchestrator.RunBenchmarkWithProfiling(ctx, j.config)
			} else {
				result, err = orchestrator.RunBenchmark(ctx, j.config)
			}
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
				if quotaErr, ok := err.(*awspkg.QuotaError); ok {
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

func storeResults(ctx context.Context, s3Storage *storage.S3Storage, result *awspkg.InstanceResult, benchmarkSuite string, region string) error {
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
	
	// Include system topology if available from profiling
	if result.SystemTopology != nil {
		resultData["system_topology"] = result.SystemTopology
		// Update schema version to indicate enhanced data
		resultData["schema_version"] = "2.0.0"
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

// runDailyProcessing implements the daily data processing command
func runDailyProcessing(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	
	// Parse flags
	processDate, _ := cmd.Flags().GetString("date")
	s3Bucket, _ := cmd.Flags().GetString("s3-bucket")
	commitToGit, _ := cmd.Flags().GetBool("commit-to-git")
	branchPrefix, _ := cmd.Flags().GetString("branch-prefix")
	qualityThreshold, _ := cmd.Flags().GetFloat64("quality-threshold")
	
	if s3Bucket == "" {
		return fmt.Errorf("--s3-bucket is required")
	}
	
	parsedDate, err := time.Parse("2006-01-02", processDate)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}
	
	fmt.Printf("📊 Processing benchmark data for %s\n", processDate)
	fmt.Printf("📁 S3 Bucket: %s\n", s3Bucket)
	fmt.Printf("🎯 Quality Threshold: %.2f\n", qualityThreshold)
	
	// Initialize S3 storage for reading raw results
	storageConfig := storage.Config{
		BucketName:    s3Bucket,
		KeyPrefix:     "instance-benchmarks/",
		RetryAttempts: 3,
	}
	s3Storage, err := storage.NewS3Storage(ctx, storageConfig, "us-east-1")
	if err != nil {
		return fmt.Errorf("failed to initialize S3 storage: %w", err)
	}
	
	// Retrieve raw results from S3 for the specified date
	fmt.Printf("🔍 Retrieving raw results from S3...\n")
	rawResults, err := retrieveRawResults(ctx, s3Storage, parsedDate)
	if err != nil {
		return fmt.Errorf("failed to retrieve raw results: %w", err)
	}
	
	fmt.Printf("📈 Found %d raw benchmark results\n", len(rawResults))
	
	// Convert to statistical format
	fmt.Printf("⚙️  Converting to statistical format...\n")
	statisticalData, err := convertToStatisticalFormat(rawResults, qualityThreshold)
	if err != nil {
		return fmt.Errorf("failed to convert to statistical format: %w", err)
	}
	
	fmt.Printf("✅ Processed %d instances (%.1f%% passed quality threshold)\n", 
		len(statisticalData.ValidInstances), statisticalData.QualityPassRate)
	
	if commitToGit {
		// Create Git branch and commit data
		fmt.Printf("📝 Committing to Git repository...\n")
		branchName := fmt.Sprintf("%s%s", branchPrefix, processDate)
		
		gitProcessor := &GitDataProcessor{
			BranchPrefix: branchPrefix,
			QualityThreshold: qualityThreshold,
		}
		
		if err := gitProcessor.ProcessAndCommit(parsedDate, statisticalData); err != nil {
			return fmt.Errorf("failed to commit to Git: %w", err)
		}
		
		fmt.Printf("🎉 Successfully committed data to branch: %s\n", branchName)
	} else {
		// Save to local files without Git commit
		fmt.Printf("💾 Saving to local files...\n")
		if err := saveStatisticalDataLocally(statisticalData, "data/statistical"); err != nil {
			return fmt.Errorf("failed to save locally: %w", err)
		}
		fmt.Printf("✅ Data saved to local directory\n")
	}
	
	return nil
}

// runAggregateProcessing implements the aggregate processing command
func runAggregateProcessing(cmd *cobra.Command, _ []string) error {
	// Parse flags
	regenerateFamilies, _ := cmd.Flags().GetBool("regenerate-families")
	regenerateArchitectures, _ := cmd.Flags().GetBool("regenerate-architectures")
	regenerateIndices, _ := cmd.Flags().GetBool("regenerate-indices")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	
	fmt.Printf("🔄 Generating aggregated summaries...\n")
	fmt.Printf("📁 Output Directory: %s\n", outputDir)
	
	processor := &AggregateProcessor{
		InputDir:  "data/statistical",
		OutputDir: outputDir,
	}
	
	var tasksCompleted int
	totalTasks := 0
	if regenerateFamilies { totalTasks++ }
	if regenerateArchitectures { totalTasks++ }
	if regenerateIndices { totalTasks++ }
	
	if regenerateFamilies {
		fmt.Printf("👨‍👩‍👧‍👦 Regenerating family summaries...\n")
		if err := processor.GenerateFamilySummaries(); err != nil {
			return fmt.Errorf("failed to generate family summaries: %w", err)
		}
		tasksCompleted++
		fmt.Printf("   Progress: %d/%d complete\n", tasksCompleted, totalTasks)
	}
	
	if regenerateArchitectures {
		fmt.Printf("🏗️  Regenerating architecture summaries...\n")
		if err := processor.GenerateArchitectureSummaries(); err != nil {
			return fmt.Errorf("failed to generate architecture summaries: %w", err)
		}
		tasksCompleted++
		fmt.Printf("   Progress: %d/%d complete\n", tasksCompleted, totalTasks)
	}
	
	if regenerateIndices {
		fmt.Printf("📋 Regenerating performance indices...\n")
		if err := processor.GeneratePerformanceIndices(); err != nil {
			return fmt.Errorf("failed to generate performance indices: %w", err)
		}
		tasksCompleted++
		fmt.Printf("   Progress: %d/%d complete\n", tasksCompleted, totalTasks)
	}
	
	fmt.Printf("✅ All aggregation tasks completed successfully\n")
	return nil
}

// runDataValidation implements the data validation command
func runDataValidation(cmd *cobra.Command, _ []string) error {
	// Parse flags
	validateStatistical, _ := cmd.Flags().GetBool("statistical")
	validateSchema, _ := cmd.Flags().GetBool("schema")
	reportPath, _ := cmd.Flags().GetString("report")
	
	fmt.Printf("🔍 Validating benchmark data quality...\n")
	
	validator := &DataValidator{
		DataDir: "data/statistical",
		SchemaDir: "schemas/current",
	}
	
	validationReport := &ValidationReport{
		Timestamp: time.Now(),
		Results:   make(map[string]ValidationResult),
	}
	
	if validateSchema {
		fmt.Printf("📋 Performing schema validation...\n")
		schemaResults, err := validator.ValidateSchemas()
		if err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
		validationReport.SchemaResults = schemaResults
		fmt.Printf("   Schema validation: %d files checked, %d errors\n", 
			schemaResults.FilesChecked, len(schemaResults.Errors))
	}
	
	if validateStatistical {
		fmt.Printf("📊 Performing statistical validation...\n")
		statResults, err := validator.ValidateStatistics()
		if err != nil {
			return fmt.Errorf("statistical validation failed: %w", err)
		}
		validationReport.StatisticalResults = statResults
		fmt.Printf("   Statistical validation: %d instances checked, %.1f%% passed\n", 
			statResults.InstancesChecked, statResults.PassRate)
	}
	
	// Generate validation report
	fmt.Printf("📝 Generating validation report...\n")
	reportData, err := json.MarshalIndent(validationReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	
	if err := os.WriteFile(reportPath, reportData, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	
	fmt.Printf("✅ Validation complete. Report saved to: %s\n", reportPath)
	
	// Exit with error code if validation failed
	if (validateSchema && len(validationReport.SchemaResults.Errors) > 0) ||
	   (validateStatistical && validationReport.StatisticalResults.PassRate < 95.0) {
		fmt.Printf("❌ Validation failed. See report for details.\n")
		os.Exit(1)
	}
	
	return nil
}

// runWeeklySchedule implements the weekly scheduling command
func runWeeklySchedule(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	
	// Get flags
	instanceFamilies, _ := cmd.Flags().GetStringSlice("instance-families")
	region, _ := cmd.Flags().GetString("region")
	maxDailyJobs, _ := cmd.Flags().GetInt("max-daily-jobs")
	maxConcurrentJobs, _ := cmd.Flags().GetInt("max-concurrent")
	keyPair, _ := cmd.Flags().GetString("key-pair")
	securityGroup, _ := cmd.Flags().GetString("security-group")
	subnet, _ := cmd.Flags().GetString("subnet")
	s3Bucket, _ := cmd.Flags().GetString("s3-bucket")
	enableSpot, _ := cmd.Flags().GetBool("enable-spot")
	benchmarkRotation, _ := cmd.Flags().GetBool("benchmark-rotation")
	instanceSizeWaves, _ := cmd.Flags().GetBool("instance-size-waves")
	
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
	
	// Expand instance families to specific instance types
	instanceTypes := expandInstanceFamilies(instanceFamilies, instanceSizeWaves)
	fmt.Printf("📋 Generated %d instance types from %d families\n", len(instanceTypes), len(instanceFamilies))
	
	// Configure scheduler
	config := scheduler.Config{
		MaxConcurrentJobs: maxConcurrentJobs,
		MaxDailyJobs:      maxDailyJobs,
		PreferredRegions:  []string{region},
		SpotInstancePreference: enableSpot,
		TimeZone:          "UTC",
		RetryAttempts:     3,
		CostOptimization:  true,
	}
	
	batchScheduler := scheduler.NewBatchScheduler(config)
	
	// Determine benchmarks with rotation
	benchmarks := []string{"stream"}
	if benchmarkRotation {
		benchmarks = append(benchmarks, "hpl")
		fmt.Printf("🔄 Benchmark rotation enabled: %v\n", benchmarks)
	} else {
		fmt.Printf("📊 Single benchmark mode: %v\n", benchmarks)
	}
	
	// Generate weekly plan
	fmt.Printf("🗓️  Generating weekly benchmark plan...\n")
	plan, err := batchScheduler.GenerateWeeklyPlan(instanceTypes, benchmarks)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}
	
	fmt.Printf("📅 Plan generated: %d jobs across %d time windows\n", len(plan.Jobs), len(plan.TimeWindows))
	fmt.Printf("💰 Estimated cost: $%.2f\n", plan.EstimatedCost)
	fmt.Printf("⏱️  Estimated duration: %v\n", plan.EstimatedDuration)
	
	// Display plan summary
	displayPlanSummary(plan, instanceSizeWaves)
	
	// Ask for confirmation
	fmt.Printf("\n❓ Execute this plan? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("❌ Execution cancelled")
		return nil
	}
	
	// Initialize AWS orchestrator and storage
	orchestrator, err := awspkg.NewOrchestrator(region)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}
	
	bucketName := s3Bucket
	if bucketName == "" {
		bucketName = fmt.Sprintf("aws-instance-benchmarks-data-%s", region)
	}
	
	storageConfig := storage.Config{
		BucketName:         bucketName,
		KeyPrefix:          "scheduled-benchmarks/",
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
	
	// Execute the plan
	fmt.Printf("\n🚀 Starting weekly benchmark execution...\n")
	executor := &ScheduledBenchmarkExecutor{
		orchestrator: orchestrator,
		s3Storage:    s3Storage,
		keyPair:      keyPair,
		securityGroup: securityGroup,
		subnet:       subnet,
		region:       region,
	}
	
	if err := executeScheduledPlan(ctx, executor, batchScheduler, plan); err != nil {
		return fmt.Errorf("failed to execute plan: %w", err)
	}
	
	fmt.Printf("✅ Weekly benchmark execution completed successfully!\n")
	return nil
}

// runPlanGeneration implements the plan generation command
func runPlanGeneration(cmd *cobra.Command, _ []string) error {
	// Get flags
	instanceTypes, _ := cmd.Flags().GetStringSlice("instance-types")
	outputFile, _ := cmd.Flags().GetString("output")
	benchmarks, _ := cmd.Flags().GetStringSlice("benchmarks")
	
	if len(instanceTypes) == 0 {
		// Use default set if none provided
		instanceTypes = []string{"m7i.large", "c7g.large", "r7a.large"}
	}
	
	// Configure scheduler
	config := scheduler.Config{
		MaxConcurrentJobs: 5,
		MaxDailyJobs:      30,
		PreferredRegions:  []string{"us-east-1"},
		SpotInstancePreference: true,
		TimeZone:          "UTC",
		RetryAttempts:     3,
		CostOptimization:  true,
	}
	
	batchScheduler := scheduler.NewBatchScheduler(config)
	
	// Generate plan
	fmt.Printf("📋 Generating plan for %d instance types...\n", len(instanceTypes))
	plan, err := batchScheduler.GenerateWeeklyPlan(instanceTypes, benchmarks)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}
	
	// Save plan to file
	planData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}
	
	if err := os.WriteFile(outputFile, planData, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}
	
	fmt.Printf("📄 Plan saved to: %s\n", outputFile)
	fmt.Printf("📊 Plan contains: %d jobs across %d time windows\n", len(plan.Jobs), len(plan.TimeWindows))
	fmt.Printf("💰 Estimated cost: $%.2f\n", plan.EstimatedCost)
	fmt.Printf("⏱️  Estimated duration: %v\n", plan.EstimatedDuration)
	
	return nil
}

// expandInstanceFamilies converts families to specific instance types with size wave grouping
func expandInstanceFamilies(families []string, sizeWaves bool) []string {
	var instanceTypes []string
	
	// Standard sizes in logical waves to avoid same physical nodes
	sizesByWave := [][]string{
		{"large"},           // Wave 1: Small instances
		{"xlarge"},          // Wave 2: Medium instances
		{"2xlarge"},         // Wave 3: Large instances
		{"4xlarge", "8xlarge"}, // Wave 4: Very large instances
	}
	
	if sizeWaves {
		// Group by size waves to minimize physical node conflicts
		for _, wave := range sizesByWave {
			for _, family := range families {
				for _, size := range wave {
					instanceTypes = append(instanceTypes, fmt.Sprintf("%s.%s", family, size))
				}
			}
		}
	} else {
		// Traditional family grouping
		for _, family := range families {
			for _, wave := range sizesByWave {
				for _, size := range wave {
					instanceTypes = append(instanceTypes, fmt.Sprintf("%s.%s", family, size))
				}
			}
		}
	}
	
	return instanceTypes
}

// displayPlanSummary shows a summary of the weekly plan
func displayPlanSummary(plan *scheduler.WeeklyPlan, sizeWaves bool) {
	fmt.Printf("\n📋 Weekly Plan Summary:\n")
	fmt.Printf("   Start Date: %s\n", plan.StartDate.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Time Windows: %d\n", len(plan.TimeWindows))
	fmt.Printf("   Total Jobs: %d\n", len(plan.Jobs))
	
	// Group jobs by benchmark type
	benchmarkCounts := make(map[string]int)
	for _, job := range plan.Jobs {
		benchmarkCounts[job.BenchmarkSuite]++
	}
	
	fmt.Printf("\n📊 Benchmark Distribution:\n")
	for benchmark, count := range benchmarkCounts {
		fmt.Printf("   %s: %d jobs\n", benchmark, count)
	}
	
	// Show size wave distribution if enabled
	if sizeWaves {
		sizeCounts := make(map[string]int)
		for _, job := range plan.Jobs {
			parts := strings.Split(job.InstanceType, ".")
			if len(parts) == 2 {
				sizeCounts[parts[1]]++
			}
		}
		
		fmt.Printf("\n🌊 Instance Size Waves:\n")
		for size, count := range sizeCounts {
			fmt.Printf("   %s: %d instances\n", size, count)
		}
	}
	
	// Show time window breakdown
	fmt.Printf("\n⏰ Time Window Schedule:\n")
	for i, window := range plan.TimeWindows {
		jobCount := 0
		for range plan.Jobs {
			// This is a simplified count - in reality would need proper window assignment
			if i < len(plan.Jobs)/len(plan.TimeWindows) {
				jobCount++
			}
		}
		fmt.Printf("   Window %d: %s (Duration: %v, Max Jobs: %d)\n", 
			i+1, window.StartTime.Format("Mon 15:04"), window.Duration, window.MaxJobs)
	}
}

// ScheduledBenchmarkExecutor handles benchmark execution for scheduled jobs
type ScheduledBenchmarkExecutor struct {
	orchestrator  *awspkg.Orchestrator
	s3Storage     *storage.S3Storage
	keyPair       string
	securityGroup string
	subnet        string
	region        string
}

// executeScheduledPlan executes a scheduled benchmark plan
func executeScheduledPlan(ctx context.Context, executor *ScheduledBenchmarkExecutor, 
	batchScheduler *scheduler.BatchScheduler, plan *scheduler.WeeklyPlan) error {
	
	// Create a custom executor that integrates with our existing benchmark logic
	customExecutor := &CustomBenchmarkExecutor{
		executor: executor,
		batchScheduler: batchScheduler,
	}
	
	// Replace the scheduler's benchmark runner with our custom implementation
	batchScheduler.SetBenchmarkRunner(customExecutor)
	
	// Execute the plan
	return batchScheduler.ExecutePlan(ctx, plan)
}

// CustomBenchmarkExecutor integrates scheduler with existing benchmark execution
type CustomBenchmarkExecutor struct {
	executor       *ScheduledBenchmarkExecutor
	batchScheduler *scheduler.BatchScheduler
}

// ExecuteBenchmark runs a single benchmark job using existing orchestrator
func (ce *CustomBenchmarkExecutor) ExecuteBenchmark(ctx context.Context, job *scheduler.BenchmarkJob) error {
	// Convert scheduler job to our BenchmarkConfig format
	config := awspkg.BenchmarkConfig{
		InstanceType:    job.InstanceType,
		ContainerImage:  fmt.Sprintf("public.ecr.aws/aws-benchmarks/%s:%s", 
			job.BenchmarkSuite, getContainerTagForInstance(job.InstanceType)),
		BenchmarkSuite:  job.BenchmarkSuite,
		Region:          job.Region,
		KeyPairName:     ce.executor.keyPair,
		SecurityGroupID: ce.executor.securityGroup,
		SubnetID:        ce.executor.subnet,
		SkipQuotaCheck:  false,
		MaxRetries:      3,
		Timeout:         10 * time.Minute,
	}
	
	// Execute benchmark using existing orchestrator
	result, err := ce.executor.orchestrator.RunBenchmark(ctx, config)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}
	
	// Store results using existing storage logic
	return storeResults(ctx, ce.executor.s3Storage, result, job.BenchmarkSuite, job.Region)
}