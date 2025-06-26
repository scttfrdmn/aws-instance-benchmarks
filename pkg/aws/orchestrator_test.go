package aws

import (
	"strings"
	"testing"
)

func TestArchitectureDetection(t *testing.T) {
	tests := []struct {
		instanceType string
		expected     string
	}{
		// Intel x86_64 instances
		{"m7i.large", "x86_64"},
		{"c7i.xlarge", "x86_64"},
		{"r7i.2xlarge", "x86_64"},
		{"m6i.medium", "x86_64"},
		{"c6i.large", "x86_64"},
		
		// AMD x86_64 instances  
		{"m7a.large", "x86_64"},
		{"c7a.xlarge", "x86_64"},
		{"r7a.2xlarge", "x86_64"},
		{"m6a.medium", "x86_64"},
		{"c6a.large", "x86_64"},
		
		// Graviton ARM64 instances
		{"m7g.large", "arm64"},
		{"c7g.xlarge", "arm64"},
		{"r7g.2xlarge", "arm64"},
		{"m6g.medium", "arm64"},
		{"c6g.large", "arm64"},
		{"t4g.micro", "arm64"},
		{"t4g.nano", "arm64"},
		
		// Edge cases
		{"m5.large", "x86_64"},    // No 'g' in name
		{"c5.xlarge", "x86_64"},   // No 'g' in name
		{"x1e.xlarge", "x86_64"},  // Different family prefix
	}
	
	for _, tt := range tests {
		t.Run(tt.instanceType, func(t *testing.T) {
			// We need to test the internal logic, so we'll extract it to a helper function
			// or test it through the actual AMI selection
			architecture := detectArchitectureFromInstanceType(tt.instanceType)
			if architecture != tt.expected {
				t.Errorf("detectArchitectureFromInstanceType(%s) = %s, want %s", 
					tt.instanceType, architecture, tt.expected)
			}
		})
	}
}

// Helper function extracted from getLatestAMI for testing
func detectArchitectureFromInstanceType(instanceType string) string {
	architecture := "x86_64"
	// Check for Graviton instances (end with 'g' after the size, e.g., m7g.large, c7g.xlarge)
	if strings.Contains(instanceType, "g.") || strings.HasSuffix(instanceType, "g") {
		if strings.HasPrefix(instanceType, "m") || strings.HasPrefix(instanceType, "c") || 
			strings.HasPrefix(instanceType, "r") || strings.HasPrefix(instanceType, "t") {
			architecture = "arm64" // Graviton instances
		}
	}
	return architecture
}

func TestOrchestratorCreation(t *testing.T) {
	orchestrator, err := NewOrchestrator("us-east-1")
	if err != nil {
		t.Fatalf("NewOrchestrator failed: %v", err)
	}
	
	if orchestrator.region != "us-east-1" {
		t.Errorf("Expected region us-east-1, got %s", orchestrator.region)
	}
	
	if orchestrator.ec2Client == nil {
		t.Error("EC2 client should not be nil")
	}
}