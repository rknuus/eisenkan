package harness

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// DragFromTo performs a simple drag from the center of src to the center of dst.
func DragFromTo(c fyne.Canvas, src, dst fyne.CanvasObject) {
	start := CenterOf(src)
	end := CenterOf(dst)
	test.Drag(c, start, end.X, end.Y)
}
