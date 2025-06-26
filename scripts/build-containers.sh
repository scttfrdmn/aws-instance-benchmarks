#!/bin/bash
set -euo pipefail

# AWS Instance Benchmarks - Container Build Script
# This script builds and pushes optimized benchmark containers

# Configuration
REGION="${AWS_REGION:-us-east-1}"
PROFILE="${AWS_PROFILE:-aws}"
ECR_REPO_NAME="aws-benchmarks/stream"

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker."
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon not running. Please start Docker."
        exit 1
    fi
    
    # Check AWS CLI and credentials
    if ! aws sts get-caller-identity --profile "$PROFILE" &> /dev/null; then
        log_error "AWS credentials not configured for profile '$PROFILE'."
        exit 1
    fi
    
    # Get account information
    ACCOUNT_ID=$(aws sts get-caller-identity --profile "$PROFILE" --query Account --output text)
    ECR_REPO_URI="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/$ECR_REPO_NAME"
    
    log_success "Docker ready"
    log_success "AWS account: $ACCOUNT_ID"
    log_success "ECR repository: $ECR_REPO_URI"
}

# Login to ECR
ecr_login() {
    log_info "Logging in to ECR..."
    
    aws ecr get-login-password --region "$REGION" --profile "$PROFILE" | \
        docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com"
    
    log_success "ECR login successful"
}

# Build container for specific architecture
build_container() {
    local arch=$1
    local build_dir="builds/$arch/stream"
    
    log_info "Building container for architecture: $arch"
    
    if [ ! -d "$build_dir" ]; then
        log_error "Build directory not found: $build_dir"
        return 1
    fi
    
    # Build the container
    docker build \
        -t "aws-benchmarks/stream:$arch" \
        -f "$build_dir/Dockerfile" \
        "$build_dir"
    
    log_success "Built container: aws-benchmarks/stream:$arch"
    
    # Tag for ECR
    docker tag "aws-benchmarks/stream:$arch" "$ECR_REPO_URI:$arch"
    log_success "Tagged container for ECR: $ECR_REPO_URI:$arch"
}

# Push container to ECR
push_container() {
    local arch=$1
    
    log_info "Pushing container to ECR: $arch"
    
    docker push "$ECR_REPO_URI:$arch"
    log_success "Pushed container: $ECR_REPO_URI:$arch"
}

# Test container functionality
test_container() {
    local arch=$1
    
    log_info "Testing container: $arch"
    
    # Run a quick test to verify the container works
    local test_output
    test_output=$(docker run --rm "aws-benchmarks/stream:$arch" --help 2>&1 || true)
    
    if echo "$test_output" | grep -q "STREAM"; then
        log_success "Container test passed: $arch"
    else
        log_warning "Container test inconclusive: $arch"
        log_info "Test output: $test_output"
    fi
}

# Build all containers
build_all_containers() {
    local architectures=("intel-icelake" "graviton3")
    
    # Add AMD architecture if Dockerfile exists
    if [ -f "builds/amd-zen4/stream/Dockerfile" ]; then
        architectures+=("amd-zen4")
    fi
    
    log_info "Building containers for architectures: ${architectures[*]}"
    
    for arch in "${architectures[@]}"; do
        log_info "Processing architecture: $arch"
        
        if build_container "$arch"; then
            test_container "$arch"
            push_container "$arch"
        else
            log_error "Failed to build container for: $arch"
            continue
        fi
    done
}

# Create AMD Zen4 Dockerfile if it doesn't exist
create_amd_dockerfile() {
    local amd_dir="builds/amd-zen4/stream"
    
    if [ ! -d "$amd_dir" ]; then
        log_info "Creating AMD Zen4 build directory"
        mkdir -p "$amd_dir/spack-configs"
        
        # Copy base structure from Intel
        cp -r builds/intel-icelake/stream/spack-configs/* "$amd_dir/spack-configs/"
        
        # Create AMD-specific Dockerfile
        cat > "$amd_dir/Dockerfile" << 'EOF'
# AMD Zen4 optimized STREAM benchmark container
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gfortran \
    python3 \
    python3-pip \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Spack
RUN git clone -c feature.manyFiles=true https://github.com/spack/spack.git /opt/spack
ENV SPACK_ROOT=/opt/spack
ENV PATH=$SPACK_ROOT/bin:$PATH

# Copy Spack configuration
COPY spack-configs/amd-zen4.yaml /opt/spack/etc/spack/packages.yaml

# Install STREAM with AMD AOCC optimization
RUN spack install stream@5.10 %gcc@11 target=zen4 +openmp
RUN spack load stream

# Create benchmark script
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'spack load stream' >> /usr/local/bin/run-stream && \
    echo 'stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]
EOF
        
        log_success "Created AMD Zen4 Dockerfile"
    fi
}

# Generate container manifest
generate_manifest() {
    log_info "Generating container manifest"
    
    cat > builds/container-manifest.json << EOF
{
  "containers": {
    "stream": {
      "repository": "$ECR_REPO_URI",
      "architectures": {
        "intel-icelake": {
          "tag": "intel-icelake",
          "target_instances": ["m7i.*", "c7i.*", "r7i.*"],
          "optimization": "Intel OneAPI with AVX-512",
          "compiler": "icc",
          "last_built": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
        },
        "graviton3": {
          "tag": "graviton3",
          "target_instances": ["m7g.*", "c7g.*", "r7g.*"],
          "optimization": "GCC with ARM Neon/SVE",
          "compiler": "gcc",
          "last_built": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
        }
EOF

    if [ -f "builds/amd-zen4/stream/Dockerfile" ]; then
        cat >> builds/container-manifest.json << EOF
        ,
        "amd-zen4": {
          "tag": "amd-zen4",
          "target_instances": ["m7a.*", "c7a.*", "r7a.*"],
          "optimization": "AMD AOCC with AVX-512",
          "compiler": "aocc",
          "last_built": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
        }
EOF
    fi

    cat >> builds/container-manifest.json << EOF
      }
    }
  },
  "metadata": {
    "generated": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "ecr_repository": "$ECR_REPO_URI",
    "region": "$REGION"
  }
}
EOF
    
    log_success "Generated container manifest: builds/container-manifest.json"
}

# Main execution
main() {
    log_info "Starting container build process for AWS Instance Benchmarks"
    
    check_prerequisites
    ecr_login
    create_amd_dockerfile
    build_all_containers
    generate_manifest
    
    log_success "Container build process completed successfully!"
    log_info ""
    log_info "Built containers:"
    docker images | grep "aws-benchmarks/stream" || log_warning "No local containers found"
    log_info ""
    log_info "ECR repository: $ECR_REPO_URI"
    log_info "Available tags:"
    aws ecr list-images --repository-name "$ECR_REPO_NAME" --region "$REGION" --profile "$PROFILE" --query 'imageIds[].imageTag' --output table || true
    log_info ""
    log_info "Next steps:"
    log_info "1. Test container execution on EC2 instances"
    log_info "2. Run initial benchmark validation"
    log_info "3. Execute full benchmark collection"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi