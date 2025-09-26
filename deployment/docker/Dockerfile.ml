# Reproducible Machine Learning Benchmark Container
# Provides consistent environment for ML computing benchmarks

FROM nvidia/cuda:12.2-devel-ubuntu22.04

LABEL maintainer="AWS Instance Benchmarks Project"
LABEL description="Machine Learning benchmarks for AWS instance performance analysis"
LABEL version="2.0"

# Avoid interactive prompts during installation
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Install system dependencies
RUN apt-get update && apt-get install -y \
    # Build essentials
    build-essential \
    cmake \
    # Scientific computing libraries
    libblas-dev \
    liblapack-dev \
    libfftw3-dev \
    # Parallel computing
    libopenmpi-dev \
    openmpi-bin \
    # Development tools
    git \
    wget \
    curl \
    unzip \
    # Python
    python3 \
    python3-pip \
    python3-dev \
    # System monitoring
    htop \
    iotop \
    sysstat \
    numactl \
    hwloc-nox \
    # Graphics and display (for headless ML)
    libgl1-mesa-glx \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    libgomp1 \
    # Cleanup
    && rm -rf /var/lib/apt/lists/*

# Set up working directory
WORKDIR /opt/ml-benchmarks

# Install Python packages for ML
RUN pip3 install --no-cache-dir \
    # Core scientific computing
    numpy \
    scipy \
    pandas \
    matplotlib \
    seaborn \
    scikit-learn \
    # Deep learning frameworks
    torch \
    torchvision \
    torchaudio \
    tensorflow \
    keras \
    # Transformers and NLP
    transformers \
    tokenizers \
    datasets \
    # Computer vision
    opencv-python \
    pillow \
    # Reinforcement learning
    gym \
    stable-baselines3 \
    # Distributed computing
    ray \
    # Performance monitoring
    psutil \
    py-cpuinfo \
    gputil \
    nvidia-ml-py3 \
    # AWS integration
    boto3 \
    # Additional ML utilities
    wandb \
    tensorboard \
    mlflow \
    optuna \
    hyperopt

# Install JAX for high-performance ML
RUN pip3 install --no-cache-dir jax jaxlib

# Install additional specialized ML packages
RUN pip3 install --no-cache-dir \
    # Graph neural networks
    torch-geometric \
    # Federated learning
    flwr \
    # Time series
    prophet \
    sktime \
    # AutoML
    auto-sklearn \
    # Model optimization
    onnx \
    onnxruntime \
    # Explainable AI
    shap \
    lime

# Copy benchmark scripts
COPY ml-benchmarks/ ./benchmarks/
COPY shared-scripts/ ./shared/

# Download pre-trained models and datasets for consistent benchmarking
RUN mkdir -p /data/models /data/datasets \
    && cd /data \
    # Download standard ML datasets
    && python3 -c "
import torchvision.datasets as datasets
import torchvision.transforms as transforms
from sklearn.datasets import fetch_openml, fetch_20newsgroups, load_digits
import pandas as pd
import numpy as np
import os

print('Downloading standard ML datasets for benchmarking...')

# CIFAR-10 for computer vision benchmarks
transform = transforms.Compose([transforms.ToTensor()])
datasets.CIFAR10(root='./datasets', train=True, download=True, transform=transform)
datasets.CIFAR10(root='./datasets', train=False, download=True, transform=transform)

# MNIST for basic ML benchmarks
datasets.MNIST(root='./datasets', train=True, download=True, transform=transform)
datasets.MNIST(root='./datasets', train=False, download=True, transform=transform)

# Fashion-MNIST
datasets.FashionMNIST(root='./datasets', train=True, download=True, transform=transform)

# Generate synthetic datasets for scalability testing
def generate_synthetic_data():
    sizes = [10000, 100000, 1000000]
    features = [100, 1000, 5000]
    
    for n_samples in sizes:
        for n_features in features:
            print(f'Generating synthetic dataset: {n_samples} samples, {n_features} features')
            X = np.random.randn(n_samples, n_features).astype(np.float32)
            y = np.random.randint(0, 10, n_samples)
            
            np.savez_compressed(f'./datasets/synthetic_{n_samples}_{n_features}.npz',
                              X=X, y=y)

generate_synthetic_data()
print('Dataset preparation complete')
"

# Create benchmark configuration
RUN echo '# ML Benchmark Configuration
export CUDA_VISIBLE_DEVICES=0
export PYTHONPATH=/opt/ml-benchmarks:$PYTHONPATH
export TRANSFORMERS_CACHE=/data/models/transformers
export HF_HOME=/data/models/huggingface
export TORCH_HOME=/data/models/torch

# Performance settings
export OMP_NUM_THREADS=1
export MKL_NUM_THREADS=1
export OPENBLAS_NUM_THREADS=1
export NUMBA_NUM_THREADS=1

# Memory settings
export PYTORCH_CUDA_ALLOC_CONF=max_split_size_mb:128
' > /etc/environment

# Set execution permissions
RUN chmod +x benchmarks/*.sh benchmarks/*.py shared/*.sh

# Create benchmark execution script
RUN echo '#!/bin/bash
set -euo pipefail

BENCHMARK_TYPE=${1:-"all"}
RESULTS_DIR=${RESULTS_DIR:-"/tmp/ml-benchmark-results"}
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

mkdir -p "$RESULTS_DIR"

echo "=== ML Benchmark Execution Started ==="
echo "Benchmark Type: $BENCHMARK_TYPE"
echo "Results Directory: $RESULTS_DIR"
echo "Timestamp: $TIMESTAMP"

# System information
echo "=== System Information ==="
python3 -c "
import torch
import tensorflow as tf
import psutil
import py_cpuinfo
import subprocess

print(f\"CPU Info: {py_cpuinfo.get_cpu_info()[\"brand_raw\"]}\")
print(f\"CPU Cores: {psutil.cpu_count()}\")
print(f\"Memory: {psutil.virtual_memory().total / (1024**3):.1f} GB\")
print(f\"PyTorch Version: {torch.__version__}\")
print(f\"CUDA Available: {torch.cuda.is_available()}\")
if torch.cuda.is_available():
    print(f\"GPU: {torch.cuda.get_device_name(0)}\")
    print(f\"GPU Memory: {torch.cuda.get_device_properties(0).total_memory / (1024**3):.1f} GB\")
print(f\"TensorFlow Version: {tf.__version__}\")
"

case "$BENCHMARK_TYPE" in
    "training")
        echo "Running ML training benchmarks..."
        python3 benchmarks/training_benchmarks.py --results-dir "$RESULTS_DIR"
        ;;
    "inference")
        echo "Running ML inference benchmarks..."
        python3 benchmarks/inference_benchmarks.py --results-dir "$RESULTS_DIR"
        ;;
    "transformers")
        echo "Running transformer model benchmarks..."
        python3 benchmarks/transformer_benchmarks.py --results-dir "$RESULTS_DIR"
        ;;
    "distributed")
        echo "Running distributed ML benchmarks..."
        python3 benchmarks/distributed_benchmarks.py --results-dir "$RESULTS_DIR"
        ;;
    "all")
        echo "Running all ML benchmarks..."
        python3 benchmarks/training_benchmarks.py --results-dir "$RESULTS_DIR"
        python3 benchmarks/inference_benchmarks.py --results-dir "$RESULTS_DIR"
        python3 benchmarks/transformer_benchmarks.py --results-dir "$RESULTS_DIR"
        ;;
    *)
        echo "Unknown benchmark type: $BENCHMARK_TYPE"
        echo "Available types: training, inference, transformers, distributed, all"
        exit 1
        ;;
esac

echo "=== ML Benchmark Execution Completed ==="
' > /usr/local/bin/run-ml-benchmarks.sh

RUN chmod +x /usr/local/bin/run-ml-benchmarks.sh

# Default command
CMD ["/usr/local/bin/run-ml-benchmarks.sh", "all"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python3 -c "import torch, tensorflow as tf; print('ML frameworks ready')" || exit 1