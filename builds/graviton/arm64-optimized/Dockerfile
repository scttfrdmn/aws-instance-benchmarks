# AWS Graviton ARM64 Optimized Benchmark Container
# Built for native AWS Graviton processors (Graviton 2/3)

FROM ubuntu:24.04

# Set container metadata
LABEL maintainer="AWS Instance Benchmarks"
LABEL version="2.1.0"
LABEL variant="graviton-arm64-optimized"
LABEL description="ARM64 Graviton-optimized benchmarks for AWS c7g/m7g/r7g instances"
LABEL architecture="arm64"

# Set non-interactive mode for apt
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Container version and variant info
ENV CONTAINER_VERSION="2.1.0"
ENV CONTAINER_VARIANT="graviton-arm64-optimized"

# Install build dependencies and optimization tools
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc-13 \
    g++-13 \
    gfortran-13 \
    libc6-dev \
    libgomp1 \
    libomp-dev \
    cmake \
    make \
    git \
    wget \
    curl \
    bc \
    jq \
    procps \
    util-linux \
    numactl \
    && rm -rf /var/lib/apt/lists/*

# Set ARM64 Graviton-optimized compiler environment
ENV CC=gcc-13
ENV CXX=g++-13  
ENV FC=gfortran-13

# Create benchmarks directory
RUN mkdir -p /opt/benchmarks
WORKDIR /opt/benchmarks

# Copy Graviton optimization scripts
COPY scripts/graviton-optimize.sh /usr/local/bin/
COPY scripts/build-benchmarks-graviton.sh /usr/local/bin/
COPY scripts/run-benchmark-graviton.sh /usr/local/bin/
COPY scripts/version-collector-graviton.sh /usr/local/bin/

# Make scripts executable
RUN chmod +x /usr/local/bin/*.sh

# Copy benchmark source code
COPY benchmarks/ /opt/benchmarks/

# Build benchmarks with Graviton ARM64 optimizations
RUN bash -c "source /usr/local/bin/graviton-optimize.sh && /usr/local/bin/build-benchmarks-graviton.sh"

# Set runtime environment for optimal Graviton performance
ENV OMP_NUM_THREADS=""
ENV GOMP_CPU_AFFINITY=""

# Health check to validate benchmarks
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD test -f /opt/benchmarks/stream/stream_benchmark_graviton && \
      test -f /opt/benchmarks/linpack/linpack_benchmark_graviton && \
      test -f /opt/benchmarks/coremark/coremark_benchmark_graviton || exit 1

# Default command runs all Graviton-optimized benchmarks
CMD ["/usr/local/bin/run-benchmark-graviton.sh", "all"]