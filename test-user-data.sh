#!/bin/bash
yum update -y
yum install -y docker
systemctl start docker
systemctl enable docker

# Get ECR login token and login to registry
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 942542972736.dkr.ecr.us-east-1.amazonaws.com

# Pull and run the benchmark container
docker run --rm 942542972736.dkr.ecr.us-east-1.amazonaws.com/aws-benchmarks/stream:universal > /tmp/stream-results.txt 2>&1

# Upload results to S3
aws s3 cp /tmp/stream-results.txt s3://aws-instance-benchmarks-data/test-results/stream-$(date +%Y%m%d-%H%M%S).txt

# Signal completion
echo "Benchmark test completed successfully" > /tmp/test-status.txt
aws s3 cp /tmp/test-status.txt s3://aws-instance-benchmarks-data/test-results/status-$(date +%Y%m%d-%H%M%S).txt