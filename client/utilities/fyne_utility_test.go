// Package utilities provides Client Utility layer components for the EisenKan system following iDesign methodology.
// This file contains comprehensive unit tests for FyneUtility following the destructive testing strategy from FyneUtility_STP.md
package utilities

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TestUnit_CreateButton tests basic button creation functionality
func TestUnit_CreateButton(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		options      []ButtonOption
		expectText   string
		expectStyle  ButtonStyle
		expectIcon   bool
		expectError  bool
	}{
		{
			name:        "simple button creation",
			text:        "Click Me",
			options:     []ButtonOption{},
			expectText:  "Click Me",
			expectStyle: PrimaryButton,
			expectIcon:  false,
		},
		{
			name: "button with primary style",
			text: "Save",
			options: []ButtonOption{
				WithButtonStyle(PrimaryButton),
			},
			expectText:  "Save",
			expectStyle: PrimaryButton,
			expectIcon:  false,
		},
		{
			name: "button with icon",
			text: "Save",
			options: []ButtonOption{
				WithButtonIcon(theme.DocumentSaveIcon()),
			},
			expectText: "Save",
			expectIcon: true,
		},
		{
			name: "button with multiple options",
			text: "Delete",
			options: []ButtonOption{
				WithButtonStyle(DestructiveButton),
				WithButtonIcon(theme.DeleteIcon()),
				WithButtonDisabled(true),
			},
			expectText:  "Delete",
			expectStyle: DestructiveButton,
			expectIcon:  true,
		},
		{
			name:        "empty text button",
			text:        "",
			options:     []ButtonOption{},
			expectText:  "",
			expectStyle: PrimaryButton,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := CreateButton(tt.text, tt.options...)

			if button == nil {
				t.Error("CreateButton() returned nil")
				return
			}

			if button.Text != tt.expectText {
				t.Errorf("CreateButton() text = %v, want %v", button.Text, tt.expectText)
			}

			if tt.expectIcon && button.Icon == nil {
				t.Error("CreateButton() expected icon but got nil")
			}

			if !tt.expectIcon && button.Icon != nil {
				t.Error("CreateButton() expected no icon but got one")
			}
		})
	}
}

// TestUnit_CreateEntry tests entry widget creation functionality
func TestUnit_CreateEntry(t *testing.T) {
	tests := []struct {
		name            string
		placeholder     string
		options         []EntryOption
		expectPlaceholder string
		expectPassword  bool
		expectMultiLine bool
		expectDisabled  bool
	}{
		{
			name:              "simple entry creation",
			placeholder:       "Enter text",
			options:           []EntryOption{},
			expectPlaceholder: "Enter text",
			expectPassword:    false,
			expectMultiLine:   false,
			expectDisabled:    false,
		},
		{
			name:        "password entry",
			placeholder: "Enter password",
			options: []EntryOption{
				WithEntryPassword(true),
			},
			expectPlaceholder: "Enter password",
			expectPassword:    true,
		},
		{
			name:        "multiline entry",
			placeholder: "Enter description",
			options: []EntryOption{
				WithEntryMultiLine(true),
			},
			expectPlaceholder: "Enter description",
			expectMultiLine:   true,
		},
		{
			name:        "disabled entry",
			placeholder: "Disabled field",
			options: []EntryOption{
				WithEntryDisabled(true),
			},
			expectPlaceholder: "Disabled field",
			expectDisabled:    true,
		},
		{
			name:              "empty placeholder",
			placeholder:       "",
			options:           []EntryOption{},
			expectPlaceholder: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CreateEntry(tt.placeholder, tt.options...)

			if entry == nil {
				t.Error("CreateEntry() returned nil")
				return
			}

			if entry.PlaceHolder != tt.expectPlaceholder {
				t.Errorf("CreateEntry() placeholder = %v, want %v", entry.PlaceHolder, tt.expectPlaceholder)
			}

			if tt.expectPassword && !entry.Password {
				t.Error("CreateEntry() expected password entry but got regular entry")
			}

			if tt.expectMultiLine && !entry.MultiLine {
				t.Error("CreateEntry() expected multiline entry but got single line")
			}

			if tt.expectDisabled && !entry.Disabled() {
				t.Error("CreateEntry() expected disabled entry but got enabled")
			}
		})
	}
}

// TestUnit_CreateLabel tests label creation functionality
func TestUnit_CreateLabel(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		options      []LabelOption
		expectText   string
		expectAlign  fyne.TextAlign
		expectWrap   fyne.TextWrap
	}{
		{
			name:        "simple label creation",
			text:        "Hello World",
			options:     []LabelOption{},
			expectText:  "Hello World",
			expectAlign: fyne.TextAlignLeading,
			expectWrap:  fyne.TextWrapOff,
		},
		{
			name: "centered label",
			text: "Centered Text",
			options: []LabelOption{
				WithLabelAlignment(fyne.TextAlignCenter),
			},
			expectText:  "Centered Text",
			expectAlign: fyne.TextAlignCenter,
		},
		{
			name: "wrapped label",
			text: "Long text that should wrap",
			options: []LabelOption{
				WithLabelWrapping(fyne.TextWrapWord),
			},
			expectText: "Long text that should wrap",
			expectWrap: fyne.TextWrapWord,
		},
		{
			name:        "empty text label",
			text:        "",
			options:     []LabelOption{},
			expectText:  "",
			expectAlign: fyne.TextAlignLeading,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := CreateLabel(tt.text, tt.options...)

			if label == nil {
				t.Error("CreateLabel() returned nil")
				return
			}

			if label.Text != tt.expectText {
				t.Errorf("CreateLabel() text = %v, want %v", label.Text, tt.expectText)
			}

			if label.Alignment != tt.expectAlign {
				t.Errorf("CreateLabel() alignment = %v, want %v", label.Alignment, tt.expectAlign)
			}

			if label.Wrapping != tt.expectWrap {
				t.Errorf("CreateLabel() wrapping = %v, want %v", label.Wrapping, tt.expectWrap)
			}
		})
	}
}

// TestUnit_CreateContainer tests container creation functionality
func TestUnit_CreateContainer(t *testing.T) {
	tests := []struct {
		name           string
		layout         LayoutType
		options        []ContainerOption
		expectLayout   LayoutType
		expectObjects  int
		expectHidden   bool
	}{
		{
			name:          "vbox container",
			layout:        VBoxLayout,
			options:       []ContainerOption{},
			expectLayout:  VBoxLayout,
			expectObjects: 0,
			expectHidden:  false,
		},
		{
			name:          "hbox container",
			layout:        HBoxLayout,
			options:       []ContainerOption{},
			expectLayout:  HBoxLayout,
			expectObjects: 0,
		},
		{
			name: "container with objects",
			layout: VBoxLayout,
			options: []ContainerOption{
				WithContainerObjects(widget.NewLabel("Test"), widget.NewButton("Button", nil)),
			},
			expectLayout:  VBoxLayout,
			expectObjects: 2,
		},
		{
			name: "hidden container",
			layout: VBoxLayout,
			options: []ContainerOption{
				WithContainerHidden(true),
			},
			expectLayout: VBoxLayout,
			expectHidden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := CreateContainer(tt.layout, tt.options...)

			if container == nil {
				t.Error("CreateContainer() returned nil")
				return
			}

			if len(container.Objects) != tt.expectObjects {
				t.Errorf("CreateContainer() objects count = %v, want %v", len(container.Objects), tt.expectObjects)
			}

			if tt.expectHidden && container.Visible() {
				t.Error("CreateContainer() expected hidden container but got visible")
			}
		})
	}
}

// TestUnit_CreateGridContainer tests grid container creation
func TestUnit_CreateGridContainer(t *testing.T) {
	tests := []struct {
		name        string
		rows        int
		cols        int
		options     []GridOption
		expectRows  int
		expectCols  int
	}{
		{
			name:       "2x2 grid",
			rows:       2,
			cols:       2,
			options:    []GridOption{},
			expectRows: 2,
			expectCols: 2,
		},
		{
			name:       "invalid dimensions - zero rows",
			rows:       0,
			cols:       3,
			options:    []GridOption{},
			expectRows: 1, // Should default to 1
			expectCols: 3,
		},
		{
			name:       "invalid dimensions - negative cols",
			rows:       2,
			cols:       -1,
			options:    []GridOption{},
			expectRows: 2,
			expectCols: 1, // Should default to 1
		},
		{
			name:       "large grid",
			rows:       10,
			cols:       10,
			options:    []GridOption{},
			expectRows: 10,
			expectCols: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := CreateGridContainer(tt.rows, tt.cols, tt.options...)

			if container == nil {
				t.Error("CreateGridContainer() returned nil")
				return
			}

			// Grid container validation is primarily structural
			// Fyne's grid container doesn't expose rows/cols directly for validation
		})
	}
}

// TestUnit_LoadIcon tests icon loading and caching functionality
func TestUnit_LoadIcon(t *testing.T) {
	tests := []struct {
		name        string
		iconName    string
		size        IconSize
		expectError bool
		expectCache bool
	}{
		{
			name:        "load save icon",
			iconName:    "save",
			size:        IconSizeMedium,
			expectError: false,
			expectCache: true,
		},
		{
			name:        "load open icon",
			iconName:    "open",
			size:        IconSizeMedium,
			expectError: false,
			expectCache: true,
		},
		{
			name:        "load delete icon",
			iconName:    "delete",
			size:        IconSizeLarge,
			expectError: false,
			expectCache: true,
		},
		{
			name:        "load nonexistent icon",
			iconName:    "nonexistent",
			size:        IconSizeMedium,
			expectError: true,
			expectCache: false,
		},
		{
			name:        "load settings icon",
			iconName:    "settings",
			size:        IconSizeSmall,
			expectError: false,
			expectCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache before test
			globalResourceCache.mu.Lock()
			globalResourceCache.cache = make(map[string]fyne.Resource)
			globalResourceCache.mu.Unlock()

			resource, err := LoadIcon(tt.iconName, tt.size)

			if tt.expectError && err == nil {
				t.Error("LoadIcon() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("LoadIcon() unexpected error: %v", err)
			}

			if !tt.expectError && resource == nil {
				t.Error("LoadIcon() returned nil resource without error")
			}

			if tt.expectCache {
				// Check if resource is cached using same key format as LoadIcon
				cacheKey := fmt.Sprintf("%s_%d", tt.iconName, tt.size)
				_, cached := GetCachedResource(cacheKey)
				if !cached && !tt.expectError {
					t.Error("LoadIcon() expected resource to be cached but wasn't")
				}
			}
		})
	}
}

// TestUnit_ShowInfoDialog tests dialog creation functionality
func TestUnit_ShowInfoDialog(t *testing.T) {
	// Create test app and window
	testApp := test.NewApp()
	testWindow := testApp.NewWindow("Test")
	defer testApp.Quit()

	tests := []struct {
		name    string
		title   string
		message string
		options []DialogOption
	}{
		{
			name:    "simple info dialog",
			title:   "Information",
			message: "This is an info message",
			options: []DialogOption{},
		},
		{
			name:    "empty title dialog",
			title:   "",
			message: "Message with empty title",
			options: []DialogOption{},
		},
		{
			name:    "empty message dialog",
			title:   "Title",
			message: "",
			options: []DialogOption{},
		},
		{
			name:    "both empty dialog",
			title:   "",
			message: "",
			options: []DialogOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies that the function doesn't panic
			// Actual dialog display testing would require UI testing framework
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ShowInfoDialog() panicked: %v", r)
				}
			}()

			ShowInfoDialog(tt.title, tt.message, testWindow, tt.options...)
		})
	}
}

// TestUnit_CreateValidatedEntry tests validated entry creation
func TestUnit_CreateValidatedEntry(t *testing.T) {
	validator := func(text string) ValidatorResult {
		result := ValidatorResult{Valid: true}
		if text == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "value is required")
		}
		if len(text) < 3 {
			result.Valid = false
			result.Errors = append(result.Errors, "value must be at least 3 characters")
		}
		return result
	}

	tests := []struct {
		name         string
		placeholder  string
		testInput    string
		expectValid  bool
		expectErrors int
	}{
		{
			name:        "valid input",
			placeholder: "Enter text",
			testInput:   "hello",
			expectValid: true,
		},
		{
			name:         "empty input",
			placeholder:  "Enter text",
			testInput:    "",
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name:         "too short input",
			placeholder:  "Enter text",
			testInput:    "hi",
			expectValid:  false,
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CreateValidatedEntry(tt.placeholder, validator)

			if entry == nil {
				t.Error("CreateValidatedEntry() returned nil")
				return
			}

			if entry.PlaceHolder != tt.placeholder {
				t.Errorf("CreateValidatedEntry() placeholder = %v, want %v", entry.PlaceHolder, tt.placeholder)
			}

			// Test validation
			if entry.Validator != nil {
				err := entry.Validator(tt.testInput)

				if tt.expectValid && err != nil {
					t.Errorf("CreateValidatedEntry() validation failed for valid input: %v", err)
				}

				if !tt.expectValid && err == nil {
					t.Error("CreateValidatedEntry() validation passed for invalid input")
				}
			}
		})
	}
}

// TestUnit_CreateNumericEntry tests numeric entry creation
func TestUnit_CreateNumericEntry(t *testing.T) {
	tests := []struct {
		name      string
		min       float64
		max       float64
		precision int
	}{
		{
			name:      "integer range",
			min:       0,
			max:       100,
			precision: 0,
		},
		{
			name:      "decimal range",
			min:       0.0,
			max:       1.0,
			precision: 2,
		},
		{
			name:      "negative range",
			min:       -100.0,
			max:       100.0,
			precision: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CreateNumericEntry(tt.min, tt.max, tt.precision)

			if entry == nil {
				t.Error("CreateNumericEntry() returned nil")
				return
			}

			// Basic validation that entry was created
			if entry.PlaceHolder == "" {
				t.Error("CreateNumericEntry() expected placeholder but got empty")
			}
		})
	}
}

// TestUnit_CreateDateEntry tests date entry creation
func TestUnit_CreateDateEntry(t *testing.T) {
	minDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		format  string
		minDate time.Time
		maxDate time.Time
	}{
		{
			name:    "ISO date format",
			format:  "2006-01-02",
			minDate: minDate,
			maxDate: maxDate,
		},
		{
			name:    "US date format",
			format:  "01/02/2006",
			minDate: minDate,
			maxDate: maxDate,
		},
		{
			name:    "European date format",
			format:  "02.01.2006",
			minDate: minDate,
			maxDate: maxDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := CreateDateEntry(tt.format, tt.minDate, tt.maxDate)

			if entry == nil {
				t.Error("CreateDateEntry() returned nil")
				return
			}

			// Basic validation that entry was created
			if entry.PlaceHolder == "" {
				t.Error("CreateDateEntry() expected placeholder but got empty")
			}
		})
	}
}

// Destructive test cases following STP guidelines

// TestUnit_FyneUtility_DestructiveWidgetCreation tests widget creation with invalid configurations
func TestUnit_FyneUtility_DestructiveWidgetCreation(t *testing.T) {
	tests := []struct {
		name    string
		testFn  func() bool
		expectPanic bool
	}{
		{
			name: "button with nil callback",
			testFn: func() bool {
				button := CreateButton("Test", WithButtonCallback(nil))
				return button != nil
			},
			expectPanic: false,
		},
		{
			name: "entry with nil validator",
			testFn: func() bool {
				entry := CreateEntry("Test", WithEntryValidator(nil))
				return entry != nil
			},
			expectPanic: false,
		},
		{
			name: "container with nil objects",
			testFn: func() bool {
				container := CreateContainer(VBoxLayout, WithContainerObjects(nil))
				return container != nil
			},
			expectPanic: false,
		},
		{
			name: "extremely long button text",
			testFn: func() bool {
				longText := strings.Repeat("A", 10000)
				button := CreateButton(longText)
				return button != nil && len(button.Text) > 0
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Function panicked unexpectedly: %v", r)
					}
				}
			}()

			result := tt.testFn()
			if !result && !tt.expectPanic {
				t.Error("Destructive test function returned false")
			}
		})
	}
}

// TestUnit_FyneUtility_ResourceCacheStress tests resource cache under stress
func TestUnit_FyneUtility_ResourceCacheStress(t *testing.T) {
	// Clear cache before test
	globalResourceCache.mu.Lock()
	globalResourceCache.cache = make(map[string]fyne.Resource)
	globalResourceCache.mu.Unlock()

	// Test cache overflow handling
	iconNames := []string{"save", "open", "delete", "settings", "search"}

	// Load many resources to test cache limits
	for i := 0; i < globalResourceCache.maxSize + 100; i++ {
		iconName := iconNames[i % len(iconNames)]
		size := IconSize(i % 4)

		_, err := LoadIcon(iconName, size)
		if err != nil {
			continue // Skip nonexistent icons
		}
	}

	// Verify cache size is limited
	globalResourceCache.mu.RLock()
	cacheSize := len(globalResourceCache.cache)
	globalResourceCache.mu.RUnlock()

	if cacheSize > globalResourceCache.maxSize {
		t.Errorf("Cache size %d exceeds maximum %d", cacheSize, globalResourceCache.maxSize)
	}
}

// TestUnit_FyneUtility_ConcurrentAccess tests thread safety
func TestUnit_FyneUtility_ConcurrentAccess(t *testing.T) {
	const numGoroutines = 50
	const numOperations = 20

	// Test concurrent widget creation
	t.Run("concurrent widget creation", func(t *testing.T) {
		results := make(chan bool, numGoroutines*numOperations)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numOperations; j++ {
					// Create various widgets concurrently
					button := CreateButton("Test")
					entry := CreateEntry("Test")
					label := CreateLabel("Test")
					container := CreateContainer(VBoxLayout)

					results <- (button != nil && entry != nil && label != nil && container != nil)
				}
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines*numOperations; i++ {
			if !<-results {
				t.Error("Concurrent widget creation failed")
			}
		}
	})

	// Test concurrent resource loading
	t.Run("concurrent resource loading", func(t *testing.T) {
		results := make(chan bool, numGoroutines*numOperations)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numOperations; j++ {
					iconNames := []string{"save", "open", "delete", "settings"}
					iconName := iconNames[j%len(iconNames)]

					_, err := LoadIcon(iconName, IconSizeMedium)
					results <- (err == nil)
				}
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numGoroutines*numOperations; i++ {
			if <-results {
				successCount++
			}
		}

		// Expect most operations to succeed
		expectedMin := (numGoroutines * numOperations) * 8 / 10 // 80% success rate
		if successCount < expectedMin {
			t.Errorf("Concurrent resource loading success rate too low: %d/%d", successCount, numGoroutines*numOperations)
		}
	})
}

// TestUnit_FyneUtility_ErrorHandling tests error handling and recovery
func TestUnit_FyneUtility_ErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		wantErr  bool
	}{
		{
			name: "center nil window",
			testFunc: func() error {
				return CenterWindow(nil)
			},
			wantErr: true,
		},
		{
			name: "load nonexistent icon",
			testFunc: func() error {
				_, err := LoadIcon("nonexistent_icon", IconSizeMedium)
				return err
			},
			wantErr: true,
		},
		{
			name: "load image with invalid path",
			testFunc: func() error {
				_, err := LoadImage("nonexistent/path/image.png")
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Performance benchmark tests for REQ-PERF-001 verification
func BenchmarkCreateButton(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateButton("Test Button")
	}
}

func BenchmarkCreateEntry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateEntry("Test Entry")
	}
}

func BenchmarkCreateContainer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateContainer(VBoxLayout)
	}
}

func BenchmarkLoadIcon(b *testing.B) {
	// Clear cache before benchmark
	globalResourceCache.mu.Lock()
	globalResourceCache.cache = make(map[string]fyne.Resource)
	globalResourceCache.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadIcon("save", IconSizeMedium)
	}
}

func BenchmarkLoadIconCached(b *testing.B) {
	// Warm up cache
	LoadIcon("save", IconSizeMedium)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadIcon("save", IconSizeMedium)
	}
}