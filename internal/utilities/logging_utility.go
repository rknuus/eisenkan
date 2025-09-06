// Package utilities provides cross-cutting infrastructure components
// Following iDesign namespace: eisenkan.Utilities
package utilities

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warning
	Error
	Fatal
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ILoggingUtility defines the contract for logging operations
type ILoggingUtility interface {
	// Log records event with message and optional arbitrary data
	Log(level LogLevel, component string, message string, data interface{})

	// LogError records error with automatic stack trace capture (always at Error level)
	LogError(component string, err error, data interface{})

	// IsLevelEnabled checks if log level would be output (performance optimization)
	IsLevelEnabled(level LogLevel) bool
}

// LoggingUtility implements ILoggingUtility
type LoggingUtility struct {
	minLevel      LogLevel
	consoleLogger *log.Logger
	fileLogger    *log.Logger
	mu            sync.RWMutex
	logFile       *os.File
}

// NewLoggingUtility creates a new LoggingUtility instance
// Configuration is handled through environment variables following stateless design
func NewLoggingUtility() ILoggingUtility {
	utility := &LoggingUtility{
		minLevel:      getLogLevelFromEnv(),
		consoleLogger: log.New(os.Stdout, "", 0), // Custom format will be applied
	}

	// Setup file logging if configured
	if logFilePath := os.Getenv("LOG_FILE"); logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("failed to open log file %s: %v", logFilePath, err))
		}
		utility.logFile = file
		utility.fileLogger = log.New(file, "", 0)
	}

	return utility
}

// Close closes the log file if it was opened
func (l *LoggingUtility) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// IsLevelEnabled checks if a log level would be output (for performance optimization)
func (l *LoggingUtility) IsLevelEnabled(level LogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.minLevel
}

// Log records event with message and optional arbitrary data
func (l *LoggingUtility) Log(level LogLevel, component string, message string, data interface{}) {
	if !l.IsLevelEnabled(level) {
		return
	}

	logEntry := l.formatLog(level, component, message, data)
	l.writeLog(logEntry)
}

// LogError records error with automatic stack trace capture (always at Error level)
func (l *LoggingUtility) LogError(component string, err error, data interface{}) {
	if !l.IsLevelEnabled(Error) {
		return
	}

	// Get caller information for stack trace
	_, file, line, ok := runtime.Caller(1)
	stackInfo := ""
	if ok {
		stackInfo = fmt.Sprintf(" [%s:%d]", file, line)
	}

	message := fmt.Sprintf("Error: %v%s", err, stackInfo)
	logEntry := l.formatLog(Error, component, message, data)
	l.writeLog(logEntry)
}

// formatLog creates a formatted log entry with optional structured data
func (l *LoggingUtility) formatLog(level LogLevel, component string, message string, data interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	baseLog := fmt.Sprintf("%s [%s] %s: %s", timestamp, level, component, message)

	// Add structured data if provided
	if data != nil {
		structuredData := l.serializeData(data, 0)
		if structuredData != "" {
			baseLog += fmt.Sprintf(" | %s", structuredData)
		}
	}

	return baseLog
}

// serializeData converts arbitrary data to structured format with depth limiting
func (l *LoggingUtility) serializeData(data interface{}, depth int) string {
	// Depth limiting to prevent infinite recursion and verbosity
	if depth >= 5 {
		return "... [max_depth_reached]"
	}

	if data == nil {
		return "null"
	}

	// Type switch for common cases (performance optimization)
	// Benefits: No reflection overhead, no JSON marshaling for primitives, CPU branch prediction
	switch v := data.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case []interface{}: // Slice of interface{} - common from JSON unmarshaling
		return l.serializeSlice(v, depth+1)
	case map[string]interface{}: // Map with string keys, interface{} values - JSON object pattern
		return l.serializeStringMap(v, depth+1)
	case map[int]interface{}: // Map with int keys
		return l.serializeIntMap(v, depth+1)
	case map[interface{}]interface{}: // Map with arbitrary keys
		return l.serializeGenericMap(v, depth+1)
	default:
		// For complex types, try JSON encoding
		jsonData, err := json.Marshal(v)
		if err != nil {
			// Fallback to string representation
			return fmt.Sprintf("\"%v\"", v)
		}
		return string(jsonData)
	}
}

// serializeSlice handles slice serialization with depth limiting
func (l *LoggingUtility) serializeSlice(slice []interface{}, depth int) string {
	if depth >= 5 {
		return "... [max_depth_reached]"
	}

	result := "["
	for i, item := range slice {
		if i > 0 {
			result += ","
		}
		result += l.serializeData(item, depth)
	}
	result += "]"
	return result
}

// serializeStringMap handles map[string]interface{} serialization with depth limiting
func (l *LoggingUtility) serializeStringMap(m map[string]interface{}, depth int) string {
	if depth >= 5 {
		return "... [max_depth_reached]"
	}

	result := "{"
	first := true
	for key, value := range m {
		if !first {
			result += ","
		}
		first = false
		result += fmt.Sprintf("\"%s\":%s", key, l.serializeData(value, depth))
	}
	result += "}"
	return result
}

// serializeIntMap handles map[int]interface{} serialization with depth limiting
func (l *LoggingUtility) serializeIntMap(m map[int]interface{}, depth int) string {
	if depth >= 5 {
		return "... [max_depth_reached]"
	}

	result := "{"
	first := true
	for key, value := range m {
		if !first {
			result += ","
		}
		first = false
		result += fmt.Sprintf("\"%d\":%s", key, l.serializeData(value, depth))
	}
	result += "}"
	return result
}

// serializeGenericMap handles map[interface{}]interface{} serialization with depth limiting
func (l *LoggingUtility) serializeGenericMap(m map[interface{}]interface{}, depth int) string {
	if depth >= 5 {
		return "... [max_depth_reached]"
	}

	result := "{"
	first := true
	for key, value := range m {
		if !first {
			result += ","
		}
		first = false
		// Serialize the key as well, since it can be any type
		keyStr := l.serializeData(key, depth)
		valueStr := l.serializeData(value, depth)
		result += fmt.Sprintf("%s:%s", keyStr, valueStr)
	}
	result += "}"
	return result
}

// writeLog writes the log entry to all configured outputs
func (l *LoggingUtility) writeLog(logEntry string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write to console
	l.consoleLogger.Println(logEntry)

	// Write to file if configured
	if l.fileLogger != nil {
		l.fileLogger.Println(logEntry)
	}
}

// getLogLevelFromEnv reads the log level from environment variables
func getLogLevelFromEnv() LogLevel {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		return Info // Default level
	}

	switch levelStr {
	case "DEBUG":
		return Debug
	case "INFO":
		return Info
	case "WARNING":
		return Warning
	case "ERROR":
		return Error
	case "FATAL":
		return Fatal
	default:
		return Info
	}
}
