package engines

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

// TestUnit_LayoutEngine_NewLayoutEngine tests layout engine creation
func TestUnit_LayoutEngine_NewLayoutEngine(t *testing.T) {
	engine := NewLayoutEngine()
	if engine == nil {
		t.Fatal("NewLayoutEngine() returned nil")
	}

	if engine.Responsive() == nil {
		t.Error("Responsive facet is nil")
	}
	if engine.Animation() == nil {
		t.Error("Animation facet is nil")
	}
	if engine.Kanban() == nil {
		t.Error("Kanban facet is nil")
	}
	if engine.DragDrop() == nil {
		t.Error("DragDrop facet is nil")
	}
}

// TestUnit_LayoutEngine_SpatialMath tests internal spatial math calculations
func TestUnit_LayoutEngine_SpatialMath(t *testing.T) {
	spatialMath := &SpatialMath{}

	t.Run("CalculateBounds_with_valid_input", func(t *testing.T) {
		spec := ComponentSpec{
			Size:   Dimensions{Width: 100, Height: 50},
			Offset: Point{X: 10, Y: 5},
		}
		container := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 200, Height: 100},
		}

		bounds, err := spatialMath.CalculateBounds(spec, container)
		if err != nil {
			t.Fatalf("CalculateBounds failed: %v", err)
		}

		expectedBounds := Rectangle{
			Origin: Point{X: 10, Y: 5},
			Size:   Dimensions{Width: 100, Height: 50},
		}

		if bounds != expectedBounds {
			t.Errorf("Expected bounds %+v, got %+v", expectedBounds, bounds)
		}
	})

	t.Run("CalculateBounds_with_negative_dimensions", func(t *testing.T) {
		spec := ComponentSpec{
			Size:   Dimensions{Width: -100, Height: 50},
			Offset: Point{X: 10, Y: 5},
		}
		container := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 200, Height: 100},
		}

		_, err := spatialMath.CalculateBounds(spec, container)
		if err == nil {
			t.Error("Expected error for negative dimensions")
		}
	})

	t.Run("CalculateBounds_with_infinite_dimensions", func(t *testing.T) {
		spec := ComponentSpec{
			Size:   Dimensions{Width: math.Inf(1), Height: 50},
			Offset: Point{X: 10, Y: 5},
		}
		container := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 200, Height: 100},
		}

		_, err := spatialMath.CalculateBounds(spec, container)
		if err == nil {
			t.Error("Expected error for infinite dimensions")
		}
	})

	t.Run("CalculateBounds_with_NaN_dimensions", func(t *testing.T) {
		spec := ComponentSpec{
			Size:   Dimensions{Width: math.NaN(), Height: 50},
			Offset: Point{X: 10, Y: 5},
		}
		container := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 200, Height: 100},
		}

		_, err := spatialMath.CalculateBounds(spec, container)
		if err == nil {
			t.Error("Expected error for NaN dimensions")
		}
	})

	t.Run("CalculateDistance", func(t *testing.T) {
		p1 := Point{X: 0, Y: 0}
		p2 := Point{X: 3, Y: 4}

		distance := spatialMath.CalculateDistance(p1, p2)
		expected := 5.0 // 3-4-5 triangle

		if math.Abs(distance-expected) > 0.001 {
			t.Errorf("Expected distance %f, got %f", expected, distance)
		}
	})

	t.Run("DetectCollision_with_overlapping_rectangles", func(t *testing.T) {
		r1 := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 100, Height: 100},
		}
		r2 := Rectangle{
			Origin: Point{X: 50, Y: 50},
			Size:   Dimensions{Width: 100, Height: 100},
		}

		collision := spatialMath.DetectCollision(r1, r2)
		if !collision {
			t.Error("Expected collision detection to return true for overlapping rectangles")
		}
	})

	t.Run("DetectCollision_with_non_overlapping_rectangles", func(t *testing.T) {
		r1 := Rectangle{
			Origin: Point{X: 0, Y: 0},
			Size:   Dimensions{Width: 50, Height: 50},
		}
		r2 := Rectangle{
			Origin: Point{X: 100, Y: 100},
			Size:   Dimensions{Width: 50, Height: 50},
		}

		collision := spatialMath.DetectCollision(r1, r2)
		if collision {
			t.Error("Expected collision detection to return false for non-overlapping rectangles")
		}
	})
}

// TestUnit_LayoutEngine_ResponsiveFacet tests responsive layout capabilities
func TestUnit_LayoutEngine_ResponsiveFacet(t *testing.T) {
	engine := NewLayoutEngine()
	responsive := engine.Responsive()

	t.Run("AdaptToSize_with_scaling", func(t *testing.T) {
		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID: "comp1",
					Bounds: Rectangle{
						Origin: Point{X: 10, Y: 10},
						Size:   Dimensions{Width: 50, Height: 50},
					},
					Visible: true,
				},
			},
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 100, Height: 100},
			},
		}

		newSize := Dimensions{Width: 200, Height: 200}
		adapted, err := responsive.AdaptToSize(layout, newSize)
		if err != nil {
			t.Fatalf("AdaptToSize failed: %v", err)
		}

		if adapted.Bounds.Size != newSize {
			t.Errorf("Expected bounds size %+v, got %+v", newSize, adapted.Bounds.Size)
		}

		// Components should be scaled 2x
		expectedComp := adapted.Components[0]
		if expectedComp.Bounds.Origin.X != 20 || expectedComp.Bounds.Origin.Y != 20 {
			t.Errorf("Expected component origin (20, 20), got (%f, %f)",
				expectedComp.Bounds.Origin.X, expectedComp.Bounds.Origin.Y)
		}
		if expectedComp.Bounds.Size.Width != 100 || expectedComp.Bounds.Size.Height != 100 {
			t.Errorf("Expected component size (100, 100), got (%f, %f)",
				expectedComp.Bounds.Size.Width, expectedComp.Bounds.Size.Height)
		}
	})

	t.Run("AdaptToSize_with_invalid_dimensions", func(t *testing.T) {
		layout := LayoutSpec{
			Bounds: Rectangle{
				Size: Dimensions{Width: 100, Height: 100},
			},
		}

		_, err := responsive.AdaptToSize(layout, Dimensions{Width: -100, Height: 200})
		if err == nil {
			t.Error("Expected error for negative width")
		}

		_, err = responsive.AdaptToSize(layout, Dimensions{Width: 200, Height: 0})
		if err == nil {
			t.Error("Expected error for zero height")
		}
	})

	t.Run("ApplyBreakpoints", func(t *testing.T) {
		layout := LayoutSpec{
			Bounds: Rectangle{
				Size: Dimensions{Width: 800, Height: 600},
			},
		}

		breakpoints := []Breakpoint{
			{
				Name:     "mobile",
				MinWidth: 0,
				MaxWidth: 600,
				Layout: BreakpointLayout{
					Columns:     1,
					Orientation: OrientationVertical,
				},
			},
			{
				Name:     "desktop",
				MinWidth: 600,
				MaxWidth: 0,
				Layout: BreakpointLayout{
					Columns:     3,
					Orientation: OrientationHorizontal,
				},
			},
		}

		result, err := responsive.ApplyBreakpoints(layout, breakpoints)
		if err != nil {
			t.Fatalf("ApplyBreakpoints failed: %v", err)
		}

		if result.Metadata.Version != "desktop" {
			t.Errorf("Expected desktop breakpoint, got %s", result.Metadata.Version)
		}
	})
}

// TestUnit_LayoutEngine_AnimationFacet tests animation support capabilities
func TestUnit_LayoutEngine_AnimationFacet(t *testing.T) {
	engine := NewLayoutEngine()
	animation := engine.Animation()

	t.Run("CaptureLayoutState", func(t *testing.T) {
		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID: "comp1",
					Bounds: Rectangle{
						Origin: Point{X: 10, Y: 10},
						Size:   Dimensions{Width: 50, Height: 50},
					},
					Visible: true,
				},
			},
		}

		state, err := animation.CaptureLayoutState(layout)
		if err != nil {
			t.Fatalf("CaptureLayoutState failed: %v", err)
		}

		if len(state.Components) != 1 {
			t.Errorf("Expected 1 component in state, got %d", len(state.Components))
		}

		comp := state.Components["comp1"]
		if comp.Bounds.Origin.X != 10 || comp.Bounds.Origin.Y != 10 {
			t.Errorf("Expected component origin (10, 10), got (%f, %f)",
				comp.Bounds.Origin.X, comp.Bounds.Origin.Y)
		}
	})

	t.Run("RestoreLayoutState", func(t *testing.T) {
		state := LayoutState{
			Components: map[string]ComponentState{
				"comp1": {
					Bounds: Rectangle{
						Origin: Point{X: 20, Y: 20},
						Size:   Dimensions{Width: 60, Height: 60},
					},
					Visible: true,
				},
			},
		}

		layout, err := animation.RestoreLayoutState(state)
		if err != nil {
			t.Fatalf("RestoreLayoutState failed: %v", err)
		}

		if len(layout.Components) != 1 {
			t.Errorf("Expected 1 component in layout, got %d", len(layout.Components))
		}

		comp := layout.Components[0]
		if comp.Bounds.Origin.X != 20 || comp.Bounds.Origin.Y != 20 {
			t.Errorf("Expected component origin (20, 20), got (%f, %f)",
				comp.Bounds.Origin.X, comp.Bounds.Origin.Y)
		}
	})

	t.Run("InterpolateLayouts", func(t *testing.T) {
		startState := LayoutState{
			Components: map[string]ComponentState{
				"comp1": {
					Bounds: Rectangle{
						Origin: Point{X: 0, Y: 0},
						Size:   Dimensions{Width: 50, Height: 50},
					},
				},
			},
		}

		endState := LayoutState{
			Components: map[string]ComponentState{
				"comp1": {
					Bounds: Rectangle{
						Origin: Point{X: 100, Y: 100},
						Size:   Dimensions{Width: 100, Height: 100},
					},
				},
			},
		}

		interpolated, err := animation.InterpolateLayouts(startState, endState, 0.5)
		if err != nil {
			t.Fatalf("InterpolateLayouts failed: %v", err)
		}

		comp := interpolated.Components["comp1"]
		if comp.Bounds.Origin.X != 50 || comp.Bounds.Origin.Y != 50 {
			t.Errorf("Expected interpolated origin (50, 50), got (%f, %f)",
				comp.Bounds.Origin.X, comp.Bounds.Origin.Y)
		}
		if comp.Bounds.Size.Width != 75 || comp.Bounds.Size.Height != 75 {
			t.Errorf("Expected interpolated size (75, 75), got (%f, %f)",
				comp.Bounds.Size.Width, comp.Bounds.Size.Height)
		}
	})

	t.Run("InterpolateLayouts_with_invalid_progress", func(t *testing.T) {
		startState := LayoutState{Components: map[string]ComponentState{}}
		endState := LayoutState{Components: map[string]ComponentState{}}

		_, err := animation.InterpolateLayouts(startState, endState, -0.1)
		if err == nil {
			t.Error("Expected error for negative progress")
		}

		_, err = animation.InterpolateLayouts(startState, endState, 1.1)
		if err == nil {
			t.Error("Expected error for progress > 1")
		}
	})

	t.Run("ValidateTransition", func(t *testing.T) {
		validParams := TransitionParams{
			Duration:  250 * time.Millisecond,
			FrameRate: 60,
		}

		err := animation.ValidateTransition(validParams)
		if err != nil {
			t.Errorf("ValidateTransition failed for valid params: %v", err)
		}

		invalidParams := TransitionParams{
			Duration:  -100 * time.Millisecond,
			FrameRate: 60,
		}

		err = animation.ValidateTransition(invalidParams)
		if err == nil {
			t.Error("Expected error for negative duration")
		}

		invalidParams = TransitionParams{
			Duration:  250 * time.Millisecond,
			FrameRate: 0,
		}

		err = animation.ValidateTransition(invalidParams)
		if err == nil {
			t.Error("Expected error for zero frame rate")
		}
	})
}

// TestUnit_LayoutEngine_KanbanFacet tests Kanban layout capabilities
func TestUnit_LayoutEngine_KanbanFacet(t *testing.T) {
	engine := NewLayoutEngine()
	kanban := engine.Kanban()

	t.Run("CalculateColumnLayout", func(t *testing.T) {
		boardSize := Dimensions{Width: 800, Height: 600}
		columnCount := 4

		layout, err := kanban.CalculateColumnLayout(boardSize, columnCount)
		if err != nil {
			t.Fatalf("CalculateColumnLayout failed: %v", err)
		}

		if len(layout.Columns) != columnCount {
			t.Errorf("Expected %d columns, got %d", columnCount, len(layout.Columns))
		}

		if layout.TotalWidth != boardSize.Width {
			t.Errorf("Expected total width %f, got %f", boardSize.Width, layout.TotalWidth)
		}

		// Check that columns don't overlap
		for i := 1; i < len(layout.Columns); i++ {
			prevColumn := layout.Columns[i-1]
			currentColumn := layout.Columns[i]

			prevRight := prevColumn.Bounds.Origin.X + prevColumn.Bounds.Size.Width
			currentLeft := currentColumn.Bounds.Origin.X

			if prevRight > currentLeft {
				t.Errorf("Columns %d and %d overlap", i-1, i)
			}
		}
	})

	t.Run("CalculateColumnLayout_with_invalid_count", func(t *testing.T) {
		boardSize := Dimensions{Width: 800, Height: 600}

		_, err := kanban.CalculateColumnLayout(boardSize, 0)
		if err == nil {
			t.Error("Expected error for zero column count")
		}

		_, err = kanban.CalculateColumnLayout(boardSize, -1)
		if err == nil {
			t.Error("Expected error for negative column count")
		}
	})

	t.Run("CalculateColumnLayout_with_insufficient_width", func(t *testing.T) {
		boardSize := Dimensions{Width: 50, Height: 600} // Very narrow
		columnCount := 10                                // Too many columns

		_, err := kanban.CalculateColumnLayout(boardSize, columnCount)
		if err == nil {
			t.Error("Expected error for insufficient width")
		}
	})

	t.Run("ArrangeTaskCards", func(t *testing.T) {
		cards := []TaskCard{
			{ID: "card1", Size: Dimensions{Width: 200, Height: 100}},
			{ID: "card2", Size: Dimensions{Width: 200, Height: 150}},
		}

		columnBounds := Rectangle{
			Origin: Point{X: 10, Y: 10},
			Size:   Dimensions{Width: 220, Height: 400},
		}

		positions, err := kanban.ArrangeTaskCards(cards, columnBounds)
		if err != nil {
			t.Fatalf("ArrangeTaskCards failed: %v", err)
		}

		if len(positions) != len(cards) {
			t.Errorf("Expected %d positions, got %d", len(cards), len(positions))
		}

		// Check that cards are positioned vertically
		if positions[0].Bounds.Origin.Y >= positions[1].Bounds.Origin.Y {
			t.Error("Cards should be arranged vertically")
		}

		// Check that cards fit within column bounds
		for _, pos := range positions {
			if pos.Bounds.Origin.X < columnBounds.Origin.X ||
				pos.Bounds.Origin.X+pos.Bounds.Size.Width > columnBounds.Origin.X+columnBounds.Size.Width {
				t.Error("Card extends outside column bounds horizontally")
			}
		}
	})
}

// TestUnit_LayoutEngine_DragDropFacet tests drag and drop spatial capabilities
func TestUnit_LayoutEngine_DragDropFacet(t *testing.T) {
	engine := NewLayoutEngine()
	dragDrop := engine.DragDrop()

	t.Run("CalculateDropZones", func(t *testing.T) {
		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID: "target1",
					Bounds: Rectangle{
						Origin: Point{X: 50, Y: 50},
						Size:   Dimensions{Width: 100, Height: 100},
					},
				},
				{
					ID: "target2",
					Bounds: Rectangle{
						Origin: Point{X: 200, Y: 200},
						Size:   Dimensions{Width: 100, Height: 100},
					},
				},
			},
		}

		dragPosition := Point{X: 80, Y: 80} // Close to target1

		zones, err := dragDrop.CalculateDropZones(layout, dragPosition)
		if err != nil {
			t.Fatalf("CalculateDropZones failed: %v", err)
		}

		if len(zones) == 0 {
			t.Error("Expected at least one drop zone")
		}

		// Should find target1 as closest
		foundTarget1 := false
		for _, zone := range zones {
			if zone.Target == "target1" {
				foundTarget1 = true
				break
			}
		}

		if !foundTarget1 {
			t.Error("Expected to find target1 as a drop zone")
		}
	})

	t.Run("ValidateDropTarget", func(t *testing.T) {
		layout := LayoutSpec{
			Bounds: Rectangle{
				Origin: Point{X: 0, Y: 0},
				Size:   Dimensions{Width: 500, Height: 500},
			},
		}

		// Valid target within layout bounds
		validTarget := DropTarget{
			ID: "valid",
			Bounds: Rectangle{
				Origin: Point{X: 100, Y: 100},
				Size:   Dimensions{Width: 50, Height: 50},
			},
		}

		valid, err := dragDrop.ValidateDropTarget(layout, validTarget)
		if err != nil {
			t.Fatalf("ValidateDropTarget failed: %v", err)
		}

		if !valid {
			t.Error("Expected valid target to be valid")
		}

		// Invalid target outside layout bounds
		invalidTarget := DropTarget{
			ID: "invalid",
			Bounds: Rectangle{
				Origin: Point{X: 600, Y: 600}, // Outside layout
				Size:   Dimensions{Width: 50, Height: 50},
			},
		}

		valid, err = dragDrop.ValidateDropTarget(layout, invalidTarget)
		if err != nil {
			t.Fatalf("ValidateDropTarget failed: %v", err)
		}

		if valid {
			t.Error("Expected invalid target to be invalid")
		}
	})

	t.Run("ComputeSnapPoints", func(t *testing.T) {
		layout := LayoutSpec{
			Components: []ComponentLayout{
				{
					ID: "comp1",
					Bounds: Rectangle{
						Origin: Point{X: 100, Y: 100},
						Size:   Dimensions{Width: 50, Height: 50},
					},
				},
			},
		}

		position := Point{X: 105, Y: 98} // Close to comp1

		snapPoints, err := dragDrop.ComputeSnapPoints(layout, position)
		if err != nil {
			t.Fatalf("ComputeSnapPoints failed: %v", err)
		}

		if len(snapPoints) == 0 {
			t.Error("Expected at least one snap point")
		}

		// Should have grid snap point
		hasGridSnap := false
		for _, snap := range snapPoints {
			if snap.Type == SnapToGrid {
				hasGridSnap = true
				break
			}
		}

		if !hasGridSnap {
			t.Error("Expected grid snap point")
		}
	})
}

// TestUnit_LayoutEngine_Optimization tests caching and performance
func TestUnit_LayoutEngine_Optimization(t *testing.T) {
	optimization := &Optimization{
		cache: make(map[string]*CachedLayout),
		mu:    &sync.RWMutex{},
	}

	t.Run("Cache_operations", func(t *testing.T) {
		key := "test-key"
		value := "test-value"

		// Test cache miss
		result, hit := optimization.GetCached(key)
		if hit {
			t.Error("Expected cache miss for new key")
		}
		if result != nil {
			t.Error("Expected nil result for cache miss")
		}

		// Test cache set and hit
		optimization.SetCached(key, value)

		result, hit = optimization.GetCached(key)
		if !hit {
			t.Error("Expected cache hit after setting value")
		}
		if result != value {
			t.Errorf("Expected cached value %v, got %v", value, result)
		}
	})

	t.Run("Cache_eviction", func(t *testing.T) {
		// Fill cache beyond limit
		for i := 0; i < 1100; i++ {
			optimization.SetCached(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
		}

		// Cache should be limited
		if len(optimization.cache) > 1000 {
			t.Errorf("Cache size %d exceeds limit", len(optimization.cache))
		}
	})
}

// Helper function for tests
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}