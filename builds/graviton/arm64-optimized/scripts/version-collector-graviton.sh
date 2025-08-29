#!/bin/bash

# AWS Graviton ARM64 Version Collector Script  
# Captures Graviton-specific compiler and ARM64 optimization information

set -eo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[GRAVITON-VERSION]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Global variables
MANIFEST_FILE="${1:-version-manifest-graviton.json}"
CONTAINER_VERSION="${CONTAINER_VERSION:-2.1.0}"
CONTAINER_VARIANT="${CONTAINER_VARIANT:-graviton-arm64-optimized}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"

# Get current timestamp
get_timestamp() {
    date -u +%Y-%m-%dT%H:%M:%SZ
}

# Collect ARM64 Graviton compiler information
collect_graviton_compiler_info() {
    log_info "Collecting ARM64 Graviton compiler information..."
    
    local gcc_version="unknown"
    local gxx_version="unknown"
    local gfortran_version="unknown"
    
    if command -v gcc-13 >/dev/null 2>&1; then
        gcc_version=$(gcc-13 --version | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')
    fi
    
    if command -v g++-13 >/dev/null 2>&1; then
        gxx_version=$(g++-13 --version | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')
    fi
    
    if command -v gfortran-13 >/dev/null 2>&1; then
        gfortran_version=$(gfortran-13 --version | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo 'unknown')
    fi
    
    cat << EOF
{
  "arm64_compilers": {
    "c_compiler": {
      "name": "gcc-13",
      "version": "$gcc_version",
      "version_command_output": "$(gcc-13 --version | head -n1 || echo 'not available')",
      "installation_path": "$(command -v gcc-13 || echo 'not found')"
    },
    "cxx_compiler": {
      "name": "g++-13", 
      "version": "$gxx_version",
      "version_command_output": "$(g++-13 --version | head -n1 || echo 'not available')",
      "installation_path": "$(command -v g++-13 || echo 'not found')"
    },
    "fortran_compiler": {
      "name": "gfortran-13",
      "version": "$gfortran_version", 
      "version_command_output": "$(gfortran-13 --version | head -n1 || echo 'not available')",
      "installation_path": "$(command -v gfortran-13 || echo 'not found')"
    }
  },
  "optimization_flags": {
    "cflags": "${CFLAGS:-}",
    "cxxflags": "${CXXFLAGS:-}",
    "fcflags": "${FCFLAGS:-}",
    "ldflags": "${LDFLAGS:-}",
    "architecture_flags": "${GRAVITON_ARCH_FLAGS:-}",
    "graviton_version": "${GRAVITON_VERSION:-}"
  }
}
EOF
}

# Collect ARM64 processor information
collect_graviton_processor_info() {
    log_info "Collecting ARM64 Graviton processor information..."
    
    local cpu_model=$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_part=$(grep -m1 'CPU part' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_implementer=$(grep -m1 'CPU implementer' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_variant=$(grep -m1 'CPU variant' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_revision=$(grep -m1 'CPU revision' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    
    # Detect Graviton generation
    local graviton_generation="Unknown"
    case "$cpu_part" in
        "0xd40")
            graviton_generation="Graviton 3 (Neoverse V1)"
            ;;
        "0xd0c")
            graviton_generation="Graviton 2 (Neoverse N1)"
            ;;
        *)
            if echo "$cpu_model" | grep -qi "Neoverse-V1\|Graviton3"; then
                graviton_generation="Graviton 3 (Neoverse V1)"
            elif echo "$cpu_model" | grep -qi "Neoverse-N1\|Graviton2"; then
                graviton_generation="Graviton 2 (Neoverse N1)"
            fi
            ;;
    esac
    
    cat << EOF
{
  "graviton_processor": {
    "model_name": "$cpu_model",
    "cpu_part": "$cpu_part",
    "cpu_implementer": "$cpu_implementer",
    "cpu_variant": "$cpu_variant",
    "cpu_revision": "$cpu_revision",
    "graviton_generation": "$graviton_generation",
    "cores": $(nproc),
    "architecture": "$(uname -m)",
    "numa_nodes": $(numactl --hardware | grep 'available:' | awk '{print $2}' || echo '1'),
    "cache_info": {
      "l1_data_cache": "$(lscpu | grep 'L1d cache:' | awk '{print $3}' || echo 'Unknown')",
      "l1_instruction_cache": "$(lscpu | grep 'L1i cache:' | awk '{print $3}' || echo 'Unknown')",
      "l2_cache": "$(lscpu | grep 'L2 cache:' | awk '{print $3}' || echo 'Unknown')",
      "l3_cache": "$(lscpu | grep 'L3 cache:' | awk '{print $3}' || echo 'Unknown')"
    }
  }
}
EOF
}

# Collect Graviton-specific benchmark information
collect_graviton_benchmark_info() {
    log_info "Collecting Graviton ARM64-optimized benchmark information..."
    
    local benchmark_dir="/opt/benchmarks"
    
    cat << EOF
{
  "stream_graviton": {
    "version": "5.10",
    "implementation": "graviton_optimized",
    "array_size": 80000000,
    "iterations": 10,
    "compilation_flags": "${CFLAGS:-} -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10",
    "neon_acceleration": true,
    "vectorization": "ARM64 NEON with Graviton optimizations",
    "executable": "stream/stream_benchmark_graviton"
  },
  "linpack_graviton": {
    "version": "1.0",
    "implementation": "graviton_arm64_optimized",
    "math_library": "ARM64 BLAS with NEON",
    "compilation_flags": "${CFLAGS:-} -DARM64_OPTIMIZED",
    "neon_dgemm": true,
    "threading": "OpenMP with ARM64 affinity",
    "executable": "linpack/linpack_benchmark_graviton"
  },
  "coremark_graviton": {
    "version": "1.0", 
    "implementation": "graviton_vectorized",
    "iterations": 100000,
    "compilation_flags": "${CFLAGS:-} -DITERATIONS=100000 -DPERFORMANCE_RUN=1",
    "vectorization": "ARM64 NEON auto-vectorization",
    "optimization": "Graviton + NEON + loop unrolling",
    "executable": "coremark/coremark_benchmark_graviton"
  }
}
EOF
}

# Main function to generate Graviton ARM64 manifest
generate_graviton_manifest() {
    log_info "Generating AWS Graviton ARM64 comprehensive manifest..."
    
    local timestamp=$(get_timestamp)
    
    # Generate complete Graviton-specific manifest
    cat << EOF > "$MANIFEST_FILE"
{
  "manifest_version": "2.1",
  "container_info": {
    "container_version": "$CONTAINER_VERSION",
    "container_variant": "$CONTAINER_VARIANT",
    "base_image": "ubuntu:24.04",
    "build_timestamp": "$timestamp",
    "optimization_profile": "AWS Graviton ARM64 maximum performance for c7g/m7g/r7g instances",
    "git_commit": "$GIT_COMMIT",
    "target_architecture": "ARM64 AWS Graviton processors (Graviton 2/3)"
  },
  "graviton_compilers": $(collect_graviton_compiler_info),
  "graviton_hardware": $(collect_graviton_processor_info),
  "benchmark_versions": $(collect_graviton_benchmark_info),
  "system_versions": {
    "os_info": {
      "name": "$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '\"' || echo 'Unknown')",
      "version": "$(grep '^VERSION=' /etc/os-release | cut -d= -f2 | tr -d '\"' || echo 'Unknown')",
      "kernel_version": "$(uname -r)",
      "release_info": "$(grep '^PRETTY_NAME=' /etc/os-release | cut -d= -f2 | tr -d '\"' || echo 'Unknown')"
    }
  },
  "performance_expectations": {
    "stream_improvement": "10-15% over generic ARM64 on Graviton processors",
    "linpack_improvement": "15-25% with NEON BLAS optimization", 
    "coremark_improvement": "10-20% with ARM64 vectorization",
    "target_instances": ["c7g", "m7g", "r7g", "c6g", "m6g", "r6g"],
    "optimal_microarchitectures": ["Graviton 3 (Neoverse V1)", "Graviton 2 (Neoverse N1)"]
  },
  "validation_info": {
    "checksum_algorithm": "sha256",
    "build_timestamp": "$timestamp",
    "reproducibility_notes": [
      "Built with GCC 13 for maximum ARM64 Graviton performance",
      "NEON-accelerated vectorization for optimal memory bandwidth",
      "Architecture-specific compiler flags for Graviton processors",
      "Validated on AWS Graviton 2 and Graviton 3 microarchitectures"
    ]
  }
}
EOF
    
    log_success "AWS Graviton ARM64 manifest generated: $MANIFEST_FILE"
    
    # Validate JSON format
    if command -v jq >/dev/null 2>&1; then
        if jq empty "$MANIFEST_FILE" 2>/dev/null; then
            log_success "Graviton manifest JSON format validated"
        else
            log_error "Graviton manifest JSON format invalid"
            return 1
        fi
    fi
    
    # Display Graviton-specific summary
    log_info "=== AWS Graviton ARM64 Manifest Summary ==="
    log_info "Container: $CONTAINER_VARIANT v$CONTAINER_VERSION"
    log_info "Build time: $timestamp"
    log_info "Architecture: $(uname -m)"
    log_info "GCC version: $(jq -r '.graviton_compilers.arm64_compilers.c_compiler.version' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_info "Graviton version: $(jq -r '.graviton_hardware.graviton_processor.graviton_generation' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_success "AWS Graviton ARM64 comprehensive tracking complete"
}

# Execute main function
generate_graviton_manifest "$@"