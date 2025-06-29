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

# Run benchmarks with statistical validation (multiple iterations)
./aws-benchmark-collector run \
    --instance-types m7i.large,c7g.large \
    --region us-east-1 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --benchmarks stream,hpl \
    --iterations 5
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

## 📈 Current Coverage

- **Instance Families**: 25+ families (m7i, c7g, r7a, inf2, trn1, etc.)
- **Architectures**: Intel Xeon, AMD EPYC, AWS Graviton, Inferentia, Trainium  
- **Instance Sizes**: nano to 96xlarge across all families
- **Regions**: Multi-region validation for consistency

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