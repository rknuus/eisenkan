// Package board_access provides BoardAccess layer components implementing the iDesign methodology.
// This package contains components that provide data access and persistence services
// to higher-level components in the application architecture.
package board_access

import (
	"context"
	"time"
)

// BoardDiscoveryResult contains discovered board paths and metadata
type BoardDiscoveryResult struct {
	BoardPath    string            `json:"board_path"`
	IsValid      bool              `json:"is_valid"`
	HasGitRepo   bool              `json:"has_git_repo"`
	ConfigExists bool              `json:"config_exists"`
	Title        string            `json:"title,omitempty"`
	Issues       []string          `json:"issues,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// BoardMetadata provides comprehensive board information
type BoardMetadata struct {
	Title          string            `json:"title"`
	Description    string            `json:"description,omitempty"`
	CreatedAt      *time.Time        `json:"created_at,omitempty"`
	ModifiedAt     *time.Time        `json:"modified_at,omitempty"`
	TaskCount      int               `json:"task_count"`
	ColumnCounts   map[string]int    `json:"column_counts"`
	Configuration  *BoardConfiguration `json:"configuration,omitempty"`
	SchemaVersion  string            `json:"schema_version,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// BoardStatistics provides calculated metrics for board analysis
type BoardStatistics struct {
	TotalTasks        int               `json:"total_tasks"`
	ActiveTasks       int               `json:"active_tasks"`
	CompletedTasks    int               `json:"completed_tasks"`
	TasksByColumn     map[string]int    `json:"tasks_by_column"`
	TasksByPriority   map[string]int    `json:"tasks_by_priority"`
	AverageTaskAge    float64           `json:"average_task_age_days"`
	OldestTaskAge     float64           `json:"oldest_task_age_days"`
	LastActivity      *time.Time        `json:"last_activity,omitempty"`
	BoardHealthScore  float64           `json:"board_health_score"` // 0.0 - 1.0
}

// BoardValidationResult contains validation status and diagnostics
type BoardValidationResult struct {
	IsValid       bool                       `json:"is_valid"`
	Issues        []BoardValidationIssue     `json:"issues,omitempty"`
	Warnings      []BoardValidationIssue     `json:"warnings,omitempty"`
	GitRepoValid  bool                       `json:"git_repo_valid"`
	ConfigValid   bool                       `json:"config_valid"`
	DataIntegrity bool                       `json:"data_integrity"`
	SchemaVersion string                     `json:"schema_version,omitempty"`
}

// BoardValidationIssue represents a specific validation problem
type BoardValidationIssue struct {
	Severity    string `json:"severity"`    // "error", "warning", "info"
	Component   string `json:"component"`   // "git", "config", "data", "structure"
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// BoardDeletionRequest contains parameters for board deletion
type BoardDeletionRequest struct {
	BoardPath       string `json:"board_path"`
	UseTrash        bool   `json:"use_trash"`           // Use OS trash if available
	CreateBackup    bool   `json:"create_backup"`       // Create backup before deletion
	BackupLocation  string `json:"backup_location,omitempty"`
	ForceDelete     bool   `json:"force_delete"`        // Skip safety checks
}

// BoardDeletionResult contains deletion confirmation and details
type BoardDeletionResult struct {
	Success        bool   `json:"success"`
	Method         string `json:"method"`               // "trash", "permanent"
	BackupCreated  bool   `json:"backup_created"`
	BackupLocation string `json:"backup_location,omitempty"`
	Message        string `json:"message,omitempty"`
}

// BoardCreationRequest contains parameters for board creation
type BoardCreationRequest struct {
	BoardPath     string               `json:"board_path"`
	Title         string               `json:"title"`
	Description   string               `json:"description,omitempty"`
	Configuration *BoardConfiguration  `json:"configuration,omitempty"`
	InitializeGit bool                 `json:"initialize_git"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
}

// BoardCreationResult contains creation confirmation and details
type BoardCreationResult struct {
	Success      bool   `json:"success"`
	BoardPath    string `json:"board_path"`
	ConfigPath   string `json:"config_path"`
	GitInitialized bool `json:"git_initialized"`
	Message      string `json:"message,omitempty"`
}

// IBoard defines the interface for board management operations
type IBoard interface {
	// Board Discovery Operations
	DiscoverBoards(ctx context.Context, directoryPath string) ([]BoardDiscoveryResult, error)

	// Board Metadata Operations
	ExtractBoardMetadata(ctx context.Context, boardPath string) (*BoardMetadata, error)
	GetBoardStatistics(ctx context.Context, boardPath string) (*BoardStatistics, error)

	// Board Validation Operations
	ValidateBoardStructure(ctx context.Context, boardPath string) (*BoardValidationResult, error)

	// Board Configuration Operations
	LoadBoardConfiguration(ctx context.Context, boardPath string, configType string) (map[string]interface{}, error)
	StoreBoardConfiguration(ctx context.Context, boardPath string, configType string, configData map[string]interface{}) error

	// Board Lifecycle Operations
	CreateBoard(ctx context.Context, request *BoardCreationRequest) (*BoardCreationResult, error)
	DeleteBoard(ctx context.Context, request *BoardDeletionRequest) (*BoardDeletionResult, error)
}