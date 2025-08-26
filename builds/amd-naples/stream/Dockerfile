# AMD EPYC Naples (1st Gen) optimized STREAM benchmark container
FROM ubuntu:22.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    gfortran \
    libomp-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with AMD EPYC Naples-specific optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=znver1 -mtune=znver1 -mavx2 -mfma -fopenmp"
ENV CXXFLAGS="-O3 -march=znver1 -mtune=znver1 -mavx2 -mfma -fopenmp"

# Create workspace
WORKDIR /opt/stream

# Download and compile STREAM with AMD EPYC Naples optimizations
RUN wget https://www.cs.virginia.edu/stream/FTP/Code/stream.c && \
    gcc $CFLAGS -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
    stream.c -o stream_omp -fopenmp

# Create wrapper script that detects system configuration
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'export OMP_NUM_THREADS=${OMP_NUM_THREADS:-$(nproc)}' >> /usr/local/bin/run-stream && \
    echo 'echo "AMD EPYC Naples (1st Gen) Optimized STREAM Benchmark"' >> /usr/local/bin/run-stream && \
    echo 'echo "Compiler: GCC with -march=znver1 -mavx2 -mfma"' >> /usr/local/bin/run-stream && \
    echo 'echo "Threads: $OMP_NUM_THREADS"' >> /usr/local/bin/run-stream && \
    echo 'echo "Array size: 80M elements (1.8GB total)"' >> /usr/local/bin/run-stream && \
    echo 'echo "----------------------------------------"' >> /usr/local/bin/run-stream && \
    echo 'cd /opt/stream && ./stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]