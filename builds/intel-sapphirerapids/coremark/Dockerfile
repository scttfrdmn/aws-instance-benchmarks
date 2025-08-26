# Intel Sapphire Rapids CoreMark benchmark container
FROM ubuntu:22.04

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    git \
    make \
    libomp-dev \
    util-linux \
    bc \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Sapphire Rapids optimizations
ENV CC=gcc
ENV CXX=g++
ENV CFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
ENV CXXFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"

# Create workspace
WORKDIR /opt/coremark

# Download and build CoreMark
RUN git clone https://github.com/eembc/coremark.git . && \
    make XCFLAGS="$CFLAGS" compile

# Create system-aware CoreMark runner
RUN cat > run_coremark.c << 'EOF'
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/sysinfo.h>

void get_cpu_info() {
    printf("=== Intel Sapphire Rapids CoreMark Benchmark ===\n");
    printf("Compiler optimizations: AVX-512 VNNI/BF16\n");
    
    FILE *fp = fopen("/proc/cpuinfo", "r");
    if (fp) {
        char line[256];
        while (fgets(line, sizeof(line), fp)) {
            if (strstr(line, "model name")) {
                printf("CPU: %s", strchr(line, ':') + 2);
                break;
            }
        }
        fclose(fp);
    }
    
    int cores = get_nprocs();
    printf("Available cores: %d\n", cores);
    
    // Get CPU frequency from /proc/cpuinfo
    fp = fopen("/proc/cpuinfo", "r");
    if (fp) {
        char line[256];
        while (fgets(line, sizeof(line), fp)) {
            if (strstr(line, "cpu MHz")) {
                float freq = atof(strchr(line, ':') + 2);
                printf("CPU frequency: %.0f MHz\n", freq);
                break;
            }
        }
        fclose(fp);
    }
}

int calculate_iterations() {
    int cores = get_nprocs();
    
    // Base iterations: 1M per core, adjusted for frequency
    // Sapphire Rapids typically runs at 2.0-3.2 GHz
    int base_iterations = 1000000;
    int total_iterations = base_iterations * cores;
    
    // Ensure minimum runtime of ~10 seconds
    if (total_iterations < 5000000) total_iterations = 5000000;
    if (total_iterations > 100000000) total_iterations = 100000000;
    
    printf("Calculated iterations: %d\n", total_iterations);
    return total_iterations;
}

int main() {
    get_cpu_info();
    
    int iterations = calculate_iterations();
    printf("============================================\n");
    
    // Run CoreMark with calculated iterations
    char command[256];
    snprintf(command, sizeof(command), "./coremark.exe 0x0 0x0 0x66 %d 7 1 2000", iterations);
    
    printf("Running CoreMark with %d iterations...\n\n", iterations);
    return system(command);
}
EOF

# Compile the runner
RUN gcc -o coremark_runner run_coremark.c

# Create wrapper script
RUN echo '#!/bin/bash' > /usr/local/bin/run-coremark && \
    echo 'cd /opt/coremark && ./coremark_runner' >> /usr/local/bin/run-coremark && \
    chmod +x /usr/local/bin/run-coremark

ENTRYPOINT ["/usr/local/bin/run-coremark"]