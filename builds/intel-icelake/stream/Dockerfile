# Multi-stage build for intel-icelake architecture
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

# Intel OneAPI setup
RUN curl -fsSL https://apt.repos.intel.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB | apt-key add - && \
    echo "deb https://apt.repos.intel.com/oneapi all main" > /etc/apt/sources.list.d/oneAPI.list && \
    apt-get update && apt-get install -y intel-oneapi-compiler-dpcpp-cpp && \
    rm -rf /var/lib/apt/lists/*


# Copy Spack configuration
COPY spack-configs/intel-icelake.yaml /opt/spack/etc/spack/packages.yaml

# Build benchmarks with architecture-specific optimizations

RUN spack install stream %gcc@11 target=intel-icelake cflags="-O3 -xCORE-AVX512 -qopt-zmm-usage=high"


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
