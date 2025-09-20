package journeys

import (
	"testing"

	"fyne.io/fyne/v2"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// AR-REQ-001..004: Startup shows BoardSelectionView, window title set.
func TestApplicationRoot_Startup_Skipped(t *testing.T) {
	t.Skip("pending real ApplicationRoot wiring")
	_, win, cleanup := h.NewDeterministicApp("", fyne.NewSize(1024, 768))
	defer cleanup()
	_ = win
}

// AR-REQ-005..013: Transition to BoardView on selection and back navigation.
func TestApplicationRoot_Transition_And_Back_Skipped(t *testing.T) {
	t.Skip("pending real ApplicationRoot wiring")
}

// AR-REQ-019..023: Shutdown via close and shortcut; AR-REQ-024..028: error handling.
func TestApplicationRoot_Shutdown_And_ErrorHandling_Skipped(t *testing.T) {
	t.Skip("pending real ApplicationRoot wiring")
}
