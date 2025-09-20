package journeys

import (
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// SRS refs: BV-REQ-001..010 (board display, task display basics).
// NOTE: placeholder smoke to ensure we can render a board-like container.
func TestBoardViewJourney_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_populated")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()

	// Simulate 4 columns using a simple grid
	col := func(name string, items []string) fyne.CanvasObject {
		list := widget.NewList(
			func() int { return len(items) },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(items[i]) },
		)
		return container.NewBorder(widget.NewLabel(name), nil, nil, nil, list)
	}
	grid := container.NewGridWithColumns(2,
		col("Urgent Important", []string{"Fix critical bug"}),
		col("Urgent Not Important", []string{"Schedule meeting"}),
		col("Not Urgent Important", []string{"Strategic planning"}),
		col("Not Urgent Not Important", []string{"Clean desk"}),
	)
	win.SetContent(grid)
	if win.Content() == nil {
		t.Fatal("window content is nil")
	}
	// Basic assertion: root is a container with 4 children
	if c, ok := win.Content().(*fyne.Container); ok {
		if got := len(c.Objects); got != 4 {
			t.Fatalf("expected 4 columns, have %d", got)
		}
	}
}

// BV-REQ-011..015: Drag-drop task movement happy path (scaffolded, skipped).
func TestBoardViewJourney_DnD_HappyPath_Skipped(t *testing.T) {
	t.Skip("pending real BoardView wiring and UI DnD")
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_populated")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()
	_ = win
}

// BV-REQ-031..035: Validation integration blocks invalid drag-drop (skipped scaffold).
func TestBoardViewJourney_DnD_InvalidBlocked_Skipped(t *testing.T) {
	t.Skip("pending real BoardView wiring and validation engine")
}

// BV-REQ-019: WIP limits enforced with visual feedback (skipped scaffold).
func TestBoardViewJourney_WIP_Enforcement_Skipped(t *testing.T) {
	t.Skip("pending real BoardView wiring and WIP configuration")
}

// BV-REQ-036..040: Event callbacks for selection/move/state changes (skipped scaffold).
func TestBoardViewJourney_EventCallbacks_Skipped(t *testing.T) {
	t.Skip("pending real BoardView event wiring")
}
