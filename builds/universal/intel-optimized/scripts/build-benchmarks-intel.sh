#!/bin/bash

# Intel oneAPI Optimized Benchmark Build Script
# Compiles STREAM, LINPACK, and CoreMark with Intel compilers and MKL optimization

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INTEL-BUILD]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize Intel oneAPI environment
source /opt/intel/oneapi/setvars.sh

# Ensure we're in the benchmarks directory
BENCHMARK_DIR="/opt/benchmarks"
cd "$BENCHMARK_DIR" || exit 1

# Run Intel optimization detection
/usr/local/bin/detect-and-optimize-intel.sh >/dev/null 2>&1

# Use Intel-optimized flags
CFLAGS=${CFLAGS:-"-O3 -xHost -mkl=parallel -qopenmp"}
CXXFLAGS=${CXXFLAGS:-"-O3 -xHost -mkl=parallel -qopenmp"}
FCFLAGS=${FCFLAGS:-"-O3 -xHost -mkl=parallel -qopenmp"}

log_info "Building benchmarks with Intel oneAPI compilers"
log_info "Compiler flags: $CFLAGS"

# Build STREAM with Intel compiler and MKL
build_stream_intel() {
    log_info "Building STREAM with Intel compiler + MKL optimization..."
    
    cd stream/
    
    if [[ -f stream.c ]]; then
        log_info "Compiling STREAM with: icc $CFLAGS"
        icc $CFLAGS -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
            stream.c -o stream_benchmark -qopenmp
        log_success "Intel-optimized STREAM benchmark compiled successfully"
    else
        log_error "STREAM source code not found!"
        return 1
    fi
    
    cd ..
}

# Build LINPACK with Intel MKL
build_linpack_intel() {
    log_info "Building LINPACK with Intel MKL optimization..."
    
    cd linpack/
    
    if [[ -f linpack.c ]]; then
        log_info "Compiling LINPACK with: icc $CFLAGS + MKL"
        # Use Intel MKL for maximum LINPACK performance
        icc $CFLAGS linpack.c -o linpack_benchmark \
            -mkl=parallel -qopenmp -DMKL_ILP64 \
            -I${MKLROOT}/include
        log_success "Intel MKL-optimized LINPACK benchmark compiled successfully"
    else
        log_error "LINPACK source code not found!"
        return 1
    fi
    
    cd ..
}

# Build CoreMark with Intel compiler
build_coremark_intel() {
    log_info "Building CoreMark with Intel compiler optimization..."
    
    cd coremark/
    
    if [[ -f core_main.c && -f coremark.h ]]; then
        log_info "Compiling CoreMark with: icc $CFLAGS"
        # Intel compiler with aggressive optimization for integer performance
        icc $CFLAGS -DITERATIONS=50000 -DPERFORMANCE_RUN=1 \
            -fp-model fast=2 -fno-alias -unroll-aggressive \
            core_main.c -o coremark_benchmark -lm
        log_success "Intel-optimized CoreMark benchmark compiled successfully"
    else
        log_error "CoreMark source code not found!"
        return 1
    fi
    
    cd ..
}

# Create Intel-optimized benchmark information
create_intel_benchmark_info() {
    log_info "Creating Intel-optimized benchmark information..."
    
    cat > benchmark_info.json << EOF
{
  "benchmarks": {
    "stream": {
      "name": "STREAM Memory Bandwidth (Intel Optimized)",
      "version": "5.10",
      "description": "Memory bandwidth measurement with Intel compiler + threading optimizations",
      "compiler": "Intel oneAPI icc",
      "optimization": "Intel-specific with aggressive vectorization",
      "units": "GB/s",
      "executable": "stream/stream_benchmark"
    },
    "linpack": {
      "name": "LINPACK CPU Performance (Intel MKL)",
      "version": "HPL with MKL",
      "description": "High Performance LINPACK with Intel Math Kernel Library optimization",
      "compiler": "Intel oneAPI icc + MKL",
      "optimization": "Intel MKL parallel BLAS/LAPACK with ILP64",
      "units": "GFLOPS", 
      "executable": "linpack/linpack_benchmark"
    },
    "coremark": {
      "name": "CoreMark Integer Performance (Intel Optimized)",
      "version": "1.0", 
      "description": "Integer performance with Intel compiler aggressive optimization",
      "compiler": "Intel oneAPI icc",
      "optimization": "Intel-specific with fast math and loop unrolling",
      "units": "CoreMark Score",
      "executable": "coremark/coremark_benchmark"
    }
  },
  "compilation": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "compiler": "Intel oneAPI $(icc --version | head -n1)",
    "optimization_profile": "Maximum Intel performance",
    "mkl_version": "$(find /opt/intel/oneapi/mkl -name 'mkl_version.h' -exec grep 'INTEL_MKL_VERSION' {} \; | head -1 | awk '{print $3}' 2>/dev/null || echo 'Unknown')",
    "cflags": "$CFLAGS",
    "cxxflags": "$CXXFLAGS",
    "fcflags": "$FCFLAGS"
  }
}
EOF

    log_success "Intel-optimized benchmark information file created"
}

# Main build process
main() {
    log_info "=== Building Intel oneAPI Optimized Benchmarks ==="
    log_info "Compiler: Intel oneAPI with MKL"
    log_info "Optimization flags: $CFLAGS"
    
    # Build each benchmark with Intel optimization
    build_stream_intel
    build_linpack_intel  
    build_coremark_intel
    
    # Create information file
    create_intel_benchmark_info
    
    log_success "=== All Intel-optimized benchmarks built successfully ==="
    log_info "Performance profile: Maximum Intel CPU performance"
    log_info "Available benchmarks:"
    log_info "  - STREAM (Intel + threading): ./stream/stream_benchmark"
    log_info "  - LINPACK (Intel MKL): ./linpack/linpack_benchmark"
    log_info "  - CoreMark (Intel optimized): ./coremark/coremark_benchmark"
}

# Execute main build process
main "$@"