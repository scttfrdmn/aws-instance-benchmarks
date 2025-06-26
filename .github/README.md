# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated benchmark collection and data processing.

## Workflows

### 1. Benchmark Collection (`benchmark-collection.yml`)

**Purpose**: Automated execution of performance benchmarks on AWS EC2 instances

**Triggers**:
- **Scheduled**: Weekly on Mondays at 2 AM UTC
- **Manual**: Via workflow dispatch with custom parameters

**Default Instance Types**:
- `m7i.large` (Intel, general purpose)
- `c7i.large` (Intel, compute optimized) 
- `r7i.large` (Intel, memory optimized)
- `m7g.large` (Graviton, general purpose)
- `c7g.large` (Graviton, compute optimized)

**Manual Execution**:
```bash
# GitHub CLI
gh workflow run benchmark-collection.yml \
  -f instance_types="m7i.xlarge,c7g.2xlarge" \
  -f region="us-west-2"

# Via GitHub UI: Actions → Benchmark Collection → Run workflow
```

## Setup Requirements

### 1. AWS IAM Role (Required)

Create an IAM role for GitHub Actions OIDC authentication:

```bash
# Replace ACCOUNT_ID and USERNAME/REPO with your values
aws iam create-role --role-name GitHubActionsBenchmarkRole --assume-role-policy-document '{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"},
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {"token.actions.githubusercontent.com:aud": "sts.amazonaws.com"},
      "StringLike": {"token.actions.githubusercontent.com:sub": "repo:USERNAME/aws-instance-benchmarks:*"}
    }
  }]
}'

# Attach the benchmark policy
aws iam attach-role-policy \
  --role-name GitHubActionsBenchmarkRole \
  --policy-arn arn:aws:iam::ACCOUNT_ID:policy/AWSInstanceBenchmarksPolicy
```

### 2. GitHub Secrets (Required)

Add to Repository Settings → Secrets and variables → Actions:

```
AWS_ROLE_ARN = arn:aws:iam::ACCOUNT_ID:role/GitHubActionsBenchmarkRole
```

### 3. OIDC Provider (One-time setup)

If not already configured in your AWS account:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

## Integration with ComputeCompass

The workflows generate JSON data accessible via GitHub Raw URLs:

```typescript
// Example integration
const response = await fetch(
  'https://raw.githubusercontent.com/USERNAME/aws-instance-benchmarks/main/data/processed/latest/memory-benchmarks.json'
)
const benchmarkData = await response.json()
```

## Security Features

- **Secure Authentication**: OIDC-based AWS access (no stored credentials)
- **Least Privilege**: IAM role with minimal required permissions
- **Private Data**: S3 bucket secured, public access via GitHub Pages only
- **Resource Cleanup**: Automatic cleanup of temporary AWS resources
- **Audit Trail**: Complete workflow execution logs

## Development

To test workflows locally:

```bash
# Install act (GitHub Actions local runner)
brew install act

# Run workflow locally (dry run)
act workflow_dispatch -W .github/workflows/benchmark-collection.yml --dryrun
```

## Monitoring

- **Workflow Status**: GitHub Actions tab
- **AWS Costs**: CloudWatch dashboard "AWSInstanceBenchmarks"  
- **Results**: Check `data/processed/latest/` directory after completion
- **Failures**: Review workflow logs for debugging

## Next Steps

1. Configure AWS IAM role and GitHub secrets
2. Test with manual workflow dispatch
3. Verify results in S3 and GitHub repository
4. Enable scheduled execution for automated collection