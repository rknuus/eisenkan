// Package resource_access provides ResourceAccess layer components implementing the iDesign methodology.
// This package contains components that provide data access and persistence services
// to higher-level components in the application architecture.
package resource_access

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// Task represents a single task in the board
type Task struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	DueDate     *time.Time        `json:"due_date,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
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
	GitUser  string              `json:"git_user"`  // Git commit author
	GitEmail string              `json:"git_email"` // Git commit email
}

// QueryCriteria defines search parameters for task retrieval
type QueryCriteria struct {
	Columns   []string   `json:"columns,omitempty"`
	Sections  []string   `json:"sections,omitempty"`
	Priority  *Priority  `json:"priority,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	DateRange *DateRange `json:"date_range,omitempty"`
	Archived  *bool      `json:"archived,omitempty"`
	DateType  string     `json:"date_type,omitempty"` // "created" or "updated"
}

// DateRange specifies temporal constraints
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// TaskWithTimestamps includes git-derived timestamps and location-derived metadata
type TaskWithTimestamps struct {
	Task      *Task          `json:"task"`
	Priority  Priority       `json:"priority"` // Derived from file location
	Status    WorkflowStatus `json:"status"`   // Derived from file location
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
}

// IBoardAccess defines the contract for board data operations
type IBoardAccess interface {
	CreateTask(task *Task, priority Priority, status WorkflowStatus) (string, error)
	GetTasksData(taskIDs []string) ([]*TaskWithTimestamps, error)
	ListTaskIdentifiers() ([]string, error)
	ChangeTaskData(taskID string, task *Task, priority Priority, status WorkflowStatus) error
	MoveTask(taskID string, priority Priority, status WorkflowStatus) error
	ArchiveTask(taskID string) error
	RemoveTask(taskID string) error
	FindTasks(criteria *QueryCriteria) ([]*TaskWithTimestamps, error)
	GetTaskHistory(taskID string, limit int) ([]utilities.CommitInfo, error)

	// Board Configuration Operations
	GetBoardConfiguration() (*BoardConfiguration, error)
	UpdateBoardConfiguration(config *BoardConfiguration) error

	// Rule Engine Helper Operations
	GetRulesData(taskID string, targetColumns []string) (*RulesData, error)

	// Utility Operations
	Close() error
}

// boardAccess implements IBoardAccess
type boardAccess struct {
	repository utilities.Repository
	logger     utilities.ILoggingUtility
	mutex      *sync.RWMutex
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

	boardAccess := &boardAccess{
		repository: repository,
		logger:     logger,
		mutex:      &sync.RWMutex{},
	}

	logger.LogMessage(utilities.Info, "BoardAccess", "BoardAccess initialized successfully")

	return boardAccess, nil
}

// CreateTask stores a new task and returns its ID
func (ba *boardAccess) CreateTask(task *Task, priority Priority, status WorkflowStatus) (string, error) {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	// Early validation to prevent nil pointer panics
	if task == nil {
		return "", fmt.Errorf("BoardAccess.CreateTask task validation failed: task cannot be nil")
	}

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Storing new task: %s", task.Title))

	// Validate task content
	if err := ba.validateTask(task); err != nil {
		return "", fmt.Errorf("BoardAccess.CreateTask task validation failed: %w", err)
	}

	// Validate priority and status
	if err := ba.validatePriority(priority); err != nil {
		return "", fmt.Errorf("BoardAccess.CreateTask priority validation failed: %w", err)
	}

	// Auto-correct priority label after validation
	priority.Label = ba.generatePriorityLabel(priority.Urgent, priority.Important)

	// Generate ID if not provided
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// Determine file path with position prefix
	filePath, err := ba.getTaskFilePath(task.ID, priority, status)
	if err != nil {
		return "", fmt.Errorf("BoardAccess.CreateTask failed to determine file path: %w", err)
	}

	// Store task and commit
	if err := ba.writeTaskFile(task, filePath); err != nil {
		return "", fmt.Errorf("BoardAccess.CreateTask failed to write task: %w", err)
	}

	if err := ba.commitChange(filePath, fmt.Sprintf("Add task: %s", task.Title)); err != nil {
		return "", fmt.Errorf("BoardAccess.CreateTask failed to commit task: %w", err)
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Task stored successfully: %s", task.ID))

	return task.ID, nil
}

// GetTasksData retrieves tasks by IDs (combined method)
func (ba *boardAccess) GetTasksData(taskIDs []string) ([]*TaskWithTimestamps, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Retrieving %d tasks", len(taskIDs)))

	var tasks []*TaskWithTimestamps
	for _, taskID := range taskIDs {
		taskWithTimestamps, err := ba.retrieveTaskWithTimestamps(taskID)
		if err != nil {
			return nil, fmt.Errorf("BoardAccess.GetTasksData failed to retrieve task %s: %w", taskID, err)
		}
		if taskWithTimestamps != nil {
			tasks = append(tasks, taskWithTimestamps)
		}
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Retrieved %d/%d tasks", len(tasks), len(taskIDs)))

	return tasks, nil
}

// ListTaskIdentifiers returns all task IDs
func (ba *boardAccess) ListTaskIdentifiers() ([]string, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", "Retrieving all task identifiers")

	taskFiles, err := ba.getAllTaskFiles()
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.ListTaskIdentifiers failed to get task files: %w", err)
	}

	var taskIDs []string
	for _, filePath := range taskFiles {
		taskID := ba.extractTaskIDFromPath(filePath)
		taskIDs = append(taskIDs, taskID)
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Retrieved %d task identifiers", len(taskIDs)))

	return taskIDs, nil
}

// ChangeTaskData updates an existing task (content + location)
func (ba *boardAccess) ChangeTaskData(taskID string, task *Task, priority Priority, status WorkflowStatus) error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Updating task: %s", taskID))

	// Early validation to prevent nil pointer panics
	if task == nil {
		return fmt.Errorf("BoardAccess.ChangeTaskData task validation failed: task cannot be nil")
	}

	// Validate inputs
	if err := ba.validateTask(task); err != nil {
		return fmt.Errorf("BoardAccess.ChangeTaskData task validation failed: %w", err)
	}

	if err := ba.validatePriority(priority); err != nil {
		return fmt.Errorf("BoardAccess.ChangeTaskData priority validation failed: %w", err)
	}

	// Auto-correct priority label
	priority.Label = ba.generatePriorityLabel(priority.Urgent, priority.Important)
	task.ID = taskID

	// Find current task file and determine new path
	oldPath, err := ba.findTaskFile(taskID)
	if err != nil {
		return fmt.Errorf("BoardAccess.ChangeTaskData failed to find task %s: %w", taskID, err)
	}

	newPath, err := ba.getTaskFilePath(taskID, priority, status)
	if err != nil {
		return fmt.Errorf("BoardAccess.ChangeTaskData failed to determine new file path: %w", err)
	}

	// Handle file location/name change
	if oldPath != newPath {
		if err := os.Remove(oldPath); err != nil {
			return fmt.Errorf("BoardAccess.ChangeTaskData failed to remove old file: %w", err)
		}

		if err := ba.writeTaskFile(task, newPath); err != nil {
			return fmt.Errorf("BoardAccess.ChangeTaskData failed to write new file: %w", err)
		}

		// Stage both old and new files, then commit
		if err := ba.stageFiles([]string{oldPath, newPath}); err != nil {
			return fmt.Errorf("BoardAccess.ChangeTaskData failed to stage files: %w", err)
		}

		if err := ba.commitWithConfig(fmt.Sprintf("Move and update task: %s", task.Title)); err != nil {
			return fmt.Errorf("BoardAccess.ChangeTaskData failed to commit move: %w", err)
		}
	} else {
		if err := ba.writeTaskFile(task, newPath); err != nil {
			return fmt.Errorf("BoardAccess.ChangeTaskData failed to write file: %w", err)
		}

		if err := ba.commitChange(newPath, fmt.Sprintf("Update task: %s", task.Title)); err != nil {
			return fmt.Errorf("BoardAccess.ChangeTaskData failed to commit: %w", err)
		}
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Task updated successfully: %s", taskID))

	return nil
}

// MoveTask moves a task without changing content (just priority/status)
func (ba *boardAccess) MoveTask(taskID string, priority Priority, status WorkflowStatus) error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Moving task: %s", taskID))

	// Validate priority
	if err := ba.validatePriority(priority); err != nil {
		return fmt.Errorf("BoardAccess.MoveTask priority validation failed: %w", err)
	}

	// Find current task file
	oldPath, err := ba.findTaskFile(taskID)
	if err != nil {
		return fmt.Errorf("BoardAccess.MoveTask failed to find task %s: %w", taskID, err)
	}

	// Load existing task content
	task, err := ba.loadTaskFromFile(oldPath)
	if err != nil {
		return fmt.Errorf("BoardAccess.MoveTask failed to load task: %w", err)
	}

	// Auto-correct priority label
	priority.Label = ba.generatePriorityLabel(priority.Urgent, priority.Important)

	// Determine new path
	newPath, err := ba.getTaskFilePath(taskID, priority, status)
	if err != nil {
		return fmt.Errorf("BoardAccess.MoveTask failed to determine new file path: %w", err)
	}

	// Only proceed if location actually changes
	if oldPath == newPath {
		ba.logger.LogMessage(utilities.Debug, "BoardAccess", "Task location unchanged, no move needed")
		return nil
	}

	// Move task file
	return ba.moveTaskFile(oldPath, task, newPath, fmt.Sprintf("Move task: %s", task.Title))
}

// ArchiveTask moves a task to the archived directory
func (ba *boardAccess) ArchiveTask(taskID string) error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Archiving task: %s", taskID))

	// Find and load current task
	oldPath, err := ba.findTaskFile(taskID)
	if err != nil {
		return fmt.Errorf("BoardAccess.ArchiveTask failed to find task %s: %w", taskID, err)
	}

	task, err := ba.loadTaskFromFile(oldPath)
	if err != nil {
		return fmt.Errorf("BoardAccess.ArchiveTask failed to load task: %w", err)
	}

	// Create archived status
	archivedStatus := WorkflowStatus{
		Column:   "archived",
		Section:  "",
		Position: 1, // Could be made configurable
	}

	// Determine archive path
	newPath, err := ba.getTaskFilePath(taskID, Priority{}, archivedStatus) // Priority doesn't matter for archived
	if err != nil {
		return fmt.Errorf("BoardAccess.ArchiveTask failed to determine archive path: %w", err)
	}

	// Move to archive
	return ba.moveTaskFile(oldPath, task, newPath, fmt.Sprintf("Archive task: %s", task.Title))
}

// RemoveTask permanently deletes a task
func (ba *boardAccess) RemoveTask(taskID string) error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Removing task: %s", taskID))

	// Find task file
	filePath, err := ba.findTaskFile(taskID)
	if err != nil {
		// Task not found - idempotent operation
		ba.logger.LogMessage(utilities.Debug, "BoardAccess", "Task not found for removal (idempotent)")
		return nil
	}

	// Remove file and commit
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("BoardAccess.RemoveTask failed to remove file: %w", err)
	}

	if err := ba.commitChange(filePath, fmt.Sprintf("Remove task: %s", taskID)); err != nil {
		return fmt.Errorf("BoardAccess.RemoveTask failed to commit removal: %w", err)
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Task removed successfully: %s", taskID))

	return nil
}

// FindTasks searches for tasks matching the given criteria
func (ba *boardAccess) FindTasks(criteria *QueryCriteria) ([]*TaskWithTimestamps, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", "Querying tasks")

	// Get all task files
	taskFiles, err := ba.getAllTaskFiles()
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.FindTasks failed to get task files: %w", err)
	}

	var matchingTasks []*TaskWithTimestamps
	for _, filePath := range taskFiles {
		taskWithTimestamps, err := ba.loadTaskWithTimestampsFromFile(filePath)
		if err != nil {
			ba.logger.Log(utilities.Warning, "BoardAccess", "Failed to load task during query", map[string]any{
				"file_path": filePath,
				"error":     err.Error(),
			})
			continue
		}

		if ba.taskMatchesCriteria(taskWithTimestamps, criteria) {
			matchingTasks = append(matchingTasks, taskWithTimestamps)
		}
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Query found %d/%d matching tasks", len(matchingTasks), len(taskFiles)))

	return matchingTasks, nil
}

// GetTaskHistory retrieves version history for a specific task with configurable limit
func (ba *boardAccess) GetTaskHistory(taskID string, limit int) ([]utilities.CommitInfo, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	if limit <= 0 {
		limit = 100 // Default limit
	}

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Getting task history: %s (limit: %d)", taskID, limit))

	// Find task file
	filePath, err := ba.findTaskFile(taskID)
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.GetTaskHistory failed to find task %s: %w", taskID, err)
	}

	// Get relative path for version control
	repoPath := ba.repository.Path()
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.GetTaskHistory failed to get relative path: %w", err)
	}

	// Get file history from version control
	history, err := ba.repository.GetFileHistory(relPath, limit)
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.GetTaskHistory failed to get file history: %w", err)
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", fmt.Sprintf("Retrieved task history: %s (%d commits)", taskID, len(history)))

	return history, nil
}

// GetBoardConfiguration retrieves the current board configuration
func (ba *boardAccess) GetBoardConfiguration() (*BoardConfiguration, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", "Getting board configuration")

	config, err := ba.loadBoardConfiguration()
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.GetBoardConfiguration failed: %w", err)
	}

	return config, nil
}

// UpdateBoardConfiguration updates the board configuration
func (ba *boardAccess) UpdateBoardConfiguration(config *BoardConfiguration) error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", fmt.Sprintf("Updating board configuration: %s", config.Name))

	if err := ba.saveBoardConfiguration(config); err != nil {
		return fmt.Errorf("BoardAccess.UpdateBoardConfiguration failed: %w", err)
	}

	ba.logger.LogMessage(utilities.Info, "BoardAccess", "Board configuration updated successfully")

	return nil
}

// Close closes the BoardAccess instance and releases resources
func (ba *boardAccess) Close() error {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	ba.logger.LogMessage(utilities.Info, "BoardAccess", "Closing BoardAccess")

	if ba.repository != nil {
		if err := ba.repository.Close(); err != nil {
			return fmt.Errorf("BoardAccess.Close failed to close repository: %w", err)
		}
	}

	return nil
}

// Helper methods

// validateTask validates task content only (no priority/status)
func (ba *boardAccess) validateTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf("task title cannot be empty")
	}

	return nil
}

// validatePriority validates priority combination (no not-urgent-not-important)
func (ba *boardAccess) validatePriority(priority Priority) error {
	if !priority.Urgent && !priority.Important {
		return fmt.Errorf("priority combination 'not-urgent-not-important' is not supported")
	}
	return nil
}

// generatePriorityLabel generates priority label from urgent/important flags
func (ba *boardAccess) generatePriorityLabel(urgent, important bool) string {
	switch {
	case urgent && important:
		return "urgent-important"
	case urgent && !important:
		return "urgent-not-important"
	case !urgent && important:
		return "not-urgent-important"
	default:
		return "invalid" // Should not happen due to validation
	}
}

// getTaskFilePath determines the file path for a task with position prefix
func (ba *boardAccess) getTaskFilePath(taskID string, _ Priority, status WorkflowStatus) (string, error) {
	repoPath := ba.repository.Path()

	// Handle archived tasks
	if status.Column == "archived" {
		fileName := fmt.Sprintf("%04d-task-%s.json", status.Position, taskID)
		return filepath.Join(repoPath, "archived", fileName), nil
	}

	// Handle active tasks with column position prefix
	columnPath := ba.getColumnPath(status.Column)
	fileName := fmt.Sprintf("%04d-task-%s.json", status.Position, taskID)

	if status.Section != "" {
		return filepath.Join(repoPath, columnPath, status.Section, fileName), nil
	}

	return filepath.Join(repoPath, columnPath, fileName), nil
}

// getColumnPath returns the directory path for a column (with position prefix)
func (ba *boardAccess) getColumnPath(column string) string {
	config, err := ba.loadBoardConfiguration()
	if err != nil {
		// Fallback to column name without position
		return column
	}

	for i, col := range config.Columns {
		if col == column {
			return fmt.Sprintf("%02d_%s", i+1, column)
		}
	}

	return column // Fallback
}

// findTaskFile searches for a task file by ID
func (ba *boardAccess) findTaskFile(taskID string) (string, error) {
	repoPath := ba.repository.Path()
	var foundPath string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking on errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if this is a task file with our ID
		if ba.isTaskFileForID(info.Name(), taskID) {
			foundPath = path
			return fmt.Errorf("found") // Use error to stop walking
		}

		return nil
	})

	if foundPath != "" {
		return foundPath, nil
	}

	if err != nil && err.Error() == "found" {
		return foundPath, nil
	}

	return "", fmt.Errorf("task file not found for ID: %s", taskID)
}

// isTaskFileForID checks if a filename matches a task ID
func (ba *boardAccess) isTaskFileForID(filename, taskID string) bool {
	// Pattern: XXXX-task-{taskID}.json
	suffix := fmt.Sprintf("task-%s.json", taskID)
	return strings.HasSuffix(filename, suffix) && strings.Contains(filename, "-task-")
}

// writeTaskFile writes a task to a file with proper directory creation
func (ba *boardAccess) writeTaskFile(task *Task, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Serialize task to JSON
	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task to JSON: %w", err)
	}

	// Write task file
	if err := os.WriteFile(filePath, taskJSON, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// commitChange stages a file and commits with board configuration
func (ba *boardAccess) commitChange(filePath, message string) error {
	if err := ba.stageFiles([]string{filePath}); err != nil {
		return err
	}

	return ba.commitWithConfig(message)
}

// stageFiles stages multiple files for commit
func (ba *boardAccess) stageFiles(filePaths []string) error {
	repoPath := ba.repository.Path()
	var relPaths []string

	for _, filePath := range filePaths {
		relPath, err := filepath.Rel(repoPath, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		relPaths = append(relPaths, relPath)
	}

	return ba.repository.Stage(relPaths)
}

// commitWithConfig commits using the repository's configured git settings
func (ba *boardAccess) commitWithConfig(message string) error {
	_, err := ba.repository.Commit(message)
	return err
}

// moveTaskFile moves a task file to a new location with commit
func (ba *boardAccess) moveTaskFile(oldPath string, task *Task, newPath string, commitMessage string) error {
	if err := os.Remove(oldPath); err != nil {
		return fmt.Errorf("failed to remove old file: %w", err)
	}

	if err := ba.writeTaskFile(task, newPath); err != nil {
		return fmt.Errorf("failed to write new file: %w", err)
	}

	if err := ba.stageFiles([]string{oldPath, newPath}); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	return ba.commitWithConfig(commitMessage)
}

// retrieveTaskWithTimestamps retrieves a task with git-derived timestamps
func (ba *boardAccess) retrieveTaskWithTimestamps(taskID string) (*TaskWithTimestamps, error) {
	filePath, err := ba.findTaskFile(taskID)
	if err != nil {
		return nil, nil // Task not found
	}

	return ba.loadTaskWithTimestampsFromFile(filePath)
}

// loadTaskFromFile loads a task from a file path
func (ba *boardAccess) loadTaskFromFile(filePath string) (*Task, error) {
	taskJSON, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var task Task
	if err := json.Unmarshal(taskJSON, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task JSON: %w", err)
	}

	return &task, nil
}

// loadTaskWithTimestampsFromFile loads a task with timestamps and location-derived metadata
func (ba *boardAccess) loadTaskWithTimestampsFromFile(filePath string) (*TaskWithTimestamps, error) {
	task, err := ba.loadTaskFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// Derive priority and status from file path
	priority, status, err := ba.extractLocationFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract location from path: %w", err)
	}

	// Get timestamps from git
	repoPath := ba.repository.Path()
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	history, err := ba.repository.GetFileHistory(relPath, 1000) // Large limit to get all history
	if err != nil || len(history) == 0 {
		// Fallback if no history available
		now := time.Now()
		return &TaskWithTimestamps{
			Task:      task,
			Priority:  priority,
			Status:    status,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	// First commit is creation, last is most recent update
	createdAt := history[len(history)-1].Timestamp
	updatedAt := history[0].Timestamp

	return &TaskWithTimestamps{
		Task:      task,
		Priority:  priority,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// getAllTaskFiles recursively finds all task files
func (ba *boardAccess) getAllTaskFiles() ([]string, error) {
	repoPath := ba.repository.Path()
	var taskFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if this is a task file (pattern: XXXX-task-*.json)
		if ba.isTaskFile(info.Name()) {
			taskFiles = append(taskFiles, path)
		}

		return nil
	})

	return taskFiles, err
}

// isTaskFile checks if a filename is a task file
func (ba *boardAccess) isTaskFile(filename string) bool {
	return strings.Contains(filename, "-task-") && strings.HasSuffix(filename, ".json") && filename != "board.json"
}

// extractTaskIDFromPath extracts task ID from file path
func (ba *boardAccess) extractTaskIDFromPath(filePath string) string {
	fileName := filepath.Base(filePath)
	// Pattern: XXXX-task-{ID}.json
	parts := strings.Split(fileName, "-task-")
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSuffix(parts[1], ".json")
}

// extractLocationFromPath derives priority and status from file path
func (ba *boardAccess) extractLocationFromPath(filePath string) (Priority, WorkflowStatus, error) {
	repoPath := ba.repository.Path()
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		return Priority{}, WorkflowStatus{}, fmt.Errorf("failed to get relative path: %w", err)
	}

	// Parse path components
	parts := strings.Split(relPath, string(filepath.Separator))
	fileName := filepath.Base(filePath)

	// Extract position from filename (XXXX-task-ID.json)
	positionStr := ""
	if strings.Contains(fileName, "-task-") {
		positionStr = strings.Split(fileName, "-task-")[0]
	}

	position := 1 // Default position
	if pos, err := strconv.Atoi(positionStr); err == nil && pos > 0 {
		position = pos
	}

	var priority Priority
	var status WorkflowStatus

	if len(parts) >= 1 && parts[0] == "archived" {
		// Archived task
		status = WorkflowStatus{
			Column:   "archived",
			Section:  "",
			Position: position,
		}
		// Priority doesn't matter for archived tasks
		priority = Priority{Urgent: false, Important: false, Label: ""}

	} else if len(parts) >= 1 {
		// Active task - extract column and section
		columnDir := parts[0]

		// Parse column (format: "01_todo" -> "todo")
		column := columnDir
		if strings.Contains(columnDir, "_") {
			column = strings.SplitN(columnDir, "_", 2)[1]
		}

		section := ""
		if len(parts) >= 2 && parts[1] != fileName {
			section = parts[1]
		}

		status = WorkflowStatus{
			Column:   column,
			Section:  section,
			Position: position,
		}

		// Derive priority from section (for todo column)
		if section != "" {
			switch section {
			case "urgent-important":
				priority = Priority{Urgent: true, Important: true, Label: "urgent-important"}
			case "urgent-not-important":
				priority = Priority{Urgent: true, Important: false, Label: "urgent-not-important"}
			case "not-urgent-important":
				priority = Priority{Urgent: false, Important: true, Label: "not-urgent-important"}
			default:
				priority = Priority{Urgent: false, Important: false, Label: ""}
			}
		} else {
			// Default priority for non-todo columns
			priority = Priority{Urgent: false, Important: false, Label: ""}
		}
	} else {
		return Priority{}, WorkflowStatus{}, fmt.Errorf("unable to parse file path: %s", relPath)
	}

	return priority, status, nil
}

// taskMatchesCriteria checks if a task matches query criteria
func (ba *boardAccess) taskMatchesCriteria(taskWithTimestamps *TaskWithTimestamps, criteria *QueryCriteria) bool {
	if criteria == nil {
		return true
	}

	task := taskWithTimestamps.Task
	priority := taskWithTimestamps.Priority
	status := taskWithTimestamps.Status

	// Check archived status
	if criteria.Archived != nil {
		isArchived := status.Column == "archived"
		if *criteria.Archived != isArchived {
			return false
		}
	}

	// Check columns
	if len(criteria.Columns) > 0 && !ba.stringInSlice(status.Column, criteria.Columns) {
		return false
	}

	// Check sections
	if len(criteria.Sections) > 0 && !ba.stringInSlice(status.Section, criteria.Sections) {
		return false
	}

	// Check priority
	if criteria.Priority != nil {
		if priority.Urgent != criteria.Priority.Urgent ||
			priority.Important != criteria.Priority.Important {
			return false
		}
	}

	// Check tags
	if len(criteria.Tags) > 0 {
		for _, requiredTag := range criteria.Tags {
			if !ba.stringInSlice(requiredTag, task.Tags) {
				return false
			}
		}
	}

	// Check date range
	if criteria.DateRange != nil {
		var targetDate time.Time
		if criteria.DateType == "updated" {
			targetDate = taskWithTimestamps.UpdatedAt
		} else {
			targetDate = taskWithTimestamps.CreatedAt // Default to created
		}

		if criteria.DateRange.From != nil && targetDate.Before(*criteria.DateRange.From) {
			return false
		}
		if criteria.DateRange.To != nil && targetDate.After(*criteria.DateRange.To) {
			return false
		}
	}

	return true
}

// stringInSlice checks if a string is in a slice
func (ba *boardAccess) stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// loadBoardConfiguration loads board config from disk or returns default
func (ba *boardAccess) loadBoardConfiguration() (*BoardConfiguration, error) {
	configPath := filepath.Join(ba.repository.Path(), "board.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return ba.getDefaultBoardConfiguration(), nil
	}

	configJSON, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read board configuration: %w", err)
	}

	var config BoardConfiguration
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to parse board configuration JSON: %w", err)
	}

	return &config, nil
}

// saveBoardConfiguration saves board config to disk and commits
func (ba *boardAccess) saveBoardConfiguration(config *BoardConfiguration) error {
	configPath := filepath.Join(ba.repository.Path(), "board.json")

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal board configuration: %w", err)
	}

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write board configuration: %w", err)
	}

	return ba.commitChange(configPath, "Update board configuration")
}

// GetRulesData retrieves all rule-related context data in a single operation
func (ba *boardAccess) GetRulesData(taskID string, targetColumns []string) (*RulesData, error) {
	ba.mutex.RLock()
	defer ba.mutex.RUnlock()

	rulesData := &RulesData{
		WIPCounts:        make(map[string]int),
		ColumnTasks:      make(map[string][]*TaskWithTimestamps),
		ColumnEnterTimes: make(map[string]time.Time),
		BoardMetadata:    make(map[string]string),
	}

	// Get all active tasks for WIP counts and column tasks
	criteria := &QueryCriteria{
		Archived: func() *bool { b := false; return &b }(), // Not archived
	}

	allTasks, err := ba.FindTasks(criteria)
	if err != nil {
		return nil, fmt.Errorf("BoardAccess.GetRulesData failed to get tasks: %w", err)
	}

	// Build WIP counts and organize tasks by column
	for _, task := range allTasks {
		// Count for WIP
		rulesData.WIPCounts[task.Status.Column]++
		
		// Group tasks by column (only for requested columns)
		if len(targetColumns) == 0 || ba.containsColumn(targetColumns, task.Status.Column) {
			rulesData.ColumnTasks[task.Status.Column] = append(
				rulesData.ColumnTasks[task.Status.Column], task)
		}
	}

	// Get task history if taskID is provided
	if taskID != "" {
		history, err := ba.GetTaskHistory(taskID, 50)
		if err != nil {
			ba.logger.LogMessage(utilities.Warning, "BoardAccess", 
				fmt.Sprintf("Failed to get task history for %s: %v", taskID, err))
			// Continue without history rather than failing
		} else {
			rulesData.TaskHistory = history
			
			// Calculate column enter times for each target column
			for _, column := range targetColumns {
				enterTime := ba.findColumnEnterTime(history, column)
				if !enterTime.IsZero() {
					rulesData.ColumnEnterTimes[column] = enterTime
				}
			}
		}
	}

	// Get board metadata
	boardConfig, err := ba.GetBoardConfiguration()
	if err != nil {
		ba.logger.LogMessage(utilities.Warning, "BoardAccess", 
			fmt.Sprintf("Failed to get board configuration: %v", err))
		// Continue with default metadata
		rulesData.BoardMetadata["board_name"] = "Unknown Board"
	} else {
		rulesData.BoardMetadata["board_name"] = boardConfig.Name
		rulesData.BoardMetadata["columns"] = strings.Join(boardConfig.Columns, ",")
	}

	ba.logger.LogMessage(utilities.Debug, "BoardAccess", 
		fmt.Sprintf("Retrieved rules data: %d columns WIP, %d column task groups, %d history entries", 
			len(rulesData.WIPCounts), len(rulesData.ColumnTasks), len(rulesData.TaskHistory)))

	return rulesData, nil
}

// Helper method to check if a column is in the target list
func (ba *boardAccess) containsColumn(columns []string, target string) bool {
	for _, col := range columns {
		if col == target {
			return true
		}
	}
	return false
}

// Helper method to find when a task entered a specific column from history
func (ba *boardAccess) findColumnEnterTime(history []utilities.CommitInfo, targetColumn string) time.Time {
	if len(history) == 0 {
		return time.Time{}
	}

	// Search backwards through history to find when task entered target column
	for i := len(history) - 1; i >= 0; i-- {
		commit := history[i]
		// This is a simplified approach - in practice, we'd need to parse commit messages
		// or maintain column transition timestamps more explicitly
		if strings.Contains(commit.Message, fmt.Sprintf("to %s", targetColumn)) {
			return commit.Timestamp
		}
	}

	// Fallback to task creation time if no specific transition found
	return history[0].Timestamp
}

// getDefaultBoardConfiguration returns a default board configuration
func (ba *boardAccess) getDefaultBoardConfiguration() *BoardConfiguration {
	return &BoardConfiguration{
		Name:    "EisenKan Board",
		Columns: []string{"todo", "doing", "done"},
		Sections: map[string][]string{
			"todo": {"urgent-important", "urgent-not-important", "not-urgent-important"},
		},
		GitUser:  "BoardAccess",
		GitEmail: "boardaccess@eisenkan.local",
	}
}
