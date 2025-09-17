package resource_access

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/rknuus/eisenkan/internal/managers/task_manager"
)

// translateServiceError converts TaskManager service errors to UI-friendly errors
func (t *taskManagerAccess) translateServiceError(operation string, err error) error {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()
	
	// Categorize error based on error message content
	category := t.categorizeError(errorMsg)
	
	// Generate user-friendly message
	message := t.generateUserFriendlyMessage(operation, errorMsg, category)
	
	// Generate recovery suggestions
	suggestions := t.generateRecoverySuggestions(operation, errorMsg, category)
	
	// Determine if error is retryable
	retryable := t.isErrorRetryable(errorMsg, category)
	
	// Log the error for debugging
	t.logger.LogError("TaskManagerAccess", err, map[string]interface{}{
		"operation": operation,
		"category":  category,
	})
	
	return UIErrorResponse{
		Category:    category,
		Message:     message,
		Details:     errorMsg,
		Suggestions: suggestions,
		Retryable:   retryable,
	}
}

// categorizeError determines the error category based on error content
func (t *taskManagerAccess) categorizeError(errorMsg string) string {
	errorLower := strings.ToLower(errorMsg)
	
	// Check for validation errors
	if strings.Contains(errorLower, "validation") || 
	   strings.Contains(errorLower, "invalid") || 
	   strings.Contains(errorLower, "required") ||
	   strings.Contains(errorLower, "missing") {
		return "validation"
	}
	
	// Check for connectivity errors
	if strings.Contains(errorLower, "connection") ||
	   strings.Contains(errorLower, "network") ||
	   strings.Contains(errorLower, "timeout") ||
	   strings.Contains(errorLower, "unavailable") ||
	   strings.Contains(errorLower, "refused") {
		return "connectivity"
	}
	
	// Default to service error
	return "service"
}

// generateUserFriendlyMessage creates user-friendly error messages
func (t *taskManagerAccess) generateUserFriendlyMessage(operation, errorMsg, category string) string {
	switch category {
	case "validation":
		return t.generateValidationMessage(operation, errorMsg)
	case "connectivity":
		return t.generateConnectivityMessage(operation)
	case "service":
		return t.generateServiceMessage(operation, errorMsg)
	default:
		return fmt.Sprintf("An error occurred during %s", operation)
	}
}

// generateValidationMessage creates validation-specific error messages
func (t *taskManagerAccess) generateValidationMessage(operation, errorMsg string) string {
	switch operation {
	case "CreateTask":
		return "Unable to create task due to validation errors"
	case "UpdateTask":
		return "Unable to update task due to validation errors"
	case "DeleteTask":
		return "Unable to delete task due to validation errors"
	case "ChangeTaskStatus":
		return "Unable to change task status due to validation errors"
	default:
		return "Task data validation failed"
	}
}

// generateConnectivityMessage creates connectivity-specific error messages
func (t *taskManagerAccess) generateConnectivityMessage(operation string) string {
	switch operation {
	case "CreateTask", "UpdateTask", "DeleteTask":
		return "Unable to save changes - please check your connection and try again"
	case "GetTask", "ListTasks", "SearchTasks":
		return "Unable to load tasks - please check your connection and try again"
	default:
		return "Connection error - please check your connection and try again"
	}
}

// generateServiceMessage creates service-specific error messages
func (t *taskManagerAccess) generateServiceMessage(operation, errorMsg string) string {
	if strings.Contains(strings.ToLower(errorMsg), "not found") {
		return "The requested task was not found"
	}
	
	if strings.Contains(strings.ToLower(errorMsg), "duplicate") {
		return "A task with similar information already exists"
	}
	
	switch operation {
	case "CreateTask":
		return "Unable to create task due to an internal error"
	case "UpdateTask":
		return "Unable to update task due to an internal error"
	case "DeleteTask":
		return "Unable to delete task due to an internal error"
	case "GetTask":
		return "Unable to retrieve task information"
	case "ListTasks":
		return "Unable to load task list"
	case "SearchTasks":
		return "Unable to search tasks"
	default:
		return "An internal error occurred"
	}
}

// generateRecoverySuggestions creates context-appropriate recovery suggestions
func (t *taskManagerAccess) generateRecoverySuggestions(operation, errorMsg, category string) []string {
	switch category {
	case "validation":
		return t.generateValidationSuggestions(operation, errorMsg)
	case "connectivity":
		return []string{
			"Check your network connection",
			"Try again in a few moments",
			"Contact support if the problem persists",
		}
	case "service":
		return t.generateServiceSuggestions(operation, errorMsg)
	default:
		return []string{"Try again later"}
	}
}

// generateValidationSuggestions creates validation-specific recovery suggestions
func (t *taskManagerAccess) generateValidationSuggestions(operation, errorMsg string) []string {
	suggestions := []string{}
	
	errorLower := strings.ToLower(errorMsg)
	
	if strings.Contains(errorLower, "description") {
		suggestions = append(suggestions, "Ensure the task description is not empty")
	}
	
	if strings.Contains(errorLower, "priority") {
		suggestions = append(suggestions, "Select a valid priority level")
	}
	
	if strings.Contains(errorLower, "deadline") {
		suggestions = append(suggestions, "Check that the deadline is in the future")
	}
	
	if strings.Contains(errorLower, "parent") {
		suggestions = append(suggestions, "Verify the parent task exists and is valid")
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Check all required fields are filled correctly")
	}
	
	return suggestions
}

// generateServiceSuggestions creates service-specific recovery suggestions
func (t *taskManagerAccess) generateServiceSuggestions(operation, errorMsg string) []string {
	if strings.Contains(strings.ToLower(errorMsg), "not found") {
		return []string{
			"Refresh the task list",
			"Verify the task still exists",
			"Try searching for the task",
		}
	}
	
	if strings.Contains(strings.ToLower(errorMsg), "duplicate") {
		return []string{
			"Check if a similar task already exists",
			"Use a different task description",
			"Update the existing task instead",
		}
	}
	
	return []string{
		"Try the operation again",
		"Refresh the application",
		"Contact support if the problem persists",
	}
}

// isErrorRetryable determines if an error condition can be retried
func (t *taskManagerAccess) isErrorRetryable(errorMsg, category string) bool {
	switch category {
	case "validation":
		return true // User can fix validation errors
	case "connectivity":
		return true // Connectivity issues are often transient
	case "service":
		// Service errors are sometimes retryable
		errorLower := strings.ToLower(errorMsg)
		if strings.Contains(errorLower, "not found") {
			return false // Not found errors are not retryable
		}
		if strings.Contains(errorLower, "duplicate") {
			return false // Duplicate errors require different action
		}
		return true // Other service errors might be transient
	default:
		return false
	}
}

// createUIError creates a structured UI error response
func (t *taskManagerAccess) createUIError(category, message, details string, suggestions []string, retryable bool) error {
	return UIErrorResponse{
		Category:    category,
		Message:     message,
		Details:     details,
		Suggestions: suggestions,
		Retryable:   retryable,
	}
}

// validateUITaskRequest performs UI-level validation on task requests
func (t *taskManagerAccess) validateUITaskRequest(request UITaskRequest) UIValidationResult {
	result := UIValidationResult{
		Valid:       true,
		FieldErrors: make(map[string]string),
		Suggestions: []string{},
	}
	
	// Validate description
	if strings.TrimSpace(request.Description) == "" {
		result.Valid = false
		result.FieldErrors["description"] = "Task description is required"
		result.Suggestions = append(result.Suggestions, "Enter a meaningful task description")
	}
	
	// Validate description length
	if len(request.Description) > 10000 {
		result.Valid = false
		result.FieldErrors["description"] = "Task description is too long (maximum 10,000 characters)"
		result.Suggestions = append(result.Suggestions, "Shorten the task description")
	}
	
	// Validate deadline
	if request.Deadline != nil && request.Deadline.Before(time.Now().Add(-24*time.Hour)) {
		result.Valid = false
		result.FieldErrors["deadline"] = "Deadline cannot be more than 24 hours in the past"
		result.Suggestions = append(result.Suggestions, "Set a deadline in the future")
	}
	
	// Validate priority promotion date
	if request.PriorityPromotionDate != nil {
		if request.Priority.Urgent {
			result.Valid = false
			result.FieldErrors["priority_promotion_date"] = "Priority promotion date is not needed for urgent tasks"
			result.Suggestions = append(result.Suggestions, "Remove priority promotion date for urgent tasks")
		}
	}
	
	// Validate tags
	for _, tag := range request.Tags {
		if strings.TrimSpace(tag) == "" {
			result.Valid = false
			result.FieldErrors["tags"] = "Empty tags are not allowed"
			result.Suggestions = append(result.Suggestions, "Remove empty tags")
			break
		}
		if len(tag) > 50 {
			result.Valid = false
			result.FieldErrors["tags"] = fmt.Sprintf("Tag '%s' is too long (maximum 50 characters)", tag)
			result.Suggestions = append(result.Suggestions, "Shorten tag names")
			break
		}
	}
	
	// Set general error if validation failed
	if !result.Valid && result.GeneralError == "" {
		result.GeneralError = "Please correct the validation errors and try again"
	}
	
	return result
}

// generateCacheKeyFromCriteria creates a cache key from query criteria
func (t *taskManagerAccess) generateCacheKeyFromCriteria(criteria UIQueryCriteria) string {
	// Create a simple hash of the criteria for cache key
	data := fmt.Sprintf("%+v", criteria)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// taskMatchesQuery checks if a task matches a search query
func (t *taskManagerAccess) taskMatchesQuery(task task_manager.TaskResponse, query string) bool {
	queryLower := strings.ToLower(query)
	
	// Search in description
	if strings.Contains(strings.ToLower(task.Description), queryLower) {
		return true
	}
	
	// Search in tags
	for _, tag := range task.Tags {
		if strings.Contains(strings.ToLower(tag), queryLower) {
			return true
		}
	}
	
	// Search in priority label
	if strings.Contains(strings.ToLower(task.Priority.Label), queryLower) {
		return true
	}
	
	return false
}