# Progressive Enhancement Roadmap for AWS Instance Benchmarks

## Overview

The progressive enhancement strategy addresses the container explosion problem (105+ containers) by implementing three phases that build upon each other, providing increasing levels of performance optimization while maintaining simplicity and broad compatibility.

## 🎯 **Phase 1: Enhanced Universal Container** ✅ **COMPLETE**

### **Goal**: Single container that works across all architectures with runtime optimization

### **Implementation Status**: ✅ Completed
- **Container**: `aws-benchmark-universal:2.0.0`
- **Base**: Ubuntu 24.04 + GCC 13
- **Runtime Detection**: Intel, AMD, ARM processor identification
- **Benchmarks**: STREAM, LINPACK, CoreMark with OpenMP
- **JSON Schema**: v2.0 for Cloud Compass integration

### **Key Achievements**:
- ✅ Eliminates 105+ container matrix complexity
- ✅ Runtime architecture-specific compiler flags
- ✅ Industry standard benchmarks (no custom implementations)
- ✅ Cloud Compass ready with performance ratings
- ✅ Proven working on ARM64 (Graviton) architecture

### **Technical Components**:
```bash
# Runtime optimization examples
Intel Sapphire Rapids: -march=sapphirerapids -mavx512vnni -mavx512bf16
AMD Zen4: -march=znver4 -mavx2 -mfma  
ARM Graviton4: -mcpu=neoverse-v2 -mtune=neoverse-v2
```

---

## 🚀 **Phase 2: Selective Vendor Compiler Optimization** ✅ **DESIGNED**

### **Goal**: Maximum performance for users requiring absolute peak performance

### **Implementation Status**: 🔄 Ready for Implementation
- **Intel Optimized**: `aws-benchmark-universal:2.1.0-intel`
- **AMD Optimized**: `aws-benchmark-universal:2.1.0-amd`
- **Native Building**: AWS SSM-based remote container building

### **Container Variants**:

#### **Universal (Default Recommendation)**
```bash
# Broad compatibility, good performance
docker run aws-benchmark-universal:2.0.0
```

#### **Intel oneAPI Optimized**
```bash
# Maximum Intel performance with MKL
docker run aws-benchmark-universal:2.1.0-intel
```
- **Compiler**: Intel oneAPI icc/icpc/ifort
- **Math Library**: Intel MKL with ILP64 support
- **Optimization**: `-xCORE-AVX512 -qopt-zmm-usage=high` (Sapphire Rapids)
- **Performance Gain**: ~15-30% on Intel processors
- **Use Case**: HPC workloads requiring maximum Intel performance

#### **AMD AOCC Optimized**
```bash
# Maximum AMD performance with AOCL
docker run aws-benchmark-universal:2.1.0-amd
```
- **Compiler**: AMD AOCC clang/clang++/flang
- **Math Library**: AMD AOCL (BLAS/LAPACK/ScaLAPACK)
- **Optimization**: `-march=znver4 -mtune=znver4 -mprefer-vector-width=256`
- **Performance Gain**: ~10-25% on AMD Zen processors
- **Use Case**: AMD-specific optimization requirements

### **Native AWS Instance Building**:
```bash
# Build containers on target hardware for maximum accuracy
./scripts/native-container-build.sh build intel-sapphirerapids intel-optimized
./scripts/native-container-build.sh build amd-zen4 amd-optimized
./scripts/native-container-build.sh build graviton4 universal
./scripts/native-container-build.sh build-all  # Build all combinations
```

### **Progressive Selection Strategy**:
1. **Start with Universal**: Works everywhere, good performance
2. **Upgrade to Vendor-Optimized**: When maximum performance needed
3. **Native Building**: For research-grade accuracy validation

---

## 📦 **Phase 3: Multi-OS Support** 🔄 **IN DESIGN**

### **Goal**: Support for enterprise Linux distributions commonly used in HPC

### **Implementation Status**: 📋 Design Phase
- **Amazon Linux 2023**: AWS-native optimization
- **Rocky Linux 9.6**: Enterprise HPC compatibility  
- **Ubuntu 24.04**: Current implementation (reference)

### **Container Matrix** (Simplified):
```
Base OS × Architecture × Compiler = Containers
  3     ×      15       ×    3     = 135 potential containers
  
BUT: Progressive approach = 3 OS × 3 variants = 9 core containers
```

#### **Ubuntu 24.04** (Current Reference)
- **Status**: ✅ Complete
- **GCC Version**: 13.3.0
- **Features**: Latest toolchain, modern kernel
- **Use Case**: Development, testing, modern workloads

#### **Amazon Linux 2023**
```dockerfile
FROM amazonlinux:2023
# AWS-optimized kernel and libraries
# Enhanced networking and storage performance
# Graviton-specific optimizations
```
- **Status**: 📋 Planned
- **Features**: AWS-optimized kernel, enhanced EC2 integration
- **Performance**: ~5-10% better on AWS infrastructure
- **Use Case**: Production AWS workloads

#### **Rocky Linux 9.6** 
```dockerfile
FROM rockylinux:9.6
# Enterprise HPC compatibility
# RHEL ecosystem libraries
# Long-term support lifecycle
```
- **Status**: 📋 Planned  
- **Features**: Enterprise stability, HPC library compatibility
- **Use Case**: Traditional HPC migrations, enterprise environments

### **OS-Specific Optimizations**:

| **Operating System** | **Kernel** | **GCC** | **Special Features** | **Performance Profile** |
|---------------------|-----------|---------|---------------------|----------------------|
| **Ubuntu 24.04** | 6.8+ | 13.3.0 | Latest toolchain, modern drivers | Cutting-edge performance |
| **Amazon Linux 2023** | 6.1 (AWS) | 11.4.1 | AWS-optimized, enhanced networking | AWS-native optimization |
| **Rocky Linux 9.6** | 5.14 (RHEL) | 11.4.1 | Enterprise stability, HPC libs | HPC ecosystem compatibility |

---

## 🎛️ **Container Selection Guide**

### **For Cloud Compass Users**:
```bash
# Recommended progression
1. Start: docker run aws-benchmark-universal:2.0.0
2. Intel: docker run aws-benchmark-universal:2.1.0-intel  
3. AMD:   docker run aws-benchmark-universal:2.1.0-amd
```

### **For HPC Research**:
```bash  
# Maximum accuracy with native building
./scripts/native-container-build.sh build intel-sapphirerapids intel-optimized
./scripts/native-container-build.sh build amd-zen4 amd-optimized
```

### **For Enterprise**:
```bash
# Stable enterprise OS with vendor optimization
docker run aws-benchmark-universal:2.1.0-intel-rockylinux
docker run aws-benchmark-universal:2.1.0-amd-rockylinux
```

---

## 📊 **Expected Performance Improvements**

### **Phase 1 → Phase 2 Gains**:
- **Intel with MKL**: 15-30% improvement in LINPACK
- **AMD with AOCL**: 10-25% improvement in mathematical operations
- **Architecture-specific**: 5-15% general performance uplift

### **Phase 2 → Phase 3 Gains**:
- **Amazon Linux**: 5-10% on AWS infrastructure
- **Rocky Linux**: Better HPC library compatibility
- **OS Optimization**: Kernel-specific performance tuning

### **Validation Strategy**:
```bash
# Compare performance across all variants
docker run aws-benchmark-universal:2.0.0 stream          # Baseline
docker run aws-benchmark-universal:2.1.0-intel stream    # Intel optimized  
docker run aws-benchmark-universal:2.1.0-amd stream      # AMD optimized

# Measure percentage improvements for each architecture
```

---

## 🛠️ **Implementation Timeline**

### **Phase 1**: ✅ Complete (August 2025)
- Universal container with runtime detection
- Industry standard benchmarks integration
- Cloud Compass JSON schema v2.0

### **Phase 2**: 🔄 Ready for Implementation (September 2025)
- Intel oneAPI + MKL integration
- AMD AOCC + AOCL integration  
- Native AWS instance building workflow

### **Phase 3**: 📋 Design Complete (October 2025)
- Amazon Linux 2023 containers
- Rocky Linux 9.6 containers
- Multi-OS performance validation

---

## 🎯 **Success Metrics**

### **Technical Metrics**:
- **Container Reduction**: 105+ → 9 core containers (91% reduction)
- **Performance Coverage**: 15+ processor generations (2009-2025)
- **OS Coverage**: 3 enterprise Linux distributions
- **Build Time**: Native building on target hardware

### **User Experience Metrics**:
- **Simplicity**: Single command execution across all architectures
- **Performance**: Maximum optimization when needed
- **Compatibility**: Broad ecosystem support
- **Accuracy**: Native builds eliminate cross-compilation artifacts

### **Community Adoption**:
- **Pull-ready containers**: `docker pull aws-benchmark-universal:variant`
- **Validation capability**: Users can verify results on their hardware
- **Research integration**: Fair architecture comparisons
- **Open source**: Complete transparency and reproducibility

---

## 🔮 **Future Considerations**

### **GPU Support** (Phase 4):
- NVIDIA H100/A100 benchmarking
- AMD Instinct MI300 support
- AWS Inferentia/Trainium integration

### **Specialized Workloads** (Phase 5):
- AI/ML specific benchmarks
- Database performance testing  
- Network-intensive workload analysis

### **Advanced Features** (Phase 6):
- Real-time performance monitoring
- Automated performance regression detection
- Cost-performance optimization recommendations

---

This progressive enhancement approach ensures that users can start simple and enhance performance as needed, while maintainers avoid the complexity of managing 105+ containers. Each phase builds on the previous one, providing clear upgrade paths and measurable performance improvements.