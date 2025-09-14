package resource_access

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnit_BoardAccess_NewBoardAccess(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "boardaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating new BoardAccess
	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer ba.Close()

	// Verify it implements the interface
	var _ IBoardAccess = ba
}

func TestUnit_BoardAccess_StoreAndGetTasksData(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "boardaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create BoardAccess
	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer ba.Close()

	// Create test task
	task := &Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Tags:        []string{"test", "sample"},
	}

	// Define priority and status as separate parameters
	priority := Priority{
		Urgent:    true,
		Important: true,
	}
	status := WorkflowStatus{
		Column:   "todo",
		Section:  "urgent-important",
		Position: 1,
	}

	// Store task with priority and status parameters
	taskID, err := ba.CreateTask(task, priority, status, nil)
	if err != nil {
		t.Fatalf("Failed to store task: %v", err)
	}

	if taskID == "" {
		t.Fatal("Task ID should not be empty")
	}

	// Retrieve task using new combined method
	tasksWithTimestamps, err := ba.GetTasksData([]string{taskID}, false)
	if err != nil {
		t.Fatalf("Failed to retrieve tasks: %v", err)
	}

	if len(tasksWithTimestamps) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasksWithTimestamps))
	}

	retrievedTaskWithTimestamps := tasksWithTimestamps[0]
	retrievedTask := retrievedTaskWithTimestamps.Task

	// Verify task data
	if retrievedTask.ID != taskID {
		t.Errorf("Expected task ID %s, got %s", taskID, retrievedTask.ID)
	}

	if retrievedTask.Title != task.Title {
		t.Errorf("Expected task title %s, got %s", task.Title, retrievedTask.Title)
	}

	if retrievedTaskWithTimestamps.Priority.Label != "urgent-important" {
		t.Errorf("Expected priority label 'urgent-important', got %s", retrievedTaskWithTimestamps.Priority.Label)
	}

	// Verify timestamps are populated from git
	if retrievedTaskWithTimestamps.CreatedAt.IsZero() {
		t.Error("CreatedAt timestamp should not be zero")
	}

	if retrievedTaskWithTimestamps.UpdatedAt.IsZero() {
		t.Error("UpdatedAt timestamp should not be zero")
	}

	// Verify file structure with position prefix
	expectedPath := filepath.Join(tempDir, "01_todo", "urgent-important", "0001-task-"+taskID+".json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected task file does not exist at %s", expectedPath)
	}
}

func TestUnit_BoardAccess_BoardConfiguration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "boardaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create BoardAccess
	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer ba.Close()

	// Get default configuration
	config, err := ba.GetBoardConfiguration()
	if err != nil {
		t.Fatalf("Failed to get board configuration: %v", err)
	}

	if config == nil {
		t.Fatal("Board configuration should not be nil")
	}

	if config.Name != "EisenKan Board" {
		t.Errorf("Expected board name 'EisenKan Board', got %s", config.Name)
	}

	if len(config.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(config.Columns))
	}

	// Verify git config is included
	if config.GitUser == "" {
		t.Error("GitUser should not be empty")
	}

	if config.GitEmail == "" {
		t.Error("GitEmail should not be empty")
	}

	// Update configuration
	config.Name = "Updated Board"
	config.GitUser = "TestUser"
	config.GitEmail = "test@example.com"

	err = ba.UpdateBoardConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to update board configuration: %v", err)
	}

	// Retrieve updated configuration
	updatedConfig, err := ba.GetBoardConfiguration()
	if err != nil {
		t.Fatalf("Failed to get updated board configuration: %v", err)
	}

	if updatedConfig.Name != "Updated Board" {
		t.Errorf("Expected updated board name 'Updated Board', got %s", updatedConfig.Name)
	}

	if updatedConfig.GitUser != "TestUser" {
		t.Errorf("Expected git user 'TestUser', got %s", updatedConfig.GitUser)
	}
}

func TestUnit_BoardAccess_ArchiveTask(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "boardaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create BoardAccess
	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer ba.Close()

	// Create and store test task
	task := &Task{
		Title: "Task to Archive",
	}

	priority := Priority{
		Urgent:    false,
		Important: true,
	}
	status := WorkflowStatus{
		Column:   "done",
		Position: 1,
	}

	taskID, err := ba.CreateTask(task, priority, status, nil)
	if err != nil {
		t.Fatalf("Failed to store task: %v", err)
	}

	// Archive task
	err = ba.ArchiveTask(taskID, NoAction)
	if err != nil {
		t.Fatalf("Failed to archive task: %v", err)
	}

	// Verify task is in archived directory with position prefix
	archivedPath := filepath.Join(tempDir, "archived", "0001-task-"+taskID+".json")
	if _, err := os.Stat(archivedPath); os.IsNotExist(err) {
		t.Errorf("Archived task file does not exist at %s", archivedPath)
	}

	// Verify original file is removed
	originalPath := filepath.Join(tempDir, "03_done", "0001-task-"+taskID+".json")
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Errorf("Original task file should be removed from %s", originalPath)
	}

	// Verify task can still be retrieved
	archivedTasks, err := ba.GetTasksData([]string{taskID}, false)
	if err != nil {
		t.Fatalf("Failed to retrieve archived task: %v", err)
	}

	if len(archivedTasks) != 1 {
		t.Fatalf("Expected 1 archived task, got %d", len(archivedTasks))
	}

	if archivedTasks[0].Status.Column != "archived" {
		t.Errorf("Expected archived task column to be 'archived', got %s", archivedTasks[0].Status.Column)
	}
}

func TestUnit_BoardAccess_FindTasks(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "boardaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create BoardAccess
	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer ba.Close()

	// Create test tasks
	testData := []struct {
		task     *Task
		priority Priority
		status   WorkflowStatus
	}{
		{
			task: &Task{
				Title: "Urgent Task",
				Tags:  []string{"urgent"},
			},
			priority: Priority{Urgent: true, Important: true},
			status:   WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1},
		},
		{
			task: &Task{
				Title: "Important Task",
				Tags:  []string{"important"},
			},
			priority: Priority{Urgent: false, Important: true},
			status:   WorkflowStatus{Column: "todo", Section: "not-urgent-important", Position: 2},
		},
		{
			task: &Task{
				Title: "Done Task",
				Tags:  []string{"completed"},
			},
			priority: Priority{Urgent: true, Important: false},
			status:   WorkflowStatus{Column: "done", Position: 1},
		},
	}

	// Store all tasks
	for _, td := range testData {
		_, err := ba.CreateTask(td.task, td.priority, td.status, nil)
		if err != nil {
			t.Fatalf("Failed to store task %s: %v", td.task.Title, err)
		}
	}

	// Query by column
	criteria := &QueryCriteria{
		Columns: []string{"todo"},
	}

	results, err := ba.FindTasks(criteria)
	if err != nil {
		t.Fatalf("Failed to query tasks: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 todo tasks, got %d", len(results))
	}

	// Query by priority
	criteria = &QueryCriteria{
		Priority: &Priority{Urgent: true, Important: true},
	}

	results, err = ba.FindTasks(criteria)
	if err != nil {
		t.Fatalf("Failed to query tasks by priority: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 urgent-important task, got %d", len(results))
	}

	if len(results) > 0 && results[0].Task.Title != "Urgent Task" {
		t.Errorf("Expected urgent task title 'Urgent Task', got %s", results[0].Task.Title)
	}
}

func TestUnit_BoardAccess_GetTaskHistory(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "boardaccess_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create BoardAccess
	ba, err := NewBoardAccess(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BoardAccess: %v", err)
	}
	defer ba.Close()

	// Create and store test task
	task := &Task{
		Title: "Test Task for History",
	}
	priority := Priority{Urgent: true, Important: true}
	status := WorkflowStatus{Column: "todo", Section: "urgent-important", Position: 1}

	taskID, err := ba.CreateTask(task, priority, status, nil)
	if err != nil {
		t.Fatalf("Failed to store task: %v", err)
	}

	// Update task to create history
	task.Title = "Updated Task Title"
	err = ba.ChangeTaskData(taskID, task, priority, status)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Get task history with configurable limit
	history, err := ba.GetTaskHistory(taskID, 10)
	if err != nil {
		t.Fatalf("Failed to get task history: %v", err)
	}

	if len(history) < 2 {
		t.Errorf("Expected at least 2 history entries, got %d", len(history))
	}

	// Test default limit
	historyDefault, err := ba.GetTaskHistory(taskID, 0) // Should use default limit of 100
	if err != nil {
		t.Fatalf("Failed to get task history with default limit: %v", err)
	}

	if len(historyDefault) != len(history) {
		t.Errorf("Expected same history length with default limit, got %d vs %d", len(historyDefault), len(history))
	}
}
