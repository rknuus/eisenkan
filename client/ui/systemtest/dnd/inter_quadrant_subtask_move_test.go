package dnd

import (
	"os/exec"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// SRS refs: BV-REQ-011..015, BV-REQ-033..035; parent-anchored variant relocation
// Placeholder DnD for subtask: parent-anchored variant relocation
func TestDnD_InterQuadrantSubtaskMove_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
		}
	}
	run("init")
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")

	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	prev, err := h.LastCommit(repo)
	if err != nil {
		t.Fatalf("last commit: %v", err)
	}

	// Move parent-anchored subtask directory to another quadrant section
	oldRel := filepath.Join("todo", "urgent-important", "task-sample-1", "001-subtask-s1-1.json")
	newRel := filepath.Join("todo", "not-urgent-important", "task-sample-1", "001-subtask-s1-1.json")
	// create target directory then move via shell tools
	if out, err := exec.Command("mkdir", "-p", filepath.Join(repoRoot, filepath.Dir(newRel))).CombinedOutput(); err != nil {
		t.Fatalf("mkdir failed: %v\n%s", err, string(out))
	}
	if out, err := exec.Command("mv", filepath.Join(repoRoot, oldRel), filepath.Join(repoRoot, newRel)).CombinedOutput(); err != nil {
		t.Fatalf("mv failed: %v\n%s", err, string(out))
	}
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "move subtask")

	curr, err := h.LastCommit(repo)
	if err != nil {
		t.Fatalf("last commit (after): %v", err)
	}
	if err := h.AssertCommitAdvanced(prev, curr); err != nil {
		t.Fatalf("commit not advanced: %v", err)
	}
	if err := h.AssertMoved(repo, oldRel, newRel); err != nil {
		t.Fatalf("repo move assert failed: %v", err)
	}
}
