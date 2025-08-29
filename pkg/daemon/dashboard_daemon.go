package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
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
)

// DashboardDaemon provides live monitoring of AWS benchmarks
type DashboardDaemon struct {
	ec2Client    *ec2.Client
	s3Client     *s3.Client
	bucketName   string
	region       string
	pollInterval time.Duration
	
	// Cached data
	mutex             sync.RWMutex
	benchmarkResults  []BenchmarkResult
	runningInstances  []RunningInstance
	lastUpdated       time.Time
}

// BenchmarkResult represents a completed benchmark
type BenchmarkResult struct {
	InstanceType   string    `json:"instance_type"`
	InstanceFamily string    `json:"instance_family"`
	Architecture   string    `json:"architecture"`
	TriadBandwidth float64   `json:"triad_bandwidth"`
	CopyBandwidth  float64   `json:"copy_bandwidth"`
	Status         string    `json:"status"`
	Timestamp      time.Time `json:"timestamp"`
	InstanceID     string    `json:"instance_id,omitempty"`
}

// RunningInstance represents a currently running benchmark
type RunningInstance struct {
	InstanceID     string    `json:"instance_id"`
	InstanceType   string    `json:"instance_type"`
	InstanceFamily string    `json:"instance_family"`
	Architecture   string    `json:"architecture"`
	Status         string    `json:"status"`
	LaunchTime     time.Time `json:"launch_time"`
	BenchmarkSuite string    `json:"benchmark_suite"`
}

// DashboardStatus represents the overall dashboard state
type DashboardStatus struct {
	RunningCount     int               `json:"running_count"`
	CompletedToday   int               `json:"completed_today"`
	InstanceTypes    int               `json:"instance_types"`
	TotalResults     int               `json:"total_results"`
	Results          []BenchmarkResult `json:"results"`
	RunningInstances []RunningInstance `json:"running_instances"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// NewDashboardDaemon creates a new dashboard daemon
func NewDashboardDaemon(ctx context.Context, bucketName, region, profile string) (*DashboardDaemon, error) {
	cfg, err := config.LoadDefaultConfig(ctx, 
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &DashboardDaemon{
		ec2Client:    ec2.NewFromConfig(cfg),
		s3Client:     s3.NewFromConfig(cfg),
		bucketName:   bucketName,
		region:       region,
		pollInterval: 10 * time.Second,
	}, nil
}

// Start begins the daemon's polling loops
func (d *DashboardDaemon) Start(ctx context.Context) error {
	log.Printf("🚀 Starting Dashboard Daemon (polling every %v)", d.pollInterval)
	
	// Initial data load
	if err := d.refreshData(ctx); err != nil {
		log.Printf("⚠️  Initial data load failed: %v", err)
	}

	// Start polling goroutines
	go d.pollLoop(ctx)
	
	return nil
}

// pollLoop continuously refreshes data
func (d *DashboardDaemon) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(d.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Dashboard daemon stopping...")
			return
		case <-ticker.C:
			if err := d.refreshData(ctx); err != nil {
				log.Printf("❌ Failed to refresh data: %v", err)
			}
		}
	}
}

// refreshData updates both S3 results and running instances
func (d *DashboardDaemon) refreshData(ctx context.Context) error {
	log.Println("🔄 Refreshing benchmark data...")
	
	// Try S3 results first, fallback to local files
	results, err := d.fetchS3Results(ctx)
	if err != nil {
		log.Printf("⚠️  Failed to fetch S3 results: %v", err)
	}

	// Refresh running instances  
	instances, err := d.fetchRunningInstances(ctx)
	if err != nil {
		log.Printf("⚠️  Failed to fetch running instances: %v", err)
	}

	// Update cached data atomically
	d.mutex.Lock()
	if err == nil {
		d.benchmarkResults = results
	}
	if err == nil {
		d.runningInstances = instances
	}
	d.lastUpdated = time.Now()
	d.mutex.Unlock()

	log.Printf("✅ Data refreshed: %d results, %d running instances", len(results), len(instances))
	return nil
}

// fetchS3Results gets benchmark results from S3
func (d *DashboardDaemon) fetchS3Results(ctx context.Context) ([]BenchmarkResult, error) {
	var allResults []BenchmarkResult
	
	// Search for results from recent dates
	now := time.Now()
	dates := []string{
		now.Format("2006/01/02"),              // Today
		now.AddDate(0, 0, 1).Format("2006/01/02"), // Tomorrow (timezone handling)
		now.AddDate(0, 0, -1).Format("2006/01/02"), // Yesterday
	}
	
	for _, date := range dates {
		resultsDir := filepath.Join("results", date)
		if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
			continue
		}
		
		log.Printf("🔍 Searching local directory: %s", resultsDir)
		
		err := filepath.WalkDir(resultsDir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			
			if !entry.IsDir() && strings.HasSuffix(path, ".json") {
				result, err := d.parseLocalResult(path)
				if err != nil {
					log.Printf("⚠️  Failed to parse %s: %v", path, err)
					return nil
				}
				
				if result != nil {
					allResults = append(allResults, *result)
				}
			}
			return nil
		})
		
		if err != nil {
			log.Printf("⚠️  Failed to walk directory %s: %v", resultsDir, err)
		}
	}
	
	log.Printf("✅ Found %d local results", len(allResults))
	return allResults, nil
}

// parseLocalResult reads and parses a local benchmark result file
func (d *DashboardDaemon) parseLocalResult(filepath string) (*BenchmarkResult, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	
	return d.parseResultData(jsonData, filepath)
}

// fetchS3Results retrieves benchmark results from S3
func (d *DashboardDaemon) parseS3Result(ctx context.Context, key string) (*BenchmarkResult, error) {
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(d.bucketName),
		Key:    aws.String(key),
	}

	resp, err := d.s3Client.GetObject(ctx, getInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return d.parseResultData(data, key)
}

// parseResultData extracts benchmark data from JSON
func (d *DashboardDaemon) parseResultData(data map[string]interface{}, key string) (*BenchmarkResult, error) {
	metadata, _ := data["metadata"].(map[string]interface{})
	if metadata == nil {
		return nil, fmt.Errorf("no metadata found")
	}

	// Extract performance data
	performance, _ := data["performance"].(map[string]interface{})
	if performance == nil {
		return nil, fmt.Errorf("no performance data found")
	}

	memory, _ := performance["memory"].(map[string]interface{})
	if memory == nil {
		return nil, fmt.Errorf("no memory performance data found")
	}

	stream, _ := memory["stream"].(map[string]interface{})
	if stream == nil {
		return nil, fmt.Errorf("no STREAM data found")
	}

	// Extract instance info
	instanceType, _ := metadata["instanceType"].(string)
	instanceFamily, _ := metadata["instanceFamily"].(string)
	architecture := d.getArchitectureName(instanceFamily)
	
	// Extract timestamp
	timestampStr, _ := metadata["timestamp"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp = time.Now()
	}

	// Extract bandwidth values
	triadBandwidth := d.extractBandwidth(stream, "triad")
	copyBandwidth := d.extractBandwidth(stream, "copy")

	if triadBandwidth == 0 {
		return nil, fmt.Errorf("no valid bandwidth data")
	}

	return &BenchmarkResult{
		InstanceType:   instanceType,
		InstanceFamily: instanceFamily,
		Architecture:   architecture,
		TriadBandwidth: triadBandwidth,
		CopyBandwidth:  copyBandwidth,
		Status:         "completed",
		Timestamp:      timestamp,
	}, nil
}

// extractBandwidth extracts bandwidth value from STREAM test data
func (d *DashboardDaemon) extractBandwidth(stream map[string]interface{}, test string) float64 {
	testData, ok := stream[test].(map[string]interface{})
	if !ok {
		return 0
	}

	bandwidth, ok := testData["bandwidth"].(float64)
	if !ok {
		return 0
	}

	return bandwidth
}

// getArchitectureName maps instance family to architecture name
func (d *DashboardDaemon) getArchitectureName(instanceFamily string) string {
	if strings.Contains(instanceFamily, "g") {
		return "AWS Graviton (ARM64)"
	}
	if strings.Contains(instanceFamily, "a") {
		return "AMD x86_64"
	}
	return "Intel x86_64"
}

// fetchRunningInstances gets currently running benchmark instances
func (d *DashboardDaemon) fetchRunningInstances(ctx context.Context) ([]RunningInstance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Purpose"),
				Values: []string{"aws-instance-benchmarks"},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "pending"},
			},
		},
	}

	resp, err := d.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []RunningInstance
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			// Extract benchmark suite from tags
			benchmarkSuite := "unknown"
			for _, tag := range instance.Tags {
				if aws.ToString(tag.Key) == "BenchmarkSuite" {
					benchmarkSuite = aws.ToString(tag.Value)
					break
				}
			}

			instanceType := string(instance.InstanceType)
			instanceFamily := strings.Split(instanceType, ".")[0]
			
			instances = append(instances, RunningInstance{
				InstanceID:     aws.ToString(instance.InstanceId),
				InstanceType:   instanceType,
				InstanceFamily: instanceFamily,
				Architecture:   d.getArchitectureName(instanceFamily),
				Status:         string(instance.State.Name),
				LaunchTime:     aws.ToTime(instance.LaunchTime),
				BenchmarkSuite: benchmarkSuite,
			})
		}
	}

	return instances, nil
}

// GetStatus returns current dashboard status
func (d *DashboardDaemon) GetStatus() DashboardStatus {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Calculate today's completions
	today := time.Now().Truncate(24 * time.Hour)
	completedToday := 0
	for _, result := range d.benchmarkResults {
		if result.Timestamp.After(today) {
			completedToday++
		}
	}

	// Calculate unique instance types
	instanceTypeMap := make(map[string]bool)
	for _, result := range d.benchmarkResults {
		instanceTypeMap[result.InstanceType] = true
	}
	for _, instance := range d.runningInstances {
		instanceTypeMap[instance.InstanceType] = true
	}

	return DashboardStatus{
		RunningCount:     len(d.runningInstances),
		CompletedToday:   completedToday,
		InstanceTypes:    len(instanceTypeMap),
		TotalResults:     len(d.benchmarkResults),
		Results:          d.benchmarkResults,
		RunningInstances: d.runningInstances,
		LastUpdated:      d.lastUpdated,
	}
}

// StartHTTPServer starts the HTTP API server
func (d *DashboardDaemon) StartHTTPServer(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/status", d.handleStatus)
	mux.HandleFunc("/api/results", d.handleResults)
	mux.HandleFunc("/api/instances", d.handleInstances)
	
	// Serve static dashboard files
	dashboardHTML := d.getDashboardHTML()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, dashboardHTML)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("🌐 Starting HTTP server on http://localhost:%d", port)
	return server.ListenAndServe()
}

// HTTP handlers
func (d *DashboardDaemon) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	status := d.GetStatus()
	json.NewEncoder(w).Encode(status)
}

func (d *DashboardDaemon) handleResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	d.mutex.RLock()
	results := d.benchmarkResults
	d.mutex.RUnlock()
	
	json.NewEncoder(w).Encode(results)
}

func (d *DashboardDaemon) handleInstances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	d.mutex.RLock()
	instances := d.runningInstances
	d.mutex.RUnlock()
	
	json.NewEncoder(w).Encode(instances)
}

// getDashboardHTML returns the embedded dashboard HTML
func (d *DashboardDaemon) getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AWS Instance Benchmarks - Live Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
    <style>
        .arch-intel { background-color: #0066cc; color: white; }
        .arch-amd { background-color: #ed1c24; color: white; }
        .arch-graviton { background-color: #ff9500; color: white; }
        .status-dot { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
        .status-dot.running { background-color: #ffc107; animation: pulse 2s infinite; }
        .status-dot.completed { background-color: #28a745; }
        @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }
        
        .sortable-header { cursor: pointer; user-select: none; }
        .sortable-header:hover { background-color: rgba(255,255,255,0.1); }
        .sortable-header .fas { opacity: 0.3; transition: opacity 0.3s; }
        .sortable-header:hover .fas { opacity: 1; }
        .sortable-header.sorted .fas { opacity: 1; color: #ffc107; }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
        <div class="container">
            <a class="navbar-brand" href="#"><i class="fas fa-chart-line"></i> AWS Instance Benchmarks <span class="badge bg-success">LOCAL</span></a>
        </div>
    </nav>
    
    <div class="container mt-4">
        <div class="row mb-4">
            <div class="col-md-3">
                <div class="card text-white bg-primary text-center">
                    <div class="card-body">
                        <i class="fas fa-server fa-2x mb-2"></i>
                        <h4 id="runningCount">-</h4>
                        <p>Running Benchmarks</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card text-white bg-success text-center">
                    <div class="card-body">
                        <i class="fas fa-check-circle fa-2x mb-2"></i>
                        <h4 id="completedToday">-</h4>
                        <p>Completed Today</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card text-white bg-info text-center">
                    <div class="card-body">
                        <i class="fas fa-microchip fa-2x mb-2"></i>
                        <h4 id="instanceTypes">-</h4>
                        <p>Instance Types</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card text-white bg-warning text-center">
                    <div class="card-body">
                        <i class="fas fa-database fa-2x mb-2"></i>
                        <h4 id="totalResults">-</h4>
                        <p>Total Results</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Running Benchmarks Section -->
        <div class="row mb-4" id="runningSection" style="display: none;">
            <div class="col-12">
                <div class="card border-warning">
                    <div class="card-header bg-warning text-dark">
                        <h5><i class="fas fa-spinner fa-pulse"></i> Currently Running Benchmarks</h5>
                    </div>
                    <div class="card-body">
                        <div class="table-responsive">
                            <table class="table table-striped">
                                <thead class="table-warning">
                                    <tr>
                                        <th>Instance ID</th>
                                        <th>Instance Type</th>
                                        <th>Architecture</th>
                                        <th>Benchmark Suite</th>
                                        <th>Runtime</th>
                                        <th>Status</th>
                                    </tr>
                                </thead>
                                <tbody id="runningTable">
                                    <tr><td colspan="6" class="text-center">No running benchmarks</td></tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Completed Results Section -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card">
                    <div class="card-header d-flex justify-content-between">
                        <h5><i class="fas fa-table"></i> Completed Benchmark Results</h5>
                        <span class="text-muted" id="lastUpdated">Loading...</span>
                    </div>
                    <div class="card-body">
                        <div class="table-responsive">
                            <table class="table table-striped sortable">
                                <thead class="table-dark">
                                    <tr>
                                        <th class="sortable-header" data-sort="instance_type">Instance <i class="fas fa-sort"></i></th>
                                        <th class="sortable-header" data-sort="architecture">Architecture <i class="fas fa-sort"></i></th>
                                        <th class="sortable-header" data-sort="triad_bandwidth">Triad Bandwidth <i class="fas fa-sort"></i></th>
                                        <th>Status</th>
                                        <th class="sortable-header" data-sort="timestamp">Completed <i class="fas fa-sort"></i></th>
                                    </tr>
                                </thead>
                                <tbody id="resultsTable">
                                    <tr><td colspan="5" class="text-center">Loading...</td></tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        function formatTimeAgo(timestamp) {
            const now = new Date();
            const time = new Date(timestamp);
            const diff = now - time;
            const hours = Math.floor(diff / (1000 * 60 * 60));
            const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
            if (hours < 1) return minutes + 'm ago';
            if (hours < 24) return hours + 'h ' + minutes + 'm ago';
            return Math.floor(hours/24) + 'd ago';
        }

        function getArchClass(arch) {
            if (arch.includes('Graviton')) return 'arch-graviton';
            if (arch.includes('AMD')) return 'arch-amd';
            return 'arch-intel';
        }

        let currentResults = [];
        let sortColumn = '';
        let sortDirection = 'asc';

        function formatRuntime(launchTime) {
            const now = new Date();
            const start = new Date(launchTime);
            const diff = now - start;
            const hours = Math.floor(diff / (1000 * 60 * 60));
            const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
            if (hours > 0) return hours + 'h ' + minutes + 'm';
            return minutes + 'm';
        }

        function updateRunningBenchmarks() {
            fetch('/api/instances')
                .then(res => res.json())
                .then(instances => {
                    const runningSection = document.getElementById('runningSection');
                    const runningTable = document.getElementById('runningTable');
                    
                    if (!instances || instances.length === 0) {
                        runningSection.style.display = 'none';
                        return;
                    }

                    runningSection.style.display = 'block';
                    runningTable.innerHTML = instances.map(instance => {
                        const archClass = getArchClass(instance.architecture);
                        const runtime = formatRuntime(instance.launch_time);
                        return '<tr>' +
                            '<td><code>' + instance.instance_id + '</code></td>' +
                            '<td><strong>' + instance.instance_type + '</strong></td>' +
                            '<td><span class="badge ' + archClass + '">' + (instance.architecture.includes('Graviton') ? 'Graviton' : (instance.architecture.includes('AMD') ? 'AMD' : 'Intel')) + '</span></td>' +
                            '<td>' + instance.benchmark_suite + '</td>' +
                            '<td>' + runtime + '</td>' +
                            '<td><span class="badge bg-warning"><i class="fas fa-spinner fa-pulse"></i> running</span></td>' +
                        '</tr>';
                    }).join('');
                })
                .catch(err => console.error('Failed to update running instances:', err));
        }

        function sortResults(column) {
            if (sortColumn === column) {
                sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
            } else {
                sortColumn = column;
                sortDirection = 'asc';
            }

            // Update header indicators
            document.querySelectorAll('.sortable-header').forEach(th => {
                th.classList.remove('sorted');
                th.querySelector('.fas').className = 'fas fa-sort';
            });

            const currentHeader = document.querySelector('[data-sort="' + column + '"]');
            currentHeader.classList.add('sorted');
            currentHeader.querySelector('.fas').className = sortDirection === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down';

            // Sort the results
            currentResults.sort((a, b) => {
                let aVal = a[column];
                let bVal = b[column];
                
                if (column === 'triad_bandwidth') {
                    aVal = parseFloat(aVal);
                    bVal = parseFloat(bVal);
                } else if (column === 'timestamp') {
                    aVal = new Date(aVal);
                    bVal = new Date(bVal);
                } else {
                    aVal = aVal.toString().toLowerCase();
                    bVal = bVal.toString().toLowerCase();
                }

                if (sortDirection === 'asc') {
                    return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
                } else {
                    return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
                }
            });

            renderResultsTable();
        }

        function renderResultsTable() {
            const tbody = document.getElementById('resultsTable');
            if (!currentResults || currentResults.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" class="text-center text-muted">No results available</td></tr>';
                return;
            }

            tbody.innerHTML = currentResults.map(result => {
                const archClass = getArchClass(result.architecture);
                const timeAgo = formatTimeAgo(result.timestamp);
                return '<tr>' +
                    '<td><strong>' + result.instance_type + '</strong></td>' +
                    '<td><span class="badge ' + archClass + '">' + (result.architecture.includes('Graviton') ? 'Graviton' : (result.architecture.includes('AMD') ? 'AMD' : 'Intel')) + '</span></td>' +
                    '<td><strong>' + result.triad_bandwidth.toFixed(1) + ' GB/s</strong></td>' +
                    '<td><span class="badge bg-success"><i class="fas fa-check-circle"></i> completed</span></td>' +
                    '<td><small>' + timeAgo + '</small></td>' +
                '</tr>';
            }).join('');
        }

        function updateDashboard() {
            fetch('/api/status')
                .then(res => res.json())
                .then(data => {
                    document.getElementById('runningCount').textContent = data.running_count;
                    document.getElementById('completedToday').textContent = data.completed_today;
                    document.getElementById('instanceTypes').textContent = data.instance_types;
                    document.getElementById('totalResults').textContent = data.total_results;
                    document.getElementById('lastUpdated').textContent = 'Updated: ' + new Date(data.last_updated).toLocaleTimeString();

                    currentResults = data.results || [];
                    renderResultsTable();
                })
                .catch(err => console.error('Failed to update dashboard:', err));
            
            updateRunningBenchmarks();
        }

        // Add click handlers for sortable columns
        document.addEventListener('DOMContentLoaded', function() {
            document.querySelectorAll('.sortable-header').forEach(header => {
                header.addEventListener('click', function() {
                    const column = this.getAttribute('data-sort');
                    sortResults(column);
                });
            });
        });

        // Update every 10 seconds
        updateDashboard();
        setInterval(updateDashboard, 10000);
    </script>
</body>
</html>`
}