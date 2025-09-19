package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// taskWidgetRenderer implements fyne.WidgetRenderer for TaskWidget
type taskWidgetRenderer struct {
	widget *TaskWidget

	// Visual components
	background   *canvas.Rectangle
	border       *canvas.Rectangle
	titleLabel   *widget.Label
	descLabel    *widget.Label
	metaLabel    *widget.Label
	statusIcon   *widget.Icon
	priorityIcon *widget.Icon
	loadingIcon  *widget.Icon
	errorIcon    *widget.Icon

	// Layout containers
	mainContainer    *fyne.Container
	contentContainer *fyne.Container
	headerContainer  *fyne.Container
	metaContainer    *fyne.Container

	// Component list for rendering
	objects []fyne.CanvasObject
}

// newTaskWidgetRenderer creates a new renderer for TaskWidget
func newTaskWidgetRenderer(widget *TaskWidget) fyne.WidgetRenderer {
	renderer := &taskWidgetRenderer{
		widget: widget,
	}

	renderer.createComponents()
	renderer.setupLayout()
	renderer.updateComponents()

	return renderer
}

// createComponents creates all visual components
func (r *taskWidgetRenderer) createComponents() {
	// Background and border
	r.background = canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	r.border = canvas.NewRectangle(color.Transparent)
	r.border.StrokeWidth = 1
	r.border.StrokeColor = theme.Color(theme.ColorNameForeground)
	r.border.FillColor = color.Transparent

	// Text labels
	r.titleLabel = widget.NewLabel("")
	r.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	r.titleLabel.Wrapping = fyne.TextWrapWord

	r.descLabel = widget.NewLabel("")
	r.descLabel.Wrapping = fyne.TextWrapWord
	r.descLabel.TextStyle = fyne.TextStyle{Italic: true}

	r.metaLabel = widget.NewLabel("")
	r.metaLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Icons
	r.statusIcon = widget.NewIcon(theme.InfoIcon())
	r.priorityIcon = widget.NewIcon(theme.WarningIcon())
	r.loadingIcon = widget.NewIcon(theme.ViewRefreshIcon())
	r.errorIcon = widget.NewIcon(theme.ErrorIcon())

	// Initially hide all icons
	r.statusIcon.Hide()
	r.priorityIcon.Hide()
	r.loadingIcon.Hide()
	r.errorIcon.Hide()
}

// setupLayout creates the container layout structure
func (r *taskWidgetRenderer) setupLayout() {
	// Header container with title and icons
	r.headerContainer = container.NewHBox()
	r.headerContainer.Add(r.titleLabel)
	r.headerContainer.Add(r.priorityIcon)
	r.headerContainer.Add(r.statusIcon)
	r.headerContainer.Add(r.loadingIcon)
	r.headerContainer.Add(r.errorIcon)

	// Metadata container
	r.metaContainer = container.NewHBox()
	r.metaContainer.Add(r.metaLabel)

	// Content container with all text elements
	r.contentContainer = container.NewVBox()
	r.contentContainer.Add(r.headerContainer)
	r.contentContainer.Add(r.descLabel)
	r.contentContainer.Add(r.metaContainer)

	// Main border container
	r.mainContainer = container.NewBorder(nil, nil, nil, nil, r.contentContainer)

	// Objects list for rendering (order matters for layering)
	r.objects = []fyne.CanvasObject{
		r.background,
		r.border,
		r.mainContainer,
	}
}

// updateComponents updates all visual components based on current widget state
func (r *taskWidgetRenderer) updateComponents() {
	state := r.widget.currentState
	if state == nil || state.Data == nil {
		r.updateEmptyState()
		return
	}

	// Update colors based on state
	r.updateColors()

	// Update text content
	r.updateTextContent()

	// Update icons based on state
	r.updateIcons()

	// Update layout based on compact mode
	r.updateLayout()
}

// updateColors updates background and border colors based on widget state
func (r *taskWidgetRenderer) updateColors() {
	bgColor, borderColor := r.widget.getStateColors()

	r.background.FillColor = bgColor
	r.border.StrokeColor = borderColor

	// Special handling for validation errors (red border)
	if len(r.widget.currentState.ValidationErrs) > 0 {
		r.border.StrokeColor = theme.Color(theme.ColorNameError)
		r.border.StrokeWidth = 2
	} else {
		r.border.StrokeWidth = 1
	}
}

// updateTextContent updates all text labels with formatted content
func (r *taskWidgetRenderer) updateTextContent() {
	title, description, metadata := r.widget.formatTaskDisplay()

	// Update title
	r.titleLabel.SetText(title)

	// Update description (hide if compact mode)
	if r.widget.compact {
		r.descLabel.Hide()
	} else {
		r.descLabel.SetText(description)
		r.descLabel.Show()
	}

	// Update metadata (hide if not showing metadata or compact mode)
	if r.widget.showMetadata && !r.widget.compact && metadata != "" {
		r.metaLabel.SetText(metadata)
		r.metaContainer.Show()
	} else {
		r.metaContainer.Hide()
	}

	// Add validation error tooltips if present
	if len(r.widget.currentState.ValidationErrs) > 0 {
		errorText := "Validation errors:"
		for field, err := range r.widget.currentState.ValidationErrs {
			errorText += "\n" + field + ": " + err
		}
		// Note: Fyne doesn't have built-in tooltips, this would need custom implementation
		// For now, we'll indicate errors through border color only
	}
}

// updateIcons shows/hides icons based on current state and task data
func (r *taskWidgetRenderer) updateIcons() {
	state := r.widget.currentState

	// Reset all icons
	r.statusIcon.Hide()
	r.priorityIcon.Hide()
	r.loadingIcon.Hide()
	r.errorIcon.Hide()

	// Show appropriate icons based on state
	if state.IsLoading {
		r.loadingIcon.Show()
		return
	}

	if state.HasError {
		r.errorIcon.Show()
		return
	}

	if state.Data != nil {
		// Show priority icon based on priority level
		switch state.Data.Priority {
		case "urgent-important", "high":
			r.priorityIcon.Resource = theme.ErrorIcon()
			r.priorityIcon.Show()
		case "urgent-not-important", "medium":
			r.priorityIcon.Resource = theme.WarningIcon()
			r.priorityIcon.Show()
		case "not-urgent-important", "low":
			r.priorityIcon.Resource = theme.InfoIcon()
			r.priorityIcon.Show()
		default:
			// No priority icon for not-urgent-not-important or unknown
		}

		// Show status icon based on status
		switch state.Data.Status {
		case "completed", "done":
			r.statusIcon.Resource = theme.ConfirmIcon()
			r.statusIcon.Show()
		case "in_progress", "doing":
			r.statusIcon.Resource = theme.MediaPlayIcon()
			r.statusIcon.Show()
		case "blocked":
			r.statusIcon.Resource = theme.MediaPauseIcon()
			r.statusIcon.Show()
		default:
			// No status icon for todo or unknown
		}
	}
}

// updateLayout adjusts layout based on compact mode and available space
func (r *taskWidgetRenderer) updateLayout() {
	// In compact mode, reduce padding and spacing
	if r.widget.compact {
		// Adjust text sizes for compact mode
		r.titleLabel.TextStyle.Bold = false
		r.descLabel.Hide()
		r.metaContainer.Hide()
	} else {
		r.titleLabel.TextStyle.Bold = true
		if r.widget.currentState.Data != nil && r.widget.currentState.Data.Description != "" {
			r.descLabel.Show()
		}
		if r.widget.showMetadata {
			r.metaContainer.Show()
		}
	}
}

// updateEmptyState handles rendering when no task data is available
func (r *taskWidgetRenderer) updateEmptyState() {
	r.titleLabel.SetText("No Task Data")
	r.descLabel.SetText("")
	r.metaLabel.SetText("")

	r.background.FillColor = theme.Color(theme.ColorNameDisabled)
	r.border.StrokeColor = theme.Color(theme.ColorNameDisabled)

	// Hide all icons
	r.statusIcon.Hide()
	r.priorityIcon.Hide()
	r.loadingIcon.Hide()
	r.errorIcon.Hide()
}

// Renderer interface implementation

// Layout positions all visual components within the given size
func (r *taskWidgetRenderer) Layout(size fyne.Size) {
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
func (r *taskWidgetRenderer) MinSize() fyne.Size {
	if r.widget.compact {
		return fyne.NewSize(150, 40)
	}
	return fyne.NewSize(200, 80)
}

// Refresh updates the visual components when the widget state changes
func (r *taskWidgetRenderer) Refresh() {
	r.updateComponents()
	r.Layout(r.widget.Size())

	// Refresh all canvas objects
	for _, obj := range r.objects {
		obj.Refresh()
	}
}

// Objects returns the list of canvas objects that make up the widget
func (r *taskWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy cleans up renderer resources
func (r *taskWidgetRenderer) Destroy() {
	// No specific cleanup needed for basic components
	// Custom resources would be cleaned up here
}