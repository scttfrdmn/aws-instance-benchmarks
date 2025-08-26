#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <immintrin.h>
#include <omp.h>
#include <string.h>

#define ARRAY_SIZE (64 * 1024 * 1024) 
#define ITERATIONS 100

void print_cpu_features() {
    printf("Intel Sapphire Rapids Vector Instruction Benchmark\n");
    printf("Compiled with: -march=sapphirerapids -mavx512vnni -mavx512bf16\n");
    
    /* Check CPU features */
    unsigned int eax, ebx, ecx, edx;
    __cpuid(1, eax, ebx, ecx, edx);
    
    printf("CPU Features Detected:\n");
    printf("SSE4.2: %s\n", (ecx & (1 << 20)) ? "Yes" : "No");
    printf("AVX: %s\n", (ecx & (1 << 28)) ? "Yes" : "No");
    
    __cpuid_count(7, 0, eax, ebx, ecx, edx);
    printf("AVX2: %s\n", (ebx & (1 << 5)) ? "Yes" : "No");
    printf("AVX512F: %s\n", (ebx & (1 << 16)) ? "Yes" : "No");
    printf("AVX512VNNI: %s\n", (ecx & (1 << 11)) ? "Yes" : "No");
    printf("AVX512BF16: %s\n", (eax & (1 << 5)) ? "Yes" : "No");
    
    printf("========================================\n");
}

double benchmark_scalar_add() {
    float *a = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    float *b = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    float *c = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    
    /* Initialize arrays */
    for (int i = 0; i < ARRAY_SIZE; i++) {
        a[i] = 1.0f + i * 0.001f;
        b[i] = 2.0f + i * 0.001f;
    }
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int iter = 0; iter < ITERATIONS; iter++) {
        for (int i = 0; i < ARRAY_SIZE; i++) {
            c[i] = a[i] + b[i];
        }
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_s = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    double gflops = (double)ARRAY_SIZE * ITERATIONS / time_s / 1e9;
    
    printf("Scalar Addition: %.2f GFLOPS\n", gflops);
    
    free(a); free(b); free(c);
    return gflops;
}

double benchmark_avx512_add() {
    float *a = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    float *b = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    float *c = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    
    /* Initialize arrays */
    for (int i = 0; i < ARRAY_SIZE; i++) {
        a[i] = 1.0f + i * 0.001f;
        b[i] = 2.0f + i * 0.001f;
    }
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int iter = 0; iter < ITERATIONS; iter++) {
        for (int i = 0; i < ARRAY_SIZE; i += 16) {
            __m512 va = _mm512_load_ps(&a[i]);
            __m512 vb = _mm512_load_ps(&b[i]);
            __m512 vc = _mm512_add_ps(va, vb);
            _mm512_store_ps(&c[i], vc);
        }
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_s = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    double gflops = (double)ARRAY_SIZE * ITERATIONS / time_s / 1e9;
    
    printf("AVX-512 Addition: %.2f GFLOPS\n", gflops);
    
    free(a); free(b); free(c);
    return gflops;
}

double benchmark_avx512_fma() {
    float *a = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    float *b = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    float *c = aligned_alloc(64, ARRAY_SIZE * sizeof(float));
    
    /* Initialize arrays */
    for (int i = 0; i < ARRAY_SIZE; i++) {
        a[i] = 1.0f + i * 0.001f;
        b[i] = 2.0f + i * 0.001f;
        c[i] = 0.5f;
    }
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int iter = 0; iter < ITERATIONS; iter++) {
        for (int i = 0; i < ARRAY_SIZE; i += 16) {
            __m512 va = _mm512_load_ps(&a[i]);
            __m512 vb = _mm512_load_ps(&b[i]);
            __m512 vc = _mm512_load_ps(&c[i]);
            vc = _mm512_fmadd_ps(va, vb, vc); /* c = a * b + c */
            _mm512_store_ps(&c[i], vc);
        }
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    double time_s = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    double gflops = (double)ARRAY_SIZE * ITERATIONS * 2 / time_s / 1e9; /* FMA is 2 ops */
    
    printf("AVX-512 FMA: %.2f GFLOPS\n", gflops);
    
    free(a); free(b); free(c);
    return gflops;
}

int main() {
    print_cpu_features();
    
    printf("Vector Instruction Performance Comparison:\n");
    printf("Array Size: %d elements (%.1f MB)\n", ARRAY_SIZE, ARRAY_SIZE * sizeof(float) / 1024.0 / 1024.0);
    printf("Iterations: %d\n", ITERATIONS);
    printf("----------------------------------------\n");
    
    double scalar_perf = benchmark_scalar_add();
    double avx512_add_perf = benchmark_avx512_add();
    double avx512_fma_perf = benchmark_avx512_fma();
    
    printf("----------------------------------------\n");
    printf("Performance Summary:\n");
    printf("AVX-512 Speedup (Add): %.1fx over scalar\n", avx512_add_perf / scalar_perf);
    printf("AVX-512 FMA Peak: %.2f GFLOPS\n", avx512_fma_perf);
    printf("Theoretical Peak (@ 3.0GHz, 16 FMA/cycle): %.0f GFLOPS\n", 
           3.0 * 16.0 * 2.0 * omp_get_max_threads());
    
    return 0;
}