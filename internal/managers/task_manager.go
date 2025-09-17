// Package managers provides Manager layer components implementing the iDesign methodology.
// This package contains components that orchestrate business workflows and coordinate
// between different components in the application architecture.
package managers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/resource_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// TaskRequest represents the input data for task operations
type TaskRequest struct {
	Description           string                   `json:"description"`
	Priority              resource_access.Priority `json:"priority"`
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
	Priority              resource_access.Priority `json:"priority"`
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
	Priority              *resource_access.Priority       `json:"priority,omitempty"`
	Tags                  []string                        `json:"tags,omitempty"`
	DateRange             *resource_access.DateRange      `json:"date_range,omitempty"`
	PriorityPromotionDate *resource_access.DateRange      `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string                         `json:"parent_task_id,omitempty"`
	Hierarchy             resource_access.HierarchyFilter `json:"hierarchy,omitempty"`
}

// ValidationResult represents the outcome of task validation
type ValidationResult struct {
	Valid      bool                    `json:"valid"`
	Violations []engines.RuleViolation `json:"violations,omitempty"`
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

	// IContext facet operations for UI context management
	IContext
}

// taskManager implements the TaskManager interface
type taskManager struct {
	mu          sync.RWMutex
	boardAccess resource_access.IBoardAccess
	ruleEngine  engines.IRuleEngine
	logger      utilities.ILoggingUtility
	boardPath   string
	IContext    // embedded context facet
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(boardAccess resource_access.IBoardAccess, ruleEngine engines.IRuleEngine, logger utilities.ILoggingUtility, repository utilities.Repository, boardPath string) TaskManager {
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
	task := &resource_access.Task{
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
	task := &resource_access.Task{
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
	err := tm.boardAccess.RemoveTask(taskID, resource_access.DeleteSubtasks) // Default cascade policy
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
	criteria := &resource_access.QueryCriteria{
		PriorityPromotionDate: &resource_access.DateRange{
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
			newPriority := resource_access.Priority{
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

// Helper methods

// validateTaskRequest validates a task request using the RuleEngine
func (tm *taskManager) validateTaskRequest(request TaskRequest) (ValidationResult, error) {
	// Create TaskEvent for rule validation
	futureState := &engines.TaskState{
		Task: &resource_access.Task{
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
		Task: &resource_access.Task{
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
func (tm *taskManager) convertToTaskResponse(taskWithTimestamps *resource_access.TaskWithTimestamps, subtasks []*resource_access.TaskWithTimestamps) TaskResponse {
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
func (tm *taskManager) convertToBoardCriteria(criteria QueryCriteria) *resource_access.QueryCriteria {
	return &resource_access.QueryCriteria{
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
func mapWorkflowStatusWithPriority(status WorkflowStatus, priority resource_access.Priority) resource_access.WorkflowStatus {
	result := resource_access.WorkflowStatus{
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
func mapFromBoardStatus(status resource_access.WorkflowStatus) WorkflowStatus {
	return WorkflowStatus(status.Column)
}
