#!/bin/bash

# AWS Instance Benchmarks Build Script
set -e

echo "🔨 Building AWS Instance Benchmarks CLI..."

# Check Go version
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22 or later."
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
echo "✅ Found Go $GO_VERSION"

# Build the main CLI tool
echo "📦 Building aws-benchmark-collector..."
go build -o aws-benchmark-collector ./cmd

echo "🧪 Running tests..."
go test ./...

echo "✅ Build completed successfully!"
echo ""
echo "🚀 Try these commands:"
echo "  ./aws-benchmark-collector --help"
echo "  ./aws-benchmark-collector analyze results/ --sort value_score"
echo "  ./aws-benchmark-collector schema validate results/"
echo ""
echo "📊 Open the web viewer:"
echo "  open tools/viewer/benchmark-viewer.html"