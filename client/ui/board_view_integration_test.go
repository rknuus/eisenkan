package ui

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/client/managers"
)

// BoardViewMockWorkflowManager provides a test implementation of managers.WorkflowManager
type BoardViewMockWorkflowManager struct {
	taskResponses map[string]any
	dragResponses map[string]any
	callLog       []string
}

func NewBoardViewMockWorkflowManager() *BoardViewMockWorkflowManager {
	return &BoardViewMockWorkflowManager{
		taskResponses: make(map[string]any),
		dragResponses: make(map[string]any),
		callLog:       make([]string, 0),
	}
}

func (m *BoardViewMockWorkflowManager) Task() managers.ITask {
	return &mockTaskWorkflows{manager: m}
}

func (m *BoardViewMockWorkflowManager) Drag() managers.IDrag {
	return &mockDragWorkflows{manager: m}
}

func (m *BoardViewMockWorkflowManager) Batch() managers.IBatch {
	return &mockBatchWorkflows{manager: m}
}

func (m *BoardViewMockWorkflowManager) Search() managers.ISearch {
	return &mockSearchWorkflows{manager: m}
}

func (m *BoardViewMockWorkflowManager) Subtask() managers.ISubtask {
	return &mockSubtaskWorkflows{manager: m}
}

// Mock task workflows
type mockTaskWorkflows struct {
	manager *BoardViewMockWorkflowManager
}

func (m *mockTaskWorkflows) CreateTaskWorkflow(ctx context.Context, request map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "CreateTaskWorkflow")
	return m.manager.taskResponses, nil
}

func (m *mockTaskWorkflows) UpdateTaskWorkflow(ctx context.Context, taskID string, request map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "UpdateTaskWorkflow")
	return m.manager.taskResponses, nil
}

func (m *mockTaskWorkflows) DeleteTaskWorkflow(ctx context.Context, taskID string) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "DeleteTaskWorkflow")
	return m.manager.taskResponses, nil
}

func (m *mockTaskWorkflows) QueryTasksWorkflow(ctx context.Context, criteria map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "QueryTasksWorkflow")

	// Return sample tasks for integration testing
	response := map[string]any{
		"tasks": []interface{}{
			map[string]interface{}{
				"id":          "task-1",
				"title":       "Urgent Important Task",
				"description": "Test task 1",
				"priority":    "urgent important",
				"status":      "todo",
				"created_at":  time.Now(),
				"updated_at":  time.Now(),
				"metadata":    map[string]interface{}{},
			},
			map[string]interface{}{
				"id":          "task-2",
				"title":       "Urgent Non-Important Task",
				"description": "Test task 2",
				"priority":    "urgent non-important",
				"status":      "todo",
				"created_at":  time.Now(),
				"updated_at":  time.Now(),
				"metadata":    map[string]interface{}{},
			},
			map[string]interface{}{
				"id":          "task-3",
				"title":       "Non-Urgent Important Task",
				"description": "Test task 3",
				"priority":    "non-urgent important",
				"status":      "todo",
				"created_at":  time.Now(),
				"updated_at":  time.Now(),
				"metadata":    map[string]interface{}{},
			},
		},
	}

	return response, nil
}

func (m *mockTaskWorkflows) ChangeTaskStatusWorkflow(ctx context.Context, taskID string, status string) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ChangeTaskStatusWorkflow")
	return m.manager.taskResponses, nil
}

func (m *mockTaskWorkflows) ChangeTaskPriorityWorkflow(ctx context.Context, taskID string, priority string) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ChangeTaskPriorityWorkflow")
	return m.manager.taskResponses, nil
}

func (m *mockTaskWorkflows) ArchiveTaskWorkflow(ctx context.Context, taskID string, options map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ArchiveTaskWorkflow")
	return m.manager.taskResponses, nil
}

// Mock drag workflows
type mockDragWorkflows struct {
	manager *BoardViewMockWorkflowManager
}

func (m *mockDragWorkflows) ProcessDragDropWorkflow(ctx context.Context, event map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ProcessDragDropWorkflow")

	// Return updated task for integration testing
	response := map[string]any{
		"updated_task": map[string]interface{}{
			"id":          event["task_id"],
			"title":       "Updated Task",
			"description": "Task moved between columns",
			"priority":    "non-urgent important", // Changed priority
			"status":      "todo",
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
			"metadata":    map[string]interface{}{},
		},
	}

	return response, nil
}

// Mock batch, search, and subtask workflows (minimal implementation)
type mockBatchWorkflows struct {
	manager *BoardViewMockWorkflowManager
}

func (m *mockBatchWorkflows) BatchStatusUpdateWorkflow(ctx context.Context, taskIDs []string, status string) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

func (m *mockBatchWorkflows) BatchPriorityUpdateWorkflow(ctx context.Context, taskIDs []string, priority string) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

func (m *mockBatchWorkflows) BatchArchiveWorkflow(ctx context.Context, taskIDs []string, options map[string]any) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

type mockSearchWorkflows struct {
	manager *BoardViewMockWorkflowManager
}

func (m *mockSearchWorkflows) SearchTasksWorkflow(ctx context.Context, query string, filters map[string]any) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

func (m *mockSearchWorkflows) ApplyFiltersWorkflow(ctx context.Context, filters map[string]any, context map[string]any) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

type mockSubtaskWorkflows struct {
	manager *BoardViewMockWorkflowManager
}

func (m *mockSubtaskWorkflows) CreateSubtaskRelationshipWorkflow(ctx context.Context, parentID string, childSpec map[string]any) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

func (m *mockSubtaskWorkflows) ProcessSubtaskCompletionWorkflow(ctx context.Context, subtaskID string, cascade map[string]any) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

func (m *mockSubtaskWorkflows) MoveSubtaskWorkflow(ctx context.Context, subtaskID string, newParentID string, position map[string]any) (map[string]any, error) {
	return m.manager.taskResponses, nil
}

// Integration Tests



// TestIntegration_BoardView_TaskMovement verifies task movement workflow integration
func TestIntegration_BoardView_TaskMovement(t *testing.T) {
	mockWM := NewBoardViewMockWorkflowManager()
	board := NewBoardView(mockWM, nil, nil)
	defer board.Destroy()

	// Load tasks first
	board.LoadBoard()
	time.Sleep(50 * time.Millisecond)

	// Test task movement without triggering complex UI operations
	err := board.MoveTask("task-1", 0, 1)
	if err != nil {
		t.Errorf("Task movement failed: %v", err)
	}

	// Verify workflow manager was called
	found := false
	for _, call := range mockWM.callLog {
		if call == "ProcessDragDropWorkflow" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected ProcessDragDropWorkflow to be called during task movement")
	}
}



// TestIntegration_BoardView_EventHandlers verifies event handler coordination
func TestIntegration_BoardView_EventHandlers(t *testing.T) {
	mockWM := NewBoardViewMockWorkflowManager()
	board := NewBoardView(mockWM, nil, nil)
	defer board.Destroy()

	// Test event handler integration without complex UI operations
	var (
		errorCalled         bool
		configChangedCalled bool
	)

	board.SetOnError(func(err error) {
		errorCalled = true
	})

	board.SetOnConfigChanged(func(config *BoardConfiguration) {
		configChangedCalled = true
	})

	// Test error handler
	board.SetError(fmt.Errorf("test error"))
	if !errorCalled {
		t.Error("Expected error handler to be called")
	}

	// Test configuration change handler
	newConfig := &BoardConfiguration{
		Title:     "Updated Board",
		BoardType: "kanban",
		Columns: []*ColumnConfiguration{
			{Title: "Test Column", Type: TodoColumn},
		},
	}
	board.SetBoardConfiguration(newConfig)
	if !configChangedCalled {
		t.Error("Expected config change handler to be called")
	}
}

// TestIntegration_BoardView_StateConsistency verifies state consistency during operations
func TestIntegration_BoardView_StateConsistency(t *testing.T) {
	mockWM := NewBoardViewMockWorkflowManager()
	board := NewBoardView(mockWM, nil, nil)
	defer board.Destroy()

	// Test state consistency during multiple operations
	initialState := board.GetBoardState()
	if initialState == nil {
		t.Fatal("Initial state is nil")
	}

	// Load board and verify state transition
	board.LoadBoard()

	// Wait for async operation to complete by polling the state
	var loadedState *BoardState
	for i := 0; i < 40; i++ { // Wait up to 2 seconds
		time.Sleep(50 * time.Millisecond)
		loadedState = board.GetBoardState()
		// Exit if loading is complete (either success or error)
		if !loadedState.IsLoading {
			break
		}
	}

	if loadedState.IsLoading {
		t.Error("Expected loading state to be false after load completion")
	}

	// Set error and verify state consistency
	testErr := fmt.Errorf("test error")
	board.SetError(testErr)
	errorState := board.GetBoardState()
	if !errorState.HasError {
		t.Error("Expected error state to be true")
	}
	if errorState.ErrorMessage != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", errorState.ErrorMessage)
	}

	// Clear error and verify state
	board.SetError(nil)
	clearedState := board.GetBoardState()
	if clearedState.HasError {
		t.Error("Expected error state to be false after clearing")
	}
	if clearedState.ErrorMessage != "" {
		t.Error("Expected error message to be cleared")
	}
}