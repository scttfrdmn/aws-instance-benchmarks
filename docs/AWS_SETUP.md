# AWS Setup for Instance Benchmarks

This document outlines the required AWS configuration, IAM policies, and infrastructure setup needed to run the AWS Instance Benchmarks tool.

## Prerequisites

### 1. AWS CLI v2 Configuration
The tool requires AWS CLI v2 to be installed and configured with the `aws` profile:

```bash
# Verify AWS CLI v2 is installed
aws --version

# Configure the 'aws' profile
aws configure --profile aws
```

### 2. Required AWS Resources

#### VPC and Networking
- **VPC**: Existing VPC with internet gateway
- **Subnet**: Public subnet with auto-assign public IP enabled
- **Security Group**: Allow SSH (port 22) and outbound internet access

#### S3 Bucket
- **Bucket Name**: `aws-instance-benchmarks-results`
- **Purpose**: Store benchmark results and logs
- **Lifecycle**: Configure lifecycle rules to manage costs

#### ECR Repository (Optional)
- **Registry**: `public.ecr.aws/aws-benchmarks` 
- **Purpose**: Store optimized benchmark containers

## IAM Policies

### 1. User/Role Policy for CLI Tool

The AWS profile used by the CLI tool requires the following policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EC2InstanceManagement",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceTypes",
                "ec2:DescribeInstanceTypeOfferings",
                "ec2:DescribeImages",
                "ec2:DescribeKeyPairs",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpcs",
                "ec2:RunInstances",
                "ec2:TerminateInstances",
                "ec2:CreateTags",
                "ec2:DescribeTags"
            ],
            "Resource": "*"
        },
        {
            "Sid": "IAMPassRole",
            "Effect": "Allow",
            "Action": [
                "iam:PassRole"
            ],
            "Resource": "arn:aws:iam::*:role/benchmark-instance-role"
        },
        {
            "Sid": "ECRAccess",
            "Effect": "Allow", 
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage"
            ],
            "Resource": "*"
        },
        {
            "Sid": "ECRPublicAccess",
            "Effect": "Allow",
            "Action": [
                "ecr-public:GetAuthorizationToken",
                "ecr-public:BatchCheckLayerAvailability", 
                "ecr-public:GetDownloadUrlForLayer",
                "ecr-public:BatchGetImage",
                "ecr-public:CreateRepository",
                "ecr-public:PutImage",
                "ecr-public:InitiateLayerUpload",
                "ecr-public:UploadLayerPart",
                "ecr-public:CompleteLayerUpload"
            ],
            "Resource": "*"
        },
        {
            "Sid": "S3ResultsAccess",
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::aws-instance-benchmarks-results",
                "arn:aws:s3:::aws-instance-benchmarks-results/*"
            ]
        }
    ]
}
```

### 2. Instance IAM Role Policy

Create an IAM role `benchmark-instance-role` with the following policy for EC2 instances:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "S3ResultsUpload",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:PutObjectAcl",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::aws-instance-benchmarks-results/*"
            ]
        },
        {
            "Sid": "CloudWatchLogs",
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "logs:DescribeLogStreams"
            ],
            "Resource": "*"
        },
        {
            "Sid": "ECRAccess",
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer", 
                "ecr:BatchGetImage"
            ],
            "Resource": "*"
        },
        {
            "Sid": "ECRPublicAccess",
            "Effect": "Allow",
            "Action": [
                "ecr-public:GetAuthorizationToken",
                "ecr-public:BatchCheckLayerAvailability",
                "ecr-public:GetDownloadUrlForLayer",
                "ecr-public:BatchGetImage"
            ],
            "Resource": "*"
        }
    ]
}
```

### 3. Instance Trust Policy

The `benchmark-instance-role` also needs this trust policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
```

## Setup Commands

### 1. Create S3 Bucket
```bash
aws s3 mb s3://aws-instance-benchmarks-results --profile aws
aws s3api put-bucket-versioning \
    --bucket aws-instance-benchmarks-results \
    --versioning-configuration Status=Enabled \
    --profile aws
```

### 2. Create IAM Role and Instance Profile
```bash
# Create the role
aws iam create-role \
    --role-name benchmark-instance-role \
    --assume-role-policy-document file://instance-trust-policy.json \
    --profile aws

# Attach the policy
aws iam put-role-policy \
    --role-name benchmark-instance-role \
    --policy-name BenchmarkInstancePolicy \
    --policy-document file://instance-policy.json \
    --profile aws

# Create instance profile
aws iam create-instance-profile \
    --instance-profile-name benchmark-instance-profile \
    --profile aws

# Add role to instance profile
aws iam add-role-to-instance-profile \
    --instance-profile-name benchmark-instance-profile \
    --role-name benchmark-instance-role \
    --profile aws
```

### 3. Create Security Group
```bash
# Create security group
aws ec2 create-security-group \
    --group-name benchmark-security-group \
    --description "Security group for benchmark instances" \
    --vpc-id vpc-xxxxxxxxx \
    --profile aws

# Allow SSH access (adjust source as needed)
aws ec2 authorize-security-group-ingress \
    --group-id sg-xxxxxxxxx \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0 \
    --profile aws

# Allow all outbound traffic (usually default)
aws ec2 authorize-security-group-egress \
    --group-id sg-xxxxxxxxx \
    --protocol all \
    --cidr 0.0.0.0/0 \
    --profile aws
```

## Usage Examples

### Basic Benchmark Run
```bash
./aws-benchmark-collector run \
    --instance-types m7i.large,c7g.large \
    --region us-east-1 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx \
    --benchmarks stream
```

### Skip Quota Checks
```bash
./aws-benchmark-collector run \
    --instance-types m7i.large \
    --skip-quota-check \
    --region us-east-1 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx
```

### Multiple Regions
```bash
# Run in us-east-1
./aws-benchmark-collector run \
    --instance-types m7i.large \
    --region us-east-1 \
    --key-pair my-key-pair \
    --security-group sg-xxxxxxxxx \
    --subnet subnet-xxxxxxxxx

# Run in eu-west-1
./aws-benchmark-collector run \
    --instance-types m7i.large \
    --region eu-west-1 \
    --key-pair my-key-pair-eu \
    --security-group sg-yyyyyyyyy \
    --subnet subnet-yyyyyyyyy
```

## Quota Management

### Built-in Quota Handling
The tool includes quota management features:

1. **Pre-flight Checks**: Validates running instance counts before launch
2. **Quota Error Detection**: Automatically detects and handles quota errors
3. **Skip Mechanisms**: `--skip-quota-check` flag to bypass validation
4. **Graceful Degradation**: Continues with other instance types on quota errors

### Common Quota Limits
- **vCPU Limits**: Each instance family has separate vCPU limits
- **Spot Instance Limits**: Lower limits for spot instances
- **Regional Limits**: Limits are per-region

### Requesting Quota Increases
```bash
# List current quotas
aws service-quotas list-service-quotas \
    --service-code ec2 \
    --profile aws

# Request quota increase (example for On-Demand instances)
aws service-quotas request-service-quota-increase \
    --service-code ec2 \
    --quota-code L-1216C47A \
    --desired-value 100 \
    --profile aws
```

## Cost Management

### Cost Optimization Features
- **Automatic Termination**: Instances are terminated after benchmark completion
- **Spot Instance Support**: Use spot instances for cost savings (future feature)
- **Resource Tagging**: All resources tagged for cost tracking

### Monitoring Costs
```bash
# Set up cost alerts
aws budgets create-budget \
    --account-id 123456789012 \
    --budget file://budget.json \
    --profile aws
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Check IAM policies and role assignments
2. **Quota Exceeded**: Use `--skip-quota-check` or request increases
3. **Network Issues**: Verify VPC, subnet, and security group configuration
4. **Container Pull Errors**: Check ECR permissions and connectivity

### Debug Commands
```bash
# Test AWS configuration
aws sts get-caller-identity --profile aws

# Check EC2 permissions
aws ec2 describe-instances --profile aws --max-items 1

# Test S3 access
aws s3 ls s3://aws-instance-benchmarks-results --profile aws
```

## Security Considerations

1. **IAM Principle of Least Privilege**: Only grant necessary permissions
2. **Network Security**: Use appropriate security groups and NACLs
3. **Data Encryption**: Enable S3 bucket encryption
4. **Access Logging**: Enable CloudTrail for audit logging
5. **Resource Cleanup**: Ensure automatic termination of instances

## Compliance

For regulated environments:
- **Data Residency**: Run benchmarks in compliant regions
- **Audit Logging**: Enable comprehensive CloudTrail logging
- **Encryption**: Use KMS keys for S3 and EBS encryption
- **Network Isolation**: Use private subnets and VPC endpoints where required