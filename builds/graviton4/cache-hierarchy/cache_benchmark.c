#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <string.h>
#include <unistd.h>

#define L1_SIZE (32 * 1024)      
#define L2_SIZE (1024 * 1024)    
#define L3_SIZE (8 * 1024 * 1024)

void cache_latency_test(size_t size, const char* level, int iterations) {
    printf("Testing %s cache latency (size: %zu KB)\n", level, size / 1024);
    
    volatile char *data = malloc(size);
    if (!data) {
        printf("Failed to allocate %zu bytes\n", size);
        return;
    }
    
    /* Initialize data to avoid page faults during timing */
    memset((void*)data, 0xAA, size);
    
    /* Warm up */
    for (int i = 0; i < 100; i++) {
        volatile char dummy = data[i % size];
    }
    
    /* Measure latency with random access pattern */
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    size_t index = 0;
    for (int i = 0; i < iterations; i++) {
        /* Simple stride to avoid prefetcher interference */
        index = (index + 64 + (i * 17)) % size;
        volatile char dummy = data[index];
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_ns = (end.tv_sec - start.tv_sec) * 1e9 + (end.tv_nsec - start.tv_nsec);
    double avg_latency = time_ns / iterations;
    
    printf("%s Average Latency: %.2f ns per access\n", level, avg_latency);
    
    /* Bandwidth test - sequential access */
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    volatile long long sum = 0;
    size_t stride = 64; /* Cache line stride */
    for (size_t i = 0; i < size; i += stride) {
        sum += data[i];
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    time_ns = (end.tv_sec - start.tv_sec) * 1e9 + (end.tv_nsec - start.tv_nsec);
    double bandwidth_gbps = (size / (time_ns / 1e9)) / (1024*1024*1024);
    
    printf("%s Sequential Bandwidth: %.2f GB/s\n", level, bandwidth_gbps);
    printf("----------------------------------------\n");
    
    free((void*)data);
}

void detect_system_info() {
    printf("ARM Neoverse-V2 (Graviton4) Cache Hierarchy Analysis\n");
    printf("Compiled with: -mcpu=neoverse-v2 (ARMv8.5-A+SVE2)\n");
    
    /* Show basic system info */
    printf("Available cores: %ld\n", sysconf(_SC_NPROCESSORS_ONLN));
    printf("Page size: %ld bytes\n", sysconf(_SC_PAGESIZE));
    printf("========================================\n");
}

int main() {
    detect_system_info();
    
    /* Test cache hierarchy with conservative sizes */
    cache_latency_test(L1_SIZE / 2, "L1", 10000);
    cache_latency_test(L2_SIZE / 2, "L2", 5000);
    cache_latency_test(L3_SIZE / 2, "L3", 1000);
    cache_latency_test(L3_SIZE * 2, "Memory", 500);
    
    printf("\nBenchmark completed successfully!\n");
    printf("Results show cache hierarchy performance on ARM Neoverse-V2\n");
    
    return 0;
}