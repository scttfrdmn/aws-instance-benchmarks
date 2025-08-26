# Intel Sapphire Rapids LINPACK (HPL) benchmark container
FROM ubuntu:22.04

# Install build dependencies including math libraries
RUN apt-get update && apt-get install -y \
    build-essential \
    gfortran \
    libblas-dev \
    liblapack-dev \
    libopenmpi-dev \
    openmpi-bin \
    libomp-dev \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Set up build environment with Intel Sapphire Rapids optimizations
ENV CC=gcc
ENV CXX=g++
ENV FC=gfortran
ENV CFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
ENV CXXFLAGS="-O3 -march=sapphirerapids -mavx512vnni -mavx512bf16 -fopenmp"
ENV FCFLAGS="-O3 -march=sapphire-rapids -mavx512vnni -mavx512bf16 -fopenmp"

# Create workspace
WORKDIR /opt/hpl

# Download and build LINPACK (HPL)
RUN wget http://www.netlib.org/benchmark/hpl/hpl-2.3.tar.gz && \
    tar -xzf hpl-2.3.tar.gz && \
    cd hpl-2.3 && \
    cp setup/Make.Linux_PII_CBLAS Make.arch && \
    sed -i 's/LAdir.*/LAdir = \/usr\/lib\/x86_64-linux-gnu/' Make.arch && \
    sed -i 's/LAinc.*/LAinc =/' Make.arch && \
    sed -i 's/LAlib.*/LAlib = -lblas -llapack/' Make.arch && \
    sed -i 's/CC.*/CC = gcc/' Make.arch && \
    sed -i 's/CCFLAGS.*/CCFLAGS = $(CFLAGS)/' Make.arch && \
    sed -i 's/LINKER.*/LINKER = gcc/' Make.arch && \
    sed -i 's/LINKFLAGS.*/LINKFLAGS = $(CFLAGS)/' Make.arch && \
    make arch=arch

# Create comprehensive benchmark runner script
RUN echo '#!/bin/bash' > /usr/local/bin/run-linpack && \
    echo 'export OMP_NUM_THREADS=${OMP_NUM_THREADS:-$(nproc)}' >> /usr/local/bin/run-linpack && \
    echo 'TOTAL_MEM_KB=$(grep MemTotal /proc/meminfo | awk "{print \$2}")' >> /usr/local/bin/run-linpack && \
    echo 'TOTAL_MEM_GB=$((TOTAL_MEM_KB / 1024 / 1024))' >> /usr/local/bin/run-linpack && \
    echo 'CORES=$(nproc)' >> /usr/local/bin/run-linpack && \
    echo '# Use 50% of memory for matrix size calculation' >> /usr/local/bin/run-linpack && \
    echo 'MEM_FOR_MATRIX=$((TOTAL_MEM_KB * 50 / 100))' >> /usr/local/bin/run-linpack && \
    echo '# N^2 * 8 bytes = memory usage, solve for N' >> /usr/local/bin/run-linpack && \
    echo 'N=$(echo "sqrt($MEM_FOR_MATRIX * 1024 / 8)" | bc -l | cut -d. -f1)' >> /usr/local/bin/run-linpack && \
    echo '# Ensure N is reasonable bounds' >> /usr/local/bin/run-linpack && \
    echo 'N=$((N < 500 ? 500 : N))' >> /usr/local/bin/run-linpack && \
    echo 'N=$((N > 10000 ? 10000 : N))' >> /usr/local/bin/run-linpack && \
    echo 'NB=256  # Block size optimized for AVX-512' >> /usr/local/bin/run-linpack && \
    echo 'echo "Intel Sapphire Rapids LINPACK Benchmark"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Processor: Intel Xeon Sapphire Rapids with AVX-512 VNNI/BF16"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Cores: $CORES"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Total Memory: ${TOTAL_MEM_GB}GB"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Matrix Size (N): $N"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Block Size (NB): $NB"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Expected GFLOPS: $(echo "scale=2; 2 * $N * $N * $N / 3 / 1000000000" | bc -l)"' >> /usr/local/bin/run-linpack && \
    echo 'echo "Memory Usage: $(echo "scale=2; $N * $N * 8 / 1024 / 1024 / 1024" | bc -l)GB"' >> /usr/local/bin/run-linpack && \
    echo 'echo "----------------------------------------"' >> /usr/local/bin/run-linpack && \
    echo 'cd /opt/hpl/hpl-2.3/bin/arch' >> /usr/local/bin/run-linpack && \
    echo 'cat > HPL.dat << EOF' >> /usr/local/bin/run-linpack && \
    echo 'HPLinpack benchmark input file' >> /usr/local/bin/run-linpack && \
    echo 'Innovative Computing Laboratory, University of Tennessee' >> /usr/local/bin/run-linpack && \
    echo 'HPL.out      output file name (if any)' >> /usr/local/bin/run-linpack && \
    echo '6            device out (6=stdout,7=stderr,file)' >> /usr/local/bin/run-linpack && \
    echo '1            # of problems sizes (N)' >> /usr/local/bin/run-linpack && \
    echo '$N           Ns' >> /usr/local/bin/run-linpack && \
    echo '1            # of NBs' >> /usr/local/bin/run-linpack && \
    echo '$NB          NBs' >> /usr/local/bin/run-linpack && \
    echo '0            PMAP process mapping (0=Row-,1=Column-major)' >> /usr/local/bin/run-linpack && \
    echo '1            # of process grids (P x Q)' >> /usr/local/bin/run-linpack && \
    echo '1            Ps' >> /usr/local/bin/run-linpack && \
    echo '1            Qs' >> /usr/local/bin/run-linpack && \
    echo '16.0         threshold' >> /usr/local/bin/run-linpack && \
    echo '1            # of panel fact' >> /usr/local/bin/run-linpack && \
    echo '2            PFACTs (0=left, 1=Crout, 2=Right)' >> /usr/local/bin/run-linpack && \
    echo '1            # of recursive stopping criterium' >> /usr/local/bin/run-linpack && \
    echo '4            NBMINs (>= 1)' >> /usr/local/bin/run-linpack && \
    echo '1            # of panels in recursion' >> /usr/local/bin/run-linpack && \
    echo '2            NDIVs' >> /usr/local/bin/run-linpack && \
    echo '1            # of recursive panel fact.' >> /usr/local/bin/run-linpack && \
    echo '1            RFACTs (0=left, 1=Crout, 2=Right)' >> /usr/local/bin/run-linpack && \
    echo '1            # of broadcast' >> /usr/local/bin/run-linpack && \
    echo '1            BCASTs (0=1rg,1=1rM,2=2rg,3=2rM,4=Lng,5=LnM)' >> /usr/local/bin/run-linpack && \
    echo '1            # of lookahead depth' >> /usr/local/bin/run-linpack && \
    echo '1            DEPTHs (>=0)' >> /usr/local/bin/run-linpack && \
    echo '2            SWAP (0=bin-exch,1=long,2=mix)' >> /usr/local/bin/run-linpack && \
    echo '64           swapping threshold' >> /usr/local/bin/run-linpack && \
    echo '0            L1 in (0=transposed,1=no-transposed) form' >> /usr/local/bin/run-linpack && \
    echo '0            U  in (0=transposed,1=no-transposed) form' >> /usr/local/bin/run-linpack && \
    echo '1            Equilibration (0=no,1=yes)' >> /usr/local/bin/run-linpack && \
    echo '8            memory alignment in double (> 0)' >> /usr/local/bin/run-linpack && \
    echo 'EOF' >> /usr/local/bin/run-linpack && \
    echo './xhpl' >> /usr/local/bin/run-linpack && \
    chmod +x /usr/local/bin/run-linpack

# Install bc for calculations
RUN apt-get update && apt-get install -y bc && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/local/bin/run-linpack"]