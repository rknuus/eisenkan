package dnd

import (
	"os/exec"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// SRS refs: BV-REQ-031..035; subtask moves between sections and columns (parent-anchored variant).
// Placeholder: subtask section-only and section+column moves for parent-anchored variant
func TestDnD_SubtaskMove_BetweenSectionsAndColumns_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed: %v", err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, string(out))
		}
	}
	run("init")
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Section-only subtask move: urgent-important -> not-urgent-important
	prev, _ := h.LastCommit(repo)
	oldRel := filepath.Join("todo", "urgent-important", "task-sample-1", "001-subtask-s1-1.json")
	newRel := filepath.Join("todo", "not-urgent-important", "task-sample-1", "001-subtask-s1-1.json")
	if out, err := exec.Command("mkdir", "-p", filepath.Join(repoRoot, filepath.Dir(newRel))).CombinedOutput(); err != nil {
		t.Fatalf("mkdir: %v\n%s", err, string(out))
	}
	if out, err := exec.Command("mv", filepath.Join(repoRoot, oldRel), filepath.Join(repoRoot, newRel)).CombinedOutput(); err != nil {
		t.Fatalf("mv: %v\n%s", err, string(out))
	}
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "subtask section move")
	curr, _ := h.LastCommit(repo)
	if err := h.AssertCommitAdvanced(prev, curr); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := h.AssertMoved(repo, oldRel, newRel); err != nil {
		t.Fatalf("moved: %v", err)
	}

	// Section+column: to doing/ (parent-anchored under task-sample-1)
	prev = curr
	oldRel = newRel
	newRel = filepath.Join("doing", "task-sample-1", "001-subtask-s1-1.json")
	if out, err := exec.Command("mkdir", "-p", filepath.Join(repoRoot, filepath.Dir(newRel))).CombinedOutput(); err != nil {
		t.Fatalf("mkdir2: %v\n%s", err, string(out))
	}
	if out, err := exec.Command("mv", filepath.Join(repoRoot, oldRel), filepath.Join(repoRoot, newRel)).CombinedOutput(); err != nil {
		t.Fatalf("mv2: %v\n%s", err, string(out))
	}
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "subtask section+column move")
	curr, _ = h.LastCommit(repo)
	if err := h.AssertCommitAdvanced(prev, curr); err != nil {
		t.Fatalf("commit2: %v", err)
	}
	if err := h.AssertMoved(repo, oldRel, newRel); err != nil {
		t.Fatalf("moved2: %v", err)
	}
}
