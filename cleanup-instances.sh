#!/bin/bash

# AWS Instance Benchmarks - Cleanup Script
# Finds and terminates orphaned/zombie instances belonging to this project

PROFILE="aws"
REGION="us-west-2"

echo "🧹 AWS Instance Benchmarks - Cleanup Orphaned Instances"
echo "======================================================="

# Check AWS authentication
if ! aws --profile $PROFILE sts get-caller-identity > /dev/null 2>&1; then
    echo "❌ Error: AWS CLI not configured with profile '$PROFILE'"
    echo "Please run: aws configure --profile $PROFILE"
    exit 1
fi

echo "✅ AWS Authentication verified"
echo ""

# Find all instances tagged with our project
echo "🔍 Finding instances belonging to aws-instance-benchmarks project..."
BENCHMARK_INSTANCES=$(aws --profile $PROFILE ec2 describe-instances \
    --region $REGION \
    --filters \
        "Name=tag:Purpose,Values=aws-instance-benchmarks" \
        "Name=instance-state-name,Values=running,pending,stopping" \
    --query 'Reservations[*].Instances[*].[InstanceId,InstanceType,State.Name,LaunchTime,Tags[?Key==`Name`]|[0].Value]' \
    --output table)

if [ -z "$BENCHMARK_INSTANCES" ] || [ "$BENCHMARK_INSTANCES" = "None" ]; then
    echo "✅ No orphaned benchmark instances found"
    exit 0
fi

echo "$BENCHMARK_INSTANCES"
echo ""

# Get instance IDs only
INSTANCE_IDS=$(aws --profile $PROFILE ec2 describe-instances \
    --region $REGION \
    --filters \
        "Name=tag:Purpose,Values=aws-instance-benchmarks" \
        "Name=instance-state-name,Values=running,pending" \
    --query 'Reservations[*].Instances[*].InstanceId' \
    --output text)

if [ -z "$INSTANCE_IDS" ]; then
    echo "✅ No running benchmark instances to clean up"
    exit 0
fi

# Count instances
INSTANCE_COUNT=$(echo $INSTANCE_IDS | wc -w)
echo "⚠️  Found $INSTANCE_COUNT benchmark instances that may be orphaned"
echo ""

# Show which instances will be terminated
echo "📋 Instances to be terminated:"
for instance_id in $INSTANCE_IDS; do
    INSTANCE_INFO=$(aws --profile $PROFILE ec2 describe-instances \
        --region $REGION \
        --instance-ids $instance_id \
        --query 'Reservations[0].Instances[0].[InstanceType,LaunchTime,Tags[?Key==`Name`]|[0].Value]' \
        --output text)
    echo "  - $instance_id ($INSTANCE_INFO)"
done
echo ""

# Confirm before termination
read -p "❓ Terminate these $INSTANCE_COUNT instances? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Cleanup cancelled"
    exit 0
fi

# Terminate instances
echo "🗑️  Terminating benchmark instances..."
for instance_id in $INSTANCE_IDS; do
    echo "  Terminating $instance_id..."
    aws --profile $PROFILE ec2 terminate-instances \
        --region $REGION \
        --instance-ids $instance_id > /dev/null
    
    if [ $? -eq 0 ]; then
        echo "  ✅ $instance_id termination initiated"
    else
        echo "  ❌ Failed to terminate $instance_id"
    fi
done

echo ""
echo "🧹 Cleanup completed!"
echo "✅ $INSTANCE_COUNT benchmark instances marked for termination"
echo ""
echo "💡 Note: It may take a few minutes for instances to fully terminate."
echo "   You can check status with: aws --profile $PROFILE ec2 describe-instances --region $REGION --filters 'Name=tag:Purpose,Values=aws-instance-benchmarks'"