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
	return &board_access.BoardConfiguration{
		Name:    "Test Board",
		Columns: []string{"todo", "doing", "done"},
		Sections: map[string][]string{
			"todo": {"urgent-important", "urgent-not-important", "not-urgent-important"},
		},
	}, nil
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

// Integration Tests for Board Operations (OP-9 to OP-13)

// TestIntegration_TaskManager_ValidateBoardDirectory tests OP-9 board validation
func TestIntegration_TaskManager_ValidateBoardDirectory(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	// Test board directory validation
	response, err := taskManager.ValidateBoardDirectory("/test/board/path")
	if err != nil {
		t.Fatalf("Expected board validation to succeed, got error: %v", err)
	}

	if !response.IsValid {
		t.Errorf("Expected board to be valid, got invalid")
	}

	if !response.GitRepoValid {
		t.Errorf("Expected git repo to be valid, got invalid")
	}

	if !response.ConfigValid {
		t.Errorf("Expected config to be valid, got invalid")
	}
}

// TestIntegration_TaskManager_GetBoardMetadata tests OP-10 metadata extraction
func TestIntegration_TaskManager_GetBoardMetadata(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	// Test board metadata extraction
	response, err := taskManager.GetBoardMetadata("/test/board/path")
	if err != nil {
		t.Fatalf("Expected board metadata extraction to succeed, got error: %v", err)
	}

	if response.Title != "Mock Board" {
		t.Errorf("Expected title to be 'Mock Board', got '%s'", response.Title)
	}

	if response.TaskCount != 0 {
		t.Errorf("Expected task count to be 0, got %d", response.TaskCount)
	}
}

// TestIntegration_TaskManager_CreateBoard tests OP-11 board creation
func TestIntegration_TaskManager_CreateBoard(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	// Test board creation
	request := BoardCreationRequest{
		BoardPath:     "/test/new/board",
		Title:         "New Test Board",
		Description:   "A new test board",
		InitializeGit: true,
		Metadata:      map[string]string{"type": "test"},
	}

	response, err := taskManager.CreateBoard(request)
	if err != nil {
		t.Fatalf("Expected board creation to succeed, got error: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected board creation to succeed, got failure")
	}

	if response.BoardPath != request.BoardPath {
		t.Errorf("Expected board path to be '%s', got '%s'", request.BoardPath, response.BoardPath)
	}

	if !response.GitInitialized {
		t.Errorf("Expected git to be initialized, got false")
	}
}

// TestIntegration_TaskManager_UpdateBoardMetadata tests OP-12 metadata update
func TestIntegration_TaskManager_UpdateBoardMetadata(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	// Test board metadata update
	metadata := BoardMetadataRequest{
		Title:       "Updated Board Title",
		Description: "Updated description",
		Metadata:    map[string]string{"version": "2.0"},
	}

	response, err := taskManager.UpdateBoardMetadata("/test/board/path", metadata)
	if err != nil {
		t.Fatalf("Expected board metadata update to succeed, got error: %v", err)
	}

	// The response comes from GetBoardMetadata, so we check mock values
	if response.Title != "Mock Board" { // Mock returns original title
		t.Errorf("Expected title from mock, got '%s'", response.Title)
	}
}

// TestIntegration_TaskManager_DeleteBoard tests OP-13 board deletion
func TestIntegration_TaskManager_DeleteBoard(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	// Test board deletion
	request := BoardDeletionRequest{
		BoardPath:      "/test/board/path",
		UseTrash:       false,
		CreateBackup:   false,
		ForceDelete:    true,
	}

	response, err := taskManager.DeleteBoard(request)
	if err != nil {
		t.Fatalf("Expected board deletion to succeed, got error: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected board deletion to succeed, got failure")
	}

	if response.Method != "permanent" {
		t.Errorf("Expected deletion method to be 'permanent', got '%s'", response.Method)
	}
}

// TestIntegration_TaskManager_BoardOperationsWorkflow tests full board workflow
func TestIntegration_TaskManager_BoardOperationsWorkflow(t *testing.T) {
	boardAccess := &MockBoardAccess{}
	ruleEngine := &MockRuleEngine{}
	logger := &MockLogger{}
	repository := &MockRepository{}
	boardPath := "/test/path"

	taskManager := NewTaskManager(boardAccess, ruleEngine, logger, repository, boardPath)

	// Step 1: Validate board directory
	validationResponse, err := taskManager.ValidateBoardDirectory("/test/workflow/board")
	if err != nil {
		t.Fatalf("Board validation failed: %v", err)
	}
	if !validationResponse.IsValid {
		t.Fatalf("Expected board to be valid")
	}

	// Step 2: Create board
	createRequest := BoardCreationRequest{
		BoardPath:     "/test/workflow/board",
		Title:         "Workflow Test Board",
		Description:   "Board for testing workflow",
		InitializeGit: true,
	}

	createResponse, err := taskManager.CreateBoard(createRequest)
	if err != nil {
		t.Fatalf("Board creation failed: %v", err)
	}
	if !createResponse.Success {
		t.Fatalf("Expected board creation to succeed")
	}

	// Step 3: Get board metadata
	metadataResponse, err := taskManager.GetBoardMetadata("/test/workflow/board")
	if err != nil {
		t.Fatalf("Board metadata extraction failed: %v", err)
	}
	if metadataResponse.Title == "" {
		t.Fatalf("Expected board metadata to have title")
	}

	// Step 4: Update board metadata
	updateRequest := BoardMetadataRequest{
		Title:       "Updated Workflow Board",
		Description: "Updated description",
	}

	_, err = taskManager.UpdateBoardMetadata("/test/workflow/board", updateRequest)
	if err != nil {
		t.Fatalf("Board metadata update failed: %v", err)
	}

	// Step 5: Delete board
	deleteRequest := BoardDeletionRequest{
		BoardPath:   "/test/workflow/board",
		ForceDelete: true,
	}

	deleteResponse, err := taskManager.DeleteBoard(deleteRequest)
	if err != nil {
		t.Fatalf("Board deletion failed: %v", err)
	}
	if !deleteResponse.Success {
		t.Fatalf("Expected board deletion to succeed")
	}
}