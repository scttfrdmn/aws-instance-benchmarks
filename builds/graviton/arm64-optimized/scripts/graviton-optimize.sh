#!/bin/bash

# AWS Graviton ARM64 Architecture Detection and Optimization Script
# Detects Graviton processor versions and applies optimal ARM64 compiler flags

set -eo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[GRAVITON-OPT]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Detect AWS Graviton architecture and apply optimal flags
detect_graviton_architecture() {
    local cpu_model=$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_part=$(grep -m1 'CPU part' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_implementer=$(grep -m1 'CPU implementer' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    
    log_info "Detected ARM64 CPU: $cpu_model"
    log_info "CPU Part: $cpu_part, Implementer: $cpu_implementer"
    
    # AWS Graviton-specific optimization flags
    local graviton_flags=""
    local graviton_version=""
    local numa_flags=""
    
    # Graviton 3 (c7g, m7g, r7g instances) - ARM Neoverse V1
    if echo "$cpu_model" | grep -qi "Neoverse-V1\|Graviton3"; then
        graviton_flags="-mcpu=neoverse-v1 -mtune=neoverse-v1 -march=armv8.2-a+fp16+dotprod"
        graviton_version="Graviton 3 (Neoverse V1)"
        numa_flags="-DNUMA_AWARE"
        log_success "Detected Graviton 3 - applying Neoverse V1 optimizations (container-safe)"
        
    # Graviton 2 (c6g, m6g, r6g instances) - ARM Neoverse N1  
    elif echo "$cpu_model" | grep -qi "Neoverse-N1\|Graviton2"; then
        graviton_flags="-mcpu=neoverse-n1 -mtune=neoverse-n1 -march=armv8.2-a+fp16"
        graviton_version="Graviton 2 (Neoverse N1)"
        numa_flags="-DNUMA_AWARE"
        log_success "Detected Graviton 2 - applying Neoverse N1 optimizations"
        
    # Generic ARM64 fallback
    else
        # Check for specific ARM CPU part numbers
        case "$cpu_part" in
            "0xd40") # Neoverse V1 (Graviton 3)
                graviton_flags="-mcpu=neoverse-v1 -mtune=neoverse-v1 -march=armv8.2-a+fp16+dotprod"
                graviton_version="Graviton 3 (Neoverse V1)"
                numa_flags="-DNUMA_AWARE"
                log_success "Detected Neoverse V1 via CPU part - applying Graviton 3 optimizations (container-safe)"
                ;;
            "0xd0c") # Neoverse N1 (Graviton 2)
                graviton_flags="-mcpu=neoverse-n1 -mtune=neoverse-n1 -march=armv8.2-a+fp16"
                graviton_version="Graviton 2 (Neoverse N1)"
                numa_flags="-DNUMA_AWARE"
                log_success "Detected Neoverse N1 via CPU part - applying Graviton 2 optimizations"
                ;;
            *)
                # Generic ARM64 optimization
                graviton_flags="-mcpu=native -mtune=native -march=native"
                graviton_version="Generic ARM64"
                numa_flags=""
                log_info "Using generic ARM64 optimizations with native detection"
                ;;
        esac
    fi
    
    # Export optimized flags
    export GRAVITON_ARCH_FLAGS="$graviton_flags"
    export GRAVITON_VERSION="$graviton_version"
    export NUMA_FLAGS="$numa_flags"
    
    # Enhanced ARM64 compiler flags for maximum performance
    export CFLAGS="-O3 $graviton_flags -fopenmp -ffast-math -funroll-loops -ftree-vectorize $numa_flags"
    export CXXFLAGS="-O3 $graviton_flags -fopenmp -ffast-math -funroll-loops -ftree-vectorize $numa_flags"
    export FCFLAGS="-O3 $graviton_flags -fopenmp -ffast-math -funroll-loops"
    
    # ARM64 linking with optimized libraries
    export LDFLAGS="-fopenmp -lm -lpthread"
    
    log_success "ARM64 Graviton compiler flags configured: $CFLAGS"
    log_success "Linking configured: $LDFLAGS"
}

# Configure ARM64 threading for optimal performance
configure_graviton_threading() {
    local cpu_cores=$(nproc)
    
    # ARM64 Graviton-specific threading configuration
    export OMP_NUM_THREADS="$cpu_cores"
    export GOMP_CPU_AFFINITY="0-$((cpu_cores-1))"
    export OMP_PROC_BIND="spread"
    export OMP_PLACES="cores"
    
    log_info "Configured ARM64 Graviton threading: $cpu_cores threads with spread affinity"
}

# Validate ARM64 compiler environment
validate_graviton_environment() {
    log_info "Validating ARM64 Graviton build environment..."
    
    # Check compilers
    if command -v gcc-13 >/dev/null 2>&1; then
        local gcc_version=$(gcc-13 --version | head -n1)
        log_success "GCC ARM64 Compiler: $gcc_version"
    else
        log_error "GCC-13 ARM64 compiler not found"
        return 1
    fi
    
    if command -v g++-13 >/dev/null 2>&1; then
        local gxx_version=$(g++-13 --version | head -n1)
        log_success "G++ ARM64 Compiler: $gxx_version"
    else
        log_error "G++-13 ARM64 compiler not found"  
        return 1
    fi
    
    # Check ARM64 architecture
    local arch=$(uname -m)
    if [[ "$arch" == "aarch64" || "$arch" == "arm64" ]]; then
        log_success "Confirmed ARM64 architecture: $arch"
    else
        log_error "Not running on ARM64 architecture: $arch"
        return 1
    fi
    
    log_success "ARM64 Graviton environment validation complete"
}

# Main execution
main() {
    log_info "=== AWS Graviton ARM64 Architecture Optimization ==="
    
    detect_graviton_architecture  
    configure_graviton_threading
    validate_graviton_environment
    
    log_success "=== Graviton Optimization Complete ==="
    log_info "Ready for benchmark compilation with ARM64 Graviton optimizations"
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi