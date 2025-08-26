# Native Container Build Workflow

## Overview

This document describes the proper workflow for building architecture-specific benchmark containers on their native AWS instances. This ensures optimal compiler optimizations and accurate performance measurements.

## Prerequisites

1. **IAM Role Setup**
   ```bash
   aws iam create-role --role-name EC2-SSM-Role --assume-role-policy-document '{
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Principal": {"Service": "ec2.amazonaws.com"},
       "Action": "sts:AssumeRole"
     }]
   }'
   aws iam attach-role-policy --role-name EC2-SSM-Role --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
   aws iam create-instance-profile --instance-profile-name EC2-SSM-Profile
   aws iam add-role-to-instance-profile --instance-profile-name EC2-SSM-Profile --role-name EC2-SSM-Role
   ```

2. **Repository with Committed Container Definitions**
   - All Dockerfiles and source files must be committed to the repository
   - Use proper base images (Ubuntu 24.04 for modern GCC support)
   - Architecture-specific compiler flags must be tested on target hardware

## Instance Launch Pattern

### 1. Get Latest AMIs and Network Info
```bash
# Get latest AMIs
AMI_X86=$(aws ec2 describe-images --owners amazon --filters "Name=name,Values=amzn2-ami-hvm-*" "Name=virtualization-type,Values=hvm" "Name=architecture,Values=x86_64" --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text)
AMI_ARM=$(aws ec2 describe-images --owners amazon --filters "Name=name,Values=amzn2-ami-hvm-*" "Name=virtualization-type,Values=hvm" "Name=architecture,Values=arm64" --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text)

# Get default VPC info
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query 'Vpcs[0].VpcId' --output text)
SUBNET_ID=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query 'Subnets[0].SubnetId' --output text)
SG_ID=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=default" "Name=vpc-id,Values=$VPC_ID" --query 'SecurityGroups[0].GroupId' --output text)
```

### 2. Launch Architecture-Specific Instances
```bash
# Intel Sapphire Rapids (7th gen)
INTEL_INSTANCE=$(aws ec2 run-instances \
  --image-id $AMI_X86 \
  --instance-type c7i.large \
  --security-group-ids $SG_ID \
  --subnet-id $SUBNET_ID \
  --iam-instance-profile Name=EC2-SSM-Profile \
  --user-data '#!/bin/bash
yum update -y
yum install -y git docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git /home/ec2-user/aws-instance-benchmarks
chown -R ec2-user:ec2-user /home/ec2-user/aws-instance-benchmarks' \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=benchmark-intel-sapphirerapids},{Key=Project,Value=aws-instance-benchmarks}]' \
  --query 'Instances[0].InstanceId' --output text)

# AMD Genoa (7th gen)  
AMD_INSTANCE=$(aws ec2 run-instances \
  --image-id $AMI_X86 \
  --instance-type c7a.large \
  --security-group-ids $SG_ID \
  --subnet-id $SUBNET_ID \
  --iam-instance-profile Name=EC2-SSM-Profile \
  --user-data '#!/bin/bash
yum update -y
yum install -y git docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git /home/ec2-user/aws-instance-benchmarks
chown -R ec2-user:ec2-user /home/ec2-user/aws-instance-benchmarks' \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=benchmark-amd-genoa},{Key=Project,Value=aws-instance-benchmarks}]' \
  --query 'Instances[0].InstanceId' --output text)

# Graviton4 (ARM)
GRAVITON_INSTANCE=$(aws ec2 run-instances \
  --image-id $AMI_ARM \
  --instance-type c7g.large \
  --security-group-ids $SG_ID \
  --subnet-id $SUBNET_ID \
  --iam-instance-profile Name=EC2-SSM-Profile \
  --user-data '#!/bin/bash
yum update -y
yum install -y git docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git /home/ec2-user/aws-instance-benchmarks
chown -R ec2-user:ec2-user /home/ec2-user/aws-instance-benchmarks' \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=benchmark-graviton4},{Key=Project,Value=aws-instance-benchmarks}]' \
  --query 'Instances[0].InstanceId' --output text)
```

### 3. Wait for Initialization
```bash
sleep 90  # Allow time for user-data script completion and SSM agent registration
```

## Container Build Pattern

### 1. Update Repository (if changes made)
```bash
COMMAND_ID=$(aws ssm send-command \
  --instance-ids $INTEL_INSTANCE \
  --document-name "AWS-RunShellScript" \
  --parameters 'commands=["cd /home/ec2-user/aws-instance-benchmarks", "git pull origin main"]' \
  --query 'Command.CommandId' --output text)

# Wait and check status
aws ssm get-command-invocation --command-id $COMMAND_ID --instance-id $INTEL_INSTANCE --query 'Status' --output text
```

### 2. Build Container Natively
```bash
COMMAND_ID=$(aws ssm send-command \
  --instance-ids $INTEL_INSTANCE \
  --document-name "AWS-RunShellScript" \
  --parameters 'commands=[
    "cd /home/ec2-user/aws-instance-benchmarks",
    "sudo docker build -t aws-instance-benchmarks/vector-benchmark:intel-sapphirerapids -f builds/intel-sapphirerapids/vector-benchmark/Dockerfile . && echo BUILD_SUCCESS"
  ]' \
  --query 'Command.CommandId' --output text)

# Monitor build progress
aws ssm get-command-invocation --command-id $COMMAND_ID --instance-id $INTEL_INSTANCE --query '[Status,StandardOutputContent]' --output text
```

### 3. Run Container and Capture Results
```bash
COMMAND_ID=$(aws ssm send-command \
  --instance-ids $INTEL_INSTANCE \
  --document-name "AWS-RunShellScript" \
  --parameters 'commands=[
    "sudo docker run --rm aws-instance-benchmarks/vector-benchmark:intel-sapphirerapids"
  ]' \
  --query 'Command.CommandId' --output text)

# Get benchmark results
aws ssm get-command-invocation --command-id $COMMAND_ID --instance-id $INTEL_INSTANCE --query 'StandardOutputContent' --output text
```

## Results Processing Pattern

### 1. Save Results with Metadata
```bash
# Create results directory structure
mkdir -p results/$(date +%Y-%m-%d)/intel-sapphirerapids/vector-benchmark/

# Save results with instance metadata
aws ssm get-command-invocation --command-id $COMMAND_ID --instance-id $INTEL_INSTANCE --query 'StandardOutputContent' --output text > results/$(date +%Y-%m-%d)/intel-sapphirerapids/vector-benchmark/raw-output.txt

# Save instance information
aws ec2 describe-instances --instance-ids $INTEL_INSTANCE --query 'Reservations[].Instances[0].[InstanceType,Placement.AvailabilityZone,CpuOptions]' --output json > results/$(date +%Y-%m-%d)/intel-sapphirerapids/vector-benchmark/instance-info.json
```

### 2. Commit Results
```bash
git add results/$(date +%Y-%m-%d)/
git commit -m "Add $(date +%Y-%m-%d) benchmark results for Intel Sapphire Rapids vector instructions

Native container execution on c7i.large instance ensuring architecture-specific optimizations.

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
git push
```

## Cleanup Pattern

### 1. Terminate Instances
```bash
aws ec2 terminate-instances --instance-ids $INTEL_INSTANCE $AMD_INSTANCE $GRAVITON_INSTANCE
```

### 2. Verify Termination
```bash
aws ec2 describe-instances --instance-ids $INTEL_INSTANCE $AMD_INSTANCE $GRAVITON_INSTANCE --query 'Reservations[].Instances[].[InstanceId,State.Name]' --output table
```

## Critical Requirements

### 1. Base Image Selection
- **Use Ubuntu 24.04** for modern GCC support (supports -march=sapphirerapids)
- **Never use Ubuntu 22.04 or older** for modern architectures
- Verify compiler support before committing containers

### 2. Architecture Flags
- **Intel Sapphire Rapids**: `-march=sapphirerapids -mavx512vnni -mavx512bf16`
- **AMD Genoa**: `-march=znver4 -mavx2 -mavx512f`
- **Graviton4**: `-mcpu=neoverse-v2 -mtune=neoverse-v2`

### 3. Never Cross-Compile
- **Always build on target architecture**
- **Never build x86 containers on ARM or vice versa**
- **Never use generic optimization flags**

### 4. Validation Steps
- Verify CPU features match compiler flags
- Test container execution before committing
- Save complete benchmark output and instance metadata

## Troubleshooting

### Build Failures
1. Check GCC version: `gcc --version`
2. Test architecture flags: `gcc -march=sapphirerapids -Q --help=target`
3. Verify CPU features: `lscpu | grep -E "Model name|Flags"`
4. Update base image if compiler too old

### Runtime Failures  
1. Check available memory: `free -h`
2. Monitor during execution: `docker stats`
3. Review container logs: `docker logs <container>`
4. Test with reduced workload sizes

### SSM Issues
1. Verify IAM role has SSM policies
2. Check instance status: `aws ssm describe-instance-information`
3. Wait longer for agent registration (up to 10 minutes)
4. Verify VPC has internet access for SSM endpoints

## Cost Optimization

- Use **c7i.large** for Intel (cost-effective for builds)
- Use **c7a.large** for AMD (cost-effective for builds)  
- Use **c7g.large** for Graviton (cost-effective for builds)
- Terminate instances immediately after builds complete
- Build multiple containers per instance session when possible

---

**Last Updated**: 2025-08-26
**Validated On**: Intel Sapphire Rapids (c7i.large), AMD Genoa (c7a.large), Graviton4 (c7g.large)