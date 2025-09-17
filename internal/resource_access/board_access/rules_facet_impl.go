// Package resource_access provides ResourceAccess layer components implementing the iDesign methodology.
// This file implements the IRules facet for rule engine helper operations.
package board_access

import (
	"sync"
	"time"

	"github.com/rknuus/eisenkan/internal/utilities"
)

// rulesFacet implements the IRules interface
type rulesFacet struct {
	taskFacet ITask
	logger    utilities.ILoggingUtility
	mutex     *sync.RWMutex
}

// newRulesFacet creates a new rules facet instance
func newRulesFacet(taskFacet ITask, logger utilities.ILoggingUtility, mutex *sync.RWMutex) IRules {
	return &rulesFacet{
		taskFacet: taskFacet,
		logger:    logger,
		mutex:     mutex,
	}
}

// GetRulesData aggregates all rule-related context data
func (rf *rulesFacet) GetRulesData(taskID string, targetColumns []string) (*RulesData, error) {
	rf.mutex.RLock()
	defer rf.mutex.RUnlock()

	// Initialize rules data structure
	rulesData := &RulesData{
		WIPCounts:        make(map[string]int),
		SubtaskWIPCounts: make(map[string]int),
		ColumnTasks:      make(map[string][]*TaskWithTimestamps),
		ColumnEnterTimes: make(map[string]time.Time),
		BoardMetadata:    make(map[string]string),
		HierarchyMap:     make(map[string][]string),
	}

	// Get all tasks
	allTasks, err := rf.taskFacet.FindTasks(&QueryCriteria{})
	if err != nil {
		return nil, err
	}

	// Process each task to build rules data
	for _, task := range allTasks {
		// Build WIP counts and organize tasks by column
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
		if len(targetColumns) == 0 || rf.containsString(targetColumns, task.Status.Column) {
			rulesData.ColumnTasks[task.Status.Column] = append(
				rulesData.ColumnTasks[task.Status.Column], task)
		}
	}

	// Get task history if taskID provided
	if taskID != "" {
		taskHistory, err := rf.taskFacet.GetTaskHistory(taskID, 10)
		if err != nil {
			rf.logger.LogMessage(utilities.Warning, "RulesFacet", "Failed to get task history")
		} else {
			rulesData.TaskHistory = taskHistory
		}

		// Mock column enter times based on task history
		for _, column := range targetColumns {
			rulesData.ColumnEnterTimes[column] = time.Now().Add(-time.Hour)
		}
	}

	// Add board metadata
	rulesData.BoardMetadata["board_name"] = "EisenKan Board"
	rulesData.BoardMetadata["wip_limit_enabled"] = "true"

	return rulesData, nil
}

// Helper function
func (rf *rulesFacet) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}