package utilities

import (
	"math"
	"strings"
	"testing"
	"time"
)

// TestUnit_ValidateString tests string validation functionality
func TestUnit_ValidateString(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		constraints StringConstraints
		expectValid bool
		expectError bool
	}{
		{
			"valid string",
			"hello world",
			StringConstraints{MinLength: 5, MaxLength: 20, Required: true},
			true,
			false,
		},
		{
			"too short",
			"hi",
			StringConstraints{MinLength: 5, MaxLength: 20, Required: true},
			false,
			false,
		},
		{
			"too long",
			"this is a very long string that exceeds the maximum length",
			StringConstraints{MinLength: 5, MaxLength: 20, Required: true},
			false,
			false,
		},
		{
			"empty not required",
			"",
			StringConstraints{MinLength: 5, MaxLength: 20, Required: false},
			true,
			false,
		},
		{
			"empty but required",
			"",
			StringConstraints{MinLength: 5, MaxLength: 20, Required: true},
			false,
			false,
		},
		{
			"pattern match",
			"hello123",
			StringConstraints{Pattern: `^[a-z]+\d+$`},
			true,
			false,
		},
		{
			"pattern no match",
			"HELLO123",
			StringConstraints{Pattern: `^[a-z]+\d+$`},
			false,
			false,
		},
		{
			"allowed characters",
			"hello",
			StringConstraints{AllowedChars: "abcdefghijklmnopqrstuvwxyz"},
			true,
			false,
		},
		{
			"disallowed characters",
			"hello123",
			StringConstraints{AllowedChars: "abcdefghijklmnopqrstuvwxyz"},
			false,
			false,
		},
		{
			"invalid pattern",
			"test",
			StringConstraints{Pattern: `[unclosed`},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateString(tt.value, tt.constraints)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateNumber tests numeric validation functionality
func TestUnit_ValidateNumber(t *testing.T) {
	min5 := 5.0
	max100 := 100.0
	precision2 := 2

	tests := []struct {
		name        string
		value       float64
		constraints NumericConstraints
		expectValid bool
	}{
		{
			"valid number in range",
			50.0,
			NumericConstraints{Min: &min5, Max: &max100},
			true,
		},
		{
			"below minimum",
			2.0,
			NumericConstraints{Min: &min5, Max: &max100},
			false,
		},
		{
			"above maximum",
			150.0,
			NumericConstraints{Min: &min5, Max: &max100},
			false,
		},
		{
			"positive only valid",
			10.0,
			NumericConstraints{PositiveOnly: true},
			true,
		},
		{
			"positive only invalid",
			-5.0,
			NumericConstraints{PositiveOnly: true},
			false,
		},
		{
			"integer only valid",
			42.0,
			NumericConstraints{IntegerOnly: true},
			true,
		},
		{
			"integer only invalid",
			42.5,
			NumericConstraints{IntegerOnly: true},
			false,
		},
		{
			"precision valid",
			12.34,
			NumericConstraints{Precision: &precision2},
			true,
		},
		{
			"precision invalid",
			12.345,
			NumericConstraints{Precision: &precision2},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateNumber(tt.value, tt.constraints)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateNumber_SpecialValues tests numeric validation with special float values
func TestUnit_ValidateNumber_SpecialValues(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		expectValid bool
	}{
		{"NaN", math.NaN(), false},
		{"positive infinity", math.Inf(1), false},
		{"negative infinity", math.Inf(-1), false},
		{"normal number", 42.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateNumber(tt.value, NumericConstraints{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateBoolean tests boolean validation and conversion
func TestUnit_ValidateBoolean(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectBool   bool
		expectValid  bool
		expectError  bool
	}{
		{"true bool", true, true, true, false},
		{"false bool", false, false, true, false},
		{"true string", "true", true, true, false},
		{"false string", "false", false, true, false},
		{"yes string", "yes", true, true, false},
		{"no string", "no", false, true, false},
		{"1 string", "1", true, true, false},
		{"0 string", "0", false, true, false},
		{"empty string", "", false, true, false},
		{"invalid string", "maybe", false, false, false},
		{"1 int", 1, true, true, false},
		{"0 int", 0, false, true, false},
		{"2 int", 2, false, false, false},
		{"invalid type", 3.14, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boolVal, result, err := ValidateBoolean(tt.value)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
			if result.Valid && boolVal != tt.expectBool {
				t.Errorf("expected bool=%v, got bool=%v", tt.expectBool, boolVal)
			}
		})
	}
}

// TestUnit_ValidateDate tests date validation functionality
func TestUnit_ValidateDate(t *testing.T) {
	format := "2006-01-02"
	minDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		value       string
		format      string
		minDate     *time.Time
		maxDate     *time.Time
		expectValid bool
	}{
		{"valid date", "2022-06-15", format, &minDate, &maxDate, true},
		{"before min", "2019-12-31", format, &minDate, &maxDate, false},
		{"after max", "2026-01-01", format, &minDate, &maxDate, false},
		{"invalid format", "15-06-2022", format, nil, nil, false},
		{"empty date", "", format, nil, nil, false},
		{"invalid date", "2022-13-45", format, nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateDate(tt.value, tt.format, tt.minDate, tt.maxDate)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateEmail tests email validation functionality
func TestUnit_ValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectValid bool
	}{
		{"valid email", "user@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"empty email", "", false},
		{"missing @", "userexample.com", false},
		{"missing domain", "user@", false},
		{"missing user", "@example.com", false},
		{"invalid characters", "user name@example.com", false},
		{"too long email", strings.Repeat("a", 250) + "@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateEmail(tt.email)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateURL tests URL validation functionality
func TestUnit_ValidateURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		allowedSchemes []string
		expectValid    bool
	}{
		{"valid http url", "http://example.com", []string{"http", "https"}, true},
		{"valid https url", "https://example.com", []string{"http", "https"}, true},
		{"invalid scheme", "ftp://example.com", []string{"http", "https"}, false},
		{"missing scheme", "example.com", nil, false},
		{"missing host", "http://", nil, false},
		{"empty url", "", nil, false},
		{"file url", "file:///path/to/file", []string{"file"}, true},
		{"malformed url", "ht tp://example.com", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateURL(tt.url, tt.allowedSchemes)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateUUID tests UUID validation functionality
func TestUnit_ValidateUUID(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		expectValid bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid v1", "12345678-1234-1234-1234-123456789012", true},
		{"invalid uuid format", "550e8400-e29b-41d4-a716", false},
		{"invalid characters", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"empty uuid", "", false},
		{"wrong length", "550e8400-e29b-41d4-a716-4466554400000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUUID(tt.uuid)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidatePattern tests pattern validation functionality
func TestUnit_ValidatePattern(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		pattern     string
		expectValid bool
		expectError bool
	}{
		{"valid pattern match", "hello123", `^[a-z]+\d+$`, true, false},
		{"invalid pattern match", "HELLO123", `^[a-z]+\d+$`, false, false},
		{"empty pattern", "test", "", false, true},
		{"invalid regex", "test", `[unclosed`, false, true},
		{"email pattern", "test@example.com", `^[^@]+@[^@]+\.[^@]+$`, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePattern(tt.value, tt.pattern)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateRequired tests required field validation
func TestUnit_ValidateRequired(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		fieldName   string
		expectValid bool
	}{
		{"valid string", "hello", "username", true},
		{"empty string", "", "username", false},
		{"nil value", nil, "username", false},
		{"valid number", 42, "age", true},
		{"zero number", 0, "age", true}, // zero is valid, not empty
		{"boolean false", false, "active", true}, // false is valid, not empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateRequired(tt.value, tt.fieldName)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
			if !result.Valid && tt.fieldName != "" {
				if _, exists := result.FieldErrors[tt.fieldName]; !exists {
					t.Errorf("expected field error for %s", tt.fieldName)
				}
			}
		})
	}
}

// TestUnit_ValidateConditional tests conditional validation functionality
func TestUnit_ValidateConditional(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		condition   func(interface{}) bool
		rules       []ValidatorRule
		expectValid bool
		expectError bool
	}{
		{
			"condition not met",
			"test",
			func(v interface{}) bool { return false },
			[]ValidatorRule{{Type: "required"}},
			true,
			false,
		},
		{
			"condition met, valid",
			"test",
			func(v interface{}) bool { return true },
			[]ValidatorRule{{Type: "required"}},
			true,
			false,
		},
		{
			"condition met, invalid",
			"",
			func(v interface{}) bool { return true },
			[]ValidatorRule{{Type: "required"}},
			false,
			false,
		},
		{
			"nil condition",
			"test",
			nil,
			[]ValidatorRule{{Type: "required"}},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateConditional(tt.value, "testField", tt.condition, tt.rules)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateCollection tests collection validation functionality
func TestUnit_ValidateCollection(t *testing.T) {
	tests := []struct {
		name            string
		collection      interface{}
		minSize         int
		maxSize         int
		elementValidator func(interface{}) (ValidatorResult, error)
		expectValid     bool
		expectError     bool
	}{
		{
			"valid string array",
			[]string{"a", "b", "c"},
			2,
			5,
			nil,
			true,
			false,
		},
		{
			"too small",
			[]string{"a"},
			2,
			5,
			nil,
			false,
			false,
		},
		{
			"too large",
			[]string{"a", "b", "c", "d", "e", "f"},
			2,
			5,
			nil,
			false,
			false,
		},
		{
			"nil collection",
			nil,
			0,
			5,
			nil,
			false,
			false,
		},
		{
			"unsupported type",
			map[string]string{"key": "value"},
			0,
			5,
			nil,
			false,
			true,
		},
		{
			"with element validation",
			[]interface{}{"valid", ""},
			0,
			5,
			func(v interface{}) (ValidatorResult, error) {
				return ValidateRequired(v, "element")
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateCollection(tt.collection, tt.minSize, tt.maxSize, tt.elementValidator)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateMap tests map validation functionality
func TestUnit_ValidateMap(t *testing.T) {
	tests := []struct {
		name           string
		mapData        map[string]interface{}
		requiredKeys   []string
		valueValidator func(string, interface{}) (ValidatorResult, error)
		expectValid    bool
		expectError    bool
	}{
		{
			"valid map",
			map[string]interface{}{"name": "test", "age": 25},
			[]string{"name"},
			nil,
			true,
			false,
		},
		{
			"missing required key",
			map[string]interface{}{"age": 25},
			[]string{"name"},
			nil,
			false,
			false,
		},
		{
			"nil map",
			nil,
			[]string{"name"},
			nil,
			false,
			false,
		},
		{
			"with value validation",
			map[string]interface{}{"name": "", "age": 25},
			[]string{"name"},
			func(key string, value interface{}) (ValidatorResult, error) {
				if key == "name" {
					return ValidateRequired(value, key)
				}
				return ValidatorResult{Valid: true}, nil
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateMap(tt.mapData, tt.requiredKeys, tt.valueValidator)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_ValidateUnique tests uniqueness validation functionality
func TestUnit_ValidateUnique(t *testing.T) {
	tests := []struct {
		name        string
		collection  interface{}
		expectValid bool
		expectError bool
	}{
		{
			"unique strings",
			[]string{"a", "b", "c"},
			true,
			false,
		},
		{
			"duplicate strings",
			[]string{"a", "b", "a"},
			false,
			false,
		},
		{
			"unique integers",
			[]int{1, 2, 3},
			true,
			false,
		},
		{
			"duplicate integers",
			[]int{1, 2, 1},
			false,
			false,
		},
		{
			"nil collection",
			nil,
			false,
			false,
		},
		{
			"unsupported type",
			map[string]string{"key": "value"},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUnique(tt.collection)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, errors=%v", tt.expectValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_InputSizeLimit tests input size limitations per SRS REQ-IMPL-003
func TestUnit_ValidationUtility_InputSizeLimit(t *testing.T) {
	// Create a string larger than MaxStringLength
	largeInput := strings.Repeat("a", MaxStringLength+1)

	result, err := ValidateString(largeInput, StringConstraints{})
	if err == nil {
		t.Error("expected error for oversized input but got none")
	}
	if result.Valid {
		t.Error("expected validation to fail for oversized input")
	}
	if !strings.Contains(err.Error(), "exceeds maximum length") {
		t.Errorf("expected size limit error, got: %v", err)
	}
}

// TestUnit_ValidationRuleLimit tests validation rule limit per SRS REQ-IMPL-003
func TestUnit_ValidationUtility_ValidationRuleLimit(t *testing.T) {
	// Create more than MaxValidationRules
	rules := make([]ValidatorRule, MaxValidationRules+1)
	for i := range rules {
		rules[i] = ValidatorRule{Type: "required"}
	}

	condition := func(interface{}) bool { return true }
	_, err := ValidateConditional("test", "field", condition, rules)
	if err == nil {
		t.Error("expected error for too many validation rules")
	}
	if !strings.Contains(err.Error(), "too many validation rules") {
		t.Errorf("expected rule limit error, got: %v", err)
	}
}

// TestUnit_ThreadSafety tests concurrent access to ensure thread safety
func TestUnit_ValidationUtility_ThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// Test concurrent execution of all validation functions
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Test all functions concurrently
				ValidateString("test", StringConstraints{Required: true})
				ValidateNumber(42.0, NumericConstraints{PositiveOnly: true})
				ValidateBoolean("true")
				ValidateDate("2024-01-01", "2006-01-02", nil, nil)
				ValidateEmail("test@example.com")
				ValidateURL("https://example.com", []string{"https"})
				ValidateUUID("550e8400-e29b-41d4-a716-446655440000")
				ValidatePattern("test123", `^[a-z]+\d+$`)
				ValidateRequired("value", "field")
				ValidateCollection([]string{"a", "b"}, 1, 5, nil)
				ValidateMap(map[string]interface{}{"key": "value"}, []string{"key"}, nil)
				ValidateUnique([]string{"a", "b", "c"})
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
func TestUnit_ValidationUtility_ErrorHandling(t *testing.T) {
	t.Run("Nil condition function", func(t *testing.T) {
		_, err := ValidateConditional("test", "field", nil, []ValidatorRule{})
		if err == nil {
			t.Error("expected error for nil condition function")
		}
	})

	t.Run("Empty pattern", func(t *testing.T) {
		_, err := ValidatePattern("test", "")
		if err == nil {
			t.Error("expected error for empty pattern")
		}
	})

	t.Run("Invalid regex in string constraints", func(t *testing.T) {
		_, err := ValidateString("test", StringConstraints{Pattern: "[unclosed"})
		if err == nil {
			t.Error("expected error for invalid regex pattern")
		}
	})
}

// Benchmark tests for performance validation per SRS REQ-PERF-001
func BenchmarkValidationUtility(b *testing.B) {
	b.Run("ValidateString", func(b *testing.B) {
		constraints := StringConstraints{MinLength: 3, MaxLength: 50, Required: true}
		for i := 0; i < b.N; i++ {
			ValidateString("test string", constraints)
		}
	})

	b.Run("ValidateNumber", func(b *testing.B) {
		min := 0.0
		max := 100.0
		constraints := NumericConstraints{Min: &min, Max: &max}
		for i := 0; i < b.N; i++ {
			ValidateNumber(42.5, constraints)
		}
	})

	b.Run("ValidateEmail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ValidateEmail("test@example.com")
		}
	})

	b.Run("ValidatePattern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ValidatePattern("test123", `^[a-z]+\d+$`)
		}
	})

	b.Run("ValidateCollection", func(b *testing.B) {
		collection := []string{"a", "b", "c", "d", "e"}
		for i := 0; i < b.N; i++ {
			ValidateCollection(collection, 1, 10, nil)
		}
	})
}