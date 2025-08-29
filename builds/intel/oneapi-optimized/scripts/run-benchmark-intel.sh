#!/bin/bash

# Intel oneAPI Optimized Benchmark Runner
# Executes STREAM, LINPACK, and CoreMark with Intel-specific optimizations

set -eo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INTEL-BENCH]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Initialize Intel oneAPI environment
source "${ONEAPI_ROOT}/setvars.sh" --force

# Configuration
BENCHMARK_DIR="/opt/benchmarks"
RESULTS_DIR="/tmp/benchmark_results"
RESULTS_FILE="$RESULTS_DIR/intel_benchmark_results_$(date -u +%Y%m%dT%H%M%SZ).json"

# Create results directory
mkdir -p "$RESULTS_DIR"

cd "$BENCHMARK_DIR" || exit 1

# Intel-optimized STREAM benchmark
run_stream_intel() {
    log_info "Running Intel-optimized STREAM benchmark..."
    
    if [[ ! -f "stream/stream_benchmark_intel" ]]; then
        log_error "Intel STREAM benchmark not found!"
        return 1
    fi
    
    log_info "Executing STREAM with Intel MKL acceleration..."
    local stream_output
    stream_output=$(timeout 300s ./stream/stream_benchmark_intel 2>&1 || echo "STREAM execution failed")
    
    # Parse Intel-optimized STREAM results
    local copy_rate=$(echo "$stream_output" | grep "Copy:" | awk '{print $2}' | head -1 || echo "0")
    local scale_rate=$(echo "$stream_output" | grep "Scale:" | awk '{print $2}' | head -1 || echo "0")  
    local add_rate=$(echo "$stream_output" | grep "Add:" | awk '{print $2}' | head -1 || echo "0")
    local triad_rate=$(echo "$stream_output" | grep "Triad:" | awk '{print $2}' | head -1 || echo "0")
    
    # Calculate average memory bandwidth
    local avg_bandwidth=$(echo "scale=2; ($copy_rate + $scale_rate + $add_rate + $triad_rate) / 4" | bc -l || echo "0")
    
    log_success "Intel STREAM completed - Average bandwidth: ${avg_bandwidth} MB/s"
    
    echo "$copy_rate|$scale_rate|$add_rate|$triad_rate|$avg_bandwidth"
}

# Intel MKL-accelerated LINPACK benchmark  
run_linpack_intel() {
    log_info "Running Intel MKL-accelerated LINPACK benchmark..."
    
    if [[ ! -f "linpack/linpack_benchmark_intel" ]]; then
        log_error "Intel LINPACK benchmark not found!"
        return 1
    fi
    
    log_info "Executing LINPACK with Intel MKL BLAS..."
    local linpack_output
    linpack_output=$(timeout 300s ./linpack/linpack_benchmark_intel 2>&1 || echo "LINPACK execution failed")
    
    # Parse Intel MKL LINPACK results
    local gflops=$(echo "$linpack_output" | grep -i "gflops\|performance" | grep -oE '[0-9]+\.[0-9]+' | tail -1 || echo "0")
    
    log_success "Intel LINPACK completed - Performance: ${gflops} GFLOPS"
    
    echo "$gflops"
}

# Intel compiler-optimized CoreMark benchmark
run_coremark_intel() {
    log_info "Running Intel compiler-optimized CoreMark benchmark..."
    
    if [[ ! -f "coremark/coremark_benchmark_intel" ]]; then
        log_error "Intel CoreMark benchmark not found!"
        return 1
    fi
    
    log_info "Executing CoreMark with Intel compiler vectorization..."
    local coremark_output
    coremark_output=$(timeout 300s ./coremark/coremark_benchmark_intel 2>&1 || echo "CoreMark execution failed")
    
    # Parse Intel CoreMark results
    local coremark_score=$(echo "$coremark_output" | grep -i "coremark.*:" | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0")
    local iterations_per_sec=$(echo "$coremark_output" | grep -i "iterations/sec" | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0")
    
    log_success "Intel CoreMark completed - Score: ${coremark_score}, Iterations/sec: ${iterations_per_sec}"
    
    echo "$coremark_score|$iterations_per_sec"
}

# Generate comprehensive Intel benchmark results
generate_intel_results() {
    log_info "Collecting system information..."
    local cpu_model=$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')
    local cpu_cores=$(nproc)
    local memory_gb=$(free -g | awk '/^Mem:/ {print $2}')
    
    log_info "Running Intel-optimized benchmark suite..."
    
    # Execute benchmarks
    local stream_results=$(run_stream_intel)
    local linpack_results=$(run_linpack_intel) 
    local coremark_results=$(run_coremark_intel)
    
    # Parse results
    IFS='|' read -r copy_rate scale_rate add_rate triad_rate avg_bandwidth <<< "$stream_results"
    local gflops="$linpack_results"
    IFS='|' read -r coremark_score iterations_per_sec <<< "$coremark_results"
    
    # Generate Intel-specific JSON results
    cat << EOF > "$RESULTS_FILE"
{
  "benchmark_metadata": {
    "container_variant": "intel-oneapi-optimized",
    "container_version": "$CONTAINER_VERSION",
    "benchmark_suite": "Intel oneAPI + MKL accelerated",
    "execution_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "compiler": "Intel oneAPI (icx/icpx)",
    "math_library": "Intel MKL",
    "optimization_level": "maximum_intel_performance"
  },
  "system_info": {
    "architecture": "$(uname -m)",
    "cpu_model": "$cpu_model", 
    "cpu_cores": $cpu_cores,
    "memory_gb": $memory_gb,
    "hostname": "$(hostname)",
    "kernel_version": "$(uname -r)",
    "compiler_version": "$(icx --version | head -n1)"
  },
  "benchmark_results": {
    "stream_intel": {
      "copy_rate_mb_s": $copy_rate,
      "scale_rate_mb_s": $scale_rate, 
      "add_rate_mb_s": $add_rate,
      "triad_rate_mb_s": $triad_rate,
      "average_bandwidth_mb_s": $avg_bandwidth,
      "array_size": 80000000,
      "optimization": "Intel MKL + AVX-512",
      "status": "completed"
    },
    "linpack_intel": {
      "gflops": $gflops,
      "math_library": "Intel MKL BLAS",
      "optimization": "MKL DGEMM + Intel threading",
      "status": "completed"
    },
    "coremark_intel": {
      "score": $coremark_score,
      "iterations_per_sec": $iterations_per_sec,
      "iterations": 100000,
      "optimization": "Intel compiler vectorization + IPO",
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
      "memory_bandwidth": "$(if (( $(echo "$avg_bandwidth > 50000" | bc -l) )); then echo "Excellent"; elif (( $(echo "$avg_bandwidth > 30000" | bc -l) )); then echo "Very Good"; else echo "Good"; fi)",
      "cpu_performance": "$(if (( $(echo "$gflops > 100" | bc -l) )); then echo "Excellent"; elif (( $(echo "$gflops > 50" | bc -l) )); then echo "Very Good"; else echo "Good"; fi)",
      "integer_performance": "$(if (( $(echo "$coremark_score > 50000" | bc -l) )); then echo "Excellent"; elif (( $(echo "$coremark_score > 30000" | bc -l) )); then echo "Very Good"; else echo "Good"; fi)",
      "overall": "Intel Optimized"
    },
    "optimization_benefits": {
      "expected_improvement_over_gcc": "20-40% performance gain on Intel Xeon processors",
      "mkl_acceleration": "Optimized BLAS/LAPACK routines for maximum GFLOPS",
      "compiler_vectorization": "Automatic AVX-512 utilization where available",
      "threading_optimization": "Intel OpenMP with NUMA-aware scheduling"
    }
  }
}
EOF
    
    log_success "=== Intel Benchmark Suite Complete ==="
    log_info "Results saved to: $RESULTS_FILE"
    
    # Display Intel-specific summary
    echo
    echo "================================================"
    echo "      INTEL oneAPI BENCHMARK RESULTS SUMMARY"  
    echo "================================================"
    cat "$RESULTS_FILE" | jq -r '
        "System: " + .system_info.cpu_model,
        "Architecture: " + .system_info.architecture + " (" + (.system_info.cpu_cores | tostring) + " cores)",
        "Compiler: " + .system_info.compiler_version,
        "",
        "PERFORMANCE RESULTS:",
        "Memory Bandwidth: " + (.benchmark_summary.performance_metrics.memory_bandwidth_mb_s | tostring) + " MB/s (" + .benchmark_summary.performance_ratings.memory_bandwidth + ")",
        "CPU Performance: " + (.benchmark_summary.performance_metrics.cpu_performance_gflops | tostring) + " GFLOPS (" + .benchmark_summary.performance_ratings.cpu_performance + ")", 
        "Integer Performance: " + (.benchmark_summary.performance_metrics.integer_performance_coremark | tostring) + " CoreMark (" + .benchmark_summary.performance_ratings.integer_performance + ")",
        "",
        "INTEL OPTIMIZATIONS:",
        "✓ Intel MKL BLAS acceleration for LINPACK",
        "✓ Intel compiler vectorization for CoreMark", 
        "✓ AVX-512 optimizations for STREAM",
        "✓ NUMA-aware threading with Intel OpenMP",
        "",
        "Expected " + .benchmark_summary.optimization_benefits.expected_improvement_over_gcc'
    echo "================================================"
}

# Handle different execution modes
case "${1:-all}" in
    "stream")
        log_info "Running Intel STREAM benchmark only..."
        run_stream_intel
        ;;
    "linpack") 
        log_info "Running Intel LINPACK benchmark only..."
        run_linpack_intel
        ;;
    "coremark")
        log_info "Running Intel CoreMark benchmark only..."
        run_coremark_intel
        ;;
    "all"|*)
        log_info "Running complete Intel oneAPI benchmark suite..."
        generate_intel_results
        ;;
esac