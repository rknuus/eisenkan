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
}

// ITask handles task-related workflows with validation
type ITask interface {
	CreateTaskWorkflow(ctx context.Context, request map[string]any) (map[string]any, error)
	UpdateTaskWorkflow(ctx context.Context, taskID string, request map[string]any) (map[string]any, error)
	DeleteTaskWorkflow(ctx context.Context, taskID string) (map[string]any, error)
	QueryTasksWorkflow(ctx context.Context, criteria map[string]any) (map[string]any, error)
}

// IDrag handles drag-drop workflows with movement validation
type IDrag interface {
	ProcessDragDropWorkflow(ctx context.Context, event map[string]any) (map[string]any, error)
}

// Data Types for workflow state management
type WorkflowType string
type WorkflowStatus string

const (
	WorkflowTypeTaskCreate WorkflowType = "task_create"
	WorkflowTypeTaskUpdate WorkflowType = "task_update"
	WorkflowTypeTaskDelete WorkflowType = "task_delete"
	WorkflowTypeTaskQuery  WorkflowType = "task_query"
	WorkflowTypeDragDrop   WorkflowType = "drag_drop"

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