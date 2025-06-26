#!/bin/bash
set -euo pipefail

# AWS Instance Benchmarks - Infrastructure Setup Script
# This script sets up the required AWS infrastructure for benchmark collection

# Configuration
BUCKET_NAME="aws-instance-benchmarks-data"
ROLE_NAME="AWSInstanceBenchmarksRole"
POLICY_NAME="AWSInstanceBenchmarksPolicy"
ECR_REPO_NAME="aws-benchmarks/stream"
REGION="${AWS_REGION:-us-east-1}"
PROFILE="${AWS_PROFILE:-aws}"

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
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI not found. Please install AWS CLI v2."
        exit 1
    fi
    
    # Check AWS CLI version
    aws_version=$(aws --version 2>&1 | cut -d/ -f2 | cut -d' ' -f1)
    log_info "AWS CLI version: $aws_version"
    
    # Check AWS credentials
    if ! aws sts get-caller-identity --profile "$PROFILE" &> /dev/null; then
        log_error "AWS credentials not configured for profile '$PROFILE'."
        log_info "Please run: aws configure --profile $PROFILE"
        exit 1
    fi
    
    # Get account information
    ACCOUNT_ID=$(aws sts get-caller-identity --profile "$PROFILE" --query Account --output text)
    USER_ARN=$(aws sts get-caller-identity --profile "$PROFILE" --query Arn --output text)
    
    log_success "AWS account: $ACCOUNT_ID"
    log_success "User/Role: $USER_ARN"
    log_success "Region: $REGION"
}

# Create S3 bucket for benchmark data
setup_s3_bucket() {
    log_info "Setting up S3 bucket: $BUCKET_NAME"
    
    # Check if bucket exists
    if aws s3api head-bucket --bucket "$BUCKET_NAME" --profile "$PROFILE" 2>/dev/null; then
        log_warning "S3 bucket $BUCKET_NAME already exists"
    else
        # Create bucket
        if [ "$REGION" = "us-east-1" ]; then
            aws s3api create-bucket \
                --bucket "$BUCKET_NAME" \
                --region "$REGION" \
                --profile "$PROFILE"
        else
            aws s3api create-bucket \
                --bucket "$BUCKET_NAME" \
                --region "$REGION" \
                --create-bucket-configuration LocationConstraint="$REGION" \
                --profile "$PROFILE"
        fi
        log_success "Created S3 bucket: $BUCKET_NAME"
    fi
    
    # Configure bucket versioning
    aws s3api put-bucket-versioning \
        --bucket "$BUCKET_NAME" \
        --versioning-configuration Status=Enabled \
        --profile "$PROFILE"
    log_success "Enabled versioning on S3 bucket"
    
    # Ensure security - keep public access blocked
    aws s3api put-public-access-block \
        --bucket "$BUCKET_NAME" \
        --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
        --profile "$PROFILE"
    log_success "Secured bucket with public access blocking"
    
    # Note: For public data access, we'll use GitHub Pages or CloudFront instead of direct S3 public access
    log_info "Public data access will be provided via GitHub Pages for security compliance"
    
    # Create initial directory structure
    aws s3api put-object \
        --bucket "$BUCKET_NAME" \
        --key "raw/" \
        --profile "$PROFILE" > /dev/null
    aws s3api put-object \
        --bucket "$BUCKET_NAME" \
        --key "processed/latest/" \
        --profile "$PROFILE" > /dev/null
    aws s3api put-object \
        --bucket "$BUCKET_NAME" \
        --key "processed/historical/" \
        --profile "$PROFILE" > /dev/null
    aws s3api put-object \
        --bucket "$BUCKET_NAME" \
        --key "schemas/" \
        --profile "$PROFILE" > /dev/null
    log_success "Created S3 directory structure"
}

# Create ECR repository for containers
setup_ecr_repository() {
    log_info "Setting up ECR repository: $ECR_REPO_NAME"
    
    # Check if repository exists
    if aws ecr describe-repositories \
        --repository-names "$ECR_REPO_NAME" \
        --region "$REGION" \
        --profile "$PROFILE" &> /dev/null; then
        log_warning "ECR repository $ECR_REPO_NAME already exists"
    else
        # Create repository
        aws ecr create-repository \
            --repository-name "$ECR_REPO_NAME" \
            --region "$REGION" \
            --profile "$PROFILE" > /dev/null
        log_success "Created ECR repository: $ECR_REPO_NAME"
    fi
    
    # Get repository URI
    REPO_URI=$(aws ecr describe-repositories \
        --repository-names "$ECR_REPO_NAME" \
        --region "$REGION" \
        --profile "$PROFILE" \
        --query 'repositories[0].repositoryUri' \
        --output text)
    
    log_success "ECR repository URI: $REPO_URI"
    echo "export ECR_REPO_URI=$REPO_URI" >> ~/.aws_benchmark_env
}

# Create CloudWatch dashboard
setup_cloudwatch_dashboard() {
    log_info "Setting up CloudWatch dashboard"
    
    cat > /tmp/dashboard.json << EOF
{
  "widgets": [
    {
      "type": "metric",
      "x": 0,
      "y": 0,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "InstanceBenchmarks", "BenchmarkExecution", "Success", "true" ],
          [ "...", "false" ]
        ],
        "period": 300,
        "stat": "Sum",
        "region": "$REGION",
        "title": "Benchmark Execution Success Rate"
      }
    },
    {
      "type": "metric",
      "x": 0,
      "y": 6,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "InstanceBenchmarks", "ExecutionDuration", "BenchmarkSuite", "stream" ],
          [ "...", "hpl" ]
        ],
        "period": 300,
        "stat": "Average",
        "region": "$REGION",
        "title": "Average Execution Duration"
      }
    },
    {
      "type": "metric",
      "x": 12,
      "y": 0,
      "width": 12,
      "height": 6,
      "properties": {
        "metrics": [
          [ "InstanceBenchmarks", "QualityScore" ]
        ],
        "period": 300,
        "stat": "Average",
        "region": "$REGION",
        "title": "Benchmark Quality Score"
      }
    }
  ]
}
EOF
    
    aws cloudwatch put-dashboard \
        --dashboard-name "AWSInstanceBenchmarks" \
        --dashboard-body file:///tmp/dashboard.json \
        --profile "$PROFILE" > /dev/null
    
    log_success "Created CloudWatch dashboard: AWSInstanceBenchmarks"
}

# Check service quotas
check_service_quotas() {
    log_info "Checking EC2 service quotas"
    
    # Check Running On-Demand instances quota
    quota=$(aws service-quotas get-service-quota \
        --service-code ec2 \
        --quota-code L-1216C47A \
        --region "$REGION" \
        --profile "$PROFILE" \
        --query 'Quota.Value' \
        --output text 2>/dev/null || echo "Unknown")
    
    log_info "Running On-Demand instances quota: $quota"
    
    if [ "$quota" != "Unknown" ] && [ "$(echo "$quota < 50" | bc -l)" = "1" ]; then
        log_warning "EC2 quota may be insufficient for large-scale benchmarking"
        log_info "Consider requesting quota increase for L-1216C47A (Running On-Demand instances)"
    fi
    
    # Check specific instance family quotas
    declare -a families=("m7i" "c7i" "r7i" "c7g" "m7g" "r7g")
    for family in "${families[@]}"; do
        log_info "Instance family $family quota check would require family-specific quota codes"
    done
}

# Create environment configuration
create_environment_config() {
    log_info "Creating environment configuration"
    
    cat > ~/.aws_benchmark_env << EOF
# AWS Instance Benchmarks Environment Configuration
export AWS_PROFILE=$PROFILE
export AWS_REGION=$REGION
export AWS_ACCOUNT_ID=$ACCOUNT_ID
export BENCHMARK_S3_BUCKET=$BUCKET_NAME
export ECR_REPO_URI=$REPO_URI

# Usage: source ~/.aws_benchmark_env before running benchmark commands
EOF
    
    log_success "Created environment configuration: ~/.aws_benchmark_env"
    log_info "Run 'source ~/.aws_benchmark_env' to load environment variables"
}

# Main execution
main() {
    log_info "Starting AWS infrastructure setup for Instance Benchmarks"
    log_info "Region: $REGION, Profile: $PROFILE"
    
    check_prerequisites
    setup_s3_bucket
    setup_ecr_repository
    setup_cloudwatch_dashboard
    check_service_quotas
    create_environment_config
    
    log_success "AWS infrastructure setup completed successfully!"
    log_info ""
    log_info "Next steps:"
    log_info "1. Source environment: source ~/.aws_benchmark_env"
    log_info "2. Build and push containers: ./scripts/build-containers.sh"
    log_info "3. Run initial benchmarks: ./aws-benchmark-collector run --help"
    log_info ""
    log_info "S3 Bucket: $BUCKET_NAME"
    log_info "ECR Repository: $REPO_URI"
    log_info "CloudWatch Dashboard: https://console.aws.amazon.com/cloudwatch/home?region=$REGION#dashboards:name=AWSInstanceBenchmarks"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi