# Universal STREAM benchmark container
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gfortran \
    libomp-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Download and compile STREAM
WORKDIR /opt/stream
RUN wget https://www.cs.virginia.edu/stream/FTP/Code/stream.c && \
    gcc -O3 -fopenmp -mcpu=native -DSTREAM_ARRAY_SIZE=80000000 -DNTIMES=10 \
    stream.c -o stream_omp

# Create benchmark script
RUN echo '#!/bin/bash' > /usr/local/bin/run-stream && \
    echo 'export OMP_NUM_THREADS=${OMP_NUM_THREADS:-$(nproc)}' >> /usr/local/bin/run-stream && \
    echo 'cd /opt/stream && ./stream_omp' >> /usr/local/bin/run-stream && \
    chmod +x /usr/local/bin/run-stream

ENTRYPOINT ["/usr/local/bin/run-stream"]