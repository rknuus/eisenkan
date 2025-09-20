package harness

import (
	"fyne.io/fyne/v2"
)

// SendKey is a placeholder; use widget APIs directly in tests.
func SendKey(canvas fyne.Canvas, name fyne.KeyName) {
	_ = name
	WaitUntilIdle(canvas)
}
