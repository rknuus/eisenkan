package access

import (
	"time"
)

// UITaskRequest represents task data optimized for UI input
type UITaskRequest struct {
	Description           string                `json:"description"`
	Priority              UIPriority           `json:"priority"`
	WorkflowStatus        UIWorkflowStatus     `json:"workflow_status"`
	Tags                  []string             `json:"tags,omitempty"`
	Deadline              *time.Time           `json:"deadline,omitempty"`
	PriorityPromotionDate *time.Time           `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string              `json:"parent_task_id,omitempty"`
}

// UITaskResponse represents task data optimized for UI display
type UITaskResponse struct {
	ID                    string               `json:"id"`
	Description           string               `json:"description"`
	Priority              UIPriority           `json:"priority"`
	WorkflowStatus        UIWorkflowStatus     `json:"workflow_status"`
	Tags                  []string             `json:"tags,omitempty"`
	Deadline              *time.Time           `json:"deadline,omitempty"`
	PriorityPromotionDate *time.Time           `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string              `json:"parent_task_id,omitempty"`
	SubtaskIDs            []string             `json:"subtask_ids,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
	
	// UI-optimized display fields
	DisplayName           string               `json:"display_name"`           // UI-optimized display text
	StatusText            string               `json:"status_text"`            // Human-readable status
	PriorityText          string               `json:"priority_text"`          // Human-readable priority
	DeadlineText          string               `json:"deadline_text"`          // Formatted deadline string
	HasSubtasks           bool                 `json:"has_subtasks"`           // Quick subtask check
	IsOverdue             bool                 `json:"is_overdue"`             // Deadline status
}

// UIPriority represents priority settings optimized for UI interaction
type UIPriority struct {
	Urgent     bool   `json:"urgent"`
	Important  bool   `json:"important"`
	Label      string `json:"label"`      // "urgent-important", "not-urgent-important", etc.
	SortOrder  int    `json:"sort_order"` // For UI sorting (1=highest priority)
}

// UIWorkflowStatus represents workflow states optimized for UI
type UIWorkflowStatus string

const (
	UITodo       UIWorkflowStatus = "todo"
	UIInProgress UIWorkflowStatus = "doing"
	UIDone       UIWorkflowStatus = "done"
)

// UIQueryCriteria represents query parameters optimized for UI filtering
type UIQueryCriteria struct {
	Columns               []string                `json:"columns,omitempty"`
	Sections              []string                `json:"sections,omitempty"`
	Priority              *UIPriority             `json:"priority,omitempty"`
	Tags                  []string                `json:"tags,omitempty"`
	DateRange             *UIDateRange            `json:"date_range,omitempty"`
	PriorityPromotionDate *UIDateRange            `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string                 `json:"parent_task_id,omitempty"`
	Hierarchy             UIHierarchyFilter       `json:"hierarchy,omitempty"`
	WorkflowStatus        []UIWorkflowStatus      `json:"workflow_status,omitempty"`
	SearchText            string                  `json:"search_text,omitempty"`
}

// UIDateRange represents date filtering for UI
type UIDateRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// UIHierarchyFilter represents hierarchy filtering options
type UIHierarchyFilter string

const (
	UIHierarchyAll         UIHierarchyFilter = "all"
	UIHierarchyParentsOnly UIHierarchyFilter = "parents_only"
	UIHierarchySubtasksOnly UIHierarchyFilter = "subtasks_only"
	UIHierarchyTopLevel    UIHierarchyFilter = "top_level"
)

// UIValidationResult represents validation results optimized for UI display
type UIValidationResult struct {
	Valid        bool                    `json:"valid"`
	FieldErrors  map[string]string       `json:"field_errors,omitempty"`  // Field -> Error message
	GeneralError string                  `json:"general_error,omitempty"` // General validation error
	Suggestions  []string                `json:"suggestions,omitempty"`   // Recovery suggestions
}

// UIErrorResponse represents errors optimized for UI display
type UIErrorResponse struct {
	Category    string   `json:"category"`    // "validation", "service", "connectivity"
	Message     string   `json:"message"`     // User-friendly error message
	Details     string   `json:"details"`     // Technical details for debugging
	Suggestions []string `json:"suggestions"` // Recovery actions for user
	Retryable   bool     `json:"retryable"`   // Whether operation can be retried
}

// UIBoardSummary represents board statistics optimized for UI display
type UIBoardSummary struct {
	TotalTasks        int                        `json:"total_tasks"`
	TasksByStatus     map[UIWorkflowStatus]int   `json:"tasks_by_status"`
	TasksByPriority   map[string]int             `json:"tasks_by_priority"`
	OverdueTasks      int                        `json:"overdue_tasks"`
	TasksDueToday     int                        `json:"tasks_due_today"`
	TasksDueThisWeek  int                        `json:"tasks_due_this_week"`
	SubtaskCounts     UISubtaskSummary           `json:"subtask_counts"`
	LastUpdated       time.Time                  `json:"last_updated"`
}

// UISubtaskSummary represents subtask statistics
type UISubtaskSummary struct {
	ParentTasks       int `json:"parent_tasks"`       // Tasks with subtasks
	TotalSubtasks     int `json:"total_subtasks"`     // Total number of subtasks
	CompletedSubtasks int `json:"completed_subtasks"` // Completed subtasks
}

// Error implements the error interface for UIErrorResponse
func (e UIErrorResponse) Error() string {
	return e.Message
}