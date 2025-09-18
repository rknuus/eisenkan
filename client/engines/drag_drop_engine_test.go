package engines

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Unit Tests for DragDropEngine

func TestNewDragDropEngine(t *testing.T) {
	engine := NewDragDropEngine()

	if engine == nil {
		t.Fatal("NewDragDropEngine returned nil")
	}

	if engine.Drag() == nil {
		t.Error("Drag facet is nil")
	}

	if engine.Drop() == nil {
		t.Error("Drop facet is nil")
	}

	if engine.Visualize() == nil {
		t.Error("Visualize facet is nil")
	}
}

// IDrag Tests

func TestUnit_DragManager_StartDrag_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drag := engine.Drag()

	widget := widget.NewLabel("Test Widget")
	params := DragParams{
		SourceWidget: widget,
		DragType:     DragTypeTask,
		Metadata:     map[string]interface{}{"test": "value"},
	}

	handle, err := drag.StartDrag(widget, params)
	if err != nil {
		t.Errorf("StartDrag failed: %v", err)
	}

	if handle == "" {
		t.Error("StartDrag returned empty handle")
	}
}

func TestUnit_DragManager_StartDrag_NilWidget(t *testing.T) {
	engine := NewDragDropEngine()
	drag := engine.Drag()

	params := DragParams{
		DragType: DragTypeTask,
	}

	_, err := drag.StartDrag(nil, params)
	if err == nil {
		t.Error("StartDrag should fail with nil widget")
	}
}

func TestUnit_DragManager_UpdateDragPosition_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drag := engine.Drag()

	widget := widget.NewLabel("Test Widget")
	params := DragParams{
		SourceWidget: widget,
		DragType:     DragTypeTask,
	}

	handle, err := drag.StartDrag(widget, params)
	if err != nil {
		t.Fatalf("StartDrag failed: %v", err)
	}

	newPos := fyne.NewPos(100, 200)
	err = drag.UpdateDragPosition(handle, newPos)
	if err != nil {
		t.Errorf("UpdateDragPosition failed: %v", err)
	}
}

func TestUnit_DragManager_UpdateDragPosition_InvalidHandle(t *testing.T) {
	engine := NewDragDropEngine()
	drag := engine.Drag()

	err := drag.UpdateDragPosition("invalid-handle", fyne.NewPos(0, 0))
	if err == nil {
		t.Error("UpdateDragPosition should fail with invalid handle")
	}
}

func TestUnit_DragManager_CompleteDrag_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drag := engine.Drag()

	widget := widget.NewLabel("Test Widget")
	params := DragParams{
		SourceWidget: widget,
		DragType:     DragTypeTask,
	}

	handle, err := drag.StartDrag(widget, params)
	if err != nil {
		t.Fatalf("StartDrag failed: %v", err)
	}

	result, err := drag.CompleteDrag(handle)
	if err != nil {
		t.Errorf("CompleteDrag failed: %v", err)
	}

	if !result.Success {
		t.Error("CompleteDrag should return success")
	}
}

func TestUnit_DragManager_CancelDrag_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drag := engine.Drag()

	widget := widget.NewLabel("Test Widget")
	params := DragParams{
		SourceWidget: widget,
		DragType:     DragTypeTask,
	}

	handle, err := drag.StartDrag(widget, params)
	if err != nil {
		t.Fatalf("StartDrag failed: %v", err)
	}

	err = drag.CancelDrag(handle)
	if err != nil {
		t.Errorf("CancelDrag failed: %v", err)
	}
}

// IDrop Tests

func TestUnit_DropManager_RegisterDropZone_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:          "test-zone",
		Bounds:      fyne.NewPos(0, 0),
		Size:        fyne.NewSize(100, 100),
		AcceptTypes: []DragType{DragTypeTask},
	}

	zoneID, err := drop.RegisterDropZone(zone)
	if err != nil {
		t.Errorf("RegisterDropZone failed: %v", err)
	}

	if zoneID != zone.ID {
		t.Errorf("Expected zone ID %s, got %s", zone.ID, zoneID)
	}
}

func TestUnit_DropManager_RegisterDropZone_EmptyID(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:   "",
		Size: fyne.NewSize(100, 100),
	}

	_, err := drop.RegisterDropZone(zone)
	if err == nil {
		t.Error("RegisterDropZone should fail with empty ID")
	}
}

func TestUnit_DropManager_RegisterDropZone_InvalidSize(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:   "test-zone",
		Size: fyne.NewSize(0, 0),
	}

	_, err := drop.RegisterDropZone(zone)
	if err == nil {
		t.Error("RegisterDropZone should fail with invalid size")
	}
}

func TestUnit_DropManager_UnregisterDropZone_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:   "test-zone",
		Size: fyne.NewSize(100, 100),
	}

	zoneID, err := drop.RegisterDropZone(zone)
	if err != nil {
		t.Fatalf("RegisterDropZone failed: %v", err)
	}

	err = drop.UnregisterDropZone(zoneID)
	if err != nil {
		t.Errorf("UnregisterDropZone failed: %v", err)
	}
}

func TestUnit_DropManager_UnregisterDropZone_NotFound(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	err := drop.UnregisterDropZone("nonexistent-zone")
	if err == nil {
		t.Error("UnregisterDropZone should fail with nonexistent zone")
	}
}

func TestUnit_DropManager_ValidateDropTarget_Success(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:          "test-zone",
		Bounds:      fyne.NewPos(0, 0),
		Size:        fyne.NewSize(100, 100),
		AcceptTypes: []DragType{DragTypeTask},
	}

	_, err := drop.RegisterDropZone(zone)
	if err != nil {
		t.Fatalf("RegisterDropZone failed: %v", err)
	}

	context := DragContext{
		DragType: DragTypeTask,
	}

	// Position inside the zone
	valid, err := drop.ValidateDropTarget(fyne.NewPos(50, 50), context)
	if err != nil {
		t.Errorf("ValidateDropTarget failed: %v", err)
	}

	if !valid {
		t.Error("Position should be valid for drop")
	}
}

func TestUnit_DropManager_ValidateDropTarget_OutsideZone(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:          "test-zone",
		Bounds:      fyne.NewPos(0, 0),
		Size:        fyne.NewSize(100, 100),
		AcceptTypes: []DragType{DragTypeTask},
	}

	_, err := drop.RegisterDropZone(zone)
	if err != nil {
		t.Fatalf("RegisterDropZone failed: %v", err)
	}

	context := DragContext{
		DragType: DragTypeTask,
	}

	// Position outside the zone
	valid, err := drop.ValidateDropTarget(fyne.NewPos(200, 200), context)
	if err != nil {
		t.Errorf("ValidateDropTarget failed: %v", err)
	}

	if valid {
		t.Error("Position should not be valid for drop")
	}
}

func TestUnit_DropManager_ValidateDropTarget_WrongType(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	zone := DropZoneSpec{
		ID:          "test-zone",
		Bounds:      fyne.NewPos(0, 0),
		Size:        fyne.NewSize(100, 100),
		AcceptTypes: []DragType{DragTypeTask}, // Only accepts tasks
	}

	_, err := drop.RegisterDropZone(zone)
	if err != nil {
		t.Fatalf("RegisterDropZone failed: %v", err)
	}

	context := DragContext{
		DragType: DragTypeSubtask, // Trying to drop a subtask
	}

	// Position inside the zone but wrong type
	valid, err := drop.ValidateDropTarget(fyne.NewPos(50, 50), context)
	if err != nil {
		t.Errorf("ValidateDropTarget failed: %v", err)
	}

	if valid {
		t.Error("Position should not be valid for wrong drag type")
	}
}

func TestUnit_DropManager_GetActiveZones(t *testing.T) {
	engine := NewDragDropEngine()
	drop := engine.Drop()

	// Start with no zones
	zones := drop.GetActiveZones()
	if len(zones) != 0 {
		t.Error("Should start with no active zones")
	}

	// Register a zone
	zone := DropZoneSpec{
		ID:   "test-zone",
		Size: fyne.NewSize(100, 100),
	}

	_, err := drop.RegisterDropZone(zone)
	if err != nil {
		t.Fatalf("RegisterDropZone failed: %v", err)
	}

	// Should now have one zone
	zones = drop.GetActiveZones()
	if len(zones) != 1 {
		t.Errorf("Expected 1 active zone, got %d", len(zones))
	}

	if zones[0].ID != zone.ID {
		t.Errorf("Expected zone ID %s, got %s", zone.ID, zones[0].ID)
	}
}

// IVisualize Tests

func TestUnit_VisualManager_CreateDragIndicator_Success(t *testing.T) {
	engine := NewDragDropEngine()
	visual := engine.Visualize()

	widget := widget.NewLabel("Test Widget")

	indicator, err := visual.CreateDragIndicator(widget)
	if err != nil {
		t.Errorf("CreateDragIndicator failed: %v", err)
	}

	if indicator == nil {
		t.Error("CreateDragIndicator returned nil indicator")
	}
}

func TestUnit_VisualManager_CreateDragIndicator_NilSource(t *testing.T) {
	engine := NewDragDropEngine()
	visual := engine.Visualize()

	_, err := visual.CreateDragIndicator(nil)
	if err == nil {
		t.Error("CreateDragIndicator should fail with nil source")
	}
}

func TestUnit_VisualManager_UpdateIndicatorPosition_Success(t *testing.T) {
	engine := NewDragDropEngine()
	visual := engine.Visualize()

	widget := widget.NewLabel("Test Widget")

	indicator, err := visual.CreateDragIndicator(widget)
	if err != nil {
		t.Fatalf("CreateDragIndicator failed: %v", err)
	}

	newPos := fyne.NewPos(100, 200)
	err = visual.UpdateIndicatorPosition(indicator, newPos)
	if err != nil {
		t.Errorf("UpdateIndicatorPosition failed: %v", err)
	}
}

func TestUnit_VisualManager_UpdateIndicatorPosition_NilIndicator(t *testing.T) {
	engine := NewDragDropEngine()
	visual := engine.Visualize()

	err := visual.UpdateIndicatorPosition(nil, fyne.NewPos(0, 0))
	if err == nil {
		t.Error("UpdateIndicatorPosition should fail with nil indicator")
	}
}

func TestUnit_VisualManager_ShowDropFeedback(t *testing.T) {
	engine := NewDragDropEngine()
	visual := engine.Visualize()

	// Test both accepted and rejected feedback
	err := visual.ShowDropFeedback("test-zone", true)
	if err != nil {
		t.Errorf("ShowDropFeedback failed for accepted: %v", err)
	}

	err = visual.ShowDropFeedback("test-zone", false)
	if err != nil {
		t.Errorf("ShowDropFeedback failed for rejected: %v", err)
	}
}

func TestUnit_VisualManager_CleanupVisuals(t *testing.T) {
	engine := NewDragDropEngine()
	visual := engine.Visualize()

	// Cleanup should not fail even if handle doesn't exist
	err := visual.CleanupVisuals("nonexistent-handle")
	if err != nil {
		t.Errorf("CleanupVisuals failed: %v", err)
	}
}

// Integration Tests

func TestIntegration_DragDropEngine_BasicWorkflow(t *testing.T) {
	engine := NewDragDropEngine()

	// Create a test widget
	testWidget := widget.NewLabel("Draggable Item")

	// Register a drop zone
	zone := DropZoneSpec{
		ID:          "target-zone",
		Bounds:      fyne.NewPos(100, 100),
		Size:        fyne.NewSize(200, 200),
		AcceptTypes: []DragType{DragTypeTask},
	}

	zoneID, err := engine.Drop().RegisterDropZone(zone)
	if err != nil {
		t.Fatalf("Failed to register drop zone: %v", err)
	}

	// Start a drag operation
	params := DragParams{
		SourceWidget: testWidget,
		DragType:     DragTypeTask,
		Metadata:     map[string]interface{}{"id": "task-123"},
	}

	handle, err := engine.Drag().StartDrag(testWidget, params)
	if err != nil {
		t.Fatalf("Failed to start drag: %v", err)
	}

	// Create drag indicator
	indicator, err := engine.Visualize().CreateDragIndicator(testWidget)
	if err != nil {
		t.Fatalf("Failed to create drag indicator: %v", err)
	}

	// Update drag position to inside the drop zone
	dropPos := fyne.NewPos(150, 150) // Inside the zone bounds
	err = engine.Drag().UpdateDragPosition(handle, dropPos)
	if err != nil {
		t.Fatalf("Failed to update drag position: %v", err)
	}

	err = engine.Visualize().UpdateIndicatorPosition(indicator, dropPos)
	if err != nil {
		t.Fatalf("Failed to update indicator position: %v", err)
	}

	// Validate the drop target
	context := DragContext{
		Handle:     handle,
		DragType:   DragTypeTask,
		CurrentPos: dropPos,
	}

	valid, err := engine.Drop().ValidateDropTarget(dropPos, context)
	if err != nil {
		t.Fatalf("Failed to validate drop target: %v", err)
	}

	if !valid {
		t.Error("Drop target should be valid")
	}

	// Show drop feedback
	err = engine.Visualize().ShowDropFeedback(zoneID, true)
	if err != nil {
		t.Fatalf("Failed to show drop feedback: %v", err)
	}

	// Complete the drag
	result, err := engine.Drag().CompleteDrag(handle)
	if err != nil {
		t.Fatalf("Failed to complete drag: %v", err)
	}

	if !result.Success {
		t.Error("Drag completion should be successful")
	}

	// Cleanup visuals
	err = engine.Visualize().CleanupVisuals(handle)
	if err != nil {
		t.Fatalf("Failed to cleanup visuals: %v", err)
	}

	// Unregister the drop zone
	err = engine.Drop().UnregisterDropZone(zoneID)
	if err != nil {
		t.Fatalf("Failed to unregister drop zone: %v", err)
	}
}

func TestIntegration_DragDropEngine_CancelWorkflow(t *testing.T) {
	engine := NewDragDropEngine()

	// Create a test widget
	testWidget := widget.NewLabel("Draggable Item")
	originalPos := fyne.NewPos(50, 50)
	testWidget.Move(originalPos)

	// Start a drag operation
	params := DragParams{
		SourceWidget: testWidget,
		DragType:     DragTypeTask,
	}

	handle, err := engine.Drag().StartDrag(testWidget, params)
	if err != nil {
		t.Fatalf("Failed to start drag: %v", err)
	}

	// Move the widget
	newPos := fyne.NewPos(200, 200)
	err = engine.Drag().UpdateDragPosition(handle, newPos)
	if err != nil {
		t.Fatalf("Failed to update drag position: %v", err)
	}

	// Cancel the drag
	err = engine.Drag().CancelDrag(handle)
	if err != nil {
		t.Fatalf("Failed to cancel drag: %v", err)
	}

	// Widget should be restored to original position
	if testWidget.Position() != originalPos {
		t.Errorf("Widget position should be restored to %v, got %v", originalPos, testWidget.Position())
	}
}
