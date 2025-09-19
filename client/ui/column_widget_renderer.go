package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// columnWidgetRenderer implements fyne.WidgetRenderer for ColumnWidget
type columnWidgetRenderer struct {
	widget *ColumnWidget

	// Visual components
	background      *canvas.Rectangle
	border          *canvas.Rectangle
	headerContainer *fyne.Container
	titleLabel      *widget.Label
	taskCountLabel  *widget.Label
	addTaskButton   *widget.Button
	settingsButton  *widget.Button
	loadingIcon     *widget.Icon
	errorIcon       *widget.Icon
	wipWarningIcon  *widget.Icon

	// Task content area
	tasksContainer  *fyne.Container
	scrollContainer *container.Scroll

	// Section headers (for Todo column)
	sectionHeaders map[EisenhowerSection]*widget.Label

	// Main layout container
	mainContainer *fyne.Container

	// Component list for rendering
	objects []fyne.CanvasObject
}

// newColumnWidgetRenderer creates a new renderer for ColumnWidget
func newColumnWidgetRenderer(columnWidget *ColumnWidget) fyne.WidgetRenderer {
	renderer := &columnWidgetRenderer{
		widget:         columnWidget,
		sectionHeaders: make(map[EisenhowerSection]*widget.Label),
	}

	renderer.createComponents()
	renderer.setupLayout()
	renderer.updateComponents()

	// Store scroll container reference in widget for external access
	columnWidget.scrollContainer = renderer.scrollContainer

	return renderer
}

// createComponents creates all visual components
func (r *columnWidgetRenderer) createComponents() {
	// Background and border
	r.background = canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	r.border = canvas.NewRectangle(color.Transparent)
	r.border.StrokeWidth = 1
	r.border.StrokeColor = theme.Color(theme.ColorNameForeground)
	r.border.FillColor = color.Transparent

	// Header components
	r.titleLabel = widget.NewLabel("Column")
	r.titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	r.taskCountLabel = widget.NewLabel("0 tasks")
	r.taskCountLabel.TextStyle = fyne.TextStyle{Italic: true}

	r.addTaskButton = widget.NewButton("+", func() {
		r.handleAddTask()
	})
	r.addTaskButton.Resize(fyne.NewSize(30, 30))

	r.settingsButton = widget.NewButton("⚙", func() {
		r.handleSettings()
	})
	r.settingsButton.Resize(fyne.NewSize(30, 30))

	// Status icons
	r.loadingIcon = widget.NewIcon(theme.ViewRefreshIcon())
	r.errorIcon = widget.NewIcon(theme.ErrorIcon())
	r.wipWarningIcon = widget.NewIcon(theme.WarningIcon())

	// Initially hide icons
	r.loadingIcon.Hide()
	r.errorIcon.Hide()
	r.wipWarningIcon.Hide()

	// Task content area
	r.tasksContainer = container.NewVBox()
	r.scrollContainer = container.NewScroll(r.tasksContainer)
	r.scrollContainer.SetMinSize(fyne.NewSize(200, 300))

	// Create section headers for Todo column
	r.createSectionHeaders()
}

// createSectionHeaders creates Eisenhower Matrix section headers for Todo columns
func (r *columnWidgetRenderer) createSectionHeaders() {
	sections := []EisenhowerSection{
		UrgentImportant,
		UrgentNotImportant,
		NotUrgentImportant,
		NotUrgentNotImportant,
	}

	sectionTitles := map[EisenhowerSection]string{
		UrgentImportant:       "🔥 Urgent & Important",
		UrgentNotImportant:    "⚡ Urgent & Not Important",
		NotUrgentImportant:    "🎯 Not Urgent & Important",
		NotUrgentNotImportant: "📝 Not Urgent & Not Important",
	}

	for _, section := range sections {
		header := widget.NewLabel(sectionTitles[section])
		header.TextStyle = fyne.TextStyle{Bold: true}
		header.Hide() // Initially hidden
		r.sectionHeaders[section] = header
	}
}

// setupLayout creates the container layout structure
func (r *columnWidgetRenderer) setupLayout() {
	// Header container with title, count, and controls
	r.headerContainer = container.NewBorder(
		nil, nil,
		container.NewVBox(r.titleLabel, r.taskCountLabel),
		container.NewHBox(r.loadingIcon, r.errorIcon, r.wipWarningIcon, r.addTaskButton, r.settingsButton),
	)

	// Main container with header and scrollable task area
	r.mainContainer = container.NewBorder(
		r.headerContainer,
		nil,
		nil,
		nil,
		r.scrollContainer,
	)

	// Objects list for rendering (order matters for layering)
	r.objects = []fyne.CanvasObject{
		r.background,
		r.border,
		r.mainContainer,
	}
}

// updateComponents updates all visual components based on current widget state
func (r *columnWidgetRenderer) updateComponents() {
	state := r.widget.currentState
	if state == nil || state.Configuration == nil {
		r.updateEmptyState()
		return
	}

	// Update colors based on state
	r.updateColors()

	// Update header content
	r.updateHeader()

	// Update status icons
	r.updateIcons()

	// Update task content
	r.updateTaskContent()

	// Update section visibility
	r.updateSectionVisibility()
}

// updateColors updates background and border colors based on widget state
func (r *columnWidgetRenderer) updateColors() {
	bgColor, borderColor := r.widget.getStateColors()

	r.background.FillColor = bgColor
	r.border.StrokeColor = borderColor

	// Special handling for WIP limit warning
	if r.widget.currentState.WIPLimitReached {
		r.border.StrokeColor = theme.Color(theme.ColorNameWarning)
		r.border.StrokeWidth = 2
	} else {
		r.border.StrokeWidth = 1
	}
}

// updateHeader updates column title and task count
func (r *columnWidgetRenderer) updateHeader() {
	config := r.widget.currentState.Configuration
	taskCount := len(r.widget.currentState.Tasks)

	// Update title
	r.titleLabel.SetText(config.Title)

	// Update task count with WIP limit if applicable
	if config.WIPLimit > 0 {
		r.taskCountLabel.SetText(fmt.Sprintf("%d/%d tasks", taskCount, config.WIPLimit))
	} else {
		r.taskCountLabel.SetText(fmt.Sprintf("%d tasks", taskCount))
	}
}

// updateIcons shows/hides status icons based on current state
func (r *columnWidgetRenderer) updateIcons() {
	state := r.widget.currentState

	// Reset all icons
	r.loadingIcon.Hide()
	r.errorIcon.Hide()
	r.wipWarningIcon.Hide()

	// Show appropriate icons based on state
	if state.IsLoading {
		r.loadingIcon.Show()
		return
	}

	if state.HasError {
		r.errorIcon.Show()
	}

	if state.WIPLimitReached {
		r.wipWarningIcon.Show()
	}
}

// updateTaskContent updates the task container with current task widgets
func (r *columnWidgetRenderer) updateTaskContent() {
	// Clear existing content
	r.tasksContainer.RemoveAll()

	state := r.widget.currentState
	if state.Configuration.Type == TodoColumn && state.Configuration.ShowSections {
		r.updateTodoColumnWithSections()
	} else {
		r.updateSimpleColumn()
	}
}

// updateTodoColumnWithSections updates Todo column with Eisenhower Matrix sections
func (r *columnWidgetRenderer) updateTodoColumnWithSections() {
	// Group tasks by priority sections
	tasksBySection := r.groupTasksBySection()

	sections := []EisenhowerSection{
		UrgentImportant,
		UrgentNotImportant,
		NotUrgentImportant,
		NotUrgentNotImportant,
	}

	for _, section := range sections {
		// Add section header
		if header, exists := r.sectionHeaders[section]; exists {
			header.Show()
			r.tasksContainer.Add(header)
		}

		// Add tasks in this section
		tasks := tasksBySection[section]
		for _, task := range tasks {
			if widget, exists := r.widget.currentState.TaskWidgets[task.ID]; exists {
				r.tasksContainer.Add(widget)
			}
		}

		// Add separator between sections (except after last)
		if section != NotUrgentNotImportant && len(tasksBySection[section]) > 0 {
			r.tasksContainer.Add(widget.NewSeparator())
		}
	}
}

// updateSimpleColumn updates column with simple task list (Doing/Done columns)
func (r *columnWidgetRenderer) updateSimpleColumn() {
	// Add all tasks in order
	for _, task := range r.widget.currentState.Tasks {
		if widget, exists := r.widget.currentState.TaskWidgets[task.ID]; exists {
			r.tasksContainer.Add(widget)
		}
	}
}

// groupTasksBySection groups tasks by Eisenhower Matrix sections
func (r *columnWidgetRenderer) groupTasksBySection() map[EisenhowerSection][]*TaskData {
	result := make(map[EisenhowerSection][]*TaskData)

	for _, task := range r.widget.currentState.Tasks {
		section := r.getTaskSection(task)
		result[section] = append(result[section], task)
	}

	return result
}

// getTaskSection determines which Eisenhower section a task belongs to
func (r *columnWidgetRenderer) getTaskSection(task *TaskData) EisenhowerSection {
	// Map task priority to Eisenhower section
	switch task.Priority {
	case "urgent-important", "high":
		return UrgentImportant
	case "urgent-not-important", "medium":
		return UrgentNotImportant
	case "not-urgent-important", "low":
		return NotUrgentImportant
	default:
		return NotUrgentNotImportant
	}
}

// updateSectionVisibility shows/hides section headers based on column type
func (r *columnWidgetRenderer) updateSectionVisibility() {
	showSections := r.widget.currentState.Configuration.Type == TodoColumn &&
		r.widget.currentState.Configuration.ShowSections

	for _, header := range r.sectionHeaders {
		if showSections {
			header.Show()
		} else {
			header.Hide()
		}
	}
}

// updateEmptyState handles rendering when no configuration is available
func (r *columnWidgetRenderer) updateEmptyState() {
	r.titleLabel.SetText("No Configuration")
	r.taskCountLabel.SetText("0 tasks")

	r.background.FillColor = theme.Color(theme.ColorNameDisabled)
	r.border.StrokeColor = theme.Color(theme.ColorNameDisabled)

	// Hide all icons
	r.loadingIcon.Hide()
	r.errorIcon.Hide()
	r.wipWarningIcon.Hide()

	// Clear task content
	r.tasksContainer.RemoveAll()
}

// Event handlers

// handleAddTask handles add task button clicks
func (r *columnWidgetRenderer) handleAddTask() {
	// Simple task creation dialog - could be enhanced
	r.widget.CreateTask("New Task", "Task description")
}

// handleSettings handles settings button clicks
func (r *columnWidgetRenderer) handleSettings() {
	// Settings dialog would be implemented here
	// For now, just log
	fmt.Println("Column settings requested")
}

// Renderer interface implementation

// Layout positions all visual components within the given size
func (r *columnWidgetRenderer) Layout(size fyne.Size) {
	// Background fills entire widget
	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	// Border outline
	r.border.Resize(size)
	r.border.Move(fyne.NewPos(0, 0))

	// Main container with small padding
	padding := theme.Padding()
	contentSize := fyne.NewSize(size.Width-padding*2, size.Height-padding*2)
	r.mainContainer.Resize(contentSize)
	r.mainContainer.Move(fyne.NewPos(padding, padding))
}

// MinSize returns the minimum size needed for the widget
func (r *columnWidgetRenderer) MinSize() fyne.Size {
	headerHeight := r.headerContainer.MinSize().Height
	minTaskArea := float32(200) // Minimum task area height

	return fyne.NewSize(250, headerHeight+minTaskArea+theme.Padding()*2)
}

// Refresh updates the visual components when the widget state changes
func (r *columnWidgetRenderer) Refresh() {
	r.updateComponents()
	r.Layout(r.widget.Size())

	// Refresh all canvas objects
	for _, obj := range r.objects {
		obj.Refresh()
	}

	// Refresh task widgets
	r.widget.stateMu.RLock()
	for _, taskWidget := range r.widget.currentState.TaskWidgets {
		taskWidget.Refresh()
	}
	r.widget.stateMu.RUnlock()
}

// Objects returns the list of canvas objects that make up the widget
func (r *columnWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy cleans up renderer resources
func (r *columnWidgetRenderer) Destroy() {
	// No specific cleanup needed for basic components
	// Custom resources would be cleaned up here
}