#!/bin/bash

# Community Benchmark Sharing Script
# This demonstrates how easy it is for others to pull and run our containers
# to compare performance across different systems

set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# System information gathering
gather_system_info() {
    echo "{"
    echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
    echo "  \"hostname\": \"$(hostname)\","
    echo "  \"kernel\": \"$(uname -r)\","
    echo "  \"architecture\": \"$(uname -m)\","
    
    # CPU information
    if command -v lscpu &> /dev/null; then
        local cpu_model=$(lscpu | grep "Model name" | cut -d: -f2 | xargs)
        local cpu_cores=$(nproc)
        local cpu_threads=$(lscpu | grep "CPU(s):" | head -1 | cut -d: -f2 | xargs)
        echo "  \"cpu\": {"
        echo "    \"model\": \"$cpu_model\","
        echo "    \"cores\": $cpu_cores,"
        echo "    \"threads\": $cpu_threads"
        echo "  },"
    fi
    
    # Memory information
    if command -v free &> /dev/null; then
        local memory_gb=$(free -g | awk '/^Mem:/{print $2}')
        echo "  \"memory_gb\": $memory_gb,"
    fi
    
    # Container runtime
    echo "  \"container_runtime\": \"$(podman --version 2>/dev/null || docker --version 2>/dev/null || echo 'unknown')\","
    
    # Cloud provider detection (if applicable)
    local cloud_provider="unknown"
    if curl -s --max-time 2 http://169.254.169.254/latest/meta-data/instance-type &> /dev/null; then
        cloud_provider="aws"
        local instance_type=$(curl -s http://169.254.169.254/latest/meta-data/instance-type)
        echo "  \"cloud_provider\": \"$cloud_provider\","
        echo "  \"instance_type\": \"$instance_type\","
    elif curl -s --max-time 2 -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/machine-type &> /dev/null; then
        cloud_provider="gcp"
        echo "  \"cloud_provider\": \"$cloud_provider\","
    elif curl -s --max-time 2 -H "Metadata: true" "http://169.254.169.254/metadata/instance?api-version=2021-02-01" &> /dev/null; then
        cloud_provider="azure"
        echo "  \"cloud_provider\": \"$cloud_provider\","
    else
        echo "  \"cloud_provider\": \"$cloud_provider\","
    fi
    
    echo "  \"benchmarks\": {"
}

# Run a single benchmark container
run_benchmark() {
    local container="$1"
    local architecture="$2"
    
    log_info "Running benchmark: $container ($architecture)"
    
    # Check if container exists locally
    if ! podman images | grep -q "$container"; then
        log_warning "Container $container not found locally"
        
        # In a real scenario, this would pull from a public registry
        log_info "In production, this would run: podman pull public.ecr.aws/aws-benchmarks/$container"
        return 1
    fi
    
    # Run the benchmark
    local start_time=$(date +%s)
    local benchmark_output
    
    if benchmark_output=$(timeout 60 podman run --rm "$container" 2>&1); then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        # Extract STREAM results (simplified parsing)
        local copy_rate=$(echo "$benchmark_output" | grep "Copy:" | awk '{print $2}')
        local scale_rate=$(echo "$benchmark_output" | grep "Scale:" | awk '{print $2}')
        local add_rate=$(echo "$benchmark_output" | grep "Add:" | awk '{print $2}')
        local triad_rate=$(echo "$benchmark_output" | grep "Triad:" | awk '{print $2}')
        
        echo "    \"$architecture\": {"
        echo "      \"container\": \"$container\","
        echo "      \"duration_seconds\": $duration,"
        echo "      \"results\": {"
        echo "        \"copy_mbps\": ${copy_rate:-null},"
        echo "        \"scale_mbps\": ${scale_rate:-null},"
        echo "        \"add_mbps\": ${add_rate:-null},"
        echo "        \"triad_mbps\": ${triad_rate:-null}"
        echo "      },"
        echo "      \"raw_output\": $(echo "$benchmark_output" | jq -Rsa . 2>/dev/null || echo "\"$benchmark_output\"")"
        echo "    },"
        
        log_success "Completed $architecture benchmark (${duration}s)"
    else
        log_error "Failed to run $architecture benchmark"
        echo "    \"$architecture\": {"
        echo "      \"container\": \"$container\","
        echo "      \"error\": \"Failed to execute benchmark\","
        echo "      \"raw_output\": $(echo "$benchmark_output" | jq -Rsa . 2>/dev/null || echo "\"$benchmark_output\"")"
        echo "    },"
    fi
}

# Generate a community contribution file
generate_contribution() {
    local output_file="community-benchmark-results-$(hostname)-$(date +%Y%m%d-%H%M%S).json"
    
    log_info "Generating community benchmark contribution: $output_file"
    
    {
        gather_system_info
        
        # Available containers to test
        local containers=(
            "localhost/aws-instance-benchmarks/stream:universal"
            "localhost/aws-instance-benchmarks/stream:intel-icelake"  
            "localhost/aws-instance-benchmarks/stream:amd-zen4"
            "localhost/aws-instance-benchmarks/stream:graviton3"
            "localhost/aws-instance-benchmarks/stream:graviton4"
        )
        
        local architectures=(
            "universal"
            "intel-icelake"
            "amd-zen4" 
            "graviton3"
            "graviton4"
        )
        
        # Run benchmarks for available containers
        for i in "${!containers[@]}"; do
            run_benchmark "${containers[$i]}" "${architectures[$i]}"
        done
        
        # Remove trailing comma and close JSON
        echo "  }"
        echo "}"
        
    } > "$output_file"
    
    # Clean up JSON formatting (remove trailing commas)
    if command -v jq &> /dev/null; then
        jq . "$output_file" > "${output_file}.tmp" && mv "${output_file}.tmp" "$output_file"
    fi
    
    log_success "Generated contribution file: $output_file"
    log_info "To contribute to the community database:"
    log_info "1. Review the results in: $output_file"
    log_info "2. Submit via GitHub PR to: https://github.com/scttfrdmn/aws-instance-benchmarks"
    log_info "3. Results will be included in the community performance database"
    
    echo
    log_info "This enables:"
    log_info "• Cross-validation of AWS performance claims"
    log_info "• Comparison between cloud and on-premises hardware"
    log_info "• Research reproducibility across institutions"
    log_info "• Open, transparent performance benchmarking"
}

# Usage information
usage() {
    cat << 'EOF'
Community Benchmark Sharing Script

This script demonstrates how anyone can:
1. Pull the same optimized containers used for AWS benchmarking
2. Run identical benchmarks on their own hardware
3. Generate standardized results for community comparison
4. Contribute to open, transparent performance data

Usage:
    ./community-benchmark.sh run          # Run benchmarks and generate results
    ./community-benchmark.sh info         # Show system information only
    ./community-benchmark.sh help         # Show this help

Examples of Community Use Cases:
    
    Research Lab: "Let's validate this AWS performance claim on our local cluster"
    $ ./community-benchmark.sh run
    
    Hardware Vendor: "Compare our new chip against published AWS results"  
    $ ./community-benchmark.sh run
    
    Academic Paper: "Reproduce benchmark results cited in this research"
    $ ./community-benchmark.sh run
    
    Open Source Project: "Test our optimization on different architectures"
    $ ./community-benchmark.sh run

The standardized containers ensure everyone runs identical tests with the same:
• Compiler versions and optimization flags  
• Benchmark parameters and configurations
• System detection and scaling algorithms
• Result formatting and validation

This creates a transparent, reproducible foundation for performance comparisons.

EOF
}

# Main execution
case "${1:-run}" in
    "run")
        generate_contribution
        ;;
    "info") 
        gather_system_info
        echo "  }"
        echo "}"
        ;;
    "help"|"-h"|"--help")
        usage
        ;;
    *)
        log_error "Unknown command: $1"
        usage
        exit 1
        ;;
esac