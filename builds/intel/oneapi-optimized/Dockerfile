# Phase 2: Intel oneAPI Optimized Container
# Maximum performance for Intel Xeon processors with oneAPI toolkit and MKL
FROM intel/oneapi-hpckit:latest

LABEL maintainer="AWS Instance Benchmarks <benchmarks@computecompass.dev>"
LABEL version="2.1.0"
LABEL description="Intel oneAPI optimized benchmark suite with MKL acceleration"
LABEL architecture="Intel Xeon (Ice Lake, Sapphire Rapids, Emerald Rapids)"

# Install additional dependencies
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    unzip \
    bc \
    util-linux \
    jq \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set Intel oneAPI environment
ENV ONEAPI_ROOT="/opt/intel/oneapi"
ENV INTEL_ROOT="${ONEAPI_ROOT}"
ENV MKLROOT="${ONEAPI_ROOT}/mkl/latest"
ENV COMPILER_ROOT="${ONEAPI_ROOT}/compiler/latest"

# Initialize oneAPI environment
RUN echo "source ${ONEAPI_ROOT}/setvars.sh --force" >> ~/.bashrc

# Set Intel-optimized compiler environment
ENV CC="icx"
ENV CXX="icpx" 
ENV FC="ifx"

# Intel-specific optimization flags
ENV CFLAGS="-O3 -xHost -ipo -qopenmp -march=native -mtune=native"
ENV CXXFLAGS="-O3 -xHost -ipo -qopenmp -march=native -mtune=native"
ENV FCFLAGS="-O3 -xHost -ipo -qopenmp -march=native -mtune=native"

# Intel MKL linking flags for maximum performance
ENV LDFLAGS="-L${MKLROOT}/lib/intel64 -lmkl_intel_lp64 -lmkl_intel_thread -lmkl_core -liomp5 -lpthread -lm -ldl"

# Set threading for optimal performance
ENV OMP_NUM_THREADS="auto"
ENV MKL_NUM_THREADS="auto" 
ENV MKL_DYNAMIC="FALSE"
ENV KMP_AFFINITY="granularity=fine,compact,1,0"

# Create benchmark workspace
WORKDIR /opt/benchmarks

# Copy Intel-optimized detection script
COPY builds/intel/oneapi-optimized/scripts/intel-optimize.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/intel-optimize.sh

# Copy benchmark sources (reuse universal sources)
COPY builds/universal/comprehensive/benchmarks/ ./

# Copy Intel-specific build script
COPY builds/intel/oneapi-optimized/scripts/build-benchmarks-intel.sh ./
COPY builds/intel/oneapi-optimized/scripts/version-collector-intel.sh ./scripts/
COPY builds/intel/oneapi-optimized/scripts/run-benchmark-intel.sh /usr/local/bin/

# Set executable permissions
RUN chmod +x ./build-benchmarks-intel.sh ./scripts/version-collector-intel.sh /usr/local/bin/run-benchmark-intel.sh

# Build benchmarks with Intel oneAPI optimizations
RUN bash -c "source ${ONEAPI_ROOT}/setvars.sh --force && ./build-benchmarks-intel.sh"

# Validate installation
RUN bash -c "source ${ONEAPI_ROOT}/setvars.sh --force && ls -la stream/ linpack/ coremark/ && echo 'Intel oneAPI build complete'"

# Container metadata
ENV CONTAINER_VERSION="2.1.0"
ENV CONTAINER_VARIANT="intel-oneapi-optimized"
ENV OPTIMIZATION_PROFILE="Intel oneAPI + MKL maximum performance"

# Set default command
CMD ["run-benchmark-intel.sh"]