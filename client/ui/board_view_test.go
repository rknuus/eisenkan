package ui

import (
	"fmt"
	"testing"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
)

// TestNewBoardView verifies BoardView creation with default Eisenhower Matrix configuration
func TestNewBoardView(t *testing.T) {
	// Create mock dependencies
	var workflowManager managers.WorkflowManager // Will be nil for now
	validationEngine := engines.NewFormValidationEngine()

	// Create BoardView with default configuration
	board := NewBoardView(workflowManager, validationEngine, nil)
	if board == nil {
		t.Fatal("NewBoardView returned nil")
	}

	// Verify initial state
	state := board.GetBoardState()
	if state == nil {
		t.Fatal("GetBoardState returned nil")
	}

	// Verify default Eisenhower Matrix configuration
	if state.Configuration == nil {
		t.Fatal("Configuration is nil")
	}

	if state.Configuration.Title != "Eisenhower Matrix" {
		t.Errorf("Expected title 'Eisenhower Matrix', got '%s'", state.Configuration.Title)
	}

	if state.Configuration.BoardType != "eisenhower" {
		t.Errorf("Expected board type 'eisenhower', got '%s'", state.Configuration.BoardType)
	}

	if len(state.Configuration.Columns) != 4 {
		t.Errorf("Expected 4 columns for Eisenhower Matrix, got %d", len(state.Configuration.Columns))
	}

	// Verify default column configurations
	expectedColumns := []string{
		"Urgent Important",
		"Urgent Non-Important",
		"Non-Urgent Important",
		"Non-Urgent Non-Important",
	}

	for i, expectedTitle := range expectedColumns {
		if i >= len(state.Configuration.Columns) {
			t.Fatalf("Missing column %d", i)
		}
		if state.Configuration.Columns[i].Title != expectedTitle {
			t.Errorf("Column %d: expected title '%s', got '%s'", i, expectedTitle, state.Configuration.Columns[i].Title)
		}
	}

	// Cleanup
	board.Destroy()
}

// TestBoardViewWithCustomConfiguration verifies BoardView creation with custom configuration
func TestBoardViewWithCustomConfiguration(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()

	// Create custom configuration
	customConfig := &BoardConfiguration{
		Title:     "Custom Kanban",
		BoardType: "kanban",
		EnableDragDrop: true,
		Columns: []*ColumnConfiguration{
			{
				Title:   "To Do",
				Type:    TodoColumn,
				WIPLimit: 5,
			},
			{
				Title:   "In Progress",
				Type:    DoingColumn,
				WIPLimit: 3,
			},
			{
				Title:   "Done",
				Type:    DoneColumn,
				WIPLimit: 0,
			},
		},
	}

	// Create BoardView with custom configuration
	board := NewBoardView(nil, validationEngine, customConfig)
	if board == nil {
		t.Fatal("NewBoardView returned nil")
	}

	// Verify custom configuration
	state := board.GetBoardState()
	if state.Configuration.Title != "Custom Kanban" {
		t.Errorf("Expected title 'Custom Kanban', got '%s'", state.Configuration.Title)
	}

	if state.Configuration.BoardType != "kanban" {
		t.Errorf("Expected board type 'kanban', got '%s'", state.Configuration.BoardType)
	}

	if len(state.Configuration.Columns) != 3 {
		t.Errorf("Expected 3 columns for custom kanban, got %d", len(state.Configuration.Columns))
	}

	// Cleanup
	board.Destroy()
}

// TestBoardViewStateManagement verifies state management operations
func TestBoardViewStateManagement(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	board := NewBoardView(nil, validationEngine, nil)
	defer board.Destroy()

	// Test loading state
	board.SetLoading(true)
	state := board.GetBoardState()
	if !state.IsLoading {
		t.Error("Expected loading state to be true")
	}

	board.SetLoading(false)
	state = board.GetBoardState()
	if state.IsLoading {
		t.Error("Expected loading state to be false")
	}

	// Test error state
	testError := "Test error message"
	board.SetError(fmt.Errorf("%s", testError))
	state = board.GetBoardState()
	if !state.HasError {
		t.Error("Expected error state to be true")
	}
	if state.ErrorMessage != testError {
		t.Errorf("Expected error message '%s', got '%s'", testError, state.ErrorMessage)
	}

	// Clear error
	board.SetError(nil)
	state = board.GetBoardState()
	if state.HasError {
		t.Error("Expected error state to be false after clearing")
	}
}

// TestBoardViewColumnManagement verifies column management operations
func TestBoardViewColumnManagement(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	board := NewBoardView(nil, validationEngine, nil)
	defer board.Destroy()

	state := board.GetBoardState()

	// Verify columns were created
	if len(state.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(state.Columns))
	}

	// Test column task retrieval (should be empty initially)
	for i := 0; i < 4; i++ {
		tasks := board.GetColumnTasks(i)
		if len(tasks) != 0 {
			t.Errorf("Column %d: expected 0 tasks, got %d", i, len(tasks))
		}
	}

	// Test invalid column indices
	tasks := board.GetColumnTasks(-1)
	if len(tasks) != 0 {
		t.Error("Expected empty task list for invalid column index")
	}

	tasks = board.GetColumnTasks(10)
	if len(tasks) != 0 {
		t.Error("Expected empty task list for invalid column index")
	}
}

// TestBoardViewTaskMovement verifies task movement validation
func TestBoardViewTaskMovement(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	board := NewBoardView(nil, validationEngine, nil)
	defer board.Destroy()

	// Test task movement with invalid indices
	err := board.MoveTask("test-task", -1, 0)
	if err == nil {
		t.Error("Expected error for invalid from column index")
	}

	err = board.MoveTask("test-task", 0, -1)
	if err == nil {
		t.Error("Expected error for invalid to column index")
	}

	err = board.MoveTask("test-task", 10, 0)
	if err == nil {
		t.Error("Expected error for out of range from column index")
	}

	err = board.MoveTask("test-task", 0, 10)
	if err == nil {
		t.Error("Expected error for out of range to column index")
	}
}

// TestBoardViewEventHandlers verifies event handler registration
func TestBoardViewEventHandlers(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	board := NewBoardView(nil, validationEngine, nil)
	defer board.Destroy()

	// Test event handler registration (should not panic)
	board.SetOnTaskMoved(func(taskID, fromColumn, toColumn string) {
		// Handler registered successfully
	})

	board.SetOnTaskSelected(func(taskID string) {
		// Handler registered successfully
	})

	board.SetOnBoardRefreshed(func() {
		// Handler registered successfully
	})

	board.SetOnError(func(error) {
		// Handler registered successfully
	})

	board.SetOnConfigChanged(func(*BoardConfiguration) {
		// Handler registered successfully
	})

	// Test that handlers were registered (indirect verification through state changes)
	board.SelectTask("test-task") // This should not panic even with nil WorkflowManager

	// Note: We can't easily test handler execution without mocking WorkflowManager
	// The handlers will be tested in integration tests
}

// TestBoardViewTaskMatching verifies task-to-column matching logic
func TestBoardViewTaskMatching(t *testing.T) {
	validationEngine := engines.NewFormValidationEngine()
	board := NewBoardView(nil, validationEngine, nil)
	defer board.Destroy()

	// Test Eisenhower Matrix task matching
	state := board.GetBoardState()

	// Create test tasks with different priorities
	tasks := []*TaskData{
		{
			ID:       "task1",
			Priority: "urgent important",
		},
		{
			ID:       "task2",
			Priority: "urgent non-important",
		},
		{
			ID:       "task3",
			Priority: "non-urgent important",
		},
		{
			ID:       "task4",
			Priority: "non-urgent non-important",
		},
	}

	// Test task matching for each column
	for i, task := range tasks {
		columnConfig := state.Configuration.Columns[i]
		matches := board.taskBelongsToColumn(task, columnConfig)
		if !matches {
			t.Errorf("Task %s should match column %s", task.ID, columnConfig.Title)
		}
	}

	// Test that task doesn't match wrong column
	wrongTask := &TaskData{
		ID:       "wrong-task",
		Priority: "urgent important",
	}

	// Should not match "Non-Urgent Non-Important" column
	wrongColumnConfig := state.Configuration.Columns[3]
	matches := board.taskBelongsToColumn(wrongTask, wrongColumnConfig)
	if matches {
		t.Error("Task should not match wrong column")
	}
}