package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Migration represents a schema migration from one version to another
type Migration interface {
	// GetSourceVersion returns the version this migration upgrades from
	GetSourceVersion() SchemaVersion
	
	// GetTargetVersion returns the version this migration upgrades to
	GetTargetVersion() SchemaVersion
	
	// Migrate performs the migration on the provided data
	Migrate(data map[string]interface{}) (map[string]interface{}, error)
	
	// GetDescription returns a human-readable description of the migration
	GetDescription() string
}

// MigrationRegistry manages available migrations
type MigrationRegistry struct {
	migrations map[string]Migration // key: "source.version->target.version"
}

// NewMigrationRegistry creates a new migration registry
func NewMigrationRegistry() *MigrationRegistry {
	registry := &MigrationRegistry{
		migrations: make(map[string]Migration),
	}
	
	// Register built-in migrations
	registry.registerBuiltInMigrations()
	
	return registry
}

// registerBuiltInMigrations registers the built-in schema migrations
func (r *MigrationRegistry) registerBuiltInMigrations() {
	// Future migrations will be registered here
	// Example: r.RegisterMigration(&Migration1_0_0To1_1_0{})
}

// RegisterMigration registers a new migration
func (r *MigrationRegistry) RegisterMigration(migration Migration) error {
	key := fmt.Sprintf("%s->%s", migration.GetSourceVersion(), migration.GetTargetVersion())
	
	if _, exists := r.migrations[key]; exists {
		return fmt.Errorf("migration %s already registered", key)
	}
	
	r.migrations[key] = migration
	return nil
}

// GetMigration returns a migration for the specified version transition
func (r *MigrationRegistry) GetMigration(from, to SchemaVersion) (Migration, error) {
	key := fmt.Sprintf("%s->%s", from, to)
	
	migration, exists := r.migrations[key]
	if !exists {
		return nil, fmt.Errorf("no migration available from %s to %s", from, to)
	}
	
	return migration, nil
}

// GetMigrationPath returns a sequence of migrations to upgrade from source to target
func (r *MigrationRegistry) GetMigrationPath(from, to SchemaVersion) ([]Migration, error) {
	// For now, only support direct migrations
	// In the future, this could implement pathfinding for multi-step migrations
	migration, err := r.GetMigration(from, to)
	if err != nil {
		return nil, err
	}
	
	return []Migration{migration}, nil
}

// ListMigrations returns all available migrations
func (r *MigrationRegistry) ListMigrations() []Migration {
	var migrations []Migration
	for _, migration := range r.migrations {
		migrations = append(migrations, migration)
	}
	return migrations
}

// Migrator handles schema migrations for benchmark data
type Migrator struct {
	registry *MigrationRegistry
}

// NewMigrator creates a new migrator with the default migration registry
func NewMigrator() *Migrator {
	return &Migrator{
		registry: NewMigrationRegistry(),
	}
}

// NewMigratorWithRegistry creates a new migrator with a custom registry
func NewMigratorWithRegistry(registry *MigrationRegistry) *Migrator {
	return &Migrator{
		registry: registry,
	}
}

// MigrateData migrates benchmark data from one schema version to another
func (m *Migrator) MigrateData(data map[string]interface{}, targetVersion SchemaVersion) (map[string]interface{}, error) {
	// Extract current version from data
	currentVersion, err := m.extractVersionFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract version from data: %w", err)
	}
	
	// Check if migration is needed
	if currentVersion.String() == targetVersion.String() {
		return data, nil // No migration needed
	}
	
	// Get migration path
	migrations, err := m.registry.GetMigrationPath(currentVersion, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find migration path: %w", err)
	}
	
	// Apply migrations in sequence
	result := data
	for _, migration := range migrations {
		fmt.Printf("Applying migration: %s\n", migration.GetDescription())
		result, err = migration.Migrate(result)
		if err != nil {
			return nil, fmt.Errorf("migration failed (%s): %w", migration.GetDescription(), err)
		}
	}
	
	return result, nil
}

// MigrateFile migrates a JSON file from one schema version to another
func (m *Migrator) MigrateFile(inputFile, outputFile string, targetVersion SchemaVersion) error {
	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	
	// Parse JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Migrate data
	migratedData, err := m.MigrateData(jsonData, targetVersion)
	if err != nil {
		return err
	}
	
	// Write output file
	output, err := json.MarshalIndent(migratedData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal migrated data: %w", err)
	}
	
	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	
	return nil
}

// extractVersionFromData extracts the schema version from benchmark data
func (m *Migrator) extractVersionFromData(data map[string]interface{}) (SchemaVersion, error) {
	// Check for schema_version field
	if versionStr, ok := data["schema_version"].(string); ok {
		return ParseVersion(versionStr)
	}
	
	// Check for legacy data_version in metadata
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if dataVersion, ok := metadata["data_version"].(string); ok {
			if dataVersion == "1.0" {
				return SchemaVersion{Major: 1, Minor: 0, Patch: 0}, nil
			}
		}
	}
	
	// Default to 1.0.0 for legacy data
	return SchemaVersion{Major: 1, Minor: 0, Patch: 0}, nil
}

// BatchMigrator handles batch migration of multiple files
type BatchMigrator struct {
	migrator *Migrator
}

// NewBatchMigrator creates a new batch migrator
func NewBatchMigrator() *BatchMigrator {
	return &BatchMigrator{
		migrator: NewMigrator(),
	}
}

// MigrateDirectory migrates all JSON files in a directory
func (b *BatchMigrator) MigrateDirectory(inputDir, outputDir string, targetVersion SchemaVersion) error {
	// Walk through input directory
	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}
		
		// Calculate output path
		outputPath := filepath.Join(outputDir, relPath)
		
		fmt.Printf("Migrating %s -> %s\n", path, outputPath)
		
		// Migrate file
		if err := b.migrator.MigrateFile(path, outputPath, targetVersion); err != nil {
			fmt.Printf("Failed to migrate %s: %v\n", path, err)
			return err
		}
		
		return nil
	})
}

// MigrationReport contains information about a migration operation
type MigrationReport struct {
	SourceVersion   SchemaVersion `json:"source_version"`
	TargetVersion   SchemaVersion `json:"target_version"`
	FilesProcessed  int          `json:"files_processed"`
	FilesSucceeded  int          `json:"files_succeeded"`
	FilesFailed     int          `json:"files_failed"`
	Errors         []string      `json:"errors,omitempty"`
}

// GenerateReport creates a migration report
func (b *BatchMigrator) GenerateReport(inputDir string, targetVersion SchemaVersion) (*MigrationReport, error) {
	report := &MigrationReport{
		TargetVersion: targetVersion,
		Errors:       []string{},
	}
	
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		
		report.FilesProcessed++
		
		// Read and analyze file
		data, err := os.ReadFile(path)
		if err != nil {
			report.FilesFailed++
			report.Errors = append(report.Errors, fmt.Sprintf("%s: failed to read file", path))
			return nil // Continue processing other files
		}
		
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			report.FilesFailed++
			report.Errors = append(report.Errors, fmt.Sprintf("%s: invalid JSON", path))
			return nil
		}
		
		// Extract version
		version, err := b.migrator.extractVersionFromData(jsonData)
		if err != nil {
			report.FilesFailed++
			report.Errors = append(report.Errors, fmt.Sprintf("%s: failed to extract version", path))
			return nil
		}
		
		// Set source version from first file
		if report.FilesProcessed == 1 {
			report.SourceVersion = version
		}
		
		// Check if migration is possible
		_, err = b.migrator.registry.GetMigrationPath(version, targetVersion)
		if err != nil {
			report.FilesFailed++
			report.Errors = append(report.Errors, fmt.Sprintf("%s: no migration path from %s to %s", path, version, targetVersion))
			return nil
		}
		
		report.FilesSucceeded++
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return report, nil
}

// Example migration implementation for future use
// This would be used when we create v1.1.0 schema

/*
type Migration1_0_0To1_1_0 struct{}

func (m *Migration1_0_0To1_1_0) GetSourceVersion() SchemaVersion {
	return SchemaVersion{Major: 1, Minor: 0, Patch: 0}
}

func (m *Migration1_0_0To1_1_0) GetTargetVersion() SchemaVersion {
	return SchemaVersion{Major: 1, Minor: 1, Patch: 0}
}

func (m *Migration1_0_0To1_1_0) GetDescription() string {
	return "Migrate from schema v1.0.0 to v1.1.0: Add CPU benchmark fields"
}

func (m *Migration1_0_0To1_1_0) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
	// Update schema version
	data["schema_version"] = "1.1.0"
	
	// Add new CPU section if it doesn't exist
	if performance, ok := data["performance"].(map[string]interface{}); ok {
		if _, hasCPU := performance["cpu"]; !hasCPU {
			performance["cpu"] = map[string]interface{}{
				"linpack": map[string]interface{}{
					"gflops": 0,
					"efficiency": 0,
				},
			}
		}
	}
	
	return data, nil
}
*/