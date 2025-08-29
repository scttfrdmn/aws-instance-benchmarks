#!/bin/bash

# Native AWS Container Build Orchestrator
# Builds benchmark containers on their target AWS instance types for maximum performance
# Uses AWS Systems Manager (SSM) for remote container building and retrieval

set -euo pipefail

# Configuration
AWS_REGION=${AWS_REGION:-us-east-1}
KEY_NAME="aws-instance-benchmarks-key"
SECURITY_GROUP="benchmark-build-sg"
IAM_ROLE="EC2-SSM-Role"
BUILD_TIMEOUT=3600  # 1 hour timeout
S3_BUCKET="aws-instance-benchmarks-containers"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[BUILD]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Architecture to instance type mapping
declare -A ARCH_INSTANCES=(
    # Intel Generations
    ["intel-sapphirerapids"]="m7i.large c7i.large r7i.large"
    ["intel-icelake"]="m6i.large c6i.large r6i.large"
    ["intel-cascadelake"]="m5n.large c5n.large r5n.large"
    ["intel-skylake"]="m5.large c5.large r5.large"
    ["intel-broadwell"]="m4.large c4.large r4.large"
    ["intel-haswell"]="m3.medium c3.large"
    ["intel-ivybridge"]="m3.medium c3.medium"
    ["intel-sandybridge"]="m2.xlarge c2.medium"
    ["intel-nehalem"]="m1.large c1.medium"
    
    # AMD Generations  
    ["amd-zen4"]="m7a.large c7a.large r7a.large"
    ["amd-zen3"]="m6a.large c6a.large r6a.large" 
    ["amd-zen2"]="m5a.large c5a.large r5a.large"
    ["amd-zen1"]="m5a.large"
    
    # AWS Graviton ARM
    ["graviton4"]="c8g.large m8g.large r8g.large"
    ["graviton3e"]="c7gn.large hpc7g.medium"
    ["graviton3"]="m7g.large c7g.large r7g.large"
    ["graviton2"]="m6g.large c6g.large r6g.large"
    ["graviton1"]="m6g.large"
)

# Container variants to build
declare -A CONTAINER_VARIANTS=(
    ["universal"]="builds/universal/comprehensive/Dockerfile"
    ["intel-optimized"]="builds/universal/intel-optimized/Dockerfile" 
    ["amd-optimized"]="builds/universal/amd-optimized/Dockerfile"
)

# Get latest AMI ID for architecture
get_latest_ami() {
    local arch="$1"
    
    case "$arch" in
        x86_64)
            aws ec2 describe-images \
                --owners amazon \
                --filters "Name=name,Values=amzn2-ami-hvm-*-x86_64-gp2" \
                          "Name=state,Values=available" \
                --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
                --output text \
                --region "$AWS_REGION"
            ;;
        arm64)
            aws ec2 describe-images \
                --owners amazon \
                --filters "Name=name,Values=amzn2-ami-hvm-*-arm64-gp2" \
                          "Name=state,Values=available" \
                --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
                --output text \
                --region "$AWS_REGION"
            ;;
    esac
}

# Launch instance for container building
launch_build_instance() {
    local instance_type="$1"
    local architecture="$2"
    local ami_id="$3"
    
    log_info "Launching $instance_type instance for native container building..."
    
    local instance_id=$(aws ec2 run-instances \
        --image-id "$ami_id" \
        --count 1 \
        --instance-type "$instance_type" \
        --key-name "$KEY_NAME" \
        --security-group-ids "$SECURITY_GROUP" \
        --iam-instance-profile Name="$IAM_ROLE" \
        --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=benchmark-build-$instance_type-$architecture},{Key=Purpose,Value=ContainerBuild}]" \
        --user-data file://scripts/container-build-userdata.sh \
        --query 'Instances[0].InstanceId' \
        --output text \
        --region "$AWS_REGION")
    
    if [[ -z "$instance_id" ]]; then
        log_error "Failed to launch instance"
        return 1
    fi
    
    log_success "Launched instance: $instance_id"
    
    # Wait for instance to be ready
    log_info "Waiting for instance to be ready for SSM commands..."
    aws ec2 wait instance-status-ok --instance-ids "$instance_id" --region "$AWS_REGION"
    
    # Additional wait for SSM agent initialization
    sleep 60
    
    echo "$instance_id"
}

# Build container via SSM
build_container_remote() {
    local instance_id="$1"
    local architecture="$2"
    local container_variant="$3"
    local dockerfile_path="$4"
    
    log_info "Building $container_variant container for $architecture on $instance_id..."
    
    # Create SSM document for container building
    local build_script=$(cat << 'EOF'
#!/bin/bash
set -euo pipefail

# Install Docker
yum update -y
yum install -y docker git
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Clone repository
cd /tmp
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks

# Build container
DOCKERFILE_PATH="$1"
ARCHITECTURE="$2"
VARIANT="$3"

docker build -t aws-benchmark-$VARIANT:$ARCHITECTURE -f "$DOCKERFILE_PATH" .

# Save container to tarball
docker save aws-benchmark-$VARIANT:$ARCHITECTURE > /tmp/aws-benchmark-$VARIANT-$ARCHITECTURE.tar

# Upload to S3
aws s3 cp /tmp/aws-benchmark-$VARIANT-$ARCHITECTURE.tar s3://aws-instance-benchmarks-containers/$VARIANT-$ARCHITECTURE.tar

echo "Container build complete: aws-benchmark-$VARIANT:$ARCHITECTURE"
EOF
)
    
    # Execute build via SSM
    local command_id=$(aws ssm send-command \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters "commands=[\"$build_script\",\"$dockerfile_path\",\"$architecture\",\"$container_variant\"]" \
        --timeout-seconds "$BUILD_TIMEOUT" \
        --query 'Command.CommandId' \
        --output text \
        --region "$AWS_REGION")
    
    log_info "SSM Command ID: $command_id"
    
    # Wait for command completion
    log_info "Waiting for container build to complete..."
    aws ssm wait command-executed \
        --command-id "$command_id" \
        --instance-id "$instance_id" \
        --region "$AWS_REGION"
    
    # Get command output
    local command_output=$(aws ssm get-command-invocation \
        --command-id "$command_id" \
        --instance-id "$instance_id" \
        --query 'StandardOutputContent' \
        --output text \
        --region "$AWS_REGION")
    
    local command_status=$(aws ssm get-command-invocation \
        --command-id "$command_id" \
        --instance-id "$instance_id" \
        --query 'Status' \
        --output text \
        --region "$AWS_REGION")
    
    if [[ "$command_status" == "Success" ]]; then
        log_success "Container build completed successfully"
        log_info "Build output: $command_output"
        return 0
    else
        log_error "Container build failed with status: $command_status"
        log_error "Build output: $command_output"
        return 1
    fi
}

# Clean up build instance
cleanup_instance() {
    local instance_id="$1"
    
    log_info "Terminating build instance: $instance_id"
    aws ec2 terminate-instances --instance-ids "$instance_id" --region "$AWS_REGION" >/dev/null
    log_success "Build instance terminated"
}

# Download built container
download_container() {
    local container_variant="$1"
    local architecture="$2"
    local output_dir="$3"
    
    log_info "Downloading built container: $container_variant-$architecture"
    
    mkdir -p "$output_dir"
    aws s3 cp "s3://$S3_BUCKET/$container_variant-$architecture.tar" \
        "$output_dir/$container_variant-$architecture.tar" \
        --region "$AWS_REGION"
    
    log_success "Container downloaded to: $output_dir/$container_variant-$architecture.tar"
}

# Build container for specific architecture
build_architecture() {
    local architecture="$1"
    local container_variant="$2"
    local output_dir="${3:-./containers}"
    
    log_info "=== Building $container_variant container for $architecture ==="
    
    # Get appropriate instance type
    local instance_types="${ARCH_INSTANCES[$architecture]:-}"
    if [[ -z "$instance_types" ]]; then
        log_error "No instances available for architecture: $architecture"
        return 1
    fi
    
    local instance_type=$(echo "$instance_types" | awk '{print $1}')
    log_info "Using instance type: $instance_type"
    
    # Determine AMI architecture
    local ami_arch="x86_64"
    if [[ "$architecture" =~ graviton ]]; then
        ami_arch="arm64"
    fi
    
    # Get latest AMI
    local ami_id=$(get_latest_ami "$ami_arch")
    if [[ -z "$ami_id" ]]; then
        log_error "Could not find suitable AMI for $ami_arch"
        return 1
    fi
    
    log_info "Using AMI: $ami_id"
    
    # Get Dockerfile path
    local dockerfile_path="${CONTAINER_VARIANTS[$container_variant]:-}"
    if [[ -z "$dockerfile_path" ]]; then
        log_error "Unknown container variant: $container_variant"
        return 1
    fi
    
    # Launch build instance
    local instance_id=$(launch_build_instance "$instance_type" "$architecture" "$ami_id")
    if [[ -z "$instance_id" ]]; then
        return 1
    fi
    
    # Build container remotely
    if build_container_remote "$instance_id" "$architecture" "$container_variant" "$dockerfile_path"; then
        # Download built container
        download_container "$container_variant" "$architecture" "$output_dir"
        local build_success=0
    else
        local build_success=1
    fi
    
    # Clean up
    cleanup_instance "$instance_id"
    
    return $build_success
}

# Build all containers for all architectures
build_all() {
    local output_dir="${1:-./containers}"
    local successful_builds=0
    local failed_builds=0
    
    log_info "=== Building All Native Containers ==="
    
    for architecture in "${!ARCH_INSTANCES[@]}"; do
        for variant in "${!CONTAINER_VARIANTS[@]}"; do
            # Skip non-compatible combinations
            if [[ "$variant" == "intel-optimized" && ! "$architecture" =~ ^intel- ]]; then
                continue
            fi
            if [[ "$variant" == "amd-optimized" && ! "$architecture" =~ ^amd- ]]; then
                continue  
            fi
            
            log_info "Building $variant for $architecture..."
            
            if build_architecture "$architecture" "$variant" "$output_dir"; then
                ((successful_builds++))
                log_success "✅ $variant-$architecture build completed"
            else
                ((failed_builds++))
                log_error "❌ $variant-$architecture build failed"
            fi
        done
    done
    
    log_info "=== Build Summary ==="
    log_success "Successful builds: $successful_builds"
    if [[ $failed_builds -gt 0 ]]; then
        log_error "Failed builds: $failed_builds"
    fi
    
    log_info "All built containers are available in: $output_dir"
    log_info "Containers have been uploaded to S3: s3://$S3_BUCKET/"
}

# Usage information
usage() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  build <architecture> <variant> [output_dir]  - Build specific container"
    echo "  build-all [output_dir]                       - Build all containers"  
    echo "  list-architectures                           - List supported architectures"
    echo "  list-variants                                - List container variants"
    echo ""
    echo "Examples:"
    echo "  $0 build intel-sapphirerapids universal ./containers"
    echo "  $0 build amd-zen4 amd-optimized ./containers"
    echo "  $0 build graviton4 universal ./containers"
    echo "  $0 build-all ./containers"
    echo ""
    echo "Container Variants:"
    echo "  universal        - GCC with runtime optimization (works on all architectures)"
    echo "  intel-optimized  - Intel oneAPI + MKL (Intel only)"
    echo "  amd-optimized    - AMD AOCC + AOCL (AMD only)"
}

# Main execution
main() {
    case "${1:-}" in
        "build")
            if [[ $# -lt 3 ]]; then
                echo "Error: Architecture and variant required"
                usage
                exit 1
            fi
            build_architecture "$2" "$3" "${4:-./containers}"
            ;;
        "build-all")
            build_all "${2:-./containers}"
            ;;
        "list-architectures")
            echo "Supported architectures:"
            for arch in "${!ARCH_INSTANCES[@]}"; do
                echo "  $arch: ${ARCH_INSTANCES[$arch]}"
            done
            ;;
        "list-variants")
            echo "Container variants:"
            for variant in "${!CONTAINER_VARIANTS[@]}"; do
                echo "  $variant: ${CONTAINER_VARIANTS[$variant]}"
            done
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"