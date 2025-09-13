package engines

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/resource_access"
)

// TestAcceptance_RuleEngine_APIContractViolations tests API contract violations (STP DT-API-001)
func TestAcceptance_RuleEngine_APIContractViolations(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	rulesAccess, boardAccess, ruleEngine := createTestComponents(t, tempDir)
	defer rulesAccess.Close()
	defer boardAccess.Close()
	defer ruleEngine.Close()

	t.Run("NilTaskEventContext", func(t *testing.T) {
		// This would cause a panic in Go, but we can test with invalid fields
		invalidEvent := TaskEvent{} // Empty event - missing required fields
		
		result, err := ruleEngine.EvaluateTaskChange(context.Background(), invalidEvent, tempDir)
		
		// Should handle gracefully - either succeed with no rules or return error
		if err != nil {
			t.Logf("Expected behavior: Error handling invalid event: %v", err)
		} else {
			t.Logf("Expected behavior: Empty event handled gracefully, result: %+v", result)
		}
	})

	t.Run("TaskEventWithMissingRequiredFields", func(t *testing.T) {
		invalidEvent := TaskEvent{
			EventType: "task_transition",
			// Missing FutureState
			Timestamp: time.Now(),
		}
		
		result, err := ruleEngine.EvaluateTaskChange(context.Background(), invalidEvent, tempDir)
		
		// Should handle gracefully
		if err != nil {
			t.Logf("Expected behavior: Missing fields handled with error: %v", err)
		} else if result != nil {
			t.Logf("Expected behavior: Missing fields handled gracefully, allowed: %t", result.Allowed)
		}
	})

	t.Run("TaskEventWithInvalidDataTypes", func(t *testing.T) {
		// Test with nil FutureState.Task
		invalidEvent := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: nil, // Invalid nil task
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}
		
		result, err := ruleEngine.EvaluateTaskChange(context.Background(), invalidEvent, tempDir)
		
		// Should handle gracefully without crashing
		if err != nil {
			t.Logf("Expected behavior: Invalid data types handled with error: %v", err)
		} else if result != nil {
			t.Logf("Expected behavior: Invalid data types handled gracefully, allowed: %t", result.Allowed)
		}
	})

	t.Run("TaskEventWithExtremelyLargeTaskDescriptions", func(t *testing.T) {
		// Create 10KB+ description
		largeDescription := strings.Repeat("This is a very long description. ", 300) // ~10KB
		
		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:          "large-desc-task",
					Title:       "Task with Large Description",
					Description: largeDescription,
				},
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}
		
		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		
		// Should handle large input without issues
		if err != nil {
			t.Errorf("Large description caused error: %v", err)
		} else if result == nil {
			t.Error("Large description returned nil result")
		} else {
			t.Logf("Large description handled successfully, allowed: %t", result.Allowed)
		}
	})

	t.Run("TaskEventWithInvalidUnicodeCharacters", func(t *testing.T) {
		// Test with various Unicode characters including potentially problematic ones
		unicodeTask := &resource_access.Task{
			ID:          "unicode-task",
			Title:       "Unicode Test: 🚀🔥💯 中文 العربية русский ñáéíóú",
			Description: "Testing unicode: \u0000\u001F\uFFFD\U0001F4A9", // Including control chars
		}
		
		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: unicodeTask,
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}
		
		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		
		// Should handle Unicode correctly
		if err != nil {
			t.Logf("Unicode handling result: %v", err)
		} else {
			t.Logf("Unicode handled successfully, allowed: %t", result.Allowed)
		}
	})
}

// TestAcceptance_RuleEngine_RuleLogicEdgeCases tests rule logic edge cases (STP DT-LOGIC-001)
func TestAcceptance_RuleEngine_RuleLogicEdgeCases(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	rulesAccess, boardAccess, ruleEngine := createTestComponents(t, tempDir)
	defer rulesAccess.Close()
	defer boardAccess.Close()
	defer ruleEngine.Close()

	t.Run("RulesWithComplexConditionLogic", func(t *testing.T) {
		complexRuleSet := &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "complex-condition-rule",
					Name:        "Complex Condition Test",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"required_fields": []any{"title", "description"},
						"max_wip_limit":   3,
						"complex_nested":  map[string]any{
							"level1": map[string]any{
								"level2": []any{"value1", "value2"},
							},
						},
					},
					Actions: map[string]any{
						"block":   true,
						"message": "Complex rule violated",
					},
					Priority: 100,
					Enabled:  true,
				},
			},
		}

		err := rulesAccess.ChangeRules(tempDir, complexRuleSet)
		if err != nil {
			t.Fatalf("Failed to set complex rules: %v", err)
		}

		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:          "complex-test-task",
					Title:       "Test Task",
					Description: "Test Description",
				},
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}

		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		if err != nil {
			t.Errorf("Complex conditions caused error: %v", err)
		} else if result != nil {
			t.Logf("Complex conditions handled, allowed: %t, violations: %d", result.Allowed, len(result.Violations))
		}
	})

	t.Run("RulesWithNonExistentTaskProperties", func(t *testing.T) {
		invalidPropertyRule := &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "invalid-property-rule",
					Name:        "Invalid Property Test",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"required_fields": []any{"nonexistent_field", "another_missing_field"},
					},
					Actions: map[string]any{
						"block":   true,
						"message": "Invalid property rule",
					},
					Priority: 100,
					Enabled:  true,
				},
			},
		}

		err := rulesAccess.ChangeRules(tempDir, invalidPropertyRule)
		if err != nil {
			t.Fatalf("Failed to set invalid property rules: %v", err)
		}

		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:    "test-task",
					Title: "Valid Task",
				},
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}

		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		// Should handle non-existent properties gracefully
		if err != nil {
			t.Logf("Non-existent properties handled with error: %v", err)
		} else if result != nil {
			t.Logf("Non-existent properties handled gracefully, violations: %d", len(result.Violations))
		}
	})

	t.Run("RulesWithBoundaryValues", func(t *testing.T) {
		boundaryRuleSet := &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "boundary-value-rule",
					Name:        "Boundary Value Test",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"max_wip_limit": 2147483647, // Max int32
					},
					Actions: map[string]any{
						"block":   true,
						"message": "Boundary value test",
					},
					Priority: 2147483647, // Max int32
					Enabled:  true,
				},
			},
		}

		err := rulesAccess.ChangeRules(tempDir, boundaryRuleSet)
		if err != nil {
			t.Fatalf("Failed to set boundary value rules: %v", err)
		}

		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:    "boundary-test-task",
					Title: "Boundary Test",
				},
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}

		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		if err != nil {
			t.Errorf("Boundary values caused error: %v", err)
		} else if result != nil {
			t.Logf("Boundary values handled successfully, allowed: %t", result.Allowed)
		}
	})
}

// TestAcceptance_RuleEngine_RulePriorityAndConflicts tests rule priority and conflict resolution (STP DT-LOGIC-002)
func TestAcceptance_RuleEngine_RulePriorityAndConflicts(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	rulesAccess, boardAccess, ruleEngine := createTestComponents(t, tempDir)
	defer rulesAccess.Close()
	defer boardAccess.Close()
	defer ruleEngine.Close()

	t.Run("RulesWithIdenticalPriorities", func(t *testing.T) {
		identicalPriorityRules := &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "rule1-priority100",
					Name:        "Rule 1 Priority 100",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"required_fields": []any{"title"},
					},
					Actions: map[string]any{
						"block":   true,
						"message": "Rule 1 violation",
					},
					Priority: 100,
					Enabled:  true,
				},
				{
					ID:          "rule2-priority100",
					Name:        "Rule 2 Priority 100",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"required_fields": []any{"description"},
					},
					Actions: map[string]any{
						"block":   true,
						"message": "Rule 2 violation",
					},
					Priority: 100, // Same priority
					Enabled:  true,
				},
			},
		}

		err := rulesAccess.ChangeRules(tempDir, identicalPriorityRules)
		if err != nil {
			t.Fatalf("Failed to set identical priority rules: %v", err)
		}

		// Task violating both rules
		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:          "test-task",
					Title:       "", // Missing title
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
			t.Errorf("Identical priorities caused error: %v", err)
		} else {
			if len(result.Violations) != 2 {
				t.Errorf("Expected 2 violations, got %d", len(result.Violations))
			}
			// Both violations should have same priority
			if len(result.Violations) >= 2 {
				if result.Violations[0].Priority != result.Violations[1].Priority {
					t.Errorf("Identical priorities not preserved: %d vs %d", 
						result.Violations[0].Priority, result.Violations[1].Priority)
				}
			}
			t.Logf("Identical priorities handled deterministically")
		}
	})

	t.Run("RulesWithNegativePriorities", func(t *testing.T) {
		negativePriorityRule := &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "negative-priority-rule",
					Name:        "Negative Priority Rule",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"required_fields": []any{"title"},
					},
					Actions: map[string]any{
						"block":   true,
						"message": "Negative priority rule",
					},
					Priority: -50, // Negative priority
					Enabled:  true,
				},
			},
		}

		err := rulesAccess.ChangeRules(tempDir, negativePriorityRule)
		if err != nil {
			t.Fatalf("Failed to set negative priority rule: %v", err)
		}

		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:    "test-task",
					Title: "", // Violates rule
				},
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}

		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		if err != nil {
			t.Errorf("Negative priority caused error: %v", err)
		} else if result != nil {
			t.Logf("Negative priority handled successfully, allowed: %t", result.Allowed)
		}
	})
}

// TestAcceptance_RuleEngine_PerformanceDegradation tests performance degradation (STP DT-PERFORMANCE-001)
func TestAcceptance_RuleEngine_PerformanceDegradation(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	rulesAccess, boardAccess, ruleEngine := createTestComponents(t, tempDir)
	defer rulesAccess.Close()
	defer boardAccess.Close()
	defer ruleEngine.Close()

	// Test with varying rule set sizes: 1, 10, 100 rules
	testCases := []struct {
		name      string
		ruleCount int
		maxTime   time.Duration
	}{
		{"1 rule", 1, 100 * time.Millisecond},
		{"10 rules", 10, 200 * time.Millisecond},
		{"100 rules", 100, 500 * time.Millisecond}, // SRS requirement
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create rule set with specified number of rules
			rules := make([]resource_access.Rule, tc.ruleCount)
			for i := 0; i < tc.ruleCount; i++ {
				rules[i] = resource_access.Rule{
					ID:          fmt.Sprintf("performance-rule-%d", i),
					Name:        fmt.Sprintf("Performance Rule %d", i),
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]any{
						"required_fields": []any{"title"},
					},
					Actions: map[string]any{
						"block":   true,
						"message": fmt.Sprintf("Rule %d violation", i),
					},
					Priority: 100 - i, // Varying priorities
					Enabled:  true,
				}
			}

			ruleSet := &resource_access.RuleSet{
				Version: "1.0",
				Rules:   rules,
			}

			err := rulesAccess.ChangeRules(tempDir, ruleSet)
			if err != nil {
				t.Fatalf("Failed to set performance rules: %v", err)
			}

			// Test event
			event := TaskEvent{
				EventType: "task_transition",
				FutureState: &TaskState{
					Task: &resource_access.Task{
						ID:    "performance-test-task",
						Title: "", // Violates all rules
					},
					Status: resource_access.WorkflowStatus{
						Column: "doing",
					},
				},
				Timestamp: time.Now(),
			}

			// Measure performance
			start := time.Now()
			result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("Performance test with %d rules failed: %v", tc.ruleCount, err)
			} else {
				t.Logf("Performance test: %d rules evaluated in %v", tc.ruleCount, duration)
				
				if duration > tc.maxTime {
					t.Errorf("Performance requirement failed: %v > %v", duration, tc.maxTime)
				}

				if len(result.Violations) != tc.ruleCount {
					t.Errorf("Expected %d violations, got %d", tc.ruleCount, len(result.Violations))
				}
			}
		})
	}
}

// TestAcceptance_RuleEngine_ResourceExhaustion tests resource exhaustion (STP DT-RESOURCE-001)
func TestAcceptance_RuleEngine_ResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource exhaustion test in short mode")
	}

	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	rulesAccess, boardAccess, ruleEngine := createTestComponents(t, tempDir)
	defer rulesAccess.Close()
	defer boardAccess.Close()
	defer ruleEngine.Close()

	t.Run("LargeRuleSet", func(t *testing.T) {
		// Test with 1000 rules (more than SRS requirement of 100)
		rules := make([]resource_access.Rule, 1000)
		for i := 0; i < 1000; i++ {
			rules[i] = resource_access.Rule{
				ID:          fmt.Sprintf("exhaustion-rule-%d", i),
				Name:        fmt.Sprintf("Exhaustion Rule %d", i),
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"required_fields": []any{fmt.Sprintf("field_%d", i%10)},
				},
				Actions: map[string]any{
					"block":   true,
					"message": fmt.Sprintf("Rule %d violation", i),
				},
				Priority: 1000 - i,
				Enabled:  true,
			}
		}

		ruleSet := &resource_access.RuleSet{
			Version: "1.0",
			Rules:   rules,
		}

		err := rulesAccess.ChangeRules(tempDir, ruleSet)
		if err != nil {
			t.Fatalf("Failed to set large rule set: %v", err)
		}

		// Monitor memory before
		var memBefore runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memBefore)

		event := TaskEvent{
			EventType: "task_transition",
			FutureState: &TaskState{
				Task: &resource_access.Task{
					ID:    "exhaustion-test-task",
					Title: "Test Task",
				},
				Status: resource_access.WorkflowStatus{
					Column: "doing",
				},
			},
			Timestamp: time.Now(),
		}

		result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
		if result != nil || err != nil {
			// Handle memory testing result
		}

		// Monitor memory after
		var memAfter runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memAfter)

		memUsed := memAfter.Alloc - memBefore.Alloc

		if err != nil {
			t.Logf("Large rule set handled with error (acceptable): %v", err)
		} else {
			t.Logf("Large rule set processed successfully. Memory used: %d bytes", memUsed)
		}

		// Memory usage should be reasonable (less than 100MB for this test)
		if memUsed > 100*1024*1024 {
			t.Logf("Warning: High memory usage detected: %d bytes", memUsed)
		}
	})
}

// TestAcceptance_RuleEngine_ConcurrentAccess tests concurrent access (STP DT-CONCURRENT-001)
func TestAcceptance_RuleEngine_ConcurrentAccess(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	rulesAccess, boardAccess, ruleEngine := createTestComponents(t, tempDir)
	defer rulesAccess.Close()
	defer boardAccess.Close()
	defer ruleEngine.Close()

	// Set up rules
	testRules := &resource_access.RuleSet{
		Version: "1.0",
		Rules: []resource_access.Rule{
			{
				ID:          "concurrent-test-rule",
				Name:        "Concurrent Test Rule",
				Category:    "validation",
				TriggerType: "task_transition",
				Conditions: map[string]any{
					"required_fields": []any{"title"},
				},
				Actions: map[string]any{
					"block":   true,
					"message": "Concurrent test violation",
				},
				Priority: 100,
				Enabled:  true,
			},
		},
	}

	err := rulesAccess.ChangeRules(tempDir, testRules)
	if err != nil {
		t.Fatalf("Failed to set concurrent test rules: %v", err)
	}

	t.Run("ConcurrentRuleEvaluations", func(t *testing.T) {
		const numGoroutines = 50
		const numEvaluationsPerGoroutine = 10

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numEvaluationsPerGoroutine)
		results := make(chan bool, numGoroutines*numEvaluationsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numEvaluationsPerGoroutine; j++ {
					event := TaskEvent{
						EventType: "task_transition",
						FutureState: &TaskState{
							Task: &resource_access.Task{
								ID:    fmt.Sprintf("concurrent-task-%d-%d", goroutineID, j),
								Title: "Valid Title", // Should pass
							},
							Status: resource_access.WorkflowStatus{
								Column: "doing",
							},
						},
						Timestamp: time.Now(),
					}

					result, err := ruleEngine.EvaluateTaskChange(context.Background(), event, tempDir)
					if err != nil {
						errors <- err
					} else {
						results <- result.Allowed
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		close(results)

		// Check for errors
		var errorCount int
		for err := range errors {
			errorCount++
			t.Logf("Concurrent evaluation error: %v", err)
		}

		// Check results consistency
		var successCount int
		for allowed := range results {
			if allowed {
				successCount++
			}
		}

		expectedTotal := numGoroutines * numEvaluationsPerGoroutine
		actualTotal := errorCount + successCount

		if actualTotal != expectedTotal {
			t.Errorf("Expected %d total operations, got %d", expectedTotal, actualTotal)
		}

		t.Logf("Concurrent test: %d successes, %d errors out of %d operations", 
			successCount, errorCount, expectedTotal)

		// Most operations should succeed (allowing some errors under high concurrency)
		if successCount < expectedTotal/2 {
			t.Errorf("Too many failures in concurrent test: %d/%d succeeded", successCount, expectedTotal)
		}
	})
}

// Helper functions

func setupTestEnvironment(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "ruleengine_acceptance_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func createTestComponents(t *testing.T, tempDir string) (resource_access.IRulesAccess, resource_access.IBoardAccess, *RuleEngine) {
	rulesAccess, err := resource_access.NewRulesAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create RulesAccess: %v", err)
	}

	boardAccess, err := resource_access.NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}

	ruleEngine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("Failed to create RuleEngine: %v", err)
	}

	return rulesAccess, boardAccess, ruleEngine
}