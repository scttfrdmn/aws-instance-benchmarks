# Intel Sapphire Rapids NUMA Topology benchmark container
FROM ubuntu:22.04

# Install build dependencies including NUMA libraries
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libnuma-dev \
    numactl \
    libomp-dev \
    util-linux \
    hwloc \
    bc \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Sapphire Rapids optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
ENV CXXFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"

# Create workspace
WORKDIR /opt/numa-bench

# Create comprehensive NUMA benchmark
RUN cat > numa_benchmark.c << 'EOF'
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <numa.h>
#include <omp.h>
#include <string.h>
#include <unistd.h>

#define TEST_SIZE (256 * 1024 * 1024) // 256MB test size
#define ITERATIONS 10

void print_numa_topology() {
    printf("Intel Sapphire Rapids NUMA Topology Analysis\n");
    printf("==============================================\n");
    
    if (numa_available() == -1) {
        printf("NUMA not available on this system\n");
        return;
    }
    
    int max_node = numa_max_node();
    printf("NUMA nodes available: %d\n", max_node + 1);
    
    for (int node = 0; node <= max_node; node++) {
        if (numa_node_size(node, NULL) > 0) {
            long long size = numa_node_size(node, NULL);
            printf("Node %d: %.2f GB memory\n", node, size / (1024.0 * 1024.0 * 1024.0));
            
            // Print CPU list for this node
            struct bitmask *cpus = numa_allocate_cpumask();
            if (numa_node_to_cpus(node, cpus) == 0) {
                printf("Node %d CPUs: ", node);
                for (int cpu = 0; cpu < numa_num_possible_cpus(); cpu++) {
                    if (numa_bitmask_isbitset(cpus, cpu)) {
                        printf("%d ", cpu);
                    }
                }
                printf("\n");
            }
            numa_free_cpumask(cpus);
        }
    }
    
    printf("==============================================\n");
}

double benchmark_local_memory(int node) {
    // Bind to specific NUMA node
    struct bitmask *old_mask = numa_get_membind();
    numa_set_preferred(node);
    
    char *data = numa_alloc_onnode(TEST_SIZE, node);
    if (!data) {
        printf("Failed to allocate memory on node %d\n", node);
        return 0.0;
    }
    
    // Initialize data
    memset(data, 0xAA, TEST_SIZE);
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    volatile long long sum = 0;
    for (int iter = 0; iter < ITERATIONS; iter++) {
        for (size_t i = 0; i < TEST_SIZE; i += 64) { // Cache line stride
            sum += data[i];
        }
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_s = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    double bandwidth = (TEST_SIZE * ITERATIONS) / time_s / (1024.0 * 1024.0 * 1024.0);
    
    numa_free(data, TEST_SIZE);
    numa_set_membind(old_mask);
    numa_free_cpumask(old_mask);
    
    return bandwidth;
}

double benchmark_remote_memory(int cpu_node, int mem_node) {
    if (cpu_node == mem_node) return 0.0;
    
    // Bind thread to CPU node
    cpu_set_t cpu_set;
    CPU_ZERO(&cpu_set);
    
    // Find first CPU in the cpu_node
    struct bitmask *cpus = numa_allocate_cpumask();
    numa_node_to_cpus(cpu_node, cpus);
    int target_cpu = -1;
    for (int cpu = 0; cpu < numa_num_possible_cpus(); cpu++) {
        if (numa_bitmask_isbitset(cpus, cpu)) {
            target_cpu = cpu;
            break;
        }
    }
    numa_free_cpumask(cpus);
    
    if (target_cpu == -1) return 0.0;
    
    CPU_SET(target_cpu, &cpu_set);
    sched_setaffinity(0, sizeof(cpu_set), &cpu_set);
    
    // Allocate memory on remote node
    char *data = numa_alloc_onnode(TEST_SIZE, mem_node);
    if (!data) {
        printf("Failed to allocate memory on node %d\n", mem_node);
        return 0.0;
    }
    
    // Initialize data
    memset(data, 0xAA, TEST_SIZE);
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    volatile long long sum = 0;
    for (int iter = 0; iter < ITERATIONS; iter++) {
        for (size_t i = 0; i < TEST_SIZE; i += 64) { // Cache line stride
            sum += data[i];
        }
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_s = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    double bandwidth = (TEST_SIZE * ITERATIONS) / time_s / (1024.0 * 1024.0 * 1024.0);
    
    numa_free(data, TEST_SIZE);
    
    return bandwidth;
}

int main() {
    print_numa_topology();
    
    if (numa_available() == -1) {
        printf("NUMA not available, running basic memory test\n");
        return 1;
    }
    
    int max_node = numa_max_node();
    
    printf("\nNUMA Memory Bandwidth Analysis:\n");
    printf("Test size: %d MB per iteration\n", TEST_SIZE / (1024 * 1024));
    printf("Iterations: %d\n", ITERATIONS);
    printf("----------------------------------------\n");
    
    // Test local memory bandwidth for each node
    for (int node = 0; node <= max_node; node++) {
        if (numa_node_size(node, NULL) > 0) {
            double local_bw = benchmark_local_memory(node);
            printf("Node %d Local Memory: %.2f GB/s\n", node, local_bw);
        }
    }
    
    printf("----------------------------------------\n");
    
    // Test remote memory access patterns
    if (max_node > 0) {
        printf("Remote Memory Access Analysis:\n");
        for (int cpu_node = 0; cpu_node <= max_node; cpu_node++) {
            for (int mem_node = 0; mem_node <= max_node; mem_node++) {
                if (cpu_node != mem_node && numa_node_size(cpu_node, NULL) > 0 && numa_node_size(mem_node, NULL) > 0) {
                    double remote_bw = benchmark_remote_memory(cpu_node, mem_node);
                    if (remote_bw > 0) {
                        printf("CPU Node %d -> Memory Node %d: %.2f GB/s\n", cpu_node, mem_node, remote_bw);
                    }
                }
            }
        }
    }
    
    printf("\nHardware Topology Summary:\n");
    system("lstopo-no-graphics --of console");
    
    return 0;
}
EOF

# Compile the NUMA benchmark
RUN gcc $CFLAGS numa_benchmark.c -o numa_benchmark -lnuma -fopenmp

# Create wrapper script
RUN echo '#!/bin/bash' > /usr/local/bin/run-numa-benchmark && \
    echo 'echo "Intel Sapphire Rapids NUMA Topology Benchmark"' >> /usr/local/bin/run-numa-benchmark && \
    echo 'echo "Analyzing memory controller and interconnect performance"' >> /usr/local/bin/run-numa-benchmark && \
    echo 'echo "Hardware Info:"' >> /usr/local/bin/run-numa-benchmark && \
    echo 'numactl --hardware' >> /usr/local/bin/run-numa-benchmark && \
    echo 'echo"' >> /usr/local/bin/run-numa-benchmark && \
    echo 'cd /opt/numa-bench && ./numa_benchmark' >> /usr/local/bin/run-numa-benchmark && \
    chmod +x /usr/local/bin/run-numa-benchmark

ENTRYPOINT ["/usr/local/bin/run-numa-benchmark"]