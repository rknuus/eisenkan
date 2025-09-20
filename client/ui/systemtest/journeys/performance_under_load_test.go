package journeys

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// Performance Under Load: outline budgets and structure (skipped scaffold).
// Budgets (soft): render ≤300ms, DnD feedback <50ms, updates ≤500ms.
func TestPerformance_UnderLoad_Skipped(t *testing.T) {
	t.Skip("pending real UI wiring and large fixture")
	_, win, cleanup := h.NewDeterministicApp("", fyne.NewSize(1024, 768))
	defer cleanup()
	_ = win

	d := h.Measure(func() {
		time.Sleep(10 * time.Millisecond)
	})
	_ = d
}
