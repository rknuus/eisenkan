package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rknuus/eisenkan/client/engines"
)

// TestUnit_CreateTaskDialog_NewCreateTaskDialog tests basic dialog creation
func TestUnit_CreateTaskDialog_NewCreateTaskDialog(t *testing.T) {
	// Create test dependencies
	workflowManager := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")

	// Create CreateTaskDialog
	dialog := NewCreateTaskDialog(
		workflowManager,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	// Verify dialog was created
	require.NotNil(t, dialog)
	assert.NotNil(t, dialog.workflowManager)
	assert.NotNil(t, dialog.formattingEngine)
	assert.NotNil(t, dialog.validationEngine)
	assert.NotNil(t, dialog.layoutEngine)
	assert.NotNil(t, dialog.dragDropEngine)
	assert.NotNil(t, dialog.parentWindow)
	assert.NotNil(t, dialog.currentState)
	assert.NotNil(t, dialog.quadrantContainers)
	assert.NotNil(t, dialog.taskWidgets)

	// Verify initial state
	assert.False(t, dialog.currentState.IsLoading)
	assert.False(t, dialog.currentState.HasError)
	assert.Empty(t, dialog.currentState.UrgentImportantTasks)
	assert.Empty(t, dialog.currentState.UrgentNonImportantTasks)
	assert.Empty(t, dialog.currentState.NonUrgentImportantTasks)
	assert.Empty(t, dialog.currentState.CreatedTasks)
	assert.Empty(t, dialog.currentState.TaskMovements)

	// Verify quadrant containers exist
	assert.Len(t, dialog.quadrantContainers, 4)
	assert.Contains(t, dialog.quadrantContainers, DialogUrgentImportant)
	assert.Contains(t, dialog.quadrantContainers, DialogUrgentNonImportant)
	assert.Contains(t, dialog.quadrantContainers, DialogNonUrgentImportant)
	assert.Contains(t, dialog.quadrantContainers, DialogNonUrgentNonImportant)
}

// TestUnit_CreateTaskDialog_GetQuadrantTasks tests quadrant task retrieval
func TestUnit_CreateTaskDialog_GetQuadrantTasks(t *testing.T) {
	// Create test dependencies
	workflowManager := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")

	// Create dialog
	dialog := NewCreateTaskDialog(
		workflowManager,
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	// Test retrieving tasks from empty quadrants
	tasks := dialog.GetQuadrantTasks(DialogUrgentImportant)
	assert.Empty(t, tasks)

	tasks = dialog.GetQuadrantTasks(DialogUrgentNonImportant)
	assert.Empty(t, tasks)

	tasks = dialog.GetQuadrantTasks(DialogNonUrgentImportant)
	assert.Empty(t, tasks)

	tasks = dialog.GetQuadrantTasks(DialogNonUrgentNonImportant)
	assert.Empty(t, tasks)

	// Test invalid quadrant
	tasks = dialog.GetQuadrantTasks(DialogQuadrant("invalid"))
	assert.Empty(t, tasks)
}

// TestUnit_CreateTaskDialog_GracefulDegradation_NoWorkflowManager tests graceful handling of missing WorkflowManager
func TestUnit_CreateTaskDialog_GracefulDegradation_NoWorkflowManager(t *testing.T) {
	// Create test dependencies (nil WorkflowManager)
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()
	dragDropEngine := engines.NewDragDropEngine()

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")

	// Create dialog with nil WorkflowManager
	dialog := NewCreateTaskDialog(
		nil, // nil WorkflowManager
		formattingEngine,
		validationEngine,
		layoutEngine,
		&dragDropEngine,
		window,
	)

	// Verify dialog was created successfully
	require.NotNil(t, dialog)
	assert.Nil(t, dialog.workflowManager)

	// Test that operations handle nil WorkflowManager gracefully
	assert.NotPanics(t, func() {
		dialog.RefreshQuadrants()
	})
}

// TestUnit_CreateTaskDialog_GracefulDegradation_NoDragDropEngine tests graceful handling of missing DragDropEngine
func TestUnit_CreateTaskDialog_GracefulDegradation_NoDragDropEngine(t *testing.T) {
	// Create test dependencies (nil DragDropEngine)
	workflowManager := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	layoutEngine := engines.NewLayoutEngine()

	// Create test window
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Test")

	// Create dialog with nil DragDropEngine
	dialog := NewCreateTaskDialog(
		workflowManager,
		formattingEngine,
		validationEngine,
		layoutEngine,
		nil, // nil DragDropEngine
		window,
	)

	// Verify dialog was created successfully
	require.NotNil(t, dialog)
	assert.Nil(t, dialog.dragDropEngine)

	// Test that drag operations handle nil DragDropEngine gracefully
	testWidget := NewDisplayTaskWidget(workflowManager, formattingEngine, validationEngine, &TaskData{
		ID:    "test-123",
		Title: "Test Task",
	})

	assert.NotPanics(t, func() {
		dialog.setupTaskWidgetForDrag(testWidget, DialogUrgentImportant)
	})
}