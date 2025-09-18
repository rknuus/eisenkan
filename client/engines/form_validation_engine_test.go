package engines

import (
	"testing"
)

// TestUnit_FormValidationEngine_NewFormValidationEngine tests the constructor
func TestUnit_FormValidationEngine_NewFormValidationEngine(t *testing.T) {
	engine := NewFormValidationEngine()
	if engine == nil {
		t.Fatal("NewFormValidationEngine should not return nil")
	}
}

// TestUnit_FormValidationEngine_ValidateFormInputs tests complete form validation
func TestUnit_FormValidationEngine_ValidateFormInputs(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name      string
		formData  map[string]interface{}
		rules     ValidationRules
		wantValid bool
	}{
		{
			name:      "nil form data",
			formData:  nil,
			rules:     ValidationRules{},
			wantValid: false,
		},
		{
			name:      "empty form data with no rules",
			formData:  map[string]interface{}{},
			rules:     ValidationRules{},
			wantValid: true,
		},
		{
			name: "valid text field",
			formData: map[string]interface{}{
				"name": "John Doe",
			},
			rules: ValidationRules{
				FieldRules: map[string]FieldRule{
					"name": {
						Required: true,
						Type:     FieldTypeText,
						Format: FormatConstraints{
							TextFormat: TextConstraints{
								MinLength: 1,
								MaxLength: 50,
							},
						},
					},
				},
			},
			wantValid: true,
		},
		{
			name: "required field missing",
			formData: map[string]interface{}{
				"email": "",
			},
			rules: ValidationRules{
				FieldRules: map[string]FieldRule{
					"name": {
						Required: true,
						Type:     FieldTypeText,
					},
				},
			},
			wantValid: false,
		},
		{
			name: "text too long",
			formData: map[string]interface{}{
				"description": "This is a very long description that exceeds the maximum length limit",
			},
			rules: ValidationRules{
				FieldRules: map[string]FieldRule{
					"description": {
						Required: true,
						Type:     FieldTypeText,
						Format: FormatConstraints{
							TextFormat: TextConstraints{
								MaxLength: 20,
							},
						},
					},
				},
			},
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateFormInputs(tc.formData, tc.rules)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v", tc.wantValid, result.Valid)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateTextFormat tests text format validation
func TestUnit_FormValidationEngine_ValidateTextFormat(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name        string
		text        string
		constraints TextConstraints
		wantValid   bool
	}{
		{
			name:        "empty text with no constraints",
			text:        "",
			constraints: TextConstraints{},
			wantValid:   true,
		},
		{
			name: "text within length limits",
			text: "Hello",
			constraints: TextConstraints{
				MinLength: 3,
				MaxLength: 10,
			},
			wantValid: true,
		},
		{
			name: "text too short",
			text: "Hi",
			constraints: TextConstraints{
				MinLength: 5,
			},
			wantValid: false,
		},
		{
			name: "text too long",
			text: "This is a very long text",
			constraints: TextConstraints{
				MaxLength: 10,
			},
			wantValid: false,
		},
		{
			name: "forbidden character present",
			text: "Hello@World",
			constraints: TextConstraints{
				ForbiddenChars: "@#$",
			},
			wantValid: false,
		},
		{
			name: "only allowed characters",
			text: "ABC123",
			constraints: TextConstraints{
				AllowedChars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			},
			wantValid: true,
		},
		{
			name: "disallowed character present",
			text: "ABC123!",
			constraints: TextConstraints{
				AllowedChars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			},
			wantValid: false,
		},
		{
			name: "ASCII only constraint - valid",
			text: "Hello World 123",
			constraints: TextConstraints{
				Encoding: "ascii",
			},
			wantValid: true,
		},
		{
			name: "ASCII only constraint - invalid",
			text: "Hello 世界",
			constraints: TextConstraints{
				Encoding: "ascii",
			},
			wantValid: false,
		},
		{
			name: "pattern match - valid",
			text: "abc123",
			constraints: TextConstraints{
				Pattern: "^[a-z]+[0-9]+$",
			},
			wantValid: true,
		},
		{
			name: "pattern match - invalid",
			text: "123abc",
			constraints: TextConstraints{
				Pattern: "^[a-z]+[0-9]+$",
			},
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateTextFormat(tc.text, tc.constraints)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateNumericFormat tests numeric validation
func TestUnit_FormValidationEngine_ValidateNumericFormat(t *testing.T) {
	engine := NewFormValidationEngine()

	minVal := 10.0
	maxVal := 100.0
	decimalPlaces := 2

	testCases := []struct {
		name        string
		value       interface{}
		constraints NumericConstraints
		wantValid   bool
	}{
		{
			name:        "valid integer",
			value:       42,
			constraints: NumericConstraints{},
			wantValid:   true,
		},
		{
			name:        "valid float",
			value:       42.5,
			constraints: NumericConstraints{},
			wantValid:   true,
		},
		{
			name:        "valid string number",
			value:       "42.5",
			constraints: NumericConstraints{},
			wantValid:   true,
		},
		{
			name:        "invalid string",
			value:       "not a number",
			constraints: NumericConstraints{},
			wantValid:   false,
		},
		{
			name:        "non-numeric type",
			value:       []int{1, 2, 3},
			constraints: NumericConstraints{},
			wantValid:   false,
		},
		{
			name:  "value within range",
			value: 50.0,
			constraints: NumericConstraints{
				MinValue: &minVal,
				MaxValue: &maxVal,
			},
			wantValid: true,
		},
		{
			name:  "value below minimum",
			value: 5.0,
			constraints: NumericConstraints{
				MinValue: &minVal,
			},
			wantValid: false,
		},
		{
			name:  "value above maximum",
			value: 150.0,
			constraints: NumericConstraints{
				MaxValue: &maxVal,
			},
			wantValid: false,
		},
		{
			name:  "integer constraint - valid",
			value: 42.0,
			constraints: NumericConstraints{
				IntegerOnly: true,
			},
			wantValid: true,
		},
		{
			name:  "integer constraint - invalid",
			value: 42.5,
			constraints: NumericConstraints{
				IntegerOnly: true,
			},
			wantValid: false,
		},
		{
			name:  "positive constraint - valid",
			value: 42.0,
			constraints: NumericConstraints{
				Positive: true,
			},
			wantValid: true,
		},
		{
			name:  "positive constraint - invalid zero",
			value: 0.0,
			constraints: NumericConstraints{
				Positive: true,
			},
			wantValid: false,
		},
		{
			name:  "positive constraint - invalid negative",
			value: -42.0,
			constraints: NumericConstraints{
				Positive: true,
			},
			wantValid: false,
		},
		{
			name:  "decimal places constraint - valid",
			value: 42.12,
			constraints: NumericConstraints{
				DecimalPlaces: &decimalPlaces,
			},
			wantValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateNumericFormat(tc.value, tc.constraints)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateDateFormat tests date format validation
func TestUnit_FormValidationEngine_ValidateDateFormat(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name      string
		dateStr   string
		format    string
		wantValid bool
	}{
		{
			name:      "valid RFC3339 date",
			dateStr:   "2023-12-25T10:30:00Z",
			format:    "",
			wantValid: true,
		},
		{
			name:      "valid custom format",
			dateStr:   "2023-12-25",
			format:    "2006-01-02",
			wantValid: true,
		},
		{
			name:      "invalid date format",
			dateStr:   "25-12-2023",
			format:    "2006-01-02",
			wantValid: false,
		},
		{
			name:      "invalid date string",
			dateStr:   "not a date",
			format:    "",
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateDateFormat(tc.dateStr, tc.format)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateEmailFormat tests email validation
func TestUnit_FormValidationEngine_ValidateEmailFormat(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name      string
		email     string
		wantValid bool
	}{
		{
			name:      "valid email",
			email:     "user@example.com",
			wantValid: true,
		},
		{
			name:      "valid email with subdomain",
			email:     "user@mail.example.com",
			wantValid: true,
		},
		{
			name:      "invalid email - no @",
			email:     "userexample.com",
			wantValid: false,
		},
		{
			name:      "invalid email - no domain",
			email:     "user@",
			wantValid: false,
		},
		{
			name:      "invalid email - no user",
			email:     "@example.com",
			wantValid: false,
		},
		{
			name:      "empty email",
			email:     "",
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateEmailFormat(tc.email)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateURLFormat tests URL validation
func TestUnit_FormValidationEngine_ValidateURLFormat(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{
			name:      "valid HTTP URL",
			url:       "http://example.com",
			wantValid: true,
		},
		{
			name:      "valid HTTPS URL",
			url:       "https://example.com/path?query=value",
			wantValid: true,
		},
		{
			name:      "valid relative URL",
			url:       "/path/to/resource",
			wantValid: true,
		},
		{
			name:      "invalid URL - spaces",
			url:       "http://example .com",
			wantValid: false,
		},
		{
			name:      "empty URL",
			url:       "",
			wantValid: true, // url.Parse accepts empty strings
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateURLFormat(tc.url)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidatePattern tests pattern validation
func TestUnit_FormValidationEngine_ValidatePattern(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name      string
		input     string
		pattern   string
		wantValid bool
	}{
		{
			name:      "valid pattern match",
			input:     "abc123",
			pattern:   "^[a-z]+[0-9]+$",
			wantValid: true,
		},
		{
			name:      "invalid pattern match",
			input:     "123abc",
			pattern:   "^[a-z]+[0-9]+$",
			wantValid: false,
		},
		{
			name:      "invalid regex pattern",
			input:     "test",
			pattern:   "[invalid",
			wantValid: false,
		},
		{
			name:      "empty input matches empty pattern",
			input:     "",
			pattern:   "",
			wantValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidatePattern(tc.input, tc.pattern)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_SanitizeInputs tests input sanitization
func TestUnit_FormValidationEngine_SanitizeInputs(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name        string
		input       map[string]interface{}
		wantChanges bool
	}{
		{
			name:        "nil input",
			input:       nil,
			wantChanges: false,
		},
		{
			name: "no HTML content",
			input: map[string]interface{}{
				"name": "John Doe",
				"age":  25,
			},
			wantChanges: false,
		},
		{
			name: "HTML content to sanitize",
			input: map[string]interface{}{
				"name":        "John <script>alert('xss')</script> Doe",
				"description": "Hello <b>world</b>",
			},
			wantChanges: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.SanitizeInputs(tc.input)
			hasChanges := len(result.Changes) > 0

			if hasChanges != tc.wantChanges {
				t.Errorf("Expected changes=%v, got changes=%v", tc.wantChanges, hasChanges)
			}

			if tc.input != nil && result.Data == nil {
				t.Error("Expected sanitized data to be non-nil")
			}
		})
	}
}

// TestUnit_FormValidationEngine_SanitizeHTML tests HTML sanitization
func TestUnit_FormValidationEngine_SanitizeHTML(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no HTML",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "script tag removal",
			input:    "Hello <script>alert('xss')</script> World",
			expected: "Hello  World",
		},
		{
			name:     "HTML entity escaping",
			input:    "Hello <b>world</b>",
			expected: "Hello &lt;b&gt;world&lt;/b&gt;",
		},
		{
			name:     "quote escaping",
			input:    `He said "Hello" and 'Goodbye'`,
			expected: "He said &quot;Hello&quot; and &#39;Goodbye&#39;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.SanitizeHTML(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateCharacterSet tests character set validation
func TestUnit_FormValidationEngine_ValidateCharacterSet(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name           string
		input          string
		allowedChars   string
		forbiddenChars string
		wantValid      bool
	}{
		{
			name:      "no constraints",
			input:     "Hello World",
			wantValid: true,
		},
		{
			name:         "allowed characters only",
			input:        "ABC123",
			allowedChars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			wantValid:    true,
		},
		{
			name:         "disallowed character present",
			input:        "ABC123!",
			allowedChars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			wantValid:    false,
		},
		{
			name:           "no forbidden characters",
			input:          "Hello World",
			forbiddenChars: "@#$",
			wantValid:      true,
		},
		{
			name:           "forbidden character present",
			input:          "Hello@World",
			forbiddenChars: "@#$",
			wantValid:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateCharacterSet(tc.input, tc.allowedChars, tc.forbiddenChars)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateDependentFields tests field dependency validation
func TestUnit_FormValidationEngine_ValidateDependentFields(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name         string
		formData     map[string]interface{}
		dependencies []map[string]interface{}
		wantValid    bool
	}{
		{
			name: "no dependencies",
			formData: map[string]interface{}{
				"field1": "value1",
			},
			dependencies: []map[string]interface{}{},
			wantValid:    true,
		},
		{
			name: "dependency satisfied",
			formData: map[string]interface{}{
				"primary":   "value",
				"dependent": "dependent_value",
			},
			dependencies: []map[string]interface{}{
				{
					"primary":   "primary",
					"dependent": "dependent",
				},
			},
			wantValid: true,
		},
		{
			name: "dependency not satisfied",
			formData: map[string]interface{}{
				"primary":   "value",
				"dependent": "",
			},
			dependencies: []map[string]interface{}{
				{
					"primary":   "primary",
					"dependent": "dependent",
				},
			},
			wantValid: false,
		},
		{
			name: "primary field empty",
			formData: map[string]interface{}{
				"primary":   "",
				"dependent": "",
			},
			dependencies: []map[string]interface{}{
				{
					"primary":   "primary",
					"dependent": "dependent",
				},
			},
			wantValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateDependentFields(tc.formData, tc.dependencies)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}

// TestUnit_FormValidationEngine_ValidateStructure tests structure validation
func TestUnit_FormValidationEngine_ValidateStructure(t *testing.T) {
	engine := NewFormValidationEngine()

	testCases := []struct {
		name      string
		data      interface{}
		schema    map[string]interface{}
		wantValid bool
	}{
		{
			name:      "nil data",
			data:      nil,
			schema:    map[string]interface{}{},
			wantValid: false,
		},
		{
			name:      "valid structure - no required fields",
			data:      map[string]interface{}{"field1": "value1"},
			schema:    map[string]interface{}{},
			wantValid: true,
		},
		{
			name: "valid structure - required fields present",
			data: map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
			schema: map[string]interface{}{
				"required": []string{"name", "email"},
			},
			wantValid: true,
		},
		{
			name: "invalid structure - required field missing",
			data: map[string]interface{}{
				"name": "John",
			},
			schema: map[string]interface{}{
				"required": []string{"name", "email"},
			},
			wantValid: false,
		},
		{
			name:      "non-map data",
			data:      "not a map",
			schema:    map[string]interface{}{},
			wantValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.ValidateStructure(tc.data, tc.schema)
			if result.Valid != tc.wantValid {
				t.Errorf("Expected valid=%v, got valid=%v, errors=%v", tc.wantValid, result.Valid, result.Errors)
			}
		})
	}
}
