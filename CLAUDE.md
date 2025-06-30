# AWS Instance Benchmarks - Development Context

## 🎯 Project Mission

Create an open, community-driven database of comprehensive performance benchmarks for AWS EC2 instances that enables data-driven instance selection for research computing workloads.

## 🏗️ Development Tenets

### **1. Quality First**
- **Test-Driven Development**: Write tests before implementation
- **85%+ Coverage**: Maintain comprehensive test coverage across all components
- **TypeScript Strict**: Use strict type checking for reliability
- **Pre-commit Hooks**: Automated linting, formatting, and testing
- **Comprehensive Documentation**: Every exported function must have detailed documentation with examples for developer onboarding

### **1.1. Data Integrity Rules (CRITICAL)**
- **NO FAKED DATA**: All benchmark results must be from actual execution on real instances
- **NO CHEATING**: Never simulate, mock, or fabricate benchmark outputs
- **NO WORKAROUNDS**: Implement real solutions, not shortcuts that bypass actual benchmarking
- **HONEST IMPLEMENTATION**: Code must accurately represent what it actually does
- **REAL EXECUTION ONLY**: SSH/SSM commands must execute genuine Docker containers with real benchmarks

### **2. Open & Transparent**
- **Open Data Format**: GitHub-hosted JSON with versioned schemas
- **Community Contributions**: Enable easy benchmark submissions
- **Reproducible Results**: Document methodology and validation
- **API-First**: Design for programmatic access

### **3. Research Computing Focus**
- **Architecture Awareness**: Intel, AMD, Graviton, Inferentia, Trainium
- **Workload Optimization**: Memory bandwidth, CPU performance, NUMA topology
- **Cost Efficiency**: Price/performance analysis across all pricing models
- **Academic Rigor**: Statistical validation with confidence intervals

### **4. Modern Engineering Practices**
- **Container-Based**: Reproducible benchmark execution environments
- **Spack Integration**: Architecture-optimized compilation
- **GitHub Actions**: Automated validation and data processing
- **Version Control**: Immutable benchmark data with historical tracking

## 📊 Technical Architecture

### **Data Structure**
```
data/
├── raw/                    # Raw benchmark outputs by date
│   └── 2024-06-26/
├── processed/              # Aggregated, validated data
│   ├── latest/            # Current dataset
│   └── historical/        # Time-series data
└── schemas/               # JSON validation schemas
```

### **Benchmark Categories**
- **Memory Performance**: STREAM benchmarks, cache hierarchy, NUMA topology
- **CPU Performance**: LINPACK, CoreMark, vectorization capabilities
- **Cost Analysis**: Price/performance ratios, spot reliability

### **Quality Standards**
- **JSON Schema Validation**: All data validated against versioned schemas
- **Statistical Rigor**: Multiple runs with confidence intervals
- **Peer Review**: Community validation of benchmark methodologies
- **Automated Checks**: CI/CD pipeline for data quality assurance

## 🔧 Development Environment

### **Required Tools**
- Python 3.9+ with Spack for benchmark compilation
- Docker/Podman for containerized execution
- AWS CLI for instance management
- Git with SSH keys for repository access

### **Testing Strategy**
- **Unit Tests**: Individual benchmark tool validation
- **Integration Tests**: End-to-end benchmark execution
- **Data Validation**: Schema compliance and statistical checks
- **Performance Tests**: Benchmark execution time optimization

### **Code Quality**
- **Linting**: flake8, black, isort for Python
- **Type Checking**: mypy for static analysis
- **Documentation**: Sphinx for API documentation
- **Pre-commit**: Automated quality checks

## 🚀 Integration with ComputeCompass

### **API Design**
- **GitHub Raw URLs**: Direct access to processed JSON data
- **Caching Strategy**: 1-hour cache for production, immediate for development
- **Error Handling**: Graceful degradation when benchmark data unavailable
- **Performance Insights**: Intelligent analysis of benchmark results

### **Data Consumption**
```typescript
// Example integration pattern
const response = await fetch('https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest/memory-benchmarks.json')
const data = await response.json()
```

### **Value Propositions**
- **Performance-Aware Recommendations**: Beyond specs to actual performance
- **Cost Optimization**: Real-world price/performance analysis
- **Research Workload Focus**: Memory-intensive and compute-heavy optimizations

## 📈 Development Roadmap

### **Phase 1: Foundation (Weeks 1-2)**
- [ ] Repository structure with proper schemas
- [ ] Initial benchmark collection tools (Spack configs, containers)
- [ ] Seed data for 10-15 popular instances (m7i, c7g, r7a families)
- [ ] Data validation pipeline

### **Phase 2: Scale (Weeks 3-4)**
- [ ] Automated benchmark collection across instance families
- [ ] Comprehensive dataset (50+ instance types)
- [ ] Community contribution guidelines
- [ ] Historical data tracking

### **Phase 3: Community (Weeks 5-6)**
- [ ] Public launch with documentation
- [ ] Academic partnerships for validation
- [ ] Integration examples and SDKs
- [ ] Performance analysis tools

## 🔬 Benchmark Methodology

### **System-Aware Parameter Scaling**
All benchmarks dynamically scale parameters based on actual system configuration:

#### **STREAM Memory Bandwidth**
- **Array Sizing**: 60% of total system memory divided by 3 arrays
- **Bounds**: Minimum 10M elements, maximum 500M elements  
- **Memory Usage**: Displayed during execution for verification
- **Compiler**: Architecture-optimized GCC with `-O3 -march=native -mtune=native`
- **ARM Optimizations**: `-mcpu=native` for Graviton processors
- **x86 Optimizations**: `-mavx2` for Intel/AMD processors

#### **HPL Matrix Multiplication**
- **Matrix Sizing**: Based on 50% of available memory (N² × 8 bytes)
- **Bounds**: Minimum 500×500, maximum 10000×10000 matrix
- **Operations**: 2×N³ FLOPs for GFLOPS calculation
- **Memory Management**: Dynamic allocation with bounds checking

#### **CoreMark Integer Performance**
- **Iteration Scaling**: Base 1M iterations × CPU cores × frequency factor
- **CPU Detection**: Cores from `nproc`, frequency from `lscpu`
- **Bounds**: Minimum 5M iterations, maximum 100M iterations
- **Workloads**: List processing, matrix operations, state machines

#### **Cache Hierarchy Testing**
- **Cache Detection**: L1/L2/L3 sizes from `lscpu` output
- **Test Sizing**: 50% of each cache level to ensure containment
- **Iteration Scaling**: Inverse relationship (100k for L1, 100 for memory)
- **Access Pattern**: Sequential with stride to measure true latency

### **Statistical Rigor**
- **Multiple Iterations**: 5 iterations minimum for statistical significance
- **Confidence Intervals**: Standard deviation and confidence calculations
- **Outlier Detection**: Automated removal of statistical outliers
- **Reproducibility**: Checksums and verification for result integrity

### **Architecture Optimizations**
- **Intel x86_64**: `-O3 -march=native -mtune=native -mavx2`
- **ARM Graviton**: `-O3 -march=native -mtune=native -mcpu=native`
- **Real Execution**: AWS Systems Manager for genuine hardware testing
- **No Fake Data**: All results from actual benchmark execution

## 📝 Data Governance

### **Schema Evolution**
- **Versioned Schemas**: Backward-compatible data format changes
- **Migration Tools**: Automated data transformation between versions
- **Breaking Changes**: Major version increments with migration guides

### **Quality Assurance**
- **Automated Validation**: GitHub Actions for schema compliance
- **Statistical Checks**: Outlier detection and confidence validation
- **Community Review**: Pull request process for new benchmarks
- **Audit Trail**: Complete history of data changes

## 🔐 Security & Compliance

### **Data Integrity**
- **Checksums**: MD5 and SHA256 for all benchmark files
- **Signed Commits**: GPG signatures for data integrity
- **Immutable History**: Git-based versioning prevents data tampering

### **Privacy**
- **No Personal Data**: Only instance performance metrics
- **Public Domain**: CC BY 4.0 license for broad usage
- **Transparent Process**: Open methodology and validation

## 🤝 Community Guidelines

### **Contribution Process**
1. **Fork Repository**: Standard GitHub workflow
2. **Benchmark Execution**: Follow documented methodology
3. **Data Validation**: Automated schema and statistical checks
4. **Peer Review**: Community validation of results
5. **Integration**: Merge after approval

### **Code of Conduct**
- **Respectful Communication**: Professional and inclusive environment
- **Scientific Rigor**: Evidence-based discussions and decisions
- **Open Collaboration**: Welcome contributions from all skill levels
- **Quality Focus**: Maintain high standards for data and code

## 🎯 Success Metrics

### **Technical Metrics**
- **Data Coverage**: 100+ instance types across all families
- **Update Frequency**: Weekly benchmark data updates
- **API Reliability**: 99.9% uptime for data access
- **Community Engagement**: 10+ active contributors

### **Impact Metrics**
- **Tool Integration**: 5+ tools consuming benchmark data
- **Academic Usage**: Research papers citing the database
- **Community Growth**: 1000+ GitHub stars and forks
- **Industry Adoption**: Enterprise tools leveraging the data

## 🔄 Continuous Improvement

### **Feedback Loops**
- **User Surveys**: Quarterly feedback from tool integrators
- **Performance Analysis**: Benchmark execution optimization
- **Methodology Updates**: Evolving best practices
- **Technology Integration**: New benchmark tools and metrics

### **Evolution Strategy**
- **Backward Compatibility**: Maintain API stability
- **Feature Deprecation**: 6-month notice for breaking changes
- **Community Input**: RFC process for major changes
- **Regular Updates**: Monthly releases with new data

---

**This document should be updated as the project evolves. Last updated: 2025-06-29**