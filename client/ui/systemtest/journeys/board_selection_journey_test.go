package journeys

import (
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// SRS refs: BSV-REQ-001..010, 021..025 (discovery, selection basics)
// NOTE: This is a smoke scaffold that only verifies we can build a root
// container for the board selection view. Real wiring to app components will
// be added when those are ready.
func TestBoardSelectionJourney_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	// Seed minimal fixture into temp dir
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	// Build deterministic app and a trivial selection container placeholder
	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()

	list := widget.NewList(
		func() int { return 1 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("Eisenhower Board (Minimal)")
		},
	)
	root := container.NewBorder(widget.NewLabel("Boards"), nil, nil, nil, list)
	win.SetContent(root)

	// Minimal assertion: content set and non-nil
	if win.Content() == nil {
		t.Fatal("window content is nil")
	}
}
