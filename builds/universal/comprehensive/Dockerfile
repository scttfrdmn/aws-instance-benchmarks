# Enhanced Universal Benchmark Container for Cloud Compass Integration
# Provides comprehensive performance benchmarks with runtime architecture optimization
FROM ubuntu:24.04

LABEL maintainer="AWS Instance Benchmarks <benchmarks@computecompass.dev>"
LABEL version="2.0.0"
LABEL description="Comprehensive benchmark suite with STREAM, LINPACK, and CoreMark"

# Install comprehensive build environment
RUN apt-get update && apt-get install -y \
    gcc-13 \
    g++-13 \
    gfortran-13 \
    libblas-dev \
    liblapack-dev \
    libomp-dev \
    build-essential \
    wget \
    curl \
    unzip \
    bc \
    util-linux \
    jq \
    && update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-13 100 \
    && update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-13 100 \
    && update-alternatives --install /usr/bin/gfortran gfortran /usr/bin/gfortran-13 100 \
    && rm -rf /var/lib/apt/lists/*

# Set environment variables
ENV CC=gcc
ENV CXX=g++
ENV FC=gfortran

# Create benchmark workspace
WORKDIR /opt/benchmarks

# Copy architecture detection and optimization script
COPY builds/universal/comprehensive/scripts/detect-and-optimize.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/detect-and-optimize.sh

# Copy benchmark sources
COPY builds/universal/comprehensive/benchmarks/ ./

# Copy execution scripts
COPY builds/universal/comprehensive/scripts/build-benchmarks.sh ./
COPY builds/universal/comprehensive/scripts/version-collector.sh ./scripts/
COPY builds/universal/comprehensive/scripts/run-benchmark.sh /usr/local/bin/
RUN chmod +x ./build-benchmarks.sh ./scripts/version-collector.sh /usr/local/bin/run-benchmark.sh

# Build all benchmarks with generic optimization during container build
ENV CFLAGS="-O3 -fopenmp"
ENV CXXFLAGS="-O3 -fopenmp"
ENV FCFLAGS="-O3 -fopenmp"
RUN bash ./build-benchmarks.sh

# Set default command
ENTRYPOINT ["/usr/local/bin/run-benchmark.sh"]
CMD ["all"]