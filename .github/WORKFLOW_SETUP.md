# GitHub Actions Workflow Setup

This document explains how to configure the GitHub Actions workflows for automated benchmark collection and data processing.

## Required Setup

### 1. AWS IAM Role for GitHub Actions

Create an IAM role that GitHub Actions can assume using OIDC:

```bash
# Create trust policy for GitHub Actions OIDC
cat > github-actions-trust-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:USERNAME/aws-instance-benchmarks:*"
        }
      }
    }
  ]
}
EOF

# Create the role
aws iam create-role \
  --role-name GitHubActionsBenchmarkRole \
  --assume-role-policy-document file://github-actions-trust-policy.json

# Attach the benchmark policy we created earlier
aws iam attach-role-policy \
  --role-name GitHubActionsBenchmarkRole \
  --policy-arn arn:aws:iam::ACCOUNT_ID:policy/AWSInstanceBenchmarksPolicy
```

### 2. GitHub Repository Secrets

Add the following secrets to your GitHub repository:

1. Go to Repository Settings → Secrets and variables → Actions
2. Add these secrets:

```
AWS_ROLE_ARN = arn:aws:iam::ACCOUNT_ID:role/GitHubActionsBenchmarkRole
```

### 3. GitHub Pages Setup

1. Go to Repository Settings → Pages
2. Source: "GitHub Actions"
3. The Pages workflow will automatically deploy the benchmark data site

### 4. OIDC Provider Setup (One-time per AWS Account)

If not already configured, add GitHub as an OIDC provider:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

## Workflow Details

### 1. Benchmark Collection (`benchmark-collection.yml`)

**Triggers:**
- Scheduled: Weekly on Mondays at 2 AM UTC
- Manual: Via workflow dispatch with customizable parameters

**Parameters:**
- `instance_types`: Comma-separated list (default: m7i.large,c7i.large,r7i.large,m7g.large,c7g.large)
- `region`: AWS region (default: us-east-1)
- `iterations`: Number of runs per instance (default: 3)

**Process:**
1. Sets up Go environment and builds CLI tool
2. Configures AWS credentials via OIDC
3. Discovers VPC/subnet/security group configuration
4. Creates temporary EC2 key pair
5. Executes benchmarks on specified instance types
6. Downloads results from S3
7. Generates summary reports
8. Commits results to repository

### 2. Data Processing (`data-processing.yml`)

**Triggers:**
- Automatic: After successful benchmark collection
- Manual: Via workflow dispatch

**Process:**
1. Downloads all benchmark data from S3
2. Processes and aggregates results
3. Generates GitHub Pages site with API endpoints
4. Uploads processed data back to S3
5. Deploys to GitHub Pages for public access

### 3. Pages Deployment (`pages.yml`)

**Triggers:**
- Push to main branch (docs or processed data changes)
- Manual trigger

**Process:**
1. Builds static site from processed data
2. Deploys to GitHub Pages
3. Provides API endpoints for ComputeCompass integration

## Usage

### Manual Benchmark Collection

```bash
# Trigger benchmark collection for specific instances
gh workflow run benchmark-collection.yml \
  -f instance_types="m7i.xlarge,c7g.2xlarge" \
  -f region="us-west-2" \
  -f iterations="5"
```

### Monitoring

- **Workflow Status**: GitHub Actions tab
- **AWS Costs**: CloudWatch dashboard "AWSInstanceBenchmarks"
- **Data Quality**: Check S3 bucket for result files

### API Access

Once deployed, benchmark data is accessible via:

```
https://USERNAME.github.io/aws-instance-benchmarks/data/processed/latest/metadata.json
```

## Integration with ComputeCompass

The workflows generate JSON endpoints compatible with ComputeCompass:

```typescript
// Example integration
const response = await fetch('https://raw.githubusercontent.com/scttfrdmn/aws-instance-benchmarks/main/data/processed/latest/memory-benchmarks.json')
const data = await response.json()
```

## Security Considerations

1. **IAM Permissions**: Role follows least privilege principle
2. **S3 Access**: Bucket remains private, public access via GitHub Pages only
3. **Temporary Resources**: EC2 key pairs auto-deleted after use
4. **Instance Cleanup**: Failed instances automatically terminated

## Troubleshooting

### Common Issues

1. **OIDC Trust Relationship**: Verify GitHub OIDC provider exists
2. **AWS Permissions**: Check IAM role has required policies
3. **Instance Limits**: Verify EC2 service quotas
4. **Network Access**: Ensure default VPC/subnet configuration

### Debugging

- Check workflow logs in GitHub Actions
- Verify AWS CloudTrail for permission issues
- Monitor S3 bucket for result uploads
- Check EC2 console for stuck instances