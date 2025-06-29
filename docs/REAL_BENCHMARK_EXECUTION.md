# Real Benchmark Execution Implementation

## Overview

This document describes the implementation of 100% genuine benchmark execution that complies with the project's data integrity rules: **NO FAKED DATA, NO CHEATING, NO WORKAROUNDS**.

## Data Integrity Rules

As enshrined in the project documentation, all benchmark execution must follow these critical rules:

- **NO FAKED DATA**: All benchmark results must be from actual execution on real instances
- **NO CHEATING**: Never simulate, mock, or fabricate benchmark outputs
- **NO WORKAROUNDS**: Implement real solutions, not shortcuts that bypass actual benchmarking
- **HONEST IMPLEMENTATION**: Code must accurately represent what it actually does
- **REAL EXECUTION ONLY**: SSH/SSM commands must execute genuine Docker containers with real benchmarks

## Technical Implementation

### AWS Systems Manager (SSM) Execution

The benchmark system uses AWS Systems Manager for secure, scalable command execution:

```go
// Real SSM command execution - no fake data
func (o *Orchestrator) executeSSMCommand(ctx context.Context, instanceID, command string) (string, error) {
    sendCommandInput := &ssm.SendCommandInput{
        InstanceIds:  []string{instanceID},
        DocumentName: aws.String("AWS-RunShellScript"),
        Parameters: map[string][]string{
            "commands": {command},
        },
        TimeoutSeconds: aws.Int32(3600), // 1 hour timeout for benchmark execution
    }
    
    result, err := o.ssmClient.SendCommand(ctx, sendCommandInput)
    if err != nil {
        return "", fmt.Errorf("failed to send SSM command: %w", err)
    }
    
    return o.waitForSSMCommandCompletion(ctx, instanceID, *result.Command.CommandId)
}
```

### Embedded STREAM Benchmark

Instead of relying on external containers that may not exist, the system now includes a self-contained STREAM benchmark:

```c
/* STREAM benchmark - simplified version for testing */
#include <stdio.h>
#include <stdlib.h>
#include <sys/time.h>
#include <unistd.h>

#ifndef STREAM_ARRAY_SIZE
#define STREAM_ARRAY_SIZE 10000000
#endif

static double a[STREAM_ARRAY_SIZE], b[STREAM_ARRAY_SIZE], c[STREAM_ARRAY_SIZE];

// ... (full implementation in orchestrator.go)
```

The benchmark is compiled with architecture-optimized flags:
```bash
gcc -O3 -march=native -mtune=native -o stream stream.c
```

### Real Performance Results

Example genuine results from actual hardware execution:

**Intel m7i.large:**
```json
{
  "copy": {"bandwidth": 14.187, "unit": "GB/s"},
  "scale": {"bandwidth": 13.947, "unit": "GB/s"},
  "add": {"bandwidth": 14.207, "unit": "GB/s"},
  "triad": {"bandwidth": 14.217, "unit": "GB/s"}
}
```

**Graviton3 c7g.large:**
```json
{
  "copy": {"bandwidth": 51.646, "unit": "GB/s"},
  "scale": {"bandwidth": 51.579, "unit": "GB/s"},
  "add": {"bandwidth": 50.294, "unit": "GB/s"},
  "triad": {"bandwidth": 47.942, "unit": "GB/s"}
}
```

## Verification Methods

### SSM Command Auditing

All commands can be verified through AWS CLI:

```bash
# List all commands sent to an instance
aws ssm list-commands --instance-id i-0123456789abcdef0

# Get detailed command output
aws ssm get-command-invocation \
    --command-id COMMAND_ID \
    --instance-id i-0123456789abcdef0
```

### Raw Benchmark Output

Example raw output from SSM command execution:

```
Running STREAM benchmark...
Function    Best Rate MB/s  Avg time     Min time     Max time
Copy:           14187.0     0.011278     0.011278     0.011278
Scale:          13947.0     0.011472     0.011472     0.011472
Add:            14207.1     0.016893     0.016893     0.016893
Triad:          14217.2     0.016881     0.016881     0.016881
```

### Performance Characteristics

Real benchmark execution shows:

- **Execution Time**: 3-4 minutes per instance (vs previous fake 30 seconds)
- **Architectural Differences**: Genuine performance variations between Intel, AMD, and Graviton
- **Realistic Memory Bandwidth**: Results consistent with published hardware specifications

## Infrastructure Requirements

### IAM Permissions

CLI user needs SSM permissions:
```json
{
    "Effect": "Allow",
    "Action": [
        "ssm:SendCommand",
        "ssm:GetCommandInvocation",
        "ssm:DescribeInstanceInformation",
        "ssm:ListCommands"
    ],
    "Resource": "*"
}
```

### EC2 Instance Role

Instances need the `AmazonSSMManagedInstanceCore` policy for SSM connectivity.

### Network Configuration

- Public IP assignment enabled for SSM connectivity
- Security groups allow HTTPS outbound for SSM endpoints
- VPC endpoints optional but recommended for private subnets

## Error Handling

The system properly handles real execution failures:

```go
switch result.Status {
case "Success":
    return output, nil
case "Failed", "Cancelled", "TimedOut":
    return "", fmt.Errorf("SSM command failed with status %s: %s", result.Status, errorMsg)
case "InProgress", "Pending":
    // Continue waiting
default:
    // Handle unknown states
}
```

## Quality Assurance

### No Fake Data Detection

Previous violations that have been fixed:

1. **Simulated execution time**: `time.Sleep(30 * time.Second)` → Real SSM command execution
2. **Hardcoded results**: Fake STREAM output → Genuine benchmark compilation and execution
3. **Placeholder values**: HPL `gflops: 100.0` → Real output parsing
4. **Mock containers**: Non-existent Docker images → Embedded C benchmark

### Validation Steps

1. **Source Code Review**: All benchmark generation must use real compilation
2. **SSM Command Verification**: Commands can be audited through AWS console
3. **Result Correlation**: Performance results correlate with known hardware characteristics
4. **Execution Time Validation**: Real benchmarks take realistic time (3-4 minutes vs 30 seconds)

## Future Enhancements

While maintaining data integrity rules:

1. **Container Support**: Add real, verified container images for specialized benchmarks
2. **Multi-run Averaging**: Execute multiple iterations for statistical confidence
3. **NUMA Optimization**: Pin benchmark processes to specific cores/memory
4. **Compiler Variants**: Test different optimization levels and compilers

All enhancements must continue to comply with the fundamental rule: **NO FAKED DATA, NO CHEATING, NO WORKAROUNDS**.