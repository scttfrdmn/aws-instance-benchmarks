package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaVersion represents a semantic version for schema compatibility
type SchemaVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// String returns the version as a string
func (v SchemaVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ParseVersion parses a version string into SchemaVersion
func ParseVersion(version string) (SchemaVersion, error) {
	var v SchemaVersion
	n, err := fmt.Sscanf(version, "%d.%d.%d", &v.Major, &v.Minor, &v.Patch)
	if err != nil || n != 3 {
		return v, fmt.Errorf("invalid version format: %s", version)
	}
	return v, nil
}

// IsCompatible checks if the current version is compatible with the required version
func (v SchemaVersion) IsCompatible(required SchemaVersion) bool {
	// Same major version and this version is >= required version
	if v.Major != required.Major {
		return false
	}
	if v.Minor > required.Minor {
		return true
	}
	if v.Minor == required.Minor && v.Patch >= required.Patch {
		return true
	}
	return false
}

// Validator handles JSON schema validation for benchmark results
type Validator struct {
	schemaPath string
	version    SchemaVersion
	schema     *gojsonschema.Schema
}

// NewValidator creates a new schema validator
func NewValidator(schemaPath string) (*Validator, error) {
	validator := &Validator{
		schemaPath: schemaPath,
	}
	
	// Load and parse the schema
	if err := validator.loadSchema(); err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}
	
	return validator, nil
}

// NewValidatorForVersion creates a validator for a specific schema version
func NewValidatorForVersion(basePath string, version SchemaVersion) (*Validator, error) {
	schemaPath := filepath.Join(basePath, fmt.Sprintf("v%d.%d", version.Major, version.Minor), "benchmark-result.json")
	return NewValidator(schemaPath)
}

// loadSchema loads the JSON schema from the file system
func (v *Validator) loadSchema() error {
	// Read schema file
	schemaFile, err := os.Open(v.schemaPath)
	if err != nil {
		return fmt.Errorf("failed to open schema file %s: %w", v.schemaPath, err)
	}
	defer schemaFile.Close()
	
	schemaBytes, err := io.ReadAll(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}
	
	// Parse schema to extract version info
	var schemaDoc map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaDoc); err != nil {
		return fmt.Errorf("failed to parse schema JSON: %w", err)
	}
	
	// Extract version from schema
	if versionStr, ok := schemaDoc["version"].(string); ok {
		version, err := ParseVersion(versionStr)
		if err != nil {
			return fmt.Errorf("invalid schema version: %w", err)
		}
		v.version = version
	} else {
		// Default to 1.0.0 if no version specified
		v.version = SchemaVersion{Major: 1, Minor: 0, Patch: 0}
	}
	
	// Compile the schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}
	
	v.schema = schema
	return nil
}

// GetVersion returns the schema version
func (v *Validator) GetVersion() SchemaVersion {
	return v.version
}

// ValidateBytes validates JSON data provided as bytes
func (v *Validator) ValidateBytes(data []byte) (*ValidationResult, error) {
	if v.schema == nil {
		return nil, fmt.Errorf("schema not loaded")
	}
	
	// Parse the data to check for schema_version field
	var dataDoc map[string]interface{}
	if err := json.Unmarshal(data, &dataDoc); err != nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("invalid JSON: %v", err)},
		}, nil
	}
	
	// Check schema version compatibility
	dataVersion := SchemaVersion{Major: 1, Minor: 0, Patch: 0} // default
	if schemaVersionStr, ok := dataDoc["schema_version"].(string); ok {
		var err error
		dataVersion, err = ParseVersion(schemaVersionStr)
		if err != nil {
			return &ValidationResult{
				Valid:  false,
				Errors: []string{fmt.Sprintf("invalid schema_version in data: %v", err)},
			}, nil
		}
	}
	
	// Validate the JSON against the schema
	documentLoader := gojsonschema.NewBytesLoader(data)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Collect validation errors
	var errors []string
	var warnings []string
	
	if !result.Valid() {
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
	}
	
	// Check version compatibility
	if !v.version.IsCompatible(dataVersion) {
		warnings = append(warnings, fmt.Sprintf("schema version mismatch: validator %s, data %s", v.version, dataVersion))
	}
	
	return &ValidationResult{
		Valid:        result.Valid(),
		Errors:       errors,
		Warnings:     warnings,
		SchemaVersion: v.version,
		DataVersion:  dataVersion,
	}, nil
}

// ValidateFile validates a JSON file
func (v *Validator) ValidateFile(filename string) (*ValidationResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	return v.ValidateBytes(data)
}

// ValidateResult validates a benchmark result interface
func (v *Validator) ValidateResult(result interface{}) (*ValidationResult, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	
	return v.ValidateBytes(data)
}

// ValidationResult contains the results of schema validation
type ValidationResult struct {
	Valid         bool            `json:"valid"`
	Errors        []string        `json:"errors,omitempty"`
	Warnings      []string        `json:"warnings,omitempty"`
	SchemaVersion SchemaVersion   `json:"schema_version"`
	DataVersion   SchemaVersion   `json:"data_version"`
}

// HasErrors returns true if there are validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// String returns a formatted string representation of the validation result
func (r *ValidationResult) String() string {
	var sb strings.Builder
	
	if r.Valid {
		sb.WriteString("✅ Validation passed")
	} else {
		sb.WriteString("❌ Validation failed")
	}
	
	sb.WriteString(fmt.Sprintf(" (Schema: %s, Data: %s)", r.SchemaVersion, r.DataVersion))
	
	if r.HasErrors() {
		sb.WriteString(fmt.Sprintf("\n\nErrors (%d):", len(r.Errors)))
		for _, err := range r.Errors {
			sb.WriteString(fmt.Sprintf("\n  - %s", err))
		}
	}
	
	if r.HasWarnings() {
		sb.WriteString(fmt.Sprintf("\n\nWarnings (%d):", len(r.Warnings)))
		for _, warning := range r.Warnings {
			sb.WriteString(fmt.Sprintf("\n  - %s", warning))
		}
	}
	
	return sb.String()
}

// SchemaManager handles multiple schema versions and migration
type SchemaManager struct {
	schemaPath string
	validators map[string]*Validator // version -> validator
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(schemaPath string) *SchemaManager {
	return &SchemaManager{
		schemaPath: schemaPath,
		validators: make(map[string]*Validator),
	}
}

// GetValidator returns a validator for the specified version
func (m *SchemaManager) GetValidator(version SchemaVersion) (*Validator, error) {
	versionStr := version.String()
	
	if validator, exists := m.validators[versionStr]; exists {
		return validator, nil
	}
	
	// Load validator for this version
	validator, err := NewValidatorForVersion(m.schemaPath, version)
	if err != nil {
		return nil, err
	}
	
	m.validators[versionStr] = validator
	return validator, nil
}

// GetLatestValidator returns the validator for the latest schema version
func (m *SchemaManager) GetLatestValidator() (*Validator, error) {
	// For now, assume latest is v1.0
	return m.GetValidator(SchemaVersion{Major: 1, Minor: 0, Patch: 0})
}

// ValidateWithVersionDetection validates data and automatically detects the appropriate schema version
func (m *SchemaManager) ValidateWithVersionDetection(data []byte) (*ValidationResult, error) {
	// First, try to parse the data to extract the schema version
	var dataDoc map[string]interface{}
	if err := json.Unmarshal(data, &dataDoc); err != nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("invalid JSON: %v", err)},
		}, nil
	}
	
	// Extract schema version from data
	var dataVersion SchemaVersion
	if schemaVersionStr, ok := dataDoc["schema_version"].(string); ok {
		var err error
		dataVersion, err = ParseVersion(schemaVersionStr)
		if err != nil {
			// Try with latest validator if version parsing fails
			validator, err := m.GetLatestValidator()
			if err != nil {
				return nil, err
			}
			return validator.ValidateBytes(data)
		}
	} else {
		// Default to 1.0.0 if no version specified
		dataVersion = SchemaVersion{Major: 1, Minor: 0, Patch: 0}
	}
	
	// Get appropriate validator
	validator, err := m.GetValidator(dataVersion)
	if err != nil {
		return nil, fmt.Errorf("no validator available for version %s: %w", dataVersion, err)
	}
	
	return validator.ValidateBytes(data)
}

// DefaultSchemaManager creates a schema manager with the default schema path
func DefaultSchemaManager() *SchemaManager {
	return NewSchemaManager("data/schemas")
}