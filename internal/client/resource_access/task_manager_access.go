// Package resource_access provides ResourceAccess layer components for the EisenKan Client following iDesign methodology.
// This package contains components that interface with service layers and provide UI-optimized data access.
// Following iDesign namespace: eisenkan.Client.ResourceAccess
package resource_access

import (
	"context"
	"fmt"
	"time"

	"github.com/rknuus/eisenkan/internal/managers"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// ITaskManagerAccess defines the interface for UI-optimized task management operations
type ITaskManagerAccess interface {
	// Task Data Operations
	CreateTaskAsync(ctx context.Context, request UITaskRequest) (<-chan UITaskResponse, <-chan error)
	UpdateTaskAsync(ctx context.Context, taskID string, request UITaskRequest) (<-chan UITaskResponse, <-chan error)
	GetTaskAsync(ctx context.Context, taskID string) (<-chan UITaskResponse, <-chan error)
	DeleteTaskAsync(ctx context.Context, taskID string) (<-chan bool, <-chan error)
	ListTasksAsync(ctx context.Context, criteria UIQueryCriteria) (<-chan []UITaskResponse, <-chan error)

	// Workflow Operations
	ChangeTaskStatusAsync(ctx context.Context, taskID string, status UIWorkflowStatus) (<-chan UITaskResponse, <-chan error)
	ValidateTaskAsync(ctx context.Context, request UITaskRequest) (<-chan UIValidationResult, <-chan error)
	ProcessPriorityPromotionsAsync(ctx context.Context) (<-chan []UITaskResponse, <-chan error)

	// Query Operations
	QueryTasksAsync(ctx context.Context, criteria UIQueryCriteria) (<-chan []UITaskResponse, <-chan error)
	GetBoardSummaryAsync(ctx context.Context) (<-chan UIBoardSummary, <-chan error)
	SearchTasksAsync(ctx context.Context, query string) (<-chan []UITaskResponse, <-chan error)
}

// ICacheUtility defines the interface for UI caching operations
type ICacheUtility interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Invalidate(key string)
	InvalidatePattern(pattern string)
}

// taskManagerAccess implements ITaskManagerAccess with simple channel-based async operations
type taskManagerAccess struct {
	taskManager managers.TaskManager
	cache       ICacheUtility
	logger      utilities.ILoggingUtility
}

// NewTaskManagerAccess creates a new TaskManagerAccess instance
func NewTaskManagerAccess(taskManager managers.TaskManager, cache ICacheUtility, logger utilities.ILoggingUtility) ITaskManagerAccess {
	return &taskManagerAccess{
		taskManager: taskManager,
		cache:       cache,
		logger:      logger,
	}
}

// CreateTaskAsync creates a new task asynchronously
func (t *taskManagerAccess) CreateTaskAsync(ctx context.Context, request UITaskRequest) (<-chan UITaskResponse, <-chan error) {
	resultChan := make(chan UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Validate input
		if validationResult := t.validateUITaskRequest(request); !validationResult.Valid {
			errorChan <- t.createUIError("validation", "Task validation failed", validationResult.GeneralError, validationResult.Suggestions, true)
			return
		}

		// Convert UI request to TaskManager format
		taskRequest, err := t.convertUIRequestToTaskRequest(request)
		if err != nil {
			errorChan <- t.createUIError("validation", "Request conversion failed", err.Error(), []string{"Check task data format"}, true)
			return
		}

		// Call TaskManager service
		response, err := t.taskManager.CreateTask(taskRequest)
		if err != nil {
			errorChan <- t.translateServiceError("CreateTask", err)
			return
		}

		// Convert response to UI format
		uiResponse := t.convertTaskResponseToUI(response)

		// Invalidate relevant cache entries
		t.cache.InvalidatePattern("tasks_*")
		t.cache.InvalidatePattern("board_summary")

		// Log operation
		t.logger.Log(utilities.Info, "TaskManagerAccess", "Task created successfully", map[string]interface{}{
			"task_id": response.ID,
		})

		resultChan <- uiResponse
	}()

	return resultChan, errorChan
}

// UpdateTaskAsync updates an existing task asynchronously
func (t *taskManagerAccess) UpdateTaskAsync(ctx context.Context, taskID string, request UITaskRequest) (<-chan UITaskResponse, <-chan error) {
	resultChan := make(chan UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Validate input
		if taskID == "" {
			errorChan <- t.createUIError("validation", "Task ID is required", "Empty task ID provided", []string{"Provide a valid task ID"}, false)
			return
		}

		if validationResult := t.validateUITaskRequest(request); !validationResult.Valid {
			errorChan <- t.createUIError("validation", "Task validation failed", validationResult.GeneralError, validationResult.Suggestions, true)
			return
		}

		// Convert UI request to TaskManager format
		taskRequest, err := t.convertUIRequestToTaskRequest(request)
		if err != nil {
			errorChan <- t.createUIError("validation", "Request conversion failed", err.Error(), []string{"Check task data format"}, true)
			return
		}

		// Call TaskManager service
		response, err := t.taskManager.UpdateTask(taskID, taskRequest)
		if err != nil {
			errorChan <- t.translateServiceError("UpdateTask", err)
			return
		}

		// Convert response to UI format
		uiResponse := t.convertTaskResponseToUI(response)

		// Invalidate relevant cache entries
		t.cache.Invalidate(fmt.Sprintf("task_%s", taskID))
		t.cache.InvalidatePattern("tasks_*")
		t.cache.InvalidatePattern("board_summary")

		// Log operation
		t.logger.Log(utilities.Info, "TaskManagerAccess", "Task updated successfully", map[string]interface{}{
			"task_id": taskID,
		})

		resultChan <- uiResponse
	}()

	return resultChan, errorChan
}

// GetTaskAsync retrieves a single task asynchronously
func (t *taskManagerAccess) GetTaskAsync(ctx context.Context, taskID string) (<-chan UITaskResponse, <-chan error) {
	resultChan := make(chan UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Validate input
		if taskID == "" {
			errorChan <- t.createUIError("validation", "Task ID is required", "Empty task ID provided", []string{"Provide a valid task ID"}, false)
			return
		}

		// Check cache first
		cacheKey := fmt.Sprintf("task_%s", taskID)
		if cached, found := t.cache.Get(cacheKey); found {
			if uiResponse, ok := cached.(UITaskResponse); ok {
				resultChan <- uiResponse
				return
			}
		}

		// Call TaskManager service
		response, err := t.taskManager.GetTask(taskID)
		if err != nil {
			errorChan <- t.translateServiceError("GetTask", err)
			return
		}

		// Convert response to UI format
		uiResponse := t.convertTaskResponseToUI(response)

		// Cache the result
		t.cache.Set(cacheKey, uiResponse, 5*time.Minute)

		resultChan <- uiResponse
	}()

	return resultChan, errorChan
}

// DeleteTaskAsync deletes a task asynchronously
func (t *taskManagerAccess) DeleteTaskAsync(ctx context.Context, taskID string) (<-chan bool, <-chan error) {
	resultChan := make(chan bool, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Validate input
		if taskID == "" {
			errorChan <- t.createUIError("validation", "Task ID is required", "Empty task ID provided", []string{"Provide a valid task ID"}, false)
			return
		}

		// Call TaskManager service
		err := t.taskManager.DeleteTask(taskID)
		if err != nil {
			errorChan <- t.translateServiceError("DeleteTask", err)
			return
		}

		// Invalidate relevant cache entries
		t.cache.Invalidate(fmt.Sprintf("task_%s", taskID))
		t.cache.InvalidatePattern("tasks_*")
		t.cache.InvalidatePattern("board_summary")

		// Log operation
		t.logger.Log(utilities.Info, "TaskManagerAccess", "Task deleted successfully", map[string]interface{}{
			"task_id": taskID,
		})

		resultChan <- true
	}()

	return resultChan, errorChan
}

// ListTasksAsync lists tasks with filtering asynchronously
func (t *taskManagerAccess) ListTasksAsync(ctx context.Context, criteria UIQueryCriteria) (<-chan []UITaskResponse, <-chan error) {
	resultChan := make(chan []UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Convert UI criteria to TaskManager format
		taskCriteria := t.convertUIQueryCriteriaToTaskCriteria(criteria)

		// Check cache
		cacheKey := fmt.Sprintf("tasks_%s", t.generateCacheKeyFromCriteria(criteria))
		if cached, found := t.cache.Get(cacheKey); found {
			if uiResponses, ok := cached.([]UITaskResponse); ok {
				resultChan <- uiResponses
				return
			}
		}

		// Call TaskManager service
		responses, err := t.taskManager.ListTasks(taskCriteria)
		if err != nil {
			errorChan <- t.translateServiceError("ListTasks", err)
			return
		}

		// Convert responses to UI format
		uiResponses := make([]UITaskResponse, len(responses))
		for i, response := range responses {
			uiResponses[i] = t.convertTaskResponseToUI(response)
		}

		// Cache the result
		t.cache.Set(cacheKey, uiResponses, 2*time.Minute)

		resultChan <- uiResponses
	}()

	return resultChan, errorChan
}

// ChangeTaskStatusAsync changes task workflow status asynchronously
func (t *taskManagerAccess) ChangeTaskStatusAsync(ctx context.Context, taskID string, status UIWorkflowStatus) (<-chan UITaskResponse, <-chan error) {
	resultChan := make(chan UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Validate input
		if taskID == "" {
			errorChan <- t.createUIError("validation", "Task ID is required", "Empty task ID provided", []string{"Provide a valid task ID"}, false)
			return
		}

		// Convert UI status to TaskManager format
		taskStatus := t.convertUIWorkflowStatusToTaskStatus(status)

		// Call TaskManager service
		response, err := t.taskManager.ChangeTaskStatus(taskID, taskStatus)
		if err != nil {
			errorChan <- t.translateServiceError("ChangeTaskStatus", err)
			return
		}

		// Convert response to UI format
		uiResponse := t.convertTaskResponseToUI(response)

		// Invalidate relevant cache entries
		t.cache.Invalidate(fmt.Sprintf("task_%s", taskID))
		t.cache.InvalidatePattern("tasks_*")
		t.cache.InvalidatePattern("board_summary")

		// Log operation
		t.logger.Log(utilities.Info, "TaskManagerAccess", "Task status changed successfully", map[string]interface{}{
			"task_id": taskID,
			"status":  string(status),
		})

		resultChan <- uiResponse
	}()

	return resultChan, errorChan
}

// ValidateTaskAsync validates task data asynchronously
func (t *taskManagerAccess) ValidateTaskAsync(ctx context.Context, request UITaskRequest) (<-chan UIValidationResult, <-chan error) {
	resultChan := make(chan UIValidationResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Perform UI-level validation first
		uiValidation := t.validateUITaskRequest(request)
		if !uiValidation.Valid {
			resultChan <- uiValidation
			return
		}

		// Convert to TaskManager format for service validation
		taskRequest, err := t.convertUIRequestToTaskRequest(request)
		if err != nil {
			uiValidation.Valid = false
			uiValidation.GeneralError = "Request conversion failed: " + err.Error()
			uiValidation.Suggestions = []string{"Check task data format"}
			resultChan <- uiValidation
			return
		}

		// Call TaskManager validation
		validation, err := t.taskManager.ValidateTask(taskRequest)
		if err != nil {
			errorChan <- t.translateServiceError("ValidateTask", err)
			return
		}

		// Convert validation result to UI format
		uiValidationResult := t.convertValidationResultToUI(validation)
		resultChan <- uiValidationResult
	}()

	return resultChan, errorChan
}

// ProcessPriorityPromotionsAsync processes priority promotions asynchronously
func (t *taskManagerAccess) ProcessPriorityPromotionsAsync(ctx context.Context) (<-chan []UITaskResponse, <-chan error) {
	resultChan := make(chan []UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Call TaskManager service
		responses, err := t.taskManager.ProcessPriorityPromotions()
		if err != nil {
			errorChan <- t.translateServiceError("ProcessPriorityPromotions", err)
			return
		}

		// Convert responses to UI format
		uiResponses := make([]UITaskResponse, len(responses))
		for i, response := range responses {
			uiResponses[i] = t.convertTaskResponseToUI(response)
		}

		// Invalidate cache if promotions occurred
		if len(uiResponses) > 0 {
			t.cache.InvalidatePattern("tasks_*")
			t.cache.InvalidatePattern("board_summary")

			// Log operation
			t.logger.Log(utilities.Info, "TaskManagerAccess", "Priority promotions processed", map[string]interface{}{
				"promoted_count": len(uiResponses),
			})
		}

		resultChan <- uiResponses
	}()

	return resultChan, errorChan
}

// QueryTasksAsync performs advanced task queries asynchronously
func (t *taskManagerAccess) QueryTasksAsync(ctx context.Context, criteria UIQueryCriteria) (<-chan []UITaskResponse, <-chan error) {
	// QueryTasksAsync is essentially the same as ListTasksAsync for this implementation
	return t.ListTasksAsync(ctx, criteria)
}

// GetBoardSummaryAsync retrieves board summary statistics asynchronously
func (t *taskManagerAccess) GetBoardSummaryAsync(ctx context.Context) (<-chan UIBoardSummary, <-chan error) {
	resultChan := make(chan UIBoardSummary, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Check cache
		cacheKey := "board_summary"
		if cached, found := t.cache.Get(cacheKey); found {
			if summary, ok := cached.(UIBoardSummary); ok {
				resultChan <- summary
				return
			}
		}

		// Get all tasks to calculate summary
		allTasks, err := t.taskManager.ListTasks(managers.QueryCriteria{})
		if err != nil {
			errorChan <- t.translateServiceError("GetBoardSummary", err)
			return
		}

		// Calculate summary statistics
		summary := t.calculateBoardSummary(allTasks)

		// Cache the result
		t.cache.Set(cacheKey, summary, 1*time.Minute)

		resultChan <- summary
	}()

	return resultChan, errorChan
}

// SearchTasksAsync performs text-based task search asynchronously
func (t *taskManagerAccess) SearchTasksAsync(ctx context.Context, query string) (<-chan []UITaskResponse, <-chan error) {
	resultChan := make(chan []UITaskResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		// Validate query
		if query == "" {
			errorChan <- t.createUIError("validation", "Search query is required", "Empty search query provided", []string{"Provide a search term"}, false)
			return
		}

		// Check cache
		cacheKey := fmt.Sprintf("search_%s", query)
		if cached, found := t.cache.Get(cacheKey); found {
			if uiResponses, ok := cached.([]UITaskResponse); ok {
				resultChan <- uiResponses
				return
			}
		}

		// Get all tasks and filter by search query
		// Note: In a real implementation, this might use a dedicated search service
		allTasks, err := t.taskManager.ListTasks(managers.QueryCriteria{})
		if err != nil {
			errorChan <- t.translateServiceError("SearchTasks", err)
			return
		}

		// Filter tasks by search query
		var matchingTasks []managers.TaskResponse
		for _, task := range allTasks {
			if t.taskMatchesQuery(task, query) {
				matchingTasks = append(matchingTasks, task)
			}
		}

		// Convert to UI format
		uiResponses := make([]UITaskResponse, len(matchingTasks))
		for i, task := range matchingTasks {
			uiResponses[i] = t.convertTaskResponseToUI(task)
		}

		// Cache the result
		t.cache.Set(cacheKey, uiResponses, 30*time.Second)

		resultChan <- uiResponses
	}()

	return resultChan, errorChan
}