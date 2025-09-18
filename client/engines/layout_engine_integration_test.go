package engines

import (
	"testing"
	"time"
)

// TestIntegration_LayoutEngine_CoreFunctionality tests the core LayoutEngine functionality
// that would be used in integration with FyneUtility
func TestIntegration_LayoutEngine_CoreFunctionality(t *testing.T) {
	engine := NewLayoutEngine()

	t.Run("Responsive_layout_integration", func(t *testing.T) {
		// Test responsive layout functionality that would work with Fyne containers
		originalLayout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID:     "container1",
					Bounds: Rectangle{Origin: Point{X: 0, Y: 0}, Size: Dimensions{Width: 200, Height: 150}},
				},
				{
					ID:     "container2",
					Bounds: Rectangle{Origin: Point{X: 0, Y: 160}, Size: Dimensions{Width: 200, Height: 150}},
				},
			},
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 800, Height: 600},
			},
		}

		// Test adaptation to smaller size (mobile responsive)
		newSize := Dimensions{Width: 400, Height: 300}
		result, err := engine.Responsive().AdaptToSize(originalLayout, newSize)
		if err != nil {
			t.Errorf("AdaptToSize failed: %v", err)
		}

		// Verify the layout was adapted to new dimensions
		if result.Bounds.Size.Width != 400 || result.Bounds.Size.Height != 300 {
			t.Errorf("Expected adapted size 400x300, got %vx%v",
				result.Bounds.Size.Width, result.Bounds.Size.Height)
		}

		// Verify components are still present
		if len(result.Components) != 2 {
			t.Errorf("Expected 2 components after adaptation, got %d", len(result.Components))
		}
	})

	t.Run("Spatial_calculations_for_container_layout", func(t *testing.T) {
		// Test spatial calculations that would be used with Fyne containers
		componentSpec := ComponentSpec{
			Size:   Dimensions{Width: 150, Height: 80},
			Offset: Point{X: 25, Y: 15},
		}

		containerRect := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 500, Height: 400},
		}

		result, err := engine.spatialMath.CalculateBounds(componentSpec, containerRect)
		if err != nil {
			t.Errorf("CalculateBounds failed: %v", err)
		}

		// Verify bounds calculation
		if result.Size.Width <= 0 || result.Size.Height <= 0 {
			t.Errorf("Invalid bounds calculated: %+v", result)
		}

		// Test collision detection between components
		rect1 := Rectangle{Origin: Point{X: 10, Y: 10}, Size: Dimensions{Width: 100, Height: 50}}
		rect2 := Rectangle{Origin: Point{X: 50, Y: 30}, Size: Dimensions{Width: 100, Height: 50}}

		collision := engine.spatialMath.DetectCollision(rect1, rect2)
		if !collision {
			t.Error("Expected collision detection to find overlap")
		}

		// Test non-overlapping rectangles
		rect3 := Rectangle{Origin: Point{X: 200, Y: 200}, Size: Dimensions{Width: 50, Height: 50}}
		collision = engine.spatialMath.DetectCollision(rect1, rect3)
		if collision {
			t.Error("Expected no collision for non-overlapping rectangles")
		}
	})

	t.Run("Animation_state_management", func(t *testing.T) {
		// Test animation state management that would work with animated Fyne layouts
		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID:     "animated-widget",
					Bounds: Rectangle{Origin: Point{X: 20, Y: 20}, Size: Dimensions{Width: 100, Height: 40}},
				},
				{
					ID:     "static-widget",
					Bounds: Rectangle{Origin: Point{X: 20, Y: 80}, Size: Dimensions{Width: 150, Height: 30}},
				},
			},
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 300, Height: 200},
			},
		}

		// Capture initial state
		initialState, err := engine.Animation().CaptureLayoutState(layout)
		if err != nil {
			t.Errorf("CaptureLayoutState failed: %v", err)
		}

		// Verify state capture
		if len(initialState.Components) != 2 {
			t.Errorf("Expected 2 components in captured state, got %d", len(initialState.Components))
		}

		// Create target state for animation
		targetLayout := layout
		targetLayout.Components[0].Bounds.Origin.X = 120 // Move first component

		targetState, err := engine.Animation().CaptureLayoutState(targetLayout)
		if err != nil {
			t.Errorf("Failed to capture target state: %v", err)
		}

		// Test interpolation between states
		interpolated, err := engine.Animation().InterpolateLayouts(initialState, targetState, 0.5)
		if err != nil {
			t.Errorf("InterpolateLayouts failed: %v", err)
		}

		// Verify interpolation result
		if len(interpolated.Components) != 2 {
			t.Errorf("Expected 2 components in interpolated state, got %d", len(interpolated.Components))
		}
	})

	t.Run("Kanban_layout_calculations", func(t *testing.T) {
		// Test Kanban layout calculations for board-style interfaces
		boardDimensions := Dimensions{Width: 1200, Height: 800}
		columnCount := 4

		result, err := engine.Kanban().CalculateColumnLayout(boardDimensions, columnCount)
		if err != nil {
			t.Errorf("CalculateColumnLayout failed: %v", err)
		}

		// Verify column count
		if len(result.Columns) != columnCount {
			t.Errorf("Expected %d columns, got %d", columnCount, len(result.Columns))
		}

		// Verify columns have reasonable dimensions
		totalWidth := float64(0)
		for i, col := range result.Columns {
			if col.Bounds.Size.Width <= 0 || col.Bounds.Size.Height <= 0 {
				t.Errorf("Column %d has invalid bounds: %+v", i, col.Bounds)
			}
			totalWidth += col.Bounds.Size.Width
		}

		// Columns should use most of the available width
		if totalWidth < boardDimensions.Width*0.7 {
			t.Errorf("Columns only use %v of %v available width", totalWidth, boardDimensions.Width)
		}

		// Test task card arrangement within a column
		cards := []TaskCard{
			{ID: "card1", Size: Dimensions{Width: 200, Height: 80}},
			{ID: "card2", Size: Dimensions{Width: 200, Height: 60}},
			{ID: "card3", Size: Dimensions{Width: 200, Height: 100}},
		}

		cardResult, err := engine.Kanban().ArrangeTaskCards(cards, result.Columns[0].Bounds)
		if err != nil {
			t.Errorf("ArrangeTaskCards failed: %v", err)
		}

		// Verify all cards were arranged
		if len(cardResult) != len(cards) {
			t.Errorf("Expected %d arranged cards, got %d", len(cards), len(cardResult))
		}
	})

	t.Run("Performance_requirements_verification", func(t *testing.T) {
		// Test that performance requirements are met for typical layouts
		layout := LayoutSpec{
			Components: make([]ComponentLayout, 20), // Moderate component count
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 500, Height: 400},
			},
		}

		// Create a realistic layout
		for i := range layout.Components {
			layout.Components[i] = ComponentLayout{
				ID:     string(rune('A' + i%26)),
				Bounds: Rectangle{Origin: Point{X: float64(i % 5 * 100), Y: float64(i / 5 * 50)}, Size: Dimensions{Width: 80, Height: 40}},
			}
		}

		// Measure responsive adaptation performance
		startTime := time.Now()
		newSize := Dimensions{Width: 600, Height: 400}
		_, err := engine.Responsive().AdaptToSize(layout, newSize)
		if err != nil {
			t.Errorf("AdaptToSize failed: %v", err)
		}
		responsiveDuration := time.Since(startTime)

		// Measure animation state capture performance
		startTime = time.Now()
		_, err = engine.Animation().CaptureLayoutState(layout)
		if err != nil {
			t.Errorf("CaptureLayoutState failed: %v", err)
		}
		captureDuration := time.Since(startTime)

		// Measure Kanban layout performance
		startTime = time.Now()
		boardDimensions := Dimensions{Width: 1000, Height: 600}
		_, err = engine.Kanban().CalculateColumnLayout(boardDimensions, 5)
		if err != nil {
			t.Errorf("CalculateColumnLayout failed: %v", err)
		}
		kanbanDuration := time.Since(startTime)

		// Verify performance requirements (2ms per SRS requirement LE-REQ-037)
		maxDuration := 2 * time.Millisecond

		if responsiveDuration > maxDuration {
			t.Errorf("Responsive adaptation took %v, expected <%v", responsiveDuration, maxDuration)
		}
		if captureDuration > maxDuration {
			t.Errorf("State capture took %v, expected <%v", captureDuration, maxDuration)
		}
		if kanbanDuration > maxDuration {
			t.Errorf("Kanban layout took %v, expected <%v", kanbanDuration, maxDuration)
		}
	})

	t.Run("Error_handling_and_validation", func(t *testing.T) {
		// Test error handling for invalid inputs that might come from Fyne integration

		// Test with invalid layout dimensions
		invalidLayout := LayoutSpec{
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: -100, Height: -100},
			},
		}

		newSize := Dimensions{Width: 400, Height: 300}
		_, err := engine.Responsive().AdaptToSize(invalidLayout, newSize)
		if err == nil {
			t.Error("Expected error for invalid layout dimensions")
		}

		// Test Kanban with invalid parameters
		boardDimensions := Dimensions{Width: 100, Height: 100}
		_, err = engine.Kanban().CalculateColumnLayout(boardDimensions, 0)
		if err == nil {
			t.Error("Expected error for zero column count")
		}

		// Test with invalid component specs
		invalidSpec := ComponentSpec{
			Size:   Dimensions{Width: -50, Height: -25},
			Offset: Point{X: 0, Y: 0},
		}

		containerRect := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 300, Height: 200},
		}

		_, err = engine.spatialMath.CalculateBounds(invalidSpec, containerRect)
		if err == nil {
			t.Error("Expected error for invalid component dimensions")
		}
	})

	t.Run("Cache_effectiveness_verification", func(t *testing.T) {
		// Test cache effectiveness as specified in SRS (80% hit rate requirement)
		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID:     "cached-component",
					Bounds: Rectangle{Origin: Point{X: 10, Y: 10}, Size: Dimensions{Width: 100, Height: 50}},
				},
			},
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 200, Height: 100},
			},
		}

		// Perform the same operation multiple times to test caching
		const iterations = 10
		var totalDuration time.Duration

		for i := 0; i < iterations; i++ {
			startTime := time.Now()
			_, err := engine.Animation().CaptureLayoutState(layout)
			if err != nil {
				t.Errorf("CaptureLayoutState failed on iteration %d: %v", i, err)
			}
			totalDuration += time.Since(startTime)
		}

		avgDuration := totalDuration / iterations

		// Later iterations should be faster due to caching
		// This is a basic test - a full cache test would require access to internal cache metrics
		if avgDuration > 500*time.Microsecond {
			t.Logf("Average operation time: %v (may indicate caching not fully effective)", avgDuration)
		}
	})
}

// TestIntegration_LayoutEngine_Concurrency tests concurrent access to LayoutEngine
func TestIntegration_LayoutEngine_Concurrency(t *testing.T) {
	engine := NewLayoutEngine()

	t.Run("Concurrent_layout_operations", func(t *testing.T) {
		// Test that LayoutEngine can handle concurrent operations safely
		const goroutines = 10
		const operationsPerGoroutine = 5

		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID:     "concurrent-test",
					Bounds: Rectangle{Origin: Point{X: 0, Y: 0}, Size: Dimensions{Width: 100, Height: 50}},
				},
			},
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 400, Height: 300},
			},
		}

		// Channel to collect errors from goroutines
		errorChan := make(chan error, goroutines*operationsPerGoroutine)
		doneChan := make(chan bool, goroutines)

		// Launch concurrent operations
		for g := 0; g < goroutines; g++ {
			go func(goroutineID int) {
				defer func() { doneChan <- true }()

				for op := 0; op < operationsPerGoroutine; op++ {
					// Test different operations concurrently
					switch op % 4 {
					case 0:
						newSize := Dimensions{Width: 300 + float64(goroutineID*10), Height: 200}
						_, err := engine.Responsive().AdaptToSize(layout, newSize)
						if err != nil {
							errorChan <- err
						}

					case 1:
						_, err := engine.Animation().CaptureLayoutState(layout)
						if err != nil {
							errorChan <- err
						}

					case 2:
						boardDim := Dimensions{Width: 800 + float64(goroutineID*50), Height: 600}
						_, err := engine.Kanban().CalculateColumnLayout(boardDim, 3)
						if err != nil {
							errorChan <- err
						}

					case 3:
						spec := ComponentSpec{
							Size:   Dimensions{Width: 80, Height: 30},
							Offset: Point{X: float64(goroutineID), Y: float64(op)},
						}
						containerRect := Rectangle{
							Origin: Point{X: 0, Y: 0},
							Size:   Dimensions{Width: 400, Height: 300},
						}
						_, err := engine.spatialMath.CalculateBounds(spec, containerRect)
						if err != nil {
							errorChan <- err
						}
					}
				}
			}(g)
		}

		// Wait for all goroutines to complete
		for i := 0; i < goroutines; i++ {
			<-doneChan
		}

		// Check for any errors
		close(errorChan)
		for err := range errorChan {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	})
}