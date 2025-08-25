# Architecture-Optimized STREAM Benchmark
# Target: AWS Graviton4 processors (ARM Neoverse-V2)
FROM --platform=linux/arm64/v8 ubuntu:22.04

# Install build dependencies and tools
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libnuma-dev \
    numactl \
    lscpu \
    hwloc \
    && rm -rf /var/lib/apt/lists/*

# Set up optimized compilation environment for Graviton4
ENV CFLAGS="-O3 -march=native -mtune=native -mcpu=neoverse-v2 -flto -ffast-math"
ENV CXXFLAGS="-O3 -march=native -mtune=native -mcpu=neoverse-v2 -flto -ffast-math"

# Create working directory
WORKDIR /opt/benchmark

# Copy STREAM benchmark source
COPY stream.c /opt/benchmark/
COPY benchmark_runner.sh /opt/benchmark/

# Compile architecture-optimized STREAM benchmark
RUN gcc ${CFLAGS} -DSTREAM_ARRAY_SIZE=100000000 -DNTIMES=10 \
    -fopenmp -mcpu=neoverse-v2 -mtune=neoverse-v2 \
    -o stream stream.c

# Make scripts executable
RUN chmod +x benchmark_runner.sh

# Run benchmark with NUMA optimizations
ENTRYPOINT ["./benchmark_runner.sh"]