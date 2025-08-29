# Multi-Architecture Container Deployment Guide

## Manual Native Building on AWS EC2 Instances

To properly build and validate the Phase 2 vendor-optimized containers, we need to compile on the target architectures using actual AWS EC2 instances.

## Build Matrix

### 1. Intel oneAPI Container
**Target Platform:** Intel x86_64  
**Build Instance:** c7i.large (Intel Sapphire Rapids)  
**Container:** `aws-instance-benchmarks:intel-v2.1`

```bash
# Launch Intel instance
aws ec2 run-instances \
  --image-id ami-0abcdef1234567890 \
  --instance-type c7i.large \
  --key-name your-keypair \
  --security-groups default \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=benchmark-build-intel}]'

# SSH and build
ssh -i ~/.ssh/your-key.pem ec2-user@<intel-instance-ip>
sudo yum update -y
sudo yum install -y docker git
sudo systemctl start docker
sudo usermod -a -G docker ec2-user

# Clone repo and build Intel container
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks
docker build -t aws-instance-benchmarks:intel-v2.1 -f builds/intel/oneapi-optimized/Dockerfile .

# Test Intel optimizations
docker run --rm aws-instance-benchmarks:intel-v2.1 run-benchmark-intel.sh stream
```

### 2. AMD AOCC Container
**Target Platform:** AMD x86_64  
**Build Instance:** c7a.large (AMD Zen 4)  
**Container:** `aws-instance-benchmarks:amd-v2.1`

```bash
# Launch AMD instance  
aws ec2 run-instances \
  --image-id ami-0abcdef1234567890 \
  --instance-type c7a.large \
  --key-name your-keypair \
  --security-groups default \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=benchmark-build-amd}]'

# SSH and build
ssh -i ~/.ssh/your-key.pem ec2-user@<amd-instance-ip>
sudo yum update -y
sudo yum install -y docker git
sudo systemctl start docker
sudo usermod -a -G docker ec2-user

# Clone repo and build AMD container
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks
docker build -t aws-instance-benchmarks:amd-v2.1 -f builds/amd/aocc-optimized/Dockerfile .

# Test AMD optimizations
docker run --rm aws-instance-benchmarks:amd-v2.1 run-benchmark-amd.sh stream
```

### 3. ARM64 Graviton Container
**Target Platform:** ARM64  
**Build Instance:** c7g.large (Graviton 3)  
**Container:** `aws-instance-benchmarks:graviton-v2.1`

```bash
# Launch Graviton instance
aws ec2 run-instances \
  --image-id ami-0abcdef1234567890 \
  --instance-type c7g.large \
  --key-name your-keypair \
  --security-groups default \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=benchmark-build-graviton}]'

# SSH and build
ssh -i ~/.ssh/your-key.pem ec2-user@<graviton-instance-ip>
sudo yum update -y
sudo yum install -y docker git
sudo systemctl start docker  
sudo usermod -a -G docker ec2-user

# Clone repo and build Graviton container (universal works on ARM64)
git clone https://github.com/scttfrdmn/aws-instance-benchmarks.git
cd aws-instance-benchmarks
docker build -t aws-instance-benchmarks:graviton-v2.1 -f builds/universal/comprehensive/Dockerfile .

# Test ARM64 optimizations
docker run --rm aws-instance-benchmarks:graviton-v2.1 run-benchmark.sh stream
```

## Push to Public Registry

After building on each platform:

```bash
# Tag for multi-arch registry
docker tag aws-instance-benchmarks:intel-v2.1 public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:intel-v2.1
docker tag aws-instance-benchmarks:amd-v2.1 public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:amd-v2.1
docker tag aws-instance-benchmarks:graviton-v2.1 public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:graviton-v2.1

# Push platform-specific images
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws
docker push public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:intel-v2.1
docker push public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:amd-v2.1
docker push public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:graviton-v2.1
```

## Performance Validation

### Validation Matrix
| Instance Family | Container Variant | Expected Improvement |
|----------------|------------------|-------------------|
| c7i, m7i, r7i | intel-v2.1 | 20-40% over universal |
| c7a, m7a, r7a | amd-v2.1 | 15-30% over universal |
| c7g, m7g, r7g | graviton-v2.1 | Native ARM64 performance |

### Validation Commands
```bash
# Run comprehensive benchmark suite on each instance
docker run --rm public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:intel-v2.1 > intel-results.json
docker run --rm public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:amd-v2.1 > amd-results.json  
docker run --rm public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:graviton-v2.1 > graviton-results.json

# Compare against universal baseline
docker run --rm public.ecr.aws/f8g1e7l5/aws-instance-benchmarks:universal-v1.0 > universal-baseline.json
```

## Next Steps

1. **Launch Build Instances**: Create c7i.large, c7a.large, c7g.large instances
2. **Native Compilation**: Build containers on target architectures  
3. **Performance Testing**: Validate optimization improvements
4. **Registry Deployment**: Push multi-arch containers to public ECR
5. **Documentation**: Update usage guides with platform-specific instructions

This manual approach ensures we get true native performance on each target architecture.