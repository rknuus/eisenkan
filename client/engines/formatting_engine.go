package engines

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rknuus/eisenkan/client/utilities"
)

// FormattingEngine provides comprehensive formatting capabilities for UI components
type FormattingEngine struct {
	text          IText
	number        INumber
	time          ITime
	datastructure IDatastructure
	template      ITemplate
	locale        ILocale
}

// NewFormattingEngine creates a new FormattingEngine instance
func NewFormattingEngine() *FormattingEngine {
	locale := &LocaleFacet{
		currentLocale: "en-US",
		mu:            &sync.RWMutex{},
	}

	return &FormattingEngine{
		text:          &TextFacet{locale: locale},
		number:        &NumberFacet{locale: locale},
		time:          &TimeFacet{locale: locale},
		datastructure: &DatastructureFacet{},
		template:      &TemplateFacet{cache: make(map[string]*CompiledTemplate), mu: &sync.RWMutex{}},
		locale:        locale,
	}
}

// Text returns the text formatting facet
func (fe *FormattingEngine) Text() IText {
	return fe.text
}

// Number returns the numeric formatting facet
func (fe *FormattingEngine) Number() INumber {
	return fe.number
}

// Time returns the temporal formatting facet
func (fe *FormattingEngine) Time() ITime {
	return fe.time
}

// Datastructure returns the data structure formatting facet
func (fe *FormattingEngine) Datastructure() IDatastructure {
	return fe.datastructure
}

// Template returns the template processing facet
func (fe *FormattingEngine) Template() ITemplate {
	return fe.template
}

// Locale returns the locale configuration facet
func (fe *FormattingEngine) Locale() ILocale {
	return fe.locale
}

// IText defines the text formatting interface
type IText interface {
	FormatText(input string, options TextOptions) (string, error)
	FormatLabel(fieldName string) string
	FormatMessage(template string, params map[string]any) (string, error)
	FormatError(err error, severity FormattingErrorSeverity) FormattedError
}

// INumber defines the numeric formatting interface
type INumber interface {
	FormatNumber(value any, precision int) (string, error)
	FormatPercentage(value float64, precision int) string
	FormatFileSize(bytes int64, unit FileSizeUnit) string
	FormatCurrency(value float64, currency string) (string, error)
}

// ITime defines the temporal formatting interface
type ITime interface {
	FormatDateTime(t time.Time, format string) string
	FormatDuration(d time.Duration) string
	FormatRelativeTime(t time.Time) string
	FormatTimeRange(start, end time.Time) string
}

// IDatastructure defines the data structure formatting interface
type IDatastructure interface {
	FormatList(items []any, options ListOptions) string
	FormatKeyValue(data map[string]any, options KeyValueOptions) string
	FormatJSON(data any, indent bool) (string, error)
	FormatHierarchy(data any, maxDepth int) string
}

// ITemplate defines the template processing interface
type ITemplate interface {
	ProcessTemplate(template string, data map[string]any) (string, error)
	ValidateTemplate(template string) error
	CacheTemplate(name string, template string) error
	GetTemplateMetadata(template string) TemplateMetadata
}

// ILocale defines the locale configuration interface
type ILocale interface {
	SetLocale(locale string) error
	SetNumberFormat(decimal, thousand string) error
	SetDateFormat(format string) error
	SetCurrencyFormat(currency, symbol string) error
	GetLocale() string
}

// Supporting types and enums

// TextOptions configures text formatting behavior
type TextOptions struct {
	CaseTransform CaseTransform
	MaxLength     int
	WrapWidth     int
	TrimSpaces    bool
	EscapeHTML    bool
}

// CaseTransform defines text case transformations
type CaseTransform int

const (
	CaseNone CaseTransform = iota
	CaseUpper
	CaseLower
	CaseTitle
	CaseSentence
)

// FormattingErrorSeverity defines error message severity levels
type FormattingErrorSeverity int

const (
	FormattingSeverityInfo FormattingErrorSeverity = iota
	FormattingSeverityWarning
	FormattingSeverityError
	FormattingSeverityCritical
)

// FormattedError represents a formatted error message
type FormattedError struct {
	Message   string                  `json:"message"`
	Severity  FormattingErrorSeverity `json:"severity"`
	Timestamp time.Time               `json:"timestamp"`
	Code      string                  `json:"code,omitempty"`
}

// FileSizeUnit defines file size display units
type FileSizeUnit int

const (
	FileSizeAuto FileSizeUnit = iota
	FileSizeBytes
	FileSizeKB
	FileSizeMB
	FileSizeGB
	FileSizeTB
)

// ListOptions configures list formatting
type ListOptions struct {
	Separator   string
	Prefix      string
	Suffix      string
	MaxItems    int
	ShowIndices bool
}

// KeyValueOptions configures key-value formatting
type KeyValueOptions struct {
	Separator    string
	Assignment   string
	SortKeys     bool
	MaxKeyLength int
}

// TemplateMetadata provides information about template structure
type TemplateMetadata struct {
	Parameters []string `json:"parameters"`
	Functions  []string `json:"functions"`
	Complexity int      `json:"complexity"`
}

// CompiledTemplate represents a cached template
type CompiledTemplate struct {
	Template   string
	Parameters []string
	CompiledAt time.Time
}

// TextFacet implements IText interface
type TextFacet struct {
	locale ILocale
}

// FormatText applies text transformations to input string
func (tf *TextFacet) FormatText(input string, options TextOptions) (string, error) {
	if input == "" {
		return "", nil
	}

	result := input

	// Trim spaces if requested
	if options.TrimSpaces {
		var err error
		result, err = utilities.TrimText(result)
		if err != nil {
			return "", err
		}
	}

	// Apply case transformation
	switch options.CaseTransform {
	case CaseUpper:
		result, _ = utilities.ConvertCase(result, utilities.TextCaseUpper)
	case CaseLower:
		result, _ = utilities.ConvertCase(result, utilities.TextCaseLower)
	case CaseTitle:
		result, _ = utilities.ConvertCase(result, utilities.TextCaseTitle)
	case CaseSentence:
		result, _ = utilities.ConvertCase(result, utilities.TextCaseSentence)
	}

	// Apply length limit
	if options.MaxLength > 0 && len(result) > options.MaxLength {
		var err error
		result, err = utilities.TruncateText(result, options.MaxLength, "...")
		if err != nil {
			return "", err
		}
	}

	// Apply wrapping
	if options.WrapWidth > 0 {
		var err error
		result, err = utilities.WrapText(result, options.WrapWidth)
		if err != nil {
			return "", err
		}
	}

	// Escape HTML if requested
	if options.EscapeHTML {
		result = html.EscapeString(result)
	}

	return result, nil
}

// FormatLabel generates consistent field labels
func (tf *TextFacet) FormatLabel(fieldName string) string {
	if fieldName == "" {
		return ""
	}

	// Convert camelCase/snake_case to Title Case
	result := regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(fieldName, "$1 $2")
	result = strings.ReplaceAll(result, "_", " ")
	result = strings.ReplaceAll(result, "-", " ")

	// Convert to title case
	words := strings.Fields(result)
	for i, word := range words {
		if len(word) > 0 {
			// Special case for all-caps words like "ID"
			if word == strings.ToUpper(word) && len(word) <= 3 {
				words[i] = word
			} else {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
	}

	return strings.Join(words, " ")
}

// FormatMessage processes template-based messages with parameter substitution
func (tf *TextFacet) FormatMessage(template string, params map[string]any) (string, error) {
	if template == "" {
		return "", nil
	}

	result := template
	for key, value := range params {
		placeholder := "{" + key + "}"
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	return result, nil
}

// FormatError standardizes error message presentation
func (tf *TextFacet) FormatError(err error, severity FormattingErrorSeverity) FormattedError {
	return FormattedError{
		Message:   err.Error(),
		Severity:  severity,
		Timestamp: time.Now(),
	}
}

// NumberFacet implements INumber interface
type NumberFacet struct {
	locale ILocale
}

// FormatNumber displays numbers with appropriate precision
func (nf *NumberFacet) FormatNumber(value any, precision int) (string, error) {
	switch v := value.(type) {
	case int:
		return utilities.FormatNumber(float64(v), precision, ",")
	case int64:
		return utilities.FormatNumber(float64(v), precision, ",")
	case float32:
		return utilities.FormatNumber(float64(v), precision, ",")
	case float64:
		return utilities.FormatNumber(v, precision, ",")
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number format: %s", v)
		}
		return utilities.FormatNumber(parsed, precision, ",")
	default:
		return "", fmt.Errorf("unsupported number type: %T", value)
	}
}

// FormatPercentage converts ratios to percentage format
func (nf *NumberFacet) FormatPercentage(value float64, precision int) string {
	result, _ := utilities.FormatPercentage(value, precision)
	return result
}

// FormatFileSize presents byte counts as human-readable sizes
func (nf *NumberFacet) FormatFileSize(bytes int64, unit FileSizeUnit) string {
	unitType := utilities.FileSizeAuto
	switch unit {
	case FileSizeBytes:
		unitType = utilities.FileSizeBytes
	case FileSizeKB:
		unitType = utilities.FileSizeKB
	case FileSizeMB:
		unitType = utilities.FileSizeMB
	case FileSizeGB:
		unitType = utilities.FileSizeGB
	case FileSizeTB:
		unitType = utilities.FileSizeTB
	}

	result, _ := utilities.FormatFileSize(bytes, unitType, 2)
	return result
}

// FormatCurrency displays monetary values with proper currency symbols
func (nf *NumberFacet) FormatCurrency(value float64, currency string) (string, error) {
	symbol := "$" // Default to USD
	switch currency {
	case "USD":
		symbol = "$"
	case "EUR":
		symbol = "€"
	case "GBP":
		symbol = "£"
	case "JPY":
		symbol = "¥"
	}

	formatted, err := utilities.FormatNumber(value, 2, ",")
	if err != nil {
		return "", err
	}
	return symbol + formatted, nil
}

// TimeFacet implements ITime interface
type TimeFacet struct {
	locale ILocale
}

// FormatDateTime presents timestamps in user-friendly formats
func (tf *TimeFacet) FormatDateTime(t time.Time, format string) string {
	if format == "" {
		format = time.RFC3339
	}

	result, _ := utilities.FormatDateTime(t, format)
	return result
}

// FormatDuration converts time spans into readable strings
func (tf *TimeFacet) FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		if hours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

// FormatRelativeTime generates relative time descriptions
func (tf *TimeFacet) FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		diff = -diff
		if diff < time.Minute {
			return "in a few seconds"
		} else if diff < time.Hour {
			minutes := int(diff.Minutes())
			return fmt.Sprintf("in %d minutes", minutes)
		} else if diff < 24*time.Hour {
			hours := int(diff.Hours())
			return fmt.Sprintf("in %d hours", hours)
		} else {
			days := int(diff.Hours()) / 24
			return fmt.Sprintf("in %d days", days)
		}
	}

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(diff.Hours()) / 24
		return fmt.Sprintf("%d days ago", days)
	}
}

// FormatTimeRange displays time periods
func (tf *TimeFacet) FormatTimeRange(start, end time.Time) string {
	startStr := tf.FormatDateTime(start, "")
	endStr := tf.FormatDateTime(end, "")
	return fmt.Sprintf("%s - %s", startStr, endStr)
}

// DatastructureFacet implements IDatastructure interface
type DatastructureFacet struct{}

// FormatList presents arrays as organized, readable lists
func (df *DatastructureFacet) FormatList(items []any, options ListOptions) string {
	if len(items) == 0 {
		return ""
	}

	// Apply item limit
	displayItems := items
	if options.MaxItems > 0 && len(items) > options.MaxItems {
		displayItems = items[:options.MaxItems]
	}

	// Convert items to strings
	var parts []string
	for i, item := range displayItems {
		var itemStr string
		if options.ShowIndices {
			itemStr = fmt.Sprintf("%d. %v", i+1, item)
		} else {
			itemStr = fmt.Sprintf("%v", item)
		}
		parts = append(parts, itemStr)
	}

	// Add truncation indicator if needed
	if options.MaxItems > 0 && len(items) > options.MaxItems {
		remaining := len(items) - options.MaxItems
		parts = append(parts, fmt.Sprintf("... and %d more", remaining))
	}

	separator := options.Separator
	if separator == "" {
		separator = ", "
	}

	result := strings.Join(parts, separator)

	if options.Prefix != "" {
		result = options.Prefix + result
	}
	if options.Suffix != "" {
		result = result + options.Suffix
	}

	return result
}

// FormatKeyValue displays key-value pairs in structured format
func (df *DatastructureFacet) FormatKeyValue(data map[string]any, options KeyValueOptions) string {
	if len(data) == 0 {
		return ""
	}

	// Get keys
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	// Sort keys if requested
	if options.SortKeys {
		sort.Strings(keys)
	}

	// Format pairs
	separator := options.Separator
	if separator == "" {
		separator = "\n"
	}

	assignment := options.Assignment
	if assignment == "" {
		assignment = ": "
	}

	var parts []string
	for _, key := range keys {
		value := data[key]

		// Apply key length limit
		displayKey := key
		if options.MaxKeyLength > 0 && len(key) > options.MaxKeyLength {
			displayKey = key[:options.MaxKeyLength] + "..."
		}

		pair := fmt.Sprintf("%s%s%v", displayKey, assignment, value)
		parts = append(parts, pair)
	}

	return strings.Join(parts, separator)
}

// FormatJSON converts data structures to formatted JSON
func (df *DatastructureFacet) FormatJSON(data any, indent bool) (string, error) {
	var result []byte
	var err error

	if indent {
		result, err = json.MarshalIndent(data, "", "  ")
	} else {
		result, err = json.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("JSON formatting error: %w", err)
	}

	return string(result), nil
}

// FormatHierarchy presents nested data with proper indentation
func (df *DatastructureFacet) FormatHierarchy(data any, maxDepth int) string {
	return df.formatHierarchyRecursive(data, 0, maxDepth)
}

func (df *DatastructureFacet) formatHierarchyRecursive(data any, depth, maxDepth int) string {
	if maxDepth > 0 && depth >= maxDepth {
		return fmt.Sprintf("%s...", strings.Repeat("  ", depth))
	}

	indent := strings.Repeat("  ", depth)

	switch v := data.(type) {
	case map[string]any:
		if len(v) == 0 {
			return indent + "{}"
		}

		var parts []string
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := v[key]
			valueStr := df.formatHierarchyRecursive(value, depth+1, maxDepth)
			parts = append(parts, fmt.Sprintf("%s%s:\n%s", indent, key, valueStr))
		}
		return strings.Join(parts, "\n")

	case []any:
		if len(v) == 0 {
			return indent + "[]"
		}

		var parts []string
		for i, item := range v {
			itemStr := df.formatHierarchyRecursive(item, depth+1, maxDepth)
			parts = append(parts, fmt.Sprintf("%s[%d]:\n%s", indent, i, itemStr))
		}
		return strings.Join(parts, "\n")

	default:
		return fmt.Sprintf("%s%v", indent, v)
	}
}

// TemplateFacet implements ITemplate interface
type TemplateFacet struct {
	cache map[string]*CompiledTemplate
	mu    *sync.RWMutex
}

// ProcessTemplate replaces template placeholders with formatted values
func (tf *TemplateFacet) ProcessTemplate(template string, data map[string]any) (string, error) {
	if template == "" {
		return "", nil
	}

	result := template
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		replacement := html.EscapeString(fmt.Sprintf("%v", value))
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	// Check for unresolved placeholders
	if strings.Contains(result, "{{") && strings.Contains(result, "}}") {
		return result, fmt.Errorf("template contains unresolved placeholders")
	}

	return result, nil
}

// ValidateTemplate verifies template syntax and parameter compatibility
func (tf *TemplateFacet) ValidateTemplate(template string) error {
	if template == "" {
		return nil
	}

	// Check for balanced braces
	openCount := strings.Count(template, "{{")
	closeCount := strings.Count(template, "}}")

	if openCount != closeCount {
		return fmt.Errorf("unbalanced template braces: %d open, %d close", openCount, closeCount)
	}

	// Check for valid parameter names
	re := regexp.MustCompile(`\{\{([^}]*)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			param := strings.TrimSpace(match[1])
			if param == "" {
				return fmt.Errorf("empty template parameter")
			}
			// Simple validation - only allow alphanumeric and underscore
			if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(param) {
				return fmt.Errorf("invalid parameter name: %s", param)
			}
		}
	}

	return nil
}

// CacheTemplate stores compiled templates for repeated use
func (tf *TemplateFacet) CacheTemplate(name string, template string) error {
	if err := tf.ValidateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Extract parameters
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	var params []string
	paramSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			param := strings.TrimSpace(match[1])
			if !paramSet[param] {
				params = append(params, param)
				paramSet[param] = true
			}
		}
	}

	compiled := &CompiledTemplate{
		Template:   template,
		Parameters: params,
		CompiledAt: time.Now(),
	}

	tf.mu.Lock()
	tf.cache[name] = compiled
	tf.mu.Unlock()

	return nil
}

// GetTemplateMetadata retrieves information about template parameters
func (tf *TemplateFacet) GetTemplateMetadata(template string) TemplateMetadata {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	var params []string
	paramSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			param := strings.TrimSpace(match[1])
			if !paramSet[param] {
				params = append(params, param)
				paramSet[param] = true
			}
		}
	}

	return TemplateMetadata{
		Parameters: params,
		Functions:  []string{}, // No functions in simple template engine
		Complexity: len(params) + strings.Count(template, "{{"),
	}
}

// LocaleFacet implements ILocale interface
type LocaleFacet struct {
	currentLocale string
	numberFormat  struct {
		decimal  string
		thousand string
	}
	dateFormat     string
	currencyFormat map[string]string
	mu             *sync.RWMutex
}

// SetLocale configures locale-specific formatting preferences
func (lf *LocaleFacet) SetLocale(locale string) error {
	if locale == "" {
		return fmt.Errorf("locale cannot be empty")
	}

	lf.mu.Lock()
	defer lf.mu.Unlock()

	lf.currentLocale = locale

	// Set default formats based on locale
	switch locale {
	case "en-US":
		lf.numberFormat.decimal = "."
		lf.numberFormat.thousand = ","
		lf.dateFormat = "MM/DD/YYYY"
	case "en-GB":
		lf.numberFormat.decimal = "."
		lf.numberFormat.thousand = ","
		lf.dateFormat = "DD/MM/YYYY"
	case "de-DE":
		lf.numberFormat.decimal = ","
		lf.numberFormat.thousand = "."
		lf.dateFormat = "DD.MM.YYYY"
	case "fr-FR":
		lf.numberFormat.decimal = ","
		lf.numberFormat.thousand = " "
		lf.dateFormat = "DD/MM/YYYY"
	default:
		// Default to en-US
		lf.numberFormat.decimal = "."
		lf.numberFormat.thousand = ","
		lf.dateFormat = "MM/DD/YYYY"
	}

	return nil
}

// SetNumberFormat defines numeric formatting rules
func (lf *LocaleFacet) SetNumberFormat(decimal, thousand string) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	lf.numberFormat.decimal = decimal
	lf.numberFormat.thousand = thousand
	return nil
}

// SetDateFormat specifies date display preferences
func (lf *LocaleFacet) SetDateFormat(format string) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	lf.dateFormat = format
	return nil
}

// SetCurrencyFormat configures monetary value presentation
func (lf *LocaleFacet) SetCurrencyFormat(currency, symbol string) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.currencyFormat == nil {
		lf.currencyFormat = make(map[string]string)
	}
	lf.currencyFormat[currency] = symbol
	return nil
}

// GetLocale returns the current locale setting
func (lf *LocaleFacet) GetLocale() string {
	lf.mu.RLock()
	defer lf.mu.RUnlock()
	return lf.currentLocale
}
