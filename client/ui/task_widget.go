// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

// WidgetMode represents the current mode of the TaskWidget
type WidgetMode int

const (
	DisplayMode WidgetMode = iota
	EditMode
	CreateMode
)

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

	// New fields for enhanced functionality
	Mode           WidgetMode
	FormData       map[string]interface{}
	IsFormDirty    bool
	CanSave        bool
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
	onTaskCreated     func(*TaskData)  // New
	onTaskUpdated     func(*TaskData)  // New
	onEditCancelled   func()           // New

	// UI configuration
	showMetadata bool
	compact      bool

	// Internal state
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTaskWidget creates a new TaskWidget with the specified dependencies, task data, and mode
func NewTaskWidget(wm managers.WorkflowManager, fe *engines.FormattingEngine, fve *engines.FormValidationEngine, taskData *TaskData, mode WidgetMode) *TaskWidget {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize form data for creation mode
	var formData map[string]interface{}
	if mode == CreateMode {
		formData = make(map[string]interface{})
	}

	widget := &TaskWidget{
		workflowManager:  wm,
		formattingEngine: fe,
		validationEngine: fve,
		currentState: &TaskWidgetState{
			Data:           taskData,
			ValidationErrs: make(map[string]string),
			Mode:           mode,
			FormData:       formData,
			IsFormDirty:    false,
			CanSave:        mode == CreateMode, // Creation mode can save initially
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

// Convenience constructors for backward compatibility and ease of use
func NewDisplayTaskWidget(wm managers.WorkflowManager, fe *engines.FormattingEngine, fve *engines.FormValidationEngine, taskData *TaskData) *TaskWidget {
	return NewTaskWidget(wm, fe, fve, taskData, DisplayMode)
}

func NewCreationTaskWidget(wm managers.WorkflowManager, fe *engines.FormattingEngine, fve *engines.FormValidationEngine) *TaskWidget {
	return NewTaskWidget(wm, fe, fve, nil, CreateMode)
}

// CreateRenderer implements fyne.Widget interface with custom renderer
func (tw *TaskWidget) CreateRenderer() fyne.WidgetRenderer {
	renderer := &TaskWidgetRenderer{
		widget: tw,
	}

	// Initialize display components
	renderer.titleLabel = widget.NewLabel("")
	renderer.descriptionLabel = widget.NewLabel("")
	renderer.metadataLabel = widget.NewLabel("")

	// Initialize form components
	renderer.titleEntry = widget.NewEntry()
	renderer.titleEntry.SetPlaceHolder("Task title...")
	renderer.titleEntry.OnChanged = renderer.onFormFieldChanged

	renderer.descriptionEntry = widget.NewMultiLineEntry()
	renderer.descriptionEntry.SetPlaceHolder("Task description...")
	renderer.descriptionEntry.OnChanged = renderer.onFormFieldChanged

	renderer.prioritySelect = widget.NewSelect([]string{"urgent important", "urgent non-important", "non-urgent important", "non-urgent non-important"}, renderer.onFormFieldChanged)
	renderer.prioritySelect.SetSelected("non-urgent non-important")

	// Initialize validation components
	renderer.validationLabel = widget.NewLabel("")
	renderer.validationLabel.Wrapping = fyne.TextWrapWord

	// Initialize action components
	renderer.saveButton = widget.NewButton("Save", renderer.onSaveClicked)
	renderer.cancelButton = widget.NewButton("Cancel", renderer.onCancelClicked)

	// Create container with initial layout
	renderer.container = container.NewVBox()
	renderer.refreshLayout()

	return renderer
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
	// Update CanSave based on validation errors
	newState.CanSave = len(errors) == 0 && newState.IsFormDirty
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

// SetOnTaskCreated sets the task creation event handler
func (tw *TaskWidget) SetOnTaskCreated(handler func(*TaskData)) {
	tw.onTaskCreated = handler
}

// SetOnTaskUpdated sets the task update event handler
func (tw *TaskWidget) SetOnTaskUpdated(handler func(*TaskData)) {
	tw.onTaskUpdated = handler
}

// SetOnEditCancelled sets the edit cancellation event handler
func (tw *TaskWidget) SetOnEditCancelled(handler func()) {
	tw.onEditCancelled = handler
}

// Mode Transition Methods

// EnterEditMode transitions the widget to edit mode for existing tasks
func (tw *TaskWidget) EnterEditMode() error {
	tw.stateMu.RLock()
	if tw.currentState.Data == nil {
		tw.stateMu.RUnlock()
		return fmt.Errorf("cannot edit widget without task data")
	}

	if tw.currentState.Mode == EditMode {
		tw.stateMu.RUnlock()
		return nil // Already in edit mode
	}

	// Prepare form data from existing task
	formData := map[string]interface{}{
		"title":       tw.currentState.Data.Title,
		"description": tw.currentState.Data.Description,
		"priority":    tw.currentState.Data.Priority,
	}
	tw.stateMu.RUnlock()

	// Copy current state and modify
	newState := tw.copyCurrentState()
	newState.Mode = EditMode
	newState.FormData = formData
	newState.IsFormDirty = false
	newState.CanSave = false

	tw.updateState(newState)
	return nil
}

// EnterCreateMode transitions the widget to creation mode for new tasks
func (tw *TaskWidget) EnterCreateMode() error {
	// Copy current state and modify
	newState := tw.copyCurrentState()
	newState.Mode = CreateMode
	newState.Data = nil // Clear task data for creation
	newState.FormData = make(map[string]interface{})
	newState.IsFormDirty = false
	newState.CanSave = false
	newState.ValidationErrs = make(map[string]string)

	tw.updateState(newState)
	return nil
}

// ExitEditMode transitions back to display mode
func (tw *TaskWidget) ExitEditMode() error {
	// Copy current state and modify
	newState := tw.copyCurrentState()
	newState.Mode = DisplayMode
	newState.FormData = nil
	newState.IsFormDirty = false
	newState.CanSave = false
	newState.ValidationErrs = make(map[string]string)

	tw.updateState(newState)
	return nil
}

// SaveTask saves the current form data using WorkflowManager
func (tw *TaskWidget) SaveTask() error {
	tw.stateMu.RLock()
	mode := tw.currentState.Mode
	formData := tw.currentState.FormData
	taskData := tw.currentState.Data
	tw.stateMu.RUnlock()

	if mode != EditMode && mode != CreateMode {
		return fmt.Errorf("cannot save in display mode")
	}

	if !tw.validateAllFields() {
		return fmt.Errorf("validation failed")
	}

	// Show loading state
	tw.SetLoading(true)

	switch mode {
	case CreateMode:
		return tw.processCreateWorkflow(formData)
	case EditMode:
		return tw.processUpdateWorkflow(taskData.ID, formData)
	}

	return nil
}

// CancelEdit cancels the current edit/create operation
func (tw *TaskWidget) CancelEdit() error {
	tw.stateMu.RLock()
	mode := tw.currentState.Mode
	tw.stateMu.RUnlock()

	if mode == DisplayMode {
		return nil // Nothing to cancel
	}

	if tw.onEditCancelled != nil {
		tw.onEditCancelled()
	}

	return tw.ExitEditMode()
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
		Mode:           tw.currentState.Mode,
		FormData:       tw.currentState.FormData,
		IsFormDirty:    tw.currentState.IsFormDirty,
		CanSave:        tw.currentState.CanSave,
	}

	for k, v := range tw.currentState.ValidationErrs {
		newState.ValidationErrs[k] = v
	}

	// Copy FormData map if it exists
	if tw.currentState.FormData != nil {
		newState.FormData = make(map[string]interface{})
		for k, v := range tw.currentState.FormData {
			newState.FormData[k] = v
		}
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

// Form Validation and Workflow Methods

// validateAllFields validates all form fields using FormValidationEngine
func (tw *TaskWidget) validateAllFields() bool {
	tw.stateMu.RLock()
	formData := tw.currentState.FormData
	tw.stateMu.RUnlock()

	if tw.validationEngine == nil || formData == nil {
		return false
	}

	// Convert form data to map[string]any for validation
	validationData := make(map[string]any)
	for k, v := range formData {
		validationData[k] = v
	}

	// Define validation rules
	rules := tw.getValidationRules()
	result := tw.validationEngine.ValidateFormInputs(validationData, rules)

	// Update validation state
	tw.updateValidationDisplay(result)

	return result.Valid
}

// getValidationRules returns validation rules for task fields
func (tw *TaskWidget) getValidationRules() engines.ValidationRules {
	fieldRules := make(map[string]engines.FieldRule)

	// Title validation: required, 1-200 characters
	fieldRules["title"] = engines.FieldRule{
		Required: true,
		Type:     engines.FieldTypeText,
		Length: engines.LengthConstraints{
			MinLength: 1,
			MaxLength: 200,
		},
	}

	// Description validation: optional, max 1000 characters
	fieldRules["description"] = engines.FieldRule{
		Required: false,
		Type:     engines.FieldTypeText,
		Length: engines.LengthConstraints{
			MaxLength: 1000,
		},
	}

	// Priority validation: required, must be valid Eisenhower value
	fieldRules["priority"] = engines.FieldRule{
		Required: true,
		Type:     engines.FieldTypeText,
		Pattern:  "^(urgent-important|urgent-not-important|not-urgent-important|not-urgent-not-important)$",
	}

	return engines.ValidationRules{
		FieldRules: fieldRules,
	}
}

// updateValidationDisplay updates validation error display
func (tw *TaskWidget) updateValidationDisplay(result engines.ValidationResult) {
	newState := tw.copyCurrentState()
	newState.ValidationErrs = make(map[string]string)

	// Convert validation errors to string map
	for _, err := range result.Errors {
		newState.ValidationErrs[err.Field] = err.Message
	}

	// Update save button state
	newState.CanSave = result.Valid && newState.IsFormDirty

	tw.updateState(newState)
}

// processCreateWorkflow handles task creation workflow
func (tw *TaskWidget) processCreateWorkflow(formData map[string]interface{}) error {
	if tw.workflowManager == nil {
		return fmt.Errorf("workflow manager unavailable")
	}

	ctx := context.Background()

	// Convert form data to workflow request
	request := map[string]any{
		"title":       formData["title"],
		"description": formData["description"],
		"priority":    formData["priority"],
		"status":      "todo",
	}

	go func() {
		defer tw.SetLoading(false)

		response, err := tw.workflowManager.Task().CreateTaskWorkflow(ctx, request)
		if err != nil {
			tw.SetError(fmt.Errorf("task creation failed: %w", err))
			return
		}

		// Convert response to TaskData
		taskData := tw.responseToTaskData(response)

		// Update widget state with new task
		newState := tw.copyCurrentState()
		newState.Data = taskData
		newState.Mode = DisplayMode
		newState.FormData = nil
		newState.IsFormDirty = false
		newState.CanSave = false
		tw.updateState(newState)

		// Notify creation success
		if tw.onTaskCreated != nil {
			tw.onTaskCreated(taskData)
		}
	}()

	return nil
}

// processUpdateWorkflow handles task update workflow
func (tw *TaskWidget) processUpdateWorkflow(taskID string, formData map[string]interface{}) error {
	if tw.workflowManager == nil {
		return fmt.Errorf("workflow manager unavailable")
	}

	ctx := context.Background()

	// Convert form data to workflow request
	request := map[string]any{
		"title":       formData["title"],
		"description": formData["description"],
		"priority":    formData["priority"],
	}

	go func() {
		defer tw.SetLoading(false)

		response, err := tw.workflowManager.Task().UpdateTaskWorkflow(ctx, taskID, request)
		if err != nil {
			tw.SetError(fmt.Errorf("task update failed: %w", err))
			return
		}

		// Convert response to TaskData
		taskData := tw.responseToTaskData(response)

		// Update widget state with updated task
		newState := tw.copyCurrentState()
		newState.Data = taskData
		newState.Mode = DisplayMode
		newState.FormData = nil
		newState.IsFormDirty = false
		newState.CanSave = false
		tw.updateState(newState)

		// Notify update success
		if tw.onTaskUpdated != nil {
			tw.onTaskUpdated(taskData)
		}
	}()

	return nil
}

// responseToTaskData converts workflow response to TaskData
func (tw *TaskWidget) responseToTaskData(response map[string]any) *TaskData {
	taskData := &TaskData{}

	if id, ok := response["id"].(string); ok {
		taskData.ID = id
	}

	if title, ok := response["title"].(string); ok {
		taskData.Title = title
	}

	if description, ok := response["description"].(string); ok {
		taskData.Description = description
	}

	if priority, ok := response["priority"].(string); ok {
		taskData.Priority = priority
	}

	if status, ok := response["status"].(string); ok {
		taskData.Status = status
	}

	// Handle metadata
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		taskData.Metadata = metadata
	} else {
		taskData.Metadata = make(map[string]interface{})
	}

	// Parse timestamps
	if createdAt, ok := response["created_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, createdAt); err == nil {
			taskData.CreatedAt = parsed
		} else {
			taskData.CreatedAt = time.Now()
		}
	} else {
		taskData.CreatedAt = time.Now()
	}

	if updatedAt, ok := response["updated_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			taskData.UpdatedAt = parsed
		} else {
			taskData.UpdatedAt = time.Now()
		}
	} else {
		taskData.UpdatedAt = time.Now()
	}

	return taskData
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

// TaskWidgetRenderer implements the renderer for TaskWidget following the unified form approach
type TaskWidgetRenderer struct {
	widget *TaskWidget

	// Display components
	titleLabel       *widget.Label
	descriptionLabel *widget.Label
	metadataLabel    *widget.Label

	// Form components (for edit/create modes)
	titleEntry       *widget.Entry
	descriptionEntry *widget.Entry
	prioritySelect   *widget.Select

	// Validation components
	validationLabel  *widget.Label

	// Action components
	saveButton       *widget.Button
	cancelButton     *widget.Button

	// Layout container
	container        *fyne.Container
}


// Objects returns all rendered objects
func (r *TaskWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.container.Objects
}

// Layout arranges the components within the given size
func (r *TaskWidgetRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

// MinSize returns the minimum size required for the widget
func (r *TaskWidgetRenderer) MinSize() fyne.Size {
	return r.container.MinSize()
}

// Refresh updates the renderer based on current widget state
func (r *TaskWidgetRenderer) Refresh() {
	r.refreshContent()
	r.refreshLayout()
	r.container.Refresh()
}

// Destroy cleans up the renderer resources
func (r *TaskWidgetRenderer) Destroy() {
	// Clean up any resources if needed
}

// refreshContent updates the content of components based on current state
func (r *TaskWidgetRenderer) refreshContent() {
	r.widget.stateMu.RLock()
	state := r.widget.currentState
	r.widget.stateMu.RUnlock()

	mode := state.Mode

	// Update display components
	if state.Data != nil {
		title, description, metadata := r.widget.formatTaskDisplay()
		r.titleLabel.SetText(title)
		r.descriptionLabel.SetText(description)
		r.metadataLabel.SetText(metadata)
	}

	// Update form components based on mode and current data
	if mode == EditMode && state.Data != nil {
		r.titleEntry.SetText(state.Data.Title)
		r.descriptionEntry.SetText(state.Data.Description)
		r.prioritySelect.SetSelected(state.Data.Priority)
	} else if mode == CreateMode {
		if formData := state.FormData; formData != nil {
			if title, ok := formData["title"].(string); ok {
				r.titleEntry.SetText(title)
			}
			if description, ok := formData["description"].(string); ok {
				r.descriptionEntry.SetText(description)
			}
			if priority, ok := formData["priority"].(string); ok {
				r.prioritySelect.SetSelected(priority)
			}
		}
	}

	// Update validation display
	if len(state.ValidationErrs) > 0 {
		var errMessages []string
		for field, err := range state.ValidationErrs {
			errMessages = append(errMessages, fmt.Sprintf("%s: %s", field, err))
		}
		r.validationLabel.SetText(strings.Join(errMessages, "\n"))
	} else {
		r.validationLabel.SetText("")
	}

	// Update action button states
	r.saveButton.Enable()
	if !state.CanSave || state.IsLoading {
		r.saveButton.Disable()
	}
}

// refreshLayout updates the layout based on current widget mode
func (r *TaskWidgetRenderer) refreshLayout() {
	r.widget.stateMu.RLock()
	mode := r.widget.currentState.Mode
	r.widget.stateMu.RUnlock()

	// Clear current layout
	r.container.Objects = nil

	switch mode {
	case DisplayMode:
		// Display mode: show labels only
		r.container.Objects = []fyne.CanvasObject{
			r.titleLabel,
			r.descriptionLabel,
			r.metadataLabel,
		}

	case EditMode, CreateMode:
		// Edit/Create mode: show form components
		r.container.Objects = []fyne.CanvasObject{
			widget.NewLabel("Title:"),
			r.titleEntry,
			widget.NewLabel("Description:"),
			r.descriptionEntry,
			widget.NewLabel("Priority:"),
			r.prioritySelect,
			r.validationLabel,
			container.NewHBox(r.saveButton, r.cancelButton),
		}
	}
}

// Event handlers for form interactions

func (r *TaskWidgetRenderer) onFormFieldChanged(value string) {
	r.widget.stateMu.Lock()
	defer r.widget.stateMu.Unlock()

	// Initialize FormData if needed
	if r.widget.currentState.FormData == nil {
		r.widget.currentState.FormData = make(map[string]interface{})
	}

	// Update FormData based on which field changed
	r.widget.currentState.FormData["title"] = r.titleEntry.Text
	r.widget.currentState.FormData["description"] = r.descriptionEntry.Text
	r.widget.currentState.FormData["priority"] = r.prioritySelect.Selected

	// Mark form as dirty
	r.widget.currentState.IsFormDirty = true

	// Trigger real-time validation
	go func() {
		if r.widget.validationEngine != nil {
			rules := r.widget.getValidationRules()
			result := r.widget.validationEngine.ValidateFormInputs(r.widget.currentState.FormData, rules)
			r.widget.updateValidationDisplay(result)
		}
	}()
}

func (r *TaskWidgetRenderer) onSaveClicked() {
	if err := r.widget.SaveTask(); err != nil {
		r.widget.SetError(fmt.Errorf("save failed: %w", err))
	}
}

func (r *TaskWidgetRenderer) onCancelClicked() {
	if err := r.widget.CancelEdit(); err != nil {
		r.widget.SetError(fmt.Errorf("cancel failed: %w", err))
	}
}