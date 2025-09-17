// Package resource_access provides ResourceAccess layer components implementing the iDesign methodology.
// This file implements the IRules facet for rule engine helper operations.
package resource_access

// IRules defines the interface for rule engine helper operations
type IRules interface {
	// Rule Engine Helper Operations
	GetRulesData(taskID string, targetColumns []string) (*RulesData, error)
}