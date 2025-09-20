package harness

import "fyne.io/fyne/v2"

// Drag performs a drag gesture from the center of src to the center of dst.
// Note: implement using fyne test APIs alongside real UI when tests are wired.
func Drag(canvas fyne.Canvas, src fyne.CanvasObject, dst fyne.CanvasObject) {
	_ = canvas
	_ = src
	_ = dst
	WaitUntilIdle(canvas)
}

// DragBy drags an object by a delta offset relative to its center.
// Note: implement using fyne test APIs alongside real UI when tests are wired.
func DragBy(canvas fyne.Canvas, obj fyne.CanvasObject, dx, dy float32) {
	_ = obj
	_ = dx
	_ = dy
	WaitUntilIdle(canvas)
}
