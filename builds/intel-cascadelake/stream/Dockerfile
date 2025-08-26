# Intel Cascade Lake optimized STREAM benchmark container
FROM ubuntu:22.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    gfortran \
    libomp-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Cascade Lake-specific optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=cascadelake -mtune=cascadelake -mavx512f -mavx512dq -mavx512cd -mavx512bw -mavx512vl -fopenmp"
ENV CXXFLAGS="-O3 -march=cascadelake -mtune=cascadelake -mavx512f -mavx512dq -mavx512cd -mavx512bw -mavx512vl -fopenmp"

# Create workspace
WORKDIR /opt/stream

# Download and compile STREAM with Intel Cascade Lake optimizations
RUN wget https://www.cs.virginia.edu/stream/FTP/Code/stream.c && \
    gcc $CFLAGS -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
    stream.c -o stream_omp -fopenmp

# Create wrapper script that detects system configuration
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'export OMP_NUM_THREADS=${OMP_NUM_THREADS:-$(nproc)}' >> /usr/local/bin/run-stream && \
    echo 'echo "Intel Cascade Lake Optimized STREAM Benchmark"' >> /usr/local/bin/run-stream && \
    echo 'echo "Compiler: GCC with -march=cascadelake -mavx512f"' >> /usr/local/bin/run-stream && \
    echo 'echo "Threads: $OMP_NUM_THREADS"' >> /usr/local/bin/run-stream && \
    echo 'echo "Array size: 80M elements (1.8GB total)"' >> /usr/local/bin/run-stream && \
    echo 'echo "----------------------------------------"' >> /usr/local/bin/run-stream && \
    echo 'cd /opt/stream && ./stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]