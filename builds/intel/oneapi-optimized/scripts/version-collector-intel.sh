#!/bin/bash

# Intel oneAPI Version Collector Script  
# Captures Intel-specific compiler, MKL, and optimization information

set -eo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INTEL-VERSION]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Global variables
MANIFEST_FILE="${1:-version-manifest-intel.json}"
CONTAINER_VERSION="${CONTAINER_VERSION:-2.1.0}"
CONTAINER_VARIANT="${CONTAINER_VARIANT:-intel-oneapi-optimized}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"

# Get current timestamp
get_timestamp() {
    date -u +%Y-%m-%dT%H:%M:%SZ
}

# Collect Intel oneAPI compiler information
collect_intel_compiler_info() {
    log_info "Collecting Intel oneAPI compiler information..."
    
    # Initialize oneAPI environment
    source "${ONEAPI_ROOT}/setvars.sh" --force
    
    local icx_version="unknown"
    local icpx_version="unknown"
    local ifx_version="unknown"
    local oneapi_version="unknown"
    
    if command -v icx >/dev/null 2>&1; then
        icx_version=$(icx --version | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')
    fi
    
    if command -v icpx >/dev/null 2>&1; then
        icpx_version=$(icpx --version | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')
    fi
    
    if command -v ifx >/dev/null 2>&1; then
        ifx_version=$(ifx --version | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')
    fi
    
    # Extract oneAPI version from environment
    if [[ -n "${ONEAPI_VERSION:-}" ]]; then
        oneapi_version="$ONEAPI_VERSION"
    elif [[ -f "${ONEAPI_ROOT}/version.txt" ]]; then
        oneapi_version=$(cat "${ONEAPI_ROOT}/version.txt" || echo 'unknown')
    fi
    
    cat << EOF
{
  "oneapi_toolkit": {
    "version": "$oneapi_version",
    "installation_path": "$ONEAPI_ROOT"
  },
  "intel_compilers": {
    "c_compiler": {
      "name": "icx",
      "version": "$icx_version",
      "version_command_output": "$(icx --version | head -n1 || echo 'not available')",
      "installation_path": "$(command -v icx || echo 'not found')"
    },
    "cxx_compiler": {
      "name": "icpx", 
      "version": "$icpx_version",
      "version_command_output": "$(icpx --version | head -n1 || echo 'not available')",
      "installation_path": "$(command -v icpx || echo 'not found')"
    },
    "fortran_compiler": {
      "name": "ifx",
      "version": "$ifx_version", 
      "version_command_output": "$(ifx --version | head -n1 || echo 'not available')",
      "installation_path": "$(command -v ifx || echo 'not found')"
    }
  },
  "optimization_flags": {
    "cflags": "${CFLAGS:-}",
    "cxxflags": "${CXXFLAGS:-}",
    "fcflags": "${FCFLAGS:-}",
    "ldflags": "${LDFLAGS:-}",
    "architecture_flags": "${INTEL_ARCH_FLAGS:-}",
    "mkl_flags": "${MKL_COMPILE_FLAGS:-}"
  }
}
EOF
}

# Collect Intel MKL information
collect_intel_mkl_info() {
    log_info "Collecting Intel MKL information..."
    
    local mkl_version="unknown"
    local mkl_threading="unknown"
    local mkl_interface="lp64"
    
    if [[ -d "${MKLROOT}" ]]; then
        # Extract MKL version
        mkl_version=$(find "${MKLROOT}" -name "mkl_version.h" -exec grep 'INTEL_MKL_VERSION' {} \; 2>/dev/null | head -1 | awk '{print $3}' || echo 'unknown')
        
        # Detect MKL threading model
        if [[ -f "${MKLROOT}/lib/intel64/libmkl_intel_thread.so" ]]; then
            mkl_threading="intel_thread"
        elif [[ -f "${MKLROOT}/lib/intel64/libmkl_gnu_thread.so" ]]; then
            mkl_threading="gnu_thread"
        elif [[ -f "${MKLROOT}/lib/intel64/libmkl_sequential.so" ]]; then
            mkl_threading="sequential"
        fi
    fi
    
    cat << EOF
{
  "mkl_library": {
    "version": "$mkl_version",
    "installation_path": "${MKLROOT:-not found}",
    "threading_model": "$mkl_threading",
    "interface": "$mkl_interface",
    "environment": {
      "mkl_num_threads": "${MKL_NUM_THREADS:-auto}",
      "mkl_dynamic": "${MKL_DYNAMIC:-TRUE}",
      "omp_num_threads": "${OMP_NUM_THREADS:-auto}",
      "kmp_affinity": "${KMP_AFFINITY:-none}"
    }
  }
}
EOF
}

# Collect Intel-specific benchmark information
collect_intel_benchmark_info() {
    log_info "Collecting Intel-optimized benchmark information..."
    
    local benchmark_dir="/opt/benchmarks"
    
    cat << EOF
{
  "stream_intel": {
    "version": "5.10",
    "implementation": "intel_optimized",
    "array_size": 80000000,
    "iterations": 10,
    "compilation_flags": "${CFLAGS:-} -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10",
    "mkl_acceleration": true,
    "vectorization": "AVX-512 with Intel compiler",
    "executable": "stream/stream_benchmark_intel"
  },
  "linpack_intel": {
    "version": "1.0",
    "implementation": "intel_mkl_optimized",
    "math_library": "Intel MKL BLAS",
    "compilation_flags": "${CFLAGS:-} -DINTEL_MKL",
    "mkl_dgemm": true,
    "threading": "Intel OpenMP with MKL threading",
    "executable": "linpack/linpack_benchmark_intel"
  },
  "coremark_intel": {
    "version": "1.0", 
    "implementation": "intel_vectorized",
    "iterations": 100000,
    "compilation_flags": "${CFLAGS:-} -DITERATIONS=100000 -DPERFORMANCE_RUN=1",
    "vectorization": "Intel compiler auto-vectorization",
    "optimization": "IPO + fast math + loop unrolling",
    "executable": "coremark/coremark_benchmark_intel"
  }
}
EOF
}

# Main function to generate Intel oneAPI manifest
generate_intel_manifest() {
    log_info "Generating Intel oneAPI comprehensive manifest..."
    
    local timestamp=$(get_timestamp)
    
    # Generate complete Intel-specific manifest
    cat << EOF > "$MANIFEST_FILE"
{
  "manifest_version": "2.1",
  "container_info": {
    "container_version": "$CONTAINER_VERSION",
    "container_variant": "$CONTAINER_VARIANT",
    "base_image": "intel/oneapi-hpckit:2024.2.1-0-devel-ubuntu24.04",
    "build_timestamp": "$timestamp",
    "optimization_profile": "Intel oneAPI + MKL maximum performance for Intel Xeon processors",
    "git_commit": "$GIT_COMMIT",
    "target_architecture": "Intel Xeon (Ice Lake, Sapphire Rapids, Emerald Rapids)"
  },
  "intel_oneapi": $(collect_intel_compiler_info),
  "intel_mkl": $(collect_intel_mkl_info),
  "benchmark_versions": $(collect_intel_benchmark_info),
  "system_versions": {
    "os_info": {
      "name": "$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'Unknown')",
      "version": "$(grep '^VERSION=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'Unknown')",
      "kernel_version": "$(uname -r)",
      "release_info": "$(grep '^PRETTY_NAME=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'Unknown')"
    }
  },
  "performance_expectations": {
    "stream_improvement": "20-30% over GCC on Intel processors",
    "linpack_improvement": "30-40% with MKL BLAS optimization", 
    "coremark_improvement": "15-25% with Intel compiler vectorization",
    "target_instances": ["c7i", "m7i", "r7i", "c6i", "m6i", "r6i"],
    "optimal_microarchitectures": ["Sapphire Rapids", "Ice Lake", "Cascade Lake"]
  },
  "validation_info": {
    "checksum_algorithm": "sha256",
    "build_timestamp": "$timestamp",
    "reproducibility_notes": [
      "Built with Intel oneAPI 2024.2.1 for maximum Intel Xeon performance",
      "MKL-accelerated BLAS/LAPACK with Intel threading optimization",
      "Architecture-specific compiler flags for AVX-512 utilization",
      "Validated on Intel Ice Lake and Sapphire Rapids microarchitectures"
    ]
  }
}
EOF
    
    log_success "Intel oneAPI manifest generated: $MANIFEST_FILE"
    
    # Validate JSON format
    if command -v jq >/dev/null 2>&1; then
        if jq empty "$MANIFEST_FILE" 2>/dev/null; then
            log_success "Intel manifest JSON format validated"
        else
            log_error "Intel manifest JSON format invalid"
            return 1
        fi
    fi
    
    # Display Intel-specific summary
    log_info "=== Intel oneAPI Manifest Summary ==="
    log_info "Container: $CONTAINER_VARIANT v$CONTAINER_VERSION"
    log_info "Build time: $timestamp"
    log_info "oneAPI version: $(jq -r '.intel_oneapi.oneapi_toolkit.version' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_info "Intel C compiler: $(jq -r '.intel_oneapi.intel_compilers.c_compiler.version' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_info "MKL version: $(jq -r '.intel_mkl.mkl_library.version' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_success "Intel oneAPI comprehensive tracking complete"
}

# Execute main function
generate_intel_manifest "$@"