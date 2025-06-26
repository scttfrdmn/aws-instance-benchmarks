package discovery

import (
	"testing"
)

const (
	// Test constants to avoid goconst linter warnings.
	x86Arch = "x86_64"
)

func TestExtractInstanceFamily(t *testing.T) {
	testCases := []struct {
		instanceType string
		expected     string
	}{
		{"m7i.large", "m7i"},
		{"c7g.xlarge", "c7g"},
		{"r7a.2xlarge", "r7a"},
		{"t3.micro", "t3"},
		{"inf2.large", "inf2"},
		{"trn1.32xlarge", "trn1"},
		{"x2gd.medium", "x2gd"},
		{"invalid", "invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.instanceType, func(t *testing.T) {
			result := extractInstanceFamily(tc.instanceType)
			if result != tc.expected {
				t.Errorf("extractInstanceFamily(%s) = %s; want %s", tc.instanceType, result, tc.expected)
			}
		})
	}
}

func TestDetermineContainerTag(t *testing.T) {
	testCases := []struct {
		name     string
		instance InstanceInfo
		expected string
	}{
		{
			name: "Intel Ice Lake (m7i)",
			instance: InstanceInfo{
				InstanceFamily: "m7i",
				Architecture:   x86Arch,
				ProcessorInfo:  "Intel",
			},
			expected: "intel-icelake",
		},
		{
			name: "Intel Skylake (m5)",
			instance: InstanceInfo{
				InstanceFamily: "m5",
				Architecture:   x86Arch,
				ProcessorInfo:  "Intel",
			},
			expected: "intel-skylake",
		},
		{
			name: "AMD Zen 4 (m7a)",
			instance: InstanceInfo{
				InstanceFamily: "m7a",
				Architecture:   x86Arch,
				ProcessorInfo:  "AMD",
			},
			expected: "amd-zen4",
		},
		{
			name: "AMD Zen 3 (m6a)",
			instance: InstanceInfo{
				InstanceFamily: "m6a",
				Architecture:   x86Arch,
				ProcessorInfo:  "AMD",
			},
			expected: "amd-zen3",
		},
		{
			name: "Graviton3 (m7g)",
			instance: InstanceInfo{
				InstanceFamily: "m7g",
				Architecture:   "arm64",
				ProcessorInfo:  "AWS",
			},
			expected: "graviton3",
		},
		{
			name: "Graviton2 (m6g)",
			instance: InstanceInfo{
				InstanceFamily: "m6g",
				Architecture:   "arm64",
				ProcessorInfo:  "AWS",
			},
			expected: "graviton2",
		},
		{
			name: "Unknown architecture",
			instance: InstanceInfo{
				InstanceFamily: "unknown",
				Architecture:   "riscv",
				ProcessorInfo:  "Unknown",
			},
			expected: "riscv-unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := determineContainerTag(tc.instance)
			if result != tc.expected {
				t.Errorf("determineContainerTag(%+v) = %s; want %s", tc.instance, result, tc.expected)
			}
		})
	}
}

func TestGenerateArchitectureMappings(t *testing.T) {
	discoverer := &InstanceDiscoverer{}
	
	instances := []InstanceInfo{
		{
			InstanceType:   "m7i.large",
			InstanceFamily: "m7i",
			Architecture:   x86Arch,
			ProcessorInfo:  "Intel",
		},
		{
			InstanceType:   "m7i.xlarge",
			InstanceFamily: "m7i",
			Architecture:   x86Arch,
			ProcessorInfo:  "Intel",
		},
		{
			InstanceType:   "c7g.large",
			InstanceFamily: "c7g",
			Architecture:   "arm64",
			ProcessorInfo:  "AWS",
		},
		{
			InstanceType:   "r7a.large",
			InstanceFamily: "r7a",
			Architecture:   x86Arch,
			ProcessorInfo:  "AMD",
		},
	}

	mappings := discoverer.GenerateArchitectureMappings(instances)

	// Should have 3 unique families (m7i, c7g, r7a)
	expectedFamilies := 3
	if len(mappings) != expectedFamilies {
		t.Errorf("Expected %d families, got %d", expectedFamilies, len(mappings))
	}

	// Check specific mappings
	if mapping, exists := mappings["m7i"]; exists {
		if mapping.ContainerTag != "intel-icelake" {
			t.Errorf("Expected m7i to map to intel-icelake, got %s", mapping.ContainerTag)
		}
		if mapping.Architecture != x86Arch {
			t.Errorf("Expected m7i architecture to be x86_64, got %s", mapping.Architecture)
		}
	} else {
		t.Error("Expected m7i family in mappings")
	}

	if mapping, exists := mappings["c7g"]; exists {
		if mapping.ContainerTag != "graviton3" {
			t.Errorf("Expected c7g to map to graviton3, got %s", mapping.ContainerTag)
		}
	} else {
		t.Error("Expected c7g family in mappings")
	}

	if mapping, exists := mappings["r7a"]; exists {
		if mapping.ContainerTag != "amd-zen4" {
			t.Errorf("Expected r7a to map to amd-zen4, got %s", mapping.ContainerTag)
		}
	} else {
		t.Error("Expected r7a family in mappings")
	}
}

func TestArchitectureMappingDeduplication(t *testing.T) {
	discoverer := &InstanceDiscoverer{}
	
	// Multiple instances of same family should result in single mapping
	instances := []InstanceInfo{
		{
			InstanceType:   "m7i.large",
			InstanceFamily: "m7i",
			Architecture:   x86Arch,
			ProcessorInfo:  "Intel",
		},
		{
			InstanceType:   "m7i.xlarge",
			InstanceFamily: "m7i",
			Architecture:   x86Arch,
			ProcessorInfo:  "Intel",
		},
		{
			InstanceType:   "m7i.2xlarge",
			InstanceFamily: "m7i",
			Architecture:   x86Arch,
			ProcessorInfo:  "Intel",
		},
	}

	mappings := discoverer.GenerateArchitectureMappings(instances)

	// Should only have 1 family mapping despite 3 instances
	if len(mappings) != 1 {
		t.Errorf("Expected 1 family mapping, got %d", len(mappings))
	}

	if _, exists := mappings["m7i"]; !exists {
		t.Error("Expected m7i family in mappings")
	}
}