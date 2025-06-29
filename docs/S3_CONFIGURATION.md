# S3 Storage Configuration

## Overview

The AWS Instance Benchmarks project uses Amazon S3 for storing benchmark results with flexible bucket configuration to support different regions and deployment scenarios.

## Configuration Options

### Command Line Configuration

The S3 bucket can be configured via the `--s3-bucket` flag:

```bash
# Use specific bucket
./aws-benchmark-collector run \
    --s3-bucket my-custom-benchmark-bucket \
    --region us-west-2 \
    --instance-types m7i.large

# Use regional default bucket (recommended)
./aws-benchmark-collector run \
    --region us-west-2 \
    --instance-types m7i.large
    # Defaults to: aws-instance-benchmarks-data-us-west-2
```

### Default Bucket Naming

When no `--s3-bucket` is specified, the tool automatically generates a bucket name:
- **Pattern**: `aws-instance-benchmarks-data-{region}`
- **Examples**: 
  - `aws-instance-benchmarks-data-us-east-1`
  - `aws-instance-benchmarks-data-us-west-2`
  - `aws-instance-benchmarks-data-eu-west-1`

## Bucket Setup

### Automatic Bucket Creation

The tool automatically creates the bucket if it doesn't exist:

```bash
# This will create the bucket in the specified region
aws s3 mb s3://aws-instance-benchmarks-data-us-west-2 --region us-west-2
```

### Manual Bucket Setup

For production deployments, create the bucket manually with appropriate policies:

```bash
# Create bucket
aws s3 mb s3://my-benchmark-bucket --region us-west-2

# Set lifecycle policy (optional)
aws s3api put-bucket-lifecycle-configuration \
    --bucket my-benchmark-bucket \
    --lifecycle-configuration file://lifecycle-policy.json
```

## Storage Organization

Benchmark results are stored with a structured key pattern:

```
s3://bucket-name/
├── raw/
│   └── YYYY/MM/DD/
│       └── region/
│           └── instance-type/
│               └── timestamp-uuid.json
└── processed/
    └── aggregated-results/
```

## AWS Profile Configuration

The S3 storage uses the configured AWS profile:

```go
// Uses 'aws' profile for testing/benchmarking
cfg, err := config.LoadDefaultConfig(ctx,
    config.WithSharedConfigProfile("aws"),
    config.WithRegion(region),
    config.WithRetryMaxAttempts(storageConfig.RetryAttempts),
)
```

### Profile Setup

```bash
# Set up AWS profile for benchmarking
aws configure --profile aws
# AWS Access Key ID: YOUR_ACCESS_KEY
# AWS Secret Access Key: YOUR_SECRET_KEY  
# Default region name: us-east-1
# Default output format: json
```

## Permissions Required

The AWS profile needs the following S3 permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket",
                "s3:GetBucketLocation"
            ],
            "Resource": [
                "arn:aws:s3:::aws-instance-benchmarks-data-*",
                "arn:aws:s3:::aws-instance-benchmarks-data-*/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:CreateBucket"
            ],
            "Resource": "arn:aws:s3:::aws-instance-benchmarks-data-*"
        }
    ]
}
```

## Troubleshooting

### Common Issues

1. **403 Forbidden Error**
   - Check AWS profile has S3 permissions
   - Verify bucket exists and is accessible
   - Ensure bucket is in the correct region

2. **301 Moved Permanently**
   - Bucket exists in different region than specified
   - Use correct region flag: `--region us-east-1`

3. **Bucket Access Validation Failed**
   - Check AWS credentials are configured
   - Verify the AWS profile is correct
   - Test bucket access: `aws s3 ls s3://bucket-name --profile aws`

### Debug Commands

```bash
# Test AWS configuration
aws sts get-caller-identity --profile aws

# Test S3 access
aws s3 ls --profile aws

# Test specific bucket access
aws s3 ls s3://aws-instance-benchmarks-data --profile aws
```

## Configuration Examples

### Single Region Deployment
```bash
./aws-benchmark-collector run \
    --region us-east-1 \
    --s3-bucket aws-instance-benchmarks-data \
    --instance-types m7i.large,c7g.large
```

### Multi-Region Deployment
```bash
# East Coast
./aws-benchmark-collector run \
    --region us-east-1 \
    --s3-bucket aws-instance-benchmarks-data-us-east-1 \
    --instance-types m7i.large

# West Coast  
./aws-benchmark-collector run \
    --region us-west-2 \
    --s3-bucket aws-instance-benchmarks-data-us-west-2 \
    --instance-types m7i.large
```

### Development vs Production
```bash
# Development
./aws-benchmark-collector run \
    --s3-bucket dev-benchmarks-$(whoami) \
    --region us-west-2

# Production
./aws-benchmark-collector run \
    --s3-bucket prod-aws-instance-benchmarks \
    --region us-east-1
```