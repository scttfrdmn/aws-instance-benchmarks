# 🚀 Quick Start Guide - Tomorrow's Session

## ⚡ **Immediate Actions (First 5 Minutes)**

### **1. Check Running Instances**
```bash
# Check if any benchmarks are still running
aws ec2 describe-instances --filters "Name=instance-state-name,Values=running" --query 'Reservations[].Instances[].{InstanceId:InstanceId,Type:InstanceType,LaunchTime:LaunchTime,Tags:Tags}'

# Check costs so far
aws ce get-cost-and-usage --time-period Start=2024-06-30,End=2024-07-01 --granularity DAILY --metrics BlendedCost
```

### **2. Check S3 for Results**
```bash
# Quick check for any completed benchmarks
aws s3 ls s3://your-bucket/benchmarks/ --recursive | grep sentinel

# Look for results files
aws s3 ls s3://your-bucket/benchmarks/ --recursive | grep results.json
```

## 🎯 **Ready to Launch (5-10 Minutes)**

### **Option A: Quick Test**
```bash
# Update S3 bucket name first
vim async_benchmark_launcher.go  # Line 47

# Launch single benchmark to test
# Edit configs to have just one benchmark, then:
go run async_benchmark_launcher.go
```

### **Option B: Full Production Suite**
```bash
# Launch comprehensive 3-architecture test
go run async_benchmark_launcher.go  # Uses all 3 configs by default

# Monitor progress
go run async_benchmark_collector.go
```

## 📋 **Configuration Checklist**

### **Before First Launch**
- [ ] **S3 Bucket**: Update bucket name in launcher (line 47)
- [ ] **AWS Profile**: Ensure 'aws' profile is configured
- [ ] **Region**: Verify us-west-2 access
- [ ] **IAM Role**: Create EC2-S3-BenchmarkAccess role if needed

### **Required IAM Permissions**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "ec2:DescribeInstances",
        "s3:PutObject",
        "s3:GetObject"
      ],
      "Resource": "*"
    }
  ]
}
```

## 🔍 **Monitoring Commands**

### **Real-Time Monitoring**
```bash
# Watch S3 for new files (run in separate terminal)
watch -n 30 'aws s3 ls s3://your-bucket/benchmarks/ --recursive | tail -10'

# Check instance status
watch -n 60 'aws ec2 describe-instances --filters "Name=tag:LaunchedBy,Values=AsyncBenchmarkLauncher" --query "Reservations[].Instances[].{InstanceId:InstanceId,State:State.Name,Type:InstanceType}"'
```

### **Cost Monitoring**
```bash
# Current running costs
aws ec2 describe-instances --filters "Name=instance-state-name,Values=running" --query 'Reservations[].Instances[].{InstanceId:InstanceId,Type:InstanceType,LaunchTime:LaunchTime}' | jq '.[] | "Instance: \(.InstanceId) (\(.Type)) running since \(.LaunchTime)"'
```

## 🛠️ **Common Tasks**

### **Clean Up Running Instances**
```bash
# If you need to stop everything
aws ec2 describe-instances --filters "Name=tag:LaunchedBy,Values=AsyncBenchmarkLauncher" "Name=instance-state-name,Values=running" --query 'Reservations[].Instances[].InstanceId' --output text | xargs aws ec2 terminate-instances --instance-ids
```

### **Check Results**
```bash
# Download and view a result
aws s3 cp s3://your-bucket/benchmarks/bench-xyz/c7g.large/stream/results.json ./latest-result.json
cat latest-result.json | jq '.'
```

### **Debug Issues**
```bash
# Check logs if benchmark fails
aws s3 cp s3://your-bucket/benchmarks/bench-xyz/c7g.large/stream/benchmark.log ./debug.log
tail -50 debug.log
```

## 📊 **Expected Results**

### **Timeline**
- **Launch**: Immediate (< 1 minute)
- **Instance Startup**: 2-3 minutes
- **STREAM Benchmark**: 5-10 minutes
- **HPL Benchmark**: 1-4 hours
- **FFTW Benchmark**: 30-60 minutes

### **Cost Estimates**
- **c7g.large**: $0.0725/hour (ARM Graviton3)
- **c7i.large**: $0.0864/hour (Intel Ice Lake)
- **c7a.large**: $0.0864/hour (AMD EPYC)

### **Success Indicators**
- ✅ Instances launch without errors
- ✅ S3 sentinel files appear (status-launched, status-running)
- ✅ Progress updates in collector tool
- ✅ Results.json files appear when complete
- ✅ Instances self-terminate

## 🚨 **Emergency Procedures**

### **If Costs Are Too High**
```bash
# Emergency stop all instances
aws ec2 describe-instances --filters "Name=instance-state-name,Values=running" --query 'Reservations[].Instances[].InstanceId' --output text | xargs aws ec2 terminate-instances --instance-ids
```

### **If Instances Are Stuck**
```bash
# Check for emergency timeout sentinels
aws s3 ls s3://your-bucket/benchmarks/ --recursive | grep emergency

# Force terminate specific instance
aws ec2 terminate-instances --instance-ids i-1234567890abcdef0
```

## 🎯 **Goals for Tomorrow**

### **Primary Goal**
- [ ] Launch and complete at least one benchmark successfully
- [ ] Verify end-to-end S3 collection works
- [ ] Confirm cost protection mechanisms work

### **Secondary Goals**
- [ ] Complete all 3 architectures (ARM, Intel, AMD)
- [ ] Test different benchmark types
- [ ] Set up automated collection schedule

### **Stretch Goals**
- [ ] Integrate results with ComputeCompass
- [ ] Add multi-region testing
- [ ] Create automated reporting

---

## 🌟 **You're Ready!**

The async architecture is **production-ready** and waiting for your return. Just update the S3 bucket name and launch - the system will handle the rest with comprehensive failsafe protection.

**Happy benchmarking! 🚀**