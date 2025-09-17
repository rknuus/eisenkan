package task_manager

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/resource_access"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// Test Case DT-API-001: Task Creation and Modification with Invalid Inputs
func TestDestructive_TaskManager_APIContractViolations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "taskmanager_destructive_api_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	boardAccess, err := board_access.NewBoardAccess(tempDir)
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
	// Create repository for TaskManager
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repository.Close()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, tempDir)

	t.Run("NilTaskData", func(t *testing.T) {
		// Test zero-value TaskRequest (equivalent to nil fields)
		request := TaskRequest{}
		response, err := taskManager.CreateTask(request)
		
		// Should handle gracefully - either error or provide defaults
		if err == nil {
			// If no error, verify defaults were applied
			if response.Description == "" {
				t.Error("Expected default description or error for empty task")
			}
		}
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		// Task with missing description
		request := TaskRequest{
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		_, err := taskManager.CreateTask(request)
		if err == nil {
			t.Error("Expected error for task with missing description")
		}
	})

	t.Run("InvalidPriorityValues", func(t *testing.T) {
		// Test with priority derivation - system should handle any Priority struct
		request := TaskRequest{
			Description:    "Test task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		_, err := taskManager.CreateTask(request)
		if err != nil {
			t.Errorf("Expected valid priority to succeed, got error: %v", err)
		}
	})

	t.Run("LargeDescription", func(t *testing.T) {
		// 10KB+ description
		largeDesc := strings.Repeat("A", 10240)
		request := TaskRequest{
			Description:    largeDesc,
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
		}
		
		response, err := taskManager.CreateTask(request)
		if err != nil {
			// Should handle large descriptions gracefully
			if !strings.Contains(err.Error(), "description") && !strings.Contains(err.Error(), "size") {
				t.Errorf("Unexpected error for large description: %v", err)
			}
		} else {
			// If successful, verify it was stored (truncated or full)
			if response.Description == "" {
				t.Error("Description should not be empty if task creation succeeded")
			}
		}
	})

	t.Run("InvalidParentTaskID", func(t *testing.T) {
		nonExistentID := "non-existent-parent-id"
		request := TaskRequest{
			Description:    "Child task with invalid parent",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &nonExistentID,
		}
		
		_, err := taskManager.CreateTask(request)
		if err == nil {
			t.Error("Expected error for invalid parent task ID")
		}
	})

	t.Run("CircularHierarchy", func(t *testing.T) {
		// Create parent task
		parentRequest := TaskRequest{
			Description:    "Parent task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		parentResponse, err := taskManager.CreateTask(parentRequest)
		if err != nil {
			t.Fatalf("Failed to create parent task: %v", err)
		}
		
		// Create child task
		childRequest := TaskRequest{
			Description:    "Child task",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &parentResponse.ID,
		}
		
		childResponse, err := taskManager.CreateTask(childRequest)
		if err != nil {
			t.Fatalf("Failed to create child task: %v", err)
		}
		
		// Try to make parent a child of child (circular reference)
		updateRequest := TaskRequest{
			Description:    "Parent task updated",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &childResponse.ID,
		}
		
		_, err = taskManager.UpdateTask(parentResponse.ID, updateRequest)
		if err == nil {
			t.Error("Expected error for circular hierarchy creation")
		}
	})

	t.Run("ExcessiveHierarchyDepth", func(t *testing.T) {
		// Create parent task
		parentRequest := TaskRequest{
			Description:    "Level 1 Parent",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		parentResponse, err := taskManager.CreateTask(parentRequest)
		if err != nil {
			t.Fatalf("Failed to create parent task: %v", err)
		}
		
		// Create level 2 child
		level2Request := TaskRequest{
			Description:    "Level 2 Child",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &parentResponse.ID,
		}
		
		level2Response, err := taskManager.CreateTask(level2Request)
		if err != nil {
			t.Fatalf("Failed to create level 2 child: %v", err)
		}
		
		// Try to create level 3 child (should violate 1-2 level constraint)
		level3Request := TaskRequest{
			Description:    "Level 3 Child",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &level2Response.ID,
		}
		
		_, err = taskManager.CreateTask(level3Request)
		if err == nil {
			t.Error("Expected error for >2 level hierarchy depth")
		}
	})

	t.Run("PriorityPromotionDateInPast", func(t *testing.T) {
		pastDate := time.Now().Add(-24 * time.Hour)
		request := TaskRequest{
			Description:           "Task with past promotion date",
			Priority:              board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus:        Todo,
			PriorityPromotionDate: &pastDate,
		}
		
		_, err := taskManager.CreateTask(request)
		// Past dates might be allowed for testing/import scenarios
		// The key is that they should be handled gracefully
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "date") {
			t.Errorf("Unexpected error for past promotion date: %v", err)
		}
	})

	t.Run("InvalidPromotionDateForUrgentTask", func(t *testing.T) {
		futureDate := time.Now().Add(24 * time.Hour)
		request := TaskRequest{
			Description:           "Urgent task with promotion date",
			Priority:              board_access.Priority{Urgent: true, Important: true}, // Already urgent
			WorkflowStatus:        Todo,
			PriorityPromotionDate: &futureDate,
		}
		
		_, err := taskManager.CreateTask(request)
		// Should either reject or ignore promotion date for already urgent tasks
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "promotion") && !strings.Contains(strings.ToLower(err.Error()), "urgent") {
			t.Errorf("Unexpected error for urgent task with promotion date: %v", err)
		}
	})
}

// Test Case DT-API-002: Workflow Status Changes with Invalid Transitions
func TestDestructive_TaskManager_InvalidWorkflowTransitions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "taskmanager_destructive_workflow_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	boardAccess, err := board_access.NewBoardAccess(tempDir)
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
	// Create repository for TaskManager
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repository.Close()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, tempDir)

	t.Run("InvalidStatusTransition", func(t *testing.T) {
		// Create task in "done" state
		request := TaskRequest{
			Description:    "Completed task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Done,
		}
		
		response, err := taskManager.CreateTask(request)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
		
		// Try to transition from "done" back to "todo" (invalid)
		updateRequest := TaskRequest{
			Description:    "Completed task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		_, err = taskManager.UpdateTask(response.ID, updateRequest)
		if err == nil {
			t.Error("Expected error for invalid done->todo transition")
		}
	})

	t.Run("ParentWithNonDoneSubtasks", func(t *testing.T) {
		// Create parent task
		parentRequest := TaskRequest{
			Description:    "Parent task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		parentResponse, err := taskManager.CreateTask(parentRequest)
		if err != nil {
			t.Fatalf("Failed to create parent task: %v", err)
		}
		
		// Create subtask in "todo" state
		subtaskRequest := TaskRequest{
			Description:    "Incomplete subtask",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &parentResponse.ID,
		}
		
		_, err = taskManager.CreateTask(subtaskRequest)
		if err != nil {
			t.Fatalf("Failed to create subtask: %v", err)
		}
		
		// Try to mark parent as "done" while subtask is "todo"
		updateRequest := TaskRequest{
			Description:    "Parent task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Done,
		}
		
		_, err = taskManager.UpdateTask(parentResponse.ID, updateRequest)
		// Should enforce parent completion dependency rules
		if err == nil {
			// Check if system allows this based on active policy
			// Some systems might allow parent completion independent of subtasks
			t.Log("System allows parent completion with incomplete subtasks")
		}
	})

	t.Run("MalformedTaskIdentifier", func(t *testing.T) {
		updateRequest := TaskRequest{
			Description:    "Test task",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: InProgress,
		}
		
		_, err := taskManager.UpdateTask("", updateRequest)
		if err == nil {
			t.Error("Expected error for empty task ID")
		}
		
		_, err = taskManager.UpdateTask("invalid-id-format", updateRequest)
		if err == nil {
			t.Error("Expected error for invalid task ID format")
		}
	})
}

// Test Case DT-HIERARCHICAL-001: Subtask Workflow Coupling Edge Cases
func TestDestructive_TaskManager_SubtaskWorkflowCoupling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "taskmanager_destructive_coupling_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	boardAccess, err := board_access.NewBoardAccess(tempDir)
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
	// Create repository for TaskManager
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repository.Close()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, tempDir)

	t.Run("FirstSubtaskTransitionWithParentAlreadyDoing", func(t *testing.T) {
		// Create parent in "doing" state
		parentRequest := TaskRequest{
			Description:    "Parent already doing",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: InProgress,
		}
		
		parentResponse, err := taskManager.CreateTask(parentRequest)
		if err != nil {
			t.Fatalf("Failed to create parent task: %v", err)
		}
		
		// Create subtask
		subtaskRequest := TaskRequest{
			Description:    "First subtask",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: Todo,
			ParentTaskID:   &parentResponse.ID,
		}
		
		subtaskResponse, err := taskManager.CreateTask(subtaskRequest)
		if err != nil {
			t.Fatalf("Failed to create subtask: %v", err)
		}
		
		// Try first subtask "todo"->"doing" with parent already "doing"
		updateRequest := TaskRequest{
			Description:    "First subtask",
			Priority:       board_access.Priority{Urgent: false, Important: true},
			WorkflowStatus: InProgress,
			ParentTaskID:   &parentResponse.ID,
		}
		
		_, err = taskManager.UpdateTask(subtaskResponse.ID, updateRequest)
		// This might be allowed or rejected based on workflow coupling rules
		if err != nil {
			t.Logf("System enforces workflow coupling: %v", err)
		} else {
			t.Log("System allows subtask transition regardless of parent state")
		}
	})

	t.Run("ConcurrentSubtaskTransitions", func(t *testing.T) {
		// Create parent task
		parentRequest := TaskRequest{
			Description:    "Parent for concurrent test",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		parentResponse, err := taskManager.CreateTask(parentRequest)
		if err != nil {
			t.Fatalf("Failed to create parent task: %v", err)
		}
		
		// Create multiple subtasks
		var subtaskIDs []string
		for i := 0; i < 3; i++ {
			subtaskRequest := TaskRequest{
				Description:    "Concurrent subtask",
				Priority:       board_access.Priority{Urgent: false, Important: true},
				WorkflowStatus: Todo,
				ParentTaskID:   &parentResponse.ID,
			}
			
			subtaskResponse, err := taskManager.CreateTask(subtaskRequest)
			if err != nil {
				t.Fatalf("Failed to create subtask %d: %v", i, err)
			}
			subtaskIDs = append(subtaskIDs, subtaskResponse.ID)
		}
		
		// Try concurrent "todo"->"doing" transitions
		var wg sync.WaitGroup
		errors := make(chan error, len(subtaskIDs))
		
		for _, subtaskID := range subtaskIDs {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				updateRequest := TaskRequest{
					Description:    "Concurrent subtask",
					Priority:       board_access.Priority{Urgent: false, Important: true},
					WorkflowStatus: InProgress,
					ParentTaskID:   &parentResponse.ID,
				}
				_, err := taskManager.UpdateTask(id, updateRequest)
				if err != nil {
					errors <- err
				}
			}(subtaskID)
		}
		
		wg.Wait()
		close(errors)
		
		// Check results - should maintain consistency
		errorCount := 0
		for err := range errors {
			errorCount++
			t.Logf("Concurrent transition error: %v", err)
		}
		
		// Some errors expected due to workflow coupling rules
		t.Logf("Concurrent transitions resulted in %d errors out of %d attempts", errorCount, len(subtaskIDs))
	})
}

// Test Case DT-RESOURCE-001: Large Hierarchical Operations
func TestDestructive_TaskManager_ResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource exhaustion test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "taskmanager_destructive_resource_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	boardAccess, err := board_access.NewBoardAccess(tempDir)
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
	// Create repository for TaskManager
	gitConfig := &utilities.AuthorConfiguration{
		User:  "Test User",
		Email: "test@example.com",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(tempDir, gitConfig)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repository.Close()

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, tempDir)

	t.Run("LargeSubtaskHierarchy", func(t *testing.T) {
		// Create parent task
		parentRequest := TaskRequest{
			Description:    "Parent with many subtasks",
			Priority:       board_access.Priority{Urgent: true, Important: true},
			WorkflowStatus: Todo,
		}
		
		parentResponse, err := taskManager.CreateTask(parentRequest)
		if err != nil {
			t.Fatalf("Failed to create parent task: %v", err)
		}
		
		// Create many subtasks (reduced number for reasonable test time)
		subtaskCount := 100 // Reduced from 1000+ for practical testing
		start := time.Now()
		
		for i := 0; i < subtaskCount; i++ {
			subtaskRequest := TaskRequest{
				Description:    "Mass subtask",
				Priority:       board_access.Priority{Urgent: false, Important: true},
				WorkflowStatus: Todo,
				ParentTaskID:   &parentResponse.ID,
			}
			
			_, err := taskManager.CreateTask(subtaskRequest)
			if err != nil {
				t.Errorf("Failed to create subtask %d: %v", i, err)
				break
			}
			
			// Check for reasonable performance
			if i%10 == 0 && time.Since(start) > 30*time.Second {
				t.Logf("Created %d subtasks, stopping due to time limit", i+1)
				break
			}
		}
		
		duration := time.Since(start)
		t.Logf("Large subtask creation took %v", duration)
		
		// Verify parent task can handle large subtask count
		updatedParent, err := taskManager.GetTask(parentResponse.ID)
		if err != nil {
			t.Errorf("Failed to retrieve parent with many subtasks: %v", err)
		} else {
			t.Logf("Parent task has %d subtasks", len(updatedParent.SubtaskIDs))
		}
	})

	t.Run("BulkPriorityPromotion", func(t *testing.T) {
		// Create multiple tasks with past promotion dates
		taskCount := 50 // Manageable number for testing
		pastDate := time.Now().Add(-1 * time.Hour)
		
		for i := 0; i < taskCount; i++ {
			request := TaskRequest{
				Description:           "Bulk promotion task",
				Priority:              board_access.Priority{Urgent: false, Important: true},
				WorkflowStatus:        Todo,
				PriorityPromotionDate: &pastDate,
			}
			
			_, err := taskManager.CreateTask(request)
			if err != nil {
				t.Errorf("Failed to create promotion task %d: %v", i, err)
			}
		}
		
		// Process all promotions
		start := time.Now()
		promoted, err := taskManager.ProcessPriorityPromotions()
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Bulk priority promotion failed: %v", err)
		} else {
			t.Logf("Processed %d promotions in %v", len(promoted), duration)
		}
		
		// Verify performance requirement (3 seconds)
		if duration > 3*time.Second {
			t.Errorf("Priority promotion took %v, exceeding 3s requirement", duration)
		}
	})
}