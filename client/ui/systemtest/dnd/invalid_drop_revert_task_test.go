package dnd

import (
	"os/exec"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"

	h "github.com/rknuus/eisenkan/client/ui/systemtest/harness"
)

// Placeholder invalid-drop: perform no file move and assert working tree stays clean
func TestDnD_InvalidDrop_Task_Reverts_Smoke(t *testing.T) {
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
	// Simulate invalid drop by making no changes
	if err := h.AssertWorkingTreeClean(repo); err != nil {
		t.Fatalf("expected clean tree: %v", err)
	}
}
