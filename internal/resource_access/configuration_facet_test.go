package resource_access

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// TestUnit_ConfigurationFacet_LoadStore tests basic Load and Store operations
func TestUnit_ConfigurationFacet_LoadStore(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repository
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repository.Close()

	// Create configuration facet
	logger := utilities.NewLoggingUtility()
	configFacet := newConfigurationFacet(repository, logger)

	// Test data
	configType := "boards"
	identifier := "default"
	testData := ConfigurationData{
		Type:       configType,
		Identifier: identifier,
		Version:    "1.0",
		Settings: map[string]interface{}{
			"name":    "Test Board",
			"columns": []string{"todo", "doing", "done"},
			"theme":   "light",
		},
		Schema: "board-v1",
		Metadata: map[string]string{
			"created_by": "test",
		},
	}

	// Test Store operation
	err = configFacet.Store(configType, identifier, testData)
	if err != nil {
		t.Fatalf("Store operation failed: %v", err)
	}

	// Test Load operation
	loadedData, err := configFacet.Load(configType, identifier)
	if err != nil {
		t.Fatalf("Load operation failed: %v", err)
	}

	// Verify loaded data matches stored data
	if loadedData.Type != testData.Type {
		t.Errorf("Type mismatch: expected %s, got %s", testData.Type, loadedData.Type)
	}
	if loadedData.Identifier != testData.Identifier {
		t.Errorf("Identifier mismatch: expected %s, got %s", testData.Identifier, loadedData.Identifier)
	}
	if loadedData.Version != testData.Version {
		t.Errorf("Version mismatch: expected %s, got %s", testData.Version, loadedData.Version)
	}
	if loadedData.Schema != testData.Schema {
		t.Errorf("Schema mismatch: expected %s, got %s", testData.Schema, loadedData.Schema)
	}

	// Verify settings content
	if loadedData.Settings["name"] != testData.Settings["name"] {
		t.Errorf("Settings content mismatch for name")
	}
	if loadedData.Settings["theme"] != testData.Settings["theme"] {
		t.Errorf("Settings content mismatch for theme")
	}

	// Verify metadata was updated during store
	if loadedData.Metadata["last_updated"] == "" {
		t.Errorf("Expected last_updated metadata to be set")
	}
	if loadedData.Metadata["storage_version"] != "1.0" {
		t.Errorf("Expected storage_version to be 1.0")
	}
}

// TestUnit_ConfigurationFacet_LoadDefault tests loading non-existent configuration returns default
func TestUnit_ConfigurationFacet_LoadDefault(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repository
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repository.Close()

	// Create configuration facet
	logger := utilities.NewLoggingUtility()
	configFacet := newConfigurationFacet(repository, logger)

	// Test loading non-existent configuration
	configType := "workflows"
	identifier := "non-existent"
	defaultData, err := configFacet.Load(configType, identifier)
	if err != nil {
		t.Fatalf("Load operation failed: %v", err)
	}

	// Verify default data structure
	if defaultData.Type != configType {
		t.Errorf("Expected type %s, got %s", configType, defaultData.Type)
	}
	if defaultData.Identifier != identifier {
		t.Errorf("Expected identifier %s, got %s", identifier, defaultData.Identifier)
	}
	if defaultData.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", defaultData.Version)
	}
	if defaultData.Settings == nil {
		t.Errorf("Expected settings map to be initialized")
	}
	if defaultData.Schema != "default" {
		t.Errorf("Expected schema to be default, got %s", defaultData.Schema)
	}
	if defaultData.Metadata["default"] != "true" {
		t.Errorf("Expected default metadata to be true")
	}
}

// TestUnit_ConfigurationFacet_ValidationErrors tests validation error conditions
func TestUnit_ConfigurationFacet_ValidationErrors(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repository
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repository.Close()

	// Create configuration facet
	logger := utilities.NewLoggingUtility()
	configFacet := newConfigurationFacet(repository, logger)

	// Test empty configuration type
	_, err = configFacet.Load("", "identifier")
	if err == nil {
		t.Errorf("Expected error for empty configuration type")
	}

	// Test empty identifier
	_, err = configFacet.Load("boards", "")
	if err == nil {
		t.Errorf("Expected error for empty identifier")
	}

	// Test invalid configuration data - missing type
	invalidData := ConfigurationData{
		Identifier: "test",
		Version:    "1.0",
		Settings:   make(map[string]interface{}),
	}
	err = configFacet.Store("boards", "test", invalidData)
	if err == nil {
		t.Errorf("Expected error for invalid configuration data")
	}

	// Test type mismatch
	mismatchData := ConfigurationData{
		Type:       "wrong-type",
		Identifier: "test",
		Version:    "1.0",
		Settings:   make(map[string]interface{}),
	}
	err = configFacet.Store("boards", "test", mismatchData)
	if err == nil {
		t.Errorf("Expected error for type mismatch")
	}

	// Test identifier mismatch
	idMismatchData := ConfigurationData{
		Type:       "boards",
		Identifier: "wrong-id",
		Version:    "1.0",
		Settings:   make(map[string]interface{}),
	}
	err = configFacet.Store("boards", "test", idMismatchData)
	if err == nil {
		t.Errorf("Expected error for identifier mismatch")
	}
}

// TestUnit_ConfigurationFacet_FileStructure tests that files are created in correct structure
func TestUnit_ConfigurationFacet_FileStructure(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize repository
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repository.Close()

	// Create configuration facet
	logger := utilities.NewLoggingUtility()
	configFacet := newConfigurationFacet(repository, logger)

	// Store test configuration
	configType := "workflows"
	identifier := "kanban"
	testData := ConfigurationData{
		Type:       configType,
		Identifier: identifier,
		Version:    "1.0",
		Settings:   map[string]interface{}{"rule": "value"},
		Schema:     "workflow-v1",
	}

	err = configFacet.Store(configType, identifier, testData)
	if err != nil {
		t.Fatalf("Store operation failed: %v", err)
	}

	// Verify file structure
	expectedPath := filepath.Join(tempDir, ".eisenkan", "config", configType, fmt.Sprintf("%s.json", identifier))
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected configuration file not created at %s", expectedPath)
	}

	// Verify directory structure
	configDir := filepath.Join(tempDir, ".eisenkan", "config", configType)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Expected configuration directory not created at %s", configDir)
	}
}