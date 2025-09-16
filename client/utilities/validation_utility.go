// Package utilities provides Client Utility layer components for the EisenKan system following iDesign methodology.
// This package contains reusable client-side components that provide infrastructure services across all client layers.
// Following iDesign namespace: eisenkan.Client.Utilities
package utilities

import (
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ValidatorResult contains the outcome of validation operations (extended from ValidationResult)
type ValidatorResult struct {
	Valid       bool              `json:"valid"`        // Whether validation passed overall
	Errors      []string          `json:"errors"`       // List of validation errors
	Warnings    []string          `json:"warnings"`     // List of validation warnings (non-blocking)
	FieldErrors map[string]string `json:"field_errors"` // Map of field-specific validation errors
}

// ValidatorRule defines a single validation rule with parameters
type ValidatorRule struct {
	Type         string      `json:"type"`          // Rule type (Required, Range, Pattern, etc.)
	Parameters   interface{} `json:"parameters"`    // Rule-specific parameters
	ErrorMessage string      `json:"error_message"` // Custom error message template
	Condition    string      `json:"condition"`     // Optional condition for conditional validation
}

// StringConstraints defines constraints for string validation
type StringConstraints struct {
	MinLength    int    `json:"min_length"`     // Minimum character count
	MaxLength    int    `json:"max_length"`     // Maximum character count
	Pattern      string `json:"pattern"`        // Regular expression pattern
	AllowedChars string `json:"allowed_chars"`  // Character set specification
	Required     bool   `json:"required"`       // Whether empty values are allowed
}

// NumericConstraints defines constraints for numeric validation
type NumericConstraints struct {
	Min          *float64 `json:"min"`           // Minimum value (inclusive)
	Max          *float64 `json:"max"`           // Maximum value (inclusive)
	Precision    *int     `json:"precision"`     // Decimal precision requirement
	PositiveOnly bool     `json:"positive_only"` // Restrict to positive values
	IntegerOnly  bool     `json:"integer_only"`  // Restrict to integer values
}

// Constants for validation limits and constraints
const (
	MaxValidationRules = 1000 // Maximum validation rules per operation per SRS REQ-IMPL-003
	MaxStringLength    = 1024 * 1024 // 1MB maximum string length for validation
)

// ValidateString validates string constraints including length, pattern, and character set
func ValidateString(value string, constraints StringConstraints) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	// Check string length limit
	if len(value) > MaxStringLength {
		return ValidatorResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("string length %d exceeds maximum limit %d", len(value), MaxStringLength)},
		}, fmt.Errorf("ValidateString: input exceeds maximum length")
	}

	// Check required constraint
	if constraints.Required && value == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "field is required but empty")
	}

	// Check length constraints if value is not empty
	if value != "" {
		valueLength := utf8.RuneCountInString(value)

		if constraints.MinLength > 0 && valueLength < constraints.MinLength {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("string length %d is below minimum %d", valueLength, constraints.MinLength))
		}

		if constraints.MaxLength > 0 && valueLength > constraints.MaxLength {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("string length %d exceeds maximum %d", valueLength, constraints.MaxLength))
		}

		// Check pattern constraint
		if constraints.Pattern != "" {
			matched, err := regexp.MatchString(constraints.Pattern, value)
			if err != nil {
				return result, fmt.Errorf("ValidateString: invalid pattern %q: %w", constraints.Pattern, err)
			}
			if !matched {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("string does not match required pattern %q", constraints.Pattern))
			}
		}

		// Check allowed characters constraint
		if constraints.AllowedChars != "" {
			for _, r := range value {
				if !strings.ContainsRune(constraints.AllowedChars, r) {
					result.Valid = false
					result.Errors = append(result.Errors, fmt.Sprintf("character '%c' is not in allowed character set", r))
					break // Only report first invalid character
				}
			}
		}
	}

	return result, nil
}

// ValidateNumber validates numeric constraints including range, precision, and sign
func ValidateNumber(value float64, constraints NumericConstraints) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	// Check for NaN and infinity
	if value != value { // NaN check
		result.Valid = false
		result.Errors = append(result.Errors, "number is NaN (Not a Number)")
		return result, nil
	}

	// Check for infinity
	if math.IsInf(value, 0) {
		result.Valid = false
		result.Errors = append(result.Errors, "number is infinite")
		return result, nil
	}

	// Check minimum constraint
	if constraints.Min != nil && value < *constraints.Min {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("number %.6f is below minimum %.6f", value, *constraints.Min))
	}

	// Check maximum constraint
	if constraints.Max != nil && value > *constraints.Max {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("number %.6f exceeds maximum %.6f", value, *constraints.Max))
	}

	// Check positive only constraint
	if constraints.PositiveOnly && value <= 0 {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("number %.6f must be positive", value))
	}

	// Check integer only constraint
	if constraints.IntegerOnly && value != float64(int64(value)) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("number %.6f must be an integer", value))
	}

	// Check precision constraint
	if constraints.Precision != nil {
		// Convert to string and check decimal places
		str := strconv.FormatFloat(value, 'f', -1, 64)
		if strings.Contains(str, ".") {
			parts := strings.Split(str, ".")
			if len(parts) == 2 && len(parts[1]) > *constraints.Precision {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("number has %d decimal places, maximum allowed is %d", len(parts[1]), *constraints.Precision))
			}
		}
	}

	return result, nil
}

// ValidateBoolean validates and converts boolean values from various representations
func ValidateBoolean(value interface{}) (bool, ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	switch v := value.(type) {
	case bool:
		return v, result, nil
	case string:
		// Handle common string representations
		lower := strings.ToLower(strings.TrimSpace(v))
		switch lower {
		case "true", "t", "yes", "y", "1", "on":
			return true, result, nil
		case "false", "f", "no", "n", "0", "off", "":
			return false, result, nil
		default:
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("cannot convert %q to boolean", v))
			return false, result, nil
		}
	case int, int8, int16, int32, int64:
		// Convert integer types
		intVal := fmt.Sprintf("%v", v)
		switch intVal {
		case "0":
			return false, result, nil
		case "1":
			return true, result, nil
		default:
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("integer %s cannot be converted to boolean (only 0 and 1 are valid)", intVal))
			return false, result, nil
		}
	default:
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("type %T cannot be converted to boolean", value))
		return false, result, fmt.Errorf("ValidateBoolean: unsupported type %T", value)
	}
}

// ValidateDate validates date formats and ranges
func ValidateDate(value string, format string, minDate, maxDate *time.Time) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if value == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "date string is empty")
		return result, nil
	}

	// Parse the date using the provided format
	parsedDate, err := time.Parse(format, value)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("date %q does not match format %q: %v", value, format, err))
		return result, nil
	}

	// Check minimum date constraint
	if minDate != nil && parsedDate.Before(*minDate) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("date %s is before minimum date %s", parsedDate.Format(format), minDate.Format(format)))
	}

	// Check maximum date constraint
	if maxDate != nil && parsedDate.After(*maxDate) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("date %s is after maximum date %s", parsedDate.Format(format), maxDate.Format(format)))
	}

	return result, nil
}

// ValidateEmail validates email address format against RFC standards
func ValidateEmail(email string) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if email == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "email address is empty")
		return result, nil
	}

	// Use Go's mail package for RFC 5322 compliance
	_, err := mail.ParseAddress(email)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid email format: %v", err))
		return result, nil
	}

	// Additional basic checks
	if len(email) > 254 { // RFC 5321 limit
		result.Valid = false
		result.Errors = append(result.Errors, "email address exceeds maximum length of 254 characters")
	}

	if !strings.Contains(email, "@") {
		result.Valid = false
		result.Errors = append(result.Errors, "email address must contain @ symbol")
	}

	return result, nil
}

// ValidateURL validates URL format and scheme
func ValidateURL(urlStr string, allowedSchemes []string) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if urlStr == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "URL is empty")
		return result, nil
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid URL format: %v", err))
		return result, nil
	}

	// Check if scheme is provided
	if parsedURL.Scheme == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "URL must include a scheme (e.g., http://, https://)")
	}

	// Check allowed schemes if specified
	if len(allowedSchemes) > 0 {
		schemeAllowed := false
		for _, scheme := range allowedSchemes {
			if strings.EqualFold(parsedURL.Scheme, scheme) {
				schemeAllowed = true
				break
			}
		}
		if !schemeAllowed {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("URL scheme %q is not allowed, allowed schemes: %v", parsedURL.Scheme, allowedSchemes))
		}
	}

	// Check for host if it's a network URL
	if parsedURL.Scheme != "file" && parsedURL.Host == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "URL must include a host")
	}

	return result, nil
}

// ValidateUUID validates UUID format
func ValidateUUID(uuidStr string) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if uuidStr == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "UUID is empty")
		return result, nil
	}

	// Use Google's UUID package for validation
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid UUID format: %v", err))
		return result, nil
	}

	return result, nil
}

// ValidatePattern validates a value against a regular expression pattern
func ValidatePattern(value, pattern string, flags ...string) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if pattern == "" {
		return result, fmt.Errorf("ValidatePattern: pattern cannot be empty")
	}

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return result, fmt.Errorf("ValidatePattern: invalid regular expression %q: %w", pattern, err)
	}

	// Test the pattern against the value
	if !regex.MatchString(value) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("value %q does not match pattern %q", value, pattern))
	}

	return result, nil
}

// ValidateRequired validates that required fields contain non-empty, non-nil values
func ValidateRequired(value interface{}, fieldName string) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if fieldName == "" {
		fieldName = "field"
	}

	// Check for nil
	if value == nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s is required but is nil", fieldName))
		result.FieldErrors[fieldName] = "required field is nil"
		return result, nil
	}

	// Check for empty string
	if str, ok := value.(string); ok && str == "" {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s is required but is empty", fieldName))
		result.FieldErrors[fieldName] = "required field is empty"
	}

	return result, nil
}

// ValidateConditional applies validation rules based on the values of other fields
func ValidateConditional(value interface{}, fieldName string, condition func(interface{}) bool, rules []ValidatorRule) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if condition == nil {
		return result, fmt.Errorf("ValidateConditional: condition function cannot be nil")
	}

	// Check if condition is met
	if !condition(value) {
		// Condition not met, validation passes
		return result, nil
	}

	// Condition is met, apply validation rules
	if len(rules) > MaxValidationRules {
		return result, fmt.Errorf("ValidateConditional: too many validation rules %d (maximum %d)", len(rules), MaxValidationRules)
	}

	// This is a simplified implementation - in a full implementation,
	// we would apply each rule based on its type
	for _, rule := range rules {
		if rule.Type == "required" {
			requiredResult, err := ValidateRequired(value, fieldName)
			if err != nil {
				return result, err
			}
			if !requiredResult.Valid {
				result.Valid = false
				result.Errors = append(result.Errors, requiredResult.Errors...)
				for k, v := range requiredResult.FieldErrors {
					result.FieldErrors[k] = v
				}
			}
		}
		// Additional rule types would be implemented here
	}

	return result, nil
}

// ValidateCollection validates array/slice constraints and elements
func ValidateCollection(collection interface{}, minSize, maxSize int, elementValidator func(interface{}) (ValidatorResult, error)) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if collection == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "collection is nil")
		return result, nil
	}

	// Handle different collection types
	var size int
	var elements []interface{}

	switch v := collection.(type) {
	case []interface{}:
		size = len(v)
		elements = v
	case []string:
		size = len(v)
		elements = make([]interface{}, len(v))
		for i, s := range v {
			elements[i] = s
		}
	case []int:
		size = len(v)
		elements = make([]interface{}, len(v))
		for i, n := range v {
			elements[i] = n
		}
	default:
		return result, fmt.Errorf("ValidateCollection: unsupported collection type %T", collection)
	}

	// Check size constraints
	if minSize > 0 && size < minSize {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("collection size %d is below minimum %d", size, minSize))
	}

	if maxSize > 0 && size > maxSize {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("collection size %d exceeds maximum %d", size, maxSize))
	}

	// Validate individual elements if validator is provided
	if elementValidator != nil {
		for i, element := range elements {
			elementResult, err := elementValidator(element)
			if err != nil {
				return result, fmt.Errorf("ValidateCollection: error validating element %d: %w", i, err)
			}
			if !elementResult.Valid {
				result.Valid = false
				for _, errMsg := range elementResult.Errors {
					result.Errors = append(result.Errors, fmt.Sprintf("element %d: %s", i, errMsg))
				}
			}
		}
	}

	return result, nil
}

// ValidateMap validates map structure and values
func ValidateMap(mapData map[string]interface{}, requiredKeys []string, valueValidator func(string, interface{}) (ValidatorResult, error)) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if mapData == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "map is nil")
		return result, nil
	}

	// Check required keys
	for _, key := range requiredKeys {
		if _, exists := mapData[key]; !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("required key %q is missing", key))
			result.FieldErrors[key] = "required key is missing"
		}
	}

	// Validate values if validator is provided
	if valueValidator != nil {
		for key, value := range mapData {
			valueResult, err := valueValidator(key, value)
			if err != nil {
				return result, fmt.Errorf("ValidateMap: error validating key %q: %w", key, err)
			}
			if !valueResult.Valid {
				result.Valid = false
				for _, errMsg := range valueResult.Errors {
					result.Errors = append(result.Errors, fmt.Sprintf("key %q: %s", key, errMsg))
				}
				for k, v := range valueResult.FieldErrors {
					result.FieldErrors[fmt.Sprintf("%s.%s", key, k)] = v
				}
			}
		}
	}

	return result, nil
}

// ValidateUnique checks for duplicate values in collections
func ValidateUnique(collection interface{}) (ValidatorResult, error) {
	result := ValidatorResult{
		Valid:       true,
		Errors:      []string{},
		Warnings:    []string{},
		FieldErrors: make(map[string]string),
	}

	if collection == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "collection is nil")
		return result, nil
	}

	seen := make(map[interface{}]int)
	duplicates := make(map[interface{}][]int)

	// Handle different collection types
	switch v := collection.(type) {
	case []interface{}:
		for i, item := range v {
			if firstIndex, exists := seen[item]; exists {
				if _, isDuplicate := duplicates[item]; !isDuplicate {
					duplicates[item] = []int{firstIndex, i}
				} else {
					duplicates[item] = append(duplicates[item], i)
				}
			} else {
				seen[item] = i
			}
		}
	case []string:
		for i, item := range v {
			if firstIndex, exists := seen[item]; exists {
				if _, isDuplicate := duplicates[item]; !isDuplicate {
					duplicates[item] = []int{firstIndex, i}
				} else {
					duplicates[item] = append(duplicates[item], i)
				}
			} else {
				seen[item] = i
			}
		}
	case []int:
		for i, item := range v {
			if firstIndex, exists := seen[item]; exists {
				if _, isDuplicate := duplicates[item]; !isDuplicate {
					duplicates[item] = []int{firstIndex, i}
				} else {
					duplicates[item] = append(duplicates[item], i)
				}
			} else {
				seen[item] = i
			}
		}
	default:
		return result, fmt.Errorf("ValidateUnique: unsupported collection type %T", collection)
	}

	// Report duplicates
	if len(duplicates) > 0 {
		result.Valid = false
		for value, indices := range duplicates {
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate value %v found at indices %v", value, indices))
		}
	}

	return result, nil
}