# Multi-Cloud Instance Benchmarks

An open database of comprehensive performance benchmarks for cloud instances across providers, designed to enable data-driven instance selection for research computing workloads.

## 🌐 Supported Cloud Providers

- **AWS EC2** (Production Ready) - Complete benchmark coverage across 27+ instance families
- **Google Cloud Compute Engine** (Planned) - Architecture designed and ready for implementation  
- **Microsoft Azure Virtual Machines** (Planned) - Provider interface implemented
- **Oracle Cloud Infrastructure** (Planned) - Multi-cloud framework supports extension

## 🎯 Mission

Provide the research computing community with deep, microarchitectural performance data for cloud instances that goes beyond published specifications. Enable tools like [ComputeCompass](https://github.com/scttfrdmn/computecompass) to make intelligent, performance-aware recommendations across cloud providers.

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
- **Real Hardware Execution**: AWS Systems Manager (SSM) command execution on live EC2 instances
- **Embedded Benchmarks**: Self-contained STREAM benchmark compiled with GCC optimizations
- **Architecture-Optimized Compilation**: `-O3 -march=native -mtune=native` for maximum performance
- **Multiple Runs**: Statistical validation with confidence intervals
- **NUMA Awareness**: Proper memory affinity and scaling analysis
- **No Fake Data**: 100% genuine results from actual benchmark execution

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
go build -o cloud-benchmark-collector ./cmd

# Verify installation
./cloud-benchmark-collector --help
```

### **Configuration-Based Usage (Recommended)**
```bash
# 1. Discover and configure AWS infrastructure (one-time setup)
./cloud-benchmark-collector discover infrastructure --region us-west-2 --profile aws

# 2. Run benchmarks using configuration
./cloud-benchmark-collector run \
    --config configs/aws-infrastructure.json \
    --environment us-west-2

# 3. Override specific settings from config
./cloud-benchmark-collector run \
    --config configs/aws-infrastructure.json \
    --environment us-west-2 \
    --instance-types m7i.large,c7g.large \
    --iterations 3
```

### **Manual Configuration (Legacy)**
```bash
# Discover AWS instance types and generate architecture mappings
./cloud-benchmark-collector discover instances --update-containers

# Build optimized benchmark containers
./cloud-benchmark-collector build \
    --architectures intel-icelake,amd-zen4,graviton3 \
    --benchmarks stream

# Run comprehensive benchmarks across multiple instance types (manual config)
./cloud-benchmark-collector run \
    --instance-types m7i.large,m7a.large,m7g.large,c7i.large,c7a.large,c7g.large \
    --region us-west-2 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --s3-bucket my-benchmark-bucket \
    --benchmarks stream,hpl,coremark,cache \
    --iterations 3 \
    --max-concurrency 8 \
    --enable-system-profiling

# Schedule systematic weekly benchmark execution
./aws-benchmark-collector schedule weekly \
    --instance-families m7i,c7g,r7a \
    --region us-east-1 \
    --max-daily-jobs 30 \
    --max-concurrent 5 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --benchmark-rotation \
    --instance-size-waves

# Generate benchmark execution plan without running
./aws-benchmark-collector schedule plan \
    --instance-types m7i.large,c7g.large,r7a.large \
    --benchmarks stream,hpl \
    --output weekly-plan.json

# Process benchmark data into Git-native statistical format
./aws-benchmark-collector process daily \
    --date 2024-06-29 \
    --s3-bucket aws-instance-benchmarks-data-us-east-1 \
    --commit-to-git

# Generate aggregated summaries and indices
./aws-benchmark-collector process aggregate \
    --regenerate-families \
    --regenerate-architectures \
    --regenerate-indices

# Validate data quality and statistical significance
./aws-benchmark-collector process validate \
    --statistical \
    --schema \
    --report validation-report.json

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

## 🔧 Infrastructure Configuration

### **Configuration File System**
The project uses JSON configuration files to eliminate trial-and-error with AWS infrastructure setup:

```json
{
  "environments": {
    "us-west-2": {
      "profile": "aws",
      "region": "us-west-2", 
      "vpc": {"vpc_id": "vpc-e7e2999f", "name": "default"},
      "networking": {
        "subnet_id": "subnet-0528a0d8c3da5acfb",
        "availability_zone": "us-west-2d",
        "security_group_id": "sg-5059b179"
      },
      "compute": {"key_pair_name": "scofri"},
      "storage": {"s3_bucket": "aws-instance-benchmarks-us-west-2-1751232301"}
    }
  },
  "benchmark_defaults": {
    "enable_system_profiling": true,
    "instance_types": ["m7i.large", "c7g.large", "r7a.large"],
    "benchmarks": ["stream"]
  }
}
```

### **Infrastructure Discovery Commands**
```bash
# Discover infrastructure for multiple regions
./cloud-benchmark-collector discover infrastructure --region us-west-2 --profile aws
./cloud-benchmark-collector discover infrastructure --region us-east-1 --profile aws

# View discovered configuration without saving
./cloud-benchmark-collector discover infrastructure --region eu-west-1 --dry-run

# Use custom config file location
./cloud-benchmark-collector discover infrastructure --config custom-config.json
```

### **Configuration Benefits**
- **Zero Trial-and-Error**: Automatically discovers VPC, subnets, security groups, key pairs
- **Multi-Region Support**: Easy switching between AWS regions with region-specific configs
- **Reproducible Builds**: Version-controlled infrastructure configuration
- **Override Flexibility**: CLI flags can override config file values
- **Team Collaboration**: Shared infrastructure configuration across team members

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
- **Microarchitecture Benchmarks**: Architecture-specific performance analysis
  - Intel: AVX-512, MKL optimization, cache hierarchy
  - AMD: Zen4 features, BLIS optimization, vectorization
  - Graviton: Neon SIMD, SVE, ARM-specific optimizations
- **Statistical Validation**: Multiple iterations with confidence intervals
- **System Profiling**: Comprehensive hardware topology discovery
  - CPU microarchitecture, clock speeds, instruction sets
  - Cache hierarchy (L1/L2/L3 sizes, associativity, latencies)
  - NUMA topology and memory controller details
  - Virtualization environment and optimization features

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
                "ssm:SendCommand",
                "ssm:GetCommandInvocation",
                "ssm:DescribeInstanceInformation",
                "ssm:ListCommands",
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

### **EC2 Instance IAM Role Requirements**
For SSM command execution, EC2 instances need an IAM role with the `AmazonSSMManagedInstanceCore` policy attached:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:UpdateInstanceInformation",
                "ssmmessages:CreateControlChannel",
                "ssmmessages:CreateDataChannel",
                "ssmmessages:OpenControlChannel",
                "ssmmessages:OpenDataChannel"
            ],
            "Resource": "*"
        }
    ]
}
```

### **Real Benchmark Execution Details**
The project now implements 100% genuine benchmark execution with no fake data:

- **SSM Command Execution**: Uses AWS Systems Manager to execute commands directly on EC2 instances
- **Embedded STREAM Benchmark**: Self-contained C implementation compiled with GCC optimizations
- **Real Performance Results**: All metrics come from actual hardware execution
- **Architecture-Specific Results**: Genuine performance differences between Intel, AMD, and Graviton processors
- **Command Verification**: All SSM commands and outputs are logged and can be audited

### Important Configuration Notes
- **Subnet Selection**: Use subnets that support modern instance types (e.g., us-east-1c, not us-east-1e)
- **Public IP Assignment**: Instances automatically get public IPs for SSM connectivity
- **S3 Bucket**: Configurable via `--s3-bucket` flag, defaults to `aws-instance-benchmarks-data-{region}`
- **Architecture Matching**: ARM64 instances require ARM64-compatible AMIs
- **Availability Zones**: Some newer instance types have limited AZ availability
- **SSM Agent**: Pre-installed on Amazon Linux 2 AMIs, requires proper IAM role

## 🏗️ Architecture & Components

### Core Packages
- **`pkg/scheduler`**: Batch scheduling system for systematic execution over time
- **`pkg/discovery`**: AWS instance type discovery and architecture mapping
- **`pkg/benchmarks`**: STREAM and HPL benchmark execution with statistical validation
- **`pkg/analysis`**: Multi-dimensional data aggregation and performance analysis
- **`pkg/storage`**: S3-based result persistence with compression and organization
- **`pkg/monitoring`**: CloudWatch metrics integration for observability
- **`pkg/aws`**: EC2 orchestration with quota management and spot instance support
- **`pkg/containers`**: Docker container build framework with architecture optimization

### Key Features
- **Git-Native Data Storage**: Versioned statistical data with complete audit trail
- **GitHub Pages Integration**: Interactive web interface with direct data access
- **Batch Scheduling**: Systematic execution across time windows to avoid quota limits
- **Microarchitecture Analysis**: Deep CPU and memory subsystem performance insights
- **Statistical Rigor**: Confidence intervals, outlier detection, quality assessment
- **NUMA Awareness**: Multi-socket system optimization and memory affinity
- **Architecture Optimization**: Intel MKL, AMD BLIS, ARM SVE optimization pipelines  
- **Real-time Monitoring**: CloudWatch integration with custom metrics and alerting
- **Quality Control**: Automated validation with statistical confidence requirements

## 🤝 Contributing

We welcome community contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Benchmark submission guidelines
- Data validation requirements  
- Instance type requests
- Tool improvements

### Running Benchmarks
See our comprehensive guides for detailed configuration instructions:
- [AWS Setup Guide](docs/AWS_SETUP.md) - AWS configuration and permissions
- [Batch Scheduling Guide](docs/BATCH_SCHEDULING.md) - Systematic execution over time
- [Microarchitecture Benchmarks](docs/MICROARCHITECTURE_BENCHMARKS.md) - Deep performance analysis
- [Data Pipeline](docs/DATA_PIPELINE.md) - GitHub-first data distribution strategy
- [Git-Native Data Storage](docs/GIT_NATIVE_DATA_STORAGE.md) - Versioned statistical data
- [GitHub Pages Integration](docs/GITHUB_PAGES_INTEGRATION.md) - Interactive web interface

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