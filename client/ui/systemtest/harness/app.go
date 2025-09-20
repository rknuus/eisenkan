package harness

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// NewDeterministicApp creates a deterministic test app and window with fixed
// theme, scale, and size. repoPath is accepted for symmetry; wiring into views
// is done by higher layers.
func NewDeterministicApp(repoPath string, size fyne.Size) (fyne.App, fyne.Window, func()) {
	a := test.NewApp()
	a.Settings().SetTheme(theme.LightTheme())

	// Start with an empty canvas; callers will set content.
	w := test.NewWindow(canvas.NewRectangle(theme.BackgroundColor()))
	w.Resize(size)

	cleanup := func() {
		w.Close()
		if ca := fyne.CurrentApp(); ca != nil {
			ca.Quit()
		}
	}
	return a, w, cleanup
}
