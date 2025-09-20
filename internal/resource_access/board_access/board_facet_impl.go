// Package board_access provides BoardAccess layer components implementing the iDesign methodology.
package board_access

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// BoardConfigurationValidator interface for validating board configurations
type BoardConfigurationValidator interface {
	EvaluateBoardConfigurationChange(ctx context.Context, event interface{}) (interface{}, error)
}

// boardFacet implements IBoard interface
type boardFacet struct {
	repository   utilities.Repository
	logger       utilities.ILoggingUtility
	mutex        *sync.RWMutex
	ruleEngine   BoardConfigurationValidator  // For board configuration validation
}

// newBoardFacet creates a new board facet implementation
func newBoardFacet(repository utilities.Repository, logger utilities.ILoggingUtility, mutex *sync.RWMutex, ruleEngine BoardConfigurationValidator) IBoard {
	return &boardFacet{
		repository: repository,
		logger:     logger,
		mutex:      mutex,
		ruleEngine: ruleEngine,
	}
}

// DiscoverBoards finds and validates board structures in a directory
func (bf *boardFacet) DiscoverBoards(ctx context.Context, directoryPath string) ([]BoardDiscoveryResult, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Discovering boards in directory: %s", directoryPath))

	// Validate directory exists and is accessible
	if directoryPath == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	stat, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", directoryPath)
		}
		return nil, fmt.Errorf("cannot access directory %s: %w", directoryPath, err)
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", directoryPath)
	}

	var results []BoardDiscoveryResult
	var allIssues []string

	// Walk through directory looking for potential board structures
	err = filepath.WalkDir(directoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Collect errors but continue processing
			allIssues = append(allIssues, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil // Continue walking
		}

		// Skip if not a directory
		if !d.IsDir() {
			return nil
		}

		// Check if this directory contains board indicators
		result := bf.evaluatePotentialBoard(path)
		if result != nil {
			results = append(results, *result)
		}

		return nil
	})

	if err != nil {
		// Report all issues together as requested in design decision
		if len(allIssues) > 0 {
			return results, fmt.Errorf("directory scan completed with issues: %s", strings.Join(allIssues, "; "))
		}
		return results, fmt.Errorf("failed to scan directory %s: %w", directoryPath, err)
	}

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Discovered %d potential boards in %s", len(results), directoryPath))
	return results, nil
}

// evaluatePotentialBoard checks if a directory looks like a board
func (bf *boardFacet) evaluatePotentialBoard(dirPath string) *BoardDiscoveryResult {
	result := &BoardDiscoveryResult{
		BoardPath: dirPath,
		Issues:    make([]string, 0),
		Metadata:  make(map[string]string),
	}

	// Check for git repository
	gitDir := filepath.Join(dirPath, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		result.HasGitRepo = true
	}

	// Check for board configuration file
	configPath := filepath.Join(dirPath, "board.json")
	if stat, err := os.Stat(configPath); err == nil && !stat.IsDir() {
		result.ConfigExists = true

		// Try to extract basic metadata
		if configData, err := os.ReadFile(configPath); err == nil {
			var config BoardConfiguration
			if json.Unmarshal(configData, &config) == nil {
				result.Title = config.Name
				result.Metadata["config_valid"] = "true"
			} else {
				result.Issues = append(result.Issues, "Invalid board.json format")
				result.Metadata["config_valid"] = "false"
			}
		}
	}

	// A valid board should have at least a git repo or config file
	result.IsValid = result.HasGitRepo || result.ConfigExists

	// Only return result if it looks like a board
	if result.IsValid {
		return result
	}

	return nil
}

// ExtractBoardMetadata retrieves comprehensive board information
func (bf *boardFacet) ExtractBoardMetadata(ctx context.Context, boardPath string) (*BoardMetadata, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Extracting metadata for board: %s", boardPath))

	if boardPath == "" {
		return nil, fmt.Errorf("board path cannot be empty")
	}

	// Validate board path exists
	if _, err := os.Stat(boardPath); err != nil {
		return nil, fmt.Errorf("board path does not exist: %s", boardPath)
	}

	metadata := &BoardMetadata{
		ColumnCounts: make(map[string]int),
		Metadata:     make(map[string]string),
	}

	// Extract configuration metadata
	configPath := filepath.Join(boardPath, "board.json")
	if configData, err := os.ReadFile(configPath); err == nil {
		var config BoardConfiguration
		if json.Unmarshal(configData, &config) == nil {
			metadata.Title = config.Name
			metadata.Configuration = &config
			metadata.Metadata["config_source"] = "board.json"
		}
	}

	// Get directory timestamps
	if stat, err := os.Stat(boardPath); err == nil {
		modTime := stat.ModTime()
		metadata.ModifiedAt = &modTime
	}

	// Count tasks by reading task files (simplified)
	taskFiles := []string{"active.json", "archived.json"}
	totalTasks := 0

	for _, taskFile := range taskFiles {
		taskPath := filepath.Join(boardPath, taskFile)
		if taskData, err := os.ReadFile(taskPath); err == nil {
			var tasks []TaskWithTimestamps
			if json.Unmarshal(taskData, &tasks) == nil {
				totalTasks += len(tasks)

				// Count tasks by column
				for _, task := range tasks {
					column := task.Status.Column
					metadata.ColumnCounts[column]++
				}
			}
		}
	}

	metadata.TaskCount = totalTasks
	metadata.SchemaVersion = "1.0" // Default version

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Extracted metadata for board %s: %d tasks", boardPath, totalTasks))
	return metadata, nil
}

// GetBoardStatistics calculates comprehensive board metrics
func (bf *boardFacet) GetBoardStatistics(ctx context.Context, boardPath string) (*BoardStatistics, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Calculating statistics for board: %s", boardPath))

	if boardPath == "" {
		return nil, fmt.Errorf("board path cannot be empty")
	}

	stats := &BoardStatistics{
		TasksByColumn:   make(map[string]int),
		TasksByPriority: make(map[string]int),
	}

	// Read active tasks
	activePath := filepath.Join(boardPath, "active.json")
	if taskData, err := os.ReadFile(activePath); err == nil {
		var tasks []TaskWithTimestamps
		if json.Unmarshal(taskData, &tasks) == nil {
			stats.ActiveTasks = len(tasks)
			stats.TotalTasks += len(tasks)

			var totalAge float64
			var oldestAge float64
			var lastActivity time.Time

			for _, task := range tasks {
				// Count by column
				stats.TasksByColumn[task.Status.Column]++

				// Count by priority
				priorityLabel := task.Priority.Label
				stats.TasksByPriority[priorityLabel]++

				// Calculate age
				age := time.Since(task.CreatedAt).Hours() / 24 // days
				totalAge += age
				if age > oldestAge {
					oldestAge = age
				}

				// Track last activity
				if task.UpdatedAt.After(lastActivity) {
					lastActivity = task.UpdatedAt
				}
			}

			if len(tasks) > 0 {
				stats.AverageTaskAge = totalAge / float64(len(tasks))
			}
			stats.OldestTaskAge = oldestAge

			if !lastActivity.IsZero() {
				stats.LastActivity = &lastActivity
			}
		}
	}

	// Read completed tasks
	archivedPath := filepath.Join(boardPath, "archived.json")
	if taskData, err := os.ReadFile(archivedPath); err == nil {
		var tasks []TaskWithTimestamps
		if json.Unmarshal(taskData, &tasks) == nil {
			stats.CompletedTasks = len(tasks)
			stats.TotalTasks += len(tasks)
		}
	}

	// Calculate health score (simplified metric)
	stats.BoardHealthScore = bf.calculateHealthScore(stats)

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Calculated statistics for board %s: %d total tasks", boardPath, stats.TotalTasks))
	return stats, nil
}

// calculateHealthScore computes a simple board health metric
func (bf *boardFacet) calculateHealthScore(stats *BoardStatistics) float64 {
	if stats.TotalTasks == 0 {
		return 1.0 // Empty board is "healthy"
	}

	// Simple health calculation based on task distribution and age
	score := 1.0

	// Penalize for very old tasks
	if stats.OldestTaskAge > 30 { // Tasks older than 30 days
		score -= 0.2
	}

	// Penalize for too many active tasks
	if stats.ActiveTasks > 50 {
		score -= 0.3
	}

	// Bonus for recent activity
	if stats.LastActivity != nil && time.Since(*stats.LastActivity).Hours() < 24 {
		score += 0.1
	}

	// Ensure score stays within bounds
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// ValidateBoardStructure verifies board integrity and structure
func (bf *boardFacet) ValidateBoardStructure(ctx context.Context, boardPath string) (*BoardValidationResult, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Validating board structure: %s", boardPath))

	if boardPath == "" {
		return nil, fmt.Errorf("board path cannot be empty")
	}

	result := &BoardValidationResult{
		IsValid: true,
		Issues:  make([]BoardValidationIssue, 0),
		Warnings: make([]BoardValidationIssue, 0),
	}

	// Validate directory exists
	if _, err := os.Stat(boardPath); err != nil {
		result.IsValid = false
		result.Issues = append(result.Issues, BoardValidationIssue{
			Severity:   "error",
			Component:  "structure",
			Message:    "Board directory does not exist",
			Details:    err.Error(),
			Suggestion: "Create the board directory",
		})
		return result, nil
	}

	// Validate git repository
	gitDir := filepath.Join(boardPath, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		result.GitRepoValid = true
	} else {
		result.Warnings = append(result.Warnings, BoardValidationIssue{
			Severity:   "warning",
			Component:  "git",
			Message:    "No git repository found",
			Suggestion: "Initialize git repository for version control",
		})
	}

	// Validate configuration file
	configPath := filepath.Join(boardPath, "board.json")
	if configData, err := os.ReadFile(configPath); err == nil {
		var config BoardConfiguration
		if json.Unmarshal(configData, &config) == nil {
			result.ConfigValid = true
			result.SchemaVersion = "1.0"

			// Validate configuration content using RuleEngine if available
			if bf.ruleEngine != nil {
				if err := bf.validateConfigWithRuleEngine(ctx, &config); err != nil {
					result.Issues = append(result.Issues, BoardValidationIssue{
						Severity:   "error",
						Component:  "config",
						Message:    "Configuration validation failed",
						Details:    err.Error(),
						Suggestion: "Fix configuration errors",
					})
					result.IsValid = false
				}
			}
		} else {
			result.ConfigValid = false
			result.Issues = append(result.Issues, BoardValidationIssue{
				Severity:   "error",
				Component:  "config",
				Message:    "Invalid board.json format",
				Details:    err.Error(),
				Suggestion: "Fix JSON syntax errors",
			})
			result.IsValid = false
		}
	} else {
		result.ConfigValid = false
		result.Warnings = append(result.Warnings, BoardValidationIssue{
			Severity:   "warning",
			Component:  "config",
			Message:    "No board.json configuration found",
			Suggestion: "Create board.json configuration file",
		})
	}

	// Validate data files integrity
	result.DataIntegrity = bf.validateDataFiles(boardPath, result)

	if !result.DataIntegrity {
		result.IsValid = false
	}

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Validation complete for board %s: valid=%t", boardPath, result.IsValid))
	return result, nil
}

// validateConfigWithRuleEngine uses RuleEngine to validate board configuration
func (bf *boardFacet) validateConfigWithRuleEngine(ctx context.Context, config *BoardConfiguration) error {
	// Create a simple event structure for validation
	event := map[string]interface{}{
		"event_type": "board_validate",
		"config": map[string]interface{}{
			"title":       config.Name,
			"columns":     config.Columns,
			"git_user":    config.GitUser,
			"git_email":   config.GitEmail,
			"timestamp":   time.Now(),
		},
	}

	// Validate using RuleEngine
	result, err := bf.ruleEngine.EvaluateBoardConfigurationChange(ctx, event)
	if err != nil {
		return fmt.Errorf("RuleEngine validation failed: %w", err)
	}

	// Check if validation passed (basic check for allowed field)
	if resultMap, ok := result.(map[string]interface{}); ok {
		if allowed, exists := resultMap["allowed"]; exists {
			if allowedBool, ok := allowed.(bool); ok && !allowedBool {
				return fmt.Errorf("board configuration validation failed")
			}
		}
	}

	return nil
}

// validateDataFiles checks integrity of task data files
func (bf *boardFacet) validateDataFiles(boardPath string, result *BoardValidationResult) bool {
	dataIntegrity := true
	taskFiles := []string{"active.json", "archived.json"}

	for _, taskFile := range taskFiles {
		taskPath := filepath.Join(boardPath, taskFile)
		if taskData, err := os.ReadFile(taskPath); err == nil {
			var tasks []TaskWithTimestamps
			if json.Unmarshal(taskData, &tasks) != nil {
				result.Issues = append(result.Issues, BoardValidationIssue{
					Severity:   "error",
					Component:  "data",
					Message:    fmt.Sprintf("Invalid JSON in %s", taskFile),
					Suggestion: "Fix JSON syntax errors in task data",
				})
				dataIntegrity = false
			}
		}
		// Missing task files are not errors (board might be empty)
	}

	return dataIntegrity
}

// LoadBoardConfiguration loads board configuration data
func (bf *boardFacet) LoadBoardConfiguration(ctx context.Context, boardPath string, configType string) (map[string]interface{}, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Loading configuration for board: %s, type: %s", boardPath, configType))

	if boardPath == "" {
		return nil, fmt.Errorf("board path cannot be empty")
	}

	configPath := filepath.Join(boardPath, "board.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default configuration
			return bf.getDefaultConfiguration(configType), nil
		}
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("invalid configuration format: %w", err)
	}

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Loaded configuration for board %s", boardPath))
	return config, nil
}

// StoreBoardConfiguration stores board configuration data
func (bf *boardFacet) StoreBoardConfiguration(ctx context.Context, boardPath string, configType string, configData map[string]interface{}) error {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Storing configuration for board: %s, type: %s", boardPath, configType))

	if boardPath == "" {
		return fmt.Errorf("board path cannot be empty")
	}

	// Validate configuration using RuleEngine if available
	if bf.ruleEngine != nil {
		if err := bf.validateConfigDataWithRuleEngine(ctx, configData); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	// Serialize configuration to JSON
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Write to file
	configPath := filepath.Join(boardPath, "board.json")
	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	// Commit to git if repository exists
	bf.commitConfigurationChange(boardPath, configType)

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Stored configuration for board %s", boardPath))
	return nil
}

// validateConfigDataWithRuleEngine validates configuration data using RuleEngine
func (bf *boardFacet) validateConfigDataWithRuleEngine(ctx context.Context, configData map[string]interface{}) error {
	// Convert generic config data to BoardConfiguration for validation
	jsonData, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	var boardConfig BoardConfiguration
	if err := json.Unmarshal(jsonData, &boardConfig); err != nil {
		return fmt.Errorf("invalid board configuration format: %w", err)
	}

	return bf.validateConfigWithRuleEngine(ctx, &boardConfig)
}

// commitConfigurationChange commits configuration changes to git
func (bf *boardFacet) commitConfigurationChange(boardPath string, configType string) {
	// Best effort git commit - don't fail if git operations fail
	message := fmt.Sprintf("Update board configuration: %s", configType)
	if _, err := bf.repository.Commit(message); err != nil {
		bf.logger.LogMessage(utilities.Warning, "BoardFacet", fmt.Sprintf("Failed to commit configuration changes: %v", err))
	}
}

// getDefaultConfiguration returns default configuration for a given type
func (bf *boardFacet) getDefaultConfiguration(configType string) map[string]interface{} {
	defaultConfig := map[string]interface{}{
		"name":    "EisenKan Board",
		"columns": []string{"todo", "doing", "done"},
		"sections": map[string][]string{
			"todo": {"urgent-important", "urgent-not-important", "not-urgent-important"},
		},
		"git_user":  "BoardAccess",
		"git_email": "boardaccess@eisenkan.local",
	}

	return defaultConfig
}

// CreateBoard initializes a new board structure
func (bf *boardFacet) CreateBoard(ctx context.Context, request *BoardCreationRequest) (*BoardCreationResult, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Creating board: %s", request.BoardPath))

	if request == nil {
		return nil, fmt.Errorf("creation request cannot be nil")
	}

	if request.BoardPath == "" {
		return nil, fmt.Errorf("board path cannot be empty")
	}

	if request.Title == "" {
		return nil, fmt.Errorf("board title cannot be empty")
	}

	result := &BoardCreationResult{
		BoardPath: request.BoardPath,
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(request.BoardPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create board directory: %w", err)
	}

	// Create board configuration
	config := request.Configuration
	if config == nil {
		config = &BoardConfiguration{
			Name:    request.Title,
			Columns: []string{"todo", "doing", "done"},
			Sections: map[string][]string{
				"todo": {"urgent-important", "urgent-not-important", "not-urgent-important"},
			},
			GitUser:  "BoardAccess",
			GitEmail: "boardaccess@eisenkan.local",
		}
	} else {
		config.Name = request.Title // Ensure title matches
	}

	// Validate configuration using RuleEngine
	if bf.ruleEngine != nil {
		if err := bf.validateConfigWithRuleEngine(ctx, config); err != nil {
			return nil, fmt.Errorf("board configuration validation failed: %w", err)
		}
	}

	// Write configuration file
	configPath := filepath.Join(request.BoardPath, "board.json")
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize board configuration: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write board configuration: %w", err)
	}

	result.ConfigPath = configPath

	// Initialize empty task files
	emptyTasks := "[]"
	for _, taskFile := range []string{"active.json", "archived.json"} {
		taskPath := filepath.Join(request.BoardPath, taskFile)
		if err := os.WriteFile(taskPath, []byte(emptyTasks), 0644); err != nil {
			return nil, fmt.Errorf("failed to create task file %s: %w", taskFile, err)
		}
	}

	// Initialize git repository if requested
	if request.InitializeGit {
		gitConfig := &utilities.AuthorConfiguration{
			User:  config.GitUser,
			Email: config.GitEmail,
		}

		if _, err := utilities.InitializeRepositoryWithConfig(request.BoardPath, gitConfig); err != nil {
			bf.logger.LogMessage(utilities.Warning, "BoardFacet", fmt.Sprintf("Failed to initialize git repository: %v", err))
		} else {
			result.GitInitialized = true
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("Board '%s' created successfully", request.Title)

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Created board %s at %s", request.Title, request.BoardPath))
	return result, nil
}

// DeleteBoard removes a board structure
func (bf *boardFacet) DeleteBoard(ctx context.Context, request *BoardDeletionRequest) (*BoardDeletionResult, error) {
	bf.logger.LogMessage(utilities.Debug, "BoardFacet", fmt.Sprintf("Deleting board: %s", request.BoardPath))

	if request == nil {
		return nil, fmt.Errorf("deletion request cannot be nil")
	}

	if request.BoardPath == "" {
		return nil, fmt.Errorf("board path cannot be empty")
	}

	result := &BoardDeletionResult{}

	// Validate board exists
	if _, err := os.Stat(request.BoardPath); err != nil {
		if os.IsNotExist(err) {
			// Idempotent operation - treat non-existent as success
			result.Success = true
			result.Method = "none"
			result.Message = "Board already deleted or does not exist"
			return result, nil
		}
		return nil, fmt.Errorf("cannot access board path: %w", err)
	}

	// Create backup if requested
	if request.CreateBackup {
		backupPath := request.BackupLocation
		if backupPath == "" {
			backupPath = request.BoardPath + ".backup." + time.Now().Format("20060102-150405")
		}

		if err := bf.copyDirectory(request.BoardPath, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}

		result.BackupCreated = true
		result.BackupLocation = backupPath
	}

	// Attempt to use OS trash if requested and available
	if request.UseTrash && bf.canUseOSTrash() {
		if err := bf.moveToTrash(request.BoardPath); err != nil {
			// Fall back to permanent deletion
			bf.logger.LogMessage(utilities.Warning, "BoardFacet", fmt.Sprintf("Failed to move to trash, using permanent deletion: %v", err))
		} else {
			result.Success = true
			result.Method = "trash"
			result.Message = "Board moved to trash"
			return result, nil
		}
	}

	// Permanent deletion
	if err := os.RemoveAll(request.BoardPath); err != nil {
		return nil, fmt.Errorf("failed to delete board: %w", err)
	}

	result.Success = true
	result.Method = "permanent"
	result.Message = "Board permanently deleted"

	bf.logger.LogMessage(utilities.Info, "BoardFacet", fmt.Sprintf("Deleted board at %s", request.BoardPath))
	return result, nil
}

// copyDirectory recursively copies a directory
func (bf *boardFacet) copyDirectory(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = dstFile.ReadFrom(srcFile)
		return err
	})
}

// canUseOSTrash checks if OS trash functionality is available
func (bf *boardFacet) canUseOSTrash() bool {
	// Simplified check - would need OS-specific implementation
	// For now, assume trash is available on common desktop environments
	return true
}

// moveToTrash moves a directory to OS trash (simplified implementation)
func (bf *boardFacet) moveToTrash(path string) error {
	// This is a simplified implementation
	// Real implementation would use OS-specific trash APIs
	return fmt.Errorf("OS trash functionality not implemented")
}