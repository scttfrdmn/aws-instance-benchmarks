#!/bin/bash
# Simplified STREAM test

cd /opt/benchmarks

echo "=== Testing STREAM benchmark execution ==="

# Run STREAM and capture output
stream_output=$(./stream/stream_benchmark 2>&1)
echo "STREAM execution completed"

# Parse results
copy_rate=$(echo "$stream_output" | grep "Copy:" | awk '{print $2}' || echo "0")
echo "Copy rate: $copy_rate"

# Test JSON generation
cat << EOF
{
  "stream_test": {
    "copy_rate": $copy_rate,
    "status": "completed"
  }
}
EOF