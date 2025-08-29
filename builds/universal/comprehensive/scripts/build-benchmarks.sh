#!/bin/bash

# Build All Benchmarks Script
# Compiles STREAM, LINPACK (HPL), and CoreMark with architecture-specific optimizations

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[BUILD]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Ensure we're in the benchmarks directory
BENCHMARK_DIR="/opt/benchmarks"
cd "$BENCHMARK_DIR" || exit 1

# Use environment variables set by detect-and-optimize.sh
CFLAGS=${CFLAGS:-"-O3 -fopenmp"}
CXXFLAGS=${CXXFLAGS:-"-O3 -fopenmp"}
FCFLAGS=${FCFLAGS:-"-O3 -fopenmp"}

log_info "Building benchmarks with flags: $CFLAGS"

# Build STREAM memory bandwidth benchmark
build_stream() {
    log_info "Building STREAM memory bandwidth benchmark..."
    
    cd stream/
    
    if [[ -f stream.c ]]; then
        log_info "Compiling STREAM with: $CFLAGS"
        gcc $CFLAGS stream.c -o stream_benchmark
        log_success "STREAM benchmark compiled successfully"
    else
        log_error "STREAM source code not found!"
        return 1
    fi
    
    cd ..
}

# Build LINPACK (HPL) CPU performance benchmark
build_linpack() {
    log_info "Building LINPACK (HPL) CPU performance benchmark..."
    
    cd linpack/
    
    if [[ -f linpack.c ]]; then
        log_info "Compiling LINPACK with: $CFLAGS"
        gcc $CFLAGS linpack.c -o linpack_benchmark -lblas -llapack -lm
        log_success "LINPACK benchmark compiled successfully"
    else
        log_error "LINPACK source code not found!"
        return 1
    fi
    
    cd ..
}

# Build CoreMark integer performance benchmark
build_coremark() {
    log_info "Building CoreMark integer performance benchmark..."
    
    cd coremark/
    
    if [[ -f core_main.c && -f coremark.h ]]; then
        log_info "Compiling CoreMark with: $CFLAGS"
        # CoreMark compilation with performance optimizations
        gcc $CFLAGS -DITERATIONS=50000 -DPERFORMANCE_RUN=1 core_main.c -o coremark_benchmark -lm
        log_success "CoreMark benchmark compiled successfully"
    else
        log_error "CoreMark source code not found!"
        return 1
    fi
    
    cd ..
}

# Create unified benchmark information
create_benchmark_info() {
    log_info "Creating benchmark information file..."
    
    cat > benchmark_info.json << EOF
{
  "benchmarks": {
    "stream": {
      "name": "STREAM Memory Bandwidth",
      "version": "5.10",
      "description": "Memory bandwidth measurement with Copy, Scale, Add, and Triad operations",
      "units": "GB/s",
      "executable": "stream/stream_benchmark"
    },
    "linpack": {
      "name": "LINPACK CPU Performance", 
      "version": "HPL",
      "description": "High Performance LINPACK for peak FLOPS measurement",
      "units": "GFLOPS",
      "executable": "linpack/linpack_benchmark"
    },
    "coremark": {
      "name": "CoreMark Integer Performance",
      "version": "1.0",
      "description": "Integer performance and efficiency benchmark",
      "units": "CoreMark Score",
      "executable": "coremark/coremark_benchmark"
    }
  },
  "compilation": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "compiler": "$(gcc --version | head -n1)",
    "cflags": "$CFLAGS",
    "cxxflags": "$CXXFLAGS",
    "fcflags": "$FCFLAGS"
  }
}
EOF

    log_success "Benchmark information file created"
}

# Main build process
main() {
    log_info "=== Building All Benchmarks ==="
    log_info "Compilation flags: $CFLAGS"
    
    # Build each benchmark
    build_stream
    build_linpack
    build_coremark
    
    # Create information file
    create_benchmark_info
    
    log_success "=== All benchmarks built successfully ==="
    log_info "Available benchmarks:"
    log_info "  - STREAM: ./stream/stream_benchmark"
    log_info "  - LINPACK: ./linpack/linpack_benchmark"  
    log_info "  - CoreMark: ./coremark/coremark_benchmark"
}

# Execute main build process
main "$@"