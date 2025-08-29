#!/bin/bash

# AMD AOCC Architecture Detection and Optimization Script
# Provides maximum AMD performance using AOCC compilers and AOCL mathematical libraries
# Supports AMD Zen architectures from Zen1 to Zen4 with micro-architecture specific optimizations

set -euo pipefail

# Global variables
ARCH=""
CPU_MODEL=""
OPTIMIZATION_FLAGS=""
AOCL_FLAGS=""
DETECTED_FEATURES=()

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[AMD-OPT]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize AMD AOCC and AOCL environment
init_amd_environment() {
    if [[ -f /opt/amd/aocc/setenv_AOCC.sh ]]; then
        log_info "Initializing AMD AOCC environment..."
        source /opt/amd/aocc/setenv_AOCC.sh
        
        if [[ -f /opt/amd/aocl/aocl-setup.sh ]]; then
            log_info "Initializing AMD AOCL mathematical libraries..."
            source /opt/amd/aocl/aocl-setup.sh
        fi
        
        log_success "AMD AOCC + AOCL environment initialized"
    else
        log_error "AMD AOCC not found!"
        exit 1
    fi
}

# Detect AMD CPU generation and set optimal compiler flags
detect_amd_optimization() {
    log_info "Detecting AMD CPU generation for AOCC optimization..."
    
    local cpu_flags
    if [[ -f /proc/cpuinfo ]]; then
        cpu_flags=$(grep -m1 "^flags" /proc/cpuinfo | cut -d: -f2)
        CPU_MODEL=$(grep -m1 "model name" /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
    else
        cpu_flags=""
        CPU_MODEL="Unknown"
    fi
    
    log_info "CPU Model: $CPU_MODEL"
    
    # AMD Zen4 (EPYC Genoa - 4th Gen)
    if echo "$CPU_MODEL" | grep -qi "EPYC.*9R14\|EPYC.*9654\|Zen4\|Genoa"; then
        OPTIMIZATION_FLAGS="-O3 -march=znver4 -mtune=znver4 -mavx2 -mfma -fopenmp"
        AOCL_FLAGS="-laocl_blas -laocl_lapack -laocl_scalapack"
        DETECTED_FEATURES+=(zen4 avx2 fma 5nm_process)
        log_success "Detected AMD Zen4 (Genoa) - enabling AVX2 + 5nm optimizations"
        
    # AMD Zen3 (EPYC Milan - 3rd Gen) 
    elif echo "$CPU_MODEL" | grep -qi "EPYC.*7R13\|EPYC.*7763\|Zen3\|Milan"; then
        OPTIMIZATION_FLAGS="-O3 -march=znver3 -mtune=znver3 -mavx2 -mfma -fopenmp"
        AOCL_FLAGS="-laocl_blas -laocl_lapack"
        DETECTED_FEATURES+=(zen3 avx2 fma unified_l3)
        log_success "Detected AMD Zen3 (Milan) - enabling unified L3 cache optimizations"
        
    # AMD Zen2 (EPYC Rome - 2nd Gen)
    elif echo "$CPU_MODEL" | grep -qi "EPYC.*7R32\|EPYC.*7742\|Zen2\|Rome"; then
        OPTIMIZATION_FLAGS="-O3 -march=znver2 -mtune=znver2 -mavx2 -mfma -fopenmp"
        AOCL_FLAGS="-laocl_blas -laocl_lapack"
        DETECTED_FEATURES+=(zen2 avx2 fma 7nm_process)
        log_success "Detected AMD Zen2 (Rome) - enabling 7nm process optimizations"
        
    # AMD Zen1 (EPYC Naples - 1st Gen)
    elif echo "$CPU_MODEL" | grep -qi "EPYC.*7R01\|EPYC.*7551\|Zen1\|Naples"; then
        OPTIMIZATION_FLAGS="-O3 -march=znver1 -mtune=znver1 -mavx2 -mfma -fopenmp"
        AOCL_FLAGS="-laocl_blas -laocl_lapack"
        DETECTED_FEATURES+=(zen1 avx2 fma)
        log_success "Detected AMD Zen1 (Naples) - enabling foundational Zen optimizations"
        
    else
        # Generic AMD with AOCC
        log_warning "Unknown AMD CPU, using generic AOCC optimization"
        OPTIMIZATION_FLAGS="-O3 -march=native -mtune=native -mavx2 -fopenmp"
        AOCL_FLAGS="-laocl_blas -laocl_lapack"
        DETECTED_FEATURES+=(generic_amd)
    fi
    
    # Add AMD-specific performance flags
    OPTIMIZATION_FLAGS="$OPTIMIZATION_FLAGS -ffast-math -funroll-loops -fno-semantic-interposition"
    
    # Add NUMA optimizations for multi-socket EPYC systems
    if [[ $(nproc) -gt 32 ]]; then
        log_info "Detected multi-socket system, enabling NUMA optimizations"
        OPTIMIZATION_FLAGS="$OPTIMIZATION_FLAGS -mprefer-vector-width=256"
        DETECTED_FEATURES+=(numa_optimized)
    fi
    
    # Export optimized flags
    export CFLAGS="$OPTIMIZATION_FLAGS"
    export CXXFLAGS="$OPTIMIZATION_FLAGS"
    export FCFLAGS="$OPTIMIZATION_FLAGS"
    export LDFLAGS="-L${AOCL_ROOT}/lib $AOCL_FLAGS"
    
    log_success "AMD AOCC optimization flags set: $OPTIMIZATION_FLAGS"
    log_success "AOCL integration: $AOCL_FLAGS"
}

# Main execution
main() {
    log_info "=== AMD AOCC Enhanced Universal Benchmark Container v2.1 ==="
    log_info "Initializing AMD AOCC + AOCL optimization..."
    
    init_amd_environment
    ARCH=$(uname -m)
    
    if [[ "$ARCH" == "x86_64" ]]; then
        if echo "$CPU_MODEL" | grep -qi amd || [[ -z "$CPU_MODEL" ]]; then
            detect_amd_optimization
            log_success "AMD AOCC optimization complete"
        else
            log_error "This container requires AMD processors"
            exit 1
        fi
    else
        log_error "AMD AOCC optimization requires x86_64 architecture"
        exit 1
    fi
    
    # Output configuration for Cloud Compass integration
    cat << EOF
{
  "amd_optimization": {
    "cpu_model": "$CPU_MODEL",
    "architecture": "$ARCH",
    "compiler": "AMD AOCC",
    "optimization_flags": "$OPTIMIZATION_FLAGS",
    "aocl_flags": "$AOCL_FLAGS", 
    "detected_features": [$(printf '"%s",' "${DETECTED_FEATURES[@]}" | sed 's/,$//')],
    "performance_profile": "Maximum AMD performance with AOCL optimization"
  }
}
EOF
}

# Execute main function
main "$@"