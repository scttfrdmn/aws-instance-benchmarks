# Phase 2: AMD AOCC Optimized Universal Container  
# Provides maximum AMD performance using AOCC compilers with AOCL mathematical libraries
FROM ubuntu:24.04

LABEL maintainer="AWS Instance Benchmarks <benchmarks@computecompass.dev>"
LABEL version="2.1.0-amd"
LABEL description="AMD AOCC optimized benchmark suite with STREAM, LINPACK, and CoreMark"
LABEL optimization="AMD AOCC compilers with AOCL optimization for Zen architectures"

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    wget \
    curl \
    unzip \
    bc \
    util-linux \
    jq \
    python3 \
    python3-pip \
    libnuma1 \
    libnuma-dev \
    && rm -rf /var/lib/apt/lists/*

# Download and install AMD AOCC compiler (community edition)
# Note: In production, this would use proper AMD licensing
RUN mkdir -p /opt/amd && cd /opt/amd && \
    wget -q https://developer.amd.com/wordpress/media/files/aocc-compiler-4.0.0.tar.xz && \
    tar -xf aocc-compiler-4.0.0.tar.xz && \
    cd aocc-compiler-4.0.0 && \
    ./install.sh -i /opt/amd/aocc

# Install AMD Optimizing CPU Libraries (AOCL)
RUN cd /opt/amd && \
    wget -q https://developer.amd.com/wordpress/media/files/aocl-linux-aocc-4.0.0.tar.gz && \
    tar -xzf aocl-linux-aocc-4.0.0.tar.gz && \
    cd aocl-linux-aocc-4.0.0 && \
    ./install.sh -i /opt/amd/aocl

# Set AMD compiler environment
ENV AMD_ROOT=/opt/amd
ENV AOCC_ROOT=/opt/amd/aocc
ENV AOCL_ROOT=/opt/amd/aocl
ENV CC=clang
ENV CXX=clang++
ENV FC=flang

# Create benchmark workspace
WORKDIR /opt/benchmarks

# Copy architecture detection script (modified for AMD compilers)
COPY builds/universal/amd-optimized/scripts/detect-and-optimize-amd.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/detect-and-optimize-amd.sh

# Copy benchmark sources
COPY builds/universal/comprehensive/benchmarks/ ./

# Copy AMD-optimized build scripts
COPY builds/universal/amd-optimized/scripts/build-benchmarks-amd.sh ./
COPY builds/universal/comprehensive/scripts/run-benchmark.sh /usr/local/bin/
RUN chmod +x ./build-benchmarks-amd.sh /usr/local/bin/run-benchmark.sh

# Build benchmarks with AMD AOCC optimization
RUN /bin/bash -c "source $AOCC_ROOT/setenv_AOCC.sh && source $AOCL_ROOT/aocl-setup.sh && ./build-benchmarks-amd.sh"

# Set default command
ENTRYPOINT ["/usr/local/bin/run-benchmark.sh"] 
CMD ["all"]