// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// DialogQuadrant represents the Eisenhower Matrix quadrants within the dialog
type DialogQuadrant string

const (
	DialogUrgentImportant    DialogQuadrant = "urgent-important"
	DialogUrgentNonImportant DialogQuadrant = "urgent-non-important"
	DialogNonUrgentImportant DialogQuadrant = "non-urgent-important"
	DialogNonUrgentNonImportant DialogQuadrant = "non-urgent-non-important"
)

// TaskMovement represents a tracked task movement operation
type TaskMovement struct {
	TaskID         string
	FromQuadrant   DialogQuadrant
	ToQuadrant     DialogQuadrant
	FromPosition   int
	ToPosition     int
	PriorityChange string
}

// DialogState represents the current state of the CreateTaskDialog
type DialogState struct {
	// Task collections per quadrant
	UrgentImportantTasks    []*TaskData
	UrgentNonImportantTasks []*TaskData
	NonUrgentImportantTasks []*TaskData
	CreatedTasks            []*TaskData // Tasks created in dialog

	// Movement tracking for deferred WorkflowManager operations
	TaskMovements []TaskMovement

	// Dialog state
	IsLoading      bool
	HasError       bool
	ErrorMessage   string
	IsCreating     bool
}

// CreateTaskDialog provides a modal Eisenhower Matrix interface for task creation and organization
type CreateTaskDialog struct {
	// Engine dependencies
	workflowManager     managers.WorkflowManager
	formattingEngine    *engines.FormattingEngine
	validationEngine    *engines.FormValidationEngine
	layoutEngine        *engines.LayoutEngine
	dragDropEngine      *engines.DragDropEngine

	// UI components
	dialog           dialog.Dialog
	parentWindow     fyne.Window
	matrixContainer  *fyne.Container
	quadrantContainers map[DialogQuadrant]*fyne.Container

	// TaskWidget management
	taskWidgets      map[string]*TaskWidget // taskID -> TaskWidget
	creationWidget   *TaskWidget

	// State management
	stateMu      sync.RWMutex
	currentState *DialogState
	stateChannel chan *DialogState

	// Event handlers
	onComplete func(*TaskData, error)
	onCancel   func()
	onTaskCreated func(*TaskData)
	onTaskMoved func(taskID, fromQuadrant, toQuadrant string)
	onValidationError func(errors map[string]string)

	// Internal state
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCreateTaskDialog creates a new CreateTaskDialog with the specified dependencies
func NewCreateTaskDialog(
	wm managers.WorkflowManager,
	fe *engines.FormattingEngine,
	fve *engines.FormValidationEngine,
	le *engines.LayoutEngine,
	dde *engines.DragDropEngine,
	parent fyne.Window,
) *CreateTaskDialog {
	ctx, cancel := context.WithCancel(context.Background())

	ctd := &CreateTaskDialog{
		workflowManager:  wm,
		formattingEngine: fe,
		validationEngine: fve,
		layoutEngine:     le,
		dragDropEngine:   dde,
		parentWindow:     parent,
		currentState: &DialogState{
			UrgentImportantTasks:    []*TaskData{},
			UrgentNonImportantTasks: []*TaskData{},
			NonUrgentImportantTasks: []*TaskData{},
			CreatedTasks:            []*TaskData{},
			TaskMovements:           []TaskMovement{},
		},
		stateChannel:       make(chan *DialogState, 10),
		taskWidgets:        make(map[string]*TaskWidget),
		quadrantContainers: make(map[DialogQuadrant]*fyne.Container),
		ctx:                ctx,
		cancel:             cancel,
	}

	// Create UI components
	ctd.createMatrixLayout()
	ctd.createDialog()
	ctd.setupDropZones()

	// Start state management goroutine
	go ctd.handleStateUpdates()

	return ctd
}

// Show displays the dialog modally and queries existing tasks
func (ctd *CreateTaskDialog) Show() {
	ctd.ShowWithData(nil)
}

// ShowWithData displays the dialog with pre-populated initial form data
func (ctd *CreateTaskDialog) ShowWithData(initialData map[string]interface{}) {
	// Set loading state
	ctd.updateState(&DialogState{
		UrgentImportantTasks:    []*TaskData{},
		UrgentNonImportantTasks: []*TaskData{},
		NonUrgentImportantTasks: []*TaskData{},
		CreatedTasks:            []*TaskData{},
		TaskMovements:           []TaskMovement{},
		IsLoading:              true,
	})

	// Create creation widget with initial data
	ctd.creationWidget = NewCreationTaskWidget(ctd.workflowManager, ctd.formattingEngine, ctd.validationEngine)

	// Set up creation widget event handlers
	ctd.creationWidget.SetOnTaskCreated(ctd.handleTaskCreated)

	if initialData != nil {
		// Pre-populate creation widget with initial data
		ctd.creationWidget.stateMu.Lock()
		if ctd.creationWidget.currentState.FormData == nil {
			ctd.creationWidget.currentState.FormData = make(map[string]interface{})
		}
		for k, v := range initialData {
			ctd.creationWidget.currentState.FormData[k] = v
		}
		ctd.creationWidget.stateMu.Unlock()
	}

	// Add creation widget to non-urgent non-important quadrant
	ctd.quadrantContainers[DialogNonUrgentNonImportant].Add(ctd.creationWidget)

	// Load existing tasks asynchronously
	go ctd.loadExistingTasks()

	// Show dialog
	ctd.dialog.Show()
}

// SetOnComplete sets the callback for dialog completion events
func (ctd *CreateTaskDialog) SetOnComplete(callback func(*TaskData, error)) {
	ctd.onComplete = callback
}

// SetOnCancel sets the callback for dialog cancellation events
func (ctd *CreateTaskDialog) SetOnCancel(callback func()) {
	ctd.onCancel = callback
}

// SetOnTaskCreated sets the callback for task creation completion events
func (ctd *CreateTaskDialog) SetOnTaskCreated(callback func(*TaskData)) {
	ctd.onTaskCreated = callback
}

// SetOnTaskMoved sets the callback for task movement events
func (ctd *CreateTaskDialog) SetOnTaskMoved(callback func(taskID, fromQuadrant, toQuadrant string)) {
	ctd.onTaskMoved = callback
}

// SetOnValidationError sets the callback for validation error events
func (ctd *CreateTaskDialog) SetOnValidationError(callback func(errors map[string]string)) {
	ctd.onValidationError = callback
}

// RefreshQuadrants reloads and redisplays quadrant contents
func (ctd *CreateTaskDialog) RefreshQuadrants() {
	go ctd.loadExistingTasks()
}

// GetQuadrantTasks retrieves task collections for specific Eisenhower quadrants
func (ctd *CreateTaskDialog) GetQuadrantTasks(quadrant DialogQuadrant) []*TaskData {
	ctd.stateMu.RLock()
	defer ctd.stateMu.RUnlock()

	switch quadrant {
	case DialogUrgentImportant:
		return ctd.currentState.UrgentImportantTasks
	case DialogUrgentNonImportant:
		return ctd.currentState.UrgentNonImportantTasks
	case DialogNonUrgentImportant:
		return ctd.currentState.NonUrgentImportantTasks
	case DialogNonUrgentNonImportant:
		return ctd.currentState.CreatedTasks
	default:
		return []*TaskData{}
	}
}

// MoveTaskToQuadrant programmatically moves tasks between quadrants
func (ctd *CreateTaskDialog) MoveTaskToQuadrant(taskID string, quadrant DialogQuadrant, position int) error {
	// Find current task location
	currentQuadrant, currentPosition := ctd.findTaskLocation(taskID)
	if currentQuadrant == "" {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Create movement record
	movement := TaskMovement{
		TaskID:         taskID,
		FromQuadrant:   DialogQuadrant(currentQuadrant),
		ToQuadrant:     quadrant,
		FromPosition:   currentPosition,
		ToPosition:     position,
		PriorityChange: ctd.quadrantToPriority(quadrant),
	}

	// Track movement for deferred WorkflowManager operation
	ctd.stateMu.Lock()
	newState := ctd.copyCurrentState()
	newState.TaskMovements = append(newState.TaskMovements, movement)
	ctd.stateMu.Unlock()

	// Apply movement to UI state
	ctd.applyTaskMovement(movement)
	ctd.updateState(newState)

	// Notify movement event
	if ctd.onTaskMoved != nil {
		ctd.onTaskMoved(taskID, string(movement.FromQuadrant), string(movement.ToQuadrant))
	}

	return nil
}

// Internal Methods

// createMatrixLayout creates the 2x2 Eisenhower Matrix layout
func (ctd *CreateTaskDialog) createMatrixLayout() {
	// Create quadrant containers
	ctd.quadrantContainers[DialogUrgentImportant] = container.NewVBox()
	ctd.quadrantContainers[DialogUrgentNonImportant] = container.NewVBox()
	ctd.quadrantContainers[DialogNonUrgentImportant] = container.NewVBox()
	ctd.quadrantContainers[DialogNonUrgentNonImportant] = container.NewVBox()

	// Add quadrant labels
	urgentImportantSection := container.NewVBox(
		widget.NewLabelWithStyle("Urgent & Important", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ctd.quadrantContainers[DialogUrgentImportant],
	)

	urgentNonImportantSection := container.NewVBox(
		widget.NewLabelWithStyle("Urgent & Non-Important", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ctd.quadrantContainers[DialogUrgentNonImportant],
	)

	nonUrgentImportantSection := container.NewVBox(
		widget.NewLabelWithStyle("Non-Urgent & Important", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ctd.quadrantContainers[DialogNonUrgentImportant],
	)

	nonUrgentNonImportantSection := container.NewVBox(
		widget.NewLabelWithStyle("New Task Creation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ctd.quadrantContainers[DialogNonUrgentNonImportant],
	)

	// Create 2x2 grid layout
	ctd.matrixContainer = container.NewGridWithColumns(2,
		urgentImportantSection,
		urgentNonImportantSection,
		nonUrgentImportantSection,
		nonUrgentNonImportantSection,
	)
}

// createDialog creates the Fyne dialog with matrix content
func (ctd *CreateTaskDialog) createDialog() {
	// Create dialog content with matrix
	content := container.NewVBox(
		widget.NewLabelWithStyle("Task Creation - Eisenhower Matrix", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ctd.matrixContainer,
	)

	// Create dialog with custom buttons
	ctd.dialog = dialog.NewCustom("Create Tasks", "Close", content, ctd.parentWindow)
	ctd.dialog.SetOnClosed(ctd.handleDialogClosed)

	// Set dialog size
	ctd.dialog.Resize(fyne.NewSize(800, 600))
}

// loadExistingTasks queries and loads existing tasks for display quadrants
func (ctd *CreateTaskDialog) loadExistingTasks() {
	if ctd.workflowManager == nil {
		ctd.updateState(&DialogState{
			UrgentImportantTasks:    []*TaskData{},
			UrgentNonImportantTasks: []*TaskData{},
			NonUrgentImportantTasks: []*TaskData{},
			CreatedTasks:            []*TaskData{},
			TaskMovements:           []TaskMovement{},
			IsLoading:              false,
		})
		return
	}

	ctx := context.Background()

	// Query tasks for each quadrant
	urgentImportantTasks := ctd.queryTasksByPriority(ctx, "urgent-important")
	urgentNonImportantTasks := ctd.queryTasksByPriority(ctx, "urgent-non-important")
	nonUrgentImportantTasks := ctd.queryTasksByPriority(ctx, "non-urgent-important")

	// Update state with loaded tasks
	newState := &DialogState{
		UrgentImportantTasks:    urgentImportantTasks,
		UrgentNonImportantTasks: urgentNonImportantTasks,
		NonUrgentImportantTasks: nonUrgentImportantTasks,
		CreatedTasks:            []*TaskData{},
		TaskMovements:           []TaskMovement{},
		IsLoading:              false,
	}

	ctd.updateState(newState)

	// Create and add TaskWidgets for existing tasks
	ctd.populateQuadrantWidgets()
}

// queryTasksByPriority queries tasks with specific priority from WorkflowManager
func (ctd *CreateTaskDialog) queryTasksByPriority(ctx context.Context, priority string) []*TaskData {
	criteria := map[string]any{
		"priority": priority,
		"status":   "todo", // Only show todo tasks in creation dialog
	}

	response, err := ctd.workflowManager.Task().QueryTasksWorkflow(ctx, criteria)
	if err != nil {
		return []*TaskData{}
	}

	// Convert response to TaskData slice
	tasks := []*TaskData{}
	if taskList, ok := response["tasks"].([]interface{}); ok {
		for _, taskData := range taskList {
			if taskMap, ok := taskData.(map[string]interface{}); ok {
				task := ctd.mapResponseToTaskData(taskMap)
				if task != nil {
					tasks = append(tasks, task)
				}
			}
		}
	}

	return tasks
}

// populateQuadrantWidgets creates TaskWidgets for existing tasks and adds them to quadrants
func (ctd *CreateTaskDialog) populateQuadrantWidgets() {
	ctd.stateMu.RLock()
	state := ctd.currentState
	ctd.stateMu.RUnlock()

	// Clear existing widgets
	for _, widget := range ctd.taskWidgets {
		widget.Destroy()
	}
	ctd.taskWidgets = make(map[string]*TaskWidget)

	// Clear quadrant containers (except creation quadrant)
	ctd.quadrantContainers[DialogUrgentImportant].RemoveAll()
	ctd.quadrantContainers[DialogUrgentNonImportant].RemoveAll()
	ctd.quadrantContainers[DialogNonUrgentImportant].RemoveAll()

	// Add widgets for urgent important tasks
	for _, task := range state.UrgentImportantTasks {
		widget := NewDisplayTaskWidget(ctd.workflowManager, ctd.formattingEngine, ctd.validationEngine, task)
		ctd.setupTaskWidgetForDrag(widget, DialogUrgentImportant)
		ctd.taskWidgets[task.ID] = widget
		ctd.quadrantContainers[DialogUrgentImportant].Add(widget)
	}

	// Add widgets for urgent non-important tasks
	for _, task := range state.UrgentNonImportantTasks {
		widget := NewDisplayTaskWidget(ctd.workflowManager, ctd.formattingEngine, ctd.validationEngine, task)
		ctd.setupTaskWidgetForDrag(widget, DialogUrgentNonImportant)
		ctd.taskWidgets[task.ID] = widget
		ctd.quadrantContainers[DialogUrgentNonImportant].Add(widget)
	}

	// Add widgets for non-urgent important tasks
	for _, task := range state.NonUrgentImportantTasks {
		widget := NewDisplayTaskWidget(ctd.workflowManager, ctd.formattingEngine, ctd.validationEngine, task)
		ctd.setupTaskWidgetForDrag(widget, DialogNonUrgentImportant)
		ctd.taskWidgets[task.ID] = widget
		ctd.quadrantContainers[DialogNonUrgentImportant].Add(widget)
	}

	// Add widgets for created tasks (in creation quadrant)
	for _, task := range state.CreatedTasks {
		widget := NewDisplayTaskWidget(ctd.workflowManager, ctd.formattingEngine, ctd.validationEngine, task)
		ctd.setupTaskWidgetForDrag(widget, DialogNonUrgentNonImportant)
		ctd.taskWidgets[task.ID] = widget
		ctd.quadrantContainers[DialogNonUrgentNonImportant].Add(widget)
	}
}

// setupTaskWidgetForDrag configures TaskWidget for drag-drop operations
func (ctd *CreateTaskDialog) setupTaskWidgetForDrag(widget *TaskWidget, quadrant DialogQuadrant) {
	// Enable drag operations through DragDropEngine
	if ctd.dragDropEngine != nil {
		// Set up drag start handler
		widget.SetOnDragStart(func(event *fyne.DragEvent) {
			ctd.handleDragStart(widget, quadrant, event)
		})

		// Set up drag end handler
		widget.SetOnDragEnd(func() {
			ctd.handleDragEnd(widget, quadrant)
		})
	}
}

// handleTaskCreated handles task creation completion from creation widget
func (ctd *CreateTaskDialog) handleTaskCreated(taskData *TaskData) {
	// Add created task to dialog state
	ctd.stateMu.Lock()
	newState := ctd.copyCurrentState()
	newState.CreatedTasks = append(newState.CreatedTasks, taskData)
	ctd.stateMu.Unlock()

	ctd.updateState(newState)

	// Create and add TaskWidget for the new task
	widget := NewDisplayTaskWidget(ctd.workflowManager, ctd.formattingEngine, ctd.validationEngine, taskData)
	ctd.setupTaskWidgetForDrag(widget, DialogNonUrgentNonImportant)
	ctd.taskWidgets[taskData.ID] = widget
	ctd.quadrantContainers[DialogNonUrgentNonImportant].Add(widget)

	// Notify external handlers
	if ctd.onTaskCreated != nil {
		ctd.onTaskCreated(taskData)
	}
}


// handleDialogClosed handles dialog closing and executes deferred WorkflowManager operations
func (ctd *CreateTaskDialog) handleDialogClosed() {
	// Execute deferred WorkflowManager operations for task movements
	ctd.executeDeferredOperations()

	// Notify completion or cancellation
	if ctd.onCancel != nil {
		ctd.onCancel()
	}

	// Clean up resources
	ctd.cleanup()
}

// handleDragStart handles the start of a drag operation
func (ctd *CreateTaskDialog) handleDragStart(widget *TaskWidget, quadrant DialogQuadrant, event *fyne.DragEvent) {
	if ctd.dragDropEngine == nil {
		return
	}

	// Set widget as dragging
	widget.SetLoading(true) // Visual feedback for drag state

	// Coordinate with DragDropEngine for spatial mechanics
	params := engines.DragParams{
		SourceWidget: widget,
		DragType:     engines.DragTypeTask,
		Metadata: map[string]any{
			"task_id":   widget.GetTaskData().ID,
			"quadrant":  string(quadrant),
		},
	}

	_, err := (*ctd.dragDropEngine).Drag().StartDrag(widget, params)
	if err != nil {
		widget.SetLoading(false)
		fmt.Printf("Failed to start drag: %v\n", err)
	}
}

// handleDragEnd handles the end of a drag operation
func (ctd *CreateTaskDialog) handleDragEnd(widget *TaskWidget, quadrant DialogQuadrant) {
	if ctd.dragDropEngine == nil {
		return
	}

	// Clear loading state
	widget.SetLoading(false)

	// Note: Actual drag completion would be handled by DragDropEngine
	// through drag handle management. This is a simplified implementation.
	// In a full implementation, we would track the drag handle and complete it.
}

// setupDropZones registers drop zones for each quadrant with DragDropEngine
func (ctd *CreateTaskDialog) setupDropZones() {
	if ctd.dragDropEngine == nil {
		return
	}

	// Register drop zones for each quadrant container
	for quadrant, container := range ctd.quadrantContainers {
		zoneSpec := engines.DropZoneSpec{
			ID:          engines.ZoneID(string(quadrant)),
			Bounds:      container.Position(),
			Size:        container.Size(),
			AcceptTypes: []engines.DragType{engines.DragTypeTask},
			Container:   container,
			AcceptanceFn: func(ctx engines.DragContext) bool {
				// Accept task drags from any source
				return ctx.DragType == engines.DragTypeTask
			},
		}

		_, err := (*ctd.dragDropEngine).Drop().RegisterDropZone(zoneSpec)
		if err != nil {
			fmt.Printf("Failed to register drop zone for %s: %v\n", quadrant, err)
		}
	}
}

// updateTaskQuadrantInState updates the dialog state to reflect task movement between quadrants
func (ctd *CreateTaskDialog) updateTaskQuadrantInState(taskID string, fromQuadrant, toQuadrant DialogQuadrant) {
	ctd.stateMu.Lock()
	defer ctd.stateMu.Unlock()

	// Find and remove task from source quadrant
	var movedTask *TaskData
	switch fromQuadrant {
	case DialogUrgentImportant:
		for i, task := range ctd.currentState.UrgentImportantTasks {
			if task.ID == taskID {
				movedTask = task
				ctd.currentState.UrgentImportantTasks = append(
					ctd.currentState.UrgentImportantTasks[:i],
					ctd.currentState.UrgentImportantTasks[i+1:]...,
				)
				break
			}
		}
	case DialogUrgentNonImportant:
		for i, task := range ctd.currentState.UrgentNonImportantTasks {
			if task.ID == taskID {
				movedTask = task
				ctd.currentState.UrgentNonImportantTasks = append(
					ctd.currentState.UrgentNonImportantTasks[:i],
					ctd.currentState.UrgentNonImportantTasks[i+1:]...,
				)
				break
			}
		}
	case DialogNonUrgentImportant:
		for i, task := range ctd.currentState.NonUrgentImportantTasks {
			if task.ID == taskID {
				movedTask = task
				ctd.currentState.NonUrgentImportantTasks = append(
					ctd.currentState.NonUrgentImportantTasks[:i],
					ctd.currentState.NonUrgentImportantTasks[i+1:]...,
				)
				break
			}
		}
	case DialogNonUrgentNonImportant:
		for i, task := range ctd.currentState.CreatedTasks {
			if task.ID == taskID {
				movedTask = task
				ctd.currentState.CreatedTasks = append(
					ctd.currentState.CreatedTasks[:i],
					ctd.currentState.CreatedTasks[i+1:]...,
				)
				break
			}
		}
	}

	// Add task to target quadrant if found
	if movedTask != nil {
		// Update task priority
		movedTask.Priority = ctd.quadrantToPriority(toQuadrant)

		switch toQuadrant {
		case DialogUrgentImportant:
			ctd.currentState.UrgentImportantTasks = append(ctd.currentState.UrgentImportantTasks, movedTask)
		case DialogUrgentNonImportant:
			ctd.currentState.UrgentNonImportantTasks = append(ctd.currentState.UrgentNonImportantTasks, movedTask)
		case DialogNonUrgentImportant:
			ctd.currentState.NonUrgentImportantTasks = append(ctd.currentState.NonUrgentImportantTasks, movedTask)
		case DialogNonUrgentNonImportant:
			ctd.currentState.CreatedTasks = append(ctd.currentState.CreatedTasks, movedTask)
		}
	}
}

// executeDeferredOperations executes all tracked task movements through WorkflowManager
func (ctd *CreateTaskDialog) executeDeferredOperations() {
	ctd.stateMu.RLock()
	movements := ctd.currentState.TaskMovements
	ctd.stateMu.RUnlock()

	if len(movements) == 0 || ctd.workflowManager == nil {
		return
	}

	ctx := context.Background()

	// Execute each task movement
	for _, movement := range movements {
		request := map[string]any{
			"priority": movement.PriorityChange,
		}

		_, err := ctd.workflowManager.Task().UpdateTaskWorkflow(ctx, movement.TaskID, request)
		if err != nil {
			// Log error but continue with other operations
			fmt.Printf("Failed to update task %s priority: %v\n", movement.TaskID, err)
		}
	}
}

// Helper Methods

// findTaskLocation finds the current quadrant and position of a task
func (ctd *CreateTaskDialog) findTaskLocation(taskID string) (string, int) {
	ctd.stateMu.RLock()
	defer ctd.stateMu.RUnlock()

	// Check each quadrant for the task
	for i, task := range ctd.currentState.UrgentImportantTasks {
		if task.ID == taskID {
			return string(DialogUrgentImportant), i
		}
	}

	for i, task := range ctd.currentState.UrgentNonImportantTasks {
		if task.ID == taskID {
			return string(DialogUrgentNonImportant), i
		}
	}

	for i, task := range ctd.currentState.NonUrgentImportantTasks {
		if task.ID == taskID {
			return string(DialogNonUrgentImportant), i
		}
	}

	for i, task := range ctd.currentState.CreatedTasks {
		if task.ID == taskID {
			return string(DialogNonUrgentNonImportant), i
		}
	}

	return "", -1
}

// applyTaskMovement applies a task movement to the UI state
func (ctd *CreateTaskDialog) applyTaskMovement(movement TaskMovement) {
	// Find the TaskWidget for this task
	widget, exists := ctd.taskWidgets[movement.TaskID]
	if !exists {
		return
	}

	// Remove from source quadrant container
	sourceContainer := ctd.quadrantContainers[movement.FromQuadrant]
	if sourceContainer != nil {
		sourceContainer.Remove(widget)
	}

	// Add to target quadrant container
	targetContainer := ctd.quadrantContainers[movement.ToQuadrant]
	if targetContainer != nil {
		targetContainer.Add(widget)
		// Update drag setup for new quadrant
		ctd.setupTaskWidgetForDrag(widget, movement.ToQuadrant)
	}

	// Update task data priority
	taskData := widget.GetTaskData()
	if taskData != nil {
		// Create updated task data with new priority
		updatedData := *taskData
		updatedData.Priority = movement.PriorityChange
		widget.SetTaskData(&updatedData)
	}

	// Update dialog state to reflect the movement
	ctd.updateTaskQuadrantInState(movement.TaskID, movement.FromQuadrant, movement.ToQuadrant)
}

// quadrantToPriority converts quadrant type to priority string
func (ctd *CreateTaskDialog) quadrantToPriority(quadrant DialogQuadrant) string {
	return string(quadrant)
}

// mapResponseToTaskData converts WorkflowManager response to TaskData
func (ctd *CreateTaskDialog) mapResponseToTaskData(data map[string]interface{}) *TaskData {
	if data == nil {
		return nil
	}

	task := &TaskData{}

	if id, ok := data["id"].(string); ok {
		task.ID = id
	}
	if title, ok := data["title"].(string); ok {
		task.Title = title
	}
	if description, ok := data["description"].(string); ok {
		task.Description = description
	}
	if priority, ok := data["priority"].(string); ok {
		task.Priority = priority
	}
	if status, ok := data["status"].(string); ok {
		task.Status = status
	}
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		task.Metadata = metadata
	} else {
		task.Metadata = make(map[string]interface{})
	}

	return task
}

// State Management

// updateState sends a new state to the state channel for processing
func (ctd *CreateTaskDialog) updateState(newState *DialogState) {
	select {
	case ctd.stateChannel <- newState:
	case <-ctd.ctx.Done():
		return
	default:
		// Channel full, skip update to prevent blocking
	}
}

// copyCurrentState creates a copy of the current state for modification
func (ctd *CreateTaskDialog) copyCurrentState() *DialogState {
	ctd.stateMu.RLock()
	defer ctd.stateMu.RUnlock()

	newState := &DialogState{
		UrgentImportantTasks:    make([]*TaskData, len(ctd.currentState.UrgentImportantTasks)),
		UrgentNonImportantTasks: make([]*TaskData, len(ctd.currentState.UrgentNonImportantTasks)),
		NonUrgentImportantTasks: make([]*TaskData, len(ctd.currentState.NonUrgentImportantTasks)),
		CreatedTasks:            make([]*TaskData, len(ctd.currentState.CreatedTasks)),
		TaskMovements:           make([]TaskMovement, len(ctd.currentState.TaskMovements)),
		IsLoading:               ctd.currentState.IsLoading,
		HasError:                ctd.currentState.HasError,
		ErrorMessage:            ctd.currentState.ErrorMessage,
		IsCreating:              ctd.currentState.IsCreating,
	}

	copy(newState.UrgentImportantTasks, ctd.currentState.UrgentImportantTasks)
	copy(newState.UrgentNonImportantTasks, ctd.currentState.UrgentNonImportantTasks)
	copy(newState.NonUrgentImportantTasks, ctd.currentState.NonUrgentImportantTasks)
	copy(newState.CreatedTasks, ctd.currentState.CreatedTasks)
	copy(newState.TaskMovements, ctd.currentState.TaskMovements)

	return newState
}

// handleStateUpdates processes state updates from the state channel
func (ctd *CreateTaskDialog) handleStateUpdates() {
	for {
		select {
		case newState := <-ctd.stateChannel:
			if newState == nil {
				return
			}

			ctd.stateMu.Lock()
			ctd.currentState = newState
			ctd.stateMu.Unlock()

			// Trigger UI refresh if dialog is visible
			if ctd.dialog != nil {
				ctd.matrixContainer.Refresh()
			}

		case <-ctd.ctx.Done():
			return
		}
	}
}

// cleanup cleans up dialog resources
func (ctd *CreateTaskDialog) cleanup() {
	// Cancel context
	if ctd.cancel != nil {
		ctd.cancel()
	}

	// Destroy TaskWidgets
	for _, widget := range ctd.taskWidgets {
		widget.Destroy()
	}
	if ctd.creationWidget != nil {
		ctd.creationWidget.Destroy()
	}

	// Clear state channel
	close(ctd.stateChannel)
}