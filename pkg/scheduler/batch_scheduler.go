// Package scheduler provides systematic batch execution of benchmarks over time
// with intelligent job distribution, quota management, and progress tracking.
//
// This package enables comprehensive testing across all AWS instance types
// by distributing workloads over time windows to avoid quota limits and
// optimize cost through spot instances and off-peak execution.
//
// Key Components:
//   - BatchScheduler: Coordinates benchmark execution across time windows
//   - JobQueue: Manages benchmark job prioritization and execution order
//   - TimeWindow: Defines execution schedules and capacity limits
//   - ProgressTracker: Monitors completion and retry logic
//
// Usage:
//   scheduler := scheduler.NewBatchScheduler(config)
//   plan := scheduler.GenerateWeeklyPlan(instanceTypes, benchmarks)
//   err := scheduler.ExecutePlan(ctx, plan)
package scheduler

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// BatchScheduler manages systematic execution of benchmarks across time windows
// with intelligent distribution and resource management.
type BatchScheduler struct {
	config         Config
	jobQueue       *JobQueue
	progressTracker *ProgressTracker
	timeWindows    []TimeWindow
	benchmarkRunner BenchmarkRunner
}

// BenchmarkRunner interface for custom benchmark execution
type BenchmarkRunner interface {
	ExecuteBenchmark(ctx context.Context, job *BenchmarkJob) error
}

// Config defines comprehensive configuration for batch benchmark execution.
type Config struct {
	// MaxConcurrentJobs limits the number of simultaneous benchmark executions
	MaxConcurrentJobs int
	
	// MaxDailyJobs limits the total number of jobs per day to manage quotas
	MaxDailyJobs int
	
	// PreferredRegions lists regions in order of preference for execution
	PreferredRegions []string
	
	// SpotInstancePreference controls use of spot instances for cost optimization
	SpotInstancePreference bool
	
	// TimeZone for scheduling (e.g., "America/New_York", "UTC")
	TimeZone string
	
	// RetryAttempts for failed benchmark executions
	RetryAttempts int
	
	// QuotaLimits defines service quota limits per region
	QuotaLimits map[string]QuotaLimit
	
	// CostOptimization enables cost-aware scheduling
	CostOptimization bool
}

// QuotaLimit defines resource limits for a region to prevent quota exceeded errors.
type QuotaLimit struct {
	// MaxInstancesPerFamily limits concurrent instances per family (e.g., m7i)
	MaxInstancesPerFamily map[string]int
	
	// MaxTotalInstances limits total concurrent instances in region
	MaxTotalInstances int
	
	// MaxGPUInstances limits concurrent GPU instances (for p3, g4, etc.)
	MaxGPUInstances int
	
	// MaxHighMemoryInstances limits high-memory instances (x2, u-*)
	MaxHighMemoryInstances int
}

// TimeWindow defines a scheduled execution period with capacity and preferences.
type TimeWindow struct {
	// StartTime defines when this window begins
	StartTime time.Time
	
	// Duration defines how long this window lasts
	Duration time.Duration
	
	// MaxJobs limits the number of jobs to execute in this window
	MaxJobs int
	
	// Priority affects job selection (higher priority windows get better jobs)
	Priority int
	
	// PreferredInstanceTypes can be used to focus on specific instances
	PreferredInstanceTypes []string
	
	// PreferredBenchmarks can be used to focus on specific benchmarks
	PreferredBenchmarks []string
	
	// RegionPreference for this time window
	RegionPreference string
}

// BenchmarkJob represents a single benchmark execution to be scheduled.
type BenchmarkJob struct {
	// ID uniquely identifies this job
	ID string
	
	// InstanceType to benchmark (e.g., "m7i.large")
	InstanceType string
	
	// BenchmarkSuite to execute (e.g., "stream", "hpl")
	BenchmarkSuite string
	
	// Region for execution
	Region string
	
	// Priority affects execution order (higher priority jobs run first)
	Priority int
	
	// EstimatedDuration for scheduling purposes
	EstimatedDuration time.Duration
	
	// EstimatedCost for cost optimization
	EstimatedCost float64
	
	// RetryCount tracks how many times this job has been attempted
	RetryCount int
	
	// Dependencies lists job IDs that must complete before this job
	Dependencies []string
	
	// PreferSpotInstance for cost optimization
	PreferSpotInstance bool
	
	// Tags for categorization and filtering
	Tags map[string]string
}

// JobQueue manages prioritized execution of benchmark jobs with intelligent scheduling.
type JobQueue struct {
	jobs     []*BenchmarkJob
	index    map[string]*BenchmarkJob
	progress map[string]JobStatus
}

// JobStatus tracks the current state of a benchmark job.
type JobStatus struct {
	Status        string    // "pending", "running", "completed", "failed", "retrying"
	StartTime     time.Time
	EndTime       time.Time
	ExecutionTime time.Duration
	ErrorMessage  string
	RetryCount    int
	ResultPath    string
}

// ProgressTracker monitors overall execution progress and provides reporting.
type ProgressTracker struct {
	totalJobs     int
	completedJobs int
	failedJobs    int
	runningJobs   int
	startTime     time.Time
}

// WeeklyPlan defines a comprehensive benchmark execution plan distributed over a week.
type WeeklyPlan struct {
	// StartDate when the plan begins execution
	StartDate time.Time
	
	// TimeWindows defines when and how benchmarks execute
	TimeWindows []TimeWindow
	
	// Jobs contains all benchmark jobs to execute
	Jobs []*BenchmarkJob
	
	// EstimatedCost for the entire plan
	EstimatedCost float64
	
	// EstimatedDuration for the entire plan
	EstimatedDuration time.Duration
	
	// Metadata for plan tracking
	Metadata map[string]interface{}
}

// NewBatchScheduler creates a new scheduler with the provided configuration.
func NewBatchScheduler(config Config) *BatchScheduler {
	return &BatchScheduler{
		config:          config,
		jobQueue:        NewJobQueue(),
		progressTracker: NewProgressTracker(),
		timeWindows:     []TimeWindow{},
		benchmarkRunner: nil,
	}
}

// SetBenchmarkRunner sets the custom benchmark execution implementation
func (bs *BatchScheduler) SetBenchmarkRunner(runner BenchmarkRunner) {
	bs.benchmarkRunner = runner
}

// NewJobQueue creates a new job queue for managing benchmark execution.
func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs:     []*BenchmarkJob{},
		index:    make(map[string]*BenchmarkJob),
		progress: make(map[string]JobStatus),
	}
}

// NewProgressTracker creates a new progress tracker for monitoring execution.
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		startTime: time.Now(),
	}
}

// GenerateWeeklyPlan creates a comprehensive benchmark plan distributed over a week.
//
// This method intelligently distributes benchmark jobs across time windows to:
//   - Avoid quota limits by spreading load over time
//   - Optimize costs by using spot instances and off-peak pricing
//   - Balance coverage across instance families and benchmark types
//   - Respect region preferences and availability constraints
func (bs *BatchScheduler) GenerateWeeklyPlan(instanceTypes []string, benchmarks []string) (*WeeklyPlan, error) {
	plan := &WeeklyPlan{
		StartDate:   time.Now(),
		TimeWindows: bs.generateTimeWindows(),
		Jobs:        []*BenchmarkJob{},
		Metadata:    make(map[string]interface{}),
	}
	
	// Generate all possible benchmark jobs
	jobs := bs.generateBenchmarkJobs(instanceTypes, benchmarks)
	
	// Prioritize and distribute jobs across time windows
	bs.distributeJobs(plan, jobs)
	
	// Calculate cost and duration estimates
	bs.calculatePlanEstimates(plan)
	
	return plan, nil
}

// generateTimeWindows creates a week's worth of execution windows with intelligent scheduling.
func (bs *BatchScheduler) generateTimeWindows() []TimeWindow {
	windows := []TimeWindow{}
	now := time.Now()
	
	// Create daily windows for a week with microarchitecture focus
	for day := 0; day < 7; day++ {
		startTime := now.AddDate(0, 0, day)
		
		// Morning window - Memory and cache benchmarks
		windows = append(windows, TimeWindow{
			StartTime:   startTime.Add(8 * time.Hour),  // 8 AM
			Duration:    4 * time.Hour,                  // 4 hours
			MaxJobs:     bs.config.MaxDailyJobs / 3,
			Priority:    1,
			PreferredBenchmarks: []string{"stream", "stream-cache", "stream-numa", "micro-cache"}, 
		})
		
		// Afternoon window - Compute and vectorization benchmarks
		windows = append(windows, TimeWindow{
			StartTime:   startTime.Add(14 * time.Hour), // 2 PM
			Duration:    6 * time.Hour,                  // 6 hours
			MaxJobs:     bs.config.MaxDailyJobs / 2,
			Priority:    2,
			PreferredBenchmarks: []string{"hpl", "hpl-vector", "stream-avx512", "stream-neon"},
		})
		
		// Evening window - Microarchitecture-specific and optimized library tests
		windows = append(windows, TimeWindow{
			StartTime:   startTime.Add(20 * time.Hour), // 8 PM
			Duration:    4 * time.Hour,                  // 4 hours
			MaxJobs:     bs.config.MaxDailyJobs / 4,
			Priority:    3,
			PreferredBenchmarks: []string{"hpl-mkl", "hpl-blis", "micro-latency", "micro-ipc"}, 
		})
	}
	
	return windows
}

// generateBenchmarkJobs creates all possible benchmark job combinations with microarchitecture support.
func (bs *BatchScheduler) generateBenchmarkJobs(instanceTypes []string, benchmarks []string) []*BenchmarkJob {
	jobs := []*BenchmarkJob{}
	jobID := 0
	
	for _, instanceType := range instanceTypes {
		// Expand benchmarks to include microarchitecture-specific tests
		expandedBenchmarks := bs.expandBenchmarksForArchitecture(instanceType, benchmarks)
		
		for _, benchmark := range expandedBenchmarks {
			for _, region := range bs.config.PreferredRegions {
				job := &BenchmarkJob{
					ID:             fmt.Sprintf("job-%d", jobID),
					InstanceType:   instanceType,
					BenchmarkSuite: benchmark,
					Region:         region,
					Priority:       bs.calculateJobPriority(instanceType, benchmark),
					EstimatedDuration: bs.estimateJobDuration(instanceType, benchmark),
					EstimatedCost:     bs.estimateJobCost(instanceType, benchmark, region),
					PreferSpotInstance: bs.shouldUseSpotInstance(instanceType),
					Tags: map[string]string{
						"instance_family": extractInstanceFamily(instanceType),
						"benchmark_type":  benchmark,
						"architecture":    bs.getArchitectureType(instanceType),
						"region":         region,
					},
				}
				jobs = append(jobs, job)
				jobID++
			}
		}
	}
	
	return jobs
}

// distributeJobs intelligently assigns jobs to time windows based on priorities and constraints.
func (bs *BatchScheduler) distributeJobs(plan *WeeklyPlan, jobs []*BenchmarkJob) {
	// Sort jobs by priority (higher priority first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Priority > jobs[j].Priority
	})
	
	// Track capacity usage per window
	windowUsage := make(map[int]int)
	
	for _, job := range jobs {
		// Find the best time window for this job
		bestWindow := bs.findBestWindow(plan.TimeWindows, job, windowUsage)
		if bestWindow != -1 {
			plan.Jobs = append(plan.Jobs, job)
			windowUsage[bestWindow]++
		}
	}
}

// findBestWindow identifies the optimal time window for a benchmark job.
func (bs *BatchScheduler) findBestWindow(windows []TimeWindow, job *BenchmarkJob, usage map[int]int) int {
	bestScore := -1
	bestWindow := -1
	
	for i, window := range windows {
		// Skip if window is at capacity
		if usage[i] >= window.MaxJobs {
			continue
		}
		
		score := bs.calculateWindowScore(window, job)
		if score > bestScore {
			bestScore = score
			bestWindow = i
		}
	}
	
	return bestWindow
}

// calculateWindowScore determines how well a job fits a time window.
func (bs *BatchScheduler) calculateWindowScore(window TimeWindow, job *BenchmarkJob) int {
	score := window.Priority * 10
	
	// Prefer windows that match benchmark preferences
	for _, preferred := range window.PreferredBenchmarks {
		if preferred == job.BenchmarkSuite {
			score += 25 // Higher weight for exact matches
		}
		// Partial matches for benchmark families
		if contains(job.BenchmarkSuite, preferred) || contains(preferred, job.BenchmarkSuite) {
			score += 15
		}
	}
	
	// Prefer windows that match instance preferences
	for _, preferred := range window.PreferredInstanceTypes {
		if preferred == job.InstanceType {
			score += 20
		}
	}
	
	// Prefer windows with matching region
	if window.RegionPreference == job.Region {
		score += 12
	}
	
	// Architectural affinity - prefer grouping similar architectures
	archType := bs.getArchitectureType(job.InstanceType)
	if (archType == "intel" && (contains(job.BenchmarkSuite, "avx") || contains(job.BenchmarkSuite, "mkl"))) ||
		(archType == "amd" && (contains(job.BenchmarkSuite, "blis") || contains(job.BenchmarkSuite, "zen"))) ||
		(archType == "graviton" && (contains(job.BenchmarkSuite, "neon") || contains(job.BenchmarkSuite, "sve"))) {
		score += 18 // Strong preference for architecture-matched benchmarks
	}
	
	return score
}

// calculatePlanEstimates computes cost and duration estimates for the entire plan.
func (bs *BatchScheduler) calculatePlanEstimates(plan *WeeklyPlan) {
	totalCost := 0.0
	maxEndTime := plan.StartDate
	
	for _, job := range plan.Jobs {
		totalCost += job.EstimatedCost
		
		// Find when this job will complete
		jobEndTime := plan.StartDate.Add(job.EstimatedDuration)
		if jobEndTime.After(maxEndTime) {
			maxEndTime = jobEndTime
		}
	}
	
	plan.EstimatedCost = totalCost
	plan.EstimatedDuration = maxEndTime.Sub(plan.StartDate)
}

// Helper methods for job estimation and prioritization

func (bs *BatchScheduler) calculateJobPriority(instanceType, benchmark string) int {
	priority := 50 // Base priority
	
	// Higher priority for newer instance families
	family := extractInstanceFamily(instanceType)
	if len(family) > 0 {
		if family[0] == '7' { // 7th generation
			priority += 30
		} else if family[0] == '6' { // 6th generation
			priority += 20
		} else if family[0] == '5' { // 5th generation
			priority += 10
		}
	}
	
	// Priority based on benchmark importance and complexity
	switch {
	case benchmark == "stream":
		priority += 20 // Memory benchmarks are foundational
	case benchmark == "hpl":
		priority += 15 // CPU benchmarks are important
	case contains(benchmark, "numa"):
		priority += 25 // NUMA topology is critical for large instances
	case contains(benchmark, "cache"):
		priority += 18 // Cache hierarchy affects all workloads
	case contains(benchmark, "vector") || contains(benchmark, "avx") || contains(benchmark, "neon"):
		priority += 12 // Vectorization is important for compute workloads
	case contains(benchmark, "micro"):
		priority += 8 // Microbenchmarks provide detailed insights
	case contains(benchmark, "mkl") || contains(benchmark, "blis"):
		priority += 10 // Optimized libraries show peak performance
	}
	
	// Boost priority for large instances that benefit from detailed analysis
	if contains(instanceType, "xlarge") {
		priority += 5
	}
	
	return priority
}

func (bs *BatchScheduler) estimateJobDuration(instanceType, benchmark string) time.Duration {
	baseDuration := 45 * time.Second
	
	// Scale duration based on instance size
	if contains(instanceType, "8xlarge") {
		baseDuration = 120 * time.Second
	} else if contains(instanceType, "4xlarge") {
		baseDuration = 90 * time.Second
	} else if contains(instanceType, "2xlarge") {
		baseDuration = 75 * time.Second
	} else if contains(instanceType, "xlarge") {
		baseDuration = 60 * time.Second
	}
	
	// Adjust duration based on benchmark complexity
	switch {
	case benchmark == "hpl":
		baseDuration = baseDuration * 3 / 2 // HPL is compute-intensive
	case contains(benchmark, "numa"):
		baseDuration = baseDuration * 2 // NUMA tests require multiple memory regions
	case contains(benchmark, "cache"):
		baseDuration = baseDuration * 4 / 3 // Cache hierarchy analysis is thorough
	case contains(benchmark, "micro"):
		baseDuration = baseDuration * 2 / 3 // Microbenchmarks are typically faster
	case contains(benchmark, "vector") || contains(benchmark, "avx") || contains(benchmark, "neon"):
		baseDuration = baseDuration * 5 / 4 // Vectorization tests need multiple passes
	case contains(benchmark, "mkl") || contains(benchmark, "blis"):
		baseDuration = baseDuration * 3 / 2 // Optimized library tests are comprehensive
	}
	
	return baseDuration
}

func (bs *BatchScheduler) estimateJobCost(instanceType, benchmark string, region string) float64 {
	// Base cost estimation (simplified)
	baseCost := 0.10 // $0.10 base
	
	if contains(instanceType, "xlarge") {
		baseCost *= 2
	}
	if contains(instanceType, "2xlarge") {
		baseCost *= 4
	}
	
	// Regional pricing differences
	if region != "us-east-1" {
		baseCost *= 1.1
	}
	
	return baseCost
}

func (bs *BatchScheduler) shouldUseSpotInstance(instanceType string) bool {
	if !bs.config.SpotInstancePreference {
		return false
	}
	
	// Use spot for larger, less critical instances
	return contains(instanceType, "xlarge") || contains(instanceType, "2xlarge")
}

// ExecutePlan executes a weekly benchmark plan with progress tracking.
func (bs *BatchScheduler) ExecutePlan(ctx context.Context, plan *WeeklyPlan) error {
	bs.progressTracker.totalJobs = len(plan.Jobs)
	
	for _, window := range plan.TimeWindows {
		// Wait until window start time
		if time.Now().Before(window.StartTime) {
			select {
			case <-time.After(time.Until(window.StartTime)):
				// Continue when window starts
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		// Execute jobs in this window
		err := bs.executeTimeWindow(ctx, window, plan.Jobs)
		if err != nil {
			return fmt.Errorf("failed to execute time window: %w", err)
		}
	}
	
	return nil
}

// executeTimeWindow executes all jobs assigned to a specific time window.
func (bs *BatchScheduler) executeTimeWindow(ctx context.Context, window TimeWindow, allJobs []*BenchmarkJob) error {
	// Filter jobs for this window
	windowJobs := bs.getJobsForWindow(window, allJobs)
	
	// Execute jobs with concurrency control
	semaphore := make(chan struct{}, bs.config.MaxConcurrentJobs)
	
	for _, job := range windowJobs {
		select {
		case semaphore <- struct{}{}:
			go func(j *BenchmarkJob) {
				defer func() { <-semaphore }()
				bs.executeJob(ctx, j)
			}(job)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return nil
}

// executeJob executes a single benchmark job with retry logic.
func (bs *BatchScheduler) executeJob(ctx context.Context, job *BenchmarkJob) error {
	// This would integrate with the existing benchmark execution logic
	// For now, this is a placeholder that would call the actual benchmark runner
	
	bs.progressTracker.runningJobs++
	defer func() { bs.progressTracker.runningJobs-- }()
	
	// Update job status
	bs.jobQueue.progress[job.ID] = JobStatus{
		Status:    "running",
		StartTime: time.Now(),
	}
	
	// Execute the actual benchmark (placeholder)
	err := bs.runBenchmark(ctx, job)
	
	if err != nil {
		bs.progressTracker.failedJobs++
		bs.jobQueue.progress[job.ID] = JobStatus{
			Status:       "failed",
			EndTime:      time.Now(),
			ErrorMessage: err.Error(),
		}
		return err
	}
	
	bs.progressTracker.completedJobs++
	bs.jobQueue.progress[job.ID] = JobStatus{
		Status:  "completed",
		EndTime: time.Now(),
	}
	
	return nil
}

// runBenchmark executes the actual benchmark using custom runner or placeholder.
func (bs *BatchScheduler) runBenchmark(ctx context.Context, job *BenchmarkJob) error {
	// Use custom benchmark runner if available
	if bs.benchmarkRunner != nil {
		return bs.benchmarkRunner.ExecuteBenchmark(ctx, job)
	}
	
	// Fallback to simulation for testing
	select {
	case <-time.After(job.EstimatedDuration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Utility functions

func extractInstanceFamily(instanceType string) string {
	// Extract family from instance type (e.g., "m7i.large" -> "m7i")
	parts := []rune(instanceType)
	for i, r := range parts {
		if r == '.' {
			return string(parts[:i])
		}
	}
	return instanceType
}

// expandBenchmarksForArchitecture adds microarchitecture-specific benchmarks
func (bs *BatchScheduler) expandBenchmarksForArchitecture(instanceType string, benchmarks []string) []string {
	expanded := make([]string, 0, len(benchmarks)*3) // Estimate expansion
	archType := bs.getArchitectureType(instanceType)
	
	for _, benchmark := range benchmarks {
		// Add base benchmark
		expanded = append(expanded, benchmark)
		
		// Add microarchitecture-specific variants
		switch benchmark {
		case "stream":
			// Memory hierarchy and NUMA topology tests
			expanded = append(expanded, 
				"stream-numa",      // NUMA-aware memory access patterns
				"stream-cache",     // Cache hierarchy analysis
				"stream-prefetch", // Hardware prefetcher evaluation
			)
			
			// Architecture-specific memory tests
			switch archType {
			case "intel":
				expanded = append(expanded, "stream-avx512") // AVX-512 memory bandwidth
			case "amd":
				expanded = append(expanded, "stream-avx2") // AMD-optimized AVX2
			case "graviton":
				expanded = append(expanded, "stream-neon") // ARM Neon SIMD
			}
			
		case "hpl":
			// CPU microarchitecture and vectorization tests
			expanded = append(expanded,
				"hpl-single",  // Single-threaded performance
				"hpl-vector",  // Vectorization efficiency
				"hpl-branch",  // Branch prediction analysis
			)
			
			// Architecture-specific compute tests
			switch archType {
			case "intel":
				expanded = append(expanded, 
					"hpl-mkl",        // Intel MKL optimizations
					"hpl-avx512-fma", // AVX-512 + FMA units
				)
			case "amd":
				expanded = append(expanded, 
					"hpl-blis",   // AMD BLIS optimizations
					"hpl-zen4",   // Zen4-specific features
				)
			case "graviton":
				expanded = append(expanded, 
					"hpl-sve",    // ARM SVE vectorization
					"hpl-neoverse", // Neoverse core optimizations
				)
			}
			
			// Add emerging benchmark types for comprehensive analysis
		case "micro":
			// Microarchitecture-specific microbenchmarks
			expanded = append(expanded,
				"micro-latency",  // Memory latency analysis
				"micro-ipc",      // Instructions per cycle
				"micro-tlb",      // TLB performance
				"micro-cache",    // Cache miss patterns
			)
		}
	}
	
	return expanded
}

// getArchitectureType determines the CPU architecture type for optimization
func (bs *BatchScheduler) getArchitectureType(instanceType string) string {
	family := extractInstanceFamily(instanceType)
	
	// Graviton (ARM64) instances
	if contains(instanceType, "g") || contains(family, "g") {
		return "graviton"
	}
	
	// AMD instances
	if contains(instanceType, "a") || contains(family, "a") {
		return "amd"
	}
	
	// Intel instances (default)
	return "intel"
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (bs *BatchScheduler) getJobsForWindow(window TimeWindow, allJobs []*BenchmarkJob) []*BenchmarkJob {
	// This would filter jobs assigned to this specific window
	// For now, return a subset based on window preferences
	var windowJobs []*BenchmarkJob
	count := 0
	
	for _, job := range allJobs {
		if count >= window.MaxJobs {
			break
		}
		
		// Check if job matches window preferences
		matches := false
		for _, preferred := range window.PreferredBenchmarks {
			if preferred == job.BenchmarkSuite {
				matches = true
				break
			}
		}
		
		if matches || len(window.PreferredBenchmarks) == 0 {
			windowJobs = append(windowJobs, job)
			count++
		}
	}
	
	return windowJobs
}