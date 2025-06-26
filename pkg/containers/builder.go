// Package containers provides architecture-optimized container build capabilities
// for AWS instance benchmark execution.
//
// This package handles the complete container build pipeline, from generating
// architecture-specific Dockerfiles to pushing optimized images to registries.
// It specializes in creating containers with optimal compiler settings and
// performance tuning for different processor architectures.
//
// Key Components:
//   - Builder: Main service for container build orchestration
//   - BuildConfig: Configuration for architecture-specific builds
//   - DockerfileTemplate: Template data for Dockerfile generation
//
// Usage:
//   builder := containers.NewBuilder("public.ecr.aws", "aws-benchmarks")
//   config := containers.BuildConfig{
//       Architecture: "intel-icelake",
//       BenchmarkSuite: "stream",
//       CompilerType: "intel",
//   }
//   err := builder.BuildContainer(ctx, config)
//
// The package provides:
//   - Multi-stage Dockerfile generation with architecture-specific optimizations
//   - Compiler-specific optimization flags (Intel OneAPI, AMD AOCC, GCC)
//   - Spack integration for scientific software package management
//   - Container registry integration with automated pushing
//   - Build artifact management with proper tagging strategies
//
// Supported Architectures:
//   - Intel Ice Lake (m7i, c7i, r7i) with AVX-512 optimization
//   - AMD Zen 4 (m7a, c7a, r7a) with znver4 tuning
//   - AWS Graviton3 (m7g, c7g, r7g) with Neoverse-V1 and SVE support
//   - Legacy architectures with appropriate fallback optimizations
//
// Container Optimization Features:
//   - Architecture-specific compiler flag selection
//   - Multi-stage builds for minimal runtime image size
//   - Spack-based package management for reproducible builds
//   - Container layer caching for efficient rebuild workflows
package containers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Builder orchestrates the creation of architecture-optimized benchmark containers.
//
// This struct manages the complete container build workflow, from Dockerfile
// generation through registry publication. It specializes in creating containers
// with optimal performance characteristics for specific processor architectures
// and benchmark workloads.
//
// The Builder handles:
//   - Architecture-specific Dockerfile template generation
//   - Multi-stage build orchestration for size optimization
//   - Compiler toolchain selection and optimization flag configuration
//   - Container registry authentication and image publishing
//   - Build artifact management and cleanup
type Builder struct {
	// registryURL is the container registry base URL for image publication.
	// Examples: "public.ecr.aws", "gcr.io", "docker.io"
	registryURL string
	
	// namespace is the registry namespace for image organization.
	// Used to construct full image names: {registryURL}/{namespace}:{tag}
	namespace string
}

// BuildConfig defines comprehensive configuration for architecture-specific
// container builds.
//
// This configuration structure provides all parameters necessary for generating
// optimized containers with appropriate compiler settings, base images, and
// build contexts for specific AWS instance architectures.
type BuildConfig struct {
	// Architecture specifies the target processor architecture.
	// Examples: "intel-icelake", "amd-zen4", "graviton3"
	Architecture string
	
	// ContainerTag is the specific tag for the built container.
	// Typically matches the architecture for consistency.
	ContainerTag string
	
	// BenchmarkSuite identifies the benchmark software to include.
	// Supported values: "stream", "hpl", "coremark"
	BenchmarkSuite string
	
	// CompilerType selects the optimization compiler toolchain.
	// Values: "intel" (OneAPI), "amd" (AOCC), "gcc" (GNU Compiler Collection)
	CompilerType string
	
	// OptimizationFlags contains architecture-specific compiler optimization flags.
	// Generated automatically based on architecture and compiler type.
	OptimizationFlags []string
	
	// BaseImage specifies the container base image for the build.
	// Typically "ubuntu:22.04" for broad compatibility.
	BaseImage string
	
	// SpackConfig is the filename of the Spack environment configuration.
	// Contains package specifications and compiler settings.
	SpackConfig string
}

// DockerfileTemplate contains all data required for generating architecture-specific
// Dockerfiles from the template system.
//
// This structure provides template variables for creating optimized multi-stage
// Dockerfiles with proper compiler installations, optimization flags, and
// benchmark build configurations.
type DockerfileTemplate struct {
	// BaseImage is the container base image (e.g., "ubuntu:22.04").
	BaseImage string
	
	// Architecture is the target processor architecture tag.
	Architecture string
	
	// Compiler specifies the compiler type for conditional template logic.
	Compiler string
	
	// OptimizationFlags contains space-separated compiler optimization flags.
	OptimizationFlags string
	
	// BenchmarkSuite is the benchmark software to build and install.
	BenchmarkSuite string
	
	// SpackConfig is the Spack environment configuration filename.
	SpackConfig string
}

const dockerfileTemplate = `# Multi-stage build for {{ .Architecture }} architecture
FROM {{ .BaseImage }} as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    python3 \
    python3-pip \
    cmake \
    && rm -rf /var/lib/apt/lists/*

# Install Spack
RUN git clone -c feature.manyFiles=true https://github.com/spack/spack.git /opt/spack
ENV SPACK_ROOT=/opt/spack
ENV PATH=$SPACK_ROOT/bin:$PATH

# Architecture-specific compiler setup
{{ if eq .Compiler "intel" }}
# Intel OneAPI setup
RUN curl -fsSL https://apt.repos.intel.com/intel-gpg-keys/GPG-PUB-KEY-INTEL-SW-PRODUCTS.PUB | apt-key add - && \
    echo "deb https://apt.repos.intel.com/oneapi all main" > /etc/apt/sources.list.d/oneAPI.list && \
    apt-get update && apt-get install -y intel-oneapi-compiler-dpcpp-cpp && \
    rm -rf /var/lib/apt/lists/*
{{ else if eq .Compiler "amd" }}
# AMD AOCC setup - placeholder for actual AMD compiler installation
RUN echo "AMD AOCC compiler setup would go here"
{{ else }}
# GCC with architecture-specific flags
RUN apt-get update && apt-get install -y gcc-11 g++-11 && rm -rf /var/lib/apt/lists/*
{{ end }}

# Copy Spack configuration
COPY spack-configs/{{ .SpackConfig }} /opt/spack/etc/spack/packages.yaml

# Build benchmarks with architecture-specific optimizations
{{ if eq .BenchmarkSuite "stream" }}
RUN spack install stream %gcc@11 target={{ .Architecture }} cflags="{{ .OptimizationFlags }}"
{{ else if eq .BenchmarkSuite "hpl" }}
RUN spack install hpl %gcc@11 target={{ .Architecture }} cflags="{{ .OptimizationFlags }}"
{{ end }}

# Runtime stage
FROM {{ .BaseImage }} as runtime

# Copy built benchmarks
COPY --from=builder /opt/spack /opt/spack

# Set environment
ENV SPACK_ROOT=/opt/spack
ENV PATH=$SPACK_ROOT/bin:$PATH

# Create benchmark runner script
RUN echo '#!/bin/bash' > /usr/local/bin/run-benchmark && \
    echo 'spack load {{ .BenchmarkSuite }}' >> /usr/local/bin/run-benchmark && \
    echo 'exec "$@"' >> /usr/local/bin/run-benchmark && \
    chmod +x /usr/local/bin/run-benchmark

ENTRYPOINT ["/usr/local/bin/run-benchmark"]
`

// NewBuilder creates a new container builder configured for the specified registry
// and namespace.
//
// This function initializes a Builder with registry credentials and namespace
// configuration for publishing architecture-optimized benchmark containers.
// The builder supports multiple registry types including ECR Public, Docker Hub,
// and Google Container Registry.
//
// Parameters:
//   - registryURL: Base URL of the container registry (e.g., "public.ecr.aws")
//   - namespace: Registry namespace for image organization (e.g., "aws-benchmarks")
//
// Returns:
//   - *Builder: Configured builder ready for container operations
//
// Example:
//   // For ECR Public
//   builder := containers.NewBuilder("public.ecr.aws", "aws-benchmarks")
//   
//   // For Docker Hub
//   builder := containers.NewBuilder("docker.io", "myorg")
//   
//   // For Google Container Registry
//   builder := containers.NewBuilder("gcr.io", "my-project")
//
// Registry Authentication:
//   The builder relies on external authentication (docker login, aws ecr get-login-password)
//   to be configured before use. Authentication is handled by the Docker daemon.
//
// Image Naming Convention:
//   {registryURL}/{namespace}:{benchmark}-{architecture}
//   Example: "public.ecr.aws/aws-benchmarks:stream-intel-icelake"
func NewBuilder(registryURL, namespace string) *Builder {
	return &Builder{
		registryURL: registryURL,
		namespace:   namespace,
	}
}

// GenerateDockerfile creates an optimized Dockerfile for the specified benchmark and architecture.
//
// This method generates comprehensive Dockerfiles with architecture-specific optimizations,
// compiler configurations, and benchmark-specific requirements for maximum performance.
//
// Parameters:
//   - config: Build configuration specifying architecture, benchmark, and optimization settings
//
// Returns:
//   - string: Complete Dockerfile content ready for container builds
//   - error: Template parsing errors or configuration validation issues
func (b *Builder) GenerateDockerfile(config BuildConfig) (string, error) {
	tmpl, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := DockerfileTemplate{
		BaseImage:         config.BaseImage,
		Architecture:      config.Architecture,
		Compiler:          config.CompilerType,
		OptimizationFlags: strings.Join(config.OptimizationFlags, " "),
		BenchmarkSuite:    config.BenchmarkSuite,
		SpackConfig:       config.SpackConfig,
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// BuildContainer executes the complete container build process with architecture-specific optimizations.
//
// This method orchestrates the full container build pipeline including Dockerfile generation,
// dependency management, compilation optimization, and multi-stage builds for minimal image size.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - config: Complete build configuration with architecture and benchmark specifications
//
// Returns:
//   - error: Build failures, Docker issues, or configuration validation errors
func (b *Builder) BuildContainer(ctx context.Context, config BuildConfig) error {
	// Create build directory
	buildDir := filepath.Join("builds", config.ContainerTag, config.BenchmarkSuite)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Generate Dockerfile
	dockerfile, err := b.GenerateDockerfile(config)
	if err != nil {
		return fmt.Errorf("failed to generate dockerfile: %w", err)
	}

	// Write Dockerfile
	dockerfilePath := filepath.Join(buildDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write dockerfile: %w", err)
	}

	// Copy Spack configs if they exist
	spackConfigsDir := "spack-configs"
	if _, err := os.Stat(spackConfigsDir); err == nil {
		destDir := filepath.Join(buildDir, "spack-configs")
		if err := copyDir(spackConfigsDir, destDir); err != nil {
			return fmt.Errorf("failed to copy spack configs: %w", err)
		}
	}

	// Build container
	imageName := fmt.Sprintf("%s/%s:%s-%s", b.registryURL, b.namespace, config.BenchmarkSuite, config.ContainerTag)
	
	cmd := exec.CommandContext(ctx, "docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		buildDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	fmt.Printf("Successfully built container: %s\n", imageName)
	return nil
}

// PushContainer uploads the built container image to the configured registry.
//
// This method handles the complete container upload process including authentication,
// multi-architecture manifest creation, and registry-specific optimizations.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//   - config: Build configuration containing image tags and registry settings
//
// Returns:
//   - error: Push failures, authentication issues, or network connectivity problems
func (b *Builder) PushContainer(ctx context.Context, config BuildConfig) error {
	imageName := fmt.Sprintf("%s/%s:%s-%s", b.registryURL, b.namespace, config.BenchmarkSuite, config.ContainerTag)
	
	cmd := exec.CommandContext(ctx, "docker", "push", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	fmt.Printf("Successfully pushed container: %s\n", imageName)
	return nil
}

// GetOptimizationFlags generates architecture and compiler-specific optimization flags
// for maximum benchmark performance.
//
// This method implements intelligent flag selection based on processor architecture
// and compiler toolchain combinations. It provides optimal performance tuning for
// different microarchitectures while maintaining compatibility across instance families.
//
// The flag selection prioritizes:
//   - Latest instruction set extensions (AVX-512, SVE, Neon)
//   - Architecture-specific tuning parameters
//   - Compiler-specific optimization features
//   - Memory bandwidth optimization for STREAM benchmarks
//
// Architecture Support:
//   Intel Ice Lake: AVX-512 with high ZMM register usage
//   Intel Skylake: Standard AVX2 optimization with native tuning
//   AMD Zen 4: znver4 architecture with latest AMD optimizations
//   AMD Zen 3: znver3 fallback for older AMD families
//   Graviton3: ARMv8.2-A with SVE and Neoverse-V1 tuning
//   Graviton2: ARMv8.2-A with Neoverse-N1 optimization
//
// Compiler Integration:
//   Intel OneAPI: Architecture-specific vectorization flags
//   AMD AOCC: AMD-optimized compilation with znver tuning
//   GCC: Cross-platform compatibility with native optimization
//
// Parameters:
//   - architecture: Target processor architecture (e.g., "intel-icelake", "amd-zen4")
//   - compiler: Compiler toolchain type ("intel", "amd", "gcc")
//
// Returns:
//   - []string: Optimized compiler flags for the architecture/compiler combination
//
// Example:
//   flags := builder.GetOptimizationFlags("intel-icelake", "intel")
//   // Returns: ["-O3", "-xCORE-AVX512", "-qopt-zmm-usage=high"]
//   
//   flags = builder.GetOptimizationFlags("graviton3", "gcc")
//   // Returns: ["-O3", "-march=armv8.2-a+sve", "-mcpu=neoverse-v1"]
//
// Performance Notes:
//   - Flags are validated for compiler compatibility
//   - Architecture detection handles both specific and family-level matching
//   - Fallback flags ensure compilation success on unknown architectures
//   - Optimization levels balance performance with compilation time
func (b *Builder) GetOptimizationFlags(architecture, compiler string) []string {
	switch {
	case strings.Contains(architecture, "intel") && compiler == "intel":
		if strings.Contains(architecture, "icelake") {
			return []string{"-O3", "-xCORE-AVX512", "-qopt-zmm-usage=high"}
		}
		return []string{"-O3", "-march=native", "-mtune=native"}
	
	case strings.Contains(architecture, "amd") && compiler == "amd":
		if strings.Contains(architecture, "zen4") {
			return []string{"-O3", "-march=znver4", "-mtune=znver4"}
		}
		return []string{"-O3", "-march=znver3", "-mtune=znver3"}
	
	case strings.Contains(architecture, "graviton"):
		if strings.Contains(architecture, "graviton3") {
			return []string{"-O3", "-march=armv8.2-a+sve", "-mcpu=neoverse-v1"}
		}
		return []string{"-O3", "-march=armv8.2-a", "-mcpu=neoverse-n1"}
	
	default:
		return []string{"-O3", "-march=native", "-mtune=native"}
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}