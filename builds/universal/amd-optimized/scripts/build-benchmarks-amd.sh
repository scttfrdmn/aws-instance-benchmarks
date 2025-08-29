#!/bin/bash

# AMD AOCC Optimized Benchmark Build Script
# Compiles STREAM, LINPACK, and CoreMark with AMD AOCC compilers and AOCL optimization

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[AMD-BUILD]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize AMD AOCC + AOCL environment
source /opt/amd/aocc/setenv_AOCC.sh
source /opt/amd/aocl/aocl-setup.sh

# Ensure we're in the benchmarks directory
BENCHMARK_DIR="/opt/benchmarks"
cd "$BENCHMARK_DIR" || exit 1

# Run AMD optimization detection
/usr/local/bin/detect-and-optimize-amd.sh >/dev/null 2>&1

# Use AMD-optimized flags
CFLAGS=${CFLAGS:-"-O3 -march=native -mtune=native -mavx2 -ffast-math -fopenmp"}
CXXFLAGS=${CXXFLAGS:-"-O3 -march=native -mtune=native -mavx2 -ffast-math -fopenmp"}
FCFLAGS=${FCFLAGS:-"-O3 -march=native -mtune=native -mavx2 -ffast-math -fopenmp"}

log_info "Building benchmarks with AMD AOCC compilers"
log_info "Compiler flags: $CFLAGS"

# Build STREAM with AMD AOCC compiler
build_stream_amd() {
    log_info "Building STREAM with AMD AOCC + threading optimization..."
    
    cd stream/
    
    if [[ -f stream.c ]]; then
        log_info "Compiling STREAM with: clang $CFLAGS"
        clang $CFLAGS -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
            stream.c -o stream_benchmark -fopenmp
        log_success "AMD AOCC-optimized STREAM benchmark compiled successfully"
    else
        log_error "STREAM source code not found!"
        return 1
    fi
    
    cd ..
}

# Build LINPACK with AMD AOCL
build_linpack_amd() {
    log_info "Building LINPACK with AMD AOCL optimization..."
    
    cd linpack/
    
    if [[ -f linpack.c ]]; then
        log_info "Compiling LINPACK with: clang $CFLAGS + AOCL"
        # Use AMD AOCL for maximum LINPACK performance on Zen architectures
        clang $CFLAGS linpack.c -o linpack_benchmark \
            -fopenmp -L${AOCL_ROOT}/lib \
            -laocl_blas -laocl_lapack \
            -I${AOCL_ROOT}/include -lm
        log_success "AMD AOCL-optimized LINPACK benchmark compiled successfully"
    else
        log_error "LINPACK source code not found!"
        return 1
    fi
    
    cd ..
}

# Build CoreMark with AMD AOCC compiler
build_coremark_amd() {
    log_info "Building CoreMark with AMD AOCC optimization..."
    
    cd coremark/
    
    if [[ -f core_main.c && -f coremark.h ]]; then
        log_info "Compiling CoreMark with: clang $CFLAGS"
        # AMD AOCC compiler with Zen-specific optimization for integer performance  
        clang $CFLAGS -DITERATIONS=50000 -DPERFORMANCE_RUN=1 \
            -ffast-math -funroll-loops -fno-semantic-interposition \
            core_main.c -o coremark_benchmark -lm
        log_success "AMD AOCC-optimized CoreMark benchmark compiled successfully"
    else
        log_error "CoreMark source code not found!"
        return 1
    fi
    
    cd ..
}

# Create AMD-optimized benchmark information
create_amd_benchmark_info() {
    log_info "Creating AMD-optimized benchmark information..."
    
    cat > benchmark_info.json << EOF
{
  "benchmarks": {
    "stream": {
      "name": "STREAM Memory Bandwidth (AMD Optimized)",
      "version": "5.10",
      "description": "Memory bandwidth measurement with AMD AOCC + Zen-specific optimizations",
      "compiler": "AMD AOCC clang",
      "optimization": "AMD Zen architecture with AVX2 vectorization",
      "units": "GB/s",
      "executable": "stream/stream_benchmark"
    },
    "linpack": {
      "name": "LINPACK CPU Performance (AMD AOCL)",
      "version": "HPL with AOCL",
      "description": "High Performance LINPACK with AMD Optimizing CPU Libraries",
      "compiler": "AMD AOCC clang + AOCL",
      "optimization": "AMD AOCL optimized BLAS/LAPACK for Zen architectures",
      "units": "GFLOPS",
      "executable": "linpack/linpack_benchmark"
    },
    "coremark": {
      "name": "CoreMark Integer Performance (AMD Optimized)",
      "version": "1.0",
      "description": "Integer performance with AMD AOCC Zen-specific optimization",
      "compiler": "AMD AOCC clang",
      "optimization": "AMD-specific with fast math and Zen microarchitecture tuning",
      "units": "CoreMark Score", 
      "executable": "coremark/coremark_benchmark"
    }
  },
  "compilation": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "compiler": "AMD AOCC $(clang --version | head -n1)",
    "optimization_profile": "Maximum AMD Zen performance",
    "aocl_version": "$(find /opt/amd/aocl -name 'aocl_version.h' -exec grep 'AOCL_VERSION' {} \; | head -1 | awk '{print $3}' 2>/dev/null || echo 'Unknown')",
    "cflags": "$CFLAGS",
    "cxxflags": "$CXXFLAGS",
    "fcflags": "$FCFLAGS"
  }
}
EOF

    log_success "AMD-optimized benchmark information file created"
}

# Main build process
main() {
    log_info "=== Building AMD AOCC Optimized Benchmarks ==="
    log_info "Compiler: AMD AOCC with AOCL"
    log_info "Optimization flags: $CFLAGS"
    
    # Build each benchmark with AMD optimization
    build_stream_amd
    build_linpack_amd
    build_coremark_amd
    
    # Create information file
    create_amd_benchmark_info
    
    log_success "=== All AMD-optimized benchmarks built successfully ==="
    log_info "Performance profile: Maximum AMD Zen performance"
    log_info "Available benchmarks:"
    log_info "  - STREAM (AMD AOCC + threading): ./stream/stream_benchmark"
    log_info "  - LINPACK (AMD AOCL): ./linpack/linpack_benchmark"
    log_info "  - CoreMark (AMD optimized): ./coremark/coremark_benchmark"
}

# Execute main build process
main "$@"