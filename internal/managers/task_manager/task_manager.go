// Package managers provides Manager layer components implementing the iDesign methodology.
// This package contains components that orchestrate business workflows and coordinate
// between different components in the application architecture.
package task_manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// TaskRequest represents the input data for task operations
type TaskRequest struct {
	Description           string                   `json:"description"`
	Priority              board_access.Priority `json:"priority"`
	WorkflowStatus        WorkflowStatus           `json:"workflow_status"`
	Tags                  []string                 `json:"tags,omitempty"`
	Deadline              *time.Time               `json:"deadline,omitempty"`
	PriorityPromotionDate *time.Time               `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string                  `json:"parent_task_id,omitempty"`
}

// TaskResponse represents the output data from task operations
type TaskResponse struct {
	ID                    string                   `json:"id"`
	Description           string                   `json:"description"`
	Priority              board_access.Priority `json:"priority"`
	WorkflowStatus        WorkflowStatus           `json:"workflow_status"`
	Tags                  []string                 `json:"tags,omitempty"`
	Deadline              *time.Time               `json:"deadline,omitempty"`
	PriorityPromotionDate *time.Time               `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string                  `json:"parent_task_id,omitempty"`
	SubtaskIDs            []string                 `json:"subtask_ids,omitempty"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
}

// WorkflowStatus represents task workflow states
type WorkflowStatus string

const (
	Todo       WorkflowStatus = "todo"
	InProgress WorkflowStatus = "doing"
	Done       WorkflowStatus = "done"
)

// QueryCriteria defines search parameters for task queries
type QueryCriteria struct {
	Columns               []string                        `json:"columns,omitempty"`
	Sections              []string                        `json:"sections,omitempty"`
	Priority              *board_access.Priority       `json:"priority,omitempty"`
	Tags                  []string                        `json:"tags,omitempty"`
	DateRange             *board_access.DateRange      `json:"date_range,omitempty"`
	PriorityPromotionDate *board_access.DateRange      `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string                         `json:"parent_task_id,omitempty"`
	Hierarchy             board_access.HierarchyFilter `json:"hierarchy,omitempty"`
}

// ValidationResult represents the outcome of task validation
type ValidationResult struct {
	Valid      bool                    `json:"valid"`
	Violations []engines.RuleViolation `json:"violations,omitempty"`
}

// Board Management Types

// BoardValidationResponse represents the result of board directory validation
type BoardValidationResponse struct {
	IsValid       bool     `json:"is_valid"`
	GitRepoValid  bool     `json:"git_repo_valid"`
	ConfigValid   bool     `json:"config_valid"`
	DataIntegrity bool     `json:"data_integrity"`
	Issues        []string `json:"issues,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
}

// BoardMetadataResponse represents board metadata for UI display
type BoardMetadataResponse struct {
	Title         string            `json:"title"`
	Description   string            `json:"description,omitempty"`
	TaskCount     int               `json:"task_count"`
	ColumnCounts  map[string]int    `json:"column_counts"`
	SchemaVersion string            `json:"schema_version,omitempty"`
	CreatedAt     *time.Time        `json:"created_at,omitempty"`
	ModifiedAt    *time.Time        `json:"modified_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// BoardMetadataRequest represents board metadata update request
type BoardMetadataRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// BoardCreationRequest represents board creation request
type BoardCreationRequest struct {
	BoardPath     string            `json:"board_path"`
	Title         string            `json:"title"`
	Description   string            `json:"description,omitempty"`
	InitializeGit bool              `json:"initialize_git"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// BoardCreationResponse represents board creation result
type BoardCreationResponse struct {
	Success        bool   `json:"success"`
	BoardPath      string `json:"board_path"`
	ConfigPath     string `json:"config_path"`
	GitInitialized bool   `json:"git_initialized"`
	Message        string `json:"message,omitempty"`
}

// BoardDeletionRequest represents board deletion request
type BoardDeletionRequest struct {
	BoardPath      string `json:"board_path"`
	UseTrash       bool   `json:"use_trash"`
	CreateBackup   bool   `json:"create_backup"`
	BackupLocation string `json:"backup_location,omitempty"`
	ForceDelete    bool   `json:"force_delete"`
}

// BoardDeletionResponse represents board deletion result
type BoardDeletionResponse struct {
	Success        bool   `json:"success"`
	Method         string `json:"method"`
	BackupCreated  bool   `json:"backup_created"`
	BackupLocation string `json:"backup_location,omitempty"`
	Message        string `json:"message,omitempty"`
}

// TaskManager defines the interface for task workflow orchestration
type TaskManager interface {
	// Task CRUD Operations
	CreateTask(request TaskRequest) (TaskResponse, error)
	UpdateTask(taskID string, request TaskRequest) (TaskResponse, error)
	GetTask(taskID string) (TaskResponse, error)
	DeleteTask(taskID string) error

	// Task Query Operations
	ListTasks(criteria QueryCriteria) ([]TaskResponse, error)

	// Workflow Operations
	ChangeTaskStatus(taskID string, status WorkflowStatus) (TaskResponse, error)

	// Validation Operations
	ValidateTask(request TaskRequest) (ValidationResult, error)

	// Priority Promotion Operations
	ProcessPriorityPromotions() ([]TaskResponse, error)

	// Board Management Operations
	ValidateBoardDirectory(directoryPath string) (BoardValidationResponse, error)
	GetBoardMetadata(boardPath string) (BoardMetadataResponse, error)
	CreateBoard(request BoardCreationRequest) (BoardCreationResponse, error)
	UpdateBoardMetadata(boardPath string, metadata BoardMetadataRequest) (BoardMetadataResponse, error)
	DeleteBoard(request BoardDeletionRequest) (BoardDeletionResponse, error)

	// IContext facet operations for UI context management
	IContext
}

// taskManager implements the TaskManager interface
type taskManager struct {
	mu          sync.RWMutex
	boardAccess board_access.IBoardAccess
	ruleEngine  engines.IRuleEngine
	logger      utilities.ILoggingUtility
	boardPath   string
	IContext    // embedded context facet
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(boardAccess board_access.IBoardAccess, ruleEngine engines.IRuleEngine, logger utilities.ILoggingUtility, repository utilities.Repository, boardPath string) TaskManager {
	return &taskManager{
		boardAccess: boardAccess,
		ruleEngine:  ruleEngine,
		logger:      logger,
		boardPath:   boardPath,
		IContext:    newContextFacet(repository),
	}
}

// CreateTask implements task creation with validation and business rule checking
func (tm *taskManager) CreateTask(request TaskRequest) (TaskResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", "Creating new task")

	// Validate business rules
	validationResult, err := tm.validateTaskRequest(request)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("task creation validation failed: %w", err)
	}
	if !validationResult.Valid {
		return TaskResponse{}, fmt.Errorf("task creation violates business rules: %v", validationResult.Violations)
	}

	// Create Task struct for BoardAccess
	task := &board_access.Task{
		Title:                 request.Description, // Using description as title for now
		Description:           request.Description,
		Tags:                  request.Tags,
		DueDate:               request.Deadline,
		PriorityPromotionDate: request.PriorityPromotionDate,
		ParentTaskID:          request.ParentTaskID,
	}

	// Store task through BoardAccess
	taskID, err := tm.boardAccess.CreateTask(task, request.Priority, mapWorkflowStatusWithPriority(request.WorkflowStatus, request.Priority), request.ParentTaskID)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("task creation failed in storage: %w", err)
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Task created successfully: %s", taskID))

	// Retrieve the created task to return complete information
	return tm.getTaskInternal(taskID)
}

// UpdateTask implements task modification with validation
func (tm *taskManager) UpdateTask(taskID string, request TaskRequest) (TaskResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Updating task: %s", taskID))

	// Validate business rules
	validationResult, err := tm.validateTaskRequest(request)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("task update validation failed: %w", err)
	}
	if !validationResult.Valid {
		return TaskResponse{}, fmt.Errorf("task update violates business rules: %v", validationResult.Violations)
	}

	// Create updated Task struct
	task := &board_access.Task{
		ID:                    taskID,
		Title:                 request.Description,
		Description:           request.Description,
		Tags:                  request.Tags,
		DueDate:               request.Deadline,
		PriorityPromotionDate: request.PriorityPromotionDate,
		ParentTaskID:          request.ParentTaskID,
	}

	// Update task through BoardAccess
	err = tm.boardAccess.ChangeTaskData(taskID, task, request.Priority, mapWorkflowStatusWithPriority(request.WorkflowStatus, request.Priority))
	if err != nil {
		return TaskResponse{}, fmt.Errorf("task update failed in storage: %w", err)
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Task updated successfully: %s", taskID))

	// Return updated task information
	return tm.getTaskInternal(taskID)
}

// GetTask retrieves a single task by ID
func (tm *taskManager) GetTask(taskID string) (TaskResponse, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.getTaskInternal(taskID)
}

// getTaskInternal is the internal implementation without locking
func (tm *taskManager) getTaskInternal(taskID string) (TaskResponse, error) {
	// Retrieve task from BoardAccess
	taskWithTimestamps, err := tm.boardAccess.GetTasksData([]string{taskID}, true)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to retrieve task %s: %w", taskID, err)
	}
	if len(taskWithTimestamps) == 0 {
		return TaskResponse{}, fmt.Errorf("task not found: %s", taskID)
	}

	// Get subtasks
	subtasks, err := tm.boardAccess.GetSubtasks(taskID)
	if err != nil {
		// Log error but don't fail the operation
		tm.logger.LogMessage(utilities.Warning, "TaskManager", fmt.Sprintf("Failed to retrieve subtasks for %s: %v", taskID, err))
	}

	return tm.convertToTaskResponse(taskWithTimestamps[0], subtasks), nil
}

// DeleteTask implements task deletion with cascade handling
func (tm *taskManager) DeleteTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Deleting task: %s", taskID))

	// Handle cascade operations for subtasks (implementation depends on cascade policy)
	err := tm.boardAccess.RemoveTask(taskID, board_access.DeleteSubtasks) // Default cascade policy
	if err != nil {
		return fmt.Errorf("task deletion failed: %w", err)
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Task deleted successfully: %s", taskID))
	return nil
}

// ListTasks retrieves tasks matching the given criteria
func (tm *taskManager) ListTasks(criteria QueryCriteria) ([]TaskResponse, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tm.logger.LogMessage(utilities.Debug, "TaskManager", "Listing tasks")

	// Convert criteria to BoardAccess format
	boardCriteria := tm.convertToBoardCriteria(criteria)

	// Query tasks through BoardAccess
	tasksWithTimestamps, err := tm.boardAccess.FindTasks(boardCriteria)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Convert to TaskResponse format
	responses := make([]TaskResponse, 0, len(tasksWithTimestamps))
	for _, taskWithTimestamps := range tasksWithTimestamps {
		// Get subtasks for each task
		subtasks, err := tm.boardAccess.GetSubtasks(taskWithTimestamps.Task.ID)
		if err != nil {
			tm.logger.LogMessage(utilities.Warning, "TaskManager", fmt.Sprintf("Failed to retrieve subtasks for %s: %v", taskWithTimestamps.Task.ID, err))
			subtasks = nil // Continue without subtasks
		}
		responses = append(responses, tm.convertToTaskResponse(taskWithTimestamps, subtasks))
	}

	return responses, nil
}

// ChangeTaskStatus implements workflow status changes with subtask coupling
func (tm *taskManager) ChangeTaskStatus(taskID string, status WorkflowStatus) (TaskResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Changing task status: %s to %s", taskID, status))

	// Get current task state
	currentTask, err := tm.getTaskInternal(taskID)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("failed to get current task state: %w", err)
	}

	// Validate workflow transition with RuleEngine
	if err := tm.validateWorkflowTransition(currentTask, status); err != nil {
		return TaskResponse{}, fmt.Errorf("workflow transition validation failed: %w", err)
	}

	// Handle subtask workflow coupling
	if err := tm.orchestrateSubtaskWorkflowCoupling(currentTask, status); err != nil {
		return TaskResponse{}, fmt.Errorf("subtask workflow coupling failed: %w", err)
	}

	// Apply the status change
	boardStatus := mapWorkflowStatusWithPriority(status, currentTask.Priority)
	err = tm.boardAccess.MoveTask(taskID, currentTask.Priority, boardStatus)
	if err != nil {
		return TaskResponse{}, fmt.Errorf("status change failed in storage: %w", err)
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Task status changed successfully: %s", taskID))

	// Return updated task
	return tm.getTaskInternal(taskID)
}

// ValidateTask validates task data without persistence
func (tm *taskManager) ValidateTask(request TaskRequest) (ValidationResult, error) {
	tm.logger.LogMessage(utilities.Debug, "TaskManager", "Validating task data")

	return tm.validateTaskRequest(request)
}

// ProcessPriorityPromotions automatically escalates tasks with reached promotion dates
func (tm *taskManager) ProcessPriorityPromotions() ([]TaskResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", "Processing priority promotions")

	// Query tasks with promotion dates that have been reached
	now := time.Now()
	criteria := &board_access.QueryCriteria{
		PriorityPromotionDate: &board_access.DateRange{
			To: &now, // Tasks with promotion date <= now
		},
	}

	tasksToPromote, err := tm.boardAccess.FindTasks(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks for promotion: %w", err)
	}

	var promotedTasks []TaskResponse

	for _, taskWithTimestamps := range tasksToPromote {
		// Only promote tasks that are not-urgent-important
		if !taskWithTimestamps.Priority.Urgent && taskWithTimestamps.Priority.Important {
			tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Promoting task priority: %s", taskWithTimestamps.Task.ID))

			// Create new priority: urgent-important
			newPriority := board_access.Priority{
				Urgent:    true,
				Important: true,
				Label:     "urgent-important",
			}

			// Clear the promotion date
			updatedTask := *taskWithTimestamps.Task
			updatedTask.PriorityPromotionDate = nil

			// Calculate new status with updated priority section
			currentStatus := mapFromBoardStatus(taskWithTimestamps.Status)
			newStatus := mapWorkflowStatusWithPriority(currentStatus, newPriority)
			
			// Update the task
			err := tm.boardAccess.ChangeTaskData(taskWithTimestamps.Task.ID, &updatedTask, newPriority, newStatus)
			if err != nil {
				tm.logger.LogMessage(utilities.Error, "TaskManager", fmt.Sprintf("Failed to promote task %s: %v", taskWithTimestamps.Task.ID, err))
				continue
			}

			// Get updated task for response
			promotedTask, err := tm.getTaskInternal(taskWithTimestamps.Task.ID)
			if err != nil {
				tm.logger.LogMessage(utilities.Warning, "TaskManager", fmt.Sprintf("Failed to retrieve promoted task %s: %v", taskWithTimestamps.Task.ID, err))
				continue
			}

			promotedTasks = append(promotedTasks, promotedTask)
		}
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Processed %d priority promotions", len(promotedTasks)))

	return promotedTasks, nil
}

// Board Management Operations

// ValidateBoardDirectory validates that a directory can be used as a board (OP-9)
func (tm *taskManager) ValidateBoardDirectory(directoryPath string) (BoardValidationResponse, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Validating board directory: %s", directoryPath))

	// Delegate to BoardAccess for validation
	ctx := context.Background()
	validationResult, err := tm.boardAccess.ValidateStructure(ctx, directoryPath)
	if err != nil {
		return BoardValidationResponse{}, fmt.Errorf("board directory validation failed: %w", err)
	}

	// Convert to TaskManager response format
	response := BoardValidationResponse{
		IsValid:       validationResult.IsValid,
		GitRepoValid:  validationResult.GitRepoValid,
		ConfigValid:   validationResult.ConfigValid,
		DataIntegrity: validationResult.DataIntegrity,
		Issues:        make([]string, 0),
		Warnings:      make([]string, 0),
	}

	// Convert issues and warnings
	for _, issue := range validationResult.Issues {
		response.Issues = append(response.Issues, issue.Message)
	}
	for _, warning := range validationResult.Warnings {
		response.Warnings = append(response.Warnings, warning.Message)
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Board directory validation completed: %s (valid: %v)", directoryPath, response.IsValid))

	return response, nil
}

// GetBoardMetadata extracts board metadata for display (OP-10)
func (tm *taskManager) GetBoardMetadata(boardPath string) (BoardMetadataResponse, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Extracting board metadata: %s", boardPath))

	// Delegate to BoardAccess for metadata extraction
	ctx := context.Background()
	metadata, err := tm.boardAccess.ExtractMetadata(ctx, boardPath)
	if err != nil {
		return BoardMetadataResponse{}, fmt.Errorf("board metadata extraction failed: %w", err)
	}

	// Convert to TaskManager response format
	response := BoardMetadataResponse{
		Title:         metadata.Title,
		Description:   metadata.Description,
		TaskCount:     metadata.TaskCount,
		ColumnCounts:  metadata.ColumnCounts,
		SchemaVersion: metadata.SchemaVersion,
		CreatedAt:     metadata.CreatedAt,
		ModifiedAt:    metadata.ModifiedAt,
		Metadata:      metadata.Metadata,
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Board metadata extracted: %s (title: %s)", boardPath, response.Title))

	return response, nil
}

// CreateBoard creates a new board with validation (OP-11)
func (tm *taskManager) CreateBoard(request BoardCreationRequest) (BoardCreationResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Creating board: %s", request.BoardPath))

	// Validate board creation request with RuleEngine
	if err := tm.validateBoardCreation(request); err != nil {
		return BoardCreationResponse{}, fmt.Errorf("board creation validation failed: %w", err)
	}

	// Convert to BoardAccess format
	boardRequest := &board_access.BoardCreationRequest{
		BoardPath:     request.BoardPath,
		Title:         request.Title,
		Description:   request.Description,
		InitializeGit: request.InitializeGit,
		Metadata:      request.Metadata,
	}

	// Delegate to BoardAccess for creation
	ctx := context.Background()
	creationResult, err := tm.boardAccess.Create(ctx, boardRequest)
	if err != nil {
		return BoardCreationResponse{}, fmt.Errorf("board creation failed: %w", err)
	}

	// Convert to TaskManager response format
	response := BoardCreationResponse{
		Success:        creationResult.Success,
		BoardPath:      creationResult.BoardPath,
		ConfigPath:     creationResult.ConfigPath,
		GitInitialized: creationResult.GitInitialized,
		Message:        creationResult.Message,
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Board created successfully: %s", request.BoardPath))

	return response, nil
}

// UpdateBoardMetadata updates board metadata with validation (OP-12)
func (tm *taskManager) UpdateBoardMetadata(boardPath string, metadata BoardMetadataRequest) (BoardMetadataResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Updating board metadata: %s", boardPath))

	// Validate board metadata update with RuleEngine
	if err := tm.validateBoardMetadataUpdate(boardPath, metadata); err != nil {
		return BoardMetadataResponse{}, fmt.Errorf("board metadata update validation failed: %w", err)
	}

	// Get current board configuration
	currentConfig, err := tm.boardAccess.GetBoardConfiguration()
	if err != nil {
		return BoardMetadataResponse{}, fmt.Errorf("failed to get current board configuration: %w", err)
	}

	// Update configuration with new metadata
	updatedConfig := *currentConfig
	updatedConfig.Name = metadata.Title
	// Note: Description is not part of BoardConfiguration, handled differently

	// Store updated configuration
	err = tm.boardAccess.UpdateBoardConfiguration(&updatedConfig)
	if err != nil {
		return BoardMetadataResponse{}, fmt.Errorf("failed to update board configuration: %w", err)
	}

	// Return updated metadata (call internal method to avoid deadlock)
	ctx := context.Background()
	updatedMetadata, err := tm.boardAccess.ExtractMetadata(ctx, boardPath)
	if err != nil {
		return BoardMetadataResponse{}, fmt.Errorf("failed to extract updated metadata: %w", err)
	}

	// Convert to TaskManager response format
	return BoardMetadataResponse{
		Title:         updatedMetadata.Title,
		Description:   updatedMetadata.Description,
		TaskCount:     updatedMetadata.TaskCount,
		ColumnCounts:  updatedMetadata.ColumnCounts,
		SchemaVersion: updatedMetadata.SchemaVersion,
		CreatedAt:     updatedMetadata.CreatedAt,
		ModifiedAt:    updatedMetadata.ModifiedAt,
		Metadata:      updatedMetadata.Metadata,
	}, nil
}

// DeleteBoard deletes a board with validation (OP-13)
func (tm *taskManager) DeleteBoard(request BoardDeletionRequest) (BoardDeletionResponse, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Deleting board: %s", request.BoardPath))

	// Validate board deletion request
	if err := tm.validateBoardDeletion(request); err != nil {
		return BoardDeletionResponse{}, fmt.Errorf("board deletion validation failed: %w", err)
	}

	// Convert to BoardAccess format
	deletionRequest := &board_access.BoardDeletionRequest{
		BoardPath:       request.BoardPath,
		UseTrash:        request.UseTrash,
		CreateBackup:    request.CreateBackup,
		BackupLocation:  request.BackupLocation,
		ForceDelete:     request.ForceDelete,
	}

	// Delegate to BoardAccess for deletion
	ctx := context.Background()
	deletionResult, err := tm.boardAccess.Delete(ctx, deletionRequest)
	if err != nil {
		return BoardDeletionResponse{}, fmt.Errorf("board deletion failed: %w", err)
	}

	// Convert to TaskManager response format
	response := BoardDeletionResponse{
		Success:        deletionResult.Success,
		Method:         deletionResult.Method,
		BackupCreated:  deletionResult.BackupCreated,
		BackupLocation: deletionResult.BackupLocation,
		Message:        deletionResult.Message,
	}

	tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Board deleted successfully: %s", request.BoardPath))

	return response, nil
}

// Helper methods

// validateTaskRequest validates a task request using the RuleEngine
func (tm *taskManager) validateTaskRequest(request TaskRequest) (ValidationResult, error) {
	// Create TaskEvent for rule validation
	futureState := &engines.TaskState{
		Task: &board_access.Task{
			Title:                 request.Description,
			Description:           request.Description,
			Tags:                  request.Tags,
			DueDate:               request.Deadline,
			PriorityPromotionDate: request.PriorityPromotionDate,
			ParentTaskID:          request.ParentTaskID,
		},
		Priority: request.Priority,
		Status:   mapWorkflowStatusWithPriority(request.WorkflowStatus, request.Priority),
	}

	event := engines.TaskEvent{
		EventType:   "task_create",
		FutureState: futureState,
		Timestamp:   time.Now(),
	}

	// Validate with RuleEngine
	result, err := tm.ruleEngine.EvaluateTaskChange(context.Background(), event, tm.boardPath)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("rule validation failed: %w", err)
	}

	return ValidationResult{
		Valid:      result.Allowed,
		Violations: result.Violations,
	}, nil
}

// validateWorkflowTransition validates workflow status transitions
func (tm *taskManager) validateWorkflowTransition(currentTask TaskResponse, newStatus WorkflowStatus) error {
	// Create TaskEvent for workflow transition validation
	futureState := &engines.TaskState{
		Task: &board_access.Task{
			ID:                    currentTask.ID,
			Title:                 currentTask.Description,
			Description:           currentTask.Description,
			Tags:                  currentTask.Tags,
			DueDate:               currentTask.Deadline,
			PriorityPromotionDate: currentTask.PriorityPromotionDate,
			ParentTaskID:          currentTask.ParentTaskID,
		},
		Priority: currentTask.Priority,
		Status:   mapWorkflowStatusWithPriority(newStatus, currentTask.Priority),
	}

	event := engines.TaskEvent{
		EventType:   "task_transition",
		FutureState: futureState,
		Timestamp:   time.Now(),
	}

	// Validate with RuleEngine
	result, err := tm.ruleEngine.EvaluateTaskChange(context.Background(), event, tm.boardPath)
	if err != nil {
		return fmt.Errorf("workflow transition rule validation failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("workflow transition violates business rules: %v", result.Violations)
	}

	return nil
}

// orchestrateSubtaskWorkflowCoupling handles parent-child workflow coupling
func (tm *taskManager) orchestrateSubtaskWorkflowCoupling(task TaskResponse, newStatus WorkflowStatus) error {
	// Implementation of subtask workflow coupling logic
	// This will be based on the requirements REQ-TASKMANAGER-016 and REQ-TASKMANAGER-017

	// If this is a subtask moving from "todo" to "doing"
	if task.ParentTaskID != nil && task.WorkflowStatus == Todo && newStatus == InProgress {
		return tm.handleFirstSubtaskTransition(*task.ParentTaskID)
	}

	// If this is a parent task moving to "done"
	if task.ParentTaskID == nil && newStatus == Done {
		return tm.handleParentTaskCompletion(task.ID)
	}

	return nil
}

// handleFirstSubtaskTransition handles the first subtask moving to "doing"
func (tm *taskManager) handleFirstSubtaskTransition(parentTaskID string) error {
	// Get parent task
	parentTask, err := tm.getTaskInternal(parentTaskID)
	if err != nil {
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	// If parent is in "todo", move it to "doing"
	if parentTask.WorkflowStatus == Todo {
		tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Moving parent task %s from todo to doing due to subtask transition", parentTaskID))

		boardStatus := mapWorkflowStatusWithPriority(InProgress, parentTask.Priority)
		err = tm.boardAccess.MoveTask(parentTaskID, parentTask.Priority, boardStatus)
		if err != nil {
			return fmt.Errorf("failed to update parent task status: %w", err)
		}
	}

	return nil
}

// handleParentTaskCompletion handles parent task moving to "done"
func (tm *taskManager) handleParentTaskCompletion(parentTaskID string) error {
	// Get all subtasks
	subtasks, err := tm.boardAccess.GetSubtasks(parentTaskID)
	if err != nil {
		return fmt.Errorf("failed to get subtasks: %w", err)
	}

	// Move all subtasks to "done"
	for _, subtask := range subtasks {
		if subtask.Status.Column != "done" {
			tm.logger.LogMessage(utilities.Info, "TaskManager", fmt.Sprintf("Moving subtask %s to done due to parent completion", subtask.Task.ID))

			boardStatus := mapWorkflowStatusWithPriority(Done, subtask.Priority)
			err = tm.boardAccess.MoveTask(subtask.Task.ID, subtask.Priority, boardStatus)
			if err != nil {
				tm.logger.LogMessage(utilities.Error, "TaskManager", fmt.Sprintf("Failed to update subtask %s: %v", subtask.Task.ID, err))
				// Continue with other subtasks
			}
		}
	}

	return nil
}

// convertToTaskResponse converts BoardAccess types to TaskManager response format
func (tm *taskManager) convertToTaskResponse(taskWithTimestamps *board_access.TaskWithTimestamps, subtasks []*board_access.TaskWithTimestamps) TaskResponse {
	subtaskIDs := make([]string, 0, len(subtasks))
	for _, subtask := range subtasks {
		subtaskIDs = append(subtaskIDs, subtask.Task.ID)
	}

	return TaskResponse{
		ID:                    taskWithTimestamps.Task.ID,
		Description:           taskWithTimestamps.Task.Description,
		Priority:              taskWithTimestamps.Priority,
		WorkflowStatus:        mapFromBoardStatus(taskWithTimestamps.Status),
		Tags:                  taskWithTimestamps.Task.Tags,
		Deadline:              taskWithTimestamps.Task.DueDate,
		PriorityPromotionDate: taskWithTimestamps.Task.PriorityPromotionDate,
		CreatedAt:             taskWithTimestamps.CreatedAt,
		UpdatedAt:             taskWithTimestamps.UpdatedAt,
		ParentTaskID:          taskWithTimestamps.Task.ParentTaskID,
		SubtaskIDs:            subtaskIDs,
	}
}

// convertToBoardCriteria converts TaskManager criteria to BoardAccess format
func (tm *taskManager) convertToBoardCriteria(criteria QueryCriteria) *board_access.QueryCriteria {
	return &board_access.QueryCriteria{
		Columns:               criteria.Columns,
		Sections:              criteria.Sections,
		Priority:              criteria.Priority,
		Tags:                  criteria.Tags,
		DateRange:             criteria.DateRange,
		PriorityPromotionDate: criteria.PriorityPromotionDate,
		ParentTaskID:          criteria.ParentTaskID,
		Hierarchy:             criteria.Hierarchy,
	}
}

// mapWorkflowStatusWithPriority converts TaskManager WorkflowStatus to BoardAccess WorkflowStatus with priority-based section
func mapWorkflowStatusWithPriority(status WorkflowStatus, priority board_access.Priority) board_access.WorkflowStatus {
	result := board_access.WorkflowStatus{
		Column: string(status),
	}
	
	// For todo column, set section based on priority
	if status == Todo {
		if priority.Urgent && priority.Important {
			result.Section = "urgent-important"
		} else if priority.Urgent && !priority.Important {
			result.Section = "urgent-not-important"
		} else if !priority.Urgent && priority.Important {
			result.Section = "not-urgent-important"
		}
	}
	
	return result
}

// mapFromBoardStatus converts BoardAccess WorkflowStatus to TaskManager WorkflowStatus
func mapFromBoardStatus(status board_access.WorkflowStatus) WorkflowStatus {
	return WorkflowStatus(status.Column)
}

// Board validation helper methods

// validateBoardCreation validates board creation request using RuleEngine
func (tm *taskManager) validateBoardCreation(request BoardCreationRequest) error {
	// Create BoardConfigurationEvent for rule validation
	config := &engines.BoardConfiguration{
		Title:       request.Title,
		Description: request.Description,
		Metadata:    request.Metadata,
	}

	event := engines.BoardConfigurationEvent{
		EventType:     "board_create",
		Configuration: config,
		Timestamp:     time.Now(),
	}

	// Validate with RuleEngine
	result, err := tm.ruleEngine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		return fmt.Errorf("board creation rule validation failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("board creation violates business rules: %v", result.Violations)
	}

	return nil
}

// validateBoardMetadataUpdate validates board metadata update using RuleEngine
func (tm *taskManager) validateBoardMetadataUpdate(boardPath string, metadata BoardMetadataRequest) error {
	// Create engines configuration for validation
	config := &engines.BoardConfiguration{
		Title:       metadata.Title,
		Description: metadata.Description,
		Metadata:    metadata.Metadata,
	}

	event := engines.BoardConfigurationEvent{
		EventType:     "board_update",
		Configuration: config,
		Timestamp:     time.Now(),
	}

	// Validate with RuleEngine
	result, err := tm.ruleEngine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		return fmt.Errorf("board metadata update rule validation failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("board metadata update violates business rules: %v", result.Violations)
	}

	return nil
}

// validateBoardDeletion validates board deletion request
func (tm *taskManager) validateBoardDeletion(request BoardDeletionRequest) error {
	// Basic validation - ensure board path is not empty
	if request.BoardPath == "" {
		return fmt.Errorf("board path cannot be empty")
	}

	// Additional validation could include checking for active tasks, etc.
	// For now, delegate validation to BoardAccess layer

	return nil
}
