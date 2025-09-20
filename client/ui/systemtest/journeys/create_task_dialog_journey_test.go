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
