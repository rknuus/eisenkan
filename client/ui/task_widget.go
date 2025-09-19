// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"context"
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// TaskData represents the immutable task data structure for TaskWidget
type TaskData struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TaskWidgetState represents the current state of the widget
type TaskWidgetState struct {
	Data           *TaskData
	IsSelected     bool
	IsEditing      bool
	IsDragging     bool
	IsLoading      bool
	HasError       bool
	ErrorMessage   string
	ValidationErrs map[string]string
}

// TaskWidget implements a Fyne widget for displaying individual task information
// following the Custom Widget + Renderer Pattern with Event Delegation
type TaskWidget struct {
	widget.BaseWidget

	// Dependencies (Constructor Injection)
	workflowManager   managers.WorkflowManager
	formattingEngine  *engines.FormattingEngine
	validationEngine  *engines.FormValidationEngine

	// State Management (Immutable with Channels)
	stateMu      sync.RWMutex
	currentState *TaskWidgetState
	stateChannel chan *TaskWidgetState

	// Event handling
	onTapped          func()
	onDoubleTapped    func()
	onRightTapped     func(*fyne.PointEvent)
	onDragStart       func(*fyne.DragEvent)
	onDragEnd         func()
	onSelectionChange func(bool)
	onError           func(error)

	// UI configuration
	showMetadata bool
	compact      bool

	// Internal state
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTaskWidget creates a new TaskWidget with the specified dependencies and initial task data
func NewTaskWidget(wm managers.WorkflowManager, fe *engines.FormattingEngine, taskData *TaskData) *TaskWidget {
	ctx, cancel := context.WithCancel(context.Background())

	widget := &TaskWidget{
		workflowManager:  wm,
		formattingEngine: fe,
		validationEngine: engines.NewFormValidationEngine(),
		currentState: &TaskWidgetState{
			Data:           taskData,
			ValidationErrs: make(map[string]string),
		},
		stateChannel: make(chan *TaskWidgetState, 10),
		showMetadata: true,
		compact:      false,
		ctx:          ctx,
		cancel:       cancel,
	}

	widget.ExtendBaseWidget(widget)

	// Start state management goroutine
	go widget.handleStateUpdates()

	return widget
}

// CreateRenderer implements fyne.Widget interface with custom renderer
func (tw *TaskWidget) CreateRenderer() fyne.WidgetRenderer {
	return newTaskWidgetRenderer(tw)
}

// Resize implements fyne.Widget interface
func (tw *TaskWidget) Resize(size fyne.Size) {
	tw.BaseWidget.Resize(size)
}

// Move implements fyne.Widget interface
func (tw *TaskWidget) Move(position fyne.Position) {
	tw.BaseWidget.Move(position)
}

// MinSize implements fyne.Widget interface
func (tw *TaskWidget) MinSize() fyne.Size {
	return fyne.NewSize(200, 80)
}

// Tapped implements fyne.Tappable interface
func (tw *TaskWidget) Tapped(event *fyne.PointEvent) {
	if tw.onTapped != nil {
		tw.onTapped()
	}
	tw.handleSelectionToggle()
}

// DoubleTapped implements fyne.DoubleTappable interface
func (tw *TaskWidget) DoubleTapped(event *fyne.PointEvent) {
	if tw.onDoubleTapped != nil {
		tw.onDoubleTapped()
	}
	tw.handleEditMode()
}

// TappedSecondary implements fyne.SecondaryTappable interface for context menu
func (tw *TaskWidget) TappedSecondary(event *fyne.PointEvent) {
	if tw.onRightTapped != nil {
		tw.onRightTapped(event)
	}
	tw.handleContextMenu(event)
}

// Dragged implements fyne.Draggable interface
func (tw *TaskWidget) Dragged(event *fyne.DragEvent) {
	tw.handleDragUpdate(event)
}

// DragEnd implements fyne.Draggable interface
func (tw *TaskWidget) DragEnd() {
	if tw.onDragEnd != nil {
		tw.onDragEnd()
	}
	tw.handleDragComplete()
}

// Public API Methods

// SetTaskData updates the widget with new task data
func (tw *TaskWidget) SetTaskData(data *TaskData) {
	if data == nil {
		return
	}

	newState := tw.copyCurrentState()
	newState.Data = data
	newState.IsLoading = false
	newState.HasError = false
	newState.ErrorMessage = ""

	tw.updateState(newState)
}

// GetTaskData returns the current task data
func (tw *TaskWidget) GetTaskData() *TaskData {
	tw.stateMu.RLock()
	defer tw.stateMu.RUnlock()
	return tw.currentState.Data
}

// SetSelected sets the selection state of the widget
func (tw *TaskWidget) SetSelected(selected bool) {
	newState := tw.copyCurrentState()
	newState.IsSelected = selected
	tw.updateState(newState)

	if tw.onSelectionChange != nil {
		tw.onSelectionChange(selected)
	}
}

// IsSelected returns the current selection state
func (tw *TaskWidget) IsSelected() bool {
	tw.stateMu.RLock()
	defer tw.stateMu.RUnlock()
	return tw.currentState.IsSelected
}

// SetLoading sets the loading state of the widget
func (tw *TaskWidget) SetLoading(loading bool) {
	newState := tw.copyCurrentState()
	newState.IsLoading = loading
	tw.updateState(newState)
}

// SetError sets an error state with message
func (tw *TaskWidget) SetError(err error) {
	message := ""
	if err != nil {
		message = err.Error()
	}

	newState := tw.copyCurrentState()
	newState.HasError = err != nil
	newState.ErrorMessage = message
	newState.IsLoading = false
	tw.updateState(newState)

	if tw.onError != nil && err != nil {
		tw.onError(err)
	}
}

// SetValidationErrors sets field-specific validation errors
func (tw *TaskWidget) SetValidationErrors(errors map[string]string) {
	newState := tw.copyCurrentState()
	newState.ValidationErrs = make(map[string]string)
	for k, v := range errors {
		newState.ValidationErrs[k] = v
	}
	tw.updateState(newState)
}

// SetCompactMode enables/disables compact display mode
func (tw *TaskWidget) SetCompactMode(compact bool) {
	tw.compact = compact
	tw.Refresh()
}

// SetShowMetadata enables/disables metadata display
func (tw *TaskWidget) SetShowMetadata(show bool) {
	tw.showMetadata = show
	tw.Refresh()
}

// Event handler setters

// SetOnTapped sets the tap event handler
func (tw *TaskWidget) SetOnTapped(handler func()) {
	tw.onTapped = handler
}

// SetOnDoubleTapped sets the double-tap event handler
func (tw *TaskWidget) SetOnDoubleTapped(handler func()) {
	tw.onDoubleTapped = handler
}

// SetOnRightTapped sets the right-click event handler
func (tw *TaskWidget) SetOnRightTapped(handler func(*fyne.PointEvent)) {
	tw.onRightTapped = handler
}

// SetOnDragStart sets the drag start event handler
func (tw *TaskWidget) SetOnDragStart(handler func(*fyne.DragEvent)) {
	tw.onDragStart = handler
}

// SetOnDragEnd sets the drag end event handler
func (tw *TaskWidget) SetOnDragEnd(handler func()) {
	tw.onDragEnd = handler
}

// SetOnSelectionChange sets the selection change event handler
func (tw *TaskWidget) SetOnSelectionChange(handler func(bool)) {
	tw.onSelectionChange = handler
}

// SetOnError sets the error event handler
func (tw *TaskWidget) SetOnError(handler func(error)) {
	tw.onError = handler
}

// Lifecycle Management

// Destroy cleans up the widget resources
func (tw *TaskWidget) Destroy() {
	if tw.cancel != nil {
		tw.cancel()
	}

	if tw.stateChannel != nil {
		close(tw.stateChannel)
		tw.stateChannel = nil
	}
}

// Internal State Management

// updateState sends a new state to the state channel for processing
func (tw *TaskWidget) updateState(newState *TaskWidgetState) {
	select {
	case tw.stateChannel <- newState:
	case <-tw.ctx.Done():
		return
	default:
		// Channel full, skip update to prevent blocking
	}
}

// copyCurrentState creates a copy of the current state for modification
func (tw *TaskWidget) copyCurrentState() *TaskWidgetState {
	tw.stateMu.RLock()
	defer tw.stateMu.RUnlock()

	newState := &TaskWidgetState{
		Data:           tw.currentState.Data,
		IsSelected:     tw.currentState.IsSelected,
		IsEditing:      tw.currentState.IsEditing,
		IsDragging:     tw.currentState.IsDragging,
		IsLoading:      tw.currentState.IsLoading,
		HasError:       tw.currentState.HasError,
		ErrorMessage:   tw.currentState.ErrorMessage,
		ValidationErrs: make(map[string]string),
	}

	for k, v := range tw.currentState.ValidationErrs {
		newState.ValidationErrs[k] = v
	}

	return newState
}

// handleStateUpdates processes state updates from the state channel
func (tw *TaskWidget) handleStateUpdates() {
	for {
		select {
		case newState := <-tw.stateChannel:
			if newState == nil {
				return
			}

			tw.stateMu.Lock()
			tw.currentState = newState
			tw.stateMu.Unlock()

			// Trigger UI refresh on main thread
			tw.Refresh()

		case <-tw.ctx.Done():
			return
		}
	}
}

// Event Handling Methods

// handleSelectionToggle handles single tap selection toggle
func (tw *TaskWidget) handleSelectionToggle() {
	newState := tw.copyCurrentState()
	newState.IsSelected = !newState.IsSelected
	tw.updateState(newState)

	if tw.onSelectionChange != nil {
		tw.onSelectionChange(newState.IsSelected)
	}
}

// handleEditMode handles double-tap edit mode activation
func (tw *TaskWidget) handleEditMode() {
	if tw.currentState.Data == nil {
		return
	}

	// Delegate edit workflow to WorkflowManager
	if tw.workflowManager != nil {
		go tw.processEditWorkflow()
	}
}

// handleContextMenu handles right-click context menu
func (tw *TaskWidget) handleContextMenu(event *fyne.PointEvent) {
	// Context menu will be handled by parent containers
	// This is a placeholder for future context menu integration
}

// handleDragUpdate handles drag position updates
func (tw *TaskWidget) handleDragUpdate(event *fyne.DragEvent) {
	if !tw.currentState.IsDragging {
		newState := tw.copyCurrentState()
		newState.IsDragging = true
		tw.updateState(newState)

		if tw.onDragStart != nil {
			tw.onDragStart(event)
		}
	}
}

// handleDragComplete handles drag operation completion
func (tw *TaskWidget) handleDragComplete() {
	if tw.currentState.IsDragging {
		newState := tw.copyCurrentState()
		newState.IsDragging = false
		tw.updateState(newState)

		// Delegate drag workflow to WorkflowManager
		if tw.workflowManager != nil {
			go tw.processDragWorkflow()
		}
	}
}

// Workflow Integration Methods

// processEditWorkflow delegates edit operations to WorkflowManager
func (tw *TaskWidget) processEditWorkflow() {
	if tw.workflowManager == nil || tw.currentState.Data == nil {
		tw.SetError(fmt.Errorf("workflow manager unavailable"))
		return
	}

	tw.SetLoading(true)

	// Create edit context for WorkflowManager
	editRequest := map[string]any{
		"taskId":    tw.currentState.Data.ID,
		"operation": "edit",
		"source":    "task_widget",
	}

	// Delegate to WorkflowManager ITask facet
	ctx, cancel := context.WithTimeout(tw.ctx, 30*time.Second)
	defer cancel()

	response, err := tw.workflowManager.Task().UpdateTaskWorkflow(ctx, tw.currentState.Data.ID, editRequest)

	tw.SetLoading(false)

	if err != nil {
		tw.SetError(fmt.Errorf("edit workflow failed: %w", err))
		return
	}

	// Process successful edit response
	tw.handleWorkflowResponse(response)
}

// processDragWorkflow delegates drag operations to WorkflowManager
func (tw *TaskWidget) processDragWorkflow() {
	if tw.workflowManager == nil || tw.currentState.Data == nil {
		tw.SetError(fmt.Errorf("workflow manager unavailable"))
		return
	}

	// Create drag context for WorkflowManager
	dragEvent := map[string]any{
		"taskId":    tw.currentState.Data.ID,
		"operation": "drag",
		"source":    "task_widget",
	}

	// Delegate to WorkflowManager IDrag facet
	ctx, cancel := context.WithTimeout(tw.ctx, 10*time.Second)
	defer cancel()

	response, err := tw.workflowManager.Drag().ProcessDragDropWorkflow(ctx, dragEvent)

	if err != nil {
		tw.SetError(fmt.Errorf("drag workflow failed: %w", err))
		return
	}

	// Process successful drag response
	tw.handleWorkflowResponse(response)
}

// handleWorkflowResponse processes responses from WorkflowManager operations
func (tw *TaskWidget) handleWorkflowResponse(response map[string]any) {
	if response == nil {
		return
	}

	// Check for validation errors
	if validationErrs, ok := response["validation_errors"].(map[string]string); ok {
		tw.SetValidationErrors(validationErrs)
		return
	}

	// Check for workflow errors
	if workflowErr, ok := response["error"].(string); ok {
		tw.SetError(fmt.Errorf("workflow error: %s", workflowErr))
		return
	}

	// Process successful response with updated task data
	if taskData, ok := response["task_data"].(map[string]any); ok {
		updatedTask := tw.mapResponseToTaskData(taskData)
		if updatedTask != nil {
			tw.SetTaskData(updatedTask)
		}
	}
}

// mapResponseToTaskData converts WorkflowManager response to TaskData
func (tw *TaskWidget) mapResponseToTaskData(data map[string]any) *TaskData {
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

// Input Validation Integration

// validateAndSanitizeInput uses FormValidationEngine for input sanitization
func (tw *TaskWidget) validateAndSanitizeInput(input map[string]string) (map[string]string, error) {
	if tw.validationEngine == nil {
		return input, fmt.Errorf("validation engine unavailable")
	}

	sanitized := make(map[string]string)
	validationErrors := make(map[string]string)

	for field, value := range input {
		// Sanitize input using FormValidationEngine
		// Note: Using placeholder until actual FormValidationEngine method is available
		sanitized[field] = value // Placeholder - implement actual sanitization
	}

	if len(validationErrors) > 0 {
		tw.SetValidationErrors(validationErrors)
		return sanitized, fmt.Errorf("validation errors occurred")
	}

	return sanitized, nil
}

// Helper Methods

// formatTaskDisplay uses FormattingEngine for text presentation
func (tw *TaskWidget) formatTaskDisplay() (title, description, metadata string) {
	if tw.currentState.Data == nil {
		return "No Data", "", ""
	}

	data := tw.currentState.Data

	// Format title
	if tw.formattingEngine != nil {
		if formattedTitle, err := tw.formattingEngine.Text().FormatText(data.Title, engines.TextOptions{MaxLength: 50}); err == nil {
			title = formattedTitle
		} else {
			title = data.Title
		}

		// Format description
		if formattedDesc, err := tw.formattingEngine.Text().FormatText(data.Description, engines.TextOptions{MaxLength: 100}); err == nil {
			description = formattedDesc
		} else {
			description = data.Description
		}

		// Format metadata
		if tw.showMetadata && len(data.Metadata) > 0 {
			formattedMeta := tw.formattingEngine.Datastructure().FormatKeyValue(data.Metadata, engines.KeyValueOptions{Separator: ", "})
			metadata = formattedMeta
		}
	} else {
		// Fallback without FormattingEngine
		title = data.Title
		description = data.Description
		if tw.showMetadata && len(data.Metadata) > 0 {
			metadata = fmt.Sprintf("%v", data.Metadata)
		}
	}

	return title, description, metadata
}

// getStateColors returns colors based on current widget state
func (tw *TaskWidget) getStateColors() (background, border color.Color) {
	// Default colors
	background = theme.Color(theme.ColorNameBackground)
	border = theme.Color(theme.ColorNameForeground)

	// State-specific colors
	switch {
	case tw.currentState.HasError:
		background = theme.Color(theme.ColorNameError)
		border = theme.Color(theme.ColorNameError)
	case len(tw.currentState.ValidationErrs) > 0:
		border = theme.Color(theme.ColorNameError)
	case tw.currentState.IsSelected:
		background = theme.Color(theme.ColorNameSelection)
	case tw.currentState.IsDragging:
		background = theme.Color(theme.ColorNameHover)
	case tw.currentState.IsLoading:
		background = theme.Color(theme.ColorNameDisabled)
	}

	return background, border
}