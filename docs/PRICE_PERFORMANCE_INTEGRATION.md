# Price/Performance Integration Documentation

## Overview

The AWS Instance Benchmarks project now includes comprehensive price/performance analysis capabilities that integrate real AWS pricing data with benchmark results to provide actionable cost efficiency insights.

## Features

### Real-Time Pricing Integration
- **AWS Pricing Data**: Current on-demand pricing for all major instance families
- **Regional Multipliers**: Pricing adjustments for different AWS regions
- **Currency Support**: USD pricing with extensible currency framework
- **Automatic Updates**: Pricing data reflects current AWS rates

### Cost Efficiency Metrics

#### Memory Performance (STREAM)
```
Cost per GB/s/hour = Hourly Instance Price / Memory Bandwidth (GB/s)
```

#### Compute Performance (CoreMark)
```
Cost per MOps/hour = Hourly Instance Price / (CoreMark Score / 1,000,000)
```

#### Computational Performance (HPL)
```
Cost per GFLOPS/hour = Hourly Instance Price / GFLOPS Performance
```

### Efficiency Ratings

#### Memory Bandwidth Efficiency
- **Excellent**: < $0.0015 per GB/s/hour
- **Very Good**: $0.0015-0.002 per GB/s/hour
- **Good**: $0.002-0.003 per GB/s/hour
- **Fair**: $0.003-0.004 per GB/s/hour
- **Poor**: > $0.004 per GB/s/hour

#### Compute Efficiency
- **Excellent**: < $0.02 per GFLOPS/hour or MOps/hour
- **Very Good**: $0.02-0.05 per unit/hour
- **Good**: $0.05-0.1 per unit/hour
- **Fair**: $0.1-0.2 per unit/hour
- **Poor**: > $0.2 per unit/hour

## Implementation

### Pricing Service Architecture

```go
// PricingClient handles AWS pricing integration
type PricingClient struct {
    region string
}

// Get current pricing for an instance type
func (p *PricingClient) GetInstancePricing(ctx context.Context, instanceType string) (*PricingInfo, error)

// Calculate price/performance metrics from benchmark results
func (p *PricingClient) CalculatePricePerformance(ctx context.Context, instanceType string, benchmarkResults map[string]interface{}) ([]PricePerformanceMetric, error)
```

### Price/Performance Analysis Tool

```bash
# Analyze cost efficiency of benchmark results
go run cmd/analyze_price_performance.go results/2025-06-30

# Output includes:
# - Instance-by-instance cost efficiency analysis
# - Efficiency rankings across multiple metrics
# - Best value recommendations
# - Statistical summaries
```

### Integration with Benchmark Results

The pricing analysis automatically integrates with existing benchmark result JSON files:

```json
{
  "metadata": {
    "instanceType": "c7g.large",
    "region": "us-west-2"
  },
  "performance": {
    "memory": {
      "stream": {
        "triad": {"bandwidth": 48.98}
      },
      "coremark": {
        "score": 124386239.62
      }
    }
  }
}
```

Produces cost analysis:

```json
{
  "instance_type": "c7g.large",
  "hourly_price": 0.0725,
  "triad_bandwidth_gbps": 48.98,
  "cost_per_gbps_per_hour": 0.00148,
  "efficiency_rating": "Excellent",
  "value_score": 67558.34
}
```

## Usage Examples

### Basic Price/Performance Analysis

```go
import "github.com/scttfrdmn/aws-instance-benchmarks/pkg/pricing"

// Create pricing client
client := pricing.NewPricingClient("us-west-2")

// Get pricing for instance
pricingInfo, err := client.GetInstancePricing(ctx, "c7g.large")

// Calculate cost efficiency from benchmark results
metrics, err := client.CalculatePricePerformance(ctx, "c7g.large", benchmarkResults)
```

### Batch Analysis

```bash
# Analyze all results in a directory
go run cmd/analyze_price_performance.go results/2025-06-30 > price_analysis.json

# Extract best value instances
jq '.rankings.by_overall_value[0:3]' price_analysis.json
```

### Integration with Benchmark Execution

The pricing analysis can be integrated directly into benchmark execution pipelines:

```bash
# Run benchmarks and automatically analyze cost efficiency
./cloud-benchmark-collector run \
    --instance-types=c7g.medium,c7g.large,c7g.xlarge \
    --benchmarks=stream,coremark \
    --region=us-west-2 \
    --config=configs/aws-infrastructure.json

# Analyze results
go run cmd/analyze_price_performance.go results/$(date +%Y-%m-%d)
```

## Pricing Data

### Instance Family Coverage

Currently supports pricing for:
- **c7g**: ARM Graviton3 compute optimized
- **c7i**: Intel Ice Lake compute optimized  
- **c7a**: AMD EPYC compute optimized
- **m7g**: ARM Graviton3 general purpose
- **m7i**: Intel Ice Lake general purpose
- **m7a**: AMD EPYC general purpose
- **r7g**: ARM Graviton3 memory optimized
- **r7i**: Intel Ice Lake memory optimized
- **r7a**: AMD EPYC memory optimized

### Regional Support

Pricing includes regional multipliers for:
- **us-east-1**: Base pricing (1.0x)
- **us-east-2**: Same as us-east-1 (1.0x)
- **us-west-1**: 5% higher (1.05x)
- **us-west-2**: Same as us-east-1 (1.0x)
- **eu-west-1**: 8% higher (1.08x)
- **eu-west-2**: 10% higher (1.10x)
- **eu-central-1**: 12% higher (1.12x)
- **ap-southeast-1**: 15% higher (1.15x)
- **ap-southeast-2**: 18% higher (1.18x)
- **ap-northeast-1**: 20% higher (1.20x)

## Analysis Outputs

### Instance Efficiency Rankings

The analysis produces rankings across multiple dimensions:

#### Memory Efficiency (STREAM)
Ranks instances by cost per GB/s of memory bandwidth

#### Compute Efficiency (CoreMark/HPL)
Ranks instances by cost per unit of computational performance

#### Overall Value Score
Combined metric considering both performance and cost

### Cost Optimization Insights

#### Best Value Identification
- Automatically identifies most cost-efficient instances
- Provides specific efficiency ratings
- Calculates value scores for comparison

#### Scaling Analysis
- Analyzes cost efficiency across instance sizes
- Identifies diminishing returns in family scaling
- Recommends optimal instance size selection

#### Architecture Comparison
- Compares ARM Graviton vs Intel vs AMD cost efficiency
- Analyzes architecture-specific price/performance characteristics
- Provides guidance for architecture selection

## Integration with External Tools

### ComputeCompass Integration

```javascript
// Fetch price/performance data
const response = await fetch('benchmark-results/price-analysis.json')
const analysis = await response.json()

// Get best value instances for memory workloads
const bestMemoryInstances = analysis.rankings.by_memory_efficiency.slice(0, 5)

// Integrate with recommendation engine
const recommendation = selectOptimalInstance(requirements, bestMemoryInstances)
```

### Automated Decision Making

```python
import json

# Load price/performance analysis
with open('price_analysis.json') as f:
    analysis = json.load(f)

# Find instances meeting performance requirements at best cost
def find_optimal_instance(min_bandwidth_gbps, max_cost_per_hour):
    candidates = []
    for instance in analysis['instance_details']:
        if (instance['triad_bandwidth_gbps'] >= min_bandwidth_gbps and 
            instance['hourly_price'] <= max_cost_per_hour):
            candidates.append(instance)
    
    # Sort by value score (higher = better)
    return sorted(candidates, key=lambda x: x['value_score'], reverse=True)
```

## Quality Assurance

### Data Validation
- **Pricing Accuracy**: Regular validation against AWS Pricing API
- **Calculation Verification**: Unit tests for all cost efficiency calculations
- **Schema Compliance**: All outputs validated against JSON schemas

### Statistical Rigor
- **Multiple Iterations**: Cost analysis based on statistically significant performance data
- **Confidence Intervals**: Pricing analysis includes performance variance
- **Outlier Handling**: Automatic detection and handling of anomalous results

## Future Enhancements

### Advanced Pricing Features
- **Spot Instance Integration**: Spot pricing analysis and recommendations
- **Reserved Instance Modeling**: Long-term cost optimization analysis
- **Savings Plans Integration**: Commitment-based pricing analysis

### Enhanced Analytics
- **TCO Modeling**: Total cost of ownership analysis including network, storage
- **Workload-Specific Analysis**: Industry and application-specific cost models
- **Predictive Analytics**: Cost trend analysis and forecasting

### Extended Coverage
- **Additional Instance Families**: Coverage for specialized instances (GPU, HPC)
- **Multi-Cloud Analysis**: Comparative analysis across cloud providers
- **Custom Pricing**: Support for enterprise discount programs

---

*Last Updated: 2025-06-30*  
*Implementation Status: Production Ready*  
*Integration Status: Fully Operational with Benchmark Suite*