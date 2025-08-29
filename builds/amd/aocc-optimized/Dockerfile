# Phase 2: AMD AOCC Optimized Container  
# Maximum performance for AMD EPYC processors with AOCC and AOCL
FROM ubuntu:24.04

LABEL maintainer="AWS Instance Benchmarks <benchmarks@computecompass.dev>"
LABEL version="2.1.0"
LABEL description="AMD AOCC optimized benchmark suite with AOCL acceleration"
LABEL architecture="AMD EPYC (Zen 3, Zen 4, Zen 5)"

# Install system dependencies
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    unzip \
    bc \
    util-linux \
    jq \
    git \
    build-essential \
    gfortran \
    libblas-dev \
    liblapack-dev \
    libomp-dev \
    && rm -rf /var/lib/apt/lists/*

# Download and install AMD AOCC 5.1 compiler
RUN cd /tmp && \
    wget --no-check-certificate -q "https://download.amd.com/developer/eula/aocc-compiler/aocc-compiler-5.1.0_1_amd64.deb" -O aocc-compiler-5.1.0_1_amd64.deb && \
    dpkg -i aocc-compiler-5.1.0_1_amd64.deb || apt-get install -f -y && \
    rm -f aocc-compiler-5.1.0_1_amd64.deb

# Download and install AMD AOCL 5.1 libraries  
RUN cd /tmp && \
    wget --no-check-certificate -q "https://download.amd.com/developer/eula/aocl/aocl-linux-aocc-5.1.0_1_amd64.deb" -O aocl-linux-aocc-5.1.0_1_amd64.deb && \
    dpkg -i aocl-linux-aocc-5.1.0_1_amd64.deb || apt-get install -f -y && \
    rm -f aocl-linux-aocc-5.1.0_1_amd64.deb

# Set AMD optimization environment variables for AOCC 5.1
ENV AMD_AOCC_ROOT="/opt/AMD/aocc-compiler-5.1.0"
ENV AMD_AOCL_ROOT="/opt/AMD/aocl-5.1.0"
ENV PATH="$AMD_AOCC_ROOT/bin:$PATH"
ENV LD_LIBRARY_PATH="$AMD_AOCC_ROOT/lib:$AMD_AOCL_ROOT/lib:$LD_LIBRARY_PATH"

# AMD AOCC compiler settings
ENV CC="clang"
ENV CXX="clang++" 
ENV FC="flang"

# AMD EPYC-specific optimization flags for AOCC
ENV CFLAGS="-O3 -march=znver4 -mtune=znver4 -fopenmp -ffast-math -funroll-loops"
ENV CXXFLAGS="-O3 -march=znver4 -mtune=znver4 -fopenmp -ffast-math -funroll-loops"
ENV FCFLAGS="-O3 -march=znver4 -mtune=znver4 -fopenmp -ffast-math -funroll-loops"

# AMD AOCL-optimized linking with real AOCL libraries
ENV LDFLAGS="-L$AMD_AOCL_ROOT/lib -laocl_blas -laocl_lapack -lm -fopenmp"

# AMD threading optimization
ENV OMP_NUM_THREADS="auto"
ENV OMP_PROC_BIND="close"
ENV OMP_PLACES="cores"

# Create benchmark workspace
WORKDIR /opt/benchmarks

# Copy AMD-optimized detection script
COPY scripts/amd-optimize.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/amd-optimize.sh

# Copy benchmark sources
COPY benchmarks/ ./

# Copy AMD-specific build scripts
COPY scripts/build-benchmarks-amd.sh ./
COPY scripts/version-collector-amd.sh ./scripts/
COPY scripts/run-benchmark-amd.sh /usr/local/bin/

# Set executable permissions
RUN chmod +x ./build-benchmarks-amd.sh ./scripts/version-collector-amd.sh /usr/local/bin/run-benchmark-amd.sh

# Build benchmarks with AMD optimizations
RUN ./build-benchmarks-amd.sh

# Validate AMD-optimized builds
RUN ls -la stream/ linpack/ coremark/ && echo 'AMD AOCC build complete'

# Container metadata
ENV CONTAINER_VERSION="2.1.0"
ENV CONTAINER_VARIANT="amd-aocc-optimized"
ENV OPTIMIZATION_PROFILE="AMD AOCC + AOCL maximum performance for AMD EPYC"

# Set default command
CMD ["run-benchmark-amd.sh"]