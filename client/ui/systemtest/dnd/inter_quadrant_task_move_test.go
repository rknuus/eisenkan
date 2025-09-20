package dnd

import (
	"os/exec"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// SRS refs: BV-REQ-011..015, BV-REQ-033..035 (movement + validation); repo verification proxy for STR.
// Placeholder DnD happy-path: simulate a move by renaming a file in the repo
// and assert repo state using harness helpers. UI drag will be added later.
func TestDnD_InterQuadrantTaskMove_Smoke(t *testing.T) {
	repoRoot := t.TempDir()
	fixture := filepath.Join("..", "fixtures", "board_eisenhower_minimal")
	if err := h.SeedRepoFromFixture(repoRoot, fixture); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	// git init + initial commit
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

	// Simulate moving a task from urgent-important to not-urgent-important
	oldRel := filepath.Join("todo", "urgent-important", "001-task-sample-1.json")
	newRel := filepath.Join("todo", "not-urgent-important", "001-task-sample-1.json")
	run("mv", oldRel, newRel)
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "move task")

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
