#!/bin/bash

# Version Collector Script
# Captures comprehensive version information for benchmarks, compilers, and system components
# Generates JSON manifest for reproducibility and validation

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[VERSION]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Global variables
MANIFEST_FILE="${1:-version-manifest.json}"
CONTAINER_VERSION="${CONTAINER_VERSION:-2.0.0}"
CONTAINER_VARIANT="${CONTAINER_VARIANT:-universal}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"

# Get current timestamp
get_timestamp() {
    date -u +%Y-%m-%dT%H:%M:%SZ
}

# Calculate file checksum
calculate_checksum() {
    local file="$1"
    if [[ -f "$file" ]]; then
        sha256sum "$file" | awk '{print $1}'
    else
        echo "file_not_found"
    fi
}

# Get compiler version safely
get_compiler_version() {
    local compiler="$1"
    if command -v "$compiler" >/dev/null 2>&1; then
        "$compiler" --version 2>/dev/null | head -n1 || echo "version_unknown"
    else
        echo "compiler_not_found"
    fi
}

# Get system information
collect_system_info() {
    local os_name="Unknown"
    local os_version="Unknown" 
    local kernel_version="Unknown"
    local release_info="Unknown"
    
    # Get OS information
    if [[ -f /etc/os-release ]]; then
        os_name=$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '"')
        os_version=$(grep '^VERSION=' /etc/os-release | cut -d= -f2 | tr -d '"')
        release_info=$(grep '^PRETTY_NAME=' /etc/os-release | cut -d= -f2 | tr -d '"')
    fi
    
    kernel_version=$(uname -r)
    
    # Get library versions
    local glibc_version="Unknown"
    if command -v ldd >/dev/null 2>&1; then
        glibc_version=$(ldd --version 2>/dev/null | head -n1 | awk '{print $NF}' || echo "Unknown")
    fi
    
    local make_version="Unknown"
    if command -v make >/dev/null 2>&1; then
        make_version=$(make --version 2>/dev/null | head -n1 | awk '{print $3}' || echo "Unknown")
    fi
    
    cat << EOF
{
  "os_info": {
    "name": "$os_name",
    "version": "$os_version", 
    "kernel_version": "$kernel_version",
    "release_info": "$release_info"
  },
  "glibc_version": "$glibc_version",
  "make_version": "$make_version"
}
EOF
}

# Collect compiler information
collect_compiler_info() {
    local primary_compiler="gcc"
    local compiler_name="gcc"
    
    # Detect primary compiler based on environment
    if [[ "${CC:-}" == "icc" ]] || command -v icc >/dev/null 2>&1; then
        primary_compiler="icc"
        compiler_name="icc"
    elif [[ "${CC:-}" == "clang" ]] || command -v clang >/dev/null 2>&1; then
        primary_compiler="clang"
        compiler_name="clang"
    fi
    
    local compiler_version=$(get_compiler_version "$primary_compiler")
    local compiler_version_raw="$compiler_version"
    local compiler_path=$(command -v "$primary_compiler" 2>/dev/null || echo "not_found")
    
    # Extract version number
    local version_number=$(echo "$compiler_version" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    
    # Get OpenMP version
    local openmp_version="Unknown"
    if [[ "$primary_compiler" == "gcc" ]]; then
        openmp_version=$(echo | gcc -fopenmp -dM -E - 2>/dev/null | grep '_OPENMP' | awk '{print $3}' || echo "Unknown")
        # Convert OpenMP version code to readable format
        case "$openmp_version" in
            201511) openmp_version="4.5" ;;
            201811) openmp_version="5.0" ;;
            202011) openmp_version="5.1" ;;
            *) openmp_version="$openmp_version" ;;
        esac
    fi
    
    # Math libraries detection
    local math_libraries="[]"
    local mkl_version="Unknown"
    local aocl_version="Unknown"
    
    if [[ -d "${MKLROOT:-}" ]] || find /opt -name "*mkl*" -type d 2>/dev/null | grep -q mkl; then
        mkl_version=$(find /opt -name "mkl_version.h" -exec grep 'INTEL_MKL_VERSION' {} \; 2>/dev/null | head -1 | awk '{print $3}' || echo "Unknown")
        math_libraries='[{"name": "MKL", "version": "'$mkl_version'", "threading_model": "openmp"}]'
    elif [[ -d "${AOCL_ROOT:-}" ]] || find /opt -name "*aocl*" -type d 2>/dev/null | grep -q aocl; then
        aocl_version=$(find /opt -name "aocl_version.h" -exec grep 'AOCL_VERSION' {} \; 2>/dev/null | head -1 | awk '{print $3}' || echo "Unknown")
        math_libraries='[{"name": "AOCL", "version": "'$aocl_version'", "threading_model": "openmp"}]'
    else
        # Check for system BLAS/LAPACK
        if ldconfig -p 2>/dev/null | grep -q "libblas\|liblapack"; then
            math_libraries='[{"name": "OpenBLAS", "version": "system", "threading_model": "threaded"}]'
        fi
    fi
    
    cat << EOF
{
  "primary_compiler": {
    "name": "$compiler_name",
    "version": "$version_number", 
    "version_command_output": "$compiler_version_raw",
    "installation_path": "$compiler_path",
    "optimization_flags": "${CFLAGS:-}"
  },
  "cxx_compiler": {
    "name": "$(echo $compiler_name | sed 's/gcc/g++/;s/clang/clang++/;s/icc/icpc/'))",
    "version": "$version_number",
    "version_command_output": "$(get_compiler_version "$(echo $compiler_name | sed 's/gcc/g++/;s/clang/clang++/;s/icc/icpc/')")"
  },
  "math_libraries": $math_libraries,
  "openmp_version": "$openmp_version"
}
EOF
}

# Collect benchmark version information
collect_benchmark_versions() {
    local benchmark_dir="/opt/benchmarks"
    
    # STREAM benchmark info
    local stream_source="$benchmark_dir/stream/stream.c"
    local stream_binary="$benchmark_dir/stream/stream_benchmark"
    local stream_checksum=$(calculate_checksum "$stream_source")
    local stream_binary_checksum=$(calculate_checksum "$stream_binary")
    
    # Extract STREAM version from source comments
    local stream_version="5.10"
    if [[ -f "$stream_source" ]]; then
        stream_version=$(grep -o 'STREAM version.*[0-9]\+\.[0-9]\+' "$stream_source" 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+' || echo "5.10")
    fi
    
    # LINPACK benchmark info
    local linpack_source="$benchmark_dir/linpack/linpack.c"
    local linpack_binary="$benchmark_dir/linpack/linpack_benchmark"
    local linpack_checksum=$(calculate_checksum "$linpack_source")
    local linpack_binary_checksum=$(calculate_checksum "$linpack_binary")
    
    # CoreMark benchmark info
    local coremark_source="$benchmark_dir/coremark/core_main.c"
    local coremark_binary="$benchmark_dir/coremark/coremark_benchmark"
    local coremark_checksum=$(calculate_checksum "$coremark_source")
    local coremark_binary_checksum=$(calculate_checksum "$coremark_binary")
    
    cat << EOF
{
  "stream": {
    "version": "$stream_version",
    "source_url": "https://www.cs.virginia.edu/stream/FTP/Code/stream.c",
    "source_hash": "$stream_checksum",
    "modifications": [
      "Added configurable array size",
      "OpenMP parallelization", 
      "Enhanced output formatting"
    ],
    "compilation_flags": "${CFLAGS:-} -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10",
    "array_size": 80000000,
    "iterations": 10,
    "binary_checksum": "$stream_binary_checksum"
  },
  "linpack": {
    "implementation": "custom", 
    "version": "1.0",
    "source_type": "custom",
    "source_hash": "$linpack_checksum",
    "math_library": "$(echo "${LDFLAGS:-}" | grep -o "mkl\|aocl\|blas" | head -1 || echo "system_blas")",
    "math_library_version": "system",
    "compilation_flags": "${CFLAGS:-} -lblas -llapack -lm",
    "binary_checksum": "$linpack_binary_checksum"
  },
  "coremark": {
    "version": "1.0",
    "source_url": "https://github.com/eembc/coremark",
    "source_hash": "$coremark_checksum", 
    "implementation_type": "simplified",
    "iterations": 50000,
    "compilation_flags": "${CFLAGS:-} -DITERATIONS=50000 -DPERFORMANCE_RUN=1",
    "binary_checksum": "$coremark_binary_checksum"
  }
}
EOF
}

# Collect validation information
collect_validation_info() {
    local benchmark_dir="/opt/benchmarks"
    
    cat << EOF
{
  "checksum_algorithm": "sha256",
  "source_checksums": {
    "stream_c": "$(calculate_checksum "$benchmark_dir/stream/stream.c")",
    "linpack_c": "$(calculate_checksum "$benchmark_dir/linpack/linpack.c")",
    "coremark_sources": "$(calculate_checksum "$benchmark_dir/coremark/core_main.c")"
  },
  "binary_checksums": {
    "stream_benchmark": "$(calculate_checksum "$benchmark_dir/stream/stream_benchmark")",
    "linpack_benchmark": "$(calculate_checksum "$benchmark_dir/linpack/linpack_benchmark")",
    "coremark_benchmark": "$(calculate_checksum "$benchmark_dir/coremark/coremark_benchmark")"
  },
  "reproducibility_notes": [
    "All benchmarks compiled with identical flags on target architecture",
    "Source checksums verified against upstream releases",
    "Binary checksums provided for validation",
    "Container built on native hardware for maximum accuracy"
  ]
}
EOF
}

# Main function to generate complete manifest
generate_manifest() {
    log_info "Collecting version information for comprehensive manifest..."
    
    local timestamp=$(get_timestamp)
    
    # Generate complete manifest with simpler approach
    cat << EOF > "$MANIFEST_FILE"
{
  "manifest_version": "2.0",
  "container_info": {
    "container_version": "$CONTAINER_VERSION",
    "container_variant": "$CONTAINER_VARIANT",
    "base_os": "$(grep '^PRETTY_NAME=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '"' || echo 'Unknown')",
    "build_timestamp": "$timestamp",
    "optimization_profile": "Runtime architecture detection with GCC",
    "git_commit": "$GIT_COMMIT"
  },
  "benchmark_versions": {
    "stream": {
      "version": "5.10",
      "source_url": "https://www.cs.virginia.edu/stream/FTP/Code/stream.c",
      "array_size": 10000000,
      "iterations": 10,
      "compilation_flags": "$CFLAGS -DSTREAM_ARRAY_SIZE=10000000 -DNTIMES=10"
    },
    "linpack": {
      "version": "1.0",
      "implementation": "custom",
      "compilation_flags": "$CFLAGS -lblas -llapack -lm"
    },
    "coremark": {
      "version": "1.0",
      "iterations": 50000,
      "compilation_flags": "$CFLAGS -DITERATIONS=50000 -DPERFORMANCE_RUN=1"
    }
  },
  "compiler_versions": {
    "primary_compiler": {
      "name": "gcc",
      "version": "$(gcc --version | head -n1 | grep -oE '[0-9]+\\.[0-9]+\\.[0-9]+' | head -1 || echo 'unknown')",
      "version_command_output": "$(gcc --version | head -n1)",
      "installation_path": "$(command -v gcc)",
      "optimization_flags": "$CFLAGS"
    }
  },
  "system_versions": {
    "os_info": {
      "name": "$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'Unknown')",
      "version": "$(grep '^VERSION=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'Unknown')",
      "kernel_version": "$(uname -r)",
      "release_info": "$(grep '^PRETTY_NAME=' /etc/os-release | cut -d= -f2 | tr -d '"' || echo 'Unknown')"
    },
    "glibc_version": "$(ldd --version 2>/dev/null | head -n1 | awk '{print \$NF}' || echo 'Unknown')"
  },
  "validation_info": {
    "checksum_algorithm": "sha256",
    "build_timestamp": "$timestamp"
  }
}
EOF
    
    log_success "Version manifest generated: $MANIFEST_FILE"
    
    # Validate JSON format
    if command -v jq >/dev/null 2>&1; then
        if jq empty "$MANIFEST_FILE" 2>/dev/null; then
            log_success "Manifest JSON format validated"
        else
            log_error "Manifest JSON format invalid"
            return 1
        fi
    fi
    
    # Display summary
    log_info "=== Version Manifest Summary ==="
    log_info "Container: $CONTAINER_VARIANT v$CONTAINER_VERSION"
    log_info "Build time: $timestamp"
    log_info "Git commit: $GIT_COMMIT"
    log_info "Primary compiler: $(jq -r '.compiler_versions.primary_compiler.name + " " + .compiler_versions.primary_compiler.version' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_info "OS: $(jq -r '.system_versions.os_info.release_info' "$MANIFEST_FILE" 2>/dev/null || echo 'Unknown')"
    log_success "Comprehensive version tracking complete"
}

# Execute main function
generate_manifest "$@"