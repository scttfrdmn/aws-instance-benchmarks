#!/bin/bash
set -eo pipefail

# AWS Instance Benchmarks - Architecture-Aware Container Build Script
# This script launches EC2 instances to build containers on their target architecture
# with optimal toolchains for maximum performance

# Configuration
REGION="${AWS_REGION:-us-west-2}"
PROFILE="${AWS_PROFILE:-aws}"
KEY_NAME="${KEY_NAME:-scofri}"
SUBNET_ID="${SUBNET_ID:-subnet-0528a0d8c3da5acfb}"
SECURITY_GROUP="${SECURITY_GROUP:-sg-0a1b2c3d4e5f6789a}"
S3_BUCKET="${S3_BUCKET:-aws-instance-benchmarks-us-west-2-1751232301}"

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

# Get instance type for architecture
get_instance_type() {
    local arch="$1"
    case "$arch" in
        "universal") echo "t3.medium" ;;
        "intel-icelake") echo "m7i.large" ;;
        "amd-zen4") echo "m7a.large" ;;
        "graviton3") echo "m7g.large" ;;
        "graviton4") echo "m8g.large" ;;
        *) echo "t3.medium" ;;
    esac
}

# Get AMI for architecture
get_ami_id() {
    local arch="$1"
    case "$arch" in
        "universal"|"intel-icelake"|"amd-zen4") 
            echo "ami-0bbc328167dee8f3c"  # x86_64 Amazon Linux 2
            ;;
        "graviton3"|"graviton4") 
            echo "ami-001cfb1564f24ce79"  # arm64 Amazon Linux 2
            ;;
        *) 
            echo "ami-0bbc328167dee8f3c"  # default x86_64
            ;;
    esac
}

# Generate architecture-specific user data script
generate_build_script() {
    local arch="$1"
    local benchmark="$2"
    
    cat << 'EOF'
#!/bin/bash
set -eo pipefail

# Update system
sudo yum update -y

# Install Docker and development tools
sudo yum install -y docker git gcc gcc-c++ make cmake wget

# Start Docker
sudo systemctl start docker
sudo usermod -a -G docker ec2-user

# Create build directory
mkdir -p /home/ec2-user/build
cd /home/ec2-user/build

EOF

    # Add architecture-specific optimizations
    case "$arch" in
        "intel-icelake")
            cat << 'EOF'
# Build with Intel optimizations
export CC=gcc
export CXX=g++
export CFLAGS="-O3 -march=native -mtune=native -mavx2"
export CXXFLAGS="-O3 -march=native -mtune=native -mavx2"

EOF
            ;;
        "amd-zen4")
            cat << 'EOF'
# Build with AMD optimizations
export CC=gcc
export CXX=g++
export CFLAGS="-O3 -march=native -mtune=native -mavx2"
export CXXFLAGS="-O3 -march=native -mtune=native -mavx2"

EOF
            ;;
        "graviton3"|"graviton4")
            cat << 'EOF'
# Build with ARM optimizations
export CC=gcc
export CXX=g++
export CFLAGS="-O3 -march=native -mtune=native -mcpu=native"
export CXXFLAGS="-O3 -march=native -mtune=native -mcpu=native"

EOF
            ;;
        "universal")
            cat << 'EOF'
# Universal build with conservative optimizations
export CC=gcc
export CXX=g++
export CFLAGS="-O3 -march=native -mtune=native"
export CXXFLAGS="-O3 -march=native -mtune=native"

EOF
            ;;
    esac

    cat << EOF
# Clone the project repository
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks

# Build the specific container on this architecture
sudo docker build -t aws-instance-benchmarks/$benchmark:$arch builds/$arch/$benchmark/

# Test the container
sudo docker run --rm aws-instance-benchmarks/$benchmark:$arch --help || true

# Save container to tarball
sudo docker save aws-instance-benchmarks/$benchmark:$arch | gzip > $benchmark-$arch.tar.gz

# Upload to S3
aws s3 cp $benchmark-$arch.tar.gz s3://$S3_BUCKET/containers/$benchmark-$arch.tar.gz

# Signal completion
aws ssm put-parameter --name "/aws-benchmarks/build-status/$arch-$benchmark" --value "completed" --type String --overwrite --region $REGION

# Self-terminate after upload
sudo shutdown -h now
EOF
}

# Launch build instance for specific architecture
launch_build_instance() {
    local arch="$1"
    local benchmark="$2"
    local instance_type
    local ami_id
    
    instance_type=$(get_instance_type "$arch")
    ami_id=$(get_ami_id "$arch")
    
    log_info "Launching $instance_type for building $benchmark:$arch"
    
    # Generate user data script
    local user_data_file="/tmp/user-data-$arch.sh"
    generate_build_script "$arch" "$benchmark" > "$user_data_file"
    
    # Launch instance
    local instance_id
    instance_id=$(aws ec2 run-instances \
        --profile "$PROFILE" \
        --region "$REGION" \
        --image-id "$ami_id" \
        --instance-type "$instance_type" \
        --key-name "$KEY_NAME" \
        --subnet-id "$SUBNET_ID" \
        --security-group-ids "$SECURITY_GROUP" \
        --user-data "file://$user_data_file" \
        --iam-instance-profile Name=benchmark-instance-profile \
        --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=build-$benchmark-$arch},{Key=Purpose,Value=container-build},{Key=Architecture,Value=$arch}]" \
        --query 'Instances[0].InstanceId' \
        --output text)
    
    log_success "Launched build instance: $instance_id ($arch)"
    
    # Clean up temp file
    rm -f "$user_data_file"
    
    echo "$instance_id"
}

# Monitor build progress
monitor_builds() {
    local instance_ids="$1"
    
    log_info "Monitoring build instances: $instance_ids"
    
    while true; do
        local completed=0
        local total=0
        
        for instance_id in $instance_ids; do
            ((total++))
            # Check instance state
            local state
            state=$(aws ec2 describe-instances \
                --profile "$PROFILE" \
                --region "$REGION" \
                --instance-ids "$instance_id" \
                --query 'Reservations[0].Instances[0].State.Name' \
                --output text 2>/dev/null || echo "unknown")
            
            if [[ "$state" == "terminated" ]]; then
                ((completed++))
            fi
        done
        
        log_info "Build progress: $completed/$total instances completed"
        
        if [[ $completed -eq $total ]]; then
            log_success "All builds completed!"
            break
        fi
        
        sleep 30
    done
}

# Download built containers
download_containers() {
    log_info "Downloading built containers from S3..."
    
    mkdir -p ./built-containers
    
    # Download all container tarballs
    aws s3 sync "s3://$S3_BUCKET/containers/" ./built-containers/ \
        --profile "$PROFILE" \
        --region "$REGION"
    
    # Load containers into local registry
    for tarball in ./built-containers/*.tar.gz; do
        if [[ -f "$tarball" ]]; then
            log_info "Loading container: $tarball"
            gunzip -c "$tarball" | podman load
        fi
    done
    
    log_success "All containers loaded into local registry"
}

# Main execution
main() {
    local arch="${1:-all}"
    local benchmark="${2:-stream}"
    
    case "$arch" in
        "all")
            log_info "Building all architectures for $benchmark"
            local instance_ids=""
            
            for architecture in universal intel-icelake amd-zen4 graviton3 graviton4; do
                local instance_id
                instance_id=$(launch_build_instance "$architecture" "$benchmark")
                instance_ids="$instance_ids $instance_id"
                
                # Stagger launches to avoid hitting API limits
                sleep 10
            done
            
            monitor_builds "$instance_ids"
            download_containers
            ;;
        *)
            log_info "Building $arch architecture for $benchmark"
            local instance_id
            instance_id=$(launch_build_instance "$arch" "$benchmark")
            monitor_builds "$instance_id"
            download_containers
            ;;
    esac
    
    log_success "Architecture-aware container builds completed!"
    log_info "Built containers are available in your local Podman registry"
    
    # Show built containers
    podman images | grep "aws-instance-benchmarks" || true
}

# Usage
if [[ "${1:-}" == "--help" ]] || [[ "${1:-}" == "-h" ]]; then
    cat << EOF
Architecture-Aware Container Build Script

Usage: $0 [ARCHITECTURE] [BENCHMARK]

Arguments:
    ARCHITECTURE    Target architecture (universal, intel-icelake, amd-zen4, graviton3, graviton4, or 'all')
    BENCHMARK       Benchmark to build (default: stream)

Examples:
    $0 all stream                    # Build stream containers for all architectures
    $0 intel-icelake stream         # Build stream container for Intel Ice Lake
    $0 graviton4 stream             # Build stream container for Graviton4

Environment Variables:
    AWS_REGION      AWS region (default: us-west-2)
    AWS_PROFILE     AWS profile (default: aws)
    KEY_NAME        EC2 key pair name (default: scofri)

This script:
1. Launches EC2 instances of each target architecture
2. Installs optimal toolchains
3. Builds containers with architecture-specific optimizations
4. Uploads containers to S3 and loads them locally
5. Terminates build instances automatically

EOF
    exit 0
fi

main "$@"