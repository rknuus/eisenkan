// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	clientEngines "github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/client/managers"
	"github.com/rknuus/eisenkan/internal/managers/task_manager"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/resource_access"
	"github.com/rknuus/eisenkan/internal/engines"
	"github.com/rknuus/eisenkan/internal/utilities"
	clientResourceAccess "github.com/rknuus/eisenkan/internal/client/resource_access"
)

// ViewType represents the different view types in the application
type ViewType int

const (
	ViewTypeBoardSelection ViewType = iota
	ViewTypeBoardView
)

// NavigationType represents the types of navigation events
type NavigationType int

const (
	NavigateToBoard NavigationType = iota
	NavigateBackToBoardSelection
	ApplicationExit
)

// NavigationEvent represents a navigation event in the application
type NavigationEvent struct {
	Type      NavigationType
	BoardPath string
	SourceView ViewType
}

// NavigationEventDispatcher handles navigation event subscription and publishing
type NavigationEventDispatcher struct {
	subscribers map[NavigationType][]func(NavigationEvent)
	mutex       sync.RWMutex
}

// NewNavigationEventDispatcher creates a new navigation event dispatcher
func NewNavigationEventDispatcher() *NavigationEventDispatcher {
	return &NavigationEventDispatcher{
		subscribers: make(map[NavigationType][]func(NavigationEvent)),
	}
}

// Subscribe adds a subscriber for a specific navigation event type
func (ned *NavigationEventDispatcher) Subscribe(eventType NavigationType, handler func(NavigationEvent)) {
	ned.mutex.Lock()
	defer ned.mutex.Unlock()

	ned.subscribers[eventType] = append(ned.subscribers[eventType], handler)
}

// Publish sends a navigation event to all subscribers
func (ned *NavigationEventDispatcher) Publish(event NavigationEvent) {
	ned.mutex.RLock()
	handlers, exists := ned.subscribers[event.Type]
	ned.mutex.RUnlock()

	if !exists {
		return
	}

	// Call handlers without holding the lock to avoid deadlocks
	for _, handler := range handlers {
		handler(event)
	}
}

// ApplicationRoot provides the main application controller and view management
type ApplicationRoot struct {
	// Fyne Application Management
	app    fyne.App
	window fyne.Window

	// View Management
	currentView        ViewType
	boardSelectionView BoardSelectionView
	boardView          *BoardView

	// Event-Driven Navigation
	eventDispatcher *NavigationEventDispatcher

	// Dependencies
	taskManager      task_manager.TaskManager
	workflowManager  managers.WorkflowManager
	formattingEngine *clientEngines.FormattingEngine
	layoutEngine     *clientEngines.LayoutEngine
	validationEngine *clientEngines.FormValidationEngine

	// Thread Safety
	mutex sync.RWMutex
}

// NewApplicationRoot creates a new ApplicationRoot instance
func NewApplicationRoot() *ApplicationRoot {
	return NewApplicationRootWithBoardDir("")
}

// NewApplicationRootWithBoardDir creates a new ApplicationRoot instance with custom board directory
func NewApplicationRootWithBoardDir(boardDir string) *ApplicationRoot {
	ar := &ApplicationRoot{
		eventDispatcher: NewNavigationEventDispatcher(),
		currentView:     ViewTypeBoardSelection, // Start with board selection
	}

	// Initialize dependencies
	if err := ar.initializeDependencies(boardDir); err != nil {
		// Return the ApplicationRoot even if dependency initialization fails
		// The error will be shown when StartApplication is called
		fmt.Printf("Warning: Failed to initialize dependencies: %v\n", err)
	}

	return ar
}

// initializeDependencies creates all the required dependencies for the application
func (ar *ApplicationRoot) initializeDependencies(customBoardDir string) error {
	// Initialize client engines
	ar.formattingEngine = clientEngines.NewFormattingEngine()
	ar.layoutEngine = clientEngines.NewLayoutEngine()
	ar.validationEngine = clientEngines.NewFormValidationEngine()

	// Determine board directory
	var boardDir string
	if customBoardDir != "" {
		boardDir = customBoardDir
	} else {
		// Create a default board directory for production use
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		boardDir = filepath.Join(homeDir, "EisenKan", "default-board")
	}

	// Ensure the board directory exists
	if err := os.MkdirAll(boardDir, 0755); err != nil {
		return fmt.Errorf("failed to create board directory: %w", err)
	}

	// Create default board configuration if it doesn't exist
	configPath := filepath.Join(boardDir, "board.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := `{
  "name": "Default Board",
  "columns": ["todo", "doing", "done"],
  "sections": {
    "todo": ["urgent-important", "urgent-not-important", "not-urgent-important"]
  },
  "git_user": "EisenKan User",
  "git_email": "user@eisenkan.local"
}`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to create default board config: %w", err)
		}
	}

	// Initialize utilities
	loggingUtility := utilities.NewLoggingUtility()
	cacheUtility := utilities.NewCacheUtility()

	// Initialize git repository for board directory
	gitConfig := &utilities.AuthorConfiguration{
		User:  "EisenKan User",
		Email: "user@eisenkan.local",
	}
	repository, err := utilities.InitializeRepositoryWithConfig(boardDir, gitConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Initialize BoardAccess
	boardAccess, err := board_access.NewBoardAccess(boardDir)
	if err != nil {
		return fmt.Errorf("failed to create BoardAccess: %w", err)
	}

	// Initialize RulesAccess
	rulesAccess, err := resource_access.NewRulesAccess(boardDir)
	if err != nil {
		return fmt.Errorf("failed to create RulesAccess: %w", err)
	}

	// Initialize RuleEngine
	ruleEngine, err := engines.NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		return fmt.Errorf("failed to create RuleEngine: %w", err)
	}

	// Initialize TaskManager
	ar.taskManager = task_manager.NewTaskManager(boardAccess, ruleEngine, loggingUtility, repository, boardDir)

	// Initialize DragDropEngine
	dragDropEngine := clientEngines.NewDragDropEngine()

	// Initialize TaskManagerAccess for WorkflowManager
	taskManagerAccess := clientResourceAccess.NewTaskManagerAccess(ar.taskManager, cacheUtility, loggingUtility)

	// Initialize WorkflowManager
	ar.workflowManager = managers.NewWorkflowManager(
		ar.validationEngine,
		ar.formattingEngine,
		dragDropEngine,
		taskManagerAccess,
	)

	return nil
}


// StartApplication initializes the Fyne application and shows the main window
func (ar *ApplicationRoot) StartApplication() error {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	// Initialize Fyne application
	ar.app = app.New()
	if ar.app == nil {
		return fmt.Errorf("failed to create Fyne application")
	}

	// Create main window
	ar.window = ar.app.NewWindow("EisenKan")
	if ar.window == nil {
		return fmt.Errorf("failed to create main window")
	}

	// Configure window
	ar.window.Resize(fyne.NewSize(1024, 768))
	ar.window.CenterOnScreen()

	// Set up navigation event handlers
	ar.setupNavigationHandlers()

	// Set up window close handler
	ar.window.SetCloseIntercept(func() {
		// Clean up and quit directly
		if ar.app != nil {
			ar.app.Quit()
		}
	})

	// Show initial view (BoardSelectionView)
	if err := ar.showBoardSelectionView(); err != nil {
		return fmt.Errorf("failed to show board selection view: %w", err)
	}

	// Show window and run
	ar.window.ShowAndRun()
	return nil
}

// setupNavigationHandlers configures the navigation event handlers
func (ar *ApplicationRoot) setupNavigationHandlers() {
	// Handle navigation to board
	ar.eventDispatcher.Subscribe(NavigateToBoard, func(event NavigationEvent) {
		if err := ar.showBoardView(event.BoardPath); err != nil {
			ar.showErrorAndExit(fmt.Errorf("failed to navigate to board: %w", err))
		}
	})

	// Handle navigation back to board selection
	ar.eventDispatcher.Subscribe(NavigateBackToBoardSelection, func(event NavigationEvent) {
		if err := ar.showBoardSelectionView(); err != nil {
			ar.showErrorAndExit(fmt.Errorf("failed to navigate back to board selection: %w", err))
		}
	})

	// Handle application exit
	ar.eventDispatcher.Subscribe(ApplicationExit, func(event NavigationEvent) {
		ar.shutdownApplication()
	})
}

// showBoardSelectionView displays the board selection view
func (ar *ApplicationRoot) showBoardSelectionView() error {
	// Hide current view if necessary
	if ar.boardView != nil {
		// Note: BoardView cleanup would happen here if needed
		ar.boardView = nil
	}

	// Create or get board selection view
	if ar.boardSelectionView == nil {
		// Check that dependencies are available
		if ar.taskManager == nil || ar.formattingEngine == nil || ar.layoutEngine == nil {
			return fmt.Errorf("BoardSelectionView initialization not yet implemented")
		}

		// Create BoardSelectionView with dependencies
		ar.boardSelectionView = NewBoardSelectionView(
			ar.taskManager,
			ar.formattingEngine,
			ar.layoutEngine,
			ar.window,
		)
	}

	// Set up board selection callbacks to publish navigation events
	ar.boardSelectionView.SetBoardSelectedCallback(func(boardPath string) {
		ar.eventDispatcher.Publish(NavigationEvent{
			Type:       NavigateToBoard,
			BoardPath:  boardPath,
			SourceView: ViewTypeBoardSelection,
		})
	})

	ar.boardSelectionView.SetBoardCreatedCallback(func(boardPath string) {
		ar.eventDispatcher.Publish(NavigationEvent{
			Type:       NavigateToBoard,
			BoardPath:  boardPath,
			SourceView: ViewTypeBoardSelection,
		})
	})

	// Update window content (only if window is available)
	if ar.window != nil {
		ar.window.SetContent(ar.boardSelectionView)
		ar.window.SetTitle("EisenKan - Select Board")
	}
	ar.currentView = ViewTypeBoardSelection

	// Refresh boards list
	if err := ar.boardSelectionView.RefreshBoards(); err != nil {
		return fmt.Errorf("failed to refresh boards: %w", err)
	}

	return nil
}

// showBoardView displays the board view with the specified board
func (ar *ApplicationRoot) showBoardView(boardPath string) error {
	if boardPath == "" {
		return fmt.Errorf("board path cannot be empty")
	}

	// Hide current view
	if ar.boardSelectionView != nil {
		// BoardSelectionView cleanup would happen here if needed
	}

	// Create or initialize board view
	if ar.boardView == nil {
		// Check that dependencies are available
		if ar.workflowManager == nil || ar.validationEngine == nil {
			return fmt.Errorf("BoardView initialization not yet implemented")
		}

		// Create BoardView with dependencies (using default Eisenhower configuration)
		ar.boardView = NewBoardView(
			ar.workflowManager,
			ar.validationEngine,
			nil, // Use default configuration
		)
	}

	// Load the specified board
	// Note: BoardPath would typically be handled by setting board configuration
	// before calling LoadBoard. For now, we'll just load the board.
	ar.boardView.LoadBoard()

	// Set up board view navigation callback
	// TODO: Implement navigation callback setup for BoardView
	// BoardView would need to provide a way to register navigation back callback

	// Update window content (only if window is available)
	if ar.window != nil {
		ar.window.SetContent(ar.boardView)
		ar.window.SetTitle(fmt.Sprintf("EisenKan - %s", boardPath))
	}
	ar.currentView = ViewTypeBoardView

	return nil
}

// showErrorAndExit displays an error dialog and exits the application
func (ar *ApplicationRoot) showErrorAndExit(err error) {
	if ar.window == nil {
		// Fallback if window is not available
		fmt.Printf("Application Error: %v\n", err)
		return
	}

	// Create error dialog
	errorDialog := dialog.NewError(err, ar.window)
	errorDialog.SetOnClosed(func() {
		ar.shutdownApplication()
	})

	// Show error dialog
	errorDialog.Show()
}

// shutdownApplication performs clean shutdown of the application
func (ar *ApplicationRoot) shutdownApplication() {
	// Simple and direct shutdown
	if ar.app != nil {
		ar.app.Quit()
	}
}

// GetCurrentView returns the currently displayed view type
func (ar *ApplicationRoot) GetCurrentView() ViewType {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	return ar.currentView
}

// GetEventDispatcher returns the navigation event dispatcher for testing
func (ar *ApplicationRoot) GetEventDispatcher() *NavigationEventDispatcher {
	return ar.eventDispatcher
}

// showPlaceholderView displays a placeholder message when dependencies are not available
func (ar *ApplicationRoot) showPlaceholderView(message string) {
	// Create a simple placeholder UI
	titleLabel := widget.NewLabelWithStyle("EisenKan Application Root", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	// Create a button to exit the application
	exitButton := widget.NewButton("Exit Application", func() {
		// Quit the application directly - this is the safest approach
		if ar.app != nil {
			ar.app.Quit()
		}
	})

	// Create container with placeholder content
	content := container.NewVBox(
		container.NewPadded(titleLabel),
		container.NewPadded(messageLabel),
		container.NewPadded(exitButton),
	)

	// Update window content
	ar.window.SetContent(content)
	ar.window.SetTitle("EisenKan - Dependency Setup Required")
}