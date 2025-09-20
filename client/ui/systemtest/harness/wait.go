package harness

import (
	"errors"
	"time"

	"fyne.io/fyne/v2"
)

// WaitUntilIdle flushes Fyne events and waits briefly to allow rendering to settle.
func WaitUntilIdle(c fyne.Canvas) {
	// In Fyne v2 test harness, events are processed synchronously in most cases.
	// Brief pause to allow any post-event drawing to complete deterministically.
	time.Sleep(5 * time.Millisecond)
}

// WaitFor waits until cond returns true or timeout elapses.
func WaitFor(cond func() bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return nil
		}
		// Yield a bit between polls
		time.Sleep(2 * time.Millisecond)
	}
	if cond() {
		return nil
	}
	return errors.New("condition not met before timeout")
}
