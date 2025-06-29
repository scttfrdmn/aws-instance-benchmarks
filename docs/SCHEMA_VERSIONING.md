# Schema Versioning and Data Migration

This document describes the comprehensive schema versioning strategy for AWS Instance Benchmarks, including validation, migration, and compatibility management.

## Overview

AWS Instance Benchmarks uses a robust schema versioning system to ensure:
- **Data integrity** across all benchmark submissions
- **Backward compatibility** for existing data consumers
- **Forward migration** capabilities for schema evolution
- **Community contribution validation** with automatic quality checks
- **API stability** for tools like ComputeCompass

## Schema Versioning Strategy

### Semantic Versioning

We use semantic versioning (semver) for schema versions with the format `MAJOR.MINOR.PATCH`:

- **MAJOR**: Breaking changes that require data migration
- **MINOR**: Backward-compatible additions (new optional fields)
- **PATCH**: Bug fixes and clarifications

```json
{
  "schema_version": "1.0.0",
  "version": "1.0.0"
}
```

### Current Schema Version: 1.0.0

The current schema (v1.0.0) includes:
- **Required top-level fields**: `schema_version`, `metadata`, `performance`, `validation`
- **Comprehensive metadata**: Instance type, architecture, environment details
- **Performance data**: STREAM memory benchmarks, HPL CPU benchmarks
- **Validation requirements**: Checksums, reproducibility metrics

## Schema Directory Structure

```
data/schemas/
├── v1.0/
│   └── benchmark-result.json          # Schema v1.0.0
├── v1.1/                              # Future minor version
│   └── benchmark-result.json
└── v2.0/                              # Future major version
    └── benchmark-result.json
```

## Schema Validation

### CLI Tool Integration

The `aws-benchmark-collector` includes built-in schema validation:

```bash
# Validate a single file
./aws-benchmark-collector schema validate results/benchmark.json --version 1.0.0

# Validate a directory
./aws-benchmark-collector schema validate data/contributions/ --version 1.0.0

# Generate migration report
./aws-benchmark-collector schema migrate data/legacy/ data/migrated/ --version 1.0.0 --report-only
```

### Automatic Validation

Schema validation is automatically performed:
1. **During benchmark execution**: All results validated before storage
2. **Community contributions**: GitHub Actions validate PRs automatically
3. **Data processing**: Aggregation tools validate input data

### Validation Output

```
Validating file: benchmark.json
✅ Validation passed (Schema: 1.0.0, Data: 1.0.0)

# On validation errors:
❌ Validation failed (Schema: 1.0.0, Data: 1.0.0)

Errors (3):
  - metadata: instanceType is required
  - performance: Required property missing
  - validation.reproducibility: runs must be >= 1
```

## Data Migration Framework

### Migration Architecture

The migration system supports:
- **Single file migration**: Transform individual benchmark files
- **Batch migration**: Process entire directories with progress tracking
- **Version detection**: Automatic source version identification
- **Migration chaining**: Multi-step migrations for complex upgrades

### Migration Commands

```bash
# Migrate a single file
./aws-benchmark-collector schema migrate \
  legacy-data.json \
  migrated-data.json \
  --version 1.1.0

# Batch migrate a directory
./aws-benchmark-collector schema migrate \
  data/legacy/ \
  data/migrated/ \
  --version 1.1.0

# Generate migration report without migrating
./aws-benchmark-collector schema migrate \
  data/legacy/ \
  data/migrated/ \
  --version 1.1.0 \
  --report-only
```

### Migration Report Example

```
Migration Report for: data/legacy/
  Source version: 1.0.0
  Target version: 1.1.0
  Files processed: 150
  Files that can be migrated: 147
  Files with issues: 3

Issues found:
  - benchmark-old.json: Missing required metadata fields
  - invalid-format.json: Invalid JSON format
  - corrupted-data.json: No migration path from 0.9.0 to 1.1.0
```

## Schema Evolution Examples

### Example: Adding CPU Benchmarks (v1.0.0 → v1.1.0)

When adding new CPU benchmark fields in a minor version update:

```go
type Migration1_0_0To1_1_0 struct{}

func (m *Migration1_0_0To1_1_0) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
    // Update schema version
    data["schema_version"] = "1.1.0"
    
    // Add CPU section if missing (backward compatible)
    if performance, ok := data["performance"].(map[string]interface{}); ok {
        if _, hasCPU := performance["cpu"]; !hasCPU {
            performance["cpu"] = map[string]interface{}{
                "linpack": map[string]interface{}{
                    "gflops": nil,      // Optional in v1.1.0
                    "efficiency": nil,  // Optional in v1.1.0
                },
            }
        }
    }
    
    return data, nil
}
```

### Example: Breaking Changes (v1.x.x → v2.0.0)

Major version changes might restructure data formats:

```go
type Migration1_x_xTo2_0_0 struct{}

func (m *Migration1_x_xTo2_0_0) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
    // Major restructuring - convert old format to new
    newData := map[string]interface{}{
        "schema_version": "2.0.0",
        "format_version": "2.0",  // New field in v2.0
    }
    
    // Transform existing data structure
    if oldMetadata, ok := data["metadata"].(map[string]interface{}); ok {
        newData["instance"] = map[string]interface{}{
            "type":         oldMetadata["instanceType"],
            "family":       oldMetadata["instanceFamily"],
            "architecture": oldMetadata["processorArchitecture"],
            "region":       oldMetadata["region"],
        }
    }
    
    return newData, nil
}
```

## Community Contribution Integration

### GitHub Actions Validation

All community contributions are automatically validated:

```yaml
- name: Schema validation with built-in tools
  run: |
    # Validate all contribution files
    find data/contributions -name "*.json" -type f | while read file; do
      ./aws-benchmark-collector schema validate "$file" --version 1.0.0
    done
```

### Contribution Requirements

Community contributions must:
1. **Pass schema validation**: Conform to current schema version
2. **Include required metadata**: Complete instance and environment information
3. **Provide validation data**: Checksums and reproducibility metrics
4. **Meet quality standards**: Statistical quality score ≥ 0.7

### Validation Script Integration

The validation script automatically uses the built-in schema validator:

```bash
# Built-in validation
if [ -f "./aws-benchmark-collector" ]; then
    ./aws-benchmark-collector schema validate "$TEST_CONTRIBUTION" --version 1.0.0
fi
```

## Compatibility Management

### Version Compatibility Matrix

| Data Version | Schema v1.0 | Schema v1.1 | Schema v2.0 |
|--------------|-------------|-------------|-------------|
| 1.0.0        | ✅ Native   | ✅ Compatible | ❌ Migration Required |
| 1.1.0        | ⚠️ Partial  | ✅ Native   | ❌ Migration Required |
| 2.0.0        | ❌ Incompatible | ❌ Incompatible | ✅ Native |

### Backward Compatibility Rules

- **Minor versions**: Must be backward compatible
- **Major versions**: May introduce breaking changes
- **Patch versions**: Bug fixes only, full compatibility

### API Stability Guarantees

For API consumers (like ComputeCompass):
- **v1.x.x schemas**: Guarantee specific field access patterns
- **Migration tools**: Available for all supported transitions
- **Deprecation notice**: 6 months minimum for breaking changes

## Implementation Details

### Schema Loading

```go
// Load validator for specific version
validator, err := schema.NewValidatorForVersion("data/schemas", 
    schema.SchemaVersion{Major: 1, Minor: 0, Patch: 0})

// Validate data
result, err := validator.ValidateBytes(jsonData)
if !result.Valid {
    // Handle validation errors
    for _, errMsg := range result.Errors {
        log.Printf("Validation error: %s", errMsg)
    }
}
```

### Version Detection

```go
// Automatic version detection from data
schemaManager := schema.DefaultSchemaManager()
result, err := schemaManager.ValidateWithVersionDetection(jsonData)

// Version compatibility checking
validator := schema.NewValidator("data/schemas/v1.0/benchmark-result.json")
if !validator.GetVersion().IsCompatible(requiredVersion) {
    return fmt.Errorf("incompatible schema version")
}
```

### Migration Registration

```go
// Register custom migrations
registry := schema.NewMigrationRegistry()
registry.RegisterMigration(&CustomMigration{})

// Get migration path
migrations, err := registry.GetMigrationPath(sourceVersion, targetVersion)
```

## Best Practices

### For Schema Design

1. **Additive changes**: Prefer adding optional fields over modifying existing ones
2. **Clear semantics**: Use descriptive field names and include documentation
3. **Future-proofing**: Consider extensibility when designing new schemas
4. **Validation rules**: Include comprehensive constraints and examples

### For Data Producers

1. **Always include version**: Set `schema_version` field in all data
2. **Validate locally**: Use the CLI tool before submitting contributions
3. **Complete metadata**: Provide all required environment and validation information
4. **Quality assurance**: Ensure multiple runs and statistical validation

### For Data Consumers

1. **Version checking**: Always validate schema compatibility
2. **Graceful degradation**: Handle missing optional fields appropriately
3. **Migration support**: Implement automatic data migration when possible
4. **Error handling**: Provide clear error messages for schema mismatches

## Troubleshooting

### Common Validation Errors

```bash
# Missing required fields
❌ metadata: instanceType is required
# Solution: Add missing metadata.instanceType field

# Invalid format
❌ metadata.region: Does not match pattern "^[a-z]+-[a-z]+-[0-9]+$"
# Solution: Use valid AWS region format (e.g., "us-east-1")

# Version mismatch
⚠️ schema version mismatch: validator 1.0.0, data 0.9.0
# Solution: Migrate data or use appropriate validator version
```

### Migration Issues

```bash
# No migration path
❌ no migration available from 0.8.0 to 1.0.0
# Solution: Implement custom migration or upgrade incrementally

# Migration failure
❌ migration failed: missing required field after transformation
# Solution: Review migration logic and add missing field handling
```

### Performance Optimization

- **Batch validation**: Process multiple files in parallel
- **Schema caching**: Reuse compiled schemas across validations  
- **Selective migration**: Only migrate files that need updates
- **Progress tracking**: Monitor large batch operations

## Future Enhancements

### Planned Features

1. **Real-time validation**: WebSocket-based validation for live data streams
2. **Schema registry**: Centralized schema management with versioning APIs
3. **Custom validators**: Plugin system for domain-specific validation rules
4. **Migration testing**: Automated test suite for migration correctness

### Integration Roadmap

- **ComputeCompass v2**: Native schema version support
- **Community tools**: SDK for schema validation in external tools
- **Cloud integration**: Schema validation in AWS Lambda functions
- **Monitoring**: Schema compliance metrics in CloudWatch

## References

- [JSON Schema Specification](https://json-schema.org/)
- [Semantic Versioning](https://semver.org/)
- [Community Contribution Workflow](COMMUNITY_WORKFLOW.md)
- [Statistical Validation Guide](STATISTICAL_VALIDATION.md)
- [CloudWatch Integration](CLOUDWATCH_INTEGRATION.md)

---

For questions about schema versioning or migration issues, please:
- Open an issue: [GitHub Issues](https://github.com/scttfrdmn/aws-instance-benchmarks/issues)
- Join discussions: [GitHub Discussions](https://github.com/scttfrdmn/aws-instance-benchmarks/discussions)
- Contact maintainers: [benchmarks@computecompass.dev](mailto:benchmarks@computecompass.dev)