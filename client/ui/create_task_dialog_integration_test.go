package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rknuus/eisenkan/client/engines"
)

// TestIntegration_CreateTaskDialog_CoreFunctionality tests core dialog functionality with real engines
func TestIntegration_CreateTaskDialog_CoreFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test dependencies with real engines
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Setup mock expectations for task loading (3 quadrants only)
	mockWM.On("Task").Return(&MockITask{mock: &mockWM.Mock})

	// Mock all 3 quadrant queries with specific responses
	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, map[string]any{
		"priority": "urgent-important",
		"status":   "todo",
	}).Return(map[string]any{
		"tasks": []interface{}{
			map[string]interface{}{
				"id":          "task-1",
				"title":       "Urgent Important Task",
				"description": "Test description",
				"priority":    "urgent-important",
				"status":      "todo",
				"metadata":    map[string]interface{}{},
			},
		},
	}, nil)

	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, map[string]any{
		"priority": "urgent-non-important",
		"status":   "todo",
	}).Return(map[string]any{
		"tasks": []interface{}{},
	}, nil)

	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, map[string]any{
		"priority": "non-urgent-important",
		"status":   "todo",
	}).Return(map[string]any{
		"tasks": []interface{}{},
	}, nil)

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Integration Test")

	// Create dialog
	dialog := NewCreateTaskDialog(
		mockWM,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	require.NotNil(t, dialog)

	// Show dialog (this triggers task loading)
	dialog.Show()

	// Wait for async task loading to complete
	time.Sleep(500 * time.Millisecond)

	// Verify tasks were loaded
	urgentImportantTasks := dialog.GetQuadrantTasks(DialogUrgentImportant)
	assert.Len(t, urgentImportantTasks, 1)
	assert.Equal(t, "task-1", urgentImportantTasks[0].ID)
	assert.Equal(t, "Urgent Important Task", urgentImportantTasks[0].Title)

	// Verify other quadrants are empty
	assert.Empty(t, dialog.GetQuadrantTasks(DialogUrgentNonImportant))
	assert.Empty(t, dialog.GetQuadrantTasks(DialogNonUrgentImportant))
	assert.Empty(t, dialog.GetQuadrantTasks(DialogNonUrgentNonImportant))

	// Cleanup
	dialog.cleanup()

	// Verify mock expectations
	mockWM.AssertExpectations(t)
}

// TestIntegration_CreateTaskDialog_TaskMovementWorkflow tests task movement between quadrants
func TestIntegration_CreateTaskDialog_TaskMovementWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test dependencies
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Setup mock expectations
	mockWM.On("Task").Return(&MockITask{mock: &mockWM.Mock})
	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, mock.Anything).Return(map[string]any{
		"tasks": []interface{}{},
	}, nil)

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Integration Test")

	// Create dialog
	dialog := NewCreateTaskDialog(
		mockWM,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	// Don't show dialog to avoid UI rendering issues in tests
	// Instead directly test the functionality
	time.Sleep(100 * time.Millisecond)

	// Create a task directly in dialog state
	testTask := &TaskData{
		ID:          "move-test-task",
		Title:       "Task to Move",
		Description: "Test task movement",
		Priority:    "non-urgent-non-important",
		Status:      "todo",
		Metadata:    map[string]interface{}{},
	}

	// Add task to creation quadrant state
	dialog.stateMu.Lock()
	dialog.currentState.CreatedTasks = append(dialog.currentState.CreatedTasks, testTask)
	dialog.stateMu.Unlock()

	// Verify task is in creation quadrant
	createdTasks := dialog.GetQuadrantTasks(DialogNonUrgentNonImportant)
	assert.Len(t, createdTasks, 1)

	// Track movement events
	movementCalled := false
	var movedTaskID, fromQuadrant, toQuadrant string
	dialog.SetOnTaskMoved(func(taskID, from, to string) {
		movementCalled = true
		movedTaskID = taskID
		fromQuadrant = from
		toQuadrant = to
	})

	// Move task to urgent important quadrant
	err := dialog.MoveTaskToQuadrant("move-test-task", DialogUrgentImportant, 0)
	assert.NoError(t, err)

	// Verify task was moved in state
	urgentTasks := dialog.GetQuadrantTasks(DialogUrgentImportant)
	assert.Len(t, urgentTasks, 1)
	assert.Equal(t, "move-test-task", urgentTasks[0].ID)
	assert.Equal(t, "urgent-important", urgentTasks[0].Priority)

	// Verify task was removed from creation quadrant
	createdTasks = dialog.GetQuadrantTasks(DialogNonUrgentNonImportant)
	assert.Empty(t, createdTasks)

	// Verify movement event was fired
	assert.True(t, movementCalled)
	assert.Equal(t, "move-test-task", movedTaskID)
	assert.Equal(t, string(DialogNonUrgentNonImportant), fromQuadrant)
	assert.Equal(t, string(DialogUrgentImportant), toQuadrant)

	// Verify movement was tracked for deferred execution
	dialog.stateMu.RLock()
	movements := dialog.currentState.TaskMovements
	dialog.stateMu.RUnlock()
	assert.Len(t, movements, 1)
	assert.Equal(t, "move-test-task", movements[0].TaskID)
	assert.Equal(t, DialogNonUrgentNonImportant, movements[0].FromQuadrant)
	assert.Equal(t, DialogUrgentImportant, movements[0].ToQuadrant)

	// Cleanup
	dialog.cleanup()
}

// TestIntegration_CreateTaskDialog_DeferredOperationsExecution tests deferred WorkflowManager operations
func TestIntegration_CreateTaskDialog_DeferredOperationsExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test dependencies
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Setup mock expectations for task loading
	mockWM.On("Task").Return(&MockITask{mock: &mockWM.Mock})
	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, mock.Anything).Return(map[string]any{
		"tasks": []interface{}{},
	}, nil)

	// Setup expectation for deferred update operation
	mockWM.Mock.On("UpdateTaskWorkflow", mock.Anything, "deferred-task", mock.MatchedBy(func(request map[string]any) bool {
		return request["priority"] == "urgent-important"
	})).Return(map[string]any{
		"id":       "deferred-task",
		"priority": "urgent-important",
	}, nil)

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Integration Test")

	// Create dialog
	dialog := NewCreateTaskDialog(
		mockWM,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	// Wait for initialization
	time.Sleep(100 * time.Millisecond)

	// Create a task movement to create a deferred operation
	testTask := &TaskData{
		ID:       "deferred-task",
		Title:    "Task for Deferred Operation",
		Priority: "non-urgent-non-important",
		Status:   "todo",
		Metadata: map[string]interface{}{},
	}

	// Add task to state and move it
	dialog.stateMu.Lock()
	dialog.currentState.CreatedTasks = append(dialog.currentState.CreatedTasks, testTask)
	dialog.stateMu.Unlock()

	err := dialog.MoveTaskToQuadrant("deferred-task", DialogUrgentImportant, 0)
	assert.NoError(t, err)

	// Simulate dialog closing to trigger deferred operations
	dialog.handleDialogClosed()

	// Verify deferred operation was executed
	mockWM.AssertExpectations(t)
}

// TestIntegration_CreateTaskDialog_EngineCoordination tests coordination between multiple engines
func TestIntegration_CreateTaskDialog_EngineCoordination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test dependencies with real engines
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Setup mock expectations
	mockWM.On("Task").Return(&MockITask{mock: &mockWM.Mock})
	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, mock.Anything).Return(map[string]any{
		"tasks": []interface{}{},
	}, nil)

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Integration Test")

	// Create dialog
	dialog := NewCreateTaskDialog(
		mockWM,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	require.NotNil(t, dialog)

	// Verify all engines are properly integrated
	assert.NotNil(t, dialog.workflowManager)
	assert.NotNil(t, dialog.formattingEngine)
	assert.NotNil(t, dialog.validationEngine)
	assert.NotNil(t, dialog.layoutEngine)
	assert.NotNil(t, dialog.dragDropEngine)

	// Test engine coordination through basic operations
	time.Sleep(100 * time.Millisecond)

	// Test FormattingEngine integration through task data mapping
	testTaskMap := map[string]interface{}{
		"id":          "format-test",
		"title":       "Test Task for Formatting",
		"description": "This tests formatting engine integration",
		"priority":    "non-urgent-non-important",
		"status":      "todo",
		"metadata":    map[string]interface{}{"tag": "test"},
	}

	taskData := dialog.mapResponseToTaskData(testTaskMap)
	assert.NotNil(t, taskData)
	assert.Equal(t, "format-test", taskData.ID)
	assert.Equal(t, "Test Task for Formatting", taskData.Title)

	// Verify ValidationEngine integration
	assert.NotNil(t, dialog.validationEngine)

	// Verify LayoutEngine integration through dialog structure
	assert.NotNil(t, dialog.matrixContainer)
	assert.Len(t, dialog.quadrantContainers, 4)

	// Verify DragDropEngine integration
	assert.NotNil(t, dialog.dragDropEngine)

	// Cleanup
	dialog.cleanup()
}

// TestIntegration_CreateTaskDialog_ErrorRecovery tests error handling and recovery scenarios
func TestIntegration_CreateTaskDialog_ErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test dependencies
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Setup mock to return error for task loading
	mockWM.On("Task").Return(&MockITask{mock: &mockWM.Mock})
	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, mock.Anything).Return(
		map[string]any{}, assert.AnError)

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Integration Test")

	// Create dialog
	dialog := NewCreateTaskDialog(
		mockWM,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	// Show dialog (this will trigger failed task loading)
	dialog.Show()
	time.Sleep(200 * time.Millisecond)

	// Verify dialog handles errors gracefully
	// All quadrants should be empty due to loading error
	assert.Empty(t, dialog.GetQuadrantTasks(DialogUrgentImportant))
	assert.Empty(t, dialog.GetQuadrantTasks(DialogUrgentNonImportant))
	assert.Empty(t, dialog.GetQuadrantTasks(DialogNonUrgentImportant))
	assert.Empty(t, dialog.GetQuadrantTasks(DialogNonUrgentNonImportant))

	// Verify dialog state is not corrupted
	dialog.stateMu.RLock()
	state := dialog.currentState
	dialog.stateMu.RUnlock()
	assert.NotNil(t, state)
	assert.False(t, state.IsLoading)

	// Test error recovery by manually refreshing
	mockWM.Mock.ExpectedCalls = nil // Clear previous expectations
	mockWM.Mock.On("QueryTasksWorkflow", mock.Anything, mock.Anything).Return(map[string]any{
		"tasks": []interface{}{},
	}, nil)

	// Test that refresh works after error
	dialog.RefreshQuadrants()
	time.Sleep(200 * time.Millisecond)

	// Dialog should now work normally
	assert.NotNil(t, dialog.currentState)

	// Cleanup
	dialog.cleanup()
}