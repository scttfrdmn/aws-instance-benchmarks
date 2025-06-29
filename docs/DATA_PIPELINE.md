# Data Pipeline Architecture

## Overview

The AWS Instance Benchmarks project uses a hybrid storage strategy that leverages GitHub for public data distribution while using S3 only for temporary collection storage. This approach eliminates egress costs, provides free global CDN access, and ensures data persistence.

## Storage Strategy

### GitHub-First Architecture
- **Primary Storage**: GitHub repository with processed benchmark data
- **Public Access**: Free global access via GitHub Raw URLs
- **Version Control**: Complete data history and change tracking
- **No Egress Costs**: GitHub serves data globally at no charge
- **CDN Performance**: GitHub's global CDN provides fast access

### S3 Temporary Storage
- **Collection Only**: Temporary storage during benchmark execution
- **Processing Pipeline**: Data aggregation and validation
- **Auto-Migration**: Automated transfer to GitHub
- **Cost Minimization**: Short retention period, lifecycle policies

## Data Structure on GitHub

```
data/
├── processed/
│   ├── latest/                           # Current dataset
│   │   ├── memory-benchmarks.json       # STREAM results
│   │   ├── cpu-benchmarks.json          # HPL results  
│   │   ├── microarch-benchmarks.json    # Microarchitecture-specific
│   │   ├── instance-rankings.json       # Performance rankings
│   │   ├── price-performance.json       # Cost analysis
│   │   └── metadata.json                # Dataset information
│   ├── v1.0/                           # Versioned snapshots
│   │   ├── 2024-06-29/                 # Date-based releases
│   │   └── weekly-aggregates/          # Weekly summaries
│   └── historical/                     # Time-series data
│       ├── 2024-Q2/                   # Quarterly aggregates
│       └── trends/                     # Performance trends
├── schemas/
│   ├── v1.0.0/                        # Schema versions
│   │   ├── memory-benchmark.schema.json
│   │   ├── cpu-benchmark.schema.json
│   │   └── microarch-benchmark.schema.json
│   └── migrations/                     # Schema migration guides
└── api/
    ├── endpoints.json                  # API endpoint definitions
    ├── filters.json                   # Supported filters
    └── examples/                      # Usage examples
```

## Data Processing Pipeline

### 1. Collection Phase (S3)
```mermaid
graph LR
    A[Benchmark Execution] --> B[S3 Raw Storage]
    B --> C[Validation Pipeline]
    C --> D[Aggregation Engine]
    D --> E[GitHub Publication]
    E --> F[S3 Cleanup]
```

### 2. Processing Components

#### Raw Data Validation
```go
type RawDataProcessor struct {
    S3Bucket        string
    ValidationRules []ValidationRule
    QualityThreshold float64
    OutputFormat    string
}

func (rdp *RawDataProcessor) ProcessDailyResults(date time.Time) (*ProcessedResults, error) {
    // 1. Fetch raw results from S3
    // 2. Validate against schemas
    // 3. Apply quality filters
    // 4. Aggregate by instance family
    // 5. Calculate statistical measures
    // 6. Generate processed dataset
}
```

#### GitHub Data Publisher
```go
type GitHubPublisher struct {
    Repository  string
    Branch      string
    AuthToken   string
    CommitInfo  CommitMetadata
}

func (gp *GitHubPublisher) PublishProcessedData(data *ProcessedResults) error {
    // 1. Update latest/ directory
    // 2. Create versioned snapshot
    // 3. Update metadata and indices
    // 4. Commit with automated message
    // 5. Trigger downstream notifications
}
```

## Data Formats for Consumption

### 1. Memory Benchmarks
```json
{
  "schema_version": "1.1.0",
  "last_updated": "2024-06-29T18:00:00Z",
  "total_instances": 67,
  "architectures": ["intel", "amd", "graviton"],
  "benchmarks": {
    "stream": {
      "m7i.large": {
        "triad_bandwidth": {
          "value": 47.9,
          "unit": "GB/s", 
          "confidence_interval": [47.2, 48.6],
          "sample_size": 5,
          "coefficient_variation": 1.2
        },
        "copy_bandwidth": { "value": 51.2, "unit": "GB/s" },
        "metadata": {
          "instance_family": "m7i",
          "architecture": "intel",
          "generation": 7,
          "processor": "Intel Xeon Scalable (Ice Lake)",
          "memory_type": "DDR4",
          "numa_nodes": 1
        }
      }
    },
    "stream_cache": {
      "m7i.large": {
        "l1_bandwidth": { "value": 892.3, "unit": "GB/s" },
        "l2_bandwidth": { "value": 234.7, "unit": "GB/s" },
        "l3_bandwidth": { "value": 156.8, "unit": "GB/s" },
        "cache_hierarchy_efficiency": 0.94
      }
    }
  },
  "rankings": {
    "triad_bandwidth": [
      {"instance": "r7i.2xlarge", "value": 95.2},
      {"instance": "m7i.2xlarge", "value": 94.8},
      {"instance": "c7i.2xlarge", "value": 93.1}
    ]
  }
}
```

### 2. CPU Benchmarks
```json
{
  "schema_version": "1.1.0",
  "benchmarks": {
    "hpl": {
      "m7i.large": {
        "gflops": {
          "value": 42.3,
          "unit": "GFLOPS",
          "theoretical_peak": 89.6,
          "efficiency_percent": 47.2
        },
        "single_thread_gflops": { "value": 5.2, "unit": "GFLOPS" },
        "vectorization_speedup": 3.8
      }
    },
    "hpl_mkl": {
      "m7i.large": {
        "gflops": { "value": 67.8, "unit": "GFLOPS" },
        "mkl_speedup": 1.6,
        "library_efficiency": 0.76
      }
    }
  }
}
```

### 3. Microarchitecture Benchmarks
```json
{
  "schema_version": "1.1.0",
  "microarchitecture": {
    "intel": {
      "avx512_performance": {
        "m7i.large": {
          "vector_bandwidth": { "value": 89.2, "unit": "GB/s" },
          "scalar_bandwidth": { "value": 47.9, "unit": "GB/s" },
          "vectorization_ratio": 1.86,
          "avx512_efficiency": 0.91
        }
      },
      "cache_analysis": {
        "m7i.large": {
          "l1_latency": { "value": 1.2, "unit": "cycles" },
          "l2_latency": { "value": 4.8, "unit": "cycles" },
          "l3_latency": { "value": 18.5, "unit": "cycles" },
          "dram_latency": { "value": 89.2, "unit": "nanoseconds" }
        }
      }
    }
  }
}
```

## Access Patterns for Sister Applications

### 1. ComputeCompass Integration

#### Direct GitHub Access
```javascript
class AWSBenchmarkClient {
  constructor(baseUrl = 'https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest') {
    this.baseUrl = baseUrl;
    this.cache = new Map();
    this.cacheTimeout = 60 * 60 * 1000; // 1 hour
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

  async fetchWithCache(endpoint) {
    const cacheKey = endpoint;
    const cached = this.cache.get(cacheKey);
    
    if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
      return cached.data;
    }

    const response = await fetch(`${this.baseUrl}/${endpoint}`);
    const data = await response.json();
    
    this.cache.set(cacheKey, {
      data,
      timestamp: Date.now()
    });
    
    return data;
  }
}
```

#### Performance-Aware Instance Selection
```javascript
class PerformanceSelector {
  constructor(benchmarkClient) {
    this.client = benchmarkClient;
  }

  async selectOptimalInstance(workloadProfile) {
    const [memory, cpu, microarch] = await Promise.all([
      this.client.getMemoryBenchmarks(),
      this.client.getCPUBenchmarks(), 
      this.client.getMicroarchBenchmarks()
    ]);

    return this.scoreInstances(workloadProfile, { memory, cpu, microarch });
  }

  scoreInstances(profile, benchmarks) {
    const scores = new Map();

    for (const [instance, data] of Object.entries(benchmarks.memory.benchmarks.stream)) {
      let score = 0;

      // Memory-intensive workload scoring
      if (profile.memoryWeight > 0) {
        score += data.triad_bandwidth.value * profile.memoryWeight;
      }

      // Compute-intensive workload scoring
      if (profile.computeWeight > 0) {
        const cpuData = benchmarks.cpu.benchmarks.hpl[instance];
        if (cpuData) {
          score += cpuData.gflops.value * profile.computeWeight;
        }
      }

      // Architecture-specific optimizations
      if (profile.vectorization && benchmarks.microarch.intel?.[instance]) {
        score *= benchmarks.microarch.intel[instance].vectorization_ratio || 1;
      }

      scores.set(instance, {
        score,
        instance,
        metadata: data.metadata
      });
    }

    return Array.from(scores.values())
      .sort((a, b) => b.score - a.score)
      .slice(0, 10);
  }
}
```

### 2. Research Tool Integration

#### Python Data Science Access
```python
import requests
import pandas as pd
import json
from typing import Dict, List, Optional

class AWSBenchmarkDataset:
    def __init__(self, base_url: str = "https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest"):
        self.base_url = base_url
        self._cache = {}
    
    def load_memory_benchmarks(self) -> pd.DataFrame:
        """Load memory benchmark data as pandas DataFrame"""
        data = self._fetch_json('memory-benchmarks.json')
        
        records = []
        for instance, benchmarks in data['benchmarks']['stream'].items():
            record = {
                'instance_type': instance,
                'instance_family': benchmarks['metadata']['instance_family'],
                'architecture': benchmarks['metadata']['architecture'],
                'triad_bandwidth': benchmarks['triad_bandwidth']['value'],
                'copy_bandwidth': benchmarks['copy_bandwidth']['value'],
                'confidence_interval_lower': benchmarks['triad_bandwidth']['confidence_interval'][0],
                'confidence_interval_upper': benchmarks['triad_bandwidth']['confidence_interval'][1],
                'sample_size': benchmarks['triad_bandwidth']['sample_size']
            }
            records.append(record)
        
        return pd.DataFrame(records)
    
    def load_microarch_analysis(self, architecture: str) -> Dict:
        """Load microarchitecture-specific analysis"""
        data = self._fetch_json('microarch-benchmarks.json')
        return data['microarchitecture'].get(architecture, {})
    
    def performance_comparison(self, instances: List[str], metric: str = 'triad_bandwidth') -> pd.DataFrame:
        """Compare performance across specific instances"""
        df = self.load_memory_benchmarks()
        filtered = df[df['instance_type'].isin(instances)]
        return filtered.sort_values(metric, ascending=False)
    
    def _fetch_json(self, endpoint: str) -> Dict:
        if endpoint in self._cache:
            return self._cache[endpoint]
        
        response = requests.get(f"{self.base_url}/{endpoint}")
        response.raise_for_status()
        data = response.json()
        
        self._cache[endpoint] = data
        return data

# Usage Example
dataset = AWSBenchmarkDataset()
memory_df = dataset.load_memory_benchmarks()

# Find best memory bandwidth instances
top_memory = memory_df.nlargest(10, 'triad_bandwidth')
print(top_memory[['instance_type', 'triad_bandwidth', 'architecture']])

# Architecture comparison
intel_analysis = dataset.load_microarch_analysis('intel')
graviton_analysis = dataset.load_microarch_analysis('graviton')
```

### 3. API-Style Access Patterns

#### RESTful-Style Endpoints via GitHub
```
# Base URL: https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest

# Core datasets
GET /memory-benchmarks.json           # All memory performance data
GET /cpu-benchmarks.json              # All CPU performance data  
GET /microarch-benchmarks.json        # Microarchitecture analysis
GET /instance-rankings.json           # Performance rankings
GET /price-performance.json           # Cost analysis

# Filtered views (generated during processing)
GET /intel-instances.json             # Intel-only instances
GET /amd-instances.json               # AMD-only instances
GET /graviton-instances.json          # Graviton-only instances
GET /compute-optimized.json           # C-family instances
GET /memory-optimized.json            # R-family instances
GET /general-purpose.json             # M-family instances

# Specific metrics
GET /stream-results.json              # STREAM-only results
GET /hpl-results.json                 # HPL-only results
GET /avx512-analysis.json             # AVX-512 specific analysis
GET /numa-performance.json            # NUMA topology analysis

# Metadata and schemas
GET /metadata.json                    # Dataset information
GET /../../schemas/v1.1.0/memory-benchmark.schema.json
```

## Data Processing Automation

### 1. GitHub Actions Pipeline
```yaml
name: Process and Publish Benchmark Data
on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM UTC
  workflow_dispatch:

jobs:
  process-data:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'
          
      - name: Process S3 Data
        run: |
          ./aws-benchmark-collector process \
            --s3-bucket aws-instance-benchmarks-data-us-east-1 \
            --output-dir data/processed/latest \
            --generate-rankings \
            --calculate-trends
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          
      - name: Validate Processed Data
        run: |
          ./aws-benchmark-collector schema validate data/processed/latest/
          
      - name: Generate API Indices
        run: |
          ./aws-benchmark-collector generate-indices \
            --input data/processed/latest \
            --output data/api
            
      - name: Commit and Push
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add data/
          git commit -m "Update benchmark data - $(date -u +%Y-%m-%d)" || exit 0
          git push
```

### 2. Data Processing Commands
```bash
# Process daily results from S3
./aws-benchmark-collector process daily \
    --date 2024-06-29 \
    --s3-bucket aws-instance-benchmarks-data-us-east-1 \
    --output data/processed/latest

# Generate architectural summaries  
./aws-benchmark-collector summarize \
    --input data/processed/latest \
    --group-by architecture \
    --output data/processed/latest/architectural-summary.json

# Create performance rankings
./aws-benchmark-collector rank \
    --input data/processed/latest \
    --metrics triad_bandwidth,gflops,cost_efficiency \
    --output data/processed/latest/instance-rankings.json

# Generate trend analysis
./aws-benchmark-collector trends \
    --historical data/processed/historical \
    --output data/processed/latest/trends.json
```

## Cost and Performance Benefits

### 1. Cost Elimination
- **No S3 Egress**: GitHub serves data globally at no cost
- **CDN Performance**: GitHub's CDN provides fast global access
- **Storage Efficiency**: Git compression reduces storage requirements
- **Bandwidth Savings**: No AWS data transfer charges

### 2. Performance Advantages
- **Global Distribution**: GitHub's worldwide CDN
- **HTTP/2 Support**: Modern protocol optimizations
- **Caching**: Browser and proxy caching support
- **Compression**: Automatic gzip compression

### 3. Operational Benefits
- **Version Control**: Complete data history and rollback capability
- **Change Tracking**: Git diffs show exactly what changed
- **Community Access**: Open source friendly distribution
- **API Stability**: Consistent URLs and data contracts

## Implementation Roadmap

### Phase 1: Core Pipeline (Week 1)
- [ ] Build S3 to GitHub data processor
- [ ] Implement daily aggregation pipeline
- [ ] Create basic processed data formats
- [ ] Set up GitHub Actions automation

### Phase 2: Enhanced Formats (Week 2)  
- [ ] Generate architectural summaries
- [ ] Build performance ranking system
- [ ] Create filtered dataset views
- [ ] Implement trend analysis

### Phase 3: Integration Support (Week 3)
- [ ] Build client libraries (JS, Python)
- [ ] Create API documentation
- [ ] Develop usage examples
- [ ] Performance optimization

### Phase 4: Advanced Features (Week 4)
- [ ] Real-time data updates
- [ ] Custom filtering endpoints
- [ ] Performance prediction models
- [ ] Community contribution pipeline

This architecture provides a scalable, cost-effective way to distribute benchmark data while enabling sophisticated analysis tools like ComputeCompass to make intelligent, data-driven instance recommendations.