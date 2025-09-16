package utilities

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestUnit_TrimText tests basic text trimming functionality
func TestUnit_TrimText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"empty string", "", "", false},
		{"no whitespace", "hello", "hello", false},
		{"leading spaces", "  hello", "hello", false},
		{"trailing spaces", "hello  ", "hello", false},
		{"both sides", "  hello  ", "hello", false},
		{"tabs and newlines", "\t\nhello\n\t", "hello", false},
		{"unicode whitespace", "\u00A0hello\u00A0", "hello", false},
		{"only whitespace", "   \t\n   ", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TrimText(tt.input)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_ConvertCase tests text case conversion functionality
func TestUnit_ConvertCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		caseType TextCaseType
		expected string
		hasError bool
	}{
		{"upper case", "hello world", TextCaseUpper, "HELLO WORLD", false},
		{"lower case", "HELLO WORLD", TextCaseLower, "hello world", false},
		{"title case", "hello world", TextCaseTitle, "Hello World", false},
		{"sentence case", "HELLO WORLD", TextCaseSentence, "Hello world", false},
		{"empty string", "", TextCaseUpper, "", false},
		{"unicode text", "café naïve", TextCaseUpper, "CAFÉ NAÏVE", false},
		{"invalid case type", "hello", TextCaseType(999), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertCase(tt.input, tt.caseType)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_TruncateText tests text truncation functionality
func TestUnit_TruncateText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		ellipsis  string
		expected  string
		hasError  bool
	}{
		{"no truncation needed", "hello", 10, "...", "hello", false},
		{"exact length", "hello", 5, "...", "hello", false},
		{"truncate with ellipsis", "hello world", 8, "...", "hello...", false},
		{"empty ellipsis", "hello world", 8, "", "hello wo", false},
		{"zero length", "hello", 0, "...", "", false},
		{"negative length", "hello", -1, "...", "", true},
		{"ellipsis longer than max", "hello", 2, "...", "..", false},
		{"empty string", "", 5, "...", "", false},
		{"unicode text", "café naïve résumé", 10, "...", "café na...", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TruncateText(tt.input, tt.maxLength, tt.ellipsis)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_WrapText tests text wrapping functionality
func TestUnit_WrapText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
		hasError bool
	}{
		{"no wrapping needed", "hello world", 20, "hello world", false},
		{"simple wrap", "hello world test", 10, "hello\nworld test", false},
		{"exact fit", "hello", 5, "hello", false},
		{"single char width", "hello", 1, "h\ne\nl\nl\no", false},
		{"zero width", "hello", 0, "", true},
		{"negative width", "hello", -1, "", true},
		{"empty string", "", 10, "", false},
		{"multiple lines", "hello\nworld", 10, "hello\nworld", false},
		{"long word", "supercalifragilisticexpialidocious", 10, "supercalifragilisticexpialidocious", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WrapText(tt.input, tt.width)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_FormatNumber tests number formatting functionality
func TestUnit_FormatNumber(t *testing.T) {
	tests := []struct {
		name               string
		value              float64
		decimalPlaces      int
		thousandsSeparator string
		expected           string
		hasError           bool
	}{
		{"simple integer", 1234, 0, "", "1234", false},
		{"with decimals", 1234.567, 2, "", "1234.57", false},
		{"with thousands separator", 1234567, 2, ",", "1,234,567.00", false},
		{"zero value", 0, 2, ",", "0.00", false},
		{"negative value", -1234.56, 1, ",", "-1,234.6", false},
		{"excessive decimal places", 123.45, 51, "", "", true},
		{"negative decimal places", 123.45, -1, "", "", true},
		{"very large number", 123456789.123, 3, ",", "123,456,789.123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatNumber(tt.value, tt.decimalPlaces, tt.thousandsSeparator)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_FormatDateTime tests date/time formatting functionality
func TestUnit_FormatDateTime(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		time     time.Time
		pattern  string
		expected string
		hasError bool
	}{
		{"ISO format", testTime, "2006-01-02T15:04:05Z07:00", "2024-03-15T14:30:45Z", false},
		{"US format", testTime, "01/02/2006", "03/15/2024", false},
		{"time only", testTime, "15:04:05", "14:30:45", false},
		{"custom format", testTime, "Jan 2, 2006", "Mar 15, 2024", false},
		{"empty pattern", testTime, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatDateTime(tt.time, tt.pattern)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_FormatFileSize tests file size formatting functionality
func TestUnit_FormatFileSize(t *testing.T) {
	tests := []struct {
		name      string
		bytes     int64
		unit      FileSizeUnit
		precision int
		expected  string
		hasError  bool
	}{
		{"bytes", 512, FileSizeBytes, 0, "512 Bytes", false},
		{"kilobytes", 1536, FileSizeKB, 1, "1.5 KB", false},
		{"megabytes", 1048576, FileSizeMB, 0, "1 MB", false},
		{"gigabytes", 3221225472, FileSizeGB, 2, "3.00 GB", false},
		{"terabytes", 1099511627776, FileSizeTB, 1, "1.0 TB", false},
		{"auto small", 512, FileSizeAuto, 0, "512 Bytes", false},
		{"auto KB", 2048, FileSizeAuto, 1, "2.0 KB", false},
		{"auto MB", 1048576, FileSizeAuto, 0, "1 MB", false},
		{"negative bytes", -100, FileSizeBytes, 0, "", true},
		{"invalid precision", 1024, FileSizeKB, -1, "", true},
		{"excessive precision", 1024, FileSizeKB, 51, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatFileSize(tt.bytes, tt.unit, tt.precision)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_FormatPercentage tests percentage formatting functionality
func TestUnit_FormatPercentage(t *testing.T) {
	tests := []struct {
		name          string
		value         float64
		decimalPlaces int
		expected      string
		hasError      bool
	}{
		{"simple percentage", 0.25, 0, "25%", false},
		{"with decimals", 0.3333, 2, "33.33%", false},
		{"zero percent", 0.0, 1, "0.0%", false},
		{"over 100%", 1.5, 1, "150.0%", false},
		{"negative percent", -0.1, 2, "-10.00%", false},
		{"excessive decimal places", 0.5, 51, "", true},
		{"negative decimal places", 0.5, -1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatPercentage(tt.value, tt.decimalPlaces)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_EscapeHTML tests HTML escaping functionality
func TestUnit_EscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"no special chars", "hello world", "hello world", false},
		{"basic HTML chars", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", false},
		{"quotes", `"hello" & 'world'`, "&#34;hello&#34; &amp; &#39;world&#39;", false},
		{"ampersand", "ben & jerry", "ben &amp; jerry", false},
		{"empty string", "", "", false},
		{"unicode", "café & naïve", "café &amp; naïve", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EscapeHTML(tt.input)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_NormalizeUnicode tests Unicode normalization functionality
func TestUnit_NormalizeUnicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"no normalization needed", "hello", "hello", false},
		{"empty string", "", "", false},
		{"basic ASCII", "Hello World 123", "Hello World 123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeUnicode(tt.input)
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnit_ValidateText tests text validation functionality
func TestUnit_ValidateText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		rules       ValidationRule
		expectValid bool
		expectError bool
	}{
		{
			"valid text",
			"hello",
			ValidationRule{MinLength: 3, MaxLength: 10, Required: true},
			true,
			false,
		},
		{
			"too short",
			"hi",
			ValidationRule{MinLength: 3, MaxLength: 10, Required: true},
			false,
			false,
		},
		{
			"too long",
			"this is a very long text",
			ValidationRule{MinLength: 3, MaxLength: 10, Required: true},
			false,
			false,
		},
		{
			"empty not required",
			"",
			ValidationRule{MinLength: 3, MaxLength: 10, Required: false},
			true,
			false,
		},
		{
			"empty but required",
			"",
			ValidationRule{MinLength: 3, MaxLength: 10, Required: true},
			false,
			false,
		},
		{
			"character set validation",
			"hello",
			ValidationRule{AllowedChars: "abcdefghijklmnopqrstuvwxyz"},
			true,
			false,
		},
		{
			"invalid character",
			"hello123",
			ValidationRule{AllowedChars: "abcdefghijklmnopqrstuvwxyz"},
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateText(tt.input, tt.rules)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_InputSizeLimit tests input size limitations as per SRS REQ-IMPL-003
func TestUnit_InputSizeLimit(t *testing.T) {
	// Create a string larger than MaxInputSize (1MB)
	largeInput := strings.Repeat("a", MaxInputSize+1)

	testFunctions := []struct {
		name string
		fn   func() error
	}{
		{"TrimText", func() error { _, err := TrimText(largeInput); return err }},
		{"ConvertCase", func() error { _, err := ConvertCase(largeInput, TextCaseUpper); return err }},
		{"TruncateText", func() error { _, err := TruncateText(largeInput, 100, "..."); return err }},
		{"WrapText", func() error { _, err := WrapText(largeInput, 80); return err }},
		{"EscapeHTML", func() error { _, err := EscapeHTML(largeInput); return err }},
		{"NormalizeUnicode", func() error { _, err := NormalizeUnicode(largeInput); return err }},
		{"ValidateText", func() error {
			result, err := ValidateText(largeInput, ValidationRule{})
			if err != nil {
				return err
			}
			if result.Valid {
				return fmt.Errorf("expected validation to fail for oversized input")
			}
			return nil
		}},
	}

	for _, tt := range testFunctions {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("expected error for oversized input but got none")
			}
			if !strings.Contains(err.Error(), "exceeds maximum limit") {
				t.Errorf("expected size limit error, got: %v", err)
			}
		})
	}
}

// TestUnit_ThreadSafety tests concurrent access to ensure thread safety
func TestUnit_ThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 1000

	done := make(chan bool, numGoroutines)

	// Test concurrent execution of all functions
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Test all functions concurrently
				TrimText("  hello world  ")
				ConvertCase("Hello World", TextCaseUpper)
				TruncateText("Hello World", 5, "...")
				WrapText("Hello World Test", 10)
				FormatNumber(1234.56, 2, ",")
				FormatDateTime(time.Now(), "2006-01-02")
				FormatFileSize(1024, FileSizeAuto, 1)
				FormatPercentage(0.75, 1)
				EscapeHTML("<script>alert('test')</script>")
				NormalizeUnicode("café")
				ValidateText("test", ValidationRule{MinLength: 1, MaxLength: 10})
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// If we reach here without panics or data races, the functions are thread-safe
}

// TestUnit_ErrorHandling tests comprehensive error conditions
func TestUnit_ErrorHandling(t *testing.T) {
	// Test various error conditions
	t.Run("Invalid enum values", func(t *testing.T) {
		_, err := ConvertCase("test", TextCaseType(999))
		if err == nil {
			t.Error("expected error for invalid case type")
		}

		_, err = FormatFileSize(1024, FileSizeUnit(999), 2)
		if err == nil {
			t.Error("expected error for invalid file size unit")
		}
	})

	t.Run("Edge case values", func(t *testing.T) {
		_, err := TruncateText("test", -1, "...")
		if err == nil {
			t.Error("expected error for negative max length")
		}

		_, err = WrapText("test", 0)
		if err == nil {
			t.Error("expected error for zero width")
		}

		_, err = FormatDateTime(time.Now(), "")
		if err == nil {
			t.Error("expected error for empty pattern")
		}
	})
}

// Benchmark tests for performance validation per SRS REQ-PERF-001
func BenchmarkFormatUtility(b *testing.B) {
	testText := "The quick brown fox jumps over the lazy dog"

	b.Run("TrimText", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			TrimText("  " + testText + "  ")
		}
	})

	b.Run("ConvertCase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ConvertCase(testText, TextCaseUpper)
		}
	})

	b.Run("TruncateText", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			TruncateText(testText, 20, "...")
		}
	})

	b.Run("FormatNumber", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FormatNumber(1234567.89, 2, ",")
		}
	})

	b.Run("EscapeHTML", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			EscapeHTML("<script>" + testText + "</script>")
		}
	})
}