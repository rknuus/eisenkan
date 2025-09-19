// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// BoardConfiguration represents the configuration for the entire board
type BoardConfiguration struct {
	Title         string                    `json:"title"`
	Columns       []*ColumnConfiguration    `json:"columns"`
	BoardType     string                    `json:"board_type"`     // "eisenhower", "kanban", "custom"
	EnableDragDrop bool                     `json:"enable_drag_drop"`
	Metadata      map[string]interface{}    `json:"metadata,omitempty"`
}

// BoardState represents the current state of the board widget
type BoardState struct {
	Configuration  *BoardConfiguration
	Columns        []*ColumnWidget
	AllTasks       []*TaskData
	IsLoading      bool
	HasError       bool
	ErrorMessage   string
	IsRefreshing   bool
	LastRefresh    time.Time
}

// BoardView implements a Fyne widget for displaying a kanban board with configurable columns
// following the Custom Widget + Renderer Pattern with Manager Integration
type BoardView struct {
	widget.BaseWidget

	// Dependencies (Constructor Injection)
	workflowManager   managers.WorkflowManager
	validationEngine  *engines.FormValidationEngine
	// Note: LayoutEngine and DragDropEngine accessed through ColumnWidget

	// State Management (Immutable with Channels)
	stateMu      sync.RWMutex
	currentState *BoardState
	stateChannel chan *BoardState

	// Event handling
	onTaskMoved       func(taskID, fromColumn, toColumn string)
	onTaskSelected    func(taskID string)
	onBoardRefreshed  func()
	onError           func(error)
	onConfigChanged   func(*BoardConfiguration)

	// Internal state
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBoardView creates a new BoardView with the specified dependencies and configuration
func NewBoardView(
	wm managers.WorkflowManager,
	ve *engines.FormValidationEngine,
	config *BoardConfiguration,
) *BoardView {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		// Default Eisenhower Matrix configuration
		config = &BoardConfiguration{
			Title:         "Eisenhower Matrix",
			BoardType:     "eisenhower",
			EnableDragDrop: true,
			Columns: []*ColumnConfiguration{
				{
					Title:       "Urgent Important",
					Type:        TodoColumn,
					ShowSections: true,
					Color:       "red",
					WIPLimit:    10,
				},
				{
					Title:       "Urgent Non-Important",
					Type:        TodoColumn,
					ShowSections: true,
					Color:       "orange",
					WIPLimit:    15,
				},
				{
					Title:       "Non-Urgent Important",
					Type:        TodoColumn,
					ShowSections: true,
					Color:       "blue",
					WIPLimit:    20,
				},
				{
					Title:       "Non-Urgent Non-Important",
					Type:        TodoColumn,
					ShowSections: true,
					Color:       "gray",
					WIPLimit:    25,
				},
			},
		}
	}

	board := &BoardView{
		workflowManager:  wm,
		validationEngine: ve,
		currentState: &BoardState{
			Configuration: config,
			Columns:       make([]*ColumnWidget, 0),
			AllTasks:      make([]*TaskData, 0),
		},
		stateChannel: make(chan *BoardState, 10),
		ctx:          ctx,
		cancel:       cancel,
	}

	board.ExtendBaseWidget(board)

	// Start state management goroutine
	go board.handleStateUpdates()

	// Initialize columns
	board.initializeColumns()

	return board
}

// CreateRenderer implements fyne.Widget interface with custom renderer
func (bv *BoardView) CreateRenderer() fyne.WidgetRenderer {
	return newBoardViewRenderer(bv)
}

// Public API Methods

// LoadBoard loads tasks from WorkflowManager and organizes them into columns
func (bv *BoardView) LoadBoard() {
	if bv.workflowManager == nil {
		bv.SetError(fmt.Errorf("workflow manager unavailable"))
		return
	}

	go bv.processLoadBoardWorkflow()
}

// RefreshBoard reloads all task data and updates column displays
func (bv *BoardView) RefreshBoard() {
	bv.setRefreshing(true)
	bv.LoadBoard()
}

// GetBoardState returns the current board state
func (bv *BoardView) GetBoardState() *BoardState {
	bv.stateMu.RLock()
	defer bv.stateMu.RUnlock()

	// Return a copy to prevent external modification
	return &BoardState{
		Configuration: bv.currentState.Configuration,
		Columns:       bv.currentState.Columns,
		AllTasks:      bv.currentState.AllTasks,
		IsLoading:     bv.currentState.IsLoading,
		HasError:      bv.currentState.HasError,
		ErrorMessage:  bv.currentState.ErrorMessage,
		IsRefreshing:  bv.currentState.IsRefreshing,
		LastRefresh:   bv.currentState.LastRefresh,
	}
}

// SetBoardConfiguration updates the board configuration and recreates columns
func (bv *BoardView) SetBoardConfiguration(config *BoardConfiguration) {
	newState := bv.copyCurrentState()
	newState.Configuration = config
	bv.updateState(newState)

	// Reinitialize columns with new configuration
	bv.initializeColumns()

	if bv.onConfigChanged != nil {
		bv.onConfigChanged(config)
	}
}

// GetColumnTasks returns task collections for a specific column
func (bv *BoardView) GetColumnTasks(columnIndex int) []*TaskData {
	bv.stateMu.RLock()
	defer bv.stateMu.RUnlock()

	if columnIndex < 0 || columnIndex >= len(bv.currentState.Columns) {
		return make([]*TaskData, 0)
	}

	return bv.currentState.Columns[columnIndex].GetTasks()
}

// MoveTask programmatically moves a task between columns with validation
func (bv *BoardView) MoveTask(taskID string, fromColumnIndex, toColumnIndex int) error {
	// Validate indices
	if fromColumnIndex < 0 || fromColumnIndex >= len(bv.currentState.Columns) {
		return fmt.Errorf("invalid from column index: %d", fromColumnIndex)
	}
	if toColumnIndex < 0 || toColumnIndex >= len(bv.currentState.Columns) {
		return fmt.Errorf("invalid to column index: %d", toColumnIndex)
	}

	// Validate with FormValidationEngine if available
	if bv.validationEngine != nil {
		validationData := map[string]any{
			"task_id":         taskID,
			"from_column":     fromColumnIndex,
			"to_column":       toColumnIndex,
			"operation_type":  "task_move",
		}

		// Create validation rules for task movement
		rules := engines.ValidationRules{
			FieldRules: map[string]engines.FieldRule{
				"task_id": {
					Required: true,
					Type:     engines.FieldTypeText,
				},
				"from_column": {
					Required: true,
					Type:     engines.FieldTypeNumeric,
				},
				"to_column": {
					Required: true,
					Type:     engines.FieldTypeNumeric,
				},
			},
		}

		result := bv.validationEngine.ValidateFormInputs(validationData, rules)
		if !result.Valid {
			return fmt.Errorf("task move validation failed: %v", result.Errors)
		}
	}

	return bv.processTaskMovement(taskID, fromColumnIndex, toColumnIndex)
}

// SelectTask highlights and selects a specific task within the board
func (bv *BoardView) SelectTask(taskID string) {
	// Find and select task across all columns
	for _, column := range bv.currentState.Columns {
		for _, task := range column.GetTasks() {
			if task.ID == taskID {
				column.SetSelected(true)
				if bv.onTaskSelected != nil {
					bv.onTaskSelected(taskID)
				}
				return
			}
		}
	}
}

// RefreshTask updates a specific task display without full board refresh
func (bv *BoardView) RefreshTask(taskID string) {
	// Find task across columns and refresh its display
	for _, column := range bv.currentState.Columns {
		for _, task := range column.GetTasks() {
			if task.ID == taskID {
				// Trigger column refresh to update task display
				column.Refresh()
				return
			}
		}
	}
}

// SetLoading sets the loading state of the board
func (bv *BoardView) SetLoading(loading bool) {
	newState := bv.copyCurrentState()
	newState.IsLoading = loading
	bv.updateState(newState)
}

// SetError sets an error state with message
func (bv *BoardView) SetError(err error) {
	message := ""
	if err != nil {
		message = err.Error()
	}

	newState := bv.copyCurrentState()
	newState.HasError = err != nil
	newState.ErrorMessage = message
	newState.IsLoading = false
	newState.IsRefreshing = false
	bv.updateState(newState)

	if bv.onError != nil && err != nil {
		bv.onError(err)
	}
}

// Event handler setters

// SetOnTaskMoved sets the task moved event handler
func (bv *BoardView) SetOnTaskMoved(handler func(taskID, fromColumn, toColumn string)) {
	bv.onTaskMoved = handler
}

// SetOnTaskSelected sets the task selected event handler
func (bv *BoardView) SetOnTaskSelected(handler func(taskID string)) {
	bv.onTaskSelected = handler
}

// SetOnBoardRefreshed sets the board refreshed event handler
func (bv *BoardView) SetOnBoardRefreshed(handler func()) {
	bv.onBoardRefreshed = handler
}

// SetOnError sets the error event handler
func (bv *BoardView) SetOnError(handler func(error)) {
	bv.onError = handler
}

// SetOnConfigChanged sets the configuration changed event handler
func (bv *BoardView) SetOnConfigChanged(handler func(*BoardConfiguration)) {
	bv.onConfigChanged = handler
}

// Lifecycle Management

// Destroy cleans up the board widget resources
func (bv *BoardView) Destroy() {
	// Destroy all column widgets
	bv.stateMu.Lock()
	for _, column := range bv.currentState.Columns {
		column.Destroy()
	}
	bv.stateMu.Unlock()

	// Cancel context and close channel
	if bv.cancel != nil {
		bv.cancel()
	}

	if bv.stateChannel != nil {
		close(bv.stateChannel)
		bv.stateChannel = nil
	}
}

// Internal State Management

// updateState sends a new state to the state channel for processing
func (bv *BoardView) updateState(newState *BoardState) {
	select {
	case bv.stateChannel <- newState:
		// For immediate state updates (like in tests), also update directly
		bv.stateMu.Lock()
		bv.currentState = newState
		bv.stateMu.Unlock()
	case <-bv.ctx.Done():
		return
	default:
		// Channel full, update directly to prevent blocking
		bv.stateMu.Lock()
		bv.currentState = newState
		bv.stateMu.Unlock()
	}
}

// copyCurrentState creates a copy of the current state for modification
func (bv *BoardView) copyCurrentState() *BoardState {
	bv.stateMu.RLock()
	defer bv.stateMu.RUnlock()

	newState := &BoardState{
		Configuration: bv.currentState.Configuration,
		Columns:       make([]*ColumnWidget, len(bv.currentState.Columns)),
		AllTasks:      make([]*TaskData, len(bv.currentState.AllTasks)),
		IsLoading:     bv.currentState.IsLoading,
		HasError:      bv.currentState.HasError,
		ErrorMessage:  bv.currentState.ErrorMessage,
		IsRefreshing:  bv.currentState.IsRefreshing,
		LastRefresh:   bv.currentState.LastRefresh,
	}

	copy(newState.Columns, bv.currentState.Columns)
	copy(newState.AllTasks, bv.currentState.AllTasks)

	return newState
}

// handleStateUpdates processes state updates from the state channel
func (bv *BoardView) handleStateUpdates() {
	for {
		select {
		case newState := <-bv.stateChannel:
			if newState == nil {
				return
			}

			bv.stateMu.Lock()
			bv.currentState = newState
			bv.stateMu.Unlock()

			// Trigger UI refresh on main thread
			bv.Refresh()

		case <-bv.ctx.Done():
			return
		}
	}
}

// setRefreshing sets the refreshing state
func (bv *BoardView) setRefreshing(refreshing bool) {
	newState := bv.copyCurrentState()
	newState.IsRefreshing = refreshing
	if !refreshing {
		newState.LastRefresh = time.Now()
	}
	bv.updateState(newState)
}

// Column Management Methods

// initializeColumns creates ColumnWidget instances based on board configuration
func (bv *BoardView) initializeColumns() {
	bv.stateMu.Lock()
	defer bv.stateMu.Unlock()

	// Destroy existing columns
	for _, column := range bv.currentState.Columns {
		column.Destroy()
	}

	// Create new columns based on configuration
	newColumns := make([]*ColumnWidget, 0, len(bv.currentState.Configuration.Columns))

	for _, columnConfig := range bv.currentState.Configuration.Columns {
		// Create engines for ColumnWidget - using placeholders for now
		dragDropEngine := engines.NewDragDropEngine() // Placeholder
		layoutEngine := engines.NewLayoutEngine()     // Placeholder

		column := NewColumnWidget(
			bv.workflowManager,
			dragDropEngine,
			layoutEngine,
			columnConfig,
		)

		// Set up column event handlers
		bv.setupColumnEventHandlers(column)

		newColumns = append(newColumns, column)
	}

	bv.currentState.Columns = newColumns
}

// setupColumnEventHandlers configures event handlers for a column widget
func (bv *BoardView) setupColumnEventHandlers(column *ColumnWidget) {
	// Handle task addition in column
	column.SetOnTaskAdded(func(task *TaskData) {
		// Update board state with new task
		bv.addTaskToBoardState(task)
	})

	// Handle task removal from column
	column.SetOnTaskRemoved(func(taskID string) {
		// Update board state to remove task
		bv.removeTaskFromBoardState(taskID)
	})

	// Handle column configuration changes
	column.SetOnConfigChanged(func(config *ColumnConfiguration) {
		// Update board configuration
		bv.updateColumnInBoardConfig(config)
	})

	// Handle column errors
	column.SetOnError(func(err error) {
		// Propagate error to board level
		bv.SetError(fmt.Errorf("column error: %w", err))
	})
}

// Workflow Integration Methods

// processLoadBoardWorkflow handles board loading workflow coordination
func (bv *BoardView) processLoadBoardWorkflow() {
	bv.SetLoading(true)

	ctx, cancel := context.WithTimeout(bv.ctx, 30*time.Second)
	defer cancel()

	// Query all tasks through WorkflowManager
	criteria := map[string]any{
		"board_type": bv.currentState.Configuration.BoardType,
		"include_archived": false,
	}

	response, err := bv.workflowManager.Task().QueryTasksWorkflow(ctx, criteria)

	bv.SetLoading(false)

	if err != nil {
		bv.SetError(fmt.Errorf("task loading failed: %w", err))
		return
	}

	// Process task data and organize into columns
	if err := bv.organizeTasksIntoColumns(response); err != nil {
		bv.SetError(fmt.Errorf("task organization failed: %w", err))
		return
	}

	bv.setRefreshing(false)

	if bv.onBoardRefreshed != nil {
		bv.onBoardRefreshed()
	}
}

// processTaskMovement handles task movement between columns through WorkflowManager
func (bv *BoardView) processTaskMovement(taskID string, fromColumnIndex, toColumnIndex int) error {
	ctx, cancel := context.WithTimeout(bv.ctx, 10*time.Second)
	defer cancel()

	// Prepare drag-drop event for WorkflowManager
	dragEvent := map[string]any{
		"task_id":      taskID,
		"from_column":  fromColumnIndex,
		"to_column":    toColumnIndex,
		"operation":    "move_task",
		"board_type":   bv.currentState.Configuration.BoardType,
	}

	response, err := bv.workflowManager.Drag().ProcessDragDropWorkflow(ctx, dragEvent)
	if err != nil {
		return fmt.Errorf("task movement workflow failed: %w", err)
	}

	// Update column states based on successful movement
	if err := bv.updateColumnsAfterTaskMovement(taskID, fromColumnIndex, toColumnIndex, response); err != nil {
		return fmt.Errorf("column update failed: %w", err)
	}

	if bv.onTaskMoved != nil {
		fromColumn := bv.currentState.Configuration.Columns[fromColumnIndex].Title
		toColumn := bv.currentState.Configuration.Columns[toColumnIndex].Title
		bv.onTaskMoved(taskID, fromColumn, toColumn)
	}

	return nil
}

// Helper Methods

// organizeTasksIntoColumns distributes tasks across columns based on their properties
func (bv *BoardView) organizeTasksIntoColumns(taskResponse map[string]any) error {
	// Extract tasks from WorkflowManager response
	tasksData, ok := taskResponse["tasks"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid task response format")
	}

	// Convert to TaskData structures
	allTasks := make([]*TaskData, 0, len(tasksData))
	for _, taskItem := range tasksData {
		if taskMap, ok := taskItem.(map[string]interface{}); ok {
			task := bv.mapResponseToTaskData(taskMap)
			if task != nil {
				allTasks = append(allTasks, task)
			}
		}
	}

	// Update board state with all tasks
	newState := bv.copyCurrentState()
	newState.AllTasks = allTasks
	bv.updateState(newState)

	// Distribute tasks to appropriate columns
	for i, column := range bv.currentState.Columns {
		columnTasks := bv.getTasksForColumn(allTasks, i)
		column.SetTasks(columnTasks)
	}

	return nil
}

// getTasksForColumn filters tasks that belong to a specific column
func (bv *BoardView) getTasksForColumn(allTasks []*TaskData, columnIndex int) []*TaskData {
	if columnIndex < 0 || columnIndex >= len(bv.currentState.Configuration.Columns) {
		return make([]*TaskData, 0)
	}

	columnConfig := bv.currentState.Configuration.Columns[columnIndex]
	columnTasks := make([]*TaskData, 0)

	for _, task := range allTasks {
		if bv.taskBelongsToColumn(task, columnConfig) {
			columnTasks = append(columnTasks, task)
		}
	}

	return columnTasks
}

// taskBelongsToColumn determines if a task belongs to a specific column
func (bv *BoardView) taskBelongsToColumn(task *TaskData, columnConfig *ColumnConfiguration) bool {
	// For Eisenhower Matrix, match based on priority and column title
	switch bv.currentState.Configuration.BoardType {
	case "eisenhower":
		return bv.taskMatchesEisenhowerColumn(task, columnConfig)
	case "kanban":
		return bv.taskMatchesKanbanColumn(task, columnConfig)
	default:
		return bv.taskMatchesGenericColumn(task, columnConfig)
	}
}

// taskMatchesEisenhowerColumn matches tasks to Eisenhower Matrix columns
func (bv *BoardView) taskMatchesEisenhowerColumn(task *TaskData, columnConfig *ColumnConfiguration) bool {
	switch columnConfig.Title {
	case "Urgent Important":
		return task.Priority == "urgent important" || task.Priority == string(UrgentImportant)
	case "Urgent Non-Important":
		return task.Priority == "urgent non-important" || task.Priority == string(UrgentNotImportant)
	case "Non-Urgent Important":
		return task.Priority == "non-urgent important" || task.Priority == string(NotUrgentImportant)
	case "Non-Urgent Non-Important":
		return task.Priority == "non-urgent non-important" || task.Priority == string(NotUrgentNotImportant)
	default:
		return false
	}
}

// taskMatchesKanbanColumn matches tasks to standard Kanban columns
func (bv *BoardView) taskMatchesKanbanColumn(task *TaskData, columnConfig *ColumnConfiguration) bool {
	switch columnConfig.Type {
	case TodoColumn:
		return task.Status == "todo" || task.Status == "backlog"
	case DoingColumn:
		return task.Status == "doing" || task.Status == "in_progress"
	case DoneColumn:
		return task.Status == "done" || task.Status == "completed"
	default:
		return false
	}
}

// taskMatchesGenericColumn matches tasks to generic columns based on metadata
func (bv *BoardView) taskMatchesGenericColumn(task *TaskData, columnConfig *ColumnConfiguration) bool {
	// Generic matching based on task metadata or configuration
	if columnMetadata, exists := columnConfig.Metadata["match_criteria"]; exists {
		if criteria, ok := columnMetadata.(map[string]interface{}); ok {
			for key, value := range criteria {
				if taskValue, exists := task.Metadata[key]; exists && taskValue == value {
					return true
				}
			}
		}
	}
	return false
}

// mapResponseToTaskData converts WorkflowManager response to TaskData
func (bv *BoardView) mapResponseToTaskData(data map[string]interface{}) *TaskData {
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

// updateColumnsAfterTaskMovement updates column states after successful task movement
func (bv *BoardView) updateColumnsAfterTaskMovement(taskID string, fromColumnIndex, toColumnIndex int, response map[string]any) error {
	// Find the task to move
	var taskToMove *TaskData
	fromColumn := bv.currentState.Columns[fromColumnIndex]

	for _, task := range fromColumn.GetTasks() {
		if task.ID == taskID {
			taskToMove = task
			break
		}
	}

	if taskToMove == nil {
		return fmt.Errorf("task %s not found in from column", taskID)
	}

	// Update task properties based on movement (priority/status changes)
	if updatedTask := bv.extractUpdatedTaskFromResponse(response, taskID); updatedTask != nil {
		taskToMove = updatedTask
	}

	// Remove from source column
	fromColumn.RemoveTask(taskID)

	// Add to destination column
	toColumn := bv.currentState.Columns[toColumnIndex]
	toColumn.AddTask(taskToMove)

	return nil
}

// extractUpdatedTaskFromResponse extracts updated task data from WorkflowManager response
func (bv *BoardView) extractUpdatedTaskFromResponse(response map[string]any, taskID string) *TaskData {
	if updatedTask, exists := response["updated_task"]; exists {
		if taskMap, ok := updatedTask.(map[string]interface{}); ok {
			return bv.mapResponseToTaskData(taskMap)
		}
	}
	return nil
}

// addTaskToBoardState adds a task to the board's task collection
func (bv *BoardView) addTaskToBoardState(task *TaskData) {
	newState := bv.copyCurrentState()
	newState.AllTasks = append(newState.AllTasks, task)
	bv.updateState(newState)
}

// removeTaskFromBoardState removes a task from the board's task collection
func (bv *BoardView) removeTaskFromBoardState(taskID string) {
	newState := bv.copyCurrentState()

	for i, task := range newState.AllTasks {
		if task.ID == taskID {
			newState.AllTasks = append(newState.AllTasks[:i], newState.AllTasks[i+1:]...)
			break
		}
	}

	bv.updateState(newState)
}

// updateColumnInBoardConfig updates a column configuration in the board config
func (bv *BoardView) updateColumnInBoardConfig(updatedConfig *ColumnConfiguration) {
	newState := bv.copyCurrentState()

	for i, columnConfig := range newState.Configuration.Columns {
		if columnConfig.Title == updatedConfig.Title {
			newState.Configuration.Columns[i] = updatedConfig
			break
		}
	}

	bv.updateState(newState)
}