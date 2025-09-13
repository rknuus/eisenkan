package engines

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rknuus/eisenkan/internal/resource_access"
)

// ExampleRuleEngineUsage demonstrates how to use the RuleEngine
func ExampleRuleEngineUsage() {
	// This example shows how TaskManager would use RuleEngine

	// Initialize ResourceAccess components (typically done in main or service layer)
	rulesAccess, err := resource_access.NewRulesAccess("/tmp/test-board")
	if err != nil {
		log.Fatalf("Failed to create RulesAccess: %v", err)
	}
	defer rulesAccess.Close()

	boardAccess, err := resource_access.NewBoardAccess("/tmp/test-board")
	if err != nil {
		log.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer boardAccess.Close()

	// Create RuleEngine
	ruleEngine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		log.Fatalf("Failed to create RuleEngine: %v", err)
	}
	defer ruleEngine.Close()

	// Example 1: Task transition validation
	fmt.Println("=== Example 1: Task Transition Validation ===")

	taskEvent := TaskEvent{
		EventType: "task_transition",
		CurrentState: &resource_access.TaskWithTimestamps{
			Task: &resource_access.Task{
				ID:    "task-123",
				Title: "Example Task",
			},
			Status: resource_access.WorkflowStatus{
				Column: "todo",
			},
		},
		FutureState: &TaskState{
			Task: &resource_access.Task{
				ID:    "task-123",
				Title: "Example Task",
			},
			Status: resource_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := ruleEngine.EvaluateTaskChange(context.Background(), taskEvent, "/tmp/test-board")
	if err != nil {
		log.Printf("Rule evaluation failed: %v", err)
		return
	}

	if result.Allowed {
		fmt.Println("✅ Task transition allowed")
	} else {
		fmt.Printf("❌ Task transition blocked. Violations:\n")
		for _, violation := range result.Violations {
			fmt.Printf("  - Rule %s (Priority %d): %s\n", violation.RuleID, violation.Priority, violation.Message)
		}
	}

	// Example 2: Task creation validation
	fmt.Println("\n=== Example 2: Task Creation Validation ===")

	creationEvent := TaskEvent{
		EventType: "task_create",
		FutureState: &TaskState{
			Task: &resource_access.Task{
				ID:          "task-456",
				Title:       "", // Missing title - may violate rules
				Description: "A task without a title",
			},
			Priority: resource_access.Priority{
				Urgent:    true,
				Important: true,
				Label:     "urgent-important",
			},
			Status: resource_access.WorkflowStatus{
				Column:  "todo",
				Section: "urgent-important",
			},
		},
		Timestamp: time.Now(),
	}

	result, err = ruleEngine.EvaluateTaskChange(context.Background(), creationEvent, "/tmp/test-board")
	if err != nil {
		log.Printf("Rule evaluation failed: %v", err)
		return
	}

	if result.Allowed {
		fmt.Println("✅ Task creation allowed")
	} else {
		fmt.Printf("❌ Task creation blocked. Violations:\n")
		for _, violation := range result.Violations {
			fmt.Printf("  - Rule %s (Priority %d): %s\n", violation.RuleID, violation.Priority, violation.Message)
			if violation.Details != "" {
				fmt.Printf("    Details: %s\n", violation.Details)
			}
		}
	}
}

// ExampleRuleSetup demonstrates how to configure rules that RuleEngine can evaluate
func ExampleRuleSetup() *resource_access.RuleSet {
	return &resource_access.RuleSet{
		Version: "1.0",
		Rules: []resource_access.Rule{
			{
				ID:          "wip-limit-doing",
				Name:        "WIP Limit for Doing Column",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"max_wip_limit": 3,
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Cannot exceed 3 tasks in doing column",
				},
				Priority: 100,
				Enabled:  true,
				Metadata: map[string]string{
					"description": "Limits work in progress to maintain team focus",
					"author":      "Team Lead",
				},
			},
			{
				ID:          "required-fields-ready",
				Name:        "Definition of Ready",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"required_fields": []any{"title", "description"},
					"target_column":   "doing",
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Tasks must have title and description before starting work",
				},
				Priority: 90,
				Enabled:  true,
			},
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
			{
				ID:          "age-limit-doing",
				Name:        "Age Limit for Doing Column",
				Category:    "automation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"max_age_days": 14,
					"column":       "doing",
				},
				Actions: map[string]any{
					"warn":    true,
					"message": "Task has been in progress for too long",
				},
				Priority: 50,
				Enabled:  true,
			},
		},
		Metadata: map[string]string{
			"board_name":   "Development Board",
			"methodology":  "Kanban",
			"last_updated": time.Now().Format(time.RFC3339),
		},
	}
}
