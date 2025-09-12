// Package utilities provides core utilities implementing the iDesign methodology.
// This package contains utility components that provide essential services
// to higher-level components in the application architecture.
package utilities

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// AuthorConfiguration represents git user configuration
type AuthorConfiguration struct {
	User  string // Git commit author name
	Email string // Git commit author email
}

// CommitInfo represents information about a single commit
type CommitInfo struct {
	ID        string    // Commit hash
	Author    string    // Author name
	Email     string    // Author email
	Timestamp time.Time // Commit timestamp
	Message   string    // Commit message
}

// RepositoryStatus represents the current state of a repository
type RepositoryStatus struct {
	CurrentBranch  string   // Active branch name
	ModifiedFiles  []string // List of changed files
	StagedFiles    []string // List of files ready for commit
	UntrackedFiles []string // List of unversioned files
	HasConflicts   bool     // Indicates presence of merge conflicts
}

// InitializeRepositoryWithConfig initializes or opens a repository with git configuration
func InitializeRepositoryWithConfig(path string, gitConfig *AuthorConfiguration) (Repository, error) {
	logger := NewLoggingUtility()

	// Validate git configuration is provided
	if gitConfig == nil {
		return nil, fmt.Errorf("InitializeRepositoryWithConfig requires AuthorConfiguration - cannot be nil")
	}

	if gitConfig.User == "" || gitConfig.Email == "" {
		return nil, fmt.Errorf("InitializeRepositoryWithConfig requires complete AuthorConfiguration - both user (%s) and email (%s) must be non-empty", gitConfig.User, gitConfig.Email)
	}

	logger.Log(Debug, "VersioningUtility", "Initializing repository", map[string]interface{}{
		"path":  path,
		"user":  gitConfig.User,
		"email": gitConfig.Email,
	})

	// Validate path is not empty
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("VersioningUtility.InitializeRepository cannot initialize repository with empty path")
	}

	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("VersioningUtility.InitializeRepository failed to create directory: %w", err)
	}

	// Initialize or open Git repository
	var gitRepo *git.Repository
	var err error

	// Try to open existing repository first
	gitRepo, err = git.PlainOpen(path)
	if err != nil {
		// Repository doesn't exist, create new one
		gitRepo, err = git.PlainInit(path, false)
		if err != nil {
			return nil, fmt.Errorf("VersioningUtility.InitializeRepository failed to initialize Git repository: %w", err)
		}
	}

	repo := &repository{
		path:      path,
		gitRepo:   gitRepo,
		gitConfig: gitConfig,
		mutex:     &sync.RWMutex{},
		logger:    logger,
	}

	logger.Log(Info, "VersioningUtility", "Repository initialized successfully", map[string]interface{}{
		"path": path,
	})

	return repo, nil
}

// Repository provides version control operations for a specific repository
type Repository interface {
	Path() string
	Status() (*RepositoryStatus, error)
	Stage(patterns []string) error
	Commit(message string) (string, error)

	// Dual approach: limited sync + unlimited streaming
	GetHistory(limit int) ([]CommitInfo, error)
	GetHistoryStream() <-chan CommitInfo

	GetFileHistory(filePath string, limit int) ([]CommitInfo, error)
	GetFileHistoryStream(filePath string) <-chan CommitInfo

	GetFileDifferences(hash1, hash2 string) ([]byte, error)

	Close() error
}

// repository implements Repository
type repository struct {
	path      string
	gitRepo   *git.Repository
	gitConfig *AuthorConfiguration
	mutex     *sync.RWMutex
	logger    ILoggingUtility
}

// Repository Implementation

// Path returns the absolute path of the repository
func (r *repository) Path() string {
	return r.path
}

// Status returns the current status of the repository
func (r *repository) Status() (*RepositoryStatus, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.statusInternal()
}

// statusInternal returns repository status without acquiring locks (for internal use)
func (r *repository) statusInternal() (*RepositoryStatus, error) {
	workTree, err := r.gitRepo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("repository.Status failed to get worktree for %s: %w", r.path, err)
	}

	status, err := workTree.Status()
	if err != nil {
		return nil, fmt.Errorf("repository.Status failed to get status for %s: %w", r.path, err)
	}

	// Get current branch
	head, err := r.gitRepo.Head()
	var currentBranch string
	if err != nil {
		currentBranch = "HEAD" // Detached HEAD or empty repo
	} else {
		currentBranch = head.Name().Short()
	}

	// Parse status
	var modifiedFiles, stagedFiles, untrackedFiles []string
	hasConflicts := false

	for filename, fileStatus := range status {
		// Priority order: staged > modified > untracked
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked {
			stagedFiles = append(stagedFiles, filename)
		} else if fileStatus.Worktree == git.Untracked || fileStatus.Staging == git.Untracked {
			untrackedFiles = append(untrackedFiles, filename)
		} else if fileStatus.Worktree != git.Unmodified {
			modifiedFiles = append(modifiedFiles, filename)
		}

		// Check for conflicts
		if fileStatus.Staging == git.UpdatedButUnmerged {
			hasConflicts = true
		}
	}

	return &RepositoryStatus{
		CurrentBranch:  currentBranch,
		ModifiedFiles:  modifiedFiles,
		StagedFiles:    stagedFiles,
		UntrackedFiles: untrackedFiles,
		HasConflicts:   hasConflicts,
	}, nil
}

// Stage stages files matching the provided patterns
func (r *repository) Stage(patterns []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	workTree, err := r.gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("repository.Stage failed to get worktree for %s: %w", r.path, err)
	}

	// First check for conflicts to prevent staging conflicted files
	status, err := r.statusInternal()
	if err != nil {
		return fmt.Errorf("repository.Stage failed to get status for %s: %w", r.path, err)
	}

	if status.HasConflicts {
		return fmt.Errorf("repository.Stage cannot stage files while conflicts exist in %s", r.path)
	}

	// Stage matching files
	for _, pattern := range patterns {
		if pattern == "." {
			// Stage all files
			_, err := workTree.Add(".")
			if err != nil {
				return fmt.Errorf("repository.Stage failed to stage all files in %s: %w", r.path, err)
			}
		} else if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
			// Handle glob patterns by expanding them
			matches, err := filepath.Glob(filepath.Join(r.path, pattern))
			if err != nil {
				r.logger.LogError("Repository", err, map[string]interface{}{
					"operation": "Stage",
					"path":      r.path,
					"pattern":   pattern,
					"error":     "failed to expand glob pattern",
				})
				continue
			}

			// Stage each matched file
			for _, match := range matches {
				// Convert absolute path to relative path from repo root
				relPath, err := filepath.Rel(r.path, match)
				if err != nil {
					r.logger.LogError("Repository", err, map[string]interface{}{
						"operation": "Stage",
						"path":      r.path,
						"file":      match,
						"error":     "failed to get relative path",
					})
					continue
				}

				_, err = workTree.Add(relPath)
				if err != nil {
					r.logger.LogError("Repository", err, map[string]interface{}{
						"operation": "Stage",
						"path":      r.path,
						"file":      relPath,
						"error":     "failed to stage matched file",
					})
					continue
				}
			}
		} else {
			// Stage specific file path
			_, err := workTree.Add(pattern)
			if err != nil {
				r.logger.LogError("Repository", err, map[string]interface{}{
					"operation": "Stage",
					"path":      r.path,
					"pattern":   pattern,
					"error":     "failed to stage file",
				})
				// Continue with other patterns instead of failing entirely
				continue
			}
		}
	}

	r.logger.Log(Debug, "Repository", "Files staged successfully", map[string]interface{}{
		"path":     r.path,
		"patterns": patterns,
	})

	return nil
}

// Commit creates a commit with all staged changes
func (r *repository) Commit(message string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Validate that git configuration is available
	if r.gitConfig == nil {
		return "", fmt.Errorf("repository.Commit no git configuration available - repository must be initialized with AuthorConfiguration")
	}

	if r.gitConfig.User == "" || r.gitConfig.Email == "" {
		return "", fmt.Errorf("repository.Commit git configuration incomplete - both user (%s) and email (%s) are required", r.gitConfig.User, r.gitConfig.Email)
	}

	workTree, err := r.gitRepo.Worktree()
	if err != nil {
		return "", fmt.Errorf("repository.Commit failed to get worktree for %s: %w", r.path, err)
	}

	commitHash, err := workTree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  r.gitConfig.User,
			Email: r.gitConfig.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("repository.Commit failed to create commit in %s: %w", r.path, err)
	}

	r.logger.Log(Info, "Repository", "Commit created successfully", map[string]interface{}{
		"path":   r.path,
		"hash":   commitHash.String(),
		"author": r.gitConfig.User,
		"email":  r.gitConfig.Email,
	})

	return commitHash.String(), nil
}

// GetHistory returns a limited number of commits from repository history
func (r *repository) GetHistory(limit int) ([]CommitInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	ref, err := r.gitRepo.Head()
	if err != nil {
		// Repository might be empty
		return []CommitInfo{}, nil
	}

	commitIter, err := r.gitRepo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("repository.GetHistory failed to get commit log for %s: %w", r.path, err)
	}

	var commits []CommitInfo
	count := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached") // Use error to break iteration
		}

		commits = append(commits, CommitInfo{
			ID:        c.Hash.String(),
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
			Message:   c.Message,
		})
		count++
		return nil
	})

	// Ignore "limit reached" error
	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("repository.GetHistory failed to iterate commits for %s: %w", r.path, err)
	}

	return commits, nil
}

// GetHistoryStream returns a channel that streams all commits
func (r *repository) GetHistoryStream() <-chan CommitInfo {
	ch := make(chan CommitInfo)

	go func() {
		defer close(ch)

		r.mutex.RLock()
		defer r.mutex.RUnlock()

		ref, err := r.gitRepo.Head()
		if err != nil {
			r.logger.LogError("Repository", err, map[string]interface{}{
				"operation": "GetHistoryStream",
				"path":      r.path,
				"error":     "failed to get HEAD reference",
			})
			return
		}

		commitIter, err := r.gitRepo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			r.logger.LogError("Repository", err, map[string]interface{}{
				"operation": "GetHistoryStream",
				"path":      r.path,
				"error":     "failed to get commit log",
			})
			return
		}

		err = commitIter.ForEach(func(c *object.Commit) error {
			select {
			case ch <- CommitInfo{
				ID:        c.Hash.String(),
				Author:    c.Author.Name,
				Email:     c.Author.Email,
				Timestamp: c.Author.When,
				Message:   c.Message,
			}:
			default:
				// Channel closed, stop iteration
				return fmt.Errorf("channel closed")
			}
			return nil
		})

		if err != nil && err.Error() != "channel closed" {
			r.logger.LogError("Repository", err, map[string]interface{}{
				"operation": "GetHistoryStream",
				"path":      r.path,
				"error":     "failed to iterate commits",
			})
		}
	}()

	return ch
}

// GetFileHistory returns a limited number of commits that modified a specific file
func (r *repository) GetFileHistory(filePath string, limit int) ([]CommitInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	ref, err := r.gitRepo.Head()
	if err != nil {
		// Repository might be empty
		return []CommitInfo{}, nil
	}

	commitIter, err := r.gitRepo.Log(&git.LogOptions{
		From:     ref.Hash(),
		FileName: &filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("repository.GetFileHistory failed to get commit log for file %s in %s: %w", filePath, r.path, err)
	}

	var commits []CommitInfo
	count := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached")
		}

		commits = append(commits, CommitInfo{
			ID:        c.Hash.String(),
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
			Message:   c.Message,
		})
		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("repository.GetFileHistory failed to iterate commits for file %s in %s: %w", filePath, r.path, err)
	}

	return commits, nil
}

// GetFileHistoryStream returns a channel that streams commits for a specific file
func (r *repository) GetFileHistoryStream(filePath string) <-chan CommitInfo {
	ch := make(chan CommitInfo)

	go func() {
		defer close(ch)

		r.mutex.RLock()
		defer r.mutex.RUnlock()

		ref, err := r.gitRepo.Head()
		if err != nil {
			r.logger.LogError("Repository", err, map[string]interface{}{
				"operation": "GetFileHistoryStream",
				"path":      r.path,
				"filePath":  filePath,
				"error":     "failed to get HEAD reference",
			})
			return
		}

		commitIter, err := r.gitRepo.Log(&git.LogOptions{
			From:     ref.Hash(),
			FileName: &filePath,
		})
		if err != nil {
			r.logger.LogError("Repository", err, map[string]interface{}{
				"operation": "GetFileHistoryStream",
				"path":      r.path,
				"filePath":  filePath,
				"error":     "failed to get commit log",
			})
			return
		}

		err = commitIter.ForEach(func(c *object.Commit) error {
			select {
			case ch <- CommitInfo{
				ID:        c.Hash.String(),
				Author:    c.Author.Name,
				Email:     c.Author.Email,
				Timestamp: c.Author.When,
				Message:   c.Message,
			}:
			default:
				return fmt.Errorf("channel closed")
			}
			return nil
		})

		if err != nil && err.Error() != "channel closed" {
			r.logger.LogError("Repository", err, map[string]interface{}{
				"operation": "GetFileHistoryStream",
				"path":      r.path,
				"filePath":  filePath,
				"error":     "failed to iterate commits",
			})
		}
	}()

	return ch
}

// GetFileDifferences returns the differences between two commits
func (r *repository) GetFileDifferences(hash1, hash2 string) ([]byte, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Get commits
	commit1, err := r.gitRepo.CommitObject(plumbing.NewHash(hash1))
	if err != nil {
		return nil, fmt.Errorf("VersioningUtility.Repository.GetFileDifferences failed to get commit %s: %w", hash1, err)
	}

	commit2, err := r.gitRepo.CommitObject(plumbing.NewHash(hash2))
	if err != nil {
		return nil, fmt.Errorf("VersioningUtility.Repository.GetFileDifferences failed to get commit %s: %w", hash2, err)
	}

	// Get trees
	tree1, err := commit1.Tree()
	if err != nil {
		return nil, fmt.Errorf("VersioningUtility.Repository.GetFileDifferences failed to get tree for commit %s: %w", hash1, err)
	}

	tree2, err := commit2.Tree()
	if err != nil {
		return nil, fmt.Errorf("VersioningUtility.Repository.GetFileDifferences failed to get tree for commit %s: %w", hash2, err)
	}

	// Get patch
	patch, err := tree1.Patch(tree2)
	if err != nil {
		return nil, fmt.Errorf("VersioningUtility.Repository.GetFileDifferences failed to generate patch between %s and %s: %w", hash1, hash2, err)
	}

	return []byte(patch.String()), nil
}

// Close releases resources associated with the repository handle
func (r *repository) Close() error {
	// go-git repositories don't need explicit closing
	// but we can log the operation
	r.logger.Log(Debug, "VersioningUtility", "Repository handle closed", map[string]interface{}{
		"path": r.path,
	})
	return nil
}
