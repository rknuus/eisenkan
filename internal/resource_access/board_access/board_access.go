// Package board_access provides BoardAccess layer components implementing the iDesign methodology.
// This package contains components that provide data access and persistence services
// to higher-level components in the application architecture.
package board_access

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// Task represents a single task in the board
type Task struct {
	ID                    string            `json:"id"`
	Title                 string            `json:"title"`
	Description           string            `json:"description,omitempty"`
	Tags                  []string          `json:"tags,omitempty"`
	DueDate               *time.Time        `json:"due_date,omitempty"`
	PriorityPromotionDate *time.Time        `json:"priority_promotion_date,omitempty"`
	Metadata              map[string]string `json:"metadata,omitempty"`
	ParentTaskID          *string           `json:"parent_task_id,omitempty"`
}

// Priority represents Eisenhower matrix categorization (excludes not-urgent-not-important)
type Priority struct {
	Urgent    bool   `json:"urgent"`
	Important bool   `json:"important"`
	Label     string `json:"label"` // "urgent-important", "urgent-not-important", "not-urgent-important"
}

// WorkflowStatus tracks current workflow position
type WorkflowStatus struct {
	Column   string `json:"column"`   // e.g., "todo", "doing", "done"
	Section  string `json:"section"`  // e.g., "urgent-important" for todo column
	Position int    `json:"position"` // Order within column/section
}

// BoardConfiguration defines the board structure (simplified)
type BoardConfiguration struct {
	Name     string              `json:"name"`
	Columns  []string            `json:"columns"`   // ["todo", "doing", "done"]
	Sections map[string][]string `json:"sections"`  // column -> sections mapping
	GitUser  string              `json:"git_user"`  // Git commit author name
	GitEmail string              `json:"git_email"` // Git commit author email
}

// HierarchyFilter defines task hierarchy filtering options
type HierarchyFilter string

const (
	AllTasks      HierarchyFilter = "all"
	TopLevelOnly  HierarchyFilter = "top_level_only"
	SubtasksOnly  HierarchyFilter = "subtasks_only"
)

// CascadePolicy defines how parent task operations affect subtasks
type CascadePolicy string

const (
	NoAction         CascadePolicy = "no_action"
	ArchiveSubtasks  CascadePolicy = "archive_subtasks"
	DeleteSubtasks   CascadePolicy = "delete_subtasks"
	PromoteSubtasks  CascadePolicy = "promote_subtasks"
)

// DateRange specifies a date range for queries
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// QueryCriteria defines search parameters for task retrieval
type QueryCriteria struct {
	Columns               []string         `json:"columns,omitempty"`
	Sections              []string         `json:"sections,omitempty"`
	Priority              *Priority        `json:"priority,omitempty"`
	Tags                  []string         `json:"tags,omitempty"`
	DateRange             *DateRange       `json:"date_range,omitempty"`
	PriorityPromotionDate *DateRange       `json:"priority_promotion_date,omitempty"`
	ParentTaskID          *string          `json:"parent_task_id,omitempty"`
	Hierarchy             HierarchyFilter  `json:"hierarchy,omitempty"`
}

// TaskWithTimestamps represents a task with creation and modification timestamps
type TaskWithTimestamps struct {
	Task      *Task          `json:"task"`
	Priority  Priority       `json:"priority"`
	Status    WorkflowStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// RulesData contains all rule-related context data in a single structure
type RulesData struct {
	WIPCounts        map[string]int                               `json:"wip_counts"`        // column -> task count
	ColumnTasks      map[string][]*TaskWithTimestamps             `json:"column_tasks"`      // column -> tasks
	TaskHistory      []utilities.CommitInfo                       `json:"task_history"`      // for age calculations
	ColumnEnterTimes map[string]time.Time                         `json:"column_enter_times"` // column -> enter timestamp
	BoardMetadata    map[string]string                            `json:"board_metadata"`    // board configuration data
	SubtaskWIPCounts map[string]int                               `json:"subtask_wip_counts"` // column -> subtask count
	HierarchyMap     map[string][]string                          `json:"hierarchy_map"`     // parent_id -> child_ids
}

// IBoardAccess defines the contract for board data operations using faceted design
type IBoardAccess interface {
	// Task and subtask operations facet
	ITask

	// Rule engine helper operations facet
	IRules

	// Configuration management operations facet
	IConfiguration

	// Board management operations facet
	IBoard

	// Utility Operations
	Close() error
}

// boardAccess implements IBoardAccess
type boardAccess struct {
	repository utilities.Repository
	logger     utilities.ILoggingUtility
	mutex      *sync.RWMutex
	ITask          // embedded task facet
	IRules         // embedded rules facet
	IConfiguration // embedded configuration facet
	IBoard         // embedded board facet
}

// NewBoardAccess creates a new BoardAccess instance
func NewBoardAccess(repositoryPath string) (IBoardAccess, error) {
	logger := utilities.NewLoggingUtility()

	logger.LogMessage(utilities.Debug, "BoardAccess", "Initializing BoardAccess")

	// Load board configuration directly from file to get git settings
	configPath := filepath.Join(repositoryPath, "board.json")
	var config *BoardConfiguration

	if configData, err := os.ReadFile(configPath); err == nil {
		// Try to parse board configuration
		var parsedConfig BoardConfiguration
		if json.Unmarshal(configData, &parsedConfig) == nil {
			config = &parsedConfig
		}
	}

	// Use default configuration if loading fails or is incomplete
	if config == nil {
		config = &BoardConfiguration{
			Name:    "EisenKan Board",
			Columns: []string{"todo", "doing", "done"},
			Sections: map[string][]string{
				"todo": {"urgent-important", "urgent-not-important", "not-urgent-important"},
			},
			GitUser:  "BoardAccess",
			GitEmail: "boardaccess@eisenkan.local",
		}
	}

	// Ensure git configuration is complete
	if config.GitUser == "" {
		config.GitUser = "BoardAccess"
	}
	if config.GitEmail == "" {
		config.GitEmail = "boardaccess@eisenkan.local"
	}

	// Initialize repository with git configuration
	gitConfig := &utilities.AuthorConfiguration{
		User:  config.GitUser,
		Email: config.GitEmail,
	}

	repository, err := utilities.InitializeRepositoryWithConfig(repositoryPath, gitConfig)
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.NewBoardAccess failed to initialize repository with config: %w", err)
	}

	mutex := &sync.RWMutex{}
	taskFacetImpl := newTaskFacet(repository, logger, mutex)

	boardAccess := &boardAccess{
		repository:     repository,
		logger:         logger,
		mutex:          mutex,
		ITask:          taskFacetImpl,
		IRules:         newRulesFacet(taskFacetImpl, logger, mutex),
		IConfiguration: newConfigurationFacet(repository, logger),
		IBoard:         newBoardFacet(repository, logger, mutex, nil),
	}

	logger.LogMessage(utilities.Info, "BoardAccess", "BoardAccess initialized successfully")

	return boardAccess, nil
}

// Close implements the utility operation to clean up resources
func (ba *boardAccess) Close() error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Info, "BoardAccess", "Closing BoardAccess")

	// Close the repository
	if err := ba.repository.Close(); err != nil {
		return fmt.Errorf("failed to close repository: %w", err)
	}

	return nil
}