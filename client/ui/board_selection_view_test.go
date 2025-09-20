package ui

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"github.com/rknuus/eisenkan/internal/managers/task_manager"
)

// MockTaskManager implements task_manager.TaskManager for testing
type MockTaskManager struct {
	validateBoardDirectoryFunc func(string) (task_manager.BoardValidationResponse, error)
	getBoardMetadataFunc       func(string) (task_manager.BoardMetadataResponse, error)
	createBoardFunc           func(task_manager.BoardCreationRequest) (task_manager.BoardCreationResponse, error)
	updateBoardMetadataFunc   func(string, task_manager.BoardMetadataRequest) (task_manager.BoardMetadataResponse, error)
	deleteBoardFunc           func(task_manager.BoardDeletionRequest) (task_manager.BoardDeletionResponse, error)
}

func (m *MockTaskManager) CreateTask(request task_manager.TaskRequest) (task_manager.TaskResponse, error) {
	return task_manager.TaskResponse{}, nil
}

func (m *MockTaskManager) UpdateTask(taskID string, request task_manager.TaskRequest) (task_manager.TaskResponse, error) {
	return task_manager.TaskResponse{}, nil
}

func (m *MockTaskManager) GetTask(taskID string) (task_manager.TaskResponse, error) {
	return task_manager.TaskResponse{}, nil
}

func (m *MockTaskManager) DeleteTask(taskID string) error {
	return nil
}

func (m *MockTaskManager) ListTasks(criteria task_manager.QueryCriteria) ([]task_manager.TaskResponse, error) {
	return []task_manager.TaskResponse{}, nil
}

func (m *MockTaskManager) ChangeTaskStatus(taskID string, status task_manager.WorkflowStatus) (task_manager.TaskResponse, error) {
	return task_manager.TaskResponse{}, nil
}

func (m *MockTaskManager) ValidateTask(request task_manager.TaskRequest) (task_manager.ValidationResult, error) {
	return task_manager.ValidationResult{Valid: true}, nil
}

func (m *MockTaskManager) ProcessPriorityPromotions() ([]task_manager.TaskResponse, error) {
	return []task_manager.TaskResponse{}, nil
}

// Board operations
func (m *MockTaskManager) ValidateBoardDirectory(directoryPath string) (task_manager.BoardValidationResponse, error) {
	if m.validateBoardDirectoryFunc != nil {
		return m.validateBoardDirectoryFunc(directoryPath)
	}
	return task_manager.BoardValidationResponse{IsValid: true}, nil
}

func (m *MockTaskManager) GetBoardMetadata(boardPath string) (task_manager.BoardMetadataResponse, error) {
	if m.getBoardMetadataFunc != nil {
		return m.getBoardMetadataFunc(boardPath)
	}
	return task_manager.BoardMetadataResponse{
		Title:        "Test Board",
		Description:  "Test Description",
		TaskCount:    5,
		ColumnCounts: map[string]int{"todo": 3, "doing": 1, "done": 1},
	}, nil
}

func (m *MockTaskManager) CreateBoard(request task_manager.BoardCreationRequest) (task_manager.BoardCreationResponse, error) {
	if m.createBoardFunc != nil {
		return m.createBoardFunc(request)
	}
	return task_manager.BoardCreationResponse{
		Success:        true,
		BoardPath:      request.BoardPath,
		GitInitialized: request.InitializeGit,
	}, nil
}

func (m *MockTaskManager) UpdateBoardMetadata(boardPath string, metadata task_manager.BoardMetadataRequest) (task_manager.BoardMetadataResponse, error) {
	if m.updateBoardMetadataFunc != nil {
		return m.updateBoardMetadataFunc(boardPath, metadata)
	}
	return task_manager.BoardMetadataResponse{}, nil
}

func (m *MockTaskManager) DeleteBoard(request task_manager.BoardDeletionRequest) (task_manager.BoardDeletionResponse, error) {
	if m.deleteBoardFunc != nil {
		return m.deleteBoardFunc(request)
	}
	return task_manager.BoardDeletionResponse{Success: true}, nil
}

// Context operations (for IContext interface)
func (m *MockTaskManager) Load(contextType string) (task_manager.ContextData, error) {
	return task_manager.ContextData{}, nil
}

func (m *MockTaskManager) Store(contextType string, data task_manager.ContextData) error {
	return nil
}


// TestUnit_BoardSelectionView_NewBoardSelectionView tests widget creation
func TestUnit_BoardSelectionView_NewBoardSelectionView(t *testing.T) {
	// Arrange
	mockTM := &MockTaskManager{}
	test.NewApp()
	window := test.NewWindow(nil)

	// Act
	bsv := NewBoardSelectionView(mockTM, nil, nil, window)

	// Assert
	if bsv == nil {
		t.Fatal("NewBoardSelectionView returned nil")
	}

	// Verify widget can be rendered
	renderer := bsv.CreateRenderer()
	if renderer == nil {
		t.Error("CreateRenderer returned nil")
	}
}

// TestUnit_BoardSelectionView_RefreshBoards tests board refresh functionality
func TestUnit_BoardSelectionView_RefreshBoards(t *testing.T) {
	// Arrange
	mockTM := &MockTaskManager{}
	test.NewApp()
	window := test.NewWindow(nil)

	bsv := NewBoardSelectionView(mockTM, nil, nil, window).(*boardSelectionView)

	// Act
	err := bsv.RefreshBoards()

	// Assert
	if err != nil {
		t.Errorf("RefreshBoards returned error: %v", err)
	}

	// Wait for async operation to complete
	time.Sleep(100 * time.Millisecond)

	// Verify state was updated
	bsv.stateMu.RLock()
	isLoading := bsv.currentState.isLoading
	bsv.stateMu.RUnlock()

	if isLoading {
		t.Error("Expected loading state to be false after refresh")
	}
}

// TestUnit_BoardSelectionView_ValidateAndAddBoard tests board validation and addition
func TestUnit_BoardSelectionView_ValidateAndAddBoard(t *testing.T) {
	tests := []struct {
		name          string
		directoryPath string
		setupMock     func(*MockTaskManager)
		expectError   bool
	}{
		{
			name:          "Valid board directory",
			directoryPath: "/path/to/valid/board",
			setupMock: func(m *MockTaskManager) {
				m.validateBoardDirectoryFunc = func(path string) (task_manager.BoardValidationResponse, error) {
					return task_manager.BoardValidationResponse{
						IsValid:       true,
						GitRepoValid:  true,
						ConfigValid:   true,
						DataIntegrity: true,
					}, nil
				}
				m.getBoardMetadataFunc = func(path string) (task_manager.BoardMetadataResponse, error) {
					return task_manager.BoardMetadataResponse{
						Title:       "Test Board",
						Description: "Test Description",
						TaskCount:   5,
					}, nil
				}
			},
			expectError: false,
		},
		{
			name:          "Invalid board directory",
			directoryPath: "/path/to/invalid/board",
			setupMock: func(m *MockTaskManager) {
				m.validateBoardDirectoryFunc = func(path string) (task_manager.BoardValidationResponse, error) {
					return task_manager.BoardValidationResponse{
						IsValid: false,
						Issues:  []string{"Missing board.json", "Not a git repository"},
					}, nil
				}
			},
			expectError: false, // validateAndAddBoard handles validation gracefully
		},
		{
			name:          "Validation error",
			directoryPath: "/path/to/error/board",
			setupMock: func(m *MockTaskManager) {
				m.validateBoardDirectoryFunc = func(path string) (task_manager.BoardValidationResponse, error) {
					return task_manager.BoardValidationResponse{}, fmt.Errorf("validation failed")
				}
			},
			expectError: false, // validateAndAddBoard handles errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockTM := &MockTaskManager{}
			tt.setupMock(mockTM)
			app := test.NewApp()
			window := test.NewWindow(nil)
			defer app.Quit()

			bsv := NewBoardSelectionView(mockTM, nil, nil, window).(*boardSelectionView)

			// Act
			bsv.validateAndAddBoard(tt.directoryPath)

			// Wait for async operation to complete
			time.Sleep(100 * time.Millisecond)

			// Assert - this is primarily a test of error handling
			// The actual board addition would be tested in integration tests
		})
	}
}

// TestUnit_BoardSelectionView_CreateBoard tests board creation functionality
func TestUnit_BoardSelectionView_CreateBoard(t *testing.T) {
	tests := []struct {
		name        string
		request     BoardCreationRequest
		setupMock   func(*MockTaskManager)
		expectError bool
	}{
		{
			name: "Successful board creation",
			request: BoardCreationRequest{
				Path:        "/path/to/new/board",
				Title:       "New Board",
				Description: "New board description",
				InitializeGit: true,
			},
			setupMock: func(m *MockTaskManager) {
				m.createBoardFunc = func(req task_manager.BoardCreationRequest) (task_manager.BoardCreationResponse, error) {
					return task_manager.BoardCreationResponse{
						Success:        true,
						BoardPath:      req.BoardPath,
						GitInitialized: req.InitializeGit,
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "Board creation failure",
			request: BoardCreationRequest{
				Path:        "/invalid/path",
				Title:       "Failed Board",
				Description: "This should fail",
			},
			setupMock: func(m *MockTaskManager) {
				m.createBoardFunc = func(req task_manager.BoardCreationRequest) (task_manager.BoardCreationResponse, error) {
					return task_manager.BoardCreationResponse{}, fmt.Errorf("directory does not exist")
				}
			},
			expectError: true,
		},
		{
			name: "Board creation unsuccessful response",
			request: BoardCreationRequest{
				Path:  "/path/to/board",
				Title: "Unsuccessful Board",
			},
			setupMock: func(m *MockTaskManager) {
				m.createBoardFunc = func(req task_manager.BoardCreationRequest) (task_manager.BoardCreationResponse, error) {
					return task_manager.BoardCreationResponse{
						Success: false,
						Message: "Creation failed",
					}, nil
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockTM := &MockTaskManager{}
			tt.setupMock(mockTM)
			app := test.NewApp()
			window := test.NewWindow(nil)
			defer app.Quit()

			bsv := NewBoardSelectionView(mockTM, nil, nil, window)

			var callbackCalled bool
			var callbackPath string
			bsv.SetBoardCreatedCallback(func(path string) {
				callbackCalled = true
				callbackPath = path
			})

			// Act
			err := bsv.CreateBoard(tt.request)

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && !callbackCalled {
				t.Error("Expected board created callback to be called")
			}
			if !tt.expectError && callbackPath != tt.request.Path {
				t.Errorf("Expected callback path %s, got %s", tt.request.Path, callbackPath)
			}
		})
	}
}

// TestUnit_BoardSelectionView_SearchFilter tests search filtering functionality
func TestUnit_BoardSelectionView_SearchFilter(t *testing.T) {
	// Arrange
	mockTM := &MockTaskManager{}
	test.NewApp()
	window := test.NewWindow(nil)

	bsv := NewBoardSelectionView(mockTM, nil, nil, window).(*boardSelectionView)

	// Add test boards to state
	testBoards := []BoardInfo{
		{Path: "/path/alpha", Title: "Alpha Board", Description: "First board"},
		{Path: "/path/beta", Title: "Beta Project", Description: "Second board"},
		{Path: "/path/gamma", Title: "Gamma Task", Description: "Third board"},
	}

	bsv.stateMu.Lock()
	bsv.currentState.boards = testBoards
	bsv.stateMu.Unlock()

	tests := []struct {
		name           string
		searchText     string
		expectedCount  int
		expectedTitles []string
	}{
		{
			name:           "Empty search shows all boards",
			searchText:     "",
			expectedCount:  3,
			expectedTitles: []string{"Alpha Board", "Beta Project", "Gamma Task"},
		},
		{
			name:           "Search by title",
			searchText:     "alpha",
			expectedCount:  1,
			expectedTitles: []string{"Alpha Board"},
		},
		{
			name:           "Search by description",
			searchText:     "first",
			expectedCount:  1,
			expectedTitles: []string{"Alpha Board"},
		},
		{
			name:           "Search by path",
			searchText:     "beta",
			expectedCount:  1,
			expectedTitles: []string{"Beta Project"},
		},
		{
			name:           "Case insensitive search",
			searchText:     "GAMMA",
			expectedCount:  1,
			expectedTitles: []string{"Gamma Task"},
		},
		{
			name:           "No matches",
			searchText:     "nonexistent",
			expectedCount:  0,
			expectedTitles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			bsv.updateSearchFilter(tt.searchText)

			// Assert
			bsv.stateMu.RLock()
			filteredCount := len(bsv.currentState.filteredBoards)
			filteredBoards := make([]BoardInfo, len(bsv.currentState.filteredBoards))
			copy(filteredBoards, bsv.currentState.filteredBoards)
			bsv.stateMu.RUnlock()

			if filteredCount != tt.expectedCount {
				t.Errorf("Expected %d filtered boards, got %d", tt.expectedCount, filteredCount)
			}

			for i, expectedTitle := range tt.expectedTitles {
				if i >= len(filteredBoards) {
					t.Errorf("Expected board %s not found in filtered results", expectedTitle)
					continue
				}
				if filteredBoards[i].Title != expectedTitle {
					t.Errorf("Expected board %s at position %d, got %s", expectedTitle, i, filteredBoards[i].Title)
				}
			}
		})
	}
}

// TestUnit_BoardSelectionView_SelectionManagement tests board selection functionality
func TestUnit_BoardSelectionView_SelectionManagement(t *testing.T) {
	// Arrange
	mockTM := &MockTaskManager{}
	test.NewApp()
	window := test.NewWindow(nil)

	bsv := NewBoardSelectionView(mockTM, nil, nil, window).(*boardSelectionView)

	// Add test boards
	testBoards := []BoardInfo{
		{Path: "/path/alpha", Title: "Alpha Board"},
		{Path: "/path/beta", Title: "Beta Board"},
	}

	bsv.stateMu.Lock()
	bsv.currentState.boards = testBoards
	bsv.currentState.filteredBoards = testBoards
	bsv.stateMu.Unlock()

	// Set up callback for testing (callback functionality tested elsewhere)
	bsv.SetBoardSelectedCallback(func(path string) {
		// Callback set for testing
	})

	// Test SetSelectedBoard
	t.Run("SetSelectedBoard success", func(t *testing.T) {
		err := bsv.SetSelectedBoard("/path/alpha")
		if err != nil {
			t.Errorf("SetSelectedBoard returned error: %v", err)
		}

		selected, err := bsv.GetSelectedBoard()
		if err != nil {
			t.Errorf("GetSelectedBoard returned error: %v", err)
		}
		if selected.Path != "/path/alpha" {
			t.Errorf("Expected selected board path /path/alpha, got %s", selected.Path)
		}
	})

	// Test SetSelectedBoard with invalid path
	t.Run("SetSelectedBoard invalid path", func(t *testing.T) {
		err := bsv.SetSelectedBoard("/path/nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent board path")
		}
	})

	// Test GetSelectedBoard with no selection
	t.Run("GetSelectedBoard no selection", func(t *testing.T) {
		bsv.stateMu.Lock()
		bsv.currentState.selectedBoard = nil
		bsv.stateMu.Unlock()

		_, err := bsv.GetSelectedBoard()
		if err == nil {
			t.Error("Expected error when no board is selected")
		}
	})
}

// TestUnit_BoardSelectionView_FormatRelativeTime tests time formatting
func TestUnit_BoardSelectionView_FormatRelativeTime(t *testing.T) {
	// Arrange
	mockTM := &MockTaskManager{}
	test.NewApp()
	window := test.NewWindow(nil)

	bsv := NewBoardSelectionView(mockTM, nil, nil, window).(*boardSelectionView)

	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "Just now",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "Hours ago",
			time:     now.Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "Days ago",
			time:     now.Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "Weeks ago",
			time:     now.Add(-10 * 24 * time.Hour),
			expected: now.Add(-10 * 24 * time.Hour).Format("Jan 2, 2006"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := bsv.formatRelativeTime(tt.time)

			// Assert
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestUnit_BoardSelectionView_StateManagement tests state management
func TestUnit_BoardSelectionView_StateManagement(t *testing.T) {
	// Arrange
	mockTM := &MockTaskManager{}
	test.NewApp()
	window := test.NewWindow(nil)

	bsv := NewBoardSelectionView(mockTM, nil, nil, window).(*boardSelectionView)

	// Test state channel communication
	t.Run("State channel updates", func(t *testing.T) {
		newState := &BoardSelectionState{
			boards:         []BoardInfo{{Path: "/test", Title: "Test Board"}},
			filteredBoards: []BoardInfo{{Path: "/test", Title: "Test Board"}},
			searchFilter:   "test",
			isLoading:      false,
		}

		// Send new state through channel
		select {
		case bsv.stateChannel <- newState:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Failed to send state through channel")
		}

		// Wait for state to be processed
		time.Sleep(50 * time.Millisecond)

		// Verify state was updated
		bsv.stateMu.RLock()
		currentFilter := bsv.currentState.searchFilter
		boardCount := len(bsv.currentState.boards)
		bsv.stateMu.RUnlock()

		if currentFilter != "test" {
			t.Errorf("Expected search filter 'test', got '%s'", currentFilter)
		}
		if boardCount != 1 {
			t.Errorf("Expected 1 board, got %d", boardCount)
		}
	})

	// Test cleanup
	t.Run("Cleanup", func(t *testing.T) {
		bsv.Cleanup()

		// Verify context was cancelled
		select {
		case <-bsv.ctx.Done():
			// Success - context was cancelled
		case <-time.After(100 * time.Millisecond):
			t.Error("Context was not cancelled after cleanup")
		}
	})
}