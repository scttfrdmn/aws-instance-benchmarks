#!/bin/bash

# Intel oneAPI Architecture Detection and Optimization Script
# Detects specific Intel microarchitectures and applies optimal compiler flags

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INTEL-OPT]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize Intel oneAPI environment
source_oneapi() {
    if [[ -f "${ONEAPI_ROOT}/setvars.sh" ]]; then
        log_info "Sourcing Intel oneAPI environment..."
        source "${ONEAPI_ROOT}/setvars.sh" --force
        log_success "Intel oneAPI environment loaded"
    else
        log_error "Intel oneAPI environment not found"
        return 1
    fi
}

# Detect Intel CPU architecture and apply optimal flags
detect_intel_architecture() {
    local cpu_model=$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
    local cpu_family=$(grep -m1 'cpu family' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
    local cpu_model_num=$(grep -m1 'model' /proc/cpuinfo | grep -v 'model name' | cut -d: -f2 | sed 's/^ *//')
    
    log_info "Detected CPU: $cpu_model"
    log_info "CPU Family: $cpu_family, Model: $cpu_model_num"
    
    # Intel-specific microarchitecture detection and optimization
    local intel_flags=""
    local mkl_flags=""
    
    # Sapphire Rapids (4th Gen Xeon Scalable - C7i instances)
    if echo "$cpu_model" | grep -qi "Platinum.*8488C\|8470C\|8280L"; then
        intel_flags="-xSAPPHIRERAPIDS -mavx512vnni -mavx512bf16 -mavx512vp2intersect"
        mkl_flags="-DMKL_ENABLE_AVX512"
        log_success "Detected Sapphire Rapids - applying maximum AVX-512 optimizations"
        
    # Ice Lake (3rd Gen Xeon Scalable - M6i, C6i instances)  
    elif echo "$cpu_model" | grep -qi "Platinum.*8375C\|8370C\|Gold.*6354"; then
        intel_flags="-xICELAKE-SERVER -mavx512f -mavx512cd -mavx512bw -mavx512dq -mavx512vl"
        mkl_flags="-DMKL_ENABLE_AVX512"
        log_success "Detected Ice Lake - applying AVX-512 optimizations"
        
    # Cascade Lake (2nd Gen Xeon Scalable - M5, C5 instances)
    elif echo "$cpu_model" | grep -qi "Platinum.*8259CL\|8175M\|Gold.*6142"; then
        intel_flags="-xCASCADELAKE -mavx512f -mavx512cd -mavx512bw -mavx512dq -mavx512vl"
        mkl_flags="-DMKL_ENABLE_AVX512"
        log_success "Detected Cascade Lake - applying AVX-512 optimizations"
        
    # Skylake (1st Gen Xeon Scalable)
    elif echo "$cpu_model" | grep -qi "Platinum.*8124M\|Gold.*5120"; then
        intel_flags="-xSKYLAKE-AVX512 -mavx512f -mavx512cd -mavx512bw -mavx512dq -mavx512vl"
        mkl_flags="-DMKL_ENABLE_AVX512"
        log_success "Detected Skylake - applying AVX-512 optimizations"
        
    # Broadwell (older instances)
    elif echo "$cpu_model" | grep -qi "E5-2676\|E5-2686"; then
        intel_flags="-xBROADWELL -mavx2 -mfma"
        mkl_flags="-DMKL_ENABLE_AVX2"
        log_success "Detected Broadwell - applying AVX2 optimizations"
        
    # Generic Intel x86_64 fallback
    else
        intel_flags="-xHost -march=native -mtune=native"
        mkl_flags="-DMKL_ENABLE_AVX2"
        log_info "Using generic Intel optimizations with host detection"
    fi
    
    # Export optimized flags
    export INTEL_ARCH_FLAGS="$intel_flags"
    export MKL_COMPILE_FLAGS="$mkl_flags"
    
    # Enhanced Intel compiler flags for maximum performance
    export CFLAGS="-O3 $intel_flags -ipo -qopenmp -fp-model fast=2 -ffast-math -funroll-loops $mkl_flags"
    export CXXFLAGS="-O3 $intel_flags -ipo -qopenmp -fp-model fast=2 -ffast-math -funroll-loops $mkl_flags"
    export FCFLAGS="-O3 $intel_flags -ipo -qopenmp -fp-model fast=2 -ffast-math -funroll-loops"
    
    # Intel MKL linking with threading optimizations
    export LDFLAGS="-L${MKLROOT}/lib/intel64 -Wl,--start-group -lmkl_intel_lp64 -lmkl_intel_thread -lmkl_core -Wl,--end-group -liomp5 -lpthread -lm -ldl"
    
    log_success "Intel compiler flags configured: $CFLAGS"
    log_success "MKL linking configured: $LDFLAGS"
}

# Configure Intel threading for optimal performance
configure_intel_threading() {
    local cpu_cores=$(nproc)
    
    # Intel-specific threading configuration
    export OMP_NUM_THREADS="$cpu_cores"
    export MKL_NUM_THREADS="$cpu_cores"
    export MKL_DYNAMIC="FALSE"
    export KMP_AFFINITY="granularity=fine,compact,1,0"
    export KMP_BLOCKTIME="0"
    export KMP_SETTINGS="1"
    
    log_info "Configured Intel threading: $cpu_cores threads with compact affinity"
}

# Validate Intel oneAPI installation
validate_intel_environment() {
    log_info "Validating Intel oneAPI environment..."
    
    # Check compilers
    if command -v icx >/dev/null 2>&1; then
        local icx_version=$(icx --version | head -n1)
        log_success "Intel C Compiler: $icx_version"
    else
        log_error "Intel C Compiler (icx) not found"
        return 1
    fi
    
    if command -v icpx >/dev/null 2>&1; then
        local icpx_version=$(icpx --version | head -n1)
        log_success "Intel C++ Compiler: $icpx_version"
    else
        log_error "Intel C++ Compiler (icpx) not found"  
        return 1
    fi
    
    # Check MKL
    if [[ -d "${MKLROOT}" ]]; then
        local mkl_version=$(find "${MKLROOT}" -name "mkl_version.h" -exec grep 'INTEL_MKL_VERSION' {} \; 2>/dev/null | head -1 | awk '{print $3}' || echo 'Unknown')
        log_success "Intel MKL: Version $mkl_version at $MKLROOT"
    else
        log_error "Intel MKL not found at $MKLROOT"
        return 1
    fi
    
    log_success "Intel oneAPI environment validation complete"
}

# Main execution
main() {
    log_info "=== Intel oneAPI Architecture Optimization ==="
    
    source_oneapi
    detect_intel_architecture  
    configure_intel_threading
    validate_intel_environment
    
    log_success "=== Intel Optimization Complete ==="
    log_info "Ready for benchmark compilation with Intel oneAPI + MKL"
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi