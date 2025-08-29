#!/bin/bash
set -euo pipefail

# Remote Container Build Script for AWS Instance Benchmarks
# Builds architecture-specific containers on the appropriate EC2 instances

INSTANCE_HOST="$1"
ARCHITECTURE="$2"
BENCHMARK="$3"
KEY_PATH="${AWS_PRIVATE_KEY:-~/.ssh/id_rsa}"

echo "Building $ARCHITECTURE/$BENCHMARK on $INSTANCE_HOST"

# Upload repository to instance
rsync -avz -e "ssh -i $KEY_PATH -o StrictHostKeyChecking=no" \
    --exclude='.git' \
    --exclude='*.log' \
    ./ ec2-user@$INSTANCE_HOST:/home/ec2-user/aws-instance-benchmarks/

# Execute build on remote instance
ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no ec2-user@$INSTANCE_HOST << EOF
set -e
cd /home/ec2-user/aws-instance-benchmarks
# Ensure proper permissions
chmod +x scripts/build-containers.sh

echo "=== Building $ARCHITECTURE/$BENCHMARK container on \$(uname -m) ===="
lscpu | head -10
echo "======================================"

# Build the specific container
./scripts/build-containers.sh build $ARCHITECTURE $BENCHMARK

# Show the built container
docker images | grep aws-instance-benchmarks || echo "No containers found"

# Test the container
echo "=== Testing container ==="
timeout 30s docker run --rm localhost/aws-instance-benchmarks/$BENCHMARK:$ARCHITECTURE || echo "Container test completed"
EOF

echo "Build complete for $ARCHITECTURE/$BENCHMARK on $INSTANCE_HOST"