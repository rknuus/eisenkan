package harness

import (
	"fmt"
	"path/filepath"
)

// ResolveTaskPath constructs a task filepath given column, optional section, position, and id.
func ResolveTaskPath(root, column, section string, position int, taskID string) string {
	name := fmt.Sprintf("%03d-task-%s.json", position, taskID)
	if section != "" {
		return filepath.Join(root, column, section, name)
	}
	return filepath.Join(root, column, name)
}

// ResolveSubtaskPathA constructs a subtask path for Variant A (parent-anchored).
func ResolveSubtaskPathA(root, column, section, parentID string, position int, subID string) string {
	name := fmt.Sprintf("%03d-subtask-%s.json", position, subID)
	if section != "" {
		return filepath.Join(root, column, section, "task-"+parentID, name)
	}
	return filepath.Join(root, column, "task-"+parentID, name)
}

// ResolveSubtaskPathB constructs a subtask path for Variant B (first-class files).
func ResolveSubtaskPathB(root, column, section string, position int, subID string) string {
	name := fmt.Sprintf("%03d-subtask-%s.json", position, subID)
	if section != "" {
		return filepath.Join(root, column, section, name)
	}
	return filepath.Join(root, column, name)
}
