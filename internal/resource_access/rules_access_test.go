package resource_access

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnit_RulesAccess_NewRulesAccess(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rulesaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating new RulesAccess
	ra, err := NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer ra.Close()

	// Verify it implements the interface
	var _ IRulesAccess = ra
}

func TestUnit_RulesAccess_ReadRulesEmptyDirectory(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rulesaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create RulesAccess
	ra, err := NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer ra.Close()

	// Read rules from empty directory
	ruleSet, err := ra.ReadRules(tempDir)
	if err != nil {
		t.Fatalf("Failed to read rules from empty directory: %v", err)
	}

	// Should return empty rule set
	if ruleSet == nil {
		t.Fatal("Rule set should not be nil")
	}

	if len(ruleSet.Rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(ruleSet.Rules))
	}

	if ruleSet.Version != "1.0" {
		t.Errorf("Expected version '1.0', got %s", ruleSet.Version)
	}
}

func TestUnit_RulesAccess_ValidateAndChangeRules(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rulesaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create RulesAccess
	ra, err := NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer ra.Close()

	// Create test rule set
	ruleSet := &RuleSet{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:          "test-rule-001",
				Name:        "Test Workflow Rule",
				Category:    "workflow",
				TriggerType: "task_transition",
				Conditions: map[string]interface{}{
					"from_status": "todo",
					"to_status":   "doing",
				},
				Actions: map[string]interface{}{
					"log_message": "Task moved to doing",
				},
				Priority: 1,
				Enabled:  true,
			},
		},
	}

	// Validate rule set
	validation, err := ra.ValidateRuleChanges(ruleSet)
	if err != nil {
		t.Fatalf("Failed to validate rule set: %v", err)
	}

	if !validation.Valid {
		t.Fatalf("Rule set validation failed: %v", validation.Errors)
	}

	// Store rule set
	err = ra.ChangeRules(tempDir, ruleSet)
	if err != nil {
		t.Fatalf("Failed to store rule set: %v", err)
	}

	// Verify rules file was created
	rulesFilePath := filepath.Join(tempDir, "rules.json")
	if _, err := os.Stat(rulesFilePath); os.IsNotExist(err) {
		t.Errorf("Rules file was not created at %s", rulesFilePath)
	}

	// Read back the rules
	retrievedRuleSet, err := ra.ReadRules(tempDir)
	if err != nil {
		t.Fatalf("Failed to read stored rules: %v", err)
	}

	// Verify rule data
	if len(retrievedRuleSet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(retrievedRuleSet.Rules))
	}

	retrievedRule := retrievedRuleSet.Rules[0]
	if retrievedRule.ID != "test-rule-001" {
		t.Errorf("Expected rule ID 'test-rule-001', got %s", retrievedRule.ID)
	}

	if retrievedRule.Name != "Test Workflow Rule" {
		t.Errorf("Expected rule name 'Test Workflow Rule', got %s", retrievedRule.Name)
	}

	if retrievedRule.Category != "workflow" {
		t.Errorf("Expected category 'workflow', got %s", retrievedRule.Category)
	}
}

func TestUnit_RulesAccess_InvalidRules(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rulesaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create RulesAccess
	ra, err := NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer ra.Close()

	t.Run("NilRuleSet", func(t *testing.T) {
		validation, err := ra.ValidateRuleChanges(nil)
		if err != nil {
			t.Fatalf("Validation should not error on nil rule set: %v", err)
		}
		if validation.Valid {
			t.Error("Nil rule set should be invalid")
		}
		if len(validation.Errors) == 0 {
			t.Error("Should have validation errors for nil rule set")
		}
	})

	t.Run("MissingVersion", func(t *testing.T) {
		ruleSet := &RuleSet{
			Rules: []Rule{},
		}
		validation, err := ra.ValidateRuleChanges(ruleSet)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if validation.Valid {
			t.Error("Rule set without version should be invalid")
		}
	})

	t.Run("DuplicateRuleIDs", func(t *testing.T) {
		ruleSet := &RuleSet{
			Version: "1.0",
			Rules: []Rule{
				{
					ID:          "duplicate-id",
					Name:        "Rule 1",
					Category:    "workflow",
					TriggerType: "task_transition",
					Conditions:  map[string]interface{}{"test": "value"},
					Actions:     map[string]interface{}{"test": "action"},
				},
				{
					ID:          "duplicate-id",
					Name:        "Rule 2",
					Category:    "workflow",
					TriggerType: "task_transition",
					Conditions:  map[string]interface{}{"test": "value"},
					Actions:     map[string]interface{}{"test": "action"},
				},
			},
		}
		validation, err := ra.ValidateRuleChanges(ruleSet)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if validation.Valid {
			t.Error("Rule set with duplicate IDs should be invalid")
		}
	})

	t.Run("InvalidCategory", func(t *testing.T) {
		ruleSet := &RuleSet{
			Version: "1.0",
			Rules: []Rule{
				{
					ID:          "test-rule",
					Name:        "Test Rule",
					Category:    "invalid-category",
					TriggerType: "task_transition",
					Conditions:  map[string]interface{}{"test": "value"},
					Actions:     map[string]interface{}{"test": "action"},
				},
			},
		}
		validation, err := ra.ValidateRuleChanges(ruleSet)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if validation.Valid {
			t.Error("Rule set with invalid category should be invalid")
		}
	})
}

func TestUnit_RulesAccess_CircularDependencies(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rulesaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create RulesAccess
	ra, err := NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer ra.Close()

	// Create rule set with circular dependencies
	ruleSet := &RuleSet{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:          "rule-a",
				Name:        "Rule A",
				Category:    "workflow",
				TriggerType: "task_transition",
				Conditions:  map[string]interface{}{"test": "value"},
				Actions:     map[string]interface{}{"test": "action"},
			},
			{
				ID:          "rule-b",
				Name:        "Rule B",
				Category:    "workflow",
				TriggerType: "task_transition",
				Conditions:  map[string]interface{}{"test": "value"},
				Actions:     map[string]interface{}{"test": "action"},
			},
		},
		Dependencies: map[string][]string{
			"rule-a": {"rule-b"},
			"rule-b": {"rule-a"}, // Circular dependency
		},
	}

	validation, err := ra.ValidateRuleChanges(ruleSet)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if validation.Valid {
		t.Error("Rule set with circular dependencies should be invalid")
	}

	// Check that circular dependency was detected
	foundCircularError := false
	for _, errMsg := range validation.Errors {
		if errMsg == "circular dependency detected involving rule: rule-a" || errMsg == "circular dependency detected involving rule: rule-b" {
			foundCircularError = true
			break
		}
	}
	if !foundCircularError {
		t.Errorf("Expected circular dependency error, got errors: %v", validation.Errors)
	}
}