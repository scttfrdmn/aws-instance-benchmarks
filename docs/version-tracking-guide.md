# Version Tracking Guide for AWS Instance Benchmarks

## Overview

Comprehensive version tracking ensures reproducibility, validation, and scientific rigor in benchmark results. This system captures versions of benchmarks, compilers, libraries, and system components to enable exact result reproduction.

## 📋 **Version Tracking Components**

### **1. Version Manifest Schema v2.0**
Central schema defining all version information captured:
```json
{
  "manifest_version": "2.0",
  "container_info": { "container_version", "build_timestamp", "git_commit" },
  "benchmark_versions": { "stream", "linpack", "coremark" },
  "compiler_versions": { "primary_compiler", "math_libraries", "openmp_version" },
  "system_versions": { "os_info", "glibc_version", "kernel_version" },
  "validation_info": { "source_checksums", "binary_checksums" }
}
```

### **2. Version Collector Script**
Automated collection of version information:
```bash
# Generates comprehensive version manifest during container build
./scripts/version-collector.sh version-manifest.json
```

### **3. Runtime Integration**
Version information included in all benchmark results:
```bash
# Benchmark results now include complete version manifest
docker run aws-benchmark-universal:2.0.0 > results_with_versions.json
```

---

## 🔍 **Tracked Components**

### **Benchmark Versions**

#### **STREAM Memory Bandwidth**
```json
{
  "version": "5.10",
  "source_url": "https://www.cs.virginia.edu/stream/FTP/Code/stream.c",
  "source_hash": "sha256:a1b2c3d4e5f6...",
  "modifications": [
    "Added configurable array size",
    "OpenMP parallelization",
    "Enhanced output formatting"
  ],
  "compilation_flags": "-O3 -fopenmp -DSTREAM_ARRAY_SIZE=80000000",
  "array_size": 80000000,
  "iterations": 10
}
```

#### **LINPACK CPU Performance**
```json
{
  "implementation": "custom",
  "version": "1.0",
  "source_type": "custom",
  "math_library": "system_blas",
  "math_library_version": "3.12.0",
  "compilation_flags": "-O3 -fopenmp -lblas -llapack -lm"
}
```

#### **CoreMark Integer Performance**
```json
{
  "version": "1.0",
  "source_url": "https://github.com/eembc/coremark",
  "source_hash": "sha256:1a2b3c4d5e6f...",
  "implementation_type": "simplified",
  "iterations": 50000,
  "compilation_flags": "-O3 -DITERATIONS=50000 -DPERFORMANCE_RUN=1"
}
```

### **Compiler Versions**

#### **GCC Universal Container**
```json
{
  "primary_compiler": {
    "name": "gcc",
    "version": "13.3.0",
    "version_command_output": "gcc (Ubuntu 13.3.0-6ubuntu2~24.04) 13.3.0",
    "optimization_flags": "-O3 -march=native -mtune=native"
  },
  "math_libraries": [
    {"name": "OpenBLAS", "version": "system", "threading_model": "threaded"}
  ],
  "openmp_version": "4.5"
}
```

#### **Intel oneAPI Optimized Container**
```json
{
  "primary_compiler": {
    "name": "icc",
    "version": "2024.0.2",
    "optimization_flags": "-O3 -xCORE-AVX512 -qopt-zmm-usage=high"
  },
  "math_libraries": [
    {"name": "MKL", "version": "2024.0", "threading_model": "openmp"}
  ]
}
```

#### **AMD AOCC Optimized Container**
```json
{
  "primary_compiler": {
    "name": "clang",
    "version": "17.0.0",
    "optimization_flags": "-O3 -march=znver4 -mtune=znver4"
  },
  "math_libraries": [
    {"name": "AOCL", "version": "4.0", "threading_model": "openmp"}
  ]
}
```

### **System Versions**
```json
{
  "os_info": {
    "name": "Ubuntu",
    "version": "24.04.1",
    "kernel_version": "6.8.0-40-generic",
    "release_info": "Ubuntu 24.04.1 LTS"
  },
  "glibc_version": "2.39",
  "make_version": "4.3"
}
```

---

## 🔐 **Validation & Reproducibility**

### **Source Code Verification**
```json
{
  "checksum_algorithm": "sha256",
  "source_checksums": {
    "stream_c": "a1b2c3d4e5f6789...",
    "linpack_c": "1a2b3c4d5e6f789...",
    "coremark_sources": "9a8b7c6d5e4f321..."
  }
}
```

### **Binary Verification**
```json
{
  "binary_checksums": {
    "stream_benchmark": "f1e2d3c4b5a6987...",
    "linpack_benchmark": "e1f2c3d4a5b6987...",
    "coremark_benchmark": "d1e2f3c4a5b6987..."
  }
}
```

### **Reproducibility Notes**
- All benchmarks compiled with identical flags on target architecture
- Source checksums verified against upstream releases
- Binary checksums provided for validation
- Container built on native hardware for maximum accuracy

---

## 🛠️ **Implementation Details**

### **Container Build Integration**
```dockerfile
# Version collector script copied during build
COPY builds/universal/comprehensive/scripts/version-collector.sh ./scripts/
RUN chmod +x ./scripts/version-collector.sh

# Version manifest generated during build process
RUN bash ./scripts/version-collector.sh version-manifest.json
```

### **Runtime Version Collection**
```bash
# Version information included in benchmark results
run_all_benchmarks() {
    # ... benchmark execution ...
    
    # Load version information
    local version_info=$(load_version_manifest)
    
    # Include in final results
    local final_results=$(jq -n \
        --argjson results "$benchmark_results" \
        --argjson version "$version_info" \
        '$results + $version')
}
```

### **Version Manifest Generation**
```bash
# Manual version manifest generation
./scripts/version-collector.sh my-version-manifest.json

# Validate version manifest
jq empty my-version-manifest.json && echo "Valid JSON" || echo "Invalid JSON"

# View version summary
jq -r '.container_info.container_version + " built on " + .container_info.build_timestamp' my-version-manifest.json
```

---

## 📊 **Use Cases & Benefits**

### **Research Reproducibility**
```bash
# Reproduce exact benchmark results
docker pull aws-benchmark-universal:2.0.0@sha256:abc123...
docker run aws-benchmark-universal:2.0.0 stream

# Verify compiler and library versions match published results
jq '.compiler_versions.primary_compiler' results.json
```

### **Performance Regression Detection**
```bash
# Compare versions between containers
diff <(jq '.benchmark_versions' old-results.json) \
     <(jq '.benchmark_versions' new-results.json)

# Detect compiler version changes
jq -r '.compiler_versions.primary_compiler.version' results.json
```

### **Scientific Publication Support**
- **Complete Version Documentation**: All components tracked for peer review
- **Reproducible Methods**: Exact compiler flags and library versions documented  
- **Validation Data**: Checksums enable independent verification
- **Historical Context**: Git commits link results to specific code versions

### **Cloud Compass Integration**
```json
{
  "performance_context": {
    "compiler_optimization": "Intel oneAPI with MKL",
    "architecture_tuning": "Sapphire Rapids AVX-512",
    "library_versions": "MKL 2024.0 with ILP64"
  }
}
```

---

## 🎯 **Version Tracking Workflow**

### **1. Container Build Phase**
```bash
# Version information captured during build
1. Base OS and kernel versions detected
2. Compiler installations verified  
3. Library versions catalogued
4. Source code checksums calculated
5. Version manifest generated
```

### **2. Benchmark Compilation Phase**
```bash  
# Compilation details tracked
1. Exact compiler flags recorded
2. Library linkage documented
3. Binary checksums calculated
4. Optimization profiles noted
```

### **3. Runtime Execution Phase**
```bash
# Runtime information collected
1. Version manifest loaded
2. System configuration detected
3. Results combined with version data
4. Complete tracking included in output
```

### **4. Result Validation Phase**
```bash
# Validation information available
1. Source code integrity verified
2. Binary checksums validated
3. Compiler version consistency checked
4. Reproducibility confirmed
```

---

## 📈 **Benefits for Different Users**

### **For Researchers**
- **Reproducibility**: Exact version information for paper methods sections
- **Validation**: Independent verification through checksums
- **Comparison**: Fair benchmarking across different systems

### **For Cloud Compass**
- **Context**: Performance results include optimization context
- **Reliability**: Version tracking ensures result consistency  
- **Insights**: Compiler/library performance impact analysis

### **For Community**
- **Transparency**: Open source with complete version tracking
- **Trust**: Cryptographic verification of benchmark integrity
- **Evolution**: Track performance improvements across versions

### **For DevOps/SRE**
- **Regression Detection**: Automated version comparison
- **Deployment Validation**: Confirm expected versions in production
- **Performance Monitoring**: Track performance across version updates

This comprehensive version tracking system ensures that every benchmark result can be independently verified, reproduced, and understood in its complete technical context.