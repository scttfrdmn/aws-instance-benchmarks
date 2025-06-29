# AWS Instance Benchmarks

An open database of comprehensive performance benchmarks for AWS EC2 instances, designed to enable data-driven instance selection for research computing workloads.

## 🎯 Mission

Provide the research computing community with deep, microarchitectural performance data for AWS instances that goes beyond published specifications. Enable tools like [ComputeCompass](https://github.com/scttfrdmn/computecompass) to make intelligent, performance-aware recommendations.

## 📊 What's Included

### Memory Performance
- **STREAM Benchmarks**: Copy, Scale, Add, Triad operations across all memory types
- **Cache Hierarchy**: L1/L2/L3 latency and bandwidth measurements  
- **NUMA Topology**: Multi-socket performance characteristics
- **Access Patterns**: Sequential, random, sparse memory access benchmarks

### CPU Performance  
- **LINPACK**: Peak GFLOPS and sustained performance
- **CoreMark**: Integer performance and efficiency metrics
- **Vectorization**: SSE, AVX, AVX-512, ARM Neon, SVE performance
- **Microarchitecture**: Pipeline efficiency, branch prediction, ILP analysis

### Cost Analysis
- **Price/Performance**: Cost per GFLOP, cost per GB/s memory bandwidth
- **Spot Pricing**: Historical availability and cost savings
- **Architecture Comparison**: Intel vs AMD vs Graviton efficiency

## 🛠️ Methodology

All benchmarks are executed using:
- **Spack**: Architecture-optimized compilation (Intel OneAPI, AMD AOCC, GCC)
- **Containers**: Reproducible environments with consistent toolchains
- **Multiple Runs**: Statistical validation with confidence intervals
- **NUMA Awareness**: Proper memory affinity and scaling analysis

## 📁 Data Structure

```
data/
├── processed/
│   ├── latest/
│   │   ├── memory-benchmarks.json      # STREAM, cache, NUMA data
│   │   ├── cpu-benchmarks.json         # LINPACK, CoreMark, vectorization
│   │   ├── instance-rankings.json      # Performance rankings by category
│   │   └── price-performance.json      # Cost efficiency analysis
│   └── historical/                     # Time-series data
├── raw/                               # Raw benchmark outputs by date
└── schemas/                           # JSON schemas for validation
```

## 🚀 Quick Start

### **CLI Tool Installation**
```bash
# Clone and build
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks
go build -o aws-benchmark-collector ./cmd

# Verify installation
./aws-benchmark-collector --help
```

### **Basic Usage**
```bash
# Discover AWS instance types and generate architecture mappings
./aws-benchmark-collector discover --update-containers

# Build optimized benchmark containers
./aws-benchmark-collector build \
    --architectures intel-icelake,amd-zen4,graviton3 \
    --benchmarks stream

# Run comprehensive benchmarks across multiple instance types
./aws-benchmark-collector run \
    --instance-types m7i.large,m7a.large,m7g.large,c7i.large,c7a.large,c7g.large \
    --region us-east-1 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --s3-bucket my-benchmark-bucket \
    --benchmarks stream,hpl \
    --iterations 3 \
    --max-concurrency 8

# Run benchmarks with custom S3 storage
./aws-benchmark-collector run \
    --instance-types m7i.large,c7g.large \
    --region us-west-2 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --s3-bucket aws-instance-benchmarks-data-us-west-2 \
    --benchmarks stream \
    --iterations 5

# Schema validation and migration
./aws-benchmark-collector schema validate results/ --version 1.0.0
./aws-benchmark-collector schema migrate legacy/ migrated/ --version 1.0.0
```

### **Using the Data**
```javascript
// Fetch latest benchmark data
const response = await fetch('https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest/memory-benchmarks.json')
const memoryData = await response.json()

// Find best memory bandwidth instances
const bestMemory = memoryData.rankings.triad_bandwidth.slice(0, 10)
```

### **Data Analysis & Processing**
```go
// Advanced data aggregation and analysis
package main

import (
    "context"
    "github.com/scttfrdmn/aws-instance-benchmarks/pkg/analysis"
)

func main() {
    // Configure multi-dimensional analysis
    config := analysis.AggregationConfig{
        GroupingDimensions: []string{"instance_family", "region"},
        StatisticalConfig: analysis.StatisticalConfig{
            ConfidenceLevel: 0.95,
            MinSampleSize:   10,
        },
        QualityThreshold: 0.8,
    }
    
    aggregator, _ := analysis.NewDataAggregator(config, dataSource)
    results, _ := aggregator.ProcessBenchmarkData(context.Background())
    
    // Access aggregated performance metrics
    for _, result := range results {
        fmt.Printf("Instance Family: %s\n", result.GroupKey.Dimensions["instance_family"])
        fmt.Printf("STREAM Triad: %.2f GB/s\n", result.PerformanceMetrics.StreamMetrics.TriadBandwidth.Mean)
        fmt.Printf("Quality Score: %.2f\n", result.QualityAssessment.OverallScore)
    }
}
```

### **Integration Examples**
- **ComputeCompass**: Performance-aware instance recommendations
- **Research Tools**: Data-driven instance selection
- **Cost Optimization**: Price/performance analysis
- **Academic Research**: HPC cloud computing studies

## 📈 Comprehensive Testing Coverage

### Instance Type Coverage (27+ types tested)
- **Compute Optimized**: c5.large, c5a.large, c6a.large, c6g.large, c6i.large, c7a.large, c7g.large, c7i.large  
- **General Purpose**: m5.large, m5a.large, m6a.large, m6g.large, m6i.large, m7a.large, m7g.large, m7i.large
- **Memory Optimized**: r5.large, r5a.large, r6a.large, r6g.large, r6i.large, r7a.large, r7g.large, r7i.large
- **Storage Optimized**: i4i.large
- **Burstable**: t3.large, t3a.large

### Architecture Coverage
- **Intel (x86_64)**: c5, c6i, c7i, m5, m6i, m7i, r5, r6i, r7i, i4i, t3
- **AMD (x86_64)**: c5a, c6a, c7a, m5a, m6a, m7a, r5a, r6a, r7a, t3a  
- **AWS Graviton (ARM64)**: c6g, c7g, m6g, m7g, r6g, r7g

### Benchmark Types
- **STREAM**: Memory bandwidth testing across all architectures
- **HPL (LINPACK)**: CPU floating-point performance
- **Statistical Validation**: Multiple iterations with confidence intervals

## ⚙️ AWS Configuration Requirements

### Prerequisites
1. **AWS CLI configured** with appropriate credentials
2. **EC2 permissions** for launching instances, managing security groups, and VPC access
3. **S3 permissions** for storing benchmark results
4. **CloudWatch permissions** for metrics publishing (optional)

### Required AWS Profile Setup
```bash
# Configure AWS profile for benchmarking (recommended: 'aws' profile)
aws configure --profile aws
# Alternatively, use default profile
aws configure
```

### AWS Permissions Required
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:RunInstances",
                "ec2:TerminateInstances",
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceTypes",
                "ec2:DescribeSubnets",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeKeyPairs",
                "s3:GetObject",
                "s3:PutObject",
                "s3:ListBucket",
                "cloudwatch:PutMetricData"
            ],
            "Resource": "*"
        }
    ]
}
```

### Important Configuration Notes
- **Subnet Selection**: Use subnets that support modern instance types (e.g., us-east-1c, not us-east-1e)
- **S3 Bucket**: Configurable via `--s3-bucket` flag, defaults to `aws-instance-benchmarks-data-{region}`
- **Architecture Matching**: ARM64 instances require ARM64-compatible AMIs
- **Availability Zones**: Some newer instance types have limited AZ availability

## 🏗️ Architecture & Components

### Core Packages
- **`pkg/discovery`**: AWS instance type discovery and architecture mapping
- **`pkg/benchmarks`**: STREAM and HPL benchmark execution with statistical validation
- **`pkg/analysis`**: Multi-dimensional data aggregation and performance analysis
- **`pkg/storage`**: S3-based result persistence with compression and organization
- **`pkg/monitoring`**: CloudWatch metrics integration for observability
- **`pkg/aws`**: EC2 orchestration with quota management and spot instance support
- **`pkg/containers`**: Docker container build framework with architecture optimization

### Key Features
- **Statistical Rigor**: Confidence intervals, outlier detection, quality assessment
- **NUMA Awareness**: Multi-socket system optimization and memory affinity
- **Architecture Optimization**: Intel MKL, AMD BLIS, and GCC optimization pipelines  
- **Real-time Monitoring**: CloudWatch integration with custom metrics and alerting
- **Quality Control**: Automated validation with statistical confidence requirements

## 🤝 Contributing

We welcome community contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Benchmark submission guidelines
- Data validation requirements  
- Instance type requests
- Tool improvements

### Running Benchmarks
See our comprehensive [AWS Setup Guide](docs/AWS_SETUP.md) for detailed configuration instructions.

```bash
# Prerequisites: AWS CLI v2 configured with 'aws' profile
aws configure --profile aws

# Build the CLI tool
go build -o aws-benchmark-collector ./cmd

# Run benchmarks with statistical validation (multiple iterations)
./aws-benchmark-collector run \
    --instance-types m7i.large,m7i.xlarge \
    --region us-east-1 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --benchmarks stream,hpl \
    --iterations 5
```

### **New Features in Phase 2**

#### Statistical Validation
- **Multiple iterations** with confidence intervals and quality scoring
- **CloudWatch integration** for real-time monitoring and alerting
- **Advanced error handling** with quota management and capacity recovery
- See [Statistical Validation Guide](docs/STATISTICAL_VALIDATION.md)

#### Community Contributions
- **Automated validation** workflow for community benchmark submissions
- **GitHub Actions integration** with quality assessment and schema validation
- **Contributor recognition** system with structured review process
- See [Community Workflow Guide](docs/COMMUNITY_WORKFLOW.md)

#### Monitoring and Observability
- **CloudWatch metrics** for execution tracking and performance analysis
- **Quality assessment** with coefficient of variation and efficiency scoring
- **Cost tracking** and price-performance analysis
- See [CloudWatch Integration Guide](docs/CLOUDWATCH_INTEGRATION.md)

#### Schema Versioning and Data Quality
- **Semantic versioning** for data schemas with migration support
- **Built-in validation** for all benchmark results and contributions
- **Community quality assurance** with automated validation workflows
- **API compatibility** guarantees for data consumers
- See [Schema Versioning Guide](docs/SCHEMA_VERSIONING.md)

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

The benchmark data is released under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) to encourage broad use in research and commercial applications.

## 🔗 Related Projects

- [ComputeCompass](https://github.com/scttfrdmn/computecompass) - AWS Instance Selector for Research Computing
- [Spack](https://github.com/spack/spack) - Package manager for HPC
- [STREAM](https://www.cs.virginia.edu/stream/) - Memory bandwidth benchmark

## 📞 Contact

- Issues: [GitHub Issues](https://github.com/scttfrdmn/aws-instance-benchmarks/issues)
- Discussions: [GitHub Discussions](https://github.com/scttfrdmn/aws-instance-benchmarks/discussions)
- Email: [benchmarks@computecompass.dev](mailto:benchmarks@computecompass.dev)

---

**Made with ❤️ for the research computing community**