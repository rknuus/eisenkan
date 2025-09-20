package harness

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FindButtonByText returns the first button whose Text matches.
func FindButtonByText(c fyne.Canvas, text string) *widget.Button {
	var found *widget.Button
	walkCanvas(c, func(o fyne.CanvasObject) bool {
		if b, ok := o.(*widget.Button); ok {
			if b.Text == text {
				found = b
				return true
			}
		}
		return false
	})
	return found
}

// FindEntryByPlaceholder finds an Entry by Placeholder text.
func FindEntryByPlaceholder(c fyne.Canvas, placeholder string) *widget.Entry {
	var found *widget.Entry
	walkCanvas(c, func(o fyne.CanvasObject) bool {
		if e, ok := o.(*widget.Entry); ok {
			if e.PlaceHolder == placeholder {
				found = e
				return true
			}
		}
		return false
	})
	return found
}

// CenterOf returns the center point of a canvas object in window coordinates.
func CenterOf(o fyne.CanvasObject) fyne.Position {
	b := o.Position()
	s := o.Size()
	return fyne.NewPos(b.X+s.Width/2, b.Y+s.Height/2)
}

// BoundsOf returns the top-left position and size of a canvas object.
func BoundsOf(o fyne.CanvasObject) (fyne.Position, fyne.Size) {
	return o.Position(), o.Size()
}

// walkCanvas walks all objects including overlays.
func walkCanvas(c fyne.Canvas, f func(o fyne.CanvasObject) bool) {
	// Walk content
	if co := c.Content(); co != nil {
		if walkObject(co, f) {
			return
		}
	}
	// Walk overlays
	for _, o := range c.Overlays().List() {
		if walkObject(o, f) {
			return
		}
	}
}

func walkObject(o fyne.CanvasObject, f func(o fyne.CanvasObject) bool) bool {
	if f(o) {
		return true
	}
	// Container-like exploration via type switches.
	switch t := o.(type) {
	case *container.AppTabs:
		for _, it := range t.Items {
			if walkObject(it.Content, f) {
				return true
			}
		}
	case *container.Scroll:
		if t.Content != nil && walkObject(t.Content, f) {
			return true
		}
	case *container.Split:
		if t.Leading != nil && walkObject(t.Leading, f) {
			return true
		}
		if t.Trailing != nil && walkObject(t.Trailing, f) {
			return true
		}
	case *container.DocTabs:
		for _, it := range t.Items {
			if walkObject(it.Content, f) {
				return true
			}
		}
	case *fyne.Container:
		for _, ch := range t.Objects {
			if walkObject(ch, f) {
				return true
			}
		}
	case *widget.Card:
		if t.Content != nil && walkObject(t.Content, f) {
			return true
		}
	case *canvas.Rectangle, *canvas.Text, *widget.Label, *widget.Button, *widget.Entry:
		// leafs
	default:
		// Try generic container interface if present
		if cc, ok := o.(interface{ Objects() []fyne.CanvasObject }); ok {
			for _, ch := range cc.Objects() {
				if walkObject(ch, f) {
					return true
				}
			}
		}
	}
	return false
}

// Must asserts non-nil with a formatted panic for harness usage.
func Must[T any](v *T, name string) *T {
	if v == nil {
		panic(fmt.Sprintf("%s not found", name))
	}
	return v
}
