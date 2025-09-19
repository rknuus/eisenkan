package ui

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// Mock WorkflowManager for testing
type MockWorkflowManager struct {
	mock.Mock
}

func (m *MockWorkflowManager) Task() managers.ITask {
	return MockITask{mock: &m.Mock}
}

func (m *MockWorkflowManager) Drag() managers.IDrag {
	return MockIDrag{mock: &m.Mock}
}

func (m *MockWorkflowManager) Batch() managers.IBatch {
	return MockIBatch{mock: &m.Mock}
}

func (m *MockWorkflowManager) Search() managers.ISearch {
	return MockISearch{mock: &m.Mock}
}

func (m *MockWorkflowManager) Subtask() managers.ISubtask {
	return MockISubtask{mock: &m.Mock}
}

type MockITask struct {
	mock *mock.Mock
}

func (m MockITask) CreateTaskWorkflow(ctx context.Context, request map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, request)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockITask) UpdateTaskWorkflow(ctx context.Context, taskID string, request map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, taskID, request)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockITask) DeleteTaskWorkflow(ctx context.Context, taskID string) (map[string]any, error) {
	args := m.mock.Called(ctx, taskID)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockITask) QueryTasksWorkflow(ctx context.Context, criteria map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, criteria)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockITask) ChangeTaskStatusWorkflow(ctx context.Context, taskID string, status string) (map[string]any, error) {
	args := m.mock.Called(ctx, taskID, status)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockITask) ChangeTaskPriorityWorkflow(ctx context.Context, taskID string, priority string) (map[string]any, error) {
	args := m.mock.Called(ctx, taskID, priority)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockITask) ArchiveTaskWorkflow(ctx context.Context, taskID string, options map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, taskID, options)
	return args.Get(0).(map[string]any), args.Error(1)
}

type MockIDrag struct {
	mock *mock.Mock
}

func (m MockIDrag) ProcessDragDropWorkflow(ctx context.Context, event map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, event)
	return args.Get(0).(map[string]any), args.Error(1)
}

type MockIBatch struct {
	mock *mock.Mock
}

func (m MockIBatch) BatchStatusUpdateWorkflow(ctx context.Context, taskIDs []string, status string) (map[string]any, error) {
	args := m.mock.Called(ctx, taskIDs, status)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockIBatch) BatchPriorityUpdateWorkflow(ctx context.Context, taskIDs []string, priority string) (map[string]any, error) {
	args := m.mock.Called(ctx, taskIDs, priority)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockIBatch) BatchArchiveWorkflow(ctx context.Context, taskIDs []string, options map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, taskIDs, options)
	return args.Get(0).(map[string]any), args.Error(1)
}

type MockISearch struct {
	mock *mock.Mock
}

func (m MockISearch) SearchTasksWorkflow(ctx context.Context, query string, filters map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, query, filters)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockISearch) ApplyFiltersWorkflow(ctx context.Context, filters map[string]any, context map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, filters, context)
	return args.Get(0).(map[string]any), args.Error(1)
}

type MockISubtask struct {
	mock *mock.Mock
}

func (m MockISubtask) CreateSubtaskRelationshipWorkflow(ctx context.Context, parentID string, childSpec map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, parentID, childSpec)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockISubtask) ProcessSubtaskCompletionWorkflow(ctx context.Context, subtaskID string, cascade map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, subtaskID, cascade)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m MockISubtask) MoveSubtaskWorkflow(ctx context.Context, subtaskID string, newParentID string, position map[string]any) (map[string]any, error) {
	args := m.mock.Called(ctx, subtaskID, newParentID, position)
	return args.Get(0).(map[string]any), args.Error(1)
}

// Test Data Helper
func createTestTaskData() *TaskData {
	return &TaskData{
		ID:          "test-task-123",
		Title:       "Test Task",
		Description: "This is a test task description",
		Priority:    "urgent-important",
		Status:      "todo",
		Metadata: map[string]interface{}{
			"category": "testing",
			"tags":     []string{"unit", "test"},
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

// Test setup helper
func setupTestApp() {
	test.NewApp()
}

// Add setupTestApp to all test functions
func addSetupToAllTests(testFunc func(*testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		setupTestApp()
		testFunc(t)
	}
}

// Unit Tests

func TestUnit_TaskWidget_NewTaskWidget(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()

	// Execute
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Verify
	assert.NotNil(t, widget)
	assert.Equal(t, taskData, widget.GetTaskData())
	assert.False(t, widget.IsSelected())
	assert.NotNil(t, widget.workflowManager)
	assert.NotNil(t, widget.formattingEngine)
	assert.NotNil(t, widget.validationEngine)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_SetTaskData(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	initialData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), initialData, DisplayMode)

	// Execute - update task data
	newData := &TaskData{
		ID:          "updated-task-456",
		Title:       "Updated Task",
		Description: "Updated description",
		Priority:    "not-urgent-important",
		Status:      "doing",
	}
	widget.SetTaskData(newData)

	// Give time for state update
	time.Sleep(200 * time.Millisecond)

	// Verify
	retrievedData := widget.GetTaskData()
	assert.Equal(t, newData.ID, retrievedData.ID)
	assert.Equal(t, newData.Title, retrievedData.Title)
	assert.Equal(t, newData.Priority, retrievedData.Priority)
	assert.Equal(t, newData.Status, retrievedData.Status)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_SetSelected(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	selectionChangeCount := 0
	var lastSelectedState bool
	widget.SetOnSelectionChange(func(selected bool) {
		selectionChangeCount++
		lastSelectedState = selected
	})

	// Execute
	widget.SetSelected(true)

	// Give time for state update
	time.Sleep(200 * time.Millisecond)

	// Verify
	assert.True(t, widget.IsSelected())
	assert.Equal(t, 1, selectionChangeCount)
	assert.True(t, lastSelectedState)

	// Test deselection
	widget.SetSelected(false)
	time.Sleep(200 * time.Millisecond)
	assert.False(t, widget.IsSelected())
	assert.Equal(t, 2, selectionChangeCount)
	assert.False(t, lastSelectedState)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_SetLoading(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Execute
	widget.SetLoading(true)

	// Give time for state update
	time.Sleep(200 * time.Millisecond)

	// Verify
	widget.stateMu.RLock()
	isLoading := widget.currentState.IsLoading
	widget.stateMu.RUnlock()

	assert.True(t, isLoading)

	// Test clearing loading
	widget.SetLoading(false)
	time.Sleep(200 * time.Millisecond)

	widget.stateMu.RLock()
	isLoading = widget.currentState.IsLoading
	widget.stateMu.RUnlock()

	assert.False(t, isLoading)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_SetError(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	errorOccurred := false
	widget.SetOnError(func(err error) {
		errorOccurred = true
		assert.Contains(t, err.Error(), "assert.AnError")
	})

	// Execute
	testError := assert.AnError
	widget.SetError(testError)

	// Give time for state update
	time.Sleep(200 * time.Millisecond)

	// Verify
	widget.stateMu.RLock()
	hasError := widget.currentState.HasError
	errorMessage := widget.currentState.ErrorMessage
	widget.stateMu.RUnlock()

	assert.True(t, hasError)
	assert.Contains(t, errorMessage, "assert.AnError")
	assert.True(t, errorOccurred)

	// Test clearing error
	widget.SetError(nil)
	time.Sleep(200 * time.Millisecond)

	widget.stateMu.RLock()
	hasError = widget.currentState.HasError
	widget.stateMu.RUnlock()

	assert.False(t, hasError)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_SetValidationErrors(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Execute
	validationErrors := map[string]string{
		"title": "Field is required",
		"priority": "Input does not match required pattern",
	}
	widget.SetValidationErrors(validationErrors)

	// Give time for state update
	time.Sleep(200 * time.Millisecond)

	// Verify
	widget.stateMu.RLock()
	currentErrors := widget.currentState.ValidationErrs
	widget.stateMu.RUnlock()

	assert.Len(t, currentErrors, 2)
	assert.Equal(t, "Field is required", currentErrors["title"])
	assert.Equal(t, "Input does not match required pattern", currentErrors["priority"])

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_CompactMode(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Verify default state
	assert.False(t, widget.compact)

	// Execute
	widget.SetCompactMode(true)

	// Verify
	assert.True(t, widget.compact)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_ShowMetadata(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Verify default state
	assert.True(t, widget.showMetadata)

	// Execute
	widget.SetShowMetadata(false)

	// Verify
	assert.False(t, widget.showMetadata)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_MinSize(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Execute & Verify
	minSize := widget.MinSize()
	assert.Equal(t, float32(200), minSize.Width)
	assert.Equal(t, float32(80), minSize.Height)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_GracefulDegradation_NoWorkflowManager(t *testing.T) {
	// Setup - nil WorkflowManager
	setupTestApp()
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(nil, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Execute - should not crash
	widget.handleEditMode()
	widget.handleDragComplete()

	// Give time for any async operations
	time.Sleep(200 * time.Millisecond)

	// Verify - widget still functions
	assert.Equal(t, taskData, widget.GetTaskData())

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_GracefulDegradation_NoFormattingEngine(t *testing.T) {
	// Setup - nil FormattingEngine
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, nil, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Execute - should not crash when formatting
	title, description, _ := widget.formatTaskDisplay()

	// Verify - fallback formatting works
	assert.Equal(t, taskData.Title, title)
	assert.Equal(t, taskData.Description, description)
	// metadata should be formatted as string representation

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_StateManagement_ConcurrentUpdates(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Execute - concurrent state updates
	go widget.SetSelected(true)
	go widget.SetLoading(true)
	go widget.SetError(assert.AnError)

	// Give time for state updates
	time.Sleep(200 * time.Millisecond)

	// Verify - no race conditions or crashes
	assert.NotNil(t, widget.GetTaskData())

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_Lifecycle_Destroy(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	taskData := createTestTaskData()
	widget := NewTaskWidget(mockWM, formattingEngine, engines.NewFormValidationEngine(), taskData, DisplayMode)

	// Verify initial state
	assert.NotNil(t, widget.stateChannel)
	assert.NotNil(t, widget.cancel)

	// Execute
	widget.Destroy()

	// Verify cleanup
	assert.Nil(t, widget.stateChannel)

	// Should not crash if called multiple times
	widget.Destroy()
}

// Enhanced TaskWidget Tests for Creation and Editing Modes

func TestUnit_TaskWidget_CreationMode(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()

	// Execute - create widget in creation mode
	widget := NewTaskWidget(mockWM, formattingEngine, validationEngine, nil, CreateMode)

	// Verify
	assert.NotNil(t, widget)
	assert.Nil(t, widget.GetTaskData()) // No task data in creation mode

	widget.stateMu.RLock()
	mode := widget.currentState.Mode
	widget.stateMu.RUnlock()

	assert.Equal(t, CreateMode, mode)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_EditMode(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	taskData := createTestTaskData()

	// Execute - create widget in edit mode
	widget := NewTaskWidget(mockWM, formattingEngine, validationEngine, taskData, EditMode)

	// Verify
	assert.NotNil(t, widget)
	assert.Equal(t, taskData, widget.GetTaskData())

	widget.stateMu.RLock()
	mode := widget.currentState.Mode
	widget.stateMu.RUnlock()

	assert.Equal(t, EditMode, mode)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_ConvenienceConstructors(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	taskData := createTestTaskData()

	// Test display widget constructor
	displayWidget := NewDisplayTaskWidget(mockWM, formattingEngine, validationEngine, taskData)
	assert.NotNil(t, displayWidget)

	displayWidget.stateMu.RLock()
	displayMode := displayWidget.currentState.Mode
	displayWidget.stateMu.RUnlock()
	assert.Equal(t, DisplayMode, displayMode)

	// Test creation widget constructor
	creationWidget := NewCreationTaskWidget(mockWM, formattingEngine, validationEngine)
	assert.NotNil(t, creationWidget)

	creationWidget.stateMu.RLock()
	createMode := creationWidget.currentState.Mode
	creationWidget.stateMu.RUnlock()
	assert.Equal(t, CreateMode, createMode)

	// Cleanup
	displayWidget.Destroy()
	creationWidget.Destroy()
}

func TestUnit_TaskWidget_ModeTransitions(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	taskData := createTestTaskData()

	widget := NewTaskWidget(mockWM, formattingEngine, validationEngine, taskData, DisplayMode)

	// Test transition to edit mode
	err := widget.EnterEditMode()
	assert.NoError(t, err)

	// Wait for state update to complete
	time.Sleep(200 * time.Millisecond)

	widget.stateMu.RLock()
	mode := widget.currentState.Mode
	widget.stateMu.RUnlock()
	assert.Equal(t, EditMode, mode)

	// Test transition back to display mode
	err = widget.ExitEditMode()
	assert.NoError(t, err)

	// Wait for state update to complete
	time.Sleep(200 * time.Millisecond)

	widget.stateMu.RLock()
	mode = widget.currentState.Mode
	widget.stateMu.RUnlock()
	assert.Equal(t, DisplayMode, mode)

	// Test transition to creation mode
	err = widget.EnterCreateMode()
	assert.NoError(t, err)

	// Wait for state update to complete
	time.Sleep(200 * time.Millisecond)

	widget.stateMu.RLock()
	mode = widget.currentState.Mode
	widget.stateMu.RUnlock()
	assert.Equal(t, CreateMode, mode)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_FormDataManagement(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()

	widget := NewTaskWidget(mockWM, formattingEngine, validationEngine, nil, CreateMode)

	// Test form data initialization
	widget.stateMu.Lock()
	widget.currentState.FormData = map[string]interface{}{
		"title":       "Test Task",
		"description": "Test Description",
		"priority":    "urgent important",
	}
	widget.currentState.IsFormDirty = true
	widget.stateMu.Unlock()

	// Verify form data is stored
	widget.stateMu.RLock()
	formData := widget.currentState.FormData
	isDirty := widget.currentState.IsFormDirty
	widget.stateMu.RUnlock()

	assert.NotNil(t, formData)
	assert.Equal(t, "Test Task", formData["title"])
	assert.Equal(t, "Test Description", formData["description"])
	assert.Equal(t, "urgent important", formData["priority"])
	assert.True(t, isDirty)

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_ValidationIntegration(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()

	widget := NewTaskWidget(mockWM, formattingEngine, validationEngine, nil, CreateMode)

	// Test validation error setting
	validationErrors := map[string]string{
		"title": "Field is required",
		"priority": "Input does not match required pattern",
	}

	widget.SetValidationErrors(validationErrors)

	// Give time for state update
	time.Sleep(200 * time.Millisecond)

	// Verify validation errors are stored
	widget.stateMu.RLock()
	currentErrors := widget.currentState.ValidationErrs
	canSave := widget.currentState.CanSave
	widget.stateMu.RUnlock()

	assert.Len(t, currentErrors, 2)
	assert.Equal(t, "Field is required", currentErrors["title"])
	assert.Equal(t, "Input does not match required pattern", currentErrors["priority"])
	assert.False(t, canSave) // Should not be able to save with validation errors

	// Cleanup
	widget.Destroy()
}

func TestUnit_TaskWidget_EventHandlers(t *testing.T) {
	// Setup
	setupTestApp()
	mockWM := &MockWorkflowManager{}
	formattingEngine := engines.NewFormattingEngine()
	validationEngine := engines.NewFormValidationEngine()
	taskData := createTestTaskData()

	widget := NewTaskWidget(mockWM, formattingEngine, validationEngine, taskData, DisplayMode)

	// Test event handler setup
	var taskCreatedCalled bool
	var taskUpdatedCalled bool
	var editCancelledCalled bool

	widget.SetOnTaskCreated(func(data *TaskData) {
		taskCreatedCalled = true
	})

	widget.SetOnTaskUpdated(func(data *TaskData) {
		taskUpdatedCalled = true
	})

	widget.SetOnEditCancelled(func() {
		editCancelledCalled = true
	})

	// Simulate events (these would normally be called internally)
	if widget.onTaskCreated != nil {
		widget.onTaskCreated(taskData)
	}
	if widget.onTaskUpdated != nil {
		widget.onTaskUpdated(taskData)
	}
	if widget.onEditCancelled != nil {
		widget.onEditCancelled()
	}

	// Verify event handlers were called
	assert.True(t, taskCreatedCalled)
	assert.True(t, taskUpdatedCalled)
	assert.True(t, editCancelledCalled)

	// Cleanup
	widget.Destroy()
}