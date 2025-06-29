# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability within AWS Instance Benchmarks, please follow these steps:

### 1. **Do NOT** create a public GitHub issue

Security vulnerabilities should be reported privately to allow us to fix them before they become public knowledge.

### 2. Send a report via email

Please email security reports to: **scott.friedman@[remove-this]gmail.com**

Include the following information:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Any suggested fixes (if available)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Varies based on severity (1-30 days)

### 4. Disclosure Policy

- We will acknowledge receipt of your vulnerability report
- We will confirm the vulnerability and determine its severity
- We will develop and test a fix
- We will prepare a security advisory
- We will release the fix and publish the advisory
- We will publicly acknowledge your responsible disclosure (if desired)

## Security Considerations

### AWS Credentials

This tool requires AWS credentials to function. Please follow these security best practices:

- **Never commit AWS credentials to the repository**
- Use IAM roles when running on EC2 instances
- Use AWS credential profiles or environment variables
- Follow the principle of least privilege for IAM permissions
- Regularly rotate access keys

### Container Security

The benchmark containers are designed with security in mind:

- Containers run with minimal privileges
- No sensitive data is stored in container images
- Base images are regularly updated for security patches

### Data Security

- Benchmark results contain only performance metrics
- No personal or sensitive information is collected
- All data is stored in accordance with AWS data protection guidelines

## Secure Development Practices

We follow these practices to ensure code security:

- Static security analysis with gosec
- Dependency vulnerability scanning
- Regular security updates for dependencies
- Code review requirements for all changes
- Automated security testing in CI/CD pipeline

## Bug Bounty Program

Currently, we do not have a formal bug bounty program. However, we greatly appreciate security researchers who responsibly disclose vulnerabilities and will publicly acknowledge their contributions (with permission).

## Questions?

If you have questions about this security policy, please contact us at the email address above.