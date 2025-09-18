// Package engines provides Engine layer components implementing the iDesign methodology.
// This package contains components that encapsulate validation logic and provide
// pure processing services to higher-level components in the application architecture.
package engines

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// ValidationResult represents the outcome of validation operations
type ValidationResult struct {
	Valid        bool                             `json:"valid"`
	Errors       []ValidationError                `json:"errors,omitempty"`
	Warnings     []ValidationWarning              `json:"warnings,omitempty"`
	FieldResults map[string]FieldValidationResult `json:"field_results,omitempty"`
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field    string         `json:"field"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Severity ErrorSeverity  `json:"severity"`
	Details  map[string]any `json:"details,omitempty"`
}

// ValidationWarning represents a non-critical validation issue
type ValidationWarning struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// FieldValidationResult represents validation result for a single field
type FieldValidationResult struct {
	Valid     bool     `json:"valid"`
	Value     any      `json:"value"`
	Errors    []string `json:"errors,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
	Sanitized bool     `json:"sanitized"`
}

// SanitizedData represents sanitized input data with change tracking
type SanitizedData struct {
	Data     map[string]any       `json:"data"`
	Changes  []SanitizationChange `json:"changes,omitempty"`
	Warnings []string             `json:"warnings,omitempty"`
}

// SanitizationChange represents a modification made during sanitization
type SanitizationChange struct {
	Field      string           `json:"field"`
	Original   string           `json:"original"`
	Sanitized  string           `json:"sanitized"`
	ChangeType SanitizationType `json:"change_type"`
}

// ErrorSeverity represents the severity level of validation errors
type ErrorSeverity string

const (
	ErrorSeverityError   ErrorSeverity = "error"
	ErrorSeverityWarning ErrorSeverity = "warning"
	ErrorSeverityInfo    ErrorSeverity = "info"
)

// SanitizationType represents the type of sanitization performed
type SanitizationType string

const (
	SanitizationTypeEscape    SanitizationType = "escape"
	SanitizationTypeRemove    SanitizationType = "remove"
	SanitizationTypeNormalize SanitizationType = "normalize"
	SanitizationTypeTruncate  SanitizationType = "truncate"
)

// ValidationRules defines the validation configuration for form inputs
type ValidationRules struct {
	FieldRules  map[string]FieldRule `json:"field_rules"`
	CrossField  []CrossFieldRule     `json:"cross_field,omitempty"`
	GlobalRules GlobalValidationRule `json:"global_rules,omitempty"`
}

// FieldRule defines validation rules for individual fields
type FieldRule struct {
	Required     bool              `json:"required"`
	Type         FieldType         `json:"type"`
	Format       FormatConstraints `json:"format,omitempty"`
	Length       LengthConstraints `json:"length,omitempty"`
	Pattern      string            `json:"pattern,omitempty"`
	CustomFormat *CustomFormat     `json:"custom_format,omitempty"`
}

// CrossFieldRule defines validation rules between multiple fields
type CrossFieldRule struct {
	Type      CrossFieldType `json:"type"`
	Fields    []string       `json:"fields"`
	Condition string         `json:"condition,omitempty"`
	Rule      FieldRule      `json:"rule"`
}

// GlobalValidationRule defines global validation settings
type GlobalValidationRule struct {
	MaxInputSize   int  `json:"max_input_size"`
	StrictMode     bool `json:"strict_mode"`
	SanitizeInputs bool `json:"sanitize_inputs"`
}

// FieldType represents the expected type of a field
type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumeric FieldType = "numeric"
	FieldTypeDate    FieldType = "date"
	FieldTypeEmail   FieldType = "email"
	FieldTypeURL     FieldType = "url"
	FieldTypeJSON    FieldType = "json"
	FieldTypeArray   FieldType = "array"
	FieldTypeBoolean FieldType = "boolean"
)

// CrossFieldType represents the type of cross-field validation
type CrossFieldType string

const (
	CrossFieldTypeDependency  CrossFieldType = "dependency"
	CrossFieldTypeConditional CrossFieldType = "conditional"
	CrossFieldTypeConsistency CrossFieldType = "consistency"
)

// FormatConstraints defines format-specific validation constraints
type FormatConstraints struct {
	DateFormat    string             `json:"date_format,omitempty"`
	EmailFormat   EmailConstraints   `json:"email_format,omitempty"`
	URLFormat     URLConstraints     `json:"url_format,omitempty"`
	NumericFormat NumericConstraints `json:"numeric_format,omitempty"`
	TextFormat    TextConstraints    `json:"text_format,omitempty"`
}

// TextConstraints defines text-specific validation constraints
type TextConstraints struct {
	MinLength      int    `json:"min_length"`
	MaxLength      int    `json:"max_length"`
	AllowedChars   string `json:"allowed_chars,omitempty"`
	ForbiddenChars string `json:"forbidden_chars,omitempty"`
	Pattern        string `json:"pattern,omitempty"`
	Encoding       string `json:"encoding,omitempty"`
}

// NumericConstraints defines numeric validation constraints
type NumericConstraints struct {
	MinValue      *float64 `json:"min_value,omitempty"`
	MaxValue      *float64 `json:"max_value,omitempty"`
	DecimalPlaces *int     `json:"decimal_places,omitempty"`
	IntegerOnly   bool     `json:"integer_only"`
	Positive      bool     `json:"positive"`
}

// EmailConstraints defines email validation constraints
type EmailConstraints struct {
	AllowLocalOnly bool `json:"allow_local_only"`
	MaxLength      int  `json:"max_length"`
}

// URLConstraints defines URL validation constraints
type URLConstraints struct {
	AllowedSchemes []string `json:"allowed_schemes,omitempty"`
	RequireScheme  bool     `json:"require_scheme"`
	MaxLength      int      `json:"max_length"`
}

// LengthConstraints defines length validation constraints
type LengthConstraints struct {
	MinLength int `json:"min_length"`
	MaxLength int `json:"max_length"`
}

// CustomFormat defines custom format validation rules
type CustomFormat struct {
	Pattern     string   `json:"pattern"`
	Description string   `json:"description,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// FormValidationEngine provides stateless form validation operations
type FormValidationEngine struct{}

// NewFormValidationEngine creates a new FormValidationEngine instance
func NewFormValidationEngine() *FormValidationEngine {
	return &FormValidationEngine{}
}

// ValidateFormInputs validates complete form data against specified validation rules
func (fve *FormValidationEngine) ValidateFormInputs(formData map[string]any, rules ValidationRules) ValidationResult {
	if formData == nil {
		return ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Field:    "form",
				Code:     "NULL_INPUT",
				Message:  "Form data cannot be null",
				Severity: ErrorSeverityError,
			}},
		}
	}

	result := ValidationResult{
		Valid:        true,
		FieldResults: make(map[string]FieldValidationResult),
	}

	// Validate global rules first
	if err := fve.validateGlobalRules(formData, rules.GlobalRules); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, *err)
		return result
	}

	// Validate individual fields
	for fieldName, fieldRule := range rules.FieldRules {
		fieldResult := fve.validateField(fieldName, formData[fieldName], fieldRule)
		result.FieldResults[fieldName] = fieldResult

		if !fieldResult.Valid {
			result.Valid = false
			for _, errorMsg := range fieldResult.Errors {
				result.Errors = append(result.Errors, ValidationError{
					Field:    fieldName,
					Code:     "FIELD_VALIDATION_ERROR",
					Message:  errorMsg,
					Severity: ErrorSeverityError,
				})
			}
		}

		for _, warningMsg := range fieldResult.Warnings {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   fieldName,
				Code:    "FIELD_VALIDATION_WARNING",
				Message: warningMsg,
			})
		}
	}

	// Validate cross-field rules
	for _, crossRule := range rules.CrossField {
		if err := fve.validateCrossFieldRule(formData, crossRule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	return result
}

// validateGlobalRules validates global validation rules
func (fve *FormValidationEngine) validateGlobalRules(formData map[string]any, globalRules GlobalValidationRule) *ValidationError {
	if globalRules.MaxInputSize > 0 {
		totalSize := fve.calculateDataSize(formData)
		if totalSize > globalRules.MaxInputSize {
			return &ValidationError{
				Field:    "form",
				Code:     "INPUT_SIZE_EXCEEDED",
				Message:  fmt.Sprintf("Total input size %d exceeds maximum %d", totalSize, globalRules.MaxInputSize),
				Severity: ErrorSeverityError,
				Details:  map[string]any{"actual_size": totalSize, "max_size": globalRules.MaxInputSize},
			}
		}
	}
	return nil
}

// calculateDataSize estimates the size of form data
func (fve *FormValidationEngine) calculateDataSize(data map[string]any) int {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return len(jsonData)
}

// validateField validates a single field against its rules
func (fve *FormValidationEngine) validateField(_ string, value any, rule FieldRule) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: value,
	}

	// Check required field
	if rule.Required && (value == nil || value == "") {
		result.Valid = false
		result.Errors = append(result.Errors, "Field is required")
		return result
	}

	// Skip validation for optional empty fields
	if !rule.Required && (value == nil || value == "") {
		return result
	}

	// Convert value to string for validation
	strValue := fmt.Sprintf("%v", value)

	// Validate based on field type
	switch rule.Type {
	case FieldTypeText:
		if fieldResult := fve.ValidateTextFormat(strValue, rule.Format.TextFormat); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	case FieldTypeNumeric:
		if fieldResult := fve.ValidateNumericFormat(value, rule.Format.NumericFormat); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	case FieldTypeDate:
		if fieldResult := fve.ValidateDateFormat(strValue, rule.Format.DateFormat); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	case FieldTypeEmail:
		if fieldResult := fve.ValidateEmailFormat(strValue); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	case FieldTypeURL:
		if fieldResult := fve.ValidateURLFormat(strValue); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	}

	// Validate custom pattern if specified
	if rule.Pattern != "" {
		if fieldResult := fve.ValidatePattern(strValue, rule.Pattern); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	}

	// Validate custom format if specified
	if rule.CustomFormat != nil {
		if fieldResult := fve.ValidateCustomFormat(strValue, *rule.CustomFormat); !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	}

	return result
}

// ValidateTextFormat validates text input against text-specific constraints
func (fve *FormValidationEngine) ValidateTextFormat(text string, constraints TextConstraints) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: text,
	}

	// Validate length constraints
	textLength := utf8.RuneCountInString(text)
	if constraints.MinLength > 0 && textLength < constraints.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Text length %d is below minimum %d", textLength, constraints.MinLength))
	}
	if constraints.MaxLength > 0 && textLength > constraints.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Text length %d exceeds maximum %d", textLength, constraints.MaxLength))
	}

	// Validate character constraints
	if constraints.ForbiddenChars != "" {
		for _, char := range constraints.ForbiddenChars {
			if strings.ContainsRune(text, char) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Text contains forbidden character: %c", char))
			}
		}
	}

	if constraints.AllowedChars != "" {
		for _, char := range text {
			if !strings.ContainsRune(constraints.AllowedChars, char) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Text contains disallowed character: %c", char))
				break
			}
		}
	}

	// Validate encoding
	if constraints.Encoding == "ascii" && !fve.isASCII(text) {
		result.Valid = false
		result.Errors = append(result.Errors, "Text contains non-ASCII characters")
	}

	// Validate pattern if specified
	if constraints.Pattern != "" {
		if matched, err := regexp.MatchString(constraints.Pattern, text); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, "Invalid pattern in constraints")
		} else if !matched {
			result.Valid = false
			result.Errors = append(result.Errors, "Text does not match required pattern")
		}
	}

	return result
}

// ValidateNumericFormat validates numeric input against numeric constraints
func (fve *FormValidationEngine) ValidateNumericFormat(value any, constraints NumericConstraints) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: value,
	}

	// Convert to float64 for validation
	var numValue float64
	var err error

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	case string:
		if numValue, err = strconv.ParseFloat(v, 64); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, "Invalid numeric format")
			return result
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, "Value is not numeric")
		return result
	}

	// Validate range constraints
	if constraints.MinValue != nil && numValue < *constraints.MinValue {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %.2f is below minimum %.2f", numValue, *constraints.MinValue))
	}
	if constraints.MaxValue != nil && numValue > *constraints.MaxValue {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Value %.2f exceeds maximum %.2f", numValue, *constraints.MaxValue))
	}

	// Validate integer constraint
	if constraints.IntegerOnly && numValue != float64(int64(numValue)) {
		result.Valid = false
		result.Errors = append(result.Errors, "Value must be an integer")
	}

	// Validate positive constraint
	if constraints.Positive && numValue <= 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Value must be positive")
	}

	// Validate decimal places
	if constraints.DecimalPlaces != nil {
		strValue := fmt.Sprintf("%.*f", *constraints.DecimalPlaces, numValue)
		if parsedValue, _ := strconv.ParseFloat(strValue, 64); parsedValue != numValue {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Value has more than %d decimal places", *constraints.DecimalPlaces))
		}
	}

	return result
}

// ValidateDateFormat validates date format against specified date format rules
func (fve *FormValidationEngine) ValidateDateFormat(dateStr string, format string) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: dateStr,
	}

	if format == "" {
		format = time.RFC3339 // Default to ISO 8601
	}

	if _, err := time.Parse(format, dateStr); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid date format, expected: %s", format))
	}

	return result
}

// ValidateEmailFormat validates email address format compliance
func (fve *FormValidationEngine) ValidateEmailFormat(email string) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: email,
	}

	if _, err := mail.ParseAddress(email); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid email format")
	}

	return result
}

// ValidateURLFormat validates URL format and structure
func (fve *FormValidationEngine) ValidateURLFormat(urlStr string) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: urlStr,
	}

	if _, err := url.Parse(urlStr); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid URL format")
	}

	return result
}

// ValidatePattern validates input against regular expression patterns
func (fve *FormValidationEngine) ValidatePattern(input string, pattern string) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: input,
	}

	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "Invalid regular expression pattern")
	} else if !matched {
		result.Valid = false
		result.Errors = append(result.Errors, "Input does not match required pattern")
	}

	return result
}

// ValidateCustomFormat validates input against user-defined format rules
func (fve *FormValidationEngine) ValidateCustomFormat(input string, formatDef CustomFormat) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: input,
	}

	if formatDef.Pattern != "" {
		return fve.ValidatePattern(input, formatDef.Pattern)
	}

	return result
}

// validateCrossFieldRule validates cross-field validation rules
func (fve *FormValidationEngine) validateCrossFieldRule(formData map[string]any, rule CrossFieldRule) *ValidationError {
	switch rule.Type {
	case CrossFieldTypeDependency:
		return fve.validateDependencyRule(formData, rule)
	case CrossFieldTypeConditional:
		return fve.validateConditionalRule(formData, rule)
	case CrossFieldTypeConsistency:
		return fve.validateConsistencyRule(formData, rule)
	default:
		return &ValidationError{
			Field:    strings.Join(rule.Fields, ","),
			Code:     "UNKNOWN_CROSS_FIELD_TYPE",
			Message:  fmt.Sprintf("Unknown cross-field validation type: %s", rule.Type),
			Severity: ErrorSeverityError,
		}
	}
}

// validateDependencyRule validates field dependency rules
func (fve *FormValidationEngine) validateDependencyRule(formData map[string]any, rule CrossFieldRule) *ValidationError {
	if len(rule.Fields) < 2 {
		return &ValidationError{
			Field:    strings.Join(rule.Fields, ","),
			Code:     "INVALID_DEPENDENCY_RULE",
			Message:  "Dependency rule requires at least 2 fields",
			Severity: ErrorSeverityError,
		}
	}

	primaryField := rule.Fields[0]
	dependentField := rule.Fields[1]

	primaryValue := formData[primaryField]
	dependentValue := formData[dependentField]

	// If primary field has value, dependent field should be required
	if primaryValue != nil && primaryValue != "" {
		if rule.Rule.Required && (dependentValue == nil || dependentValue == "") {
			return &ValidationError{
				Field:    dependentField,
				Code:     "DEPENDENT_FIELD_REQUIRED",
				Message:  fmt.Sprintf("Field '%s' is required when '%s' has a value", dependentField, primaryField),
				Severity: ErrorSeverityError,
			}
		}
	}

	return nil
}

// validateConditionalRule validates conditional validation rules
func (fve *FormValidationEngine) validateConditionalRule(formData map[string]any, rule CrossFieldRule) *ValidationError {
	// Basic conditional validation - can be extended based on specific needs
	if rule.Condition == "" {
		return nil
	}

	// Simple condition evaluation (can be made more sophisticated)
	for _, fieldName := range rule.Fields {
		value := formData[fieldName]
		if value == nil || value == "" {
			continue
		}

		fieldResult := fve.validateField(fieldName, value, rule.Rule)
		if !fieldResult.Valid {
			return &ValidationError{
				Field:    fieldName,
				Code:     "CONDITIONAL_VALIDATION_FAILED",
				Message:  fmt.Sprintf("Conditional validation failed for field '%s'", fieldName),
				Severity: ErrorSeverityError,
			}
		}
	}

	return nil
}

// validateConsistencyRule validates consistency between related fields
func (fve *FormValidationEngine) validateConsistencyRule(formData map[string]any, rule CrossFieldRule) *ValidationError {
	// Example: date range consistency (start date before end date)
	if len(rule.Fields) == 2 {
		field1Value := formData[rule.Fields[0]]
		field2Value := formData[rule.Fields[1]]

		if field1Value != nil && field2Value != nil {
			str1 := fmt.Sprintf("%v", field1Value)
			str2 := fmt.Sprintf("%v", field2Value)

			// For date consistency, try to parse as dates
			if date1, err1 := time.Parse(time.RFC3339, str1); err1 == nil {
				if date2, err2 := time.Parse(time.RFC3339, str2); err2 == nil {
					if date1.After(date2) {
						return &ValidationError{
							Field:    rule.Fields[0],
							Code:     "INCONSISTENT_DATE_RANGE",
							Message:  "Start date must be before end date",
							Severity: ErrorSeverityError,
						}
					}
				}
			}
		}
	}

	return nil
}

// SanitizeInputs sanitizes and cleans input data for safe processing
func (fve *FormValidationEngine) SanitizeInputs(inputData map[string]any) SanitizedData {
	result := SanitizedData{
		Data:     make(map[string]any),
		Changes:  []SanitizationChange{},
		Warnings: []string{},
	}

	if inputData == nil {
		result.Warnings = append(result.Warnings, "Input data is nil")
		return result
	}

	for key, value := range inputData {
		if strValue, ok := value.(string); ok {
			sanitized := fve.SanitizeHTML(strValue)
			result.Data[key] = sanitized

			if sanitized != strValue {
				result.Changes = append(result.Changes, SanitizationChange{
					Field:      key,
					Original:   strValue,
					Sanitized:  sanitized,
					ChangeType: SanitizationTypeEscape,
				})
			}
		} else {
			result.Data[key] = value
		}
	}

	return result
}

// ValidateFieldFormats validates individual field formats against format rules
func (fve *FormValidationEngine) ValidateFieldFormats(fieldMap map[string]string, formatRules map[string]FieldRule) ValidationResult {
	result := ValidationResult{
		Valid:        true,
		FieldResults: make(map[string]FieldValidationResult),
	}

	for fieldName, fieldValue := range fieldMap {
		if rule, exists := formatRules[fieldName]; exists {
			fieldResult := fve.validateField(fieldName, fieldValue, rule)
			result.FieldResults[fieldName] = fieldResult

			if !fieldResult.Valid {
				result.Valid = false
				for _, errorMsg := range fieldResult.Errors {
					result.Errors = append(result.Errors, ValidationError{
						Field:    fieldName,
						Code:     "FORMAT_ERROR",
						Message:  errorMsg,
						Severity: ErrorSeverityError,
					})
				}
			}
		}
	}

	return result
}

// ValidateStructure validates data structure against defined schema
func (fve *FormValidationEngine) ValidateStructure(data any, schema map[string]any) ValidationResult {
	result := ValidationResult{
		Valid:        true,
		FieldResults: make(map[string]FieldValidationResult),
	}

	// Basic structure validation - can be extended based on schema complexity
	if data == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:    "data",
			Code:     "NULL_DATA",
			Message:  "Data cannot be null",
			Severity: ErrorSeverityError,
		})
		return result
	}

	// Validate JSON structure if data is a map
	if dataMap, ok := data.(map[string]any); ok {
		// Check required fields from schema
		if requiredFields, exists := schema["required"]; exists {
			if required, ok := requiredFields.([]string); ok {
				for _, fieldName := range required {
					if _, fieldExists := dataMap[fieldName]; !fieldExists {
						result.Valid = false
						result.Errors = append(result.Errors, ValidationError{
							Field:    fieldName,
							Code:     "REQUIRED_FIELD_MISSING",
							Message:  fmt.Sprintf("Required field '%s' is missing", fieldName),
							Severity: ErrorSeverityError,
						})
					}
				}
			}
		}
	}

	return result
}

// ValidateCharacterSet validates input against allowed/forbidden character sets
func (fve *FormValidationEngine) ValidateCharacterSet(input string, allowedChars, forbiddenChars string) FieldValidationResult {
	result := FieldValidationResult{
		Valid: true,
		Value: input,
	}

	// Check forbidden characters
	if forbiddenChars != "" {
		for _, char := range forbiddenChars {
			if strings.ContainsRune(input, char) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Input contains forbidden character: %c", char))
			}
		}
	}

	// Check allowed characters
	if allowedChars != "" {
		for _, char := range input {
			if !strings.ContainsRune(allowedChars, char) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Input contains disallowed character: %c", char))
				break
			}
		}
	}

	return result
}

// ValidateDependentFields validates dependent field relationships
func (fve *FormValidationEngine) ValidateDependentFields(formData map[string]any, dependencies []map[string]any) ValidationResult {
	result := ValidationResult{
		Valid: true,
	}

	for _, dependency := range dependencies {
		primaryField, primaryExists := dependency["primary"].(string)
		dependentField, dependentExists := dependency["dependent"].(string)

		if !primaryExists || !dependentExists {
			continue
		}

		primaryValue := formData[primaryField]
		dependentValue := formData[dependentField]

		// If primary field has value, dependent field should also have value
		if primaryValue != nil && primaryValue != "" {
			if dependentValue == nil || dependentValue == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:    dependentField,
					Code:     "DEPENDENT_FIELD_REQUIRED",
					Message:  fmt.Sprintf("Field '%s' is required when '%s' has a value", dependentField, primaryField),
					Severity: ErrorSeverityError,
				})
			}
		}
	}

	return result
}

// ValidateConditionalRules applies conditional validation rules based on field conditions
func (fve *FormValidationEngine) ValidateConditionalRules(formData map[string]any, rules []map[string]any) ValidationResult {
	result := ValidationResult{
		Valid: true,
	}

	for _, rule := range rules {
		condition, conditionExists := rule["condition"].(string)
		field, fieldExists := rule["field"].(string)

		if !conditionExists || !fieldExists {
			continue
		}

		fieldValue := formData[field]

		// Simple condition evaluation - can be extended for more complex conditions
		if condition == "not_empty" && (fieldValue == nil || fieldValue == "") {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:    field,
				Code:     "CONDITIONAL_VALIDATION_FAILED",
				Message:  fmt.Sprintf("Field '%s' must not be empty", field),
				Severity: ErrorSeverityError,
			})
		}
	}

	return result
}

// SanitizeHTML removes or escapes HTML content from input
func (fve *FormValidationEngine) SanitizeHTML(input string) string {
	// Basic HTML sanitization - remove script tags and escape HTML entities
	input = regexp.MustCompile(`(?i)<script\b[^>]*>[\s\S]*?</script>`).ReplaceAllString(input, "")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#39;")
	return input
}

// EscapeSpecialCharacters escapes special characters according to rules
func (fve *FormValidationEngine) EscapeSpecialCharacters(input string, escapeRules map[string]string) string {
	result := input
	for char, replacement := range escapeRules {
		result = strings.ReplaceAll(result, char, replacement)
	}
	return result
}

// NormalizeUnicode normalizes Unicode text to consistent forms
func (fve *FormValidationEngine) NormalizeUnicode(input string, normForm string) string {
	// Basic Unicode normalization - can be extended with proper Unicode normalization
	if !utf8.ValidString(input) {
		// Remove invalid UTF-8 sequences
		return strings.ToValidUTF8(input, "")
	}
	return input
}

// EnforceLength enforces maximum length limits on input
func (fve *FormValidationEngine) EnforceLength(input string, maxLength int) string {
	if maxLength <= 0 {
		return input
	}

	runes := []rune(input)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return input
}

// isASCII checks if a string contains only ASCII characters
func (fve *FormValidationEngine) isASCII(s string) bool {
	for _, char := range s {
		if char > unicode.MaxASCII {
			return false
		}
	}
	return true
}
