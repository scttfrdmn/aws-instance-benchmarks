# Phase 2: Intel oneAPI Optimized Universal Container
# Provides maximum Intel performance using oneAPI compilers with Math Kernel Library
FROM ubuntu:24.04

LABEL maintainer="AWS Instance Benchmarks <benchmarks@computecompass.dev>"
LABEL version="2.1.0-intel"
LABEL description="Intel oneAPI optimized benchmark suite with STREAM, LINPACK, and CoreMark"
LABEL optimization="Intel oneAPI compilers with MKL and AVX-512 optimization"

# Install Intel oneAPI repository and dependencies
RUN apt-get update && apt-get install -y \
    wget \
    gnupg2 \
    lsb-release \
    && wget -O- https://apt.repos.intel.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB | gpg --dearmor | tee /usr/share/keyrings/oneapi-archive-keyring.gpg > /dev/null \
    && echo "deb [signed-by=/usr/share/keyrings/oneapi-archive-keyring.gpg] https://apt.repos.intel.com/oneapi all main" | tee /etc/apt/sources.list.d/oneAPI.list

# Install Intel oneAPI toolkit components
RUN apt-get update && apt-get install -y \
    intel-oneapi-compiler-dpcpp-cpp \
    intel-oneapi-compiler-fortran \
    intel-oneapi-mkl-devel \
    intel-oneapi-tbb-devel \
    build-essential \
    wget \
    curl \
    unzip \
    bc \
    util-linux \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Set Intel compiler environment
ENV ONEAPI_ROOT=/opt/intel/oneapi
ENV CC=icc
ENV CXX=icpc
ENV FC=ifort
ENV MKLROOT=/opt/intel/oneapi/mkl/latest

# Create benchmark workspace
WORKDIR /opt/benchmarks

# Copy architecture detection script (modified for Intel compilers)
COPY builds/universal/intel-optimized/scripts/detect-and-optimize-intel.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/detect-and-optimize-intel.sh

# Copy benchmark sources
COPY builds/universal/comprehensive/benchmarks/ ./

# Copy Intel-optimized build scripts
COPY builds/universal/intel-optimized/scripts/build-benchmarks-intel.sh ./
COPY builds/universal/comprehensive/scripts/run-benchmark.sh /usr/local/bin/
RUN chmod +x ./build-benchmarks-intel.sh /usr/local/bin/run-benchmark.sh

# Build benchmarks with Intel oneAPI optimization
RUN /bin/bash -c "source $ONEAPI_ROOT/setvars.sh && ./build-benchmarks-intel.sh"

# Set default command
ENTRYPOINT ["/usr/local/bin/run-benchmark.sh"]
CMD ["all"]