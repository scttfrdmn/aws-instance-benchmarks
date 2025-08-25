#!/bin/bash

# Graviton4-Optimized STREAM Benchmark Runner
# Optimized for AWS Graviton4 (ARM Neoverse-V2)

set -euo pipefail

echo "🚀 AWS Graviton4 STREAM Benchmark"
echo "===================================="

# System information
echo "📊 SYSTEM INFORMATION"
echo "======================"
echo "Date: $(date)"
echo "Hostname: $(hostname)"
echo "Architecture: $(uname -m)"
echo "Kernel: $(uname -r)"

# CPU Information
echo -e "\n🔧 CPU INFORMATION"
echo "=================="
lscpu | grep -E "(Model name|Architecture|CPU\(s\)|Thread|Core|Socket|Cache|Flags)"

# NUMA Information
echo -e "\n🧠 NUMA TOPOLOGY"
echo "=================="
numactl --hardware

# Memory Information
echo -e "\n💾 MEMORY INFORMATION"
echo "===================="
cat /proc/meminfo | grep -E "(MemTotal|MemAvailable|MemFree)"

# Determine optimal array size (60% of total memory / 3 arrays)
TOTAL_MEM_KB=$(cat /proc/meminfo | grep MemTotal | awk '{print $2}')
TOTAL_MEM_BYTES=$((TOTAL_MEM_KB * 1024))
ARRAY_SIZE_BYTES=$((TOTAL_MEM_BYTES * 60 / 100 / 3))  # 60% of memory / 3 arrays
ARRAY_SIZE_ELEMENTS=$((ARRAY_SIZE_BYTES / 8))  # 8 bytes per double

# Bounds checking
MIN_ELEMENTS=10000000    # 10M elements minimum
MAX_ELEMENTS=500000000   # 500M elements maximum

if [ $ARRAY_SIZE_ELEMENTS -lt $MIN_ELEMENTS ]; then
    ARRAY_SIZE_ELEMENTS=$MIN_ELEMENTS
elif [ $ARRAY_SIZE_ELEMENTS -gt $MAX_ELEMENTS ]; then
    ARRAY_SIZE_ELEMENTS=$MAX_ELEMENTS
fi

echo -e "\n⚙️  BENCHMARK CONFIGURATION"
echo "=========================="
echo "Array Size: $ARRAY_SIZE_ELEMENTS elements"
echo "Memory Usage: $((ARRAY_SIZE_ELEMENTS * 8 * 3 / 1024 / 1024)) MB"
echo "Compiler Flags: $CFLAGS"

# Set NUMA policy for optimal performance
echo -e "\n🎯 NUMA OPTIMIZATION"
echo "===================="
export OMP_NUM_THREADS=$(nproc)
export OMP_PROC_BIND=true
export OMP_PLACES=cores

# Run benchmark with NUMA optimizations
echo -e "\n▶️  RUNNING STREAM BENCHMARK"
echo "============================"

# Run on local NUMA node for best performance
numactl --localalloc --cpunodebind=0 ./stream

echo -e "\n✅ BENCHMARK COMPLETE"
echo "===================="
echo "Graviton4 STREAM benchmark completed successfully"