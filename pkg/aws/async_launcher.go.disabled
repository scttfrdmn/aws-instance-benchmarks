package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// AsyncLauncher handles fire-and-forget benchmark execution
type AsyncLauncher struct {
	orchestrator *Orchestrator
	s3Client     *s3.Client
}

// NewAsyncLauncher creates a new async benchmark launcher
func NewAsyncLauncher(region string) (*AsyncLauncher, error) {
	orchestrator, err := NewOrchestrator(region)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}

	s3Client := s3.NewFromConfig(orchestrator.cfg)

	return &AsyncLauncher{
		orchestrator: orchestrator,
		s3Client:     s3Client,
	}, nil
}

// LaunchBenchmarks launches multiple benchmarks asynchronously
func (l *AsyncLauncher) LaunchBenchmarks(ctx context.Context, request *LaunchRequest) (*LaunchResponse, error) {
	response := &LaunchResponse{
		Jobs: make([]*AsyncBenchmarkJob, 0, len(request.Configs)),
	}

	fmt.Printf("🚀 LAUNCHING ASYNC BENCHMARKS\n")
	fmt.Printf("===============================\n")
	fmt.Printf("   Configs: %d\n", len(request.Configs))
	fmt.Printf("   S3 Bucket: %s\n", request.S3Bucket)
	fmt.Printf("   Max Runtime: %v\n", request.MaxRuntime)
	fmt.Printf("===============================\n\n")

	for i, config := range request.Configs {
		job, err := l.LaunchSingleBenchmark(ctx, config, request.S3Bucket, 
			fmt.Sprintf("%s-%d", request.JobNamePrefix, i+1), request.MaxRuntime)
		
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, fmt.Sprintf("Config %d: %v", i+1, err))
			fmt.Printf("❌ Failed to launch config %d: %v\n", i+1, err)
			continue
		}

		response.Jobs = append(response.Jobs, job)
		response.LaunchedCount++
		
		fmt.Printf("✅ Launched job %d: %s (%s on %s)\n", 
			i+1, job.BenchmarkID, job.BenchmarkConfig.BenchmarkSuite, job.BenchmarkConfig.InstanceType)
	}

	fmt.Printf("\n🎯 LAUNCH SUMMARY\n")
	fmt.Printf("================\n")
	fmt.Printf("   Successful: %d/%d\n", response.LaunchedCount, len(request.Configs))
	fmt.Printf("   Failed: %d/%d\n", response.FailedCount, len(request.Configs))
	fmt.Printf("   Jobs running independently...\n\n")

	return response, nil
}

// LaunchSingleBenchmark launches a single benchmark asynchronously
func (l *AsyncLauncher) LaunchSingleBenchmark(ctx context.Context, config BenchmarkConfig, 
	s3Bucket, jobName string, maxRuntime time.Duration) (*AsyncBenchmarkJob, error) {

	// Generate unique benchmark ID
	benchmarkID := fmt.Sprintf("bench-%s-%s", 
		strings.ReplaceAll(time.Now().Format("20060102-150405"), "-", ""),
		uuid.New().String()[:8])

	// Create S3 prefix for this benchmark
	s3Prefix := fmt.Sprintf("benchmarks/%s/%s/%s/", 
		benchmarkID, config.InstanceType, config.BenchmarkSuite)

	// Create job metadata
	job := &AsyncBenchmarkJob{
		BenchmarkID:     benchmarkID,
		JobName:         jobName,
		BenchmarkConfig: config,
		S3Bucket:        s3Bucket,
		S3Prefix:        s3Prefix,
		Status:          JobStatusLaunched,
		LaunchedAt:      time.Now(),
		Region:          config.Region,
	}

	// Upload job metadata to S3 first
	if err := l.uploadJobMetadata(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to upload job metadata: %w", err)
	}

	// Launch EC2 instance with self-contained benchmark execution
	instanceID, err := l.launchBenchmarkInstance(ctx, job, maxRuntime)
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}

	job.InstanceID = instanceID
	job.EstimatedCost = l.estimateJobCost(config.InstanceType, maxRuntime)

	// Update job metadata with instance ID
	if err := l.uploadJobMetadata(ctx, job); err != nil {
		fmt.Printf("⚠️  Warning: failed to update job metadata: %v\n", err)
	}

	// Upload launched sentinel
	if err := l.uploadSentinel(ctx, job, JobStatusLaunched); err != nil {
		fmt.Printf("⚠️  Warning: failed to upload launched sentinel: %v\n", err)
	}

	return job, nil
}

// launchBenchmarkInstance launches an EC2 instance with self-contained benchmark execution
func (l *AsyncLauncher) launchBenchmarkInstance(ctx context.Context, job *AsyncBenchmarkJob, 
	maxRuntime time.Duration) (string, error) {

	config := job.BenchmarkConfig

	// Generate comprehensive user data script
	userData := l.generateAsyncUserDataScript(job, maxRuntime)
	userDataEncoded := base64.StdEncoding.EncodeToString([]byte(userData))

	// Get AMI ID for the instance type
	amiID, err := l.orchestrator.getOptimalAMI(ctx, config.InstanceType)
	if err != nil {
		return "", fmt.Errorf("failed to get AMI: %w", err)
	}

	// Configure instance
	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: ec2.InstanceType(config.InstanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		UserData:     aws.String(userDataEncoded),
		
		// Networking
		SecurityGroupIds: []string{config.SecurityGroupID},
		SubnetId:         aws.String(config.SubnetID),
		
		// Instance profile for S3 access
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String("EC2-S3-BenchmarkAccess"), // Needs to be created
		},
		
		// Tagging
		TagSpecifications: []ec2.TagSpecification{
			{
				ResourceType: ec2.ResourceTypeInstance,
				Tags: []ec2.Tag{
					{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("AsyncBenchmark-%s", job.BenchmarkID))},
					{Key: aws.String("BenchmarkID"), Value: aws.String(job.BenchmarkID)},
					{Key: aws.String("BenchmarkSuite"), Value: aws.String(config.BenchmarkSuite)},
					{Key: aws.String("InstanceType"), Value: aws.String(config.InstanceType)},
					{Key: aws.String("LaunchedBy"), Value: aws.String("AsyncBenchmarkLauncher")},
					{Key: aws.String("S3Bucket"), Value: aws.String(job.S3Bucket)},
					{Key: aws.String("AutoTerminate"), Value: aws.String("true")},
				},
			},
		},
	}

	// Add key pair if specified
	if config.KeyPairName != "" {
		runInput.KeyName = aws.String(config.KeyPairName)
	}

	// Launch instance
	result, err := l.orchestrator.ec2Client.RunInstances(ctx, runInput)
	if err != nil {
		return "", fmt.Errorf("failed to launch instance: %w", err)
	}

	if len(result.Instances) == 0 {
		return "", fmt.Errorf("no instances returned from launch request")
	}

	instanceID := *result.Instances[0].InstanceId
	
	fmt.Printf("   🚀 Instance launched: %s\n", instanceID)
	fmt.Printf("   📊 Benchmark: %s on %s\n", config.BenchmarkSuite, config.InstanceType)
	fmt.Printf("   💰 Est. cost: $%.4f (max runtime: %v)\n", 
		l.estimateJobCost(config.InstanceType, maxRuntime), maxRuntime)
	fmt.Printf("   📍 S3 tracking: s3://%s/%s\n", job.S3Bucket, job.S3Prefix)

	return instanceID, nil
}

// generateAsyncUserDataScript creates a comprehensive self-contained benchmark script
func (l *AsyncLauncher) generateAsyncUserDataScript(job *AsyncBenchmarkJob, maxRuntime time.Duration) string {
	config := job.BenchmarkConfig
	sentinels := NewS3SentinelFiles(job.S3Prefix)

	return fmt.Sprintf(`#!/bin/bash
set -e

# AWS Instance Async Benchmark Execution
# Benchmark ID: %s
# Instance Type: %s
# Benchmark Suite: %s
# S3 Bucket: %s
# Max Runtime: %v

echo "🚀 Starting async benchmark execution at $(date)"

# Install dependencies
yum update -y
yum install -y gcc gcc-c++ make wget unzip git htop iostat

# Install AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

# Set up logging
LOGFILE="/tmp/benchmark.log"
exec 1> >(tee -a "$LOGFILE")
exec 2> >(tee -a "$LOGFILE" >&2)

# Function to upload to S3 with retry
upload_to_s3() {
    local file="$1"
    local s3_path="$2"
    local retries=3
    
    for i in $(seq 1 $retries); do
        if aws s3 cp "$file" "$s3_path"; then
            echo "✅ Uploaded $file to $s3_path"
            return 0
        else
            echo "⚠️  Upload attempt $i failed, retrying..."
            sleep 5
        fi
    done
    echo "❌ Failed to upload $file after $retries attempts"
    return 1
}

# Function to upload sentinel
upload_sentinel() {
    local status="$1"
    echo "$(date): $status" > "/tmp/status-$status.sentinel"
    upload_to_s3 "/tmp/status-$status.sentinel" "s3://%s/%sstatus-$status.sentinel"
}

# Function to upload progress
upload_progress() {
    local iteration="$1"
    local total="$2"
    local message="$3"
    
    cat > /tmp/status-progress.json <<EOF
{
    "current_iteration": $iteration,
    "total_iterations": $total,
    "last_update": "$(date -Iseconds)",
    "message": "$message",
    "percent_complete": $(( iteration * 100 / total ))
}
EOF
    upload_to_s3 "/tmp/status-progress.json" "s3://%s/%sstatus-progress.json"
}

# Function to collect system info
collect_system_info() {
    cat > /tmp/system-info.json <<EOF
{
    "instance_id": "$(curl -s http://169.254.169.254/latest/meta-data/instance-id)",
    "instance_type": "$(curl -s http://169.254.169.254/latest/meta-data/instance-type)",
    "availability_zone": "$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone)",
    "cpu_info": "$(lscpu | grep 'Model name' | cut -d: -f2 | xargs)",
    "memory_gb": $(free -g | awk '/^Mem:/{print $2}'),
    "architecture": "$(uname -m)",
    "kernel": "$(uname -r)",
    "timestamp": "$(date -Iseconds)"
}
EOF
    upload_to_s3 "/tmp/system-info.json" "s3://%s/%ssystem-info.json"
}

# Function to handle cleanup and termination
cleanup_and_terminate() {
    echo "🧹 Performing cleanup..."
    
    # Upload final logs
    upload_to_s3 "$LOGFILE" "s3://%s/%sbenchmark.log"
    
    # Self-terminate instance
    INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
    echo "🔚 Self-terminating instance $INSTANCE_ID"
    aws ec2 terminate-instances --region %s --instance-ids "$INSTANCE_ID"
}

# Set up signal handlers for cleanup
trap cleanup_and_terminate EXIT
trap cleanup_and_terminate SIGTERM
trap cleanup_and_terminate SIGINT

# Set maximum runtime enforcement with failsafe
timeout_seconds=%d
failsafe_timeout_seconds=$((timeout_seconds + 3600))  # Add 1 hour failsafe buffer

# Primary timeout - graceful termination
if [ $timeout_seconds -gt 0 ]; then
    (
        sleep $timeout_seconds
        echo "⏰ Maximum runtime exceeded, attempting graceful termination"
        upload_sentinel "TIMED_OUT"
        
        # Try to terminate gracefully first
        kill -TERM $$
        
        # Wait 10 minutes for graceful termination
        sleep 600
        
        # Force kill if still running
        echo "🚨 Graceful termination failed, forcing kill"
        kill -KILL $$
    ) &
    TIMEOUT_PID=$!
fi

# Failsafe timeout - absolute maximum runtime (prevents runaway instances)
if [ $failsafe_timeout_seconds -gt 0 ]; then
    (
        sleep $failsafe_timeout_seconds
        echo "🚨 FAILSAFE TIMEOUT: Instance has exceeded absolute maximum runtime"
        echo "🚨 This indicates a serious problem - forcing immediate termination"
        
        # Upload emergency sentinel
        echo "$(date): EMERGENCY_TIMEOUT - Instance exceeded failsafe limit" > "/tmp/status-emergency.sentinel"
        upload_to_s3 "/tmp/status-emergency.sentinel" "s3://%s/%sstatus-emergency.sentinel"
        
        # Force immediate instance termination
        INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
        aws ec2 terminate-instances --region %s --instance-ids "$INSTANCE_ID" || true
        
        # If EC2 termination fails, try shutdown
        sudo shutdown -h now
        
        # Ultimate fallback - kernel panic to force termination
        echo 1 > /proc/sys/kernel/sysrq
        echo c > /proc/sysrq-trigger
    ) &
    FAILSAFE_PID=$!
fi

echo "📋 Collecting system information..."
collect_system_info

echo "🏁 Starting benchmark execution..."
upload_sentinel "RUNNING"

# Generate and execute benchmark
%s

echo "✅ Benchmark execution completed at $(date)"
upload_sentinel "COMPLETED"

# Kill timeout processes if still running
if [ ! -z "$TIMEOUT_PID" ]; then
    kill $TIMEOUT_PID 2>/dev/null || true
fi
if [ ! -z "$FAILSAFE_PID" ]; then
    kill $FAILSAFE_PID 2>/dev/null || true
fi

echo "🎉 Async benchmark completed successfully!"
`,
		job.BenchmarkID,
		config.InstanceType, 
		config.BenchmarkSuite,
		job.S3Bucket,
		maxRuntime,
		job.S3Bucket, job.S3Prefix,
		job.S3Bucket, job.S3Prefix,
		job.S3Bucket, job.S3Prefix,
		job.S3Bucket, job.S3Prefix,
		config.Region,
		int(maxRuntime.Seconds()),
		job.S3Bucket, job.S3Prefix,
		config.Region,
		l.generateBenchmarkExecutionScript(config))
}

// generateBenchmarkExecutionScript creates the actual benchmark execution code
func (l *AsyncLauncher) generateBenchmarkExecutionScript(config BenchmarkConfig) string {
	// Use the existing benchmark command generation from orchestrator
	baseCommand := l.orchestrator.generateBenchmarkCommand(config)
	
	// Wrap with result collection and S3 upload
	return fmt.Sprintf(`
# Execute benchmark with result collection
echo "🔬 Executing %s benchmark..."

# Run the benchmark and capture output
BENCHMARK_OUTPUT_FILE="/tmp/benchmark_output.txt"
RESULTS_FILE="/tmp/results.json"

# Execute the actual benchmark
%s 2>&1 | tee "$BENCHMARK_OUTPUT_FILE"
BENCHMARK_EXIT_CODE=${PIPESTATUS[0]}

echo "📊 Processing benchmark results..."

# Parse results based on benchmark type
case "%s" in
    "stream")
        # Parse STREAM results
        COPY_BANDWIDTH=$(grep "Copy:" "$BENCHMARK_OUTPUT_FILE" | awk '{print $2}' | tail -1)
        SCALE_BANDWIDTH=$(grep "Scale:" "$BENCHMARK_OUTPUT_FILE" | awk '{print $2}' | tail -1)
        ADD_BANDWIDTH=$(grep "Add:" "$BENCHMARK_OUTPUT_FILE" | awk '{print $2}' | tail -1)
        TRIAD_BANDWIDTH=$(grep "Triad:" "$BENCHMARK_OUTPUT_FILE" | awk '{print $2}' | tail -1)
        
        cat > "$RESULTS_FILE" <<EOF
{
    "benchmark_suite": "stream",
    "results": {
        "copy_bandwidth_mbps": ${COPY_BANDWIDTH:-0},
        "scale_bandwidth_mbps": ${SCALE_BANDWIDTH:-0},
        "add_bandwidth_mbps": ${ADD_BANDWIDTH:-0},
        "triad_bandwidth_mbps": ${TRIAD_BANDWIDTH:-0}
    },
    "success": $([ $BENCHMARK_EXIT_CODE -eq 0 ] && echo "true" || echo "false"),
    "exit_code": $BENCHMARK_EXIT_CODE,
    "timestamp": "$(date -Iseconds)"
}
EOF
        ;;
    "hpl")
        # Parse HPL results
        HPL_GFLOPS=$(grep "WR" "$BENCHMARK_OUTPUT_FILE" | awk '{print $7}' | tail -1)
        
        cat > "$RESULTS_FILE" <<EOF
{
    "benchmark_suite": "hpl",
    "results": {
        "peak_gflops": ${HPL_GFLOPS:-0}
    },
    "success": $([ $BENCHMARK_EXIT_CODE -eq 0 ] && echo "true" || echo "false"),
    "exit_code": $BENCHMARK_EXIT_CODE,
    "timestamp": "$(date -Iseconds)"
}
EOF
        ;;
    *)
        # Generic result format
        cat > "$RESULTS_FILE" <<EOF
{
    "benchmark_suite": "%s",
    "results": {
        "raw_output": "$(cat "$BENCHMARK_OUTPUT_FILE" | base64 -w 0)"
    },
    "success": $([ $BENCHMARK_EXIT_CODE -eq 0 ] && echo "true" || echo "false"),
    "exit_code": $BENCHMARK_EXIT_CODE,
    "timestamp": "$(date -Iseconds)"
}
EOF
        ;;
esac

# Upload results to S3
upload_to_s3 "$RESULTS_FILE" "s3://%s/%sresults.json"

if [ $BENCHMARK_EXIT_CODE -eq 0 ]; then
    echo "✅ Benchmark completed successfully"
else
    echo "❌ Benchmark failed with exit code $BENCHMARK_EXIT_CODE"
    upload_sentinel "FAILED"
    exit $BENCHMARK_EXIT_CODE
fi
`, config.BenchmarkSuite, baseCommand, config.BenchmarkSuite, config.BenchmarkSuite, config.BenchmarkSuite, config.BenchmarkSuite)
}

// uploadJobMetadata uploads job metadata to S3
func (l *AsyncLauncher) uploadJobMetadata(ctx context.Context, job *AsyncBenchmarkJob) error {
	metadata, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal job metadata: %w", err)
	}

	sentinels := NewS3SentinelFiles(job.S3Prefix)
	
	_, err = l.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(job.S3Bucket),
		Key:    aws.String(sentinels.JobMetadata),
		Body:   strings.NewReader(string(metadata)),
		ContentType: aws.String("application/json"),
	})

	return err
}

// uploadSentinel uploads a status sentinel file to S3
func (l *AsyncLauncher) uploadSentinel(ctx context.Context, job *AsyncBenchmarkJob, status JobStatus) error {
	sentinelContent := fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), status)
	sentinelKey := fmt.Sprintf("%sstatus-%s.sentinel", job.S3Prefix, strings.ToLower(string(status)))
	
	_, err := l.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(job.S3Bucket),
		Key:    aws.String(sentinelKey),
		Body:   strings.NewReader(sentinelContent),
		ContentType: aws.String("text/plain"),
	})

	return err
}

// estimateJobCost estimates the cost of running a benchmark job
func (l *AsyncLauncher) estimateJobCost(instanceType string, maxRuntime time.Duration) float64 {
	// Approximate hourly costs (update with current AWS pricing)
	costs := map[string]float64{
		"c7g.large": 0.0725,
		"c7i.large": 0.0864,
		"c7a.large": 0.0864,
		"m7i.large": 0.1008,
		"r7i.large": 0.2016,
	}
	
	if cost, ok := costs[instanceType]; ok {
		return cost * maxRuntime.Hours()
	}
	
	return 0.10 * maxRuntime.Hours() // Default estimate
}