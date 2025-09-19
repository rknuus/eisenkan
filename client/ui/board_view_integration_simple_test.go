package ui

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// SimpleMockWorkflowManager provides a simple test implementation without UI interactions
type SimpleMockWorkflowManager struct {
	callLog []string
}

func NewSimpleMockWorkflowManager() *SimpleMockWorkflowManager {
	return &SimpleMockWorkflowManager{
		callLog: make([]string, 0),
	}
}

func (m *SimpleMockWorkflowManager) Task() managers.ITask {
	return &simpleTaskWorkflows{manager: m}
}

func (m *SimpleMockWorkflowManager) Drag() managers.IDrag {
	return &simpleDragWorkflows{manager: m}
}

func (m *SimpleMockWorkflowManager) Batch() managers.IBatch {
	return &simpleBatchWorkflows{manager: m}
}

func (m *SimpleMockWorkflowManager) Search() managers.ISearch {
	return &simpleSearchWorkflows{manager: m}
}

func (m *SimpleMockWorkflowManager) Subtask() managers.ISubtask {
	return &simpleSubtaskWorkflows{manager: m}
}

// Simple implementations that don't trigger UI
type simpleTaskWorkflows struct {
	manager *SimpleMockWorkflowManager
}

func (m *simpleTaskWorkflows) CreateTaskWorkflow(ctx context.Context, request map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "CreateTaskWorkflow")
	return map[string]any{}, nil
}

func (m *simpleTaskWorkflows) UpdateTaskWorkflow(ctx context.Context, taskID string, request map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "UpdateTaskWorkflow")
	return map[string]any{}, nil
}

func (m *simpleTaskWorkflows) DeleteTaskWorkflow(ctx context.Context, taskID string) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "DeleteTaskWorkflow")
	return map[string]any{}, nil
}

func (m *simpleTaskWorkflows) QueryTasksWorkflow(ctx context.Context, criteria map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "QueryTasksWorkflow")

	// Return simple tasks without triggering UI
	return map[string]any{
		"tasks": []interface{}{
			map[string]interface{}{
				"id":          "task-1",
				"title":       "Test Task 1",
				"description": "Test task",
				"priority":    "urgent important",
				"status":      "todo",
			},
		},
	}, nil
}

func (m *simpleTaskWorkflows) ChangeTaskStatusWorkflow(ctx context.Context, taskID string, status string) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ChangeTaskStatusWorkflow")
	return map[string]any{}, nil
}

func (m *simpleTaskWorkflows) ChangeTaskPriorityWorkflow(ctx context.Context, taskID string, priority string) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ChangeTaskPriorityWorkflow")
	return map[string]any{}, nil
}

func (m *simpleTaskWorkflows) ArchiveTaskWorkflow(ctx context.Context, taskID string, options map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ArchiveTaskWorkflow")
	return map[string]any{}, nil
}

type simpleDragWorkflows struct {
	manager *SimpleMockWorkflowManager
}

func (m *simpleDragWorkflows) ProcessDragDropWorkflow(ctx context.Context, event map[string]any) (map[string]any, error) {
	m.manager.callLog = append(m.manager.callLog, "ProcessDragDropWorkflow")
	return map[string]any{
		"updated_task": map[string]interface{}{
			"id":       event["task_id"],
			"priority": "non-urgent important",
		},
	}, nil
}

type simpleBatchWorkflows struct {
	manager *SimpleMockWorkflowManager
}

func (m *simpleBatchWorkflows) BatchStatusUpdateWorkflow(ctx context.Context, taskIDs []string, status string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (m *simpleBatchWorkflows) BatchPriorityUpdateWorkflow(ctx context.Context, taskIDs []string, priority string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (m *simpleBatchWorkflows) BatchArchiveWorkflow(ctx context.Context, taskIDs []string, options map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

type simpleSearchWorkflows struct {
	manager *SimpleMockWorkflowManager
}

func (m *simpleSearchWorkflows) SearchTasksWorkflow(ctx context.Context, query string, filters map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

func (m *simpleSearchWorkflows) ApplyFiltersWorkflow(ctx context.Context, filters map[string]any, context map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

type simpleSubtaskWorkflows struct {
	manager *SimpleMockWorkflowManager
}

func (m *simpleSubtaskWorkflows) CreateSubtaskRelationshipWorkflow(ctx context.Context, parentID string, childSpec map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

func (m *simpleSubtaskWorkflows) ProcessSubtaskCompletionWorkflow(ctx context.Context, subtaskID string, cascade map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

func (m *simpleSubtaskWorkflows) MoveSubtaskWorkflow(ctx context.Context, subtaskID string, newParentID string, position map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

// Simple Integration Tests (Avoiding UI race conditions)

// TestSimpleIntegration_BoardView_BasicWorkflowIntegration verifies basic workflow integration
func TestSimpleIntegration_BoardView_BasicWorkflowIntegration(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test that board was created successfully
	if board == nil {
		t.Fatal("BoardView creation failed")
	}

	// Test basic state
	state := board.GetBoardState()
	if state == nil {
		t.Fatal("GetBoardState returned nil")
	}

	// Test configuration
	if state.Configuration == nil {
		t.Fatal("Configuration is nil")
	}

	if state.Configuration.Title != "Eisenhower Matrix" {
		t.Errorf("Expected title 'Eisenhower Matrix', got '%s'", state.Configuration.Title)
	}

	// Test column creation
	if len(state.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(state.Columns))
	}
}

// TestSimpleIntegration_BoardView_WorkflowManagerCalls verifies WorkflowManager method calls
func TestSimpleIntegration_BoardView_WorkflowManagerCalls(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test LoadBoard calls WorkflowManager
	board.LoadBoard()

	// Give some time for async operation without triggering UI
	time.Sleep(50 * time.Millisecond)

	// Check that WorkflowManager was called
	found := false
	for _, call := range mockWM.callLog {
		if call == "QueryTasksWorkflow" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected QueryTasksWorkflow to be called during LoadBoard")
	}
}

// TestSimpleIntegration_BoardView_TaskMovementValidation verifies task movement validation
func TestSimpleIntegration_BoardView_TaskMovementValidation(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test validation with invalid parameters
	err := board.MoveTask("", -1, -1)
	if err == nil {
		t.Error("Expected validation error for invalid parameters")
	}

	// Test validation with valid parameters but non-existent task
	err = board.MoveTask("valid-task-id", 0, 1)
	if err != nil {
		// This may fail due to workflow execution, but validation should pass
		t.Logf("Task movement failed (expected for non-existent task): %v", err)
	}

	// Check that ProcessDragDropWorkflow was called for valid movement attempt
	found := false
	for _, call := range mockWM.callLog {
		if call == "ProcessDragDropWorkflow" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected ProcessDragDropWorkflow to be called during valid task movement")
	}
}

// TestSimpleIntegration_BoardView_ValidationEngineIntegration verifies FormValidationEngine integration
func TestSimpleIntegration_BoardView_ValidationEngineIntegration(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test that validation engine is properly integrated
	if board.validationEngine == nil {
		t.Error("Validation engine not properly integrated")
	}

	// Test validation during task movement
	err := board.MoveTask("", 0, 1) // Empty task ID should fail validation
	if err == nil {
		t.Error("Expected validation error for empty task ID")
	}

	// Test with invalid column indices
	err = board.MoveTask("test-task", -1, 0)
	if err == nil {
		t.Error("Expected error for invalid from column index")
	}

	err = board.MoveTask("test-task", 0, 10)
	if err == nil {
		t.Error("Expected error for invalid to column index")
	}
}

// TestSimpleIntegration_BoardView_StateManagement verifies state management during operations
func TestSimpleIntegration_BoardView_StateManagement(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test initial state
	initialState := board.GetBoardState()
	if initialState.HasError {
		t.Error("Initial state should not have error")
	}
	if initialState.IsLoading {
		t.Error("Initial state should not be loading")
	}

	// Test loading state
	board.SetLoading(true)
	loadingState := board.GetBoardState()
	if !loadingState.IsLoading {
		t.Error("Expected loading state to be true")
	}

	board.SetLoading(false)
	notLoadingState := board.GetBoardState()
	if notLoadingState.IsLoading {
		t.Error("Expected loading state to be false")
	}

	// Test error state
	testError := fmt.Errorf("test error")
	board.SetError(testError)
	errorState := board.GetBoardState()
	if !errorState.HasError {
		t.Error("Expected error state to be true")
	}
	if errorState.ErrorMessage != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", errorState.ErrorMessage)
	}

	// Clear error
	board.SetError(nil)
	clearedState := board.GetBoardState()
	if clearedState.HasError {
		t.Error("Expected error state to be false after clearing")
	}
}

// TestSimpleIntegration_BoardView_EventHandlerRegistration verifies event handler registration
func TestSimpleIntegration_BoardView_EventHandlerRegistration(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test event handler registration (should not panic)
	var handlersCalled int

	board.SetOnTaskMoved(func(taskID, fromColumn, toColumn string) {
		handlersCalled++
	})

	board.SetOnTaskSelected(func(taskID string) {
		handlersCalled++
	})

	board.SetOnBoardRefreshed(func() {
		handlersCalled++
	})

	board.SetOnError(func(err error) {
		handlersCalled++
	})

	board.SetOnConfigChanged(func(config *BoardConfiguration) {
		handlersCalled++
	})

	// Test error handler
	board.SetError(fmt.Errorf("test error"))
	if handlersCalled == 0 {
		t.Error("Expected at least one event handler to be called")
	}
}

// TestSimpleIntegration_BoardView_ConfigurationManagement verifies configuration management
func TestSimpleIntegration_BoardView_ConfigurationManagement(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	mockWM := NewSimpleMockWorkflowManager()

	board := NewBoardView(mockWM, validationEngine, nil)
	defer board.Destroy()

	// Test initial configuration
	initialState := board.GetBoardState()
	if initialState.Configuration.BoardType != "eisenhower" {
		t.Errorf("Expected initial board type 'eisenhower', got '%s'", initialState.Configuration.BoardType)
	}

	// Test configuration change
	newConfig := &BoardConfiguration{
		Title:     "Custom Board",
		BoardType: "kanban",
		Columns: []*ColumnConfiguration{
			{Title: "To Do", Type: TodoColumn},
			{Title: "In Progress", Type: DoingColumn},
			{Title: "Done", Type: DoneColumn},
		},
	}

	board.SetBoardConfiguration(newConfig)

	// Verify configuration was updated
	updatedState := board.GetBoardState()
	if updatedState.Configuration.Title != "Custom Board" {
		t.Errorf("Expected updated title 'Custom Board', got '%s'", updatedState.Configuration.Title)
	}
	if updatedState.Configuration.BoardType != "kanban" {
		t.Errorf("Expected updated board type 'kanban', got '%s'", updatedState.Configuration.BoardType)
	}
	if len(updatedState.Configuration.Columns) != 3 {
		t.Errorf("Expected 3 columns after update, got %d", len(updatedState.Configuration.Columns))
	}
}