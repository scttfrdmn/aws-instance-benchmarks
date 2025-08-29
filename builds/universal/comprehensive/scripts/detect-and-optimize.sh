#!/bin/bash

# Architecture Detection and Optimization Script
# Detects CPU architecture and sets optimal compiler flags for benchmarks
# Supports Intel (Sapphire Rapids to Nehalem), AMD (Zen4 to Zen1), and ARM (Graviton1-4)

set -euo pipefail

# Global variables
ARCH=""
CPU_MODEL=""
OPTIMIZATION_FLAGS=""
DETECTED_FEATURES=()

# Color output for better visibility
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Detect system architecture and CPU model
detect_architecture() {
    log_info "Detecting system architecture..."
    
    ARCH=$(uname -m)
    log_info "System architecture: $ARCH"
    
    # Get detailed CPU information
    if [[ -f /proc/cpuinfo ]]; then
        CPU_MODEL=$(grep -m1 "model name" /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//')
        log_info "CPU Model: $CPU_MODEL"
    else
        log_warning "Cannot read /proc/cpuinfo, using generic settings"
        CPU_MODEL="Unknown"
    fi
}

# Detect Intel CPU generation and set optimization flags
detect_intel_optimization() {
    log_info "Detecting Intel CPU generation..."
    
    local cpu_flags
    if [[ -f /proc/cpuinfo ]]; then
        cpu_flags=$(grep -m1 "^flags" /proc/cpuinfo | cut -d: -f2)
    else
        cpu_flags=""
    fi
    
    # Detect Intel generations based on model name and CPU flags
    if echo "$CPU_MODEL" | grep -qi "Platinum.*8488C\|Xeon.*8488"; then
        # Intel Sapphire Rapids
        OPTIMIZATION_FLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
        DETECTED_FEATURES+=("avx512f" "avx512vnni" "avx512bf16" "avx512_bitalg")
        log_success "Detected Intel Sapphire Rapids - enabling AVX-512 VNNI/BF16"
        
    elif echo "$CPU_MODEL" | grep -qi "Platinum.*8375C\|Platinum.*8370C"; then
        # Intel Ice Lake
        OPTIMIZATION_FLAGS="-O3 -march=icelake-server -mavx512f -mavx512dq -fopenmp"
        DETECTED_FEATURES+=("avx512f" "avx512dq" "avx512cd")
        log_success "Detected Intel Ice Lake - enabling AVX-512"
        
    elif echo "$CPU_MODEL" | grep -qi "Platinum.*8275CL\|Platinum.*8259CL"; then
        # Intel Cascade Lake
        OPTIMIZATION_FLAGS="-O3 -march=cascadelake -mavx512f -fopenmp"
        DETECTED_FEATURES+=("avx512f" "avx512dq")
        log_success "Detected Intel Cascade Lake - enabling AVX-512"
        
    elif echo "$CPU_MODEL" | grep -qi "Platinum.*8175M\|Xeon.*8175"; then
        # Intel Skylake
        OPTIMIZATION_FLAGS="-O3 -march=skylake -mavx2 -mfma -fopenmp"
        DETECTED_FEATURES+=("avx2" "fma")
        log_success "Detected Intel Skylake - enabling AVX2"
        
    elif echo "$CPU_MODEL" | grep -qi "E5-.*v2\|E5-.*v3"; then
        # Intel Ivy Bridge / Haswell
        OPTIMIZATION_FLAGS="-O3 -march=ivybridge -mavx -msse4.2 -fopenmp"
        DETECTED_FEATURES+=("avx" "sse4_2")
        log_success "Detected Intel Ivy Bridge/Haswell - enabling AVX"
        
    else
        # Generic Intel with native detection
        log_warning "Unknown Intel CPU, using -march=native"
        OPTIMIZATION_FLAGS="-O3 -march=native -mtune=native -fopenmp"
        DETECTED_FEATURES+=("native")
    fi
}

# Detect AMD CPU generation and set optimization flags  
detect_amd_optimization() {
    log_info "Detecting AMD CPU generation..."
    
    if echo "$CPU_MODEL" | grep -qi "EPYC.*9R14\|EPYC.*9654"; then
        # AMD Zen4 (Genoa)
        OPTIMIZATION_FLAGS="-O3 -march=znver4 -mavx512f -fopenmp"
        DETECTED_FEATURES+=("avx512f" "avx2" "zen4")
        log_success "Detected AMD Zen4 (Genoa) - enabling AVX-512"
        
    elif echo "$CPU_MODEL" | grep -qi "EPYC.*7R13\|EPYC.*7763"; then
        # AMD Zen3 (Milan)
        OPTIMIZATION_FLAGS="-O3 -march=znver3 -mavx2 -fopenmp"
        DETECTED_FEATURES+=("avx2" "zen3")
        log_success "Detected AMD Zen3 (Milan) - enabling AVX2"
        
    elif echo "$CPU_MODEL" | grep -qi "EPYC.*7R32\|EPYC.*7742"; then
        # AMD Zen2 (Rome)
        OPTIMIZATION_FLAGS="-O3 -march=znver2 -mavx2 -fopenmp"
        DETECTED_FEATURES+=("avx2" "zen2")
        log_success "Detected AMD Zen2 (Rome) - enabling AVX2"
        
    elif echo "$CPU_MODEL" | grep -qi "EPYC.*7R01\|EPYC.*7551"; then
        # AMD Zen1 (Naples)
        OPTIMIZATION_FLAGS="-O3 -march=znver1 -mavx2 -fopenmp"
        DETECTED_FEATURES+=("avx2" "zen1")
        log_success "Detected AMD Zen1 (Naples) - enabling AVX2"
        
    else
        # Generic AMD with native detection
        log_warning "Unknown AMD CPU, using -march=native"
        OPTIMIZATION_FLAGS="-O3 -march=native -mtune=native -fopenmp"
        DETECTED_FEATURES+=("native")
    fi
}

# Detect ARM/Graviton CPU and set optimization flags
detect_arm_optimization() {
    log_info "Detecting ARM/Graviton CPU generation..."
    
    if echo "$CPU_MODEL" | grep -qi "Neoverse-V2"; then
        # Graviton4
        OPTIMIZATION_FLAGS="-O3 -mcpu=neoverse-v2 -mtune=neoverse-v2 -fopenmp"
        DETECTED_FEATURES+=("neoverse-v2" "sve2" "armv8.5-a")
        log_success "Detected ARM Graviton4 (Neoverse-V2) - enabling SVE2"
        
    elif echo "$CPU_MODEL" | grep -qi "Neoverse-V1"; then
        # Graviton3/3E
        OPTIMIZATION_FLAGS="-O3 -mcpu=neoverse-v1 -mtune=neoverse-v1 -fopenmp"
        DETECTED_FEATURES+=("neoverse-v1" "sve" "armv8.4-a")
        log_success "Detected ARM Graviton3 (Neoverse-V1) - enabling SVE"
        
    elif echo "$CPU_MODEL" | grep -qi "Neoverse-N1"; then
        # Graviton2
        OPTIMIZATION_FLAGS="-O3 -mcpu=neoverse-n1 -mtune=neoverse-n1 -fopenmp"
        DETECTED_FEATURES+=("neoverse-n1" "armv8.2-a")
        log_success "Detected ARM Graviton2 (Neoverse-N1) - enabling ARMv8.2-A"
        
    elif echo "$CPU_MODEL" | grep -qi "Cortex-A72"; then
        # Graviton1
        OPTIMIZATION_FLAGS="-O3 -mcpu=cortex-a72 -mtune=cortex-a72 -fopenmp"
        DETECTED_FEATURES+=("cortex-a72" "armv8.0-a")
        log_success "Detected ARM Graviton1 (Cortex-A72) - enabling ARMv8.0-A"
        
    else
        # Generic ARM with native detection
        log_warning "Unknown ARM CPU, using -mcpu=native"
        OPTIMIZATION_FLAGS="-O3 -mcpu=native -mtune=native -fopenmp"
        DETECTED_FEATURES+=("native")
    fi
}

# Main architecture detection logic
detect_optimization_flags() {
    detect_architecture
    
    case "$ARCH" in
        x86_64)
            if echo "$CPU_MODEL" | grep -qi intel; then
                detect_intel_optimization
            elif echo "$CPU_MODEL" | grep -qi amd; then
                detect_amd_optimization
            else
                log_warning "Unknown x86_64 CPU vendor, using generic flags"
                OPTIMIZATION_FLAGS="-O3 -march=native -mtune=native -fopenmp"
                DETECTED_FEATURES+=("native")
            fi
            ;;
        aarch64|arm64)
            detect_arm_optimization
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            OPTIMIZATION_FLAGS="-O3 -fopenmp"
            DETECTED_FEATURES+=("generic")
            ;;
    esac
    
    # Export flags for use by build scripts
    export CFLAGS="$OPTIMIZATION_FLAGS"
    export CXXFLAGS="$OPTIMIZATION_FLAGS"
    export FCFLAGS="$OPTIMIZATION_FLAGS"
    
    log_success "Optimization flags set: $OPTIMIZATION_FLAGS"
}

# Build benchmarks with detected optimization
build_benchmarks() {
    log_info "Building benchmarks with architecture-specific optimizations..."
    
    cd /opt/benchmarks
    
    if [[ -f build-benchmarks.sh ]]; then
        log_info "Running build script with flags: $OPTIMIZATION_FLAGS"
        bash ./build-benchmarks.sh
        log_success "All benchmarks built successfully"
    else
        log_error "Build script not found!"
        exit 1
    fi
}

# Output detection results in JSON format
output_detection_results() {
    cat << EOF
{
  "architecture_detection": {
    "system_architecture": "$ARCH",
    "cpu_model": "$CPU_MODEL",
    "optimization_flags": "$OPTIMIZATION_FLAGS",
    "detected_features": [$(printf '"%s",' "${DETECTED_FEATURES[@]}" | sed 's/,$//')]
  },
  "compiler_info": {
    "gcc_version": "$(gcc --version | head -n1)",
    "compile_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
}

# Main execution
main() {
    log_info "=== Enhanced Universal Benchmark Container v2.0 ==="
    log_info "Architecture detection and optimization starting..."
    
    detect_optimization_flags
    
    if [[ "${1:-}" == "--build-only" ]]; then
        build_benchmarks
        log_success "Build complete. Flags: $OPTIMIZATION_FLAGS"
    elif [[ "${1:-}" == "--detect-only" ]]; then
        output_detection_results
    else
        # Normal execution - detect and build
        build_benchmarks
        log_success "=== Architecture detection and build complete ==="
        log_info "Container ready for benchmark execution"
    fi
}

# Execute main function with all arguments
main "$@"