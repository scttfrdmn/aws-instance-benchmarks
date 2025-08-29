# AWS Instance Benchmarks

Open database of comprehensive performance benchmarks for AWS EC2 instances, designed to enable data-driven instance selection for research computing workloads.

## 🚀 Quick Start

### Run Benchmarks

```bash
# Intel instances
docker run --rm public.ecr.aws/x5e6z7q0/benchmarks-intel-spack:latest

# AMD instances  
docker run --rm public.ecr.aws/x5e6z7q0/benchmarks-amd-spack:latest

# ARM64/Graviton instances
docker run --rm public.ecr.aws/x5e6z7q0/benchmarks-graviton-spack:latest

# Universal (any architecture)
docker run --rm public.ecr.aws/x5e6z7q0/benchmarks-universal-spack:latest
```

## 📊 What's Included

### Memory Performance
- **STREAM Benchmarks**: Memory bandwidth (Copy, Scale, Add, Triad)
- **Cache Hierarchy**: L1/L2/L3 performance analysis
- **NUMA Topology**: Multi-socket performance characterization

### CPU Performance  
- **LINPACK**: Peak GFLOPS performance
- **CoreMark**: Integer performance and efficiency
- **Vectorization**: Architecture-specific SIMD optimization

### Architecture Coverage
- **Intel**: Sapphire Rapids, Ice Lake, Cascade Lake, Skylake
- **AMD**: Zen4, Zen3, Zen2 with real AMD AOCC compiler
- **ARM64**: Graviton 4, 3, 2 with NEON/SVE optimizations

## 🛠️ Container Architecture

### Spack-Based Normalization
All containers built with consistent Ubuntu 24.04 + Spack approach:

```dockerfile
FROM ubuntu:24.04

# Install Spack
RUN git clone https://github.com/spack/spack.git /opt/spack && \
    . /opt/spack/share/spack/setup-env.sh && \
    spack mirror add binary_mirror https://binaries.spack.io/develop && \
    spack buildcache keys --install --trust

# Install architecture-specific compilers
RUN . /opt/spack/share/spack/setup-env.sh && \
    spack install intel-oneapi-compilers    # Intel containers
    # OR spack install aocc+license-agreed  # AMD containers  
    # OR spack install gcc@13              # ARM64 containers
```

### Native Hardware Builds
- **Intel containers**: Built on c7i.large Intel instances
- **AMD containers**: Built on c7a.large AMD instances  
- **ARM64 containers**: Built on c7g.large Graviton instances
- **No cross-compilation**: Maximum performance accuracy

### Available Container Registry
```
public.ecr.aws/x5e6z7q0/benchmarks-intel-spack:latest
public.ecr.aws/x5e6z7q0/benchmarks-amd-spack:latest
public.ecr.aws/x5e6z7q0/benchmarks-graviton-spack:latest
public.ecr.aws/x5e6z7q0/benchmarks-universal-spack:latest
```

## 📁 Project Structure

```
builds/
├── intel/oneapi-optimized/          # Intel oneAPI + MKL
├── amd/aocc-optimized/              # AMD AOCC compiler
├── graviton/arm64-optimized/        # ARM64 GCC + OpenBLAS
└── universal/comprehensive/         # Cross-platform GCC
```

## 🔧 Container Development

### Building Containers

Containers must be built on their native architectures:

```bash
# Intel (build on c7i instance)
docker build -f builds/intel/oneapi-optimized/Dockerfile.spack .

# AMD (build on c7a instance)  
docker build -f builds/amd/aocc-optimized/Dockerfile.spack .

# ARM64 (build on c7g instance)
docker build -f builds/graviton/arm64-optimized/Dockerfile.spack .
```

### Spack Configuration

All containers use AWS binary cache for fast installation:

```bash
# Configure Spack binary cache
spack mirror add binary_mirror https://binaries.spack.io/develop
spack buildcache keys --install --trust

# Install compilers
spack install intel-oneapi-compilers  # 20x faster with cache
```

## 🤝 Contributing

1. Fork the repository
2. Build containers on native hardware
3. Test benchmark execution
4. Submit pull request

### Development Guidelines

- **Ubuntu 24.04 base**: All containers use same base OS
- **Spack normalization**: Consistent package management
- **Native builds**: No cross-compilation allowed
- **Real compilers**: Intel oneAPI, AMD AOCC, not GCC simulation
- **No fallbacks**: Fix issues properly, no workarounds

## 📄 License

MIT License - see [LICENSE](LICENSE)

Benchmark data: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)

## 🔗 Related Projects

- [ComputeCompass](https://github.com/scttfrdmn/computecompass) - Instance selector
- [Spack](https://spack.io) - Package manager