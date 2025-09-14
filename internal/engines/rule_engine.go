// Package engines provides Engine layer components implementing the iDesign methodology.
// This package contains components that encapsulate business logic and provide
// pure processing services to higher-level components in the application architecture.
package engines

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rknuus/eisenkan/internal/resource_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// TaskEvent represents a task state change event for rule evaluation
type TaskEvent struct {
	EventType        string                              `json:"event_type"` // "task_transition", "task_update", "task_create"
	CurrentState     *resource_access.TaskWithTimestamps `json:"current_state,omitempty"`
	FutureState      *TaskState                          `json:"future_state"`
	Timestamp        time.Time                           `json:"timestamp"`
	ParentTask       *resource_access.TaskWithTimestamps `json:"parent_task,omitempty"`
	AffectedSubtasks []*resource_access.TaskWithTimestamps `json:"affected_subtasks,omitempty"`
}

// TaskState represents the intended state of a task
type TaskState struct {
	Task     *resource_access.Task          `json:"task"`
	Priority resource_access.Priority       `json:"priority"`
	Status   resource_access.WorkflowStatus `json:"status"`
}

// RuleViolation represents a single rule violation
type RuleViolation struct {
	RuleID   string `json:"rule_id"`
	Priority int    `json:"priority"`
	Message  string `json:"message"`
	Category string `json:"category"` // "validation", "workflow", "automation", "notification"
	Details  string `json:"details,omitempty"`
}

// RuleEvaluationResult contains the outcome of rule evaluation
type RuleEvaluationResult struct {
	Allowed    bool            `json:"allowed"`
	Violations []RuleViolation `json:"violations,omitempty"`
}

// EnrichedContext contains all context needed for rule evaluation
type EnrichedContext struct {
	Event            TaskEvent                                        `json:"event"`
	WIPCounts        map[string]int                                   `json:"wip_counts"`        // column -> task count
	SubtaskWIPCounts map[string]int                                   `json:"subtask_wip_counts"` // column -> subtask count
	TaskHistory      []utilities.CommitInfo                           `json:"task_history"`      // for age calculations
	Subtasks         []*resource_access.TaskWithTimestamps            `json:"subtasks"`          // for dependency rules
	ColumnTasks      map[string][]*resource_access.TaskWithTimestamps `json:"column_tasks"`      // for priority comparisons
	ColumnEnterTimes map[string]time.Time                             `json:"column_enter_times"` // column -> enter timestamp
	BoardMetadata    map[string]string                                `json:"board_metadata"`    // for custom rules
	HierarchyMap     map[string][]string                              `json:"hierarchy_map"`     // parent -> subtasks mapping
}

// IRuleEngine defines the interface for rule evaluation operations
type IRuleEngine interface {
	// EvaluateTaskChange evaluates whether a task change can be applied
	EvaluateTaskChange(ctx context.Context, event TaskEvent, boardPath string) (*RuleEvaluationResult, error)

	// Close releases any resources held by the engine
	Close() error
}

// RuleEngine implements IRuleEngine interface
type RuleEngine struct {
	rulesAccess resource_access.IRulesAccess
	boardAccess resource_access.IBoardAccess
	logger      utilities.ILoggingUtility
}

// NewRuleEngine creates a new RuleEngine instance
func NewRuleEngine(rulesAccess resource_access.IRulesAccess, boardAccess resource_access.IBoardAccess) (*RuleEngine, error) {
	if rulesAccess == nil {
		return nil, fmt.Errorf("RuleEngine.NewRuleEngine: rulesAccess cannot be nil")
	}
	if boardAccess == nil {
		return nil, fmt.Errorf("RuleEngine.NewRuleEngine: boardAccess cannot be nil")
	}

	logger := utilities.NewLoggingUtility()
	logger.LogMessage(utilities.Debug, "RuleEngine", "Initializing RuleEngine")

	engine := &RuleEngine{
		rulesAccess: rulesAccess,
		boardAccess: boardAccess,
		logger:      logger,
	}

	logger.LogMessage(utilities.Info, "RuleEngine", "RuleEngine initialized successfully")
	return engine, nil
}

// EvaluateTaskChange evaluates whether a task change can be applied
func (re *RuleEngine) EvaluateTaskChange(ctx context.Context, event TaskEvent, boardPath string) (*RuleEvaluationResult, error) {
	re.logger.LogMessage(utilities.Debug, "RuleEngine", fmt.Sprintf("Evaluating task change for event type: %s", event.EventType))

	// Get applicable rules for the board
	ruleSet, err := re.rulesAccess.ReadRules(boardPath)
	if err != nil {
		return nil, fmt.Errorf("RuleEngine.EvaluateTaskChange failed to read rules: %w", err)
	}

	// Filter rules based on event type and enabled status
	applicableRules := re.filterApplicableRules(ruleSet.Rules, event.EventType)
	if len(applicableRules) == 0 {
		re.logger.LogMessage(utilities.Debug, "RuleEngine", "No applicable rules found, allowing task change")
		return &RuleEvaluationResult{Allowed: true}, nil
	}

	// Enrich context with board data
	enrichedContext, err := re.enrichContext(ctx, event, boardPath)
	if err != nil {
		return nil, fmt.Errorf("RuleEngine.EvaluateTaskChange failed to enrich context: %w", err)
	}

	// Evaluate all applicable rules using complete sequential processor
	violations := re.evaluateRules(applicableRules, enrichedContext)

	// Sort violations by priority (higher priority first)
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Priority > violations[j].Priority
	})

	result := &RuleEvaluationResult{
		Allowed:    len(violations) == 0,
		Violations: violations,
	}

	re.logger.LogMessage(utilities.Info, "RuleEngine",
		fmt.Sprintf("Rule evaluation completed: allowed=%t, violations=%d", result.Allowed, len(violations)))

	return result, nil
}

// filterApplicableRules filters rules based on event type and enabled status
func (re *RuleEngine) filterApplicableRules(rules []resource_access.Rule, eventType string) []resource_access.Rule {
	var applicable []resource_access.Rule

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule triggers match the event type
		if rule.TriggerType == eventType || rule.TriggerType == "all" {
			applicable = append(applicable, rule)
		}
	}

	re.logger.LogMessage(utilities.Debug, "RuleEngine",
		fmt.Sprintf("Filtered %d applicable rules from %d total rules", len(applicable), len(rules)))

	return applicable
}

// enrichContext gathers board context needed for rule evaluation
func (re *RuleEngine) enrichContext(ctx context.Context, event TaskEvent, boardPath string) (*EnrichedContext, error) {
	// Determine what data we need for rules evaluation
	var taskID string
	var targetColumns []string
	
	if event.CurrentState != nil && event.CurrentState.Task.ID != "" {
		taskID = event.CurrentState.Task.ID
	}
	
	if event.FutureState != nil {
		targetColumns = append(targetColumns, event.FutureState.Status.Column)
	}
	if event.CurrentState != nil {
		// Also get current column for age calculations
		targetColumns = append(targetColumns, event.CurrentState.Status.Column)
	}

	// Get all rule-related data in a single call
	rulesData, err := re.boardAccess.GetRulesData(taskID, targetColumns)
	if err != nil {
		return nil, fmt.Errorf("RuleEngine.enrichContext failed to get rules data: %w", err)
	}

	// Get subtasks for dependency rules using BoardAccess
	var subtasks []*resource_access.TaskWithTimestamps
	if taskID != "" {
		var err error
		subtasks, err = re.boardAccess.GetSubtasks(taskID)
		if err != nil {
			re.logger.LogMessage(utilities.Warning, "RuleEngine", fmt.Sprintf("Failed to get subtasks: %v", err))
			subtasks = []*resource_access.TaskWithTimestamps{} // Continue with empty list
		}
	}

	enriched := &EnrichedContext{
		Event:            event,
		WIPCounts:        rulesData.WIPCounts,
		SubtaskWIPCounts: rulesData.SubtaskWIPCounts,
		TaskHistory:      rulesData.TaskHistory,
		Subtasks:         subtasks,
		ColumnTasks:      rulesData.ColumnTasks,
		ColumnEnterTimes: rulesData.ColumnEnterTimes,
		BoardMetadata:    rulesData.BoardMetadata,
		HierarchyMap:     rulesData.HierarchyMap,
	}

	return enriched, nil
}


// getSubtasks retrieves subtasks for dependency rules (placeholder implementation)
func (re *RuleEngine) getSubtasks(ctx context.Context, event TaskEvent) ([]*resource_access.TaskWithTimestamps, error) {
	// TODO: Implement subtask retrieval when subtask support is added to BoardAccess
	// For now, return empty list as subtasks are not yet implemented in BoardAccess
	return []*resource_access.TaskWithTimestamps{}, nil
}


// evaluateRules evaluates all rules sequentially and aggregates violations
func (re *RuleEngine) evaluateRules(rules []resource_access.Rule, context *EnrichedContext) []RuleViolation {
	var violations []RuleViolation

	for _, rule := range rules {
		violation := re.evaluateRule(rule, context)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// evaluateRule evaluates a single rule against the enriched context
func (re *RuleEngine) evaluateRule(rule resource_access.Rule, context *EnrichedContext) *RuleViolation {
	re.logger.LogMessage(utilities.Debug, "RuleEngine", fmt.Sprintf("Evaluating rule: %s (%s)", rule.ID, rule.Name))

	// Evaluate rule based on category and conditions
	switch rule.Category {
	case "validation":
		return re.evaluateValidationRule(rule, context)
	case "workflow":
		return re.evaluateWorkflowRule(rule, context)
	case "automation":
		return re.evaluateAutomationRule(rule, context)
	case "notification":
		return re.evaluateNotificationRule(rule, context)
	default:
		re.logger.LogMessage(utilities.Warning, "RuleEngine", fmt.Sprintf("Unknown rule category: %s", rule.Category))
		return &RuleViolation{
			RuleID:   rule.ID,
			Priority: rule.Priority,
			Message:  fmt.Sprintf("Unknown rule category: %s", rule.Category),
			Category: rule.Category,
		}
	}
}

// evaluateValidationRule evaluates validation rules (e.g., required fields, WIP limits)
func (re *RuleEngine) evaluateValidationRule(rule resource_access.Rule, context *EnrichedContext) *RuleViolation {
	// WIP Limit Rule for top-level tasks
	if maxWIP, exists := rule.Conditions["max_wip_limit"]; exists {
		targetColumn := context.Event.FutureState.Status.Column

		// Convert maxWIP to integer
		maxWIPInt, err := re.parseIntValue(maxWIP)
		if err != nil {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  fmt.Sprintf("Invalid max_wip_limit value: %v", maxWIP),
				Category: rule.Category,
			}
		}

		// Check if this is a subtask
		isSubtask := context.Event.FutureState != nil && context.Event.FutureState.Task.ParentTaskID != nil
		var currentWIP int
		if isSubtask {
			// Use subtask WIP counts for subtasks
			currentWIP = context.SubtaskWIPCounts[targetColumn]
		} else {
			// Use regular WIP counts for top-level tasks
			currentWIP = context.WIPCounts[targetColumn]
		}

		// If moving TO this column (not already in it), check if it would exceed limit
		if context.Event.CurrentState == nil || context.Event.CurrentState.Status.Column != targetColumn {
			if currentWIP >= maxWIPInt {
				taskType := "tasks"
				if isSubtask {
					taskType = "subtasks"
				}
				return &RuleViolation{
					RuleID:   rule.ID,
					Priority: rule.Priority,
					Message:  fmt.Sprintf("WIP limit exceeded: column '%s' has %d %s, limit is %d", targetColumn, currentWIP, taskType, maxWIPInt),
					Category: rule.Category,
					Details:  fmt.Sprintf("Current WIP: %d, Limit: %d", currentWIP, maxWIPInt),
				}
			}
		}
	}

	// Subtask-specific WIP Limit Rule 
	if maxSubtaskWIP, exists := rule.Conditions["max_subtask_wip_limit"]; exists {
		targetColumn := context.Event.FutureState.Status.Column

		// Convert maxSubtaskWIP to integer
		maxSubtaskWIPInt, err := re.parseIntValue(maxSubtaskWIP)
		if err != nil {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  fmt.Sprintf("Invalid max_subtask_wip_limit value: %v", maxSubtaskWIP),
				Category: rule.Category,
			}
		}

		// Only apply to subtasks
		isSubtask := context.Event.FutureState != nil && context.Event.FutureState.Task.ParentTaskID != nil
		if isSubtask {
			currentSubtaskWIP := context.SubtaskWIPCounts[targetColumn]

			// If moving TO this column (not already in it), check if it would exceed limit
			if context.Event.CurrentState == nil || context.Event.CurrentState.Status.Column != targetColumn {
				if currentSubtaskWIP >= maxSubtaskWIPInt {
					return &RuleViolation{
						RuleID:   rule.ID,
						Priority: rule.Priority,
						Message:  fmt.Sprintf("Subtask WIP limit exceeded: column '%s' has %d subtasks, limit is %d", targetColumn, currentSubtaskWIP, maxSubtaskWIPInt),
						Category: rule.Category,
						Details:  fmt.Sprintf("Current Subtask WIP: %d, Limit: %d", currentSubtaskWIP, maxSubtaskWIPInt),
					}
				}
			}
		}
	}

	// Required Fields Rule
	if requiredFields, exists := rule.Conditions["required_fields"]; exists {
		if fields, ok := requiredFields.([]interface{}); ok {
			for _, field := range fields {
				fieldName := fmt.Sprintf("%v", field)
				if violation := re.checkRequiredField(fieldName, context.Event.FutureState.Task, rule); violation != nil {
					return violation
				}
			}
		}
	}

	return nil // No violation
}

// evaluateWorkflowRule evaluates workflow rules (e.g., column transitions)
func (re *RuleEngine) evaluateWorkflowRule(rule resource_access.Rule, context *EnrichedContext) *RuleViolation {
	// Column Transition Rule
	if allowedTransitions, exists := rule.Conditions["allowed_transitions"]; exists {
		if context.Event.CurrentState != nil {
			currentColumn := context.Event.CurrentState.Status.Column
			targetColumn := context.Event.FutureState.Status.Column

			if currentColumn != targetColumn {
				if !re.isAllowedTransition(currentColumn, targetColumn, allowedTransitions) {
					return &RuleViolation{
						RuleID:   rule.ID,
						Priority: rule.Priority,
						Message:  fmt.Sprintf("Invalid column transition from '%s' to '%s'", currentColumn, targetColumn),
						Category: rule.Category,
						Details:  fmt.Sprintf("Allowed transitions: %v", allowedTransitions),
					}
				}
			}
		}
	}

	return nil // No violation
}

// evaluateAutomationRule evaluates automation rules (e.g., age limits)
func (re *RuleEngine) evaluateAutomationRule(rule resource_access.Rule, context *EnrichedContext) *RuleViolation {
	// Age Limit Rule
	if maxAgeDays, exists := rule.Conditions["max_age_days"]; exists {
		maxAgeDaysInt, err := re.parseIntValue(maxAgeDays)
		if err != nil {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  fmt.Sprintf("Invalid max_age_days value: %v", maxAgeDays),
				Category: rule.Category,
			}
		}

		var columnEnterTime time.Time
		if context.Event.CurrentState != nil {
			currentColumn := context.Event.CurrentState.Status.Column
			if enterTime, exists := context.ColumnEnterTimes[currentColumn]; exists {
				columnEnterTime = enterTime
			}
		}
		if !columnEnterTime.IsZero() {
			age := time.Since(columnEnterTime)
			maxAge := time.Duration(maxAgeDaysInt) * 24 * time.Hour

			if age > maxAge {
				return &RuleViolation{
					RuleID:   rule.ID,
					Priority: rule.Priority,
					Message:  fmt.Sprintf("Task has been in column too long: %d days (limit: %d days)", int(age.Hours()/24), maxAgeDaysInt),
					Category: rule.Category,
					Details:  fmt.Sprintf("Age: %.1f days, Limit: %d days", age.Hours()/24, maxAgeDaysInt),
				}
			}
		}
	}

	return nil // No violation
}

// evaluateNotificationRule evaluates notification rules (currently placeholder)
func (re *RuleEngine) evaluateNotificationRule(rule resource_access.Rule, context *EnrichedContext) *RuleViolation {
	// Notification rules don't typically block actions, they trigger notifications
	// For now, we'll treat them as non-blocking
	return nil
}

// Helper methods

func (re *RuleEngine) parseIntValue(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot parse %v as integer", value)
	}
}

func (re *RuleEngine) checkRequiredField(fieldName string, task *resource_access.Task, rule resource_access.Rule) *RuleViolation {
	switch fieldName {
	case "title":
		if strings.TrimSpace(task.Title) == "" {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  "Task title is required",
				Category: rule.Category,
			}
		}
	case "description":
		if strings.TrimSpace(task.Description) == "" {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  "Task description is required",
				Category: rule.Category,
			}
		}
	}
	return nil
}

func (re *RuleEngine) isAllowedTransition(from, to string, allowedTransitions interface{}) bool {
	// Parse allowed transitions (could be map or array format)
	switch transitions := allowedTransitions.(type) {
	case map[string]any:
		if allowedList, exists := transitions[from]; exists {
			if allowed, ok := allowedList.([]interface{}); ok {
				for _, allowedTo := range allowed {
					if fmt.Sprintf("%v", allowedTo) == to {
						return true
					}
				}
			}
		}
	case []any:
		// Simple format: ["todo->doing", "doing->done"]
		transitionStr := fmt.Sprintf("%s->%s", from, to)
		for _, transition := range transitions {
			if fmt.Sprintf("%v", transition) == transitionStr {
				return true
			}
		}
	}
	return false
}


// Close releases any resources held by the engine
func (re *RuleEngine) Close() error {
	re.logger.LogMessage(utilities.Info, "RuleEngine", "Closing RuleEngine")
	return nil
}
