package engines

import (
	"errors"
	"testing"
	"time"
)

// TestUnit_FormattingEngine_NewFormattingEngine tests the constructor
func TestUnit_FormattingEngine_NewFormattingEngine(t *testing.T) {
	engine := NewFormattingEngine()

	if engine == nil {
		t.Fatal("NewFormattingEngine returned nil")
	}

	if engine.Text() == nil {
		t.Error("Text facet is nil")
	}

	if engine.Number() == nil {
		t.Error("Number facet is nil")
	}

	if engine.Time() == nil {
		t.Error("Time facet is nil")
	}

	if engine.Datastructure() == nil {
		t.Error("Datastructure facet is nil")
	}

	if engine.Template() == nil {
		t.Error("Template facet is nil")
	}

	if engine.Locale() == nil {
		t.Error("Locale facet is nil")
	}

	// Verify default locale
	if engine.Locale().GetLocale() != "en-US" {
		t.Errorf("Expected default locale 'en-US', got '%s'", engine.Locale().GetLocale())
	}
}

// TestUnit_FormattingEngine_TextFacet tests the IText interface
func TestUnit_FormattingEngine_TextFacet(t *testing.T) {
	engine := NewFormattingEngine()
	textFacet := engine.Text()

	t.Run("FormatText with basic options", func(t *testing.T) {
		options := TextOptions{
			CaseTransform: CaseUpper,
			MaxLength:     10,
			TrimSpaces:    true,
		}

		result, err := textFacet.FormatText("  hello world  ", options)
		if err != nil {
			t.Fatalf("FormatText failed: %v", err)
		}

		expected := "HELLO W..."
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("FormatText with HTML escaping", func(t *testing.T) {
		options := TextOptions{
			EscapeHTML: true,
		}

		result, err := textFacet.FormatText("<script>alert('xss')</script>", options)
		if err != nil {
			t.Fatalf("FormatText failed: %v", err)
		}

		if result == "<script>alert('xss')</script>" {
			t.Error("HTML was not escaped")
		}
	})

	t.Run("FormatLabel converts field names", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"firstName", "First Name"},
			{"user_name", "User Name"},
			{"email-address", "Email Address"},
			{"ID", "ID"},
			{"", ""},
		}

		for _, tc := range testCases {
			result := textFacet.FormatLabel(tc.input)
			if result != tc.expected {
				t.Errorf("FormatLabel(%s): expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})

	t.Run("FormatMessage substitutes parameters", func(t *testing.T) {
		template := "Hello {name}, you have {count} messages"
		params := map[string]any{
			"name":  "Alice",
			"count": 5,
		}

		result, err := textFacet.FormatMessage(template, params)
		if err != nil {
			t.Fatalf("FormatMessage failed: %v", err)
		}

		expected := "Hello Alice, you have 5 messages"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("FormatError creates formatted error", func(t *testing.T) {
		err := errors.New("test error")
		formatted := textFacet.FormatError(err, FormattingSeverityError)

		if formatted.Message != "test error" {
			t.Errorf("Expected message 'test error', got '%s'", formatted.Message)
		}

		if formatted.Severity != FormattingSeverityError {
			t.Errorf("Expected severity %d, got %d", FormattingSeverityError, formatted.Severity)
		}

		if formatted.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
	})
}

// TestUnit_FormattingEngine_NumberFacet tests the INumber interface
func TestUnit_FormattingEngine_NumberFacet(t *testing.T) {
	engine := NewFormattingEngine()
	numberFacet := engine.Number()

	t.Run("FormatNumber with different types", func(t *testing.T) {
		testCases := []struct {
			value     any
			precision int
			shouldErr bool
		}{
			{123, 2, false},
			{123.456, 2, false},
			{int64(999), 0, false},
			{float32(3.14), 1, false},
			{"123.45", 2, false},
			{"invalid", 2, true},
			{[]int{1, 2, 3}, 2, true},
		}

		for _, tc := range testCases {
			result, err := numberFacet.FormatNumber(tc.value, tc.precision)
			if tc.shouldErr {
				if err == nil {
					t.Errorf("FormatNumber(%v) should have failed", tc.value)
				}
			} else {
				if err != nil {
					t.Errorf("FormatNumber(%v) failed: %v", tc.value, err)
				}
				if result == "" {
					t.Errorf("FormatNumber(%v) returned empty result", tc.value)
				}
			}
		}
	})

	t.Run("FormatPercentage", func(t *testing.T) {
		result := numberFacet.FormatPercentage(0.1234, 2)
		if !contains(result, "12.34") && !contains(result, "%") {
			t.Errorf("Expected percentage to contain percentage value, got '%s'", result)
		}
	})

	t.Run("FormatFileSize", func(t *testing.T) {
		testCases := []struct {
			bytes    int64
			unit     FileSizeUnit
			expected string
		}{
			{1024, FileSizeAuto, "1.00 KB"},
			{1024, FileSizeBytes, "1,024 bytes"},
			{1048576, FileSizeAuto, "1.00 MB"},
		}

		for _, tc := range testCases {
			result := numberFacet.FormatFileSize(tc.bytes, tc.unit)
			if !contains(result, "1.00") && !contains(result, "1024") {
				t.Errorf("FormatFileSize(%d, %d): unexpected result '%s'", tc.bytes, tc.unit, result)
			}
		}
	})

	t.Run("FormatCurrency", func(t *testing.T) {
		testCases := []struct {
			value    float64
			currency string
			symbol   string
		}{
			{123.45, "USD", "$"},
			{99.99, "EUR", "€"},
			{50.00, "GBP", "£"},
		}

		for _, tc := range testCases {
			result, err := numberFacet.FormatCurrency(tc.value, tc.currency)
			if err != nil {
				t.Errorf("FormatCurrency failed: %v", err)
			}
			if !contains(result, tc.symbol) {
				t.Errorf("Expected currency symbol '%s' in result '%s'", tc.symbol, result)
			}
		}
	})
}

// TestUnit_FormattingEngine_TimeFacet tests the ITime interface
func TestUnit_FormattingEngine_TimeFacet(t *testing.T) {
	engine := NewFormattingEngine()
	timeFacet := engine.Time()

	testTime := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)

	t.Run("FormatDateTime", func(t *testing.T) {
		result := timeFacet.FormatDateTime(testTime, "ISO")
		if result == "" {
			t.Error("FormatDateTime returned empty result")
		}
	})

	t.Run("FormatDuration", func(t *testing.T) {
		testCases := []struct {
			duration time.Duration
			contains string
		}{
			{30 * time.Second, "30s"},
			{5 * time.Minute, "5m"},
			{2 * time.Hour, "2h"},
			{25 * time.Hour, "1d"},
		}

		for _, tc := range testCases {
			result := timeFacet.FormatDuration(tc.duration)
			if !contains(result, tc.contains) {
				t.Errorf("FormatDuration(%v): expected to contain '%s', got '%s'", tc.duration, tc.contains, result)
			}
		}
	})

	t.Run("FormatRelativeTime", func(t *testing.T) {
		now := time.Now()
		pastTime := now.Add(-2 * time.Hour)
		futureTime := now.Add(3 * time.Hour)

		pastResult := timeFacet.FormatRelativeTime(pastTime)
		if !contains(pastResult, "ago") {
			t.Errorf("Expected past time to contain 'ago', got '%s'", pastResult)
		}

		futureResult := timeFacet.FormatRelativeTime(futureTime)
		if !contains(futureResult, "in") {
			t.Errorf("Expected future time to contain 'in', got '%s'", futureResult)
		}
	})

	t.Run("FormatTimeRange", func(t *testing.T) {
		start := testTime
		end := testTime.Add(2 * time.Hour)

		result := timeFacet.FormatTimeRange(start, end)
		if !contains(result, "-") {
			t.Errorf("Expected time range to contain '-', got '%s'", result)
		}
	})
}

// TestUnit_FormattingEngine_DatastructureFacet tests the IDatastructure interface
func TestUnit_FormattingEngine_DatastructureFacet(t *testing.T) {
	engine := NewFormattingEngine()
	dataFacet := engine.Datastructure()

	t.Run("FormatList", func(t *testing.T) {
		items := []any{"apple", "banana", "cherry"}
		options := ListOptions{
			Separator:   ", ",
			ShowIndices: false,
		}

		result := dataFacet.FormatList(items, options)
		expected := "apple, banana, cherry"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("FormatList with indices", func(t *testing.T) {
		items := []any{"first", "second"}
		options := ListOptions{
			ShowIndices: true,
			Separator:   "\n",
		}

		result := dataFacet.FormatList(items, options)
		if !contains(result, "1. first") || !contains(result, "2. second") {
			t.Errorf("Expected numbered list, got '%s'", result)
		}
	})

	t.Run("FormatList with max items", func(t *testing.T) {
		items := []any{"a", "b", "c", "d", "e"}
		options := ListOptions{
			MaxItems: 3,
		}

		result := dataFacet.FormatList(items, options)
		if !contains(result, "and 2 more") {
			t.Errorf("Expected truncation message, got '%s'", result)
		}
	})

	t.Run("FormatKeyValue", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}
		options := KeyValueOptions{
			SortKeys:   true,
			Assignment: ": ",
			Separator:  "\n",
		}

		result := dataFacet.FormatKeyValue(data, options)
		if !contains(result, "age: 30") || !contains(result, "name: John") {
			t.Errorf("Expected key-value pairs, got '%s'", result)
		}
	})

	t.Run("FormatJSON", func(t *testing.T) {
		data := map[string]any{
			"test": true,
			"num":  42,
		}

		result, err := dataFacet.FormatJSON(data, true)
		if err != nil {
			t.Fatalf("FormatJSON failed: %v", err)
		}

		if !contains(result, "test") || !contains(result, "42") {
			t.Errorf("Expected JSON content, got '%s'", result)
		}
	})

	t.Run("FormatHierarchy", func(t *testing.T) {
		data := map[string]any{
			"level1": map[string]any{
				"level2": "value",
			},
		}

		result := dataFacet.FormatHierarchy(data, 3)
		if !contains(result, "level1") || !contains(result, "level2") {
			t.Errorf("Expected hierarchical structure, got '%s'", result)
		}
	})
}

// TestUnit_FormattingEngine_TemplateFacet tests the ITemplate interface
func TestUnit_FormattingEngine_TemplateFacet(t *testing.T) {
	engine := NewFormattingEngine()
	templateFacet := engine.Template()

	t.Run("ProcessTemplate", func(t *testing.T) {
		template := "Hello {{name}}, your score is {{score}}"
		data := map[string]any{
			"name":  "Alice",
			"score": 95,
		}

		result, err := templateFacet.ProcessTemplate(template, data)
		if err != nil {
			t.Fatalf("ProcessTemplate failed: %v", err)
		}

		expected := "Hello Alice, your score is 95"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("ProcessTemplate with unresolved placeholders", func(t *testing.T) {
		template := "Hello {{name}}, your score is {{score}}"
		data := map[string]any{
			"name": "Alice",
			// missing score
		}

		result, err := templateFacet.ProcessTemplate(template, data)
		if err == nil {
			t.Error("Expected error for unresolved placeholders")
		}
		if !contains(result, "{{score}}") {
			t.Errorf("Expected unresolved placeholder in result '%s'", result)
		}
	})

	t.Run("ValidateTemplate", func(t *testing.T) {
		validTemplates := []string{
			"Hello {{name}}",
			"{{greeting}} {{name}}!",
			"No placeholders here",
			"",
		}

		for _, template := range validTemplates {
			err := templateFacet.ValidateTemplate(template)
			if err != nil {
				t.Errorf("ValidateTemplate('%s') should be valid: %v", template, err)
			}
		}

		invalidTemplates := []string{
			"Hello {{name}",     // unbalanced
			"Hello {{}}",        // empty parameter
			"Hello {{na-me}}",   // invalid character
			"Hello {{123name}}", // starts with number
		}

		for _, template := range invalidTemplates {
			err := templateFacet.ValidateTemplate(template)
			if err == nil {
				t.Errorf("ValidateTemplate('%s') should be invalid", template)
			}
		}
	})

	t.Run("CacheTemplate", func(t *testing.T) {
		template := "Hello {{name}}"
		err := templateFacet.CacheTemplate("greeting", template)
		if err != nil {
			t.Fatalf("CacheTemplate failed: %v", err)
		}

		// Try to cache invalid template
		invalidTemplate := "Hello {{name}"
		err = templateFacet.CacheTemplate("invalid", invalidTemplate)
		if err == nil {
			t.Error("Expected error for invalid template")
		}
	})

	t.Run("GetTemplateMetadata", func(t *testing.T) {
		template := "Hello {{name}}, you have {{count}} {{items}}"
		metadata := templateFacet.GetTemplateMetadata(template)

		expectedParams := 3 // name, count, items
		if len(metadata.Parameters) != expectedParams {
			t.Errorf("Expected %d parameters, got %d", expectedParams, len(metadata.Parameters))
		}

		if metadata.Complexity == 0 {
			t.Error("Expected non-zero complexity")
		}
	})
}

// TestUnit_FormattingEngine_LocaleFacet tests the ILocale interface
func TestUnit_FormattingEngine_LocaleFacet(t *testing.T) {
	engine := NewFormattingEngine()
	localeFacet := engine.Locale()

	t.Run("SetLocale", func(t *testing.T) {
		testLocales := []string{"en-US", "en-GB", "de-DE", "fr-FR", "ja-JP"}

		for _, locale := range testLocales {
			err := localeFacet.SetLocale(locale)
			if err != nil {
				t.Errorf("SetLocale('%s') failed: %v", locale, err)
			}

			current := localeFacet.GetLocale()
			if current != locale {
				t.Errorf("Expected locale '%s', got '%s'", locale, current)
			}
		}

		// Test invalid locale
		err := localeFacet.SetLocale("")
		if err == nil {
			t.Error("Expected error for empty locale")
		}
	})

	t.Run("SetNumberFormat", func(t *testing.T) {
		err := localeFacet.SetNumberFormat(",", ".")
		if err != nil {
			t.Errorf("SetNumberFormat failed: %v", err)
		}
	})

	t.Run("SetDateFormat", func(t *testing.T) {
		err := localeFacet.SetDateFormat("DD/MM/YYYY")
		if err != nil {
			t.Errorf("SetDateFormat failed: %v", err)
		}
	})

	t.Run("SetCurrencyFormat", func(t *testing.T) {
		err := localeFacet.SetCurrencyFormat("USD", "$")
		if err != nil {
			t.Errorf("SetCurrencyFormat failed: %v", err)
		}
	})
}

// TestUnit_FormattingEngine_Integration tests integration between facets
func TestUnit_FormattingEngine_Integration(t *testing.T) {
	engine := NewFormattingEngine()

	t.Run("Text and Template integration", func(t *testing.T) {
		// First format some text
		options := TextOptions{
			CaseTransform: CaseTitle,
			TrimSpaces:    true,
		}

		formattedName, err := engine.Text().FormatText("  john doe  ", options)
		if err != nil {
			t.Fatalf("Text formatting failed: %v", err)
		}

		// Then use it in a template
		template := "Welcome, {{name}}!"
		data := map[string]any{
			"name": formattedName,
		}

		result, err := engine.Template().ProcessTemplate(template, data)
		if err != nil {
			t.Fatalf("Template processing failed: %v", err)
		}

		if !contains(result, "John Doe") {
			t.Errorf("Expected formatted name in result '%s'", result)
		}
	})

	t.Run("Number and Template integration", func(t *testing.T) {
		// Format a number
		amount, err := engine.Number().FormatNumber(1234.56, 2)
		if err != nil {
			t.Fatalf("Number formatting failed: %v", err)
		}

		// Use in template
		template := "Total: {{amount}}"
		data := map[string]any{
			"amount": amount,
		}

		result, err := engine.Template().ProcessTemplate(template, data)
		if err != nil {
			t.Fatalf("Template processing failed: %v", err)
		}

		if !contains(result, "Total:") {
			t.Errorf("Expected total in result '%s'", result)
		}
	})

	t.Run("Locale affects formatting", func(t *testing.T) {
		// Set German locale
		err := engine.Locale().SetLocale("de-DE")
		if err != nil {
			t.Fatalf("SetLocale failed: %v", err)
		}

		// Verify locale change
		if engine.Locale().GetLocale() != "de-DE" {
			t.Error("Locale was not set correctly")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
