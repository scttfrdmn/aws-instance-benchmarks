/*
 * Simple LINPACK benchmark for measuring CPU FLOPS performance
 * Based on original LINPACK routines for solving linear systems
 * Adapted for AWS Instance Benchmarking with Cloud Compass integration
 */

#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <time.h>
#include <sys/time.h>
#include <omp.h>

#ifndef N
#define N 2000  /* Default matrix size */
#endif

/* Function prototypes */
double mysecond(void);
double dgefa(double **a, int n, int *ipvt);
void dgesl(double **a, int n, int *ipvt, double *b, int job);
double **allocate_matrix(int n);
void free_matrix(double **a, int n);
void matgen(double **a, int n, double *b);

/* Timer function */
double mysecond(void)
{
    struct timeval tp;
    struct timezone tzp;
    gettimeofday(&tp, &tzp);
    return ((double) tp.tv_sec + (double) tp.tv_usec * 1.e-6);
}

/* Allocate matrix memory */
double **allocate_matrix(int n)
{
    int i;
    double **a = (double **) malloc(n * sizeof(double *));
    if (!a) return NULL;
    
    for (i = 0; i < n; i++) {
        a[i] = (double *) malloc(n * sizeof(double));
        if (!a[i]) {
            for (int j = 0; j < i; j++) free(a[j]);
            free(a);
            return NULL;
        }
    }
    return a;
}

/* Free matrix memory */
void free_matrix(double **a, int n)
{
    int i;
    for (i = 0; i < n; i++) {
        free(a[i]);
    }
    free(a);
}

/* Generate test matrix and RHS */
void matgen(double **a, int n, double *b)
{
    int i, j;
    
    /* Initialize matrix A and vector b */
    for (i = 0; i < n; i++) {
        for (j = 0; j < n; j++) {
            a[i][j] = (double)(i - j) / (double)n + 1.0;
        }
        b[i] = (double)(i + 1);
    }
}

/* Simple LU decomposition with partial pivoting */
double dgefa(double **a, int n, int *ipvt)
{
    int i, j, k, l;
    double t;
    
    /* Initialize pivot vector */
    for (i = 0; i < n; i++) {
        ipvt[i] = i;
    }
    
    /* Gaussian elimination with partial pivoting */
    for (k = 0; k < n - 1; k++) {
        /* Find pivot */
        l = k;
        for (i = k + 1; i < n; i++) {
            if (fabs(a[i][k]) > fabs(a[l][k])) {
                l = i;
            }
        }
        ipvt[k] = l;
        
        /* Interchange rows if needed */
        if (l != k) {
            for (j = 0; j < n; j++) {
                t = a[l][j];
                a[l][j] = a[k][j];
                a[k][j] = t;
            }
        }
        
        /* Skip if pivot is zero */
        if (a[k][k] == 0.0) continue;
        
        /* Compute multipliers */
        for (i = k + 1; i < n; i++) {
            a[i][k] = -a[i][k] / a[k][k];
        }
        
        /* Apply transformations */
        for (j = k + 1; j < n; j++) {
            t = a[l][j];
            if (l != k) {
                a[l][j] = a[k][j];
                a[k][j] = t;
            }
            for (i = k + 1; i < n; i++) {
                a[i][j] += a[i][k] * t;
            }
        }
    }
    
    return 0.0;
}

/* Solve linear system using LU decomposition */
void dgesl(double **a, int n, int *ipvt, double *b, int job)
{
    int k, l;
    double t;
    
    /* Forward solve L*y = b */
    for (k = 0; k < n - 1; k++) {
        l = ipvt[k];
        t = b[l];
        if (l != k) {
            b[l] = b[k];
            b[k] = t;
        }
        for (int i = k + 1; i < n; i++) {
            b[i] += a[i][k] * t;
        }
    }
    
    /* Back solve U*x = y */
    for (k = n - 1; k >= 0; k--) {
        b[k] = b[k] / a[k][k];
        t = -b[k];
        for (int i = 0; i < k; i++) {
            b[i] += a[i][k] * t;
        }
    }
}

int main(void)
{
    double **a;
    double *b, *x;
    int *ipvt;
    double t1, t2, total_time;
    double ops, mflops;
    int n = N;
    int i;
    
    printf("=====================================\n");
    printf("LINPACK Benchmark\n");
    printf("=====================================\n");
    printf("Matrix size: %d x %d\n", n, n);
    printf("Precision: Double\n");
    
#ifdef _OPENMP
    printf("OpenMP threads: %d\n", omp_get_max_threads());
#endif
    
    /* Allocate memory */
    a = allocate_matrix(n);
    b = (double *) malloc(n * sizeof(double));
    x = (double *) malloc(n * sizeof(double));
    ipvt = (int *) malloc(n * sizeof(int));
    
    if (!a || !b || !x || !ipvt) {
        printf("Error: Cannot allocate memory for matrices\n");
        return 1;
    }
    
    /* Generate test matrix */
    printf("Generating matrix...\n");
    matgen(a, n, b);
    
    /* Copy b to x for solution */
    for (i = 0; i < n; i++) {
        x[i] = b[i];
    }
    
    printf("Starting LINPACK benchmark...\n");
    printf("=====================================\n");
    
    /* Start timing */
    t1 = mysecond();
    
    /* Factor the matrix */
    dgefa(a, n, ipvt);
    
    /* Solve the system */
    dgesl(a, n, ipvt, x, 0);
    
    /* Stop timing */
    t2 = mysecond();
    total_time = t2 - t1;
    
    /* Calculate performance metrics */
    ops = (2.0 * n * n * n) / 3.0 + 2.0 * n * n; /* LU + solve operations */
    mflops = (ops / 1.0e6) / total_time;
    
    /* Display results */
    printf("=====================================\n");
    printf("LINPACK Results\n");
    printf("=====================================\n");
    printf("Matrix size:           %d\n", n);
    printf("Leading dimension:     %d\n", n);
    printf("Time (seconds):        %.6f\n", total_time);
    printf("GFLOPS:               %.3f\n", mflops / 1000.0);
    printf("MFLOPS:               %.3f\n", mflops);
    
    /* Simple residual check */
    matgen(a, n, b);  /* Regenerate original matrix */
    double residual = 0.0;
    for (i = 0; i < n; i++) {
        double sum = 0.0;
        for (int j = 0; j < n; j++) {
            sum += a[i][j] * x[j];
        }
        double diff = sum - b[i];
        residual += diff * diff;
    }
    residual = sqrt(residual);
    
    printf("Residual:             %.6e\n", residual);
    
    if (residual < 1.0e-10) {
        printf("Solution: PASSED\n");
    } else {
        printf("Solution: FAILED (residual too large)\n");
    }
    
    printf("=====================================\n");
    
    /* Memory usage information */
    double memory_mb = (double)(n * n * sizeof(double)) / (1024.0 * 1024.0);
    printf("Memory used:          %.1f MB\n", memory_mb);
    printf("Memory per GFLOPS:    %.1f MB/GFLOPS\n", memory_mb / (mflops / 1000.0));
    
    /* Clean up */
    free_matrix(a, n);
    free(b);
    free(x);
    free(ipvt);
    
    return 0;
}