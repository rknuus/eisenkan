package harness

import (
	"fmt"
	"time"
)

// Measure runs fn and returns its duration.
func Measure(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// AssertUnder returns error if d exceeds budget.
func AssertUnder(d, budget time.Duration) error {
	if d > budget {
		return fmt.Errorf("duration %v exceeds budget %v", d, budget)
	}
	return nil
}
