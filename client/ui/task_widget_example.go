// Package ui provides an example demonstrating TaskWidget usage
package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/rknuus/eisenkan/client/engines"
)

// ExampleTaskWidget demonstrates basic TaskWidget usage
func ExampleTaskWidget() {
	// Create a new Fyne app
	myApp := app.New()
	myWindow := myApp.NewWindow("TaskWidget Example")
	myWindow.Resize(fyne.NewSize(400, 300))

	// Create sample task data
	taskData := &TaskData{
		ID:          "example-001",
		Title:       "Complete TaskWidget Implementation",
		Description: "Implement the TaskWidget component with all required features including event handling, state management, and integration with WorkflowManager.",
		Priority:    "urgent-important",
		Status:      "in_progress",
		Metadata: map[string]interface{}{
			"category":  "development",
			"tags":      []string{"ui", "widget", "fyne"},
			"assignee":  "developer",
			"estimate":  "4 hours",
		},
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Create FormattingEngine for text formatting
	formattingEngine := engines.NewFormattingEngine()

	// Create TaskWidget (with nil WorkflowManager for demo)
	taskWidget := NewTaskWidget(nil, formattingEngine, taskData)

	// Set up event handlers
	taskWidget.SetOnTapped(func() {
		println("Task tapped!")
	})

	taskWidget.SetOnDoubleTapped(func() {
		println("Task double-tapped - entering edit mode!")
	})

	taskWidget.SetOnSelectionChange(func(selected bool) {
		if selected {
			println("Task selected")
		} else {
			println("Task deselected")
		}
	})

	// Create some control buttons for demonstration
	selectBtn := widget.NewButton("Toggle Selection", func() {
		taskWidget.SetSelected(!taskWidget.IsSelected())
	})

	loadingBtn := widget.NewButton("Toggle Loading", func() {
		taskWidget.SetLoading(!taskWidget.currentState.IsLoading)
	})

	errorBtn := widget.NewButton("Set Error", func() {
		if taskWidget.currentState.HasError {
			taskWidget.SetError(nil)
		} else {
			taskWidget.SetError(fmt.Errorf("example error occurred"))
		}
	})

	compactBtn := widget.NewButton("Toggle Compact", func() {
		taskWidget.SetCompactMode(!taskWidget.compact)
	})

	metadataBtn := widget.NewButton("Toggle Metadata", func() {
		taskWidget.SetShowMetadata(!taskWidget.showMetadata)
	})

	// Create another task for comparison
	taskData2 := &TaskData{
		ID:          "example-002",
		Title:       "Review Implementation",
		Description: "Review the TaskWidget implementation for completeness",
		Priority:    "not-urgent-important",
		Status:      "todo",
		Metadata: map[string]interface{}{
			"category": "review",
			"tags":     []string{"code-review"},
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	taskWidget2 := NewTaskWidget(nil, formattingEngine, taskData2)
	taskWidget2.SetCompactMode(true)

	// Layout the widgets
	controlsContainer := container.NewHBox(
		selectBtn,
		loadingBtn,
		errorBtn,
		compactBtn,
		metadataBtn,
	)

	tasksContainer := container.NewVBox(
		widget.NewLabel("Full Task Widget:"),
		taskWidget,
		widget.NewSeparator(),
		widget.NewLabel("Compact Task Widget:"),
		taskWidget2,
	)

	mainContainer := container.NewBorder(
		controlsContainer,
		nil,
		nil,
		nil,
		tasksContainer,
	)

	myWindow.SetContent(mainContainer)

	// Show window and run
	myWindow.ShowAndRun()

	// Cleanup
	taskWidget.Destroy()
	taskWidget2.Destroy()
}

// This example would be run like:
// go run -tags example ./client/ui/task_widget_example.go