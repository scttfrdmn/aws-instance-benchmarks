# Intel Sapphire Rapids Cache Hierarchy benchmark container
FROM ubuntu:22.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libomp-dev \
    libnuma-dev \
    numactl \
    util-linux \
    wget \
    bc \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Sapphire Rapids optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
ENV CXXFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"

# Create workspace
WORKDIR /opt/cache-bench

# Create comprehensive cache hierarchy benchmark
RUN cat > cache_benchmark.c << 'EOF'
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <string.h>
#include <unistd.h>
#include <omp.h>

// Cache size detection from /proc/cpuinfo or reasonable defaults for Sapphire Rapids
#define L1_SIZE (48 * 1024)      // 48KB L1D typical for Sapphire Rapids
#define L2_SIZE (2 * 1024 * 1024) // 2MB L2 typical 
#define L3_SIZE (112 * 1024 * 1024) // 112MB L3 for Sapphire Rapids

void cache_latency_test(size_t size, const char* level, int iterations) {
    printf("Testing %s cache latency (size: %zu KB)\n", level, size / 1024);
    
    volatile char *data = malloc(size);
    if (!data) {
        printf("Failed to allocate %zu bytes\n", size);
        return;
    }
    
    // Initialize data to avoid page faults during timing
    memset((void*)data, 0xAA, size);
    
    // Warm up
    for (int i = 0; i < 1000; i++) {
        volatile char dummy = data[i % size];
    }
    
    // Measure latency with random access pattern
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    size_t index = 0;
    for (int i = 0; i < iterations; i++) {
        // Random stride to defeat prefetcher
        index = (index + 64 + (i * 17)) % size;
        volatile char dummy = data[index];
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_ns = (end.tv_sec - start.tv_sec) * 1e9 + (end.tv_nsec - start.tv_nsec);
    double avg_latency = time_ns / iterations;
    
    printf("%s Average Latency: %.2f ns per access\n", level, avg_latency);
    printf("%s Bandwidth Test:\n", level);
    
    // Bandwidth test - sequential access
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    volatile long long sum = 0;
    for (size_t i = 0; i < size; i += 64) { // Cache line stride
        sum += data[i];
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    time_ns = (end.tv_sec - start.tv_sec) * 1e9 + (end.tv_nsec - start.tv_nsec);
    double bandwidth_gbps = (size / (time_ns / 1e9)) / (1024*1024*1024);
    
    printf("%s Sequential Bandwidth: %.2f GB/s\n", level, bandwidth_gbps);
    printf("----------------------------------------\n");
    
    free((void*)data);
}

void detect_cache_sizes() {
    printf("Intel Sapphire Rapids Cache Hierarchy Analysis\n");
    printf("Detecting actual cache sizes from system...\n");
    
    system("lscpu | grep -i cache");
    printf("----------------------------------------\n");
}

int main() {
    printf("Intel Sapphire Rapids Cache Hierarchy Benchmark\n");
    printf("AVX-512 VNNI/BF16 Optimized Build\n");
    printf("Cores available: %d\n", omp_get_max_threads());
    printf("========================================\n");
    
    detect_cache_sizes();
    
    // Test L1 cache - use 50% of estimated size
    cache_latency_test(L1_SIZE / 2, "L1", 100000);
    
    // Test L2 cache - use 50% of estimated size  
    cache_latency_test(L2_SIZE / 2, "L2", 10000);
    
    // Test L3 cache - use 50% of estimated size
    cache_latency_test(L3_SIZE / 2, "L3", 1000);
    
    // Test main memory - use size larger than L3
    cache_latency_test(L3_SIZE * 4, "Memory", 100);
    
    printf("\nNUMA Topology Information:\n");
    system("numactl --hardware");
    
    return 0;
}
EOF

# Compile the cache benchmark
RUN gcc $CFLAGS cache_benchmark.c -o cache_benchmark -fopenmp

# Create wrapper script
RUN echo '#!/bin/bash' > /usr/local/bin/run-cache-benchmark && \
    echo 'echo "Intel Sapphire Rapids Cache Hierarchy Benchmark"' >> /usr/local/bin/run-cache-benchmark && \
    echo 'echo "Target: AVX-512 VNNI/BF16 optimized for AI/ML workloads"' >> /usr/local/bin/run-cache-benchmark && \
    echo 'echo "CPU Info:"' >> /usr/local/bin/run-cache-benchmark && \
    echo 'lscpu | head -20' >> /usr/local/bin/run-cache-benchmark && \
    echo 'echo "========================================"' >> /usr/local/bin/run-cache-benchmark && \
    echo 'cd /opt/cache-bench && ./cache_benchmark' >> /usr/local/bin/run-cache-benchmark && \
    chmod +x /usr/local/bin/run-cache-benchmark

ENTRYPOINT ["/usr/local/bin/run-cache-benchmark"]