# AWS Graviton3E optimized STREAM benchmark container
FROM ubuntu:22.04

# Install build dependencies including NUMA libraries for enhanced performance
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    gfortran \
    libomp-dev \
    libnuma-dev \
    numactl \
    util-linux \
    hwloc \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Graviton3E-specific optimizations
# Graviton3E has 35% higher vector performance than standard Graviton3
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -mcpu=neoverse-v1 -mtune=neoverse-v1 -fopenmp -funroll-loops"
ENV CXXFLAGS="-O3 -mcpu=neoverse-v1 -mtune=neoverse-v1 -fopenmp -funroll-loops"

# Create workspace
WORKDIR /opt/stream

# Download and compile STREAM with Graviton3E optimizations
RUN wget https://www.cs.virginia.edu/stream/FTP/Code/stream.c && \
    gcc $CFLAGS -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
    stream.c -o stream_omp -fopenmp

# Create wrapper script that detects system configuration
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'export OMP_NUM_THREADS=${OMP_NUM_THREADS:-$(nproc)}' >> /usr/local/bin/run-stream && \
    echo 'echo "AWS Graviton3E Enhanced STREAM Benchmark"' >> /usr/local/bin/run-stream && \
    echo 'echo "Compiler: GCC with -mcpu=neoverse-v1 -mtune=neoverse-v1 (ARMv8.4-A+SVE+Enhanced)"' >> /usr/local/bin/run-stream && \
    echo 'echo "Enhanced Features: 35% higher vector performance, 200Gbps networking optimized"' >> /usr/local/bin/run-stream && \
    echo 'echo "Target Instances: c7gn (network-optimized), hpc7g (HPC-optimized)"' >> /usr/local/bin/run-stream && \
    echo 'echo "Threads: $OMP_NUM_THREADS"' >> /usr/local/bin/run-stream && \
    echo 'echo "Array size: 80M elements (1.8GB total)"' >> /usr/local/bin/run-stream && \
    echo 'echo "----------------------------------------"' >> /usr/local/bin/run-stream && \
    echo 'cd /opt/stream && ./stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]