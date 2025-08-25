package aws

import (
	"time"
)

// AsyncBenchmarkJob represents a fire-and-forget benchmark execution
type AsyncBenchmarkJob struct {
	// Job identification
	BenchmarkID  string `json:"benchmark_id"`
	InstanceID   string `json:"instance_id"`
	JobName      string `json:"job_name"`
	
	// Configuration
	BenchmarkConfig BenchmarkConfig `json:"benchmark_config"`
	
	// S3 tracking
	S3Bucket     string `json:"s3_bucket"`
	S3Prefix     string `json:"s3_prefix"`
	
	// Status tracking
	Status       JobStatus `json:"status"`
	LaunchedAt   time.Time `json:"launched_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	
	// Cost tracking
	EstimatedCost float64 `json:"estimated_cost"`
	Region       string  `json:"region"`
}

// JobStatus represents the current state of an async benchmark job
type JobStatus string

const (
	JobStatusLaunched       JobStatus = "LAUNCHED"         // Instance launched, waiting for startup
	JobStatusRunning        JobStatus = "RUNNING"          // Benchmark executing
	JobStatusCompleted      JobStatus = "COMPLETED"        // Benchmark finished successfully
	JobStatusFailed         JobStatus = "FAILED"           // Benchmark failed
	JobStatusTimedOut       JobStatus = "TIMED_OUT"        // Benchmark exceeded maximum runtime
	JobStatusEmergencyStop  JobStatus = "EMERGENCY_STOP"   // Failsafe timeout triggered
	JobStatusTerminated     JobStatus = "TERMINATED"       // Instance terminated
)

// S3SentinelFiles defines the S3 file structure for job tracking
type S3SentinelFiles struct {
	JobMetadata       string // job-metadata.json
	StatusLaunched    string // status-launched.sentinel  
	StatusRunning     string // status-running.sentinel
	StatusProgress    string // status-progress.json
	StatusCompleted   string // status-completed.sentinel
	StatusFailed      string // status-failed.sentinel
	StatusEmergency   string // status-emergency.sentinel
	BenchmarkResults  string // results.json
	BenchmarkLogs     string // benchmark.log
	SystemInfo        string // system-info.json
}

// NewS3SentinelFiles creates the S3 file paths for a job
func NewS3SentinelFiles(s3Prefix string) *S3SentinelFiles {
	return &S3SentinelFiles{
		JobMetadata:      s3Prefix + "job-metadata.json",
		StatusLaunched:   s3Prefix + "status-launched.sentinel",
		StatusRunning:    s3Prefix + "status-running.sentinel", 
		StatusProgress:   s3Prefix + "status-progress.json",
		StatusCompleted:  s3Prefix + "status-completed.sentinel",
		StatusFailed:     s3Prefix + "status-failed.sentinel",
		StatusEmergency:  s3Prefix + "status-emergency.sentinel",
		BenchmarkResults: s3Prefix + "results.json",
		BenchmarkLogs:    s3Prefix + "benchmark.log",
		SystemInfo:       s3Prefix + "system-info.json",
	}
}

// BenchmarkProgress tracks benchmark execution progress
type BenchmarkProgress struct {
	CurrentIteration int       `json:"current_iteration"`
	TotalIterations  int       `json:"total_iterations"`
	LastUpdate       time.Time `json:"last_update"`
	Message          string    `json:"message"`
	PercentComplete  float64   `json:"percent_complete"`
}

// AsyncBenchmarkResult contains the final results from an async benchmark
type AsyncBenchmarkResult struct {
	Job           *AsyncBenchmarkJob `json:"job"`
	BenchmarkData map[string]interface{} `json:"benchmark_data"`
	SystemInfo    map[string]interface{} `json:"system_info"`
	ExecutionTime time.Duration `json:"execution_time"`
	Success       bool `json:"success"`
	Error         string `json:"error,omitempty"`
}

// LaunchRequest contains parameters for launching async benchmarks
type LaunchRequest struct {
	Configs      []BenchmarkConfig `json:"configs"`
	S3Bucket     string           `json:"s3_bucket"`
	JobNamePrefix string          `json:"job_name_prefix"`
	MaxRuntime   time.Duration    `json:"max_runtime"`
	Tags         map[string]string `json:"tags"`
}

// LaunchResponse contains the results of launching async benchmarks
type LaunchResponse struct {
	Jobs         []*AsyncBenchmarkJob `json:"jobs"`
	LaunchedCount int                 `json:"launched_count"`
	FailedCount   int                 `json:"failed_count"`
	Errors       []string            `json:"errors,omitempty"`
}