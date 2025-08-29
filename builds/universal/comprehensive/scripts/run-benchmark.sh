#!/bin/bash

# Unified Benchmark Execution Script
# Runs individual benchmarks or complete suite with comprehensive JSON output
# Supports STREAM, LINPACK, and CoreMark with Cloud Compass integration

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[BENCHMARK]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Global variables
BENCHMARK_DIR="/opt/benchmarks"
OUTPUT_DIR="/tmp/benchmark_results"
TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
RESULTS_FILE="$OUTPUT_DIR/benchmark_results_$TIMESTAMP.json"

# System information collection
collect_system_info() {
    local cpu_model=$(grep -m1 "model name" /proc/cpuinfo 2>/dev/null | cut -d: -f2 | sed 's/^ *//' || echo "Unknown")
    local cpu_arch=$(uname -m)
    local memory_gb=$(free -g | awk 'NR==2{printf "%.1f", $2}')
    local cpu_cores=$(nproc)
    local cpu_freq=$(lscpu | grep "CPU max MHz" | awk '{print $4}' 2>/dev/null || echo "Unknown")
    
    cat << EOF
{
  "system_info": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "hostname": "$(hostname)",
    "architecture": "$cpu_arch",
    "cpu_model": "$cpu_model",
    "cpu_cores": $cpu_cores,
    "cpu_frequency_mhz": "$cpu_freq",
    "memory_gb": $memory_gb,
    "kernel": "$(uname -r)",
    "os_release": "$(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"')"
  }
}
EOF
}

# AWS instance metadata collection
collect_aws_metadata() {
    local instance_id=""
    local instance_type=""
    local availability_zone=""
    local region=""
    
    # Try to get AWS metadata (will fail if not on EC2)
    if curl -s --max-time 2 http://169.254.169.254/latest/meta-data/instance-id >/dev/null 2>&1; then
        instance_id=$(curl -s --max-time 2 http://169.254.169.254/latest/meta-data/instance-id 2>/dev/null || echo "unknown")
        instance_type=$(curl -s --max-time 2 http://169.254.169.254/latest/meta-data/instance-type 2>/dev/null || echo "unknown")
        availability_zone=$(curl -s --max-time 2 http://169.254.169.254/latest/meta-data/placement/availability-zone 2>/dev/null || echo "unknown")
        region=$(echo "$availability_zone" | sed 's/[a-z]$//' 2>/dev/null || echo "unknown")
    fi
    
    cat << EOF
{
  "aws_metadata": {
    "instance_id": "$instance_id",
    "instance_type": "$instance_type",
    "availability_zone": "$availability_zone",
    "region": "$region",
    "is_aws_instance": $([ -n "$instance_id" ] && echo "true" || echo "false")
  }
}
EOF
}

# Run STREAM memory bandwidth benchmark
run_stream_benchmark() {
    log_info "Running STREAM memory bandwidth benchmark..."
    
    local stream_executable="$BENCHMARK_DIR/stream/stream_benchmark"
    local stream_output=""
    local stream_results=""
    
    if [[ ! -x "$stream_executable" ]]; then
        log_error "STREAM benchmark not found or not executable: $stream_executable"
        return 1
    fi
    
    # Capture STREAM output
    stream_output=$(timeout 300s "$stream_executable" 2>&1 || echo "STREAM execution failed")
    
    # Parse STREAM results (handle potential parsing errors)
    local copy_rate=$(echo "$stream_output" | grep "Copy:" | awk '{print $2}' 2>/dev/null || echo "0")
    local scale_rate=$(echo "$stream_output" | grep "Scale:" | awk '{print $2}' 2>/dev/null || echo "0")
    local add_rate=$(echo "$stream_output" | grep "Add:" | awk '{print $2}' 2>/dev/null || echo "0")
    local triad_rate=$(echo "$stream_output" | grep "Triad:" | awk '{print $2}' 2>/dev/null || echo "0")
    
    # Ensure numeric values (default to 0 if parsing fails)
    copy_rate=$(echo "$copy_rate" | grep -E '^[0-9]+(\.[0-9]+)?$' || echo "0")
    scale_rate=$(echo "$scale_rate" | grep -E '^[0-9]+(\.[0-9]+)?$' || echo "0")
    add_rate=$(echo "$add_rate" | grep -E '^[0-9]+(\.[0-9]+)?$' || echo "0")
    triad_rate=$(echo "$triad_rate" | grep -E '^[0-9]+(\.[0-9]+)?$' || echo "0")
    
    # Extract additional metrics
    local array_size=$(echo "$stream_output" | grep "Array size" | awk -F'= ' '{print $2}' | awk '{print $1}' || echo "0")
    local memory_per_array=$(echo "$stream_output" | grep "Memory per array" | awk '{print $4}' || echo "0")
    local total_memory=$(echo "$stream_output" | grep "Total memory required" | awk '{print $4}' || echo "0")
    local validation=$(echo "$stream_output" | grep -o "Solution Validates\|Failed Validation" || echo "Unknown")
    
    # Calculate best rate safely
    local best_rate=$(echo "$copy_rate $scale_rate $add_rate $triad_rate" | tr ' ' '\n' | sort -nr | head -1)
    best_rate=$(echo "$best_rate" | grep -E '^[0-9]+(\.[0-9]+)?$' || echo "0")
    
    # Escape JSON output safely
    local raw_output_json
    if command -v jq >/dev/null 2>&1; then
        raw_output_json=$(echo "$stream_output" | jq -Rs . 2>/dev/null || echo '"STREAM output parsing failed"')
    else
        raw_output_json='"jq not available"'
    fi
    
    stream_results=$(cat << EOF
{
  "stream_benchmark": {
    "status": "completed",
    "execution_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "results": {
      "copy_rate_mb_s": $copy_rate,
      "scale_rate_mb_s": $scale_rate,
      "add_rate_mb_s": $add_rate,
      "triad_rate_mb_s": $triad_rate,
      "best_rate_mb_s": $best_rate,
      "array_size_elements": "$array_size",
      "memory_per_array_mb": "$memory_per_array",
      "total_memory_mb": "$total_memory",
      "validation": "$validation"
    },
    "raw_output": $raw_output_json
  }
}
EOF
)
    
    echo "$stream_results"
    log_success "STREAM benchmark completed"
}

# Run LINPACK CPU performance benchmark
run_linpack_benchmark() {
    log_info "Running LINPACK CPU performance benchmark..."
    
    local linpack_executable="$BENCHMARK_DIR/linpack/linpack_benchmark"
    local linpack_output=""
    local linpack_results=""
    
    if [[ ! -x "$linpack_executable" ]]; then
        log_error "LINPACK benchmark not found or not executable: $linpack_executable"
        return 1
    fi
    
    # Capture LINPACK output
    linpack_output=$(timeout 600s "$linpack_executable" 2>&1 || echo "LINPACK execution failed")
    
    # Parse LINPACK results
    local matrix_size=$(echo "$linpack_output" | grep "Matrix size:" | awk '{print $3}' || echo "0")
    local time_seconds=$(echo "$linpack_output" | grep "Time (seconds):" | awk '{print $3}' || echo "0")
    local gflops=$(echo "$linpack_output" | grep "GFLOPS:" | awk '{print $2}' || echo "0")
    local mflops=$(echo "$linpack_output" | grep "MFLOPS:" | awk '{print $2}' || echo "0")
    local residual=$(echo "$linpack_output" | grep "Residual:" | awk '{print $2}' || echo "0")
    local solution_status=$(echo "$linpack_output" | grep "Solution:" | awk '{print $2}' || echo "Unknown")
    local memory_mb=$(echo "$linpack_output" | grep "Memory used:" | awk '{print $3}' || echo "0")
    
    linpack_results=$(cat << EOF
{
  "linpack_benchmark": {
    "status": "completed", 
    "execution_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "results": {
      "matrix_size": $matrix_size,
      "execution_time_seconds": $time_seconds,
      "gflops": $gflops,
      "mflops": $mflops,
      "residual": "$residual",
      "solution_status": "$solution_status",
      "memory_used_mb": $memory_mb
    },
    "raw_output": $(echo "$linpack_output" | jq -Rs .)
  }
}
EOF
)
    
    echo "$linpack_results"
    log_success "LINPACK benchmark completed"
}

# Run CoreMark integer performance benchmark
run_coremark_benchmark() {
    log_info "Running CoreMark integer performance benchmark..."
    
    local coremark_executable="$BENCHMARK_DIR/coremark/coremark_benchmark"
    local coremark_output=""
    local coremark_results=""
    
    if [[ ! -x "$coremark_executable" ]]; then
        log_error "CoreMark benchmark not found or not executable: $coremark_executable"
        return 1
    fi
    
    # Capture CoreMark output
    coremark_output=$(timeout 300s "$coremark_executable" 2>&1 || echo "CoreMark execution failed")
    
    # Parse CoreMark results
    local iterations=$(echo "$coremark_output" | grep "Iterations:" | awk '{print $2}' || echo "0")
    local total_time=$(echo "$coremark_output" | grep "Total time (sec):" | awk '{print $4}' || echo "0")
    local iterations_per_sec=$(echo "$coremark_output" | grep "Iterations per second:" | awk '{print $4}' || echo "0")
    local coremark_score=$(echo "$coremark_output" | grep "CoreMark Score:" | awk '{print $3}' || echo "0")
    local coremark_per_mhz=$(echo "$coremark_output" | grep "CoreMark/MHz:" | awk '{print $2}' || echo "0")
    local validation=$(echo "$coremark_output" | grep "Validation:" | awk '{print $2}' || echo "Unknown")
    local performance_rating=$(echo "$coremark_output" | grep "Performance Rating:" | awk '{print $3}' || echo "Unknown")
    
    coremark_results=$(cat << EOF
{
  "coremark_benchmark": {
    "status": "completed",
    "execution_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "results": {
      "iterations": $iterations,
      "execution_time_seconds": $total_time,
      "iterations_per_second": $iterations_per_sec,
      "coremark_score": $coremark_score,
      "coremark_per_mhz": $coremark_per_mhz,
      "performance_rating": "$performance_rating",
      "validation": "$validation"
    },
    "raw_output": $(echo "$coremark_output" | jq -Rs .)
  }
}
EOF
)
    
    echo "$coremark_results"
    log_success "CoreMark benchmark completed"
}

# Load version manifest information
load_version_manifest() {
    local manifest_file="$BENCHMARK_DIR/version-manifest.json"
    local version_info="{}"
    
    if [[ -f "$manifest_file" ]]; then
        version_info=$(cat "$manifest_file" 2>/dev/null || echo "{}")
    else
        log_warning "Version manifest not found, generating minimal version info"
        # Generate basic version info if manifest missing
        version_info=$(cat << EOF
{
  "container_info": {
    "container_version": "2.0.0",
    "container_variant": "universal",
    "build_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  },
  "benchmark_versions": {
    "stream": {"version": "5.10"},
    "linpack": {"implementation": "custom", "version": "1.0"}, 
    "coremark": {"version": "1.0"}
  },
  "compiler_versions": {
    "primary_compiler": {
      "name": "gcc",
      "version": "$(gcc --version 2>/dev/null | head -n1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo 'unknown')"
    }
  }
}
EOF
)
    fi
    
    echo "$version_info"
}

# Generate comprehensive results summary
generate_summary() {
    local stream_result="$1"
    local linpack_result="$2"
    local coremark_result="$3"
    
    # Extract key metrics for summary
    local stream_best=$(echo "$stream_result" | jq -r '.stream_benchmark.results.best_rate_mb_s // 0' 2>/dev/null || echo "0")
    local linpack_gflops=$(echo "$linpack_result" | jq -r '.linpack_benchmark.results.gflops // 0' 2>/dev/null || echo "0")
    local coremark_score=$(echo "$coremark_result" | jq -r '.coremark_benchmark.results.coremark_score // 0' 2>/dev/null || echo "0")
    
    # Calculate composite performance score (weighted average)
    local memory_weight=0.3
    local cpu_weight=0.4
    local integer_weight=0.3
    
    # Normalize scores (rough approximation)
    local norm_memory=$(echo "$stream_best / 100000" | bc -l 2>/dev/null || echo "0")
    local norm_cpu=$(echo "$linpack_gflops / 500" | bc -l 2>/dev/null || echo "0")
    local norm_integer=$(echo "$coremark_score / 50000" | bc -l 2>/dev/null || echo "0")
    
    local composite_score=$(echo "$norm_memory * $memory_weight + $norm_cpu * $cpu_weight + $norm_integer * $integer_weight" | bc -l 2>/dev/null || echo "0")
    
    cat << EOF
{
  "benchmark_summary": {
    "execution_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "version": "2.0.0",
    "schema_version": "2.0",
    "performance_metrics": {
      "memory_bandwidth_mb_s": $stream_best,
      "cpu_performance_gflops": $linpack_gflops,
      "integer_performance_coremark": $coremark_score,
      "composite_performance_score": $(printf "%.3f" "$composite_score")
    },
    "performance_ratings": {
      "memory": $([ $(echo "$stream_best > 50000" | bc -l) -eq 1 ] && echo "\"Excellent\"" || echo "\"Good\""),
      "cpu": $([ $(echo "$linpack_gflops > 100" | bc -l) -eq 1 ] && echo "\"Excellent\"" || echo "\"Good\""),
      "integer": $([ $(echo "$coremark_score > 30000" | bc -l) -eq 1 ] && echo "\"Excellent\"" || echo "\"Good\""),
      "overall": $([ $(echo "$composite_score > 0.7" | bc -l) -eq 1 ] && echo "\"Excellent\"" || [ $(echo "$composite_score > 0.5" | bc -l) -eq 1 ] && echo "\"Good\"" || echo "\"Fair\"")
    },
    "benchmark_status": {
      "stream_completed": $(echo "$stream_result" | jq -r '.stream_benchmark.status == "completed"' 2>/dev/null || echo "false"),
      "linpack_completed": $(echo "$linpack_result" | jq -r '.linpack_benchmark.status == "completed"' 2>/dev/null || echo "false"),
      "coremark_completed": $(echo "$coremark_result" | jq -r '.coremark_benchmark.status == "completed"' 2>/dev/null || echo "false")
    }
  }
}
EOF
}

# Main benchmark execution
run_all_benchmarks() {
    log_info "=== Running Complete Benchmark Suite ==="
    
    # Ensure output directory exists
    mkdir -p "$OUTPUT_DIR"
    
    # Collect system information
    log_info "Collecting system information..."
    
    # Run individual benchmarks with error handling
    log_info "Running STREAM benchmark..."
    local stream_output=$(timeout 300s "$BENCHMARK_DIR/stream/stream_benchmark" 2>&1 || echo "STREAM execution failed")
    local copy_rate=$(echo "$stream_output" | grep "Copy:" | awk '{print $2}' 2>/dev/null || echo "0")
    local scale_rate=$(echo "$stream_output" | grep "Scale:" | awk '{print $2}' 2>/dev/null || echo "0") 
    local add_rate=$(echo "$stream_output" | grep "Add:" | awk '{print $2}' 2>/dev/null || echo "0")
    local triad_rate=$(echo "$stream_output" | grep "Triad:" | awk '{print $2}' 2>/dev/null || echo "0")
    
    # Get best rate
    local best_rate=$(echo "$copy_rate $scale_rate $add_rate $triad_rate" | tr ' ' '\n' | sort -nr | head -1)
    
    log_info "Running LINPACK benchmark..."
    local linpack_output=$(timeout 600s "$BENCHMARK_DIR/linpack/linpack_benchmark" 2>&1 || echo "LINPACK execution failed")
    local gflops=$(echo "$linpack_output" | grep "GFLOPS:" | awk '{print $2}' 2>/dev/null || echo "0")
    
    log_info "Running CoreMark benchmark..."
    local coremark_output=$(timeout 300s "$BENCHMARK_DIR/coremark/coremark_benchmark" 2>&1 || echo "CoreMark execution failed")
    local coremark_score=$(echo "$coremark_output" | grep "CoreMark Score:" | awk '{print $3}' 2>/dev/null || echo "0")
    
    # Generate simplified JSON results
    cat > "$RESULTS_FILE" << EOF
{
  "benchmark_results": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "container_version": "2.0.0",
    "stream": {
      "copy_rate_mb_s": $copy_rate,
      "scale_rate_mb_s": $scale_rate,
      "add_rate_mb_s": $add_rate,
      "triad_rate_mb_s": $triad_rate,
      "best_rate_mb_s": $best_rate,
      "status": "completed"
    },
    "linpack": {
      "gflops": $gflops,
      "status": "completed"
    },
    "coremark": {
      "score": $coremark_score,
      "status": "completed"
    },
    "system_info": {
      "architecture": "$(uname -m)",
      "cpu_model": "$(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed 's/^ *//' || echo 'Unknown')",
      "cpu_cores": $(nproc),
      "hostname": "$(hostname)"
    }
  }
}
EOF
    
    log_success "=== Benchmark Suite Complete ==="
    log_info "Results saved to: $RESULTS_FILE"
    
    # Display summary to stdout
    echo
    echo "==============================================="
    echo "           BENCHMARK RESULTS SUMMARY"
    echo "==============================================="
    cat "$RESULTS_FILE" | jq -r '
        "System: " + .system_info.cpu_model,
        "Architecture: " + .system_info.architecture + " (" + (.system_info.cpu_cores | tostring) + " cores)",
        "Memory Bandwidth: " + (.benchmark_summary.performance_metrics.memory_bandwidth_mb_s | tostring) + " MB/s",
        "CPU Performance: " + (.benchmark_summary.performance_metrics.cpu_performance_gflops | tostring) + " GFLOPS", 
        "Integer Performance: " + (.benchmark_summary.performance_metrics.integer_performance_coremark | tostring) + " CoreMark",
        "Overall Rating: " + .benchmark_summary.performance_ratings.overall'
    echo "==============================================="
}

# Handle different execution modes
case "${1:-all}" in
    "stream")
        log_info "Running STREAM benchmark..."
        stream_output=$(timeout 300s "$BENCHMARK_DIR/stream/stream_benchmark" 2>&1 || echo "STREAM execution failed")
        copy_rate=$(echo "$stream_output" | grep "Copy:" | awk '{print $2}' 2>/dev/null || echo "0")
        echo "{\"stream\": {\"copy_rate_mb_s\": $copy_rate, \"status\": \"completed\"}}"
        ;;
    "linpack") 
        log_info "Running LINPACK benchmark..."
        linpack_output=$(timeout 600s "$BENCHMARK_DIR/linpack/linpack_benchmark" 2>&1 || echo "LINPACK execution failed")
        gflops=$(echo "$linpack_output" | grep "GFLOPS:" | awk '{print $2}' 2>/dev/null || echo "0")
        echo "{\"linpack\": {\"gflops\": $gflops, \"status\": \"completed\"}}"
        ;;
    "coremark")
        log_info "Running CoreMark benchmark..."
        coremark_output=$(timeout 300s "$BENCHMARK_DIR/coremark/coremark_benchmark" 2>&1 || echo "CoreMark execution failed")
        coremark_score=$(echo "$coremark_output" | grep "CoreMark Score:" | awk '{print $3}' 2>/dev/null || echo "0")
        echo "{\"coremark\": {\"score\": $coremark_score, \"status\": \"completed\"}}"
        ;;
    "all"|*)
        cd "$BENCHMARK_DIR"
        run_all_benchmarks
        ;;
esac