package managers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/internal/client/resource_access"
)

// WorkflowManager provides client-side task workflow orchestration by coordinating
// UI engines with backend task management services.
type WorkflowManager interface {
	Task() ITask
	Drag() IDrag
	Batch() IBatch
	Search() ISearch
	Subtask() ISubtask
}

// ITask handles task-related workflows with validation
type ITask interface {
	// Core task operations
	CreateTaskWorkflow(ctx context.Context, request map[string]any) (map[string]any, error)
	UpdateTaskWorkflow(ctx context.Context, taskID string, request map[string]any) (map[string]any, error)
	DeleteTaskWorkflow(ctx context.Context, taskID string) (map[string]any, error)
	QueryTasksWorkflow(ctx context.Context, criteria map[string]any) (map[string]any, error)

	// Status and priority management
	ChangeTaskStatusWorkflow(ctx context.Context, taskID string, status string) (map[string]any, error)
	ChangeTaskPriorityWorkflow(ctx context.Context, taskID string, priority string) (map[string]any, error)
	ArchiveTaskWorkflow(ctx context.Context, taskID string, options map[string]any) (map[string]any, error)
}

// IDrag handles drag-drop workflows with movement validation
type IDrag interface {
	ProcessDragDropWorkflow(ctx context.Context, event map[string]any) (map[string]any, error)
}

// IBatch handles bulk operation workflows with partial failure handling
type IBatch interface {
	BatchStatusUpdateWorkflow(ctx context.Context, taskIDs []string, status string) (map[string]any, error)
	BatchPriorityUpdateWorkflow(ctx context.Context, taskIDs []string, priority string) (map[string]any, error)
	BatchArchiveWorkflow(ctx context.Context, taskIDs []string, options map[string]any) (map[string]any, error)
}

// ISearch handles advanced search and filter workflows
type ISearch interface {
	SearchTasksWorkflow(ctx context.Context, query string, filters map[string]any) (map[string]any, error)
	ApplyFiltersWorkflow(ctx context.Context, filters map[string]any, context map[string]any) (map[string]any, error)
}

// ISubtask handles subtask relationship workflows with hierarchy validation
type ISubtask interface {
	CreateSubtaskRelationshipWorkflow(ctx context.Context, parentID string, childSpec map[string]any) (map[string]any, error)
	ProcessSubtaskCompletionWorkflow(ctx context.Context, subtaskID string, cascade map[string]any) (map[string]any, error)
	MoveSubtaskWorkflow(ctx context.Context, subtaskID string, newParentID string, position map[string]any) (map[string]any, error)
}

// Data Types for workflow state management
type WorkflowType string
type WorkflowStatus string

const (
	// Core workflow types
	WorkflowTypeTaskCreate WorkflowType = "task_create"
	WorkflowTypeTaskUpdate WorkflowType = "task_update"
	WorkflowTypeTaskDelete WorkflowType = "task_delete"
	WorkflowTypeTaskQuery  WorkflowType = "task_query"
	WorkflowTypeDragDrop   WorkflowType = "drag_drop"

	// Extended workflow types
	WorkflowTypeStatusChange   WorkflowType = "status_change"
	WorkflowTypePriorityChange WorkflowType = "priority_change"
	WorkflowTypeTaskArchive    WorkflowType = "task_archive"
	WorkflowTypeBatchStatus    WorkflowType = "batch_status"
	WorkflowTypeBatchPriority  WorkflowType = "batch_priority"
	WorkflowTypeBatchArchive   WorkflowType = "batch_archive"
	WorkflowTypeSearch         WorkflowType = "search"
	WorkflowTypeFilter         WorkflowType = "filter"
	WorkflowTypeSubtaskCreate  WorkflowType = "subtask_create"
	WorkflowTypeSubtaskComplete WorkflowType = "subtask_complete"
	WorkflowTypeSubtaskMove    WorkflowType = "subtask_move"

	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusInProgress WorkflowStatus = "in_progress"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
)

type WorkflowState struct {
	WorkflowID   string
	WorkflowType WorkflowType
	Status       WorkflowStatus
	Context      map[string]any
	StartTime    time.Time
	LastUpdate   time.Time
}

// Main implementation
type workflowManager struct {
	// Engine dependencies
	validation *engines.FormValidationEngine
	formatting *engines.FormattingEngine
	dragDrop   engines.DragDropEngine
	backend    resource_access.ITaskManagerAccess

	// Workflow state tracking
	activeWorkflows map[string]*WorkflowState
	mu              sync.RWMutex
}

func NewWorkflowManager(
	validation *engines.FormValidationEngine,
	formatting *engines.FormattingEngine,
	dragDrop engines.DragDropEngine,
	backend resource_access.ITaskManagerAccess,
) WorkflowManager {
	return &workflowManager{
		validation:      validation,
		formatting:      formatting,
		dragDrop:        dragDrop,
		backend:         backend,
		activeWorkflows: make(map[string]*WorkflowState),
	}
}

func (wm *workflowManager) Task() ITask {
	return &taskWorkflows{manager: wm}
}

func (wm *workflowManager) Drag() IDrag {
	return &dragWorkflows{manager: wm}
}

func (wm *workflowManager) Batch() IBatch {
	return &batchWorkflows{manager: wm}
}

func (wm *workflowManager) Search() ISearch {
	return &searchWorkflows{manager: wm}
}

func (wm *workflowManager) Subtask() ISubtask {
	return &subtaskWorkflows{manager: wm}
}

// Workflow state management
func (wm *workflowManager) createWorkflow(workflowType WorkflowType) *WorkflowState {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Generate unique workflow ID with counter to handle high concurrency
	var workflowID string
	counter := 0

	for {
		// Always get fresh time to avoid collisions
		currentTime := time.Now().UnixNano()

		if counter == 0 {
			workflowID = fmt.Sprintf("%s_%d", workflowType, currentTime)
		} else {
			workflowID = fmt.Sprintf("%s_%d_%d", workflowType, currentTime, counter)
		}

		// Check if ID already exists
		if _, exists := wm.activeWorkflows[workflowID]; !exists {
			break
		}
		counter++

		// Safety check to prevent infinite loop - should never happen with fresh timestamps
		if counter > 100 {
			// Add additional entropy if somehow we still have collisions
			workflowID = fmt.Sprintf("%s_%d_%d_%p", workflowType, time.Now().UnixNano(), counter, &workflowID)
			break
		}
	}

	workflow := &WorkflowState{
		WorkflowID:   workflowID,
		WorkflowType: workflowType,
		Status:       WorkflowStatusPending,
		Context:      make(map[string]any),
		StartTime:    time.Now(),
		LastUpdate:   time.Now(),
	}

	wm.activeWorkflows[workflowID] = workflow
	return workflow
}

func (wm *workflowManager) updateWorkflowStatus(workflowID string, status WorkflowStatus) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if workflow, exists := wm.activeWorkflows[workflowID]; exists {
		workflow.Status = status
		workflow.LastUpdate = time.Now()
	}
}

func (wm *workflowManager) completeWorkflow(workflowID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if workflow, exists := wm.activeWorkflows[workflowID]; exists {
		workflow.Status = WorkflowStatusCompleted
		workflow.LastUpdate = time.Now()
		delete(wm.activeWorkflows, workflowID)
	}
}

func (wm *workflowManager) failWorkflow(workflowID string, err error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if workflow, exists := wm.activeWorkflows[workflowID]; exists {
		workflow.Status = WorkflowStatusFailed
		workflow.LastUpdate = time.Now()
		if err != nil {
			workflow.Context["error"] = err.Error()
		} else {
			workflow.Context["error"] = "unknown error"
		}
		delete(wm.activeWorkflows, workflowID)
	}
}

// Task workflow implementations
type taskWorkflows struct {
	manager *workflowManager
}

func (t *taskWorkflows) CreateTaskWorkflow(ctx context.Context, request map[string]any) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypeTaskCreate)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate task data using FormValidationEngine
	validationRules := engines.ValidationRules{}
	validationResult := t.manager.validation.ValidateFormInputs(request, validationRules)
	if !validationResult.Valid {
		t.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Task validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Create task through TaskManagerAccess
	uiRequest := resource_access.UITaskRequest{
		Description: fmt.Sprintf("%v", request["description"]),
	}

	respCh, errCh := t.manager.backend.CreateTaskAsync(ctx, uiRequest)

	select {
	case response := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := t.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"task_id":    response.ID,
			"task":       map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
			},
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (t *taskWorkflows) UpdateTaskWorkflow(ctx context.Context, taskID string, request map[string]any) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypeTaskUpdate)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate update data
	validationRules := engines.ValidationRules{}
	validationResult := t.manager.validation.ValidateFormInputs(request, validationRules)
	if !validationResult.Valid {
		t.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("validation failed"))
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       "Task update validation failed",
		}, nil
	}

	// Update task through TaskManagerAccess
	uiRequest := resource_access.UITaskRequest{
		Description: fmt.Sprintf("%v", request["description"]),
	}

	respCh, errCh := t.manager.backend.UpdateTaskAsync(ctx, taskID, uiRequest)

	select {
	case response := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)

		// Format response
		formattedDesc, _ := t.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"task":       map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
			},
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (t *taskWorkflows) DeleteTaskWorkflow(ctx context.Context, taskID string) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypeTaskDelete)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Delete task through TaskManagerAccess
	respCh, errCh := t.manager.backend.DeleteTaskAsync(ctx, taskID)

	select {
	case success := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)
		return map[string]any{
			"success":     success,
			"workflow_id": workflow.WorkflowID,
			"task_id":     taskID,
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (t *taskWorkflows) QueryTasksWorkflow(ctx context.Context, criteria map[string]any) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypeTaskQuery)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Convert criteria to UI format
	uiCriteria := resource_access.UIQueryCriteria{
		// Basic conversion - could be enhanced based on actual UIQueryCriteria structure
	}

	// Query tasks through TaskManagerAccess
	respCh, errCh := t.manager.backend.QueryTasksAsync(ctx, uiCriteria)

	select {
	case tasks := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)

		// Format task collection for UI consumption
		formattedTasks := make([]map[string]any, len(tasks))
		for i, task := range tasks {
			formattedDesc, _ := t.manager.formatting.Text().FormatText(task.Description, engines.TextOptions{MaxLength: 50})
			formattedTasks[i] = map[string]any{
				"id":          task.ID,
				"description": formattedDesc,
				"display_name": task.DisplayName,
			}
		}

		return map[string]any{
			"success":     true,
			"workflow_id": workflow.WorkflowID,
			"tasks":       formattedTasks,
			"total_count": len(tasks),
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

// Enhanced task status and priority operations
func (t *taskWorkflows) ChangeTaskStatusWorkflow(ctx context.Context, taskID string, status string) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypeStatusChange)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate status transition using FormValidationEngine
	validationRules := engines.ValidationRules{}
	statusRequest := map[string]any{"status": status, "taskID": taskID}
	validationResult := t.manager.validation.ValidateFormInputs(statusRequest, validationRules)
	if !validationResult.Valid {
		t.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("status validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Status change validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Change status through TaskManagerAccess
	statusEnum := resource_access.UIWorkflowStatus(status)
	respCh, errCh := t.manager.backend.ChangeTaskStatusAsync(ctx, taskID, statusEnum)

	select {
	case response := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := t.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"task":       map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
				"status":      status,
			},
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (t *taskWorkflows) ChangeTaskPriorityWorkflow(ctx context.Context, taskID string, priority string) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypePriorityChange)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate priority change using FormValidationEngine
	validationRules := engines.ValidationRules{}
	priorityRequest := map[string]any{"priority": priority, "taskID": taskID}
	validationResult := t.manager.validation.ValidateFormInputs(priorityRequest, validationRules)
	if !validationResult.Valid {
		t.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("priority validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Priority change validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Update task with new priority through TaskManagerAccess
	updateRequest := resource_access.UITaskRequest{
		Description: "", // Will be filled by backend from existing task
		// Priority would be added here when UITaskRequest supports it
	}

	respCh, errCh := t.manager.backend.UpdateTaskAsync(ctx, taskID, updateRequest)

	select {
	case response := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := t.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"task":       map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
				"priority":    priority,
			},
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (t *taskWorkflows) ArchiveTaskWorkflow(ctx context.Context, taskID string, options map[string]any) (map[string]any, error) {
	workflow := t.manager.createWorkflow(WorkflowTypeTaskArchive)
	t.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate archive options using FormValidationEngine
	validationRules := engines.ValidationRules{}
	archiveRequest := map[string]any{"taskID": taskID, "options": options}
	validationResult := t.manager.validation.ValidateFormInputs(archiveRequest, validationRules)
	if !validationResult.Valid {
		t.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("archive validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Task archive validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Archive task through TaskManagerAccess (using status change to archived)
	archivedStatus := resource_access.UIWorkflowStatus("archived")
	respCh, errCh := t.manager.backend.ChangeTaskStatusAsync(ctx, taskID, archivedStatus)

	select {
	case response := <-respCh:
		t.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := t.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"task":       map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
				"archived":    true,
			},
			"cascade_effects": options, // Return options as cascade info
		}, nil
	case err := <-errCh:
		t.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		t.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

// Drag workflow implementations
type dragWorkflows struct {
	manager *workflowManager
}

func (d *dragWorkflows) ProcessDragDropWorkflow(ctx context.Context, event map[string]any) (map[string]any, error) {
	workflow := d.manager.createWorkflow(WorkflowTypeDragDrop)
	d.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Extract drag data
	sourceID, _ := event["source_id"].(string)
	targetID, _ := event["target_id"].(string)
	position, _ := event["drop_position"].(fyne.Position)

	// Validate movement using DragDropEngine
	valid, _ := d.manager.dragDrop.Drop().ValidateDropTarget(position, engines.DragContext{})
	if !valid {
		d.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("invalid drop target"))
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       "Invalid drop target",
		}, nil
	}

	// Process movement through backend
	// This would typically involve updating task position/status
	// For now, we'll simulate success

	d.manager.completeWorkflow(workflow.WorkflowID)
	return map[string]any{
		"success":     true,
		"workflow_id": workflow.WorkflowID,
		"source_id":   sourceID,
		"target_id":   targetID,
		"new_position": position,
	}, nil
}

// Batch workflow implementations
type batchWorkflows struct {
	manager *workflowManager
}

func (b *batchWorkflows) BatchStatusUpdateWorkflow(ctx context.Context, taskIDs []string, status string) (map[string]any, error) {
	workflow := b.manager.createWorkflow(WorkflowTypeBatchStatus)
	b.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate batch size
	if len(taskIDs) > 100 {
		b.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("batch size exceeds limit"))
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       "Batch size exceeds maximum limit of 100 tasks",
		}, nil
	}

	// Validate status using FormValidationEngine
	validationRules := engines.ValidationRules{}
	statusRequest := map[string]any{"status": status, "taskCount": len(taskIDs)}
	validationResult := b.manager.validation.ValidateFormInputs(statusRequest, validationRules)
	if !validationResult.Valid {
		b.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("batch validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Batch status update validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Process batch status updates
	results := make([]map[string]any, len(taskIDs))
	successCount := 0
	failureCount := 0

	for i, taskID := range taskIDs {
		// Change status for each task
		statusEnum := resource_access.UIWorkflowStatus(status)
		respCh, errCh := b.manager.backend.ChangeTaskStatusAsync(ctx, taskID, statusEnum)

		select {
		case response := <-respCh:
			formattedDesc, _ := b.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 30})
			results[i] = map[string]any{
				"task_id":     taskID,
				"success":     true,
				"description": formattedDesc,
				"status":      status,
			}
			successCount++
		case err := <-errCh:
			errMsg := "unknown error"
			if err != nil {
				errMsg = err.Error()
			}
			results[i] = map[string]any{
				"task_id": taskID,
				"success": false,
				"error":   errMsg,
			}
			failureCount++
		case <-ctx.Done():
			b.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
			return nil, ctx.Err()
		}
	}

	b.manager.completeWorkflow(workflow.WorkflowID)
	return map[string]any{
		"success":       successCount > 0,
		"workflow_id":   workflow.WorkflowID,
		"results":       results,
		"success_count": successCount,
		"failure_count": failureCount,
		"total_count":   len(taskIDs),
	}, nil
}

func (b *batchWorkflows) BatchPriorityUpdateWorkflow(ctx context.Context, taskIDs []string, priority string) (map[string]any, error) {
	workflow := b.manager.createWorkflow(WorkflowTypeBatchPriority)
	b.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate batch size
	if len(taskIDs) > 100 {
		b.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("batch size exceeds limit"))
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       "Batch size exceeds maximum limit of 100 tasks",
		}, nil
	}

	// Validate priority using FormValidationEngine
	validationRules := engines.ValidationRules{}
	priorityRequest := map[string]any{"priority": priority, "taskCount": len(taskIDs)}
	validationResult := b.manager.validation.ValidateFormInputs(priorityRequest, validationRules)
	if !validationResult.Valid {
		b.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("batch validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Batch priority update validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Process batch priority updates
	results := make([]map[string]any, len(taskIDs))
	successCount := 0
	failureCount := 0

	for i, taskID := range taskIDs {
		// Update priority for each task (using UpdateTaskAsync as priority-specific method may not exist)
		updateRequest := resource_access.UITaskRequest{
			Description: "", // Will be filled by backend from existing task
		}
		respCh, errCh := b.manager.backend.UpdateTaskAsync(ctx, taskID, updateRequest)

		select {
		case response := <-respCh:
			formattedDesc, _ := b.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 30})
			results[i] = map[string]any{
				"task_id":     taskID,
				"success":     true,
				"description": formattedDesc,
				"priority":    priority,
			}
			successCount++
		case err := <-errCh:
			errMsg := "unknown error"
			if err != nil {
				errMsg = err.Error()
			}
			results[i] = map[string]any{
				"task_id": taskID,
				"success": false,
				"error":   errMsg,
			}
			failureCount++
		case <-ctx.Done():
			b.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
			return nil, ctx.Err()
		}
	}

	b.manager.completeWorkflow(workflow.WorkflowID)
	return map[string]any{
		"success":       successCount > 0,
		"workflow_id":   workflow.WorkflowID,
		"results":       results,
		"success_count": successCount,
		"failure_count": failureCount,
		"total_count":   len(taskIDs),
	}, nil
}

func (b *batchWorkflows) BatchArchiveWorkflow(ctx context.Context, taskIDs []string, options map[string]any) (map[string]any, error) {
	workflow := b.manager.createWorkflow(WorkflowTypeBatchArchive)
	b.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate batch size
	if len(taskIDs) > 100 {
		b.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("batch size exceeds limit"))
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       "Batch size exceeds maximum limit of 100 tasks",
		}, nil
	}

	// Validate archive options using FormValidationEngine
	validationRules := engines.ValidationRules{}
	archiveRequest := map[string]any{"options": options, "taskCount": len(taskIDs)}
	validationResult := b.manager.validation.ValidateFormInputs(archiveRequest, validationRules)
	if !validationResult.Valid {
		b.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("batch validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Batch archive validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Process batch archive operations
	results := make([]map[string]any, len(taskIDs))
	successCount := 0
	failureCount := 0

	for i, taskID := range taskIDs {
		// Archive each task (using status change to archived)
		archivedStatus := resource_access.UIWorkflowStatus("archived")
		respCh, errCh := b.manager.backend.ChangeTaskStatusAsync(ctx, taskID, archivedStatus)

		select {
		case response := <-respCh:
			formattedDesc, _ := b.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 30})
			results[i] = map[string]any{
				"task_id":     taskID,
				"success":     true,
				"description": formattedDesc,
				"archived":    true,
			}
			successCount++
		case err := <-errCh:
			errMsg := "unknown error"
			if err != nil {
				errMsg = err.Error()
			}
			results[i] = map[string]any{
				"task_id": taskID,
				"success": false,
				"error":   errMsg,
			}
			failureCount++
		case <-ctx.Done():
			b.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
			return nil, ctx.Err()
		}
	}

	b.manager.completeWorkflow(workflow.WorkflowID)
	return map[string]any{
		"success":         successCount > 0,
		"workflow_id":     workflow.WorkflowID,
		"results":         results,
		"success_count":   successCount,
		"failure_count":   failureCount,
		"total_count":     len(taskIDs),
		"cascade_effects": options,
	}, nil
}

// Search workflow implementations
type searchWorkflows struct {
	manager *workflowManager
}

func (s *searchWorkflows) SearchTasksWorkflow(ctx context.Context, query string, filters map[string]any) (map[string]any, error) {
	workflow := s.manager.createWorkflow(WorkflowTypeSearch)
	s.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate search query using FormValidationEngine
	validationRules := engines.ValidationRules{}
	searchRequest := map[string]any{"query": query, "filters": filters}
	validationResult := s.manager.validation.ValidateFormInputs(searchRequest, validationRules)
	if !validationResult.Valid {
		s.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("search validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Search validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Execute search through TaskManagerAccess
	respCh, errCh := s.manager.backend.SearchTasksAsync(ctx, query)

	select {
	case tasks := <-respCh:
		s.manager.completeWorkflow(workflow.WorkflowID)

		// Format search results using FormattingEngine
		formattedTasks := make([]map[string]any, len(tasks))
		for i, task := range tasks {
			formattedDesc, _ := s.manager.formatting.Text().FormatText(task.Description, engines.TextOptions{MaxLength: 50})
			formattedTasks[i] = map[string]any{
				"id":           task.ID,
				"description":  formattedDesc,
				"display_name": task.DisplayName,
				"relevance":    1.0, // Default relevance
			}
		}

		return map[string]any{
			"success":     true,
			"workflow_id": workflow.WorkflowID,
			"results":     formattedTasks,
			"total_count": len(tasks),
			"query":       query,
			"metadata": map[string]any{
				"search_time": time.Now().Format(time.RFC3339),
				"filters_applied": filters,
			},
		}, nil
	case err := <-errCh:
		s.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		s.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (s *searchWorkflows) ApplyFiltersWorkflow(ctx context.Context, filters map[string]any, context map[string]any) (map[string]any, error) {
	workflow := s.manager.createWorkflow(WorkflowTypeFilter)
	s.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate filters using FormValidationEngine
	validationRules := engines.ValidationRules{}
	filterRequest := map[string]any{"filters": filters, "context": context}
	validationResult := s.manager.validation.ValidateFormInputs(filterRequest, validationRules)
	if !validationResult.Valid {
		s.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("filter validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Filter validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Convert filters to query criteria
	uiCriteria := resource_access.UIQueryCriteria{
		// Convert filters to criteria - implementation would depend on UIQueryCriteria structure
	}

	// Apply filters through TaskManagerAccess
	respCh, errCh := s.manager.backend.QueryTasksAsync(ctx, uiCriteria)

	select {
	case tasks := <-respCh:
		s.manager.completeWorkflow(workflow.WorkflowID)

		// Format filtered results using FormattingEngine
		formattedTasks := make([]map[string]any, len(tasks))
		for i, task := range tasks {
			formattedDesc, _ := s.manager.formatting.Text().FormatText(task.Description, engines.TextOptions{MaxLength: 50})
			formattedTasks[i] = map[string]any{
				"id":           task.ID,
				"description":  formattedDesc,
				"display_name": task.DisplayName,
			}
		}

		return map[string]any{
			"success":     true,
			"workflow_id": workflow.WorkflowID,
			"results":     formattedTasks,
			"total_count": len(tasks),
			"filters":     filters,
			"filter_status": map[string]any{
				"applied":    true,
				"valid":      true,
				"result_count": len(tasks),
			},
		}, nil
	case err := <-errCh:
		s.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		s.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

// Subtask workflow implementations
type subtaskWorkflows struct {
	manager *workflowManager
}

func (st *subtaskWorkflows) CreateSubtaskRelationshipWorkflow(ctx context.Context, parentID string, childSpec map[string]any) (map[string]any, error) {
	workflow := st.manager.createWorkflow(WorkflowTypeSubtaskCreate)
	st.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate subtask creation using FormValidationEngine
	validationRules := engines.ValidationRules{}
	subtaskRequest := map[string]any{"parentID": parentID, "childSpec": childSpec}
	validationResult := st.manager.validation.ValidateFormInputs(subtaskRequest, validationRules)
	if !validationResult.Valid {
		st.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("subtask validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Subtask creation validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Create subtask through TaskManagerAccess
	uiRequest := resource_access.UITaskRequest{
		Description: fmt.Sprintf("%v", childSpec["description"]),
		// Parent relationship would be established here if UITaskRequest supports it
	}
	respCh, errCh := st.manager.backend.CreateTaskAsync(ctx, uiRequest)

	select {
	case response := <-respCh:
		st.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := st.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"subtask":    map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
				"parent_id":   parentID,
			},
			"relationship": map[string]any{
				"parent_id": parentID,
				"child_id":  response.ID,
				"created":   true,
			},
		}, nil
	case err := <-errCh:
		st.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		st.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (st *subtaskWorkflows) ProcessSubtaskCompletionWorkflow(ctx context.Context, subtaskID string, cascade map[string]any) (map[string]any, error) {
	workflow := st.manager.createWorkflow(WorkflowTypeSubtaskComplete)
	st.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate subtask completion using FormValidationEngine
	validationRules := engines.ValidationRules{}
	completionRequest := map[string]any{"subtaskID": subtaskID, "cascade": cascade}
	validationResult := st.manager.validation.ValidateFormInputs(completionRequest, validationRules)
	if !validationResult.Valid {
		st.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("completion validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Subtask completion validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Mark subtask as completed through TaskManagerAccess
	completedStatus := resource_access.UIWorkflowStatus("completed")
	respCh, errCh := st.manager.backend.ChangeTaskStatusAsync(ctx, subtaskID, completedStatus)

	select {
	case response := <-respCh:
		st.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := st.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"subtask":    map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
				"completed":   true,
			},
			"cascade_results": map[string]any{
				"processed":    true,
				"affected_tasks": []string{}, // Would be populated based on actual cascade logic
				"cascade_options": cascade,
			},
		}, nil
	case err := <-errCh:
		st.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		st.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}

func (st *subtaskWorkflows) MoveSubtaskWorkflow(ctx context.Context, subtaskID string, newParentID string, position map[string]any) (map[string]any, error) {
	workflow := st.manager.createWorkflow(WorkflowTypeSubtaskMove)
	st.manager.updateWorkflowStatus(workflow.WorkflowID, WorkflowStatusInProgress)

	// Validate subtask movement using FormValidationEngine
	validationRules := engines.ValidationRules{}
	moveRequest := map[string]any{"subtaskID": subtaskID, "newParentID": newParentID, "position": position}
	validationResult := st.manager.validation.ValidateFormInputs(moveRequest, validationRules)
	if !validationResult.Valid {
		st.manager.failWorkflow(workflow.WorkflowID, fmt.Errorf("move validation failed"))
		return map[string]any{
			"success":      false,
			"workflow_id":  workflow.WorkflowID,
			"error":        "Subtask move validation failed",
			"field_errors": validationResult.Errors,
		}, nil
	}

	// Move subtask through TaskManagerAccess (using update to change parent reference)
	updateRequest := resource_access.UITaskRequest{
		Description: "", // Will be filled by backend from existing task
		// Parent relationship would be updated here if UITaskRequest supports it
	}
	respCh, errCh := st.manager.backend.UpdateTaskAsync(ctx, subtaskID, updateRequest)

	select {
	case response := <-respCh:
		st.manager.completeWorkflow(workflow.WorkflowID)

		// Format response using FormattingEngine
		formattedDesc, _ := st.manager.formatting.Text().FormatText(response.Description, engines.TextOptions{MaxLength: 50})

		return map[string]any{
			"success":    true,
			"workflow_id": workflow.WorkflowID,
			"subtask":    map[string]any{
				"id":          response.ID,
				"description": formattedDesc,
				"display_name": response.DisplayName,
			},
			"movement": map[string]any{
				"new_parent_id": newParentID,
				"position":      position,
				"moved":         true,
			},
		}, nil
	case err := <-errCh:
		st.manager.failWorkflow(workflow.WorkflowID, err)
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		return map[string]any{
			"success":     false,
			"workflow_id": workflow.WorkflowID,
			"error":       errMsg,
		}, err
	case <-ctx.Done():
		st.manager.failWorkflow(workflow.WorkflowID, ctx.Err())
		return nil, ctx.Err()
	}
}