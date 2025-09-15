package managers

import (
	"os"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/resource_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// Integration tests for TaskManager with real dependencies
func TestIntegration_TaskManager_WithRealDependencies(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "taskmanager_integration_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real dependencies
	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	ruleEngine, err := engines.NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	logger := utilities.NewLoggingUtility()

	// Create TaskManager with real dependencies
	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, tempDir)
	if taskManager == nil {
		t.Fatal("Expected TaskManager to be created, got nil")
	}

	t.Run("CreateTask", func(t *testing.T) {
		request := TaskRequest{
			Description:    "Integration test task",
			Priority:       resource_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
			Tags:           []string{"integration", "test"},
		}

		response, err := taskManager.CreateTask(request)
		if err != nil {
			t.Fatalf("Expected task creation to succeed, got error: %v", err)
		}

		if response.ID == "" {
			t.Error("Expected task ID to be set")
		}
		if response.Description != "Integration test task" {
			t.Errorf("Expected description 'Integration test task', got '%s'", response.Description)
		}
		if response.Priority.Label != "urgent-important" {
			t.Errorf("Expected priority label 'urgent-important', got '%s'", response.Priority.Label)
		}
	})
}

func TestIntegration_TaskManager_PriorityPromotion(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "taskmanager_promotion_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real dependencies
	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	ruleEngine, err := engines.NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	logger := utilities.NewLoggingUtility()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, tempDir)

	// Create task with promotion date in the past
	pastDate := time.Now().Add(-24 * time.Hour)
	request := TaskRequest{
		Description:           "Task needing promotion",
		Priority:              resource_access.Priority{Urgent: false, Important: true},
		WorkflowStatus:        Todo,
		PriorityPromotionDate: &pastDate,
	}

	// Create the task
	response, err := taskManager.CreateTask(request)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify initial priority
	if response.Priority.Label != "not-urgent-important" {
		t.Errorf("Expected initial priority 'not-urgent-important', got '%s'", response.Priority.Label)
	}

	// Process priority promotions
	promoted, err := taskManager.ProcessPriorityPromotions()
	if err != nil {
		t.Fatalf("Failed to process priority promotions: %v", err)
	}

	if len(promoted) != 1 {
		t.Errorf("Expected 1 promoted task, got %d", len(promoted))
	}

	// Verify the task was promoted
	if len(promoted) > 0 && promoted[0].Priority.Label != "urgent-important" {
		t.Errorf("Expected promoted priority 'urgent-important', got '%s'", promoted[0].Priority.Label)
	}
}

func TestIntegration_TaskManager_SubtaskWorkflows(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "taskmanager_subtasks_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real dependencies
	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	ruleEngine, err := engines.NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	logger := utilities.NewLoggingUtility()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, tempDir)

	// Create parent task
	parentRequest := TaskRequest{
		Description:    "Parent task",
		Priority:       resource_access.Priority{Urgent: true, Important: true},
		WorkflowStatus: Todo,
	}

	parentResponse, err := taskManager.CreateTask(parentRequest)
	if err != nil {
		t.Fatalf("Failed to create parent task: %v", err)
	}

	// Create subtask
	subtaskRequest := TaskRequest{
		Description:    "Subtask",
		Priority:       resource_access.Priority{Urgent: false, Important: true},
		WorkflowStatus: Todo,
		ParentTaskID:   &parentResponse.ID,
	}

	subtaskResponse, err := taskManager.CreateTask(subtaskRequest)
	if err != nil {
		t.Fatalf("Failed to create subtask: %v", err)
	}

	// Verify parent-child relationship
	if subtaskResponse.ParentTaskID == nil || *subtaskResponse.ParentTaskID != parentResponse.ID {
		t.Errorf("Expected subtask to have parent ID %s", parentResponse.ID)
	}

	// Get parent task and verify it shows the subtask
	updatedParent, err := taskManager.GetTask(parentResponse.ID)
	if err != nil {
		t.Fatalf("Failed to get updated parent task: %v", err)
	}

	if len(updatedParent.SubtaskIDs) != 1 || updatedParent.SubtaskIDs[0] != subtaskResponse.ID {
		t.Errorf("Expected parent to have subtask ID %s", subtaskResponse.ID)
	}
}

func TestIntegration_TaskManager_RuleEngineIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "taskmanager_rules_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real dependencies
	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	ruleEngine, err := engines.NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	logger := utilities.NewLoggingUtility()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, tempDir)

	// Test that rule engine is being called (should succeed with default empty rules)
	request := TaskRequest{
		Description:    "Rule engine test task",
		Priority:       resource_access.Priority{Urgent: true, Important: true},
		WorkflowStatus: Todo,
	}

	response, err := taskManager.CreateTask(request)
	if err != nil {
		t.Fatalf("Expected task creation to succeed with rule validation, got error: %v", err)
	}

	// Verify task was created successfully
	if response.ID == "" {
		t.Error("Expected task ID to be set")
	}
	if response.WorkflowStatus != Todo {
		t.Errorf("Expected task status to be 'todo', got '%s'", response.WorkflowStatus)
	}
}

func TestIntegration_TaskManager_FullWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "taskmanager_workflow_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real dependencies
	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	ruleEngine, err := engines.NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	logger := utilities.NewLoggingUtility()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, tempDir)

	// Test workflow: Create -> Update
	request := TaskRequest{
		Description:    "Workflow test task",
		Priority:       resource_access.Priority{Urgent: true, Important: true},
		WorkflowStatus: Todo,
		Tags:           []string{"workflow", "test"},
	}

	// 1. Create task
	response, err := taskManager.CreateTask(request)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	taskID := response.ID

	// 2. Update task
	updateRequest := TaskRequest{
		Description:    "Updated workflow test task",
		Priority:       resource_access.Priority{Urgent: true, Important: true},
		WorkflowStatus: Todo,
		Tags:           []string{"workflow", "test", "updated"},
	}

	_, err = taskManager.UpdateTask(taskID, updateRequest)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// 3. Verify task was updated
	updatedTask, err := taskManager.GetTask(taskID)
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if updatedTask.Description != "Updated workflow test task" {
		t.Errorf("Expected updated description, got '%s'", updatedTask.Description)
	}

	if len(updatedTask.Tags) != 3 || updatedTask.Tags[2] != "updated" {
		t.Errorf("Expected 3 tags including 'updated', got %v", updatedTask.Tags)
	}
}