# GitHub Pages Integration

## Overview

GitHub Pages provides a powerful platform for creating interactive web interfaces that can directly access benchmark data stored in the repository. This enables real-time data visualization, instance comparison tools, and public access to performance insights without requiring a separate hosting infrastructure.

## Web Interface Architecture

### 1. Static Site with Dynamic Data
- **GitHub Pages**: Static hosting with custom domain support
- **Client-Side Processing**: JavaScript-based data fetching and visualization
- **Direct Data Access**: No API server required, direct JSON consumption
- **Real-Time Updates**: Always reflects latest committed data

### 2. Site Structure
```
docs/                              # GitHub Pages source
├── index.html                     # Main dashboard
├── instance-selector/             # Instance comparison tool
│   ├── index.html
│   ├── js/
│   │   ├── selector.js
│   │   ├── comparison.js
│   │   └── visualization.js
│   └── css/
│       └── selector.css
├── architecture-analysis/         # Deep-dive analysis
│   ├── intel.html
│   ├── amd.html
│   ├── graviton.html
│   └── js/microarch.js
├── trends/                        # Performance trends
│   ├── index.html
│   └── js/trends.js
├── api-docs/                      # Data format documentation
│   ├── index.html
│   └── examples/
└── shared/                        # Common components
    ├── js/
    │   ├── data-client.js
    │   ├── charts.js
    │   └── utils.js
    └── css/
        └── common.css
```

## Interactive Components

### 1. Main Dashboard (`docs/index.html`)
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AWS Instance Benchmarks - Performance Database</title>
    <link rel="stylesheet" href="shared/css/common.css">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/d3@7"></script>
</head>
<body>
    <header>
        <h1>🚀 AWS Instance Benchmarks</h1>
        <nav>
            <a href="instance-selector/">Compare Instances</a>
            <a href="architecture-analysis/">Architecture Analysis</a>
            <a href="trends/">Performance Trends</a>
            <a href="api-docs/">API Documentation</a>
        </nav>
    </header>

    <main>
        <section class="overview">
            <div class="metrics-grid">
                <div class="metric-card" id="total-instances">
                    <h3>Instance Types</h3>
                    <span class="metric-value" id="instance-count">Loading...</span>
                </div>
                <div class="metric-card" id="architectures">
                    <h3>Architectures</h3>
                    <span class="metric-value" id="arch-count">Loading...</span>
                </div>
                <div class="metric-card" id="last-updated">
                    <h3>Last Updated</h3>
                    <span class="metric-value" id="update-date">Loading...</span>
                </div>
                <div class="metric-card" id="benchmark-runs">
                    <h3>Total Benchmark Runs</h3>
                    <span class="metric-value" id="run-count">Loading...</span>
                </div>
            </div>
        </section>

        <section class="performance-overview">
            <h2>🏆 Top Performers</h2>
            <div class="rankings-grid">
                <div class="ranking-card">
                    <h3>Memory Bandwidth (STREAM Triad)</h3>
                    <canvas id="memory-ranking-chart"></canvas>
                    <div id="memory-top-instances"></div>
                </div>
                <div class="ranking-card">
                    <h3>CPU Performance (HPL GFLOPS)</h3>
                    <canvas id="cpu-ranking-chart"></canvas>
                    <div id="cpu-top-instances"></div>
                </div>
            </div>
        </section>

        <section class="architecture-comparison">
            <h2>🏗️ Architecture Comparison</h2>
            <div class="arch-comparison-grid">
                <div class="arch-card intel">
                    <h3>Intel x86_64</h3>
                    <div class="arch-metrics" id="intel-metrics"></div>
                    <a href="architecture-analysis/intel.html" class="explore-btn">Explore Intel →</a>
                </div>
                <div class="arch-card amd">
                    <h3>AMD x86_64</h3>
                    <div class="arch-metrics" id="amd-metrics"></div>
                    <a href="architecture-analysis/amd.html" class="explore-btn">Explore AMD →</a>
                </div>
                <div class="arch-card graviton">
                    <h3>AWS Graviton ARM64</h3>
                    <div class="arch-metrics" id="graviton-metrics"></div>
                    <a href="architecture-analysis/graviton.html" class="explore-btn">Explore Graviton →</a>
                </div>
            </div>
        </section>

        <section class="quick-selector">
            <h2>🎯 Quick Instance Selector</h2>
            <div class="selector-form">
                <div class="workload-selector">
                    <label for="workload-type">Workload Type:</label>
                    <select id="workload-type">
                        <option value="general">General Purpose</option>
                        <option value="memory-intensive">Memory Intensive</option>
                        <option value="compute-intensive">Compute Intensive</option>
                        <option value="hpc">High Performance Computing</option>
                        <option value="ml-ai">Machine Learning / AI</option>
                    </select>
                </div>
                <div class="size-selector">
                    <label for="instance-size">Instance Size:</label>
                    <select id="instance-size">
                        <option value="small">Small (large)</option>
                        <option value="medium">Medium (xlarge)</option>
                        <option value="large">Large (2xlarge+)</option>
                        <option value="any">Any Size</option>
                    </select>
                </div>
                <button id="find-instances" class="primary-btn">Find Optimal Instances</button>
            </div>
            <div id="quick-results" class="results-grid"></div>
        </section>
    </main>

    <script src="shared/js/data-client.js"></script>
    <script src="shared/js/charts.js"></script>
    <script src="js/dashboard.js"></script>
</body>
</html>
```

### 2. Instance Selector Tool (`docs/instance-selector/index.html`)
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Instance Selector - AWS Benchmarks</title>
    <link rel="stylesheet" href="../shared/css/common.css">
    <link rel="stylesheet" href="css/selector.css">
</head>
<body>
    <header>
        <h1>🎯 Instance Selector & Comparison</h1>
        <nav>
            <a href="../">← Dashboard</a>
            <a href="../architecture-analysis/">Architecture Analysis</a>
            <a href="../trends/">Trends</a>
        </nav>
    </header>

    <main>
        <section class="filter-panel">
            <h2>Filter & Select Instances</h2>
            <div class="filters-grid">
                <div class="filter-group">
                    <label>Instance Families:</label>
                    <div class="checkbox-group" id="family-filters">
                        <!-- Dynamically populated -->
                    </div>
                </div>
                <div class="filter-group">
                    <label>Architectures:</label>
                    <div class="checkbox-group" id="arch-filters">
                        <label><input type="checkbox" value="intel" checked> Intel</label>
                        <label><input type="checkbox" value="amd" checked> AMD</label>
                        <label><input type="checkbox" value="graviton" checked> Graviton</label>
                    </div>
                </div>
                <div class="filter-group">
                    <label>Performance Metrics:</label>
                    <div class="checkbox-group" id="metric-filters">
                        <label><input type="checkbox" value="memory" checked> Memory Bandwidth</label>
                        <label><input type="checkbox" value="cpu" checked> CPU Performance</label>
                        <label><input type="checkbox" value="cost" checked> Price/Performance</label>
                        <label><input type="checkbox" value="microarch"> Microarchitecture</label>
                    </div>
                </div>
                <div class="filter-group">
                    <label>Workload Optimization:</label>
                    <select id="workload-optimization">
                        <option value="balanced">Balanced Performance</option>
                        <option value="memory-bound">Memory Bound</option>
                        <option value="cpu-bound">CPU Bound</option>
                        <option value="vectorization">Vectorization Heavy</option>
                        <option value="cache-sensitive">Cache Sensitive</option>
                        <option value="numa-aware">NUMA Aware</option>
                    </select>
                </div>
            </div>
            <button id="apply-filters" class="primary-btn">Apply Filters</button>
            <button id="clear-filters" class="secondary-btn">Clear All</button>
        </section>

        <section class="comparison-view">
            <div class="view-controls">
                <div class="view-selector">
                    <button class="view-btn active" data-view="table">📊 Table View</button>
                    <button class="view-btn" data-view="chart">📈 Chart View</button>
                    <button class="view-btn" data-view="radar">🎯 Radar View</button>
                </div>
                <div class="sort-controls">
                    <label for="sort-metric">Sort by:</label>
                    <select id="sort-metric">
                        <option value="overall-score">Overall Score</option>
                        <option value="memory-bandwidth">Memory Bandwidth</option>
                        <option value="cpu-performance">CPU Performance</option>
                        <option value="price-performance">Price/Performance</option>
                        <option value="instance-name">Instance Name</option>
                    </select>
                </div>
            </div>

            <div id="comparison-results">
                <!-- Dynamic content based on selected view -->
            </div>
        </section>

        <section class="detailed-comparison" id="selected-instances">
            <h2>Selected Instances Comparison</h2>
            <div class="comparison-grid" id="comparison-grid">
                <!-- Populated when instances are selected -->
            </div>
        </section>
    </main>

    <script src="../shared/js/data-client.js"></script>
    <script src="../shared/js/charts.js"></script>
    <script src="js/selector.js"></script>
    <script src="js/comparison.js"></script>
</body>
</html>
```

### 3. Data Client Library (`docs/shared/js/data-client.js`)
```javascript
class AWSBenchmarkClient {
    constructor() {
        this.baseUrl = 'https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest';
        this.cache = new Map();
        this.cacheTimeout = 30 * 60 * 1000; // 30 minutes
    }

    async getMetadata() {
        return this.fetchWithCache('metadata.json');
    }

    async getMemoryBenchmarks() {
        return this.fetchWithCache('memory-benchmarks.json');
    }

    async getCPUBenchmarks() {
        return this.fetchWithCache('cpu-benchmarks.json');
    }

    async getMicroarchBenchmarks() {
        return this.fetchWithCache('microarch-benchmarks.json');
    }

    async getInstanceRankings() {
        return this.fetchWithCache('instance-rankings.json');
    }

    async getPricePerformance() {
        return this.fetchWithCache('price-performance.json');
    }

    async getArchitectureData(architecture) {
        const data = await this.fetchWithCache(`${architecture}-instances.json`);
        return data;
    }

    async getAllData() {
        const [metadata, memory, cpu, microarch, rankings, pricing] = await Promise.all([
            this.getMetadata(),
            this.getMemoryBenchmarks(),
            this.getCPUBenchmarks(),
            this.getMicroarchBenchmarks(),
            this.getInstanceRankings(),
            this.getPricePerformance()
        ]);

        return { metadata, memory, cpu, microarch, rankings, pricing };
    }

    async fetchWithCache(endpoint) {
        const cacheKey = endpoint;
        const cached = this.cache.get(cacheKey);
        
        if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
            return cached.data;
        }

        try {
            const response = await fetch(`${this.baseUrl}/${endpoint}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            this.cache.set(cacheKey, {
                data,
                timestamp: Date.now()
            });
            
            return data;
        } catch (error) {
            console.error(`Failed to fetch ${endpoint}:`, error);
            // Return cached data if available, even if stale
            if (cached) {
                console.warn(`Using stale cached data for ${endpoint}`);
                return cached.data;
            }
            throw error;
        }
    }

    // Performance analysis utilities
    scoreInstance(instanceData, workloadProfile) {
        let score = 0;
        let maxScore = 0;

        // Memory performance scoring
        if (workloadProfile.memoryWeight && instanceData.memory) {
            const memoryScore = instanceData.memory.triad_bandwidth?.value || 0;
            score += memoryScore * workloadProfile.memoryWeight;
            maxScore += 100 * workloadProfile.memoryWeight; // Assume 100 GB/s as max
        }

        // CPU performance scoring
        if (workloadProfile.computeWeight && instanceData.cpu) {
            const cpuScore = instanceData.cpu.gflops?.value || 0;
            score += cpuScore * workloadProfile.computeWeight;
            maxScore += 1000 * workloadProfile.computeWeight; // Assume 1000 GFLOPS as max
        }

        // Price performance scoring (inverse - lower price is better)
        if (workloadProfile.costWeight && instanceData.pricing) {
            const costEfficiency = 1 / (instanceData.pricing.hourly_price || 1);
            score += costEfficiency * workloadProfile.costWeight * 100;
            maxScore += 100 * workloadProfile.costWeight;
        }

        return maxScore > 0 ? (score / maxScore) * 100 : 0;
    }

    filterInstances(data, filters) {
        let instances = Object.entries(data.memory.benchmarks.stream);

        // Filter by instance families
        if (filters.families && filters.families.length > 0) {
            instances = instances.filter(([instance, data]) => {
                const family = data.metadata.instance_family;
                return filters.families.includes(family);
            });
        }

        // Filter by architectures
        if (filters.architectures && filters.architectures.length > 0) {
            instances = instances.filter(([instance, data]) => {
                const architecture = data.metadata.architecture;
                return filters.architectures.includes(architecture);
            });
        }

        // Filter by minimum performance thresholds
        if (filters.minMemoryBandwidth) {
            instances = instances.filter(([instance, data]) => {
                return data.triad_bandwidth.value >= filters.minMemoryBandwidth;
            });
        }

        return instances;
    }
}

// Export for use in other scripts
window.AWSBenchmarkClient = AWSBenchmarkClient;
```

### 4. Chart Visualization Library (`docs/shared/js/charts.js`)
```javascript
class BenchmarkCharts {
    constructor() {
        this.defaultColors = {
            intel: '#0071c5',
            amd: '#ed1c24', 
            graviton: '#ff9900'
        };
    }

    createRankingChart(ctx, data, metric, title) {
        const labels = data.map(item => item.instance);
        const values = data.map(item => item.value);
        const colors = data.map(item => this.getArchitectureColor(item.architecture));

        return new Chart(ctx, {
            type: 'horizontalBar',
            data: {
                labels: labels,
                datasets: [{
                    label: title,
                    data: values,
                    backgroundColor: colors,
                    borderColor: colors,
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: metric
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            afterLabel: function(context) {
                                const item = data[context.dataIndex];
                                return [
                                    `Architecture: ${item.architecture}`,
                                    `Family: ${item.family}`,
                                    `Generation: ${item.generation}`
                                ];
                            }
                        }
                    }
                }
            }
        });
    }

    createArchitectureComparison(ctx, data, metric) {
        const architectures = ['intel', 'amd', 'graviton'];
        const datasets = [];

        architectures.forEach(arch => {
            const archData = data.filter(item => item.architecture === arch);
            const avgValue = archData.reduce((sum, item) => sum + item.value, 0) / archData.length;
            
            datasets.push({
                label: arch.charAt(0).toUpperCase() + arch.slice(1),
                data: [avgValue],
                backgroundColor: this.defaultColors[arch],
                borderColor: this.defaultColors[arch],
                borderWidth: 1
            });
        });

        return new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [metric],
                datasets: datasets
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    createRadarChart(ctx, instances, metrics) {
        const datasets = instances.map((instance, index) => ({
            label: instance.name,
            data: metrics.map(metric => instance.data[metric] || 0),
            borderColor: this.getInstanceColor(instance.name, index),
            backgroundColor: this.getInstanceColor(instance.name, index, 0.2),
            pointBackgroundColor: this.getInstanceColor(instance.name, index),
            pointBorderColor: '#fff',
            pointHoverBackgroundColor: '#fff',
            pointHoverBorderColor: this.getInstanceColor(instance.name, index)
        }));

        return new Chart(ctx, {
            type: 'radar',
            data: {
                labels: metrics.map(m => this.formatMetricName(m)),
                datasets: datasets
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    r: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });
    }

    createTrendChart(ctx, historicalData, metric) {
        const datasets = Object.keys(historicalData).map((instance, index) => ({
            label: instance,
            data: historicalData[instance].map(point => ({
                x: point.date,
                y: point[metric]
            })),
            borderColor: this.getInstanceColor(instance, index),
            backgroundColor: this.getInstanceColor(instance, index, 0.1),
            fill: false,
            tension: 0.1
        }));

        return new Chart(ctx, {
            type: 'line',
            data: {
                datasets: datasets
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'time',
                        time: {
                            unit: 'day'
                        }
                    },
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    getArchitectureColor(architecture) {
        return this.defaultColors[architecture] || '#666666';
    }

    getInstanceColor(instanceName, index, alpha = 1) {
        const architecture = this.extractArchitecture(instanceName);
        const baseColor = this.defaultColors[architecture] || '#666666';
        
        if (alpha === 1) {
            return baseColor;
        }
        
        // Convert hex to rgba with alpha
        const r = parseInt(baseColor.slice(1, 3), 16);
        const g = parseInt(baseColor.slice(3, 5), 16);
        const b = parseInt(baseColor.slice(5, 7), 16);
        
        return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    }

    extractArchitecture(instanceName) {
        if (instanceName.includes('g.') || instanceName.endsWith('g')) return 'graviton';
        if (instanceName.includes('a.') || instanceName.endsWith('a')) return 'amd';
        return 'intel';
    }

    formatMetricName(metric) {
        return metric.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
    }
}

window.BenchmarkCharts = BenchmarkCharts;
```

### 5. Dashboard Logic (`docs/js/dashboard.js`)
```javascript
class Dashboard {
    constructor() {
        this.client = new AWSBenchmarkClient();
        this.charts = new BenchmarkCharts();
        this.data = null;
        
        this.init();
    }

    async init() {
        try {
            await this.loadData();
            this.renderOverview();
            this.renderTopPerformers();
            this.renderArchitectureComparison();
            this.setupQuickSelector();
        } catch (error) {
            console.error('Failed to initialize dashboard:', error);
            this.showError('Failed to load benchmark data. Please try again later.');
        }
    }

    async loadData() {
        this.showLoading();
        this.data = await this.client.getAllData();
        this.hideLoading();
    }

    renderOverview() {
        const { metadata, memory } = this.data;
        
        // Update metrics cards
        document.getElementById('instance-count').textContent = metadata.total_instances;
        document.getElementById('arch-count').textContent = metadata.architectures.length;
        document.getElementById('update-date').textContent = new Date(metadata.last_updated).toLocaleDateString();
        document.getElementById('run-count').textContent = this.calculateTotalRuns();
    }

    renderTopPerformers() {
        const { rankings } = this.data;
        
        // Memory bandwidth top performers
        const memoryCtx = document.getElementById('memory-ranking-chart').getContext('2d');
        const topMemory = rankings.triad_bandwidth.slice(0, 10);
        this.charts.createRankingChart(memoryCtx, topMemory, 'Memory Bandwidth (GB/s)', 'Top Memory Performance');
        
        this.renderTopInstancesList('memory-top-instances', topMemory);

        // CPU performance top performers  
        const cpuCtx = document.getElementById('cpu-ranking-chart').getContext('2d');
        const topCPU = rankings.gflops.slice(0, 10);
        this.charts.createRankingChart(cpuCtx, topCPU, 'GFLOPS', 'Top CPU Performance');
        
        this.renderTopInstancesList('cpu-top-instances', topCPU);
    }

    renderTopInstancesList(elementId, instances) {
        const container = document.getElementById(elementId);
        const html = instances.slice(0, 5).map((instance, index) => `
            <div class="top-instance-item">
                <span class="rank">#${index + 1}</span>
                <span class="instance-name">${instance.instance}</span>
                <span class="instance-value">${instance.value.toFixed(1)}</span>
            </div>
        `).join('');
        
        container.innerHTML = html;
    }

    renderArchitectureComparison() {
        const { memory, cpu } = this.data;
        
        // Calculate architecture averages
        const architectures = ['intel', 'amd', 'graviton'];
        
        architectures.forEach(arch => {
            const instances = this.getInstancesByArchitecture(arch);
            const avgMemory = this.calculateAverage(instances, 'memory', 'triad_bandwidth');
            const avgCPU = this.calculateAverage(instances, 'cpu', 'gflops');
            
            const metricsHtml = `
                <div class="metric">
                    <span class="metric-label">Avg Memory:</span>
                    <span class="metric-value">${avgMemory.toFixed(1)} GB/s</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Avg CPU:</span>
                    <span class="metric-value">${avgCPU.toFixed(1)} GFLOPS</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Instance Count:</span>
                    <span class="metric-value">${instances.length}</span>
                </div>
            `;
            
            document.getElementById(`${arch}-metrics`).innerHTML = metricsHtml;
        });
    }

    setupQuickSelector() {
        const findBtn = document.getElementById('find-instances');
        findBtn.addEventListener('click', () => this.handleQuickSelection());
    }

    handleQuickSelection() {
        const workloadType = document.getElementById('workload-type').value;
        const instanceSize = document.getElementById('instance-size').value;
        
        const workloadProfiles = {
            'general': { memoryWeight: 0.4, computeWeight: 0.4, costWeight: 0.2 },
            'memory-intensive': { memoryWeight: 0.7, computeWeight: 0.2, costWeight: 0.1 },
            'compute-intensive': { memoryWeight: 0.2, computeWeight: 0.7, costWeight: 0.1 },
            'hpc': { memoryWeight: 0.5, computeWeight: 0.5, costWeight: 0.0 },
            'ml-ai': { memoryWeight: 0.3, computeWeight: 0.6, costWeight: 0.1 }
        };
        
        const profile = workloadProfiles[workloadType];
        const recommendations = this.getRecommendations(profile, instanceSize);
        this.renderQuickResults(recommendations);
    }

    getRecommendations(profile, sizeFilter) {
        const { memory, cpu, pricing } = this.data;
        const instances = [];
        
        for (const [instanceName, memoryData] of Object.entries(memory.benchmarks.stream)) {
            // Apply size filter
            if (sizeFilter !== 'any') {
                const size = this.extractInstanceSize(instanceName);
                if (!this.matchesSize(size, sizeFilter)) continue;
            }
            
            const cpuData = cpu.benchmarks.hpl[instanceName];
            const pricingData = pricing.instances[instanceName];
            
            if (!cpuData || !pricingData) continue;
            
            const instanceData = {
                memory: memoryData,
                cpu: cpuData,
                pricing: pricingData
            };
            
            const score = this.client.scoreInstance(instanceData, profile);
            
            instances.push({
                name: instanceName,
                score,
                data: instanceData
            });
        }
        
        return instances.sort((a, b) => b.score - a.score).slice(0, 6);
    }

    renderQuickResults(recommendations) {
        const container = document.getElementById('quick-results');
        
        if (recommendations.length === 0) {
            container.innerHTML = '<p>No instances match your criteria.</p>';
            return;
        }
        
        const html = recommendations.map(rec => `
            <div class="recommendation-card">
                <h4>${rec.name}</h4>
                <div class="score">Score: ${rec.score.toFixed(0)}/100</div>
                <div class="metrics">
                    <div>Memory: ${rec.data.memory.triad_bandwidth.value.toFixed(1)} GB/s</div>
                    <div>CPU: ${rec.data.cpu.gflops.value.toFixed(1)} GFLOPS</div>
                    <div>Price: $${rec.data.pricing.hourly_price.toFixed(3)}/hr</div>
                </div>
                <button onclick="window.open('./instance-selector/?instance=${rec.name}', '_blank')" class="detail-btn">
                    View Details →
                </button>
            </div>
        `).join('');
        
        container.innerHTML = html;
    }

    // Utility methods
    getInstancesByArchitecture(architecture) {
        const { memory } = this.data;
        return Object.entries(memory.benchmarks.stream)
            .filter(([instance, data]) => data.metadata.architecture === architecture)
            .map(([instance, data]) => ({ instance, ...data }));
    }

    calculateAverage(instances, dataType, metric) {
        const { cpu } = this.data;
        let sum = 0;
        let count = 0;
        
        instances.forEach(instance => {
            let value;
            if (dataType === 'memory') {
                value = instance[metric]?.value;
            } else if (dataType === 'cpu') {
                const cpuData = cpu.benchmarks.hpl[instance.instance];
                value = cpuData?.[metric]?.value;
            }
            
            if (value) {
                sum += value;
                count++;
            }
        });
        
        return count > 0 ? sum / count : 0;
    }

    calculateTotalRuns() {
        const { metadata } = this.data;
        return metadata.total_benchmark_runs || 0;
    }

    extractInstanceSize(instanceName) {
        const parts = instanceName.split('.');
        return parts[1] || 'unknown';
    }

    matchesSize(size, filter) {
        const sizeMap = {
            'small': ['large'],
            'medium': ['xlarge'],
            'large': ['2xlarge', '4xlarge', '8xlarge', '16xlarge']
        };
        
        return sizeMap[filter]?.includes(size) || false;
    }

    showLoading() {
        // Add loading indicator
        document.body.classList.add('loading');
    }

    hideLoading() {
        document.body.classList.remove('loading');
    }

    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;
        document.body.prepend(errorDiv);
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new Dashboard();
});
```

## GitHub Pages Configuration

### 1. Repository Settings
```yaml
# _config.yml
title: AWS Instance Benchmarks
description: Open database of comprehensive performance benchmarks for AWS EC2 instances
baseurl: "/aws-instance-benchmarks"
url: "https://scttfrdmn.github.io"

# GitHub Pages settings
source: docs
plugins:
  - jekyll-sitemap
  - jekyll-feed

# Custom domain (optional)
# custom_domain: benchmarks.computecompass.dev
```

### 2. Automated Deployment
```yaml
# .github/workflows/pages.yml
name: Deploy GitHub Pages
on:
  push:
    branches: [ main ]
    paths: 
      - 'docs/**'
      - 'data/**'
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pages: write
      id-token: write
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Pages
        uses: actions/configure-pages@v2
        
      - name: Build site
        run: |
          # Optional: Build step if using Jekyll or other static site generator
          echo "Building static site..."
          
      - name: Upload artifact
        uses: actions/upload-pages-artifact@v1
        with:
          path: docs
          
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v1
```

## Integration with ComputeCompass

### 1. Widget Embedding
```html
<!-- Embed performance widget in ComputeCompass -->
<iframe 
    src="https://scttfrdmn.github.io/aws-instance-benchmarks/instance-selector/?embed=true&instances=m7i.large,c7g.large"
    width="100%" 
    height="400"
    frameborder="0">
</iframe>
```

### 2. Direct API Integration
```javascript
// In ComputeCompass application
import { AWSBenchmarkClient } from 'https://scttfrdmn.github.io/aws-instance-benchmarks/shared/js/data-client.js';

class ComputeCompassIntegration {
    constructor() {
        this.benchmarkClient = new AWSBenchmarkClient();
    }
    
    async enhanceInstanceRecommendations(instances) {
        const benchmarkData = await this.benchmarkClient.getAllData();
        
        return instances.map(instance => ({
            ...instance,
            performance: this.getBenchmarkSummary(instance.type, benchmarkData),
            score: this.calculateComputeCompassScore(instance, benchmarkData)
        }));
    }
}
```

## Benefits of GitHub Pages Integration

### 1. Zero Infrastructure Costs
- **Free Hosting**: GitHub Pages provides free static hosting
- **Global CDN**: Automatic worldwide distribution
- **SSL/TLS**: Free HTTPS certificates
- **Custom Domains**: Support for custom domain names

### 2. Developer Experience
- **Direct Data Access**: No API server required
- **Real-time Updates**: Reflects latest committed data
- **Version Control**: Complete history of data and interface changes
- **Community Contributions**: Easy forking and pull requests

### 3. Performance Benefits
- **Client-side Processing**: Leverages user's browser compute
- **Caching**: Browser and CDN caching for fast repeated access
- **Responsive Design**: Works on desktop and mobile devices
- **Progressive Enhancement**: Graceful degradation for older browsers

This GitHub Pages integration provides a powerful, cost-effective way to make benchmark data accessible through an interactive web interface while enabling seamless integration with tools like ComputeCompass.