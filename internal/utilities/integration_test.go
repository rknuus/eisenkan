// Package utilities_test provides integration tests for LoggingUtility
// These tests validate the utility works correctly within the system architecture
package utilities

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestLoggingUtility_Integration_ArchitecturalCompliance validates architectural layer rules
func TestIntegration_LoggingUtility_ArchitecturalCompliance(t *testing.T) {
	// Test: Utilities can be called by all layers
	logger := NewLoggingUtility()
	defer func() {
		if concreteLogger, ok := logger.(*LoggingUtility); ok {
			concreteLogger.Close()
		}
	}()

	// Simulate calls from different architectural layers
	testCases := []struct {
		layer     string
		component string
	}{
		{"Client", "EisenKanClient"},
		{"Manager", "TaskManager"},
		{"Engine", "ValidationEngine"},
		{"ResourceAccess", "TasksAccess"},
	}

	for _, tc := range testCases {
		// Each layer should be able to use the logging utility
		logger.Log(Info, tc.component, fmt.Sprintf("Test message from %s layer", tc.layer), nil)

		// Structured logging should work from any layer
		data := map[string]interface{}{
			"layer": tc.layer,
			"test":  true,
		}
		logger.Log(Info, tc.component, "Integration test message", data)
	}
}

// TestLoggingUtility_Integration_PerformanceImpact validates logging doesn't degrade system performance
func TestIntegration_LoggingUtility_PerformanceImpact(t *testing.T) {
	logger := NewLoggingUtility()
	defer func() {
		if concreteLogger, ok := logger.(*LoggingUtility); ok {
			concreteLogger.Close()
		}
	}()

	// Measure performance impact
	const iterations = 10000

	// Measure baseline (no logging)
	start := time.Now()
	for range iterations {
		// Simulate business operation
		_ = fmt.Sprintf("Business operation %d", 42)
	}
	baselineDuration := time.Since(start)

	// Measure with logging
	start = time.Now()
	for i := range iterations {
		// Simulate business operation with logging
		result := fmt.Sprintf("Business operation %d", i)
		logger.Log(Debug, "PerformanceTest", result, nil)
	}
	loggingDuration := time.Since(start)

	// Logging overhead should be reasonable (less than 400% of baseline)
	overhead := float64(loggingDuration) / float64(baselineDuration)
	if overhead > 4.0 {
		t.Errorf("Logging overhead too high: %.2fx (baseline: %v, with logging: %v)",
			overhead, baselineDuration, loggingDuration)
	}

	t.Logf("Performance test: baseline=%v, logging=%v, overhead=%.2fx",
		baselineDuration, loggingDuration, overhead)
}

// TestLoggingUtility_Integration_ConfigurationIntegration tests environment variable integration
func TestIntegration_LoggingUtility_ConfigurationIntegration(t *testing.T) {
	testCases := []struct {
		name     string
		level    string
		file     string
		expected LogLevel
	}{
		{"Default", "", "", Info},
		{"Debug", "DEBUG", "", Debug},
		{"WithFile", "WARNING", "/tmp/integration_test.log", Warning},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment
			if tc.level != "" {
				os.Setenv("LOG_LEVEL", tc.level)
				defer os.Unsetenv("LOG_LEVEL")
			}
			if tc.file != "" {
				os.Setenv("LOG_FILE", tc.file)
				defer os.Unsetenv("LOG_FILE")
				defer os.Remove(tc.file)
			}

			// Create utility
			logger := NewLoggingUtility()
			defer func() {
				if concreteLogger, ok := logger.(*LoggingUtility); ok {
					concreteLogger.Close()
				}
			}()

			// Test configuration is applied
			if !logger.IsLevelEnabled(tc.expected) {
				t.Errorf("Expected level %v to be enabled", tc.expected)
			}

			// Test logging works
			logger.Log(tc.expected, "ConfigTest", "Configuration integration test", nil)

			// If file logging configured, verify file exists
			if tc.file != "" {
				if _, err := os.Stat(tc.file); os.IsNotExist(err) {
					t.Error("Expected log file to be created")
				}
			}
		})
	}
}

// TestLoggingUtility_Integration_ErrorScenarios tests error handling in integration scenarios
func TestIntegration_LoggingUtility_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		scenario    func() (recovered interface{})
		expectPanic bool
	}{
		{
			name: "InvalidLogFilePath",
			scenario: func() (recovered interface{}) {
				defer func() {
					recovered = recover()
				}()

				os.Setenv("LOG_FILE", "/invalid/path/that/does/not/exist/test.log")
				defer os.Unsetenv("LOG_FILE")

				NewLoggingUtility()
				return nil
			},
			expectPanic: true,
		},
		{
			name: "ReadOnlyLogFile",
			scenario: func() (recovered interface{}) {
				defer func() {
					recovered = recover()
				}()

				// Create a read-only directory
				tempDir := "/tmp/readonly_test"
				os.Mkdir(tempDir, 0444)
				defer os.RemoveAll(tempDir)

				logFile := tempDir + "/test.log"
				os.Setenv("LOG_FILE", logFile)
				defer os.Unsetenv("LOG_FILE")

				NewLoggingUtility()
				return nil
			},
			expectPanic: true,
		},
		{
			name: "NormalOperation",
			scenario: func() (recovered interface{}) {
				defer func() {
					recovered = recover()
				}()

				logger := NewLoggingUtility()
				defer func() {
					if concreteLogger, ok := logger.(*LoggingUtility); ok {
						concreteLogger.Close()
					}
				}()

				// Normal operations should not panic
				logger.Log(Info, "ErrorTest", "Normal operation", nil)
				return nil
			},
			expectPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recovered := tc.scenario()

			if tc.expectPanic && recovered == nil {
				t.Error("Expected panic but got none")
			}
			if !tc.expectPanic && recovered != nil {
				t.Errorf("Unexpected panic: %v", recovered)
			}
		})
	}
}

// TestLoggingUtility_Integration_ConcurrentUsage tests thread safety in realistic scenarios
func TestIntegration_LoggingUtility_ConcurrentUsage(t *testing.T) {
	logger := NewLoggingUtility()
	defer func() {
		if concreteLogger, ok := logger.(*LoggingUtility); ok {
			concreteLogger.Close()
		}
	}()

	// Simulate realistic concurrent usage from different system components
	components := []string{
		"TaskManager",
		"ValidationEngine",
		"TasksAccess",
		"EisenKanClient",
		"RuleEngine",
	}

	done := make(chan bool, len(components))

	// Start concurrent logging from different components
	for _, component := range components {
		go func(comp string) {
			for i := range 100 {
				// Mix different logging approaches
				switch i % 3 {
				case 0:
					logger.Log(Info, comp, fmt.Sprintf("Simple log %d", i), nil)
				case 1:
					data := map[string]interface{}{
						"iteration": i,
						"timestamp": time.Now().Unix(),
					}
					logger.Log(Debug, comp, "Structured log", data)
				case 2:
					testError := fmt.Errorf("test error %d", i)
					errorData := map[string]interface{}{
						"errorCode": i,
					}
					logger.LogError(comp, testError, errorData)
				}
			}
			done <- true
		}(component)
	}

	// Wait for all goroutines to complete
	for range components {
		<-done
	}

	// Test should complete without deadlocks or data races
}

// TestLoggingUtility_Integration_UseCaseValidation validates all SRS use cases work in practice
func TestIntegration_LoggingUtility_UseCaseValidation(t *testing.T) {
	// Set up comprehensive logging environment
	tmpFile := "/tmp/usecase_validation.log"
	defer os.Remove(tmpFile)

	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_FILE", tmpFile)
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FILE")
	}()

	logger := NewLoggingUtility()
	defer func() {
		if concreteLogger, ok := logger.(*LoggingUtility); ok {
			concreteLogger.Close()
		}
	}()

	// Use Case 1: Log Application Events with appropriate levels
	levels := []LogLevel{Debug, Info, Warning, Error, Fatal}
	for _, level := range levels {
		logger.Log(level, "UseCaseTest", fmt.Sprintf("Testing level %s", level), nil)
	}

	// Use Case 2: Log Structured Data
	structuredData := map[string]interface{}{
		"testType":   "integration",
		"phase":      "usecase_validation",
		"timestamp":  time.Now().Unix(),
		"components": len(levels),
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": "nested_test",
			},
		},
	}
	logger.Log(Info, "IntegrationTest", "Structured data test", structuredData)

	// Use Case 3: Multiple Output Destinations (console + file)
	logger.Log(Info, "OutputTest", "Testing dual output destinations", nil)

	// Use Case 4: Filter by Log Level (performance optimization)
	if logger.IsLevelEnabled(Debug) {
		expensiveData := make([]string, 1000)
		for i := range expensiveData {
			expensiveData[i] = fmt.Sprintf("data-%d", i)
		}
		debugData := map[string]interface{}{
			"dataSize": len(expensiveData),
			"sample":   expensiveData[:5], // Just first 5 for readability
		}
		logger.Log(Debug, "PerformanceTest", "Expensive debug operation", debugData)
	}

	// Use Case 5: Error Logging with Stack Trace
	testError := fmt.Errorf("integration test error")
	errorData := map[string]interface{}{
		"context": "use_case_validation",
		"step":    5,
	}
	logger.LogError("IntegrationTest", testError, errorData)

	// Validate file output exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Expected log file to exist after logging operations")
	}

	// Validate file contains expected content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	expectedStrings := []string{
		"DEBUG", "INFO", "WARNING", "ERROR", "FATAL",
		"UseCaseTest",
		"Structured data test",
		"testType",
		"integration",
		"OutputTest",
		"PerformanceTest",
		"integration test error",
	}

	for _, expected := range expectedStrings {
		if !contains(logContent, expected) {
			t.Errorf("Log file missing expected content: %s", expected)
		}
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// findSubstring finds a substring in a string
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
