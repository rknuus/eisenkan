package journeys

import (
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// SRS refs: CTD-REQ-001..010 (dialog display + creation basics).
// Smoke scaffold: shows a placeholder dialog container; real dialog later.
func TestCreateTaskDialogJourney_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()

	form := container.NewVBox(
		widget.NewLabel("Create Task"),
		widget.NewEntry(),
		widget.NewButton("Create", func() {}),
	)
	win.SetContent(form)
	if win.Content() == nil {
		t.Fatal("window content is nil")
	}
}

// CTD-REQ-001..010: Matrix display and creation flow (scaffolded, skipped).
func TestCreateTaskDialogJourney_CreateFlow_Skipped(t *testing.T) {
	t.Skip("pending real CreateTaskDialog wiring")
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()
	_ = win
}

// CTD-REQ-011..020: DnD movement and reordering (scaffolded, skipped).
func TestCreateTaskDialogJourney_DnD_Organization_Skipped(t *testing.T) {
	t.Skip("pending real CreateTaskDialog wiring and UI DnD")
}

// CTD-REQ-021..025: Lifecycle open/with-data/close/cancel; cleanup (skipped scaffold).
func TestCreateTaskDialogJourney_Lifecycle_Skipped(t *testing.T) {
	t.Skip("pending real CreateTaskDialog wiring for lifecycle")
}

// CTD-REQ-026..030: Error handling for validation/workflow/DnD failures (skipped scaffold).
func TestCreateTaskDialogJourney_ErrorHandling_Skipped(t *testing.T) {
	t.Skip("pending real CreateTaskDialog wiring for error handling")
}

// CTD-REQ-031..035: Integrations with DragDropEngine, FormValidationEngine, WorkflowManager (skipped scaffold).
func TestCreateTaskDialogJourney_Integrations_Skipped(t *testing.T) {
	t.Skip("pending real CreateTaskDialog wiring for integrations")
}
