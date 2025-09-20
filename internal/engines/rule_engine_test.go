package engines

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rknuus/eisenkan/internal/resource_access"
	"github.com/rknuus/eisenkan/internal/resource_access/board_access"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// Mock implementations for testing

type mockRulesAccess struct {
	ruleSet *resource_access.RuleSet
	err     error
}

func (m *mockRulesAccess) ReadRules(boardDirPath string) (*resource_access.RuleSet, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ruleSet, nil
}

func (m *mockRulesAccess) ValidateRuleChanges(ruleSet *resource_access.RuleSet) (*resource_access.ValidationResult, error) {
	return &resource_access.ValidationResult{Valid: true}, nil
}

func (m *mockRulesAccess) ChangeRules(boardDirPath string, ruleSet *resource_access.RuleSet) error {
	return nil
}

func (m *mockRulesAccess) Close() error {
	return nil
}

type mockBoardAccess struct {
	tasks       []*board_access.TaskWithTimestamps
	history     []utilities.CommitInfo
	config      *board_access.BoardConfiguration
	err         error
}

func (m *mockBoardAccess) CreateTask(task *board_access.Task, priority board_access.Priority, status board_access.WorkflowStatus, parentTaskID *string) (string, error) {
	return "", nil
}

func (m *mockBoardAccess) GetTasksData(taskIDs []string, includeHierarchy bool) ([]*board_access.TaskWithTimestamps, error) {
	return m.tasks, m.err
}

func (m *mockBoardAccess) ListTaskIdentifiers(hierarchyFilter board_access.HierarchyFilter) ([]string, error) {
	return nil, nil
}

func (m *mockBoardAccess) ChangeTaskData(taskID string, task *board_access.Task, priority board_access.Priority, status board_access.WorkflowStatus) error {
	return nil
}

func (m *mockBoardAccess) MoveTask(taskID string, priority board_access.Priority, status board_access.WorkflowStatus) error {
	return nil
}

func (m *mockBoardAccess) ArchiveTask(taskID string, cascadePolicy board_access.CascadePolicy) error {
	return nil
}

func (m *mockBoardAccess) RemoveTask(taskID string, cascadePolicy board_access.CascadePolicy) error {
	return nil
}

func (m *mockBoardAccess) FindTasks(criteria *board_access.QueryCriteria) ([]*board_access.TaskWithTimestamps, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasks, nil
}

func (m *mockBoardAccess) GetTaskHistory(taskID string, limit int) ([]utilities.CommitInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.history, nil
}

func (m *mockBoardAccess) GetBoardConfiguration() (*board_access.BoardConfiguration, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.config == nil {
		return &board_access.BoardConfiguration{Name: "Test Board"}, nil
	}
	return m.config, nil
}

func (m *mockBoardAccess) UpdateBoardConfiguration(config *board_access.BoardConfiguration) error {
	return nil
}

func (m *mockBoardAccess) GetRulesData(taskID string, targetColumns []string) (*board_access.RulesData, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	rulesData := &board_access.RulesData{
		WIPCounts:        make(map[string]int),
		SubtaskWIPCounts: make(map[string]int),
		ColumnTasks:      make(map[string][]*board_access.TaskWithTimestamps),
		ColumnEnterTimes: make(map[string]time.Time),
		BoardMetadata:    make(map[string]string),
		HierarchyMap:     make(map[string][]string),
	}
	
	// Build WIP counts and organize tasks by column
	for _, task := range m.tasks {
		// Separate WIP counts for tasks and subtasks
		if task.Task.ParentTaskID == nil {
			// Top-level task
			rulesData.WIPCounts[task.Status.Column]++
		} else {
			// Subtask
			rulesData.SubtaskWIPCounts[task.Status.Column]++
			
			// Build hierarchy map (parent -> subtasks)
			parentID := *task.Task.ParentTaskID
			rulesData.HierarchyMap[parentID] = append(rulesData.HierarchyMap[parentID], task.Task.ID)
		}
		
		// Group tasks by column (only for requested columns)
		if len(targetColumns) == 0 || containsString(targetColumns, task.Status.Column) {
			rulesData.ColumnTasks[task.Status.Column] = append(
				rulesData.ColumnTasks[task.Status.Column], task)
		}
	}
	
	// Mock task history if taskID provided
	if taskID != "" {
		rulesData.TaskHistory = m.history
		
		// Mock column enter times
		for _, column := range targetColumns {
			rulesData.ColumnEnterTimes[column] = time.Now().Add(-time.Hour)
		}
	}
	
	// Mock board metadata
	if m.config != nil {
		rulesData.BoardMetadata["board_name"] = m.config.Name
	} else {
		rulesData.BoardMetadata["board_name"] = "Test Board"
	}
	
	return rulesData, nil
}

// Helper function for mock
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (m *mockBoardAccess) Close() error {
	return nil
}

// GetSubtasks retrieves all subtasks for a given parent task
func (m *mockBoardAccess) GetSubtasks(parentTaskID string) ([]*board_access.TaskWithTimestamps, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	// Filter mock tasks that have this parent ID
	var subtasks []*board_access.TaskWithTimestamps
	for _, task := range m.tasks {
		if task.Task.ParentTaskID != nil && *task.Task.ParentTaskID == parentTaskID {
			subtasks = append(subtasks, task)
		}
	}
	return subtasks, nil
}

// GetParentTask retrieves the parent task for a given subtask
func (m *mockBoardAccess) GetParentTask(subtaskID string) (*board_access.TaskWithTimestamps, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	// Find the subtask first
	var subtask *board_access.TaskWithTimestamps
	for _, task := range m.tasks {
		if task.Task.ID == subtaskID {
			subtask = task
			break
		}
	}
	
	if subtask == nil || subtask.Task.ParentTaskID == nil {
		return nil, nil // No parent or subtask doesn't exist
	}
	
	// Find the parent task
	parentID := *subtask.Task.ParentTaskID
	for _, task := range m.tasks {
		if task.Task.ID == parentID {
			return task, nil
		}
	}
	return nil, nil // Parent doesn't exist
}

// IConfiguration facet mock methods
func (m *mockBoardAccess) Load(configType string, identifier string) (board_access.ConfigurationData, error) {
	// Return empty configuration data for tests
	return board_access.ConfigurationData{
		Type:       configType,
		Identifier: identifier,
		Version:    "1.0",
		Settings:   make(map[string]interface{}),
		Schema:     "default",
		Metadata:   make(map[string]string),
	}, nil
}

func (m *mockBoardAccess) Store(configType string, identifier string, data board_access.ConfigurationData) error {
	// Mock store operation does nothing for tests
	return nil
}

// ChangeTask updates task data, priority, and status
func (m *mockBoardAccess) ChangeTask(taskID string, task *board_access.Task, priority board_access.Priority, status board_access.WorkflowStatus) error {
	return nil
}

// IBoard facet mock methods
func (m *mockBoardAccess) Discover(ctx context.Context, directoryPath string) ([]board_access.BoardDiscoveryResult, error) {
	return []board_access.BoardDiscoveryResult{}, nil
}

func (m *mockBoardAccess) ExtractMetadata(ctx context.Context, boardPath string) (*board_access.BoardMetadata, error) {
	return &board_access.BoardMetadata{
		Title:     "Mock Board",
		TaskCount: len(m.tasks),
	}, nil
}

func (m *mockBoardAccess) GetStatistics(ctx context.Context, boardPath string) (*board_access.BoardStatistics, error) {
	return &board_access.BoardStatistics{
		TotalTasks:    len(m.tasks),
		ActiveTasks:   len(m.tasks),
		TasksByColumn: make(map[string]int),
	}, nil
}

func (m *mockBoardAccess) ValidateStructure(ctx context.Context, boardPath string) (*board_access.BoardValidationResult, error) {
	return &board_access.BoardValidationResult{
		IsValid:       true,
		GitRepoValid:  true,
		ConfigValid:   true,
		DataIntegrity: true,
	}, nil
}

func (m *mockBoardAccess) LoadConfiguration(ctx context.Context, boardPath string, configType string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":    "Mock Board",
		"columns": []string{"todo", "doing", "done"},
	}, nil
}

func (m *mockBoardAccess) StoreConfiguration(ctx context.Context, boardPath string, configType string, configData map[string]interface{}) error {
	return nil
}

func (m *mockBoardAccess) Create(ctx context.Context, request *board_access.BoardCreationRequest) (*board_access.BoardCreationResult, error) {
	return &board_access.BoardCreationResult{
		Success:        true,
		BoardPath:      request.BoardPath,
		GitInitialized: request.InitializeGit,
	}, nil
}

func (m *mockBoardAccess) Delete(ctx context.Context, request *board_access.BoardDeletionRequest) (*board_access.BoardDeletionResult, error) {
	return &board_access.BoardDeletionResult{
		Success: true,
		Method:  "permanent",
	}, nil
}

// Test helper functions

func createMockTask(id, title, column string) *board_access.TaskWithTimestamps {
	return &board_access.TaskWithTimestamps{
		Task: &board_access.Task{
			ID:    id,
			Title: title,
		},
		Priority: board_access.Priority{
			Urgent:    false,
			Important: true,
			Label:     "not-urgent-important",
		},
		Status: board_access.WorkflowStatus{
			Column:   column,
			Section:  "not-urgent-important",
			Position: 1,
		},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
}

func TestNewRuleEngine(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		rulesAccess := &mockRulesAccess{}
		boardAccess := &mockBoardAccess{}

		engine, err := NewRuleEngine(rulesAccess, boardAccess)

		if err != nil {
			t.Errorf("NewRuleEngine() error = %v, want nil", err)
		}
		if engine == nil {
			t.Error("NewRuleEngine() returned nil engine")
		}
	})

	t.Run("nil rulesAccess", func(t *testing.T) {
		boardAccess := &mockBoardAccess{}

		engine, err := NewRuleEngine(nil, boardAccess)

		if err == nil {
			t.Error("NewRuleEngine() with nil rulesAccess should return error")
		}
		if engine != nil {
			t.Error("NewRuleEngine() with nil rulesAccess should return nil engine")
		}
	})

	t.Run("nil boardAccess", func(t *testing.T) {
		rulesAccess := &mockRulesAccess{}

		engine, err := NewRuleEngine(rulesAccess, nil)

		if err == nil {
			t.Error("NewRuleEngine() with nil boardAccess should return error")
		}
		if engine != nil {
			t.Error("NewRuleEngine() with nil boardAccess should return nil engine")
		}
	})
}

func TestEvaluateTaskChange_NoRules(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules:   []resource_access.Rule{},
		},
	}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:    "task1",
				Title: "Test Task",
			},
			Status: board_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if !result.Allowed {
		t.Error("EvaluateTaskChange() with no rules should allow task change")
	}
	if len(result.Violations) != 0 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 0", len(result.Violations))
	}
}

func TestEvaluateTaskChange_WIPLimit(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "wip-limit-doing",
					Name:        "WIP Limit for Doing Column",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]interface{}{
						"max_wip_limit": 2,
					},
					Priority: 100,
					Enabled:  true,
				},
			},
		},
	}

	// Mock board access with 2 tasks already in "doing" column
	boardAccess := &mockBoardAccess{
		tasks: []*board_access.TaskWithTimestamps{
			createMockTask("task1", "Existing Task 1", "doing"),
			createMockTask("task2", "Existing Task 2", "doing"),
		},
	}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Try to move a task from "todo" to "doing" (would exceed WIP limit)
	event := TaskEvent{
		EventType: "task_transition",
		CurrentState: createMockTask("task3", "Moving Task", "todo"),
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:    "task3",
				Title: "Moving Task",
			},
			Status: board_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateTaskChange() should not allow task change when WIP limit exceeded")
	}
	if len(result.Violations) != 1 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 1", len(result.Violations))
	}
	if len(result.Violations) > 0 && result.Violations[0].RuleID != "wip-limit-doing" {
		t.Errorf("EvaluateTaskChange() violation rule ID = %s, want wip-limit-doing", result.Violations[0].RuleID)
	}
}

func TestEvaluateTaskChange_RequiredFields(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "required-fields",
					Name:        "Required Fields Rule",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]interface{}{
						"required_fields": []interface{}{"title", "description"},
					},
					Priority: 90,
					Enabled:  true,
				},
			},
		},
	}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Task with missing description
	event := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:          "task1",
				Title:       "Test Task",
				Description: "", // Missing description
			},
			Status: board_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateTaskChange() should not allow task with missing required fields")
	}
	if len(result.Violations) != 1 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 1", len(result.Violations))
	}
}

func TestEvaluateTaskChange_WorkflowTransition(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "workflow-transition",
					Name:        "Workflow Transition Rule",
					Category:    "workflow",
					TriggerType: "task_transition",
					Conditions: map[string]interface{}{
						"allowed_transitions": []interface{}{
							"todo->doing",
							"doing->done",
						},
					},
					Priority: 80,
					Enabled:  true,
				},
			},
		},
	}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Invalid transition: todo -> done (skipping doing)
	event := TaskEvent{
		EventType: "task_transition",
		CurrentState: createMockTask("task1", "Test Task", "todo"),
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:    "task1",
				Title: "Test Task",
			},
			Status: board_access.WorkflowStatus{
				Column: "done",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateTaskChange() should not allow invalid workflow transition")
	}
	if len(result.Violations) != 1 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 1", len(result.Violations))
	}
}

func TestEvaluateTaskChange_MultipleRules(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "high-priority-rule",
					Name:        "High Priority Rule",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]interface{}{
						"required_fields": []interface{}{"title"},
					},
					Priority: 100,
					Enabled:  true,
				},
				{
					ID:          "low-priority-rule",
					Name:        "Low Priority Rule",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]interface{}{
						"required_fields": []interface{}{"description"},
					},
					Priority: 50,
					Enabled:  true,
				},
			},
		},
	}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Task violating both rules
	event := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:          "task1",
				Title:       "", // Missing title (high priority)
				Description: "", // Missing description (low priority)
			},
			Status: board_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateTaskChange() should not allow task with multiple rule violations")
	}
	if len(result.Violations) != 2 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 2", len(result.Violations))
	}

	// Check that violations are sorted by priority (high priority first)
	if len(result.Violations) >= 2 {
		if result.Violations[0].Priority < result.Violations[1].Priority {
			t.Error("EvaluateTaskChange() violations should be sorted by priority (descending)")
		}
	}
}

func TestEvaluateTaskChange_DisabledRule(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "disabled-rule",
					Name:        "Disabled Rule",
					Category:    "validation",
					TriggerType: "task_transition",
					Conditions: map[string]interface{}{
						"required_fields": []interface{}{"description"},
					},
					Priority: 100,
					Enabled:  false, // Disabled
				},
			},
		},
	}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Task would violate the rule if it were enabled
	event := TaskEvent{
		EventType: "task_transition",
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:          "task1",
				Title:       "Test Task",
				Description: "", // Missing description
			},
			Status: board_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if !result.Allowed {
		t.Error("EvaluateTaskChange() should allow task change when rules are disabled")
	}
	if len(result.Violations) != 0 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 0", len(result.Violations))
	}
}

func TestEvaluateTaskChange_WrongEventType(t *testing.T) {
	rulesAccess := &mockRulesAccess{
		ruleSet: &resource_access.RuleSet{
			Version: "1.0",
			Rules: []resource_access.Rule{
				{
					ID:          "task-transition-rule",
					Name:        "Task Transition Rule",
					Category:    "validation",
					TriggerType: "task_transition", // Only triggers on transitions
					Conditions: map[string]interface{}{
						"required_fields": []interface{}{"description"},
					},
					Priority: 100,
					Enabled:  true,
				},
			},
		},
	}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Event type doesn't match rule trigger
	event := TaskEvent{
		EventType: "task_update", // Different from rule trigger type
		FutureState: &TaskState{
			Task: &board_access.Task{
				ID:          "task1",
				Title:       "Test Task",
				Description: "", // Missing description
			},
			Status: board_access.WorkflowStatus{
				Column: "doing",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateTaskChange(context.Background(), event, "/test/board")

	if err != nil {
		t.Errorf("EvaluateTaskChange() error = %v, want nil", err)
	}
	if !result.Allowed {
		t.Error("EvaluateTaskChange() should allow task change when event type doesn't match rule trigger")
	}
	if len(result.Violations) != 0 {
		t.Errorf("EvaluateTaskChange() violations = %d, want 0", len(result.Violations))
	}
}

func TestParseIntValue(t *testing.T) {
	engine := &RuleEngine{} // Don't need full setup for unit test

	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"int value", 42, 42, false},
		{"float64 value", 42.0, 42, false},
		{"string value", "42", 42, false},
		{"invalid string", "not-a-number", 0, true},
		{"nil value", nil, 0, true},
		{"bool value", true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.parseIntValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseIntValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAllowedTransition(t *testing.T) {
	engine := &RuleEngine{} // Don't need full setup for unit test

	tests := []struct {
		name               string
		from               string
		to                 string
		allowedTransitions interface{}
		want               bool
	}{
		{
			name: "simple array format - allowed",
			from: "todo",
			to:   "doing",
			allowedTransitions: []interface{}{
				"todo->doing",
				"doing->done",
			},
			want: true,
		},
		{
			name: "simple array format - not allowed",
			from: "todo",
			to:   "done",
			allowedTransitions: []interface{}{
				"todo->doing",
				"doing->done",
			},
			want: false,
		},
		{
			name: "map format - allowed",
			from: "todo",
			to:   "doing",
			allowedTransitions: map[string]interface{}{
				"todo":  []interface{}{"doing"},
				"doing": []interface{}{"done"},
			},
			want: true,
		},
		{
			name: "map format - not allowed",
			from: "todo",
			to:   "done",
			allowedTransitions: map[string]interface{}{
				"todo":  []interface{}{"doing"},
				"doing": []interface{}{"done"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.isAllowedTransition(tt.from, tt.to, tt.allowedTransitions)
			if got != tt.want {
				t.Errorf("isAllowedTransition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	err = engine.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// Board Configuration Validation Tests

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_ValidConfiguration(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title:       "My Project Board",
			Description: "A valid board configuration",
			Metadata: map[string]string{
				"project": "test",
				"team":    "dev",
			},
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if !result.Allowed {
		t.Errorf("EvaluateBoardConfigurationChange() Allowed = %v, want true", result.Allowed)
	}
	if len(result.Violations) != 0 {
		t.Errorf("EvaluateBoardConfigurationChange() violations = %d, want 0", len(result.Violations))
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_ValidConfigurationWithHyphens(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title: "Project-Board 2024 Version-2",
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if !result.Allowed {
		t.Errorf("EvaluateBoardConfigurationChange() Allowed = %v, want true", result.Allowed)
	}
	if len(result.Violations) != 0 {
		t.Errorf("EvaluateBoardConfigurationChange() violations = %d, want 0", len(result.Violations))
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_NilConfiguration(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := BoardConfigurationEvent{
		EventType:     "board_create",
		Configuration: nil,
		Timestamp:     time.Now(),
	}

	_, err = engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err == nil {
		t.Error("EvaluateBoardConfigurationChange() with nil configuration should return error")
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_EmptyTitle(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title:       "",
			Description: "Board with empty title",
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should not allow empty title")
	}
	if len(result.Violations) == 0 {
		t.Error("EvaluateBoardConfigurationChange() should have violations for empty title")
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_WhitespaceOnlyTitle(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title: "   \t\n  ", // only whitespace
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should not allow whitespace-only title")
	}
	if len(result.Violations) == 0 {
		t.Error("EvaluateBoardConfigurationChange() should have violations for whitespace-only title")
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_TitleTooLong(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Create 101-character title (exceeds limit)
	longTitle := strings.Repeat("A", 101)
	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title:       longTitle,
			Description: "Board with title exceeding 100 character limit",
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should not allow title exceeding 100 characters")
	}
	if len(result.Violations) == 0 {
		t.Error("EvaluateBoardConfigurationChange() should have violations for long title")
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_TitleBoundary100Characters(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Test exactly 100 characters (should pass)
	exactly100 := strings.Repeat("A", 100)
	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title: exactly100,
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if !result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should allow title with exactly 100 characters")
	}
	if len(result.Violations) != 0 {
		t.Errorf("EvaluateBoardConfigurationChange() violations = %d, want 0", len(result.Violations))
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_InvalidCharacters(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title:       "Board@#$%Title",
			Description: "Board with invalid characters in title",
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should not allow title with invalid characters")
	}
	if len(result.Violations) == 0 {
		t.Error("EvaluateBoardConfigurationChange() should have violations for invalid characters")
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_DescriptionTooLong(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Create 501-character description (exceeds limit)
	longDescription := strings.Repeat("A", 501)
	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title:       "Valid Title",
			Description: longDescription,
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should not allow description exceeding 500 characters")
	}
	if len(result.Violations) == 0 {
		t.Error("EvaluateBoardConfigurationChange() should have violations for long description")
	}
}

func TestUnit_RuleEngine_EvaluateBoardConfigurationChange_MultipleViolations(t *testing.T) {
	rulesAccess := &mockRulesAccess{}
	boardAccess := &mockBoardAccess{}

	engine, err := NewRuleEngine(rulesAccess, boardAccess)
	if err != nil {
		t.Fatalf("NewRuleEngine() error = %v", err)
	}

	// Configuration with multiple violations
	event := BoardConfigurationEvent{
		EventType: "board_create",
		Configuration: &BoardConfiguration{
			Title:       "Board@#$", // Invalid characters
			Description: strings.Repeat("A", 501), // Too long
		},
		Timestamp: time.Now(),
	}

	result, err := engine.EvaluateBoardConfigurationChange(context.Background(), event)
	if err != nil {
		t.Errorf("EvaluateBoardConfigurationChange() error = %v, want nil", err)
	}
	if result.Allowed {
		t.Error("EvaluateBoardConfigurationChange() should not allow configuration with multiple violations")
	}
	if len(result.Violations) < 2 {
		t.Errorf("EvaluateBoardConfigurationChange() violations = %d, want at least 2", len(result.Violations))
	}
}