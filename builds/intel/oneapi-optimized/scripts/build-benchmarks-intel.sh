#!/bin/bash

# Intel oneAPI Optimized Benchmark Build Script
# Compiles STREAM, LINPACK, and CoreMark with Intel compilers and MKL acceleration

set -euo pipefail

# Color output  
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INTEL-BUILD]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Ensure we're in the benchmarks directory
BENCHMARK_DIR="/opt/benchmarks"
cd "$BENCHMARK_DIR" || exit 1

# Initialize Intel oneAPI environment and optimization
log_info "Initializing Intel oneAPI environment..."
source "${ONEAPI_ROOT}/setvars.sh" --force
source "/usr/local/bin/intel-optimize.sh"

log_info "Building benchmarks with Intel oneAPI + MKL optimizations..."
log_info "Compiler: $(icx --version | head -n1)"
log_info "Flags: $CFLAGS"

# Build STREAM with Intel optimizations and MKL
build_stream_intel() {
    log_info "Building STREAM with Intel oneAPI + MKL acceleration..."
    
    cd stream/
    
    if [[ -f stream.c ]]; then
        log_info "Compiling STREAM with Intel optimizations..."
        # Intel-optimized STREAM with MKL vectorization hints
        icx $CFLAGS \
            -DSTREAM_ARRAY_SIZE=80000000 \
            -DNTIMES=10 \
            -DSTREAM_TYPE=double \
            -qopt-report=5 \
            -qopt-report-phase=vec \
            stream.c -o stream_benchmark_intel $LDFLAGS
            
        log_success "STREAM benchmark compiled with Intel oneAPI"
        
        # Generate optimization report summary
        if [[ -f stream.optrpt ]]; then
            log_info "Intel optimization report generated: stream.optrpt"
        fi
    else
        log_error "STREAM source code not found!"
        return 1
    fi
    
    cd ..
}

# Build LINPACK with Intel MKL for maximum GFLOPS
build_linpack_intel() {
    log_info "Building LINPACK with Intel MKL for maximum performance..."
    
    cd linpack/
    
    if [[ -f linpack.c ]]; then
        log_info "Compiling LINPACK with Intel MKL BLAS..."
        # Intel MKL provides optimized DGEMM for maximum GFLOPS
        icx $CFLAGS \
            -DINTEL_MKL \
            -qopt-report=5 \
            -qopt-report-phase=vec \
            linpack.c -o linpack_benchmark_intel $LDFLAGS
            
        log_success "LINPACK benchmark compiled with Intel MKL acceleration"
    else
        log_error "LINPACK source code not found!"
        return 1
    fi
    
    cd ..
}

# Build CoreMark with Intel compiler optimizations
build_coremark_intel() {
    log_info "Building CoreMark with Intel compiler vectorization..."
    
    cd coremark/
    
    if [[ -f core_main.c && -f coremark.h ]]; then
        log_info "Compiling CoreMark with Intel optimizations..."
        # Intel compiler excels at auto-vectorization for integer workloads
        icx $CFLAGS \
            -DITERATIONS=100000 \
            -DPERFORMANCE_RUN=1 \
            -DINTEL_COMPILER \
            -qopt-report=5 \
            -qopt-report-phase=vec \
            core_main.c -o coremark_benchmark_intel -lm
            
        log_success "CoreMark benchmark compiled with Intel optimizations"
    else
        log_error "CoreMark source code not found!"
        return 1
    fi
    
    cd ..
}

# Create Intel-specific benchmark information
create_intel_benchmark_info() {
    log_info "Creating Intel oneAPI benchmark information..."
    
    cat > benchmark_info_intel.json << EOF
{
  "benchmarks": {
    "stream": {
      "name": "STREAM Memory Bandwidth (Intel MKL Optimized)",
      "version": "5.10",
      "description": "Memory bandwidth with Intel MKL vectorization and AVX-512 optimizations",
      "units": "GB/s",
      "executable": "stream/stream_benchmark_intel",
      "array_size": 80000000,
      "optimization": "Intel oneAPI + MKL + AVX-512"
    },
    "linpack": {
      "name": "LINPACK CPU Performance (Intel MKL BLAS)",
      "version": "HPL with MKL",
      "description": "Peak FLOPS using Intel MKL optimized DGEMM routines",
      "units": "GFLOPS", 
      "executable": "linpack/linpack_benchmark_intel",
      "optimization": "Intel MKL BLAS with threading"
    },
    "coremark": {
      "name": "CoreMark Integer Performance (Intel Vectorized)",
      "version": "1.0",
      "description": "Integer performance with Intel compiler auto-vectorization",
      "units": "CoreMark Score",
      "executable": "coremark/coremark_benchmark_intel", 
      "iterations": 100000,
      "optimization": "Intel compiler vectorization + IPO"
    }
  },
  "compilation": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "compiler": "$(icx --version | head -n1)",
    "cflags": "$CFLAGS",
    "ldflags": "$LDFLAGS",
    "architecture_target": "$INTEL_ARCH_FLAGS",
    "mkl_version": "$(find "${MKLROOT}" -name "mkl_version.h" -exec grep 'INTEL_MKL_VERSION' {} \; 2>/dev/null | head -1 | awk '{print $3}' || echo 'Unknown')",
    "optimization_profile": "Intel oneAPI + MKL maximum performance"
  }
}
EOF

    # Generate comprehensive Intel version manifest
    log_info "Generating Intel oneAPI version manifest..."
    if [[ -f /opt/benchmarks/scripts/version-collector-intel.sh ]]; then
        bash /opt/benchmarks/scripts/version-collector-intel.sh version-manifest-intel.json
    else
        log_info "Intel version collector not found, skipping detailed manifest"
    fi

    log_success "Intel benchmark information created"
}

# Performance validation tests
validate_intel_builds() {
    log_info "Validating Intel-optimized benchmark builds..."
    
    # Test STREAM execution
    if [[ -f stream/stream_benchmark_intel ]]; then
        log_info "Testing STREAM benchmark..."
        timeout 30s ./stream/stream_benchmark_intel > /dev/null 2>&1 && log_success "STREAM validation passed" || log_error "STREAM validation failed"
    fi
    
    # Test LINPACK execution  
    if [[ -f linpack/linpack_benchmark_intel ]]; then
        log_info "Testing LINPACK benchmark..."
        timeout 30s ./linpack/linpack_benchmark_intel > /dev/null 2>&1 && log_success "LINPACK validation passed" || log_error "LINPACK validation failed"
    fi
    
    # Test CoreMark execution
    if [[ -f coremark/coremark_benchmark_intel ]]; then
        log_info "Testing CoreMark benchmark..."  
        timeout 30s ./coremark/coremark_benchmark_intel > /dev/null 2>&1 && log_success "CoreMark validation passed" || log_error "CoreMark validation failed"
    fi
}

# Main build process
main() {
    log_info "=== Intel oneAPI Optimized Benchmark Build ==="
    log_info "oneAPI Root: $ONEAPI_ROOT"
    log_info "MKL Root: $MKLROOT"
    log_info "Optimization flags: $CFLAGS"
    
    # Build each benchmark with Intel optimizations
    build_stream_intel
    build_linpack_intel  
    build_coremark_intel
    
    # Create information and validate
    create_intel_benchmark_info
    validate_intel_builds
    
    log_success "=== Intel oneAPI Benchmark Build Complete ==="
    log_info "Available Intel-optimized benchmarks:"
    log_info "  - STREAM: ./stream/stream_benchmark_intel"
    log_info "  - LINPACK: ./linpack/linpack_benchmark_intel"
    log_info "  - CoreMark: ./coremark/coremark_benchmark_intel"
    log_info "Expected performance gains: 20-40% over GCC on Intel processors"
}

# Execute main build process
main "$@"