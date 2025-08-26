# Graviton4 Cache Hierarchy benchmark container (lightweight test version)
FROM ubuntu:22.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libomp-dev \
    util-linux \
    bc \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Graviton4-specific optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -mcpu=neoverse-v2 -mtune=neoverse-v2 -fopenmp"
ENV CXXFLAGS="-O3 -mcpu=neoverse-v2 -mtune=neoverse-v2 -fopenmp"

# Create workspace
WORKDIR /opt/cache-bench

# Copy cache benchmark source
COPY builds/graviton4/cache-hierarchy/cache_benchmark.c .

# Compile the cache benchmark
RUN gcc $CFLAGS cache_benchmark.c -o cache_benchmark

# Create wrapper script
RUN echo '#!/bin/bash' > /usr/local/bin/run-cache-benchmark && \
    echo 'echo "Graviton4 Cache Hierarchy Benchmark"' >> /usr/local/bin/run-cache-benchmark && \
    echo 'echo "ARM Neoverse-V2 optimized with SVE2 support"' >> /usr/local/bin/run-cache-benchmark && \
    echo 'echo "Running memory hierarchy analysis..."' >> /usr/local/bin/run-cache-benchmark && \
    echo 'cd /opt/cache-bench && ./cache_benchmark' >> /usr/local/bin/run-cache-benchmark && \
    chmod +x /usr/local/bin/run-cache-benchmark

ENTRYPOINT ["/usr/local/bin/run-cache-benchmark"]