# Intel Sapphire Rapids Vector Instruction benchmark container
FROM ubuntu:24.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libomp-dev \
    util-linux \
    bc \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Sapphire Rapids optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
ENV CXXFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"

# Create workspace
WORKDIR /opt/vector-bench

# Copy vector benchmark source
COPY builds/intel-sapphirerapids/vector-benchmark/vector_benchmark.c .

# Compile the vector benchmark
RUN gcc $CFLAGS vector_benchmark.c -o vector_benchmark

# Create wrapper script
RUN echo '#!/bin/bash' > /usr/local/bin/run-vector-benchmark && \
    echo 'echo "Intel Sapphire Rapids Vector Instruction Benchmark"' >> /usr/local/bin/run-vector-benchmark && \
    echo 'echo "Testing AVX-512 VNNI/BF16 capabilities for AI/ML workloads"' >> /usr/local/bin/run-vector-benchmark && \
    echo 'cd /opt/vector-bench && ./vector_benchmark' >> /usr/local/bin/run-vector-benchmark && \
    chmod +x /usr/local/bin/run-vector-benchmark

ENTRYPOINT ["/usr/local/bin/run-vector-benchmark"]