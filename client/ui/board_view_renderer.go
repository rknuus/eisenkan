// Package ui provides Client UI layer components for the EisenKan system following iDesign methodology.
// This package contains UI components that integrate with Manager and Engine layers.
// Following iDesign namespace: eisenkan.Client.UI
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BoardViewRenderer implements fyne.WidgetRenderer for BoardView
type BoardViewRenderer struct {
	widget       *BoardView
	container    *fyne.Container
	loadingLabel *widget.Label
	errorLabel   *widget.Label
	titleLabel   *widget.Label
	background   *canvas.Rectangle
	objects      []fyne.CanvasObject
}

// newBoardViewRenderer creates a new renderer for BoardView
func newBoardViewRenderer(board *BoardView) *BoardViewRenderer {
	r := &BoardViewRenderer{
		widget: board,
	}

	// Create background
	r.background = canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

	// Create title label
	r.titleLabel = widget.NewLabel("")
	r.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	r.titleLabel.Alignment = fyne.TextAlignCenter

	// Create loading indicator
	r.loadingLabel = widget.NewLabel("Loading...")
	r.loadingLabel.Alignment = fyne.TextAlignCenter
	r.loadingLabel.Hide()

	// Create error label
	r.errorLabel = widget.NewLabel("")
	r.errorLabel.Wrapping = fyne.TextWrapWord
	r.errorLabel.Alignment = fyne.TextAlignCenter
	r.errorLabel.Hide()

	// Create main container with dynamic layout
	r.container = container.NewVBox()

	r.objects = []fyne.CanvasObject{
		r.background,
		r.container,
	}

	r.refreshLayout()

	return r
}

// Layout arranges the child objects within the specified size
func (r *BoardViewRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	r.container.Resize(size)
	r.container.Move(fyne.NewPos(0, 0))
}

// MinSize returns the minimum size required for the widget
func (r *BoardViewRenderer) MinSize() fyne.Size {
	state := r.widget.GetBoardState()

	// Base minimum size
	minWidth := float32(200)
	minHeight := float32(300)

	// Calculate minimum width based on columns
	if len(state.Columns) > 0 {
		columnMinWidth := float32(250) // Minimum width per column
		padding := float32(10)         // Padding between columns
		totalMinWidth := float32(len(state.Columns))*columnMinWidth + float32(len(state.Columns)-1)*padding

		if totalMinWidth > minWidth {
			minWidth = totalMinWidth
		}
	}

	return fyne.NewSize(minWidth, minHeight)
}

// Refresh updates the visual representation
func (r *BoardViewRenderer) Refresh() {
	r.refreshLayout()
	r.updateColors()
	canvas.Refresh(r.widget)
}

// Objects returns all the objects that should be rendered
func (r *BoardViewRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy cleans up renderer resources
func (r *BoardViewRenderer) Destroy() {
	// No specific cleanup needed for this renderer
}

// refreshLayout rebuilds the layout based on current board state
func (r *BoardViewRenderer) refreshLayout() {
	state := r.widget.GetBoardState()

	// Update title
	if state.Configuration != nil {
		r.titleLabel.SetText(state.Configuration.Title)
	}

	// Clear container
	r.container.Objects = nil

	// Add title
	r.container.Add(r.titleLabel)

	// Handle different states
	switch {
	case state.HasError:
		r.showError(state.ErrorMessage)
	case state.IsLoading:
		r.showLoading()
	default:
		r.showBoard(state)
	}
}

// showError displays error state
func (r *BoardViewRenderer) showError(errorMessage string) {
	r.loadingLabel.Hide()
	r.errorLabel.SetText("Error: " + errorMessage)
	r.errorLabel.Show()
	r.container.Add(r.errorLabel)
}

// showLoading displays loading state
func (r *BoardViewRenderer) showLoading() {
	r.errorLabel.Hide()
	r.loadingLabel.Show()
	r.container.Add(r.loadingLabel)
}

// showBoard displays the main board with columns
func (r *BoardViewRenderer) showBoard(state *BoardState) {
	r.loadingLabel.Hide()
	r.errorLabel.Hide()

	if len(state.Columns) == 0 {
		emptyLabel := widget.NewLabel("No columns configured")
		emptyLabel.Alignment = fyne.TextAlignCenter
		r.container.Add(emptyLabel)
		return
	}

	// Create columns layout based on board type and configuration
	columnsContainer := r.createColumnsLayout(state)
	r.container.Add(columnsContainer)

	// Add refresh indicator if refreshing
	if state.IsRefreshing {
		refreshLabel := widget.NewLabel("Refreshing...")
		refreshLabel.Alignment = fyne.TextAlignCenter
		r.container.Add(refreshLabel)
	}
}

// createColumnsLayout creates the layout container for columns
func (r *BoardViewRenderer) createColumnsLayout(state *BoardState) *fyne.Container {
	columnObjects := make([]fyne.CanvasObject, 0, len(state.Columns))

	for _, column := range state.Columns {
		columnObjects = append(columnObjects, column)
	}

	// Choose layout based on board type and number of columns
	return r.selectOptimalLayout(state, columnObjects)
}

// selectOptimalLayout selects the best layout for the given columns
func (r *BoardViewRenderer) selectOptimalLayout(state *BoardState, columnObjects []fyne.CanvasObject) *fyne.Container {
	numColumns := len(columnObjects)

	switch state.Configuration.BoardType {
	case "eisenhower":
		// Eisenhower Matrix: 2x2 grid
		if numColumns == 4 {
			return container.NewGridWithColumns(2, columnObjects...)
		}
		// Fallback to horizontal if not exactly 4 columns
		return container.NewHBox(columnObjects...)

	case "kanban":
		// Kanban: horizontal layout
		return container.NewHBox(columnObjects...)

	default:
		// Generic: choose layout based on number of columns
		if numColumns <= 2 {
			return container.NewHBox(columnObjects...)
		} else if numColumns <= 4 {
			return container.NewGridWithColumns(2, columnObjects...)
		} else if numColumns <= 6 {
			return container.NewGridWithColumns(3, columnObjects...)
		} else {
			// For many columns, use horizontal layout
			return container.NewHBox(columnObjects...)
		}
	}
}

// updateColors updates colors based on current state
func (r *BoardViewRenderer) updateColors() {
	state := r.widget.GetBoardState()

	// Update background color based on state
	switch {
	case state.HasError:
		r.background.FillColor = color.NRGBA{R: 255, G: 240, B: 240, A: 255} // Light red
		r.errorLabel.Importance = widget.DangerImportance
	case state.IsLoading:
		r.background.FillColor = color.NRGBA{R: 240, G: 240, B: 255, A: 255} // Light blue
	case state.IsRefreshing:
		r.background.FillColor = color.NRGBA{R: 255, G: 255, B: 240, A: 255} // Light yellow
	default:
		r.background.FillColor = theme.Color(theme.ColorNameBackground)
	}

	r.background.Refresh()
}

// onFormFieldChanged handles form field changes (for future form integration)
func (r *BoardViewRenderer) onFormFieldChanged(value string) {
	// Placeholder for future form integration
}

// onConfigurationChanged handles board configuration changes
func (r *BoardViewRenderer) onConfigurationChanged() {
	r.refreshLayout()
}

// Drag-Drop Integration Methods (Board-Level Coordination)

// onDragStart handles the start of a drag operation
func (r *BoardViewRenderer) onDragStart(event *fyne.DragEvent) {
	// Board-level drag coordination will be implemented here
	// This will coordinate with WorkflowManager for cross-column drags
}

// onDragMove handles drag movement across the board
func (r *BoardViewRenderer) onDragMove(event *fyne.DragEvent) {
	// Board-level drag movement coordination
	// Determine which column the drag is over and provide visual feedback
}

// onDragEnd handles the end of a drag operation
func (r *BoardViewRenderer) onDragEnd(event *fyne.DragEvent) {
	// Board-level drag completion coordination
	// Execute WorkflowManager.Drag() for cross-column task movement
}

// Helper Methods

// calculateColumnBounds calculates bounds for a specific column
func (r *BoardViewRenderer) calculateColumnBounds(columnIndex int) (fyne.Position, fyne.Size) {
	// This will be used for drag-drop zone calculations
	// Implementation depends on the layout strategy
	return fyne.NewPos(0, 0), fyne.NewSize(0, 0)
}

// findColumnAtPosition determines which column is at a given position
func (r *BoardViewRenderer) findColumnAtPosition(pos fyne.Position) int {
	// This will be used for drag-drop target detection
	// Implementation depends on the layout strategy
	return -1
}

// createDropIndicator creates visual feedback for drop operations
func (r *BoardViewRenderer) createDropIndicator(pos fyne.Position, size fyne.Size) *canvas.Rectangle {
	indicator := canvas.NewRectangle(color.NRGBA{R: 100, G: 200, B: 100, A: 128})
	indicator.Move(pos)
	indicator.Resize(size)
	return indicator
}

// removeDropIndicator removes visual feedback for drop operations
func (r *BoardViewRenderer) removeDropIndicator() {
	// Remove any active drop indicators
	r.refreshLayout()
}