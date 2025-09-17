package resource_access

import (
	"fmt"
	"strings"
	"time"

	"github.com/rknuus/eisenkan/internal/managers/task_manager"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
)

// convertUIRequestToTaskRequest converts UI request to TaskManager request format
func (t *taskManagerAccess) convertUIRequestToTaskRequest(uiRequest UITaskRequest) (task_manager.TaskRequest, error) {
	// Convert UI priority to board_access priority
	priority := board_access.Priority{
		Urgent:    uiRequest.Priority.Urgent,
		Important: uiRequest.Priority.Important,
	}

	// Convert UI workflow status to TaskManager workflow status
	var workflowStatus task_manager.WorkflowStatus
	switch uiRequest.WorkflowStatus {
	case UITodo:
		workflowStatus = task_manager.Todo
	case UIInProgress:
		workflowStatus = task_manager.InProgress
	case UIDone:
		workflowStatus = task_manager.Done
	default:
		workflowStatus = task_manager.Todo // Default fallback
	}

	return task_manager.TaskRequest{
		Description:           uiRequest.Description,
		Priority:              priority,
		WorkflowStatus:        workflowStatus,
		Tags:                  uiRequest.Tags,
		Deadline:              uiRequest.Deadline,
		PriorityPromotionDate: uiRequest.PriorityPromotionDate,
		ParentTaskID:          uiRequest.ParentTaskID,
	}, nil
}

// convertTaskResponseToUI converts TaskManager response to UI format
func (t *taskManagerAccess) convertTaskResponseToUI(response task_manager.TaskResponse) UITaskResponse {
	// Convert priority
	uiPriority := UIPriority{
		Urgent:    response.Priority.Urgent,
		Important: response.Priority.Important,
		Label:     response.Priority.Label,
		SortOrder: t.calculatePrioritySortOrder(response.Priority),
	}

	// Convert workflow status
	var uiStatus UIWorkflowStatus
	switch response.WorkflowStatus {
	case task_manager.Todo:
		uiStatus = UITodo
	case task_manager.InProgress:
		uiStatus = UIInProgress
	case task_manager.Done:
		uiStatus = UIDone
	default:
		uiStatus = UITodo
	}

	// Generate UI-optimized display fields
	displayName := t.generateDisplayName(response)
	statusText := t.generateStatusText(uiStatus)
	priorityText := t.generatePriorityText(uiPriority)
	deadlineText := t.generateDeadlineText(response.Deadline)
	hasSubtasks := len(response.SubtaskIDs) > 0
	isOverdue := t.isTaskOverdue(response.Deadline)

	return UITaskResponse{
		ID:                    response.ID,
		Description:           response.Description,
		Priority:              uiPriority,
		WorkflowStatus:        uiStatus,
		Tags:                  response.Tags,
		Deadline:              response.Deadline,
		PriorityPromotionDate: response.PriorityPromotionDate,
		ParentTaskID:          response.ParentTaskID,
		SubtaskIDs:            response.SubtaskIDs,
		CreatedAt:             response.CreatedAt,
		UpdatedAt:             response.UpdatedAt,
		DisplayName:           displayName,
		StatusText:            statusText,
		PriorityText:          priorityText,
		DeadlineText:          deadlineText,
		HasSubtasks:           hasSubtasks,
		IsOverdue:             isOverdue,
	}
}

// convertUIQueryCriteriaToTaskCriteria converts UI criteria to TaskManager format
func (t *taskManagerAccess) convertUIQueryCriteriaToTaskCriteria(uiCriteria UIQueryCriteria) task_manager.QueryCriteria {
	criteria := task_manager.QueryCriteria{
		Columns:      uiCriteria.Columns,
		Sections:     uiCriteria.Sections,
		Tags:         uiCriteria.Tags,
		ParentTaskID: uiCriteria.ParentTaskID,
	}

	// Convert priority if specified
	if uiCriteria.Priority != nil {
		criteria.Priority = &board_access.Priority{
			Urgent:    uiCriteria.Priority.Urgent,
			Important: uiCriteria.Priority.Important,
		}
	}

	// Convert date range if specified
	if uiCriteria.DateRange != nil {
		criteria.DateRange = &board_access.DateRange{
			From: uiCriteria.DateRange.Start,
			To:   uiCriteria.DateRange.End,
		}
	}

	// Convert priority promotion date range if specified
	if uiCriteria.PriorityPromotionDate != nil {
		criteria.PriorityPromotionDate = &board_access.DateRange{
			From: uiCriteria.PriorityPromotionDate.Start,
			To:   uiCriteria.PriorityPromotionDate.End,
		}
	}

	// Convert hierarchy filter
	switch uiCriteria.Hierarchy {
	case UIHierarchyTopLevel:
		criteria.Hierarchy = board_access.TopLevelOnly
	case UIHierarchySubtasksOnly:
		criteria.Hierarchy = board_access.SubtasksOnly
	default:
		criteria.Hierarchy = board_access.AllTasks
	}

	return criteria
}

// convertUIWorkflowStatusToTaskStatus converts UI status to TaskManager format
func (t *taskManagerAccess) convertUIWorkflowStatusToTaskStatus(uiStatus UIWorkflowStatus) task_manager.WorkflowStatus {
	switch uiStatus {
	case UITodo:
		return task_manager.Todo
	case UIInProgress:
		return task_manager.InProgress
	case UIDone:
		return task_manager.Done
	default:
		return task_manager.Todo
	}
}

// convertValidationResultToUI converts TaskManager validation to UI format
func (t *taskManagerAccess) convertValidationResultToUI(validation task_manager.ValidationResult) UIValidationResult {
	// Convert rule violations to field errors
	fieldErrors := make(map[string]string)
	var suggestions []string
	var generalError string

	for _, violation := range validation.Violations {
		// Map rule violations to field-specific errors
		// This would need to be enhanced based on actual rule violation types
		if strings.Contains(violation.Message, "description") {
			fieldErrors["description"] = violation.Message
		} else if strings.Contains(violation.Message, "priority") {
			fieldErrors["priority"] = violation.Message
		} else if strings.Contains(violation.Message, "deadline") {
			fieldErrors["deadline"] = violation.Message
		} else {
			if generalError == "" {
				generalError = violation.Message
			} else {
				generalError += "; " + violation.Message
			}
		}

		// Add any suggestions from rule violations
		suggestions = append(suggestions, "Check business rules configuration")
	}

	return UIValidationResult{
		Valid:        validation.Valid,
		FieldErrors:  fieldErrors,
		GeneralError: generalError,
		Suggestions:  suggestions,
	}
}

// calculatePrioritySortOrder determines sort order for UI priority display
func (t *taskManagerAccess) calculatePrioritySortOrder(priority board_access.Priority) int {
	if priority.Urgent && priority.Important {
		return 1 // Highest priority
	} else if priority.Urgent && !priority.Important {
		return 2
	} else if !priority.Urgent && priority.Important {
		return 3
	} else {
		return 4 // Lowest priority
	}
}

// generateDisplayName creates UI-optimized display text for tasks
func (t *taskManagerAccess) generateDisplayName(response task_manager.TaskResponse) string {
	displayName := response.Description
	
	// Truncate long descriptions for display
	if len(displayName) > 100 {
		displayName = displayName[:97] + "..."
	}
	
	// Add indicators for special states
	if len(response.SubtaskIDs) > 0 {
		displayName += fmt.Sprintf(" (%d subtasks)", len(response.SubtaskIDs))
	}
	
	return displayName
}

// generateStatusText creates human-readable status text
func (t *taskManagerAccess) generateStatusText(status UIWorkflowStatus) string {
	switch status {
	case UITodo:
		return "To Do"
	case UIInProgress:
		return "In Progress"
	case UIDone:
		return "Done"
	default:
		return "Unknown"
	}
}

// generatePriorityText creates human-readable priority text
func (t *taskManagerAccess) generatePriorityText(priority UIPriority) string {
	if priority.Urgent && priority.Important {
		return "Urgent & Important"
	} else if priority.Urgent && !priority.Important {
		return "Urgent"
	} else if !priority.Urgent && priority.Important {
		return "Important"
	} else {
		return "Normal"
	}
}

// generateDeadlineText creates formatted deadline text for UI display
func (t *taskManagerAccess) generateDeadlineText(deadline *time.Time) string {
	if deadline == nil {
		return ""
	}
	
	now := time.Now()
	diff := deadline.Sub(now)
	
	if diff < 0 {
		return fmt.Sprintf("Overdue by %s", formatDuration(-diff))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("Due in %s", formatDuration(diff))
	} else {
		return deadline.Format("Jan 2, 2006")
	}
}

// formatDuration creates human-readable duration text
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	} else {
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	}
}

// isTaskOverdue checks if a task is past its deadline
func (t *taskManagerAccess) isTaskOverdue(deadline *time.Time) bool {
	if deadline == nil {
		return false
	}
	return time.Now().After(*deadline)
}

// calculateBoardSummary generates board statistics from task list
func (t *taskManagerAccess) calculateBoardSummary(tasks []task_manager.TaskResponse) UIBoardSummary {
	summary := UIBoardSummary{
		TotalTasks:      len(tasks),
		TasksByStatus:   make(map[UIWorkflowStatus]int),
		TasksByPriority: make(map[string]int),
		LastUpdated:     time.Now(),
	}
	
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekEnd := today.AddDate(0, 0, 7)
	
	var parentTasks, totalSubtasks, completedSubtasks int
	
	for _, task := range tasks {
		// Count by status
		uiStatus := t.convertWorkflowStatusToUI(task.WorkflowStatus)
		summary.TasksByStatus[uiStatus]++
		
		// Count by priority
		summary.TasksByPriority[task.Priority.Label]++
		
		// Check deadline status
		if task.Deadline != nil {
			if task.Deadline.Before(now) {
				summary.OverdueTasks++
			} else if task.Deadline.Before(today.AddDate(0, 0, 1)) {
				summary.TasksDueToday++
			} else if task.Deadline.Before(weekEnd) {
				summary.TasksDueThisWeek++
			}
		}
		
		// Count subtask information
		if len(task.SubtaskIDs) > 0 {
			parentTasks++
		}
		if task.ParentTaskID != nil {
			totalSubtasks++
			if task.WorkflowStatus == task_manager.Done {
				completedSubtasks++
			}
		}
	}
	
	summary.SubtaskCounts = UISubtaskSummary{
		ParentTasks:       parentTasks,
		TotalSubtasks:     totalSubtasks,
		CompletedSubtasks: completedSubtasks,
	}
	
	return summary
}

// convertWorkflowStatusToUI converts TaskManager status to UI format
func (t *taskManagerAccess) convertWorkflowStatusToUI(status task_manager.WorkflowStatus) UIWorkflowStatus {
	switch status {
	case task_manager.Todo:
		return UITodo
	case task_manager.InProgress:
		return UIInProgress
	case task_manager.Done:
		return UIDone
	default:
		return UITodo
	}
}