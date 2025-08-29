#!/bin/bash
set -euo pipefail

# AWS Instance Benchmarks - Container Build Management Script
# This script builds all required benchmark containers for different architectures
# Supports both local development (Podman) and ECR deployment (Docker)

# Configuration
REGISTRY_PREFIX="${REGISTRY_PREFIX:-localhost/aws-instance-benchmarks}"
BUILD_DIR="$(dirname "$0")/../builds"
CONTAINER_TOOL="${CONTAINER_TOOL:-podman}"
ECR_MODE="${ECR_MODE:-false}"
REGION="${AWS_REGION:-us-east-1}"
PROFILE="${AWS_PROFILE:-aws}"
ECR_REPO_NAME="aws-benchmarks"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if container tool is available
check_container_tool() {
    if ! command -v "$CONTAINER_TOOL" &> /dev/null; then
        log_error "Container tool '$CONTAINER_TOOL' not found. Please install $CONTAINER_TOOL or set CONTAINER_TOOL environment variable."
        exit 1
    fi
    log_info "Using container tool: $CONTAINER_TOOL"
}

# Build a single container
build_container() {
    local arch="$1"
    local benchmark="$2"
    local build_path="$BUILD_DIR/$arch/$benchmark"
    local image_tag="$REGISTRY_PREFIX/$benchmark:$arch"
    
    if [[ ! -d "$build_path" ]]; then
        log_warning "Build directory not found: $build_path"
        return 1
    fi
    
    if [[ ! -f "$build_path/Dockerfile" ]]; then
        log_warning "Dockerfile not found in: $build_path"
        return 1
    fi
    
    log_info "Building container: $image_tag"
    log_info "Build context: $build_path"
    
    if $CONTAINER_TOOL build -t "$image_tag" "$build_path"; then
        log_success "Successfully built: $image_tag"
        return 0
    else
        log_error "Failed to build: $image_tag"
        return 1
    fi
}

# List available containers
list_containers() {
    log_info "Available container builds:"
    for arch_dir in "$BUILD_DIR"/*; do
        if [[ -d "$arch_dir" ]]; then
            arch=$(basename "$arch_dir")
            echo -e "  ${YELLOW}$arch:${NC}"
            for benchmark_dir in "$arch_dir"/*; do
                if [[ -d "$benchmark_dir" ]] && [[ -f "$benchmark_dir/Dockerfile" ]]; then
                    benchmark=$(basename "$benchmark_dir")
                    echo "    - $benchmark"
                fi
            done
        fi
    done
}

# Build all containers for a specific architecture
build_arch() {
    local arch="$1"
    local arch_dir="$BUILD_DIR/$arch"
    
    if [[ ! -d "$arch_dir" ]]; then
        log_error "Architecture directory not found: $arch_dir"
        return 1
    fi
    
    log_info "Building all containers for architecture: $arch"
    
    local success_count=0
    local total_count=0
    
    for benchmark_dir in "$arch_dir"/*; do
        if [[ -d "$benchmark_dir" ]] && [[ -f "$benchmark_dir/Dockerfile" ]]; then
            benchmark=$(basename "$benchmark_dir")
            ((total_count++))
            if build_container "$arch" "$benchmark"; then
                ((success_count++))
            fi
        fi
    done
    
    log_info "Built $success_count/$total_count containers for $arch"
    return $((total_count - success_count))
}

# Build all containers
build_all() {
    log_info "Building all benchmark containers..."
    
    local total_success=0
    local total_containers=0
    
    for arch_dir in "$BUILD_DIR"/*; do
        if [[ -d "$arch_dir" ]]; then
            arch=$(basename "$arch_dir")
            for benchmark_dir in "$arch_dir"/*; do
                if [[ -d "$benchmark_dir" ]] && [[ -f "$benchmark_dir/Dockerfile" ]]; then
                    benchmark=$(basename "$benchmark_dir")
                    ((total_containers++))
                    if build_container "$arch" "$benchmark"; then
                        ((total_success++))
                    fi
                fi
            done
        fi
    done
    
    log_info "Build summary: $total_success/$total_containers containers built successfully"
    
    if [[ $total_success -eq $total_containers ]]; then
        log_success "All containers built successfully!"
        return 0
    else
        log_warning "Some containers failed to build"
        return 1
    fi
}

# Clean up containers
clean_containers() {
    log_info "Cleaning up benchmark containers..."
    
    # Remove containers matching our registry prefix
    local images
    images=$($CONTAINER_TOOL images --format "{{.Repository}}:{{.Tag}}" | grep "^$REGISTRY_PREFIX" || true)
    
    if [[ -z "$images" ]]; then
        log_info "No benchmark containers found to clean"
        return 0
    fi
    
    echo "$images" | while read -r image; do
        log_info "Removing image: $image"
        $CONTAINER_TOOL rmi "$image" || log_warning "Failed to remove $image"
    done
    
    log_success "Container cleanup completed"
}

# Show container status
show_status() {
    log_info "Current benchmark containers:"
    $CONTAINER_TOOL images | grep "$REGISTRY_PREFIX" || log_info "No benchmark containers found"
}

# Usage information
usage() {
    cat << EOF
Container Build Management Script for AWS Instance Benchmarks

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    list                    List all available container builds
    build-all              Build all containers
    build-arch ARCH        Build all containers for specific architecture
    build ARCH BENCHMARK   Build specific container
    clean                  Remove all benchmark containers
    status                 Show current benchmark containers
    help                   Show this help message

Architectures:
    universal             Universal compatibility (fallback)
    intel-icelake         Intel Ice Lake optimized
    amd-zen4              AMD Zen 4 optimized  
    graviton3             AWS Graviton3 optimized
    graviton4             AWS Graviton4 optimized

Benchmarks:
    stream                Memory bandwidth benchmark

Environment Variables:
    CONTAINER_TOOL        Container tool to use (default: podman)
    REGISTRY_PREFIX       Local registry prefix (default: localhost/aws-instance-benchmarks)

Examples:
    $0 list                                 # List available builds
    $0 build-all                           # Build everything
    $0 build-arch universal                # Build all universal containers
    $0 build universal stream              # Build specific container
    $0 clean                               # Clean up all containers
    $0 status                              # Show current containers

EOF
}

# Main script logic
main() {
    check_container_tool
    
    case "${1:-help}" in
        "list")
            list_containers
            ;;
        "build-all")
            build_all
            ;;
        "build-arch")
            if [[ $# -lt 2 ]]; then
                log_error "Architecture required for build-arch command"
                usage
                exit 1
            fi
            build_arch "$2"
            ;;
        "build")
            if [[ $# -lt 3 ]]; then
                log_error "Architecture and benchmark required for build command"
                usage
                exit 1
            fi
            build_container "$2" "$3"
            ;;
        "clean")
            clean_containers
            ;;
        "status")
            show_status
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
}

# Run main function with all arguments
main "$@"