// Package resource_access provides ResourceAccess layer components implementing the iDesign methodology.
// This file implements the ITask facet for task and subtask operations.
package resource_access

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rknuus/eisenkan/internal/utilities"
)

// taskFacet implements the ITask interface
type taskFacet struct {
	repository utilities.Repository
	logger     utilities.ILoggingUtility
	mutex      *sync.RWMutex
}

// newTaskFacet creates a new task facet instance
func newTaskFacet(repository utilities.Repository, logger utilities.ILoggingUtility, mutex *sync.RWMutex) ITask {
	return &taskFacet{
		repository: repository,
		logger:     logger,
		mutex:      mutex,
	}
}

// CreateTask implements task creation with hierarchical support
func (tf *taskFacet) CreateTask(task *Task, priority Priority, status WorkflowStatus, parentTaskID *string) (string, error) {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	if task == nil {
		return "", fmt.Errorf("task cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(task.Title) == "" {
		return "", fmt.Errorf("task title cannot be empty")
	}

	// Validate priority (exclude not-urgent-not-important)
	if !priority.Urgent && !priority.Important {
		return "", fmt.Errorf("not-urgent-not-important priority is not allowed")
	}

	// Set priority label based on urgent/important combination
	if priority.Urgent && priority.Important {
		priority.Label = "urgent-important"
	} else if priority.Urgent && !priority.Important {
		priority.Label = "urgent-not-important"
	} else if !priority.Urgent && priority.Important {
		priority.Label = "not-urgent-important"
	}

	// Generate unique task ID
	taskID := uuid.New().String()
	task.ID = taskID

	// Set parent task ID if provided
	if parentTaskID != nil {
		task.ParentTaskID = parentTaskID
	}

	// Create TaskWithTimestamps
	now := time.Now()
	taskWithTimestamps := &TaskWithTimestamps{
		Task:      task,
		Priority:  priority,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to storage
	if err := tf.saveTaskToStorage(taskWithTimestamps); err != nil {
		return "", fmt.Errorf("failed to save task: %w", err)
	}

	tf.logger.LogMessage(utilities.Info, "TaskFacet", fmt.Sprintf("Task created: %s", taskID))
	return taskID, nil
}

// GetTasksData retrieves tasks by IDs with optional hierarchy information
func (tf *taskFacet) GetTasksData(taskIDs []string, includeHierarchy bool) ([]*TaskWithTimestamps, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	var results []*TaskWithTimestamps

	// Load all tasks first
	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Create a map for quick lookup
	taskMap := make(map[string]*TaskWithTimestamps)
	for _, task := range allTasks {
		taskMap[task.Task.ID] = task
	}

	// Find requested tasks
	for _, taskID := range taskIDs {
		if task, exists := taskMap[taskID]; exists {
			results = append(results, task)
		}
	}

	return results, nil
}

// ListTaskIdentifiers returns all task IDs with optional hierarchy filtering
func (tf *taskFacet) ListTaskIdentifiers(hierarchyFilter HierarchyFilter) ([]string, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	var taskIDs []string
	for _, task := range allTasks {
		// Apply hierarchy filter
		switch hierarchyFilter {
		case TopLevelOnly:
			if task.Task.ParentTaskID == nil {
				taskIDs = append(taskIDs, task.Task.ID)
			}
		case SubtasksOnly:
			if task.Task.ParentTaskID != nil {
				taskIDs = append(taskIDs, task.Task.ID)
			}
		case AllTasks:
			taskIDs = append(taskIDs, task.Task.ID)
		}
	}

	return taskIDs, nil
}

// ChangeTaskData updates task data, priority, and status
func (tf *taskFacet) ChangeTaskData(taskID string, task *Task, priority Priority, status WorkflowStatus) error {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	// Load existing task
	existingTask, err := tf.getTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}
	if existingTask == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Update the task data
	task.ID = taskID // Ensure ID is preserved
	updatedTask := &TaskWithTimestamps{
		Task:      task,
		Priority:  priority,
		Status:    status,
		CreatedAt: existingTask.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Save updated task
	if err := tf.saveTaskToStorage(updatedTask); err != nil {
		return fmt.Errorf("failed to save updated task: %w", err)
	}

	tf.logger.LogMessage(utilities.Info, "TaskFacet", fmt.Sprintf("Task updated: %s", taskID))
	return nil
}

// MoveTask updates task status/priority (workflow transitions)
func (tf *taskFacet) MoveTask(taskID string, priority Priority, status WorkflowStatus) error {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	// Load existing task
	existingTask, err := tf.getTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}
	if existingTask == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Update only priority and status
	updatedTask := &TaskWithTimestamps{
		Task:      existingTask.Task,
		Priority:  priority,
		Status:    status,
		CreatedAt: existingTask.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Save updated task
	if err := tf.saveTaskToStorage(updatedTask); err != nil {
		return fmt.Errorf("failed to save moved task: %w", err)
	}

	tf.logger.LogMessage(utilities.Info, "TaskFacet", fmt.Sprintf("Task moved: %s", taskID))
	return nil
}

// ArchiveTask archives a task with cascade policy
func (tf *taskFacet) ArchiveTask(taskID string, cascadePolicy CascadePolicy) error {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	// Get the task to archive
	task, err := tf.getTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task for archival: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Handle cascade operations for subtasks
	if err := tf.handleCascadeOperation(taskID, cascadePolicy, true); err != nil {
		return fmt.Errorf("failed to handle cascade archival: %w", err)
	}

	// Archive the task (move to archived tasks file)
	if err := tf.archiveTaskInStorage(task); err != nil {
		return fmt.Errorf("failed to archive task in storage: %w", err)
	}

	tf.logger.LogMessage(utilities.Info, "TaskFacet", fmt.Sprintf("Task archived: %s", taskID))
	return nil
}

// RemoveTask permanently deletes a task with cascade policy
func (tf *taskFacet) RemoveTask(taskID string, cascadePolicy CascadePolicy) error {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	// Get the task to remove
	task, err := tf.getTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task for removal: %w", err)
	}
	if task == nil {
		// Task doesn't exist - idempotent operation
		return nil
	}

	// Handle cascade operations for subtasks
	if err := tf.handleCascadeOperation(taskID, cascadePolicy, false); err != nil {
		return fmt.Errorf("failed to handle cascade removal: %w", err)
	}

	// Remove the task from storage
	if err := tf.removeTaskFromStorage(taskID); err != nil {
		return fmt.Errorf("failed to remove task from storage: %w", err)
	}

	tf.logger.LogMessage(utilities.Info, "TaskFacet", fmt.Sprintf("Task removed: %s", taskID))
	return nil
}

// FindTasks searches for tasks based on criteria
func (tf *taskFacet) FindTasks(criteria *QueryCriteria) ([]*TaskWithTimestamps, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	if criteria == nil {
		criteria = &QueryCriteria{}
	}

	// Load all tasks
	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks for search: %w", err)
	}

	var results []*TaskWithTimestamps

	// Apply filters
	for _, task := range allTasks {
		if tf.matchesCriteria(task, criteria) {
			results = append(results, task)
		}
	}

	return results, nil
}

// GetTaskHistory retrieves task history from git commits
func (tf *taskFacet) GetTaskHistory(taskID string, limit int) ([]utilities.CommitInfo, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	// Use repository to get file history for this task
	// This is a simplified implementation - in practice, you'd track which file contains the task
	history, err := tf.repository.GetHistory(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get task history: %w", err)
	}

	// Filter commits that might be related to this task
	var taskHistory []utilities.CommitInfo
	for _, commit := range history {
		if strings.Contains(commit.Message, taskID) || strings.Contains(commit.Message, "task") {
			taskHistory = append(taskHistory, commit)
		}
	}

	return taskHistory, nil
}

// GetSubtasks retrieves all subtasks for a given parent task
func (tf *taskFacet) GetSubtasks(parentTaskID string) ([]*TaskWithTimestamps, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	return tf.getSubtasksInternal(parentTaskID)
}

// getSubtasksInternal retrieves subtasks without acquiring locks (for internal use)
func (tf *taskFacet) getSubtasksInternal(parentTaskID string) ([]*TaskWithTimestamps, error) {
	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks for subtask search: %w", err)
	}

	var subtasks []*TaskWithTimestamps
	for _, task := range allTasks {
		if task.Task.ParentTaskID != nil && *task.Task.ParentTaskID == parentTaskID {
			subtasks = append(subtasks, task)
		}
	}

	return subtasks, nil
}

// GetParentTask retrieves the parent task for a given subtask
func (tf *taskFacet) GetParentTask(subtaskID string) (*TaskWithTimestamps, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	// First find the subtask
	subtask, err := tf.getTaskByID(subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}
	if subtask == nil || subtask.Task.ParentTaskID == nil {
		return nil, nil // No parent or subtask doesn't exist
	}

	// Get the parent task
	parentTask, err := tf.getTaskByID(*subtask.Task.ParentTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent task: %w", err)
	}

	return parentTask, nil
}

// Helper methods

func (tf *taskFacet) getTaskByID(taskID string) (*TaskWithTimestamps, error) {
	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return nil, err
	}

	for _, task := range allTasks {
		if task.Task.ID == taskID {
			return task, nil
		}
	}
	return nil, nil
}

func (tf *taskFacet) loadAllTasks() ([]*TaskWithTimestamps, error) {
	// Load from active tasks file
	tasksFile := filepath.Join(tf.repository.Path(), "tasks.json")

	data, err := os.ReadFile(tasksFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*TaskWithTimestamps{}, nil
		}
		return nil, err
	}

	var tasks []*TaskWithTimestamps
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks file: %w", err)
	}

	return tasks, nil
}

func (tf *taskFacet) saveTaskToStorage(task *TaskWithTimestamps) error {
	// Load all existing tasks
	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return err
	}

	// Update or add the task
	found := false
	for i, existingTask := range allTasks {
		if existingTask.Task.ID == task.Task.ID {
			allTasks[i] = task
			found = true
			break
		}
	}
	if !found {
		allTasks = append(allTasks, task)
	}

	// Save back to file
	return tf.saveAllTasks(allTasks)
}

func (tf *taskFacet) saveAllTasks(tasks []*TaskWithTimestamps) error {
	tasksFile := filepath.Join(tf.repository.Path(), "tasks.json")

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(tasksFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	// Stage and commit changes
	if err := tf.repository.Stage([]string{"tasks.json"}); err != nil {
		return fmt.Errorf("failed to stage tasks file: %w", err)
	}

	_, err = tf.repository.Commit("Update tasks")
	return err
}

func (tf *taskFacet) removeTaskFromStorage(taskID string) error {
	allTasks, err := tf.loadAllTasks()
	if err != nil {
		return err
	}

	// Remove the task
	filteredTasks := make([]*TaskWithTimestamps, 0, len(allTasks))
	for _, task := range allTasks {
		if task.Task.ID != taskID {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return tf.saveAllTasks(filteredTasks)
}

func (tf *taskFacet) archiveTaskInStorage(task *TaskWithTimestamps) error {
	// Remove from active tasks
	if err := tf.removeTaskFromStorage(task.Task.ID); err != nil {
		return err
	}

	// Add to archived tasks (simplified - would use separate file in practice)
	tf.logger.LogMessage(utilities.Info, "TaskFacet", fmt.Sprintf("Task archived to archive storage: %s", task.Task.ID))
	return nil
}

func (tf *taskFacet) handleCascadeOperation(parentTaskID string, policy CascadePolicy, archive bool) error {
	// Get subtasks (internal version without locking)
	subtasks, err := tf.getSubtasksInternal(parentTaskID)
	if err != nil {
		return err
	}

	for _, subtask := range subtasks {
		switch policy {
		case NoAction:
			// Do nothing to subtasks
		case ArchiveSubtasks:
			if archive {
				if err := tf.ArchiveTask(subtask.Task.ID, policy); err != nil {
					return err
				}
			}
		case DeleteSubtasks:
			if err := tf.RemoveTask(subtask.Task.ID, policy); err != nil {
				return err
			}
		case PromoteSubtasks:
			// Clear parent task ID to promote to top level
			subtask.Task.ParentTaskID = nil
			if err := tf.saveTaskToStorage(subtask); err != nil {
				return err
			}
		}
	}

	return nil
}

func (tf *taskFacet) matchesCriteria(task *TaskWithTimestamps, criteria *QueryCriteria) bool {
	// Apply various filters based on criteria
	// This is a simplified implementation - expand as needed

	// Column filter
	if len(criteria.Columns) > 0 {
		hasColumn := false
		for _, column := range criteria.Columns {
			if task.Status.Column == column {
				hasColumn = true
				break
			}
		}
		if !hasColumn {
			return false
		}
	}

	// Priority filter
	if criteria.Priority != nil {
		if task.Priority.Urgent != criteria.Priority.Urgent || task.Priority.Important != criteria.Priority.Important {
			return false
		}
	}

	// Tags filter
	if len(criteria.Tags) > 0 {
		hasTag := false
		for _, criteriaTag := range criteria.Tags {
			for _, taskTag := range task.Task.Tags {
				if taskTag == criteriaTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// Parent task filter
	if criteria.ParentTaskID != nil {
		if task.Task.ParentTaskID == nil || *task.Task.ParentTaskID != *criteria.ParentTaskID {
			return false
		}
	}

	// Hierarchy filter
	switch criteria.Hierarchy {
	case TopLevelOnly:
		if task.Task.ParentTaskID != nil {
			return false
		}
	case SubtasksOnly:
		if task.Task.ParentTaskID == nil {
			return false
		}
	}

	// Priority promotion date filter
	if criteria.PriorityPromotionDate != nil {
		if task.Task.PriorityPromotionDate == nil {
			return false
		}

		if criteria.PriorityPromotionDate.From != nil && task.Task.PriorityPromotionDate.Before(*criteria.PriorityPromotionDate.From) {
			return false
		}

		if criteria.PriorityPromotionDate.To != nil && task.Task.PriorityPromotionDate.After(*criteria.PriorityPromotionDate.To) {
			return false
		}
	}

	return true
}