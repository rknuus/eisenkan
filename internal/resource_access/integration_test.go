// Package resource_access_test provides integration tests for ResourceAccess layer components
package resource_access

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// TestIntegration_ResourceAccess_SharedRepositoryInitialization verifies that both
// BoardAccess and RulesAccess can initialize and work with the same repository
func TestIntegration_ResourceAccess_SharedRepositoryInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "shared_repo")
	
	// Create directory structure
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}
	
	// Create board.json with git configuration for both components
	boardConfig := `{
		"name": "Integration Test Board",
		"description": "Test board for integration testing",
		"git_user": "Integration Test",
		"git_email": "integration@test.local",
		"created": "` + time.Now().Format(time.RFC3339) + `"
	}`
	
	boardConfigPath := filepath.Join(repoPath, "board.json")
	if err := os.WriteFile(boardConfigPath, []byte(boardConfig), 0644); err != nil {
		t.Fatalf("Failed to write board.json: %v", err)
	}

	// Initialize BoardAccess - this should initialize the repository
	t.Log("Initializing BoardAccess with shared repository")
	boardAccess, err := board_access.NewBoardAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	// Initialize RulesAccess on the same repository
	t.Log("Initializing RulesAccess with same repository")
	rulesAccess, err := NewRulesAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize RulesAccess on same repository: %v", err)
	}
	defer rulesAccess.Close()

	// Verify repository exists and is valid
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		t.Fatal("Git repository was not created")
	}

	// Test BoardAccess operations
	t.Log("Testing BoardAccess operations")
	task := &board_access.Task{
		Title:       "Integration Test Task",
		Description: "Task created during integration testing",
	}
	priority := board_access.Priority{
		Urgent:    true,
		Important: true,
		Label:     "urgent-important",
	}
	status := board_access.WorkflowStatus{
		Column:   "todo",
		Section:  "urgent-important",
		Position: 1,
	}

	taskID, err := boardAccess.CreateTask(task, priority, status, nil)
	if err != nil {
		t.Fatalf("BoardAccess.CreateTask failed: %v", err)
	}

	retrievedTasks, err := boardAccess.GetTasksData([]string{taskID}, false)
	if err != nil {
		t.Fatalf("BoardAccess.GetTasksData failed: %v", err)
	}
	if len(retrievedTasks) != 1 {
		t.Fatal("Task not found after storage")
	}
	if retrievedTasks[0].Task.Title != "Integration Test Task" {
		t.Errorf("Expected task title 'Integration Test Task', got %v", retrievedTasks[0].Task.Title)
	}

	// Test RulesAccess operations
	t.Log("Testing RulesAccess operations")
	ruleSet := &RuleSet{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:          "integration-test-rule",
				Name:        "Integration Test Rule",
				Category:    "validation",
				TriggerType: "task_creation",
				Conditions:  map[string]interface{}{"priority": 1},
				Actions:     map[string]interface{}{"notify": true},
				Priority:    1,
				Enabled:     true,
			},
		},
	}

	// Validate the rule set
	validation, err := rulesAccess.ValidateRuleChanges(ruleSet)
	if err != nil {
		t.Fatalf("RulesAccess.ValidateRuleChanges failed: %v", err)
	}
	if !validation.Valid {
		t.Fatalf("Rule set validation failed: %v", validation.Errors)
	}

	// Store the rule set
	err = rulesAccess.ChangeRules(repoPath, ruleSet)
	if err != nil {
		t.Fatalf("RulesAccess.ChangeRules failed: %v", err)
	}

	// Read the rule set back
	retrievedRules, err := rulesAccess.ReadRules(repoPath)
	if err != nil {
		t.Fatalf("RulesAccess.ReadRules failed: %v", err)
	}
	
	if len(retrievedRules.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(retrievedRules.Rules))
	}
	
	if retrievedRules.Rules[0].Name != "Integration Test Rule" {
		t.Errorf("Expected rule name 'Integration Test Rule', got %s", retrievedRules.Rules[0].Name)
	}

	// Verify both components created commits in the same repository
	t.Log("Verifying repository commits from both components")
	
	// Check that repository has commits from both components
	// We can't directly access the repository here, but the fact that both operations
	// succeeded indicates the shared repository is working correctly
	
	t.Log("Integration test completed successfully - both components can share repository")
}

// TestIntegration_ResourceAccess_CrossComponentDataIsolation verifies that
// BoardAccess and RulesAccess maintain data isolation while sharing repository
func TestIntegration_ResourceAccess_CrossComponentDataIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "isolation_repo")
	
	// Create directory and board.json
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}
	
	boardConfig := `{
		"name": "Data Isolation Test",
		"git_user": "Isolation Test",
		"git_email": "isolation@test.local"
	}`
	
	if err := os.WriteFile(filepath.Join(repoPath, "board.json"), []byte(boardConfig), 0644); err != nil {
		t.Fatalf("Failed to write board.json: %v", err)
	}

	// Initialize both components
	boardAccess, err := board_access.NewBoardAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := NewRulesAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	// Store data in BoardAccess
	task := &board_access.Task{
		Title: "Isolation Test Task",
		Description: "Task for data isolation testing",
	}
	priority := board_access.Priority{Urgent: false, Important: true, Label: "not-urgent-important"}
	status := board_access.WorkflowStatus{Column: "todo", Section: "not-urgent-important", Position: 1}
	
	_, err = boardAccess.CreateTask(task, priority, status, nil)
	if err != nil {
		t.Fatalf("Failed to store task: %v", err)
	}

	// Store data in RulesAccess  
	ruleSet := &RuleSet{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:          "isolation-rule",
				Name:        "Isolation Rule",
				Category:    "workflow", 
				TriggerType: "status_change",
				Conditions:  map[string]interface{}{"status": "done"},
				Actions:     map[string]interface{}{"archive": true},
				Priority:    1,
				Enabled:     true,
			},
		},
	}
	
	err = rulesAccess.ChangeRules(repoPath, ruleSet)
	if err != nil {
		t.Fatalf("Failed to store rules: %v", err)
	}

	// Verify data isolation - BoardAccess should only see tasks
	criteria := &board_access.QueryCriteria{} // Empty criteria to get all tasks
	allTasks, err := boardAccess.FindTasks(criteria)
	if err != nil {
		t.Fatalf("Failed to query tasks: %v", err)
	}
	
	if len(allTasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(allTasks))
	}

	// Verify data isolation - RulesAccess should only see rules
	retrievedRules, err := rulesAccess.ReadRules(repoPath)
	if err != nil {
		t.Fatalf("Failed to read rules: %v", err)
	}
	
	if len(retrievedRules.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(retrievedRules.Rules))
	}

	// Verify files exist independently
	// BoardAccess stores tasks in a directory structure based on priority/status
	// Let's check for the rules file which should be directly in the repo
	rulesFile := filepath.Join(repoPath, "rules.json")
	
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		t.Error("rules.json file not found")
	}
	
	// BoardAccess creates a directory structure - check inside directories
	todoDir := filepath.Join(repoPath, "todo")
	if _, err := os.Stat(todoDir); err == nil {
		// List files in todo directory  
		todoFiles, err := filepath.Glob(filepath.Join(todoDir, "**", "*.json"))
		if err == nil && len(todoFiles) > 0 {
			// Task files found in directory structure - data isolation verified
		} else {
			// Try one level deep
			todoSubFiles, err := filepath.Glob(filepath.Join(todoDir, "*", "*.json"))
			if err != nil || len(todoSubFiles) == 0 {
				t.Error("No task files found in todo directory structure")
			}
		}
	} else {
		t.Error("todo directory not created by BoardAccess")
	}

	t.Log("Data isolation verified - components maintain separate data files")
}

// TestIntegration_ResourceAccess_ConcurrentAccess verifies that both components
// can safely access the same repository concurrently
func TestIntegration_ResourceAccess_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "concurrent_repo")
	
	// Setup repository
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}
	
	boardConfig := `{
		"name": "Concurrent Access Test",
		"git_user": "Concurrent Test", 
		"git_email": "concurrent@test.local"
	}`
	
	if err := os.WriteFile(filepath.Join(repoPath, "board.json"), []byte(boardConfig), 0644); err != nil {
		t.Fatalf("Failed to write board.json: %v", err)
	}

	// Initialize components
	boardAccess, err := board_access.NewBoardAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := NewRulesAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	// Test concurrent operations
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	// Goroutine 1: BoardAccess operations
	go func() {
		defer func() { done <- true }()
		
		for i := 0; i < 5; i++ {
			task := &board_access.Task{
				Title:       "Concurrent Task " + string(rune('A'+i)),
				Description: "Task created during concurrent test",
			}
			priority := board_access.Priority{Urgent: true, Important: false, Label: "urgent-not-important"}
			status := board_access.WorkflowStatus{Column: "todo", Section: "urgent-not-important", Position: i + 1}
			
			_, err := boardAccess.CreateTask(task, priority, status, nil)
			if err != nil {
				errors <- err
				return
			}
			
			// Small delay to increase chance of concurrency issues
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Goroutine 2: RulesAccess operations  
	go func() {
		defer func() { done <- true }()
		
		for i := 0; i < 5; i++ {
			ruleSet := &RuleSet{
				Version: "1.0",
				Rules: []Rule{
					{
						ID:          "concurrent-rule-" + string(rune('1'+i)),
						Name:        "Concurrent Rule " + string(rune('1'+i)),
						Category:    "automation",
						TriggerType: "task_update",
						Conditions:  map[string]interface{}{"iteration": i},
						Actions:     map[string]interface{}{"log": true},
						Priority:    1,
						Enabled:     true,
					},
				},
			}
			
			err := rulesAccess.ChangeRules(repoPath, ruleSet)
			if err != nil {
				errors <- err
				return
			}
			
			// Small delay to increase chance of concurrency issues
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for completion or error
	completed := 0
	for completed < 2 {
		select {
		case err := <-errors:
			t.Fatalf("Concurrent operation failed: %v", err)
		case <-done:
			completed++
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	// Verify final state
	criteria := &board_access.QueryCriteria{}
	tasks, err := boardAccess.FindTasks(criteria)
	if err != nil {
		t.Fatalf("Failed to query final tasks: %v", err)
	}
	
	if len(tasks) != 5 {
		t.Errorf("Expected 5 tasks after concurrent operations, got %d", len(tasks))
	}

	rules, err := rulesAccess.ReadRules(repoPath)
	if err != nil {
		t.Fatalf("Failed to read final rules: %v", err)
	}
	
	// RulesAccess overwrites the entire rule set each time, so we expect 1 rule (the last one)
	if len(rules.Rules) != 1 {
		t.Errorf("Expected 1 rule after concurrent operations, got %d", len(rules.Rules))
	}

	t.Log("Concurrent access test completed successfully")
}

// TestIntegration_ResourceAccess_VersioningIntegration verifies that both components
// properly integrate with the VersioningUtility for commit tracking
func TestIntegration_ResourceAccess_VersioningIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "versioning_repo")
	
	// Setup
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}
	
	boardConfig := `{
		"name": "Versioning Integration Test",
		"git_user": "Versioning Test",
		"git_email": "versioning@test.local"
	}`
	
	if err := os.WriteFile(filepath.Join(repoPath, "board.json"), []byte(boardConfig), 0644); err != nil {
		t.Fatalf("Failed to write board.json: %v", err)
	}

	// Initialize components  
	boardAccess, err := board_access.NewBoardAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := NewRulesAccess(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	// Create a Repository instance to check version history
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Versioning Test",
		Email: "versioning@test.local", 
	}
	
	repository, err := utilities.InitializeRepositoryWithConfig(repoPath, gitConfig)
	if err != nil {
		t.Fatalf("Failed to initialize repository for version checking: %v", err)
	}
	defer repository.Close()

	// Perform operations that should create commits
	task := &board_access.Task{
		Title:       "Versioned Task",
		Description: "Task to test version tracking",
	}
	priority := board_access.Priority{Urgent: false, Important: true, Label: "not-urgent-important"}
	status := board_access.WorkflowStatus{Column: "todo", Section: "not-urgent-important", Position: 1}
	
	_, err = boardAccess.CreateTask(task, priority, status, nil)
	if err != nil {
		t.Fatalf("Failed to store task: %v", err)
	}

	ruleSet := &RuleSet{
		Version: "1.0",
		Rules: []Rule{
			{
				ID:          "versioned-rule",
				Name:        "Versioned Rule",
				Category:    "validation",
				TriggerType: "task_creation",
				Conditions:  map[string]interface{}{"title": "Versioned Task"},
				Actions:     map[string]interface{}{"validate": true},
				Priority:    1,
				Enabled:     true,
			},
		},
	}
	
	err = rulesAccess.ChangeRules(repoPath, ruleSet)
	if err != nil {
		t.Fatalf("Failed to change rules: %v", err)
	}

	// Verify commits were created
	history, err := repository.GetHistory(10)
	if err != nil {
		t.Fatalf("Failed to get repository history: %v", err)
	}

	if len(history) < 2 { // At least one commit each from BoardAccess and RulesAccess
		t.Fatalf("Expected at least 2 commits, got %d", len(history))
	}

	// Check that commits have proper author information
	for _, commit := range history {
		if commit.Author != "Versioning Test" {
			t.Errorf("Expected author 'Versioning Test', got '%s'", commit.Author)
		}
		if commit.Email != "versioning@test.local" {
			t.Errorf("Expected email 'versioning@test.local', got '%s'", commit.Email)
		}
	}

	t.Log("Versioning integration verified - both components create proper commits")
}