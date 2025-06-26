# Multi-stage build for graviton3 architecture
FROM ubuntu:22.04 as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    python3 \
    python3-pip \
    cmake \
    && rm -rf /var/lib/apt/lists/*

# Install Spack
RUN git clone -c feature.manyFiles=true https://github.com/spack/spack.git /opt/spack
ENV SPACK_ROOT=/opt/spack
ENV PATH=$SPACK_ROOT/bin:$PATH

# Architecture-specific compiler setup

# GCC with architecture-specific flags
RUN apt-get update && apt-get install -y gcc-11 g++-11 && rm -rf /var/lib/apt/lists/*


# Copy Spack configuration
COPY spack-configs/graviton3.yaml /opt/spack/etc/spack/packages.yaml

# Build benchmarks with architecture-specific optimizations

RUN spack install stream %gcc@11 target=graviton3 cflags="-O3 -march=armv8.2-a+sve -mcpu=neoverse-v1"


# Runtime stage
FROM ubuntu:22.04 as runtime

# Copy built benchmarks
COPY --from=builder /opt/spack /opt/spack

# Set environment
ENV SPACK_ROOT=/opt/spack
ENV PATH=$SPACK_ROOT/bin:$PATH

# Create benchmark runner script
RUN echo '#!/bin/bash' > /usr/local/bin/run-benchmark && \
    echo 'spack load stream' >> /usr/local/bin/run-benchmark && \
    echo 'exec "$@"' >> /usr/local/bin/run-benchmark && \
    chmod +x /usr/local/bin/run-benchmark

ENTRYPOINT ["/usr/local/bin/run-benchmark"]
