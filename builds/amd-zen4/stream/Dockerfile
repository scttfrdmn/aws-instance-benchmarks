# AMD Zen4 optimized STREAM benchmark container
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gfortran \
    python3 \
    python3-pip \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Spack
RUN git clone -c feature.manyFiles=true https://github.com/spack/spack.git /opt/spack
ENV SPACK_ROOT=/opt/spack
ENV PATH=$SPACK_ROOT/bin:$PATH

# Copy Spack configuration
COPY spack-configs/amd-zen4.yaml /opt/spack/etc/spack/packages.yaml

# Install STREAM with AMD AOCC optimization
RUN spack install stream@5.10 %gcc@11 target=zen4 +openmp
RUN spack load stream

# Create benchmark script
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'spack load stream' >> /usr/local/bin/run-stream && \
    echo 'stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]
