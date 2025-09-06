// Package utilities_test provides constructive tests for LoggingUtility
package utilities

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
)

// TestLoggingUtility_NewLoggingUtility tests utility creation
func TestLoggingUtility_NewLoggingUtility(t *testing.T) {
	// Test default creation
	utility := NewLoggingUtility()
	if utility == nil {
		t.Fatal("Expected utility instance, got nil")
	}

	// Cleanup
	if concreteUtility, ok := utility.(*LoggingUtility); ok {
		concreteUtility.Close()
	}
}

// TestLoggingUtility_NewLoggingUtility_WithFileLogging tests file logging configuration
func TestLoggingUtility_NewLoggingUtility_WithFileLogging(t *testing.T) {
	// Create temporary log file
	tmpFile := "/tmp/test_eisenkan.log"
	defer os.Remove(tmpFile)

	// Set environment variable
	os.Setenv("LOG_FILE", tmpFile)
	defer os.Unsetenv("LOG_FILE")

	utility := NewLoggingUtility()

	// Cleanup
	if concreteUtility, ok := utility.(*LoggingUtility); ok {
		concreteUtility.Close()
	}

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

// TestLoggingUtility_IsLevelEnabled tests log level filtering
func TestLoggingUtility_IsLevelEnabled(t *testing.T) {
	// Set log level to Warning
	os.Setenv("LOG_LEVEL", "WARNING")
	defer os.Unsetenv("LOG_LEVEL")

	utility := NewLoggingUtility()
	defer func() {
		if concreteUtility, ok := utility.(*LoggingUtility); ok {
			concreteUtility.Close()
		}
	}()

	testCases := []struct {
		level    LogLevel
		expected bool
	}{
		{Debug, false},
		{Info, false},
		{Warning, true},
		{Error, true},
		{Fatal, true},
	}

	for _, tc := range testCases {
		result := utility.IsLevelEnabled(tc.level)
		if result != tc.expected {
			t.Errorf("Level %v: expected %v, got %v", tc.level, tc.expected, result)
		}
	}
}

// TestLoggingUtility_Log tests simple logging
func TestLoggingUtility_Log(t *testing.T) {
	// Capture console output
	var buf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Debug,
		consoleLogger: log.New(&buf, "", 0),
	}

	// Test simple logging without data
	utility.Log(Info, "TestComponent", "Test message", nil)

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Expected output to contain log level")
	}
	if !strings.Contains(output, "TestComponent") {
		t.Error("Expected output to contain component name")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("Expected output to contain message")
	}
}

// TestLoggingUtility_Log_WithStructuredData tests structured logging
func TestLoggingUtility_Log_WithStructuredData(t *testing.T) {
	var buf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Debug,
		consoleLogger: log.New(&buf, "", 0),
	}

	// Test with simple data types
	testData := map[string]interface{}{
		"stringField": "test value",
		"intField":    42,
		"boolField":   true,
	}

	utility.Log(Info, "TestComponent", "Structured test message", testData)

	output := buf.String()
	expectedStrings := []string{
		"INFO",
		"Structured test message",
		"TestComponent",
		"stringField",
		"test value",
		"intField",
		"42",
		"boolField",
		"true",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// TestLoggingUtility_Log_WithVariousMapTypes tests different map key types
func TestLoggingUtility_Log_WithVariousMapTypes(t *testing.T) {
	var buf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Debug,
		consoleLogger: log.New(&buf, "", 0),
	}

	testCases := []struct {
		name string
		data interface{}
		contains []string
	}{
		{
			name: "StringMap",
			data: map[string]interface{}{"key": "value"},
			contains: []string{"key", "value"},
		},
		{
			name: "IntMap", 
			data: map[int]interface{}{1: "first", 2: "second"},
			contains: []string{"1", "first", "2", "second"},
		},
		{
			name: "GenericMap",
			data: map[interface{}]interface{}{42: "answer", "key": 123},
			contains: []string{"42", "answer", "key", "123"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			utility.Log(Info, "TestComponent", "Map test", tc.data)
			output := buf.String()
			
			for _, expected := range tc.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

// TestLoggingUtility_LogError tests error logging with stack trace
func TestLoggingUtility_LogError(t *testing.T) {
	var buf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Debug,
		consoleLogger: log.New(&buf, "", 0),
	}

	testError := fmt.Errorf("test error message")
	testData := map[string]interface{}{
		"errorCode": 500,
	}

	utility.LogError("ErrorComponent", testError, testData)

	output := buf.String()
	expectedStrings := []string{
		"ERROR",
		"Error: test error message",
		"ErrorComponent",
		"errorCode",
		"500",
		".go:", // Should contain file reference
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

// TestLoggingUtility_SerializeData_DepthLimiting tests depth limiting
func TestLoggingUtility_SerializeData_DepthLimiting(t *testing.T) {
	utility := &LoggingUtility{}
	
	// Create deeply nested structure
	deepData := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"level5": map[string]interface{}{
							"level6": "too deep",
						},
					},
				},
			},
		},
	}

	result := utility.serializeData(deepData, 0)
	
	// Should contain max depth reached indicator
	if !strings.Contains(result, "[max_depth_reached]") {
		t.Error("Expected depth limiting to trigger, but it didn't")
	}
}

// TestLoggingUtility_ThreadSafety tests concurrent logging
func TestLoggingUtility_ThreadSafety(t *testing.T) {
	var buf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Debug,
		consoleLogger: log.New(&buf, "", 0),
	}

	const numGoroutines = 10
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range messagesPerGoroutine {
				utility.Log(Info, fmt.Sprintf("Component%d", id), fmt.Sprintf("Message %d", j), nil)
			}
		}(i)
	}

	wg.Wait()

	// Count log lines
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	expectedLines := numGoroutines * messagesPerGoroutine

	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify each line contains expected elements
	for i, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "INFO") {
			t.Errorf("Line %d missing log level: %s", i, line)
		}
		if !strings.Contains(line, "Component") {
			t.Errorf("Line %d missing component: %s", i, line)
		}
		if !strings.Contains(line, "Message") {
			t.Errorf("Line %d missing message: %s", i, line)
		}
	}
}

// TestLoggingUtility_LevelFiltering tests that levels below minimum are filtered
func TestLoggingUtility_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Warning, // Only Warning and above
		consoleLogger: log.New(&buf, "", 0),
	}

	// These should not appear
	utility.Log(Debug, "Component", "Debug message", nil)
	utility.Log(Info, "Component", "Info message", nil)

	// These should appear
	utility.Log(Warning, "Component", "Warning message", nil)
	utility.Log(Error, "Component", "Error message", nil)

	output := buf.String()
	
	// Should not contain filtered levels
	if strings.Contains(output, "Debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "Info message") {
		t.Error("Info message should be filtered out")
	}

	// Should contain allowed levels
	if !strings.Contains(output, "Warning message") {
		t.Error("Warning message should be present")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Error message should be present")
	}
}

// TestLoggingUtility_FileAndConsoleOutput tests dual output functionality
func TestLoggingUtility_FileAndConsoleOutput(t *testing.T) {
	// Create temporary log file
	tmpFile := "/tmp/test_dual_output.log"
	defer os.Remove(tmpFile)

	// Create file logger
	file, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var consoleBuf bytes.Buffer
	utility := &LoggingUtility{
		minLevel:      Debug,
		consoleLogger: log.New(&consoleBuf, "", 0),
		fileLogger:    log.New(file, "", 0),
		logFile:       file,
	}

	testMessage := "Dual output test message"
	utility.Log(Info, "TestComponent", testMessage, nil)

	// Close file to flush
	utility.Close()

	// Check console output
	consoleOutput := consoleBuf.String()
	if !strings.Contains(consoleOutput, testMessage) {
		t.Error("Message not found in console output")
	}

	// Check file output
	fileContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	fileOutput := string(fileContent)
	if !strings.Contains(fileOutput, testMessage) {
		t.Error("Message not found in file output")
	}
}

// TestLogLevel_String tests LogLevel string representation
func TestLogLevel_String(t *testing.T) {
	testCases := []struct {
		level    LogLevel
		expected string
	}{
		{Debug, "DEBUG"},
		{Info, "INFO"},
		{Warning, "WARNING"},
		{Error, "ERROR"},
		{Fatal, "FATAL"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tc := range testCases {
		result := tc.level.String()
		if result != tc.expected {
			t.Errorf("Level %d: expected %s, got %s", tc.level, tc.expected, result)
		}
	}
}

// TestGetLogLevelFromEnv tests environment variable parsing
func TestGetLogLevelFromEnv(t *testing.T) {
	testCases := []struct {
		envValue string
		expected LogLevel
	}{
		{"", Info},         // Default
		{"DEBUG", Debug},
		{"INFO", Info},
		{"WARNING", Warning},
		{"ERROR", Error},
		{"FATAL", Fatal},
		{"INVALID", Info},  // Invalid defaults to Info
	}

	for _, tc := range testCases {
		if tc.envValue == "" {
			os.Unsetenv("LOG_LEVEL")
		} else {
			os.Setenv("LOG_LEVEL", tc.envValue)
		}

		result := getLogLevelFromEnv()
		if result != tc.expected {
			t.Errorf("Env value '%s': expected %v, got %v", tc.envValue, tc.expected, result)
		}

		os.Unsetenv("LOG_LEVEL")
	}
}

// TestLoggingUtility_InvalidFilePathPanic tests that invalid file paths cause panic
func TestLoggingUtility_InvalidFilePathPanic(t *testing.T) {
	// Set invalid file path
	os.Setenv("LOG_FILE", "/invalid/path/that/does/not/exist/test.log")
	defer os.Unsetenv("LOG_FILE")

	// Should panic due to our fail-fast design
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid file path, but no panic occurred")
		}
	}()

	NewLoggingUtility()
}