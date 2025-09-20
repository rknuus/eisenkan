package task_manager

import (
	"context"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// MockBoardAccess implements IBoardAccess for testing
type MockBoardAccess struct{}

func (m *MockBoardAccess) CreateTask(task *board_access.Task, priority board_access.Priority, status board_access.WorkflowStatus, parentTaskID *string) (string, error) {
	return "test-task-id", nil
}

func (m *MockBoardAccess) GetTasksData(taskIDs []string, includeHierarchy bool) ([]*board_access.TaskWithTimestamps, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	
	return []*board_access.TaskWithTimestamps{
		{
			Task: &board_access.Task{
				ID:          taskIDs[0],
				Title:       "Test Task",
				Description: "Test Description",
			},
			Priority:  board_access.Priority{Urgent: false, Important: true, Label: "not-urgent-important"},
			Status:    board_access.WorkflowStatus{Column: "todo"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}, nil
}

func (m *MockBoardAccess) ListTaskIdentifiers(hierarchyFilter board_access.HierarchyFilter) ([]string, error) {
	return []string{"test-task-id"}, nil
}

func (m *MockBoardAccess) ChangeTaskData(taskID string, task *board_access.Task, priority board_access.Priority, status board_access.WorkflowStatus) error {
	return nil
}

func (m *MockBoardAccess) MoveTask(taskID string, priority board_access.Priority, status board_access.WorkflowStatus) error {
	return nil
}

func (m *MockBoardAccess) ArchiveTask(taskID string, cascadePolicy board_access.CascadePolicy) error {
	return nil
}

func (m *MockBoardAccess) RemoveTask(taskID string, cascadePolicy board_access.CascadePolicy) error {
	return nil
}

func (m *MockBoardAccess) FindTasks(criteria *board_access.QueryCriteria) ([]*board_access.TaskWithTimestamps, error) {
	return []*board_access.TaskWithTimestamps{}, nil
}

func (m *MockBoardAccess) GetTaskHistory(taskID string, limit int) ([]utilities.CommitInfo, error) {
	return []utilities.CommitInfo{}, nil
}

func (m *MockBoardAccess) GetSubtasks(parentTaskID string) ([]*board_access.TaskWithTimestamps, error) {
	return []*board_access.TaskWithTimestamps{}, nil
}

func (m *MockBoardAccess) GetParentTask(subtaskID string) (*board_access.TaskWithTimestamps, error) {
	return nil, nil
}

func (m *MockBoardAccess) GetBoardConfiguration() (*board_access.BoardConfiguration, error) {
	return &board_access.BoardConfiguration{}, nil
}

func (m *MockBoardAccess) UpdateBoardConfiguration(config *board_access.BoardConfiguration) error {
	return nil
}

func (m *MockBoardAccess) GetRulesData(taskID string, targetColumns []string) (*board_access.RulesData, error) {
	return &board_access.RulesData{}, nil
}

func (m *MockBoardAccess) Close() error {
	return nil
}

// IConfiguration facet mock methods
func (m *MockBoardAccess) Load(configType string, identifier string) (board_access.ConfigurationData, error) {
	return board_access.ConfigurationData{
		Type:       configType,
		Identifier: identifier,
		Version:    "1.0",
		Settings:   make(map[string]interface{}),
		Schema:     "default",
		Metadata:   make(map[string]string),
	}, nil
}

func (m *MockBoardAccess) Store(configType string, identifier string, data board_access.ConfigurationData) error {
	return nil
}

// IBoard facet mock methods
func (m *MockBoardAccess) Discover(ctx context.Context, directoryPath string) ([]board_access.BoardDiscoveryResult, error) {
	return []board_access.BoardDiscoveryResult{}, nil
}

func (m *MockBoardAccess) ExtractMetadata(ctx context.Context, boardPath string) (*board_access.BoardMetadata, error) {
	return &board_access.BoardMetadata{
		Title:     "Mock Board",
		TaskCount: 0,
	}, nil
}

func (m *MockBoardAccess) GetStatistics(ctx context.Context, boardPath string) (*board_access.BoardStatistics, error) {
	return &board_access.BoardStatistics{
		TotalTasks:    0,
		ActiveTasks:   0,
		TasksByColumn: make(map[string]int),
	}, nil
}

func (m *MockBoardAccess) ValidateStructure(ctx context.Context, boardPath string) (*board_access.BoardValidationResult, error) {
	return &board_access.BoardValidationResult{
		IsValid:       true,
		GitRepoValid:  true,
		ConfigValid:   true,
		DataIntegrity: true,
	}, nil
}

func (m *MockBoardAccess) LoadConfiguration(ctx context.Context, boardPath string, configType string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":    "Mock Board",
		"columns": []string{"todo", "doing", "done"},
	}, nil
}

func (m *MockBoardAccess) StoreConfiguration(ctx context.Context, boardPath string, configType string, configData map[string]interface{}) error {
	return nil
}

func (m *MockBoardAccess) Create(ctx context.Context, request *board_access.BoardCreationRequest) (*board_access.BoardCreationResult, error) {
	return &board_access.BoardCreationResult{
		Success:        true,
		BoardPath:      request.BoardPath,
		GitInitialized: request.InitializeGit,
	}, nil
}

func (m *MockBoardAccess) Delete(ctx context.Context, request *board_access.BoardDeletionRequest) (*board_access.BoardDeletionResult, error) {
	return &board_access.BoardDeletionResult{
		Success: true,
		Method:  "permanent",
	}, nil
}

// MockRepository implements Repository for testing
type MockRepository struct{}

func (m *MockRepository) Path() string {
	return "/mock/path"
}

func (m *MockRepository) Status() (*utilities.RepositoryStatus, error) {
	return &utilities.RepositoryStatus{}, nil
}

func (m *MockRepository) Stage(patterns []string) error {
	return nil
}

func (m *MockRepository) Commit(message string) (string, error) {
	return "mock-hash", nil
}

func (m *MockRepository) GetHistory(limit int) ([]utilities.CommitInfo, error) {
	return []utilities.CommitInfo{}, nil
}

func (m *MockRepository) GetHistoryStream() <-chan utilities.CommitInfo {
	ch := make(chan utilities.CommitInfo)
	close(ch)
	return ch
}

func (m *MockRepository) GetFileHistory(filePath string, limit int) ([]utilities.CommitInfo, error) {
	return []utilities.CommitInfo{}, nil
}

func (m *MockRepository) GetFileHistoryStream(filePath string) <-chan utilities.CommitInfo {
	ch := make(chan utilities.CommitInfo)
	close(ch)
	return ch
}

func (m *MockRepository) GetFileDifferences(hash1, hash2 string) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockRepository) ValidateRepositoryAndPaths(request utilities.RepositoryValidationRequest) (*utilities.RepositoryValidationResult, error) {
	return &utilities.RepositoryValidationResult{
		RepositoryValid: true,
		ExistingPaths:   []string{},
		MissingPaths:    []string{},
	}, nil
}

func (m *MockRepository) Close() error {
	return nil
}

// MockRuleEngine implements IRuleEngine for testing
type MockRuleEngine struct{}

func (m *MockRuleEngine) EvaluateTaskChange(ctx context.Context, event engines.TaskEvent, boardPath string) (*engines.RuleEvaluationResult, error) {
	return &engines.RuleEvaluationResult{
		Allowed:    true,
		Violations: []engines.RuleViolation{},
	}, nil
}

func (m *MockRuleEngine) EvaluateBoardConfigurationChange(ctx context.Context, event engines.BoardConfigurationEvent) (*engines.RuleEvaluationResult, error) {
	return &engines.RuleEvaluationResult{
		Allowed:    true,
		Violations: []engines.RuleViolation{},
	}, nil
}

func (m *MockRuleEngine) Close() error {
	return nil
}

// MockLogger implements ILoggingUtility for testing  
type MockLogger struct{}

func (m *MockLogger) Log(level utilities.LogLevel, component string, message string, data interface{}) {}

func (m *MockLogger) LogMessage(level utilities.LogLevel, component string, message string) {}

func (m *MockLogger) LogError(component string, err error, data interface{}) {}

func (m *MockLogger) IsLevelEnabled(level utilities.LogLevel) bool {
	return false
}

func TestTaskManagerCreation(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)
	if taskManager == nil {
		t.Fatal("Expected TaskManager to be created, got nil")
	}
}

func TestCreateTask(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	request := TaskRequest{
		Description:    "Test task",
		Priority:       board_access.Priority{Urgent: false, Important: true, Label: "not-urgent-important"},
		WorkflowStatus: Todo,
		Tags:           []string{"test"},
	}

	response, err := taskManager.CreateTask(request)
	if err != nil {
		t.Fatalf("Expected task creation to succeed, got error: %v", err)
	}

	if response.ID != "test-task-id" {
		t.Errorf("Expected task ID to be 'test-task-id', got '%s'", response.ID)
	}

	if response.Description != "Test Description" {
		t.Errorf("Expected description to be 'Test Description', got '%s'", response.Description)
	}
}