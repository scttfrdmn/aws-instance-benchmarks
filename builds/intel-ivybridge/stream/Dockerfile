# Intel Ivy Bridge optimized STREAM benchmark container (AWS 3rd Generation)
FROM ubuntu:22.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    gfortran \
    libomp-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Ivy Bridge-specific optimizations
# Ivy Bridge introduced AVX, 22nm process, improved Turbo Boost
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=ivybridge -mtune=ivybridge -msse4.2 -mavx -fopenmp"
ENV CXXFLAGS="-O3 -march=ivybridge -mtune=ivybridge -msse4.2 -mavx -fopenmp"

# Create workspace
WORKDIR /opt/stream

# Download and compile STREAM with Intel Ivy Bridge optimizations
RUN wget https://www.cs.virginia.edu/stream/FTP/Code/stream.c && \
    gcc $CFLAGS -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
    stream.c -o stream_omp -fopenmp

# Create wrapper script that detects system configuration
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'export OMP_NUM_THREADS=${OMP_NUM_THREADS:-$(nproc)}' >> /usr/local/bin/run-stream && \
    echo 'echo "Intel Ivy Bridge Optimized STREAM Benchmark (AWS 3rd Gen)"' >> /usr/local/bin/run-stream && \
    echo 'echo "Processor: Intel Xeon E5-2670 v2 (Ivy Bridge, 22nm, 2013)"' >> /usr/local/bin/run-stream && \
    echo 'echo "Compiler: GCC with -march=ivybridge -mavx -msse4.2"' >> /usr/local/bin/run-stream && \
    echo 'echo "Target Instances: c3, m3, r3 (Historical AWS 3rd Generation)"' >> /usr/local/bin/run-stream && \
    echo 'echo "Threads: $OMP_NUM_THREADS"' >> /usr/local/bin/run-stream && \
    echo 'echo "Array size: 80M elements (1.8GB total)"' >> /usr/local/bin/run-stream && \
    echo 'echo "Historical Note: First AWS generation with AVX support"' >> /usr/local/bin/run-stream && \
    echo 'echo "----------------------------------------"' >> /usr/local/bin/run-stream && \
    echo 'cd /opt/stream && ./stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]