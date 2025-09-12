// Package resource_access provides ResourceAccess layer components implementing the iDesign methodology.
// This package contains components that provide data access and persistence services
// to higher-level components in the application architecture.
package resource_access

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rknuus/eisenkan/internal/utilities"
)

const (
	// rulesFileName defines the standard filename for rule sets
	rulesFileName = "rules.json"
)

// Rule represents a single business rule in the rule set
type Rule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`     // validation, workflow, automation, notification
	TriggerType string                 `json:"trigger_type"` // task_transition, status_change, due_date, etc.
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     map[string]interface{} `json:"actions"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// RuleSet represents the complete collection of rules for a board directory
type RuleSet struct {
	Version      string              `json:"version"`
	Rules        []Rule              `json:"rules"`
	Dependencies map[string][]string `json:"dependencies,omitempty"` // rule_id -> [dependent_rule_ids]
	Metadata     map[string]string   `json:"metadata,omitempty"`
}

// ValidationResult contains validation status and error details
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings,omitempty"`
}

// IRulesAccess defines the interface for rule data access operations
type IRulesAccess interface {
	// ReadRules retrieves the complete rule set for a board directory
	ReadRules(boardDirPath string) (*RuleSet, error)

	// ValidateRuleChanges validates a rule set without storing it
	ValidateRuleChanges(ruleSet *RuleSet) (*ValidationResult, error)

	// ChangeRules validates and stores a complete rule set
	ChangeRules(boardDirPath string, ruleSet *RuleSet) error

	// Close releases any resources held by the service
	Close() error
}

// RulesAccess implements IRulesAccess interface
type RulesAccess struct {
	repository utilities.Repository
	logger     utilities.ILoggingUtility
	mutex      sync.RWMutex
}

// NewRulesAccess creates a new RulesAccess instance
func NewRulesAccess(repositoryPath string) (*RulesAccess, error) {
	logger := utilities.NewLoggingUtility()

	logger.LogMessage(utilities.Debug, "RulesAccess", "Initializing RulesAccess")

	// Load board configuration to get git settings
	boardConfigPath := filepath.Join(repositoryPath, "board.json")
	var gitConfig *utilities.AuthorConfiguration

	if configData, err := os.ReadFile(boardConfigPath); err == nil {
		// Try to parse board configuration
		var boardConfig struct {
			GitUser  string `json:"git_user"`
			GitEmail string `json:"git_email"`
		}
		if json.Unmarshal(configData, &boardConfig) == nil && boardConfig.GitUser != "" && boardConfig.GitEmail != "" {
			gitConfig = &utilities.AuthorConfiguration{
				User:  boardConfig.GitUser,
				Email: boardConfig.GitEmail,
			}
		}
	}

	// Fall back to default if no config found
	if gitConfig == nil {
		gitConfig = &utilities.AuthorConfiguration{
			User:  "Eisen Kan",
			Email: "eisenkan@board.local",
		}
		logger.LogMessage(utilities.Warning, "RulesAccess", "Using default git configuration - board.json not found or incomplete")
	}

	// Initialize repository with git configuration
	repository, err := utilities.InitializeRepositoryWithConfig(repositoryPath, gitConfig)
	if err != nil {
		return nil, fmt.Errorf("RulesAccess.NewRulesAccess failed to initialize repository: %w", err)
	}

	ra := &RulesAccess{
		repository: repository,
		logger:     logger,
	}

	ra.logger.LogMessage(utilities.Info, "RulesAccess", "RulesAccess initialized successfully")
	return ra, nil
}

// ReadRules retrieves the complete rule set for a board directory
func (ra *RulesAccess) ReadRules(boardDirPath string) (*RuleSet, error) {
	ra.mutex.RLock()
	defer ra.mutex.RUnlock()

	rulesFilePath := filepath.Join(boardDirPath, rulesFileName)

	// Check if rules file exists
	if _, err := os.Stat(rulesFilePath); os.IsNotExist(err) {
		// Return empty rule set if no rules are configured
		ra.logger.LogMessage(utilities.Info, "RulesAccess", fmt.Sprintf("No rules file found, returning empty rule set for %s", boardDirPath))
		return &RuleSet{
			Version: "1.0",
			Rules:   []Rule{},
		}, nil
	}

	// Read rules file
	data, err := os.ReadFile(rulesFilePath)
	if err != nil {
		return nil, fmt.Errorf("RulesAccess.ReadRules failed to read rules file %s: %w", rulesFilePath, err)
	}

	// Parse JSON
	var ruleSet RuleSet
	if err := json.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("RulesAccess.ReadRules failed to parse rules JSON from %s: %w", rulesFilePath, err)
	}

	ra.logger.LogMessage(utilities.Info, "RulesAccess", fmt.Sprintf("Retrieved %d rules from %s", len(ruleSet.Rules), boardDirPath))
	return &ruleSet, nil
}

// ValidateRuleChanges validates a rule set without storing it
func (ra *RulesAccess) ValidateRuleChanges(ruleSet *RuleSet) (*ValidationResult, error) {
	if ruleSet == nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []string{"rule set cannot be nil"},
		}, nil
	}

	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate basic structure
	if ruleSet.Version == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "rule set version is required")
	}

	// Validate individual rules
	ruleIDs := make(map[string]bool)
	for i, rule := range ruleSet.Rules {
		// Check for duplicate rule IDs
		if rule.ID == "" {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("rule at index %d missing ID", i))
			continue
		}

		if ruleIDs[rule.ID] {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate rule ID: %s", rule.ID))
		}
		ruleIDs[rule.ID] = true

		// Validate required fields
		if rule.Name == "" {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("rule %s missing name", rule.ID))
		}

		// Validate category
		validCategories := map[string]bool{
			"validation":   true,
			"workflow":     true,
			"automation":   true,
			"notification": true,
		}
		if !validCategories[rule.Category] {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("rule %s has invalid category: %s", rule.ID, rule.Category))
		}

		// Validate trigger type
		if rule.TriggerType == "" {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("rule %s missing trigger type", rule.ID))
		}

		// Validate conditions and actions exist
		if len(rule.Conditions) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("rule %s missing conditions", rule.ID))
		}

		if len(rule.Actions) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("rule %s missing actions", rule.ID))
		}
	}

	// Validate dependencies
	if err := ra.validateDependencies(ruleSet, ruleIDs, result); err != nil {
		return nil, fmt.Errorf("RulesAccess.ValidateRuleChanges dependency validation failed: %w", err)
	}

	if result.Valid {
		ra.logger.LogMessage(utilities.Debug, "RulesAccess", fmt.Sprintf("Rule set validation passed with %d rules", len(ruleSet.Rules)))
	} else {
		ra.logger.LogMessage(utilities.Warning, "RulesAccess", fmt.Sprintf("Rule set validation failed with %d errors", len(result.Errors)))
	}

	return result, nil
}

// validateDependencies checks for circular dependencies and invalid references
func (ra *RulesAccess) validateDependencies(ruleSet *RuleSet, ruleIDs map[string]bool, result *ValidationResult) error {
	if ruleSet.Dependencies == nil {
		return nil
	}

	// Check that all dependency references exist
	for ruleID, deps := range ruleSet.Dependencies {
		if !ruleIDs[ruleID] {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("dependency reference to non-existent rule: %s", ruleID))
			continue
		}

		for _, depID := range deps {
			if !ruleIDs[depID] {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("rule %s depends on non-existent rule: %s", ruleID, depID))
			}
		}
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for ruleID := range ruleIDs {
		if !visited[ruleID] {
			if ra.hasCycle(ruleID, ruleSet.Dependencies, visited, recStack) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("circular dependency detected involving rule: %s", ruleID))
			}
		}
	}

	return nil
}

// hasCycle detects circular dependencies using DFS
func (ra *RulesAccess) hasCycle(ruleID string, dependencies map[string][]string, visited, recStack map[string]bool) bool {
	visited[ruleID] = true
	recStack[ruleID] = true

	// Check all dependencies of current rule
	if deps, exists := dependencies[ruleID]; exists {
		for _, depID := range deps {
			if !visited[depID] {
				if ra.hasCycle(depID, dependencies, visited, recStack) {
					return true
				}
			} else if recStack[depID] {
				return true
			}
		}
	}

	recStack[ruleID] = false
	return false
}

// ChangeRules validates and stores a complete rule set
func (ra *RulesAccess) ChangeRules(boardDirPath string, ruleSet *RuleSet) error {
	ra.mutex.Lock()
	defer ra.mutex.Unlock()

	// Validate rule set first
	validation, err := ra.ValidateRuleChanges(ruleSet)
	if err != nil {
		return fmt.Errorf("RulesAccess.ChangeRules validation failed: %w", err)
	}

	if !validation.Valid {
		return fmt.Errorf("RulesAccess.ChangeRules rule set validation failed: %v", validation.Errors)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(ruleSet, "", "  ")
	if err != nil {
		return fmt.Errorf("RulesAccess.ChangeRules failed to marshal rule set to JSON: %w", err)
	}

	// Write to rules file
	rulesFilePath := filepath.Join(boardDirPath, rulesFileName)
	if err := os.WriteFile(rulesFilePath, data, 0644); err != nil {
		return fmt.Errorf("RulesAccess.ChangeRules failed to write rules file %s: %w", rulesFilePath, err)
	}

	// Stage and commit changes via Repository
	if err := ra.repository.Stage([]string{rulesFileName}); err != nil {
		return fmt.Errorf("RulesAccess.ChangeRules failed to stage changes: %w", err)
	}

	commitMessage := fmt.Sprintf("Update rule set with %d rules", len(ruleSet.Rules))
	if _, err := ra.repository.Commit(commitMessage); err != nil {
		return fmt.Errorf("RulesAccess.ChangeRules failed to commit changes: %w", err)
	}

	ra.logger.LogMessage(utilities.Info, "RulesAccess", fmt.Sprintf("Rule set updated successfully with %d rules in %s", len(ruleSet.Rules), boardDirPath))
	return nil
}

// Close releases any resources held by the service
func (ra *RulesAccess) Close() error {
	ra.mutex.Lock()
	defer ra.mutex.Unlock()

	var errors []error

	if ra.repository != nil {
		if err := ra.repository.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Repository: %w", err))
		}
	}

	if ra.logger != nil {
		ra.logger.LogMessage(utilities.Info, "RulesAccess", "Closing RulesAccess")
	}

	if len(errors) > 0 {
		return fmt.Errorf("RulesAccess.Close encountered errors: %v", errors)
	}

	return nil
}
