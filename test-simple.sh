#!/bin/bash
# Simple test script to debug JSON parsing

echo "Testing basic JSON generation..."

# Test 1: Simple JSON
echo '{"test": 123}' | jq empty && echo "Test 1: PASS" || echo "Test 1: FAIL"

# Test 2: Number parsing
echo '{"number": 123.45}' | jq empty && echo "Test 2: PASS" || echo "Test 2: FAIL"

# Test 3: String parsing  
echo '{"string": "hello"}' | jq empty && echo "Test 3: PASS" || echo "Test 3: FAIL"

# Test 4: Complex JSON
cat << EOF | jq empty && echo "Test 4: PASS" || echo "Test 4: FAIL"
{
  "results": {
    "value": 123.45,
    "status": "completed"
  }
}
EOF

echo "JSON tests complete"