// Package managers provides Manager layer components implementing the iDesign methodology.
// This file implements the IContext facet for TaskManager to handle UI context persistence.
package task_manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// IContext defines the interface for context management operations
type IContext interface {
	Load(contextType string) (ContextData, error)
	Store(contextType string, data ContextData) error
}

// ContextData represents structured context information for UI state management
type ContextData struct {
	Type     string                 `json:"type"`
	Version  string                 `json:"version"`
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]string      `json:"metadata"`
}

// contextFacet implements the IContext interface
type contextFacet struct {
	repository utilities.Repository
}

// newContextFacet creates a new context facet instance
func newContextFacet(repository utilities.Repository) IContext {
	return &contextFacet{
		repository: repository,
	}
}

// Load retrieves context data from git-based JSON storage
func (cf *contextFacet) Load(contextType string) (ContextData, error) {
	if contextType == "" {
		return ContextData{}, fmt.Errorf("context type cannot be empty")
	}

	// Construct file path for context data
	contextFile := fmt.Sprintf("%s.json", contextType)
	contextDir := filepath.Join(cf.repository.Path(), ".eisenkan", "context")
	contextPath := filepath.Join(contextDir, contextFile)

	// Read context file
	content, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default context data if file doesn't exist
			return cf.createDefaultContextData(contextType), nil
		}
		return ContextData{}, fmt.Errorf("failed to read context file %s: %w", contextPath, err)
	}

	// Parse JSON content
	var contextData ContextData
	if err := json.Unmarshal(content, &contextData); err != nil {
		return ContextData{}, fmt.Errorf("failed to parse context data for type %s: %w", contextType, err)
	}

	// Validate context data structure
	if err := cf.validateContextData(contextData); err != nil {
		return ContextData{}, fmt.Errorf("invalid context data structure for type %s: %w", contextType, err)
	}

	return contextData, nil
}

// Store persists context data to git-based JSON storage with atomic operations
func (cf *contextFacet) Store(contextType string, data ContextData) error {
	if contextType == "" {
		return fmt.Errorf("context type cannot be empty")
	}

	// Validate context data structure
	if err := cf.validateContextData(data); err != nil {
		return fmt.Errorf("invalid context data structure: %w", err)
	}

	// Ensure context type matches data type
	if data.Type != contextType {
		return fmt.Errorf("context type mismatch: expected %s, got %s", contextType, data.Type)
	}

	// Update metadata
	if data.Metadata == nil {
		data.Metadata = make(map[string]string)
	}
	data.Metadata["last_updated"] = time.Now().UTC().Format(time.RFC3339)
	data.Metadata["storage_version"] = "1.0"

	// Serialize to JSON
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize context data: %w", err)
	}

	// Construct file path
	contextFile := fmt.Sprintf("%s.json", contextType)
	contextDir := filepath.Join(cf.repository.Path(), ".eisenkan", "context")
	contextPath := filepath.Join(contextDir, contextFile)

	// Ensure directory exists
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	// Write file with atomic operations
	if err := os.WriteFile(contextPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	// Stage changes using relative path
	relativeContextPath := filepath.Join(".eisenkan", "context", contextFile)
	if err := cf.repository.Stage([]string{relativeContextPath}); err != nil {
		return fmt.Errorf("failed to stage context changes: %w", err)
	}

	// Commit changes
	commitMessage := fmt.Sprintf("Update %s context data", contextType)
	if _, err := cf.repository.Commit(commitMessage); err != nil {
		return fmt.Errorf("failed to commit context changes: %w", err)
	}

	return nil
}

// validateContextData validates the structure and content of context data
func (cf *contextFacet) validateContextData(data ContextData) error {
	if data.Type == "" {
		return fmt.Errorf("context type is required")
	}

	if data.Version == "" {
		return fmt.Errorf("context version is required")
	}

	if data.Data == nil {
		return fmt.Errorf("context data map cannot be nil")
	}

	// Validate data can be serialized to JSON
	_, err := json.Marshal(data.Data)
	if err != nil {
		return fmt.Errorf("context data contains non-serializable content: %w", err)
	}

	return nil
}

// createDefaultContextData creates default context data for a given type
func (cf *contextFacet) createDefaultContextData(contextType string) ContextData {
	return ContextData{
		Type:    contextType,
		Version: "1.0",
		Data:    make(map[string]interface{}),
		Metadata: map[string]string{
			"created":         time.Now().UTC().Format(time.RFC3339),
			"storage_version": "1.0",
			"default":         "true",
		},
	}
}