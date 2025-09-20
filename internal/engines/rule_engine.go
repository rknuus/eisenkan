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
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// TaskEvent represents a task state change event for rule evaluation
type TaskEvent struct {
	EventType        string                             `json:"event_type"` // "task_transition", "task_update", "task_create"
	CurrentState     *board_access.TaskWithTimestamps   `json:"current_state,omitempty"`
	FutureState      *TaskState                         `json:"future_state"`
	Timestamp        time.Time                          `json:"timestamp"`
	ParentTask       *board_access.TaskWithTimestamps   `json:"parent_task,omitempty"`
	AffectedSubtasks []*board_access.TaskWithTimestamps `json:"affected_subtasks,omitempty"`
}

// BoardConfigurationEvent represents a board configuration validation event
type BoardConfigurationEvent struct {
	EventType     string              `json:"event_type"` // "board_create", "board_update"
	Configuration *BoardConfiguration `json:"configuration"`
	Timestamp     time.Time           `json:"timestamp"`
}

// BoardConfiguration contains board metadata for validation
type BoardConfiguration struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// TaskState represents the intended state of a task
type TaskState struct {
	Task     *board_access.Task          `json:"task"`
	Priority board_access.Priority       `json:"priority"`
	Status   board_access.WorkflowStatus `json:"status"`
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
	Event            TaskEvent                                     `json:"event"`
	WIPCounts        map[string]int                                `json:"wip_counts"`         // column -> task count
	SubtaskWIPCounts map[string]int                                `json:"subtask_wip_counts"` // column -> subtask count
	TaskHistory      []utilities.CommitInfo                        `json:"task_history"`       // for age calculations
	Subtasks         []*board_access.TaskWithTimestamps            `json:"subtasks"`           // for dependency rules
	ColumnTasks      map[string][]*board_access.TaskWithTimestamps `json:"column_tasks"`       // for priority comparisons
	ColumnEnterTimes map[string]time.Time                          `json:"column_enter_times"` // column -> enter timestamp
	BoardMetadata    map[string]string                             `json:"board_metadata"`     // for custom rules
	HierarchyMap     map[string][]string                           `json:"hierarchy_map"`      // parent -> subtasks mapping
}

// IRuleEngine defines the interface for rule evaluation operations
type IRuleEngine interface {
	// EvaluateTaskChange evaluates whether a task change can be applied
	EvaluateTaskChange(ctx context.Context, event TaskEvent, boardPath string) (*RuleEvaluationResult, error)

	// EvaluateBoardConfigurationChange evaluates whether a board configuration change can be applied
	EvaluateBoardConfigurationChange(ctx context.Context, event BoardConfigurationEvent) (*RuleEvaluationResult, error)

	// Close releases any resources held by the engine
	Close() error
}

// RuleEngine implements IRuleEngine interface
type RuleEngine struct {
	rulesAccess resource_access.IRulesAccess
	boardAccess board_access.IBoardAccess
	logger      utilities.ILoggingUtility
}

// NewRuleEngine creates a new RuleEngine instance
func NewRuleEngine(rulesAccess resource_access.IRulesAccess, boardAccess board_access.IBoardAccess) (*RuleEngine, error) {
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

// EvaluateBoardConfigurationChange evaluates whether a board configuration change can be applied
func (re *RuleEngine) EvaluateBoardConfigurationChange(ctx context.Context, event BoardConfigurationEvent) (*RuleEvaluationResult, error) {
	re.logger.LogMessage(utilities.Debug, "RuleEngine", fmt.Sprintf("Evaluating board configuration change for event type: %s", event.EventType))

	// Validate input event
	if event.Configuration == nil {
		return nil, fmt.Errorf("RuleEngine.EvaluateBoardConfigurationChange: configuration cannot be nil")
	}

	// For board configuration validation, we use default rules since we don't have a specific board path
	// This could be extended to support board-specific rules in the future
	ruleSet := re.getDefaultBoardRules()

	// Filter rules based on event type and enabled status
	applicableRules := re.filterApplicableBoardRules(ruleSet, event.EventType)
	if len(applicableRules) == 0 {
		re.logger.LogMessage(utilities.Debug, "RuleEngine", "No applicable board rules found, allowing configuration change")
		return &RuleEvaluationResult{Allowed: true}, nil
	}

	// Evaluate all applicable rules using board configuration context
	violations := re.evaluateBoardRules(applicableRules, event)

	// Sort violations by priority (higher priority first)
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Priority > violations[j].Priority
	})

	result := &RuleEvaluationResult{
		Allowed:    len(violations) == 0,
		Violations: violations,
	}

	re.logger.LogMessage(utilities.Info, "RuleEngine",
		fmt.Sprintf("Board configuration evaluation completed: allowed=%t, violations=%d", result.Allowed, len(violations)))

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

// filterApplicableBoardRules filters board rules based on event type and enabled status
func (re *RuleEngine) filterApplicableBoardRules(rules []resource_access.Rule, eventType string) []resource_access.Rule {
	var applicable []resource_access.Rule

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Only consider board configuration rules
		if rule.Category != "board_configuration" {
			continue
		}

		// Check if rule triggers match the event type
		if rule.TriggerType == eventType || rule.TriggerType == "all" {
			applicable = append(applicable, rule)
		}
	}

	re.logger.LogMessage(utilities.Debug, "RuleEngine",
		fmt.Sprintf("Filtered %d applicable board rules from %d total rules", len(applicable), len(rules)))

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
	var subtasks []*board_access.TaskWithTimestamps
	if taskID != "" {
		var err error
		subtasks, err = re.boardAccess.GetSubtasks(taskID)
		if err != nil {
			re.logger.LogMessage(utilities.Warning, "RuleEngine", fmt.Sprintf("Failed to get subtasks: %v", err))
			subtasks = []*board_access.TaskWithTimestamps{} // Continue with empty list
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
func (re *RuleEngine) getSubtasks(ctx context.Context, event TaskEvent) ([]*board_access.TaskWithTimestamps, error) {
	// TODO: Implement subtask retrieval when subtask support is added to BoardAccess
	// For now, return empty list as subtasks are not yet implemented in BoardAccess
	return []*board_access.TaskWithTimestamps{}, nil
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

// evaluateBoardRules evaluates all board rules sequentially and aggregates violations
func (re *RuleEngine) evaluateBoardRules(rules []resource_access.Rule, event BoardConfigurationEvent) []RuleViolation {
	var violations []RuleViolation

	for _, rule := range rules {
		violation := re.evaluateBoardRule(rule, event)
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
	case "board_configuration":
		// Board configuration rules should not be evaluated in task context
		re.logger.LogMessage(utilities.Warning, "RuleEngine", "Board configuration rule encountered in task evaluation context")
		return nil
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

func (re *RuleEngine) checkRequiredField(fieldName string, task *board_access.Task, rule resource_access.Rule) *RuleViolation {
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

// evaluateBoardRule evaluates a single board configuration rule
func (re *RuleEngine) evaluateBoardRule(rule resource_access.Rule, event BoardConfigurationEvent) *RuleViolation {
	re.logger.LogMessage(utilities.Debug, "RuleEngine", fmt.Sprintf("Evaluating board rule: %s (%s)", rule.ID, rule.Name))

	config := event.Configuration

	// Board Title Validation
	if _, hasCondition := rule.Conditions["validate_title"]; hasCondition {
		if violation := re.validateBoardTitle(config.Title, rule); violation != nil {
			return violation
		}
	}

	// Board Description Validation
	if _, hasCondition := rule.Conditions["validate_description"]; hasCondition {
		if violation := re.validateBoardDescription(config.Description, rule); violation != nil {
			return violation
		}
	}

	// Board Configuration Format Validation
	if _, hasCondition := rule.Conditions["validate_format"]; hasCondition {
		if violation := re.validateBoardFormat(config, rule); violation != nil {
			return violation
		}
	}

	return nil // No violation
}

// getDefaultBoardRules returns the default set of board configuration validation rules
func (re *RuleEngine) getDefaultBoardRules() []resource_access.Rule {
	return []resource_access.Rule{
		{
			ID:          "board-title-validation",
			Name:        "Board Title Validation",
			Category:    "board_configuration",
			TriggerType: "all",
			Conditions: map[string]interface{}{
				"validate_title": true,
			},
			Actions:  map[string]interface{}{},
			Priority: 100,
			Enabled:  true,
		},
		{
			ID:          "board-description-validation",
			Name:        "Board Description Validation",
			Category:    "board_configuration",
			TriggerType: "all",
			Conditions: map[string]interface{}{
				"validate_description": true,
			},
			Actions:  map[string]interface{}{},
			Priority: 90,
			Enabled:  true,
		},
		{
			ID:          "board-format-validation",
			Name:        "Board Configuration Format Validation",
			Category:    "board_configuration",
			TriggerType: "all",
			Conditions: map[string]interface{}{
				"validate_format": true,
			},
			Actions:  map[string]interface{}{},
			Priority: 80,
			Enabled:  true,
		},
	}
}

// validateBoardTitle validates board title according to business rules
func (re *RuleEngine) validateBoardTitle(title string, rule resource_access.Rule) *RuleViolation {
	title = strings.TrimSpace(title)

	// Check if title is non-empty
	if title == "" {
		return &RuleViolation{
			RuleID:   rule.ID,
			Priority: rule.Priority,
			Message:  "Board title is required and cannot be empty",
			Category: rule.Category,
		}
	}

	// Check title length (max 100 characters)
	if len(title) > 100 {
		return &RuleViolation{
			RuleID:   rule.ID,
			Priority: rule.Priority,
			Message:  fmt.Sprintf("Board title exceeds maximum length: %d characters (limit: 100)", len(title)),
			Category: rule.Category,
			Details:  fmt.Sprintf("Title length: %d, Limit: 100", len(title)),
		}
	}

	// Check for valid characters (alphanumeric, spaces, and hyphens only)
	for _, char := range title {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == ' ' || char == '-') {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  fmt.Sprintf("Board title contains invalid character: '%c' (only alphanumeric, spaces, and hyphens allowed)", char),
				Category: rule.Category,
				Details:  fmt.Sprintf("Invalid character: '%c' at position %d", char, strings.IndexRune(title, char)),
			}
		}
	}

	return nil // No violation
}

// validateBoardDescription validates board description according to business rules
func (re *RuleEngine) validateBoardDescription(description string, rule resource_access.Rule) *RuleViolation {
	// Description is optional, but if provided, must meet length constraints
	if description != "" {
		description = strings.TrimSpace(description)

		// Check description length (max 500 characters when provided)
		if len(description) > 500 {
			return &RuleViolation{
				RuleID:   rule.ID,
				Priority: rule.Priority,
				Message:  fmt.Sprintf("Board description exceeds maximum length: %d characters (limit: 500)", len(description)),
				Category: rule.Category,
				Details:  fmt.Sprintf("Description length: %d, Limit: 500", len(description)),
			}
		}
	}

	return nil // No violation
}

// validateBoardFormat validates board configuration format and structure
func (re *RuleEngine) validateBoardFormat(config *BoardConfiguration, rule resource_access.Rule) *RuleViolation {
	// Check required fields are present
	if config == nil {
		return &RuleViolation{
			RuleID:   rule.ID,
			Priority: rule.Priority,
			Message:  "Board configuration cannot be null",
			Category: rule.Category,
		}
	}

	// Title is required (checked by validateBoardTitle, but also verify structure)
	if strings.TrimSpace(config.Title) == "" {
		return &RuleViolation{
			RuleID:   rule.ID,
			Priority: rule.Priority,
			Message:  "Board configuration must include a valid title field",
			Category: rule.Category,
		}
	}

	// Metadata validation (basic structure check)
	if config.Metadata != nil {
		for key, value := range config.Metadata {
			if strings.TrimSpace(key) == "" {
				return &RuleViolation{
					RuleID:   rule.ID,
					Priority: rule.Priority,
					Message:  "Board configuration metadata contains empty key",
					Category: rule.Category,
				}
			}
			if strings.TrimSpace(value) == "" {
				return &RuleViolation{
					RuleID:   rule.ID,
					Priority: rule.Priority,
					Message:  fmt.Sprintf("Board configuration metadata key '%s' has empty value", key),
					Category: rule.Category,
				}
			}
		}
	}

	return nil // No violation
}

// Close releases any resources held by the engine
func (re *RuleEngine) Close() error {
	re.logger.LogMessage(utilities.Info, "RuleEngine", "Closing RuleEngine")
	return nil
}
