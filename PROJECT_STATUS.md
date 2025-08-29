# Project Status

## Container Normalization (Spack-based)

### ✅ Completed
- **AMD AOCC**: Real AMD AOCC 5.0.0 compiler via Spack on c7a.large
- **Universal**: GCC 13 + OpenBLAS via Spack (cross-platform)
- **GitHub Actions**: Disabled to prevent interference
- **AWS Binary Cache**: Configured for 20x faster builds

### 🔄 In Progress  
- **Intel oneAPI**: Building on c7i.large with correct AWS binary cache

### ⏳ Pending
- **ARM64 Graviton**: Build on c7g.large with GCC 13 + OpenBLAS
- **Validation**: Verify all containers use consistent Ubuntu 24.04 + Spack

## Container Architecture

All containers follow identical approach:
```
Ubuntu 24.04 base
↓
Spack installation  
↓
AWS binary cache (https://binaries.spack.io/develop)
↓
Architecture-specific compiler (Intel oneAPI / AMD AOCC / GCC 13)
↓ 
Build benchmarks on native hardware
```

## Registry Status

**Public ECR**: `public.ecr.aws/x5e6z7q0/`

| Container | Status | Architecture | Compiler |
|-----------|--------|--------------|----------|
| benchmarks-amd-spack | ✅ | AMD Zen4 | AOCC 5.0.0 |
| benchmarks-universal-spack | ✅ | Universal | GCC 13 |
| benchmarks-intel-spack | 🔄 | Intel x86_64 | oneAPI 2025.2.1 |
| benchmarks-graviton-spack | ⏳ | ARM64 | GCC 13 |

## Development Principles

- **No cross-compilation**: Build on native hardware only
- **No workarounds**: Fix issues properly
- **No fallbacks**: Use real vendor compilers
- **Consistent base**: Ubuntu 24.04 + Spack for all
- **Real execution**: No simulated or fake data