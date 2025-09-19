package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rknuus/eisenkan/client/engines"
)

// Note: MockWorkflowManager and related mocks are defined in task_widget_test.go

// Mock DragDropEngine for testing
type MockDragDropEngine struct {
	mock.Mock
}

func (m *MockDragDropEngine) Drag() engines.IDrag {
	return MockIDragEngine{mock: &m.Mock}
}

func (m *MockDragDropEngine) Drop() engines.IDrop {
	return MockIDropEngine{mock: &m.Mock}
}

func (m *MockDragDropEngine) Visualize() engines.IVisualize {
	return MockIVisualizeEngine{mock: &m.Mock}
}

type MockIDragEngine struct {
	mock *mock.Mock
}

func (m MockIDragEngine) StartDrag(widget fyne.CanvasObject, params engines.DragParams) (engines.DragHandle, error) {
	args := m.mock.Called(widget, params)
	return args.Get(0).(engines.DragHandle), args.Error(1)
}

func (m MockIDragEngine) UpdateDragPosition(handle engines.DragHandle, position fyne.Position) error {
	args := m.mock.Called(handle, position)
	return args.Error(0)
}

func (m MockIDragEngine) CompleteDrag(handle engines.DragHandle) (engines.DropResult, error) {
	args := m.mock.Called(handle)
	return args.Get(0).(engines.DropResult), args.Error(1)
}

func (m MockIDragEngine) CancelDrag(handle engines.DragHandle) error {
	args := m.mock.Called(handle)
	return args.Error(0)
}

type MockIDropEngine struct {
	mock *mock.Mock
}

func (m MockIDropEngine) RegisterDropZone(zone engines.DropZoneSpec) (engines.ZoneID, error) {
	args := m.mock.Called(zone)
	return args.Get(0).(engines.ZoneID), args.Error(1)
}

func (m MockIDropEngine) UnregisterDropZone(zoneID engines.ZoneID) error {
	args := m.mock.Called(zoneID)
	return args.Error(0)
}

func (m MockIDropEngine) ValidateDropTarget(position fyne.Position, dragContext engines.DragContext) (bool, error) {
	args := m.mock.Called(position, dragContext)
	return args.Bool(0), args.Error(1)
}

func (m MockIDropEngine) GetActiveZones() []engines.DropZoneSpec {
	args := m.mock.Called()
	return args.Get(0).([]engines.DropZoneSpec)
}

type MockIVisualizeEngine struct {
	mock *mock.Mock
}

func (m MockIVisualizeEngine) CreateDragIndicator(source fyne.CanvasObject) (fyne.CanvasObject, error) {
	args := m.mock.Called(source)
	return args.Get(0).(fyne.CanvasObject), args.Error(1)
}

func (m MockIVisualizeEngine) UpdateIndicatorPosition(indicator fyne.CanvasObject, position fyne.Position) error {
	args := m.mock.Called(indicator, position)
	return args.Error(0)
}

func (m MockIVisualizeEngine) ShowDropFeedback(zoneID engines.ZoneID, accepted bool) error {
	args := m.mock.Called(zoneID, accepted)
	return args.Error(0)
}

func (m MockIVisualizeEngine) CleanupVisuals(dragHandle engines.DragHandle) error {
	args := m.mock.Called(dragHandle)
	return args.Error(0)
}

// Test data helpers
func createTestColumnConfiguration(colType ColumnType) *ColumnConfiguration {
	return &ColumnConfiguration{
		Title:        "Test Column",
		Type:         colType,
		WIPLimit:     5,
		Color:        "#blue",
		ShowSections: colType == TodoColumn,
		SortOrder:    "created_desc",
		Metadata:     make(map[string]interface{}),
	}
}

func createTestTasksCollection() []*TaskData {
	return []*TaskData{
		{
			ID:          "task-1",
			Title:       "First Task",
			Description: "Description 1",
			Priority:    "urgent-important",
			Status:      "todo",
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
			Metadata:    make(map[string]interface{}),
		},
		{
			ID:          "task-2",
			Title:       "Second Task",
			Description: "Description 2",
			Priority:    "not-urgent-important",
			Status:      "todo",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now(),
			Metadata:    make(map[string]interface{}),
		},
	}
}

// Unit Tests

func TestUnit_ColumnWidget_NewColumnWidget(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(TodoColumn)

	// Mock drop zone registration and unregistration
	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	// Execute
	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)

	// Verify
	assert.NotNil(t, widget)
	assert.Equal(t, config, widget.GetConfiguration())
	assert.Empty(t, widget.GetTasks())
	assert.False(t, widget.IsSelected())
	assert.NotNil(t, widget.workflowManager)
	assert.NotNil(t, widget.dragDropEngine)
	assert.NotNil(t, widget.layoutEngine)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_SetTasks(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoingColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)
	tasks := createTestTasksCollection()

	// Execute
	widget.SetTasks(tasks)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify - tasks should be sorted by created_desc (newest first)
	retrievedTasks := widget.GetTasks()
	assert.Len(t, retrievedTasks, 2)
	assert.Equal(t, tasks[1].ID, retrievedTasks[0].ID) // task-2 (newer) first
	assert.Equal(t, tasks[0].ID, retrievedTasks[1].ID) // task-1 (older) second

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_AddTask(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoneColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)

	taskAddedCalled := false
	widget.SetOnTaskAdded(func(task *TaskData) {
		taskAddedCalled = true
		assert.Equal(t, "new-task", task.ID)
	})

	newTask := &TaskData{
		ID:          "new-task",
		Title:       "New Task",
		Description: "New task description",
		Priority:    "medium",
		Status:      "done",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Execute
	widget.AddTask(newTask)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify
	tasks := widget.GetTasks()
	assert.Len(t, tasks, 1)
	assert.Equal(t, "new-task", tasks[0].ID)
	assert.True(t, taskAddedCalled)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_RemoveTask(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoingColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)
	tasks := createTestTasksCollection()
	widget.SetTasks(tasks)

	taskRemovedCalled := false
	widget.SetOnTaskRemoved(func(taskID string) {
		taskRemovedCalled = true
		assert.Equal(t, "task-1", taskID)
	})

	// Give time for initial state update
	time.Sleep(50 * time.Millisecond)

	// Execute
	widget.RemoveTask("task-1")

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify
	remainingTasks := widget.GetTasks()
	assert.Len(t, remainingTasks, 1)
	assert.Equal(t, "task-2", remainingTasks[0].ID)
	assert.True(t, taskRemovedCalled)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_SetConfiguration(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	initialConfig := createTestColumnConfiguration(TodoColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, initialConfig)

	configChangedCalled := false
	widget.SetOnConfigChanged(func(config *ColumnConfiguration) {
		configChangedCalled = true
		assert.Equal(t, "Updated Column", config.Title)
	})

	// Execute
	newConfig := &ColumnConfiguration{
		Title:        "Updated Column",
		Type:         DoingColumn,
		WIPLimit:     10,
		Color:        "#red",
		ShowSections: false,
		SortOrder:    "priority_asc",
		Metadata:     make(map[string]interface{}),
	}
	widget.SetConfiguration(newConfig)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify
	retrievedConfig := widget.GetConfiguration()
	assert.Equal(t, "Updated Column", retrievedConfig.Title)
	assert.Equal(t, DoingColumn, retrievedConfig.Type)
	assert.Equal(t, 10, retrievedConfig.WIPLimit)
	assert.True(t, configChangedCalled)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_SetLoading(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(TodoColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)

	// Execute
	widget.SetLoading(true)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify
	widget.stateMu.RLock()
	isLoading := widget.currentState.IsLoading
	widget.stateMu.RUnlock()

	assert.True(t, isLoading)

	// Test clearing loading
	widget.SetLoading(false)
	time.Sleep(50 * time.Millisecond)

	widget.stateMu.RLock()
	isLoading = widget.currentState.IsLoading
	widget.stateMu.RUnlock()

	assert.False(t, isLoading)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_SetError(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoingColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)

	errorOccurred := false
	widget.SetOnError(func(err error) {
		errorOccurred = true
		assert.Contains(t, err.Error(), "general error for testing")
	})

	// Execute
	testError := assert.AnError
	widget.SetError(testError)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

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
	time.Sleep(50 * time.Millisecond)

	widget.stateMu.RLock()
	hasError = widget.currentState.HasError
	widget.stateMu.RUnlock()

	assert.False(t, hasError)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_SetSelected(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoneColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)

	selectionChangeCount := 0
	var lastSelectedState bool
	widget.SetOnSelectionChange(func(selected bool) {
		selectionChangeCount++
		lastSelectedState = selected
	})

	// Execute
	widget.SetSelected(true)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify
	assert.True(t, widget.IsSelected())
	assert.Equal(t, 1, selectionChangeCount)
	assert.True(t, lastSelectedState)

	// Test deselection
	widget.SetSelected(false)
	time.Sleep(50 * time.Millisecond)
	assert.False(t, widget.IsSelected())
	assert.Equal(t, 2, selectionChangeCount)
	assert.False(t, lastSelectedState)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_ColumnTypes(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	// Test Todo Column
	todoConfig := createTestColumnConfiguration(TodoColumn)
	todoWidget := NewColumnWidget(mockWM, mockDDE, layoutEngine, todoConfig)
	assert.Equal(t, TodoColumn, todoWidget.GetConfiguration().Type)
	assert.True(t, todoWidget.GetConfiguration().ShowSections)

	// Test Doing Column
	doingConfig := createTestColumnConfiguration(DoingColumn)
	doingConfig.ShowSections = false
	doingWidget := NewColumnWidget(mockWM, mockDDE, layoutEngine, doingConfig)
	assert.Equal(t, DoingColumn, doingWidget.GetConfiguration().Type)
	assert.False(t, doingWidget.GetConfiguration().ShowSections)

	// Test Done Column
	doneConfig := createTestColumnConfiguration(DoneColumn)
	doneConfig.ShowSections = false
	doneWidget := NewColumnWidget(mockWM, mockDDE, layoutEngine, doneConfig)
	assert.Equal(t, DoneColumn, doneWidget.GetConfiguration().Type)

	// Cleanup
	todoWidget.Destroy()
	doingWidget.Destroy()
	doneWidget.Destroy()
}

func TestUnit_ColumnWidget_GracefulDegradation_NoWorkflowManager(t *testing.T) {
	// Setup - nil WorkflowManager
	test.NewApp()
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(TodoColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(nil, mockDDE, layoutEngine, config)

	// Execute - should not crash
	widget.CreateTask("Test Task", "Description")

	// Give time for any async operations
	time.Sleep(50 * time.Millisecond)

	// Verify - widget still functions
	assert.Equal(t, config, widget.GetConfiguration())

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_GracefulDegradation_NoDragDropEngine(t *testing.T) {
	// Setup - nil DragDropEngine
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoingColumn)

	widget := NewColumnWidget(mockWM, nil, layoutEngine, config)

	// Execute - should not crash
	tasks := createTestTasksCollection()
	widget.SetTasks(tasks)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify - widget still functions
	assert.Len(t, widget.GetTasks(), 2)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_WIPLimitHandling(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(DoingColumn)
	config.WIPLimit = 2 // Set low limit for testing

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)
	tasks := createTestTasksCollection() // 2 tasks

	// Execute - add tasks up to WIP limit
	widget.SetTasks(tasks)

	// Give time for state update
	time.Sleep(50 * time.Millisecond)

	// Verify WIP limit reached
	widget.stateMu.RLock()
	wipReached := widget.currentState.WIPLimitReached
	widget.stateMu.RUnlock()

	assert.True(t, wipReached)

	// Test removing task reduces WIP
	widget.RemoveTask("task-1")
	time.Sleep(50 * time.Millisecond)

	widget.stateMu.RLock()
	wipReached = widget.currentState.WIPLimitReached
	widget.stateMu.RUnlock()

	assert.False(t, wipReached)

	// Cleanup
	widget.Destroy()
}

func TestUnit_ColumnWidget_Lifecycle_Destroy(t *testing.T) {
	// Setup
	test.NewApp()
	mockWM := &MockWorkflowManager{}
	mockDDE := &MockDragDropEngine{}
	layoutEngine := engines.NewLayoutEngine()
	config := createTestColumnConfiguration(TodoColumn)

	mockDDE.On("RegisterDropZone", mock.AnythingOfType("engines.DropZoneSpec")).Return(engines.ZoneID("test-zone"), nil)
	mockDDE.On("UnregisterDropZone", engines.ZoneID("test-zone")).Return(nil)

	widget := NewColumnWidget(mockWM, mockDDE, layoutEngine, config)

	// Add some tasks to test cleanup
	tasks := createTestTasksCollection()
	widget.SetTasks(tasks)
	time.Sleep(50 * time.Millisecond)

	// Verify initial state
	assert.NotNil(t, widget.stateChannel)
	assert.NotNil(t, widget.cancel)

	// Execute
	widget.Destroy()

	// Verify cleanup
	assert.Nil(t, widget.stateChannel)

	// Should not crash if called multiple times
	widget.Destroy()

	// Verify mocks were called
	mockDDE.AssertExpectations(t)
}