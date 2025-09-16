// Package utilities provides Client Utility layer components for the EisenKan system following iDesign methodology.
// This package contains reusable client-side components that provide infrastructure services across all client layers.
// Following iDesign namespace: eisenkan.Client.Utilities
package utilities

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Core type definitions for FyneUtility configuration
type (
	// Button styling and behavior options
	ButtonStyle int
	ButtonSize  int
	IconPosition int

	// Layout and container options
	LayoutType int

	// Theme and styling options
	ThemeType int

	// Window management options
	WindowSize struct {
		Width  float32
		Height float32
	}

	// Input validation state
	ValidationState int

	// Resource management
	IconSize int
)

// Enumeration constants for styling and configuration
const (
	// Button styles - following Material Design principles
	PrimaryButton ButtonStyle = iota
	SecondaryButton
	DestructiveButton
	TextButton
	OutlineButton

	// Button sizes
	ButtonSizeSmall ButtonSize = iota
	ButtonSizeMedium
	ButtonSizeLarge

	// Icon positions
	IconLeft IconPosition = iota
	IconRight
	IconTop
	IconBottom

	// Layout types for containers
	VBoxLayout LayoutType = iota
	HBoxLayout
	GridLayout
	BorderLayout
	FlowLayout

	// Theme types
	DefaultTheme ThemeType = iota
	DarkTheme
	LightTheme
	HighContrastTheme

	// Validation states for input widgets
	ValidationNeutral ValidationState = iota
	ValidationValid
	ValidationInvalid
	ValidationPending

	// Icon sizes
	IconSizeSmall IconSize = iota
	IconSizeMedium
	IconSizeLarge
	IconSizeExtraLarge
)

// Configuration structures for advanced options
type (
	// ButtonConfig holds complete button configuration
	ButtonConfig struct {
		Style       ButtonStyle
		Size        ButtonSize
		Icon        fyne.Resource
		IconPos     IconPosition
		OnTap       func()
		Disabled    bool
		Important   bool
		Validator   func() bool
		Tooltip     string
	}

	// EntryConfig holds enhanced entry widget configuration
	EntryConfig struct {
		Placeholder string
		Password    bool
		MultiLine   bool
		Validator   func(string) ValidatorResult
		Formatter   func(string) string
		OnChanged   func(string)
		OnSubmitted func(string)
		Disabled    bool
		Required    bool
	}

	// LabelConfig holds label styling configuration
	LabelConfig struct {
		Alignment fyne.TextAlign
		TextStyle fyne.TextStyle
		Color     color.Color
		Wrapping  fyne.TextWrap
		Truncation fyne.TextTruncation
	}

	// ContainerConfig holds container styling and layout configuration
	ContainerConfig struct {
		Layout    LayoutType
		Spacing   float32
		Padding   float32
		Objects   []fyne.CanvasObject
		Hidden    bool
	}

	// GridConfig holds grid layout specific configuration
	GridConfig struct {
		Rows      int
		Cols      int
		Spacing   float32
		Padding   float32
	}

	// WindowConfig holds window creation and management configuration
	WindowConfig struct {
		Title       string
		Size        WindowSize
		Resizable   bool
		FixedSize   bool
		CenterOnScreen bool
		Icon        fyne.Resource
		OnClose     func()
		MainWindow  bool
	}

	// ThemeConfig holds theme application configuration
	ThemeConfig struct {
		Type         ThemeType
		PrimaryColor color.Color
		AccentColor  color.Color
		CustomTheme  fyne.Theme
	}

	// DialogConfig holds dialog creation configuration
	DialogConfig struct {
		Title    string
		Message  string
		Icon     fyne.Resource
		Modal    bool
		OnSubmit func(bool)
		Buttons  []string
	}
)

// Resource cache for efficient resource management
type resourceCache struct {
	mu    sync.RWMutex
	cache map[string]fyne.Resource
	maxSize int
}

var (
	// Global resource cache for efficient resource management
	globalResourceCache = &resourceCache{
		cache:   make(map[string]fyne.Resource),
		maxSize: 1000, // Limit cache to 1000 resources
	}
)

// Functional option types for configuration
type (
	ButtonOption    func(*ButtonConfig)
	EntryOption     func(*EntryConfig)
	LabelOption     func(*LabelConfig)
	ContainerOption func(*ContainerConfig)
	GridOption      func(*GridConfig)
	WindowOption    func(*WindowConfig)
	ThemeOption     func(*ThemeConfig)
	DialogOption    func(*DialogConfig)
)

// =============================================================================
// Widget Creation Operations (REQ-WIDGET-001 to REQ-WIDGET-003)
// =============================================================================

// CreateButton creates a styled button with optional configuration
func CreateButton(text string, options ...ButtonOption) *widget.Button {
	config := &ButtonConfig{
		Style: PrimaryButton,
		Size:  ButtonSizeMedium,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Create base button
	button := widget.NewButton(text, config.OnTap)

	// Apply styling based on configuration
	applyButtonStyle(button, config)

	return button
}

// CreateEntry creates an enhanced entry widget with validation and formatting
func CreateEntry(placeholder string, options ...EntryOption) *widget.Entry {
	config := &EntryConfig{
		Placeholder: placeholder,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Create appropriate entry type
	var entry *widget.Entry
	if config.Password {
		entry = widget.NewPasswordEntry()
	} else if config.MultiLine {
		entry = widget.NewMultiLineEntry()
	} else {
		entry = widget.NewEntry()
	}

	// Configure entry
	entry.SetPlaceHolder(config.Placeholder)

	if config.OnChanged != nil {
		entry.OnChanged = config.OnChanged
	}

	if config.OnSubmitted != nil {
		entry.OnSubmitted = config.OnSubmitted
	}

	if config.Disabled {
		entry.Disable()
	}

	// Add validation if specified
	if config.Validator != nil {
		entry.Validator = func(text string) error {
			result := config.Validator(text)
			if !result.Valid {
				if len(result.Errors) > 0 {
					return fmt.Errorf("%s", result.Errors[0])
				}
				return fmt.Errorf("validation failed")
			}
			return nil
		}
	}

	return entry
}

// CreateLabel creates a styled label with optional configuration
func CreateLabel(text string, options ...LabelOption) *widget.Label {
	config := &LabelConfig{
		Alignment: fyne.TextAlignLeading,
		Wrapping:  fyne.TextWrapOff,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Create label
	label := widget.NewLabel(text)
	label.Alignment = config.Alignment
	label.Wrapping = config.Wrapping
	label.Truncation = config.Truncation
	label.TextStyle = config.TextStyle

	return label
}

// =============================================================================
// Layout Management Operations (REQ-LAYOUT-001 to REQ-LAYOUT-003)
// =============================================================================

// CreateContainer creates a responsive container with specified layout
func CreateContainer(layout LayoutType, options ...ContainerOption) *fyne.Container {
	config := &ContainerConfig{
		Layout:  layout,
		Spacing: 4, // Default spacing
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Create container with appropriate layout
	var cont *fyne.Container

	switch config.Layout {
	case VBoxLayout:
		cont = container.NewVBox(config.Objects...)
	case HBoxLayout:
		cont = container.NewHBox(config.Objects...)
	case GridLayout:
		// Grid requires additional configuration
		cont = container.NewGridWithColumns(2, config.Objects...) // Default 2 columns
	case BorderLayout:
		// Border layout requires specific object arrangement
		cont = container.NewBorder(nil, nil, nil, nil, config.Objects...)
	default:
		cont = container.NewVBox(config.Objects...)
	}

	if config.Hidden {
		cont.Hide()
	}

	return cont
}

// CreateGridContainer creates a grid-based container with specified dimensions
func CreateGridContainer(rows, cols int, options ...GridOption) *fyne.Container {
	config := &GridConfig{
		Rows:    rows,
		Cols:    cols,
		Spacing: 4,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Validate grid dimensions
	if config.Rows <= 0 {
		config.Rows = 1
	}
	if config.Cols <= 0 {
		config.Cols = 1
	}

	// Create grid container
	cont := container.NewGridWithColumns(config.Cols)

	return cont
}

// CreateBorderContainer creates a border layout container
func CreateBorderContainer(center fyne.CanvasObject, options ...ContainerOption) *fyne.Container {
	config := &ContainerConfig{}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Create border container with center object
	cont := container.NewBorder(nil, nil, nil, nil, center)

	return cont
}

// =============================================================================
// Theme and Styling Operations (REQ-THEME-001 to REQ-THEME-003)
// =============================================================================

// ApplyTheme applies theme configuration to a canvas object
func ApplyTheme(obj fyne.CanvasObject, options ...ThemeOption) error {
	config := &ThemeConfig{
		Type: DefaultTheme,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Theme application is handled at the app level in Fyne
	// Individual objects inherit from app theme
	// This function can be extended for custom styling

	return nil
}

// GetStandardTheme returns a standard theme configuration
func GetStandardTheme(themeType ThemeType) fyne.Theme {
	switch themeType {
	case DarkTheme:
		return theme.DarkTheme()
	case LightTheme:
		return theme.DefaultTheme()
	default:
		return theme.DefaultTheme()
	}
}

// =============================================================================
// Resource Management Operations (REQ-RESOURCE-001 to REQ-RESOURCE-003)
// =============================================================================

// LoadIcon loads an icon resource with caching
func LoadIcon(name string, size IconSize) (fyne.Resource, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s_%d", name, size)

	globalResourceCache.mu.RLock()
	if resource, exists := globalResourceCache.cache[cacheKey]; exists {
		globalResourceCache.mu.RUnlock()
		return resource, nil
	}
	globalResourceCache.mu.RUnlock()

	// For now, return theme icons as examples
	// In a real implementation, this would load from files or embedded resources
	var resource fyne.Resource

	switch name {
	case "save":
		resource = theme.DocumentSaveIcon()
	case "open":
		resource = theme.FolderOpenIcon()
	case "delete":
		resource = theme.DeleteIcon()
	case "settings":
		resource = theme.SettingsIcon()
	case "search":
		resource = theme.SearchIcon()
	default:
		return nil, fmt.Errorf("icon '%s' not found", name)
	}

	// Cache the resource
	globalResourceCache.mu.Lock()
	if len(globalResourceCache.cache) >= globalResourceCache.maxSize {
		// Simple cache eviction - remove oldest entry
		for k := range globalResourceCache.cache {
			delete(globalResourceCache.cache, k)
			break
		}
	}
	globalResourceCache.cache[cacheKey] = resource
	globalResourceCache.mu.Unlock()

	return resource, nil
}

// LoadImage loads an image resource with caching
func LoadImage(path string) (fyne.Resource, error) {
	// Check cache first
	globalResourceCache.mu.RLock()
	if resource, exists := globalResourceCache.cache[path]; exists {
		globalResourceCache.mu.RUnlock()
		return resource, nil
	}
	globalResourceCache.mu.RUnlock()

	// Load image from path
	// In a real implementation, this would use fyne's storage or file loading
	// For now, return a placeholder
	return nil, fmt.Errorf("image loading not implemented for path: %s", path)
}

// GetCachedResource retrieves a resource from cache
func GetCachedResource(key string) (fyne.Resource, bool) {
	globalResourceCache.mu.RLock()
	defer globalResourceCache.mu.RUnlock()

	resource, exists := globalResourceCache.cache[key]
	return resource, exists
}

// =============================================================================
// Window Management Operations (REQ-WINDOW-001 to REQ-WINDOW-003)
// =============================================================================

// CreateWindow creates a new window with specified configuration
func CreateWindow(title string, options ...WindowOption) fyne.Window {
	config := &WindowConfig{
		Title:     title,
		Size:      WindowSize{Width: 800, Height: 600},
		Resizable: true,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	// Create application if needed (this would typically be done once)
	a := app.New()

	// Create window
	var window fyne.Window
	if config.MainWindow {
		window = a.NewWindow(config.Title)
	} else {
		window = a.NewWindow(config.Title)
	}

	// Configure window
	if config.FixedSize {
		window.SetFixedSize(true)
	}

	window.Resize(fyne.NewSize(config.Size.Width, config.Size.Height))

	if config.CenterOnScreen {
		window.CenterOnScreen()
	}

	if config.Icon != nil {
		window.SetIcon(config.Icon)
	}

	if config.OnClose != nil {
		window.SetCloseIntercept(config.OnClose)
	}

	return window
}

// CenterWindow centers a window on the screen
func CenterWindow(window fyne.Window) error {
	if window == nil {
		return fmt.Errorf("window cannot be nil")
	}

	window.CenterOnScreen()
	return nil
}

// =============================================================================
// Event Handling Operations (REQ-EVENT-001 to REQ-EVENT-003)
// =============================================================================

// BindEventHandler is a placeholder for event binding utilities
// Fyne handles events through callbacks on widgets directly
func BindEventHandler(widget fyne.Widget, eventType string, handler func()) error {
	// Event binding in Fyne is done through widget-specific callbacks
	// This function serves as a placeholder for standardized event handling
	return fmt.Errorf("event binding not implemented for widget type %T", widget)
}

// =============================================================================
// Dialog Management Operations (REQ-DIALOG-001 to REQ-DIALOG-003)
// =============================================================================

// ShowInfoDialog displays an information dialog
func ShowInfoDialog(title, message string, parent fyne.Window, options ...DialogOption) {
	config := &DialogConfig{
		Title:   title,
		Message: message,
		Modal:   true,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	info := dialog.NewInformation(config.Title, config.Message, parent)
	info.Show()
}

// ShowErrorDialog displays an error dialog
func ShowErrorDialog(title, message string, parent fyne.Window, options ...DialogOption) {
	config := &DialogConfig{
		Title:   title,
		Message: message,
		Modal:   true,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	err := dialog.NewError(fmt.Errorf("%s", config.Message), parent)
	err.Show()
}

// ShowConfirmDialog displays a confirmation dialog
func ShowConfirmDialog(title, message string, callback func(bool), parent fyne.Window, options ...DialogOption) {
	config := &DialogConfig{
		Title:    title,
		Message:  message,
		Modal:    true,
		OnSubmit: callback,
	}

	// Apply functional options
	for _, option := range options {
		option(config)
	}

	confirm := dialog.NewConfirm(config.Title, config.Message, config.OnSubmit, parent)
	confirm.Show()
}

// =============================================================================
// Input Enhancement Operations (REQ-INPUT-001 to REQ-INPUT-003)
// =============================================================================

// CreateValidatedEntry creates an entry widget with integrated validation feedback
func CreateValidatedEntry(placeholder string, validator func(string) ValidatorResult, options ...EntryOption) *widget.Entry {
	// Use CreateEntry with validation
	opts := append(options, WithEntryValidator(validator))
	return CreateEntry(placeholder, opts...)
}

// CreateNumericEntry creates a numeric input entry widget
func CreateNumericEntry(min, max float64, precision int, options ...EntryOption) *widget.Entry {
	validator := func(text string) ValidatorResult {
		// Implement numeric validation using ValidationUtility
		// This would integrate with existing ValidateNumber function
		result := ValidatorResult{Valid: true}

		// Basic validation - in real implementation would use ValidationUtility
		if text == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "value is required")
		}

		return result
	}

	opts := append(options, WithEntryValidator(validator))
	return CreateEntry("Enter number", opts...)
}

// CreateDateEntry creates a date input entry widget
func CreateDateEntry(format string, minDate, maxDate time.Time, options ...EntryOption) *widget.Entry {
	validator := func(text string) ValidatorResult {
		// Implement date validation using ValidationUtility
		// This would integrate with existing ValidateDate function
		result := ValidatorResult{Valid: true}

		// Basic validation - in real implementation would use ValidationUtility
		if text == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "date is required")
		}

		return result
	}

	opts := append(options, WithEntryValidator(validator))
	return CreateEntry("Enter date", opts...)
}

// =============================================================================
// Functional Option Implementations
// =============================================================================

// Button option functions
func WithButtonStyle(style ButtonStyle) ButtonOption {
	return func(config *ButtonConfig) {
		config.Style = style
	}
}

func WithButtonIcon(icon fyne.Resource) ButtonOption {
	return func(config *ButtonConfig) {
		config.Icon = icon
	}
}

func WithButtonCallback(callback func()) ButtonOption {
	return func(config *ButtonConfig) {
		config.OnTap = callback
	}
}

func WithButtonDisabled(disabled bool) ButtonOption {
	return func(config *ButtonConfig) {
		config.Disabled = disabled
	}
}

func WithButtonTooltip(tooltip string) ButtonOption {
	return func(config *ButtonConfig) {
		config.Tooltip = tooltip
	}
}

// Entry option functions
func WithEntryValidator(validator func(string) ValidatorResult) EntryOption {
	return func(config *EntryConfig) {
		config.Validator = validator
	}
}

func WithEntryFormatter(formatter func(string) string) EntryOption {
	return func(config *EntryConfig) {
		config.Formatter = formatter
	}
}

func WithEntryPassword(password bool) EntryOption {
	return func(config *EntryConfig) {
		config.Password = password
	}
}

func WithEntryMultiLine(multiLine bool) EntryOption {
	return func(config *EntryConfig) {
		config.MultiLine = multiLine
	}
}

func WithEntryOnChanged(callback func(string)) EntryOption {
	return func(config *EntryConfig) {
		config.OnChanged = callback
	}
}

func WithEntryDisabled(disabled bool) EntryOption {
	return func(config *EntryConfig) {
		config.Disabled = disabled
	}
}

// Label option functions
func WithLabelAlignment(alignment fyne.TextAlign) LabelOption {
	return func(config *LabelConfig) {
		config.Alignment = alignment
	}
}

func WithLabelStyle(style fyne.TextStyle) LabelOption {
	return func(config *LabelConfig) {
		config.TextStyle = style
	}
}

func WithLabelWrapping(wrapping fyne.TextWrap) LabelOption {
	return func(config *LabelConfig) {
		config.Wrapping = wrapping
	}
}

func WithLabelColor(color color.Color) LabelOption {
	return func(config *LabelConfig) {
		config.Color = color
	}
}

// Container option functions
func WithContainerSpacing(spacing float32) ContainerOption {
	return func(config *ContainerConfig) {
		config.Spacing = spacing
	}
}

func WithContainerPadding(padding float32) ContainerOption {
	return func(config *ContainerConfig) {
		config.Padding = padding
	}
}

func WithContainerObjects(objects ...fyne.CanvasObject) ContainerOption {
	return func(config *ContainerConfig) {
		config.Objects = objects
	}
}

func WithContainerHidden(hidden bool) ContainerOption {
	return func(config *ContainerConfig) {
		config.Hidden = hidden
	}
}

// Window option functions
func WithWindowSize(width, height float32) WindowOption {
	return func(config *WindowConfig) {
		config.Size = WindowSize{Width: width, Height: height}
	}
}

func WithWindowResizable(resizable bool) WindowOption {
	return func(config *WindowConfig) {
		config.Resizable = resizable
	}
}

func WithWindowCentered(centered bool) WindowOption {
	return func(config *WindowConfig) {
		config.CenterOnScreen = centered
	}
}

func WithWindowIcon(icon fyne.Resource) WindowOption {
	return func(config *WindowConfig) {
		config.Icon = icon
	}
}

func WithWindowOnClose(callback func()) WindowOption {
	return func(config *WindowConfig) {
		config.OnClose = callback
	}
}

// =============================================================================
// Internal Helper Functions
// =============================================================================

// applyButtonStyle applies styling to a button based on configuration
func applyButtonStyle(button *widget.Button, config *ButtonConfig) {
	// Set icon if provided
	if config.Icon != nil {
		button.SetIcon(config.Icon)
	}

	// Apply disabled state
	if config.Disabled {
		button.Disable()
	}

	// Style application in Fyne is primarily handled through themes
	// Button-specific styling would be implemented here based on style type
	switch config.Style {
	case PrimaryButton:
		button.Importance = widget.HighImportance
	case SecondaryButton:
		button.Importance = widget.MediumImportance
	case DestructiveButton:
		button.Importance = widget.DangerImportance
	case TextButton:
		button.Importance = widget.LowImportance
	}
}