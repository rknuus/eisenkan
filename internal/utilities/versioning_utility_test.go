// Package utilities_test provides unit tests for VersioningUtility
package utilities

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to create test AuthorConfiguration
func testAuthorConfig() *AuthorConfiguration {
	return &AuthorConfiguration{
		User:  "Test Author",
		Email: "test@example.com",
	}
}

// TestVersioningUtility_InitializeRepository_FactoryFunction tests factory function availability
func TestVersioningUtility_InitializeRepository_FactoryFunction(t *testing.T) {
	// Test that the factory function is available and works
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "factory_test")
	
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Factory function failed: %v", err)
	}
	defer repo.Close()
	
	if repo == nil {
		t.Fatal("Factory function returned nil repository")
	}
}

// TestVersioningUtility_InitializeRepository_NewRepository tests creating new repository
func TestVersioningUtility_InitializeRepository_NewRepository(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test_repo")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Expected successful initialization, got error: %v", err)
	}
	defer repo.Close()

	if repo.Path() != repoPath {
		t.Errorf("Expected path %s, got %s", repoPath, repo.Path())
	}

	// Verify .git directory was created
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("Expected .git directory to be created")
	}
}

// TestVersioningUtility_InitializeRepository_ExistingRepository tests opening existing repository
func TestVersioningUtility_InitializeRepository_ExistingRepository(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "existing_repo")
	
	// Removed old utility pattern
	
	// Create repository first
	repo1, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to create initial repository: %v", err)
	}
	repo1.Close()

	// Open existing repository
	repo2, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Expected successful opening of existing repository, got error: %v", err)
	}
	defer repo2.Close()

	if repo2.Path() != repoPath {
		t.Errorf("Expected path %s, got %s", repoPath, repo2.Path())
	}
}

// TestVersioningUtility_InitializeRepository_InvalidPath tests invalid path handling
func TestVersioningUtility_InitializeRepository_InvalidPath(t *testing.T) {
	// Removed old utility pattern
	
	// Test with read-only parent directory (simulated)
	invalidPath := "/dev/null/invalid_repo"
	_, err := InitializeRepositoryWithConfig(invalidPath, testAuthorConfig())
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestVersioningUtility_GetRepositoryStatus tests repository status retrieval
func TestVersioningUtility_GetRepositoryStatus(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "status_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Test empty repository status
	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Failed to get repository status: %v", err)
	}

	if status == nil {
		t.Fatal("Expected status object, got nil")
	}

	// Create a test file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get status with untracked file
	status, err = repo.Status()
	if err != nil {
		t.Fatalf("Failed to get repository status: %v", err)
	}

	if len(status.UntrackedFiles) == 0 {
		t.Error("Expected untracked files, got none")
	}

	if !containsString(status.UntrackedFiles, "test.txt") {
		t.Error("Expected test.txt to be in untracked files")
	}
}

// TestVersioningUtility_StageChanges tests file staging
func TestVersioningUtility_StageChanges(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "stage_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Create test files
	testFile1 := filepath.Join(repoPath, "file1.txt")
	testFile2 := filepath.Join(repoPath, "file2.txt")
	
	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Stage all files
	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage changes: %v", err)
	}

	// Check status
	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status.StagedFiles) == 0 {
		t.Error("Expected staged files, got none")
	}
}

// TestVersioningUtility_StageChanges_SelectiveStaging tests pattern-based staging
func TestVersioningUtility_StageChanges_SelectiveStaging(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "selective_stage_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Create test files
	txtFile := filepath.Join(repoPath, "test.txt")
	mdFile := filepath.Join(repoPath, "readme.md")
	
	if err := os.WriteFile(txtFile, []byte("text content"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}
	if err := os.WriteFile(mdFile, []byte("# Readme"), 0644); err != nil {
		t.Fatalf("Failed to create md file: %v", err)
	}

	// Use the repo directly for staging to ensure files are visible
	err = repo.Stage([]string{"*.txt"})
	if err != nil {
		t.Fatalf("Failed to stage txt files: %v", err)
	}

	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}


	// Should have staged txt file but not md file
	if !containsString(status.StagedFiles, "test.txt") {
		t.Error("Expected test.txt to be staged")
	}
	if containsString(status.StagedFiles, "readme.md") {
		t.Error("Expected readme.md to NOT be staged")
	}
}

// TestVersioningUtility_CommitChanges tests commit creation
func TestVersioningUtility_CommitChanges(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "commit_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Create and stage a test file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage changes: %v", err)
	}

	// Create commit
	commitHash, err := repo.Commit("Initial commit")
	if err != nil {
		t.Fatalf("Failed to commit changes: %v", err)
	}

	if commitHash == "" {
		t.Error("Expected commit hash, got empty string")
	}

	if len(commitHash) != 40 { // SHA-1 hash length
		t.Errorf("Expected 40 character hash, got %d characters", len(commitHash))
	}
}

// TestVersioningUtility_GetRepositoryHistory tests commit history retrieval
func TestVersioningUtility_GetRepositoryHistory(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "history_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Test empty repository
	history, err := repo.GetHistory(10)
	if err != nil {
		t.Fatalf("Failed to get history from empty repo: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d commits", len(history))
	}

	// Create commits
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(repoPath, "file"+string(rune('0'+i))+".txt")
		content := []byte("content " + string(rune('0'+i)))
		
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		err = repo.Stage([]string{"."})
		if err != nil {
			t.Fatalf("Failed to stage changes %d: %v", i, err)
		}

		_, err = repo.Commit("Commit "+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("Failed to commit %d: %v", i, err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Get full history
	history, err = repo.GetHistory(0)
	if err != nil {
		t.Fatalf("Failed to get repository history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 commits, got %d", len(history))
	}

	// Verify commit info
	for _, commit := range history {
		if commit.ID == "" {
			t.Error("Expected commit ID, got empty")
		}
		if commit.Author != "Test Author" {
			t.Errorf("Expected author 'Test Author', got '%s'", commit.Author)
		}
		if commit.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", commit.Email)
		}
		if commit.Timestamp.IsZero() {
			t.Error("Expected timestamp, got zero time")
		}
		if commit.Message == "" {
			t.Error("Expected commit message, got empty")
		}
	}

	// Test with limit
	limitedHistory, err := repo.GetHistory(2)
	if err != nil {
		t.Fatalf("Failed to get limited history: %v", err)
	}

	if len(limitedHistory) != 2 {
		t.Errorf("Expected 2 commits with limit, got %d", len(limitedHistory))
	}
}

// TestVersioningUtility_GetRepositoryHistoryStream tests streaming commit history
func TestVersioningUtility_GetRepositoryHistoryStream(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "stream_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Create a commit
	testFile := filepath.Join(repoPath, "stream_test.txt")
	if err := os.WriteFile(testFile, []byte("stream content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage changes: %v", err)
	}

	_, err = repo.Commit("Stream test commit")
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test streaming
	commitChan := repo.GetHistoryStream()
	
	var receivedCommits []CommitInfo
	for commit := range commitChan {
		receivedCommits = append(receivedCommits, commit)
	}

	if len(receivedCommits) != 1 {
		t.Errorf("Expected 1 commit from stream, got %d", len(receivedCommits))
	}

	if len(receivedCommits) > 0 {
		commit := receivedCommits[0]
		if commit.Author != "Test Author" {
			t.Errorf("Expected author 'Test Author', got '%s'", commit.Author)
		}
		if !strings.Contains(commit.Message, "Stream test commit") {
			t.Errorf("Expected commit message to contain 'Stream test commit', got '%s'", commit.Message)
		}
	}
}

// TestVersioningUtility_GetFileHistory tests file-specific history
func TestVersioningUtility_GetFileHistory(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "file_history_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	testFile := filepath.Join(repoPath, "tracked_file.txt")
	otherFile := filepath.Join(repoPath, "other_file.txt")

	// Create initial commit with tracked file
	if err := os.WriteFile(testFile, []byte("version 1"), 0644); err != nil {
		t.Fatalf("Failed to create tracked file: %v", err)
	}
	if err := os.WriteFile(otherFile, []byte("other content"), 0644); err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage initial changes: %v", err)
	}

	_, err = repo.Commit("Initial commit")
	if err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Modify only the tracked file
	if err := os.WriteFile(testFile, []byte("version 2"), 0644); err != nil {
		t.Fatalf("Failed to modify tracked file: %v", err)
	}

	err = repo.Stage([]string{"tracked_file.txt"})
	if err != nil {
		t.Fatalf("Failed to stage tracked file: %v", err)
	}

	_, err = repo.Commit("Updated tracked file")
	if err != nil {
		t.Fatalf("Failed to commit tracked file update: %v", err)
	}

	// Get file history
	fileHistory, err := repo.GetFileHistory("tracked_file.txt", 0)
	if err != nil {
		t.Fatalf("Failed to get file history: %v", err)
	}

	if len(fileHistory) != 2 {
		t.Errorf("Expected 2 commits in file history, got %d", len(fileHistory))
	}

	// Test non-existent file
	nonExistentHistory, err := repo.GetFileHistory("nonexistent.txt", 0)
	if err != nil {
		t.Fatalf("Failed to get history for non-existent file: %v", err)
	}

	if len(nonExistentHistory) != 0 {
		t.Errorf("Expected 0 commits for non-existent file, got %d", len(nonExistentHistory))
	}
}

// TestVersioningUtility_GetFileDifferences tests file difference retrieval
func TestVersioningUtility_GetFileDifferences(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "diff_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	testFile := filepath.Join(repoPath, "diff_file.txt")

	// Create first commit
	if err := os.WriteFile(testFile, []byte("line 1\nline 2\nline 3\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage first version: %v", err)
	}

	hash1, err := repo.Commit("First version")
	if err != nil {
		t.Fatalf("Failed to commit first version: %v", err)
	}

	// Create second commit
	if err := os.WriteFile(testFile, []byte("line 1\nmodified line 2\nline 3\nline 4\n"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage second version: %v", err)
	}

	hash2, err := repo.Commit("Second version")
	if err != nil {
		t.Fatalf("Failed to commit second version: %v", err)
	}

	// Get differences
	diff, err := repo.GetFileDifferences(hash1, hash2)
	if err != nil {
		t.Fatalf("Failed to get file differences: %v", err)
	}

	if len(diff) == 0 {
		t.Error("Expected diff content, got empty")
	}

	diffString := string(diff)
	if !strings.Contains(diffString, "modified line 2") {
		t.Error("Expected diff to contain 'modified line 2'")
	}
	if !strings.Contains(diffString, "line 4") {
		t.Error("Expected diff to contain 'line 4'")
	}
}

// TestVersioningUtility_InvalidCommitHash tests error handling for invalid commit hashes
func TestVersioningUtility_InvalidCommitHash(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "invalid_hash_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Test with invalid hash
	_, err = repo.GetFileDifferences("invalid_hash", "another_invalid_hash")
	if err == nil {
		t.Error("Expected error for invalid commit hashes, got nil")
	}
}

// TestRepositoryHandle_StatusAndOperations tests repository repo operations
func TestRepositoryHandle_StatusAndOperations(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "repo_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Test path
	if repo.Path() != repoPath {
		t.Errorf("Expected path %s, got %s", repoPath, repo.Path())
	}

	// Create test file
	testFile := filepath.Join(repoPath, "repo_test.txt")
	if err := os.WriteFile(testFile, []byte("repo test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test status
	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Failed to get status from repo: %v", err)
	}

	if len(status.UntrackedFiles) == 0 {
		t.Error("Expected untracked files from repo status")
	}

	// Test staging
	err = repo.Stage([]string{"."})
	if err != nil {
		t.Fatalf("Failed to stage via repo: %v", err)
	}

	// Test commit
	commitHash, err := repo.Commit("Handle test commit")
	if err != nil {
		t.Fatalf("Failed to commit via repo: %v", err)
	}

	if commitHash == "" {
		t.Error("Expected commit hash from repo, got empty")
	}
}

// TestRepositoryHandle_ConflictDetection tests conflict detection
func TestRepositoryHandle_ConflictDetection(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "conflict_test")
	
	// Removed old utility pattern
	repo, err := InitializeRepositoryWithConfig(repoPath, testAuthorConfig())
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// For now, just test that conflict detection doesn't crash
	// Full conflict testing would require more complex git state manipulation
	status, err := repo.Status()
	if err != nil {
		t.Fatalf("Failed to get status for conflict test: %v", err)
	}

	// In a clean repo, there should be no conflicts
	if status.HasConflicts {
		t.Error("Expected no conflicts in clean repository")
	}
}

// Helper function to check if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}