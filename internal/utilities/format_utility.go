// Package utilities provides Utility layer components for the EisenKan system following iDesign methodology.
// This package contains reusable components that provide infrastructure services across all system layers.
package utilities

import (
	"fmt"
	"html"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// TextCaseType represents different text case transformation options
type TextCaseType int

const (
	TextCaseUpper TextCaseType = iota
	TextCaseLower
	TextCaseTitle
	TextCaseSentence
)

// String returns the string representation of TextCaseType
func (t TextCaseType) String() string {
	switch t {
	case TextCaseUpper:
		return "UPPER"
	case TextCaseLower:
		return "LOWER"
	case TextCaseTitle:
		return "TITLE"
	case TextCaseSentence:
		return "SENTENCE"
	default:
		return "UNKNOWN"
	}
}

// FileSizeUnit represents file size unit options
type FileSizeUnit int

const (
	FileSizeBytes FileSizeUnit = iota
	FileSizeKB
	FileSizeMB
	FileSizeGB
	FileSizeTB
	FileSizeAuto
)

// String returns the string representation of FileSizeUnit
func (f FileSizeUnit) String() string {
	switch f {
	case FileSizeBytes:
		return "Bytes"
	case FileSizeKB:
		return "KB"
	case FileSizeMB:
		return "MB"
	case FileSizeGB:
		return "GB"
	case FileSizeTB:
		return "TB"
	case FileSizeAuto:
		return "Auto"
	default:
		return "Unknown"
	}
}

// ValidationRule defines text validation constraints
type ValidationRule struct {
	MinLength    int    // Minimum character count
	MaxLength    int    // Maximum character count
	AllowedChars string // Character set specification (empty means all allowed)
	Required     bool   // Whether empty values are allowed
}

// ValidationResult contains validation outcome and details
type ValidationResult struct {
	Valid   bool     // Whether validation passed
	Errors  []string // List of validation errors
	Details string   // Technical details about failure
}

// Constants for input size limits and processing constraints
const (
	MaxInputSize = 1024 * 1024 // 1MB maximum input size per SRS REQ-IMPL-003
	MaxPrecision = 50          // Maximum decimal precision for number formatting
)

// TrimText removes leading and trailing whitespace including spaces, tabs, and newlines
func TrimText(text string) (string, error) {
	if err := validateInputSize(text); err != nil {
		return "", fmt.Errorf("TrimText: %w", err)
	}

	// Handle nil and empty cases
	if text == "" {
		return "", nil
	}

	// Use strings.TrimSpace which handles all Unicode whitespace
	return strings.TrimSpace(text), nil
}

// ConvertCase transforms text case according to specified type
func ConvertCase(text string, caseType TextCaseType) (string, error) {
	if err := validateInputSize(text); err != nil {
		return "", fmt.Errorf("ConvertCase: %w", err)
	}

	if text == "" {
		return "", nil
	}

	switch caseType {
	case TextCaseUpper:
		return strings.ToUpper(text), nil
	case TextCaseLower:
		return strings.ToLower(text), nil
	case TextCaseTitle:
		return strings.Title(text), nil
	case TextCaseSentence:
		return toSentenceCase(text), nil
	default:
		return "", fmt.Errorf("ConvertCase: invalid case type %d", caseType)
	}
}

// TruncateText truncates text to specified length and appends ellipsis if needed
func TruncateText(text string, maxLength int, ellipsis string) (string, error) {
	if err := validateInputSize(text); err != nil {
		return "", fmt.Errorf("TruncateText: %w", err)
	}

	if maxLength < 0 {
		return "", fmt.Errorf("TruncateText: maxLength cannot be negative: %d", maxLength)
	}

	if maxLength == 0 {
		return "", nil
	}

	if text == "" {
		return "", nil
	}

	// Handle UTF-8 properly by counting runes, not bytes
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text, nil
	}

	// Don't add default ellipsis if empty string is explicitly provided

	ellipsisRunes := []rune(ellipsis)
	if len(ellipsisRunes) >= maxLength {
		// If ellipsis is longer than max length, just return truncated ellipsis
		return string(ellipsisRunes[:maxLength]), nil
	}

	truncateAt := maxLength - len(ellipsisRunes)
	return string(runes[:truncateAt]) + ellipsis, nil
}

// WrapText breaks text at specified width boundaries while preserving word boundaries where possible
func WrapText(text string, width int) (string, error) {
	if err := validateInputSize(text); err != nil {
		return "", fmt.Errorf("WrapText: %w", err)
	}

	if width <= 0 {
		return "", fmt.Errorf("WrapText: width must be positive: %d", width)
	}

	if text == "" {
		return "", nil
	}

	// Split text into lines first
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if line == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}

		wrapped := wrapLine(line, width)
		wrappedLines = append(wrappedLines, wrapped...)
	}

	return strings.Join(wrappedLines, "\n"), nil
}

// FormatNumber formats numeric values with separators and precision
func FormatNumber(value float64, decimalPlaces int, thousandsSeparator string) (string, error) {
	if decimalPlaces < 0 || decimalPlaces > MaxPrecision {
		return "", fmt.Errorf("FormatNumber: invalid decimal places %d (must be 0-%d)", decimalPlaces, MaxPrecision)
	}

	// Handle special float values
	if value != value { // NaN check
		return "NaN", nil
	}
	if value == float64(1<<63-1) { // +Inf approximation
		return "+Inf", nil
	}
	if value == float64(-1<<63) { // -Inf approximation
		return "-Inf", nil
	}

	// Format with specified decimal places
	formatStr := fmt.Sprintf("%%.%df", decimalPlaces)
	formatted := fmt.Sprintf(formatStr, value)

	// Add thousands separator if specified and number is large enough
	if thousandsSeparator != "" {
		formatted = addThousandsSeparator(formatted, thousandsSeparator)
	}

	return formatted, nil
}

// FormatDateTime formats date/time with standard patterns
func FormatDateTime(t time.Time, pattern string) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("FormatDateTime: pattern cannot be empty")
	}

	// Support standard Go time format patterns
	// Pattern is expected to use Go's reference time format
	return t.Format(pattern), nil
}

// FormatFileSize converts bytes to human-readable units
func FormatFileSize(bytes int64, unit FileSizeUnit, precision int) (string, error) {
	if bytes < 0 {
		return "", fmt.Errorf("FormatFileSize: bytes cannot be negative: %d", bytes)
	}

	if precision < 0 || precision > MaxPrecision {
		return "", fmt.Errorf("FormatFileSize: invalid precision %d (must be 0-%d)", precision, MaxPrecision)
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	var value float64
	var unitStr string

	switch unit {
	case FileSizeBytes:
		return fmt.Sprintf("%d Bytes", bytes), nil
	case FileSizeKB:
		value = float64(bytes) / KB
		unitStr = "KB"
	case FileSizeMB:
		value = float64(bytes) / MB
		unitStr = "MB"
	case FileSizeGB:
		value = float64(bytes) / GB
		unitStr = "GB"
	case FileSizeTB:
		value = float64(bytes) / TB
		unitStr = "TB"
	case FileSizeAuto:
		if bytes >= TB {
			value = float64(bytes) / TB
			unitStr = "TB"
		} else if bytes >= GB {
			value = float64(bytes) / GB
			unitStr = "GB"
		} else if bytes >= MB {
			value = float64(bytes) / MB
			unitStr = "MB"
		} else if bytes >= KB {
			value = float64(bytes) / KB
			unitStr = "KB"
		} else {
			return fmt.Sprintf("%d Bytes", bytes), nil
		}
	default:
		return "", fmt.Errorf("FormatFileSize: invalid unit %d", unit)
	}

	formatStr := fmt.Sprintf("%%.%df %s", precision, unitStr)
	return fmt.Sprintf(formatStr, value), nil
}

// FormatPercentage converts decimal to percentage representation
func FormatPercentage(value float64, decimalPlaces int) (string, error) {
	if decimalPlaces < 0 || decimalPlaces > MaxPrecision {
		return "", fmt.Errorf("FormatPercentage: invalid decimal places %d (must be 0-%d)", decimalPlaces, MaxPrecision)
	}

	// Handle special float values
	if value != value { // NaN check
		return "NaN%", nil
	}

	percentage := value * 100
	formatStr := fmt.Sprintf("%%.%df", decimalPlaces)
	return fmt.Sprintf(formatStr, percentage) + "%", nil
}

// EscapeHTML escapes special characters to prevent injection attacks
func EscapeHTML(text string) (string, error) {
	if err := validateInputSize(text); err != nil {
		return "", fmt.Errorf("EscapeHTML: %w", err)
	}

	// Use Go's built-in HTML escaping
	return html.EscapeString(text), nil
}

// NormalizeUnicode applies Unicode normalization to ensure consistent character representation
func NormalizeUnicode(text string) (string, error) {
	if err := validateInputSize(text); err != nil {
		return "", fmt.Errorf("NormalizeUnicode: %w", err)
	}

	// Apply NFC (Canonical Decomposition followed by Canonical Composition)
	// This is the most common normalization form for text processing
	normalized := norm.NFC.String(text)
	return normalized, nil
}

// ValidateText validates text against specified constraints
func ValidateText(text string, rules ValidationRule) (ValidationResult, error) {
	result := ValidationResult{Valid: true, Errors: []string{}}

	// Check input size limit
	if err := validateInputSize(text); err != nil {
		return ValidationResult{
			Valid:   false,
			Errors:  []string{err.Error()},
			Details: "Input exceeds maximum size limit",
		}, fmt.Errorf("ValidateText: %w", err)
	}

	// Check required field
	if rules.Required && text == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "field is required but empty")
	}

	if text != "" {
		// Check length constraints
		textLength := utf8.RuneCountInString(text)

		if rules.MinLength > 0 && textLength < rules.MinLength {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("text length %d is below minimum %d", textLength, rules.MinLength))
		}

		if rules.MaxLength > 0 && textLength > rules.MaxLength {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("text length %d exceeds maximum %d", textLength, rules.MaxLength))
		}

		// Check character set constraints
		if rules.AllowedChars != "" {
			if err := validateCharacterSet(text, rules.AllowedChars); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, err.Error())
			}
		}
	}

	if len(result.Errors) > 0 {
		result.Details = fmt.Sprintf("Validation failed with %d errors", len(result.Errors))
	}

	return result, nil
}

// Helper functions

// validateInputSize checks if input exceeds maximum size limit
func validateInputSize(text string) error {
	if len(text) > MaxInputSize {
		return fmt.Errorf("input size %d exceeds maximum limit %d bytes", len(text), MaxInputSize)
	}
	return nil
}

// toSentenceCase converts text to sentence case (first letter uppercase, rest lowercase)
func toSentenceCase(text string) string {
	if text == "" {
		return ""
	}

	runes := []rune(text)

	// Find first letter and make it uppercase
	firstLetterFound := false
	for i, r := range runes {
		if unicode.IsLetter(r) {
			if !firstLetterFound {
				runes[i] = unicode.ToUpper(r)
				firstLetterFound = true
			} else {
				runes[i] = unicode.ToLower(r)
			}
		}
	}

	return string(runes)
}

// wrapLine wraps a single line of text at the specified width
func wrapLine(line string, width int) []string {
	if line == "" {
		return []string{""}
	}

	// Special case for width = 1: split by characters
	if width == 1 {
		runes := []rune(line)
		var lines []string
		for _, r := range runes {
			if !unicode.IsSpace(r) { // Skip whitespace when wrapping to single chars
				lines = append(lines, string(r))
			}
		}
		if len(lines) == 0 {
			return []string{""}
		}
		return lines
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine []string
	currentLength := 0

	for _, word := range words {
		wordLength := utf8.RuneCountInString(word)

		// If this is the first word on the line, or if adding it doesn't exceed width
		if len(currentLine) == 0 || currentLength+1+wordLength <= width {
			currentLine = append(currentLine, word)
			if len(currentLine) == 1 {
				currentLength = wordLength
			} else {
				currentLength += 1 + wordLength // +1 for space
			}
		} else {
			// Start a new line
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentLength = wordLength
		}
	}

	// Add the last line
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	return lines
}

// addThousandsSeparator adds thousands separator to a formatted number string
func addThousandsSeparator(formatted string, separator string) string {
	// Find decimal point if it exists
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]

	// Handle negative numbers
	var prefix string
	if strings.HasPrefix(integerPart, "-") {
		prefix = "-"
		integerPart = integerPart[1:]
	}

	// Add separators from right to left
	if len(integerPart) <= 3 {
		return formatted // No need for separators
	}

	var result strings.Builder
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 {
			result.WriteString(separator)
		}
		result.WriteRune(digit)
	}

	formatted = prefix + result.String()

	// Add decimal part back if it exists
	if len(parts) > 1 {
		formatted += "." + parts[1]
	}

	return formatted
}

// validateCharacterSet checks if text contains only allowed characters
func validateCharacterSet(text string, allowedChars string) error {
	// Simple implementation - check if each character in text is in allowedChars
	// This is a basic implementation; more sophisticated character set validation
	// could be added based on specific requirements

	for _, r := range text {
		if !strings.ContainsRune(allowedChars, r) {
			return fmt.Errorf("character '%c' is not in allowed character set", r)
		}
	}

	return nil
}