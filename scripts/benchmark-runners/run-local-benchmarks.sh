#!/bin/bash
# Local benchmark execution script for on-premises testing
# Provides reproducible benchmark execution outside of AWS

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
RESULTS_DIR="${BENCHMARK_RESULTS_DIR:-$PROJECT_ROOT/results}"
DATA_DIR="${BENCHMARK_DATA_DIR:-$PROJECT_ROOT/data}"
LOG_FILE="${RESULTS_DIR}/benchmark-execution-$(date +%Y%m%d-%H%M%S).log"

# Default settings
BENCHMARK_SUITE="${1:-all}"
INSTANCE_TYPE="${BENCHMARK_INSTANCE_TYPE:-local}"
PARALLEL_JOBS="${BENCHMARK_PARALLEL_JOBS:-1}"
DOCKER_MODE="${BENCHMARK_DOCKER_MODE:-true}"
UPLOAD_RESULTS="${BENCHMARK_UPLOAD_RESULTS:-false}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date -Iseconds)
    
    case "$level" in
        "INFO")  echo -e "${BLUE}[INFO]${NC}  $timestamp - $message" | tee -a "$LOG_FILE" ;;
        "WARN")  echo -e "${YELLOW}[WARN]${NC}  $timestamp - $message" | tee -a "$LOG_FILE" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $timestamp - $message" | tee -a "$LOG_FILE" ;;
        "SUCCESS") echo -e "${GREEN}[SUCCESS]${NC} $timestamp - $message" | tee -a "$LOG_FILE" ;;
    esac
}

# System information collection
collect_system_info() {
    log "INFO" "Collecting system information"
    
    local system_info_file="$RESULTS_DIR/system-info-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$system_info_file" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "instance_type": "$INSTANCE_TYPE",
  "benchmark_suite": "$BENCHMARK_SUITE",
  "system": {
    "os": "$(uname -s)",
    "kernel": "$(uname -r)",
    "architecture": "$(uname -m)",
    "hostname": "$(hostname)",
    "uptime": "$(uptime -p 2>/dev/null || uptime)"
  },
  "cpu": {
    "model": "$(grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)",
    "cores": "$(nproc)",
    "threads": "$(grep -c ^processor /proc/cpuinfo)",
    "frequency": "$(grep 'cpu MHz' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs) MHz"
  },
  "memory": {
    "total": "$(free -h | grep '^Mem:' | awk '{print $2}')",
    "available": "$(free -h | grep '^Mem:' | awk '{print $7}')",
    "swap": "$(free -h | grep '^Swap:' | awk '{print $2}')"
  },
  "storage": {
    "root_filesystem": "$(df -h / | tail -1 | awk '{print $2 " total, " $4 " available"}')"
  },
  "network": {
    "interfaces": $(ip -j addr show | jq '[.[] | select(.ifname != "lo") | {name: .ifname, state: .operstate}]')
  }
}
EOF

    # Add GPU information if nvidia-smi is available
    if command -v nvidia-smi &> /dev/null; then
        log "INFO" "GPU detected, adding GPU information"
        local gpu_info=$(nvidia-smi --query-gpu=name,memory.total,memory.free,driver_version --format=csv,noheader,nounits | head -1)
        local gpu_name=$(echo "$gpu_info" | cut -d, -f1 | xargs)
        local gpu_memory=$(echo "$gpu_info" | cut -d, -f2 | xargs)
        local gpu_driver=$(echo "$gpu_info" | cut -d, -f4 | xargs)
        
        jq --arg name "$gpu_name" --arg memory "$gpu_memory" --arg driver "$gpu_driver" \
           '.gpu = {name: $name, memory: ($memory + " MB"), driver: $driver}' \
           "$system_info_file" > "${system_info_file}.tmp" && \
           mv "${system_info_file}.tmp" "$system_info_file"
    fi
    
    log "SUCCESS" "System information saved to $system_info_file"
}

# Docker-based benchmark execution
run_docker_benchmarks() {
    local suite="$1"
    
    log "INFO" "Running Docker-based benchmarks for suite: $suite"
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        log "ERROR" "Docker is not installed. Please install Docker or set BENCHMARK_DOCKER_MODE=false"
        return 1
    fi
    
    # Check if Docker Compose is available
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log "ERROR" "Docker Compose is not installed. Please install Docker Compose"
        return 1
    fi
    
    # Navigate to project root for Docker context
    cd "$PROJECT_ROOT"
    
    # Set up environment variables for Docker Compose
    export BENCHMARK_RESULTS_DIR="$RESULTS_DIR"
    export BENCHMARK_DATA_DIR="$DATA_DIR"
    export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
    export COMPOSE_PROJECT_NAME="aws-benchmarks-$(date +%s)"
    
    # Run benchmarks based on suite
    case "$suite" in
        "genomics")
            log "INFO" "Running genomics benchmarks in Docker"
            docker-compose -f deployment/docker/docker-compose.yml --profile genomics up --build
            ;;
        "ml")
            log "INFO" "Running ML benchmarks in Docker"
            docker-compose -f deployment/docker/docker-compose.yml --profile ml up --build
            ;;
        "climate")
            log "INFO" "Running climate benchmarks in Docker"
            docker-compose -f deployment/docker/docker-compose.yml --profile climate up --build
            ;;
        "chemistry")
            log "INFO" "Running chemistry benchmarks in Docker"
            docker-compose -f deployment/docker/docker-compose.yml --profile chemistry up --build
            ;;
        "all")
            log "INFO" "Running all benchmark suites in Docker"
            docker-compose -f deployment/docker/docker-compose.yml --profile all up --build
            ;;
        *)
            log "ERROR" "Unknown benchmark suite: $suite"
            return 1
            ;;
    esac
    
    # Clean up containers
    docker-compose -f deployment/docker/docker-compose.yml down
    
    log "SUCCESS" "Docker benchmarks completed for suite: $suite"
}

# Native benchmark execution (without Docker)
run_native_benchmarks() {
    local suite="$1"
    
    log "INFO" "Running native benchmarks for suite: $suite"
    
    case "$suite" in
        "genomics")
            if [[ -f "$SCRIPT_DIR/run-genomics-benchmarks.sh" ]]; then
                bash "$SCRIPT_DIR/run-genomics-benchmarks.sh"
            else
                log "ERROR" "Genomics benchmark script not found"
                return 1
            fi
            ;;
        "ml")
            if [[ -f "$SCRIPT_DIR/run-ml-benchmarks.sh" ]]; then
                bash "$SCRIPT_DIR/run-ml-benchmarks.sh"
            else
                log "ERROR" "ML benchmark script not found"
                return 1
            fi
            ;;
        "all")
            log "INFO" "Running all available native benchmarks"
            for script in "$SCRIPT_DIR"/run-*-benchmarks.sh; do
                if [[ -f "$script" && "$script" != "$0" ]]; then
                    log "INFO" "Executing: $(basename "$script")"
                    bash "$script"
                fi
            done
            ;;
        *)
            log "WARN" "Native benchmark for '$suite' not implemented, trying generic approach"
            if [[ -f "$SCRIPT_DIR/run-${suite}-benchmarks.sh" ]]; then
                bash "$SCRIPT_DIR/run-${suite}-benchmarks.sh"
            else
                log "ERROR" "No benchmark script found for suite: $suite"
                return 1
            fi
            ;;
    esac
    
    log "SUCCESS" "Native benchmarks completed for suite: $suite"
}

# Results processing and validation
process_results() {
    log "INFO" "Processing benchmark results"
    
    # Find all result files
    local result_files=($(find "$RESULTS_DIR" -name "*.json" -type f -newer "$LOG_FILE" 2>/dev/null || true))
    
    if [[ ${#result_files[@]} -eq 0 ]]; then
        log "WARN" "No result files found"
        return 0
    fi
    
    log "INFO" "Found ${#result_files[@]} result files"
    
    # Validate JSON format
    local valid_files=0
    local invalid_files=0
    
    for file in "${result_files[@]}"; do
        if jq empty "$file" &>/dev/null; then
            ((valid_files++))
            log "INFO" "Valid result file: $(basename "$file")"
        else
            ((invalid_files++))
            log "ERROR" "Invalid JSON in file: $(basename "$file")"
        fi
    done
    
    log "INFO" "Results validation: $valid_files valid, $invalid_files invalid files"
    
    # Generate summary report
    local summary_file="$RESULTS_DIR/benchmark-summary-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$summary_file" << EOF
{
  "benchmark_execution": {
    "timestamp": "$(date -Iseconds)",
    "suite": "$BENCHMARK_SUITE",
    "instance_type": "$INSTANCE_TYPE",
    "execution_mode": "$([ "$DOCKER_MODE" = "true" ] && echo "docker" || echo "native")",
    "results": {
      "total_files": ${#result_files[@]},
      "valid_files": $valid_files,
      "invalid_files": $invalid_files,
      "files": [$(printf '"%s",' "${result_files[@]%$RESULTS_DIR/}" | sed 's/,$//')] 
    }
  }
}
EOF
    
    log "SUCCESS" "Summary report saved to $summary_file"
}

# Upload results to S3 (if configured)
upload_results() {
    if [[ "$UPLOAD_RESULTS" != "true" ]]; then
        log "INFO" "Result upload disabled (BENCHMARK_UPLOAD_RESULTS=false)"
        return 0
    fi
    
    if [[ -z "${S3_BENCHMARK_BUCKET:-}" ]]; then
        log "WARN" "S3_BENCHMARK_BUCKET not set, skipping upload"
        return 0
    fi
    
    if ! command -v aws &> /dev/null; then
        log "WARN" "AWS CLI not installed, skipping upload"
        return 0
    fi
    
    log "INFO" "Uploading results to S3: s3://$S3_BENCHMARK_BUCKET"
    
    local s3_path="s3://$S3_BENCHMARK_BUCKET/results/$(date +%Y-%m-%d)/$INSTANCE_TYPE/"
    
    if aws s3 sync "$RESULTS_DIR" "$s3_path" --exclude "*" --include "*.json" --include "*.log"; then
        log "SUCCESS" "Results uploaded to $s3_path"
    else
        log "ERROR" "Failed to upload results to S3"
        return 1
    fi
}

# Main execution function
main() {
    log "INFO" "Starting AWS Instance Benchmarks execution"
    log "INFO" "Suite: $BENCHMARK_SUITE, Instance: $INSTANCE_TYPE, Docker: $DOCKER_MODE"
    
    # Create directories
    mkdir -p "$RESULTS_DIR" "$DATA_DIR"
    
    # Collect system information
    collect_system_info
    
    # Run benchmarks
    if [[ "$DOCKER_MODE" == "true" ]]; then
        run_docker_benchmarks "$BENCHMARK_SUITE"
    else
        run_native_benchmarks "$BENCHMARK_SUITE"
    fi
    
    # Process and validate results
    process_results
    
    # Upload results if configured
    upload_results
    
    log "SUCCESS" "Benchmark execution completed successfully"
    echo ""
    echo "Results available at: $RESULTS_DIR"
    echo "Log file: $LOG_FILE"
    echo ""
    echo "To view results:"
    echo "  ls -la $RESULTS_DIR"
    echo ""
    echo "To run specific benchmark suites:"
    echo "  $0 genomics"
    echo "  $0 ml"
    echo "  $0 chemistry"
    echo "  $0 all"
}

# Usage information
usage() {
    cat << EOF
Usage: $0 [BENCHMARK_SUITE]

Benchmark Suites:
  genomics     - Genomics and bioinformatics benchmarks
  ml           - Machine learning and AI benchmarks  
  climate      - Climate and weather modeling benchmarks
  hep          - High energy physics benchmarks
  chemistry    - Computational chemistry benchmarks
  cfd          - Computational fluid dynamics benchmarks
  astronomy    - Astronomy and astrophysics benchmarks
  all          - All available benchmark suites (default)

Environment Variables:
  BENCHMARK_RESULTS_DIR     - Results output directory (default: ./results)
  BENCHMARK_DATA_DIR        - Data directory (default: ./data)
  BENCHMARK_INSTANCE_TYPE   - Instance type identifier (default: local)
  BENCHMARK_DOCKER_MODE     - Use Docker containers (default: true)
  BENCHMARK_UPLOAD_RESULTS  - Upload results to S3 (default: false)
  BENCHMARK_PARALLEL_JOBS   - Number of parallel jobs (default: 1)
  S3_BENCHMARK_BUCKET       - S3 bucket for result uploads
  AWS_DEFAULT_REGION        - AWS region (default: us-east-1)

Examples:
  # Run all benchmarks locally
  $0

  # Run ML benchmarks in Docker
  BENCHMARK_DOCKER_MODE=true $0 ml
  
  # Run genomics benchmarks natively
  BENCHMARK_DOCKER_MODE=false $0 genomics
  
  # Run with custom settings
  BENCHMARK_RESULTS_DIR=/tmp/results BENCHMARK_INSTANCE_TYPE=c7i.xlarge $0 chemistry

For more information, see: https://github.com/scttfrdmn/aws-instance-benchmarks
EOF
}

# Command line argument processing
if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    usage
    exit 0
fi

# Execute main function
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi