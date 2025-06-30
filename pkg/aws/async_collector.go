package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AsyncCollector checks for completed benchmarks and collects results
type AsyncCollector struct {
	s3Client *s3.Client
	region   string
}

// NewAsyncCollector creates a new benchmark result collector
func NewAsyncCollector(region string) (*AsyncCollector, error) {
	cfg, err := LoadAWSConfig(region)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AsyncCollector{
		s3Client: s3.NewFromConfig(cfg),
		region:   region,
	}, nil
}

// CollectionResult represents the result of a collection cycle
type CollectionResult struct {
	Completed   []*AsyncBenchmarkResult `json:"completed"`
	Failed      []*AsyncBenchmarkResult `json:"failed"`
	InProgress  []*AsyncBenchmarkJob    `json:"in_progress"`
	TimedOut    []*AsyncBenchmarkJob    `json:"timed_out"`
	Summary     CollectionSummary       `json:"summary"`
}

// CollectionSummary provides aggregate statistics
type CollectionSummary struct {
	TotalJobs       int     `json:"total_jobs"`
	CompletedJobs   int     `json:"completed_jobs"`
	FailedJobs      int     `json:"failed_jobs"`
	InProgressJobs  int     `json:"in_progress_jobs"`
	TimedOutJobs    int     `json:"timed_out_jobs"`
	TotalCost       float64 `json:"total_cost"`
	SuccessRate     float64 `json:"success_rate"`
}

// CheckAllBenchmarks scans S3 for all benchmark jobs and returns status
func (c *AsyncCollector) CheckAllBenchmarks(ctx context.Context, s3Bucket string) (*CollectionResult, error) {
	fmt.Printf("🔍 SCANNING FOR BENCHMARK RESULTS\n")
	fmt.Printf("==================================\n")
	fmt.Printf("   S3 Bucket: %s\n", s3Bucket)
	fmt.Printf("   Region: %s\n", c.region)
	fmt.Printf("==================================\n\n")

	result := &CollectionResult{
		Completed:  make([]*AsyncBenchmarkResult, 0),
		Failed:     make([]*AsyncBenchmarkResult, 0),
		InProgress: make([]*AsyncBenchmarkJob, 0),
		TimedOut:   make([]*AsyncBenchmarkJob, 0),
	}

	// List all benchmark jobs in S3
	jobs, err := c.listAllJobs(ctx, s3Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	fmt.Printf("📊 Found %d benchmark jobs\n\n", len(jobs))

	// Check status of each job
	for i, job := range jobs {
		fmt.Printf("🔬 Checking job %d/%d: %s\n", i+1, len(jobs), job.BenchmarkID)
		
		status, err := c.checkJobStatus(ctx, job)
		if err != nil {
			fmt.Printf("   ⚠️  Error checking status: %v\n", err)
			continue
		}

		switch status {
		case JobStatusCompleted:
			benchmarkResult, err := c.collectCompletedJob(ctx, job)
			if err != nil {
				fmt.Printf("   ❌ Failed to collect results: %v\n", err)
				continue
			}
			result.Completed = append(result.Completed, benchmarkResult)
			fmt.Printf("   ✅ Completed: %s on %s\n", 
				job.BenchmarkConfig.BenchmarkSuite, job.BenchmarkConfig.InstanceType)

		case JobStatusFailed:
			benchmarkResult, err := c.collectFailedJob(ctx, job)
			if err != nil {
				fmt.Printf("   ⚠️  Failed to collect failure details: %v\n", err)
				continue
			}
			result.Failed = append(result.Failed, benchmarkResult)
			fmt.Printf("   ❌ Failed: %s on %s\n", 
				job.BenchmarkConfig.BenchmarkSuite, job.BenchmarkConfig.InstanceType)

		case JobStatusTimedOut:
			result.TimedOut = append(result.TimedOut, job)
			fmt.Printf("   ⏰ Timed out: %s on %s\n", 
				job.BenchmarkConfig.BenchmarkSuite, job.BenchmarkConfig.InstanceType)

		case JobStatusEmergencyStop:
			result.TimedOut = append(result.TimedOut, job) // Treat as timeout for grouping
			fmt.Printf("   🚨 Emergency stop: %s on %s (failsafe triggered)\n", 
				job.BenchmarkConfig.BenchmarkSuite, job.BenchmarkConfig.InstanceType)

		case JobStatusRunning, JobStatusLaunched:
			result.InProgress = append(result.InProgress, job)
			fmt.Printf("   🔄 In progress: %s on %s\n", 
				job.BenchmarkConfig.BenchmarkSuite, job.BenchmarkConfig.InstanceType)

		default:
			fmt.Printf("   ❓ Unknown status: %s\n", status)
		}
	}

	// Calculate summary statistics
	result.Summary = c.calculateSummary(result)

	fmt.Printf("\n📈 COLLECTION SUMMARY\n")
	fmt.Printf("====================\n")
	fmt.Printf("   Total Jobs: %d\n", result.Summary.TotalJobs)
	fmt.Printf("   Completed: %d (%.1f%%)\n", result.Summary.CompletedJobs, 
		float64(result.Summary.CompletedJobs)/float64(result.Summary.TotalJobs)*100)
	fmt.Printf("   Failed: %d (%.1f%%)\n", result.Summary.FailedJobs,
		float64(result.Summary.FailedJobs)/float64(result.Summary.TotalJobs)*100)
	fmt.Printf("   In Progress: %d\n", result.Summary.InProgressJobs)
	fmt.Printf("   Timed Out: %d\n", result.Summary.TimedOutJobs)
	fmt.Printf("   Success Rate: %.1f%%\n", result.Summary.SuccessRate)
	fmt.Printf("   Total Cost: $%.4f\n", result.Summary.TotalCost)

	return result, nil
}

// WaitForCompletion waits for specific jobs to complete
func (c *AsyncCollector) WaitForCompletion(ctx context.Context, jobs []*AsyncBenchmarkJob, 
	checkInterval time.Duration) (*CollectionResult, error) {
	
	fmt.Printf("⏳ WAITING FOR BENCHMARK COMPLETION\n")
	fmt.Printf("===================================\n")
	fmt.Printf("   Jobs: %d\n", len(jobs))
	fmt.Printf("   Check interval: %v\n", checkInterval)
	fmt.Printf("===================================\n\n")

	completedJobs := make(map[string]bool)
	
	for {
		allComplete := true
		
		for _, job := range jobs {
			if completedJobs[job.BenchmarkID] {
				continue // Already completed
			}
			
			status, err := c.checkJobStatus(ctx, job)
			if err != nil {
				fmt.Printf("⚠️  Error checking %s: %v\n", job.BenchmarkID, err)
				continue
			}
			
			switch status {
			case JobStatusCompleted, JobStatusFailed, JobStatusTimedOut:
				completedJobs[job.BenchmarkID] = true
				fmt.Printf("🎯 Job completed: %s (%s)\n", job.BenchmarkID, status)
			case JobStatusRunning, JobStatusLaunched:
				allComplete = false
				
				// Show progress if available
				progress, err := c.getJobProgress(ctx, job)
				if err == nil && progress != nil {
					fmt.Printf("📊 %s: %s (%.1f%% complete)\n", 
						job.BenchmarkID, progress.Message, progress.PercentComplete)
				} else {
					fmt.Printf("🔄 %s: %s\n", job.BenchmarkID, status)
				}
			}
		}
		
		if allComplete {
			fmt.Printf("\n🎉 All jobs completed!\n")
			break
		}
		
		fmt.Printf("   ⏱️  Next check in %v...\n\n", checkInterval)
		time.Sleep(checkInterval)
	}
	
	// Collect all results
	return c.CheckSpecificJobs(ctx, jobs)
}

// CheckSpecificJobs checks the status of specific jobs
func (c *AsyncCollector) CheckSpecificJobs(ctx context.Context, jobs []*AsyncBenchmarkJob) (*CollectionResult, error) {
	result := &CollectionResult{
		Completed:  make([]*AsyncBenchmarkResult, 0),
		Failed:     make([]*AsyncBenchmarkResult, 0),
		InProgress: make([]*AsyncBenchmarkJob, 0),
		TimedOut:   make([]*AsyncBenchmarkJob, 0),
	}

	for _, job := range jobs {
		status, err := c.checkJobStatus(ctx, job)
		if err != nil {
			continue
		}

		switch status {
		case JobStatusCompleted:
			if benchmarkResult, err := c.collectCompletedJob(ctx, job); err == nil {
				result.Completed = append(result.Completed, benchmarkResult)
			}
		case JobStatusFailed:
			if benchmarkResult, err := c.collectFailedJob(ctx, job); err == nil {
				result.Failed = append(result.Failed, benchmarkResult)
			}
		case JobStatusTimedOut:
			result.TimedOut = append(result.TimedOut, job)
		default:
			result.InProgress = append(result.InProgress, job)
		}
	}

	result.Summary = c.calculateSummary(result)
	return result, nil
}

// listAllJobs finds all benchmark jobs in S3
func (c *AsyncCollector) listAllJobs(ctx context.Context, s3Bucket string) ([]*AsyncBenchmarkJob, error) {
	var jobs []*AsyncBenchmarkJob

	// List all objects under benchmarks/ prefix
	paginator := s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Bucket),
		Prefix: aws.String("benchmarks/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Look for job-metadata.json files
			if strings.HasSuffix(*obj.Key, "job-metadata.json") {
				job, err := c.loadJobMetadata(ctx, s3Bucket, *obj.Key)
				if err != nil {
					fmt.Printf("⚠️  Failed to load job metadata from %s: %v\n", *obj.Key, err)
					continue
				}
				jobs = append(jobs, job)
			}
		}
	}

	return jobs, nil
}

// loadJobMetadata loads job metadata from S3
func (c *AsyncCollector) loadJobMetadata(ctx context.Context, bucket, key string) (*AsyncBenchmarkJob, error) {
	resp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var job AsyncBenchmarkJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// checkJobStatus determines the current status of a job
func (c *AsyncCollector) checkJobStatus(ctx context.Context, job *AsyncBenchmarkJob) (JobStatus, error) {
	sentinels := NewS3SentinelFiles(job.S3Prefix)

	// Check sentinels in order of preference
	if c.objectExists(ctx, job.S3Bucket, sentinels.StatusCompleted) {
		return JobStatusCompleted, nil
	}
	if c.objectExists(ctx, job.S3Bucket, sentinels.StatusFailed) {
		return JobStatusFailed, nil
	}
	if c.objectExists(ctx, job.S3Bucket, sentinels.StatusEmergency) {
		return JobStatusEmergencyStop, nil
	}
	if c.objectExists(ctx, job.S3Bucket, sentinels.StatusTimedOut) {
		return JobStatusTimedOut, nil
	}
	if c.objectExists(ctx, job.S3Bucket, sentinels.StatusRunning) {
		return JobStatusRunning, nil
	}
	if c.objectExists(ctx, job.S3Bucket, sentinels.StatusLaunched) {
		return JobStatusLaunched, nil
	}

	return JobStatusLaunched, nil // Default status
}

// objectExists checks if an S3 object exists
func (c *AsyncCollector) objectExists(ctx context.Context, bucket, key string) bool {
	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err == nil
}

// getJobProgress retrieves job progress information
func (c *AsyncCollector) getJobProgress(ctx context.Context, job *AsyncBenchmarkJob) (*BenchmarkProgress, error) {
	sentinels := NewS3SentinelFiles(job.S3Prefix)
	
	resp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.S3Bucket),
		Key:    aws.String(sentinels.StatusProgress),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var progress BenchmarkProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, err
	}

	return &progress, nil
}

// collectCompletedJob collects results from a completed job
func (c *AsyncCollector) collectCompletedJob(ctx context.Context, job *AsyncBenchmarkJob) (*AsyncBenchmarkResult, error) {
	sentinels := NewS3SentinelFiles(job.S3Prefix)

	// Load benchmark results
	resultsResp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.S3Bucket),
		Key:    aws.String(sentinels.BenchmarkResults),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load benchmark results: %w", err)
	}
	defer resultsResp.Body.Close()

	resultsData, err := io.ReadAll(resultsResp.Body)
	if err != nil {
		return nil, err
	}

	var benchmarkData map[string]interface{}
	if err := json.Unmarshal(resultsData, &benchmarkData); err != nil {
		return nil, err
	}

	// Load system info
	systemInfo := make(map[string]interface{})
	if systemResp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.S3Bucket),
		Key:    aws.String(sentinels.SystemInfo),
	}); err == nil {
		defer systemResp.Body.Close()
		if systemData, err := io.ReadAll(systemResp.Body); err == nil {
			json.Unmarshal(systemData, &systemInfo)
		}
	}

	// Calculate execution time
	var executionTime time.Duration
	if job.StartedAt != nil && job.CompletedAt != nil {
		executionTime = job.CompletedAt.Sub(*job.StartedAt)
	}

	return &AsyncBenchmarkResult{
		Job:           job,
		BenchmarkData: benchmarkData,
		SystemInfo:    systemInfo,
		ExecutionTime: executionTime,
		Success:       true,
	}, nil
}

// collectFailedJob collects details from a failed job
func (c *AsyncCollector) collectFailedJob(ctx context.Context, job *AsyncBenchmarkJob) (*AsyncBenchmarkResult, error) {
	// Try to get whatever results we can
	result := &AsyncBenchmarkResult{
		Job:           job,
		BenchmarkData: make(map[string]interface{}),
		SystemInfo:    make(map[string]interface{}),
		Success:       false,
		Error:         "Benchmark execution failed",
	}

	// Try to load logs for error details
	sentinels := NewS3SentinelFiles(job.S3Prefix)
	if logResp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(job.S3Bucket),
		Key:    aws.String(sentinels.BenchmarkLogs),
	}); err == nil {
		defer logResp.Body.Close()
		if logData, err := io.ReadAll(logResp.Body); err == nil {
			result.BenchmarkData["error_logs"] = string(logData)
		}
	}

	return result, nil
}

// calculateSummary computes aggregate statistics
func (c *AsyncCollector) calculateSummary(result *CollectionResult) CollectionSummary {
	total := len(result.Completed) + len(result.Failed) + len(result.InProgress) + len(result.TimedOut)
	completed := len(result.Completed)
	failed := len(result.Failed)
	
	var successRate float64
	if total > 0 {
		successRate = float64(completed) / float64(total) * 100
	}

	var totalCost float64
	for _, job := range result.Completed {
		totalCost += job.Job.EstimatedCost
	}
	for _, job := range result.Failed {
		totalCost += job.Job.EstimatedCost
	}

	return CollectionSummary{
		TotalJobs:      total,
		CompletedJobs:  completed,
		FailedJobs:     failed,
		InProgressJobs: len(result.InProgress),
		TimedOutJobs:   len(result.TimedOut),
		TotalCost:      totalCost,
		SuccessRate:    successRate,
	}
}