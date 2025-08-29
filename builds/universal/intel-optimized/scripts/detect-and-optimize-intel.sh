#!/bin/bash

# Intel oneAPI Architecture Detection and Optimization Script
# Provides maximum Intel performance using Intel compilers and Math Kernel Library
# Supports Intel processors from Nehalem to Sapphire Rapids with processor-specific optimizations

set -euo pipefail

# Global variables
ARCH=""
CPU_MODEL=""
OPTIMIZATION_FLAGS=""
MKL_FLAGS=""
DETECTED_FEATURES=()

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INTEL-OPT]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize Intel oneAPI environment
init_intel_environment() {
    if [[ -f /opt/intel/oneapi/setvars.sh ]]; then
        log_info "Initializing Intel oneAPI environment..."
        source /opt/intel/oneapi/setvars.sh
        log_success "Intel oneAPI environment initialized"
    else
        log_error "Intel oneAPI not found!"
        exit 1
    fi
}

# Detect Intel CPU generation and set optimal compiler flags
detect_intel_optimization() {
    log_info "Detecting Intel CPU generation for oneAPI optimization..."
    
    local cpu_flags
    if [[ -f /proc/cpuinfo ]]; then
        cpu_flags=$(grep -m1 "^flags" /proc/cpuinfo | cut -d: -f2)
        CPU_MODEL=$(grep -m1 "model name" /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
    else
        cpu_flags=""
        CPU_MODEL="Unknown"
    fi
    
    log_info "CPU Model: $CPU_MODEL"
    
    # Intel Sapphire Rapids (4th Gen Xeon Scalable)
    if echo "$CPU_MODEL" | grep -qi "Platinum.*8488C\|Xeon.*8488\|Sapphire Rapids"; then
        OPTIMIZATION_FLAGS="-O3 -xCORE-AVX512 -qopt-zmm-usage=high -march=sapphirerapids"
        MKL_FLAGS="-mkl=parallel -qopenmp -DMKL_ILP64"
        DETECTED_FEATURES+=(avx512f avx512vnni avx512bf16 amx)
        log_success "Detected Intel Sapphire Rapids - enabling AVX-512 VNNI/BF16 + AMX"
        
    # Intel Ice Lake (3rd Gen Xeon Scalable)
    elif echo "$CPU_MODEL" | grep -qi "Platinum.*8375C\|Platinum.*8370C\|Ice Lake"; then
        OPTIMIZATION_FLAGS="-O3 -xCORE-AVX512 -march=icelake-server"
        MKL_FLAGS="-mkl=parallel -qopenmp -DMKL_ILP64"
        DETECTED_FEATURES+=(avx512f avx512dq avx512cd avx512vnni)
        log_success "Detected Intel Ice Lake - enabling AVX-512 with VNNI"
        
    # Intel Cascade Lake (2nd Gen Xeon Scalable)
    elif echo "$CPU_MODEL" | grep -qi "Platinum.*8275CL\|Platinum.*8259CL\|Cascade Lake"; then
        OPTIMIZATION_FLAGS="-O3 -xCORE-AVX512 -march=cascadelake"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(avx512f avx512dq avx512cd)
        log_success "Detected Intel Cascade Lake - enabling AVX-512"
        
    # Intel Skylake (1st Gen Xeon Scalable)
    elif echo "$CPU_MODEL" | grep -qi "Platinum.*8175M\|Xeon.*8175\|Skylake"; then
        OPTIMIZATION_FLAGS="-O3 -xCORE-AVX2 -march=skylake"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(avx2 fma)
        log_success "Detected Intel Skylake - enabling AVX2 with FMA"
        
    # Intel Broadwell
    elif echo "$CPU_MODEL" | grep -qi "E5-.*v4\|Broadwell"; then
        OPTIMIZATION_FLAGS="-O3 -xCORE-AVX2 -march=broadwell"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(avx2 fma)
        log_success "Detected Intel Broadwell - enabling AVX2"
        
    # Intel Haswell 
    elif echo "$CPU_MODEL" | grep -qi "E5-.*v3\|Haswell"; then
        OPTIMIZATION_FLAGS="-O3 -xCORE-AVX2 -march=haswell"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(avx2 fma)
        log_success "Detected Intel Haswell - enabling AVX2 with FMA"
        
    # Intel Ivy Bridge
    elif echo "$CPU_MODEL" | grep -qi "E5-.*v2\|Ivy Bridge"; then
        OPTIMIZATION_FLAGS="-O3 -xAVX -march=ivybridge"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(avx)
        log_success "Detected Intel Ivy Bridge - enabling AVX"
        
    # Intel Sandy Bridge
    elif echo "$CPU_MODEL" | grep -qi "E5-.*v1\|Sandy Bridge"; then
        OPTIMIZATION_FLAGS="-O3 -xAVX -march=sandybridge"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(avx)
        log_success "Detected Intel Sandy Bridge - enabling AVX"
        
    # Intel Nehalem/Westmere (original EC2)
    elif echo "$CPU_MODEL" | grep -qi "E5520\|L5520\|Nehalem\|Westmere"; then
        OPTIMIZATION_FLAGS="-O3 -xSSE4.2 -march=nehalem"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(sse4_2)
        log_success "Detected Intel Nehalem/Westmere - enabling SSE4.2"
        
    else
        # Generic Intel with oneAPI
        log_warning "Unknown Intel CPU, using generic oneAPI optimization"
        OPTIMIZATION_FLAGS="-O3 -xHost"
        MKL_FLAGS="-mkl=parallel -qopenmp"
        DETECTED_FEATURES+=(generic_intel)
    fi
    
    # Add Intel-specific performance flags
    OPTIMIZATION_FLAGS="$OPTIMIZATION_FLAGS -fp-model fast=2 -fno-alias -unroll-aggressive"
    
    # Export optimized flags
    export CFLAGS="$OPTIMIZATION_FLAGS $MKL_FLAGS"
    export CXXFLAGS="$OPTIMIZATION_FLAGS $MKL_FLAGS"
    export FCFLAGS="$OPTIMIZATION_FLAGS $MKL_FLAGS"
    export LDFLAGS="-L${MKLROOT}/lib/intel64 -lmkl_intel_ilp64 -lmkl_intel_thread -lmkl_core -liomp5"
    
    log_success "Intel oneAPI optimization flags set: $OPTIMIZATION_FLAGS"
    log_success "MKL integration: $MKL_FLAGS"
}

# Main execution
main() {
    log_info "=== Intel oneAPI Enhanced Universal Benchmark Container v2.1 ==="
    log_info "Initializing Intel oneAPI optimization..."
    
    init_intel_environment
    ARCH=$(uname -m)
    
    if [[ "$ARCH" == "x86_64" ]]; then
        if echo "$CPU_MODEL" | grep -qi intel || [[ -z "$CPU_MODEL" ]]; then
            detect_intel_optimization
            log_success "Intel oneAPI optimization complete"
        else
            log_error "This container requires Intel processors"
            exit 1
        fi
    else
        log_error "Intel oneAPI optimization requires x86_64 architecture"
        exit 1
    fi
    
    # Output configuration for Cloud Compass integration
    cat << EOF
{
  "intel_optimization": {
    "cpu_model": "$CPU_MODEL",
    "architecture": "$ARCH", 
    "compiler": "Intel oneAPI",
    "optimization_flags": "$OPTIMIZATION_FLAGS",
    "mkl_flags": "$MKL_FLAGS",
    "detected_features": [$(printf '"%s",' "${DETECTED_FEATURES[@]}" | sed 's/,$//')],
    "performance_profile": "Maximum Intel performance with MKL optimization"
  }
}
EOF
}

# Execute main function
main "$@"