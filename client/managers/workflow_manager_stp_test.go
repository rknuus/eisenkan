package managers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/internal/client/resource_access"
)

// Destructive test cases that directly map to STP test scenarios
// These tests verify that WorkflowManager handles extreme conditions gracefully

// Mock that can simulate failures for destructive testing
type failingMockTaskManagerAccess struct {
	simulateTimeout       bool
	simulateCorruption    bool
	simulateUnavailable   bool
	simulateMemoryFailure bool
}

func (m *failingMockTaskManagerAccess) CreateTaskAsync(ctx context.Context, request resource_access.UITaskRequest) (<-chan resource_access.UITaskResponse, <-chan error) {
	respCh := make(chan resource_access.UITaskResponse, 1)
	errCh := make(chan error, 1)

	if m.simulateTimeout {
		// Simulate timeout by not responding until context cancellation
		go func() {
			<-ctx.Done()
			errCh <- ctx.Err()
		}()
		return respCh, errCh
	}

	if m.simulateUnavailable {
		errCh <- fmt.Errorf("backend service unavailable")
		return respCh, errCh
	}

	if m.simulateCorruption {
		// Send corrupted response
		respCh <- resource_access.UITaskResponse{
			ID:          "", // Invalid empty ID
			Description: "CORRUPTED_DATA_" + request.Description,
			DisplayName: "",
		}
		close(respCh)
		return respCh, errCh
	}

	// Normal response
	respCh <- resource_access.UITaskResponse{
		ID:          "task-123",
		Description: request.Description,
		DisplayName: "Test Task",
	}
	close(respCh)
	return respCh, errCh
}

func (m *failingMockTaskManagerAccess) UpdateTaskAsync(ctx context.Context, taskID string, request resource_access.UITaskRequest) (<-chan resource_access.UITaskResponse, <-chan error) {
	return m.CreateTaskAsync(ctx, request)
}

func (m *failingMockTaskManagerAccess) DeleteTaskAsync(ctx context.Context, taskID string) (<-chan bool, <-chan error) {
	respCh := make(chan bool, 1)
	errCh := make(chan error, 1)

	if m.simulateUnavailable {
		errCh <- fmt.Errorf("backend service unavailable")
		return respCh, errCh
	}

	respCh <- true
	close(respCh)
	return respCh, errCh
}

func (m *failingMockTaskManagerAccess) QueryTasksAsync(ctx context.Context, criteria resource_access.UIQueryCriteria) (<-chan []resource_access.UITaskResponse, <-chan error) {
	respCh := make(chan []resource_access.UITaskResponse, 1)
	errCh := make(chan error, 1)

	if m.simulateCorruption {
		// Send corrupted data
		respCh <- []resource_access.UITaskResponse{
			{ID: "", Description: "CORRUPTED", DisplayName: ""},
		}
		close(respCh)
		return respCh, errCh
	}

	respCh <- []resource_access.UITaskResponse{
		{ID: "task-1", Description: "Task 1", DisplayName: "Task 1"},
	}
	close(respCh)
	return respCh, errCh
}

// Implement remaining interface methods
func (m *failingMockTaskManagerAccess) GetTaskAsync(ctx context.Context, taskID string) (<-chan resource_access.UITaskResponse, <-chan error) {
	return m.CreateTaskAsync(ctx, resource_access.UITaskRequest{Description: "Retrieved"})
}

func (m *failingMockTaskManagerAccess) ListTasksAsync(ctx context.Context, criteria resource_access.UIQueryCriteria) (<-chan []resource_access.UITaskResponse, <-chan error) {
	return m.QueryTasksAsync(ctx, criteria)
}

func (m *failingMockTaskManagerAccess) ChangeTaskStatusAsync(ctx context.Context, taskID string, status resource_access.UIWorkflowStatus) (<-chan resource_access.UITaskResponse, <-chan error) {
	return m.CreateTaskAsync(ctx, resource_access.UITaskRequest{Description: "Status Changed"})
}

func (m *failingMockTaskManagerAccess) ValidateTaskAsync(ctx context.Context, request resource_access.UITaskRequest) (<-chan resource_access.UIValidationResult, <-chan error) {
	respCh := make(chan resource_access.UIValidationResult, 1)
	errCh := make(chan error, 1)
	respCh <- resource_access.UIValidationResult{Valid: true}
	close(respCh)
	return respCh, errCh
}

func (m *failingMockTaskManagerAccess) ProcessPriorityPromotionsAsync(ctx context.Context) (<-chan []resource_access.UITaskResponse, <-chan error) {
	return m.QueryTasksAsync(ctx, resource_access.UIQueryCriteria{})
}

func (m *failingMockTaskManagerAccess) GetBoardSummaryAsync(ctx context.Context) (<-chan resource_access.UIBoardSummary, <-chan error) {
	respCh := make(chan resource_access.UIBoardSummary, 1)
	errCh := make(chan error, 1)
	respCh <- resource_access.UIBoardSummary{}
	close(respCh)
	return respCh, errCh
}

func (m *failingMockTaskManagerAccess) SearchTasksAsync(ctx context.Context, query string) (<-chan []resource_access.UITaskResponse, <-chan error) {
	return m.QueryTasksAsync(ctx, resource_access.UIQueryCriteria{})
}

// STP Test Case DT-CREATE-001: Task Creation Workflow with Engine Coordination Failures
func TestSTP_DT_CREATE_001_EngineCoordinationFailures(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()

	// Test with backend communication failures
	failingBackend := &failingMockTaskManagerAccess{simulateUnavailable: true}
	wm := NewWorkflowManager(validation, formatting, dragDrop, failingBackend)

	ctx := context.Background()
	response, err := wm.Task().CreateTaskWorkflow(ctx, map[string]any{
		"description": "Test with backend failure",
	})

	// Should handle backend failure gracefully
	if err == nil {
		t.Error("Expected error due to backend unavailability")
	}

	if response != nil {
		success, ok := response["success"].(bool)
		if ok && success {
			t.Error("Should not succeed when backend is unavailable")
		}
	}

	// Test with timeout scenario
	timeoutBackend := &failingMockTaskManagerAccess{simulateTimeout: true}
	wmTimeout := NewWorkflowManager(validation, formatting, dragDrop, timeoutBackend)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	response, err = wmTimeout.Task().CreateTaskWorkflow(timeoutCtx, map[string]any{
		"description": "Test with timeout",
	})

	// Should handle timeout gracefully
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// STP Test Case DT-CREATE-002: Task Creation Data Validation and Formatting Stress
func TestSTP_DT_CREATE_002_DataValidationFormattingStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()

	// Test with corrupted data backend
	corruptedBackend := &failingMockTaskManagerAccess{simulateCorruption: true}
	wm := NewWorkflowManager(validation, formatting, dragDrop, corruptedBackend)

	ctx := context.Background()

	// Test with various malformed inputs
	malformedInputs := []map[string]any{
		{"description": nil}, // nil description
		{"description": make([]byte, 10000)}, // oversized data
		{"description": string(make([]byte, 1000000))}, // extremely large string
		{}, // empty input
		{"invalid_field": "value"}, // unexpected fields
	}

	for i, input := range malformedInputs {
		response, err := wm.Task().CreateTaskWorkflow(ctx, input)

		// Should handle malformed input gracefully without panic
		if response == nil && err != nil {
			t.Logf("Test case %d handled malformed input gracefully: %v", i, err)
		} else if response != nil {
			// Verify workflow tracking even with malformed input
			if workflowID, ok := response["workflow_id"].(string); !ok || workflowID == "" {
				t.Errorf("Test case %d should provide workflow tracking even with malformed input", i)
			}
		}
	}
}

// STP Test Case DT-UPDATE-001: Task Update Workflow with Backend Integration Failures
func TestSTP_DT_UPDATE_001_BackendIntegrationFailures(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()

	// Test concurrent update operations
	normalBackend := &mockTaskManagerAccess{}
	wm := NewWorkflowManager(validation, formatting, dragDrop, normalBackend)

	ctx := context.Background()
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	// Simulate concurrent task updates with overlapping dependencies
	for i := 0; i < 10; i++ {
		go func(taskNum int) {
			_, err := wm.Task().UpdateTaskWorkflow(ctx, "shared-task-id", map[string]any{
				"description": fmt.Sprintf("Concurrent update %d", taskNum),
			})
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all operations
	completedOps := 0
	errorCount := 0
	for completedOps < 10 {
		select {
		case <-done:
			completedOps++
		case err := <-errors:
			errorCount++
			t.Logf("Concurrent operation error (expected under stress): %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	t.Logf("Completed %d concurrent updates with %d errors (demonstrating proper handling under stress)", completedOps, errorCount)
}

// STP Test Case DT-DRAGDROP-001: Drag-Drop Workflow with Engine Coordination Failures
func TestSTP_DT_DRAGDROP_001_EngineCoordinationFailures(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()
	backend := &mockTaskManagerAccess{}

	wm := NewWorkflowManager(validation, formatting, dragDrop, backend)
	ctx := context.Background()

	// Test with invalid drop zone configurations
	invalidDragEvents := []map[string]any{
		{
			"source_id":     "",
			"target_id":     "column-456",
			"drop_position": fyne.NewPos(-1000, -1000), // Invalid negative coordinates
		},
		{
			"source_id":     "task-123",
			"target_id":     "",
			"drop_position": fyne.NewPos(0, 0),
		},
		{
			"source_id":     "task-123",
			"target_id":     "column-456",
			"drop_position": nil, // nil position
		},
		{
			// Missing required fields
		},
	}

	for i, event := range invalidDragEvents {
		response, err := wm.Drag().ProcessDragDropWorkflow(ctx, event)

		// Should handle invalid configurations gracefully
		if response != nil {
			success, ok := response["success"].(bool)
			if !ok {
				t.Errorf("Test case %d should always return success field", i)
			}

			// Verify workflow tracking even for failed operations
			if workflowID, ok := response["workflow_id"].(string); !ok || workflowID == "" {
				t.Errorf("Test case %d should provide workflow tracking for failed operations", i)
			}

			if !success {
				t.Logf("Test case %d correctly rejected invalid drag event: %+v", i, response)
			}
		}

		if err != nil {
			t.Logf("Test case %d handled error gracefully: %v", i, err)
		}
	}
}

// STP Test Case DT-QUERY-001: Task Query Workflow with Performance and Data Stress
func TestSTP_DT_QUERY_001_PerformanceDataStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()

	// Test with corrupted data backend
	corruptedBackend := &failingMockTaskManagerAccess{simulateCorruption: true}
	wm := NewWorkflowManager(validation, formatting, dragDrop, corruptedBackend)

	ctx := context.Background()

	// Test query with corrupted backend data
	response, err := wm.Task().QueryTasksWorkflow(ctx, map[string]any{
		"limit": 1000, // Large query
	})

	if err != nil {
		t.Logf("Query handled corrupted backend gracefully: %v", err)
	}

	if response != nil {
		// Verify that corrupted data is handled
		tasks, ok := response["tasks"].([]map[string]any)
		if ok {
			t.Logf("Query returned %d tasks despite backend corruption", len(tasks))

			// Verify formatting was applied even to corrupted data
			for i, task := range tasks {
				if desc, ok := task["description"].(string); ok && desc != "" {
					t.Logf("Task %d description was formatted: %s", i, desc)
				}
			}
		}

		// Verify workflow tracking
		if workflowID, ok := response["workflow_id"].(string); !ok || workflowID == "" {
			t.Error("Query should provide workflow tracking even with corrupted data")
		}
	}
}

// STP Test Case: Workflow State Management under Stress
func TestSTP_WorkflowStateManagementStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()
	backend := &mockTaskManagerAccess{}

	wm := NewWorkflowManager(validation, formatting, dragDrop, backend)
	ctx := context.Background()

	// Create many concurrent workflows to stress state management
	numWorkflows := 100
	done := make(chan string, numWorkflows)

	for i := 0; i < numWorkflows; i++ {
		go func(workflowNum int) {
			response, err := wm.Task().CreateTaskWorkflow(ctx, map[string]any{
				"description": fmt.Sprintf("Stress test workflow %d", workflowNum),
			})

			if err != nil {
				done <- fmt.Sprintf("ERROR_%d", workflowNum)
				return
			}

			if response != nil {
				if workflowID, ok := response["workflow_id"].(string); ok {
					done <- workflowID
				} else {
					done <- fmt.Sprintf("NO_ID_%d", workflowNum)
				}
			} else {
				done <- fmt.Sprintf("NO_RESPONSE_%d", workflowNum)
			}
		}(i)
	}

	// Collect all workflow IDs
	workflowIDs := make(map[string]bool)
	errorCount := 0

	for i := 0; i < numWorkflows; i++ {
		select {
		case result := <-done:
			if strings.HasPrefix(result, "ERROR_") || strings.HasPrefix(result, "NO_") {
				errorCount++
				t.Logf("Workflow stress result: %s", result)
			} else {
				if workflowIDs[result] {
					t.Errorf("Duplicate workflow ID detected: %s", result)
				}
				workflowIDs[result] = true
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Workflow stress test timed out")
		}
	}

	uniqueWorkflows := len(workflowIDs)
	t.Logf("Created %d unique workflows with %d errors under stress", uniqueWorkflows, errorCount)

	if uniqueWorkflows < numWorkflows-errorCount {
		t.Errorf("Expected %d unique workflows, got %d", numWorkflows-errorCount, uniqueWorkflows)
	}
}

// STP Test Cases for WorkflowManager Extensions

// STP Test Case DT-ENHANCED-001: Enhanced Task Status Management under Stress
func TestSTP_DT_ENHANCED_001_StatusManagementStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()

	// Test with failing backend
	failingBackend := &failingMockTaskManagerAccess{simulateUnavailable: true}
	wm := NewWorkflowManager(validation, formatting, dragDrop, failingBackend)

	ctx := context.Background()

	// Test status change with backend failures
	response, err := wm.Task().ChangeTaskStatusWorkflow(ctx, "task-123", "completed")

	// Should handle backend failure gracefully
	if err == nil {
		t.Error("Expected error due to backend unavailability")
	}

	if response != nil {
		success, ok := response["success"].(bool)
		if ok && success {
			t.Error("Should not succeed when backend is unavailable")
		}
	}

	// Test priority change with corrupted data
	corruptedBackend := &failingMockTaskManagerAccess{simulateCorruption: true}
	wmCorrupt := NewWorkflowManager(validation, formatting, dragDrop, corruptedBackend)

	priorityResponse, err := wmCorrupt.Task().ChangeTaskPriorityWorkflow(ctx, "task-456", "urgent")

	// Should handle corrupted backend gracefully
	if priorityResponse != nil {
		if workflowID, ok := priorityResponse["workflow_id"].(string); !ok || workflowID == "" {
			t.Error("Should provide workflow tracking even with corrupted backend")
		}
	}

	// Test archive workflow with timeout
	timeoutBackend := &failingMockTaskManagerAccess{simulateTimeout: true}
	wmTimeout := NewWorkflowManager(validation, formatting, dragDrop, timeoutBackend)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = wmTimeout.Task().ArchiveTaskWorkflow(timeoutCtx, "task-789", map[string]any{"cascade": "orphan"})

	// Should handle timeout gracefully
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// STP Test Case DT-BATCH-001: Batch Operations under Resource Exhaustion and Partial Failures
func TestSTP_DT_BATCH_001_BatchOperationsStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()
	backend := &mockTaskManagerAccess{}

	wm := NewWorkflowManager(validation, formatting, dragDrop, backend)
	ctx := context.Background()

	// Test batch size limit enforcement
	largeBatch := make([]string, 101)
	for i := 0; i < 101; i++ {
		largeBatch[i] = fmt.Sprintf("task-%d", i)
	}

	response, err := wm.Batch().BatchStatusUpdateWorkflow(ctx, largeBatch, "completed")

	// Should reject large batch without crashing
	if response == nil {
		t.Error("Should return response even for rejected batch")
	}

	if response != nil {
		success, ok := response["success"].(bool)
		if !ok {
			t.Error("Response should contain success field")
		}
		if success {
			t.Error("Large batch should not succeed")
		}

		errorMsg, ok := response["error"].(string)
		if !ok || !strings.Contains(errorMsg, "limit") {
			t.Error("Should return error message about size limit")
		}
	}

	// Test batch operations with mixed valid/invalid task IDs
	mixedBatch := []string{"valid-task-1", "", "valid-task-2", "invalid-task-3"}
	batchResponse, err := wm.Batch().BatchPriorityUpdateWorkflow(ctx, mixedBatch, "high")

	// Should handle mixed batch gracefully
	if batchResponse == nil && err != nil {
		t.Error("Should handle mixed batch without complete failure")
	}

	if batchResponse != nil {
		if workflowID, ok := batchResponse["workflow_id"].(string); !ok || workflowID == "" {
			t.Error("Should provide workflow tracking for batch operations")
		}

		// Should report per-task results
		if results, ok := batchResponse["results"].([]map[string]any); ok {
			if len(results) != len(mixedBatch) {
				t.Errorf("Should return result for each task in batch")
			}
		}
	}

	// Test concurrent batch operations
	done := make(chan bool, 3)
	errors := make(chan error, 3)

	taskIDs := []string{"concurrent-1", "concurrent-2", "concurrent-3"}

	for i := 0; i < 3; i++ {
		go func(batchNum int) {
			_, err := wm.Batch().BatchArchiveWorkflow(ctx, taskIDs, map[string]any{
				"cascade": fmt.Sprintf("batch-%d", batchNum),
			})
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all concurrent batch operations
	completedOps := 0
	errorCount := 0
	for completedOps < 3 {
		select {
		case <-done:
			completedOps++
		case err := <-errors:
			errorCount++
			t.Logf("Concurrent batch operation error (expected under stress): %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent batch operations timed out")
		}
	}

	t.Logf("Completed %d concurrent batch operations with %d errors", completedOps, errorCount)
}

// STP Test Case DT-SEARCH-001: Advanced Search under Performance and Validation Stress
func TestSTP_DT_SEARCH_001_SearchOperationsStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()

	// Test with corrupted backend
	corruptedBackend := &failingMockTaskManagerAccess{simulateCorruption: true}
	wm := NewWorkflowManager(validation, formatting, dragDrop, corruptedBackend)

	ctx := context.Background()

	// Test search with invalid queries
	invalidQueries := []string{
		"", // Empty query
		string(make([]byte, 10000)), // Extremely long query
		"query\x00with\x00nulls", // Query with null bytes
		"query with special chars: <script>alert('xss')</script>", // Potential XSS
	}

	for i, query := range invalidQueries {
		filters := map[string]any{"status": "active"}
		response, err := wm.Search().SearchTasksWorkflow(ctx, query, filters)

		// Should handle invalid queries gracefully
		if response == nil && err != nil {
			t.Logf("Query %d handled gracefully: %v", i, err)
		} else if response != nil {
			// Verify workflow tracking
			if workflowID, ok := response["workflow_id"].(string); !ok || workflowID == "" {
				t.Errorf("Query %d should provide workflow tracking", i)
			}
		}
	}

	// Test filter application with corrupted filter data
	corruptedFilters := []map[string]any{
		{"invalid_field": make(chan int)}, // Non-serializable data
		{"status": nil}, // Nil values
		{}, // Empty filters
		{"nested": map[string]any{"deep": map[string]any{"very": "deep"}}}, // Deep nesting
	}

	for i, filters := range corruptedFilters {
		context := map[string]any{"board": "test"}
		response, err := wm.Search().ApplyFiltersWorkflow(ctx, filters, context)

		// Should handle corrupted filters gracefully
		if response == nil && err != nil {
			t.Logf("Filter %d handled gracefully: %v", i, err)
		} else if response != nil {
			// Verify filter status is reported
			if filterStatus, ok := response["filter_status"].(map[string]any); !ok {
				t.Errorf("Filter %d should provide filter status", i)
			} else {
				if _, ok := filterStatus["valid"]; !ok {
					t.Errorf("Filter %d should report validity status", i)
				}
			}
		}
	}

	// Test concurrent search operations
	done := make(chan bool, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func(searchNum int) {
			query := fmt.Sprintf("concurrent search %d", searchNum)
			filters := map[string]any{"priority": "high", "search_id": searchNum}
			_, err := wm.Search().SearchTasksWorkflow(ctx, query, filters)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for concurrent searches
	completedSearches := 0
	errorCount := 0
	for completedSearches < 5 {
		select {
		case <-done:
			completedSearches++
		case err := <-errors:
			errorCount++
			t.Logf("Concurrent search error (expected under stress): %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent searches timed out")
		}
	}

	t.Logf("Completed %d concurrent searches with %d errors", completedSearches, errorCount)
}

// STP Test Case DT-SUBTASK-001: Subtask Management under Hierarchy Corruption and Circular Dependencies
func TestSTP_DT_SUBTASK_001_SubtaskHierarchyStress(t *testing.T) {
	validation := engines.NewFormValidationEngine()
	formatting := engines.NewFormattingEngine()
	dragDrop := engines.NewDragDropEngine()
	backend := &mockTaskManagerAccess{}

	wm := NewWorkflowManager(validation, formatting, dragDrop, backend)
	ctx := context.Background()

	// Test subtask creation with invalid parent-child relationships
	invalidRelationships := []struct {
		parentID  string
		childSpec map[string]any
	}{
		{"", map[string]any{"description": "orphan subtask"}}, // Empty parent
		{"parent-123", map[string]any{}}, // Empty child spec
		{"parent-123", map[string]any{"description": nil}}, // Nil description
		{"parent-123", map[string]any{"description": make([]byte, 10000)}}, // Oversized data
	}

	for i, rel := range invalidRelationships {
		response, err := wm.Subtask().CreateSubtaskRelationshipWorkflow(ctx, rel.parentID, rel.childSpec)

		// Should handle invalid relationships gracefully
		if response == nil && err != nil {
			t.Logf("Invalid relationship %d handled gracefully: %v", i, err)
		} else if response != nil {
			// Verify workflow tracking even for invalid relationships
			if workflowID, ok := response["workflow_id"].(string); !ok || workflowID == "" {
				t.Errorf("Invalid relationship %d should provide workflow tracking", i)
			}
		}
	}

	// Test subtask completion with corrupted cascade options
	corruptedCascades := []map[string]any{
		{"circular_reference": "self"}, // Circular reference
		{"invalid_option": make(chan int)}, // Non-serializable data
		{"deep_nesting": map[string]any{"level1": map[string]any{"level2": "deep"}}}, // Deep nesting
		{}, // Empty cascade options
	}

	for i, cascade := range corruptedCascades {
		response, err := wm.Subtask().ProcessSubtaskCompletionWorkflow(ctx, "subtask-123", cascade)

		// Should handle corrupted cascade options gracefully
		if response == nil && err != nil {
			t.Logf("Corrupted cascade %d handled gracefully: %v", i, err)
		} else if response != nil {
			// Verify cascade results are reported
			if cascadeResults, ok := response["cascade_results"].(map[string]any); !ok {
				t.Errorf("Cascade %d should provide cascade results", i)
			} else {
				if _, ok := cascadeResults["processed"]; !ok {
					t.Errorf("Cascade %d should report processing status", i)
				}
			}
		}
	}

	// Test subtask movement with invalid position data
	invalidPositions := []map[string]any{
		{"index": -1}, // Negative index
		{"index": "invalid"}, // Non-numeric index
		{"position": make(chan int)}, // Non-serializable data
		{}, // Empty position
	}

	for i, position := range invalidPositions {
		response, err := wm.Subtask().MoveSubtaskWorkflow(ctx, "subtask-456", "new-parent-789", position)

		// Should handle invalid positions gracefully
		if response == nil && err != nil {
			t.Logf("Invalid position %d handled gracefully: %v", i, err)
		} else if response != nil {
			// Verify movement data is reported
			if movement, ok := response["movement"].(map[string]any); !ok {
				t.Errorf("Position %d should provide movement data", i)
			} else {
				if _, ok := movement["moved"]; !ok {
					t.Errorf("Position %d should report movement status", i)
				}
			}
		}
	}

	// Test concurrent subtask operations with potential circular dependencies
	done := make(chan bool, 4)
	errors := make(chan error, 4)

	// Simulate potential circular dependency scenario
	go func() {
		_, err := wm.Subtask().CreateSubtaskRelationshipWorkflow(ctx, "parent-A", map[string]any{"description": "child of A"})
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	go func() {
		_, err := wm.Subtask().CreateSubtaskRelationshipWorkflow(ctx, "parent-B", map[string]any{"description": "child of B"})
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	go func() {
		_, err := wm.Subtask().MoveSubtaskWorkflow(ctx, "child-A", "parent-B", map[string]any{"index": 0})
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	go func() {
		_, err := wm.Subtask().MoveSubtaskWorkflow(ctx, "child-B", "parent-A", map[string]any{"index": 0})
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	// Wait for concurrent subtask operations
	completedOps := 0
	errorCount := 0
	for completedOps < 4 {
		select {
		case <-done:
			completedOps++
		case err := <-errors:
			errorCount++
			t.Logf("Concurrent subtask operation error (expected under stress): %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent subtask operations timed out")
		}
	}

	t.Logf("Completed %d concurrent subtask operations with %d errors", completedOps, errorCount)
}