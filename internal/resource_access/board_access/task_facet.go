// Package board_access provides BoardAccess layer components implementing the iDesign methodology.
// This file implements the ITask facet for task and subtask operations.
package board_access

import (
	"github.com/rknuus/eisenkan/internal/utilities"
)

// ITask defines the interface for task and subtask operations
type ITask interface {
	// Task CRUD Operations
	CreateTask(task *Task, priority Priority, status WorkflowStatus, parentTaskID *string) (string, error)
	GetTasksData(taskIDs []string, includeHierarchy bool) ([]*TaskWithTimestamps, error)
	ListTaskIdentifiers(hierarchyFilter HierarchyFilter) ([]string, error)
	ChangeTaskData(taskID string, task *Task, priority Priority, status WorkflowStatus) error
	MoveTask(taskID string, priority Priority, status WorkflowStatus) error
	ArchiveTask(taskID string, cascadePolicy CascadePolicy) error
	RemoveTask(taskID string, cascadePolicy CascadePolicy) error
	FindTasks(criteria *QueryCriteria) ([]*TaskWithTimestamps, error)
	GetTaskHistory(taskID string, limit int) ([]utilities.CommitInfo, error)

	// Subtask Operations
	GetSubtasks(parentTaskID string) ([]*TaskWithTimestamps, error)
	GetParentTask(subtaskID string) (*TaskWithTimestamps, error)
}