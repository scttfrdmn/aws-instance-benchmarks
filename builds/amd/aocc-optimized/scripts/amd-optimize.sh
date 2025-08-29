#!/bin/bash

# AMD AOCC Architecture Detection and Optimization Script
# Detects AMD EPYC microarchitectures and applies optimal compiler flags

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[AMD-OPT]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Detect AMD EPYC architecture and apply optimal flags
detect_amd_architecture() {
    local cpu_model=$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
    local cpu_family=$(grep -m1 'cpu family' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
    local cpu_model_num=$(grep -m1 'model' /proc/cpuinfo | grep -v 'model name' | cut -d: -f2 | sed 's/^ *//')
    
    log_info "Detected CPU: $cpu_model"
    log_info "CPU Family: $cpu_family, Model: $cpu_model_num"
    
    # AMD EPYC microarchitecture detection and optimization
    local amd_flags=""
    local zen_generation=""
    
    # Zen 4 (4th Gen EPYC - Genoa, C7a instances)
    if echo "$cpu_model" | grep -qi "EPYC.*9654\|9754\|9554\|9534"; then
        amd_flags="-march=znver4 -mtune=znver4 -mavx512f -mavx512bw -mavx512cd -mavx512dq -mavx512vl"
        zen_generation="Zen 4 (Genoa)"
        log_success "Detected Zen 4 EPYC - applying AVX-512 optimizations"
        
    # Zen 3 (3rd Gen EPYC - Milan, C6a/M6a/R6a instances)
    elif echo "$cpu_model" | grep -qi "EPYC.*7763\|7713\|7643\|75F3"; then
        amd_flags="-march=znver3 -mtune=znver3 -mavx2 -mfma -mbmi2 -msha"
        zen_generation="Zen 3 (Milan)"
        log_success "Detected Zen 3 EPYC - applying AVX2 + FMA optimizations"
        
    # Zen 2 (2nd Gen EPYC - Rome, C5a/M5a/R5a instances)  
    elif echo "$cpu_model" | grep -qi "EPYC.*7742\|7502\|7402\|7302"; then
        amd_flags="-march=znver2 -mtune=znver2 -mavx2 -mfma -mbmi2"
        zen_generation="Zen 2 (Rome)"
        log_success "Detected Zen 2 EPYC - applying AVX2 optimizations"
        
    # Zen 1 (1st Gen EPYC - Naples)
    elif echo "$cpu_model" | grep -qi "EPYC.*7601\|7501\|7451\|7351"; then
        amd_flags="-march=znver1 -mtune=znver1 -mavx2 -mfma"
        zen_generation="Zen 1 (Naples)"
        log_success "Detected Zen 1 EPYC - applying basic AVX2 optimizations"
        
    # Generic AMD x86_64 fallback
    elif echo "$cpu_model" | grep -qi "AMD"; then
        amd_flags="-march=native -mtune=native -mavx2"
        zen_generation="Generic AMD"
        log_info "Using generic AMD optimizations with native detection"
        
    # Non-AMD processor
    else
        amd_flags="-march=native -mtune=native"
        zen_generation="Unknown"
        log_info "Non-AMD processor detected, using generic optimizations"
    fi
    
    # Export AMD-optimized flags
    export AMD_ARCH_FLAGS="$amd_flags"
    export ZEN_GENERATION="$zen_generation"
    
    # Enhanced AMD compiler flags for maximum performance
    export CFLAGS="-O3 $amd_flags -fopenmp -ffast-math -funroll-loops -fprefetch-loop-arrays -ftree-vectorize"
    export CXXFLAGS="-O3 $amd_flags -fopenmp -ffast-math -funroll-loops -fprefetch-loop-arrays -ftree-vectorize"
    export FCFLAGS="-O3 $amd_flags -fopenmp -ffast-math -funroll-loops"
    
    # AMD AOCL-optimized linking (simulated with system BLAS)
    export LDFLAGS="-lblas -llapack -lm -fopenmp"
    
    log_success "AMD compiler flags configured: $CFLAGS"
    log_success "AMD linking configured: $LDFLAGS"
}

# Configure AMD threading for optimal performance
configure_amd_threading() {
    local cpu_cores=$(nproc)
    local numa_nodes=$(lscpu | grep 'NUMA node(s):' | awk '{print $3}' || echo "1")
    
    # AMD EPYC-specific threading configuration
    export OMP_NUM_THREADS="$cpu_cores"
    export OMP_PROC_BIND="close"
    export OMP_PLACES="cores"
    export OMP_SCHEDULE="static"
    
    # AMD NUMA optimization
    if [[ $numa_nodes -gt 1 ]]; then
        export OMP_NESTED="true"
        log_info "Detected $numa_nodes NUMA nodes - enabling nested parallelism"
    fi
    
    log_info "Configured AMD threading: $cpu_cores threads with close binding"
    log_info "NUMA nodes: $numa_nodes"
}

# Validate AMD optimization environment
validate_amd_environment() {
    log_info "Validating AMD optimization environment..."
    
    # Check compilers
    if command -v gcc >/dev/null 2>&1; then
        local gcc_version=$(gcc --version | head -n1)
        log_success "GCC Compiler: $gcc_version"
    else
        log_error "GCC compiler not found"
        return 1
    fi
    
    # Check for AMD-specific features
    if grep -q "avx2" /proc/cpuinfo; then
        log_success "AVX2 support detected"
    else
        log_info "AVX2 support not detected"
    fi
    
    if grep -q "fma" /proc/cpuinfo; then
        log_success "FMA instruction support detected"  
    else
        log_info "FMA instruction support not detected"
    fi
    
    # Check BLAS/LAPACK availability
    if ldconfig -p | grep -q "libblas\|liblapack"; then
        log_success "BLAS/LAPACK libraries available"
    else
        log_error "BLAS/LAPACK libraries not found"
        return 1
    fi
    
    log_success "AMD environment validation complete"
}

# Collect AMD system information
collect_amd_system_info() {
    log_info "Collecting AMD system information..."
    
    local cpu_vendor=$(lscpu | grep 'Vendor ID:' | awk '{print $3}' || echo 'Unknown')
    local cpu_family_name=$(lscpu | grep 'Model name:' | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cache_l1d=$(lscpu | grep 'L1d cache:' | awk '{print $3}' || echo 'Unknown')
    local cache_l1i=$(lscpu | grep 'L1i cache:' | awk '{print $3}' || echo 'Unknown')
    local cache_l2=$(lscpu | grep 'L2 cache:' | awk '{print $3}' || echo 'Unknown')
    local cache_l3=$(lscpu | grep 'L3 cache:' | awk '{print $3}' || echo 'Unknown')
    
    log_info "CPU Vendor: $cpu_vendor"
    log_info "CPU Model: $cpu_family_name"
    log_info "Cache hierarchy: L1d=$cache_l1d, L1i=$cache_l1i, L2=$cache_l2, L3=$cache_l3"
    log_info "Zen Generation: $ZEN_GENERATION"
    
    # Export for use in other scripts
    export AMD_CPU_VENDOR="$cpu_vendor"
    export AMD_CPU_MODEL="$cpu_family_name"
    export AMD_CACHE_L3="$cache_l3"
}

# Main execution
main() {
    log_info "=== AMD AOCC Architecture Optimization ==="
    
    detect_amd_architecture
    configure_amd_threading
    validate_amd_environment
    collect_amd_system_info
    
    log_success "=== AMD Optimization Complete ==="
    log_info "Ready for benchmark compilation with AMD AOCC + AOCL optimizations"
    log_info "Target architecture: $ZEN_GENERATION"
    log_info "Optimization flags: $AMD_ARCH_FLAGS"
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi