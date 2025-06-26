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

### Using the Data
```javascript
// Fetch latest benchmark data
const response = await fetch('https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest/memory-benchmarks.json')
const memoryData = await response.json()

// Find best memory bandwidth instances
const bestMemory = memoryData.rankings.triad_bandwidth.slice(0, 10)
```

### Integration Examples
- **ComputeCompass**: Performance-aware instance recommendations
- **Research Tools**: Data-driven instance selection
- **Cost Optimization**: Price/performance analysis
- **Academic Research**: HPC cloud computing studies

## 📈 Current Coverage

- **Instance Families**: 25+ families (m7i, c7g, r7a, inf2, trn1, etc.)
- **Architectures**: Intel Xeon, AMD EPYC, AWS Graviton, Inferentia, Trainium  
- **Instance Sizes**: nano to 96xlarge across all families
- **Regions**: Multi-region validation for consistency

## 🤝 Contributing

We welcome community contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Benchmark submission guidelines
- Data validation requirements  
- Instance type requests
- Tool improvements

### Running Benchmarks
```bash
# Clone the repository
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git

# Deploy benchmark suite
cd tools/deployment
./deploy-benchmarks.sh --family m7i --sizes "large,xlarge,2xlarge"
```

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