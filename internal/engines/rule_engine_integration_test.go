package engines

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/resource_access"
)

// TestIntegration_RuleEngine_WithRealComponents tests RuleEngine with actual ResourceAccess components
func TestIntegration_RuleEngine_WithRealComponents(t *testing.T) {
	// Create temporary directory for test board
	tempDir, err := os.MkdirTemp("", "ruleengine_integration_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize real ResourceAccess components
	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	// Create RuleEngine with real components
	ruleEngine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	t.Run("WIP Limit Rule Integration", func(t *testing.T) {
		testWIPLimitIntegration(t, ruleEngine, rulesAccess, boardAccess, tempDir)
	})

	t.Run("Required Fields Rule Integration", func(t *testing.T) {
		testRequiredFieldsIntegration(t, ruleEngine, rulesAccess, boardAccess, tempDir)
	})

	t.Run("Workflow Transition Rule Integration", func(t *testing.T) {
		testWorkflowTransitionIntegration(t, ruleEngine, rulesAccess, boardAccess, tempDir)
	})

	t.Run("Multiple Rules Integration", func(t *testing.T) {
		testMultipleRulesIntegration(t, ruleEngine, rulesAccess, boardAccess, tempDir)
	})
}

func testWIPLimitIntegration(t *testing.T, ruleEngine *RuleEngine, rulesAccess resource_access.IRulesAccess, boardAccess resource_access.IBoardAccess, tempDir string) {
	// Set up WIP limit rule
	wipLimitRuleSet := &resource_access.RuleSet{
		Version: "1.0",
		Rules: []resource_access.Rule{
			{
				ID:          "wip-limit-doing",
				Name:        "WIP Limit for Doing Column",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"max_wip_limit": 2,
				},
				Actions: map[string]any{
					"block":   true,
					"message": "WIP limit exceeded",
				},
				Priority: 100,
				Enabled:  true,
			},
		},
	}

	err := rulesAccess.ChangeRules(tempDir, wipLimitRuleSet)
	if err != nil {
		t.Fatalf("Failed to set rules: %v", err)
	}

	// Create tasks to fill the WIP limit
	task1 := &resource_access.Task{ID: "task1", Title: "First Task"}
	task2 := &resource_access.Task{ID: "task2", Title: "Second Task"}

	priority := resource_access.Priority{
		Urgent:    false,
		Important: true,
		Label:     "not-urgent-important",
	}

	doingStatus := resource_access.WorkflowStatus{
		Column:   "doing",
		Section:  "not-urgent-important",
		Position: 1,
	}

	// Add tasks to doing column (should reach WIP limit)
	_, err = boardAccess.CreateTask(task1, priority, doingStatus, nil)
	if err != nil {
		t.Fatalf("Failed to create task1: %v", err)
	}

	_, err = boardAccess.CreateTask(task2, priority, doingStatus, nil)
	if err != nil {
		t.Fatalf("Failed to create task2: %v", err)
	}

	// Try to add a third task (should violate WIP limit)
	event := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &resource_access.Task{
				ID:    "task3",
				Title: "Third Task",
			},
			Status: doingStatus,
		},
		Timestamp: time.Now(),
	}

	result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
	if err != nil {
		t.Errorf("EvaluateTaskChange failed: %v", err)
	}

	// Should be blocked due to WIP limit
	if result.Allowed {
		t.Error("Expected task change to be blocked due to WIP limit, but it was allowed")
	}

	if len(result.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(result.Violations))
	}

	if len(result.Violations) > 0 {
		violation := result.Violations[0]
		if violation.RuleID != "wip-limit-doing" {
			t.Errorf("Expected rule ID 'wip-limit-doing', got '%s'", violation.RuleID)
		}
		if violation.Priority != 100 {
			t.Errorf("Expected priority 100, got %d", violation.Priority)
		}
		t.Logf("WIP Limit violation message: %s", violation.Message)
	}
}

func testRequiredFieldsIntegration(t *testing.T, ruleEngine *RuleEngine, rulesAccess resource_access.IRulesAccess, boardAccess resource_access.IBoardAccess, tempDir string) {
	// Set up required fields rule
	requiredFieldsRuleSet := &resource_access.RuleSet{
		Version: "1.0",
		Rules: []resource_access.Rule{
			{
				ID:          "required-fields-ready",
				Name:        "Definition of Ready",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"required_fields": []any{"title", "description"},
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Missing required fields",
				},
				Priority: 90,
				Enabled:  true,
			},
		},
	}

	err := rulesAccess.ChangeRules(tempDir, requiredFieldsRuleSet)
	if err != nil {
		t.Fatalf("Failed to set rules: %v", err)
	}

	// Test task with missing description
	event := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &resource_access.Task{
				ID:          "incomplete-task",
				Title:       "Task with Title Only",
				Description: "", // Missing description
			},
			Status: resource_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
	if err != nil {
		t.Errorf("EvaluateTaskChange failed: %v", err)
	}

	// Should be blocked due to missing required field
	if result.Allowed {
		t.Error("Expected task change to be blocked due to missing description, but it was allowed")
	}

	if len(result.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(result.Violations))
	}

	if len(result.Violations) > 0 {
		violation := result.Violations[0]
		if violation.RuleID != "required-fields-ready" {
			t.Errorf("Expected rule ID 'required-fields-ready', got '%s'", violation.RuleID)
		}
		t.Logf("Required fields violation message: %s", violation.Message)
	}

	// Test task with all required fields (should be allowed)
	completeEvent := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &resource_access.Task{
				ID:          "complete-task",
				Title:       "Complete Task",
				Description: "This task has all required fields",
			},
			Status: resource_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	completeResult, err := ruleEngine.EvaluateTaskChange(context.Background(), completeEvent, tempDir)
	if err != nil {
		t.Errorf("EvaluateTaskChange failed: %v", err)
	}

	// Should be allowed
	if !completeResult.Allowed {
		t.Error("Expected complete task to be allowed, but it was blocked")
	}

	if len(completeResult.Violations) != 0 {
		t.Errorf("Expected 0 violations for complete task, got %d", len(completeResult.Violations))
	}
}

func testWorkflowTransitionIntegration(t *testing.T, ruleEngine *RuleEngine, rulesAccess resource_access.IRulesAccess, boardAccess resource_access.IBoardAccess, tempDir string) {
	// Set up workflow transition rule
	workflowRuleSet := &resource_access.RuleSet{
		Version: "1.0",
		Rules: []resource_access.Rule{
			{
				ID:          "workflow-transitions",
				Name:        "Allowed Workflow Transitions",
				Category:    "workflow",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"allowed_transitions": []any{
						"todo->doing",
						"doing->done",
						"doing->todo", // Allow moving back
					},
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Invalid workflow transition",
				},
				Priority: 80,
				Enabled:  true,
			},
		},
	}

	err := rulesAccess.ChangeRules(tempDir, workflowRuleSet)
	if err != nil {
		t.Fatalf("Failed to set rules: %v", err)
	}

	// Create a task in todo column first
	task := &resource_access.Task{
		ID:    "workflow-task",
		Title: "Workflow Test Task",
	}

	priority := resource_access.Priority{
		Urgent:    false,
		Important: true,
		Label:     "not-urgent-important",
	}

	todoStatus := resource_access.WorkflowStatus{
		Column:   "todo",
		Section:  "not-urgent-important",
		Position: 1,
	}

	taskID, err := boardAccess.CreateTask(task, priority, todoStatus, nil)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Get the created task to use as current state
	tasks, err := boardAccess.GetTasksData([]string{taskID}, false)
	if err != nil {
		t.Fatalf("Failed to get task data: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	currentTask := tasks[0]

	// Test invalid transition: todo -> done (skipping doing)
	invalidEvent := TaskEvent{
		EventType:    "task_transition",
		CurrentState: currentTask,
		FutureState: &TaskState{
			Task: task,
			Status: resource_access.WorkflowStatus{
				Column: "done", // Invalid direct jump from todo to done
			},
		},
		Timestamp: time.Now(),
	}

	result, err := ruleEngine.EvaluateTaskChange(context.Background(), invalidEvent, tempDir)
	if err != nil {
		t.Errorf("EvaluateTaskChange failed: %v", err)
	}

	// Should be blocked due to invalid transition
	if result.Allowed {
		t.Error("Expected invalid transition to be blocked, but it was allowed")
	}

	if len(result.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(result.Violations))
	}

	// Test valid transition: todo -> doing
	validEvent := TaskEvent{
		EventType:    "task_transition",
		CurrentState: currentTask,
		FutureState: &TaskState{
			Task: task,
			Status: resource_access.WorkflowStatus{
				Column: "doing", // Valid transition
			},
		},
		Timestamp: time.Now(),
	}

	validResult, err := ruleEngine.EvaluateTaskChange(context.Background(), validEvent, tempDir)
	if err != nil {
		t.Errorf("EvaluateTaskChange failed: %v", err)
	}

	// Should be allowed
	if !validResult.Allowed {
		t.Error("Expected valid transition to be allowed, but it was blocked")
	}

	if len(validResult.Violations) != 0 {
		t.Errorf("Expected 0 violations for valid transition, got %d", len(validResult.Violations))
	}
}

func testMultipleRulesIntegration(t *testing.T, ruleEngine *RuleEngine, rulesAccess resource_access.IRulesAccess, boardAccess resource_access.IBoardAccess, tempDir string) {
	// Set up multiple rules with different priorities
	multiRuleSet := &resource_access.RuleSet{
		Version: "1.0",
		Rules: []resource_access.Rule{
			{
				ID:          "high-priority-required-title",
				Name:        "High Priority Title Rule",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"required_fields": []any{"title"},
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Title is required",
				},
				Priority: 100, // High priority
				Enabled:  true,
			},
			{
				ID:          "low-priority-required-description",
				Name:        "Low Priority Description Rule",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"required_fields": []any{"description"},
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Description is required",
				},
				Priority: 50, // Low priority
				Enabled:  true,
			},
			{
				ID:          "medium-priority-wip-limit",
				Name:        "Medium Priority WIP Limit",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"max_wip_limit": 1, // Very low limit to trigger easily
				},
				Actions: map[string]any{
					"block":   true,
					"message": "WIP limit exceeded",
				},
				Priority: 75, // Medium priority
				Enabled:  true,
			},
		},
	}

	err := rulesAccess.ChangeRules(tempDir, multiRuleSet)
	if err != nil {
		t.Fatalf("Failed to set rules: %v", err)
	}

	// Create a task in doing column to trigger WIP limit
	existingTask := &resource_access.Task{
		ID:    "existing-task",
		Title: "Existing Task",
	}

	priority := resource_access.Priority{
		Urgent:    false,
		Important: true,
		Label:     "not-urgent-important",
	}

	doingStatus := resource_access.WorkflowStatus{
		Column:   "doing",
		Section:  "not-urgent-important",
		Position: 1,
	}

	_, err = boardAccess.CreateTask(existingTask, priority, doingStatus, nil)
	if err != nil {
		t.Fatalf("Failed to create existing task: %v", err)
	}

	// Test task that violates all three rules
	violatingEvent := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &resource_access.Task{
				ID:          "violating-task",
				Title:       "", // Violates high priority rule
				Description: "", // Violates low priority rule
			},
			Status: doingStatus, // Violates WIP limit (medium priority)
		},
		Timestamp: time.Now(),
	}

	result, err := ruleEngine.EvaluateTaskChange(context.Background(), violatingEvent, tempDir)
	if err != nil {
		t.Errorf("EvaluateTaskChange failed: %v", err)
	}

	// Should be blocked with multiple violations
	if result.Allowed {
		t.Error("Expected task with multiple violations to be blocked, but it was allowed")
	}

	if len(result.Violations) != 3 {
		t.Errorf("Expected 3 violations, got %d", len(result.Violations))
	}

	// Verify violations are sorted by priority (descending)
	if len(result.Violations) >= 3 {
		priorities := make([]int, len(result.Violations))
		for i, violation := range result.Violations {
			priorities[i] = violation.Priority
		}

		// Check that priorities are in descending order
		for i := 0; i < len(priorities)-1; i++ {
			if priorities[i] < priorities[i+1] {
				t.Errorf("Violations not sorted by priority: %v", priorities)
				break
			}
		}

		t.Logf("Multiple rule violations (priority sorted):")
		for i, violation := range result.Violations {
			t.Logf("  %d. Rule %s (Priority %d): %s", i+1, violation.RuleID, violation.Priority, violation.Message)
		}
	}
}

// TestIntegration_RuleEngine_GetRulesDataPerformance tests the performance of the consolidated GetRulesData
func TestIntegration_RuleEngine_GetRulesDataPerformance(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ruleengine_performance_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	// Create several tasks across different columns
	priority := resource_access.Priority{
		Urgent:    false,
		Important: true,
		Label:     "not-urgent-important",
	}

	columns := []string{"todo", "doing", "done"}
	taskIDs := make([]string, 0, 15)

	for _, column := range columns {
		for i := 0; i < 5; i++ {
			task := &resource_access.Task{
				ID:          "",
				Title:       fmt.Sprintf("Task %d in %s", i+1, column),
				Description: fmt.Sprintf("Description for task %d", i+1),
			}

			status := resource_access.WorkflowStatus{
				Column:   column,
				Section:  "not-urgent-important",
				Position: i + 1,
			}

			taskID, err := boardAccess.CreateTask(task, priority, status, nil)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
			taskIDs = append(taskIDs, taskID)
		}
	}

	// Test GetRulesData performance
	start := time.Now()

	rulesData, err := boardAccess.GetRulesData(taskIDs[0], []string{"todo", "doing"})
	if err != nil {
		t.Fatalf("GetRulesData failed: %v", err)
	}

	duration := time.Since(start)

	// Verify data completeness
	if len(rulesData.WIPCounts) == 0 {
		t.Error("Expected WIP counts, got empty map")
	}

	expectedWIPCounts := map[string]int{
		"todo":  5,
		"doing": 5,
		"done":  5,
	}

	for column, expectedCount := range expectedWIPCounts {
		if actualCount, exists := rulesData.WIPCounts[column]; !exists || actualCount != expectedCount {
			t.Errorf("Expected WIP count for %s: %d, got: %d", column, expectedCount, actualCount)
		}
	}

	if len(rulesData.ColumnTasks) != 2 { // Only requested todo and doing
		t.Errorf("Expected 2 column task groups, got %d", len(rulesData.ColumnTasks))
	}

	if len(rulesData.TaskHistory) == 0 {
		t.Error("Expected task history, got empty slice")
	}

	if rulesData.BoardMetadata["board_name"] == "" {
		t.Error("Expected board metadata, got empty")
	}

	t.Logf("GetRulesData completed in %v with %d WIP counts, %d column groups, %d history entries",
		duration, len(rulesData.WIPCounts), len(rulesData.ColumnTasks), len(rulesData.TaskHistory))

	// Performance should be reasonable (less than 1 second for this small dataset)
	if duration > time.Second {
		t.Errorf("GetRulesData took too long: %v", duration)
	}
}