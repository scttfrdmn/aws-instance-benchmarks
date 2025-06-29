# Git-Native Data Storage Strategy

## Overview

Using Git as the primary data storage mechanism provides powerful versioning, change tracking, and historical analysis capabilities for benchmark data. This approach treats performance data as "code" with full Git benefits: diffs, branches, tags, and comprehensive audit trails.

## Git-Based Data Architecture

### 1. Repository Structure
```
aws-instance-benchmarks/
├── data/
│   ├── statistical/                    # Git-tracked statistical data
│   │   ├── memory/
│   │   │   ├── stream/
│   │   │   │   ├── m7i.large.json     # Instance-specific stats
│   │   │   │   ├── m7a.large.json
│   │   │   │   └── ...
│   │   │   ├── stream-cache/
│   │   │   └── stream-numa/
│   │   ├── cpu/
│   │   │   ├── hpl/
│   │   │   ├── hpl-mkl/
│   │   │   └── hpl-vector/
│   │   └── microarch/
│   │       ├── intel/
│   │       ├── amd/
│   │       └── graviton/
│   ├── aggregated/                     # Cross-instance analysis
│   │   ├── family-summaries/
│   │   │   ├── m7i-family.json        # Family-level statistics
│   │   │   ├── c7g-family.json
│   │   │   └── ...
│   │   ├── architecture-summaries/
│   │   │   ├── intel-summary.json     # Architecture aggregates
│   │   │   ├── amd-summary.json
│   │   │   └── graviton-summary.json
│   │   └── temporal/
│   │       ├── daily-averages/
│   │       ├── weekly-trends/
│   │       └── monthly-reports/
│   ├── metadata/
│   │   ├── collection-info.json       # When/how data was collected
│   │   ├── schema-versions.json       # Data format evolution
│   │   └── quality-reports.json       # Data quality assessments
│   └── indices/                       # Fast lookup structures
│       ├── by-performance.json
│       ├── by-architecture.json
│       └── by-cost-efficiency.json
├── schemas/                           # Data validation schemas
│   ├── v1.0/
│   ├── v1.1/
│   └── current -> v1.1
└── tools/                            # Data processing scripts
    ├── collectors/
    ├── processors/
    └── analyzers/
```

### 2. Statistical Data Format

#### Individual Instance Statistics (`data/statistical/memory/stream/m7i.large.json`)
```json
{
  "schema_version": "1.1.0",
  "instance_type": "m7i.large",
  "metadata": {
    "instance_family": "m7i",
    "architecture": "intel",
    "generation": 7,
    "processor_model": "Intel Xeon Scalable (Ice Lake)",
    "vcpu_count": 2,
    "memory_gb": 8,
    "last_updated": "2024-06-29T18:00:00Z",
    "collection_periods": [
      {
        "start": "2024-06-20",
        "end": "2024-06-29",
        "sample_count": 15,
        "collection_method": "automated"
      }
    ]
  },
  "benchmarks": {
    "stream": {
      "triad_bandwidth": {
        "statistics": {
          "mean": 47.92,
          "median": 47.85,
          "std_dev": 0.58,
          "min": 46.9,
          "max": 49.1,
          "coefficient_variation": 1.21,
          "sample_count": 15,
          "confidence_interval_95": {
            "lower": 47.61,
            "upper": 48.23
          },
          "outliers_removed": 1,
          "quality_score": 0.96
        },
        "unit": "GB/s",
        "measurement_conditions": {
          "numa_policy": "interleave",
          "compiler": "gcc-11",
          "optimization": "-O3 -march=native -mavx2",
          "array_size": "100MB",
          "iterations": 10
        },
        "historical_data": [
          {
            "date": "2024-06-20",
            "value": 47.3,
            "run_id": "run-20240620-001"
          },
          {
            "date": "2024-06-21", 
            "value": 48.1,
            "run_id": "run-20240621-001"
          }
        ]
      },
      "copy_bandwidth": {
        "statistics": {
          "mean": 51.24,
          "median": 51.18,
          "std_dev": 0.62,
          "min": 50.1,
          "max": 52.3,
          "coefficient_variation": 1.21,
          "sample_count": 15,
          "confidence_interval_95": {
            "lower": 50.89,
            "upper": 51.59
          }
        },
        "unit": "GB/s"
      },
      "scale_bandwidth": {
        "statistics": {
          "mean": 50.78,
          "std_dev": 0.71,
          "sample_count": 15,
          "confidence_interval_95": {
            "lower": 50.40,
            "upper": 51.16
          }
        },
        "unit": "GB/s"
      },
      "add_bandwidth": {
        "statistics": {
          "mean": 48.34,
          "std_dev": 0.59,
          "sample_count": 15,
          "confidence_interval_95": {
            "lower": 48.02,
            "upper": 48.66
          }
        },
        "unit": "GB/s"
      }
    }
  },
  "performance_characteristics": {
    "memory_consistency": {
      "coefficient_variation": 1.21,
      "performance_stability": "excellent"
    },
    "thermal_behavior": {
      "sustained_performance_ratio": 0.98,
      "thermal_throttling_observed": false
    },
    "numa_efficiency": {
      "local_vs_remote_ratio": 2.34,
      "numa_balancing_effectiveness": 0.89
    }
  },
  "data_provenance": {
    "collection_runs": [
      {
        "run_id": "batch-20240620-001",
        "timestamp": "2024-06-20T14:30:00Z",
        "region": "us-east-1",
        "az": "us-east-1c",
        "instance_id": "i-0123456789abcdef0",
        "ami_id": "ami-0abcd1234efgh5678",
        "container_image": "public.ecr.aws/aws-benchmarks/stream:intel-icelake",
        "scheduler_job_id": "weekly-plan-job-142"
      }
    ],
    "validation": {
      "schema_validated": true,
      "statistical_tests_passed": true,
      "outlier_detection_applied": true,
      "peer_review_status": "validated"
    }
  }
}
```

#### Family-Level Aggregation (`data/aggregated/family-summaries/m7i-family.json`)
```json
{
  "schema_version": "1.1.0",
  "family": "m7i",
  "metadata": {
    "architecture": "intel",
    "generation": 7,
    "processor_family": "Intel Xeon Scalable (Ice Lake)",
    "instance_sizes": ["large", "xlarge", "2xlarge", "4xlarge", "8xlarge"],
    "total_instances_analyzed": 5,
    "last_updated": "2024-06-29T18:00:00Z",
    "analysis_period": {
      "start": "2024-06-20",
      "end": "2024-06-29",
      "total_samples": 75
    }
  },
  "performance_scaling": {
    "memory_bandwidth": {
      "stream_triad": {
        "scaling_factor": 1.98,
        "scaling_efficiency": 0.99,
        "per_vcpu_performance": {
          "mean": 23.96,
          "std_dev": 0.34,
          "unit": "GB/s per vCPU"
        },
        "scaling_curve": [
          {"size": "large", "vcpu": 2, "bandwidth": 47.92},
          {"size": "xlarge", "vcpu": 4, "bandwidth": 95.84},
          {"size": "2xlarge", "vcpu": 8, "bandwidth": 189.2},
          {"size": "4xlarge", "vcpu": 16, "bandwidth": 374.8},
          {"size": "8xlarge", "vcpu": 32, "bandwidth": 742.1}
        ]
      }
    },
    "cpu_performance": {
      "hpl_gflops": {
        "scaling_factor": 1.95,
        "scaling_efficiency": 0.97,
        "per_vcpu_performance": {
          "mean": 21.15,
          "std_dev": 0.89,
          "unit": "GFLOPS per vCPU"
        }
      }
    }
  },
  "microarchitecture_analysis": {
    "avx512_efficiency": {
      "mean": 0.91,
      "std_dev": 0.03,
      "across_sizes": "consistent"
    },
    "cache_hierarchy": {
      "l3_cache_per_vcpu": 1.25,
      "cache_scaling": "linear",
      "cache_efficiency": 0.94
    },
    "memory_subsystem": {
      "numa_topology": "2_nodes_per_socket",
      "memory_channels": 8,
      "memory_bandwidth_per_channel": "12.5_GB/s"
    }
  },
  "cost_analysis": {
    "price_per_gflops": {
      "mean": 0.0047,
      "std_dev": 0.0003,
      "unit": "$/hour/GFLOPS",
      "trend": "decreasing_with_size"
    },
    "price_per_gb_bandwidth": {
      "mean": 0.0021,
      "std_dev": 0.0001,
      "unit": "$/hour/(GB/s)",
      "sweet_spot": "2xlarge"
    }
  }
}
```

#### Architecture Summary (`data/aggregated/architecture-summaries/intel-summary.json`)
```json
{
  "schema_version": "1.1.0",
  "architecture": "intel",
  "metadata": {
    "processor_generations": ["skylake", "cascade_lake", "ice_lake"],
    "instance_families": ["m5", "m6i", "m7i", "c5", "c6i", "c7i", "r5", "r6i", "r7i"],
    "total_instances": 45,
    "last_updated": "2024-06-29T18:00:00Z"
  },
  "performance_characteristics": {
    "memory_performance": {
      "stream_triad_range": {
        "min": 12.3,
        "max": 742.1,
        "median": 94.7,
        "unit": "GB/s"
      },
      "per_vcpu_efficiency": {
        "mean": 23.2,
        "std_dev": 2.1,
        "unit": "GB/s per vCPU"
      }
    },
    "cpu_performance": {
      "hpl_range": {
        "min": 8.4,
        "max": 1247.3,
        "median": 156.7,
        "unit": "GFLOPS"
      },
      "per_vcpu_efficiency": {
        "mean": 19.8,
        "std_dev": 3.2,
        "unit": "GFLOPS per vCPU"
      }
    },
    "vectorization": {
      "avx2_speedup": {
        "mean": 3.2,
        "std_dev": 0.4
      },
      "avx512_speedup": {
        "mean": 5.8,
        "std_dev": 0.7,
        "availability": "ice_lake_and_newer"
      }
    }
  },
  "generation_comparison": {
    "skylake": {
      "families": ["m5", "c5", "r5"],
      "avg_memory_bandwidth_per_vcpu": 18.4,
      "avg_cpu_performance_per_vcpu": 16.2,
      "vectorization_support": "avx2"
    },
    "cascade_lake": {
      "families": ["m5n", "c5n", "r5n"],
      "avg_memory_bandwidth_per_vcpu": 20.1,
      "avg_cpu_performance_per_vcpu": 17.8,
      "vectorization_support": "avx512_limited"
    },
    "ice_lake": {
      "families": ["m6i", "c6i", "r6i", "m7i", "c7i", "r7i"],
      "avg_memory_bandwidth_per_vcpu": 24.6,
      "avg_cpu_performance_per_vcpu": 21.3,
      "vectorization_support": "avx512_full"
    }
  },
  "optimization_insights": {
    "compiler_recommendations": {
      "preferred": "intel_oneapi",
      "flags": "-O3 -march=native -mtune=native -mavx512f",
      "library_optimizations": {
        "mkl": "recommended",
        "performance_gain": "15-30%"
      }
    },
    "numa_considerations": {
      "numa_aware_allocation": "critical_for_large_instances",
      "performance_impact": "up_to_40%_degradation_if_ignored"
    }
  }
}
```

## Git Workflow for Data Management

### 1. Data Collection Workflow
```bash
# Create feature branch for new data collection
git checkout -b data-collection-2024-06-29

# Automated data collection updates files
./aws-benchmark-collector process daily \
    --date 2024-06-29 \
    --update-git-data \
    --validate-statistics

# Git tracks all changes automatically
git add data/statistical/
git add data/aggregated/

# Descriptive commit with statistical summary
git commit -m "Add benchmark data for 2024-06-29

- 67 instances benchmarked across stream and hpl suites
- Statistical validation: 98.5% of samples passed quality checks
- New instance families: m7i, c7i, r7i with complete microarch analysis
- Average coefficient of variation: 1.2% (excellent consistency)
- Confidence intervals: 95% CI calculated for all metrics

Detailed changes:
- Updated 67 instance-specific statistical files
- Refreshed family aggregations for m7*, c7*, r7* families  
- Regenerated architecture summaries with latest performance data
- Quality score improvements: +0.03 average across all metrics

🤖 Generated with automated collection pipeline
Co-Authored-By: AWS Benchmark Collector <noreply@benchmarks.dev>"

# Merge to main after validation
git checkout main
git merge data-collection-2024-06-29
git tag -a v2024.06.29 -m "Benchmark data release 2024-06-29"
git push origin main --tags
```

### 2. Data Analysis with Git History
```bash
# View performance evolution over time
git log --oneline --grep="benchmark data" data/statistical/memory/stream/m7i.large.json

# Compare performance between dates
git diff v2024.06.20..v2024.06.29 data/statistical/memory/stream/m7i.large.json

# Show statistical changes for specific metric
git log -p --follow data/statistical/memory/stream/m7i.large.json | grep -A5 -B5 "triad_bandwidth"

# Generate performance trend report
git log --since="30 days ago" --pretty=format:"%h %ad %s" --date=short \
    data/aggregated/family-summaries/m7i-family.json
```

### 3. Advanced Git Data Operations

#### Statistical Diff Tool
```bash
#!/bin/bash
# tools/git-stat-diff.sh - Custom diff tool for statistical data

TEMP_OLD=$(mktemp)
TEMP_NEW=$(mktemp)

# Extract just the statistics for comparison
jq '.benchmarks.stream.triad_bandwidth.statistics' "$1" > "$TEMP_OLD"
jq '.benchmarks.stream.triad_bandwidth.statistics' "$2" > "$TEMP_NEW"

echo "=== Statistical Changes ==="
echo "Mean: $(jq -r '.mean' "$TEMP_OLD") → $(jq -r '.mean' "$TEMP_NEW")"
echo "Std Dev: $(jq -r '.std_dev' "$TEMP_OLD") → $(jq -r '.std_dev' "$TEMP_NEW")"
echo "Sample Count: $(jq -r '.sample_count' "$TEMP_OLD") → $(jq -r '.sample_count' "$TEMP_NEW")"
echo "Quality Score: $(jq -r '.quality_score' "$TEMP_OLD") → $(jq -r '.quality_score' "$TEMP_NEW")"

# Calculate statistical significance
python3 -c "
import json
import scipy.stats as stats

with open('$TEMP_OLD') as f:
    old = json.load(f)
with open('$TEMP_NEW') as f:
    new = json.load(f)

# T-test for mean difference significance
t_stat, p_value = stats.ttest_ind_from_stats(
    old['mean'], old['std_dev'], old['sample_count'],
    new['mean'], new['std_dev'], new['sample_count']
)

print(f'Statistical significance: p = {p_value:.6f}')
if p_value < 0.05:
    print('✅ Statistically significant change')
else:
    print('❌ Not statistically significant')
"

rm "$TEMP_OLD" "$TEMP_NEW"
```

#### Performance Regression Detection
```bash
#!/bin/bash
# tools/regression-detector.sh

echo "🔍 Checking for performance regressions..."

# Compare last 5 commits for significant drops
for commit in $(git log --oneline -5 --pretty=format:"%h"); do
    echo "Analyzing commit: $commit"
    
    # Extract all triad_bandwidth means from this commit
    git show "$commit:data/statistical/memory/stream/*.json" 2>/dev/null | \
    jq -r '.benchmarks.stream.triad_bandwidth.statistics.mean' | \
    awk '{sum+=$1; count++} END {print "Average bandwidth:", sum/count, "GB/s"}'
done
```

### 4. Automated Data Quality with Git Hooks

#### Pre-commit Hook (`.git/hooks/pre-commit`)
```bash
#!/bin/bash
# Validate data quality before allowing commits

echo "🔍 Validating benchmark data quality..."

# Check schema compliance
for file in $(git diff --cached --name-only --diff-filter=AM | grep "data/statistical.*\.json$"); do
    echo "Validating schema: $file"
    if ! ./tools/validate-schema.py "$file"; then
        echo "❌ Schema validation failed for $file"
        exit 1
    fi
done

# Statistical quality checks
for file in $(git diff --cached --name-only --diff-filter=AM | grep "data/statistical.*\.json$"); do
    echo "Checking statistical quality: $file"
    
    # Extract coefficient of variation
    cv=$(jq -r '.benchmarks.stream.triad_bandwidth.statistics.coefficient_variation' "$file" 2>/dev/null)
    
    if [ "$cv" != "null" ] && [ "$(echo "$cv > 5.0" | bc)" -eq 1 ]; then
        echo "❌ High coefficient of variation ($cv%) in $file"
        echo "   Data quality may be insufficient (target: <5%)"
        exit 1
    fi
    
    # Check sample size
    samples=$(jq -r '.benchmarks.stream.triad_bandwidth.statistics.sample_count' "$file" 2>/dev/null)
    
    if [ "$samples" != "null" ] && [ "$samples" -lt 3 ]; then
        echo "❌ Insufficient samples ($samples) in $file"
        echo "   Minimum 3 samples required for statistical validity"
        exit 1
    fi
done

echo "✅ All data quality checks passed"
```

#### Post-commit Hook (`.git/hooks/post-commit`)
```bash
#!/bin/bash
# Generate updated aggregations after new data commits

# Check if this commit includes statistical data
if git diff-tree --no-commit-id --name-only -r HEAD | grep -q "data/statistical"; then
    echo "📊 Regenerating aggregated statistics..."
    
    # Update family summaries
    ./tools/generate-family-summaries.py
    
    # Update architecture summaries  
    ./tools/generate-architecture-summaries.py
    
    # Update performance indices
    ./tools/generate-performance-indices.py
    
    # Auto-commit the generated aggregations
    git add data/aggregated/
    git commit --amend --no-edit
    
    echo "✅ Aggregated statistics updated"
fi
```

## Data Processing Pipeline Integration

### 1. S3 to Git Data Processor
```go
// tools/collectors/s3-to-git.go
package main

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
)

type GitDataProcessor struct {
    repoPath     string
    branchPrefix string
    commitTemplate string
}

func (gdp *GitDataProcessor) ProcessDailyResults(date time.Time, s3Results []S3BenchmarkResult) error {
    // Create feature branch for this collection
    branchName := fmt.Sprintf("%s%s", gdp.branchPrefix, date.Format("2006-01-02"))
    
    if err := gdp.createBranch(branchName); err != nil {
        return fmt.Errorf("failed to create branch: %w", err)
    }
    
    // Process each result into statistical format
    for _, result := range s3Results {
        statsData := gdp.convertToStatisticalFormat(result)
        
        // Update or create the instance statistics file
        filePath := fmt.Sprintf("data/statistical/memory/stream/%s.json", result.InstanceType)
        if err := gdp.updateInstanceStats(filePath, statsData); err != nil {
            return fmt.Errorf("failed to update instance stats: %w", err)
        }
    }
    
    // Generate aggregated summaries
    if err := gdp.generateAggregations(); err != nil {
        return fmt.Errorf("failed to generate aggregations: %w", err)
    }
    
    // Commit with descriptive message
    commitMsg := gdp.generateCommitMessage(date, s3Results)
    if err := gdp.commitChanges(commitMsg); err != nil {
        return fmt.Errorf("failed to commit changes: %w", err)
    }
    
    // Merge to main and tag
    if err := gdp.mergeToMain(branchName); err != nil {
        return fmt.Errorf("failed to merge to main: %w", err)
    }
    
    return gdp.createTag(date)
}

func (gdp *GitDataProcessor) updateInstanceStats(filePath string, newData StatisticalData) error {
    // Read existing data if file exists
    existingData, err := gdp.readExistingStats(filePath)
    if err != nil && !os.IsNotExist(err) {
        return err
    }
    
    // Merge new data with existing historical data
    mergedData := gdp.mergeStatisticalData(existingData, newData)
    
    // Recalculate aggregate statistics
    mergedData.RecalculateStatistics()
    
    // Write back to file
    return gdp.writeStatsFile(filePath, mergedData)
}

func (gdp *GitDataProcessor) generateCommitMessage(date time.Time, results []S3BenchmarkResult) string {
    summary := gdp.calculateCollectionSummary(results)
    
    return fmt.Sprintf(`Add benchmark data for %s

- %d instances benchmarked across %d benchmark suites
- Statistical validation: %.1f%% of samples passed quality checks
- Average coefficient of variation: %.1f%% (excellent consistency)
- Confidence intervals: 95%% CI calculated for all metrics

Performance highlights:
- Top memory bandwidth: %s (%.1f GB/s)
- Top CPU performance: %s (%.1f GFLOPS)
- New performance records: %d instances

Quality metrics:
- Total samples collected: %d
- Outliers removed: %d (%.1f%%)
- Quality score improvements: +%.2f average

🤖 Generated with automated collection pipeline
Co-Authored-By: AWS Benchmark Collector <noreply@benchmarks.dev>`,
        date.Format("2006-01-02"),
        summary.InstanceCount,
        summary.BenchmarkSuites,
        summary.QualityPassRate,
        summary.AvgCoefficientVariation,
        summary.TopMemoryInstance, summary.TopMemoryBandwidth,
        summary.TopCPUInstance, summary.TopCPUPerformance,
        summary.NewRecords,
        summary.TotalSamples,
        summary.OutliersRemoved, summary.OutlierRate,
        summary.QualityImprovement)
}
```

## Benefits of Git-Native Data Storage

### 1. Complete Audit Trail
- **Performance Evolution**: Track how instance performance changes over time
- **Data Quality History**: See improvements in measurement consistency
- **Methodology Changes**: Understand impact of benchmark methodology updates
- **Statistical Validation**: Every change is validated and documented

### 2. Advanced Analytics
```bash
# Performance regression analysis
git log --oneline --since="1 month ago" data/statistical/memory/stream/m7i.large.json | \
while read commit msg; do
    bandwidth=$(git show "$commit:data/statistical/memory/stream/m7i.large.json" | \
               jq -r '.benchmarks.stream.triad_bandwidth.statistics.mean')
    echo "$commit: $bandwidth GB/s"
done

# Statistical significance testing between versions
git-stat-diff HEAD~1:data/statistical/memory/stream/m7i.large.json \
               HEAD:data/statistical/memory/stream/m7i.large.json

# Identify best-performing commit for specific instance
git log --oneline data/statistical/memory/stream/m7i.large.json | \
while read commit msg; do
    perf=$(git show "$commit:data/statistical/memory/stream/m7i.large.json" | \
           jq -r '.benchmarks.stream.triad_bandwidth.statistics.mean')
    echo "$perf $commit $msg"
done | sort -nr | head -1
```

### 3. Data Science Integration
```python
# tools/analyzers/git-performance-analysis.py
import subprocess
import json
import pandas as pd
from datetime import datetime

class GitPerformanceAnalyzer:
    def __init__(self, repo_path):
        self.repo_path = repo_path
    
    def get_performance_timeline(self, instance_type, metric='triad_bandwidth'):
        """Extract performance timeline from git history"""
        file_path = f"data/statistical/memory/stream/{instance_type}.json"
        
        # Get all commits that modified this file
        cmd = f"git log --oneline --follow -- {file_path}"
        commits = subprocess.check_output(cmd, shell=True, cwd=self.repo_path).decode().strip().split('\n')
        
        timeline = []
        for commit_line in commits:
            commit_hash = commit_line.split()[0]
            
            # Get file content at this commit
            cmd = f"git show {commit_hash}:{file_path}"
            try:
                content = subprocess.check_output(cmd, shell=True, cwd=self.repo_path).decode()
                data = json.loads(content)
                
                # Extract metric value
                value = data['benchmarks']['stream'][metric]['statistics']['mean']
                
                # Get commit date
                cmd = f"git show -s --format=%ci {commit_hash}"
                date_str = subprocess.check_output(cmd, shell=True, cwd=self.repo_path).decode().strip()
                commit_date = datetime.strptime(date_str.split()[0], '%Y-%m-%d')
                
                timeline.append({
                    'date': commit_date,
                    'commit': commit_hash,
                    'value': value,
                    'instance': instance_type
                })
                
            except subprocess.CalledProcessError:
                continue  # Skip if file didn't exist in this commit
        
        return pd.DataFrame(timeline).sort_values('date')
    
    def detect_performance_regressions(self, instance_type, threshold=0.05):
        """Detect statistically significant performance regressions"""
        timeline = self.get_performance_timeline(instance_type)
        
        regressions = []
        for i in range(1, len(timeline)):
            current = timeline.iloc[i]
            previous = timeline.iloc[i-1]
            
            change_pct = (current['value'] - previous['value']) / previous['value']
            
            if change_pct < -threshold:  # Performance dropped by more than threshold
                regressions.append({
                    'date': current['date'],
                    'commit': current['commit'],
                    'change_percent': change_pct * 100,
                    'previous_value': previous['value'],
                    'current_value': current['value']
                })
        
        return regressions
```

This Git-native approach transforms benchmark data from static files into a living, versioned dataset with complete provenance tracking and powerful analytical capabilities. Every performance change is documented, validated, and can be precisely attributed to specific collection periods or methodology improvements.