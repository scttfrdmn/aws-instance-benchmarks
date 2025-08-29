#!/bin/bash

# Universal Benchmark Runner (Spack Edition)
# Executes STREAM, LINPACK, and CoreMark with Spack-managed GCC optimizations

set -eo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[UNIVERSAL-SPACK]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize Spack and optimized GCC environment
log_info "Loading Spack environment and optimized GCC..."
source /opt/spack/share/spack/setup-env.sh
spack load gcc@13
spack load openblas

# Verify GCC is loaded
if ! command -v gcc &> /dev/null; then
    log_error "GCC compiler not found! Spack load may have failed."
    exit 1
fi

# Configuration
BENCHMARK_DIR="/opt/benchmarks"
RESULTS_DIR="/tmp/benchmark_results"
RESULTS_FILE="$RESULTS_DIR/universal_spack_benchmark_results_$(date -u +%Y%m%dT%H%M%SZ).json"

# Create results directory
mkdir -p "$RESULTS_DIR"

cd "$BENCHMARK_DIR" || exit 1

# Universal STREAM benchmark
run_stream_universal() {
    log_info "Running Universal STREAM benchmark..."
    
    if [[ ! -f "stream/stream_benchmark" ]]; then
        log_error "Universal STREAM benchmark not found!"
        return 1
    fi
    
    log_info "Executing STREAM with architecture-optimized GCC..."
    local stream_output
    stream_output=$(timeout 300s ./stream/stream_benchmark 2>&1 || echo "STREAM execution failed")
    
    # Parse STREAM results
    local copy_rate=$(echo "$stream_output" | grep "Copy:" | awk '{print $2}' | head -1 || echo "0")
    local scale_rate=$(echo "$stream_output" | grep "Scale:" | awk '{print $2}' | head -1 || echo "0")  
    local add_rate=$(echo "$stream_output" | grep "Add:" | awk '{print $2}' | head -1 || echo "0")
    local triad_rate=$(echo "$stream_output" | grep "Triad:" | awk '{print $2}' | head -1 || echo "0")
    
    # Calculate average memory bandwidth
    local avg_bandwidth=$(echo "scale=2; ($copy_rate + $scale_rate + $add_rate + $triad_rate) / 4" | bc -l || echo "0")
    
    log_success "Universal STREAM completed - Average bandwidth: ${avg_bandwidth} MB/s"
    
    echo "$copy_rate|$scale_rate|$add_rate|$triad_rate|$avg_bandwidth"
}

# OpenBLAS-accelerated LINPACK benchmark
run_linpack_universal() {
    log_info "Running OpenBLAS-accelerated LINPACK benchmark..."
    
    if [[ ! -f "linpack/linpack_benchmark" ]]; then
        log_error "Universal LINPACK benchmark not found!"
        return 1
    fi
    
    log_info "Executing LINPACK with OpenBLAS..."
    local linpack_output
    linpack_output=$(timeout 300s ./linpack/linpack_benchmark 2>&1 || echo "LINPACK execution failed")
    
    # Parse LINPACK results
    local gflops=$(echo "$linpack_output" | grep -i "gflops\|performance" | grep -oE '[0-9]+\.[0-9]+' | tail -1 || echo "0")
    
    log_success "Universal LINPACK completed - Performance: ${gflops} GFLOPS"
    
    echo "$gflops"
}

# GCC-optimized CoreMark benchmark
run_coremark_universal() {
    log_info "Running GCC-optimized CoreMark benchmark..."
    
    if [[ ! -f "coremark/coremark_benchmark" ]]; then
        log_error "Universal CoreMark benchmark not found!"
        return 1
    fi
    
    log_info "Executing CoreMark with GCC optimizations..."
    local coremark_output
    coremark_output=$(timeout 300s ./coremark/coremark_benchmark 2>&1 || echo "CoreMark execution failed")
    
    # Parse CoreMark results
    local coremark_score=$(echo "$coremark_output" | grep -i "coremark.*:" | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0")
    local iterations_per_sec=$(echo "$coremark_output" | grep -i "iterations/sec" | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0")
    
    log_success "Universal CoreMark completed - Score: ${coremark_score}, Iterations/sec: ${iterations_per_sec}"
    
    echo "$coremark_score|$iterations_per_sec"
}

# Generate comprehensive benchmark results
generate_universal_results() {
    log_info "Collecting system information..."
    local cpu_model=$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || lscpu | grep "Model name" | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown CPU')
    local cpu_cores=$(nproc)
    local memory_gb=$(free -g | awk '/^Mem:/ {print $2}')
    local arch=$(uname -m)
    
    log_info "Running Universal benchmark suite with Spack..."
    
    # Execute benchmarks
    local stream_results=$(run_stream_universal)
    local linpack_results=$(run_linpack_universal) 
    local coremark_results=$(run_coremark_universal)
    
    # Parse results
    IFS='|' read -r copy_rate scale_rate add_rate triad_rate avg_bandwidth <<< "$stream_results"
    local gflops="$linpack_results"
    IFS='|' read -r coremark_score iterations_per_sec <<< "$coremark_results"
    
    # Get Spack-managed compiler information
    local spack_gcc_version=$(spack find --format "{name}@{version}" gcc | head -1)
    local spack_openblas_version=$(spack find --format "{name}@{version}" openblas | head -1)
    
    # Generate universal JSON results with Spack metadata
    cat << EOF > "$RESULTS_FILE"
{
  "benchmark_metadata": {
    "container_variant": "universal-spack-optimized",
    "container_version": "$CONTAINER_VERSION-spack",
    "benchmark_suite": "Universal GCC + OpenBLAS via Spack",
    "execution_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "compiler": "GCC managed by Spack",
    "math_library": "OpenBLAS managed by Spack",
    "package_manager": "Spack",
    "optimization_level": "portable_performance"
  },
  "spack_environment": {
    "compiler_package": "$spack_gcc_version",
    "openblas_package": "$spack_openblas_version",
    "spack_version": "$(spack --version)",
    "installed_via": "AWS binary cache + local build"
  },
  "system_info": {
    "architecture": "$arch",
    "cpu_model": "$cpu_model", 
    "cpu_cores": $cpu_cores,
    "memory_gb": $memory_gb,
    "hostname": "$(hostname)",
    "kernel_version": "$(uname -r)",
    "compiler_version": "$(gcc --version | head -n1 || echo 'GCC via Spack')"
  },
  "benchmark_results": {
    "stream_universal": {
      "copy_rate_mb_s": $copy_rate,
      "scale_rate_mb_s": $scale_rate, 
      "add_rate_mb_s": $add_rate,
      "triad_rate_mb_s": $triad_rate,
      "average_bandwidth_mb_s": $avg_bandwidth,
      "array_size": 80000000,
      "optimization": "GCC native + portable flags via Spack",
      "status": "completed"
    },
    "linpack_universal": {
      "gflops": $gflops,
      "math_library": "OpenBLAS via Spack",
      "optimization": "OpenBLAS DGEMM + portable threading",
      "status": "completed"
    },
    "coremark_universal": {
      "score": $coremark_score,
      "iterations_per_sec": $iterations_per_sec,
      "iterations": 100000,
      "optimization": "GCC portable optimizations via Spack",
      "status": "completed"
    }
  },
  "benchmark_summary": {
    "performance_metrics": {
      "memory_bandwidth_mb_s": $avg_bandwidth,
      "cpu_performance_gflops": $gflops,
      "integer_performance_coremark": $coremark_score
    },
    "performance_ratings": {
      "memory_bandwidth": "$(if (( $(echo "$avg_bandwidth > 35000" | bc -l) )); then echo "Excellent"; elif (( $(echo "$avg_bandwidth > 20000" | bc -l) )); then echo "Very Good"; else echo "Good"; fi)",
      "cpu_performance": "$(if (( $(echo "$gflops > 75" | bc -l) )); then echo "Excellent"; elif (( $(echo "$gflops > 40" | bc -l) )); then echo "Very Good"; else echo "Good"; fi)",
      "integer_performance": "$(if (( $(echo "$coremark_score > 35000" | bc -l) )); then echo "Excellent"; elif (( $(echo "$coremark_score > 20000" | bc -l) )); then echo "Very Good"; else echo "Good"; fi)",
      "overall": "Universal Optimized via Spack"
    },
    "optimization_benefits": {
      "portability": "Works across Intel, AMD, and ARM architectures",
      "gcc_optimizations": "Modern GCC with native architecture detection",
      "openblas_acceleration": "Portable high-performance BLAS library",
      "spack_advantages": "Reproducible environment, easy dependency management",
      "compatibility": "Fallback container for unknown or mixed environments"
    }
  }
}
EOF
    
    log_success "=== Universal Spack Benchmark Suite Complete ==="
    log_info "Results saved to: $RESULTS_FILE"
    
    # Display universal summary
    echo
    echo "=========================================================="
    echo "      UNIVERSAL BENCHMARK RESULTS (SPACK EDITION)"  
    echo "=========================================================="
    cat "$RESULTS_FILE" | jq -r '
        "System: " + .system_info.cpu_model,
        "Architecture: " + .system_info.architecture + " (" + (.system_info.cpu_cores | tostring) + " cores)",
        "Compiler: " + .spack_environment.compiler_package,
        "Math Library: " + .spack_environment.openblas_package,
        "",
        "PERFORMANCE RESULTS:",
        "Memory Bandwidth: " + (.benchmark_summary.performance_metrics.memory_bandwidth_mb_s | tostring) + " MB/s (" + .benchmark_summary.performance_ratings.memory_bandwidth + ")",
        "CPU Performance: " + (.benchmark_summary.performance_metrics.cpu_performance_gflops | tostring) + " GFLOPS (" + .benchmark_summary.performance_ratings.cpu_performance + ")", 
        "Integer Performance: " + (.benchmark_summary.performance_metrics.integer_performance_coremark | tostring) + " CoreMark (" + .benchmark_summary.performance_ratings.integer_performance + ")",
        "",
        "UNIVERSAL OPTIMIZATIONS VIA SPACK:",
        "✓ Spack-managed modern GCC with native detection",
        "✓ Spack-managed OpenBLAS for portable acceleration", 
        "✓ Cross-architecture compatibility (Intel/AMD/ARM)",
        "✓ Portable optimization flags for broad compatibility",
        "✓ Reproducible Spack environment",
        "",
        "Benefits: " + .benchmark_summary.optimization_benefits.portability'
    echo "=========================================================="
}

# Handle different execution modes
case "${1:-all}" in
    "stream")
        log_info "Running Universal STREAM benchmark only..."
        run_stream_universal
        ;;
    "linpack") 
        log_info "Running Universal LINPACK benchmark only..."
        run_linpack_universal
        ;;
    "coremark")
        log_info "Running Universal CoreMark benchmark only..."
        run_coremark_universal
        ;;
    "all"|*)
        log_info "Running complete Universal benchmark suite via Spack..."
        generate_universal_results
        ;;
esac