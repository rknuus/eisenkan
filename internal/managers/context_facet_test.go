package managers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// TestUnit_ContextFacet_LoadStore tests basic Load and Store operations
func TestUnit_ContextFacet_LoadStore(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_context_test_*")
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

	// Create context facet
	contextFacet := newContextFacet(repository)

	// Test data
	contextType := "ui-state"
	testData := ContextData{
		Type:    contextType,
		Version: "1.0",
		Data: map[string]interface{}{
			"window_width":  1024,
			"window_height": 768,
			"theme":         "dark",
		},
		Metadata: map[string]string{
			"created_by": "test",
		},
	}

	// Test Store operation
	err = contextFacet.Store(contextType, testData)
	if err != nil {
		t.Fatalf("Store operation failed: %v", err)
	}

	// Test Load operation
	loadedData, err := contextFacet.Load(contextType)
	if err != nil {
		t.Fatalf("Load operation failed: %v", err)
	}

	// Verify loaded data matches stored data
	if loadedData.Type != testData.Type {
		t.Errorf("Type mismatch: expected %s, got %s", testData.Type, loadedData.Type)
	}
	if loadedData.Version != testData.Version {
		t.Errorf("Version mismatch: expected %s, got %s", testData.Version, loadedData.Version)
	}

	// Verify data content (JSON unmarshaling converts numbers to float64)
	if loadedData.Data["window_width"] != float64(1024) {
		t.Errorf("Data content mismatch for window_width: expected 1024, got %v", loadedData.Data["window_width"])
	}
	if loadedData.Data["theme"] != testData.Data["theme"] {
		t.Errorf("Data content mismatch for theme")
	}

	// Verify metadata was updated during store
	if loadedData.Metadata["last_updated"] == "" {
		t.Errorf("Expected last_updated metadata to be set")
	}
	if loadedData.Metadata["storage_version"] != "1.0" {
		t.Errorf("Expected storage_version to be 1.0")
	}
}

// TestUnit_ContextFacet_LoadDefault tests loading non-existent context returns default
func TestUnit_ContextFacet_LoadDefault(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_context_test_*")
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

	// Create context facet
	contextFacet := newContextFacet(repository)

	// Test loading non-existent context
	contextType := "non-existent"
	defaultData, err := contextFacet.Load(contextType)
	if err != nil {
		t.Fatalf("Load operation failed: %v", err)
	}

	// Verify default data structure
	if defaultData.Type != contextType {
		t.Errorf("Expected type %s, got %s", contextType, defaultData.Type)
	}
	if defaultData.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", defaultData.Version)
	}
	if defaultData.Data == nil {
		t.Errorf("Expected data map to be initialized")
	}
	if defaultData.Metadata["default"] != "true" {
		t.Errorf("Expected default metadata to be true")
	}
}

// TestUnit_ContextFacet_ValidationErrors tests validation error conditions
func TestUnit_ContextFacet_ValidationErrors(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_context_test_*")
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

	// Create context facet
	contextFacet := newContextFacet(repository)

	// Test empty context type
	_, err = contextFacet.Load("")
	if err == nil {
		t.Errorf("Expected error for empty context type")
	}

	// Test invalid context data - missing type
	invalidData := ContextData{
		Version: "1.0",
		Data:    make(map[string]interface{}),
	}
	err = contextFacet.Store("test", invalidData)
	if err == nil {
		t.Errorf("Expected error for invalid context data")
	}

	// Test type mismatch
	mismatchData := ContextData{
		Type:    "wrong-type",
		Version: "1.0",
		Data:    make(map[string]interface{}),
	}
	err = contextFacet.Store("test", mismatchData)
	if err == nil {
		t.Errorf("Expected error for type mismatch")
	}
}

// TestUnit_ContextFacet_FileStructure tests that files are created in correct structure
func TestUnit_ContextFacet_FileStructure(t *testing.T) {
	// Create temporary directory for test repository
	tempDir, err := os.MkdirTemp("", "eisenkan_context_test_*")
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

	// Create context facet
	contextFacet := newContextFacet(repository)

	// Store test context
	contextType := "preferences"
	testData := ContextData{
		Type:    contextType,
		Version: "1.0",
		Data:    map[string]interface{}{"setting": "value"},
	}

	err = contextFacet.Store(contextType, testData)
	if err != nil {
		t.Fatalf("Store operation failed: %v", err)
	}

	// Verify file structure
	expectedPath := filepath.Join(tempDir, ".eisenkan", "context", fmt.Sprintf("%s.json", contextType))
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected context file not created at %s", expectedPath)
	}

	// Verify directory structure
	contextDir := filepath.Join(tempDir, ".eisenkan", "context")
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		t.Errorf("Expected context directory not created at %s", contextDir)
	}
}