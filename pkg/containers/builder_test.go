package containers

import (
	"strings"
	"testing"
)

func TestGenerateDockerfile(t *testing.T) {
	builder := NewBuilder("test-registry", "test-namespace")
	
	config := BuildConfig{
		Architecture:      "graviton3",
		ContainerTag:      "graviton3",
		BenchmarkSuite:    "stream",
		CompilerType:      "gcc",
		OptimizationFlags: []string{"-O3", "-march=armv8.2-a+sve"},
		BaseImage:         "ubuntu:22.04",
		SpackConfig:       "graviton3.yaml",
	}

	dockerfile, err := builder.GenerateDockerfile(config)
	if err != nil {
		t.Fatalf("GenerateDockerfile failed: %v", err)
	}

	// Check that dockerfile contains expected content
	if !strings.Contains(dockerfile, "FROM ubuntu:22.04") {
		t.Error("Dockerfile should contain base image")
	}

	if !strings.Contains(dockerfile, "graviton3") {
		t.Error("Dockerfile should contain architecture reference")
	}

	if !strings.Contains(dockerfile, "stream") {
		t.Error("Dockerfile should contain benchmark suite")
	}

	if !strings.Contains(dockerfile, "gcc-11") {
		t.Error("Dockerfile should contain GCC compiler installation")
	}

	// Should NOT contain Intel compiler for Graviton
	if strings.Contains(dockerfile, "intel-oneapi") {
		t.Error("Dockerfile should not contain Intel compiler for Graviton")
	}
}

func TestGetOptimizationFlags(t *testing.T) {
	builder := NewBuilder("test-registry", "test-namespace")

	testCases := []struct {
		architecture string
		compiler     string
		expected     []string
	}{
		{
			architecture: "intel-icelake",
			compiler:     "intel",
			expected:     []string{"-O3", "-xCORE-AVX512", "-qopt-zmm-usage=high"},
		},
		{
			architecture: "amd-zen4",
			compiler:     "amd",
			expected:     []string{"-O3", "-march=znver4", "-mtune=znver4"},
		},
		{
			architecture: "graviton3",
			compiler:     "gcc",
			expected:     []string{"-O3", "-march=armv8.2-a+sve", "-mcpu=neoverse-v1"},
		},
		{
			architecture: "unknown",
			compiler:     "gcc",
			expected:     []string{"-O3", "-march=native", "-mtune=native"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.architecture+"/"+tc.compiler, func(t *testing.T) {
			result := builder.GetOptimizationFlags(tc.architecture, tc.compiler)
			
			if len(result) != len(tc.expected) {
				t.Fatalf("Expected %d flags, got %d", len(tc.expected), len(result))
			}

			for i, flag := range result {
				if flag != tc.expected[i] {
					t.Errorf("Expected flag %d to be %s, got %s", i, tc.expected[i], flag)
				}
			}
		})
	}
}

func TestDockerfileTemplate(t *testing.T) {
	builder := NewBuilder("test-registry", "test-namespace")
	
	// Test all compiler types generate valid Dockerfiles
	compilerTypes := []string{"intel", "amd", "gcc"}
	
	for _, compilerType := range compilerTypes {
		t.Run("compiler_"+compilerType, func(t *testing.T) {
			config := BuildConfig{
				Architecture:      "test-arch",
				ContainerTag:      "test-tag",
				BenchmarkSuite:    "stream",
				CompilerType:      compilerType,
				OptimizationFlags: []string{"-O3"},
				BaseImage:         "ubuntu:22.04",
				SpackConfig:       "test.yaml",
			}

			dockerfile, err := builder.GenerateDockerfile(config)
			if err != nil {
				t.Fatalf("GenerateDockerfile failed for %s: %v", compilerType, err)
			}

			// Basic validation that dockerfile is not empty
			if len(dockerfile) < 100 {
				t.Errorf("Generated dockerfile seems too short: %d characters", len(dockerfile))
			}

			// Should contain multi-stage build
			if !strings.Contains(dockerfile, "FROM ubuntu:22.04 as builder") {
				t.Error("Dockerfile should use multi-stage build with builder stage")
			}

			if !strings.Contains(dockerfile, "FROM ubuntu:22.04 as runtime") {
				t.Error("Dockerfile should use multi-stage build with runtime stage")
			}
		})
	}
}