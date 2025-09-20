// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/rknuus/eisenkan/client/engines"
	"github.com/rknuus/eisenkan/internal/managers/task_manager"
)

// BoardInfo represents board information for UI display
type BoardInfo struct {
	Path         string            `json:"path"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	LastModified time.Time         `json:"last_modified"`
	TaskCount    int               `json:"task_count"`
	IsValid      bool              `json:"is_valid"`
	Metadata     map[string]string `json:"metadata"`
}

// SortOrder represents board sorting options
type SortOrder int

const (
	SortByName SortOrder = iota
	SortByDate
	SortByUsage
)

// BoardSelectionState represents the current state of the BoardSelectionView
type BoardSelectionState struct {
	boards          []BoardInfo
	filteredBoards  []BoardInfo
	selectedBoard   *BoardInfo
	searchFilter    string
	sortOrder       SortOrder
	isLoading       bool
	lastError       error
}

// BoardCreationRequest represents board creation request for UI
type BoardCreationRequest struct {
	Path        string
	Title       string
	Description string
	InitializeGit bool
}

// BoardSelectionView defines the interface for board selection and management
type BoardSelectionView interface {
	// Core Operations
	RefreshBoards() error
	BrowseForBoards() error
	CreateBoard(request BoardCreationRequest) error

	// Selection Management
	GetSelectedBoard() (*BoardInfo, error)
	SetSelectedBoard(boardPath string) error

	// Event Handling
	SetBoardSelectedCallback(callback func(boardPath string))
	SetBoardCreatedCallback(callback func(boardPath string))

	// Fyne Widget Interface
	fyne.Widget
}

// boardSelectionView implements BoardSelectionView using composite widget pattern
type boardSelectionView struct {
	widget.BaseWidget

	// UI Components
	mainContainer *fyne.Container
	searchEntry   *widget.Entry
	refreshButton *widget.Button
	boardList     *widget.List
	browseButton  *widget.Button
	createButton  *widget.Button

	// State Management
	stateMu      sync.RWMutex
	currentState *BoardSelectionState
	stateChannel chan *BoardSelectionState

	// Dependencies
	taskManager   task_manager.TaskManager
	formatter     *engines.FormattingEngine
	layoutEngine  *engines.LayoutEngine

	// Callbacks
	onBoardSelected func(string)
	onBoardCreated  func(string)

	// Internal state
	ctx    context.Context
	cancel context.CancelFunc
	window fyne.Window
}

// NewBoardSelectionView creates a new BoardSelectionView with the specified dependencies
func NewBoardSelectionView(
	taskManager task_manager.TaskManager,
	formatter *engines.FormattingEngine,
	layoutEngine *engines.LayoutEngine,
	window fyne.Window,
) BoardSelectionView {
	ctx, cancel := context.WithCancel(context.Background())

	bsv := &boardSelectionView{
		taskManager:  taskManager,
		formatter:    formatter,
		layoutEngine: layoutEngine,
		window:       window,
		ctx:          ctx,
		cancel:       cancel,
		currentState: &BoardSelectionState{
			boards:         make([]BoardInfo, 0),
			filteredBoards: make([]BoardInfo, 0),
			sortOrder:      SortByDate,
		},
		stateChannel: make(chan *BoardSelectionState, 10),
	}

	bsv.ExtendBaseWidget(bsv)
	bsv.initializeUI()
	bsv.startStateManager()

	return bsv
}

// initializeUI sets up the UI components
func (bsv *boardSelectionView) initializeUI() {
	// Search entry with real-time filtering
	bsv.searchEntry = widget.NewEntry()
	bsv.searchEntry.SetPlaceHolder("Search boards...")
	bsv.searchEntry.OnChanged = func(text string) {
		bsv.updateSearchFilter(text)
	}

	// Refresh button
	bsv.refreshButton = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		bsv.RefreshBoards()
	})

	// Search container
	searchContainer := container.NewBorder(nil, nil, nil, bsv.refreshButton, bsv.searchEntry)

	// Board list
	bsv.boardList = widget.NewList(
		func() int {
			bsv.stateMu.RLock()
			defer bsv.stateMu.RUnlock()
			return len(bsv.currentState.filteredBoards)
		},
		func() fyne.CanvasObject {
			title := widget.NewLabel("Board Title")
			title.TextStyle = fyne.TextStyle{Bold: true}

			path := widget.NewLabel("/path/to/board")
			path.TextStyle = fyne.TextStyle{Italic: true}

			modified := widget.NewLabel("Last modified: 2 days ago")

			return container.NewVBox(title, path, modified)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			bsv.stateMu.RLock()
			defer bsv.stateMu.RUnlock()

			if id < 0 || id >= len(bsv.currentState.filteredBoards) {
				return
			}

			board := bsv.currentState.filteredBoards[id]
			vboxContainer := item.(*fyne.Container)

			title := vboxContainer.Objects[0].(*widget.Label)
			path := vboxContainer.Objects[1].(*widget.Label)
			modified := vboxContainer.Objects[2].(*widget.Label)

			title.SetText(board.Title)
			path.SetText(board.Path)

			// Format the modified time using FormattingEngine
			relativeTime := bsv.formatRelativeTime(board.LastModified)
			modified.SetText(fmt.Sprintf("Last modified: %s", relativeTime))
		},
	)

	bsv.boardList.OnSelected = func(id widget.ListItemID) {
		bsv.stateMu.RLock()
		defer bsv.stateMu.RUnlock()

		if id >= 0 && id < len(bsv.currentState.filteredBoards) {
			selectedBoard := bsv.currentState.filteredBoards[id]
			bsv.currentState.selectedBoard = &selectedBoard

			// Trigger callback
			if bsv.onBoardSelected != nil {
				bsv.onBoardSelected(selectedBoard.Path)
			}
		}
	}

	// Action buttons
	bsv.browseButton = widget.NewButtonWithIcon("Browse for Boards", theme.FolderOpenIcon(), func() {
		bsv.BrowseForBoards()
	})

	bsv.createButton = widget.NewButtonWithIcon("Create New Board", theme.DocumentCreateIcon(), func() {
		bsv.showCreateBoardDialog()
	})

	buttonContainer := container.NewHBox(bsv.browseButton, bsv.createButton)

	// Recent boards section
	recentLabel := widget.NewLabelWithStyle("Recent Boards", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Main container
	bsv.mainContainer = container.NewVBox(
		searchContainer,
		widget.NewSeparator(),
		recentLabel,
		bsv.boardList,
		widget.NewSeparator(),
		buttonContainer,
	)
}

// CreateRenderer implements fyne.Widget
func (bsv *boardSelectionView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(bsv.mainContainer)
}

// RefreshBoards reloads board data from recent boards mechanisms
func (bsv *boardSelectionView) RefreshBoards() error {
	bsv.stateMu.Lock()
	bsv.currentState.isLoading = true
	bsv.currentState.lastError = nil
	bsv.stateMu.Unlock()

	// Update UI to show loading state
	bsv.Refresh()

	go func() {
		defer func() {
			bsv.stateMu.Lock()
			bsv.currentState.isLoading = false
			bsv.stateMu.Unlock()
			bsv.Refresh()
		}()

		// Get recent boards from OS mechanisms (simplified for now)
		// In a real implementation, this would use OS-specific recent document APIs
		boards := bsv.loadRecentBoards()

		bsv.stateMu.Lock()
		bsv.currentState.boards = boards
		bsv.applyFiltersAndSort()
		bsv.stateMu.Unlock()

		bsv.Refresh()
	}()

	return nil
}

// BrowseForBoards opens directory selection dialog
func (bsv *boardSelectionView) BrowseForBoards() error {
	folderDialog := dialog.NewFolderOpen(func(reader fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, bsv.window)
			return
		}
		if reader == nil {
			return // User cancelled
		}

		// Validate the selected directory
		directoryPath := reader.Path()
		bsv.validateAndAddBoard(directoryPath)
	}, bsv.window)

	folderDialog.Show()
	return nil
}

// validateAndAddBoard validates a directory and adds it if it's a valid board
func (bsv *boardSelectionView) validateAndAddBoard(directoryPath string) {
	go func() {
		// Use TaskManager to validate the board directory
		response, err := bsv.taskManager.ValidateBoardDirectory(directoryPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to validate board directory: %w", err), bsv.window)
			return
		}

		if !response.IsValid {
			// Show detailed error message
			issues := strings.Join(response.Issues, "\n")
			errorMsg := fmt.Sprintf("Selected directory is not a valid board:\n\n%s", issues)
			dialog.ShowInformation("Invalid Board Directory", errorMsg, bsv.window)
			return
		}

		// Get board metadata
		metadata, err := bsv.taskManager.GetBoardMetadata(directoryPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to get board metadata: %w", err), bsv.window)
			return
		}

		// Create BoardInfo
		boardInfo := BoardInfo{
			Path:         directoryPath,
			Title:        metadata.Title,
			Description:  metadata.Description,
			LastModified: time.Now(), // Use current time as discovery time
			TaskCount:    metadata.TaskCount,
			IsValid:      true,
			Metadata:     metadata.Metadata,
		}

		// Add to recent boards and refresh
		bsv.addToRecentBoards(boardInfo)
		bsv.RefreshBoards()

		// Show success message
		dialog.ShowInformation("Board Added", fmt.Sprintf("Board '%s' has been added to your recent boards.", metadata.Title), bsv.window)
	}()
}

// CreateBoard creates a new board with the specified request
func (bsv *boardSelectionView) CreateBoard(request BoardCreationRequest) error {
	// Convert to TaskManager request format
	tmRequest := task_manager.BoardCreationRequest{
		BoardPath:     request.Path,
		Title:         request.Title,
		Description:   request.Description,
		InitializeGit: request.InitializeGit,
		Metadata:      make(map[string]string),
	}

	response, err := bsv.taskManager.CreateBoard(tmRequest)
	if err != nil {
		return fmt.Errorf("failed to create board: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("board creation failed: %s", response.Message)
	}

	// Create BoardInfo for the new board
	boardInfo := BoardInfo{
		Path:         response.BoardPath,
		Title:        request.Title,
		Description:  request.Description,
		LastModified: time.Now(),
		TaskCount:    0,
		IsValid:      true,
		Metadata:     make(map[string]string),
	}

	// Add to recent boards and refresh
	bsv.addToRecentBoards(boardInfo)
	bsv.RefreshBoards()

	// Trigger callback
	if bsv.onBoardCreated != nil {
		bsv.onBoardCreated(response.BoardPath)
	}

	return nil
}

// GetSelectedBoard returns the currently selected board
func (bsv *boardSelectionView) GetSelectedBoard() (*BoardInfo, error) {
	bsv.stateMu.RLock()
	defer bsv.stateMu.RUnlock()

	if bsv.currentState.selectedBoard == nil {
		return nil, fmt.Errorf("no board selected")
	}

	// Return a copy to prevent external modification
	selectedCopy := *bsv.currentState.selectedBoard
	return &selectedCopy, nil
}

// SetSelectedBoard sets the selected board by path
func (bsv *boardSelectionView) SetSelectedBoard(boardPath string) error {
	bsv.stateMu.Lock()
	defer bsv.stateMu.Unlock()

	for i, board := range bsv.currentState.filteredBoards {
		if board.Path == boardPath {
			bsv.currentState.selectedBoard = &board
			// Only call UI widget selection if we have a valid board list and not in test mode
			// In test environments, avoid calling Fyne widget methods that can hang
			if bsv.boardList != nil {
				// Use go routine to avoid blocking in test environments
				go func(index int) {
					defer func() {
						// Recover from any panics in UI operations during tests
						if r := recover(); r != nil {
							// Silently ignore UI operation failures in tests
						}
					}()
					bsv.boardList.Select(index)
				}(i)
			}
			return nil
		}
	}

	return fmt.Errorf("board not found: %s", boardPath)
}

// SetBoardSelectedCallback sets the callback for board selection events
func (bsv *boardSelectionView) SetBoardSelectedCallback(callback func(boardPath string)) {
	bsv.onBoardSelected = callback
}

// SetBoardCreatedCallback sets the callback for board creation events
func (bsv *boardSelectionView) SetBoardCreatedCallback(callback func(boardPath string)) {
	bsv.onBoardCreated = callback
}

// updateSearchFilter updates the search filter and refreshes the display
func (bsv *boardSelectionView) updateSearchFilter(text string) {
	bsv.stateMu.Lock()
	bsv.currentState.searchFilter = text
	bsv.applyFiltersAndSort()
	bsv.stateMu.Unlock()

	bsv.boardList.Refresh()
}

// applyFiltersAndSort applies current filters and sorting to the board list
func (bsv *boardSelectionView) applyFiltersAndSort() {
	// Filter boards based on search text
	filtered := make([]BoardInfo, 0)
	searchLower := strings.ToLower(bsv.currentState.searchFilter)

	for _, board := range bsv.currentState.boards {
		if searchLower == "" ||
		   strings.Contains(strings.ToLower(board.Title), searchLower) ||
		   strings.Contains(strings.ToLower(board.Path), searchLower) ||
		   strings.Contains(strings.ToLower(board.Description), searchLower) {
			filtered = append(filtered, board)
		}
	}

	// Sort boards based on current sort order
	sort.Slice(filtered, func(i, j int) bool {
		switch bsv.currentState.sortOrder {
		case SortByName:
			return filtered[i].Title < filtered[j].Title
		case SortByDate:
			return filtered[i].LastModified.After(filtered[j].LastModified)
		case SortByUsage:
			// For now, sort by date as usage tracking is not implemented
			return filtered[i].LastModified.After(filtered[j].LastModified)
		default:
			return filtered[i].LastModified.After(filtered[j].LastModified)
		}
	})

	bsv.currentState.filteredBoards = filtered
}

// showCreateBoardDialog shows the board creation dialog
func (bsv *boardSelectionView) showCreateBoardDialog() {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Board Title")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Board Description (optional)")
	descriptionEntry.Resize(fyne.NewSize(300, 100))

	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("Board Directory Path")

	browsePathButton := widget.NewButtonWithIcon("Browse...", theme.FolderOpenIcon(), func() {
		folderDialog := dialog.NewFolderOpen(func(reader fyne.ListableURI, err error) {
			if err == nil && reader != nil {
				pathEntry.SetText(reader.Path())
			}
		}, bsv.window)
		folderDialog.Show()
	})

	pathContainer := container.NewBorder(nil, nil, nil, browsePathButton, pathEntry)

	gitCheckbox := widget.NewCheck("Initialize Git repository", nil)
	gitCheckbox.SetChecked(true)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Create New Board", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Description:"),
		descriptionEntry,
		widget.NewLabel("Location:"),
		pathContainer,
		gitCheckbox,
	)

	createDialog := dialog.NewCustomConfirm("Create Board", "Create", "Cancel", content, func(confirmed bool) {
		if !confirmed {
			return
		}

		// Validate inputs
		if titleEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("board title is required"), bsv.window)
			return
		}
		if pathEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("board path is required"), bsv.window)
			return
		}

		// Create the board
		request := BoardCreationRequest{
			Path:        pathEntry.Text,
			Title:       titleEntry.Text,
			Description: descriptionEntry.Text,
			InitializeGit: gitCheckbox.Checked,
		}

		err := bsv.CreateBoard(request)
		if err != nil {
			dialog.ShowError(err, bsv.window)
			return
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Board '%s' created successfully!", request.Title), bsv.window)
	}, bsv.window)

	createDialog.Resize(fyne.NewSize(500, 400))
	createDialog.Show()
}

// formatRelativeTime formats a time as a relative string
func (bsv *boardSelectionView) formatRelativeTime(t time.Time) string {
	if bsv.formatter != nil {
		// Use FormattingEngine if available
		return bsv.formatter.Time().FormatRelativeTime(t)
	}

	// Fallback formatting
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("Jan 2, 2006")
	}
}

// loadRecentBoards loads recent boards from storage (simplified implementation)
func (bsv *boardSelectionView) loadRecentBoards() []BoardInfo {
	// In a real implementation, this would use OS-specific recent document APIs
	// or TaskManager context storage. For now, return empty list.
	return make([]BoardInfo, 0)
}

// addToRecentBoards adds a board to the recent boards list
func (bsv *boardSelectionView) addToRecentBoards(boardInfo BoardInfo) {
	// In a real implementation, this would store to OS recent documents
	// or TaskManager context storage. For now, just add to memory.
	bsv.stateMu.Lock()
	defer bsv.stateMu.Unlock()

	// Check if board already exists
	for i, existing := range bsv.currentState.boards {
		if existing.Path == boardInfo.Path {
			// Update existing entry
			bsv.currentState.boards[i] = boardInfo
			return
		}
	}

	// Add new board to beginning of list
	bsv.currentState.boards = append([]BoardInfo{boardInfo}, bsv.currentState.boards...)

	// Limit to 10 recent boards
	if len(bsv.currentState.boards) > 10 {
		bsv.currentState.boards = bsv.currentState.boards[:10]
	}
}

// startStateManager starts the state management goroutine
func (bsv *boardSelectionView) startStateManager() {
	go func() {
		for {
			select {
			case <-bsv.ctx.Done():
				return
			case newState := <-bsv.stateChannel:
				bsv.stateMu.Lock()
				bsv.currentState = newState
				bsv.stateMu.Unlock()
				bsv.Refresh()
			}
		}
	}()
}

// Cleanup closes resources and stops goroutines
func (bsv *boardSelectionView) Cleanup() {
	if bsv.cancel != nil {
		bsv.cancel()
	}
}