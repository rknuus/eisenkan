package journeys

import (
	"path/filepath"
	"runtime"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	ui "github.com/rknuus/eisenkan/client/ui"
	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
	tm "github.com/rknuus/eisenkan/internal/managers/task_manager"
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

	// Prepare recent store pointing to two boards
	store := h.NewRecentStore(repoRoot)
	_ = store.Add(filepath.Join(repoRoot, "todo"))

	// Build deterministic app
	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()

	// Set env for recent store file
	t.Setenv("EISENKAN_RECENT_STORE", store.Path)

	// Instantiate real BoardSelectionView widget minimally
	// Note: pass nil for manager/engines; only list rendering is used
	var manager tm.TaskManager = nil
	bsv := ui.NewBoardSelectionView(manager, nil, nil, win)
	// Trigger refresh to load recent boards
	_ = bsv.RefreshBoards()
	win.SetContent(bsv.(fyne.CanvasObject))

	if win.Content() == nil {
		t.Fatal("window content is nil")
	}
}

// BSV-REQ-001..006, 009: Browse valid/invalid directories and error messaging.
func TestBoardSelectionJourney_Browse_ValidAndInvalid_Skipped(t *testing.T) {
	if runtime.GOOS == "js" { // placeholder guard when running in constrained envs
		t.Skip("skipping in constrained env")
	}
	t.Skip("pending real BoardSelectionView wiring")
	repoRoot := t.TempDir()
	// Seed minimal valid fixture
	valid := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, valid); err != nil {
		t.Fatalf("seed valid: %v", err)
	}
	// Prepare invalid and corrupt fixtures
	invalid := filepath.Join("..", "fixtures", "board_eisenhower_invalid")
	corrupt := filepath.Join("..", "fixtures", "board_eisenhower_corrupt")

	_, win, cleanup := h.NewDeterministicApp(repoRoot, fyne.NewSize(1024, 768))
	defer cleanup()

	// Placeholder UI
	list := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {},
	)
	root := container.NewBorder(widget.NewLabel("Boards"), nil, nil, nil, list)
	win.SetContent(root)

	_ = invalid
	_ = corrupt
}

// BSV-REQ-021..023: Recent boards behavior (limit 10); selection activation by Enter/double-click.
func TestBoardSelectionJourney_Recent_Selection_Skipped(t *testing.T) {
	t.Skip("pending real BoardSelectionView wiring; using recent store shim")
	repoRoot := t.TempDir()
	store := h.NewRecentStore(repoRoot)
	_ = store.Add("/path/to/board1")
	_ = store.Add("/path/to/board2")
}

// BSV-REQ-011..019: Search/filter/sort behavior with highlight (skipped scaffold).
func TestBoardSelectionJourney_Search_Filter_Sort_Skipped(t *testing.T) {
	t.Skip("pending real BoardSelectionView wiring and FormattingEngine fake")
}

// BSV-REQ-025..033: CRUD flows: create/edit/delete with validation and errors (skipped scaffold).
func TestBoardSelectionJourney_CRUD_Skipped(t *testing.T) {
	t.Skip("pending real BoardSelectionView wiring and TaskManager integration")
}
