package engines

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
)

// DragDropEngine provides drag-and-drop coordination for kanban-style task management interfaces.
// It bridges Fyne's basic draggable interface with complex inter-container movement requirements.
type DragDropEngine interface {
	Drag() IDrag
	Drop() IDrop
	Visualize() IVisualize
}

// IDrag handles drag operation lifecycle management
type IDrag interface {
	StartDrag(widget fyne.CanvasObject, params DragParams) (DragHandle, error)
	UpdateDragPosition(handle DragHandle, position fyne.Position) error
	CompleteDrag(handle DragHandle) (DropResult, error)
	CancelDrag(handle DragHandle) error
}

// IDrop manages drop zone registration and validation
type IDrop interface {
	RegisterDropZone(zone DropZoneSpec) (ZoneID, error)
	UnregisterDropZone(zoneID ZoneID) error
	ValidateDropTarget(position fyne.Position, dragContext DragContext) (bool, error)
	GetActiveZones() []DropZoneSpec
}

// IVisualize handles visual feedback for drag operations
type IVisualize interface {
	CreateDragIndicator(source fyne.CanvasObject) (fyne.CanvasObject, error)
	UpdateIndicatorPosition(indicator fyne.CanvasObject, position fyne.Position) error
	ShowDropFeedback(zoneID ZoneID, accepted bool) error
	CleanupVisuals(dragHandle DragHandle) error
}

// Data Types

// DragHandle uniquely identifies an active drag operation
type DragHandle string

// ZoneID uniquely identifies a registered drop zone
type ZoneID string

// DragParams contains parameters for initiating a drag operation
type DragParams struct {
	SourceWidget fyne.CanvasObject
	DragType     DragType
	Metadata     map[string]any
}

// DragType categorizes the type of drag operation
type DragType string

const (
	DragTypeTask    DragType = "task"
	DragTypeSubtask DragType = "subtask"
)

// DropZoneSpec defines a drop zone configuration
type DropZoneSpec struct {
	ID           ZoneID
	Bounds       fyne.Position // Top-left corner
	Size         fyne.Size
	AcceptTypes  []DragType
	Container    fyne.CanvasObject
	AcceptanceFn func(DragContext) bool // Optional custom acceptance logic
}

// DragContext provides context information for drag operations
type DragContext struct {
	Handle       DragHandle
	DragType     DragType
	SourceWidget fyne.CanvasObject
	CurrentPos   fyne.Position
	Metadata     map[string]any
}

// DropResult contains the result of a drop operation
type DropResult struct {
	Success    bool
	TargetZone ZoneID
	Position   fyne.Position
	Error      error
}

// Implementation

type dragDropEngine struct {
	drag      IDrag
	drop      IDrop
	visualize IVisualize
}

// NewDragDropEngine creates a new DragDropEngine instance
func NewDragDropEngine() DragDropEngine {
	dragMgr := &dragManager{
		activeDrags: make(map[DragHandle]*activeDrag),
		nextHandle:  1,
	}

	dropMgr := &dropManager{
		zones: make(map[ZoneID]DropZoneSpec),
	}

	visualMgr := &visualManager{
		indicators: make(map[DragHandle]fyne.CanvasObject),
	}

	return &dragDropEngine{
		drag:      dragMgr,
		drop:      dropMgr,
		visualize: visualMgr,
	}
}

func (e *dragDropEngine) Drag() IDrag {
	return e.drag
}

func (e *dragDropEngine) Drop() IDrop {
	return e.drop
}

func (e *dragDropEngine) Visualize() IVisualize {
	return e.visualize
}

// Drag Manager Implementation

type activeDrag struct {
	handle     DragHandle
	context    DragContext
	indicator  fyne.CanvasObject
	startPos   fyne.Position
	currentPos fyne.Position
}

type dragManager struct {
	mu          sync.RWMutex
	activeDrags map[DragHandle]*activeDrag
	nextHandle  int
}

func (d *dragManager) StartDrag(widget fyne.CanvasObject, params DragParams) (DragHandle, error) {
	if widget == nil {
		return "", fmt.Errorf("widget cannot be nil")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	handle := DragHandle(fmt.Sprintf("drag-%d", d.nextHandle))
	d.nextHandle++

	context := DragContext{
		Handle:       handle,
		DragType:     params.DragType,
		SourceWidget: widget,
		CurrentPos:   widget.Position(),
		Metadata:     params.Metadata,
	}

	d.activeDrags[handle] = &activeDrag{
		handle:     handle,
		context:    context,
		startPos:   widget.Position(),
		currentPos: widget.Position(),
	}

	return handle, nil
}

func (d *dragManager) UpdateDragPosition(handle DragHandle, position fyne.Position) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	drag, exists := d.activeDrags[handle]
	if !exists {
		return fmt.Errorf("drag handle %s not found", handle)
	}

	drag.currentPos = position
	drag.context.CurrentPos = position

	return nil
}

func (d *dragManager) CompleteDrag(handle DragHandle) (DropResult, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	drag, exists := d.activeDrags[handle]
	if !exists {
		return DropResult{}, fmt.Errorf("drag handle %s not found", handle)
	}

	// Clean up the drag operation
	delete(d.activeDrags, handle)

	return DropResult{
		Success:  true,
		Position: drag.currentPos,
	}, nil
}

func (d *dragManager) CancelDrag(handle DragHandle) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	drag, exists := d.activeDrags[handle]
	if !exists {
		return fmt.Errorf("drag handle %s not found", handle)
	}

	// Restore original position
	if drag.context.SourceWidget != nil {
		drag.context.SourceWidget.Move(drag.startPos)
	}

	// Clean up the drag operation
	delete(d.activeDrags, handle)

	return nil
}


// Drop Manager Implementation

type dropManager struct {
	mu    sync.RWMutex
	zones map[ZoneID]DropZoneSpec
}

func (d *dropManager) RegisterDropZone(zone DropZoneSpec) (ZoneID, error) {
	if zone.ID == "" {
		return "", fmt.Errorf("zone ID cannot be empty")
	}

	if zone.Size.Width <= 0 || zone.Size.Height <= 0 {
		return "", fmt.Errorf("zone size must be positive")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.zones[zone.ID] = zone
	return zone.ID, nil
}

func (d *dropManager) UnregisterDropZone(zoneID ZoneID) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.zones[zoneID]; !exists {
		return fmt.Errorf("zone %s not found", zoneID)
	}

	delete(d.zones, zoneID)
	return nil
}

func (d *dropManager) ValidateDropTarget(position fyne.Position, dragContext DragContext) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, zone := range d.zones {
		if d.isPositionInZone(position, zone) {
			// Check if zone accepts this drag type
			if !d.zoneAcceptsDragType(zone, dragContext.DragType) {
				continue
			}

			// Check custom acceptance function if provided
			if zone.AcceptanceFn != nil && !zone.AcceptanceFn(dragContext) {
				continue
			}

			return true, nil
		}
	}

	return false, nil
}

func (d *dropManager) GetActiveZones() []DropZoneSpec {
	d.mu.RLock()
	defer d.mu.RUnlock()

	zones := make([]DropZoneSpec, 0, len(d.zones))
	for _, zone := range d.zones {
		zones = append(zones, zone)
	}

	return zones
}

func (d *dropManager) isPositionInZone(pos fyne.Position, zone DropZoneSpec) bool {
	return pos.X >= zone.Bounds.X &&
		pos.Y >= zone.Bounds.Y &&
		pos.X <= zone.Bounds.X+zone.Size.Width &&
		pos.Y <= zone.Bounds.Y+zone.Size.Height
}

func (d *dropManager) zoneAcceptsDragType(zone DropZoneSpec, dragType DragType) bool {
	if len(zone.AcceptTypes) == 0 {
		return true // Accept all types if none specified
	}

	// Check if dragType is in the AcceptTypes slice
	for _, acceptedType := range zone.AcceptTypes {
		if acceptedType == dragType {
			return true
		}
	}

	return false
}


// Visual Manager Implementation

type visualManager struct {
	mu         sync.RWMutex
	indicators map[DragHandle]fyne.CanvasObject
}

func (v *visualManager) CreateDragIndicator(source fyne.CanvasObject) (fyne.CanvasObject, error) {
	if source == nil {
		return nil, fmt.Errorf("source widget cannot be nil")
	}

	// Create a simple visual indicator
	// In a real implementation, this might clone the source widget's appearance
	indicator := source // Simplified - reuse the source widget as indicator

	return indicator, nil
}

func (v *visualManager) UpdateIndicatorPosition(indicator fyne.CanvasObject, position fyne.Position) error {
	if indicator == nil {
		return fmt.Errorf("indicator cannot be nil")
	}

	indicator.Move(position)
	return nil
}

func (v *visualManager) ShowDropFeedback(zoneID ZoneID, accepted bool) error {
	// Visual feedback implementation would go here
	// This might involve highlighting drop zones, changing cursor, etc.
	return nil
}

func (v *visualManager) CleanupVisuals(dragHandle DragHandle) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if indicator, exists := v.indicators[dragHandle]; exists {
		// Hide or remove the visual indicator
		indicator.Hide()
		delete(v.indicators, dragHandle)
	}

	return nil
}
