// Package utilities provides example usage of LoggingUtility
// This demonstrates how other components should use the logging utility
package utilities

import (
	"fmt"
)

// ExampleLoggingUsage demonstrates how to use LoggingUtility in practice
// This shows proper integration patterns for other system components
func ExampleLoggingUsage() {
	// Create logging utility (typically done once at application startup)
	logger := NewLoggingUtility()

	// Ensure cleanup (typically done at application shutdown)
	defer func() {
		if concreteLogger, ok := logger.(*LoggingUtility); ok {
			concreteLogger.Close()
		}
	}()

	// Example 1: Simple logging (good for basic operations)
	logger.Log(Info, "ExampleComponent", "Application started successfully", nil)

	// Example 2: Structured logging with data (recommended for business operations)
	taskData := map[string]interface{}{
		"taskId":     "task-789",
		"priority":   "High",
		"durationMs": 25,
		"metadata": map[string]interface{}{
			"source": "user_input",
			"tags":   []interface{}{"urgent", "feature"},
		},
	}
	logger.Log(Info, "TaskManager", "Task created successfully", taskData)

	// Example 3: Error logging with context data
	simulatedError := fmt.Errorf("database connection failed")
	errorData := map[string]interface{}{
		"database": "tasks_db",
		"retries":  3,
		"timeout":  "5s",
	}
	logger.LogError("TasksAccess", simulatedError, errorData)

	// Example 4: Performance optimization using level checking
	if logger.IsLevelEnabled(Debug) {
		expensiveDebugData := generateExpensiveDebugData()
		debugData := map[string]interface{}{
			"dataPoints": len(expensiveDebugData),
			"memoryMB":   42,
			"metrics":    expensiveDebugData,
		}
		logger.Log(Debug, "PerformanceMonitor", "Performance analysis completed", debugData)
	}

	// Example 5: Different map types
	intMapData := map[int]interface{}{
		1: "first priority",
		2: "second priority", 
		3: "third priority",
	}
	logger.Log(Info, "PriorityManager", "Priority mapping updated", intMapData)

	// Example 6: Generic map with mixed key types
	mixedMapData := map[interface{}]interface{}{
		42:        "answer to everything",
		"status": "active",
		true:     "enabled flag",
	}
	logger.Log(Debug, "ConfigManager", "Mixed configuration loaded", mixedMapData)
}

// generateExpensiveDebugData simulates expensive debug data generation
// This would only run if Debug level is enabled
func generateExpensiveDebugData() []string {
	return []string{"cpu_usage", "memory_stats", "network_io", "disk_stats"}
}