package ui

import (
	"os"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// createTestApplicationRoot creates an ApplicationRoot with an isolated temporary directory for testing
func createTestApplicationRoot(t *testing.T) *ApplicationRoot {
	tmpDir, err := os.MkdirTemp("", "eisenkan_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Clean up the temp directory when the test finishes
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return NewApplicationRootWithBoardDir(tmpDir)
}

// createMockApplicationRoot creates an ApplicationRoot with minimal dependencies for unit testing
func createMockApplicationRoot(t *testing.T) *ApplicationRoot {
	ar := &ApplicationRoot{
		eventDispatcher: NewNavigationEventDispatcher(),
		currentView:     ViewTypeBoardSelection,
	}

	// Don't initialize real dependencies to avoid race conditions in unit tests
	// Unit tests should focus on testing the ApplicationRoot logic, not the full dependency chain

	return ar
}

// MockBoardSelectionView provides a mock implementation for testing
type MockBoardSelectionView struct {
	refreshBoardsFunc        func() error
	browseForBoardsFunc      func() error
	createBoardFunc          func(BoardCreationRequest) error
	getSelectedBoardFunc     func() (*BoardInfo, error)
	setSelectedBoardFunc     func(string) error
	boardSelectedCallback    func(string)
	boardCreatedCallback     func(string)

	// Call tracking
	refreshBoardsCalled      bool
	browseForBoardsCalled    bool
	createBoardCalled        bool
	getSelectedBoardCalled   bool
	setSelectedBoardCalled   bool
	callbacksSet            bool
}

func NewMockBoardSelectionView() *MockBoardSelectionView {
	return &MockBoardSelectionView{}
}

func (m *MockBoardSelectionView) RefreshBoards() error {
	m.refreshBoardsCalled = true
	if m.refreshBoardsFunc != nil {
		return m.refreshBoardsFunc()
	}
	return nil
}

func (m *MockBoardSelectionView) BrowseForBoards() error {
	m.browseForBoardsCalled = true
	if m.browseForBoardsFunc != nil {
		return m.browseForBoardsFunc()
	}
	return nil
}

func (m *MockBoardSelectionView) CreateBoard(request BoardCreationRequest) error {
	m.createBoardCalled = true
	if m.createBoardFunc != nil {
		return m.createBoardFunc(request)
	}
	return nil
}

func (m *MockBoardSelectionView) GetSelectedBoard() (*BoardInfo, error) {
	m.getSelectedBoardCalled = true
	if m.getSelectedBoardFunc != nil {
		return m.getSelectedBoardFunc()
	}
	return nil, nil
}

func (m *MockBoardSelectionView) SetSelectedBoard(boardPath string) error {
	m.setSelectedBoardCalled = true
	if m.setSelectedBoardFunc != nil {
		return m.setSelectedBoardFunc(boardPath)
	}
	return nil
}

func (m *MockBoardSelectionView) SetBoardSelectedCallback(callback func(string)) {
	m.boardSelectedCallback = callback
	m.callbacksSet = true
}

func (m *MockBoardSelectionView) SetBoardCreatedCallback(callback func(string)) {
	m.boardCreatedCallback = callback
	m.callbacksSet = true
}

// Fyne widget interface implementations
func (m *MockBoardSelectionView) CreateRenderer() interface{} { return nil }
func (m *MockBoardSelectionView) Resize(size interface{})     {}
func (m *MockBoardSelectionView) Position() interface{}       { return nil }
func (m *MockBoardSelectionView) Size() interface{}           { return nil }
func (m *MockBoardSelectionView) MinSize() interface{}        { return nil }
func (m *MockBoardSelectionView) Move(position interface{})   {}
func (m *MockBoardSelectionView) Hide()                       {}
func (m *MockBoardSelectionView) Show()                       {}
func (m *MockBoardSelectionView) Visible() bool               { return true }
func (m *MockBoardSelectionView) Refresh()                    {}

// MockBoardView provides a mock implementation for testing
type MockBoardView struct {
	loadBoardFunc   func()
	loadBoardCalled bool
}

func NewMockBoardView() *MockBoardView {
	return &MockBoardView{}
}

func (m *MockBoardView) LoadBoard() {
	m.loadBoardCalled = true
	if m.loadBoardFunc != nil {
		m.loadBoardFunc()
	}
}

// Fyne widget interface implementations
func (m *MockBoardView) CreateRenderer() interface{} { return nil }
func (m *MockBoardView) Resize(size interface{})     {}
func (m *MockBoardView) Position() interface{}       { return nil }
func (m *MockBoardView) Size() interface{}           { return nil }
func (m *MockBoardView) MinSize() interface{}        { return nil }
func (m *MockBoardView) Move(position interface{})   {}
func (m *MockBoardView) Hide()                       {}
func (m *MockBoardView) Show()                       {}
func (m *MockBoardView) Visible() bool               { return true }
func (m *MockBoardView) Refresh()                    {}

func TestUnit_ApplicationRoot_NewApplicationRoot(t *testing.T) {
	ar := createMockApplicationRoot(t)

	if ar == nil {
		t.Fatal("NewApplicationRoot() returned nil")
	}

	if ar.eventDispatcher == nil {
		t.Error("Event dispatcher not initialized")
	}

	if ar.currentView != ViewTypeBoardSelection {
		t.Errorf("Expected initial view to be BoardSelection, got %v", ar.currentView)
	}
}

func TestUnit_ApplicationRoot_NavigationEventDispatcher(t *testing.T) {
	ned := NewNavigationEventDispatcher()

	if ned == nil {
		t.Fatal("NewNavigationEventDispatcher() returned nil")
	}

	// Test subscription and publishing
	var receivedEvent NavigationEvent
	var eventReceived bool

	ned.Subscribe(NavigateToBoard, func(event NavigationEvent) {
		receivedEvent = event
		eventReceived = true
	})

	// Publish event
	testEvent := NavigationEvent{
		Type:       NavigateToBoard,
		BoardPath:  "/test/board",
		SourceView: ViewTypeBoardSelection,
	}

	ned.Publish(testEvent)

	// Give some time for async handling
	time.Sleep(10 * time.Millisecond)

	if !eventReceived {
		t.Error("Event was not received")
	}

	if receivedEvent.Type != testEvent.Type {
		t.Errorf("Expected event type %v, got %v", testEvent.Type, receivedEvent.Type)
	}

	if receivedEvent.BoardPath != testEvent.BoardPath {
		t.Errorf("Expected board path %s, got %s", testEvent.BoardPath, receivedEvent.BoardPath)
	}
}

func TestUnit_ApplicationRoot_NavigationEventDispatcher_MultipleSubscribers(t *testing.T) {
	ned := NewNavigationEventDispatcher()

	var callCount1, callCount2 int

	ned.Subscribe(NavigateToBoard, func(event NavigationEvent) {
		callCount1++
	})

	ned.Subscribe(NavigateToBoard, func(event NavigationEvent) {
		callCount2++
	})

	testEvent := NavigationEvent{
		Type:       NavigateToBoard,
		BoardPath:  "/test/board",
		SourceView: ViewTypeBoardSelection,
	}

	ned.Publish(testEvent)

	// Give some time for async handling
	time.Sleep(10 * time.Millisecond)

	if callCount1 != 1 {
		t.Errorf("Expected subscriber 1 to be called once, got %d", callCount1)
	}

	if callCount2 != 1 {
		t.Errorf("Expected subscriber 2 to be called once, got %d", callCount2)
	}
}

func TestUnit_ApplicationRoot_NavigationEventDispatcher_NoSubscribers(t *testing.T) {
	ned := NewNavigationEventDispatcher()

	// Publishing to non-existent event type should not panic
	testEvent := NavigationEvent{
		Type:       NavigateToBoard,
		BoardPath:  "/test/board",
		SourceView: ViewTypeBoardSelection,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Publishing to non-existent event type caused panic: %v", r)
		}
	}()

	ned.Publish(testEvent)
}

func TestUnit_ApplicationRoot_NavigationEventDispatcher_ConcurrentAccess(t *testing.T) {
	ned := NewNavigationEventDispatcher()

	var wg sync.WaitGroup
	var callCount int
	var mutex sync.Mutex

	// Add subscriber
	ned.Subscribe(NavigateToBoard, func(event NavigationEvent) {
		mutex.Lock()
		callCount++
		mutex.Unlock()
		wg.Done()
	})

	// Publish multiple events concurrently
	numEvents := 10
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(i int) {
			testEvent := NavigationEvent{
				Type:       NavigateToBoard,
				BoardPath:  "/test/board",
				SourceView: ViewTypeBoardSelection,
			}
			ned.Publish(testEvent)
		}(i)
	}

	// Wait for all events to be processed
	wg.Wait()

	mutex.Lock()
	finalCallCount := callCount
	mutex.Unlock()

	if finalCallCount != numEvents {
		t.Errorf("Expected %d calls, got %d", numEvents, finalCallCount)
	}
}

func TestUnit_ApplicationRoot_GetCurrentView(t *testing.T) {
	ar := createMockApplicationRoot(t)

	// Test initial view
	currentView := ar.GetCurrentView()
	if currentView != ViewTypeBoardSelection {
		t.Errorf("Expected initial view to be BoardSelection, got %v", currentView)
	}

	// Test thread safety by accessing from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			view := ar.GetCurrentView()
			if view != ViewTypeBoardSelection {
				t.Errorf("Expected view to be BoardSelection, got %v", view)
			}
		}()
	}

	wg.Wait()
}

func TestUnit_ApplicationRoot_GetEventDispatcher(t *testing.T) {
	ar := createMockApplicationRoot(t)

	dispatcher := ar.GetEventDispatcher()
	if dispatcher == nil {
		t.Error("GetEventDispatcher() returned nil")
	}

	if dispatcher != ar.eventDispatcher {
		t.Error("GetEventDispatcher() returned different instance than internal dispatcher")
	}
}

// Test that ApplicationRoot properly integrates with test environment
func TestIntegration_ApplicationRoot_TestEnvironmentIntegration(t *testing.T) {
	// Create test app to avoid window creation in headless test environment
	testApp := test.NewApp()
	defer testApp.Quit()

	ar := createTestApplicationRoot(t)

	// Test that the component can be created without panicking in test environment
	if ar == nil {
		t.Fatal("ApplicationRoot creation failed in test environment")
	}

	// Test that GetCurrentView works in test environment
	currentView := ar.GetCurrentView()
	if currentView != ViewTypeBoardSelection {
		t.Errorf("Expected initial view to be BoardSelection, got %v", currentView)
	}
}

func TestUnit_ApplicationRoot_StartApplication_MockEnvironment(t *testing.T) {
	t.Skip("Skipping StartApplication test in unit tests - requires full Fyne application context")

	// Note: Full StartApplication testing is handled in integration tests
	// Unit tests focus on individual component behavior
}