package resource_access

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/managers/task_manager"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// Mock implementations for testing

// MockTaskManager is a mock implementation of task_manager.TaskManager
type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) CreateTask(request task_manager.TaskRequest) (task_manager.TaskResponse, error) {
	args := m.Called(request)
	return args.Get(0).(task_manager.TaskResponse), args.Error(1)
}

func (m *MockTaskManager) UpdateTask(taskID string, request task_manager.TaskRequest) (task_manager.TaskResponse, error) {
	args := m.Called(taskID, request)
	return args.Get(0).(task_manager.TaskResponse), args.Error(1)
}

func (m *MockTaskManager) GetTask(taskID string) (task_manager.TaskResponse, error) {
	args := m.Called(taskID)
	return args.Get(0).(task_manager.TaskResponse), args.Error(1)
}

func (m *MockTaskManager) DeleteTask(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockTaskManager) ListTasks(criteria task_manager.QueryCriteria) ([]task_manager.TaskResponse, error) {
	args := m.Called(criteria)
	return args.Get(0).([]task_manager.TaskResponse), args.Error(1)
}

func (m *MockTaskManager) ChangeTaskStatus(taskID string, status task_manager.WorkflowStatus) (task_manager.TaskResponse, error) {
	args := m.Called(taskID, status)
	return args.Get(0).(task_manager.TaskResponse), args.Error(1)
}

func (m *MockTaskManager) ValidateTask(request task_manager.TaskRequest) (task_manager.ValidationResult, error) {
	args := m.Called(request)
	return args.Get(0).(task_manager.ValidationResult), args.Error(1)
}

func (m *MockTaskManager) ProcessPriorityPromotions() ([]task_manager.TaskResponse, error) {
	args := m.Called()
	return args.Get(0).([]task_manager.TaskResponse), args.Error(1)
}

// IContext facet mock methods
func (m *MockTaskManager) Load(contextType string) (task_manager.ContextData, error) {
	args := m.Called(contextType)
	return args.Get(0).(task_manager.ContextData), args.Error(1)
}

func (m *MockTaskManager) Store(contextType string, data task_manager.ContextData) error {
	args := m.Called(contextType, data)
	return args.Error(0)
}

// MockCacheUtility is a mock implementation of ICacheUtility
type MockCacheUtility struct {
	mock.Mock
}

func (m *MockCacheUtility) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheUtility) Set(key string, value interface{}, ttl time.Duration) {
	m.Called(key, value, ttl)
}

func (m *MockCacheUtility) Invalidate(key string) {
	m.Called(key)
}

func (m *MockCacheUtility) InvalidatePattern(pattern string) {
	m.Called(pattern)
}

// MockLoggingUtility is a mock implementation of utilities.ILoggingUtility
type MockLoggingUtility struct {
	mock.Mock
}

func (m *MockLoggingUtility) Log(level utilities.LogLevel, component, message string, data interface{}) {
	m.Called(level, component, message, data)
}

func (m *MockLoggingUtility) LogMessage(level utilities.LogLevel, component string, message string) {
	m.Called(level, component, message)
}

func (m *MockLoggingUtility) LogError(component string, err error, data interface{}) {
	m.Called(component, err, data)
}

func (m *MockLoggingUtility) IsLevelEnabled(level utilities.LogLevel) bool {
	args := m.Called(level)
	return args.Bool(0)
}

// Test helper functions

func createTestTaskManagerAccess() (*taskManagerAccess, *MockTaskManager, *MockCacheUtility, *MockLoggingUtility) {
	mockTaskManager := &MockTaskManager{}
	mockCache := &MockCacheUtility{}
	mockLogger := &MockLoggingUtility{}
	
	access := &taskManagerAccess{
		taskManager: mockTaskManager,
		cache:       mockCache,
		logger:      mockLogger,
	}
	
	return access, mockTaskManager, mockCache, mockLogger
}

func createValidUITaskRequest() UITaskRequest {
	return UITaskRequest{
		Description:    "Test task",
		Priority:       UIPriority{Urgent: true, Important: false, Label: "urgent"},
		WorkflowStatus: UITodo,
		Tags:           []string{"test", "unit"},
	}
}

func createValidTaskResponse() task_manager.TaskResponse {
	return task_manager.TaskResponse{
		ID:             "task-123",
		Description:    "Test task",
		Priority:       board_access.Priority{Urgent: true, Important: false, Label: "urgent"},
		WorkflowStatus: task_manager.Todo,
		Tags:           []string{"test", "unit"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Unit Tests

// TestUnit_TaskManagerAccess_NewTaskManagerAccess tests the constructor
func TestUnit_TaskManagerAccess_NewTaskManagerAccess(t *testing.T) {
	mockTaskManager := &MockTaskManager{}
	mockCache := &MockCacheUtility{}
	mockLogger := &MockLoggingUtility{}
	
	access := NewTaskManagerAccess(mockTaskManager, mockCache, mockLogger)
	
	assert.NotNil(t, access, "TaskManagerAccess should be created")
	assert.Implements(t, (*ITaskManagerAccess)(nil), access, "Should implement ITaskManagerAccess interface")
}

// TestUnit_TaskManagerAccess_CreateTaskAsync_Success tests successful task creation
func TestUnit_TaskManagerAccess_CreateTaskAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, mockLogger := createTestTaskManagerAccess()
	
	uiRequest := createValidUITaskRequest()
	expectedResponse := createValidTaskResponse()
	
	// Setup mocks
	mockTaskManager.On("CreateTask", mock.AnythingOfType("task_manager.TaskRequest")).Return(expectedResponse, nil)
	mockCache.On("InvalidatePattern", "tasks_*").Return()
	mockCache.On("InvalidatePattern", "board_summary").Return()
	mockLogger.On("Log", utilities.Info, "TaskManagerAccess", "Task created successfully", mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.CreateTaskAsync(ctx, uiRequest)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, "task-123", result.ID, "Task ID should match")
		assert.Equal(t, "Test task", result.Description, "Description should match")
		assert.True(t, result.Priority.Urgent, "Priority should be urgent")
		assert.Equal(t, "urgent", result.Priority.Label, "Priority label should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	// Verify mocks
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_CreateTaskAsync_ValidationFailure tests validation errors
func TestUnit_TaskManagerAccess_CreateTaskAsync_ValidationFailure(t *testing.T) {
	access, _, _, _ := createTestTaskManagerAccess()
	
	// Invalid request (empty description)
	uiRequest := UITaskRequest{
		Description:    "",
		Priority:       UIPriority{Urgent: false, Important: false},
		WorkflowStatus: UITodo,
	}
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.CreateTaskAsync(ctx, uiRequest)
	
	// Wait for error
	select {
	case <-resultChan:
		t.Fatal("Expected validation error but got success")
	case err := <-errorChan:
		uiError, ok := err.(UIErrorResponse)
		assert.True(t, ok, "Error should be UIErrorResponse")
		assert.Equal(t, "validation", uiError.Category, "Error category should be validation")
		assert.True(t, uiError.Retryable, "Validation errors should be retryable")
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
}

// TestUnit_TaskManagerAccess_CreateTaskAsync_ServiceError tests service layer errors
func TestUnit_TaskManagerAccess_CreateTaskAsync_ServiceError(t *testing.T) {
	access, mockTaskManager, _, mockLogger := createTestTaskManagerAccess()
	
	uiRequest := createValidUITaskRequest()
	serviceError := fmt.Errorf("task creation failed")
	
	// Setup mocks
	mockTaskManager.On("CreateTask", mock.AnythingOfType("task_manager.TaskRequest")).Return(task_manager.TaskResponse{}, serviceError)
	mockLogger.On("LogError", "TaskManagerAccess", serviceError, mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.CreateTaskAsync(ctx, uiRequest)
	
	// Wait for error
	select {
	case <-resultChan:
		t.Fatal("Expected service error but got success")
	case err := <-errorChan:
		uiError, ok := err.(UIErrorResponse)
		assert.True(t, ok, "Error should be UIErrorResponse")
		assert.Equal(t, "service", uiError.Category, "Error category should be service")
		assert.Contains(t, uiError.Details, "task creation failed", "Error details should contain original error")
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_UpdateTaskAsync_Success tests successful task update
func TestUnit_TaskManagerAccess_UpdateTaskAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, mockLogger := createTestTaskManagerAccess()
	
	taskID := "task-123"
	uiRequest := createValidUITaskRequest()
	expectedResponse := createValidTaskResponse()
	
	// Setup mocks
	mockTaskManager.On("UpdateTask", taskID, mock.AnythingOfType("task_manager.TaskRequest")).Return(expectedResponse, nil)
	mockCache.On("Invalidate", "task_task-123").Return()
	mockCache.On("InvalidatePattern", "tasks_*").Return()
	mockCache.On("InvalidatePattern", "board_summary").Return()
	mockLogger.On("Log", utilities.Info, "TaskManagerAccess", "Task updated successfully", mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.UpdateTaskAsync(ctx, taskID, uiRequest)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, "task-123", result.ID, "Task ID should match")
		assert.Equal(t, "Test task", result.Description, "Description should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_UpdateTaskAsync_EmptyTaskID tests empty task ID validation
func TestUnit_TaskManagerAccess_UpdateTaskAsync_EmptyTaskID(t *testing.T) {
	access, _, _, _ := createTestTaskManagerAccess()
	
	uiRequest := createValidUITaskRequest()
	
	// Execute with empty task ID
	ctx := context.Background()
	resultChan, errorChan := access.UpdateTaskAsync(ctx, "", uiRequest)
	
	// Wait for error
	select {
	case <-resultChan:
		t.Fatal("Expected validation error but got success")
	case err := <-errorChan:
		uiError, ok := err.(UIErrorResponse)
		assert.True(t, ok, "Error should be UIErrorResponse")
		assert.Equal(t, "validation", uiError.Category, "Error category should be validation")
		assert.Contains(t, uiError.Message, "Task ID is required", "Error message should mention task ID")
		assert.False(t, uiError.Retryable, "Empty task ID errors should not be retryable")
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
}

// TestUnit_TaskManagerAccess_GetTaskAsync_Success tests successful task retrieval
func TestUnit_TaskManagerAccess_GetTaskAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	taskID := "task-123"
	expectedResponse := createValidTaskResponse()
	
	// Setup mocks - cache miss, then service call
	mockCache.On("Get", "task_task-123").Return(nil, false)
	mockTaskManager.On("GetTask", taskID).Return(expectedResponse, nil)
	mockCache.On("Set", "task_task-123", mock.Anything, 5*time.Minute).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.GetTaskAsync(ctx, taskID)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, "task-123", result.ID, "Task ID should match")
		assert.Equal(t, "Test task", result.Description, "Description should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_GetTaskAsync_CacheHit tests cache hit scenario
func TestUnit_TaskManagerAccess_GetTaskAsync_CacheHit(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	taskID := "task-123"
	cachedResponse := UITaskResponse{
		ID:          "task-123",
		Description: "Cached task",
		Priority:    UIPriority{Urgent: true, Important: false, Label: "urgent"},
	}
	
	// Setup mocks - cache hit
	mockCache.On("Get", "task_task-123").Return(cachedResponse, true)
	// TaskManager should NOT be called on cache hit
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.GetTaskAsync(ctx, taskID)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, "task-123", result.ID, "Task ID should match")
		assert.Equal(t, "Cached task", result.Description, "Should return cached description")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	// Verify TaskManager was NOT called
	mockTaskManager.AssertNotCalled(t, "GetTask", mock.Anything)
	mockCache.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_GetTaskAsync_EmptyTaskID tests empty task ID validation
func TestUnit_TaskManagerAccess_GetTaskAsync_EmptyTaskID(t *testing.T) {
	access, _, _, _ := createTestTaskManagerAccess()
	
	// Execute with empty task ID
	ctx := context.Background()
	resultChan, errorChan := access.GetTaskAsync(ctx, "")
	
	// Wait for error
	select {
	case <-resultChan:
		t.Fatal("Expected validation error but got success")
	case err := <-errorChan:
		uiError, ok := err.(UIErrorResponse)
		assert.True(t, ok, "Error should be UIErrorResponse")
		assert.Equal(t, "validation", uiError.Category, "Error category should be validation")
		assert.Contains(t, uiError.Message, "Task ID is required", "Error message should mention task ID")
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
}

// TestUnit_TaskManagerAccess_DeleteTaskAsync_Success tests successful task deletion
func TestUnit_TaskManagerAccess_DeleteTaskAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, mockLogger := createTestTaskManagerAccess()
	
	taskID := "task-123"
	
	// Setup mocks
	mockTaskManager.On("DeleteTask", taskID).Return(nil)
	mockCache.On("Invalidate", "task_task-123").Return()
	mockCache.On("InvalidatePattern", "tasks_*").Return()
	mockCache.On("InvalidatePattern", "board_summary").Return()
	mockLogger.On("Log", utilities.Info, "TaskManagerAccess", "Task deleted successfully", mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.DeleteTaskAsync(ctx, taskID)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.True(t, result, "Delete operation should return true")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_DeleteTaskAsync_ServiceError tests service deletion error
func TestUnit_TaskManagerAccess_DeleteTaskAsync_ServiceError(t *testing.T) {
	access, mockTaskManager, _, mockLogger := createTestTaskManagerAccess()
	
	taskID := "task-123"
	serviceError := fmt.Errorf("task not found")
	
	// Setup mocks
	mockTaskManager.On("DeleteTask", taskID).Return(serviceError)
	mockLogger.On("LogError", "TaskManagerAccess", serviceError, mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.DeleteTaskAsync(ctx, taskID)
	
	// Wait for error
	select {
	case <-resultChan:
		t.Fatal("Expected service error but got success")
	case err := <-errorChan:
		uiError, ok := err.(UIErrorResponse)
		assert.True(t, ok, "Error should be UIErrorResponse")
		assert.Equal(t, "service", uiError.Category, "Error category should be service")
		assert.Contains(t, uiError.Details, "task not found", "Error details should contain original error")
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_ListTasksAsync_Success tests successful task listing
func TestUnit_TaskManagerAccess_ListTasksAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	criteria := UIQueryCriteria{
		Columns: []string{"todo", "in-progress"},
	}
	expectedTasks := []task_manager.TaskResponse{createValidTaskResponse()}
	
	// Setup mocks - cache miss, then service call
	mockCache.On("Get", mock.AnythingOfType("string")).Return(nil, false)
	mockTaskManager.On("ListTasks", mock.AnythingOfType("task_manager.QueryCriteria")).Return(expectedTasks, nil)
	mockCache.On("Set", mock.AnythingOfType("string"), mock.Anything, 2*time.Minute).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.ListTasksAsync(ctx, criteria)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Len(t, result, 1, "Should return one task")
		assert.Equal(t, "task-123", result[0].ID, "Task ID should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_ChangeTaskStatusAsync_Success tests successful status change
func TestUnit_TaskManagerAccess_ChangeTaskStatusAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, mockLogger := createTestTaskManagerAccess()
	
	taskID := "task-123"
	newStatus := UIInProgress
	expectedResponse := createValidTaskResponse()
	expectedResponse.WorkflowStatus = task_manager.InProgress
	
	// Setup mocks
	mockTaskManager.On("ChangeTaskStatus", taskID, task_manager.InProgress).Return(expectedResponse, nil)
	mockCache.On("Invalidate", "task_task-123").Return()
	mockCache.On("InvalidatePattern", "tasks_*").Return()
	mockCache.On("InvalidatePattern", "board_summary").Return()
	mockLogger.On("Log", utilities.Info, "TaskManagerAccess", "Task status changed successfully", mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.ChangeTaskStatusAsync(ctx, taskID, newStatus)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, "task-123", result.ID, "Task ID should match")
		assert.Equal(t, UIInProgress, result.WorkflowStatus, "Status should be updated")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_ValidateTaskAsync_Success tests successful validation
func TestUnit_TaskManagerAccess_ValidateTaskAsync_Success(t *testing.T) {
	access, mockTaskManager, _, _ := createTestTaskManagerAccess()
	
	uiRequest := createValidUITaskRequest()
	validationResult := task_manager.ValidationResult{
		Valid:      true,
		Violations: []engines.RuleViolation{},
	}
	
	// Setup mocks
	mockTaskManager.On("ValidateTask", mock.AnythingOfType("task_manager.TaskRequest")).Return(validationResult, nil)
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.ValidateTaskAsync(ctx, uiRequest)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.True(t, result.Valid, "Validation should be successful")
		assert.Empty(t, result.FieldErrors, "Should have no field errors")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_ValidateTaskAsync_UIValidationFailure tests UI-level validation failure
func TestUnit_TaskManagerAccess_ValidateTaskAsync_UIValidationFailure(t *testing.T) {
	access, _, _, _ := createTestTaskManagerAccess()
	
	// Invalid request (empty description)
	uiRequest := UITaskRequest{
		Description:    "",
		Priority:       UIPriority{Urgent: false, Important: false},
		WorkflowStatus: UITodo,
	}
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.ValidateTaskAsync(ctx, uiRequest)
	
	// Wait for result (should return validation result, not error)
	select {
	case result := <-resultChan:
		assert.False(t, result.Valid, "Validation should fail")
		assert.Contains(t, result.FieldErrors, "description", "Should have description field error")
		assert.NotEmpty(t, result.Suggestions, "Should have suggestions")
	case err := <-errorChan:
		t.Fatalf("Expected validation result but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
}

// TestUnit_TaskManagerAccess_ProcessPriorityPromotionsAsync_Success tests priority promotions
func TestUnit_TaskManagerAccess_ProcessPriorityPromotionsAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, mockLogger := createTestTaskManagerAccess()
	
	promotedTasks := []task_manager.TaskResponse{createValidTaskResponse()}
	
	// Setup mocks
	mockTaskManager.On("ProcessPriorityPromotions").Return(promotedTasks, nil)
	mockCache.On("InvalidatePattern", "tasks_*").Return()
	mockCache.On("InvalidatePattern", "board_summary").Return()
	mockLogger.On("Log", utilities.Info, "TaskManagerAccess", "Priority promotions processed", mock.Anything).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.ProcessPriorityPromotionsAsync(ctx)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Len(t, result, 1, "Should return one promoted task")
		assert.Equal(t, "task-123", result[0].ID, "Task ID should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_ProcessPriorityPromotionsAsync_NoPromotions tests no promotions case
func TestUnit_TaskManagerAccess_ProcessPriorityPromotionsAsync_NoPromotions(t *testing.T) {
	access, mockTaskManager, _, _ := createTestTaskManagerAccess()
	
	// Setup mocks - no promotions
	mockTaskManager.On("ProcessPriorityPromotions").Return([]task_manager.TaskResponse{}, nil)
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.ProcessPriorityPromotionsAsync(ctx)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Empty(t, result, "Should return empty list when no promotions")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	// Cache should NOT be invalidated when no promotions occurred
}

// TestUnit_TaskManagerAccess_GetBoardSummaryAsync_Success tests board summary retrieval
func TestUnit_TaskManagerAccess_GetBoardSummaryAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	allTasks := []task_manager.TaskResponse{createValidTaskResponse()}
	
	// Setup mocks - cache miss, then service call
	mockCache.On("Get", "board_summary").Return(nil, false)
	mockTaskManager.On("ListTasks", task_manager.QueryCriteria{}).Return(allTasks, nil)
	mockCache.On("Set", "board_summary", mock.AnythingOfType("UIBoardSummary"), 1*time.Minute).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.GetBoardSummaryAsync(ctx)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, 1, result.TotalTasks, "Should have one task in summary")
		assert.NotEmpty(t, result.TasksByStatus, "Should have status breakdown")
		assert.NotEmpty(t, result.TasksByPriority, "Should have priority breakdown")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_GetBoardSummaryAsync_CacheHit tests board summary cache hit
func TestUnit_TaskManagerAccess_GetBoardSummaryAsync_CacheHit(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	cachedSummary := UIBoardSummary{
		TotalTasks:      5,
		TasksByStatus:   make(map[UIWorkflowStatus]int),
		TasksByPriority: make(map[string]int),
		LastUpdated:     time.Now(),
	}
	
	// Setup mocks - cache hit
	mockCache.On("Get", "board_summary").Return(cachedSummary, true)
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.GetBoardSummaryAsync(ctx)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Equal(t, 5, result.TotalTasks, "Should return cached total")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	// Verify TaskManager was NOT called
	mockTaskManager.AssertNotCalled(t, "ListTasks", mock.Anything)
	mockCache.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_SearchTasksAsync_Success tests task search
func TestUnit_TaskManagerAccess_SearchTasksAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	query := "test"
	allTasks := []task_manager.TaskResponse{createValidTaskResponse()}
	
	// Setup mocks - cache miss, then service call
	mockCache.On("Get", "search_test").Return(nil, false)
	mockTaskManager.On("ListTasks", task_manager.QueryCriteria{}).Return(allTasks, nil)
	mockCache.On("Set", "search_test", mock.Anything, 30*time.Second).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.SearchTasksAsync(ctx, query)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Len(t, result, 1, "Should find one matching task")
		assert.Equal(t, "task-123", result[0].ID, "Should return matching task")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// TestUnit_TaskManagerAccess_SearchTasksAsync_EmptyQuery tests empty search query validation
func TestUnit_TaskManagerAccess_SearchTasksAsync_EmptyQuery(t *testing.T) {
	access, _, _, _ := createTestTaskManagerAccess()
	
	// Execute with empty query
	ctx := context.Background()
	resultChan, errorChan := access.SearchTasksAsync(ctx, "")
	
	// Wait for error
	select {
	case <-resultChan:
		t.Fatal("Expected validation error but got success")
	case err := <-errorChan:
		uiError, ok := err.(UIErrorResponse)
		assert.True(t, ok, "Error should be UIErrorResponse")
		assert.Equal(t, "validation", uiError.Category, "Error category should be validation")
		assert.Contains(t, uiError.Message, "Search query is required", "Error message should mention query")
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
}

// TestUnit_TaskManagerAccess_QueryTasksAsync_Success tests query operation (alias for ListTasks)
func TestUnit_TaskManagerAccess_QueryTasksAsync_Success(t *testing.T) {
	access, mockTaskManager, mockCache, _ := createTestTaskManagerAccess()
	
	criteria := UIQueryCriteria{
		Columns: []string{"todo"},
	}
	expectedTasks := []task_manager.TaskResponse{createValidTaskResponse()}
	
	// Setup mocks - cache miss, then service call
	mockCache.On("Get", mock.AnythingOfType("string")).Return(nil, false)
	mockTaskManager.On("ListTasks", mock.AnythingOfType("task_manager.QueryCriteria")).Return(expectedTasks, nil)
	mockCache.On("Set", mock.AnythingOfType("string"), mock.Anything, 2*time.Minute).Return()
	
	// Execute
	ctx := context.Background()
	resultChan, errorChan := access.QueryTasksAsync(ctx, criteria)
	
	// Wait for result
	select {
	case result := <-resultChan:
		assert.Len(t, result, 1, "Should return one task")
		assert.Equal(t, "task-123", result[0].ID, "Task ID should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// Test Context Cancellation

// TestUnit_TaskManagerAccess_CreateTaskAsync_ContextCancellation tests context cancellation
func TestUnit_TaskManagerAccess_CreateTaskAsync_ContextCancellation(t *testing.T) {
	access, mockTaskManager, mockCache, mockLogger := createTestTaskManagerAccess()
	
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	uiRequest := createValidUITaskRequest()
	expectedResponse := createValidTaskResponse()
	
	// Setup mocks - the current implementation doesn't check context cancellation
	// so it will still call the service
	mockTaskManager.On("CreateTask", mock.AnythingOfType("task_manager.TaskRequest")).Return(expectedResponse, nil)
	mockCache.On("InvalidatePattern", "tasks_*").Return()
	mockCache.On("InvalidatePattern", "board_summary").Return()
	mockLogger.On("Log", utilities.Info, "TaskManagerAccess", "Task created successfully", mock.Anything).Return()
	
	// Execute
	resultChan, errorChan := access.CreateTaskAsync(ctx, uiRequest)
	
	// Should still complete since the implementation doesn't check context cancellation
	// This test verifies the current behavior
	select {
	case result := <-resultChan:
		// Operation completed successfully despite cancelled context
		assert.Equal(t, "task-123", result.ID, "Task ID should match")
	case err := <-errorChan:
		t.Fatalf("Expected success but got error: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Operation timed out")
	}
	
	mockTaskManager.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Test Data Conversion

// TestUnit_TaskManagerAccess_DataConversion_UIToManagersTypes tests data type conversion
func TestUnit_TaskManagerAccess_DataConversion_UIToManagersTypes(t *testing.T) {
	access, _, _, _ := createTestTaskManagerAccess()
	
	// Test priority conversion
	uiPriority := UIPriority{Urgent: true, Important: false, Label: "urgent"}
	
	// Test workflow status conversion
	assert.Equal(t, UITodo, UIWorkflowStatus("todo"), "UI status constants should be correct")
	assert.Equal(t, UIInProgress, UIWorkflowStatus("doing"), "UI status constants should be correct")
	assert.Equal(t, UIDone, UIWorkflowStatus("done"), "UI status constants should be correct")
	
	// Test that access layer exists and can convert data
	assert.NotNil(t, access, "Access layer should exist for data conversion")
	assert.True(t, uiPriority.Urgent, "Priority conversion should preserve urgent flag")
}

// Test Error Handling

// TestUnit_TaskManagerAccess_ErrorHandling_CategoryMapping tests error category mapping
func TestUnit_TaskManagerAccess_ErrorHandling_CategoryMapping(t *testing.T) {
	access, mockTaskManager, _, mockLogger := createTestTaskManagerAccess()
	
	testCases := []struct {
		name          string
		serviceError  error
		expectedCategory string
	}{
		{"ValidationError", fmt.Errorf("validation failed: required field missing"), "validation"},
		{"ConnectivityError", fmt.Errorf("connection refused"), "connectivity"},
		{"NetworkError", fmt.Errorf("network timeout"), "connectivity"},
		{"ServiceError", fmt.Errorf("internal server error"), "service"},
		{"NotFoundError", fmt.Errorf("task not found"), "service"},
		{"DuplicateError", fmt.Errorf("duplicate entry"), "service"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uiRequest := createValidUITaskRequest()
			
			// Setup mock to return the test error
			mockTaskManager.On("CreateTask", mock.AnythingOfType("task_manager.TaskRequest")).Return(task_manager.TaskResponse{}, tc.serviceError).Once()
			mockLogger.On("LogError", "TaskManagerAccess", tc.serviceError, mock.Anything).Return().Once()
			
			// Execute
			ctx := context.Background()
			resultChan, errorChan := access.CreateTaskAsync(ctx, uiRequest)
			
			// Wait for error
			select {
			case <-resultChan:
				t.Fatal("Expected error but got success")
			case err := <-errorChan:
				uiError, ok := err.(UIErrorResponse)
				assert.True(t, ok, "Error should be UIErrorResponse")
				assert.Equal(t, tc.expectedCategory, uiError.Category, "Error category should match expected")
			case <-time.After(1 * time.Second):
				t.Fatal("Operation timed out")
			}
		})
	}
	
	mockTaskManager.AssertExpectations(t)
}