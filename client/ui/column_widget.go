// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"context"
	"fmt"
	"image/color"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// ColumnType defines the type of column in the kanban board
type ColumnType int

const (
	TodoColumn ColumnType = iota
	DoingColumn
	DoneColumn
)

// EisenhowerSection represents priority sections within Todo column
type EisenhowerSection string

const (
	UrgentImportant     EisenhowerSection = "urgent-important"
	UrgentNotImportant  EisenhowerSection = "urgent-not-important"
	NotUrgentImportant  EisenhowerSection = "not-urgent-important"
	NotUrgentNotImportant EisenhowerSection = "not-urgent-not-important"
)

// ColumnConfiguration represents column settings and behavior
type ColumnConfiguration struct {
	Title       string                   `json:"title"`
	Type        ColumnType               `json:"type"`
	WIPLimit    int                      `json:"wip_limit,omitempty"`
	Color       string                   `json:"color,omitempty"`
	ShowSections bool                    `json:"show_sections"`
	SortOrder   string                   `json:"sort_order,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// ColumnState represents the current state of the column widget
type ColumnState struct {
	Configuration  *ColumnConfiguration
	Tasks          []*TaskData
	TaskWidgets    map[string]*TaskWidget // taskID -> widget mapping
	IsLoading      bool
	HasError       bool
	ErrorMessage   string
	IsSelected     bool
	DropZoneActive bool
	WIPLimitReached bool
}

// ColumnWidget implements a Fyne widget for displaying a collection of tasks in a kanban column
// following the Custom Widget + Renderer Pattern with Engine Integration
type ColumnWidget struct {
	widget.BaseWidget

	// Dependencies (Constructor Injection)
	workflowManager  managers.WorkflowManager
	dragDropEngine   engines.DragDropEngine
	layoutEngine     *engines.LayoutEngine
	// Note: FyneUtility interface to be defined - using placeholder

	// State Management (Immutable with Channels)
	stateMu      sync.RWMutex
	currentState *ColumnState
	stateChannel chan *ColumnState

	// Drop Zone Management
	dropZoneID engines.ZoneID

	// Event handling
	onTaskAdded       func(*TaskData)
	onTaskRemoved     func(string)
	onTaskMoved       func(string, int)
	onConfigChanged   func(*ColumnConfiguration)
	onSelectionChange func(bool)
	onError           func(error)

	// Internal state
	ctx        context.Context
	cancel     context.CancelFunc
	scrollContainer *container.Scroll
}

// NewColumnWidget creates a new ColumnWidget with the specified dependencies and configuration
func NewColumnWidget(
	wm managers.WorkflowManager,
	dde engines.DragDropEngine,
	le *engines.LayoutEngine,
	config *ColumnConfiguration,
) *ColumnWidget {
	ctx, cancel := context.WithCancel(context.Background())

	widget := &ColumnWidget{
		workflowManager: wm,
		dragDropEngine:  dde,
		layoutEngine:    le,
		currentState: &ColumnState{
			Configuration: config,
			Tasks:         make([]*TaskData, 0),
			TaskWidgets:   make(map[string]*TaskWidget),
		},
		stateChannel: make(chan *ColumnState, 10),
		ctx:          ctx,
		cancel:       cancel,
	}

	widget.ExtendBaseWidget(widget)

	// Start state management goroutine
	go widget.handleStateUpdates()

	// Initialize drop zone
	widget.initializeDropZone()

	return widget
}

// CreateRenderer implements fyne.Widget interface with custom renderer
func (cw *ColumnWidget) CreateRenderer() fyne.WidgetRenderer {
	return newColumnWidgetRenderer(cw)
}

// Public API Methods

// SetTasks updates the column with new task collection
func (cw *ColumnWidget) SetTasks(tasks []*TaskData) {
	newState := cw.copyCurrentState()
	newState.Tasks = make([]*TaskData, len(tasks))
	copy(newState.Tasks, tasks)

	// Sort tasks for consistent display order
	cw.sortTasks(newState.Tasks)

	cw.updateState(newState)
	cw.recreateTaskWidgets()
}

// AddTask adds a new task to the column
func (cw *ColumnWidget) AddTask(task *TaskData) {
	newState := cw.copyCurrentState()
	newState.Tasks = append(newState.Tasks, task)
	cw.sortTasks(newState.Tasks)

	cw.updateState(newState)
	cw.createTaskWidget(task)

	if cw.onTaskAdded != nil {
		cw.onTaskAdded(task)
	}
}

// RemoveTask removes a task from the column
func (cw *ColumnWidget) RemoveTask(taskID string) {
	newState := cw.copyCurrentState()

	// Remove from tasks slice
	for i, task := range newState.Tasks {
		if task.ID == taskID {
			newState.Tasks = append(newState.Tasks[:i], newState.Tasks[i+1:]...)
			break
		}
	}

	// Remove widget
	if widget, exists := newState.TaskWidgets[taskID]; exists {
		widget.Destroy()
		delete(newState.TaskWidgets, taskID)
	}

	cw.updateState(newState)

	if cw.onTaskRemoved != nil {
		cw.onTaskRemoved(taskID)
	}
}

// GetTasks returns the current task collection
func (cw *ColumnWidget) GetTasks() []*TaskData {
	cw.stateMu.RLock()
	defer cw.stateMu.RUnlock()

	tasks := make([]*TaskData, len(cw.currentState.Tasks))
	copy(tasks, cw.currentState.Tasks)
	return tasks
}

// SetConfiguration updates column configuration
func (cw *ColumnWidget) SetConfiguration(config *ColumnConfiguration) {
	newState := cw.copyCurrentState()
	newState.Configuration = config
	cw.updateState(newState)

	if cw.onConfigChanged != nil {
		cw.onConfigChanged(config)
	}
}

// GetConfiguration returns current column configuration
func (cw *ColumnWidget) GetConfiguration() *ColumnConfiguration {
	cw.stateMu.RLock()
	defer cw.stateMu.RUnlock()
	return cw.currentState.Configuration
}

// SetLoading sets the loading state of the column
func (cw *ColumnWidget) SetLoading(loading bool) {
	newState := cw.copyCurrentState()
	newState.IsLoading = loading
	cw.updateState(newState)
}

// SetError sets an error state with message
func (cw *ColumnWidget) SetError(err error) {
	message := ""
	if err != nil {
		message = err.Error()
	}

	newState := cw.copyCurrentState()
	newState.HasError = err != nil
	newState.ErrorMessage = message
	newState.IsLoading = false
	cw.updateState(newState)

	if cw.onError != nil && err != nil {
		cw.onError(err)
	}
}

// SetSelected sets the selection state of the column
func (cw *ColumnWidget) SetSelected(selected bool) {
	newState := cw.copyCurrentState()
	newState.IsSelected = selected
	cw.updateState(newState)

	if cw.onSelectionChange != nil {
		cw.onSelectionChange(selected)
	}
}

// IsSelected returns the current selection state
func (cw *ColumnWidget) IsSelected() bool {
	cw.stateMu.RLock()
	defer cw.stateMu.RUnlock()
	return cw.currentState.IsSelected
}

// CreateTask initiates task creation workflow within column context
func (cw *ColumnWidget) CreateTask(title, description string) {
	if cw.workflowManager == nil {
		cw.SetError(fmt.Errorf("workflow manager unavailable"))
		return
	}

	go cw.processCreateTaskWorkflow(title, description)
}

// Event handler setters

// SetOnTaskAdded sets the task added event handler
func (cw *ColumnWidget) SetOnTaskAdded(handler func(*TaskData)) {
	cw.onTaskAdded = handler
}

// SetOnTaskRemoved sets the task removed event handler
func (cw *ColumnWidget) SetOnTaskRemoved(handler func(string)) {
	cw.onTaskRemoved = handler
}

// SetOnTaskMoved sets the task moved event handler
func (cw *ColumnWidget) SetOnTaskMoved(handler func(string, int)) {
	cw.onTaskMoved = handler
}

// SetOnConfigChanged sets the configuration changed event handler
func (cw *ColumnWidget) SetOnConfigChanged(handler func(*ColumnConfiguration)) {
	cw.onConfigChanged = handler
}

// SetOnSelectionChange sets the selection change event handler
func (cw *ColumnWidget) SetOnSelectionChange(handler func(bool)) {
	cw.onSelectionChange = handler
}

// SetOnError sets the error event handler
func (cw *ColumnWidget) SetOnError(handler func(error)) {
	cw.onError = handler
}

// Lifecycle Management

// Destroy cleans up the column widget resources
func (cw *ColumnWidget) Destroy() {
	// Cleanup drop zone
	if cw.dragDropEngine != nil && cw.dropZoneID != "" {
		cw.dragDropEngine.Drop().UnregisterDropZone(cw.dropZoneID)
	}

	// Destroy all task widgets
	cw.stateMu.Lock()
	for _, taskWidget := range cw.currentState.TaskWidgets {
		taskWidget.Destroy()
	}
	cw.stateMu.Unlock()

	// Cancel context and close channel
	if cw.cancel != nil {
		cw.cancel()
	}

	if cw.stateChannel != nil {
		close(cw.stateChannel)
		cw.stateChannel = nil
	}
}

// Internal State Management

// updateState sends a new state to the state channel for processing
func (cw *ColumnWidget) updateState(newState *ColumnState) {
	select {
	case cw.stateChannel <- newState:
	case <-cw.ctx.Done():
		return
	default:
		// Channel full, skip update to prevent blocking
	}
}

// copyCurrentState creates a copy of the current state for modification
func (cw *ColumnWidget) copyCurrentState() *ColumnState {
	cw.stateMu.RLock()
	defer cw.stateMu.RUnlock()

	newState := &ColumnState{
		Configuration: cw.currentState.Configuration,
		Tasks:         make([]*TaskData, len(cw.currentState.Tasks)),
		TaskWidgets:   make(map[string]*TaskWidget),
		IsLoading:     cw.currentState.IsLoading,
		HasError:      cw.currentState.HasError,
		ErrorMessage:  cw.currentState.ErrorMessage,
		IsSelected:    cw.currentState.IsSelected,
		DropZoneActive: cw.currentState.DropZoneActive,
		WIPLimitReached: cw.currentState.WIPLimitReached,
	}

	copy(newState.Tasks, cw.currentState.Tasks)
	for k, v := range cw.currentState.TaskWidgets {
		newState.TaskWidgets[k] = v
	}

	return newState
}

// handleStateUpdates processes state updates from the state channel
func (cw *ColumnWidget) handleStateUpdates() {
	for {
		select {
		case newState := <-cw.stateChannel:
			if newState == nil {
				return
			}

			cw.stateMu.Lock()
			cw.currentState = newState
			cw.stateMu.Unlock()

			// Update WIP limit check
			cw.checkWIPLimit()

			// Trigger UI refresh on main thread
			cw.Refresh()

		case <-cw.ctx.Done():
			return
		}
	}
}

// Task Management Methods

// createTaskWidget creates a new TaskWidget for the given task
func (cw *ColumnWidget) createTaskWidget(task *TaskData) {
	if task == nil {
		return
	}

	// Use existing formatting engine from a TaskWidget or create minimal one
	var formattingEngine *engines.FormattingEngine
	if len(cw.currentState.TaskWidgets) > 0 {
		// Get formatting engine from existing widget
		for _, widget := range cw.currentState.TaskWidgets {
			formattingEngine = widget.formattingEngine
			break
		}
	}
	if formattingEngine == nil {
		formattingEngine = engines.NewFormattingEngine()
	}

	// Use display mode for column-embedded widgets, add validation engine if needed
	validationEngine := engines.NewFormValidationEngine()
	taskWidget := NewTaskWidget(cw.workflowManager, formattingEngine, validationEngine, task, DisplayMode)

	// Set up task widget event handlers
	taskWidget.SetOnSelectionChange(func(selected bool) {
		// Handle task selection within column context
		cw.handleTaskSelection(task.ID, selected)
	})

	taskWidget.SetOnError(func(err error) {
		// Propagate task errors to column level
		cw.SetError(fmt.Errorf("task %s error: %w", task.ID, err))
	})

	cw.stateMu.Lock()
	cw.currentState.TaskWidgets[task.ID] = taskWidget
	cw.stateMu.Unlock()
}

// recreateTaskWidgets recreates all task widgets for current task collection
func (cw *ColumnWidget) recreateTaskWidgets() {
	cw.stateMu.Lock()
	defer cw.stateMu.Unlock()

	// Destroy existing widgets
	for _, widget := range cw.currentState.TaskWidgets {
		widget.Destroy()
	}
	cw.currentState.TaskWidgets = make(map[string]*TaskWidget)

	// Create new widgets
	for _, task := range cw.currentState.Tasks {
		cw.createTaskWidget(task)
	}
}

// sortTasks sorts tasks based on column configuration and type
func (cw *ColumnWidget) sortTasks(tasks []*TaskData) {
	if len(tasks) <= 1 {
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		// Default sort by creation time (newest first)
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
}

// Drop Zone Management

// initializeDropZone registers the column as a drop zone with DragDropEngine
func (cw *ColumnWidget) initializeDropZone() {
	if cw.dragDropEngine == nil {
		return
	}

	// Create drop zone specification for entire column
	dropZoneSpec := engines.DropZoneSpec{
		Bounds:      cw.Position(),
		Size:        cw.Size(),
		AcceptTypes: []engines.DragType{engines.DragTypeTask},
		Container:   cw,
	}

	zoneID, err := cw.dragDropEngine.Drop().RegisterDropZone(dropZoneSpec)
	if err != nil {
		cw.SetError(fmt.Errorf("failed to register drop zone: %w", err))
		return
	}

	cw.dropZoneID = zoneID
}

// calculateColumnBounds returns the current bounds of the column for drop zone registration
func (cw *ColumnWidget) calculateColumnBounds() (fyne.Position, fyne.Size) {
	// Use widget size and position for drop zone bounds
	return cw.Position(), cw.Size()
}

// calculateInsertionPosition determines where to insert a dropped task based on Y coordinate
func (cw *ColumnWidget) calculateInsertionPosition(mouseY float32) int {
	cw.stateMu.RLock()
	defer cw.stateMu.RUnlock()

	if len(cw.currentState.Tasks) == 0 {
		return 0
	}

	// Simple calculation - could be enhanced with LayoutEngine
	taskHeight := float32(80) // Approximate task widget height
	taskIndex := int(mouseY / taskHeight)

	if taskIndex < 0 {
		return 0
	}
	if taskIndex >= len(cw.currentState.Tasks) {
		return len(cw.currentState.Tasks)
	}

	return taskIndex
}

// determineEisenhowerSection determines priority section for Todo column drops
func (cw *ColumnWidget) determineEisenhowerSection(position int) EisenhowerSection {
	if cw.currentState.Configuration.Type != TodoColumn {
		return NotUrgentNotImportant // Default for non-Todo columns
	}

	// Simple section division - could be enhanced with LayoutEngine
	sectionSize := len(cw.currentState.Tasks) / 4
	if sectionSize == 0 {
		sectionSize = 1
	}

	switch {
	case position < sectionSize:
		return UrgentImportant
	case position < sectionSize*2:
		return UrgentNotImportant
	case position < sectionSize*3:
		return NotUrgentImportant
	default:
		return NotUrgentNotImportant
	}
}

// Workflow Integration Methods

// processCreateTaskWorkflow handles task creation workflow coordination
func (cw *ColumnWidget) processCreateTaskWorkflow(title, description string) {
	cw.SetLoading(true)

	// Create task context based on column type and configuration
	taskRequest := map[string]any{
		"title":       title,
		"description": description,
		"column_type": cw.currentState.Configuration.Type,
		"column_id":   cw.currentState.Configuration.Title,
	}

	// Add column-specific properties
	switch cw.currentState.Configuration.Type {
	case TodoColumn:
		taskRequest["status"] = "todo"
		taskRequest["priority"] = string(NotUrgentNotImportant) // Default priority
	case DoingColumn:
		taskRequest["status"] = "doing"
	case DoneColumn:
		taskRequest["status"] = "done"
	}

	ctx, cancel := context.WithTimeout(cw.ctx, 30*time.Second)
	defer cancel()

	response, err := cw.workflowManager.Task().CreateTaskWorkflow(ctx, taskRequest)

	cw.SetLoading(false)

	if err != nil {
		cw.SetError(fmt.Errorf("task creation failed: %w", err))
		return
	}

	// Process successful task creation
	if taskData := cw.mapResponseToTaskData(response); taskData != nil {
		cw.AddTask(taskData)
	}
}

// mapResponseToTaskData converts WorkflowManager response to TaskData
func (cw *ColumnWidget) mapResponseToTaskData(data map[string]any) *TaskData {
	if data == nil {
		return nil
	}

	task := &TaskData{
		Metadata: make(map[string]interface{}),
	}

	if id, ok := data["id"].(string); ok {
		task.ID = id
	}
	if title, ok := data["title"].(string); ok {
		task.Title = title
	}
	if desc, ok := data["description"].(string); ok {
		task.Description = desc
	}
	if priority, ok := data["priority"].(string); ok {
		task.Priority = priority
	}
	if status, ok := data["status"].(string); ok {
		task.Status = status
	}
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		task.Metadata = metadata
	}
	if createdAt, ok := data["created_at"].(time.Time); ok {
		task.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		task.UpdatedAt = updatedAt
	}

	return task
}

// Helper Methods

// checkWIPLimit checks if WIP limit is reached and updates state
func (cw *ColumnWidget) checkWIPLimit() {
	config := cw.currentState.Configuration
	if config.WIPLimit <= 0 {
		return
	}

	wipReached := len(cw.currentState.Tasks) >= config.WIPLimit
	if wipReached != cw.currentState.WIPLimitReached {
		newState := cw.copyCurrentState()
		newState.WIPLimitReached = wipReached
		cw.updateState(newState)
	}
}

// handleTaskSelection handles selection changes within task widgets
func (cw *ColumnWidget) handleTaskSelection(taskID string, selected bool) {
	// Could implement multi-selection logic here
	// For now, just propagate to parent
}

// getStateColors returns colors based on current column state
func (cw *ColumnWidget) getStateColors() (background, border color.Color) {
	// Default colors
	background = theme.Color(theme.ColorNameBackground)
	border = theme.Color(theme.ColorNameForeground)

	// State-specific colors
	switch {
	case cw.currentState.HasError:
		background = theme.Color(theme.ColorNameError)
		border = theme.Color(theme.ColorNameError)
	case cw.currentState.WIPLimitReached:
		border = theme.Color(theme.ColorNameWarning)
	case cw.currentState.IsSelected:
		background = theme.Color(theme.ColorNameSelection)
	case cw.currentState.DropZoneActive:
		background = theme.Color(theme.ColorNameHover)
	case cw.currentState.IsLoading:
		background = theme.Color(theme.ColorNameDisabled)
	}

	return background, border
}