// Package resource_access provides ResourceAccess layer components implementing the iDesign methodology.
// This file implements the IConfiguration facet for BoardAccess to handle board configuration persistence.
package resource_access

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// IConfiguration defines the interface for configuration management operations
type IConfiguration interface {
	Load(configType string, identifier string) (ConfigurationData, error)
	Store(configType string, identifier string, data ConfigurationData) error

	// Board Configuration Operations (migrated from IBoardAccess)
	GetBoardConfiguration() (*BoardConfiguration, error)
	UpdateBoardConfiguration(config *BoardConfiguration) error
}

// ConfigurationData represents structured configuration information for board-level settings
type ConfigurationData struct {
	Type       string                 `json:"type"`
	Identifier string                 `json:"identifier"`
	Version    string                 `json:"version"`
	Settings   map[string]interface{} `json:"settings"`
	Schema     string                 `json:"schema"`
	Metadata   map[string]string      `json:"metadata"`
}

// configurationFacet implements the IConfiguration interface
type configurationFacet struct {
	versioningUtility utilities.ILoggingUtility
	repository        utilities.Repository
}

// newConfigurationFacet creates a new configuration facet instance
func newConfigurationFacet(repository utilities.Repository, logger utilities.ILoggingUtility) IConfiguration {
	return &configurationFacet{
		versioningUtility: logger,
		repository:        repository,
	}
}

// Load retrieves configuration data from git-based JSON storage
func (cf *configurationFacet) Load(configType string, identifier string) (ConfigurationData, error) {
	if configType == "" {
		return ConfigurationData{}, fmt.Errorf("configuration type cannot be empty")
	}
	if identifier == "" {
		return ConfigurationData{}, fmt.Errorf("configuration identifier cannot be empty")
	}

	// Construct file path for configuration data
	configFile := fmt.Sprintf("%s.json", identifier)
	configDir := filepath.Join(cf.repository.Path(), ".eisenkan", "config", configType)
	configPath := filepath.Join(configDir, configFile)

	// Read configuration file
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default configuration data if file doesn't exist
			return cf.createDefaultConfigurationData(configType, identifier), nil
		}
		return ConfigurationData{}, fmt.Errorf("failed to read configuration file %s: %w", configPath, err)
	}

	// Parse JSON content
	var configData ConfigurationData
	if err := json.Unmarshal(content, &configData); err != nil {
		return ConfigurationData{}, fmt.Errorf("failed to parse configuration data for type %s, identifier %s: %w", configType, identifier, err)
	}

	// Validate configuration data structure
	if err := cf.validateConfigurationData(configData); err != nil {
		return ConfigurationData{}, fmt.Errorf("invalid configuration data structure for type %s, identifier %s: %w", configType, identifier, err)
	}

	return configData, nil
}

// Store persists configuration data to git-based JSON storage with atomic operations and versioning
func (cf *configurationFacet) Store(configType string, identifier string, data ConfigurationData) error {
	if configType == "" {
		return fmt.Errorf("configuration type cannot be empty")
	}
	if identifier == "" {
		return fmt.Errorf("configuration identifier cannot be empty")
	}

	// Validate configuration data structure
	if err := cf.validateConfigurationData(data); err != nil {
		return fmt.Errorf("invalid configuration data structure: %w", err)
	}

	// Ensure configuration type and identifier match data
	if data.Type != configType {
		return fmt.Errorf("configuration type mismatch: expected %s, got %s", configType, data.Type)
	}
	if data.Identifier != identifier {
		return fmt.Errorf("configuration identifier mismatch: expected %s, got %s", identifier, data.Identifier)
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
		return fmt.Errorf("failed to serialize configuration data: %w", err)
	}

	// Construct file path
	configFile := fmt.Sprintf("%s.json", identifier)
	configDir := filepath.Join(cf.repository.Path(), ".eisenkan", "config", configType)
	configPath := filepath.Join(configDir, configFile)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %w", err)
	}

	// Write file with atomic operations
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	// Stage changes using relative path
	relativeConfigPath := filepath.Join(".eisenkan", "config", configType, configFile)
	if err := cf.repository.Stage([]string{relativeConfigPath}); err != nil {
		return fmt.Errorf("failed to stage configuration changes: %w", err)
	}

	// Commit changes
	commitMessage := fmt.Sprintf("Update %s configuration: %s", configType, identifier)
	if _, err := cf.repository.Commit(commitMessage); err != nil {
		return fmt.Errorf("failed to commit configuration changes: %w", err)
	}

	return nil
}

// validateConfigurationData validates the structure and content of configuration data
func (cf *configurationFacet) validateConfigurationData(data ConfigurationData) error {
	if data.Type == "" {
		return fmt.Errorf("configuration type is required")
	}

	if data.Identifier == "" {
		return fmt.Errorf("configuration identifier is required")
	}

	if data.Version == "" {
		return fmt.Errorf("configuration version is required")
	}

	if data.Settings == nil {
		return fmt.Errorf("configuration settings map cannot be nil")
	}

	// Validate settings can be serialized to JSON
	_, err := json.Marshal(data.Settings)
	if err != nil {
		return fmt.Errorf("configuration settings contain non-serializable content: %w", err)
	}

	return nil
}

// createDefaultConfigurationData creates default configuration data for a given type and identifier
func (cf *configurationFacet) createDefaultConfigurationData(configType string, identifier string) ConfigurationData {
	return ConfigurationData{
		Type:       configType,
		Identifier: identifier,
		Version:    "1.0",
		Settings:   make(map[string]interface{}),
		Schema:     "default",
		Metadata: map[string]string{
			"created":         time.Now().UTC().Format(time.RFC3339),
			"storage_version": "1.0",
			"default":         "true",
		},
	}
}

// GetBoardConfiguration retrieves the board configuration (migrated from IBoardAccess)
func (cf *configurationFacet) GetBoardConfiguration() (*BoardConfiguration, error) {
	// Load board configuration using the generic Load method
	configData, err := cf.Load("boards", "default")
	if err != nil {
		return nil, fmt.Errorf("failed to load board configuration: %w", err)
	}

	// Convert ConfigurationData to BoardConfiguration
	boardConfig := &BoardConfiguration{}

	// Handle default case (empty settings)
	if len(configData.Settings) == 0 {
		return &BoardConfiguration{
			Name:    "EisenKan Board",
			Columns: []string{"todo", "doing", "done"},
			Sections: map[string][]string{
				"todo": {"urgent-important", "urgent-not-important", "not-urgent-important"},
			},
			GitUser:  "BoardAccess",
			GitEmail: "boardaccess@eisenkan.local",
		}, nil
	}

	// Extract fields from settings
	if name, ok := configData.Settings["name"].(string); ok {
		boardConfig.Name = name
	}
	if columns, ok := configData.Settings["columns"].([]interface{}); ok {
		boardConfig.Columns = make([]string, len(columns))
		for i, col := range columns {
			if colStr, ok := col.(string); ok {
				boardConfig.Columns[i] = colStr
			}
		}
	}
	if sections, ok := configData.Settings["sections"].(map[string]interface{}); ok {
		boardConfig.Sections = make(map[string][]string)
		for col, secs := range sections {
			if secList, ok := secs.([]interface{}); ok {
				boardConfig.Sections[col] = make([]string, len(secList))
				for i, sec := range secList {
					if secStr, ok := sec.(string); ok {
						boardConfig.Sections[col][i] = secStr
					}
				}
			}
		}
	}
	if gitUser, ok := configData.Settings["git_user"].(string); ok {
		boardConfig.GitUser = gitUser
	}
	if gitEmail, ok := configData.Settings["git_email"].(string); ok {
		boardConfig.GitEmail = gitEmail
	}

	return boardConfig, nil
}

// UpdateBoardConfiguration updates the board configuration (migrated from IBoardAccess)
func (cf *configurationFacet) UpdateBoardConfiguration(config *BoardConfiguration) error {
	if config == nil {
		return fmt.Errorf("board configuration cannot be nil")
	}

	// Convert BoardConfiguration to ConfigurationData
	configData := ConfigurationData{
		Type:       "boards",
		Identifier: "default",
		Version:    "1.0",
		Schema:     "board-v1",
		Settings: map[string]interface{}{
			"name":      config.Name,
			"columns":   config.Columns,
			"sections":  config.Sections,
			"git_user":  config.GitUser,
			"git_email": config.GitEmail,
		},
	}

	// Store using the generic Store method
	return cf.Store("boards", "default", configData)
}